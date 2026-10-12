package processor

import (
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"automation-hub/internal/config"
)

func TestPatternExtractor(t *testing.T) {
	tests := []struct {
		name      string
		pattern   string
		text      string
		wantCode  string
		wantFound bool
	}{
		{
			name:      "capture group skips the hex color",
			pattern:   `[^#]\b(\d{6})\b`,
			text:      "color:#202123; padding: 0\nIngresa este codigo:\n662575\n",
			wantCode:  "662575",
			wantFound: true,
		},
		{name: "whole match without a group", pattern: `\b\d{6}\b`, text: "Code: 123456", wantCode: "123456", wantFound: true},
		{name: "default pattern", pattern: "", text: "Use 4821 now", wantCode: "4821", wantFound: true},
		{name: "no match", pattern: `\b\d{6}\b`, text: "no digits here", wantFound: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := newTestExtractor(t, config.ServiceProcessorConfig{CodePattern: tt.pattern}, config.DefaultLogPreviewLength)
			code, found := extractor.Extract(tt.text)
			if found != tt.wantFound || code != tt.wantCode {
				t.Errorf("Extract() = (%q, %v), want (%q, %v)", code, found, tt.wantCode, tt.wantFound)
			}
		})
	}
}

// TestPatternExtractor_LogPreviewCapped checks that a not-found warning shows at most the configured preview.
func TestPatternExtractor_LogPreviewCapped(t *testing.T) {
	core, logs := observer.New(zap.WarnLevel)
	extractor, err := newCodeExtractor(config.ServiceProcessorConfig{CodePattern: `\b\d{6}\b`}, 10, zap.New(core))
	if err != nil {
		t.Fatalf("newCodeExtractor() error = %v", err)
	}
	extractor.Extract("a long email body without any code")

	entries := logs.FilterMessage("Code not found in email").All()
	if len(entries) != 1 {
		t.Fatalf("got %d warnings, want 1", len(entries))
	}
	preview, _ := entries[0].ContextMap()["text_preview"].(string)
	if len(preview) != 13 || preview != "a long ema..." {
		t.Errorf("text_preview = %q, want 10 characters plus ...", preview)
	}
}

func TestMarkerExtractor(t *testing.T) {
	perplexity := config.ServiceProcessorConfig{
		Extractor:      config.ExtractorMarker,
		CodeMarkers:    []string{"directly:", "directamente:"},
		CodeCandidates: []string{`\b(\d{5,6})\b`, `\b([a-zA-Z0-9]+-[a-zA-Z0-9]+)\b`},
		CodePattern:    `(?:\d{5,6}|[a-zA-Z0-9]+-[a-zA-Z0-9]+)`,
	}
	tests := []struct {
		name      string
		text      string
		wantCode  string
		wantFound bool
	}{
		{name: "numeric code after the Spanish marker", text: "Inicie sesión directamente:\n123456", wantCode: "123456", wantFound: true},
		{name: "hyphenated code after the English marker", text: "Log in directly:\naw9s5-y1zoy", wantCode: "aw9s5-y1zoy", wantFound: true},
		{name: "marker matches case-insensitively", text: "LOG IN DIRECTLY:\naw9s5-y1zoy", wantCode: "aw9s5-y1zoy", wantFound: true},
		{name: "marker order wins over position", text: "directamente: 111111 ... directly: 222222", wantCode: "222222", wantFound: true},
		{name: "marker missing", text: "Welcome to Perplexity. Click here to confirm.", wantFound: false},
		{name: "marker present without a valid code", text: "Log in directly: !!!", wantFound: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := newTestExtractor(t, perplexity, config.DefaultLogPreviewLength)
			code, found := extractor.Extract(tt.text)
			if found != tt.wantFound || code != tt.wantCode {
				t.Errorf("Extract() = (%q, %v), want (%q, %v)", code, found, tt.wantCode, tt.wantFound)
			}
		})
	}
}

func TestMarkerExtractor_FallsBackToCodePattern(t *testing.T) {
	cfg := config.ServiceProcessorConfig{
		Extractor:   config.ExtractorMarker,
		CodeMarkers: []string{"directly:"},
		CodePattern: `[a-z]{3}`,
	}
	extractor := newTestExtractor(t, cfg, config.DefaultLogPreviewLength)
	code, found := extractor.Extract("Log in directly: abc")
	if !found || code != "abc" {
		t.Errorf("Extract() = (%q, %v), want (abc, true)", code, found)
	}
}

func newTestExtractor(t *testing.T, cfg config.ServiceProcessorConfig, previewLength int) CodeExtractor {
	t.Helper()
	extractor, err := newCodeExtractor(cfg, previewLength, zap.NewNop())
	if err != nil {
		t.Fatalf("newCodeExtractor() error = %v", err)
	}
	return extractor
}
