package depreport

import (
	"fmt"
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
	imagesAt := strings.Index(out, `id="images" role="tabpanel"`)
	depsAt := strings.Index(out, `id="deps" role="tabpanel"`)
	if imagesAt < 0 || depsAt < imagesAt {
		t.Fatal("tab panes not found in document")
	}
	return out[imagesAt:depsAt], out[depsAt:]
}

func TestDocumentTabColumns(t *testing.T) {
	view := documentViewOf(t, summaryAt(loadFixtureReport(t)))
	if n := len(view.Tabs[0].Columns); n != 5 {
		t.Errorf("images columns = %d, want 5", n)
	}
	if n := len(view.Tabs[1].Columns); n != 6 {
		t.Errorf("dependencies columns = %d, want 6", n)
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
	if !strings.Contains(out, `data-copy="nginx:1.0.0&#34; onmouseover=&#34;alert(1)"`) {
		t.Error("copy value not escaped inside its attribute")
	}
	if strings.Contains(out, `" onmouseover=`) || strings.Contains(out, ` onmouseover="`) {
		t.Error("version created an onmouseover attribute")
	}
}

func TestDocumentErrorsSection(t *testing.T) {
	out := presentDocument(t, nil, summaryAt(loadFixtureReport(t)))
	if !strings.Contains(out, "devidence-dev/QuantWarden: Failed to look up dependency") {
		t.Error("lookup failure missing from the errors section")
	}
	warn := reportOf(t, `"devidence-dev/warn-only":{"problems":[{"level":40,"msg":"Just a warning"}]}`)
	if out := presentDocument(t, nil, summaryAt(warn)); strings.Contains(out, "<h5>Errores</h5>") {
		t.Error("errors section shown for a level 40 problem")
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
		`<div class="metric-value" style="color: var(--accent-yellow);">0</div>`,
		`<div class="metric-value" style="color: var(--accent-green);">12</div>`,
		`<div class="metric-value" style="color: var(--accent-green);">30</div>`,
		"Todas las dependencias están al día", //nolint:misspell // Spanish text fixed by the spec
	} {
		if !strings.Contains(out, want) {
			t.Errorf("document lacks %q", want)
		}
	}
	if strings.Contains(out, "<table") {
		t.Error("table rendered without rows")
	}
}

func TestDocumentLangOverride(t *testing.T) {
	out := presentDocument(t, map[string]any{"lang": "en"}, summaryAt(loadFixtureReport(t)))
	if !strings.Contains(out, `<html lang="en">`) {
		t.Error(`document lacks <html lang="en">`)
	}
	if !strings.Contains(out, "Actualizaciones pendientes") {
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
	if out := presentDocument(t, nil, s); !strings.Contains(out, "discord-tts-bot (fork, solo informativo)") { //nolint:misspell // Spanish text fixed by the spec
		t.Error("default document fork tag missing")
	}
	out := presentDocument(t, map[string]any{"fork": map[string]any{"tag_document": " (fork)"}}, s)
	if !strings.Contains(out, "devidence-dev/discord-tts-bot (fork)<") || strings.Contains(out, "solo informativo") { //nolint:misspell // Spanish text fixed by the spec
		t.Error("overridden fork tag missing")
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
	if total := view.Tabs[0].Total + view.Tabs[1].Total; total != s.Images.Total+s.Deps.Total {
		t.Errorf("document totals = %d, summary = %d", total, s.Images.Total+s.Deps.Total)
	}
}
