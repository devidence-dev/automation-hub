package main

import (
	"context"
	"errors"
	"fmt"
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

	// The zone is validated by config.Load and loaded from the embedded database.
	loc := cfg.Location
	logger.Info("Timezone loaded", zap.String("timezone", loc.String()))

	// Initialize services
	telegramClient := newTelegramClient(cfg, cfg.Telegram.BotToken, logger)
	workflowBotClient, botHandler, sched := setupWorkflowBot(cfg, loc, logger)

	// Initialize processor manager with dynamic configuration
	processorManager, err := processor.NewProcessorManager(cfg.Email, telegramClient, logger)
	if err != nil {
		logger.Fatal("Failed to build email processors", zap.Error(err))
	}

	// Start email monitoring: the poller reads unread mail and dispatches it through the manager
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	poller := email.NewPoller(email.NewIMAPSource(cfg.Email, logger), processorManager, cfg.Email.EffectivePollingInterval(), logger)
	go poller.Run(ctx)
	if workflowBotClient != nil {
		go workflowBotClient.StartPolling(ctx, botHandler.Handle)
	}

	srv := newHTTPServer(cfg, telegramClient, logger)
	startServer(srv, logger)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownServer(srv, sched, cfg.Server.ShutdownTimeout, logger)
	logger.Info("Server exited")
}

// setupWorkflowBot builds the command bot and its handler when any workflow,
// restart or job is configured, registers their commands and schedules, and
// starts the scheduler. Without commands it returns nil values.
func setupWorkflowBot(cfg *config.Config, loc *time.Location, logger *zap.Logger) (*telegram.Client, *handlers.BotHandler, *scheduler.Scheduler) {
	if len(cfg.Workflows) == 0 && len(cfg.Restarts) == 0 && len(cfg.Jobs) == 0 {
		return nil, nil, nil
	}
	if cfg.WorkflowBot.BotToken == "" {
		logger.Fatal("workflow_bot bot_token is required when workflows, restarts or jobs are configured")
	}
	workflowBotClient := newTelegramClient(cfg, cfg.WorkflowBot.BotToken, logger)

	kubeClient := newKubeClient(cfg, logger)
	deps := handlers.BotHandlerDeps{
		GitHubClient:   newGitHubClient(cfg, logger),
		KubeClient:     kubeClient,
		TelegramClient: workflowBotClient,
		JobClient:      kubeClient,
		Workflows:      cfg.Workflows,
		Restarts:       cfg.Restarts,
		Jobs:           cfg.Jobs,
		GitHubWebURL:   cfg.GitHub.WebURL,
		RestartReady:   cfg.RestartReady,
		Catalog:        cfg.Catalog,
		Logger:         logger,
	}
	specs := handlers.NewCommandSpecs(deps)
	botHandler := handlers.NewBotHandler(deps, specs)

	lazy := &lazyScheduler{loc: loc, logger: logger}
	if err := registerCommands(workflowBotClient, lazy, handlers.NewScheduledRunner(deps), specs); err != nil {
		logger.Fatal("Failed to register bot commands", zap.Error(err))
	}
	sched := lazy.sched
	if sched != nil {
		sched.Start()
	}
	return workflowBotClient, botHandler, sched
}

// menuPublisher publishes the command menu to Telegram.
type menuPublisher interface {
	SetMyCommands(commands []tgbotapi.BotCommand) error
}

// scheduleAdder registers a named cron entry.
type scheduleAdder interface {
	Add(name, spec string, fn func()) error
}

// registerCommands makes one pass over specs: each command goes into the Telegram menu, and each
// scheduled one is registered through adder. The menu is published last, in config order.
func registerCommands(menu menuPublisher, adder scheduleAdder, runner *handlers.ScheduledRunner, specs []handlers.CommandSpec) error {
	commands := make([]tgbotapi.BotCommand, 0, len(specs))
	for _, spec := range specs {
		commands = append(commands, tgbotapi.BotCommand{Command: spec.Name(), Description: spec.Description()})
		if spec.Schedule() == "" {
			continue
		}
		if err := adder.Add(spec.Name(), spec.Schedule(), runner.Wrap(spec)); err != nil {
			return fmt.Errorf("schedule %q: %w", spec.Name(), err)
		}
	}
	if err := menu.SetMyCommands(commands); err != nil {
		return fmt.Errorf("set Telegram commands: %w", err)
	}
	return nil
}

// lazyScheduler creates the scheduler when the first schedule is added, so no scheduler exists
// when no command has a schedule.
type lazyScheduler struct {
	loc    *time.Location
	logger *zap.Logger
	sched  *scheduler.Scheduler
}

func (l *lazyScheduler) Add(name, spec string, fn func()) error {
	if l.sched == nil {
		l.sched = scheduler.New(l.loc, l.logger)
	}
	return l.sched.Add(name, spec, fn)
}

