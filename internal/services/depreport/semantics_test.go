package depreport

import (
	"regexp"
	"strings"
	"testing"
)

func nothingPendingDocument(t *testing.T) string {
	t.Helper()
	return presentDocument(t, nil, summaryAt(nothingPendingReport(t, 0)))
}

func documentMarkup(out string) string {
	return regexp.MustCompile(`(?s)<style.*?</style>|<script.*?</script>`).ReplaceAllString(out, "")
}

func TestDocumentSemantics(t *testing.T) {
	for name, out := range map[string]string{"fixture": presentDocument(t, nil, summaryAt(loadFixtureReport(t))), "nothing pending": nothingPendingDocument(t)} {
		t.Run(name, func(t *testing.T) { checkSemantics(t, documentMarkup(out)) })
	}
}

func checkSemantics(t *testing.T, markup string) {
	t.Helper()
	if n := len(regexp.MustCompile(`<h1(?:\s|>)`).FindAllString(markup, -1)); n != 1 {
		t.Errorf("h1 count = %d, want 1", n)
	}
	tags := regexp.MustCompile(`<[^>]+>`).FindAllString(markup, -1)
	ids := map[string]bool{}
	idRE := regexp.MustCompile(`\bid="([^"]+)"`)
	for _, tag := range tags {
		if m := idRE.FindStringSubmatch(tag); m != nil {
			ids[m[1]] = true
		}
	}
	for _, tag := range tags {
		checkSemanticTag(t, tag, ids)
	}
	for _, button := range regexp.MustCompile(`(?s)<button\b[^>]*>.*?</button>`).FindAllString(markup, -1) {
		opening := button[:strings.Index(button, ">")+1]
		body := regexp.MustCompile(`<[^>]+>`).ReplaceAllString(button[len(opening):len(button)-len("</button>")], "")
		if strings.TrimSpace(body) == "" && !regexp.MustCompile(`aria-label="[^"]+"|aria-labelledby="[^"]+"`).MatchString(opening) {
			t.Errorf("button lacks accessible name: %s", opening)
		}
	}
}

func checkSemanticTag(t *testing.T, tag string, ids map[string]bool) {
	t.Helper()
	if strings.HasPrefix(tag, "<button ") && !strings.Contains(tag, `type="`) {
		t.Errorf("button lacks type: %s", tag)
	}
	if strings.HasPrefix(tag, "<svg ") && !strings.Contains(tag, `aria-hidden="true"`) && (!strings.Contains(tag, `role="img"`) || (!strings.Contains(tag, `aria-label="`) && !strings.Contains(tag, `aria-labelledby="`))) {
		t.Errorf("SVG is neither hidden nor named: %s", tag)
	}
	for _, attr := range []string{"aria-controls", "aria-labelledby"} {
		checkReference(t, tag, attr, ids)
	}
}

func checkReference(t *testing.T, tag, attr string, ids map[string]bool) {
	t.Helper()
	re := regexp.MustCompile(`\b` + attr + `="([^"]+)"`)
	if m := re.FindStringSubmatch(tag); m != nil {
		for _, id := range strings.Fields(m[1]) {
			if !ids[id] {
				t.Errorf("%s target %q absent", attr, id)
			}
		}
	}
}

func TestTabListIsNamed(t *testing.T) {
	for _, tc := range []struct{ value, want string }{{"", "Report sections"}, {"Secciones del informe", "Secciones del informe"}, {`Sections "A" & <B>`, `Sections &#34;A&#34; &amp; &lt;B&gt;`}} {
		var overrides map[string]any
		if tc.value != "" {
			overrides = map[string]any{"document": map[string]any{"tabs": map[string]any{"label": tc.value}}}
		}
		out := presentDocument(t, overrides, summaryAt(loadFixtureReport(t)))
		tag := regexp.MustCompile(`<div[^>]*role="tablist"[^>]*>`).FindString(out)
		if !strings.Contains(tag, `aria-label="`+tc.want+`"`) {
			t.Errorf("tablist = %s, want %q", tag, tc.want)
		}
	}
}

func TestDocumentLandmarksAndLists(t *testing.T) {
	for name, out := range map[string]string{"fixture": presentDocument(t, nil, summaryAt(loadFixtureReport(t))), "nothing": nothingPendingDocument(t)} {
		t.Run(name, func(t *testing.T) { checkLandmarks(t, documentMarkup(out), name == "fixture") })
	}
}

func checkLandmarks(t *testing.T, markup string, fixture bool) {
	t.Helper()
	if strings.Count(markup, "<main>") != 1 || strings.Count(markup, "</main>") != 1 {
		t.Fatal("want one main landmark")
	}
	main := markup[strings.Index(markup, "<main>"):strings.Index(markup, "</main>")]
	if strings.Contains(main, "<header") || !strings.Contains(main, `role="tablist"`) || strings.Count(main, `role="tabpanel"`) != 3 {
		t.Error("header or tab landmarks are misplaced")
	}
	statuses := 1
	if fixture {
		statuses = 2
		checkFixtureLists(t, main, markup)
	} else if strings.Contains(main, `<ul role="list">`) {
		t.Error("nothing-pending document has entry lists")
	}
	if n := strings.Count(markup, `role="status"`); n != statuses {
		t.Errorf("status regions = %d, want %d", n, statuses)
	}
	if strings.Contains(markup, "aria-live=") {
		t.Error("redundant aria-live")
	}
}

func checkFixtureLists(t *testing.T, main, markup string) {
	t.Helper()
	if !strings.Contains(main, `id="errors-heading"`) {
		t.Error("errors card outside main")
	}
	lists := regexp.MustCompile(`(?s)<ul role="list">(.*?)</ul>`).FindAllStringSubmatch(main, -1)
	want := []int{4, 45, 1, 2, 1, 1}
	if len(lists) != len(want) {
		t.Errorf("entry lists = %d, want %d", len(lists), len(want))
	}
	rows := 0
	for i, list := range lists {
		n := strings.Count(list[1], `<li class="dep-row"`)
		rows += n
		if i < len(want) && n != want[i] {
			t.Errorf("list %d has %d items, want %d", i, n, want[i])
		}
	}
	if rows != 54 {
		t.Errorf("entry list items = %d, want 54", rows)
	}
	dialog := regexp.MustCompile(`(?s)<dialog .*?</dialog>`).FindString(markup)
	if strings.Count(dialog, `role="status"`) != 1 {
		t.Error("dialog needs one status region")
	}
}
