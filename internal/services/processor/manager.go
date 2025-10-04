package processor

import (
	"context"
	"sync"

	"go.uber.org/zap"

	"automation-hub/internal/config"
	"automation-hub/internal/models"
	"automation-hub/internal/services/telegram"
)

type Manager struct {
	processors []models.EmailProcessor
	telegram   *telegram.Client
	logger     *zap.Logger
	wg         sync.WaitGroup
}

func NewProcessorManager(emailConfig config.EmailConfig, telegram *telegram.Client, logger *zap.Logger) *Manager {
	manager := &Manager{
		telegram: telegram,
		logger:   logger,
	}

	// Create processors dynamically from the configuration
	for _, serviceConfig := range emailConfig.Services {
		processor := NewGenericEmailProcessor(
			serviceConfig.Name,
			serviceConfig.Config,
			telegram,
			logger,
		)
		manager.processors = append(manager.processors, processor)
		logger.Info("Loaded email processor",
			zap.String("service", serviceConfig.Name),
			zap.String("email_from", serviceConfig.Config.EmailFrom),
			zap.Strings("email_subjects", serviceConfig.Config.EmailSubject))
	}

	return manager
}

func (pm *Manager) GetProcessors() []models.EmailProcessor {
	return pm.processors
}

func (pm *Manager) ProcessEmailsConcurrently(ctx context.Context, emails []models.Email) {
	if len(emails) == 0 {
		return
	}

	pm.logger.Info("Processing emails concurrently", zap.Int("email_count", len(emails)))

	// Process each email in its own goroutine
	for _, email := range emails {
		pm.wg.Add(1)
		go pm.processEmailAsync(ctx, email)
	}

	// Wait for all goroutines to finish
	pm.wg.Wait()
}

func (pm *Manager) processEmailAsync(ctx context.Context, email models.Email) {
	defer pm.wg.Done()

	// Check if the context was canceled
	select {
	case <-ctx.Done():
		return
	default:
	}

	// Find a processor that can handle this email
	for _, processor := range pm.processors {
		if processor.ShouldProcess(email) {
			pm.logger.Info("Processing email",
				zap.String("subject", email.Subject),
				zap.String("from", email.From))

			if err := processor.Process(email); err != nil {
				pm.logger.Error("Failed to process email",
					zap.String("subject", email.Subject),
					zap.Error(err))
			} else {
				pm.logger.Info("Email processed successfully",
					zap.String("subject", email.Subject))
			}
			return // Only the first matching processor should process the email
		}
	}

	pm.logger.Info("Email ignored (no matching processor)",
		zap.String("subject", email.Subject),
		zap.String("from", email.From))
}
