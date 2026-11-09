package messages

import (
	"strings"
	"testing"
)

func mustLoad(t *testing.T, overrides map[string]any) *Catalog {
	t.Helper()
	cat, err := Load(overrides)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	return cat
}

func TestLoadNestedOverride(t *testing.T) {
	defaults := mustLoad(t, nil)
	cat := mustLoad(t, map[string]any{"caption": map[string]any{"header": "🔔 <b>Deps</b> · {date}"}})

	if got, _ := cat.Lookup("caption.header"); got != "🔔 <b>Deps</b> · {date}" {
		t.Errorf("caption.header = %q", got)
	}
	for _, key := range defaults.Keys() {
		if key == "caption.header" {
			continue
		}
		want, _ := defaults.Lookup(key)
		if got, _ := cat.Lookup(key); got != want {
			t.Errorf("%s = %q, want the default %q", key, got, want)
		}
	}
}

func TestDefaultLangIsEnglish(t *testing.T) {
	if got, _ := mustLoad(t, nil).Lookup("lang"); got != "en" {
		t.Errorf("default lang = %q, want en", got)
	}
}

func TestLoadLangOverrideKeepsLabels(t *testing.T) {
	defaults := mustLoad(t, nil)
	cat := mustLoad(t, map[string]any{"lang": "es"})

	if got, _ := cat.Lookup("lang"); got != "es" {
		t.Errorf("lang = %q, want es", got)
	}
	for _, key := range defaults.Keys() {
		if key == "lang" {
			continue
		}
		want, _ := defaults.Lookup(key)
		if got, _ := cat.Lookup(key); got != want {
			t.Errorf("%s = %q, want the English default %q", key, got, want)
		}
	}
}

func TestLoadRejectsUnknownKey(t *testing.T) {
	_, err := Load(map[string]any{"caption": map[string]any{"headr": "x"}})
	if err == nil || !strings.Contains(err.Error(), "caption.headr") {
		t.Errorf("Load() error = %v, want one naming caption.headr", err)
	}
}

func TestLoadRejectsDroppedPlaceholder(t *testing.T) {
	_, err := Load(map[string]any{"caption": map[string]any{"more_repos": "… más"}})
	if err == nil || !strings.Contains(err.Error(), "caption.more_repos") || !strings.Contains(err.Error(), "{count}") {
		t.Errorf("Load() error = %v, want one naming caption.more_repos and {count}", err)
	}
}

func TestLoadRejectsAddedPlaceholder(t *testing.T) {
	_, err := Load(map[string]any{"caption": map[string]any{"errors": "⚠️ {count} {extra}"}})
	if err == nil || !strings.Contains(err.Error(), "{extra}") {
		t.Errorf("Load() error = %v, want one naming {extra}", err)
	}
}

func TestLoadRejectsMapWhereTextExpected(t *testing.T) {
	tests := []struct {
		name      string
		overrides map[string]any
		wantKey   string
	}{
		{name: "map under a text", overrides: map[string]any{"fork": map[string]any{"tag_caption": map[string]any{"x": "y"}}}, wantKey: "fork.tag_caption"},
		{name: "text where a group is expected", overrides: map[string]any{"caption": "Dependency report"}, wantKey: "caption"},
		{name: "number where a text is expected", overrides: map[string]any{"lang": 5}, wantKey: "lang"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Load(tt.overrides)
			if err == nil || !strings.Contains(err.Error(), tt.wantKey) {
				t.Errorf("Load() error = %v, want one naming %s", err, tt.wantKey)
			}
		})
	}
}

func TestLoadRejectsInvalidDefaults(t *testing.T) {
	if _, err := load([]byte("a: [unclosed"), nil); err == nil {
		t.Error("load() accepted invalid YAML")
	}
	if _, err := load([]byte("a:\n  b: 1\n"), nil); err == nil || !strings.Contains(err.Error(), "a.b") {
		t.Errorf("load() error = %v, want one naming a.b", err)
	}
}

