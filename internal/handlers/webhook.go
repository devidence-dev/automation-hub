package handlers

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"automation-hub/internal/config"
	"automation-hub/internal/models"
	"automation-hub/internal/services/processor"
	"automation-hub/internal/services/telegram"
)

// WebhookHandler turns the webhook of one configured hook into a Telegram message.
type WebhookHandler struct {
	telegramClient *telegram.Client
	hook           *config.WebhookProcessorConfig
	logger         *zap.Logger
}

// NewWebhookHandler returns the handler for hook. The hook is read once at registration, not per request.
func NewWebhookHandler(telegramClient *telegram.Client, hook config.WebhookProcessorConfig, logger *zap.Logger) *WebhookHandler {
	return &WebhookHandler{
		telegramClient: telegramClient,
		hook:           &hook,
		logger:         logger,
	}
}

func (h *WebhookHandler) HandleTorrentComplete(w http.ResponseWriter, r *http.Request) {
	var notification models.TorrentNotification

	if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Create processor dynamically
	torrentProc := processor.NewTorrentProcessor(h.telegramClient, h.hook, h.logger)

	if err := torrentProc.Process(notification); err != nil {
		h.logger.Error("Failed to process torrent notification", zap.Error(err))
		http.Error(w, "Processing failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "success"}); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}
