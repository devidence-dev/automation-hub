package telegram

import (
	"testing"

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
	// Create client with dummy token
	client := &Client{
		bot:    nil,
		logger: logger,
	}

	err := client.SendMessage("invalid_chat_id", "Hello World")
	if err == nil {
		t.Error("Expected error when sending message with invalid chat ID, got nil")
	}
}
