package config

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	_ "time/tzdata" // zone database: the image may have no /usr/share/zoneinfo

	"github.com/spf13/viper"

	"automation-hub/internal/messages"
	"automation-hub/internal/services/scheduler"
)

var telegramCommandPattern = regexp.MustCompile(`^[a-z0-9_]{1,32}$`)

// Defaults for the values the config may leave out. They keep today's behavior.
const (
	defaultTimezone     = "America/Guayaquil"
	defaultReportPath   = "/webhook/dependency-report"
	defaultTokenHeader  = "X-Report-Token"
	defaultMaxBodyBytes = 10 << 20 // 10 MiB
	defaultDocumentName = "dependency-report-{date}.html"
	defaultDateLayout   = "2006-01-02"
)

type Config struct {
	// Timezone is the zone of every schedule and of the dependency report's dates.
	Timezone    string                  `mapstructure:"timezone"`
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
	// Messages overrides single keys of the message catalog, with the nested keys of the defaults.
	Messages map[string]any `mapstructure:"messages"`
	// Location and Catalog are set by Load from Timezone and Messages.
	Location *time.Location    `mapstructure:"-"`
	Catalog  *messages.Catalog `mapstructure:"-"`
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
	// Schedule is an optional 5-field cron expression evaluated in the top-level timezone.
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
	// Schedule is an optional 5-field cron expression evaluated in the top-level timezone.
	Schedule string `mapstructure:"schedule"`
}

// DependencyReportConfig configures the dependency report webhook. The defaults of the
// settings below are filled by applyDefaults only when the section exists.
type DependencyReportConfig struct {
	WebhookToken   string `mapstructure:"webhook_token"`
	TelegramChatID string `mapstructure:"telegram_chat_id"`
	// Forks are repositories the report tags as informational-only forks.
	// In production the value is an unquoted Infisical reference holding a JSON
	// array, so YAML reads a list; a quoted reference becomes a string and is rejected.
	Forks []string `mapstructure:"forks"`
	// Path is the route the Job posts to, TokenHeader the header that carries WebhookToken.
	Path        string `mapstructure:"path"`
	TokenHeader string `mapstructure:"token_header"`
	// MaxBodyBytes caps the multipart body; a larger one gets 413.
	MaxBodyBytes int64 `mapstructure:"max_body_bytes"`
	// DocumentName is the file name pattern, with {date} replaced by the day in the timezone.
	DocumentName string `mapstructure:"document_name"`
	// DateLayout is the Go time layout of that date.
	DateLayout string `mapstructure:"date_layout"`
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

	if err := resolve(&config); err != nil {
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

// resolve fills the defaults, then the zone and the message catalog the config names.
func resolve(cfg *Config) error {
	applyDefaults(cfg)
	loc, err := loadLocation(cfg.Timezone)
	if err != nil {
		return err
	}
	catalog, err := messages.Load(cfg.Messages)
	if err != nil {
		return err
	}
	cfg.Location = loc
	cfg.Catalog = catalog
	return nil
}

// applyDefaults fills the values the config leaves out. The dependency report defaults apply
// only when its section exists, so an absent section keeps its webhook off.
func applyDefaults(cfg *Config) {
	if cfg.Timezone == "" {
		cfg.Timezone = defaultTimezone
	}
	if cfg.DependencyReport != nil {
		applyReportDefaults(cfg.DependencyReport)
	}
}

// applyReportDefaults checks viper.IsSet, not the zero value, so an explicit empty or zero
// setting reaches validation instead of silently taking the default.
func applyReportDefaults(r *DependencyReportConfig) {
	if !viper.IsSet("dependency_report.path") {
		r.Path = defaultReportPath
	}
	if !viper.IsSet("dependency_report.token_header") {
		r.TokenHeader = defaultTokenHeader
	}
	if !viper.IsSet("dependency_report.max_body_bytes") {
		r.MaxBodyBytes = defaultMaxBodyBytes
	}
	if !viper.IsSet("dependency_report.document_name") {
		r.DocumentName = defaultDocumentName
	}
	if !viper.IsSet("dependency_report.date_layout") {
		r.DateLayout = defaultDateLayout
	}
}

// loadLocation loads the configured zone from the embedded database.
func loadLocation(name string) (*time.Location, error) {
	loc, err := time.LoadLocation(name)
	if err != nil {
		return nil, fmt.Errorf("timezone %q: %w", name, err)
	}
	return loc, nil
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
	return validateReportSettings(report)
}

// validateReportSettings checks the webhook settings. Each error names the setting.
func validateReportSettings(r *DependencyReportConfig) error {
	switch {
	case !strings.HasPrefix(r.Path, "/"):
		return fmt.Errorf("dependency_report.path must start with /: %q", r.Path)
	case r.TokenHeader == "" || strings.ContainsAny(r.TokenHeader, " \t"):
		return fmt.Errorf("dependency_report.token_header must be a header name without spaces")
	case r.MaxBodyBytes <= 0:
		return fmt.Errorf("dependency_report.max_body_bytes must be a positive integer, got %d", r.MaxBodyBytes)
	case !strings.Contains(r.DocumentName, "{date}") || strings.ContainsAny(r.DocumentName, `/\`):
		return fmt.Errorf("dependency_report.document_name must contain {date} and no / or \\: %q", r.DocumentName)
	case r.DateLayout == "":
		return fmt.Errorf("dependency_report.date_layout must not be empty")
	}
	return nil
}
