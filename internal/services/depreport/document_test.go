package depreport

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

// spanishDefaultsPath lists the Spanish default texts of the catalog, one per line. They live in a data file so
// that the test sources hold no Spanish text.
const spanishDefaultsPath = "../../handlers/testdata/spanish-defaults.txt"

// presentDocument renders s with the given overrides, in Guayaquil, with the default file names.
func presentDocument(t *testing.T, overrides map[string]any, s Summary) string {
	t.Helper()
	p := NewDocumentPresenter(catalogWith(t, overrides), loadZone(t, "America/Guayaquil"), "dependency-updates-{date}.yaml", "2006-01-02")
	out, err := p.Present(s)
	if err != nil {
		t.Fatalf("Present() error = %v", err)
	}
	return string(out)
}

// documentViewOf builds the view of s with the default texts.
func documentViewOf(t *testing.T, s Summary) documentView {
	t.Helper()
	r := &renderer{cat: catalogWith(t, nil)}
	view := buildDocumentView(r, loadZone(t, "America/Guayaquil"), s, "")
	if r.err != nil {
		t.Fatal(r.err)
	}
	return view
}

// dataYAML returns the data-yaml attributes of the export checkboxes, joined, so the checks read them alone.
func dataYAML(out string) string {
	return strings.Join(regexp.MustCompile(`data-yaml="[^"]*"`).FindAllString(out, -1), "\n")
}

// panes splits the document at the three tab panes.
func panes(t *testing.T, out string) (images, deps, charts string) {
	t.Helper()
	depsAt := strings.Index(out, `data-panel="deps"`)
	imagesAt := strings.Index(out, `data-panel="images"`)
	chartsAt := strings.Index(out, `data-panel="charts"`)
	if depsAt < 0 || imagesAt < depsAt || chartsAt < imagesAt {
		t.Fatal("tab panes not found in Dependencies-then-Images-then-Charts order")
	}
	charts = out[chartsAt:]
	if at := strings.Index(charts, "<dialog"); at >= 0 {
		charts = charts[:at]
	}
	return out[imagesAt:chartsAt], out[depsAt:imagesAt], charts
}

// TestDocumentPresenterPresent checks that the presenter renders the ReportDocument component: an HTML5
// document in English.
func TestDocumentPresenterPresent(t *testing.T) {
	out := presentDocument(t, nil, summaryAt(loadFixtureReport(t)))
	if !strings.HasPrefix(out, "<!doctype html>") { // templ writes the doctype in lower case
		t.Errorf("document starts with %.20q, want <!doctype html>", out)
	}
	if !strings.Contains(out, `<html lang="en">`) {
		t.Error(`document lacks <html lang="en">`)
	}
}

// TestDocumentStartsOnAutomatic checks that the page carries no stored theme: the html tag has no
// data-theme, and the Auto button is the pressed one.
func TestDocumentStartsOnAutomatic(t *testing.T) {
	out := presentDocument(t, nil, summaryAt(loadFixtureReport(t)))
	htmlTag := out[strings.Index(out, "<html"):]
	htmlTag = htmlTag[:strings.Index(htmlTag, ">")]
	if strings.Contains(htmlTag, "data-theme") {
		t.Errorf("html tag %q sets data-theme", htmlTag)
	}
	if !strings.Contains(out, `data-theme-set="system" title="Auto" aria-label="Auto" aria-pressed="true"`) {
		t.Error("the Auto button is not pressed on load")
	}
}

