package depreport

import (
	"os"
	"testing"
)

// fixturePath is the Renovate report the handler tests use too.
const fixturePath = "../../handlers/testdata/renovate-report.json"

func loadFixtureReport(t *testing.T) Report {
	t.Helper()
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	rep, err := ParseReport(data)
	if err != nil {
		t.Fatalf("ParseReport() error = %v", err)
	}
	return rep
}

func TestParseReportFixture(t *testing.T) {
	rep := loadFixtureReport(t)
	if len(rep.Repositories) != 6 {
		t.Errorf("repositories = %d, want 6", len(rep.Repositories))
	}
	var found bool
	for _, file := range rep.Repositories["devidence-dev/QuantWarden"].PackageFiles["gomod"] {
		for _, dep := range file.Deps {
			if dep.DepName != nil && *dep.DepName == "golang.org/x/net" && dep.Datasource != nil && *dep.Datasource == "go" {
				found = true
			}
		}
	}
	if !found {
		t.Error("golang.org/x/net (datasource go) not found in QuantWarden gomod")
	}
}

func TestParseReportInvalidJSON(t *testing.T) {
	if _, err := ParseReport([]byte("{not json")); err == nil {
		t.Error("ParseReport() accepted invalid JSON")
	}
}

func TestReportFilter(t *testing.T) {
	rep := reportOf(t, `"o/r":{"packageFiles":{"helm":[{"packageFile":"a/on.yaml","deps":[]},{"packageFile":"a/off.yaml","deps":[]}]}}`)
	got := rep.Filter(func(_, file string) bool { return file == "a/off.yaml" })
	files := got.Repositories["o/r"].PackageFiles["helm"]
	if len(files) != 1 || files[0].PackageFile != "a/on.yaml" {
		t.Errorf("Filter() files = %+v, want only a/on.yaml", files)
	}
}
