package processor

import (
	"automation-hub/internal/messages"
	"automation-hub/internal/services/media"
	"html"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestCardPresenter(t *testing.T) {
	cat, err := messages.Load(nil)
	if err != nil {
		t.Fatal(err)
	}
	p := CardPresenter{Catalog: cat}
	m := &media.Metadata{Title: "<b>Tom & Jerry", Year: "2026", IMDbID: "tt123", IMDbRating: "7.3", RottenTomatoes: "85%", Genre: "Comedy", Runtime: "103 min", Poster: "url"}
	text, err := p.Present(m, media.Release{Tags: []string{"WEB-DL", "4k"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{"&lt;b&gt;Tom &amp; Jerry", "2026", "tt123", "7.3", "85%", "Comedy", "103 min", "4k"} {
		if !strings.Contains(text, s) {
			t.Errorf("missing %q in %q", s, text)
		}
	}
	m.Genre = strings.Repeat("x", 2000)
	text, err = p.Present(m, media.Release{})
	if err != nil || utf8.RuneCountInString(text) > 1024 || !strings.Contains(text, "tt123") || !strings.Contains(text, "Tom") {
		t.Fatalf("long caption %q: %v", text, err)
	}
	longTitle := strings.Repeat("<&", 1200)
	text, err = p.Present(&media.Metadata{Title: longTitle, IMDbID: "tt123"}, media.Release{})
	if err != nil || utf8.RuneCountInString(text) > 1024 || strings.Contains(text, "tt123") || strings.Contains(text, "&lt<b>") {
		t.Fatalf("long title %q: %v", text, err)
	}
	text, err = p.Present(&media.Metadata{Title: "Title"}, media.Release{})
	if err != nil || strings.Contains(text, "IMDb") {
		t.Fatalf("missing fields %q: %v", text, err)
	}
}

func TestCardPresenterLongTitle(t *testing.T) {
	for _, tt := range []struct{ name, title, format string }{
		{"large unicode and escaping", strings.Repeat("🌊<&", 12000), "<b>{title}</b>"},
		{"repeated title override", strings.Repeat("é&<", 12000), "<b>{title}</b> / <i>{title}</i>"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cat, err := messages.Load(map[string]any{"torrent": map[string]any{"card": map[string]any{"title": tt.format}}})
			if err != nil {
				t.Fatal(err)
			}
			got, err := (CardPresenter{Catalog: cat}).Present(&media.Metadata{Title: tt.title, IMDbID: "tt123"}, media.Release{})
			if err != nil {
				t.Fatal(err)
			}
			escaped := checkLongTitleCaption(t, tt.title, got)
			if tt.name == "repeated title override" && !strings.Contains(got, "<i>"+escaped+"</i>") {
				t.Fatalf("repeated title differs: %q", got)
			}
		})
	}
}

// checkLongTitleCaption asserts that a caption with a long title fits, has no IMDb link, and keeps a valid
// escaped prefix of the title. It returns that escaped prefix.
func checkLongTitleCaption(t *testing.T, title, got string) string {
	t.Helper()
	if utf8.RuneCountInString(got) > telegramCaptionLimit || strings.Contains(got, "tt123") {
		t.Fatalf("invalid caption length or link: %d", utf8.RuneCountInString(got))
	}
	titlePart, _, ok := strings.Cut(got, "</b>")
	if !ok {
		t.Fatalf("missing title closing tag: %q", got)
	}
	escaped := strings.TrimPrefix(titlePart, "<b>")
	raw := html.UnescapeString(escaped)
	if raw == "" || !strings.HasPrefix(title, raw) || html.EscapeString(raw) != escaped {
		t.Fatalf("invalid escaped title prefix: %q", escaped)
	}
	return escaped
}

func TestCardPresenterOverlongOverrides(t *testing.T) {
	for _, tt := range []struct {
		name, key, override string
		metadata            media.Metadata
		release             media.Release
		want, absent        string
	}{
		{name: "link override", key: "link", override: strings.Repeat("x", 1200) + "{id}", metadata: media.Metadata{Title: "Dune", IMDbID: "tt123", IMDbRating: "7.3"}, want: "<b>Dune</b>", absent: "tt123"},
		{name: "title override", key: "title", override: "<b>" + strings.Repeat("x", 1200) + "{title}</b>", metadata: media.Metadata{Title: "Dune", IMDbID: "tt123"}, want: "Dune", absent: "tt123"},
		{name: "year override", key: "year", override: strings.Repeat("x", 1200) + "{year}", metadata: media.Metadata{Title: "Dune", Year: "2021", IMDbID: "tt123"}, want: "<b>Dune</b>", absent: "tt123"},
		{name: "episode override", key: "episode", override: strings.Repeat("x", 1200) + "{episode}", metadata: media.Metadata{Title: "Dune", IMDbID: "tt123"}, release: media.Release{Episode: "S01E01"}, want: "<b>Dune</b>", absent: "tt123"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cat, err := messages.Load(map[string]any{"torrent": map[string]any{"card": map[string]any{tt.key: tt.override}}})
			if err != nil {
				t.Fatal(err)
			}
			got, err := (CardPresenter{Catalog: cat}).Present(&tt.metadata, tt.release)
			if err != nil {
				t.Fatal(err)
			}
			if utf8.RuneCountInString(got) > telegramCaptionLimit || !strings.Contains(got, tt.want) || strings.Contains(got, tt.absent) {
				t.Fatalf("unexpected caption (%d): %q", utf8.RuneCountInString(got), got)
			}
			if strings.Contains(got, "<b>") != strings.Contains(got, "</b>") {
				t.Fatalf("unbalanced title: %q", got)
			}
		})
	}
}
