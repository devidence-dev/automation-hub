package depreport

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

// presentDocument renders s with the given overrides, in Guayaquil.
func presentDocument(t *testing.T, overrides map[string]any, s Summary) string {
	t.Helper()
	p := NewDocumentPresenter(catalogWith(t, overrides), loadZone(t, "America/Guayaquil"))
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
	view := buildDocumentView(r, loadZone(t, "America/Guayaquil"), s)
	if r.err != nil {
		t.Fatal(r.err)
	}
	return view
}

// panes splits the document at the two tab panes.
func panes(t *testing.T, out string) (images, deps string) {
	t.Helper()
	depsAt := strings.Index(out, `data-panel="deps"`)
	imagesAt := strings.Index(out, `data-panel="images"`)
	if depsAt < 0 || imagesAt < depsAt {
		t.Fatal("tab panes not found in document, or not in Dependencias-then-Imágenes order") //nolint:misspell // Spanish text fixed by the spec
	}
	return out[imagesAt:], out[depsAt:imagesAt]
}

// TestDocumentPresenterPresent checks that the presenter renders the ReportDocument component: an HTML5
// document in Spanish.
func TestDocumentPresenterPresent(t *testing.T) {
	out := presentDocument(t, nil, summaryAt(loadFixtureReport(t)))
	if !strings.HasPrefix(out, "<!doctype html>") { // templ writes the doctype in lower case
		t.Errorf("document starts with %.20q, want <!doctype html>", out)
	}
	if !strings.Contains(out, `<html lang="es">`) {
		t.Error(`document lacks <html lang="es">`)
	}
}

// TestDocumentStartsOnAutomatic checks that the page carries no stored theme: the html tag has no
// data-theme, and the Automático button is the pressed one.
func TestDocumentStartsOnAutomatic(t *testing.T) {
	out := presentDocument(t, nil, summaryAt(loadFixtureReport(t)))
	htmlTag := out[strings.Index(out, "<html"):]
	htmlTag = htmlTag[:strings.Index(htmlTag, ">")]
	if strings.Contains(htmlTag, "data-theme") {
		t.Errorf("html tag %q sets data-theme", htmlTag)
	}
	if !strings.Contains(out, `data-theme-set="system" title="Automático" aria-label="Automático" aria-pressed="true"`) {
		t.Error("the Automático button is not pressed on load")
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
	imagesPane, depsPane := panes(t, out)
	if !strings.Contains(imagesPane, "ghcr.io/immich-app/immich-server:v1.119.0") || strings.Contains(depsPane, "ghcr.io/immich-app/immich-server") {
		t.Error("image row not in the images pane only")
	}
}

func TestDocumentNonImageRowOnlyInDependencies(t *testing.T) {
	out := presentDocument(t, nil, summaryAt(loadFixtureReport(t)))
	imagesPane, depsPane := panes(t, out)
	if !strings.Contains(depsPane, "golang.org/x/net") || strings.Contains(imagesPane, "golang.org/x/net") {
		t.Error("dependency row not in the dependencies pane only")
	}
}

func TestDocumentScriptEscaped(t *testing.T) {
	out := presentDocument(t, nil, summaryAt(loadFixtureReport(t)))
	if !strings.Contains(out, "&lt;script&gt;alert(1)&lt;/script&gt;") || strings.Contains(out, "<script>alert(1)</script>") {
		t.Error("a dependency name with a script is not escaped")
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
}

func TestDocumentErrorsSection(t *testing.T) {
	out := presentDocument(t, nil, summaryAt(loadFixtureReport(t)))
	if !strings.Contains(out, "Failed to look up dependency") || !strings.Contains(out, "Errores del análisis") {
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
		"pendientes de 30 · 30 al día",        //nolint:misspell // Spanish text fixed by the spec
		"Todas las dependencias están al día", //nolint:misspell // Spanish text fixed by the spec
		"Todas las imágenes están al día",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("document lacks %q", want)
		}
	}
	if strings.Contains(out, "<details") || strings.Contains(out, `class="bar-seg`) {
		t.Error("repo cards or bar segments rendered without pending updates")
	}
}

func TestDocumentLangOverride(t *testing.T) {
	out := presentDocument(t, map[string]any{"lang": "en"}, summaryAt(loadFixtureReport(t)))
	if !strings.Contains(out, `<html lang="en">`) {
		t.Error(`document lacks <html lang="en">`)
	}
	if !strings.Contains(out, "pendientes de") {
		t.Error("labels are not Spanish after a lang override")
	}
}

func TestDocumentDefaultsHaveNoEnglishLabels(t *testing.T) {
	out := presentDocument(t, nil, summaryAt(loadFixtureReport(t)))
	for _, label := range []string{
		"Dependency Report", "Pending Updates", "Up to Date", "Update Distribution", "Available Updates",
		"Package File", ">Repository<", "Current Image", "Latest Image", ">Dependency<", ">Current<", ">New<",
		">Type<", "<h5>Errors</h5>", "Copy current image", "Copy latest image", "Copied!",
		"All dependencies are up to date", "updates available",
	} {
		if strings.Contains(out, label) {
			t.Errorf("document has the English label %q", label)
		}
	}
	if !strings.Contains(out, `<html lang="es">`) {
		t.Error(`default document lacks <html lang="es">`)
	}
}

func TestDocumentForkTag(t *testing.T) {
	s := summaryAt(loadFixtureReport(t))
	if out := presentDocument(t, nil, s); !strings.Contains(out, "fork · solo informativo") { //nolint:misspell // Spanish text fixed by the spec
		t.Error("default document fork note missing from the summary row")
	}
	out := presentDocument(t, map[string]any{"fork": map[string]any{"tag_document": "fork"}}, s)
	if !strings.Contains(out, ">fork</span>") || strings.Contains(out, "solo informativo") { //nolint:misspell // Spanish text fixed by the spec
		t.Error("overridden fork note missing from the summary row")
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
	wantStats := "📊 " + strconv.Itoa(s.Images.WithUpdates+s.Deps.WithUpdates) + " desactualizadas · " + strconv.Itoa(s.Images.UpToDate+s.Deps.UpToDate) + " al día"
	if !strings.Contains(caption, wantStats) {
		t.Errorf("caption lacks %q", wantStats)
	}
	view := documentViewOf(t, s)
	if n := view.Tabs[0].Count + view.Tabs[1].Count; n != s.Images.Updates+s.Deps.Updates {
		t.Errorf("document badges = %d, summary = %d", n, s.Images.Updates+s.Deps.Updates)
	}
}
