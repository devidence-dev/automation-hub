package handlers

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	_ "time/tzdata" // the hub embeds the zone database too (cmd/automation-hub)
)

// imageRows and depRows render the rows of a tab as "|"-separated strings so the
// table tests compare whole rows at once.
func imageRows(tab reportTab) []string {
	var out []string
	for _, r := range tab.Rows {
		out = append(out, strings.Join([]string{r.Repo, r.PackageFile, r.CurrentImage, r.LatestImage, r.UpdateType}, "|"))
	}
	return out
}

func depRows(tab reportTab) []string {
	var out []string
	for _, r := range tab.Rows {
		out = append(out, strings.Join([]string{r.Repo, r.PackageFile, r.Name, r.Version, r.NewVersion, r.UpdateType}, "|"))
	}
	return out
}

func TestBuildDocumentData(t *testing.T) {
	tests := []struct {
		name       string
		json       string
		forks      []string
		wantImages []string
		wantDeps   []string
	}{
		{
			name: "image dependency",
			json: `{"repositories":{"devidence-dev/homelab":{"packageFiles":{"helm-values":[{"packageFile":"kubernetes/apps/media/immich/values.yaml","deps":[
				{"depName":"ghcr.io/immich-app/immich-server","currentValue":"v1.118.0","currentVersion":"v1.118.0","datasource":"docker","updates":[{"updateType":"minor","newVersion":"v1.119.0","newValue":"v1.119.0"}]}
			]}]}}}}`,
			wantImages: []string{"devidence-dev/homelab|kubernetes/apps/media/immich/values.yaml|ghcr.io/immich-app/immich-server:v1.118.0|ghcr.io/immich-app/immich-server:v1.119.0|minor"},
		},
		{
			name: "non-image dependency",
			json: `{"repositories":{"devidence-dev/QuantWarden":{"packageFiles":{"gomod":[{"packageFile":"go.mod","deps":[
				{"depName":"golang.org/x/net","currentValue":"0.30.0","currentVersion":"0.30.0","datasource":"go","updates":[{"updateType":"major","newVersion":"1.0.0","newValue":"1.0.0"}]}
			]}]}}}}`,
			wantDeps: []string{"devidence-dev/QuantWarden|go.mod|golang.org/x/net|0.30.0|1.0.0|major"},
		},
		{
			name: "Dockerfile image and missing datasource",
			json: `{"repositories":{"devidence-dev/discord-tts-bot":{"packageFiles":{"dockerfile":[{"packageFile":"Dockerfile","deps":[
				{"depName":"golang","currentValue":"1.26","currentVersion":"1.26","datasource":"docker","updates":[{"updateType":"minor","newVersion":"1.27","newValue":"1.27"}]},
				{"depName":"tool","currentValue":"1.0.0","currentVersion":"1.0.0","updates":[{"updateType":"patch","newVersion":"1.0.1","newValue":"1.0.1"}]}
			]}]}}}}`,
			wantImages: []string{"devidence-dev/discord-tts-bot|Dockerfile|golang:1.26|golang:1.27|minor"},
			wantDeps:   []string{"devidence-dev/discord-tts-bot|Dockerfile|tool|1.0.0|1.0.1|patch"},
		},
		{
			name: "digest only",
			json: `{"repositories":{"devidence-dev/digest-only":{"packageFiles":{"dockerfile":[{"packageFile":"Dockerfile","deps":[
				{"depName":"alpine","currentValue":"3.20","currentVersion":"3.20","datasource":"docker","updates":[{"updateType":"digest","newVersion":"3.20","newValue":"3.20"}]}
			]}]}}}}`,
		},
		{
			name: "several updates for one dep",
			json: `{"repositories":{"devidence-dev/QuantWarden":{"packageFiles":{"gomod":[{"packageFile":"go.mod","deps":[
				{"depName":"github.com/gorilla/mux","currentValue":"1.8.0","currentVersion":"1.8.0","datasource":"go","updates":[
					{"updateType":"minor","newVersion":"1.9.0","newValue":"1.9.0"},
					{"updateType":"major","newVersion":"2.0.0","newValue":"2.0.0"},
					{"updateType":"digest","newVersion":"1.8.0","newValue":"1.8.0"}
				]}
			]}]}}}}`,
			wantDeps: []string{
				"devidence-dev/QuantWarden|go.mod|github.com/gorilla/mux|1.8.0|2.0.0|major",
				"devidence-dev/QuantWarden|go.mod|github.com/gorilla/mux|1.8.0|1.9.0|minor",
			},
		},
		{
			name:  "fork tag",
			forks: []string{"devidence-dev/discord-tts-bot"},
			json: `{"repositories":{"devidence-dev/discord-tts-bot":{"packageFiles":{"npm":[{"packageFile":"package.json","deps":[
				{"depName":"discord.js","currentValue":"14.1.0","currentVersion":"14.1.0","datasource":"npm","updates":[{"updateType":"minor","newVersion":"14.2.0","newValue":"14.2.0"}]}
			]}]}}}}`,
			wantDeps: []string{"devidence-dev/discord-tts-bot" + testForkTag + "|package.json|discord.js|14.1.0|14.2.0|minor"},
		},
		{
			name: "duplicate deps are listed once",
			json: `{"repositories":{"devidence-dev/QuantWarden":{"packageFiles":{"gomod":[{"packageFile":"go.mod","deps":[
				{"depName":"go.uber.org/zap","currentValue":"1.27.0","currentVersion":"1.27.0","datasource":"go","updates":[{"updateType":"patch","newVersion":"1.27.1","newValue":"1.27.1"}]},
				{"depName":"go.uber.org/zap","currentValue":"1.27.0","currentVersion":"1.27.0","datasource":"go","updates":[{"updateType":"patch","newVersion":"1.27.1","newValue":"1.27.1"}]}
			]}]}}}}`,
			wantDeps: []string{"devidence-dev/QuantWarden|go.mod|go.uber.org/zap|1.27.0|1.27.1|patch"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rep, err := parseReport([]byte(tt.json))
			if err != nil {
				t.Fatal(err)
			}
			data := buildDocumentData(rep, tt.forks)
			if got := imageRows(data.Images); !reflect.DeepEqual(got, tt.wantImages) {
				t.Errorf("image rows = %q, want %q", got, tt.wantImages)
			}
			if got := depRows(data.Deps); !reflect.DeepEqual(got, tt.wantDeps) {
				t.Errorf("dependency rows = %q, want %q", got, tt.wantDeps)
			}
		})
	}
}

