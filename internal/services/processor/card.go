package processor

import (
	"html"
	"strings"
	"unicode"
	"unicode/utf8"

	"automation-hub/internal/messages"
	"automation-hub/internal/services/media"
)

const telegramCaptionLimit = 1024

// CardPresenter renders a metadata card using catalog texts.
type CardPresenter struct {
	Catalog *messages.Catalog
	// PlotMaxChars is the plot length limit, 0 hides the plot.
	PlotMaxChars int
}

// Card is a rendered metadata card: the HTML caption and the IMDb page URL, empty without an IMDb id.
type Card struct{ Caption, URL string }

// card holds the rendered parts of one caption until the Telegram length limit is applied.
type card struct {
	title  string   // rendered title line, linked to IMDb when there is an id
	suffix string   // year and episode, appended to the title
	lines  []string // rating, detail, tag, director, cast and plot lines in display order
}

// field is one catalog entry for a value that is skipped when empty.
type field struct{ key, variable, value string }

// Present renders the caption of m and release and the IMDb URL of the card.
func (p CardPresenter) Present(m *media.Metadata, release media.Release) (Card, error) {
	title := m.Title
	if title == "" {
		title = release.Title
	}
	imdbURL, err := p.imdbURL(m)
	if err != nil {
		return Card{}, err
	}
	c, err := p.buildCard(m, release, title, imdbURL)
	if err != nil {
		return Card{}, err
	}
	c.dropLines()
	if !c.fits() {
		if err := p.fitTitle(&c, title, imdbURL); err != nil {
			return Card{}, err
		}
	}
	return Card{Caption: c.text(), URL: imdbURL}, nil
}

// imdbURL renders the IMDb page of m, empty without an IMDb id.
func (p CardPresenter) imdbURL(m *media.Metadata) (string, error) {
	if m.IMDbID == "" {
		return "", nil
	}
	return p.render("imdb_url", map[string]string{"id": html.EscapeString(m.IMDbID)})
}

// buildCard renders every field in display order.
func (p CardPresenter) buildCard(m *media.Metadata, release media.Release, title, imdbURL string) (card, error) {
	var c card
	var err error
	if c.title, err = p.renderTitleLine(title, imdbURL); err != nil {
		return card{}, err
	}
	if c.suffix, err = p.renderSuffix(m, release); err != nil {
		return card{}, err
	}
	if c.lines, err = p.renderLines(m, release); err != nil {
		return card{}, err
	}
	return c, nil
}

// renderTitleLine renders the title, linked to imdbURL when there is one.
func (p CardPresenter) renderTitleLine(title, imdbURL string) (string, error) {
	if imdbURL == "" {
		return p.renderValue("title", "title", title)
	}
	return p.render("title_linked", map[string]string{"url": html.EscapeString(imdbURL), "title": html.EscapeString(title)})
}

// renderSuffix renders the year and the episode that follow the title.
func (p CardPresenter) renderSuffix(m *media.Metadata, release media.Release) (string, error) {
	parts, err := p.renderFields([]field{{"year", "year", m.Year}, {"episode", "episode", release.Episode}})
	if err != nil {
		return "", err
	}
	return strings.Join(parts, ""), nil
}

// renderLines renders the lines after the title in display order. Empty lines are left out.
func (p CardPresenter) renderLines(m *media.Metadata, release media.Release) ([]string, error) {
	separator, err := p.render("separator", nil)
	if err != nil {
		return nil, err
	}
	tagSeparator, err := p.render("tags_separator", nil)
	if err != nil {
		return nil, err
	}
	ratings, err := p.renderRatings(m, separator)
	if err != nil {
		return nil, err
	}
	details, err := p.renderJoined([]field{{"genre", "genre", m.Genre}, {"runtime", "runtime", m.Runtime}, {"rated", "rated", m.Rated}}, separator)
	if err != nil {
		return nil, err
	}
	tags, err := p.renderFields([]field{{"tags", "tags", strings.Join(release.Tags, tagSeparator)}})
	if err != nil {
		return nil, err
	}
	people, err := p.renderFields([]field{{"director", "director", m.Director}, {"cast", "cast", m.Actors}})
	if err != nil {
		return nil, err
	}
	plot, err := p.renderPlot(m.Plot)
	if err != nil {
		return nil, err
	}
	lines := appendNonEmpty(nil, ratings, details)
	lines = append(lines, tags...)
	lines = append(lines, people...)
	return appendNonEmpty(lines, plot), nil
}

// renderRatings renders the IMDb, Rotten Tomatoes and Metacritic ratings joined on one line.
func (p CardPresenter) renderRatings(m *media.Metadata, separator string) (string, error) {
	imdb, err := p.renderIMDbRating(m)
	if err != nil {
		return "", err
	}
	others, err := p.renderFields([]field{{"rotten_tomatoes", "rating", m.RottenTomatoes}, {"metacritic", "rating", m.Metacritic}})
	if err != nil {
		return "", err
	}
	return strings.Join(append(appendNonEmpty(nil, imdb), others...), separator), nil
}

