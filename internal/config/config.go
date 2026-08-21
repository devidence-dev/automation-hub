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

	if err := validateWorkflowCommands(config.Workflows); err != nil {
		return nil, err
	}

	return &config, nil
}

func validateWorkflowCommands(workflows []WorkflowCommandConfig) error {
	seen := make(map[string]struct{}, len(workflows))
	for _, workflow := range workflows {
		if !telegramCommandPattern.MatchString(workflow.Command) {
			return fmt.Errorf("invalid workflow command %q: must match %s", workflow.Command, telegramCommandPattern.String())
		}
		if _, exists := seen[workflow.Command]; exists {
			return fmt.Errorf("duplicate workflow command %q", workflow.Command)
		}
		seen[workflow.Command] = struct{}{}
	}
	return nil
}
