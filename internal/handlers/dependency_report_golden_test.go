package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap"
)

// goldenDir holds the output the hub sends for the fixture, captured before the refactor.
// Set UPDATE_GOLDEN=1 to rewrite the files after an intended change.
const goldenDir = "testdata/golden"

// goldenFailurePod is the pod name of the failure golden.
const goldenFailurePod = "dependency-report-manual-1791763200-x7k2p"

// upToDateFixture is the fixture with every update list emptied: the same repos and deps, none pending.
func upToDateFixture(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/renovate-report.json")
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	for _, repo := range doc["repositories"].(map[string]any) {
		for _, files := range repo.(map[string]any)["packageFiles"].(map[string]any) {
			for _, file := range files.([]any) {
				for _, dep := range file.(map[string]any)["deps"].([]any) {
					dep.(map[string]any)["updates"] = []any{}
				}
			}
		}
	}
	out, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// goldenOutputs maps each golden file to the output the hub sends for it.
func goldenOutputs(t *testing.T) map[string][]byte {
	t.Helper()
	sent := sendAll(t, validReport(t))
	upToDate := sendAll(t, reportRequest{
		token:  testToken,
		fields: map[string]string{"exit_code": "0", "pod": "p"},
		files:  map[string]string{"report": string(upToDateFixture(t))},
	})
	failure := sendAll(t, reportRequest{
		token:  testToken,
		fields: map[string]string{"exit_code": "1", "pod": goldenFailurePod},
		files:  map[string]string{"report": "{}", "log_tail": "<html>\nfoo & bar\n"},
	})
	return map[string][]byte{
		"caption.txt":            []byte(sent.caption),
		"document.html":          sent.document,
		"caption_up_to_date.txt": []byte(upToDate.caption),
		"failure.txt":            []byte(failure.caption),
	}
}

// goldenSends is what one request sent: the caption (or failure text) and the document, if any.
type goldenSends struct {
	caption  string
	document []byte
}

// sendAll posts one request and collects its caption and document.
func sendAll(t *testing.T, req reportRequest) goldenSends {
	t.Helper()
	sender := &fakeSender{}
	rec := httptest.NewRecorder()
	newReportHandlerAt(t, sender, fixedClock, zap.NewNop()).ServeHTTP(rec, req.build(t))
	if rec.Code != 202 {
		t.Fatalf("status = %d, want 202 (%s)", rec.Code, rec.Body.String())
	}
	calls := sender.waitForCalls(t, 1)
	var out goldenSends
	for _, c := range calls {
		if c.kind == "document" {
			out.document = c.data
		}
		out.caption = c.text
	}
	return out
}

// TestGoldenBaseline pins the caption, the document and the failure message of the fixture.
func TestGoldenBaseline(t *testing.T) {
	outputs := goldenOutputs(t)
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.MkdirAll(goldenDir, 0o750); err != nil {
			t.Fatal(err)
		}
		for name, data := range outputs {
			if err := os.WriteFile(filepath.Join(goldenDir, name), data, 0o600); err != nil {
				t.Fatal(err)
			}
		}
	}
	for name, got := range outputs {
		want, err := os.ReadFile(filepath.Join(goldenDir, name)) //nolint:gosec // G304: fixed golden files in testdata
		if err != nil {
			t.Fatalf("read golden %s: %v", name, err)
		}
		if string(got) != string(want) {
			t.Errorf("%s differs from the golden file:\n%s", name, got)
		}
	}
}

// readGolden returns one golden file as text.
func readGolden(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(goldenDir, name)) //nolint:gosec // G304: fixed golden files in testdata
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
