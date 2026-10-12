package processor

import (
	"go.uber.org/zap"

	"automation-hub/internal/config"
	"automation-hub/internal/models"
)

// Manager holds the email processors in config order and dispatches each email to the first one that matches.
type Manager struct {
	processors []models.EmailProcessor
	logger     *zap.Logger
}

// NewProcessorManager builds one processor per configured service. It fails when a service's regexes do not compile.
func NewProcessorManager(emailConfig config.EmailConfig, sender messageSender, logger *zap.Logger) (*Manager, error) {
	manager := &Manager{logger: logger}
	previewLength := emailConfig.EffectiveLogPreviewLength()

	for _, serviceConfig := range emailConfig.Services {
		processor, err := NewGenericEmailProcessor(serviceConfig.Name, serviceConfig.Config, previewLength, sender, logger)
		if err != nil {
			return nil, err
		}
		manager.processors = append(manager.processors, processor)
		logLoaded(logger, serviceConfig)
	}

	return manager, nil
}

// logLoaded logs the extractor of a service, and warns when it runs on the generic default pattern.
func logLoaded(logger *zap.Logger, service config.ServiceConfig) {
	extractor := service.Config.EffectiveExtractor()
	logger.Info("Loaded email processor",
		zap.String("service", service.Name),
		zap.String("email_from", service.Config.EmailFrom),
		zap.Strings("email_subjects", service.Config.EmailSubject),
		zap.String("extractor", extractor))
	if extractor == config.ExtractorPattern && service.Config.CodePattern == "" {
		logger.Warn("Email service uses the generic default code pattern",
			zap.String("service", service.Name),
			zap.String("code_pattern", config.DefaultCodePattern))
	}
}

// Senders returns the email_from of every service that sets one, in config order.
func (pm *Manager) Senders() []string {
	var senders []string
	for _, processor := range pm.processors {
		if sender := processor.GetSender(); sender != "" {
			senders = append(senders, sender)
		}
	}
	return senders
}

// Dispatch offers the email to the services in config order and processes it with the first one that
// matches. It returns that processor, or nil when none matches. A failed Process returns the processor with its error.
func (pm *Manager) Dispatch(email models.Email) (models.EmailProcessor, error) {
	for _, processor := range pm.processors {
		if !processor.ShouldProcess(email) {
			continue
		}
		pm.logger.Info("Processing email",
			zap.String("subject", email.Subject),
			zap.String("from", email.From))

		if err := processor.Process(email); err != nil {
			pm.logger.Error("Failed to process email",
				zap.String("subject", email.Subject),
				zap.Error(err))
			return processor, err
		}

		pm.logger.Info("Email processed successfully",
			zap.String("subject", email.Subject))
		return processor, nil
	}

	pm.logger.Info("Email ignored (no matching processor)",
		zap.String("subject", email.Subject),
		zap.String("from", email.From))
	return nil, nil
}
