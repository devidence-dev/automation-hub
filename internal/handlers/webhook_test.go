package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"

	"automation-hub/internal/config"
)

func TestNewWebhookHandler(t *testing.T) {
	logger := zap.NewNop()
	handler := NewWebhookHandler(nil, config.WebhookProcessorConfig{}, logger)

	if handler == nil {
		t.Fatal("Expected NewWebhookHandler to return non-nil instance")
	}
}

func TestHandleTorrentComplete_InvalidJSON(t *testing.T) {
	logger := zap.NewNop()
	handler := NewWebhookHandler(nil, config.WebhookProcessorConfig{}, logger)

	req := httptest.NewRequest("POST", "/webhook/qbittorrent", bytes.NewBufferString("{invalid_json"))
	w := httptest.NewRecorder()

	handler.HandleTorrentComplete(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400 Bad Request, got %d", resp.StatusCode)
	}
}

func TestHandleTorrentComplete_Success(t *testing.T) {
	logger := zap.NewNop()
	hook := config.WebhookProcessorConfig{
		TelegramChatID:  "123",
		TelegramMessage: "Downloaded: %s at %s",
	}
	handler := NewWebhookHandler(nil, hook, logger)

	validPayload := `{"torrent_name": "Debian ISO", "save_path": "/downloads/iso"}`
	req := httptest.NewRequest("POST", "/webhook/qbittorrent", bytes.NewBufferString(validPayload))
	w := httptest.NewRecorder()

	handler.HandleTorrentComplete(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d", resp.StatusCode)
	}
	if body := strings.TrimSpace(w.Body.String()); body != `{"status":"success"}` {
		t.Errorf("body = %q, want the success status", body)
	}
}
