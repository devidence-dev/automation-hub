package depreport

import (
	"strings"
	"testing"
)

// TestLogoAssetsAreClean checks the stripped sources that are inlined into the document: no script, no
// C2PA metadata and no inline handler, since the file is sent alone.
func TestLogoAssetsAreClean(t *testing.T) {
	for name, src := range map[string]string{"logo.svg": logoSVG, "icon.svg": iconSVG} {
		for _, bad := range []string{"<script", "<metadata", "c2pa"} {
			if strings.Contains(src, bad) {
				t.Errorf("%s contains %q", name, bad)
			}
		}
		if inlineHandler.MatchString(src) {
			t.Errorf("%s has an on* attribute", name)
		}
		if !strings.Contains(src, `role="img"`) || !strings.Contains(src, `aria-label="devidence.dev"`) {
			t.Errorf("%s lacks role=img or the accessible label", name)
		}
	}
}

// TestLogoWordmarkUsesTokens checks that the wordmark fills are the logo tokens, so the theme recolors
// the logo, and that no fixed wordmark color is left in the source.
func TestLogoWordmarkUsesTokens(t *testing.T) {
	if !strings.Contains(logoSVG, "fill:var(--logo-main)") || !strings.Contains(logoSVG, "fill:var(--logo-sub)") {
		t.Error("logo.svg does not take its wordmark fills from the logo tokens")
	}
	for _, hex := range []string{"#0B1B25", "#334C5A", "#F2F7FB", "#9FB6C6"} {
		if strings.Contains(logoSVG, hex) {
			t.Errorf("logo.svg still has the fixed wordmark color %s", hex)
		}
	}
	css := cssCompact(ReportCSS)
	for _, token := range []string{"--logo-main:#0b1b25", "--logo-main:#f2f7fb", "--logo-sub:#334c5a", "--logo-sub:#9fb6c6"} {
		if !strings.Contains(strings.ToLower(css), token) {
			t.Errorf("compiled CSS lacks the logo token %s", token)
		}
	}
}

// TestFaviconLinkIsNotSanitized checks that the favicon data URI survives templ's URL sanitizer.
func TestFaviconLinkIsNotSanitized(t *testing.T) {
	out := presentDocument(t, nil, summaryAt(loadFixtureReport(t)))
	if strings.Contains(out, "TemplFailedSanitizationURL") || !strings.Contains(out, `<link rel="icon" href="data:image/svg+xml;base64,`) {
		t.Error("the favicon data URI was sanitized away")
	}
}

// TestReportScriptEmbedded checks that the report script is the embedded report.js, inlined once at the
// end of the body, and that the document has exactly one inline script.
func TestReportScriptEmbedded(t *testing.T) {
	out := presentDocument(t, nil, summaryAt(loadFixtureReport(t)))
	if !strings.Contains(out, "<script>"+reportScript+"</script>") {
		t.Error("the document does not inline report.js in <script>")
	}
	if n := strings.Count(out, "<script"); n != 1 {
		t.Errorf("document has %d <script> elements, want 1 (report)", n)
	}
	if !strings.Contains(out, "</body>") || strings.Index(out, reportScript) < strings.Index(out, "</div>") {
		t.Error("report.js is not at the end of the body")
	}
}

// TestReportScriptHasNoStorage checks that report.js reads and writes no storage: the theme choice lasts
// for the page view only, so every open starts on Automático and no storage error can happen.
func TestReportScriptHasNoStorage(t *testing.T) {
	for _, api := range []string{"localStorage", "sessionStorage", "dv-theme"} {
		if strings.Contains(reportScript, api) {
			t.Errorf("report.js uses %s, the theme choice must not be stored", api)
		}
	}
}