// TestDocumentStaticBehavior checks what the document keeps without a framework: one inline script (the
// report script), one stylesheet link (the Google Fonts link), no network file, and repo
// cards that are collapsed and have no name.
func TestDocumentStaticBehavior(t *testing.T) {
	out := presentDocument(t, nil, summaryAt(loadFixtureReport(t)))
	if n := strings.Count(out, "<script"); n != 1 {
		t.Errorf("script elements = %d, want 1", n)
	}
	if strings.Contains(out, "<script src") || strings.Contains(out, "<img") {
		t.Error("document loads a script or an image file")
	}
	if !strings.Contains(out, `<meta name="color-scheme" content="light dark">`) || !strings.Contains(out, `<meta name="viewport" content="width=device-width, initial-scale=1">`) {
		t.Error("document lacks the color-scheme or viewport meta")
	}
	links := regexp.MustCompile(`<link rel="stylesheet"[^>]*>`).FindAllString(out, -1)
	if len(links) != 1 || !strings.HasPrefix(links[0], `<link rel="stylesheet" href="https://fonts.googleapis.com/`) {
		t.Errorf("stylesheet links = %q, want only the Google Fonts link", links)
	}
	if strings.Contains(out, "cdn.jsdelivr.net") || strings.Contains(out, "bootstrap") {
		t.Error("document loads a file from a CDN or uses the old framework")
	}
	details := regexp.MustCompile(`<details[^>]*>`).FindAllString(out, -1)
	if len(details) == 0 {
		t.Fatal("fixture document has no repo sections")
	}
	openOrNamed := regexp.MustCompile(`\s(open|name)\b`)
	for _, tag := range details {
		if openOrNamed.MatchString(tag) {
			t.Errorf("repo section tag %s has an open or name attribute", tag)
		}
	}
}

func TestDocumentImageRowOnlyInImages(t *testing.T) {
	out := presentDocument(t, nil, summaryAt(loadFixtureReport(t)))
	imagesPane, depsPane, _ := panes(t, out)
	if !strings.Contains(imagesPane, "ghcr.io/immich-app/immich-server:v1.119.0") || strings.Contains(depsPane, "ghcr.io/immich-app/immich-server") {
		t.Error("image row not in the images pane only")
	}
}

func TestDocumentChartRowOnlyInCharts(t *testing.T) {
	const path = "kubernetes/infrastructure/traefik/application.yaml"
	rep := reportOf(t, viewRepo("devidence-dev/homelab", viewFile(path, viewDep("traefik", "helm", "42.3.0", "major:43.0.0"))))
	images, deps, charts := panes(t, presentDocument(t, nil, summaryAt(rep)))
	if strings.Contains(images, "traefik") || strings.Contains(deps, "traefik") {
		t.Error("chart row appeared outside the charts pane")
	}
	dir, short := splitPath(path)
	for _, want := range []string{`traefik`, `data-copy="traefik@43.0.0"`, `aria-label="Copy name@version"`, "homelab", `<span class="text-muted-2">` + dir + `</span>` + short, "42.3.0", `class="chip tone-major">1 major</span>`} {
		if !strings.Contains(charts, want) {
			t.Errorf("charts pane lacks %q", want)
		}
	}
}

func TestDocumentNonImageRowOnlyInDependencies(t *testing.T) {
	out := presentDocument(t, nil, summaryAt(loadFixtureReport(t)))
	imagesPane, depsPane, _ := panes(t, out)
	if !strings.Contains(depsPane, "golang.org/x/net") || strings.Contains(imagesPane, "golang.org/x/net") {
		t.Error("dependency row not in the dependencies pane only")
	}
}

func TestDocumentScriptEscaped(t *testing.T) {
	out := presentDocument(t, nil, summaryAt(loadFixtureReport(t)))
	if !strings.Contains(out, "&lt;script&gt;alert(1)&lt;/script&gt;") || strings.Contains(out, "<script>alert(1)</script>") {
		t.Error("a dependency name with a script is not escaped")
	}
	if want := `name: &#34;&lt;script&gt;alert(1)&lt;/script&gt;&#34;`; !strings.Contains(dataYAML(out), want) {
		t.Errorf("export attribute lacks the escaped name %q", want)
	}
}

func TestDocumentQuoteInVersionNoAttribute(t *testing.T) {
	rep := reportOf(t, `"devidence-dev/QuantWarden":{"packageFiles":{"dockerfile":[{"packageFile":"Dockerfile","deps":[
		{"depName":"nginx","currentValue":"1.0.0\" onmouseover=\"alert(1)","currentVersion":"1.0.0\" onmouseover=\"alert(1)","datasource":"docker","updates":[{"updateType":"patch","newVersion":"1.0.1","newValue":"1.0.1"}]}
	]}]}}`)
	out := presentDocument(t, nil, summaryAt(rep))
	if !strings.Contains(out, `1.0.0&#34; onmouseover=&#34;alert(1)`) {
		t.Error("current version not escaped as text")
	}
	if !strings.Contains(out, `data-copy="nginx:1.0.1"`) {
		t.Error("copy button does not copy the new reference")
	}
	if strings.Contains(out, `" onmouseover`) {
		t.Error("version created an onmouseover attribute")
	}
	if want := `current: &#34;1.0.0\&#34; onmouseover=\&#34;alert(1)&#34;`; !strings.Contains(dataYAML(out), want) {
		t.Errorf("export attribute lacks the escaped version %q", want)
	}
}

