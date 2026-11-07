package depreport

import (
	"regexp"
	"testing"
)

func TestDocumentWeight(t *testing.T) {
	fixture := presentDocument(t, nil, summaryAt(loadFixtureReport(t)))
	empty := nothingPendingDocument(t)
	rows := regexp.MustCompile(`(?s)<li class="dep-row".*?</button></li>`).FindAllString(fixture, -1)
	if len(rows) != 54 {
		t.Fatalf("fixture has %d entry rows, want 54", len(rows))
	}
	bytes := 0
	for _, row := range rows {
		bytes += len(row)
	}
	for _, c := range []struct {
		name       string
		got, limit int
	}{
		{"fixture document", len(fixture), 130000}, {"compiled CSS", len(ReportCSS), 22000}, {"average entry row", bytes / len(rows), 1000}, {"nothing-pending document", len(empty), 60000},
	} {
		t.Logf("%s: %d bytes (budget %d)", c.name, c.got, c.limit)
		if c.got > c.limit {
			t.Errorf("%s = %d bytes, exceeds budget %d", c.name, c.got, c.limit)
		}
	}
}
