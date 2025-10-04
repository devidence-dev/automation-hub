package processor

import (
	"fmt"

	"go.uber.org/zap"

	"automation-hub/internal/config"
	"automation-hub/internal/models"
	"automation-hub/internal/services/telegram"
)

type TorrentProcessor struct {
	telegram *telegram.Client
	logger   *zap.Logger
	config   *config.WebhookProcessorConfig
}

func NewTorrentProcessor(telegram *telegram.Client, webhookConfig *config.WebhookProcessorConfig, logger *zap.Logger) *TorrentProcessor {
	return &TorrentProcessor{
		telegram: telegram,
		logger:   logger,
		config:   webhookConfig,
	}
}

// NewTorrentProcessorLegacy maintains compatibility with the existing code
func NewTorrentProcessorLegacy(telegram *telegram.Client, chatID string, logger *zap.Logger) *TorrentProcessor {
	return &TorrentProcessor{
		telegram: telegram,
		logger:   logger,
		config: &config.WebhookProcessorConfig{
			TelegramChatID: chatID,
			TelegramMessage: `📥 **Download completed successfully!** 🎬

🔍 **Name:**  
%s

📍 **Path:**  
%s`,
		},
	}
}

func (p *TorrentProcessor) Process(notification models.TorrentNotification) error {
	message := fmt.Sprintf(p.config.TelegramMessage, notification.TorrentName, notification.SavePath)
	return p.telegram.SendMessage(p.config.TelegramChatID, message)
}

// GetWebhookConfig searches for the configuration of a specific webhook by name
func GetWebhookConfig(cfg *config.Config, webhookName string) *config.WebhookProcessorConfig {
	for _, hook := range cfg.Hook {
		if hook.Name == webhookName {
			return &hook.Config
		}
	}
	return nil
}