func TestBuildDocumentDataUnnamedDepsCountedSeparately(t *testing.T) {
	// Two Dockerfile stage refs have no depName (Renovate's null); each is still one dep.
	rep, err := parseReport([]byte(`{"repositories":{"devidence-dev/QuantWarden":{"packageFiles":{"dockerfile":[{"packageFile":"Dockerfile","deps":[
		{"datasource":"docker","updates":[]},
		{"datasource":"docker","updates":[]},
		{"depName":"alpine","currentValue":"3.20","currentVersion":"3.20","datasource":"docker","updates":[]}
	]}]}}}}`))
	if err != nil {
		t.Fatal(err)
	}
	if got := buildDocumentData(rep, nil).Images.Total; got != 3 {
		t.Errorf("image tab total = %d, want 3", got)
	}
}

func TestBuildDocumentDataTotals(t *testing.T) {
	// 3 docker deps, one with a minor update; 5 other deps, two with updates (patch and major).
	var images, deps []string
	for i := 0; i < 3; i++ {
		update := `[]`
		if i == 0 {
			update = `[{"updateType":"minor","newVersion":"2.0.0","newValue":"2.0.0"}]`
		}
		images = append(images, fmt.Sprintf(`{"depName":"img-%d","currentValue":"1.0.0","currentVersion":"1.0.0","datasource":"docker","updates":%s}`, i, update))
	}
	for i := 0; i < 5; i++ {
		update := `[]`
		switch i {
		case 0:
			update = `[{"updateType":"patch","newVersion":"1.0.1","newValue":"1.0.1"}]`
		case 1:
			update = `[{"updateType":"major","newVersion":"2.0.0","newValue":"2.0.0"}]`
		}
		deps = append(deps, fmt.Sprintf(`{"depName":"dep-%d","currentValue":"1.0.0","currentVersion":"1.0.0","datasource":"go","updates":%s}`, i, update))
	}
	json := fmt.Sprintf(`{"repositories":{"devidence-dev/QuantWarden":{"packageFiles":{"dockerfile":[{"packageFile":"Dockerfile","deps":[%s]},{"packageFile":"go.mod","deps":[%s]}]}}}}`,
		strings.Join(images, ","), strings.Join(deps, ","))
	rep, err := parseReport([]byte(json))
	if err != nil {
		t.Fatal(err)
	}
	data := buildDocumentData(rep, nil)

	wantImages := reportTab{Total: 3, WithUpdates: 1, UpToDate: 2, Distribution: []distItem{
		{Type: "patch", Class: "dist-patch"}, {Type: "minor", Count: 1, Class: "dist-minor"}, {Type: "major", Class: "dist-major"},
	}}
	wantDeps := reportTab{Total: 5, WithUpdates: 2, UpToDate: 3, Distribution: []distItem{
		{Type: "patch", Count: 1, Class: "dist-patch"}, {Type: "minor", Class: "dist-minor"}, {Type: "major", Count: 1, Class: "dist-major"},
	}}
	if got := data.Images; got.Total != wantImages.Total || got.WithUpdates != wantImages.WithUpdates || got.UpToDate != wantImages.UpToDate || !reflect.DeepEqual(got.Distribution, wantImages.Distribution) {
		t.Errorf("image tab totals = %+v, want %+v", got, wantImages)
	}
	if got := data.Deps; got.Total != wantDeps.Total || got.WithUpdates != wantDeps.WithUpdates || got.UpToDate != wantDeps.UpToDate || !reflect.DeepEqual(got.Distribution, wantDeps.Distribution) {
		t.Errorf("dependency tab totals = %+v, want %+v", got, wantDeps)
	}
}