// renderIMDbRating renders the IMDb rating followed by its votes. Without a rating the votes are left out too.
func (p CardPresenter) renderIMDbRating(m *media.Metadata) (string, error) {
	if m.IMDbRating == "" {
		return "", nil
	}
	rating, err := p.renderValue("imdb_rating", "rating", m.IMDbRating)
	if err != nil {
		return "", err
	}
	if m.IMDbVotes == "" {
		return rating, nil
	}
	votes, err := p.renderValue("imdb_votes", "votes", m.IMDbVotes)
	if err != nil {
		return "", err
	}
	return rating + votes, nil
}

// renderPlot renders the plot cut to PlotMaxChars inside the plot block. It returns "" when the plot is hidden or empty.
func (p CardPresenter) renderPlot(plot string) (string, error) {
	if p.PlotMaxChars <= 0 || plot == "" {
		return "", nil
	}
	text, cut := cutPlot(plot, p.PlotMaxChars)
	body := html.EscapeString(text)
	if cut {
		var err error
		if body, err = p.render("plot_truncated", map[string]string{"text": body}); err != nil {
			return "", err
		}
	}
	return p.render("plot", map[string]string{"plot": body})
}

// cutPlot returns the first limit runes of plot, cut at the last space inside the limit when there is one,
// and whether the plot was cut.
func cutPlot(plot string, limit int) (string, bool) {
	runes := []rune(plot)
	if len(runes) <= limit {
		return plot, false
	}
	cut := limit
	if i := lastSpace(runes[:limit]); i > 0 {
		cut = i
	}
	return string(runes[:cut]), true
}

// lastSpace returns the index of the last space after the first rune, or 0 when there is none.
func lastSpace(runes []rune) int {
	for i := len(runes) - 1; i > 0; i-- {
		if unicode.IsSpace(runes[i]) {
			return i
		}
	}
	return 0
}

// fitTitle shortens the title until the caption fits. The shortened title keeps its link and suffix.
func (p CardPresenter) fitTitle(c *card, title, imdbURL string) error {
	// Search raw rune prefixes, then escape them whole so no HTML entity is split.
	runes := []rune(title)
	n, err := longestPrefix(runes, func(prefix string) (bool, error) {
		line, err := p.renderTitleLine(prefix, imdbURL)
		if err != nil {
			return false, err
		}
		return c.withTitle(line).fits(), nil
	})
	if err != nil {
		return err
	}
	c.title, err = p.renderTitleLine(string(runes[:n]), imdbURL)
	if err != nil {
		return err
	}
	if c.fits() {
		return nil
	}
	// If even the empty template is too long, keep the escaped title without
	// its oversized wrapper and its suffix, cutting only between original runes.
	n, _ = longestPrefix(runes, func(prefix string) (bool, error) {
		return utf8.RuneCountInString(html.EscapeString(prefix)) <= telegramCaptionLimit, nil
	})
	c.title = html.EscapeString(string(runes[:n]))
	c.suffix = ""
	return nil
}

// render renders one catalog entry of the card.
func (p CardPresenter) render(key string, vars map[string]string) (string, error) {
	return p.Catalog.Render("torrent.card."+key, vars)
}

// renderValue renders one catalog entry with its value escaped under the variable name.
func (p CardPresenter) renderValue(key, variable, value string) (string, error) {
	return p.render(key, map[string]string{variable: html.EscapeString(value)})
}

// renderFields renders the non-empty fields in order.
func (p CardPresenter) renderFields(fields []field) ([]string, error) {
	var out []string
	for _, f := range fields {
		if f.value == "" {
			continue
		}
		s, err := p.renderValue(f.key, f.variable, f.value)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}

// renderJoined renders the non-empty fields and joins them with separator.
func (p CardPresenter) renderJoined(fields []field, separator string) (string, error) {
	parts, err := p.renderFields(fields)
	if err != nil {
		return "", err
	}
	return strings.Join(parts, separator), nil
}

// appendNonEmpty appends the non-empty parts to lines in order.
func appendNonEmpty(lines []string, parts ...string) []string {
	for _, s := range parts {
		if s != "" {
			lines = append(lines, s)
		}
	}
	return lines
}

// text joins the parts into the caption, one part per line.
func (c card) text() string {
	parts := []string{c.title + c.suffix}
	parts = append(parts, c.lines...)
	return strings.Join(parts, "\n")
}

// fits reports whether the caption is within the Telegram caption limit.
func (c card) fits() bool {
	return utf8.RuneCountInString(c.text()) <= telegramCaptionLimit
}

// withTitle returns a copy of the card with another title line.
func (c card) withTitle(title string) card {
	c.title = title
	return c
}

// dropLines removes trailing lines until the caption fits.
func (c *card) dropLines() {
	for len(c.lines) > 0 && !c.fits() {
		c.lines = c.lines[:len(c.lines)-1]
	}
}

// longestPrefix returns the largest n for which fits(string(runes[:n])) holds, assuming
// that a prefix which fits keeps every shorter prefix fitting.
func longestPrefix(runes []rune, fits func(prefix string) (bool, error)) (int, error) {
	low, high := 0, len(runes)
	for low < high {
		mid := low + (high-low+1)/2
		ok, err := fits(string(runes[:mid]))
		if err != nil {
			return 0, err
		}
		if ok {
			low = mid
		} else {
			high = mid - 1
		}
	}
	return low, nil
}
