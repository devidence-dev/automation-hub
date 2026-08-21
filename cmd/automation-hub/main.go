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

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"automation-hub/internal/config"
	"automation-hub/internal/handlers"
	"automation-hub/internal/services/email"
	"automation-hub/internal/services/github"
	"automation-hub/internal/services/processor"
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

	// Initialize services
	telegramClient := telegram.NewClient(cfg.Telegram.BotToken, logger)
	imapClient := email.NewIMAPClient(cfg.Email, logger)
	var workflowBotClient *telegram.Client
	var botHandler *handlers.BotHandler

	if len(cfg.Workflows) > 0 {
		if cfg.WorkflowBot.BotToken == "" {
			logger.Fatal("workflow_bot bot_token is required when workflows are configured")
		}
		if cfg.GitHub.Token == "" {
			logger.Fatal("GitHub token is required when workflows are configured")
		}
		workflowBotClient = telegram.NewClient(cfg.WorkflowBot.BotToken, logger)
		commands := make([]tgbotapi.BotCommand, 0, len(cfg.Workflows))
		for _, workflow := range cfg.Workflows {
			commands = append(commands, tgbotapi.BotCommand{Command: workflow.Command, Description: workflow.Description})
		}
		if err := workflowBotClient.SetMyCommands(commands); err != nil {
			logger.Fatal("Failed to set Telegram commands", zap.Error(err))
		}
		githubClient := github.NewClient(cfg.GitHub.Token, logger)
		botHandler = handlers.NewBotHandler(githubClient, workflowBotClient, cfg.Workflows, logger)
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

	// Register webhook routes dynamically from configuration
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

	logger.Info("Shutting down server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}
