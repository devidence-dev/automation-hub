package email

import (
	"bytes"
	"testing"

	"github.com/emersion/go-imap"
	"go.uber.org/zap"

	"automation-hub/internal/config"
	"automation-hub/internal/models"
)

type mockNamedProcessor struct {
	name   string
	sender string
}

func (m *mockNamedProcessor) GetName() string {
	return m.name
}

func (m *mockNamedProcessor) GetSender() string {
	return m.sender
}

func (m *mockNamedProcessor) ShouldProcess(email models.Email) bool {
	return true
}

func (m *mockNamedProcessor) Process(email models.Email) error {
	return nil
}

func TestNewIMAPClient(t *testing.T) {
	logger := zap.NewNop()
	emailCfg := config.EmailConfig{
		Host:            "imap.test.com",
		Port:            993,
		PollingInterval: 15,
	}

	client := NewIMAPClient(emailCfg, logger)
	if client == nil {
		t.Fatal("Expected NewIMAPClient to return non-nil instance")
	}
}

func TestExtractTextPlain(t *testing.T) {
	logger := zap.NewNop()
	client := NewIMAPClient(config.EmailConfig{}, logger)

	if got := client.extractTextPlain(nil); got != "" {
		t.Errorf("Expected empty string for nil body, got %q", got)
	}

	buf := bytes.NewBufferString("Hello World Body")
	if got := client.extractTextPlain(buf); got != "Hello World Body" {
		t.Errorf("extractTextPlain() = %q, expected %q", got, "Hello World Body")
	}
}

func TestParseMessage(t *testing.T) {
	logger := zap.NewNop()
	client := NewIMAPClient(config.EmailConfig{}, logger)

	msg := &imap.Message{
		Envelope: &imap.Envelope{
			MessageId: "msg-123",
			Subject:   "Test Email Subject",
			From: []*imap.Address{
				{
					PersonalName: "Sender Name",
					MailboxName:  "support",
					HostName:     "service.com",
				},
			},
		},
	}

	email := client.parseMessage(msg)
	if email.ID != "msg-123" {
		t.Errorf("Expected ID msg-123, got %s", email.ID)
	}
	if email.Subject != "Test Email Subject" {
		t.Errorf("Expected Subject Test Email Subject, got %s", email.Subject)
	}
	if email.From != "support@service.com" {
		t.Errorf("Expected From support@service.com, got %s", email.From)
	}
}

func TestHandlePostProcessing(t *testing.T) {
	logger := zap.NewNop()
	client := NewIMAPClient(config.EmailConfig{}, logger)

	email := models.Email{Subject: "Test"}
	msg := &imap.Message{SeqNum: 1}

	// Processor without GetName interface
	structProcessor := &models.TorrentNotification{}
	client.handlePostProcessing(nil, nil, msg, email)
	_ = structProcessor

	// Processor named "perplexity"
	perplexityProc := &mockNamedProcessor{name: "perplexity", sender: "perplexity@test.com"}
	client.handlePostProcessing(nil, perplexityProc, msg, email)

	// Processor named "cloudflare"
	cfProc := &mockNamedProcessor{name: "cloudflare", sender: "cf@test.com"}
	client.handlePostProcessing(nil, cfProc, msg, email)

	// Processor named "other"
	otherProc := &mockNamedProcessor{name: "generic", sender: "other@test.com"}
	client.handlePostProcessing(nil, otherProc, msg, email)
}

func TestMarkAsReadAndUnreadNilClient(t *testing.T) {
	logger := zap.NewNop()
	client := NewIMAPClient(config.EmailConfig{}, logger)

	client.markAsRead(nil, 1)
	client.markAsUnread(nil, 1)
}
