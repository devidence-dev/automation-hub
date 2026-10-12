package depreport

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"automation-hub/internal/messages"
)

// scanLayout is the header time of the document. It is a Go layout, not a catalog text.
const scanLayout = "2006-01-02 15:04"

// DocumentName is the file name of a report received at scannedAt: the pattern with {date}
// replaced by that day in loc, formatted with layout.
func DocumentName(scannedAt time.Time, loc *time.Location, pattern, layout string) string {
	return strings.ReplaceAll(pattern, "{date}", scannedAt.In(loc).Format(layout))
}

// DocumentPresenter renders the HTML report from the catalog and the summary.
type DocumentPresenter struct {
	cat *messages.Catalog
	loc *time.Location
}

// NewDocumentPresenter returns a presenter that shows the scan time in loc.
func NewDocumentPresenter(cat *messages.Catalog, loc *time.Location) *DocumentPresenter {
	return &DocumentPresenter{cat: cat, loc: loc}
}

// Present returns the HTML document of the summary: the view built from the catalog, rendered by the
// ReportDocument component.
func (p *DocumentPresenter) Present(s Summary) ([]byte, error) {
	r := &renderer{cat: p.cat}
	view := buildDocumentView(r, p.loc, s)
	if r.err != nil {
		return nil, r.err
	}
	var buf bytes.Buffer
	if err := ReportDocument(view).Render(context.Background(), &buf); err != nil {
		return nil, fmt.Errorf("render dependency report: %w", err)
	}
	return buf.Bytes(), nil
}

// documentView is the data the document components render.
type documentView struct {
	Lang          string
	Title         string
	ScanTimestamp string
	Labels        labelSet
	Tabs          []TabView
	Errors        []ErrorView
	ErrorsChip    string // the count of the errors card, empty when there are none
}

// labelSet holds the texts of the document that are not per tab.
type labelSet struct {
	ErrorsHeading  string
	CopyLatest     string
	CopyNewVersion string
	Copied         string
	CopiedBefore   string
	Current        string
	New            string
	ThemeGroup     string
	ThemeSystem    string
	ThemeLight     string
	ThemeDark      string
	TypePatch      string
	TypeMinor      string
	TypeMajor      string
}

// TabView is one tab of the document: its label and count badge, its panel and its repo cards.
type TabView struct {
	ID       string
	Label    string
	Count    int // listed updates, the badge of the tab
	Selected bool
	Stats    StatsView
	Summary  SummaryLine
	Groups   []RepoGroupView
}

// SummaryLine is the line under the panel: a bold part and the rest of it. The rest is empty when
// the whole line is the bold part.
type SummaryLine struct {
	Bold string
	Rest string
}

// ErrorView is one scan error of the errors card: the repo or the report root, and the message.
type ErrorView struct {
	Where string
	Msg   string
}

// tabSpec is what the builder needs to know about one tab before it reads its totals.
type tabSpec struct {
	ID         string
	LabelKey   string
	EyebrowKey string
	Images     bool
	Selected   bool
}

// docBuilder builds the view of the document. It reads the catalog through r.
type docBuilder struct {
	r      *renderer
	labels labelSet
}

// buildDocumentView builds the view. The dependencies tab comes first and is selected; the images
// tab is second, as the liquid glass design requires.
func buildDocumentView(r *renderer, loc *time.Location, s Summary) documentView {
	b := &docBuilder{r: r, labels: newLabels(r)}
	return documentView{
		Lang:          r.text(KeyLang, nil),
		Title:         r.text(KeyDocTitle, nil),
		ScanTimestamp: s.ReceivedAt.In(loc).Format(scanLayout) + " " + loc.String(),
		Labels:        b.labels,
		Tabs: []TabView{
			b.tab(tabSpec{ID: "deps", LabelKey: KeyDocTabDeps, EyebrowKey: KeyDocEyebrowDeps, Selected: true}, s.Deps),
			b.tab(tabSpec{ID: "images", LabelKey: KeyDocTabImages, EyebrowKey: KeyDocEyebrowImages, Images: true}, s.Images),
		},
		Errors:     b.errorRows(s.Errors),
		ErrorsChip: b.errorsChip(len(s.Errors)),
	}
}

