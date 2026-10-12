package telegram

import (
	"context"
	"fmt"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"

	"automation-hub/internal/httpx"
)

// Config holds what the client needs. The caller maps its own configuration onto it.
type Config struct {
	Token  string
	Logger *zap.Logger
	// HTTP are the transport settings of the client's HTTP client.
	HTTP httpx.Settings
	// PollTimeout is the long-polling window sent to Telegram in getUpdates.
	PollTimeout time.Duration
	// MaxAttempts is the number of attempts per send, including the first.
	MaxAttempts int
	// RetryBackoff is the wait after the first failed attempt; it doubles after each further one.
	RetryBackoff time.Duration
	// MaxRetryAfter is the longest wait the client honours; a retry_after above it fails at once.
	MaxRetryAfter time.Duration
}

type Client struct {
	bot         *tgbotapi.BotAPI
	retrier     *RetryingSender // nil when the client is uninitialized
	logger      *zap.Logger
	pollTimeout time.Duration
}

var telegramAPIEndpoint = tgbotapi.APIEndpoint

// NewClient builds a client whose HTTP traffic uses an httpx client built from cfg.HTTP.
func NewClient(cfg Config) *Client {
	return newClient(cfg, httpx.NewClient(cfg.HTTP), telegramAPIEndpoint, time.Sleep)
}

// newClient builds a client that talks to apiEndpoint through httpClient and waits between
// retries with sleep.
func newClient(cfg Config, httpClient tgbotapi.HTTPClient, apiEndpoint string, sleep func(time.Duration)) *Client {
	bot, err := tgbotapi.NewBotAPIWithClient(cfg.Token, apiEndpoint, httpClient)
	if err != nil {
		cfg.Logger.Fatal("Failed to create Telegram bot", zap.Error(err))
	}

	return newClientWithBot(cfg, bot, sleep)
}

// newClientWithBot wraps an existing bot with the retry policy of cfg.
func newClientWithBot(cfg Config, bot *tgbotapi.BotAPI, sleep func(time.Duration)) *Client {
	backoff := ExponentialBackoff{Base: cfg.RetryBackoff, Ceiling: cfg.MaxRetryAfter}
	retrier := NewRetryingSender(bot, backoff, cfg.MaxAttempts, sleep, cfg.Logger)
	return &Client{
		bot:         bot,
		retrier:     &retrier,
		logger:      cfg.Logger,
		pollTimeout: cfg.PollTimeout,
	}
}

// SendMessage sends a Markdown message.
func (c *Client) SendMessage(chatID, message string) error {
	return c.sendText(chatID, message, tgbotapi.ModeMarkdown)
}

// SendHTML sends a message with HTML parse mode. The caller must escape variable text.
func (c *Client) SendHTML(chatID, message string) error {
	return c.sendText(chatID, message, tgbotapi.ModeHTML)
}

// SendDocument sends data as a file named filename, with caption as its HTML caption.
// The caller must escape variable text in the caption.
func (c *Client) SendDocument(chatID, filename, caption string, data []byte) error {
	if !c.ready() {
		return nil
	}
	id, err := resolveChat(chatID)
	if err != nil {
		return err
	}

	doc := tgbotapi.NewDocument(id, tgbotapi.FileBytes{Name: filename, Bytes: data})
	doc.Caption = caption
	doc.ParseMode = tgbotapi.ModeHTML
	return c.send(chatID, doc)
}

func (c *Client) sendText(chatID, message, parseMode string) error {
	if !c.ready() {
		return nil
	}
	id, err := resolveChat(chatID)
	if err != nil {
		return err
	}

	msg := tgbotapi.NewMessage(id, message)
	msg.ParseMode = parseMode
	return c.send(chatID, msg)
}

// ready reports whether the client can send. A nil or uninitialized client sends nothing.
func (c *Client) ready() bool {
	return c != nil && c.retrier != nil
}

// resolveChat parses a chat ID, which Telegram takes as a base-10 int64.
func resolveChat(chatID string) (int64, error) {
	id, err := parseInt64(chatID)
	if err != nil {
		return 0, fmt.Errorf("invalid chat ID: %w", err)
	}
	return id, nil
}

// send hands chattable to a per-call copy of the retrying sender, whose logs carry chatID, and
// logs the outcome for chatID.
func (c *Client) send(chatID string, chattable tgbotapi.Chattable) error {
	retrier := c.retrier.withLogger(c.logger.With(zap.String("chatID", chatID)))
	if _, err := retrier.Send(chattable); err != nil {
		c.logger.Error("Failed to send Telegram message",
			zap.String("chatID", chatID),
			zap.Error(err))
		return err
	}
	c.logger.Info("Telegram message sent successfully", zap.String("chatID", chatID))
	return nil
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
	updateConfig.Timeout = int(c.pollTimeout.Seconds())
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
