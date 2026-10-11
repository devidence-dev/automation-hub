package config

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/viper"

	"automation-hub/internal/services/scheduler"
)

var telegramCommandPattern = regexp.MustCompile(`^[a-z0-9_]{1,32}$`)

type Config struct {
	Server      ServerConfig            `mapstructure:"server"`
	Email       EmailConfig             `mapstructure:"email"`
	Telegram    TelegramConfig          `mapstructure:"telegram"`
	WorkflowBot WorkflowBotConfig       `mapstructure:"workflow_bot"`
	GitHub      GitHubConfig            `mapstructure:"github"`
	Workflows   []WorkflowCommandConfig `mapstructure:"workflows"`
	Restarts    []RestartCommandConfig  `mapstructure:"restarts"`
	Jobs        []JobCommandConfig      `mapstructure:"jobs"`
	// DependencyReport is nil when the section is absent, which disables its webhook.
	DependencyReport *DependencyReportConfig `mapstructure:"dependency_report"`
	Hook             []WebhookConfig         `mapstructure:"hook"`
}

type ServerConfig struct {
	Address string `mapstructure:"address"`
}

type EmailConfig struct {
	Host            string          `mapstructure:"host"`
	Port            int             `mapstructure:"port"`
	Username        string          `mapstructure:"username"`
	Password        string          `mapstructure:"password"`
	PollingInterval int             `mapstructure:"polling_interval"` // en segundos
	Services        []ServiceConfig `mapstructure:"services"`
}

type ServiceConfig struct {
	Name   string                 `mapstructure:"name"`
	Config ServiceProcessorConfig `mapstructure:"config"`
}

type ServiceProcessorConfig struct {
	EmailFrom       string   `mapstructure:"email_from"`
	EmailSubject    []string `mapstructure:"email_subject"`
	TelegramChatID  string   `mapstructure:"telegram_chat_id"`
	TelegramMessage string   `mapstructure:"telegram_message"`
	CodePattern     string   `mapstructure:"code_pattern,omitempty"` // optional custom regex
	MarkAsRead      *bool    `mapstructure:"mark_as_read,omitempty"` // optional; defaults to true
}

// ShouldMarkAsRead reports whether a processed email should be flagged as
// read. Unset (nil) defaults to true, since leaving it unread means the
// same code gets re-forwarded to Telegram on every poll cycle.
func (c ServiceProcessorConfig) ShouldMarkAsRead() bool {
	if c.MarkAsRead == nil {
		return true
	}
	return *c.MarkAsRead
}

type TelegramConfig struct {
	BotToken string            `mapstructure:"bot_token"`
	ChatIDs  map[string]string `mapstructure:"chat_ids"`
}

// WorkflowBotConfig controls the github-runner bot that receives workflow commands.
type WorkflowBotConfig struct {
	BotToken string `mapstructure:"bot_token"`
}

type GitHubConfig struct {
	Token string `mapstructure:"token"`
}

type WorkflowCommandConfig struct {
	Command        string   `mapstructure:"command"`
	Description    string   `mapstructure:"description"`
	Owner          string   `mapstructure:"owner"`
	Repo           string   `mapstructure:"repo"`
	WorkflowFile   string   `mapstructure:"workflow_file"`
	Ref            string   `mapstructure:"ref"`
	AllowedChatIDs []string `mapstructure:"allowed_chat_ids"`
	// Schedule is an optional 5-field cron expression evaluated in America/Guayaquil.
	Schedule string `mapstructure:"schedule"`
}

// RestartCommandConfig maps a Telegram command to a Kubernetes Deployment
// rollout restart (equivalent to `kubectl rollout restart deployment/...`).
type RestartCommandConfig struct {
	Command        string   `mapstructure:"command"`
	Description    string   `mapstructure:"description"`
	Namespace      string   `mapstructure:"namespace"`
	Deployment     string   `mapstructure:"deployment"`
	AllowedChatIDs []string `mapstructure:"allowed_chat_ids"`
}

// JobCommandConfig maps a Telegram command (and an optional schedule) to a
// Kubernetes Job created from a CronJob template, like
// `kubectl create job --from=cronjob/<cronjob>`.
type JobCommandConfig struct {
	Command        string   `mapstructure:"command"`
	Description    string   `mapstructure:"description"`
	Namespace      string   `mapstructure:"namespace"`
	CronJob        string   `mapstructure:"cronjob"`
	AllowedChatIDs []string `mapstructure:"allowed_chat_ids"`
	// Schedule is an optional 5-field cron expression evaluated in America/Guayaquil.
	Schedule string `mapstructure:"schedule"`
}