func newLabels(r *renderer) labelSet {
	return labelSet{
		ErrorsHeading:  r.text(KeyErrorsHeading, nil),
		CopyLatest:     r.text(KeyCopyLatestImage, nil),
		CopyNewVersion: r.text(KeyCopyNewVersion, nil),
		Copied:         r.text(KeyCopyCopied, nil),
		CopiedBefore:   r.text(KeyCopyCopiedBefore, nil),
		Current:        r.text(KeyColCurrent, nil),
		New:            r.text(KeyColNew, nil),
		ThemeGroup:     r.text(KeyDocThemeGroup, nil),
		ThemeSystem:    r.text(KeyDocThemeSystem, nil),
		ThemeLight:     r.text(KeyDocThemeLight, nil),
		ThemeDark:      r.text(KeyDocThemeDark, nil),
		TypePatch:      r.text(KeyDocTypePatch, nil),
		TypeMinor:      r.text(KeyDocTypeMinor, nil),
		TypeMajor:      r.text(KeyDocTypeMajor, nil),
	}
}

// typeName is the catalog name of an update type, as the legend and the row chips show it.
func (l labelSet) typeName(t UpdateType) string {
	switch t {
	case UpdateMajor:
		return l.TypeMajor
	case UpdateMinor:
		return l.TypeMinor
	case UpdatePatch:
		return l.TypePatch
	default:
		return ""
	}
}

// tab builds one tab from its spec and its totals. The rows are already sorted.
func (b *docBuilder) tab(spec tabSpec, tab Tab) TabView {
	return TabView{
		ID:       spec.ID,
		Label:    b.r.text(spec.LabelKey, nil),
		Count:    tab.Updates,
		Selected: spec.Selected,
		Stats:    b.stats(spec, tab),
		Summary:  b.summary(spec, tab),
		Groups:   groupTab(b.r, tab, b.labels, spec.Images),
	}
}

// stats builds the panel of a tab: its eyebrow, the listed updates, the sub line, the bar and the legend.
func (b *docBuilder) stats(spec tabSpec, tab Tab) StatsView {
	return StatsView{
		Eyebrow:  b.r.text(spec.EyebrowKey, nil),
		Big:      strconv.Itoa(tab.Updates),
		Sub:      b.r.text(KeyDocPanelSub, map[string]string{"total": strconv.Itoa(tab.Total), "up_to_date": strconv.Itoa(tab.UpToDate)}),
		Segments: segmentsOf(tab),
		Legend:   legendOf(tab, b.labels),
	}
}

// summary is the line under the panel. With listed updates the bold part holds the count of updates
// and the rest the count of dependencies (or images); when everything is up to date the bold part is
// the up-to-date text of the tab.
func (b *docBuilder) summary(spec tabSpec, tab Tab) SummaryLine {
	if tab.Updates == 0 {
		key := KeyDocStatusUpToDate
		if spec.Images {
			key = KeyDocStatusImagesUpTo
		}
		return SummaryLine{Bold: b.r.text(key, nil)}
	}
	vars := map[string]string{"updates": strconv.Itoa(tab.Updates), "deps": strconv.Itoa(tab.WithUpdates)}
	bold := pick(tab.Updates == 1, KeyDocStatusUpdatesBold1, KeyDocStatusUpdatesBold)
	rest := pick(tab.WithUpdates == 1, KeyDocStatusUpdatesRest1, KeyDocStatusUpdatesRest)
	if spec.Images {
		rest = pick(tab.WithUpdates == 1, KeyDocStatusImagesRest1, KeyDocStatusImagesRest)
	}
	return SummaryLine{Bold: b.r.text(bold, vars), Rest: b.r.text(rest, vars)}
}

// pick returns one of two catalog keys: one when the count is 1, the other otherwise.
func pick(one bool, oneKey, manyKey string) string {
	if one {
		return oneKey
	}
	return manyKey
}

// errorRows lists the scan errors. The report root is named by its own catalog text.
func (b *docBuilder) errorRows(errs []ScanError) []ErrorView {
	root := b.r.text(KeyErrorsRoot, nil)
	rows := make([]ErrorView, 0, len(errs))
	for _, e := range errs {
		where := e.Repo
		if where == "" {
			where = root
		}
		rows = append(rows, ErrorView{Where: where, Msg: e.Msg})
	}
	return rows
}

// errorsChip is the count of the errors card, or empty when there are no errors.
func (b *docBuilder) errorsChip(n int) string {
	if n == 0 {
		return ""
	}
	key := pick(n == 1, KeyErrorsCountOne, KeyErrorsCount)
	return b.r.text(key, map[string]string{"count": strconv.Itoa(n)})
}
