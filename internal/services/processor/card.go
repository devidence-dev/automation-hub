package processor

import (
	"html"
	"strings"
	"unicode/utf8"

	"automation-hub/internal/messages"
	"automation-hub/internal/services/media"
)

const telegramCaptionLimit = 1024

// CardPresenter renders a metadata card using catalog texts.
type CardPresenter struct{ Catalog *messages.Catalog }

// card holds the rendered parts of one caption until the Telegram length limit is applied.
type card struct {
	title string   // rendered title line with year and episode
	lines []string // rating and detail lines in display order
	link  string   // rendered IMDb link, empty when absent or dropped
}

// field is one catalog entry for a value that is skipped when empty.
type field struct{ key, variable, value string }

func (p CardPresenter) Present(m *media.Metadata, release media.Release) (string, error) {
	title := m.Title
	if title == "" {
		title = release.Title
	}
	c, err := p.buildCard(m, release, title)
	if err != nil {
		return "", err
	}
	c.dropLines()
	if !c.fits() {
		c.link = ""
	}
	if !c.fits() {
		if err := p.fitTitle(&c, title); err != nil {
			return "", err
		}
	}
	return c.text(), nil
}

// buildCard renders every field in display order.
func (p CardPresenter) buildCard(m *media.Metadata, release media.Release, title string) (card, error) {
	var c card
	var err error
	if c.title, err = p.renderTitle(title, m, release); err != nil {
		return card{}, err
	}
	if m.IMDbID != "" {
		if c.link, err = p.render("link", map[string]string{"id": html.EscapeString(m.IMDbID)}); err != nil {
			return card{}, err
		}
	}
	separator, err := p.render("separator", nil)
	if err != nil {
		return card{}, err
	}
	if c.lines, err = p.renderLines(m, release, separator); err != nil {
		return card{}, err
	}
	return c, nil
}

// renderTitle renders the title with the year and episode appended when present.
func (p CardPresenter) renderTitle(title string, m *media.Metadata, release media.Release) (string, error) {
	titleLine, err := p.renderValue("title", "title", title)
	if err != nil {
		return "", err
	}
	suffixes, err := p.renderFields([]field{{"year", "year", m.Year}, {"episode", "episode", release.Episode}})
	if err != nil {
		return "", err
	}
	return titleLine + strings.Join(suffixes, ""), nil
}

// renderLines renders the rating, detail and tag lines.
func (p CardPresenter) renderLines(m *media.Metadata, release media.Release, separator string) ([]string, error) {
	lines, err := p.renderFields([]field{{"imdb_rating", "rating", m.IMDbRating}, {"rotten_tomatoes", "rating", m.RottenTomatoes}})
	if err != nil {
		return nil, err
	}
	detail, err := p.renderFields([]field{{"genre", "genre", m.Genre}, {"runtime", "runtime", m.Runtime}, {"rated", "rated", m.Rated}})
	if err != nil {
		return nil, err
	}
	if len(detail) > 0 {
		lines = append(lines, strings.Join(detail, separator))
	}
	tagSeparator, err := p.render("tags_separator", nil)
	if err != nil {
		return nil, err
	}
	tags, err := p.renderFields([]field{{"tags", "tags", strings.Join(release.Tags, tagSeparator)}})
	if err != nil {
		return nil, err
	}
	return append(lines, tags...), nil
}

// fitTitle shortens the title until the caption fits.
func (p CardPresenter) fitTitle(c *card, title string) error {
	// Search raw rune prefixes, then escape them whole so no HTML entity is split.
	runes := []rune(title)
	n, err := longestPrefix(runes, func(prefix string) (bool, error) {
		line, err := p.renderValue("title", "title", prefix)
		if err != nil {
			return false, err
		}
		return c.withTitle(line).fits(), nil
	})
	if err != nil {
		return err
	}
	c.title, err = p.renderValue("title", "title", string(runes[:n]))
	if err != nil {
		return err
	}
	if c.fits() {
		return nil
	}
	// If even the empty template is too long, keep the escaped title without
	// its oversized wrapper, cutting only between original runes.
	n, _ = longestPrefix(runes, func(prefix string) (bool, error) {
		return utf8.RuneCountInString(html.EscapeString(prefix)) <= telegramCaptionLimit, nil
	})
	c.title = html.EscapeString(string(runes[:n]))
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

// text joins the parts into the caption, one part per line.
func (c card) text() string {
	parts := []string{c.title}
	parts = append(parts, c.lines...)
	if c.link != "" {
		parts = append(parts, c.link)
	}
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
