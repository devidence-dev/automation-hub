package config

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/viper"
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
	Hook        []WebhookConfig         `mapstructure:"hook"`
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
	CodePattern     string   `mapstructure:"code_pattern,omitempty"` // regex personalizado opcional
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

type WebhookConfig struct {
	Name   string                 `mapstructure:"name"`
	Path   string                 `mapstructure:"path"`
	Config WebhookProcessorConfig `mapstructure:"config"`
}

type WebhookProcessorConfig struct {
	TelegramChatID  string `mapstructure:"telegram_chat_id"`
	TelegramMessage string `mapstructure:"telegram_message"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/app")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("/app/configs")
	viper.AddConfigPath(".")

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

	if err := validateBotCommands(config.Workflows, config.Restarts); err != nil {
		return nil, err
	}

	return &config, nil
}

// validateBotCommands checks command names against Telegram's format rules
// and rejects duplicates across workflows AND restarts together, since both
// are registered on the same bot and share one command namespace.
func validateBotCommands(workflows []WorkflowCommandConfig, restarts []RestartCommandConfig) error {
	seen := make(map[string]struct{}, len(workflows)+len(restarts))
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
	return nil
}
