package config

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
	_ "time/tzdata" // zone database: the image may have no /usr/share/zoneinfo

	"github.com/spf13/viper"

	"automation-hub/internal/httpx"
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

// Defaults of the bot settings. The handlers fall back to them too, so a zero value never reaches a command.
const (
	DefaultGitHubWebURL        = "https://github.com"
	DefaultRestartTimeout      = 3 * time.Minute
	DefaultRestartPollInterval = 5 * time.Second
)

// Defaults of the HTTP settings. Each equals the value the code used before the settings were configurable.
const (
	defaultServerReadTimeout     = 15 * time.Second
	defaultServerWriteTimeout    = 15 * time.Second
	defaultServerShutdownTimeout = 30 * time.Second
	defaultGitHubAPIURL          = "https://api.github.com"
	defaultGitHubTimeout         = 30 * time.Second
	defaultGitHubDialTimeout     = 30 * time.Second
	defaultKubernetesTimeout     = 15 * time.Second
	defaultDialTimeout           = 10 * time.Second // Telegram and Kubernetes
	defaultTLSHandshakeTimeout   = 10 * time.Second
	defaultKeepAlive             = 30 * time.Second
	defaultMaxIdleConns          = 100
	defaultIdleConnTimeout       = 90 * time.Second
	defaultTelegramPollTimeout   = 30 * time.Second
	// The Telegram request and header timeouts are poll_timeout plus these margins, when their keys are absent.
	defaultTelegramTimeoutMargin = 10 * time.Second
	defaultTelegramHeaderMargin  = 5 * time.Second
	defaultExpectContinueTimeout = time.Second
	defaultMaxAttempts           = 3
	defaultRetryBackoff          = time.Second
	defaultMaxRetryAfter         = 30 * time.Second
)

type Config struct {
	// Timezone is the zone of every schedule and of the dependency report's dates.
	Timezone    string                  `mapstructure:"timezone"`
	Server      ServerConfig            `mapstructure:"server"`
	Email       EmailConfig             `mapstructure:"email"`
	Telegram    TelegramConfig          `mapstructure:"telegram"`
	WorkflowBot WorkflowBotConfig       `mapstructure:"workflow_bot"`
	GitHub      GitHubConfig            `mapstructure:"github"`
	Kubernetes  KubernetesConfig        `mapstructure:"kubernetes"`
	Workflows   []WorkflowCommandConfig `mapstructure:"workflows"`
	Restarts    []RestartCommandConfig  `mapstructure:"restarts"`
	Jobs        []JobCommandConfig      `mapstructure:"jobs"`
	// RestartReady bounds the wait for a restarted Deployment to report ready, for every restart command.
	RestartReady RestartReadyConfig `mapstructure:"restart_ready"`
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
	// ReadTimeout and WriteTimeout bound each webhook request; ShutdownTimeout bounds the graceful stop.
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
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
	// HTTP is shared by the notifications bot and the workflow bot.
	HTTP TelegramHTTPConfig `mapstructure:"http"`
}

// WorkflowBotConfig controls the github-runner bot that receives workflow commands.
type WorkflowBotConfig struct {
	BotToken string `mapstructure:"bot_token"`
}

type GitHubConfig struct {
	Token string `mapstructure:"token"`
	// WebURL is the GitHub web address the run links fall back to. Trailing slashes are trimmed.
	WebURL string `mapstructure:"web_url"`
	// APIURL is the base of the REST API requests, GitHub Enterprise Server included. Trailing slashes are trimmed.
	APIURL string           `mapstructure:"api_url"`
	HTTP   GitHubHTTPConfig `mapstructure:"http"`
}

// KubernetesConfig holds the settings of the in-cluster Kubernetes client.
type KubernetesConfig struct {
	HTTP KubernetesHTTPConfig `mapstructure:"http"`
}

// HTTPConfig holds the transport values of one client section. Each section embeds it, so the keys
// are the same under telegram.http, github.http and kubernetes.http.
type HTTPConfig struct {
	Timeout             time.Duration `mapstructure:"timeout"`
	DialTimeout         time.Duration `mapstructure:"dial_timeout"`
	TLSHandshakeTimeout time.Duration `mapstructure:"tls_handshake_timeout"`
	KeepAlive           time.Duration `mapstructure:"keep_alive"`
	MaxIdleConns        int           `mapstructure:"max_idle_conns"`
	IdleConnTimeout     time.Duration `mapstructure:"idle_conn_timeout"`
}

