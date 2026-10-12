package handlers

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

// fixtureCaption is the caption of testdata/renovate-report.json with the fork of the handler
// tests, received on fixedClock. Checked by hand against the fixture: QuantWarden has two distinct
// majors (golang.org/x/net is listed twice and counts once), so 53 deps have updates (1 image,
// 52 others), 5 are up to date, and there are 3 errors (2 in QuantWarden, 1 at the root).
const fixtureCaption = "📦 <b>Dependency report</b> · 2026-10-12\n\n" +
	"🔴 QuantWarden — 2 major · 1 patch\n" +
	"🟡 big-monorepo — 45 minor\n" +
	"🟡 homelab — 1 minor · 3 patch\n" +
	"🟡 discord-tts-bot (fork) — 1 minor\n" +
	"✅ 2 repos al día\n\n" +
	"📊 53 desactualizadas · 5 al día\n" +
	"🐳 Imágenes: 1 · 📚 Paquetes: 52\n" +
	"⚠️ 3 errores en el scan"

// reportWith parses a report made of repo entries (see repoWithUpdates).
func reportWith(t *testing.T, repos ...string) renovateReport {
	t.Helper()
	rep, err := parseReport([]byte(`{"repositories":{` + strings.Join(repos, ",") + `}}`))
	if err != nil {
		t.Fatalf("parseReport() error = %v", err)
	}
	return rep
}

// repoWithUpdates returns one repo entry with one dep per type in types ("" is a dep without updates).
func repoWithUpdates(name string, types ...string) string {
	deps := make([]string, 0, len(types))
	for i, typ := range types {
		updates := "[]"
		if typ != "" {
			updates = fmt.Sprintf(`[{"updateType":"%s","newVersion":"2.0.0","newValue":"2.0.0"}]`, typ)
		}
		deps = append(deps, fmt.Sprintf(`{"depName":"dep-%d","currentValue":"1.0.0","currentVersion":"1.0.0","datasource":"go","updates":%s}`, i, updates))
	}
	return fmt.Sprintf(`"%s":{"packageFiles":{"gomod":[{"packageFile":"go.mod","deps":[%s]}]}}`, name, strings.Join(deps, ","))
}

// buildCaptionFor renders the caption received on fixedClock.
func buildCaptionFor(t *testing.T, rep renovateReport, forks []string) string {
	t.Helper()
	caption, err := buildCaption(rep, forks, fixedClock)
	if err != nil {
		t.Fatalf("buildCaption() error = %v", err)
	}
	return caption
}

// captionRepoLinesOf returns the repo lines of a caption: the lines that start with 🔴 or 🟡.
func captionRepoLinesOf(caption string) []string {
	var lines []string
	for _, line := range strings.Split(caption, "\n") {
		if strings.HasPrefix(line, "🔴 ") || strings.HasPrefix(line, "🟡 ") {
			lines = append(lines, line)
		}
	}
	return lines
}

// assertCaptionText checks that the caption has every text in contains and none in absent.
func assertCaptionText(t *testing.T, caption string, contains, absent []string) {
	t.Helper()
	for _, want := range contains {
		if !strings.Contains(caption, want) {
			t.Errorf("caption lacks %q:\n%s", want, caption)
		}
	}
	for _, unwanted := range absent {
		if strings.Contains(caption, unwanted) {
			t.Errorf("caption has %q, want it omitted:\n%s", unwanted, caption)
		}
	}
}

func TestBuildCaptionFixture(t *testing.T) {
	scannedAt := time.Date(2026, 10, 12, 14, 7, 0, 0, time.UTC)
	got, err := buildCaption(loadFixture(t), []string{testForkRepo}, scannedAt)
	if err != nil {
		t.Fatalf("buildCaption() error = %v", err)
	}
	if got != fixtureCaption {
		t.Errorf("caption =\n%s\nwant\n%s", got, fixtureCaption)
	}
	if n := utf8.RuneCountInString(got); n > captionLimit {
		t.Errorf("caption has %d runes, over %d", n, captionLimit)
	}
}

