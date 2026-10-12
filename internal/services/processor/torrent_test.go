package processor

import (
	"testing"

	"go.uber.org/zap"

	"automation-hub/internal/config"
	"automation-hub/internal/models"
)

func TestNewTorrentProcessor(t *testing.T) {
	logger := zap.NewNop()
	webhookCfg := &config.WebhookProcessorConfig{
		TelegramChatID:  "123",
		TelegramMessage: "Done %s",
	}

	proc := NewTorrentProcessor(nil, webhookCfg, logger)
	if proc == nil {
		t.Fatal("Expected NewTorrentProcessor to return non-nil instance")
	}
	if proc.notifier.chatID != "123" || proc.notifier.template != "Done %s" {
		t.Errorf("notifier = {chat %q, template %q}, want the hook's chat and template", proc.notifier.chatID, proc.notifier.template)
	}
}

func TestTorrentProcessor_Process(t *testing.T) {
	sender := &fakeSender{}
	webhookCfg := &config.WebhookProcessorConfig{
		TelegramChatID:  "123",
		TelegramMessage: "Download finished: %s at %s",
	}

	proc := NewTorrentProcessor(sender, webhookCfg, zap.NewNop())
	err := proc.Process(models.TorrentNotification{TorrentName: "Dune.2021.1080p", SavePath: "/downloads/movies"})
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if len(sender.messages) != 1 || sender.messages[0] != "Download finished: Dune.2021.1080p at /downloads/movies" {
		t.Errorf("sent %q, want the filled template", sender.messages)
	}
	if len(sender.chatIDs) != 1 || sender.chatIDs[0] != "123" {
		t.Errorf("chat IDs = %q, want [123]", sender.chatIDs)
	}
}
