package telegram

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type failingHTTPClient struct{}

func (failingHTTPClient) Do(*http.Request) (*http.Response, error) {
	return nil, errors.New("network unavailable")
}

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	return newClient("test-token", zap.NewNop(), server.Client(), server.URL+"/bot%s/%s")
}

func writeTelegramResponse(w http.ResponseWriter, result interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "result": result})
}

func TestParseInt64(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int64
		wantErr bool
	}{
		{
			name:    "Valid positive integer",
			input:   "123456789",
			want:    123456789,
			wantErr: false,
		},
		{
			name:    "Valid negative integer",
			input:   "-987654321",
			want:    -987654321,
			wantErr: false,
		},
		{
			name:    "Invalid string",
			input:   "not_a_number",
			want:    0,
			wantErr: true,
		},
		{
			name:    "Empty string",
			input:   "",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseInt64(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseInt64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSendMessageInvalidChatID(t *testing.T) {
	logger := zap.NewNop()
	client := &Client{
		bot:    &tgbotapi.BotAPI{},
		logger: logger,
	}

	err := client.SendMessage("invalid_chat_id", "Hello World")
	if err == nil {
		t.Error("Expected error when sending message with invalid chat ID, got nil")
	}
}

func TestSendMessageNilClientOrBot(t *testing.T) {
	logger := zap.NewNop()
	var nilClient *Client
	if err := nilClient.SendMessage("123456", "Hello"); err != nil {
		t.Errorf("Expected nil error for nil client, got %v", err)
	}

	clientWithNilBot := &Client{
		bot:    nil,
		logger: logger,
	}
	if err := clientWithNilBot.SendMessage("123456", "Hello"); err != nil {
		t.Errorf("Expected nil error for client with nil bot, got %v", err)
	}
}

func TestSendMessageSuccess(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/bottest-token/getMe":
			writeTelegramResponse(w, map[string]interface{}{"id": 1, "is_bot": true, "first_name": "Test"})
		case "/bottest-token/sendMessage":
			if err := r.ParseForm(); err != nil {
				t.Fatal(err)
			}
			if r.Form.Get("chat_id") != "123" || r.Form.Get("text") != "Hello" {
				t.Errorf("unexpected form: %v", r.Form)
			}
			writeTelegramResponse(w, map[string]interface{}{"message_id": 1, "date": 1, "chat": map[string]interface{}{"id": 123, "type": "private"}, "text": "Hello"})
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	})

	if err := client.SendMessage("123", "Hello"); err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}
}

func TestSendParseMode(t *testing.T) {
	tests := []struct {
		name     string
		send     func(*Client) error
		wantMode string
	}{
		{name: "SendMessage keeps Markdown", send: func(c *Client) error { return c.SendMessage("123", "<b>x</b>") }, wantMode: "Markdown"},
		{name: "SendHTML uses HTML", send: func(c *Client) error { return c.SendHTML("123", "<b>x</b>") }, wantMode: "HTML"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotMode, gotText string
			client := newTestClient(t, sendRecorder(t, &gotMode, &gotText))
			if err := tt.send(client); err != nil {
				t.Fatalf("send error = %v", err)
			}
			if gotMode != tt.wantMode || gotText != "<b>x</b>" {
				t.Errorf("parse_mode = %q text = %q, want %q and the text unchanged", gotMode, gotText, tt.wantMode)
			}
		})
	}
}

// sendRecorder answers getMe and sendMessage, and records the parse mode and text of each send.
func sendRecorder(t *testing.T, mode, text *string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/bottest-token/getMe":
			writeTelegramResponse(w, map[string]interface{}{"id": 1, "is_bot": true, "first_name": "Test"})
		case "/bottest-token/sendMessage":
			if err := r.ParseForm(); err != nil {
				t.Fatal(err)
			}
			*mode, *text = r.Form.Get("parse_mode"), r.Form.Get("text")
			writeTelegramResponse(w, map[string]interface{}{"message_id": 1, "date": 1, "chat": map[string]interface{}{"id": 123, "type": "private"}})
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}
}

func TestNewClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bottest-token/getMe" {
			t.Errorf("path = %s, want getMe", r.URL.Path)
		}
		writeTelegramResponse(w, map[string]interface{}{"id": 1, "is_bot": true, "first_name": "Test"})
	}))
	defer server.Close()

	previousEndpoint := telegramAPIEndpoint
	telegramAPIEndpoint = server.URL + "/bot%s/%s"
	t.Cleanup(func() { telegramAPIEndpoint = previousEndpoint })

	client := NewClient("test-token", zap.NewNop())
	if client == nil || client.bot == nil {
		t.Fatal("NewClient() returned an uninitialized client")
	}
}

func TestSendMessageReturnsErrorAfterRetries(t *testing.T) {
	client := &Client{bot: &tgbotapi.BotAPI{Client: failingHTTPClient{}}, logger: zap.NewNop()}

	if err := client.SendMessage("123", "Hello"); err == nil {
		t.Fatal("SendMessage() error = nil, want an error")
	}
}

func TestSetMyCommands(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/bottest-token/getMe":
			writeTelegramResponse(w, map[string]interface{}{"id": 1, "is_bot": true, "first_name": "Test"})
		case "/bottest-token/setMyCommands":
			if err := r.ParseForm(); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(r.Form.Get("commands"), "run_cleanup") {
				t.Errorf("commands = %q", r.Form.Get("commands"))
			}
			writeTelegramResponse(w, true)
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	})

	if err := client.SetMyCommands([]tgbotapi.BotCommand{{Command: "run_cleanup", Description: "Cleanup"}}); err != nil {
		t.Fatalf("SetMyCommands() error = %v", err)
	}
}