func TestBuildCaptionRepoLines(t *testing.T) {
	tests := []struct {
		name         string
		repos        []string
		forks        []string
		wantLines    []string
		wantContains []string
		wantAbsent   []string
	}{
		{
			name:      "fork tag and counts",
			repos:     []string{repoWithUpdates("devidence-dev/discord-tts-bot", "minor")},
			forks:     []string{testForkRepo},
			wantLines: []string{"🟡 discord-tts-bot (fork) — 1 minor"},
		},
		{
			name:      "a dep with a minor and a major update counts once, as major",
			repos:     []string{`"devidence-dev/QuantWarden":{"packageFiles":{"gomod":[{"packageFile":"go.mod","deps":[{"depName":"github.com/gorilla/mux","currentValue":"1.8.0","updates":[{"updateType":"minor","newValue":"1.9.0"},{"updateType":"major","newValue":"2.0.0"}]}]}]}}`},
			wantLines: []string{"🔴 QuantWarden — 1 major"},
		},
		{
			name: "majors first, then total descending, then name",
			repos: []string{
				repoWithUpdates("devidence-dev/Z", "patch", "patch"),
				repoWithUpdates("devidence-dev/B", "minor", "minor", "minor"),
				repoWithUpdates("devidence-dev/Y", "major"),
				repoWithUpdates("devidence-dev/X", "major", "minor", "minor", "minor", "minor"),
				repoWithUpdates("devidence-dev/A", "patch", "patch", "patch"),
			},
			wantLines:  []string{"🔴 X — 1 major · 4 minor", "🔴 Y — 1 major", "🟡 A — 3 patch", "🟡 B — 3 minor", "🟡 Z — 2 patch"},
			wantAbsent: []string{"repos al día"},
		},
		{
			name:         "repos without pending updates are counted, not listed",
			repos:        []string{repoWithUpdates("devidence-dev/P", "minor"), repoWithUpdates("devidence-dev/Q", "")},
			wantLines:    []string{"🟡 P — 1 minor"},
			wantContains: []string{"\n✅ 1 repos al día\n"},
		},
		{
			name:         "digest only is nothing pending",
			repos:        []string{repoWithUpdates("devidence-dev/digest", "digest")},
			wantContains: []string{"✅ Todo al día"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			caption := buildCaptionFor(t, reportWith(t, tt.repos...), tt.forks)
			if got := captionRepoLinesOf(caption); !reflect.DeepEqual(got, tt.wantLines) {
				t.Errorf("repo lines = %q, want %q", got, tt.wantLines)
			}
			assertCaptionText(t, caption, tt.wantContains, tt.wantAbsent)
		})
	}
}

func TestBuildCaptionNothingPending(t *testing.T) {
	rep := reportWith(t, `"devidence-dev/digest-only":{"packageFiles":{"dockerfile":[{"packageFile":"Dockerfile","deps":[`+
		`{"depName":"alpine","currentValue":"3.20","datasource":"docker","updates":[{"updateType":"digest","newValue":"3.20"}]},`+
		`{"depName":"go","currentValue":"1.27","datasource":"go","updates":[]}]}]}}`)
	want := "📦 <b>Dependency report</b> · 2026-10-12\n\n✅ Todo al día\n\n📊 0 desactualizadas · 2 al día\n🐳 Imágenes: 0 · 📚 Paquetes: 0"
	if got := buildCaptionFor(t, rep, nil); got != want {
		t.Errorf("caption = %q, want %q", got, want)
	}
}

func TestBuildCaptionEscapesNames(t *testing.T) {
	caption := buildCaptionFor(t, reportWith(t, repoWithUpdates("devidence-dev/o<r>", "patch")), nil)
	if got := captionRepoLinesOf(caption); !reflect.DeepEqual(got, []string{"🟡 o&lt;r&gt; — 1 patch"}) {
		t.Errorf("repo lines = %q, want the escaped name", got)
	}
	if strings.Contains(caption, "<r>") {
		t.Errorf("caption has a raw tag from the repo name: %q", caption)
	}
}

