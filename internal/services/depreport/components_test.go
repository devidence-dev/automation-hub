package depreport

import (
	"bytes"
	"context"
	"regexp"
	"strings"
	"testing"
	"time"

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
		Copy: CopyView{Value: "github.com/gorilla/mux@2.0.0", Label: "Copy name@version"},
	}
}

func TestEntryEscapesValues(t *testing.T) {
	e := muxEntry()
	e.Name = "<script>alert(1)</script>"
	e.Current = `1.0.0" onmouseover="alert(1)`
	e.Copy.Value = e.Name + "@2.0.0"
	out := renderOf(t, DepRow(e, rowLabels{Current: "Current", New: "New"}))
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
		if !strings.Contains(out, "chip "+class) {
			t.Errorf("TypeChip(%v) lacks class %q:\n%s", typ, class, out)
		}
	}
}

func TestRepoSummaryHasDotChipsAndNoEmoji(t *testing.T) {
	g := RepoGroupView{
		Repo: "devidence-dev/discord-tts-bot", Short: "discord-tts-bot", Note: "fork · informational only",
		Dot:   UpdateMinor,
		Chips: []CountChip{{Type: UpdateMinor, Text: "1 minor"}},
	}
	out := renderOf(t, RepoGroup(g, rowLabels{}))
	expectContains(t, out,
		[]string{`<details class="glass group`, "tone-minor", "discord-tts-bot", "fork · informational only", "1 minor", `<svg class="icon chev`},
		[]string{" open", ` name="`, "🔴", "🟡", "🚨", "⚠️", "♻️", "tone-major"})
}

func TestRepoGroupHasNoOpenOrName(t *testing.T) {
	if out := renderOf(t, RepoGroup(fixtureView(t).Tabs[0].Groups[0], rowLabels{})); strings.Contains(strings.SplitN(out, ">", 2)[0], "overflow-hidden") {
		t.Error("repo card clips its focus outline")
	}
	view := fixtureView(t)
	out := renderOf(t, RepoGroup(view.Tabs[0].Groups[0], view.RowLabels))
	if strings.Contains(out, "<details") && (strings.Contains(out, " open") || strings.Contains(out, ` name="`)) {
		t.Errorf("repo card opens or names itself:\n%s", out)
	}
}

func TestDepRowColumnsAndCopy(t *testing.T) {
	css := cssCompact(ReportCSS)
	if !strings.Contains(css, ".dep-row{grid-template-columns:minmax(0,1fr)auto") && !strings.Contains(css, "grid-template-columns:minmax(0,1fr)auto") {
		t.Error("two column layout missing")
	}
	if !strings.Contains(css, "@media(min-width:601px){.dep-row{grid-template-columns:minmax(0,1fr)auto58px30px") {
		t.Error("desktop columns missing")
	}
	out := renderOf(t, DepRow(muxEntry(), rowLabels{Current: "Current", New: "New"}))
	expectContains(t, out,
		[]string{`class="dep-row"`, `aria-label="Copy name@version"`, `data-copy="github.com/gorilla/mux@2.0.0"`, "tone-major", "major"},
		[]string{"tone-minor", "<table"})
}

func TestCopyButtonStates(t *testing.T) {
	out := renderOf(t, CopyButton(muxEntry().Copy))
	expectContains(t, out,
		[]string{`data-copy=`, "group-data-[state=copied]:hidden", "icon-btn"},
		[]string{"<label", "<a "})
	if n := strings.Count(out, "<svg"); n != 2 {
		t.Errorf("copy button has %d icons, want 2 (clipboard and check)", n)
	}
}

// TestExportDialogCornerClose checks the icon-only close button beside the title: it is the one dismiss button,
// it sits in the dialog form before the sections, and its name and title come from the catalog label that the
// bottom action shares.
func TestExportDialogCornerClose(t *testing.T) {
	for _, tt := range []struct {
		name      string
		overrides map[string]any
		want      string
	}{
		{"default", nil, "Close"},
		{"override", map[string]any{"document": map[string]any{"export": map[string]any{"close": "Cerrar"}}}, "Cerrar"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := &renderer{cat: catalogWith(t, tt.overrides)}
			view := buildDocumentView(r, loadZone(t, "America/Guayaquil"), summaryAt(loadFixtureReport(t)), "")
			if r.err != nil {
				t.Fatal(r.err)
			}
			checkCornerClose(t, renderOf(t, ExportDialog(view.Export, view.Labels)), tt.want)
		})
	}
}