// Settings returns the transport values for httpx.NewClient.
func (h HTTPConfig) Settings() httpx.Settings {
	return httpx.Settings{
		Timeout:             h.Timeout,
		DialTimeout:         h.DialTimeout,
		TLSHandshakeTimeout: h.TLSHandshakeTimeout,
		KeepAlive:           h.KeepAlive,
		MaxIdleConns:        h.MaxIdleConns,
		IdleConnTimeout:     h.IdleConnTimeout,
	}
}

// GitHubHTTPConfig is the github.http section. MaxResponseBytes caps every response body the client reads.
type GitHubHTTPConfig struct {
	HTTPConfig       `mapstructure:",squash"`
	MaxResponseBytes int64 `mapstructure:"max_response_bytes"`
}

// Settings returns the transport values for httpx.NewClient. GitHub keeps the environment proxy and
// HTTP/2 of the default transport it used before this section existed, so both are on here only.
func (h GitHubHTTPConfig) Settings() httpx.Settings {
	s := h.HTTPConfig.Settings()
	s.EnvProxy = true
	s.ForceHTTP2 = true
	return s
}

// KubernetesHTTPConfig is the kubernetes.http section. MaxResponseBytes caps every response body the client reads.
type KubernetesHTTPConfig struct {
	HTTPConfig       `mapstructure:",squash"`
	MaxResponseBytes int64 `mapstructure:"max_response_bytes"`
}

// TelegramHTTPConfig is the telegram.http section: the transport, the long-poll window and the send retry policy.
type TelegramHTTPConfig struct {
	HTTPConfig            `mapstructure:",squash"`
	PollTimeout           time.Duration `mapstructure:"poll_timeout"`
	ResponseHeaderTimeout time.Duration `mapstructure:"response_header_timeout"`
	ExpectContinueTimeout time.Duration `mapstructure:"expect_continue_timeout"`
	MaxAttempts           int           `mapstructure:"max_attempts"`
	RetryBackoff          time.Duration `mapstructure:"retry_backoff"`
	MaxRetryAfter         time.Duration `mapstructure:"max_retry_after"`
}

// Settings returns the transport values for httpx.NewClient, including the Telegram header and expect-continue limits.
func (h TelegramHTTPConfig) Settings() httpx.Settings {
	s := h.HTTPConfig.Settings()
	s.ResponseHeaderTimeout = h.ResponseHeaderTimeout
	s.ExpectContinueTimeout = h.ExpectContinueTimeout
	return s
}

