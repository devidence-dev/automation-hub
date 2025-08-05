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

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"automation-hub/internal/config"
	"automation-hub/internal/handlers"
	"automation-hub/internal/services/email"
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

	// Initialize processor manager with dynamic configuration
	processorManager := processor.NewProcessorManager(cfg.Email, telegramClient, logger)

	// Initialize torrent processor (still using the old way for webhooks)
	torrentProc := processor.NewTorrentProcessor(telegramClient, cfg.Telegram.ChatIDs["torrent"], logger)

	// Start email monitoring with dynamic processors
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go imapClient.StartMonitoring(ctx, processorManager.GetProcessors()...)

	// Setup HTTP server for webhooks
	router := mux.NewRouter()
	webhookHandler := handlers.NewWebhookHandler(torrentProc, logger)
	router.HandleFunc("/webhook/qbitorrent", webhookHandler.HandleTorrentComplete).Methods("POST")

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
