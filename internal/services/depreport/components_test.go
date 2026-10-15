package depreport

import (
	"bytes"
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

// renderOf renders a component into a string.
func renderOf(t *testing.T, c templ.Component) string {
	t.Helper()
	var buf bytes.Buffer
	if err := c.Render(context.Background(), &buf); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	return buf.String()
}

// expectContains fails for each want substring missing from out, and for each absent one present in out.
func expectContains(t *testing.T, out string, want, absent []string) {
	t.Helper()
	for _, w := range want {
		if !strings.Contains(out, w) {
			t.Errorf("output lacks %q:\n%s", w, out)
		}
	}
	for _, a := range absent {
		if strings.Contains(out, a) {
			t.Errorf("output has %q, want none:\n%s", a, out)
		}
	}
}

// inlineHandler matches an on* attribute, which the document must never carry.
var inlineHandler = regexp.MustCompile(`\son[a-z]+=`)

// fixtureView is the view of the fixture report, with the default texts.
func fixtureView(t *testing.T) documentView {
	t.Helper()
	return documentViewOf(t, summaryAt(loadFixtureReport(t)))
}

// muxEntry is one listed dependency update of the fixture kind.
func muxEntry() EntryView {
	return EntryView{
		Name: "github.com/gorilla/mux", Current: "1.8.0", New: "2.0.0", Type: UpdateMajor, TypeLabel: "major",
		CurrentLabel: "Actual", NewLabel: "Nueva",
		Copy: CopyView{Value: "github.com/gorilla/mux@2.0.0", Label: "Copiar nombre@versión", Copied: "¡Copiado!", CopiedBefore: "Copiar (copiado antes)"},
	}
}

func TestEntryEscapesValues(t *testing.T) {
	e := muxEntry()
	e.Name = "<script>alert(1)</script>"
	e.Current = `1.0.0" onmouseover="alert(1)`
	e.Copy.Value = e.Name + "@2.0.0"
	out := renderOf(t, DepRow(e))
	expectContains(t, out, []string{"&lt;script&gt;alert(1)&lt;/script&gt;", "1.0.0&#34; onmouseover=&#34;alert(1)"}, []string{"<script>alert(1)", "<script", `" onmouseover`})
}

func TestToneClasses(t *testing.T) {
	want := map[UpdateType]string{UpdateMajor: "tone-major", UpdateMinor: "tone-minor", UpdatePatch: "tone-patch", UpdateUnknown: ""}
	for typ, class := range want {
		if got := toneOf(typ); got != class {
			t.Errorf("toneOf(%v) = %q, want %q", typ, got, class)
		}
	}
	for typ, class := range want {
		if class == "" {
			continue
		}
		out := renderOf(t, TypeChip(typ, "x"))
		if !strings.Contains(out, class) {
			t.Errorf("TypeChip(%v) lacks class %q:\n%s", typ, class, out)
		}
	}
}

func TestRepoSummaryHasDotChipsAndNoEmoji(t *testing.T) {
	g := RepoGroupView{
		Repo: "devidence-dev/discord-tts-bot", Short: "discord-tts-bot", Note: "fork · solo informativo", //nolint:misspell // Spanish text fixed by the spec
		Dot:   UpdateMinor,
		Chips: []CountChip{{Type: UpdateMinor, Text: "1 minor"}},
	}
	out := renderOf(t, RepoGroup(g))
	expectContains(t, out,
		[]string{`<details class="glass group`, "tone-minor", "discord-tts-bot", "fork · solo informativo", "1 minor", `<svg class="chev`}, //nolint:misspell // Spanish text fixed by the spec
		[]string{" open", ` name="`, "🔴", "🟡", "🚨", "⚠️", "♻️", "tone-major"})
}

func TestRepoGroupHasNoOpenOrName(t *testing.T) {
	view := fixtureView(t)
	out := renderOf(t, RepoGroup(view.Tabs[0].Groups[0]))
	if strings.Contains(out, "<details") && (strings.Contains(out, " open") || strings.Contains(out, ` name="`)) {
		t.Errorf("repo card opens or names itself:\n%s", out)
	}
}

func TestDepRowColumnsAndCopy(t *testing.T) {
	out := renderOf(t, DepRow(muxEntry()))
	expectContains(t, out,
		[]string{"grid-cols-[minmax(0,1fr)_auto_58px_30px]", "min-[601px]:grid-cols-", `aria-label="Copiar nombre@versión"`, `data-copy="github.com/gorilla/mux@2.0.0"`, "tone-major", "major"},
		[]string{"tone-minor", "<table"})
}

func TestCopyButtonStates(t *testing.T) {
	out := renderOf(t, CopyButton(muxEntry().Copy))
	expectContains(t, out,
		[]string{`data-copied-text="¡Copiado!"`, `data-copied-before="Copiar (copiado antes)"`, "group-data-[state=copied]:hidden", "data-[state=copied]:text-ok"},
		[]string{"<label", "<a "})
	if n := strings.Count(out, "<svg"); n != 2 {
		t.Errorf("copy button has %d icons, want 2 (clipboard and check)", n)
	}
}

func TestIconsAreFixedAndInert(t *testing.T) {
	for name := range iconMarkup {
		out := renderOf(t, Icon(name, "size-3.5"))
		expectContains(t, out, []string{`viewBox="0 0 14 14"`, `aria-hidden="true"`, "size-3.5"}, []string{"<script", "<text", "href="})
		if inlineHandler.MatchString(out) {
			t.Errorf("icon %s has an inline handler:\n%s", name, out)
		}
	}
	if out := renderOf(t, Icon("missing", "")); !strings.Contains(out, "<svg") {
		t.Errorf("unknown icon renders no svg:\n%s", out)
	}
}

func TestStatsPanelValues(t *testing.T) {
	s := StatsView{
		Eyebrow: "Reporte de dependencias", Big: "53", Sub: "pendientes de 53 · 1 al día", //nolint:misspell // Spanish text fixed by the spec
		Segments: []BarSegment{{Type: UpdatePatch, Percent: "7.7%"}, {Type: UpdateMinor, Percent: "88.5%"}, {Type: UpdateMajor, Percent: "3.8%"}},
		Legend:   []LegendItem{{Type: UpdatePatch, Label: "patch", Count: "4"}, {Type: UpdateMinor, Label: "minor", Count: "46"}, {Type: UpdateMajor, Label: "major", Count: "2"}},
	}
	out := renderOf(t, StatsPanel(s))
	expectContains(t, out,
		[]string{"Reporte de dependencias", ">53<", "pendientes de 53 · 1 al día", `style="width:7.7%;"`, `style="width:88.5%;"`, `style="width:3.8%;"`, "patch", "<b class=\"font-mono text-xs font-medium text-ink\">4</b>", "tone-patch", "tone-major"}, //nolint:misspell // Spanish text fixed by the spec
		nil)
	if strings.Index(out, "width:7.7%") > strings.Index(out, "width:3.8%") {
		t.Error("bar segments are not in patch, minor, major order")
	}
}

func TestStatsPanelZeroSegmentOmittedLegendKept(t *testing.T) {
	view := fixtureView(t)
	out := renderOf(t, Legend(view.Tabs[1].Stats.Legend))
	if strings.Count(out, "<i ") != 3 {
		t.Errorf("legend has %d swatches, want 3 with zeros kept:\n%s", strings.Count(out, "<i "), out)
	}
	if got := renderOf(t, DistributionBar(nil)); strings.Contains(got, "bar-seg") {
		t.Errorf("empty bar has segments:\n%s", got)
	}
}

func TestSummaryTextBoldAndRest(t *testing.T) {
	out := renderOf(t, SummaryText(SummaryLine{Bold: "52 actualizaciones", Rest: "para 52 dependencias"})) //nolint:misspell // Spanish text fixed by the spec
	expectContains(t, out, []string{"<b", "52 actualizaciones</b>", "para 52 dependencias"}, nil)          //nolint:misspell // Spanish text fixed by the spec
	if out := renderOf(t, SummaryText(SummaryLine{Bold: "Todas las imágenes están al día"})); strings.Contains(out, "para") {
		t.Errorf("up-to-date summary has a rest:\n%s", out)
	}
}

func TestErrorsCardRows(t *testing.T) {
	view := documentView{
		Labels:     labelSet{ErrorsHeading: "Errores del análisis"},
		Errors:     []ErrorView{{Where: "raíz", Msg: "Fatal error outside any repo"}, {Where: "QuantWarden", Msg: "Failed to look up dependency"}},
		ErrorsChip: "2 errores",
	}
	out := renderOf(t, ErrorsCard(view))
	expectContains(t, out, []string{"Errores del análisis", "2 errores", "raíz", "Fatal error outside any repo", "QuantWarden", "tone-err"}, []string{"<table"})
}

func TestTabsMarkup(t *testing.T) {
	view := fixtureView(t)
	out := renderOf(t, Tabs(view))
	expectContains(t, out,
		[]string{`role="tablist"`, `role="tab"`, `aria-selected="true"`, `aria-selected="false"`, `tabindex="0"`, `tabindex="-1"`, `role="tabpanel"`, "hidden"},
		[]string{"<input", `type="radio"`})
	deps := strings.Index(out, `data-tab="deps"`)
	images := strings.Index(out, `data-tab="images"`)
	if deps < 0 || images < 0 || deps > images {
		t.Errorf("tabs are not Dependencias then Imágenes (deps=%d images=%d)", deps, images) //nolint:misspell // Spanish text fixed by the spec
	}
	if !strings.Contains(out, `aria-selected="true"`) || strings.Index(out, `aria-selected="true"`) > images {
		t.Error("the deps tab is not the selected one")
	}
}

func TestThemeControlMarkup(t *testing.T) {
	out := renderOf(t, ThemeControl(labelSet{ThemeGroup: "Tema", ThemeSystem: "Automático", ThemeLight: "Claro", ThemeDark: "Oscuro"}))
	expectContains(t, out,
		[]string{`role="group"`, `aria-label="Tema"`, `data-theme-set="system" title="Automático" aria-label="Automático" aria-pressed="true"`,
			`data-theme-set="light" title="Claro" aria-label="Claro" aria-pressed="false"`, `data-theme-set="dark" title="Oscuro" aria-label="Oscuro" aria-pressed="false"`},
		nil)
}

func TestLogoInline(t *testing.T) {
	out := renderOf(t, Logo())
	expectContains(t, out, []string{`<svg`, `role="img"`, `aria-label="devidence.dev"`, "var(--logo-main)", "var(--logo-sub)"}, []string{"<metadata", "<img", "#0B1B25", "#F2F7FB"})
}

func TestFaviconIsDataURI(t *testing.T) {
	href := faviconHref()
	if !strings.HasPrefix(href, "data:image/svg+xml;base64,") {
		t.Errorf("favicon = %.40q, want a base64 svg data URI", href)
	}
}

// TestFileGroupPathHasNoGap checks the muted folder and the rest of a path render with no space between.
func TestFileGroupPathHasNoGap(t *testing.T) {
	out := renderOf(t, FileGroup(FileGroupView{Path: "kubernetes/apps/x/values.yaml", Dir: "kubernetes/apps/", Short: "x/values.yaml"}))
	expectContains(t, out, []string{"kubernetes/apps/</span>x/values.yaml"}, nil)
}