func TestSetMyCommandsReturnsTelegramError(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/bottest-token/getMe":
			writeTelegramResponse(w, map[string]interface{}{"id": 1, "is_bot": true, "first_name": "Test"})
		case "/bottest-token/setMyCommands":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "description": "Bad Request"})
		}
	})

	if err := client.SetMyCommands(nil); err == nil {
		t.Fatal("SetMyCommands() error = nil, want an error")
	}
}

func TestStartPollingNilClient(t *testing.T) {
	var client *Client
	client.StartPolling(context.Background(), func(tgbotapi.Update) {})
}

func TestStartPollingReceivesUpdatesUntilContextIsCancelled(t *testing.T) {
	requests := 0
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/bottest-token/getMe":
			writeTelegramResponse(w, map[string]interface{}{"id": 1, "is_bot": true, "first_name": "Test"})
		case "/bottest-token/getUpdates":
			requests++
			if requests == 1 {
				writeTelegramResponse(w, []interface{}{map[string]interface{}{"update_id": 1, "message": map[string]interface{}{"message_id": 1, "date": 1, "chat": map[string]interface{}{"id": 123, "type": "private"}, "text": "/run_cleanup"}}})
				return
			}
			writeTelegramResponse(w, []interface{}{})
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	received := make(chan tgbotapi.Update, 1)
	done := make(chan struct{})
	go func() {
		client.StartPolling(ctx, func(update tgbotapi.Update) {
			received <- update
			cancel()
		})
		close(done)
	}()

	select {
	case update := <-received:
		if update.UpdateID != 1 {
			t.Errorf("update ID = %d, want 1", update.UpdateID)
		}
	case <-time.After(time.Second):
		t.Fatal("polling did not receive the update")
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("polling did not stop after context cancellation")
	}
}

// maxDocumentBody bounds the multipart upload the fake endpoint reads.
const maxDocumentBody = 1 << 20

func TestSendDocumentSuccess(t *testing.T) {
	data := []byte("<html>\n<b>report</b> & more\n</html>")
	const caption = "<b>📦 Dependency report</b> · 2026-10-12"
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/bottest-token/getMe":
			writeTelegramResponse(w, map[string]interface{}{"id": 1, "is_bot": true, "first_name": "Test"})
		case strings.HasSuffix(r.URL.Path, "/sendDocument"):
			checkDocumentRequest(t, w, r, caption, data)
			writeTelegramResponse(w, map[string]interface{}{"message_id": 1, "date": 1, "chat": map[string]interface{}{"id": -1001234567890, "type": "supergroup"}})
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	})

	if err := client.SendDocument("-1001234567890", "dependency-report-2026-10-12.html", caption, data); err != nil {
		t.Fatalf("SendDocument() error = %v", err)
	}
}

// checkDocumentRequest checks the multipart form of a sendDocument request: chat, caption,
// parse mode, file name and file bytes.
func checkDocumentRequest(t *testing.T, w http.ResponseWriter, r *http.Request, caption string, data []byte) {
	t.Helper()
	r.Body = http.MaxBytesReader(w, r.Body, maxDocumentBody)
	mr, err := r.MultipartReader()
	if err != nil {
		t.Fatal(err)
	}
	form, err := mr.ReadForm(maxDocumentBody)
	if err != nil {
		t.Fatal(err)
	}
	checkFormValue(t, form, "chat_id", "-1001234567890")
	checkFormValue(t, form, "caption", caption)
	checkFormValue(t, form, "parse_mode", "HTML")
	files := form.File["document"]
	if len(files) != 1 || files[0].Filename != "dependency-report-2026-10-12.html" {
		t.Fatalf("document part = %+v", files)
	}
	checkFileBytes(t, files[0], data)
}

// checkFormValue checks that a multipart form has exactly one value for key, equal to want.
func checkFormValue(t *testing.T, form *multipart.Form, key, want string) {
	t.Helper()
	if got := form.Value[key]; len(got) != 1 || got[0] != want {
		t.Errorf("%s = %q, want %q", key, got, want)
	}
}

// checkFileBytes checks the content of an uploaded file part.
func checkFileBytes(t *testing.T, fh *multipart.FileHeader, want []byte) {
	t.Helper()
	f, err := fh.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	got, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Errorf("file bytes = %q, want %q", got, want)
	}
}

func TestSendDocumentInvalidChatID(t *testing.T) {
	client := &Client{bot: &tgbotapi.BotAPI{Client: failingHTTPClient{}}, logger: zap.NewNop()}
	if err := client.SendDocument("not-a-number", "r.html", "c", []byte("x")); err == nil {
		t.Fatal("SendDocument() error = nil, want an error for an invalid chat ID")
	}
}

func TestSendDocumentNilClient(t *testing.T) {
	var nilClient *Client
	if err := nilClient.SendDocument("123", "r.html", "c", []byte("x")); err != nil {
		t.Errorf("SendDocument() on nil client = %v, want nil", err)
	}
	if err := (&Client{logger: zap.NewNop()}).SendDocument("123", "r.html", "c", []byte("x")); err != nil {
		t.Errorf("SendDocument() on client without bot = %v, want nil", err)
	}
}

func TestSendDocumentReturnsErrorAfterRetries(t *testing.T) {
	attempts := 0
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/bottest-token/getMe":
			writeTelegramResponse(w, map[string]interface{}{"id": 1, "is_bot": true, "first_name": "Test"})
		case strings.HasSuffix(r.URL.Path, "/sendDocument"):
			attempts++
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "description": "Bad Request"})
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	})

	if err := client.SendDocument("123", "r.html", "c", []byte("x")); err == nil {
		t.Fatal("SendDocument() error = nil, want an error")
	}
	if attempts != 3 {
		t.Errorf("sendDocument attempts = %d, want 3", attempts)
	}
}
