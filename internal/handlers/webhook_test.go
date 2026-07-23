package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"

	"automation-hub/internal/config"
)

func TestNewWebhookHandler(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{}
	handler := NewWebhookHandler(nil, cfg, logger)

	if handler == nil {
		t.Fatal("Expected NewWebhookHandler to return non-nil instance")
	}
}

func TestHandleTorrentComplete_InvalidJSON(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{}
	handler := NewWebhookHandler(nil, cfg, logger)

	req := httptest.NewRequest("POST", "/webhook/qbittorrent", bytes.NewBufferString("{invalid_json"))
	w := httptest.NewRecorder()

	handler.HandleTorrentComplete(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400 Bad Request, got %d", resp.StatusCode)
	}
}

func TestHandleTorrentComplete_MissingWebhookConfig(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{
		Hook: []config.WebhookConfig{}, // No qbittorrent hook configured
	}
	handler := NewWebhookHandler(nil, cfg, logger)

	validPayload := `{"torrent_name": "Ubuntu 24.04 ISO", "save_path": "/downloads"}`
	req := httptest.NewRequest("POST", "/webhook/qbittorrent", bytes.NewBufferString(validPayload))
	w := httptest.NewRecorder()

	handler.HandleTorrentComplete(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 500 Internal Server Error, got %d", resp.StatusCode)
	}
}

func TestHandleTorrentComplete_Success(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{
		Hook: []config.WebhookConfig{
			{
				Name: "qbittorrent",
				Path: "/webhook/qbittorrent",
				Config: config.WebhookProcessorConfig{
					TelegramChatID:  "123",
					TelegramMessage: "Downloaded: %s at %s",
				},
			},
		},
	}
	handler := NewWebhookHandler(nil, cfg, logger)

	validPayload := `{"torrent_name": "Debian ISO", "save_path": "/downloads/iso"}`
	req := httptest.NewRequest("POST", "/webhook/qbittorrent", bytes.NewBufferString(validPayload))
	w := httptest.NewRecorder()

	handler.HandleTorrentComplete(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d", resp.StatusCode)
	}
}
