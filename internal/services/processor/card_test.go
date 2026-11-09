package processor

import (
	"html"
	"strings"
	"testing"
	"unicode/utf8"

	"automation-hub/internal/messages"
	"automation-hub/internal/services/media"
)

// coyoteLink is the title line of fullMetadata: the IMDb anchor that a caption must keep.
const coyoteLink = `<a href="https://www.imdb.com/title/tt1756855/"><b>Coyote Vs Acme</b></a> (2026)`

// fullMetadata returns a title with every field set and a plot of 400 runes.
func fullMetadata() media.Metadata {
	return media.Metadata{
		Title: "Coyote Vs Acme", Year: "2026", IMDbID: "tt1756855", IMDbRating: "7.6", IMDbVotes: "12,345",
		RottenTomatoes: "85%", Metacritic: "62/100", Genre: "Comedy", Runtime: "103 min", Rated: "PG",
		Director: "Dave Green", Actors: "Will Coyote, Road Runner", Plot: strings.Repeat("word ", 80),
	}
}

// newPresenter returns a presenter over the default catalog with the given overrides and plot limit.
func newPresenter(t *testing.T, overrides map[string]any, plotMaxChars int) CardPresenter {
	t.Helper()
	cat, err := messages.Load(overrides)
	if err != nil {
		t.Fatal(err)
	}
	return CardPresenter{Catalog: cat, PlotMaxChars: plotMaxChars}
}

// cardOverride returns the catalog overrides that replace one torrent.card key with text.
func cardOverride(key, text string) map[string]any {
	return map[string]any{"torrent": map[string]any{"card": map[string]any{key: text}}}
}

// mustPresent renders the card of m and release, and fails the test on error.
func mustPresent(t *testing.T, p CardPresenter, m *media.Metadata, release media.Release) Card {
	t.Helper()
	card, err := p.Present(m, release)
	if err != nil {
		t.Fatal(err)
	}
	return card
}

// checkCaption asserts that the caption fits the Telegram limit, holds every want part in order, and holds none of absent.
func checkCaption(t *testing.T, caption string, want, absent []string) {
	t.Helper()
	if n := utf8.RuneCountInString(caption); n > telegramCaptionLimit {
		t.Fatalf("caption has %d runes, want at most %d: %q", n, telegramCaptionLimit, caption)
	}
	rest := caption
	for _, part := range want {
		i := strings.Index(rest, part)
		if i < 0 {
			t.Fatalf("missing %q after the previous part in %q", part, caption)
		}
		rest = rest[i+len(part):]
	}
	for _, s := range absent {
		if strings.Contains(caption, s) {
			t.Fatalf("caption contains %q: %q", s, caption)
		}
	}
}

type contentCase struct {
	name     string
	metadata media.Metadata
	release  media.Release
	plotMax  int
	want     []string
	absent   []string
}

func TestCardPresenterContent(t *testing.T) {
	for _, tt := range []contentCase{
		{
			name: "full card", metadata: fullMetadata(), release: media.Release{Tags: []string{"WEB-DL", "4k", "HDR"}}, plotMax: 200,
			want:   []string{coyoteLink, "IMDb: 7.6 (12,345 votes)", "Rotten Tomatoes: 85%", "Metacritic: 62/100", "Comedy", "103 min", "PG", "WEB-DL, 4k, HDR", "Director: Dave Green", "Cast: Will Coyote, Road Runner", "<blockquote expandable>word word", "word…</blockquote>"},
			absent: []string{"/downloads"},
		},
		{name: "missing fields", metadata: media.Metadata{Title: "Dune"}, plotMax: 200, want: []string{"<b>Dune</b>"}, absent: []string{"IMDb", "Director:", "Cast:", "blockquote", "\n"}},
		{name: "plot hidden at zero", metadata: media.Metadata{Title: "Dune", Plot: "A plot."}, plotMax: 0, want: []string{"<b>Dune</b>"}, absent: []string{"blockquote", "A plot."}},
		{name: "hostile title is escaped", metadata: media.Metadata{Title: "<b>Tom & Jerry"}, want: []string{"&lt;b&gt;Tom &amp; Jerry"}, absent: []string{"<b>Tom"}},
		{
			name: "hostile values are escaped", metadata: media.Metadata{Title: "Dune", Director: "<i>x</i> & y", Actors: "<script>", Plot: "<b>bold</b> & more"}, plotMax: 200,
			want:   []string{"Director: &lt;i&gt;x&lt;/i&gt; &amp; y", "Cast: &lt;script&gt;", "&lt;b&gt;bold&lt;/b&gt; &amp; more"},
			absent: []string{"<script>", "<i>x"},
		},
		{name: "votes without a rating are left out", metadata: media.Metadata{Title: "Dune", IMDbVotes: "10"}, plotMax: 200, want: []string{"<b>Dune</b>"}, absent: []string{"votes"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p := newPresenter(t, nil, tt.plotMax)
			checkCaption(t, mustPresent(t, p, &tt.metadata, tt.release).Caption, tt.want, tt.absent)
		})
	}
}

type plotCase struct {
	name   string
	plot   string
	limit  int
	want   string
	hidden bool
}

