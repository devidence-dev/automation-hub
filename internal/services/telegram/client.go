package telegram

import (
	"fmt"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type Client struct {
	bot    *tgbotapi.BotAPI
	logger *zap.Logger
}

func NewClient(token string, logger *zap.Logger) *Client {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		logger.Fatal("Failed to create Telegram bot", zap.Error(err))
	}

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

	_, err = c.bot.Send(msg)
	if err != nil {
		c.logger.Error("Failed to send Telegram message",
			zap.String("chatID", chatID),
			zap.Error(err))
		return err
	}

	return nil
}

func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
