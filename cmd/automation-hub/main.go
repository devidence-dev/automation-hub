package main

import (
	"context"
	"errors"
	_ "log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata" // embedded zone database: the image may have no /usr/share/zoneinfo

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"automation-hub/internal/config"
	"automation-hub/internal/handlers"
	"automation-hub/internal/services/email"
	"automation-hub/internal/services/github"
	"automation-hub/internal/services/kube"
	"automation-hub/internal/services/processor"
	"automation-hub/internal/services/scheduler"
	"automation-hub/internal/services/telegram"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer func(logger *zap.Logger) {
		_ = logger.Sync() // Ignore sync errors for stdout/stderr
	}(logger)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	// Schedules are evaluated in this fixed zone, whatever the host or image provides.
	loc, err := time.LoadLocation("America/Guayaquil")
	if err != nil {
		logger.Fatal("Failed to load timezone", zap.Error(err))
	}
	logger.Info("Timezone loaded", zap.String("timezone", loc.String()))

	// Initialize services
	telegramClient := telegram.NewClient(cfg.Telegram.BotToken, logger)
	imapClient := email.NewIMAPClient(cfg.Email, logger)
	var workflowBotClient *telegram.Client
	var botHandler *handlers.BotHandler
	var sched *scheduler.Scheduler

	if len(cfg.Workflows) > 0 || len(cfg.Restarts) > 0 || len(cfg.Jobs) > 0 {
		if cfg.WorkflowBot.BotToken == "" {
			logger.Fatal("workflow_bot bot_token is required when workflows, restarts or jobs are configured")
		}
		workflowBotClient = telegram.NewClient(cfg.WorkflowBot.BotToken, logger)

		commands := make([]tgbotapi.BotCommand, 0, len(cfg.Workflows)+len(cfg.Restarts)+len(cfg.Jobs))
		for _, workflow := range cfg.Workflows {
			commands = append(commands, tgbotapi.BotCommand{Command: workflow.Command, Description: workflow.Description})
		}
		for _, restart := range cfg.Restarts {
			commands = append(commands, tgbotapi.BotCommand{Command: restart.Command, Description: restart.Description})
		}
		for _, job := range cfg.Jobs {
			commands = append(commands, tgbotapi.BotCommand{Command: job.Command, Description: job.Description})
		}
		if err := workflowBotClient.SetMyCommands(commands); err != nil {
			logger.Fatal("Failed to set Telegram commands", zap.Error(err))
		}

		var githubClient *github.Client
		if len(cfg.Workflows) > 0 {
			if cfg.GitHub.Token == "" {
				logger.Fatal("GitHub token is required when workflows are configured")
			}
			githubClient = github.NewClient(cfg.GitHub.Token, logger)
		}

		var kubeClient *kube.Client
		if len(cfg.Restarts) > 0 || len(cfg.Jobs) > 0 {
			kubeClient, err = kube.NewInClusterClient(logger)
			if err != nil {
				logger.Fatal("Failed to build Kubernetes client for restarts or jobs", zap.Error(err))
			}
		}

		botHandler = handlers.NewBotHandler(githubClient, kubeClient, workflowBotClient, cfg.Workflows, cfg.Restarts, kubeClient, cfg.Jobs, logger)

		for _, workflow := range cfg.Workflows {
			if workflow.Schedule == "" {
				continue
			}
			if sched == nil {
				sched = scheduler.New(loc, logger)
			}
			if err := sched.Add(workflow.Command, workflow.Schedule, func() { botHandler.RunScheduled(workflow) }); err != nil {
				logger.Fatal("Failed to register scheduled workflow", zap.Error(err))
			}
		}
		for _, job := range cfg.Jobs {
			if job.Schedule == "" {
				continue
			}
			if sched == nil {
				sched = scheduler.New(loc, logger)
			}
			if err := sched.Add(job.Command, job.Schedule, func() { botHandler.RunScheduledJob(job) }); err != nil {
				logger.Fatal("Failed to register scheduled job", zap.Error(err))
			}
		}
		if sched != nil {
			sched.Start()
		}
	}

	// Initialize processor manager with dynamic configuration
	processorManager := processor.NewProcessorManager(cfg.Email, telegramClient, logger)

	// Start email monitoring with dynamic processors
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go imapClient.StartMonitoring(ctx, processorManager.GetProcessors()...)
	if workflowBotClient != nil {
		go workflowBotClient.StartPolling(ctx, botHandler.Handle)
	}

	// Setup HTTP server for webhooks
	router := mux.NewRouter()
	webhookHandler := handlers.NewWebhookHandler(telegramClient, cfg, logger)

	registerRoutes(router, cfg, webhookHandler, telegramClient, logger)

	srv := &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// Start server
	go func() {
		logger.Info("Starting server", zap.String("address", cfg.Server.Address))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	if sched != nil {
		logger.Info("Stopping scheduler...")
		sched.Stop()
	}
	logger.Info("Shutting down server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}

// reportSender is what the dependency-report webhook needs from the notifications bot.
type reportSender interface {
	SendHTML(chatID, text string) error
	SendDocument(chatID, filename string, data []byte) error
}

// registerRoutes wires the webhook routes from the configuration.
func registerRoutes(router *mux.Router, cfg *config.Config, webhookHandler *handlers.WebhookHandler, sender reportSender, logger *zap.Logger) {
	for _, hook := range cfg.Hook {
		switch hook.Name {
		case "qbittorrent":
			router.HandleFunc(hook.Path, webhookHandler.HandleTorrentComplete).Methods("POST")
			logger.Info("Registered webhook route",
				zap.String("name", hook.Name),
				zap.String("path", hook.Path))
		default:
			logger.Warn("Unknown webhook type", zap.String("name", hook.Name))
		}
	}

	// Not part of the switch above: the section is top-level with a fixed path.
	if cfg.DependencyReport != nil {
		router.Handle("/webhook/dependency-report", handlers.NewDependencyReportHandler(*cfg.DependencyReport, sender, logger)).Methods("POST")
		logger.Info("Registered webhook route", zap.String("name", "dependency-report"), zap.String("path", "/webhook/dependency-report"))
	}
}
