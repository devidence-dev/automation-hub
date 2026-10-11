package config

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestLoadSuccess(t *testing.T) {
	// Create a temporary directory for test configs
	tmpDir, err := os.MkdirTemp("", "config_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

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
	if err := os.WriteFile(configFilePath, []byte(configContent), 0600); err != nil {
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
	// Search only an empty directory: Load() also checks /app, which is the
	// repo when tests run in Docker and may contain a real configs/config.yaml.
	viper.Reset()

	_, err := loadFrom([]string{t.TempDir()})
	if err == nil {
		t.Error("Expected error when config file does not exist, got nil")
	}
}

func TestLoadInvalidYaml(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config_invalid_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	invalidContent := `
server:
  address: :8080
email:
  invalid_yaml: [unclosed bracket
`
	if err := os.WriteFile(filepath.Join(tmpDir, "config.yaml"), []byte(invalidContent), 0600); err != nil {
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

func TestValidateBotCommands(t *testing.T) {
	tests := []struct {
		name      string
		workflows []WorkflowCommandConfig
		restarts  []RestartCommandConfig
		wantErr   bool
	}{
		{name: "valid workflow", workflows: []WorkflowCommandConfig{{Command: "run_cleanup"}}},
		{name: "valid restart", restarts: []RestartCommandConfig{{Command: "restart_runner"}}},
		{name: "invalid characters", workflows: []WorkflowCommandConfig{{Command: "run-cleanup"}}, wantErr: true},
		{name: "duplicate within workflows", workflows: []WorkflowCommandConfig{{Command: "run_cleanup"}, {Command: "run_cleanup"}}, wantErr: true},
		{
			name:      "duplicate across workflows and restarts",
			workflows: []WorkflowCommandConfig{{Command: "run_cleanup"}},
			restarts:  []RestartCommandConfig{{Command: "run_cleanup"}},
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateBotCommands(tt.workflows, tt.restarts); (err != nil) != tt.wantErr {
				t.Errorf("validateBotCommands() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSchedules(t *testing.T) {
	tests := []struct {
		name     string
		workflow WorkflowCommandConfig
		wantErr  []string
	}{
		{name: "weekday list accepted", workflow: WorkflowCommandConfig{Command: "run_check_updates", Schedule: "0 9 * * 1,3,5"}},
		{name: "empty schedule accepted", workflow: WorkflowCommandConfig{Command: "run_cleanup"}},
		{name: "four fields rejected", workflow: WorkflowCommandConfig{Command: "run_cleanup", Schedule: "0 9 * *"}, wantErr: []string{"run_cleanup", "0 9 * *"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSchedules([]WorkflowCommandConfig{tt.workflow})
			if len(tt.wantErr) == 0 {
				if err != nil {
					t.Fatalf("validateSchedules() error = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatal("validateSchedules() error = nil, want error")
			}
			for _, want := range tt.wantErr {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error %q does not contain %q", err, want)
				}
			}
		})
	}
}

// TestLoadExampleConfig keeps configs/config.yaml.example loadable. Its
// {{PLACEHOLDER}} values are filled with a dummy value first, since they are
// not valid YAML until replaced.
func TestLoadExampleConfig(t *testing.T) {
	example, err := os.ReadFile(filepath.Join("..", "..", "configs", "config.yaml.example"))
	if err != nil {
		t.Fatalf("read example config: %v", err)
	}
	filled := regexp.MustCompile(`\{\{[A-Z_]+\}\}`).ReplaceAll(example, []byte("123"))

	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "config.yaml"), filled, 0600); err != nil { //nolint:gosec // G703 false positive: tmpDir comes from t.TempDir()
		t.Fatalf("write config: %v", err)
	}

	viper.Reset()
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(tmpDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(cfg.Workflows) != 2 {
		t.Errorf("workflows = %d, want 2", len(cfg.Workflows))
	}
}
