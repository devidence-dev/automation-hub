package depreport

import (
	"cmp"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"
)

// unknownValue stands in for a dep name or version that Renovate does not give.
const unknownValue = "?"

// dockerDatasource is the Renovate datasource of container images. Those deps go to the images
// tab; every other dep, including one without a datasource, goes to the dependencies tab.
const dockerDatasource = "docker"

// errorLevel is the Renovate log level from which a problem is listed as a scan error.
const errorLevel = 50

// Row is one listed update: one dep and one update of type major, minor or patch.
type Row struct {
	Repo        string // report name, for example devidence-dev/QuantWarden
	Fork        bool   // the repo is one of the configured forks
	PackageFile string
	Name        string
	Version     string // current version
	NewVersion  string
	Type        UpdateType
}

// Tab is one of the two tabs of the document.
type Tab struct {
	Rows        []Row // sorted by repo, package file, dep, update type and versions
	Updates     int   // listed rows, one per dep and update
	Total       int   // distinct deps in the tab, with or without listed updates
	WithUpdates int   // distinct deps with at least one listed row
	UpToDate    int   // Total minus WithUpdates
	Counts      map[UpdateType]int
}

// RepoPending counts the deps of one repo that have a listed update, each under its highest type.
type RepoPending struct {
	Repo  string
	Fork  bool
	Major int
	Minor int
	Patch int
}

// Total is the number of deps of the repo with a listed update.
func (p RepoPending) Total() int { return p.Major + p.Minor + p.Patch }

// ScanError is a problem of level 50 or more. Repo is empty for a problem of the report root.
type ScanError struct {
	Repo string
	Msg  string
}

// Summary is what the report shows. Aggregate computes it once; every presenter reads the same value.
type Summary struct {
	ReceivedAt time.Time // when the hub received the report, set by the handler
	Images     Tab
	Deps       Tab
	Repos      int           // repos in the report, with or without pending updates
	Pending    []RepoPending // repos with a listed update, in caption order
	Errors     []ScanError
}

// Aggregate computes the summary of a report. forks are the repo names tagged as forks.
func Aggregate(rep Report, forks []string) Summary {
	images, deps := newTabBuilder(), newTabBuilder()
	for _, name := range sortedRepoNames(rep) {
		fork := slices.Contains(forks, name)
		for _, files := range rep.Repositories[name].PackageFiles {
			for _, file := range files {
				addFileDeps(images, deps, name, fork, file)
			}
		}
	}
	return Summary{
		Images:  images.finish(),
		Deps:    deps.finish(),
		Repos:   len(rep.Repositories),
		Pending: pendingByRepo(forks, images, deps),
		Errors:  scanErrors(rep),
	}
}

// ShortName is the last segment of a repo name: payments-api for acme-corp/payments-api.
func ShortName(repo string) string {
	return repo[strings.LastIndex(repo, "/")+1:]
}

// depKey identifies a dep in the report: the same dep may have several updates.
type depKey struct{ repo, file, name string }

// tabBuilder collects the rows and the dep identities of one tab.
type tabBuilder struct {
	deps     map[depKey]bool
	updated  map[depKey]bool
	highest  map[depKey]UpdateType // highest listed type of each dep with a listed update
	seenRows map[Row]bool
	rows     []Row
}

func newTabBuilder() *tabBuilder {
	return &tabBuilder{
		deps:     map[depKey]bool{},
		updated:  map[depKey]bool{},
		highest:  map[depKey]UpdateType{},
		seenRows: map[Row]bool{},
	}
}

// addRow records a listed row once; exact duplicates are dropped.
func (b *tabBuilder) addRow(key depKey, row Row) {
	b.updated[key] = true
	if cur, ok := b.highest[key]; !ok || row.Type.Rank() < cur.Rank() {
		b.highest[key] = row.Type
	}
	if !b.seenRows[row] {
		b.seenRows[row] = true
		b.rows = append(b.rows, row)
	}
}

// finish sorts the rows and computes the totals of the tab.
func (b *tabBuilder) finish() Tab {
	slices.SortStableFunc(b.rows, compareRows)
	tab := Tab{
		Rows:        b.rows,
		Updates:     len(b.rows),
		Total:       len(b.deps),
		WithUpdates: len(b.updated),
		Counts:      map[UpdateType]int{},
	}
	tab.UpToDate = tab.Total - tab.WithUpdates
	for _, row := range b.rows {
		tab.Counts[row.Type]++
	}
	return tab
}