// RestartReadyConfig sets how long a restart waits for the rollout, and how often it checks.
type RestartReadyConfig struct {
	Timeout      time.Duration `mapstructure:"timeout"`
	PollInterval time.Duration `mapstructure:"poll_interval"`
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
	setBotDefaults()
	setHTTPDefaults()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	if err := validateDurationUnits(); err != nil {
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
	if err := validateGitHubWebURL(config.GitHub.WebURL); err != nil {
		return nil, err
	}
	if err := validateRestartReady(config.RestartReady); err != nil {
		return nil, err
	}
	if err := validateHTTPSettings(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// setHTTPDefaults registers the defaults of the HTTP keys with viper, so an absent key decodes to them.
// The Telegram request and header timeouts are not registered: they derive from poll_timeout, and
// applyTelegramHTTPDefaults fills them.
func setHTTPDefaults() {
	viper.SetDefault("server.read_timeout", defaultServerReadTimeout)
	viper.SetDefault("server.write_timeout", defaultServerWriteTimeout)
	viper.SetDefault("server.shutdown_timeout", defaultServerShutdownTimeout)
	viper.SetDefault("github.api_url", defaultGitHubAPIURL)
	viper.SetDefault("github.http.timeout", defaultGitHubTimeout)
	viper.SetDefault("github.http.dial_timeout", defaultGitHubDialTimeout)
	viper.SetDefault("github.http.max_response_bytes", httpx.DefaultMaxResponseBytes)
	viper.SetDefault("kubernetes.http.timeout", defaultKubernetesTimeout)
	viper.SetDefault("kubernetes.http.dial_timeout", defaultDialTimeout)
	viper.SetDefault("kubernetes.http.max_response_bytes", httpx.DefaultMaxResponseBytes)
	viper.SetDefault("telegram.http.dial_timeout", defaultDialTimeout)
	viper.SetDefault("telegram.http.poll_timeout", defaultTelegramPollTimeout)
	viper.SetDefault("telegram.http.expect_continue_timeout", defaultExpectContinueTimeout)
	viper.SetDefault("telegram.http.max_attempts", defaultMaxAttempts)
	viper.SetDefault("telegram.http.retry_backoff", defaultRetryBackoff)
	viper.SetDefault("telegram.http.max_retry_after", defaultMaxRetryAfter)
	for _, section := range []string{"telegram.http", "github.http", "kubernetes.http"} {
		viper.SetDefault(section+".tls_handshake_timeout", defaultTLSHandshakeTimeout)
		viper.SetDefault(section+".keep_alive", defaultKeepAlive)
		viper.SetDefault(section+".max_idle_conns", defaultMaxIdleConns)
		viper.SetDefault(section+".idle_conn_timeout", defaultIdleConnTimeout)
	}
}

// durationKeys are the keys of every time.Duration value. A bare YAML integer decodes to nanoseconds
// without an error, so each of these must be a quoted string with a unit.
var durationKeys = []string{
	"server.read_timeout", "server.write_timeout", "server.shutdown_timeout",
	"telegram.http.timeout", "telegram.http.dial_timeout", "telegram.http.tls_handshake_timeout",
	"telegram.http.keep_alive", "telegram.http.idle_conn_timeout", "telegram.http.poll_timeout",
	"telegram.http.response_header_timeout", "telegram.http.expect_continue_timeout",
	"telegram.http.retry_backoff", "telegram.http.max_retry_after",
	"github.http.timeout", "github.http.dial_timeout", "github.http.tls_handshake_timeout",
	"github.http.keep_alive", "github.http.idle_conn_timeout",
	"kubernetes.http.timeout", "kubernetes.http.dial_timeout", "kubernetes.http.tls_handshake_timeout",
	"kubernetes.http.keep_alive", "kubernetes.http.idle_conn_timeout",
	"restart_ready.timeout", "restart_ready.poll_interval",
}

// validateDurationUnits rejects a duration key that the config file sets to a non-string value.
// It runs before unmarshal, while the raw value is still visible. Defaults and environment
// variables are not checked: defaults are durations already, and environment values are strings.
func validateDurationUnits() error {
	for _, key := range durationKeys {
		if !viper.InConfig(key) {
			continue
		}
		value := viper.Get(key)
		if _, isString := value.(string); !isString {
			return fmt.Errorf("%s: a unit is required, for example \"30s\", got %v", key, value)
		}
	}
	return nil
}

// namedDuration pairs a config key with its value, so a validation error can name the key.
type namedDuration struct {
	key   string
	value time.Duration
}

// requirePositive returns an error naming the first duration that is not positive.
func requirePositive(durations ...namedDuration) error {
	for _, d := range durations {
		if d.value <= 0 {
			return fmt.Errorf("%s must be positive, got %s", d.key, d.value)
		}
	}
	return nil
}

// validateHTTPSettings checks the HTTP keys of the server, the three clients and the GitHub API base URL.
func validateHTTPSettings(cfg *Config) error {
	if err := validateServerTimeouts(cfg.Server); err != nil {
		return err
	}
	if err := validateTelegramHTTP(cfg.Telegram.HTTP); err != nil {
		return err
	}
	if err := validateGitHubAPIURL(cfg.GitHub.APIURL); err != nil {
		return err
	}
	if err := validateClientHTTP("github.http", cfg.GitHub.HTTP.HTTPConfig, cfg.GitHub.HTTP.MaxResponseBytes); err != nil {
		return err
	}
	return validateClientHTTP("kubernetes.http", cfg.Kubernetes.HTTP.HTTPConfig, cfg.Kubernetes.HTTP.MaxResponseBytes)
}

func validateServerTimeouts(s ServerConfig) error {
	return requirePositive(
		namedDuration{"server.read_timeout", s.ReadTimeout},
		namedDuration{"server.write_timeout", s.WriteTimeout},
		namedDuration{"server.shutdown_timeout", s.ShutdownTimeout},
	)
}

// validateGitHubAPIURL requires an absolute https URL, the base of the GitHub API requests.
func validateGitHubAPIURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return fmt.Errorf("github.api_url must be an absolute https URL: %q", raw)
	}
	return nil
}

// validateTransport requires positive durations and a positive idle connection count for one section.
func validateTransport(section string, h HTTPConfig) error {
	if err := requirePositive(
		namedDuration{section + ".timeout", h.Timeout},
		namedDuration{section + ".dial_timeout", h.DialTimeout},
		namedDuration{section + ".tls_handshake_timeout", h.TLSHandshakeTimeout},
		namedDuration{section + ".keep_alive", h.KeepAlive},
		namedDuration{section + ".idle_conn_timeout", h.IdleConnTimeout},
	); err != nil {
		return err
	}
	if h.MaxIdleConns <= 0 {
		return fmt.Errorf("%s.max_idle_conns must be positive, got %d", section, h.MaxIdleConns)
	}
	return nil
}

// validateClientHTTP checks the transport of a client section and its response body cap.
func validateClientHTTP(section string, h HTTPConfig, maxResponseBytes int64) error {
	if err := validateTransport(section, h); err != nil {
		return err
	}
	if maxResponseBytes <= 0 {
		return fmt.Errorf("%s.max_response_bytes must be positive, got %d", section, maxResponseBytes)
	}
	return nil
}

// validateTelegramHTTP checks the Telegram transport, the long-poll window and the retry policy. The
// request timeout and the response header timeout must each be greater than poll_timeout.
func validateTelegramHTTP(h TelegramHTTPConfig) error {
	if err := validateTransport("telegram.http", h.HTTPConfig); err != nil {
		return err
	}
	if err := requirePositive(
		namedDuration{"telegram.http.poll_timeout", h.PollTimeout},
		namedDuration{"telegram.http.response_header_timeout", h.ResponseHeaderTimeout},
		namedDuration{"telegram.http.expect_continue_timeout", h.ExpectContinueTimeout},
		namedDuration{"telegram.http.retry_backoff", h.RetryBackoff},
		namedDuration{"telegram.http.max_retry_after", h.MaxRetryAfter},
	); err != nil {
		return err
	}
	if h.Timeout <= h.PollTimeout {
		return fmt.Errorf("telegram.http.timeout %s must be greater than poll_timeout %s", h.Timeout, h.PollTimeout)
	}
	if h.ResponseHeaderTimeout <= h.PollTimeout {
		return fmt.Errorf("telegram.http.response_header_timeout %s must be greater than poll_timeout %s", h.ResponseHeaderTimeout, h.PollTimeout)
	}
	if h.MaxAttempts < 1 {
		return fmt.Errorf("telegram.http.max_attempts must be at least 1, got %d", h.MaxAttempts)
	}
	return nil
}

// setBotDefaults registers the defaults of the bot keys with viper, so an absent key decodes to them.
func setBotDefaults() {
	viper.SetDefault("github.web_url", DefaultGitHubWebURL)
	viper.SetDefault("restart_ready.timeout", DefaultRestartTimeout)
	viper.SetDefault("restart_ready.poll_interval", DefaultRestartPollInterval)
}

// validateGitHubWebURL requires an absolute http or https URL, the base of the run links.
func validateGitHubWebURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return fmt.Errorf("github.web_url must be an absolute http or https URL: %q", raw)
	}
	return nil
}