func TestDocumentName(t *testing.T) {
	tests := []struct {
		name       string
		scannedAt  string
		wantName   string
		wantHeader string
	}{
		{name: "afternoon UTC", scannedAt: "2026-10-12T14:07:00Z", wantName: "dependency-report-2026-10-12.html", wantHeader: "2026-10-12 09:07 America/Guayaquil"},
		{name: "after midnight UTC is still the previous day in Guayaquil", scannedAt: "2026-10-13T02:30:00Z", wantName: "dependency-report-2026-10-12.html", wantHeader: "2026-10-12 21:30 America/Guayaquil"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertDocumentName(t, tt.scannedAt, tt.wantName, tt.wantHeader)
		})
	}
}

// assertDocumentName checks the file name and the header of the report received at scannedAt (RFC 3339).
func assertDocumentName(t *testing.T, scannedAt, wantName, wantHeader string) {
	t.Helper()
	at, err := time.Parse(time.RFC3339, scannedAt)
	if err != nil {
		t.Fatal(err)
	}
	name, err := documentName(at)
	if err != nil {
		t.Fatalf("documentName() error = %v", err)
	}
	if name != wantName {
		t.Errorf("documentName() = %q, want %q", name, wantName)
	}

	rep, err := parseReport([]byte(`{"repositories":{}}`))
	if err != nil {
		t.Fatal(err)
	}
	doc, err := buildReportDocument(rep, nil, at)
	if err != nil {
		t.Fatalf("buildReportDocument() error = %v", err)
	}
	if !strings.Contains(string(doc), wantHeader) {
		t.Errorf("document header does not contain %q", wantHeader)
	}
}

func TestBuildReportDocument(t *testing.T) {
	scannedAt := time.Date(2026, 10, 12, 14, 7, 0, 0, time.UTC)
	doc, err := buildReportDocument(loadFixture(t), []string{testForkRepo}, scannedAt)
	if err != nil {
		t.Fatalf("buildReportDocument() error = %v", err)
	}
	out := string(doc)

	// The panes are split at their ids: everything before the "deps" pane is "Imágenes".
	imagesAt := strings.Index(out, `id="images" role="tabpanel"`)
	depsAt := strings.Index(out, `id="deps" role="tabpanel"`)
	if imagesAt < 0 || depsAt < imagesAt {
		t.Fatalf("tab panes not found in document")
	}
	imagesPane, depsPane := out[imagesAt:depsAt], out[depsAt:]

	t.Run("tab labels", func(t *testing.T) {
		assertTabLabels(t, out)
	})

	t.Run("rows stay in their own tab", func(t *testing.T) {
		assertRowsStayInTab(t, imagesPane, depsPane)
	})

	t.Run("script in a dependency name is escaped", func(t *testing.T) {
		assertScriptEscaped(t, out)
	})

	t.Run("errors section", func(t *testing.T) {
		assertErrorsSection(t, out)
	})

	t.Run("quote in a version creates no attribute", func(t *testing.T) {
		assertQuoteEscaped(t, scannedAt)
	})

	t.Run("warning-level problems are not listed", func(t *testing.T) {
		assertWarningNotListed(t, scannedAt)
	})

	t.Run("all up to date renders zero updates and no rows", func(t *testing.T) {
		assertAllUpToDate(t, scannedAt)
	})
}

func assertTabLabels(t *testing.T, out string) {
	t.Helper()
	for _, label := range []string{">Imágenes</button>", ">Dependencias</button>"} { //nolint:misspell // Spanish tab label fixed by the spec
		if !strings.Contains(out, label) {
			t.Errorf("document has no tab %q", label)
		}
	}
}

