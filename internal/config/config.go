package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig     `mapstructure:"server"`
	Email    EmailConfig      `mapstructure:"email"`
	Telegram TelegramConfig   `mapstructure:"telegram"`
	Hook     []WebhookConfig  `mapstructure:"hook"`
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
	viper.AddConfigPath("/app/configs")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("/app")
	viper.AddConfigPath(".")

	// Environment variables override
	viper.AutomaticEnv()
	viper.SetEnvPrefix("AUTOMATION")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
