package telegram

import (
	"context"
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

var telegramAPIEndpoint = tgbotapi.APIEndpoint

// pollTimeout is how long Telegram holds a getUpdates request open while
// waiting for a new update before responding empty (Telegram long-polling).
const pollTimeout = 30 * time.Second

func NewClient(token string, logger *zap.Logger) *Client {
	return newClient(token, logger, newHTTPClient(), telegramAPIEndpoint)
}

func newClient(token string, logger *zap.Logger, httpClient tgbotapi.HTTPClient, apiEndpoint string) *Client {
	bot, err := tgbotapi.NewBotAPIWithClient(token, apiEndpoint, httpClient)
	if err != nil {
		logger.Fatal("Failed to create Telegram bot", zap.Error(err))
	}

	return &Client{bot: bot, logger: logger}
}

func newHTTPClient() *http.Client {
	// Create a custom HTTP client with proper timeout settings.
	// Long polling asks Telegram to hold getUpdates open for up to pollTimeout
	// seconds (see StartPolling), so every timeout here must exceed that window
	// or Telegram's normal "no updates yet" wait gets mistaken for a hang.
	return &http.Client{
		Timeout: pollTimeout + 10*time.Second,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: pollTimeout + 5*time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			MaxIdleConns:          100,
		},
	}
}

func (c *Client) SendMessage(chatID, message string) error {
	if c == nil || c.bot == nil {
		return nil
	}

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

// SetMyCommands synchronizes the command menu displayed by Telegram clients.
func (c *Client) SetMyCommands(commands []tgbotapi.BotCommand) error {
	if c == nil || c.bot == nil {
		return fmt.Errorf("telegram client is not initialized")
	}
	_, err := c.bot.Request(tgbotapi.SetMyCommandsConfig{Commands: commands})
	if err != nil {
		return fmt.Errorf("set Telegram commands: %w", err)
	}
	return nil
}

// StartPolling receives updates through Telegram long polling until ctx is cancelled.
func (c *Client) StartPolling(ctx context.Context, handler func(tgbotapi.Update)) {
	if c == nil || c.bot == nil {
		if c != nil && c.logger != nil {
			c.logger.Error("Cannot start Telegram polling: client is not initialized")
		}
		return
	}

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = int(pollTimeout.Seconds())
	updates := c.bot.GetUpdatesChan(updateConfig)

	for {
		select {
		case <-ctx.Done():
			c.bot.StopReceivingUpdates()
			return
		case update, ok := <-updates:
			if !ok {
				return
			}
			handler(update)
		}
	}
}

func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