// checkCornerClose checks the corner close button of a rendered dialog and the label of the bottom action.
func checkCornerClose(t *testing.T, out, want string) {
	t.Helper()
	tags := regexp.MustCompile(`<button\b[^>]*\sdata-export-dismiss(?:[\s=][^>]*)?>`).FindAllString(out, -1)
	if len(tags) != 1 {
		t.Fatalf("data-export-dismiss buttons = %d, want 1", len(tags))
	}
	start := strings.Index(out, tags[0])
	before := out[:start]
	if at := strings.LastIndex(before, "<form"); at < 0 || !strings.HasPrefix(out[at:], `<form method="dialog"`) || strings.Contains(before[at:], "</form>") {
		t.Error("the corner button is not inside the dialog form")
	}
	if strings.Index(out, "<fieldset") < start {
		t.Error("the corner button comes after the first section")
	}
	end := strings.Index(out[start:], "</button>")
	if end < 0 {
		t.Fatal("corner button unterminated")
	}
	expectContains(t, out[start:start+end], []string{`class="btn icon-btn"`, `type="submit"`, `aria-label="` + want + `"`, `title="` + want + `"`, iconMarkup["close"], `aria-hidden="true"`}, nil)
	if label := buttonLabel(out, tags[0]); label != "" {
		t.Errorf("corner button has text %q, want icon only", label)
	}
	if label := buttonLabel(out, exportButtonTag("data-export-close").FindString(out)); label != want {
		t.Errorf("bottom action label = %q, want %q", label, want)
	}
}