func TestCardPresenterPlot(t *testing.T) {
	for _, tt := range []plotCase{
		{name: "short plot kept", plot: "Short plot.", limit: 200, want: "Short plot.</blockquote>"},
		{name: "cut at the last space", plot: "one two three four", limit: 10, want: "one two…</blockquote>"},
		{name: "hard cut without a space", plot: "abcdefghijkl", limit: 5, want: "abcde…</blockquote>"},
		{name: "cut text is escaped", plot: "x&y zz", limit: 4, want: "x&amp;y…</blockquote>"},
		{name: "leading space is no cut point", plot: " abcdefgh", limit: 4, want: "> abc…</blockquote>"},
		{name: "hidden at zero", plot: "Short plot.", limit: 0, hidden: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p := newPresenter(t, nil, tt.limit)
			caption := mustPresent(t, p, &media.Metadata{Title: "Dune", Plot: tt.plot}, media.Release{}).Caption
			if tt.hidden {
				checkCaption(t, caption, nil, []string{"blockquote"})
				return
			}
			checkCaption(t, caption, []string{tt.want}, nil)
		})
	}
}

type trimCase struct {
	name   string
	edit   func(*media.Metadata)
	want   []string
	absent []string
}

func TestCardPresenterTrimming(t *testing.T) {
	for _, tt := range []trimCase{
		{name: "plot dropped first", edit: func(m *media.Metadata) { m.Actors = strings.Repeat("c", 700) }, want: []string{coyoteLink, "Cast: "}, absent: []string{"blockquote"}},
		{name: "cast dropped after the plot", edit: func(m *media.Metadata) { m.Actors = strings.Repeat("c", 1000) }, want: []string{coyoteLink, "Director: Dave Green"}, absent: []string{"Cast:", "blockquote"}},
		{name: "overlong genre drops every line after the ratings", edit: func(m *media.Metadata) { m.Genre = strings.Repeat("g", 2000) }, want: []string{coyoteLink, "IMDb: 7.6"}, absent: []string{"Director:", "Cast:", "blockquote", "ggg"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			m := fullMetadata()
			tt.edit(&m)
			p := newPresenter(t, nil, 200)
			checkCaption(t, mustPresent(t, p, &m, media.Release{}).Caption, tt.want, tt.absent)
		})
	}
}

func TestCardPresenterLongTitle(t *testing.T) {
	for _, tt := range []struct {
		name, title, format string
		linked, repeated    bool
	}{
		{name: "escaped title without link", title: strings.Repeat("🌊<&", 12000), format: "<b>{title}</b>"},
		{name: "repeated title override", title: strings.Repeat("é&<", 12000), format: "<b>{title}</b> / <i>{title}</i>", repeated: true},
		{name: "linked title keeps its anchor", title: strings.Repeat("🌊<&", 12000), linked: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var overrides map[string]any
			if tt.format != "" {
				overrides = cardOverride("title", tt.format)
			}
			p := newPresenter(t, overrides, 200)
			m := media.Metadata{Title: tt.title}
			if tt.linked {
				m.IMDbID = "tt1756855"
			}
			caption := mustPresent(t, p, &m, media.Release{}).Caption
			escaped := checkLongTitle(t, tt.title, caption, tt.linked)
			if tt.repeated && !strings.Contains(caption, "<i>"+escaped+"</i>") {
				t.Fatalf("repeated title differs: %q", caption)
			}
		})
	}
}

// checkLongTitle asserts that a caption with a long title fits, has the IMDb link exactly when linked, and keeps
// a valid escaped prefix of the title in its bold text. It returns that escaped prefix.
func checkLongTitle(t *testing.T, title, caption string, linked bool) string {
	t.Helper()
	if n := utf8.RuneCountInString(caption); n > telegramCaptionLimit {
		t.Fatalf("caption has %d runes, want at most %d", n, telegramCaptionLimit)
	}
	if hasLink := strings.Contains(caption, "imdb.com"); hasLink != linked {
		t.Fatalf("IMDb link present %v, want %v: %q", hasLink, linked, caption)
	}
	_, rest, ok := strings.Cut(caption, "<b>")
	if !ok {
		t.Fatalf("missing title opening tag: %q", caption)
	}
	escaped, _, ok := strings.Cut(rest, "</b>")
	if !ok {
		t.Fatalf("missing title closing tag: %q", caption)
	}
	raw := html.UnescapeString(escaped)
	if raw == "" || !strings.HasPrefix(title, raw) || html.EscapeString(raw) != escaped {
		t.Fatalf("invalid escaped title prefix: %q", escaped)
	}
	return escaped
}

func TestCardPresenterOverlongOverrides(t *testing.T) {
	long := strings.Repeat("x", 1200)
	for _, tt := range []struct {
		name, key, override string
		metadata            media.Metadata
		release             media.Release
	}{
		{name: "imdb_url override", key: "imdb_url", override: long + "{id}", metadata: media.Metadata{Title: "Dune", IMDbID: "tt123", IMDbRating: "7.3"}},
		{name: "title_linked override", key: "title_linked", override: long + `<a href="{url}"><b>{title}</b></a>`, metadata: media.Metadata{Title: "Dune", IMDbID: "tt123"}},
		{name: "title override", key: "title", override: "<b>" + long + "{title}</b>", metadata: media.Metadata{Title: "Dune"}},
		{name: "year override", key: "year", override: long + "{year}", metadata: media.Metadata{Title: "Dune", Year: "2021"}},
		{name: "episode override", key: "episode", override: long + "{episode}", metadata: media.Metadata{Title: "Dune"}, release: media.Release{Episode: "S01E01"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p := newPresenter(t, cardOverride(tt.key, tt.override), 200)
			caption := mustPresent(t, p, &tt.metadata, tt.release).Caption
			checkCaption(t, caption, []string{"Dune"}, []string{"tt123", "xxxx"})
			if strings.Contains(caption, "<b>") != strings.Contains(caption, "</b>") {
				t.Fatalf("unbalanced title: %q", caption)
			}
		})
	}
}