func assertRowsStayInTab(t *testing.T, imagesPane, depsPane string) {
	t.Helper()
	if !strings.Contains(imagesPane, "ghcr.io/immich-app/immich-server:v1.119.0") || strings.Contains(depsPane, "ghcr.io/immich-app/immich-server") {
		t.Error("image row not in the Imágenes pane only")
	}
	if !strings.Contains(depsPane, "golang.org/x/net") || strings.Contains(imagesPane, "golang.org/x/net") {
		t.Error("dependency row not in the Dependencias pane only") //nolint:misspell // Spanish tab name in the message
	}
}

func assertScriptEscaped(t *testing.T, out string) {
	t.Helper()
	if !strings.Contains(out, "&lt;script&gt;alert(1)&lt;/script&gt;") {
		t.Error("escaped script not found")
	}
	if strings.Contains(out, "<script>alert(1)</script>") {
		t.Error("script rendered as markup")
	}
}

func assertErrorsSection(t *testing.T, out string) {
	t.Helper()
	if !strings.Contains(out, "devidence-dev/QuantWarden: Failed to look up dependency") {
		t.Error("lookup failure missing from the errors section")
	}
}

func assertQuoteEscaped(t *testing.T, scannedAt time.Time) {
	t.Helper()
	json := `{"repositories":{"devidence-dev/QuantWarden":{"packageFiles":{"dockerfile":[{"packageFile":"Dockerfile","deps":[
			{"depName":"nginx","currentValue":"1.0.0\" onmouseover=\"alert(1)","currentVersion":"1.0.0\" onmouseover=\"alert(1)","datasource":"docker","updates":[{"updateType":"patch","newVersion":"1.0.1","newValue":"1.0.1"}]}
		]}]}}}}`
	rep, err := parseReport([]byte(json))
	if err != nil {
		t.Fatal(err)
	}
	doc, err := buildReportDocument(rep, nil, scannedAt)
	if err != nil {
		t.Fatal(err)
	}
	out := string(doc)
	// The quotes are escaped, so the text stays inside data-copy and no attribute is created.
	if !strings.Contains(out, `data-copy="nginx:1.0.0&#34; onmouseover=&#34;alert(1)"`) {
		t.Error("copy button value not escaped inside its attribute")
	}
	if strings.Contains(out, `" onmouseover=`) || strings.Contains(out, ` onmouseover="`) {
		t.Error("version created an onmouseover attribute")
	}
}

func assertWarningNotListed(t *testing.T, scannedAt time.Time) {
	t.Helper()
	rep, err := parseReport([]byte(`{"repositories":{"devidence-dev/warn-only":{"problems":[{"level":40,"msg":"Just a warning"}]}}}`))
	if err != nil {
		t.Fatal(err)
	}
	doc, err := buildReportDocument(rep, nil, scannedAt)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(doc), "<h5>Errors</h5>") {
		t.Error("errors section shown for a level 40 problem")
	}
}

func assertAllUpToDate(t *testing.T, scannedAt time.Time) {
	t.Helper()
	var docker, others []string
	for i := 0; i < 12; i++ {
		docker = append(docker, fmt.Sprintf(`{"depName":"img-%02d","currentValue":"1","currentVersion":"1","datasource":"docker","updates":[]}`, i))
	}
	for i := 0; i < 30; i++ {
		others = append(others, fmt.Sprintf(`{"depName":"dep-%02d","currentValue":"1","currentVersion":"1","datasource":"go","updates":[]}`, i))
	}
	json := fmt.Sprintf(`{"repositories":{"devidence-dev/QuantWarden":{"packageFiles":{"dockerfile":[{"packageFile":"Dockerfile","deps":[%s]},{"packageFile":"go.mod","deps":[%s]}]}}}}`,
		strings.Join(docker, ","), strings.Join(others, ","))
	rep, err := parseReport([]byte(json))
	if err != nil {
		t.Fatal(err)
	}
	doc, err := buildReportDocument(rep, nil, scannedAt)
	if err != nil {
		t.Fatal(err)
	}
	out := string(doc)
	if strings.Contains(out, "<table") || strings.Contains(out, "<tr>") {
		t.Error("table rendered without rows")
	}
	for _, want := range []string{
		`<div class="metric-value" style="color: var(--accent-yellow);">0</div>`,
		`<div class="metric-value" style="color: var(--accent-blue);">12</div>`,
		`<div class="metric-value" style="color: var(--accent-blue);">30</div>`,
		"All dependencies are up to date",
	} {
		if strings.Count(out, want) < 1 {
			t.Errorf("document lacks %q", want)
		}
	}
}