func TestIconsAreFixedAndInert(t *testing.T) {
	for _, name := range []string{"download", "close"} {
		if iconMarkup[name] == "" {
			t.Errorf("missing icon %s", name)
		}
	}
	for name := range iconMarkup {
		out := renderOf(t, Icon(name, "test-class"))
		expectContains(t, out, []string{`viewBox="0 0 14 14"`, `aria-hidden="true"`, `class="icon test-class"`}, []string{"<script", "<text", "href="})
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
		Eyebrow: "Dependency report", Big: "53", Sub: "pending out of 53 · 1 up to date",
		Segments: []BarSegment{{Type: UpdatePatch, Percent: "7.7%"}, {Type: UpdateMinor, Percent: "88.5%"}, {Type: UpdateMajor, Percent: "3.8%"}},
		Legend:   []LegendItem{{Type: UpdatePatch, Label: "patch", Count: "4"}, {Type: UpdateMinor, Label: "minor", Count: "46"}, {Type: UpdateMajor, Label: "major", Count: "2"}},
	}
	out := renderOf(t, StatsPanel(s))
	expectContains(t, out,
		[]string{"Dependency report", ">53<", "pending out of 53 · 1 up to date", `style="width:7.7%;"`, `style="width:88.5%;"`, `style="width:3.8%;"`, "patch", "<b class=\"font-mono text-xs font-medium text-ink\">4</b>", "tone-patch", "tone-major"},
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
	out := renderOf(t, SummaryText(SummaryLine{Bold: "52 updates", Rest: "across 52 dependencies"}))
	expectContains(t, out, []string{"<b", "52 updates</b>", "across 52 dependencies"}, nil)
	if out := renderOf(t, SummaryText(SummaryLine{Bold: "All images are up to date"})); strings.Contains(out, "across") {
		t.Errorf("up-to-date summary has a rest:\n%s", out)
	}
}

func TestErrorsCardRows(t *testing.T) {
	view := ErrorsView{
		Heading: "Scan errors",
		Rows:    []ErrorView{{Where: "root", Msg: "Fatal error outside any repo"}, {Where: "QuantWarden", Msg: "Failed to look up dependency"}},
		Chip:    "2 errors",
	}
	out := renderOf(t, ErrorsCard(view))
	expectContains(t, out, []string{"Scan errors", "2 errors", "root", "Fatal error outside any repo", "QuantWarden", "tone-err"}, []string{"<table"})
}

func TestTabsMarkup(t *testing.T) {
	view := fixtureView(t)
	out := renderOf(t, Tabs(view))
	expectContains(t, out,
		[]string{`role="tablist"`, `role="tab"`, `aria-selected="true"`, `aria-selected="false"`, `tabindex="0"`, `tabindex="-1"`, `role="tabpanel"`, "hidden"},
		[]string{"<input", `type="radio"`})
	deps := strings.Index(out, `data-tab="deps"`)
	images := strings.Index(out, `data-tab="images"`)
	charts := strings.Index(out, `data-tab="charts"`)
	if deps < 0 || images < deps || charts < images {
		t.Errorf("tabs are not Dependencies, Images, then Charts (deps=%d images=%d charts=%d)", deps, images, charts)
	}
	for _, tab := range regexp.MustCompile(`<button[^>]*role="tab"[^>]*>`).FindAllString(out, -1) {
		for _, variant := range []string{"narrow:flex-col", "narrow:gap-1", "narrow:px-1"} {
			if !strings.Contains(tab, variant) {
				t.Errorf("tab lacks %q: %s", variant, tab)
			}
		}
	}
	if !strings.Contains(out, `aria-selected="true"`) || strings.Index(out, `aria-selected="true"`) > images {
		t.Error("the deps tab is not the selected one")
	}
}

func TestThemeControlMarkup(t *testing.T) {
	out := renderOf(t, ThemeControl(chromeLabels{ThemeGroup: "Theme", ThemeSystem: "Auto", ThemeLight: "Light", ThemeDark: "Dark"}))
	expectContains(t, out,
		[]string{`role="group"`, `aria-label="Theme"`, `data-theme-set="system" title="Auto" aria-label="Auto" aria-pressed="true"`,
			`data-theme-set="light" title="Light" aria-label="Light" aria-pressed="false"`, `data-theme-set="dark" title="Dark" aria-label="Dark" aria-pressed="false"`},
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
	out := renderOf(t, FileGroup(FileGroupView{Path: "kubernetes/apps/x/values.yaml", Dir: "kubernetes/apps/", Short: "x/values.yaml"}, rowLabels{}))
	expectContains(t, out, []string{"kubernetes/apps/</span>x/values.yaml"}, nil)
}

func TestCopyConstantsOnce(t *testing.T) {
	out := presentDocument(t, nil, summaryAt(loadFixtureReport(t)))
	out = regexp.MustCompile(`(?s)<dialog .*?</dialog>|<script.*?</script>`).ReplaceAllString(out, "")
	for _, attr := range []string{"data-copied-text", "data-copied-before"} {
		if n := strings.Count(out, attr+"="); n != 1 {
			t.Errorf("%s occurs %d times outside dialog", attr, n)
		}
		for _, tag := range regexp.MustCompile(`<button[^>]*data-copy=[^>]*>`).FindAllString(out, -1) {
			if strings.Contains(tag, attr) {
				t.Errorf("copy button contains %s", attr)
			}
		}
	}
}

func TestHeaderTimeElement(t *testing.T) {
	for _, tt := range []struct{ zone, datetime, visible string }{
		{"America/Guayaquil", "2026-10-12T09:07:00-05:00", "2026-10-12 09:07 America/Guayaquil"},
		{"Asia/Kolkata", "2026-10-12T19:37:00+05:30", "2026-10-12 19:37 Asia/Kolkata"},
	} {
		t.Run(tt.zone, func(t *testing.T) {
			loc, err := time.LoadLocation(tt.zone)
			if err != nil {
				t.Fatal(err)
			}
			s := summaryAt(loadFixtureReport(t))
			s.ReceivedAt = time.Date(2026, 10, 12, 14, 7, 0, 0, time.UTC)
			view := buildDocumentView(&renderer{cat: catalogWith(t, nil)}, loc, s, "")
			out := renderOf(t, Header(view.Header))
			expectContains(t, out, []string{`datetime="` + tt.datetime + `" title="` + tt.visible + `">` + tt.visible + `</time>`}, nil)
		})
	}
}

func TestHeaderOrderAndAreas(t *testing.T) {
	out := renderOf(t, Header(fixtureView(t).Header))
	expectContains(t, out, []string{"grid-cols-[1fr_auto]", "[grid-template-areas:'logo_theme'_'meta_meta']", "min-[601px]:[grid-template-areas:'logo_meta_theme']", "min-[601px]:grid-cols-[auto_1fr_auto]", "min-w-0 flex-1 truncate", "min-[601px]:flex-initial", "glass-pill shrink-0"}, nil)
	logo, theme, stamp, export := strings.Index(out, "[grid-area:logo]"), strings.Index(out, "data-theme-set"), strings.Index(out, "<time"), strings.Index(out, "data-export-open")
	if logo < 0 || logo >= theme || theme >= stamp || stamp >= export {
		t.Errorf("incorrect header order: %d %d %d %d", logo, theme, stamp, export)
	}
}

func TestExportButtonLabelVisible(t *testing.T) {
	out := renderOf(t, ExportButton(fixtureView(t).Labels))
	expectContains(t, out, []string{`aria-label="Export"`, `<span>Export</span>`}, []string{"sr-only"})
}

func TestExportActionIcons(t *testing.T) {
	out := renderOf(t, ExportDialog(fixtureView(t).Export, fixtureView(t).Labels))
	for _, tt := range []struct{ attr, label, path string }{
		{"data-export-copy", "Copy YAML", iconMarkup["copy"]}, {"data-export-download", "Download YAML", iconMarkup["download"]}, {"data-export-close", "Close", iconMarkup["close"]},
	} {
		t.Run(tt.attr, func(t *testing.T) {
			tag := exportButtonTag(tt.attr).FindString(out)
			start := strings.Index(out, tag)
			if tag == "" {
				t.Fatal("button missing")
			}
			end := strings.Index(out[start:], "</button>")
			if end < 0 {
				t.Fatal("button unterminated")
			}
			button := out[start : start+end]
			expectContains(t, button, []string{"inline-flex items-center gap-1.5", `aria-hidden="true"`, tt.path, tt.label}, nil)
			if strings.Index(button, tt.path) > strings.Index(button, tt.label) {
				t.Error("icon follows label")
			}
			if tt.attr == "data-export-copy" {
				expectContains(t, button, []string{iconMarkup["check"], "group-data-[state=copied]:hidden", "hidden group-data-[state=copied]:block"}, nil)
			} else if strings.Count(button, "<svg") != 1 {
				t.Error("expected one icon")
			}
		})
	}
}