// compareRows orders rows by repo, package file, dep, then major before minor before patch, then versions.
func compareRows(a, b Row) int {
	return cmp.Or(
		cmp.Compare(a.Repo, b.Repo),
		cmp.Compare(a.PackageFile, b.PackageFile),
		cmp.Compare(a.Name, b.Name),
		cmp.Compare(a.Type.Rank(), b.Type.Rank()),
		cmp.Compare(a.Version, b.Version),
		cmp.Compare(a.NewVersion, b.NewVersion),
	)
}

// addFileDeps adds the deps of one package file to the images tab or the dependencies tab.
func addFileDeps(images, deps *tabBuilder, repo string, fork bool, file PackageFile) {
	for i, dep := range file.Deps {
		tab := deps
		if dep.Datasource != nil && *dep.Datasource == dockerDatasource {
			tab = images
		}
		name := firstNonNil(unknownValue, dep.DepName)
		// Unnamed deps (for example Dockerfile stage refs) are counted one by one.
		keyName := name
		if dep.DepName == nil {
			keyName = fmt.Sprintf("%s#%d", unknownValue, i)
		}
		key := depKey{repo: repo, file: file.PackageFile, name: keyName}
		tab.deps[key] = true
		base := Row{
			Repo:        repo,
			Fork:        fork,
			PackageFile: file.PackageFile,
			Name:        name,
			Version:     firstNonNil(unknownValue, dep.CurrentVersion, dep.CurrentValue),
		}
		tab.addDepRows(key, base, dep)
	}
}

// addDepRows adds one row per listed update of a dep; other update types are ignored.
func (b *tabBuilder) addDepRows(key depKey, base Row, dep Dep) {
	for _, u := range dep.Updates {
		typ := ParseUpdateType(u.UpdateType)
		if !typ.Listed() {
			continue
		}
		row := base
		row.NewVersion = firstNonNil(unknownValue, u.NewVersion, u.NewValue)
		row.Type = typ
		b.addRow(key, row)
	}
}

// pendingByRepo counts, per repo, the deps with a listed update under their highest type.
// The result is in caption order.
func pendingByRepo(forks []string, tabs ...*tabBuilder) []RepoPending {
	byRepo := map[string]*RepoPending{}
	for _, tab := range tabs {
		for key, typ := range tab.highest {
			p := byRepo[key.repo]
			if p == nil {
				p = &RepoPending{Repo: key.repo, Fork: slices.Contains(forks, key.repo)}
				byRepo[key.repo] = p
			}
			switch typ {
			case UpdateMajor:
				p.Major++
			case UpdateMinor:
				p.Minor++
			case UpdatePatch:
				p.Patch++
			}
		}
	}
	pending := make([]RepoPending, 0, len(byRepo))
	for _, p := range byRepo {
		pending = append(pending, *p)
	}
	sort.Slice(pending, func(i, j int) bool { return pendingLess(pending[i], pending[j]) })
	return pending
}

// pendingLess orders the caption lines: a repo with a major first, then the larger total, then the name.
func pendingLess(a, b RepoPending) bool {
	if (a.Major > 0) != (b.Major > 0) {
		return a.Major > 0
	}
	if a.Total() != b.Total() {
		return a.Total() > b.Total()
	}
	return ShortName(a.Repo) < ShortName(b.Repo)
}

// scanErrors lists each problem of level 50 or more: the report root's first, then each repo's, repos sorted.
func scanErrors(rep Report) []ScanError {
	var errs []ScanError
	for _, p := range rep.Problems {
		if p.Level >= errorLevel {
			errs = append(errs, ScanError{Msg: p.Msg})
		}
	}
	for _, name := range sortedRepoNames(rep) {
		for _, p := range rep.Repositories[name].Problems {
			if p.Level >= errorLevel {
				errs = append(errs, ScanError{Repo: name, Msg: p.Msg})
			}
		}
	}
	return errs
}

// sortedRepoNames returns the report's repo names in order.
func sortedRepoNames(rep Report) []string {
	names := make([]string, 0, len(rep.Repositories))
	for name := range rep.Repositories {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// firstNonNil returns the first value that is set, or def.
func firstNonNil(def string, values ...*string) string {
	for _, v := range values {
		if v != nil {
			return *v
		}
	}
	return def
}
