package processor

import (
	"testing"

	"go.uber.org/zap"

	"automation-hub/internal/config"
	"automation-hub/internal/models"
)

func TestNewGenericEmailProcessor_CustomPattern(t *testing.T) {
	logger := zap.NewNop()
	cfg := config.ServiceProcessorConfig{
		EmailFrom:       "test@example.com",
		EmailSubject:    []string{"Security Code"},
		TelegramChatID:  "123",
		TelegramMessage: "Code is %s",
		CodePattern:     `\b[0-9]{4}\b`,
	}

	p := NewGenericEmailProcessor("custom_service", cfg, nil, logger)

	if p.GetName() != "custom_service" {
		t.Errorf("Expected name custom_service, got %s", p.GetName())
	}
	if p.GetSender() != "test@example.com" {
		t.Errorf("Expected sender test@example.com, got %s", p.GetSender())
	}
}

func TestNewGenericEmailProcessor_InvalidCustomPatternFallback(t *testing.T) {
	logger := zap.NewNop()
	cfg := config.ServiceProcessorConfig{
		EmailFrom:       "test@example.com",
		EmailSubject:    []string{"Code"},
		CodePattern:     `[invalid regex (`,
	}

	p := NewGenericEmailProcessor("default", cfg, nil, logger)
	if p.codePattern == nil {
		t.Error("Expected fallback codePattern to be initialized")
	}
}

func TestNewGenericEmailProcessor_BuiltInPatterns(t *testing.T) {
	logger := zap.NewNop()
	cfg := config.ServiceProcessorConfig{
		EmailFrom: "test@example.com",
	}

	cfProcessor := NewGenericEmailProcessor("cloudflare", cfg, nil, logger)
	if cfProcessor.codePattern == nil {
		t.Error("Expected Cloudflare default pattern")
	}

	pxProcessor := NewGenericEmailProcessor("perplexity", cfg, nil, logger)
	if pxProcessor.codePattern == nil {
		t.Error("Expected Perplexity default pattern")
	}
}

func TestShouldProcess(t *testing.T) {
	logger := zap.NewNop()
	cfg := config.ServiceProcessorConfig{
		EmailFrom:    "alert@service.com",
		EmailSubject: []string{"Verification", "Login Code"},
	}

	p := NewGenericEmailProcessor("test", cfg, nil, logger)

	tests := []struct {
		name     string
		email    models.Email
		expected bool
	}{
		{
			name: "Matching sender and subject",
			email: models.Email{
				From:    "alert@service.com",
				Subject: "Your Verification Code",
			},
			expected: true,
		},
		{
			name: "Matching sender, second subject",
			email: models.Email{
				From:    "support@alert@service.com",
				Subject: "Your Login Code",
			},
			expected: true,
		},
		{
			name: "Mismatched sender",
			email: models.Email{
				From:    "other@domain.com",
				Subject: "Your Verification Code",
			},
			expected: false,
		},
		{
			name: "Matching sender, mismatched subject",
			email: models.Email{
				From:    "alert@service.com",
				Subject: "Weekly Newsletter",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p.ShouldProcess(tt.email)
			if result != tt.expected {
				t.Errorf("ShouldProcess() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestExtractCode(t *testing.T) {
	logger := zap.NewNop()
	cfg := config.ServiceProcessorConfig{
		EmailFrom:   "test@example.com",
		CodePattern: `\b\d{6}\b`,
	}

	p := NewGenericEmailProcessor("default", cfg, nil, logger)

	bodyWithHeaders := "Header: Value\n\nYour code is 123456"
	code := p.extractCode(bodyWithHeaders)
	if code != "123456" {
		t.Errorf("Expected extracted code 123456, got %s", code)
	}

	noMatchBody := "Header: Value\n\nNo numbers here"
	codeNotFound := p.extractCode(noMatchBody)
	if codeNotFound != NotFoundCode {
		t.Errorf("Expected %s, got %s", NotFoundCode, codeNotFound)
	}

	cfProc := NewGenericEmailProcessor("cloudflare", config.ServiceProcessorConfig{EmailFrom: "test@example.com"}, nil, logger)
	cfCode := cfProc.extractCode("Your Cloudflare verification code is 654321")
	if cfCode != "654321" {
		t.Errorf("Expected Cloudflare code 654321, got %s", cfCode)
	}
}

func TestExtractPerplexityCode(t *testing.T) {
	logger := zap.NewNop()
	cfg := config.ServiceProcessorConfig{
		EmailFrom: "perplexity@example.com",
	}

	p := NewGenericEmailProcessor("perplexity", cfg, nil, logger)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Perplexity Spanish marker with numeric code",
			input:    "Inicie sesión directamente:\n123456",
			expected: "123456",
		},
		{
			name:     "Perplexity English marker with hyphenated code",
			input:    "Log in directly:\naw9s5-y1zoy",
			expected: "aw9s5-y1zoy",
		},
		{
			name:     "Perplexity marker missing",
			input:    "Welcome to Perplexity. Click here to confirm.",
			expected: NotFoundCode,
		},
		{
			name:     "Perplexity marker present but no valid code",
			input:    "Log in directly: !!!",
			expected: NotFoundCode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.extractCode(tt.input)
			if got != tt.expected {
				t.Errorf("extractPerplexityCode() = %s, expected %s", got, tt.expected)
			}
		})
	}
}

func TestStripMIMEHeaders(t *testing.T) {
	logger := zap.NewNop()
	p := NewGenericEmailProcessor("test", config.ServiceProcessorConfig{}, nil, logger)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Unix newlines header separator",
			input:    "From: foo\nSubject: bar\n\nBody content",
			expected: "Body content",
		},
		{
			name:     "Windows newlines header separator",
			input:    "From: foo\r\nSubject: bar\r\n\r\nBody content",
			expected: "Body content",
		},
		{
			name:     "No header separator",
			input:    "Plain body without headers",
			expected: "Plain body without headers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.stripMIMEHeaders(tt.input)
			if got != tt.expected {
				t.Errorf("stripMIMEHeaders() = %q, expected %q", got, tt.expected)
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	if got := truncateString("hello world", 5); got != "hello..." {
		t.Errorf("truncateString() = %s, expected hello...", got)
	}
	if got := truncateString("short", 10); got != "short" {
		t.Errorf("truncateString() = %s, expected short", got)
	}
}

func TestDecodeQuotedPrintable(t *testing.T) {
	logger := zap.NewNop()
	p := NewGenericEmailProcessor("test", config.ServiceProcessorConfig{}, nil, logger)

	quoted := "Hello=20World=21"
	decoded := p.decodeQuotedPrintable(quoted)
	if decoded != "Hello World!" {
		t.Errorf("decodeQuotedPrintable() = %s, expected Hello World!", decoded)
	}

	plain := "Hello World"
	if p.decodeQuotedPrintable(plain) != plain {
		t.Errorf("decodeQuotedPrintable() plain text modified")
	}
}
