package depreport

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// These selectors are emitted from words in template comments; task 2.1 removes them.
var knownUnusedClassSelectors = map[string]bool{}

// Known redundant markup markers, removed with their components in tasks 4.3 and 4.5.
var knownUnstyledMarkupClasses = map[string]bool{}
var scriptClassHooks = map[string]bool{"no-anim": true}

var classAttribute = regexp.MustCompile(`\bclass="([^"]*)"`)

// cssClasses reads selectors from rule preludes and decodes Tailwind's CSS escapes.
func cssClasses(css string) map[string]bool {
	classes := map[string]bool{}
	css = regexp.MustCompile(`/\*.*?\*/`).ReplaceAllString(css, "")
	start := 0
	for i := 0; i < len(css); i++ {
		switch css[i] {
		case '{':
			addPreludeClasses(classes, css[start:i])
			start = i + 1
		case '}':
			start = i + 1
		}
	}
	return classes
}

func addPreludeClasses(classes map[string]bool, prelude string) {
	if strings.HasPrefix(strings.TrimSpace(prelude), "@") {
		return
	}
	re := regexp.MustCompile(`\.(?:\\[0-9a-fA-F]{1,6} ?|\\.|[a-zA-Z0-9_-])+`)
	for _, raw := range re.FindAllString(prelude, -1) {
		classes[decodeCSSClass(raw[1:])] = true
	}
}

func decodeCSSClass(raw string) string {
	re := regexp.MustCompile(`\\(?:[0-9a-fA-F]{1,6} ?|.)`)
	return re.ReplaceAllStringFunc(raw, func(escape string) string {
		value := strings.TrimSpace(escape[1:])
		if n, err := strconv.ParseInt(value, 16, 32); err == nil {
			return string(rune(n))
		}
		return escape[1:]
	})
}

func TestClassCoverage(t *testing.T) {
	out := presentDocument(t, nil, summaryAt(loadFixtureReport(t)))
	used := map[string]bool{}
	for _, m := range classAttribute.FindAllStringSubmatch(out, -1) {
		for _, c := range strings.Fields(m[1]) {
			used[c] = true
		}
	}
	selectors := cssClasses(ReportCSS)
	for c := range used {
		if !selectors[c] && !scriptClassHooks[c] && !knownUnstyledMarkupClasses[c] {
			t.Errorf("rendered class %q has no CSS selector", c)
		}
	}
	for c := range selectors {
		if !used[c] && !knownUnusedClassSelectors[c] && !scriptClassHooks[c] {
			t.Errorf("CSS class %q is unused by fixture", c)
		}
	}
}
