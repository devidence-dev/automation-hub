package depreport

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

// viewDep is a Renovate dep. Each update is given as "type:newVersion"; datasource "" is left out.
func viewDep(name, datasource, current string, updates ...string) string {
	var ups []string
	for _, u := range updates {
		typ, next, _ := strings.Cut(u, ":")
		ups = append(ups, fmt.Sprintf(`{"updateType":%q,"newVersion":%q}`, typ, next))
	}
	ds := ""
	if datasource != "" {
		ds = fmt.Sprintf(`,"datasource":%q`, datasource)
	}
	return fmt.Sprintf(`{"depName":%q%s,"currentVersion":%q,"updates":[%s]}`, name, ds, current, strings.Join(ups, ","))
}

// viewFile is one package file of a repo with the given deps.
func viewFile(path string, deps ...string) string {
	return fmt.Sprintf(`{"packageFile":%q,"deps":[%s]}`, path, strings.Join(deps, ","))
}

// viewRepo is one repo entry with the given package files.
func viewRepo(name string, files ...string) string {
	return fmt.Sprintf(`%q:{"packageFiles":{"gomod":[%s]}}`, name, strings.Join(files, ","))
}

// groupsOf returns the repo sections of the images tab and of the dependencies tab.
func groupsOf(t *testing.T, repos ...string) (images, deps []RepoGroupView) {
	t.Helper()
	view := documentViewOf(t, summaryAt(reportOf(t, repos...)))
	return view.Tabs[1].Groups, view.Tabs[0].Groups
}

// layout describes the entries of the sections as "repo file: name type, ...", one line per file.
func layout(groups []RepoGroupView) []string {
	var out []string
	for _, g := range groups {
		for _, f := range g.Files {
			var entries []string
			for _, e := range f.Entries {
				entries = append(entries, e.Name+" "+e.Type.String())
			}
			out = append(out, g.Repo+" "+f.Path+": "+strings.Join(entries, ", "))
		}
	}
	return out
}

