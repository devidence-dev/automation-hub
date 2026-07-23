package processor

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"

	"automation-hub/internal/config"
	"automation-hub/internal/models"
)

func TestProcessorManager(t *testing.T) {
	logger := zap.NewNop()
	emailCfg := config.EmailConfig{
		Services: []config.ServiceConfig{
			{
				Name: "cloudflare",
				Config: config.ServiceProcessorConfig{
					EmailFrom:       "no-reply@cloudflare.com",
					EmailSubject:    []string{"Verification Code"},
					TelegramChatID:  "123",
					TelegramMessage: "Code: %s",
				},
			},
			{
				Name: "perplexity",
				Config: config.ServiceProcessorConfig{
					EmailFrom:       "no-reply@perplexity.ai",
					EmailSubject:    []string{"Sign In"},
					TelegramChatID:  "123",
					TelegramMessage: "Login: %s",
				},
			},
		},
	}

	mgr := NewProcessorManager(emailCfg, nil, logger)

	processors := mgr.GetProcessors()
	if len(processors) != 2 {
		t.Fatalf("Expected 2 processors, got %d", len(processors))
	}

	// Test processing empty email slice
	ctx := context.Background()
	mgr.ProcessEmailsConcurrently(ctx, nil)

	// Test processing matching and non-matching emails
	emails := []models.Email{
		{
			From:      "no-reply@cloudflare.com",
			Subject:   "Your Verification Code is 123456",
			TextPlain: "Code: 123456",
		},
		{
			From:      "unrelated@spam.com",
			Subject:   "Buy drugs online",
			TextPlain: "Click here",
		},
	}

	mgr.ProcessEmailsConcurrently(ctx, emails)

	// Test context cancellation
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	mgr.ProcessEmailsConcurrently(canceledCtx, emails)
}

func TestProcessorManager_CanceledContextAsync(t *testing.T) {
	logger := zap.NewNop()
	emailCfg := config.EmailConfig{
		Services: []config.ServiceConfig{
			{
				Name: "cloudflare",
				Config: config.ServiceProcessorConfig{
					EmailFrom: "no-reply@cloudflare.com",
				},
			},
		},
	}

	mgr := NewProcessorManager(emailCfg, nil, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	time.Sleep(2 * time.Millisecond)
	cancel()

	email := models.Email{
		From:    "no-reply@cloudflare.com",
		Subject: "Verification Code",
	}

	// processEmailAsync should return early if context is canceled
	mgr.wg.Add(1)
	mgr.processEmailAsync(ctx, email)
}
