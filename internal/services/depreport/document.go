package depreport

import (
	"bytes"
	_ "embed" // dependency_report.html
	"fmt"
	"html/template"
	"strconv"
	"strings"
	"time"

	"automation-hub/internal/messages"
)

//go:embed assets/dependency_report.html
var documentHTML string

var documentTemplate = template.Must(template.New("dependency_report").Parse(documentHTML))

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

// Present returns the HTML document of the summary.
func (p *DocumentPresenter) Present(s Summary) ([]byte, error) {
	r := &renderer{cat: p.cat}
	view := buildDocumentView(r, p.loc, s)
	if r.err != nil {
		return nil, r.err
	}
	var buf bytes.Buffer
	if err := documentTemplate.Execute(&buf, view); err != nil {
		return nil, fmt.Errorf("render dependency report: %w", err)
	}
	return buf.Bytes(), nil
}

// documentView is the data the template renders.
type documentView struct {
	Lang          string
	Title         string
	Brand         string
	ScanTimestamp string
	Labels        labelSet
	Tabs          []TabView
	Errors        []string
}

// labelSet holds the texts of the document that are not per tab.
type labelSet struct {
	Total         string
	Pending       string
	UpToDate      string
	Distribution  string
	Available     string
	ErrorsHeading string
	Copied        string
	CopiedBefore  string
	CopyCurrent   string
	CopyLatest    string
}

// TabView is one tab of the document: its label, totals, status, columns and rows.
type TabView struct {
	ID           string
	Label        string
	Total        int
	Updates      int
	UpToDate     int
	Distribution []distItem
	StatusBold   string // bold part of the status line
	StatusRest   string // the rest of the status line, after the bold part
	Columns      []columnView
	Rows         []rowView
}

// distItem counts the rows of one update type.
type distItem struct {
	Label string
	Count int
	Class string
}

// columnView is one header of the table.
type columnView struct {
	Header string
	Width  string
}

// rowView is one table row: one cell per column.
type rowView struct {
	Cells []cell
}

// cell is one table cell. Kind tells the template how to show it.
type cell struct {
	Kind  string // file, repo, copy, badge or code
	Text  string
	Class string // badge class
	Title string // copy button title
}

// docBuilder builds the view of the document. It reads the catalog through r.
type docBuilder struct {
	r       *renderer
	forkTag string
	labels  labelSet
}

// columnSpec defines one table column: its header key, its width and how its cell is built.
type columnSpec struct {
	key   string
	width string
	value func(b *docBuilder, row Row) cell
}

// imageColumns is the table of the images tab.
var imageColumns = []columnSpec{
	{key: KeyColPackageFile, width: "15%", value: fileCell},
	{key: KeyColRepository, width: "15%", value: repoCell},
	{key: KeyColCurrentImage, width: "30%", value: currentImageCell},
	{key: KeyColLatestImage, width: "30%", value: latestImageCell},
	{key: KeyColType, width: "10%", value: typeCell},
}

// depColumns is the table of the dependencies tab.
var depColumns = []columnSpec{
	{key: KeyColPackageFile, width: "15%", value: fileCell},
	{key: KeyColRepository, width: "15%", value: repoCell},
	{key: KeyColDependency, width: "20%", value: nameCell},
	{key: KeyColCurrent, width: "15%", value: currentCell},
	{key: KeyColNew, width: "15%", value: newCell},
	{key: KeyColType, width: "10%", value: typeCell},
}

func buildDocumentView(r *renderer, loc *time.Location, s Summary) documentView {
	b := &docBuilder{r: r, forkTag: r.text(KeyForkTagDocument, nil), labels: newLabels(r)}
	return documentView{
		Lang:          r.text(KeyLang, nil),
		Title:         r.text(KeyDocTitle, nil),
		Brand:         r.text(KeyDocBrand, nil),
		ScanTimestamp: s.ReceivedAt.In(loc).Format(scanLayout) + " " + loc.String(),
		Labels:        b.labels,
		Tabs: []TabView{
			b.tab("images", KeyDocTabImages, s.Images, imageColumns),
			b.tab("deps", KeyDocTabDeps, s.Deps, depColumns),
		},
		Errors: b.errorLines(s.Errors),
	}
}