func TestRenderFillsPlaceholders(t *testing.T) {
	cat := mustLoad(t, nil)
	got, err := cat.Render("caption.header", map[string]string{"date": "2026-10-12"})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if want := "📦 <b>Dependency report</b> · 2026-10-12"; got != want {
		t.Errorf("Render() = %q, want %q", got, want)
	}
}

func TestRenderMissingValueErrors(t *testing.T) {
	cat := mustLoad(t, nil)
	if _, err := cat.Render("caption.header", nil); err == nil || !strings.Contains(err.Error(), "{date}") {
		t.Errorf("Render() error = %v, want one naming {date}", err)
	}
}

func TestRenderUnknownKeyErrors(t *testing.T) {
	cat := mustLoad(t, nil)
	if _, err := cat.Render("caption.headr", nil); err == nil {
		t.Error("Render() accepted an unknown key")
	}
}

func TestRenderKeepsValuesWhole(t *testing.T) {
	// A value that looks like a placeholder is inserted as written, not filled again.
	cat := mustLoad(t, nil)
	got, err := cat.Render("caption.header", map[string]string{"date": "{count}"})
	if err != nil || got != "📦 <b>Dependency report</b> · {count}" {
		t.Errorf("Render() = %q, %v", got, err)
	}
}

func TestTorrentCardOverride(t *testing.T) {
	cat := mustLoad(t, map[string]any{"torrent": map[string]any{"card": map[string]any{"tags": "Tags: {tags}"}}})
	got, err := cat.Render("torrent.card.tags", map[string]string{"tags": "4k"})
	if err != nil || got != "Tags: 4k" {
		t.Fatalf("render %q: %v", got, err)
	}
}

func TestTorrentCardTexts(t *testing.T) {
	cat := mustLoad(t, nil)
	tests := []struct {
		key  string
		vars map[string]string
		want string
	}{
		{key: "torrent.card.imdb_url", vars: map[string]string{"id": "tt1756855"}, want: "https://www.imdb.com/title/tt1756855/"},
		{key: "torrent.card.title_linked", vars: map[string]string{"url": "https://www.imdb.com/title/tt1/", "title": "Tom"}, want: `<a href="https://www.imdb.com/title/tt1/"><b>Tom</b></a>`},
		{key: "torrent.card.imdb_votes", vars: map[string]string{"votes": "12,345"}, want: " (12,345 votes)"},
		{key: "torrent.card.metacritic", vars: map[string]string{"rating": "62/100"}, want: "Ⓜ️ Metacritic: 62/100"},
		{key: "torrent.card.director", vars: map[string]string{"director": "Dave Green"}, want: "🎬 Director: Dave Green"},
		{key: "torrent.card.cast", vars: map[string]string{"cast": "A, B"}, want: "👥 Cast: A, B"},
		{key: "torrent.card.plot", vars: map[string]string{"plot": "Short."}, want: "<blockquote expandable>Short.</blockquote>"},
		{key: "torrent.card.plot_truncated", vars: map[string]string{"text": "Short"}, want: "Short…"},
		{key: "torrent.card.button", want: "Open on IMDb"},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got, err := cat.Render(tt.key, tt.vars)
			if err != nil || got != tt.want {
				t.Errorf("Render(%s) = %q, %v; want %q", tt.key, got, err, tt.want)
			}
		})
	}
}

func TestTorrentCardButtonOverride(t *testing.T) {
	cat := mustLoad(t, map[string]any{"torrent": map[string]any{"card": map[string]any{"button": "Ver en IMDb"}}})
	got, err := cat.Render("torrent.card.button", nil)
	if err != nil || got != "Ver en IMDb" {
		t.Fatalf("render %q: %v", got, err)
	}
}

func TestTorrentCardLinkKeyRemoved(t *testing.T) {
	if _, ok := mustLoad(t, nil).Lookup("torrent.card.link"); ok {
		t.Error("torrent.card.link is still a default key")
	}
	_, err := Load(map[string]any{"torrent": map[string]any{"card": map[string]any{"link": "x"}}})
	if err == nil || !strings.Contains(err.Error(), "torrent.card.link") {
		t.Errorf("Load() error = %v, want one naming torrent.card.link", err)
	}
}
