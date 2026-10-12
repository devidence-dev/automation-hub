package handlers

import (
	"bytes"
	_ "embed" // dependency_report.html
	"fmt"
	"html/template"
	"slices"
	"sort"
	"time"
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

// documentData is the content of the HTML report, split into its two tabs.
type documentData struct {
	Images reportTab
	Deps   reportTab
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
type depKey struct{ repo, file, name string }

// tabBuilder collects rows and dep identities for one tab.
type tabBuilder struct {
	deps     map[depKey]bool
	updated  map[depKey]bool
	seenRows map[docRow]bool
	rows     []docRow
}

func newTabBuilder() *tabBuilder {
	return &tabBuilder{deps: map[depKey]bool{}, updated: map[depKey]bool{}, seenRows: map[docRow]bool{}}
}

func (b *tabBuilder) addRow(key depKey, row docRow) {
	b.updated[key] = true
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

// buildDocumentData splits the report's deps into the image and dependency tabs.
// Only major, minor and patch updates become rows; digest and pin are ignored.
func buildDocumentData(rep renovateReport, forks []string) documentData {
	images, deps := newTabBuilder(), newTabBuilder()

	repos := make([]string, 0, len(rep.Repositories))
	for name := range rep.Repositories {
		repos = append(repos, name)
	}
	sort.Strings(repos)

	for _, name := range repos {
		display := name
		if slices.Contains(forks, name) {
			display += forkTag
		}
		for _, files := range rep.Repositories[name].PackageFiles {
			for _, file := range files {
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
					key := depKey{repo: display, file: file.PackageFile, name: keyName}
					tab.deps[key] = true

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
							PackageFile:  file.PackageFile,
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
			}
		}
	}

	return documentData{Images: images.finish(), Deps: deps.finish()}
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
		Errors:        scanErrors(rep),
	}

	var buf bytes.Buffer
	if err := documentTemplate.Execute(&buf, view); err != nil {
		return nil, fmt.Errorf("render dependency report: %w", err)
	}
	return buf.Bytes(), nil
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

// scanErrors lists each problem with level >= 50 as "<repo>: <msg>", repos sorted.
func scanErrors(rep renovateReport) []string {
	repos := make([]string, 0, len(rep.Repositories))
	for name := range rep.Repositories {
		repos = append(repos, name)
	}
	sort.Strings(repos)

	var errs []string
	for _, name := range repos {
		for _, p := range rep.Repositories[name].Problems {
			if p.Level >= 50 {
				errs = append(errs, name+": "+p.Msg)
			}
		}
	}
	return errs
}
