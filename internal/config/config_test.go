package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestLoadSuccess(t *testing.T) {
	// Create a temporary directory for test configs
	tmpDir, err := os.MkdirTemp("", "config_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configContent := `
server:
  address: ":8080"
email:
  host: "imap.example.com"
  port: 993
  username: "user@example.com"
  password: "secretpassword"
  polling_interval: 30
  services:
    - name: "cloudflare"
      config:
        email_from: "no-reply@cloudflare.com"
        email_subject: ["Verification Code"]
        telegram_chat_id: "123456"
        telegram_message: "Code: %s"
telegram:
  bot_token: "test_bot_token"
  chat_ids:
    admin: "123456"
hook:
  - name: "qbittorrent"
    path: "/webhook/qbittorrent"
    config:
      telegram_chat_id: "123456"
      telegram_message: "Downloaded: %s"
`
	configFilePath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configFilePath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	viper.Reset()
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(tmpDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	if cfg.Server.Address != ":8080" {
		t.Errorf("Expected Server.Address :8080, got %s", cfg.Server.Address)
	}
	if cfg.Email.Host != "imap.example.com" {
		t.Errorf("Expected Email.Host imap.example.com, got %s", cfg.Email.Host)
	}
	if cfg.Email.Port != 993 {
		t.Errorf("Expected Email.Port 993, got %d", cfg.Email.Port)
	}
	if len(cfg.Email.Services) != 1 {
		t.Fatalf("Expected 1 service, got %d", len(cfg.Email.Services))
	}
	if cfg.Email.Services[0].Name != "cloudflare" {
		t.Errorf("Expected service name cloudflare, got %s", cfg.Email.Services[0].Name)
	}
	if cfg.Telegram.BotToken != "test_bot_token" {
		t.Errorf("Expected Telegram.BotToken test_bot_token, got %s", cfg.Telegram.BotToken)
	}
	if len(cfg.Hook) != 1 {
		t.Fatalf("Expected 1 hook, got %d", len(cfg.Hook))
	}
	if cfg.Hook[0].Name != "qbittorrent" {
		t.Errorf("Expected hook name qbittorrent, got %s", cfg.Hook[0].Name)
	}
}

func TestLoadErrorMissingConfig(t *testing.T) {
	viper.Reset()
	viper.SetConfigName("non_existent_config_file_name_12345")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(os.TempDir())

	_, err := Load()
	if err == nil {
		t.Error("Expected error when config file does not exist, got nil")
	}
}

func TestLoadInvalidYaml(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config_invalid_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	invalidContent := `
server:
  address: :8080
email:
  invalid_yaml: [unclosed bracket
`
	if err := os.WriteFile(filepath.Join(tmpDir, "config.yaml"), []byte(invalidContent), 0644); err != nil {
		t.Fatalf("Failed to write invalid config file: %v", err)
	}

	viper.Reset()
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(tmpDir)

	_, err = Load()
	if err == nil {
		t.Error("Expected error for invalid YAML syntax, got nil")
	}
}
