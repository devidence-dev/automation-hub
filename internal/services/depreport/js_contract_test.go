package depreport

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

// Attributes that are intentionally advertised before the script uses them.
var unusedScriptAttributes = map[string]bool{}

func scriptContract(script, markup string) error {
	markup = documentMarkup(markup)
	calls := regexp.MustCompile(`(?s)\.(querySelectorAll|querySelector|closest|matches|getElementById)\(\s*(?:"([^"]*)"|'([^']*)')`).FindAllStringSubmatch(script, -1)
	if len(calls) == 0 {
		return fmt.Errorf("no DOM selectors found")
	}
	tags := regexp.MustCompile(`<([a-zA-Z][a-zA-Z0-9-]*)\b[^>]*>`).FindAllString(markup, -1)
	selectors := ""
	for _, call := range calls {
		selector := call[2]
		if selector == "" {
			selector = call[3]
		}
		if call[1] == "getElementById" {
			selector = "#" + selector
		}
		selectors += " " + selector
		if err := checkSelector(selector, tags); err != nil {
			return err
		}
	}
	for _, attr := range []string{"data-pill-group"} {
		if !strings.Contains(markup, attr) {
			return fmt.Errorf("expected attribute %s missing", attr)
		}
		if !unusedScriptAttributes[attr] && !strings.Contains(selectors, "["+attr+"]") {
			return fmt.Errorf("attribute %s is not used by a script selector", attr)
		}
	}
	return nil
}

func checkSelector(selector string, tags []string) error {
	for _, alternative := range strings.Split(selector, ",") {
		alternative = strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(alternative), "[open]", ""), ":checked", "")
		for _, part := range strings.Fields(alternative) {
			if !regexp.MustCompile(`^(?:[a-zA-Z][a-zA-Z0-9-]*)?(?:(?:[.#][a-zA-Z0-9_-]+)|(?:\[[a-zA-Z0-9-]+(?:=["']?[^\]"']+["']?)?\]))+$|^[a-zA-Z][a-zA-Z0-9-]*$`).MatchString(part) {
				return fmt.Errorf("unsupported selector %q", part)
			}
			if !matchesAny(part, tags) {
				return fmt.Errorf("selector %q has no matching markup for %q", selector, part)
			}
		}
	}
	return nil
}

func matchesAny(part string, tags []string) bool {
	for _, tag := range tags {
		if matchesPart(part, tag) {
			return true
		}
	}
	return false
}

func matchesPart(part, tag string) bool {
	name := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9-]*`).FindString(part)
	if name != "" && !strings.HasPrefix(tag, "<"+name+" ") && tag != "<"+name+">" {
		return false
	}
	for _, class := range regexp.MustCompile(`\.([a-zA-Z0-9_-]+)`).FindAllStringSubmatch(part, -1) {
		if !regexp.MustCompile(`\bclass="[^"]*\b` + regexp.QuoteMeta(class[1]) + `\b`).MatchString(tag) {
			return false
		}
	}
	for _, id := range regexp.MustCompile(`#([a-zA-Z0-9_-]+)`).FindAllStringSubmatch(part, -1) {
		if !strings.Contains(tag, `id="`+id[1]+`"`) {
			return false
		}
	}
	return matchesAttributes(part, tag)
}

func matchesAttributes(part, tag string) bool {
	attrs := regexp.MustCompile(`\[([a-zA-Z0-9-]+)(?:=(["']?)([^\]"']+)?)?["']?\]`).FindAllStringSubmatch(part, -1)
	for _, attr := range attrs {
		if !regexp.MustCompile(`\b` + regexp.QuoteMeta(attr[1]) + `(?:[=\s>])`).MatchString(tag) {
			return false
		}
		if attr[3] != "" && !strings.Contains(tag, attr[1]+`="`+attr[3]+`"`) {
			return false
		}
	}
	return true
}

func TestReportScriptContract(t *testing.T) {
	markup := presentDocument(t, nil, summaryAt(loadFixtureReport(t)))
	if err := scriptContract(reportScript, markup); err != nil {
		t.Fatal(err)
	}
	broken := strings.Replace(reportScript, `"[data-copy]"`, `"[data-copy-misspelled]"`, 1)
	if broken == reportScript {
		t.Fatal("negative case did not modify selector")
	}
	if err := scriptContract(broken, markup); err == nil {
		t.Error("misspelled selector passed contract")
	}
	broken = strings.Replace(reportScript, `".pill"`, `".pill.misspelled"`, 1)
	if broken == reportScript || scriptContract(broken, markup) == nil {
		t.Error("misspelled additional class passed contract")
	}
}
