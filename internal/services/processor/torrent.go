package processor

import (
	"fmt"

	"go.uber.org/zap"

	"automation-hub/internal/models"
	"automation-hub/internal/services/telegram"
)

type TorrentProcessor struct {
	telegram *telegram.Client
	logger   *zap.Logger
	chatID   string
}

func NewTorrentProcessor(telegram *telegram.Client, chatID string, logger *zap.Logger) *TorrentProcessor {
	return &TorrentProcessor{
		telegram: telegram,
		logger:   logger,
		chatID:   chatID,
	}
}

func (p *TorrentProcessor) Process(notification models.TorrentNotification) error {
	message := fmt.Sprintf(`📥 **¡Descarga completada exitosamente!** 🎬

🔍 **Nombre:**  
%s

📍 **Ruta:**  
%s`, notification.TorrentName, notification.SavePath)

	return p.telegram.SendMessage(p.chatID, message)
}
