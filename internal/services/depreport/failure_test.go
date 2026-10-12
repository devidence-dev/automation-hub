package depreport

import (
	"fmt"
	"strings"
	"testing"
)

const failurePod = "dependency-report-manual-1791763200-x7k2p"

func buildFailure(t *testing.T, overrides map[string]any, exitCode int, pod, logTail string) string {
	t.Helper()
	got, err := BuildFailure(catalogWith(t, overrides), exitCode, pod, logTail)
	if err != nil {
		t.Fatalf("BuildFailure() error = %v", err)
	}
	return got
}

// assertKeepsLastLines checks that a long log keeps its newest lines whole.
func assertKeepsLastLines(t *testing.T, pod string) {
	t.Helper()
	var lines []string
	for i := 0; i < 400; i++ {
		lines = append(lines, fmt.Sprintf("line %03d <err> & more text here", i))
	}
	got := buildFailure(t, nil, 2, pod, strings.Join(lines, "\n")) // about 12 KB
	if len(got) > telegramLimit {
		t.Fatalf("message has %d bytes", len(got))
	}
	if !strings.Contains(got, "line 399 &lt;err&gt; &amp; more text here</pre>") || strings.Contains(got, "line 000") {
		t.Errorf("tail not cut from the front: %q", got[:200])
	}
	body := got[strings.Index(got, "<pre>")+5 : len(got)-len("</pre>")]
	if first := strings.SplitN(body, "\n", 2)[0]; !strings.HasPrefix(first, "line ") {
		t.Errorf("first kept line is partial: %q", first)
	}
}

func TestBuildFailureMessageEscapesTail(t *testing.T) {
	got := buildFailure(t, nil, 1, failurePod, "<html>\nfoo & bar\n")
	want := "❌ Renovate falló (exit 1) en el pod " + failurePod + "\n<pre>&lt;html&gt;\nfoo &amp; bar</pre>"
	if got != want {
		t.Errorf("message = %q, want %q", got, want)
	}
}

func TestBuildFailureMessageKeepsLastLines(t *testing.T) {
	assertKeepsLastLines(t, failurePod)
}

func TestBuildFailureMessageHugeLineIsCut(t *testing.T) {
	got := buildFailure(t, nil, 1, failurePod, strings.Repeat("<", 10000))
	if len(got) > telegramLimit {
		t.Fatalf("message has %d bytes", len(got))
	}
	body := got[strings.Index(got, "<pre>")+5 : len(got)-len("</pre>")]
	if strings.ReplaceAll(body, "&lt;", "") != "" {
		t.Errorf("partial entity in %q", body[:20])
	}
}

func TestBuildFailureMessageEmptyTail(t *testing.T) {
	if got := buildFailure(t, nil, 1, failurePod, ""); got != "❌ Renovate falló (exit 1) en el pod "+failurePod {
		t.Errorf("message = %q", got)
	}
}

func TestBuildFailureMessageEscapesPod(t *testing.T) {
	got := buildFailure(t, nil, 1, "<x>", "")
	if want := "❌ Renovate falló (exit 1) en el pod &lt;x&gt;"; got != want {
		t.Errorf("message = %q, want %q", got, want)
	}
}

func TestEscapedSuffixKeepsWholeEntities(t *testing.T) {
	// "&gt;" costs 4 bytes: a budget of 4 keeps only it, and 9 keeps "<c>" whole as "&lt;c&gt;".
	if got := escapedSuffix("ab<c>", 4); got != "&gt;" {
		t.Errorf("escapedSuffix() = %q, want &gt;", got)
	}
	if got := escapedSuffix("ab<c>", 9); got != "&lt;c&gt;" {
		t.Errorf("escapedSuffix() = %q, want &lt;c&gt;", got)
	}
}

func TestFailureTitleOverride(t *testing.T) {
	got := buildFailure(t, map[string]any{"failure": map[string]any{"title": "❌ Renovate failed (exit {exit}) in pod {pod}"}}, 1, failurePod, "")
	if want := "❌ Renovate failed (exit 1) in pod " + failurePod; got != want {
		t.Errorf("message = %q, want %q", got, want)
	}
}
