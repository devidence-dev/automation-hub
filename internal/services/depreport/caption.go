package depreport

import (
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"automation-hub/internal/messages"
)

// captionLimit is Telegram's caption limit. The check counts the runes of the escaped caption,
// tags included, which over-counts and so never exceeds it.
const captionLimit = 1024

// CaptionPresenter renders the summary caption sent with the document, as Telegram HTML.
type CaptionPresenter struct {
	cat        *messages.Catalog
	loc        *time.Location
	dateLayout string
}

// NewCaptionPresenter returns a presenter that dates the caption in loc with dateLayout.
func NewCaptionPresenter(cat *messages.Catalog, loc *time.Location, dateLayout string) *CaptionPresenter {
	return &CaptionPresenter{cat: cat, loc: loc, dateLayout: dateLayout}
}

// Present returns the caption of the summary.
func (p *CaptionPresenter) Present(s Summary) ([]byte, error) {
	r := &renderer{cat: p.cat}
	header := r.text(KeyCaptionHeader, map[string]string{"date": s.ReceivedAt.In(p.loc).Format(p.dateLayout)})

	var repoLines, after []string
	if len(s.Pending) == 0 {
		after = []string{r.text(KeyCaptionAllUpToDate, nil)}
	} else {
		repoLines = p.repoLines(r, s.Pending)
		if upToDate := s.Repos - len(s.Pending); upToDate > 0 {
			after = []string{r.text(KeyCaptionReposUpToDate, map[string]string{"count": strconv.Itoa(upToDate)})}
		}
	}
	caption := p.fit(r, header, repoLines, after, p.stats(r, s))
	if r.err != nil {
		return nil, r.err
	}
	return []byte(caption), nil
}

// repoLines formats one line per repo: emoji, short name (escaped, with the fork tag) and counts.
func (p *CaptionPresenter) repoLines(r *renderer, pending []RepoPending) []string {
	lines := make([]string, 0, len(pending))
	for _, rp := range pending {
		emoji := r.text(KeyCaptionEmojiOther, nil)
		if rp.Major > 0 {
			emoji = r.text(KeyCaptionEmojiMajor, nil)
		}
		name := htmlEscaper.Replace(ShortName(rp.Repo))
		if rp.Fork {
			name += r.text(KeyForkTagCaption, nil)
		}
		lines = append(lines, r.text(KeyCaptionRepoLine, map[string]string{
			"emoji":  emoji,
			"name":   name,
			"counts": countsText(r, rp),
		}))
	}
	return lines
}

// countsText lists the non-zero counts of a repo in the order major, minor, patch.
func countsText(r *renderer, rp RepoPending) string {
	counts := []struct {
		key string
		n   int
	}{{KeyCountMajor, rp.Major}, {KeyCountMinor, rp.Minor}, {KeyCountPatch, rp.Patch}}
	var parts []string
	for _, c := range counts {
		if c.n > 0 {
			parts = append(parts, r.text(c.key, map[string]string{"count": strconv.Itoa(c.n)}))
		}
	}
	return strings.Join(parts, r.text(KeyCountSeparator, nil))
}

// stats are the totals lines, from the tab totals of the summary.
func (p *CaptionPresenter) stats(r *renderer, s Summary) []string {
	updated := s.Images.WithUpdates + s.Deps.WithUpdates
	current := s.Images.UpToDate + s.Deps.UpToDate
	lines := []string{
		"",
		r.text(KeyCaptionStats, map[string]string{"updated": strconv.Itoa(updated), "current": strconv.Itoa(current)}),
		r.text(KeyCaptionTabs, map[string]string{
			"images":   strconv.Itoa(s.Images.WithUpdates),
			"packages": strconv.Itoa(s.Deps.WithUpdates),
		}),
	}
	if n := len(s.Errors); n > 0 {
		lines = append(lines, r.text(KeyCaptionErrors, map[string]string{"count": strconv.Itoa(n)}))
	}
	return lines
}

// fit joins the caption lines. While the caption is over captionLimit it drops repo lines from
// the end and puts a "… y N repos más" line after the last kept one. The header and the footer
// always stay, so a caption with no repo lines may still be over the limit.
func (p *CaptionPresenter) fit(r *renderer, header string, repoLines, after, footer []string) string {
	for shown := len(repoLines); ; shown-- {
		lines := append([]string{header, ""}, repoLines[:shown]...)
		if shown < len(repoLines) {
			lines = append(lines, r.text(KeyCaptionMoreRepos, map[string]string{"count": strconv.Itoa(len(repoLines) - shown)}))
		}
		lines = append(append(lines, after...), footer...)
		caption := strings.Join(lines, "\n")
		if utf8.RuneCountInString(caption) <= captionLimit || shown == 0 {
			return caption
		}
	}
}
