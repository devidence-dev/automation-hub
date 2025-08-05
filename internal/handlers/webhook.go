package handlers

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"automation-hub/internal/models"
	"automation-hub/internal/services/processor"
)

type WebhookHandler struct {
	torrentProcessor *processor.TorrentProcessor
	logger           *zap.Logger
}

func NewWebhookHandler(torrentProcessor *processor.TorrentProcessor, logger *zap.Logger) *WebhookHandler {
	return &WebhookHandler{
		torrentProcessor: torrentProcessor,
		logger:           logger,
	}
}

func (h *WebhookHandler) HandleTorrentComplete(w http.ResponseWriter, r *http.Request) {
	var notification models.TorrentNotification

	if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := h.torrentProcessor.Process(notification); err != nil {
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
