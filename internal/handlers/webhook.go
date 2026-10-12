package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"go.uber.org/zap"

	"automation-hub/internal/config"
	"automation-hub/internal/models"
	"automation-hub/internal/services/processor"
)

// messageSender is the Telegram call the handler needs. *telegram.Client satisfies it.
type messageSender interface {
	SendMessage(chatID, message string) error
}

// WebhookHandler turns the webhook of one configured hook into a Telegram message.
type WebhookHandler struct {
	sender messageSender
	hook   *config.WebhookProcessorConfig
	logger *zap.Logger
}

// NewWebhookHandler returns the handler for hook. The hook is read once at registration, not per request.
func NewWebhookHandler(sender messageSender, hook config.WebhookProcessorConfig, logger *zap.Logger) *WebhookHandler {
	return &WebhookHandler{
		sender: sender,
		hook:   &hook,
		logger: logger,
	}
}

func (h *WebhookHandler) HandleTorrentComplete(w http.ResponseWriter, r *http.Request) {
	var notification models.TorrentNotification

	// Read the body up to the hook's cap first, so a body over the cap is rejected whole, even when its JSON is valid.
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, h.hook.EffectiveMaxBodyBytes()))
	if err != nil {
		h.rejectBody(w, err)
		return
	}

	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&notification); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	torrentProc := processor.NewTorrentProcessor(h.sender, h.hook, h.logger)

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

// rejectBody answers a body that could not be read: 413 when it is over the cap, 400 otherwise.
func (h *WebhookHandler) rejectBody(w http.ResponseWriter, err error) {
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		h.logger.Warn("Webhook body over the size cap", zap.Int64("max_body_bytes", tooLarge.Limit))
		http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
		return
	}
	h.logger.Error("Failed to read request", zap.Error(err))
	http.Error(w, "Invalid request", http.StatusBadRequest)
}