func TestDocumentErrorsSection(t *testing.T) {
	out := presentDocument(t, nil, summaryAt(loadFixtureReport(t)))
	if !strings.Contains(out, "Failed to look up dependency") || !strings.Contains(out, "Scan errors") {
		t.Error("lookup failure missing from the errors card")
	}
	warn := reportOf(t, `"devidence-dev/warn-only":{"problems":[{"level":40,"msg":"Just a warning"}]}`)
	if out := presentDocument(t, nil, summaryAt(warn)); strings.Contains(out, `id="errors-heading"`) {
		t.Error("errors card shown for a level 40 problem")
	}
}

func TestDocumentNothingPending(t *testing.T) {
	var docker, others []string
	for i := 0; i < 12; i++ {
		docker = append(docker, fmt.Sprintf(`{"depName":"img-%02d","currentValue":"1","currentVersion":"1","datasource":"docker","updates":[]}`, i))
	}
	for i := 0; i < 30; i++ {
		others = append(others, fmt.Sprintf(`{"depName":"dep-%02d","currentValue":"1","currentVersion":"1","datasource":"go","updates":[]}`, i))
	}
	rep := reportOf(t, `"devidence-dev/QuantWarden":{"packageFiles":{"dockerfile":[{"packageFile":"Dockerfile","deps":[`+strings.Join(docker, ",")+`]},{"packageFile":"go.mod","deps":[`+strings.Join(others, ",")+`]}]}}`)
	out := presentDocument(t, nil, summaryAt(rep))
	for _, want := range []string{
		`>0</span>`,
		"pending out of 30 · 30 up to date",
		"All dependencies are up to date",
		"All images are up to date",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("document lacks %q", want)
		}
	}
	if strings.Contains(out, "<details") || strings.Contains(out, `class="bar-seg`) {
		t.Error("repo cards or bar segments rendered without pending updates")
	}
	if regexp.MustCompile(`<button[^>]*data-export-open`).MatchString(out) || strings.Contains(out, "<dialog") {
		t.Error("export button or dialog rendered without pending updates")
	}
}

// TestDocumentLangOverride checks that overriding lang changes only the html lang attribute: the labels
// stay English.
// TestNothingPendingWithCharts checks a nothing-pending document with charts: the charts tab reads as up to date
// too, and the caption counts the charts in the totals.
func TestNothingPendingWithCharts(t *testing.T) {
	s := summaryAt(nothingPendingReport(t, 4))
	out := presentDocument(t, nil, s)
	for _, want := range []string{"All charts are up to date", "pending out of 4 · 4 up to date"} {
		if !strings.Contains(out, want) {
			t.Errorf("document lacks %q", want)
		}
	}
	caption := presentCaption(t, nil, s)
	for _, want := range []string{"✅ All up to date", "📊 0 outdated · 46 up to date"} {
		if !strings.Contains(caption, want) {
			t.Errorf("caption lacks %q: %s", want, caption)
		}
	}
}

func TestDocumentLangOverride(t *testing.T) {
	out := presentDocument(t, map[string]any{"lang": "es"}, summaryAt(loadFixtureReport(t)))
	if !strings.Contains(out, `<html lang="es">`) {
		t.Error(`document lacks <html lang="es">`)
	}
	if !strings.Contains(out, "pending out of") {
		t.Error("labels are not English after a lang override")
	}
	if strings.Contains(out, "pendientes de") {
		t.Error("labels are Spanish after a lang override that only sets lang")
	}
}