// validateRestartReady requires positive durations, and a poll interval that fits in the timeout.
func validateRestartReady(r RestartReadyConfig) error {
	switch {
	case r.Timeout <= 0:
		return fmt.Errorf("restart_ready.timeout must be positive, got %s", r.Timeout)
	case r.PollInterval <= 0:
		return fmt.Errorf("restart_ready.poll_interval must be positive, got %s", r.PollInterval)
	case r.PollInterval > r.Timeout:
		return fmt.Errorf("restart_ready.poll_interval %s must not exceed restart_ready.timeout %s", r.PollInterval, r.Timeout)
	}
	return nil
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
	cfg.GitHub.WebURL = strings.TrimRight(cfg.GitHub.WebURL, "/")
	cfg.GitHub.APIURL = strings.TrimRight(cfg.GitHub.APIURL, "/")
	applyTelegramHTTPDefaults(&cfg.Telegram.HTTP)
	if cfg.DependencyReport != nil {
		applyReportDefaults(cfg.DependencyReport)
	}
}

// applyTelegramHTTPDefaults derives the Telegram request and header timeouts from poll_timeout when
// their keys are absent. An explicit value, even an invalid one, reaches validation.
func applyTelegramHTTPDefaults(h *TelegramHTTPConfig) {
	if !viper.IsSet("telegram.http.timeout") {
		h.Timeout = h.PollTimeout + defaultTelegramTimeoutMargin
	}
	if !viper.IsSet("telegram.http.response_header_timeout") {
		h.ResponseHeaderTimeout = h.PollTimeout + defaultTelegramHeaderMargin
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
