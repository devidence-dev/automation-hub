package depreport

import (
	"os"
	"testing"
	"time"

	"automation-hub/internal/messages"
)

// goldenDir is the golden output the handler tests pin (see internal/handlers/dependency_report_golden_test.go).
const goldenDir = "../../handlers/testdata/golden"

const testForkRepo = "devidence-dev/discord-tts-bot"

// fixedAt is when the hub receives the fixture in the tests: 2026-10-12 09:07 in Guayaquil.
var fixedAt = time.Date(2026, 10, 12, 14, 7, 0, 0, time.UTC)

func loadZone(t *testing.T, name string) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation(name)
	if err != nil {
		t.Fatal(err)
	}
	return loc
}

func catalogWith(t *testing.T, overrides map[string]any) *messages.Catalog {
	t.Helper()
	cat, err := messages.Load(overrides)
	if err != nil {
		t.Fatalf("messages.Load() error = %v", err)
	}
	return cat
}

// summaryAt aggregates rep with the test fork and the fixed receive time.
func summaryAt(rep Report) Summary {
	return Aggregate(rep, []string{testForkRepo}, fixedAt)
}

func readGolden(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(goldenDir + "/" + name) //nolint:gosec // G304: fixed golden files in testdata
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
