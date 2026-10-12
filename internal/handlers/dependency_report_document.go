package handlers

import (
	"bytes"
	_ "embed" // dependency_report.html
	"fmt"
	"html/template"
	"slices"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

//go:embed assets/dependency_report.html
var dependencyReportHTML string

var documentTemplate = template.Must(template.New("dependency_report").Parse(dependencyReportHTML))

// reportTimeZone is the zone of the document header and file name.
const reportTimeZone = "America/Guayaquil"

// dockerDatasource is the Renovate datasource of container images. Those deps
// go to the images tab; every other dep, including one without a datasource,
// goes to the dependencies tab.
const dockerDatasource = "docker"

// captionLimit is Telegram's caption limit. The check counts the runes of the
// escaped caption, tags included, which over-counts and so never exceeds it.
const captionLimit = 1024

// documentData is the content of the HTML report and of its caption, split into its two tabs.
type documentData struct {
	Images  reportTab
	Deps    reportTab
	Repos   int           // repos in the report, with or without pending updates
	Pending []repoPending // repos with a listed update, in caption order
	Errors  []string      // problems with level >= 50: "<repo>: <msg>" or "root: <msg>"
}

// reportTab holds the rows of one tab and its totals.
type reportTab struct {
	Rows         []docRow
	Updates      int // listed rows, one per dep and update
	Total        int // distinct deps in the tab, with or without listed updates
	WithUpdates  int // distinct deps with at least one listed row
	UpToDate     int // Total minus WithUpdates
	Distribution []distItem
}

// docRow is one listed update: one dep and one update of type major, minor or patch.
type docRow struct {
	Repo         string // repo name, with the fork tag when it is a fork
	PackageFile  string
	Name         string // depName, "?" when Renovate gives none
	Version      string // current version
	NewVersion   string
	CurrentImage string // images tab only: Name:Version
	LatestImage  string // images tab only: Name:NewVersion
	UpdateType   string
	BadgeClass   string
}

// distItem counts the rows of one update type.
type distItem struct {
	Type  string
	Count int
	Class string
}

// depKey identifies a dep in the report: the same dep may have several updates.
// repo is the report name, without the fork tag.
type depKey struct{ repo, file, name string }

// tabBuilder collects rows and dep identities for one tab.
type tabBuilder struct {
	deps     map[depKey]bool
	updated  map[depKey]bool
	highest  map[depKey]string // highest update type of each dep with a listed update
	seenRows map[docRow]bool
	rows     []docRow
}

func newTabBuilder() *tabBuilder {
	return &tabBuilder{
		deps:     map[depKey]bool{},
		updated:  map[depKey]bool{},
		highest:  map[depKey]string{},
		seenRows: map[docRow]bool{},
	}
}

func (b *tabBuilder) addRow(key depKey, row docRow) {
	b.updated[key] = true
	if cur, ok := b.highest[key]; !ok || updateRank(row.UpdateType) < updateRank(cur) {
		b.highest[key] = row.UpdateType
	}
	if !b.seenRows[row] {
		b.seenRows[row] = true
		b.rows = append(b.rows, row)
	}
}

// finish sorts the rows (repo, package file, dep, then major before minor before patch)
// and computes the totals.
func (b *tabBuilder) finish() reportTab {
	sort.SliceStable(b.rows, func(i, j int) bool {
		a, c := b.rows[i], b.rows[j]
		if a.Repo != c.Repo {
			return a.Repo < c.Repo
		}
		if a.PackageFile != c.PackageFile {
			return a.PackageFile < c.PackageFile
		}
		if a.Name != c.Name {
			return a.Name < c.Name
		}
		if a.UpdateType != c.UpdateType {
			return updateRank(a.UpdateType) < updateRank(c.UpdateType)
		}
		if a.Version != c.Version {
			return a.Version < c.Version
		}
		return a.NewVersion < c.NewVersion
	})

	counts := map[string]int{}
	for _, row := range b.rows {
		counts[row.UpdateType]++
	}
	tab := reportTab{
		Rows:        b.rows,
		Updates:     len(b.rows),
		Total:       len(b.deps),
		WithUpdates: len(b.updated),
	}
	tab.UpToDate = tab.Total - tab.WithUpdates
	for _, typ := range []string{"patch", "minor", "major"} {
		tab.Distribution = append(tab.Distribution, distItem{Type: typ, Count: counts[typ], Class: "dist-" + typ})
	}
	return tab
}

// updateRank orders update types for the table: major first, then minor, then patch.
func updateRank(updateType string) int {
	switch updateType {
	case "major":
		return 0
	case "minor":
		return 1
	default:
		return 2
	}
}

// repoPending counts the deps of one repo that have a listed update, each under its highest update type.
type repoPending struct {
	Repo  string // report name, for example devidence-dev/QuantWarden
	Fork  bool
	Major int
	Minor int
	Patch int
}

func (p repoPending) total() int { return p.Major + p.Minor + p.Patch }

// buildDocumentData splits the report's deps into the image and dependency tabs.
// Only major, minor and patch updates become rows; digest and pin are ignored.
// The caption and the document both read this data, so their counts agree.
func buildDocumentData(rep renovateReport, forks []string) documentData {
	images, deps := newTabBuilder(), newTabBuilder()

	for _, name := range sortedRepoNames(rep) {
		display := name
		if slices.Contains(forks, name) {
			display += forkTag
		}
		for _, files := range rep.Repositories[name].PackageFiles {
			for _, file := range files {
				addFileDeps(images, deps, name, display, file)
			}
		}
	}

	return documentData{
		Images:  images.finish(),
		Deps:    deps.finish(),
		Repos:   len(rep.Repositories),
		Pending: pendingByRepo(forks, images, deps),
		Errors:  scanErrors(rep),
	}
}

// addFileDeps adds the deps of one package file to the image tab or the dependency tab.
// repo is the report name, used as the key; display is the name shown in the rows.
func addFileDeps(images, deps *tabBuilder, repo, display string, file packageFile) {
	for i, dep := range file.Deps {
		tab := deps
		if dep.Datasource != nil && *dep.Datasource == dockerDatasource {
			tab = images
		}
		depName := firstNonNil("?", dep.DepName)
		// Unnamed deps (for example Dockerfile stage refs) are counted one by one.
		keyName := depName
		if dep.DepName == nil {
			keyName = fmt.Sprintf("?#%d", i)
		}
		key := depKey{repo: repo, file: file.PackageFile, name: keyName}
		tab.deps[key] = true
		addDepRows(tab, key, display, file.PackageFile, depName, dep)
	}
}

// addDepRows adds one row per major, minor or patch update of a dep.
func addDepRows(tab *tabBuilder, key depKey, display, pkgFile, depName string, dep reportDep) {
	version := firstNonNil("?", dep.CurrentVersion, dep.CurrentValue)
	for _, u := range dep.Updates {
		switch u.UpdateType {
		case "major", "minor", "patch":
		default:
			continue
		}
		newVersion := firstNonNil("?", u.NewVersion, u.NewValue)
		tab.addRow(key, docRow{
			Repo:         display,
			PackageFile:  pkgFile,
			Name:         depName,
			Version:      version,
			NewVersion:   newVersion,
			CurrentImage: fmt.Sprintf("%s:%s", depName, version),
			LatestImage:  fmt.Sprintf("%s:%s", depName, newVersion),
			UpdateType:   u.UpdateType,
			BadgeClass:   "badge-" + u.UpdateType,
		})
	}
}

// pendingByRepo counts, per repo, the deps with a listed update under their highest update type.
// The result is in caption order: repos with a major first, then by total descending, then by name.
func pendingByRepo(forks []string, tabs ...*tabBuilder) []repoPending {
	byRepo := map[string]*repoPending{}
	for _, tab := range tabs {
		for key, typ := range tab.highest {
			p := byRepo[key.repo]
			if p == nil {
				p = &repoPending{Repo: key.repo, Fork: slices.Contains(forks, key.repo)}
				byRepo[key.repo] = p
			}
			switch typ {
			case "major":
				p.Major++
			case "minor":
				p.Minor++
			default:
				p.Patch++
			}
		}
	}
	pending := make([]repoPending, 0, len(byRepo))
	for _, p := range byRepo {
		pending = append(pending, *p)
	}
	sort.Slice(pending, func(i, j int) bool { return pendingLess(pending[i], pending[j]) })
	return pending
}

// pendingLess orders the caption lines: a repo with a major first, then the larger total, then the name.
func pendingLess(a, b repoPending) bool {
	if (a.Major > 0) != (b.Major > 0) {
		return a.Major > 0
	}
	if a.total() != b.total() {
		return a.total() > b.total()
	}
	return shortRepoName(a.Repo) < shortRepoName(b.Repo)
}

// reportView is the data the document template renders.
type reportView struct {
	ScanTimestamp string
	Images        reportTab
	Deps          reportTab
	Errors        []string
}

// buildReportDocument renders the HTML report. scannedAt is when the hub received
// the Renovate report; it sets the header and, through documentName, the file name.
func buildReportDocument(rep renovateReport, forks []string, scannedAt time.Time) ([]byte, error) {
	loc, err := reportLocation()
	if err != nil {
		return nil, err
	}
	data := buildDocumentData(rep, forks)
	view := reportView{
		ScanTimestamp: scannedAt.In(loc).Format("2006-01-02 15:04") + " " + loc.String(),
		Images:        data.Images,
		Deps:          data.Deps,
		Errors:        data.Errors,
	}

	var buf bytes.Buffer
	if err := documentTemplate.Execute(&buf, view); err != nil {
		return nil, fmt.Errorf("render dependency report: %w", err)
	}
	return buf.Bytes(), nil
}

// buildCaption renders the summary sent with the document, as Telegram HTML. It reads the
// same data as the document, so its totals and counts agree with the tabs.
func buildCaption(rep renovateReport, forks []string, scannedAt time.Time) (string, error) {
	loc, err := reportLocation()
	if err != nil {
		return "", err
	}
	data := buildDocumentData(rep, forks)
	header := fmt.Sprintf("📦 <b>Dependency report</b> · %s", scannedAt.In(loc).Format("2006-01-02"))

	var repoLines, after []string
	if len(data.Pending) == 0 {
		after = []string{"✅ Todo al día"}
	} else {
		repoLines = captionRepoLines(data.Pending)
		if upToDate := data.Repos - len(data.Pending); upToDate > 0 {
			after = []string{fmt.Sprintf("✅ %d repos al día", upToDate)}
		}
	}
	return fitCaption(header, repoLines, after, captionStats(data)), nil
}

// captionRepoLines formats one line per repo: emoji, short name (escaped, with the fork tag) and counts.
func captionRepoLines(pending []repoPending) []string {
	lines := make([]string, 0, len(pending))
	for _, p := range pending {
		emoji := "🟡"
		if p.Major > 0 {
			emoji = "🔴"
		}
		name := htmlEscaper.Replace(shortRepoName(p.Repo))
		if p.Fork {
			name += " (fork)"
		}
		lines = append(lines, fmt.Sprintf("%s %s — %s", emoji, name, countsText(p)))
	}
	return lines
}

// countsText lists the non-zero counts of a repo in the order major, minor, patch.
func countsText(p repoPending) string {
	var parts []string
	if p.Major > 0 {
		parts = append(parts, fmt.Sprintf("%d major", p.Major))
	}
	if p.Minor > 0 {
		parts = append(parts, fmt.Sprintf("%d minor", p.Minor))
	}
	if p.Patch > 0 {
		parts = append(parts, fmt.Sprintf("%d patch", p.Patch))
	}
	return strings.Join(parts, " · ")
}

// captionStats are the totals lines. They use the tab totals of the document.
func captionStats(data documentData) []string {
	pendingDeps := data.Images.WithUpdates + data.Deps.WithUpdates
	upToDate := data.Images.UpToDate + data.Deps.UpToDate
	lines := []string{
		"",
		fmt.Sprintf("📊 %d desactualizadas · %d al día", pendingDeps, upToDate),
		fmt.Sprintf("🐳 Imágenes: %d · 📚 Paquetes: %d", data.Images.WithUpdates, data.Deps.WithUpdates),
	}
	if n := len(data.Errors); n > 0 {
		lines = append(lines, fmt.Sprintf("⚠️ %d errores en el scan", n))
	}
	return lines
}

// fitCaption joins the caption lines. While the caption is longer than captionLimit, it drops
// repo lines from the end and puts a "… y N repos más" line after the last kept one. The header
// and the footer are always kept, so a caption with no repo lines may still be over the limit.
func fitCaption(header string, repoLines, after, footer []string) string {
	for shown := len(repoLines); ; shown-- {
		lines := append([]string{header, ""}, repoLines[:shown]...)
		if shown < len(repoLines) {
			lines = append(lines, fmt.Sprintf("… y %d repos más (ver reporte)", len(repoLines)-shown))
		}
		lines = append(append(lines, after...), footer...)
		caption := strings.Join(lines, "\n")
		if utf8.RuneCountInString(caption) <= captionLimit || shown == 0 {
			return caption
		}
	}
}

// reportLocation loads the zone of the document header and file name.
func reportLocation() (*time.Location, error) {
	loc, err := time.LoadLocation(reportTimeZone)
	if err != nil {
		return nil, fmt.Errorf("load %s: %w", reportTimeZone, err)
	}
	return loc, nil
}

// documentName is the file name of the report received at scannedAt, dated in reportTimeZone.
func documentName(scannedAt time.Time) (string, error) {
	loc, err := reportLocation()
	if err != nil {
		return "", err
	}
	return "dependency-report-" + scannedAt.In(loc).Format("2006-01-02") + ".html", nil
}

// scanErrors lists each problem with level >= 50: the report root's first, then each repo's, repos sorted.
func scanErrors(rep renovateReport) []string {
	var errs []string
	for _, p := range rep.Problems {
		if p.Level >= 50 {
			errs = append(errs, "root: "+p.Msg)
		}
	}
	for _, name := range sortedRepoNames(rep) {
		for _, p := range rep.Repositories[name].Problems {
			if p.Level >= 50 {
				errs = append(errs, name+": "+p.Msg)
			}
		}
	}
	return errs
}

// sortedRepoNames returns the report's repo names in order.
func sortedRepoNames(rep renovateReport) []string {
	names := make([]string, 0, len(rep.Repositories))
	for name := range rep.Repositories {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// shortRepoName drops the owner prefix the captions leave out.
func shortRepoName(name string) string {
	return strings.TrimPrefix(name, "devidence-dev/")
}
