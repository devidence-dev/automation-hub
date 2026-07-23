package processor

import (
	"testing"

	"go.uber.org/zap"

	"automation-hub/internal/config"
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
	if proc.config != webhookCfg {
		t.Errorf("Expected config to match webhookCfg")
	}
}

func TestNewTorrentProcessorLegacy(t *testing.T) {
	logger := zap.NewNop()
	proc := NewTorrentProcessorLegacy(nil, "999", logger)

	if proc == nil {
		t.Fatal("Expected NewTorrentProcessorLegacy to return non-nil instance")
	}
	if proc.config.TelegramChatID != "999" {
		t.Errorf("Expected TelegramChatID 999, got %s", proc.config.TelegramChatID)
	}
}

func TestGetWebhookConfig(t *testing.T) {
	cfg := &config.Config{
		Hook: []config.WebhookConfig{
			{
				Name: "qbittorrent",
				Path: "/webhook/qbittorrent",
				Config: config.WebhookProcessorConfig{
					TelegramChatID:  "123",
					TelegramMessage: "Download finished: %s",
				},
			},
			{
				Name: "sonarr",
				Path: "/webhook/sonarr",
				Config: config.WebhookProcessorConfig{
					TelegramChatID:  "456",
					TelegramMessage: "Series imported: %s",
				},
			},
		},
	}

	qbitHook := GetWebhookConfig(cfg, "qbittorrent")
	if qbitHook == nil {
		t.Fatalf("Expected qbittorrent webhook config, got nil")
	}
	if qbitHook.TelegramChatID != "123" {
		t.Errorf("Expected TelegramChatID 123, got %s", qbitHook.TelegramChatID)
	}

	nonExistentHook := GetWebhookConfig(cfg, "non_existent")
	if nonExistentHook != nil {
		t.Errorf("Expected nil for non_existent webhook, got %v", nonExistentHook)
	}
}
