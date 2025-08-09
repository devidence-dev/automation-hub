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

type WebhookHandler struct {
	telegramClient *telegram.Client
	config         *config.Config
	logger         *zap.Logger
}

func NewWebhookHandler(telegramClient *telegram.Client, config *config.Config, logger *zap.Logger) *WebhookHandler {
	return &WebhookHandler{
		telegramClient: telegramClient,
		config:         config,
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

	// Buscar configuración del webhook qbittorrent
	webhookConfig := processor.GetWebhookConfig(h.config, "qbittorrent")
	if webhookConfig == nil {
		h.logger.Error("qbittorrent webhook configuration not found")
		http.Error(w, "Webhook configuration not found", http.StatusInternalServerError)
		return
	}

	// Crear procesador dinámicamente
	torrentProc := processor.NewTorrentProcessor(h.telegramClient, webhookConfig, h.logger)

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
