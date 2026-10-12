package depreport

import (
	"fmt"
	"reflect"
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
			name: "repo then file, the repo with a major first",
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
			note:  "fork · solo informativo", //nolint:misspell // Spanish text fixed by the spec
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
	if dep.Current != "0.30.0" || dep.New != "1.0.0" || dep.CurrentLabel != "Actual" || dep.NewLabel != "Nueva" || dep.TypeLabel != "major" {
		t.Errorf("dependency entry = %+v, want versions 0.30.0 to 1.0.0 with labels Actual and Nueva", dep)
	}
	if dep.Copy.Value != "golang.org/x/net@1.0.0" {
		t.Errorf("dependency copy = %q, want name@new version", dep.Copy.Value)
	}
	if label := dep.Copy.Label; label != "Copiar nombre@versión" {
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
			want:  SummaryLine{Bold: "2 actualizaciones", Rest: "para 2 dependencias"}, //nolint:misspell // Spanish text fixed by the spec
		},
		{
			name:  "one image",
			repos: []string{viewRepo("acme/a", viewFile("Dockerfile", viewDep("ghcr.io/x/y", "docker", "v1", "minor:v2")))},
			tab:   1,
			want:  SummaryLine{Bold: "1 actualización", Rest: "para 1 imagen"},
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
