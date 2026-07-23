package telegram

import (
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

func TestParseInt64(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int64
		wantErr bool
	}{
		{
			name:    "Valid positive integer",
			input:   "123456789",
			want:    123456789,
			wantErr: false,
		},
		{
			name:    "Valid negative integer",
			input:   "-987654321",
			want:    -987654321,
			wantErr: false,
		},
		{
			name:    "Invalid string",
			input:   "not_a_number",
			want:    0,
			wantErr: true,
		},
		{
			name:    "Empty string",
			input:   "",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseInt64(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseInt64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSendMessageInvalidChatID(t *testing.T) {
	logger := zap.NewNop()
	client := &Client{
		bot:    &tgbotapi.BotAPI{},
		logger: logger,
	}

	err := client.SendMessage("invalid_chat_id", "Hello World")
	if err == nil {
		t.Error("Expected error when sending message with invalid chat ID, got nil")
	}
}

func TestSendMessageNilClientOrBot(t *testing.T) {
	logger := zap.NewNop()
	var nilClient *Client
	if err := nilClient.SendMessage("123456", "Hello"); err != nil {
		t.Errorf("Expected nil error for nil client, got %v", err)
	}

	clientWithNilBot := &Client{
		bot:    nil,
		logger: logger,
	}
	if err := clientWithNilBot.SendMessage("123456", "Hello"); err != nil {
		t.Errorf("Expected nil error for client with nil bot, got %v", err)
	}
}
