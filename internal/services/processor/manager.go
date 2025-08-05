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

	// Crear procesadores dinámicamente desde la configuración
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

	// Procesar cada email en su propia goroutine
	for _, email := range emails {
		pm.wg.Add(1)
		go pm.processEmailAsync(ctx, email)
	}

	// Esperar a que terminen todas las goroutines
	pm.wg.Wait()
}

func (pm *Manager) processEmailAsync(ctx context.Context, email models.Email) {
	defer pm.wg.Done()

	// Verificar si el contexto fue cancelado
	select {
	case <-ctx.Done():
		return
	default:
	}

	// Buscar un procesador que pueda manejar este email
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
			return // Solo el primer procesador que coincida debe procesar el email
		}
	}

	pm.logger.Info("Email ignored (no matching processor)",
		zap.String("subject", email.Subject),
		zap.String("from", email.From))
}
