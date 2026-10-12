package config

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
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
		jobs      []JobCommandConfig
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
		{name: "valid job", jobs: []JobCommandConfig{{Command: "run_dependency_report"}}},
		{
			name:      "duplicate across workflows and jobs",
			workflows: []WorkflowCommandConfig{{Command: "run_check_updates"}},
			jobs:      []JobCommandConfig{{Command: "run_check_updates"}},
			wantErr:   true,
		},
		{
			name:     "duplicate across restarts and jobs",
			restarts: []RestartCommandConfig{{Command: "restart_runner"}},
			jobs:     []JobCommandConfig{{Command: "restart_runner"}},
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateBotCommands(tt.workflows, tt.restarts, tt.jobs); (err != nil) != tt.wantErr {
				t.Errorf("validateBotCommands() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSchedules(t *testing.T) {
	tests := []struct {
		name     string
		workflow WorkflowCommandConfig
		job      JobCommandConfig
		wantErr  []string
	}{
		{name: "job schedule accepted", job: JobCommandConfig{Command: "run_dependency_report", Schedule: "0 9 * * 1"}},
		{name: "job four fields rejected", job: JobCommandConfig{Command: "run_dependency_report", Schedule: "0 9 * *"}, wantErr: []string{"run_dependency_report", "0 9 * *"}},
		{name: "weekday list accepted", workflow: WorkflowCommandConfig{Command: "run_check_updates", Schedule: "0 9 * * 1,3,5"}},
		{name: "empty schedule accepted", workflow: WorkflowCommandConfig{Command: "run_cleanup"}},
		{name: "four fields rejected", workflow: WorkflowCommandConfig{Command: "run_cleanup", Schedule: "0 9 * *"}, wantErr: []string{"run_cleanup", "0 9 * *"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSchedules([]WorkflowCommandConfig{tt.workflow}, []JobCommandConfig{tt.job})
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

// loadYAML writes content as config.yaml in a temp dir and loads it.
func loadYAML(t *testing.T, content string) (*Config, error) {
	t.Helper()
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "config.yaml"), []byte(content), 0600); err != nil { //nolint:gosec // G703 false positive: tmpDir comes from t.TempDir()
		t.Fatalf("write config: %v", err)
	}
	viper.Reset()
	return loadFrom([]string{tmpDir})
}

func TestLoadJobsAndDependencyReport(t *testing.T) {
	const jobsOnly = `
jobs:
  - command: "run_dependency_report"
    namespace: "internal-apps"
    cronjob: "dependency-report"
    schedule: "0 9 * * 1"
    allowed_chat_ids: ["123"]
`
	const report = `
dependency_report:
  webhook_token: "tok"
  telegram_chat_id: "456"
  forks: %s
`
	tests := []struct {
		name      string
		content   string
		wantErr   string
		wantJobs  int
		wantForks []string
	}{
		{name: "jobs-only config accepted", content: jobsOnly, wantJobs: 1},
		{name: "duplicate across types rejected", wantErr: "run_dependency_report", content: jobsOnly + `
workflows:
  - command: "run_dependency_report"
`},
		{name: "invalid job schedule rejected", wantErr: "0 9 * *", content: `
jobs:
  - {command: "run_dependency_report", namespace: "ns", cronjob: "cj", schedule: "0 9 * *"}
`},
		{name: "job without cronjob rejected", wantErr: "namespace and cronjob", content: `
jobs:
  - {command: "run_dependency_report", namespace: "ns"}
`},
		{name: "unquoted JSON array is a list", content: strings.Replace(report, "%s", `["devidence-dev/a", "devidence-dev/b"]`, 1), wantForks: []string{"devidence-dev/a", "devidence-dev/b"}},
		{name: "quoted array is rejected", wantErr: "forks", content: strings.Replace(report, "%s", `'["devidence-dev/a", "devidence-dev/b"]'`, 1)},
		{name: "quoted plain string is rejected", wantErr: "forks", content: strings.Replace(report, "%s", `"devidence-dev/a"`, 1)},
		{name: "missing token rejected", wantErr: "webhook_token", content: "dependency_report:\n  telegram_chat_id: \"456\"\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertJobsAndReportLoad(t, tt.content, tt.wantErr, tt.wantJobs, tt.wantForks)
		})
	}
}

// assertJobsAndReportLoad loads content and checks either the error it must
// contain (wantErr) or the number of jobs and the forks of the report section.
func assertJobsAndReportLoad(t *testing.T, content, wantErr string, wantJobs int, wantForks []string) {
	t.Helper()
	cfg, err := loadYAML(t, content)
	if wantErr != "" {
		if err == nil || !strings.Contains(err.Error(), wantErr) {
			t.Fatalf("Load() error = %v, want it to contain %q", err, wantErr)
		}
		return
	}
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(cfg.Jobs) != wantJobs {
		t.Errorf("jobs = %d, want %d", len(cfg.Jobs), wantJobs)
	}
	if wantJobs > 0 && cfg.DependencyReport != nil {
		t.Errorf("DependencyReport = %+v, want nil when the section is absent", cfg.DependencyReport)
	}
	if wantForks != nil && !reflect.DeepEqual(cfg.DependencyReport.Forks, wantForks) {
		t.Errorf("forks = %v, want %v", cfg.DependencyReport.Forks, wantForks)
	}
}

// TestLoadExampleConfig keeps configs/config.yaml.example loadable. Its
// {{PLACEHOLDER}} values are filled with a dummy value first, since they are
// not valid YAML until replaced. The ${INFISICAL_REFERENCE} values are
// substituted by Infisical in production: the forks one becomes a JSON array.
func TestLoadExampleConfig(t *testing.T) {
	example, err := os.ReadFile(filepath.Join("..", "..", "configs", "config.yaml.example"))
	if err != nil {
		t.Fatalf("read example config: %v", err)
	}
	filled := regexp.MustCompile(`\{\{[A-Z_]+\}\}`).ReplaceAll(example, []byte("123"))
	filled = bytes.ReplaceAll(filled, []byte("${DEPENDENCY_REPORT_FORKS}"), []byte(`["devidence-dev/discord-tts-bot"]`))
	filled = regexp.MustCompile(`\$\{[A-Z_]+\}`).ReplaceAll(filled, []byte("123"))

	cfg, err := loadYAML(t, string(filled))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	// run_check_updates was retired with icr (the example lists run_cleanup only).
	if len(cfg.Workflows) != 1 || cfg.Workflows[0].Command != "run_cleanup" {
		t.Errorf("workflows = %+v, want only run_cleanup", cfg.Workflows)
	}
	if len(cfg.Jobs) != 1 || cfg.Jobs[0].Schedule != "0 9 * * 1" {
		t.Errorf("jobs = %+v, want run_dependency_report with schedule 0 9 * * 1", cfg.Jobs)
	}
	if cfg.DependencyReport == nil || len(cfg.DependencyReport.Forks) != 1 {
		t.Errorf("dependency_report = %+v, want one fork", cfg.DependencyReport)
	}
}