func TestGroupsLayout(t *testing.T) {
	tests := []struct {
		name  string
		repos []string
		deps  bool // true checks the dependencies tab, false the images tab
		want  []string
	}{
		{
			name: "repo with a major first, then files by path",
			repos: []string{
				viewRepo("acme/a",
					viewFile("go.mod", viewDep("x", "go", "1.0.0", "patch:1.0.1")),
					viewFile("sub/go.mod", viewDep("y", "go", "1.0.0", "minor:1.1.0"))),
				viewRepo("acme/b", viewFile("go.mod", viewDep("z", "go", "1.0.0", "major:2.0.0"))),
			},
			deps: true,
			want: []string{"acme/b go.mod: z major", "acme/a go.mod: x patch", "acme/a sub/go.mod: y minor"},
		},
		{
			name: "entries by type, files by path",
			repos: []string{viewRepo("acme/a",
				viewFile("a.mod", viewDep("p1", "go", "1", "minor:2"), viewDep("p2", "go", "1", "minor:2")),
				viewFile("b.mod", viewDep("q", "go", "1", "minor:2")),
				viewFile("c.mod", viewDep("r1", "go", "1", "patch:2"), viewDep("r2", "go", "1", "major:2")))},
			deps: true,
			want: []string{"acme/a a.mod: p1 minor, p2 minor", "acme/a b.mod: q minor", "acme/a c.mod: r2 major, r1 patch"},
		},
		{
			name: "two updates of one dep stay two entries, major first",
			repos: []string{viewRepo("acme/a", viewFile("go.mod",
				viewDep("github.com/gorilla/mux", "go", "1.8.0", "minor:1.9.0", "major:2.0.0")))},
			deps: true,
			want: []string{"acme/a go.mod: github.com/gorilla/mux major, github.com/gorilla/mux minor"},
		},
		{
			name: "entries of a file sort by name",
			repos: []string{viewRepo("acme/a", viewFile("go.mod",
				viewDep("b", "go", "1.0.0", "patch:1.0.1"), viewDep("a", "go", "1.0.0", "patch:1.0.1")))},
			deps: true,
			want: []string{"acme/a go.mod: a patch, b patch"},
		},
		{
			name: "docker deps only in the images tab",
			repos: []string{viewRepo("acme/a",
				viewFile("Dockerfile", viewDep("golang", "docker", "1.22", "minor:1.23")),
				viewFile("go.mod", viewDep("z", "go", "1.0.0", "patch:1.0.1")))},
			deps: false,
			want: []string{"acme/a Dockerfile: golang minor"},
		},
		{
			name: "a digest-only dep has no entry",
			repos: []string{viewRepo("acme/a", viewFile("go.mod",
				viewDep("z", "go", "1.0.0", "digest:"), viewDep("q", "go", "1.0.0")))},
			deps: true,
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			images, deps := groupsOf(t, tt.repos...)
			groups := images
			if tt.deps {
				groups = deps
			}
			if got := layout(groups); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("layout = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGroupsSummaryRow(t *testing.T) {
	tests := []struct {
		name    string
		repos   []string
		deps    bool
		dot     UpdateType
		short   string
		note    string
		chips   []string
		noEmoji bool
	}{
		{
			name:  "minor only: minor dot and one chip",
			repos: []string{viewRepo("acme/homelab", viewFile("go.mod", viewDep("a", "go", "1", "minor:2"), viewDep("b", "go", "1", "minor:2")))},
			deps:  true,
			dot:   UpdateMinor,
			short: "homelab",
			chips: []string{"2 minor"},
		},
		{
			name: "a dep with major and minor counts once, as major, like the caption",
			repos: []string{viewRepo("devidence-dev/QuantWarden", viewFile("go.mod",
				viewDep("github.com/gorilla/mux", "go", "1.8.0", "major:2.0.0", "minor:1.9.0"),
				viewDep("golang.org/x/net", "go", "0.30.0", "major:1.0.0"),
				viewDep("go.uber.org/zap", "go", "1.27.0", "patch:1.27.1")))},
			deps:  true,
			dot:   UpdateMajor,
			short: "QuantWarden",
			chips: []string{"2 major", "1 patch"},
		},
		{
			name:  "fork note on the summary row",
			repos: []string{viewRepo(testForkRepo, viewFile("go.mod", viewDep("a", "go", "1", "minor:2")))},
			deps:  true,
			dot:   UpdateMinor,
			short: "discord-tts-bot",
			note:  "fork · informational only",
			chips: []string{"1 minor"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			images, deps := groupsOf(t, tt.repos...)
			groups := images
			if tt.deps {
				groups = deps
			}
			if len(groups) != 1 {
				t.Fatalf("sections = %d, want 1", len(groups))
			}
			g := groups[0]
			var chips []string
			for _, c := range g.Chips {
				chips = append(chips, c.Text)
			}
			if g.Dot != tt.dot || g.Short != tt.short || g.Note != tt.note || !reflect.DeepEqual(chips, tt.chips) {
				t.Errorf("summary = (%v, %q, %q, %q), want (%v, %q, %q, %q)", g.Dot, g.Short, g.Note, chips, tt.dot, tt.short, tt.note, tt.chips)
			}
		})
	}
}

func TestGroupsSectionOrderAndFullName(t *testing.T) {
	images, deps := groupsOf(t,
		viewRepo("acme/patches", viewFile("go.mod", viewDep("x", "go", "1", "patch:2"), viewDep("y", "go", "1", "patch:2"), viewDep("w", "go", "1", "patch:2"))),
		viewRepo("acme/major", viewFile("go.mod", viewDep("z", "go", "1", "major:2"))),
		viewRepo("acme/mix", viewFile("Dockerfile", viewDep("golang", "docker", "1.22", "patch:1.22.1")),
			viewFile("go.mod", viewDep("q", "go", "1", "major:2"))),
	)
	if got, want := []string{deps[0].Repo, deps[1].Repo, deps[2].Repo}, []string{"acme/major", "acme/mix", "acme/patches"}; !reflect.DeepEqual(got, want) {
		t.Errorf("dependency sections = %q, want %q (major first, then the larger total, then the name)", got, want)
	}
	if n := len(images); n != 1 || images[0].Repo != "acme/mix" || len(images[0].Chips) != 1 || images[0].Dot != UpdatePatch {
		t.Errorf("image sections = %+v, want one acme/mix section with a patch dot and one chip", images)
	}
}

func TestGroupsPerTabCounts(t *testing.T) {
	// The repo has a patch in the images tab and a major in the dependencies tab: one section per tab.
	images, deps := groupsOf(t, viewRepo("acme/mix",
		viewFile("Dockerfile", viewDep("golang", "docker", "1.22", "patch:1.22.1")),
		viewFile("go.mod", viewDep("q", "go", "1", "major:2"))))
	if len(images) != 1 || len(images[0].Chips) != 1 || images[0].Chips[0].Text != "1 patch" {
		t.Errorf("images section chips = %+v, want 1 patch", images)
	}
	if len(deps) != 1 || deps[0].Dot != UpdateMajor || deps[0].Chips[0].Text != "1 major" {
		t.Errorf("dependencies section = %+v, want a major dot and 1 major", deps)
	}
}

func TestGroupsEntryFields(t *testing.T) {
	images, deps := groupsOf(t,
		viewRepo("acme/a", viewFile("Dockerfile", viewDep("ghcr.io/x/y", "docker", "v1.0.0", "minor:v1.1.0"))),
		viewRepo("acme/b", viewFile("go.mod", viewDep("golang.org/x/net", "go", "0.30.0", "major:1.0.0"))),
	)
	img := images[0].Files[0].Entries[0]
	if img.Name != "ghcr.io/x/y" || img.Current != "v1.0.0" || img.New != "v1.1.0" || img.TypeLabel != "minor" {
		t.Errorf("image entry = %+v, want the image name with its two tags and the minor label", img)
	}
	if img.Copy.Value != "ghcr.io/x/y:v1.1.0" {
		t.Errorf("image copy = %q, want the latest reference only", img.Copy.Value)
	}
	dep := deps[0].Files[0].Entries[0]
	if dep.Current != "0.30.0" || dep.New != "1.0.0" || dep.TypeLabel != "major" {
		t.Errorf("dependency entry = %+v, want versions 0.30.0 to 1.0.0 with labels Current and New", dep)
	}
	if dep.Copy.Value != "golang.org/x/net@1.0.0" {
		t.Errorf("dependency copy = %q, want name@new version", dep.Copy.Value)
	}
	if label := dep.Copy.Label; label != "Copy name@version" {
		t.Errorf("dependency copy label = %q", label)
	}
}

// TestSummaryLinesSingularAndPlural checks the line under the panel for one and for several updates, of
// dependencies and of images. The dependencies tab is Tabs[0] and the images tab Tabs[1].
func TestSummaryLinesSingularAndPlural(t *testing.T) {
	tests := []struct {
		name  string
		repos []string
		tab   int
		want  SummaryLine
	}{
		{
			name:  "several dependencies",
			repos: []string{viewRepo("acme/a", viewFile("go.mod", viewDep("x", "go", "1", "patch:2"), viewDep("y", "go", "1", "minor:2")))},
			tab:   0,
			want:  SummaryLine{Bold: "2 updates", Rest: "across 2 dependencies"},
		},
		{
			name:  "one image",
			repos: []string{viewRepo("acme/a", viewFile("Dockerfile", viewDep("ghcr.io/x/y", "docker", "v1", "minor:v2")))},
			tab:   1,
			want:  SummaryLine{Bold: "1 update", Rest: "across 1 image"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view := documentViewOf(t, summaryAt(reportOf(t, tt.repos...)))
			if got := view.Tabs[tt.tab].Summary; got != tt.want {
				t.Errorf("summary = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestSplitPath(t *testing.T) {
	tests := []struct{ path, dir, short string }{
		{"go.mod", "", "go.mod"},
		{".github/workflows/ci.yml", ".github/", "workflows/ci.yml"},
		{"kubernetes/apps/bitwarden/values.yaml", "kubernetes/apps/", "bitwarden/values.yaml"},
		{".github/ci.yml", "", ".github/ci.yml"},
	}
	for _, tt := range tests {
		if dir, short := splitPath(tt.path); dir != tt.dir || short != tt.short {
			t.Errorf("splitPath(%q) = (%q, %q), want (%q, %q)", tt.path, dir, short, tt.dir, tt.short)
		}
	}
}

// TestExportButton checks the header button: one button named Export before the theme control when there
// is something to export, and none otherwise.
func TestExportButton(t *testing.T) {
	tests := []struct {
		name string
		rep  Report
		want int // export buttons in the document
	}{
		{name: "updates listed", rep: loadFixtureReport(t), want: 1},
		{name: "nothing listed", rep: reportOf(t, viewRepo("acme/a", viewFile("go.mod", viewDep("x", "go", "1.0.0")))), want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := presentDocument(t, nil, summaryAt(tt.rep))
			// report.js names the attribute in its selector, so the count matches the button element only.
			if n := len(regexp.MustCompile(`<button[^>]*data-export-open`).FindAllString(out, -1)); n != tt.want {
				t.Fatalf("export buttons = %d, want %d", n, tt.want)
			}
			if tt.want == 0 {
				return
			}
			button := regexp.MustCompile(`<button[^>]*data-export-open[^>]*>`).FindString(out)
			buttonAt := strings.Index(out, button)
			expectContains(t, button,
				[]string{`type="button"`, `aria-haspopup="dialog"`, `aria-controls="export-dialog"`, `aria-label="Export"`,
					"glass-pill", "hover:[--glass-bg:var(--btn-hover)]"}, []string{"hover:text-accent", "hover:text-ink"})
			expectContains(t, out, []string{"min-[601px]:not-sr-only", ">Export</span>"}, nil)
			if buttonAt > strings.Index(out, "data-theme-set") {
				t.Error("export button is not before the theme control")
			}
		})
	}
}

// exportButtonTag matches the opening tag of the button that carries the data attribute attr as a whole name.
func exportButtonTag(attr string) *regexp.Regexp {
	// The optional group consumes a separator but never the closing >, so the match ends at this tag.
	return regexp.MustCompile(`<button\b[^>]*\s` + regexp.QuoteMeta(attr) + `(?:[\s=][^>]*)?>`)
}

// TestEveryButtonUsesBtn checks that each button of the fixture document carries the shared btn base, which
// gives the pointer cursor and the disabled state. The loop covers every button, so a new one is checked too.
func TestEveryButtonUsesBtn(t *testing.T) {
	out := presentDocument(t, nil, summaryAt(loadFixtureReport(t)))
	tags := regexp.MustCompile(`<button\b[^>]*>`).FindAllString(out, -1)
	if len(tags) == 0 {
		t.Fatal("the fixture document has no buttons")
	}
	classAttr := regexp.MustCompile(`\sclass="([^"]*)"`)
	for _, tag := range tags {
		m := classAttr.FindStringSubmatch(tag)
		if m == nil || !strings.Contains(" "+m[1]+" ", " btn ") {
			t.Errorf("button lacks the btn class: %s", tag)
		}
	}
}

// buttonLabel returns the text of the button after its opening tag up to </button>, with nested tags removed.
func buttonLabel(dialog, tag string) string {
	start := strings.Index(dialog, tag) + len(tag)
	end := strings.Index(dialog[start:], "</button>")
	if end < 0 {
		return ""
	}
	return strings.TrimSpace(regexp.MustCompile(`<[^>]+>`).ReplaceAllString(dialog[start:start+end], ""))
}

// TestExportActionButtons checks the three dialog actions. Each is found by its opening tag: it shares the
// btn class and the enabled-only hover, and keeps the data attribute that
// report.js reads. Copy and download are plain buttons; close is the submit button of the dialog form.
func TestExportActionButtons(t *testing.T) {
	out := presentDocument(t, nil, summaryAt(loadFixtureReport(t)))
	dialog := regexp.MustCompile(`(?s)<dialog .*?</dialog>`).FindString(out)
	shared := []string{"enabled:hover:bg-btn-hover", "btn"}
	absent := []string{" hover:bg-btn-hover", " hover:text-ink"}
	tests := []struct {
		attr, typ string
		// label is the button text without tags. The copy label holds two spans, the label and the copied text.
		label string
		want  []string
	}{
		{"data-export-copy", `type="button"`, "Copy YAMLCopied!", []string{`data-copied-text="Copied!"`}},
		{"data-export-download", `type="button"`, "Download YAML", nil},
		{"data-export-close", `type="submit"`, "Close", nil},
	}
	for _, tt := range tests {
		tag := exportButtonTag(tt.attr).FindString(dialog)
		if tag == "" {
			t.Errorf("no button opening tag with %s", tt.attr)
			continue
		}
		expectContains(t, tag, append(append([]string{`class="btn group `, tt.typ, tt.attr}, tt.want...), shared...), absent)
		if got := buttonLabel(dialog, tag); got != tt.label {
			t.Errorf("%s label = %q, want %q", tt.attr, got, tt.label)
		}
	}
	if at := strings.Index(dialog, `<form method="dialog"`); at < 0 || strings.Index(dialog, "data-export-close") < at {
		t.Error("the close button is not inside the dialog form")
	}
}

// TestExportDialogMarkup checks the dialog of the fixture: its attributes, its two sections in order, one
// checked checkbox per repo card and the three actions.
func TestExportDialogMarkup(t *testing.T) {
	s := summaryAt(loadFixtureReport(t))
	out := presentDocument(t, nil, s)
	if n := strings.Count(out, "<dialog"); n != 1 {
		t.Fatalf("dialogs = %d, want 1", n)
	}
	if strings.Index(out, "<dialog") < strings.Index(out, `data-panel="images"`) {
		t.Error("the dialog is not rendered after the tabs")
	}
	dialog := regexp.MustCompile(`(?s)<dialog .*?</dialog>`).FindString(out)
	expectContains(t, dialog, []string{
		`id="export-dialog"`, "data-export-dialog", `aria-labelledby="export-title"`,
		`data-file-name="dependency-updates-2026-10-12.yaml"`,
		`data-header="scanned_at: &#34;2026-10-12 09:07 America/Guayaquil&#34;` + "\n" + `"`,
		`<h2 id="export-title"`, "Export updates",
		"data-export-copy", `data-copied-text="Copied!"`, "Copy YAML",
		"data-export-download", "Download YAML",
		`<form method="dialog"`, "data-export-close", "Close",
	}, []string{"<script"})

	sections := regexp.MustCompile(`(?s)<fieldset data-export-section data-key="(\w+)".*?</fieldset>`).FindAllStringSubmatch(dialog, -1)
	wantSections := []struct{ key, label string }{{"images", "Images"}, {"dependencies", "Dependencies"}}
	if len(sections) != len(wantSections) {
		t.Fatalf("sections = %d, want %d", len(sections), len(wantSections))
	}
	for i, w := range wantSections {
		if sections[i][1] != w.key || !strings.Contains(sections[i][0], w.label) {
			t.Errorf("section %d = %q, want key %s with label %s", i, sections[i][1], w.key, w.label)
		}
	}
	if n := strings.Count(dialog, "data-export-all checked"); n != len(wantSections) {
		t.Errorf("checked section checkboxes = %d, want %d", n, len(wantSections))
	}

	view := documentViewOf(t, s)
	repos := regexp.MustCompile(`<input type="checkbox" data-export-repo[^>]*>`).FindAllString(dialog, -1)
	if want := len(view.Tabs[0].Groups) + len(view.Tabs[1].Groups); len(repos) != want {
		t.Fatalf("repo checkboxes = %d, want one per card (%d)", len(repos), want)
	}
	for _, in := range repos {
		if !strings.Contains(in, " checked") {
			t.Errorf("repo checkbox is not checked: %s", in)
		}
	}
}

// TestExportStatusRegion checks that the dialog holds its own live region. A modal makes the page outside it
// inert, so report.js announces the copies inside the open dialog.
func TestExportStatusRegion(t *testing.T) {
	out := presentDocument(t, nil, summaryAt(loadFixtureReport(t)))
	dialog := regexp.MustCompile(`(?s)<dialog .*?</dialog>`).FindString(out)
	statuses := regexp.MustCompile(`<[a-z]+ [^>]*role="status"[^>]*>`).FindAllString(dialog, -1)
	if len(statuses) != 1 {
		t.Fatalf("status regions in the dialog = %d, want 1", len(statuses))
	}
	expectContains(t, statuses[0], []string{"data-export-status", `class="sr-only"`}, []string{"aria-live="})
}

// TestExportCopyFeedback checks the copy action: its label swaps for the copied text while the button carries
// data-state="copied", as the entry copy button swaps its icon.
func TestExportCopyFeedback(t *testing.T) {
	out := presentDocument(t, nil, summaryAt(loadFixtureReport(t)))
	dialog := regexp.MustCompile(`(?s)<dialog .*?</dialog>`).FindString(out)
	copyBtn := regexp.MustCompile(`(?s)<button\b[^>]*data-export-copy[^>]*>.*?</button>`).FindString(dialog)
	expectContains(t, copyBtn, []string{
		`<span class="group-data-[state=copied]:hidden">Copy YAML</span>`,
		`<span class="hidden group-data-[state=copied]:inline">Copied!</span>`,
	}, nil)
}

// TestHeaderWraps checks that the header and its right-hand group wrap, so at 320px no part overflows the row.
func TestHeaderWraps(t *testing.T) {
	out := presentDocument(t, nil, summaryAt(loadFixtureReport(t)))
	start, end := strings.Index(out, "<header"), strings.Index(out, "</header>")
	if start < 0 || end < start {
		t.Fatal("header not found in document")
	}
	header := out[start:end]
	expectContains(t, header, []string{`<header class="flex flex-wrap items-center justify-between gap-x-4 gap-y-2 px-1 pt-1">`, `<div class="flex flex-wrap items-center gap-2">`, "data-theme-set", "data-export-open", "glass-pill"}, nil)
}