// TestDocumentDefaultsHaveNoSpanishLabels checks the English defaults: the document has lang="en", none of
// the Spanish default texts of the catalog, and no Spanish accent.
func TestDocumentDefaultsHaveNoSpanishLabels(t *testing.T) {
	out := presentDocument(t, nil, summaryAt(loadFixtureReport(t)))
	data, err := os.ReadFile(spanishDefaultsPath) //nolint:gosec // G304: fixed data file in testdata
	if err != nil {
		t.Fatal(err)
	}
	for _, label := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if strings.Contains(out, label) {
			t.Errorf("document has the Spanish label %q", label)
		}
	}
	if strings.ContainsAny(out, "áéíóúñ¿¡") {
		t.Error("document has a Spanish accent")
	}
	if !strings.Contains(out, `<html lang="en">`) {
		t.Error(`default document lacks <html lang="en">`)
	}
}

func TestDocumentForkTag(t *testing.T) {
	s := summaryAt(loadFixtureReport(t))
	if out := presentDocument(t, nil, s); !strings.Contains(out, "fork · informational only") {
		t.Error("default document fork note missing from the summary row")
	}
	out := presentDocument(t, map[string]any{"fork": map[string]any{"tag_document": "fork"}}, s)
	if !strings.Contains(out, ">fork</span>") || strings.Contains(out, "informational only") {
		t.Error("overridden fork note missing from the summary row")
	}
	view := documentViewOf(t, s)
	forks := 0
	for _, tab := range view.Tabs {
		for _, g := range tab.Groups {
			if g.Fork != (g.Repo == testForkRepo) {
				t.Errorf("group %s Fork = %v, want %v", g.Repo, g.Fork, g.Repo == testForkRepo)
			}
			if g.Fork {
				forks++
			}
		}
	}
	if forks == 0 {
		t.Error("no fork group in the fixture")
	}
}

func TestDocumentTimestampZone(t *testing.T) {
	out := presentDocument(t, nil, summaryAt(loadFixtureReport(t)))
	if !strings.Contains(out, "2026-10-12 09:07 America/Guayaquil") {
		t.Error("document header lacks the Guayaquil scan time")
	}
}

func TestDocumentName(t *testing.T) {
	ghe := loadZone(t, "America/Guayaquil")
	madrid := loadZone(t, "Europe/Madrid")
	tests := []struct {
		name      string
		scannedAt time.Time
		loc       *time.Location
		pattern   string
		layout    string
		want      string
	}{
		{name: "default", scannedAt: time.Date(2026, 10, 12, 14, 7, 0, 0, time.UTC), loc: ghe,
			pattern: "dependency-report-{date}.html", layout: "2006-01-02", want: "dependency-report-2026-10-12.html"},
		{name: "near midnight", scannedAt: time.Date(2026, 10, 13, 2, 30, 0, 0, time.UTC), loc: ghe,
			pattern: "dependency-report-{date}.html", layout: "2006-01-02", want: "dependency-report-2026-10-12.html"},
		{name: "configured zone", scannedAt: time.Date(2026, 10, 12, 22, 30, 0, 0, time.UTC), loc: madrid,
			pattern: "dependency-report-{date}.html", layout: "2006-01-02", want: "dependency-report-2026-10-13.html"},
		{name: "custom pattern", scannedAt: time.Date(2026, 10, 12, 14, 7, 0, 0, time.UTC), loc: ghe,
			pattern: "report-{date}.html", layout: "20060102", want: "report-20261012.html"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DocumentName(tt.scannedAt, tt.loc, tt.pattern, tt.layout); got != tt.want {
				t.Errorf("DocumentName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPresentersShareSummary(t *testing.T) {
	s := summaryAt(loadFixtureReport(t))
	caption := presentCaption(t, nil, s)
	wantStats := "📊 " + strconv.Itoa(s.Images.WithUpdates+s.Deps.WithUpdates) + " outdated · " + strconv.Itoa(s.Images.UpToDate+s.Deps.UpToDate) + " up to date"
	if !strings.Contains(caption, wantStats) {
		t.Errorf("caption lacks %q", wantStats)
	}
	view := documentViewOf(t, s)
	if n := view.Tabs[0].Count + view.Tabs[1].Count; n != s.Images.Updates+s.Deps.Updates {
		t.Errorf("document badges = %d, summary = %d", n, s.Images.Updates+s.Deps.Updates)
	}
}
