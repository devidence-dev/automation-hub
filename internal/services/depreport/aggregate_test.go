package depreport

import (
	"fmt"
	"strings"
	"testing"
)

// reportOf parses a report made of repo entries (see repoJSON).
func reportOf(t *testing.T, repos ...string) Report {
	t.Helper()
	rep, err := ParseReport([]byte(`{"repositories":{` + strings.Join(repos, ",") + `}}`))
	if err != nil {
		t.Fatalf("ParseReport() error = %v", err)
	}
	return rep
}

// repoJSON is one repo with one gomod dep per given update type, each dep with its own name.
func repoJSON(name string, types ...string) string {
	var deps []string
	for i, typ := range types {
		deps = append(deps, fmt.Sprintf(`{"depName":"dep-%d","datasource":"go","currentVersion":"1.0.0","updates":[{"updateType":%q,"newVersion":"2.0.0"}]}`, i, typ))
	}
	return fmt.Sprintf(`%q:{"packageFiles":{"gomod":[{"packageFile":"go.mod","deps":[%s]}]}}`, name, strings.Join(deps, ","))
}

func TestAggregateFixtureTotals(t *testing.T) {
	sum := Aggregate(loadFixtureReport(t), []string{"devidence-dev/discord-tts-bot"}, fixedAt)
	if sum.Repos != 6 {
		t.Errorf("Repos = %d, want 6", sum.Repos)
	}
	if sum.Images.WithUpdates != 1 || sum.Deps.WithUpdates != 52 {
		t.Errorf("deps with updates = %d images, %d deps; want 1 and 52", sum.Images.WithUpdates, sum.Deps.WithUpdates)
	}
	if sum.Images.UpToDate+sum.Deps.UpToDate != 5 {
		t.Errorf("up to date = %d, want 5", sum.Images.UpToDate+sum.Deps.UpToDate)
	}
	if len(sum.Errors) != 3 {
		t.Errorf("errors = %d, want 3", len(sum.Errors))
	}
}

func TestPendingOrder(t *testing.T) {
	rep := reportOf(t,
		repoJSON("devidence-dev/X", "major", "minor", "minor", "minor", "minor"),
		repoJSON("devidence-dev/Y", "major"),
		repoJSON("devidence-dev/A", "patch", "patch", "patch"),
		repoJSON("devidence-dev/B", "minor", "minor", "minor"),
		repoJSON("devidence-dev/Z", "patch", "patch"),
		repoJSON("devidence-dev/discord-tts-bot", "minor"),
	)
	sum := Aggregate(rep, []string{"devidence-dev/discord-tts-bot"}, fixedAt)
	var got []string
	for _, p := range sum.Pending {
		got = append(got, ShortName(p.Repo))
	}
	if want := "X Y A B Z discord-tts-bot"; strings.Join(got, " ") != want {
		t.Errorf("order = %v, want %s", got, want)
	}
	if last := sum.Pending[len(sum.Pending)-1]; !last.Fork {
		t.Error("discord-tts-bot is not marked as a fork")
	}
}

func TestAggregateDigestOnlyRepoHasNoPending(t *testing.T) {
	sum := Aggregate(reportOf(t, repoJSON("devidence-dev/digest-only", "digest")), nil, fixedAt)
	if len(sum.Pending) != 0 {
		t.Errorf("pending = %+v, want none", sum.Pending)
	}
	if sum.Deps.Total != 1 || sum.Deps.WithUpdates != 0 || sum.Deps.Updates != 0 {
		t.Errorf("deps tab = %+v, want one dep up to date", sum.Deps)
	}
}

