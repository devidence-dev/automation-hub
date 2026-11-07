package depreport

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

// exportOneImage is the "One image" scenario of the export spec, byte for byte, without its header line.
const exportOneImage = `  - repo: "devidence-dev/homelab"
    fork: false
    files:
      - path: "kubernetes/apps/immich/values.yaml"
        updates:
          - name: "ghcr.io/immich-app/immich-server"
            current: "v1.118.0"
            new: "v1.119.0"
            type: "minor"
`

func TestExportOneImage(t *testing.T) {
	rep := reportOf(t, viewRepo("devidence-dev/homelab",
		viewFile("kubernetes/apps/immich/values.yaml", viewDep("ghcr.io/immich-app/immich-server", "docker", "v1.118.0", "minor:v1.119.0"))))
	view := documentViewOf(t, summaryAt(rep))
	if len(view.Export.Sections) != 1 || view.Export.Sections[0].Key != "images" {
		t.Fatalf("sections = %+v, want one images section", view.Export.Sections)
	}
	got := view.Export.Header + "images:\n" + view.Export.Sections[0].Repos[0].YAML
	want := `scanned_at: "2026-10-12 09:07 America/Guayaquil"` + "\n" + "images:\n" + exportOneImage
	if got != want {
		t.Errorf("export =\n%s\nwant\n%s", got, want)
	}
}

func TestExportFork(t *testing.T) {
	tests := []struct {
		name string
		repo string
		want string
	}{
		{name: "fork repo", repo: testForkRepo, want: "    fork: true\n"},
		{name: "plain repo", repo: "devidence-dev/homelab", want: "    fork: false\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rep := reportOf(t, viewRepo(tt.repo, viewFile("go.mod", viewDep("x", "go", "1.0.0", "patch:1.0.1"))))
			view := documentViewOf(t, summaryAt(rep))
			if got := view.Export.Sections[0].Repos[0].YAML; !strings.Contains(got, tt.want) {
				t.Errorf("fragment lacks %q:\n%s", tt.want, got)
			}
		})
	}
}

func TestExportQuotedScalars(t *testing.T) {
	tests := []struct {
		name    string
		depName string
		current string
		want    string // the line the fragment must hold
	}{
		{
			name:    "quote and newline in a version stay one scalar",
			depName: "nginx",
			current: "1.0.0\"\n  injected: true",
			want:    `current: "1.0.0\"\n  injected: true"`,
		},
		{
			name:    "a script in a dependency name stays text",
			depName: "<script>alert(1)</script>",
			current: "1.0.0",
			want:    `name: "<script>alert(1)</script>"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rep := reportOf(t, viewRepo("acme/a", viewFile("go.mod", viewDep(tt.depName, "go", tt.current, "patch:1.0.1"))))
			view := documentViewOf(t, summaryAt(rep))
			if len(view.Export.Sections) != 1 {
				t.Fatalf("sections = %d, want 1", len(view.Export.Sections))
			}
			fragment := view.Export.Sections[0].Repos[0].YAML
			if !strings.Contains(fragment, tt.want) {
				t.Errorf("fragment lacks %s:\n%s", tt.want, fragment)
			}
			for _, line := range strings.Split(fragment, "\n") {
				if strings.HasPrefix(strings.TrimSpace(line), "injected:") {
					t.Errorf("a value added a key: %q", line)
				}
			}
		})
	}
}

func TestExportSections(t *testing.T) {
	tests := []struct {
		name string
		rep  Report
		keys []string
	}{
		{name: "fixture: images, then dependencies", rep: loadFixtureReport(t), keys: []string{"images", "dependencies"}},
		{
			name: "dependencies only",
			rep:  reportOf(t, viewRepo("acme/a", viewFile("go.mod", viewDep("x", "go", "1.0.0", "patch:1.0.1")))),
			keys: []string{"dependencies"},
		},
		{
			name: "nothing listed",
			rep:  reportOf(t, viewRepo("acme/a", viewFile("go.mod", viewDep("x", "go", "1.0.0")))),
			keys: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view := documentViewOf(t, summaryAt(tt.rep))
			if keys := sectionKeys(view.Export.Sections); !reflect.DeepEqual(keys, tt.keys) {
				t.Errorf("section keys = %q, want %q", keys, tt.keys)
			}
			for _, s := range view.Export.Sections {
				assertSectionOrder(t, view, s)
			}
		})
	}
}

// sectionKeys lists the keys of the sections, in order.
func sectionKeys(sections []ExportSection) []string {
	var keys []string
	for _, s := range sections {
		keys = append(keys, s.Key)
	}
	return keys
}

// assertSectionOrder checks that a section has the label of its tab and the repos of its tab's cards, in
// card order.
func assertSectionOrder(t *testing.T, view documentView, s ExportSection) {
	t.Helper()
	tab := view.Tabs[0]
	if s.Key == "images" {
		tab = view.Tabs[1]
	}
	var repos, cards []string
	for _, r := range s.Repos {
		repos = append(repos, r.Repo)
	}
	for _, g := range tab.Groups {
		cards = append(cards, g.Repo)
	}
	if !reflect.DeepEqual(repos, cards) || s.Label != tab.Label {
		t.Errorf("section %s = (%q, %v), want the tab (%q, %v)", s.Key, s.Label, repos, tab.Label, cards)
	}
}

func TestExportFileName(t *testing.T) {
	when := time.Date(2026, 10, 13, 2, 30, 0, 0, time.UTC) // 2026-10-12 21:30 in Guayaquil
	s := Aggregate(loadFixtureReport(t), []string{testForkRepo}, when)
	tests := []struct {
		name    string
		pattern string
		layout  string
		want    string
	}{
		{name: "default pattern", pattern: "dependency-updates-{date}.yaml", layout: "2006-01-02", want: "dependency-updates-2026-10-12.yaml"},
		{name: "configured pattern", pattern: "renovate-{date}.yaml", layout: "2006-01-02", want: "renovate-2026-10-12.yaml"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewDocumentPresenter(catalogWith(t, nil), loadZone(t, "America/Guayaquil"), tt.pattern, tt.layout)
			out, err := p.Present(s)
			if err != nil {
				t.Fatalf("Present() error = %v", err)
			}
			if !strings.Contains(string(out), `data-file-name="`+tt.want+`"`) {
				t.Errorf("document lacks data-file-name=%q", tt.want)
			}
		})
	}
}