// newTelegramClient builds a client for token. The notifications bot and the workflow bot share the
// telegram.http settings, each with its own token.
func newTelegramClient(cfg *config.Config, token string, logger *zap.Logger) *telegram.Client {
	h := cfg.Telegram.HTTP
	return telegram.NewClient(telegram.Config{
		Token:         token,
		Logger:        logger,
		HTTP:          h.Settings(),
		PollTimeout:   h.PollTimeout,
		MaxAttempts:   h.MaxAttempts,
		RetryBackoff:  h.RetryBackoff,
		MaxRetryAfter: h.MaxRetryAfter,
	})
}

// newGitHubClient returns nil, an untyped nil interface, when no workflow is configured.
func newGitHubClient(cfg *config.Config, logger *zap.Logger) handlers.WorkflowDispatcher {
	if len(cfg.Workflows) == 0 {
		return nil
	}
	if cfg.GitHub.Token == "" {
		logger.Fatal("GitHub token is required when workflows are configured")
	}
	return github.NewClient(github.Config{
		Token:            cfg.GitHub.Token,
		APIURL:           cfg.GitHub.APIURL,
		HTTP:             cfg.GitHub.HTTP.Settings(),
		MaxResponseBytes: cfg.GitHub.HTTP.MaxResponseBytes,
	}, logger)
}

// newKubeClient returns nil, an untyped nil interface, when no restart or job is configured.
func newKubeClient(cfg *config.Config, logger *zap.Logger) handlers.KubeClient {
	if len(cfg.Restarts) == 0 && len(cfg.Jobs) == 0 {
		return nil
	}
	kubeClient, err := kube.NewInClusterClient(logger, cfg.Kubernetes.HTTP.Settings(), cfg.Kubernetes.HTTP.MaxResponseBytes)
	if err != nil {
		logger.Fatal("Failed to build Kubernetes client for restarts or jobs", zap.Error(err))
	}
	return kubeClient
}

// newHTTPServer wires the webhook routes and returns the server, not yet listening.
func newHTTPServer(cfg *config.Config, telegramClient *telegram.Client, logger *zap.Logger) *http.Server {
	router := mux.NewRouter()
	registerRoutes(router, cfg, handlers.RouteEnv{
		Telegram:         telegramClient,
		DependencyReport: cfg.DependencyReport,
		Catalog:          cfg.Catalog,
		Location:         cfg.Location,
		Sender:           telegramClient,
		Logger:           logger,
	})

	return &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}
}

// startServer listens in the background; a listen failure is fatal.
func startServer(srv *http.Server, logger *zap.Logger) {
	go func() {
		logger.Info("Starting server", zap.String("address", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()
}

// shutdownServer stops the scheduler, then the HTTP server, within the shutdown timeout.
func shutdownServer(srv *http.Server, sched *scheduler.Scheduler, timeout time.Duration, logger *zap.Logger) {
	if sched != nil {
		logger.Info("Stopping scheduler...")
		sched.Stop()
	}
	logger.Info("Shutting down server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), timeout)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}
}

// routeEntry is one route to mount: a hook entry, or the dependency_report section.
type routeEntry struct {
	hook        config.WebhookConfig
	fromSection bool
}

// registerRoutes mounts a POST route for each configured hook, plus the dependency report when its
// section is set. Each route comes from the factory registered for its name and source.
func registerRoutes(router *mux.Router, cfg *config.Config, env handlers.RouteEnv) {
	entries := make([]routeEntry, 0, len(cfg.Hook)+1)
	for _, hook := range cfg.Hook {
		entries = append(entries, routeEntry{hook: hook})
	}
	// The section is top-level and its path comes from its settings, not from a hook entry.
	if cfg.DependencyReport != nil {
		entries = append(entries, routeEntry{hook: config.WebhookConfig{Name: handlers.DependencyReportRoute}, fromSection: true})
	}
	for _, entry := range entries {
		mountRoute(router, entry, env)
	}
}

// mountRoute builds the route of an entry through its factory and mounts it for POST. An unknown
// name, or a name the entry's source may not use, is logged and skipped.
func mountRoute(router *mux.Router, entry routeEntry, env handlers.RouteEnv) {
	hook := entry.hook
	factory, ok := handlers.LookupRouteFactory(hook.Name, entry.fromSection)
	if !ok {
		env.Logger.Warn("Unknown webhook type", zap.String("name", hook.Name))
		return
	}
	route, err := factory(hook, env)
	if err != nil {
		env.Logger.Warn("Skipped webhook route", zap.String("name", hook.Name), zap.Error(err))
		return
	}
	router.Handle(route.Path, route.Handler).Methods("POST")
	env.Logger.Info("Registered webhook route", zap.String("name", hook.Name), zap.String("path", route.Path))
}
