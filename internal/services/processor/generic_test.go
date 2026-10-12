package processor

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"automation-hub/internal/config"
	"automation-hub/internal/models"
)

// migratedCloudflare and migratedPerplexity are the service configs of the CONFIG_YAML migration (task 6.1).
var (
	migratedCloudflare = config.ServiceProcessorConfig{
		EmailFrom:        "noreply@notify.cloudflare.com",
		EmailSubject:     []string{"devidence.dev"},
		TelegramChatID:   "42",
		TelegramMessage:  "🛡️ Cloudflare App Code: \n```%s```",
		CodePattern:      `\b\d{6}\b`,
		StripMIMEHeaders: boolPtr(false),
	}
	migratedPerplexity = config.ServiceProcessorConfig{
		EmailFrom:       "team@mail.perplexity.ai",
		EmailSubject:    []string{"Sign in to Perplexity"},
		TelegramChatID:  "43",
		TelegramMessage: "🔮 Perplexity Code: ```%s```",
		Extractor:       config.ExtractorMarker,
		CodeMarkers:     []string{"directly:", "directamente:"},
		CodeCandidates:  []string{`\b(\d{5,6})\b`, `\b([a-zA-Z0-9]+-[a-zA-Z0-9]+)\b`},
		CodePattern:     `(?:\d{5,6}|[a-zA-Z0-9]+-[a-zA-Z0-9]+)`,
	}
)

func boolPtr(b bool) *bool {
	return &b
}

// newTestProcessor builds a processor with a nil sender, for tests that call its extraction only.
func newTestProcessor(t *testing.T, name string, cfg config.ServiceProcessorConfig) *GenericEmailProcessor {
	t.Helper()
	p, err := NewGenericEmailProcessor(name, cfg, config.DefaultLogPreviewLength, nil, zap.NewNop())
	if err != nil {
		t.Fatalf("NewGenericEmailProcessor(%s) error = %v", name, err)
	}
	return p
}

func TestNewGenericEmailProcessor_CustomPattern(t *testing.T) {
	cfg := config.ServiceProcessorConfig{
		EmailFrom:       "test@example.com",
		EmailSubject:    []string{"Security Code"},
		TelegramChatID:  "123",
		TelegramMessage: "Code is %s",
		CodePattern:     `\b[0-9]{4}\b`,
	}

	p := newTestProcessor(t, "custom_service", cfg)

	if p.GetName() != "custom_service" {
		t.Errorf("Expected name custom_service, got %s", p.GetName())
	}
	if p.GetSender() != "test@example.com" {
		t.Errorf("Expected sender test@example.com, got %s", p.GetSender())
	}
}

func TestNewGenericEmailProcessor_InvalidPatternFails(t *testing.T) {
	tests := []struct {
		name string
		cfg  config.ServiceProcessorConfig
	}{
		{name: "invalid code_pattern", cfg: config.ServiceProcessorConfig{CodePattern: `[invalid regex (`}},
		{name: "invalid candidate", cfg: config.ServiceProcessorConfig{Extractor: config.ExtractorMarker, CodeMarkers: []string{"x"}, CodeCandidates: []string{"("}}},
		{name: "unknown extractor", cfg: config.ServiceProcessorConfig{Extractor: "regex"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewGenericEmailProcessor("github", tt.cfg, config.DefaultLogPreviewLength, nil, zap.NewNop()); err == nil {
				t.Error("NewGenericEmailProcessor() error = nil, want an error")
			}
		})
	}
}

// TestNewGenericEmailProcessor_ExplicitConfig builds the two migrated services from explicit config only.
func TestNewGenericEmailProcessor_ExplicitConfig(t *testing.T) {
	cloudflare := newTestProcessor(t, "cloudflare", migratedCloudflare)
	perplexity := newTestProcessor(t, "perplexity", migratedPerplexity)
	if cloudflare.extractor == nil || perplexity.extractor == nil {
		t.Error("Expected both explicit services to get an extractor")
	}
}