func TestAggregateSplitsImagesFromDeps(t *testing.T) {
	rep := reportOf(t, `"devidence-dev/homelab":{"packageFiles":{"helm":[{"packageFile":"values.yaml","deps":[
		{"depName":"ghcr.io/immich-app/immich-server","datasource":"docker","currentVersion":"v1.118.0","updates":[{"updateType":"minor","newVersion":"v1.119.0"}]},
		{"depName":"golang.org/x/net","datasource":"go","currentVersion":"0.30.0","updates":[{"updateType":"major","newVersion":"1.0.0"}]},
		{"depName":"golang","datasource":"docker","currentVersion":"1.27","updates":[{"updateType":"patch","newVersion":"1.27.1"}]},
		{"depName":"legacy-tool","currentVersion":"1.0.0","updates":[{"updateType":"patch","newVersion":"1.0.1"}]}
	]}]}}`)
	sum := Aggregate(rep, nil, fixedAt)
	if len(sum.Images.Rows) != 2 || sum.Images.Rows[0].Name != "ghcr.io/immich-app/immich-server" || sum.Images.Rows[1].Name != "golang" {
		t.Errorf("images rows = %+v", sum.Images.Rows)
	}
	// A dep without a datasource is listed under dependencies, as the spec's Dockerfile scenario requires.
	if len(sum.Deps.Rows) != 2 || sum.Deps.Rows[0].Name != "golang.org/x/net" || sum.Deps.Rows[0].Type != UpdateMajor || sum.Deps.Rows[1].Name != "legacy-tool" {
		t.Errorf("deps rows = %+v", sum.Deps.Rows)
	}
}

func TestScanErrorsIncludeRoot(t *testing.T) {
	rep := reportOf(t, `"devidence-dev/QuantWarden":{"problems":[{"level":50,"msg":"Repo failed"},{"level":40,"msg":"Just a warning"}]}`)
	rep.Problems = []Problem{{Level: 50, Msg: "Fatal error outside any repo"}, {Level: 40, Msg: "Warning at root"}}
	sum := Aggregate(rep, nil, fixedAt)
	want := []ScanError{{Msg: "Fatal error outside any repo"}, {Repo: "devidence-dev/QuantWarden", Msg: "Repo failed"}}
	if fmt.Sprint(sum.Errors) != fmt.Sprint(want) {
		t.Errorf("errors = %+v, want %+v", sum.Errors, want)
	}
}

func TestShortName(t *testing.T) {
	tests := map[string]string{
		"acme-corp/payments-api": "payments-api",
		"payments-api":           "payments-api",
	}
	for in, want := range tests {
		if got := ShortName(in); got != want {
			t.Errorf("ShortName(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestOneDepTwoUpdates is one dep with a major and a minor update: the repo counts it once, as
// major, and the dependencies tab lists both rows, in either order Renovate gives them.
func TestOneDepTwoUpdates(t *testing.T) {
	tests := []struct {
		name    string
		updates string // the dep's update list, as Renovate writes it
	}{
		{name: "major first", updates: `{"updateType":"major","newVersion":"2.0.0"},{"updateType":"minor","newVersion":"1.9.0"}`},
		{name: "minor first", updates: `{"updateType":"minor","newVersion":"1.9.0"},{"updateType":"major","newVersion":"2.0.0"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dep := `{"depName":"github.com/gorilla/mux","datasource":"go","currentVersion":"1.8.0","updates":[` + tt.updates + `]}`
			rep := reportOf(t, `"devidence-dev/QuantWarden":{"packageFiles":{"gomod":[{"packageFile":"go.mod","deps":[`+dep+`]}]}}`)
			sum := Aggregate(rep, nil, fixedAt)
			if len(sum.Pending) != 1 || sum.Pending[0].Major != 1 || sum.Pending[0].Minor != 0 || sum.Pending[0].Patch != 0 {
				t.Errorf("pending = %+v, want one repo with one major dep", sum.Pending)
			}
			rows := sum.Deps.Rows
			if len(rows) != 2 || rows[0].Type != UpdateMajor || rows[1].Type != UpdateMinor {
				t.Errorf("dependency rows = %+v, want a major row, then a minor row", rows)
			}
			if sum.Deps.Total != 1 || sum.Deps.WithUpdates != 1 || sum.Deps.Updates != 2 {
				t.Errorf("deps tab = total %d, with updates %d, rows %d; want 1, 1, 2", sum.Deps.Total, sum.Deps.WithUpdates, sum.Deps.Updates)
			}
		})
	}
}
