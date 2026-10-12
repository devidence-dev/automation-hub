package config

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestEffectiveDefaults(t *testing.T) {
	strip := false
	tests := []struct {
		name string
		got  any
		want any
	}{
		{"polling interval unset", EmailConfig{}.EffectivePollingInterval(), 60 * time.Second},
		{"polling interval set", EmailConfig{PollingInterval: 20}.EffectivePollingInterval(), 20 * time.Second},
		{"preview unset", EmailConfig{}.EffectiveLogPreviewLength(), 200},
		{"preview set", EmailConfig{LogPreviewLength: 50}.EffectiveLogPreviewLength(), 50},
		{"extractor unset", ServiceProcessorConfig{}.EffectiveExtractor(), ExtractorPattern},
		{"extractor set", ServiceProcessorConfig{Extractor: ExtractorMarker}.EffectiveExtractor(), ExtractorMarker},
		{"code pattern unset", ServiceProcessorConfig{}.EffectiveCodePattern(), `\b[a-zA-Z0-9]{4,8}\b`},
		{"code pattern set", ServiceProcessorConfig{CodePattern: `\d{6}`}.EffectiveCodePattern(), `\d{6}`},
		{"not found unset", ServiceProcessorConfig{}.EffectiveNotFoundText(), "Not found"},
		{"not found set", ServiceProcessorConfig{NotFoundText: "No code in email"}.EffectiveNotFoundText(), "No code in email"},
		{"strip unset", ServiceProcessorConfig{}.EffectiveStripMIMEHeaders(), true},
		{"strip false", ServiceProcessorConfig{StripMIMEHeaders: &strip}.EffectiveStripMIMEHeaders(), false},
		{"body cap unset", WebhookProcessorConfig{}.EffectiveMaxBodyBytes(), int64(1048576)},
		{"body cap set", WebhookProcessorConfig{MaxBodyBytes: 1024}.EffectiveMaxBodyBytes(), int64(1024)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(tt.got, tt.want) {
				t.Errorf("got %v (%T), want %v (%T)", tt.got, tt.got, tt.want, tt.want)
			}
		})
	}
}

func TestLoad_NegativePollingInterval(t *testing.T) {
	_, err := loadYAML(t, "email:\n  polling_interval: -5\n")
	if err == nil || !strings.Contains(err.Error(), "email.polling_interval") {
		t.Fatalf("error = %v, want one naming email.polling_interval", err)
	}
}

func TestLoad_NegativeMaxBodyBytes(t *testing.T) {
	content := "hook:\n  - name: qbittorrent\n    path: /webhook/qbittorrent\n    config:\n      max_body_bytes: -1\n"
	_, err := loadYAML(t, content)
	if err == nil || !strings.Contains(err.Error(), "hook.qbittorrent.config.max_body_bytes") {
		t.Fatalf("error = %v, want one naming hook.qbittorrent.config.max_body_bytes", err)
	}
}

func TestLoad_EmailDefaults(t *testing.T) {
	cfg, err := loadYAML(t, "email:\n  host: imap.example.com\n")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got := cfg.Email.EffectivePollingInterval(); got != 60*time.Second {
		t.Errorf("poll interval = %s, want 60s", got)
	}
	if got := cfg.Email.EffectiveLogPreviewLength(); got != 200 {
		t.Errorf("log preview = %d, want 200", got)
	}
}

func TestValidateEmailServices(t *testing.T) {
	const header = "email:\n  services:\n"
	tests := []struct {
		name    string
		service string
		wantErr string
	}{
		{
			name:    "invalid pattern names the service",
			service: "    - name: github\n      config:\n        code_pattern: \"[invalid regex (\"\n",
			wantErr: `service "github"`,
		},
		{
			name:    "unknown extractor names the value",
			service: "    - name: perplexity\n      config:\n        extractor: regex\n",
			wantErr: `"regex"`,
		},
		{
			name:    "marker without markers names the service",
			service: "    - name: perplexity\n      config:\n        extractor: marker\n",
			wantErr: `service "perplexity"`,
		},
		{
			name:    "invalid candidate names the service",
			service: "    - name: perplexity\n      config:\n        extractor: marker\n        code_markers: [\"directly:\"]\n        code_candidates: [\"(\"]\n",
			wantErr: `service "perplexity"`,
		},
		{
			name: "minimal service loads with defaults",
			service: "    - name: github\n      config:\n        email_from: noreply@github.com\n" +
				"        email_subject: [\"verification code\"]\n        telegram_chat_id: \"1\"\n" +
				"        telegram_message: \"Code: %s\"\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := loadYAML(t, header+tt.service)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Load() error = %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %v, want one containing %s", err, tt.wantErr)
			}
		})
	}
}