func TestShouldProcess(t *testing.T) {
	cfg := config.ServiceProcessorConfig{
		EmailFrom:    "alert@service.com",
		EmailSubject: []string{"Verification", "Login Code"},
	}

	p := newTestProcessor(t, "test", cfg)

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

func TestShouldMarkAsRead(t *testing.T) {
	tests := []struct {
		name       string
		markAsRead *bool
		want       bool
	}{
		{name: "unset means read", markAsRead: nil, want: true},
		{name: "explicit true", markAsRead: boolPtr(true), want: true},
		{name: "explicit false", markAsRead: boolPtr(false), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newTestProcessor(t, "service", config.ServiceProcessorConfig{MarkAsRead: tt.markAsRead})
			if got := p.ShouldMarkAsRead(); got != tt.want {
				t.Errorf("ShouldMarkAsRead() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestExtractedCodeIsNotLogged checks that the Info log of a found code has the code length, and no field
// or message holds the code.
func TestExtractedCodeIsNotLogged(t *testing.T) {
	const code = "654321"
	core, logs := observer.New(zap.InfoLevel)
	p, err := NewGenericEmailProcessor("cloudflare", migratedCloudflare, config.DefaultLogPreviewLength, &fakeSender{}, zap.New(core))
	if err != nil {
		t.Fatalf("NewGenericEmailProcessor() error = %v", err)
	}
	if got := p.extractCode("Your code is " + code); got != code {
		t.Fatalf("extractCode() = %q, want %q", got, code)
	}

	entries := logs.FilterMessage("Code extracted successfully").All()
	if len(entries) != 1 {
		t.Fatalf("got %d extraction logs, want 1", len(entries))
	}
	fields := entries[0].ContextMap()
	if fields["code_length"] != int64(len(code)) || fields["service"] != "cloudflare" {
		t.Errorf("fields = %v, want service cloudflare and code_length %d", fields, len(code))
	}
	for key, value := range fields {
		if strings.Contains(fmt.Sprint(value), code) {
			t.Errorf("field %s holds the code", key)
		}
	}
	if strings.Contains(entries[0].Message, code) {
		t.Errorf("message holds the code")
	}
}

func TestExtractCode(t *testing.T) {
	p := newTestProcessor(t, "default", config.ServiceProcessorConfig{
		EmailFrom:   "test@example.com",
		CodePattern: `\b\d{6}\b`,
	})

	bodyWithHeaders := "Header: Value\n\nYour code is 123456"
	code := p.extractCode(bodyWithHeaders)
	if code != "123456" {
		t.Errorf("Expected extracted code 123456, got %s", code)
	}

	noMatchBody := "Header: Value\n\nNo numbers here"
	codeNotFound := p.extractCode(noMatchBody)
	if codeNotFound != config.DefaultNotFoundText {
		t.Errorf("Expected %s, got %s", config.DefaultNotFoundText, codeNotFound)
	}

	cfProc := newTestProcessor(t, "cloudflare", config.ServiceProcessorConfig{
		EmailFrom:        "test@example.com",
		CodePattern:      `\b\d{6}\b`,
		StripMIMEHeaders: boolPtr(false),
	})
	cfCode := cfProc.extractCode("Your Cloudflare verification code is 654321")
	if cfCode != "654321" {
		t.Errorf("Expected Cloudflare code 654321, got %s", cfCode)
	}
}

// TestExtractCode_IgnoresHexColor guards against a real production bug: an
// HTML email body with inline CSS hex colors (e.g. "color:#202123") got
// matched by a bare \b\d{6}\b before ever reaching the real 6-digit code
// further down. A pattern with a capturing group like "[^#]\b(\d{6})\b"
// (OpenAI's configured code_pattern) must skip the hex color and return the
// actual code via submatch group 1.
func TestExtractCode_IgnoresHexColor(t *testing.T) {
	cfg := config.ServiceProcessorConfig{
		EmailFrom:   "noreply@tm.openai.com",
		CodePattern: `[^#]\b(\d{6})\b`,
	}

	p := newTestProcessor(t, "openai", cfg)

	body := "background-color: #ffffff;color:#202123; padding: 0\n" +
		"Ingresa este codigo:\n662575\n"
	code := p.extractCode(body)
	if code != "662575" {
		t.Errorf("Expected 662575 (ignoring the #202123 hex color), got %s", code)
	}
}

// TestNoNameDefaults checks that a service's name alone changes nothing: "cloudflare" without keys
// strips the headers and uses the generic pattern.
func TestNoNameDefaults(t *testing.T) {
	p := newTestProcessor(t, "cloudflare", config.ServiceProcessorConfig{EmailFrom: "test@example.com"})
	got := p.extractCode("Subject: Code 111111\n\nYour code is 222222")
	if got != "Your" {
		t.Errorf("extractCode() = %q, want Your (generic pattern after the headers)", got)
	}
}

// TestMigratedServices proves that the migrated keys give today's Cloudflare and Perplexity behavior.
func TestMigratedServices(t *testing.T) {
	tests := []struct {
		name string
		cfg  config.ServiceProcessorConfig
		text string
		want string
	}{
		{name: "cloudflare plain code", cfg: migratedCloudflare, text: "Your Cloudflare verification code is 654321", want: "654321"},
		{name: "cloudflare keeps headers", cfg: migratedCloudflare, text: "Subject: Code 111111\n\nYour code is 222222", want: "111111"},
		{name: "perplexity Spanish marker", cfg: migratedPerplexity, text: "Inicie sesión directamente:\n123456", want: "123456"},
		{name: "perplexity English marker", cfg: migratedPerplexity, text: "Log in directly:\naw9s5-y1zoy", want: "aw9s5-y1zoy"},
		{name: "perplexity marker missing", cfg: migratedPerplexity, text: "Welcome to Perplexity.", want: config.DefaultNotFoundText},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newTestProcessor(t, "service", tt.cfg)
			if got := p.extractCode(tt.text); got != tt.want {
				t.Errorf("extractCode() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGenericEmailProcessor_Process(t *testing.T) {
	tests := []struct {
		name     string
		cfg      config.ServiceProcessorConfig
		text     string
		sendErr  error
		wantText string
		wantErr  bool
	}{
		{
			name:     "cloudflare example",
			cfg:      migratedCloudflare,
			text:     "Use this code directly: 123456",
			wantText: "🛡️ Cloudflare App Code: \n```123456```",
		},
		{
			name: "configured not-found text is sent",
			cfg: config.ServiceProcessorConfig{
				TelegramChatID:  "42",
				TelegramMessage: "🤖 ChatGPT Code: ```%s```",
				CodePattern:     `\b\d{6}\b`,
				NotFoundText:    "No code in email",
			},
			text:     "There is no code here",
			wantText: "🤖 ChatGPT Code: ```No code in email```",
		},
		{
			name:    "sender error is returned",
			cfg:     migratedCloudflare,
			text:    "Use this code directly: 123456",
			sendErr: errors.New("telegram unavailable"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sender := &fakeSender{err: tt.sendErr}
			p, err := NewGenericEmailProcessor("service", tt.cfg, config.DefaultLogPreviewLength, sender, zap.NewNop())
			if err != nil {
				t.Fatalf("NewGenericEmailProcessor() error = %v", err)
			}

			err = p.Process(models.Email{Subject: "subject", From: "from", TextPlain: tt.text})
			if (err != nil) != tt.wantErr {
				t.Fatalf("Process() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if len(sender.messages) != 1 || sender.messages[0] != tt.wantText {
				t.Errorf("sent %q, want %q", sender.messages, tt.wantText)
			}
		})
	}
}

func TestStripMIMEHeaders(t *testing.T) {
	p := newTestProcessor(t, "test", config.ServiceProcessorConfig{})

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
	p := newTestProcessor(t, "test", config.ServiceProcessorConfig{})

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