func newLabels(r *renderer) labelSet {
	return labelSet{
		Total:         r.text(KeyDocTotal, nil),
		Pending:       r.text(KeyDocPending, nil),
		UpToDate:      r.text(KeyDocUpToDate, nil),
		Distribution:  r.text(KeyDocDistribution, nil),
		Available:     r.text(KeyDocAvailable, nil),
		ErrorsHeading: r.text(KeyErrorsHeading, nil),
		Copied:        r.text(KeyCopyCopied, nil),
		CopiedBefore:  r.text(KeyCopyCopiedBefore, nil),
		CopyCurrent:   r.text(KeyCopyCurrentImage, nil),
		CopyLatest:    r.text(KeyCopyLatestImage, nil),
	}
}

// tab builds one tab from its totals and its rows, which are already sorted.
func (b *docBuilder) tab(id, labelKey string, tab Tab, columns []columnSpec) TabView {
	bold, rest := b.status(tab)
	view := TabView{
		ID:           id,
		Label:        b.r.text(labelKey, nil),
		Total:        tab.Total,
		Updates:      tab.Updates,
		UpToDate:     tab.UpToDate,
		Distribution: distribution(tab),
		StatusBold:   bold,
		StatusRest:   rest,
	}
	for _, c := range columns {
		view.Columns = append(view.Columns, columnView{Header: b.r.text(c.key, nil), Width: c.width})
	}
	for _, row := range tab.Rows {
		cells := make([]cell, 0, len(columns))
		for _, c := range columns {
			cells = append(cells, c.value(b, row))
		}
		view.Rows = append(view.Rows, rowView{Cells: cells})
	}
	return view
}

// status is the line under the tab's title, as a bold part and a rest. With pending updates the
// bold part holds the count and the rest names the dependencies; when everything is up to date the
// whole line is the bold part.
func (b *docBuilder) status(tab Tab) (bold, rest string) {
	if tab.Updates == 0 {
		return b.r.text(KeyDocStatusUpToDate, nil), ""
	}
	vars := map[string]string{
		"updates": strconv.Itoa(tab.Updates),
		"deps":    strconv.Itoa(tab.WithUpdates),
	}
	return b.r.text(KeyDocStatusUpdatesBold, vars), b.r.text(KeyDocStatusUpdatesRest, vars)
}

// errorLines lists the scan errors. The report root is named by its own catalog text.
func (b *docBuilder) errorLines(errs []ScanError) []string {
	root := b.r.text(KeyErrorsRoot, nil)
	lines := make([]string, 0, len(errs))
	for _, e := range errs {
		where := e.Repo
		if where == "" {
			where = root
		}
		lines = append(lines, b.r.text(KeyErrorsItem, map[string]string{"where": where, "msg": e.Msg}))
	}
	return lines
}

// distribution counts the rows of each update type, patch first.
func distribution(tab Tab) []distItem {
	var items []distItem
	for _, typ := range []UpdateType{UpdatePatch, UpdateMinor, UpdateMajor} {
		items = append(items, distItem{Label: typ.String(), Count: tab.Counts[typ], Class: "dist-" + typ.String()})
	}
	return items
}

func fileCell(_ *docBuilder, row Row) cell {
	return cell{Kind: "file", Text: row.PackageFile}
}

func repoCell(b *docBuilder, row Row) cell {
	text := row.Repo
	if row.Fork {
		text += b.forkTag
	}
	return cell{Kind: "repo", Text: text}
}

func currentImageCell(b *docBuilder, row Row) cell {
	return cell{Kind: "copy", Text: row.Name + ":" + row.Version, Title: b.labels.CopyCurrent}
}

func latestImageCell(b *docBuilder, row Row) cell {
	return cell{Kind: "copy", Text: row.Name + ":" + row.NewVersion, Title: b.labels.CopyLatest}
}

func nameCell(_ *docBuilder, row Row) cell {
	return cell{Kind: "code", Text: row.Name}
}

func currentCell(_ *docBuilder, row Row) cell {
	return cell{Kind: "code", Text: row.Version}
}

func newCell(_ *docBuilder, row Row) cell {
	return cell{Kind: "code", Text: row.NewVersion}
}

func typeCell(_ *docBuilder, row Row) cell {
	return cell{Kind: "badge", Text: row.Type.String(), Class: "badge-" + row.Type.String()}
}