func TestBuildCaptionOverflow(t *testing.T) {
	var repos []string
	for i := 0; i < 60; i++ {
		repos = append(repos, repoWithUpdates(fmt.Sprintf("devidence-dev/%s-%02d", strings.Repeat("long-repository-name", 2), i), "patch"))
	}
	caption := buildCaptionFor(t, reportWith(t, repos...), nil)
	if n := utf8.RuneCountInString(caption); n > captionLimit {
		t.Fatalf("caption has %d runes, over %d", n, captionLimit)
	}

	kept := captionRepoLinesOf(caption)
	dropped := len(repos) - len(kept)
	if dropped <= 0 {
		t.Fatalf("no repo line was dropped: %d kept of %d", len(kept), len(repos))
	}
	for _, line := range kept {
		if !strings.HasSuffix(line, " — 1 patch") {
			t.Errorf("kept line is cut: %q", line)
		}
	}
	overflow := fmt.Sprintf("… y %d repos más (ver reporte)", dropped)
	if !strings.Contains(caption, kept[len(kept)-1]+"\n"+overflow+"\n") {
		t.Errorf("overflow line %q does not follow the last kept repo line:\n%s", overflow, caption)
	}
	for _, want := range []string{"📦 <b>Dependency report</b> · 2026-10-12", "📊 60 desactualizadas · 0 al día", "🐳 Imágenes: 0 · 📚 Paquetes: 60"} {
		if !strings.Contains(caption, want) {
			t.Errorf("caption lacks %q", want)
		}
	}
}

func TestBuildDocumentDataPending(t *testing.T) {
	tests := []struct {
		name        string
		repos       []string
		forks       []string
		wantPending []repoPending
		wantRepos   int
	}{
		{
			name:        "highest update type counts once",
			repos:       []string{`"devidence-dev/QuantWarden":{"packageFiles":{"gomod":[{"packageFile":"go.mod","deps":[{"depName":"github.com/gorilla/mux","currentValue":"1.8.0","updates":[{"updateType":"minor","newValue":"1.9.0"},{"updateType":"major","newValue":"2.0.0"}]}]}]}}`},
			wantPending: []repoPending{{Repo: "devidence-dev/QuantWarden", Major: 1}},
			wantRepos:   1,
		},
		{
			name:        "fork flag and repos without updates",
			repos:       []string{repoWithUpdates(testForkRepo, "minor"), repoWithUpdates("devidence-dev/clean", "")},
			forks:       []string{testForkRepo},
			wantPending: []repoPending{{Repo: testForkRepo, Fork: true, Minor: 1}},
			wantRepos:   2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := buildDocumentData(reportWith(t, tt.repos...), tt.forks)
			if !reflect.DeepEqual(data.Pending, tt.wantPending) {
				t.Errorf("pending = %+v, want %+v", data.Pending, tt.wantPending)
			}
			if data.Repos != tt.wantRepos {
				t.Errorf("repos = %d, want %d", data.Repos, tt.wantRepos)
			}
		})
	}
}

func TestScanErrorsIncludeRoot(t *testing.T) {
	tests := []struct {
		name string
		json string
		want []string
	}{
		{
			name: "root and repo problems at level 50 and above",
			json: `{"problems":[{"level":50,"msg":"Fatal"},{"level":40,"msg":"warn"}],"repositories":{"a/b":{"problems":[{"level":50,"msg":"failed"}]}}}`,
			want: []string{"root: Fatal", "a/b: failed"},
		},
		{
			name: "warning level only",
			json: `{"problems":[{"level":40,"msg":"warn"}],"repositories":{}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rep, err := parseReport([]byte(tt.json))
			if err != nil {
				t.Fatal(err)
			}
			if got := scanErrors(rep); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("scanErrors() = %q, want %q", got, tt.want)
			}
		})
	}
}
