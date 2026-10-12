package depreport

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"
)

func presentCaption(t *testing.T, overrides map[string]any, s Summary) string {
	t.Helper()
	p := NewCaptionPresenter(catalogWith(t, overrides), loadZone(t, "America/Guayaquil"), "2006-01-02")
	out, err := p.Present(s)
	if err != nil {
		t.Fatalf("Present() error = %v", err)
	}
	return string(out)
}

func TestCaptionFixture(t *testing.T) {
	got := presentCaption(t, nil, summaryAt(loadFixtureReport(t)))
	if want := readGolden(t, "caption.txt"); got != want {
		t.Errorf("caption =\n%s\nwant\n%s", got, want)
	}
	if n := utf8.RuneCountInString(got); n > captionLimit {
		t.Errorf("caption has %d runes, over %d", n, captionLimit)
	}
}

func TestCaptionNothingPending(t *testing.T) {
	s := Summary{ReceivedAt: fixedAt, Repos: 1, Deps: Tab{Total: 42, UpToDate: 42}}
	want := "📦 <b>Dependency report</b> · 2026-10-12\n\n✅ Todo al día\n\n📊 0 desactualizadas · 42 al día\n🐳 Imágenes: 0 · 📚 Paquetes: 0"
	if got := presentCaption(t, nil, s); got != want {
		t.Errorf("caption = %q, want %q", got, want)
	}
}

func TestCaptionRepoLines(t *testing.T) {
	tests := []struct {
		name  string
		repos []string
		want  string
	}{
		{name: "major and patch", repos: []string{repoJSON("devidence-dev/QuantWarden", "major", "patch")}, want: "🔴 QuantWarden — 1 major · 1 patch"},
		{name: "minor only", repos: []string{repoJSON("devidence-dev/homelab", "minor")}, want: "🟡 homelab — 1 minor"},
		{name: "fork", repos: []string{repoJSON(testForkRepo, "minor")}, want: "🟡 discord-tts-bot (fork) — 1 minor"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := presentCaption(t, nil, summaryAt(reportOf(t, tt.repos...)))
			if !strings.Contains(got, tt.want+"\n") {
				t.Errorf("caption lacks %q:\n%s", tt.want, got)
			}
		})
	}
}

func TestCaptionEscapesNames(t *testing.T) {
	got := presentCaption(t, nil, summaryAt(reportOf(t, repoJSON("devidence-dev/o<r>", "patch"))))
	if !strings.Contains(got, "o&lt;r&gt;") || strings.Contains(got, "<r>") {
		t.Errorf("caption does not escape the name:\n%s", got)
	}
}

func TestCaptionOverflow(t *testing.T) {
	var repos []string
	for i := 0; i < 60; i++ {
		name := "devidence-dev/" + strings.Repeat("long-repository-name-", 3) + string(rune('a'+i%26)) + strings.Repeat("x", i/26)
		repos = append(repos, repoJSON(name, "patch"))
	}
	got := presentCaption(t, nil, summaryAt(reportOf(t, repos...)))
	if n := utf8.RuneCountInString(got); n > captionLimit {
		t.Fatalf("caption has %d runes, over %d", n, captionLimit)
	}
	kept := strings.Count(got, "🟡 ")
	more := fmt.Sprintf("… y %d repos más (ver reporte)", 60-kept)
	if kept == 60 || !strings.Contains(got, more) {
		t.Errorf("kept %d lines; caption lacks %q:\n%s", kept, more, got)
	}
}

func TestCaptionHeaderOverride(t *testing.T) {
	got := presentCaption(t, map[string]any{"caption": map[string]any{"header": "🔔 <b>Deps</b> · {date}"}}, summaryAt(loadFixtureReport(t)))
	if !strings.HasPrefix(got, "🔔 <b>Deps</b> · 2026-10-12\n\n") {
		t.Errorf("caption = %q, want the overridden header", got[:40])
	}
}

func TestCaptionDateLayout(t *testing.T) {
	p := NewCaptionPresenter(catalogWith(t, nil), loadZone(t, "America/Guayaquil"), "20060102")
	out, err := p.Present(summaryAt(loadFixtureReport(t)))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(out), "📦 <b>Dependency report</b> · 20261012\n") {
		t.Errorf("caption = %q, want the date in the layout", string(out)[:50])
	}
}
