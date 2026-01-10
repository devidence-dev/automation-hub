package telegram

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type Client struct {
	bot    *tgbotapi.BotAPI
	logger *zap.Logger
}

func NewClient(token string, logger *zap.Logger) *Client {
	// Create a custom HTTP client with proper timeout settings
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			MaxIdleConns:          100,
		},
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		logger.Fatal("Failed to create Telegram bot", zap.Error(err))
	}

	// Set the custom HTTP client
	bot.Client = httpClient

	return &Client{
		bot:    bot,
		logger: logger,
	}
}

func (c *Client) SendMessage(chatID string, message string) error {
	chatIDInt, err := parseInt64(chatID)
	if err != nil {
		return fmt.Errorf("invalid chat ID: %w", err)
	}

	msg := tgbotapi.NewMessage(chatIDInt, message)
	msg.ParseMode = "Markdown"

	// Retry logic for transient network errors
	maxRetries := 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		_, err = c.bot.Send(msg)
		if err == nil {
			c.logger.Info("Telegram message sent successfully",
				zap.String("chatID", chatID),
				zap.Int("attempt", attempt))
			return nil
		}

		lastErr = err
		c.logger.Warn("Failed to send Telegram message",
			zap.String("chatID", chatID),
			zap.Error(err),
			zap.Int("attempt", attempt),
			zap.Int("maxRetries", maxRetries))

		// Don't retry on last attempt
		if attempt < maxRetries {
			// Exponential backoff: 1s, 2s, 4s
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			c.logger.Info("Retrying Telegram message send",
				zap.Duration("backoff", backoff))
			time.Sleep(backoff)
		}
	}

	c.logger.Error("Failed to send Telegram message after retries",
		zap.String("chatID", chatID),
		zap.Error(lastErr),
		zap.Int("attempts", maxRetries))
	return fmt.Errorf("failed to send message after %d attempts: %w", maxRetries, lastErr)
}

func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