// DependencyReportConfig configures POST /webhook/dependency-report.
type DependencyReportConfig struct {
	WebhookToken   string `mapstructure:"webhook_token"`
	TelegramChatID string `mapstructure:"telegram_chat_id"`
	// Forks are repositories the report tags as informational-only forks.
	// In production the value is an unquoted Infisical reference holding a JSON
	// array, so YAML reads a list; a quoted reference becomes a string and is rejected.
	Forks []string `mapstructure:"forks"`
}

type WebhookConfig struct {
	Name   string                 `mapstructure:"name"`
	Path   string                 `mapstructure:"path"`
	Config WebhookProcessorConfig `mapstructure:"config"`
}

type WebhookProcessorConfig struct {
	TelegramChatID  string `mapstructure:"telegram_chat_id"`
	TelegramMessage string `mapstructure:"telegram_message"`
}

// defaultConfigPaths are searched in order: container paths first, then the working directory.
var defaultConfigPaths = []string{"/app", "./configs", "/app/configs", "."}

func Load() (*Config, error) {
	return loadFrom(defaultConfigPaths)
}

func loadFrom(configPaths []string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	for _, path := range configPaths {
		viper.AddConfigPath(path)
	}

	// Environment variables override
	viper.AutomaticEnv()
	viper.SetEnvPrefix("AUTOMATION")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	_ = viper.BindEnv("github.token")
	_ = viper.BindEnv("telegram.bot_token")
	_ = viper.BindEnv("workflow_bot.bot_token")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	if err := validateBotCommands(config.Workflows, config.Restarts, config.Jobs); err != nil {
		return nil, err
	}
	if err := validateSchedules(config.Workflows, config.Jobs); err != nil {
		return nil, err
	}
	if err := validateJobs(config.Jobs); err != nil {
		return nil, err
	}
	if err := validateDependencyReport(config.DependencyReport); err != nil {
		return nil, err
	}

	return &config, nil
}

// validateBotCommands checks command names against Telegram's format rules
// and rejects duplicates across workflows, restarts AND jobs together, since all
// are registered on the same bot and share one command namespace.
func validateBotCommands(workflows []WorkflowCommandConfig, restarts []RestartCommandConfig, jobs []JobCommandConfig) error {
	seen := make(map[string]struct{}, len(workflows)+len(restarts)+len(jobs))
	check := func(command string) error {
		if !telegramCommandPattern.MatchString(command) {
			return fmt.Errorf("invalid bot command %q: must match %s", command, telegramCommandPattern.String())
		}
		if _, exists := seen[command]; exists {
			return fmt.Errorf("duplicate bot command %q", command)
		}
		seen[command] = struct{}{}
		return nil
	}
	for _, workflow := range workflows {
		if err := check(workflow.Command); err != nil {
			return err
		}
	}
	for _, restart := range restarts {
		if err := check(restart.Command); err != nil {
			return err
		}
	}
	for _, job := range jobs {
		if err := check(job.Command); err != nil {
			return err
		}
	}
	return nil
}

// validateSchedules parses every non-empty schedule with the scheduler's parser,
// so an invalid expression stops startup before any goroutine starts.
func validateSchedules(workflows []WorkflowCommandConfig, jobs []JobCommandConfig) error {
	check := func(command, schedule string) error {
		if schedule == "" {
			return nil
		}
		if _, err := scheduler.ParseSpec(schedule); err != nil {
			return fmt.Errorf("invalid schedule %q for command %q: %w", schedule, command, err)
		}
		return nil
	}
	for _, workflow := range workflows {
		if err := check(workflow.Command, workflow.Schedule); err != nil {
			return err
		}
	}
	for _, job := range jobs {
		if err := check(job.Command, job.Schedule); err != nil {
			return err
		}
	}
	return nil
}

func validateJobs(jobs []JobCommandConfig) error {
	for _, job := range jobs {
		if job.Namespace == "" || job.CronJob == "" {
			return fmt.Errorf("job command %q requires namespace and cronjob", job.Command)
		}
	}
	return nil
}

// validateDependencyReport rejects an empty token (it would match a missing
// header) and a non-list forks value. viper's weak typing would otherwise
// split a quoted string on commas instead of failing.
func validateDependencyReport(report *DependencyReportConfig) error {
	if report == nil {
		return nil
	}
	if report.WebhookToken == "" || report.TelegramChatID == "" {
		return fmt.Errorf("dependency_report requires webhook_token and telegram_chat_id")
	}
	if _, isString := viper.Get("dependency_report.forks").(string); isString {
		return fmt.Errorf("dependency_report.forks must be a YAML list: do not quote the reference")
	}
	return nil
}
