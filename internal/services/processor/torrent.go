package processor

import (
	"go.uber.org/zap"

	"automation-hub/internal/config"
	"automation-hub/internal/models"
)

// TorrentProcessor sends the completion message of a torrent webhook through the hook's template.
type TorrentProcessor struct {
	notifier TemplateNotifier
	logger   *zap.Logger
}

func NewTorrentProcessor(sender messageSender, webhookConfig *config.WebhookProcessorConfig, logger *zap.Logger) *TorrentProcessor {
	return &TorrentProcessor{
		notifier: TemplateNotifier{
			sender:   sender,
			chatID:   webhookConfig.TelegramChatID,
			template: webhookConfig.TelegramMessage,
		},
		logger: logger,
	}
}

func (p *TorrentProcessor) Process(notification models.TorrentNotification) error {
	p.logger.Debug("Sending torrent notification", zap.String("torrent", notification.TorrentName))
	return p.notifier.Notify(notification.TorrentName, notification.SavePath)
}
