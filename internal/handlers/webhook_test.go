package handlers

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"

	"automation-hub/internal/config"
	"automation-hub/internal/messages"
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
	handler := NewWebhookHandler(&webhookSender{}, hook, logger)

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

// webhookSender records the messages the handler sends, and fails them when err is set.
type webhookSender struct {
	messages []string
	err      error
}

func (s *webhookSender) SendMessage(chatID, message string) error {
	s.messages = append(s.messages, message)
	return s.err
}

func TestHandleTorrentCompleteProcessingError(t *testing.T) {
	sender := &webhookSender{err: errors.New("telegram unavailable")}
	hook := config.WebhookProcessorConfig{TelegramChatID: "123", TelegramMessage: "Downloaded: %s at %s"}
	handler := NewWebhookHandler(sender, hook, zap.NewNop())

	req := httptest.NewRequest("POST", "/webhook/qbittorrent", strings.NewReader(`{"torrent_name":"Debian ISO","save_path":"/downloads/iso"}`))
	w := httptest.NewRecorder()

	handler.HandleTorrentComplete(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
	if body := strings.TrimSpace(w.Body.String()); body != "Processing failed" {
		t.Errorf("body = %q, want Processing failed", body)
	}
}

// torrentBody returns a JSON body of exactly size bytes: the torrent name is padded with 'a'.
func torrentBody(size int) string {
	prefix := `{"torrent_name":"`
	suffix := `","save_path":"/d"}`
	return prefix + strings.Repeat("a", size-len(prefix)-len(suffix)) + suffix
}

func TestTorrentCompleteBodyCap(t *testing.T) {
	const defaultCap = 1 << 20
	tests := []struct {
		name     string
		maxBody  int64
		size     int
		wantCode int
	}{
		{name: "default cap accepts exactly 1 MiB", size: defaultCap, wantCode: http.StatusOK},
		{name: "default cap rejects one byte more", size: defaultCap + 1, wantCode: http.StatusRequestEntityTooLarge},
		{name: "configured cap rejects 2048 bytes", maxBody: 1024, size: 2048, wantCode: http.StatusRequestEntityTooLarge},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sender := &webhookSender{}
			hook := config.WebhookProcessorConfig{TelegramChatID: "123", TelegramMessage: "%s %s", MaxBodyBytes: tt.maxBody}
			handler := NewWebhookHandler(sender, hook, zap.NewNop())

			req := httptest.NewRequest("POST", "/webhook/qbittorrent", strings.NewReader(torrentBody(tt.size)))
			w := httptest.NewRecorder()
			handler.HandleTorrentComplete(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("status = %d, want %d", w.Code, tt.wantCode)
			}
			wantSent := 0
			if tt.wantCode == http.StatusOK {
				wantSent = 1
			}
			if len(sender.messages) != wantSent {
				t.Errorf("sent %d messages, want %d", len(sender.messages), wantSent)
			}
		})
	}
}

type metadataWebhookSender struct {
	photoErr, textErr        error
	photos, texts, templates int
}

func (s *metadataWebhookSender) SendMessage(string, string) error { s.templates++; return s.textErr }
func (s *metadataWebhookSender) SendPhoto(string, string, string) error {
	s.photos++
	return s.photoErr
}
func (s *metadataWebhookSender) SendHTML(string, string) error { s.texts++; return s.textErr }
func TestWebhookMetadata(t *testing.T) {
	for _, tt := range []struct {
		name, body                     string
		photoErr, textErr              error
		want, photos, texts, templates int
	}{
		{name: "full", body: `{"Response":"True","Title":"Dune","Poster":"https://example.com/poster"}`, want: 200, photos: 1},
		{name: "fallback", body: `{"Response":"False"}`, want: 200, templates: 1},
		{name: "send failure", body: `{"Response":"True","Title":"Dune","Poster":"https://example.com/poster"}`, photoErr: errors.New("failed"), textErr: errors.New("failed"), want: 500, photos: 1, texts: 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(tt.body)) }))
			defer server.Close()
			cat, err := messages.Load(nil)
			if err != nil {
				t.Fatal(err)
			}
			sender := &metadataWebhookSender{photoErr: tt.photoErr, textErr: tt.textErr}
			hook := config.WebhookConfig{Name: "qbittorrent", Path: "/webhook/qbittorrent", Config: config.WebhookProcessorConfig{TelegramChatID: "123", TelegramMessage: "%s %s", Metadata: &config.MetadataConfig{Provider: "omdb", APIKey: "secret", BaseURL: server.URL, Timeout: time.Second}}}
			route, err := newTorrentRoute(hook, RouteEnv{TorrentSender: sender, Catalog: cat, Logger: zap.NewNop()})
			if err != nil {
				t.Fatal(err)
			}
			req := httptest.NewRequest("POST", "/webhook/qbittorrent", strings.NewReader(`{"torrent_name":"Dune.2021.1080p","save_path":"/downloads"}`))
			w := httptest.NewRecorder()
			route.Handler.ServeHTTP(w, req)
			if w.Code != tt.want || sender.photos != tt.photos || sender.texts != tt.texts || sender.templates != tt.templates {
				t.Fatalf("status %d, sender %#v", w.Code, sender)
			}
		})
	}
}