// TestLoad_MigratedEmailServices loads the CONFIG_YAML migration of task 6.1 and checks the resolved
// extractor settings of cloudflare and perplexity.
func TestLoad_MigratedEmailServices(t *testing.T) {
	content := `email:
  services:
    - name: "cloudflare"
      config:
        email_from: "noreply@notify.cloudflare.com"
        email_subject: ["devidence.dev"]
        telegram_chat_id: "1"
        telegram_message: "🛡️ Cloudflare App Code: \n` + "```%s```" + `"
        strip_mime_headers: false
        code_pattern: "\\b\\d{6}\\b"
    - name: "perplexity"
      config:
        email_from: "team@mail.perplexity.ai"
        email_subject: ["Sign in to Perplexity"]
        telegram_chat_id: "2"
        telegram_message: "🔮 Perplexity Code: ` + "```%s```" + `"
        extractor: marker
        code_markers: ["directly:", "directamente:"]
        code_candidates: ["\\b(\\d{5,6})\\b", "\\b([a-zA-Z0-9]+-[a-zA-Z0-9]+)\\b"]
        code_pattern: "(?:\\d{5,6}|[a-zA-Z0-9]+-[a-zA-Z0-9]+)"
`
	cfg, err := loadYAML(t, content)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(cfg.Email.Services) != 2 {
		t.Fatalf("services = %d, want 2", len(cfg.Email.Services))
	}
	cloudflare := cfg.Email.Services[0].Config
	if got := cloudflare.EffectiveExtractor(); got != ExtractorPattern {
		t.Errorf("cloudflare extractor = %q, want pattern", got)
	}
	if got := cloudflare.EffectiveCodePattern(); got != `\b\d{6}\b` {
		t.Errorf("cloudflare code pattern = %q", got)
	}
	if cloudflare.EffectiveStripMIMEHeaders() {
		t.Error("cloudflare keeps its MIME headers, want strip_mime_headers false")
	}
	perplexity := cfg.Email.Services[1].Config
	if got := perplexity.EffectiveExtractor(); got != ExtractorMarker {
		t.Errorf("perplexity extractor = %q, want marker", got)
	}
	if want := []string{"directly:", "directamente:"}; !reflect.DeepEqual(perplexity.CodeMarkers, want) {
		t.Errorf("perplexity markers = %q, want %q", perplexity.CodeMarkers, want)
	}
	wantCandidates := []string{`\b(\d{5,6})\b`, `\b([a-zA-Z0-9]+-[a-zA-Z0-9]+)\b`}
	if !reflect.DeepEqual(perplexity.CodeCandidates, wantCandidates) {
		t.Errorf("perplexity candidates = %q, want %q", perplexity.CodeCandidates, wantCandidates)
	}
	if want := `(?:\d{5,6}|[a-zA-Z0-9]+-[a-zA-Z0-9]+)`; perplexity.CodePattern != want {
		t.Errorf("perplexity code pattern = %q, want %q", perplexity.CodePattern, want)
	}
	if !perplexity.EffectiveStripMIMEHeaders() {
		t.Error("perplexity strips headers by default, got strip_mime_headers false")
	}
}

func TestFindWebhook(t *testing.T) {
	cfg := &Config{
		Hook: []WebhookConfig{
			{Name: "qbittorrent", Path: "/webhook/qbittorrent", Config: WebhookProcessorConfig{TelegramChatID: "123", TelegramMessage: "Download finished: %s"}},
			{Name: "sonarr", Path: "/webhook/sonarr", Config: WebhookProcessorConfig{TelegramChatID: "456", TelegramMessage: "Series imported: %s"}},
		},
	}
	qbit := FindWebhook(cfg, "qbittorrent")
	if qbit == nil {
		t.Fatal("qbittorrent not found")
	}
	if qbit.TelegramChatID != "123" {
		t.Errorf("TelegramChatID = %s, want 123", qbit.TelegramChatID)
	}
	if FindWebhook(cfg, "non_existent") != nil {
		t.Error("non_existent found, want nil")
	}
}
