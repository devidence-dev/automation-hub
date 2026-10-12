package depreport

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// The tokens of the liquid glass design (liquid-glass-report, openspec design D8), as the three blocks of
// assets/input.css declare them: light on :root, dark for the system choice and for a stored dark choice.
var lightTokens = map[string]string{
	"bg": "#EDF3F8", "ink": "#0B1B25", "ink-2": "#33475A", "muted": "#56697A", "accent": "#0A63C2",
	"major-fg": "#B93A24", "minor-fg": "#0A63C2", "patch-fg": "#07707F", "err-fg": "#C4141F", "ok": "#07707F",
	"logo-main": "#0B1B25", "logo-sub": "#334C5A", "minor": "#0A84FF", "major": "#FF6B57", "patch": "#16CBE7",
}

var darkTokens = map[string]string{
	"bg": "#04101A", "ink": "#EAF2F8", "ink-2": "#C9D8E2", "muted": "#9FB6C6", "accent": "#7FD9EA",
	"major-fg": "#FFA391", "minor-fg": "#8CC8FF", "patch-fg": "#7AF2FF", "err-fg": "#FF8A82", "ok": "#6CF5FF",
	"logo-main": "#F2F7FB", "logo-sub": "#9FB6C6",
}

// textTokens are the text colors, checked against their backgrounds at 4.5:1 (WCAG AA, normal text).
var textTokens = []string{"ink", "ink-2", "muted", "accent", "major-fg", "minor-fg", "patch-fg", "err-fg", "ok"}

var (
	cssTokenRE = regexp.MustCompile(`--([a-z0-9-]+):(#[0-9a-fA-F]{3,6})`)
	// A declaration, not a media feature: "(min-width:601px)" in a media query is the breakpoint.
	minWidthRE = regexp.MustCompile(`[;{]min-width:(\d+)px`)
)

// cssCompact drops the whitespace of a stylesheet, so the checks do not depend on formatting.
func cssCompact(css string) string {
	return strings.Join(strings.Fields(css), "")
}

// canonHex returns a color as lowercase #rrggbb, so #FFFFFF, #fff and #ffffff compare equal.
func canonHex(c string) string {
	c = strings.ToLower(c)
	if len(c) == 4 {
		return "#" + string(c[1]) + string(c[1]) + string(c[2]) + string(c[2]) + string(c[3]) + string(c[3])
	}
	return c
}

// cssTokens returns the color tokens declared in css, in canonical form.
func cssTokens(css string) map[string]string {
	tokens := map[string]string{}
	for _, m := range cssTokenRE.FindAllStringSubmatch(css, -1) {
		tokens[m[1]] = canonHex(m[2])
	}
	return tokens
}

// relLuminance is the WCAG 2.x relative luminance of a #rrggbb color.
func relLuminance(hex string) float64 {
	channel := func(s string) float64 {
		v, _ := strconv.ParseUint(s, 16, 8)
		c := float64(v) / 255
		if c <= 0.03928 {
			return c / 12.92
		}
		return math.Pow((c+0.055)/1.055, 2.4)
	}
	return 0.2126*channel(hex[1:3]) + 0.7152*channel(hex[3:5]) + 0.0722*channel(hex[5:7])
}

// contrastRatio is the WCAG 2.x contrast ratio of two #rrggbb colors.
func contrastRatio(a, b string) float64 {
	la, lb := relLuminance(a), relLuminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

// schemeBlocks returns the compiled CSS of the three token blocks: the light :root, the system dark block
// and the stored dark block.
func schemeBlocks(t *testing.T) (light, systemDark, storedDark string) {
	t.Helper()
	css := cssCompact(ReportCSS)
	dark := strings.Index(css, "@media(prefers-color-scheme:dark)")
	stored := strings.Index(css, `:root[data-theme=dark]`)
	if dark < 0 || stored < 0 {
		t.Fatal("compiled CSS lacks the dark token blocks")
	}
	return css[:dark], css[dark:stored], css[stored:]
}

func TestReportCSSLightTokens(t *testing.T) {
	light, _, _ := schemeBlocks(t)
	got := cssTokens(light)
	for name, value := range lightTokens {
		if g := got[name]; g != canonHex(value) {
			t.Errorf("light token --%s = %q, want %q", name, g, canonHex(value))
		}
	}
}

func TestReportCSSDarkTokensInBothBlocks(t *testing.T) {
	_, systemDark, storedDark := schemeBlocks(t)
	for block, css := range map[string]string{"system dark": systemDark, "stored dark": storedDark} {
		got := cssTokens(css)
		for name, value := range darkTokens {
			if g := got[name]; g != canonHex(value) {
				t.Errorf("%s token --%s = %q, want %q", block, name, g, canonHex(value))
			}
		}
	}
	if !strings.Contains(systemDark, `:root:not([data-theme=light])`) {
		t.Error("the system dark block does not exclude a stored light choice")
	}
}

func TestReportCSSTextContrast(t *testing.T) {
	light, systemDark, _ := schemeBlocks(t)
	lightTok, darkTok := cssTokens(light), cssTokens(systemDark)
	for _, name := range textTokens {
		// The light text sits on the card (white) and on the page background.
		for _, base := range []string{"#ffffff", "#edf3f8"} {
			if ratio := contrastRatio(lightTok[name], base); ratio < 4.5 {
				t.Errorf("light %s on %s = %.2f:1, want at least 4.5:1", name, base, ratio)
			}
		}
		if ratio := contrastRatio(darkTok[name], darkTok["bg"]); ratio < 4.5 {
			t.Errorf("dark %s on %s = %.2f:1, want at least 4.5:1", name, darkTok["bg"], ratio)
		}
	}
}

func TestReportCSSUtilitiesAreEmitted(t *testing.T) {
	css := cssCompact(ReportCSS)
	for _, want := range []string{
		".glass{", ".pill{", ".app-bg{", ".tone-major{", ".tone-minor{", ".tone-patch{", ".tone-err{",
		".bar-seg{", ".swatch{", "details[open]>summary.chev", "rotate(180deg)",
		"@media(prefers-reduced-motion:reduce)", "::view-transition-old(root)", "[hidden]{display:none}",
	} {
		if !strings.Contains(css, want) {
			t.Errorf("compiled CSS lacks %s, so the rule is missing or was dropped", want)
		}
	}
}

func TestReportCSSRemovedTokens(t *testing.T) {
	css := strings.ToLower(cssCompact(ReportCSS))
	for _, gone := range []string{"--c-", "success", "brand-gradient", "linear-gradient(105deg"} {
		if strings.Contains(css, gone) {
			t.Errorf("compiled CSS still has %q, which the liquid glass design removes", gone)
		}
	}
}

func TestReportCSSHasNoOldFrameworkOrWideMinWidth(t *testing.T) {
	css := strings.ToLower(cssCompact(ReportCSS))
	if strings.Contains(css, "--bs-") {
		t.Error("compiled CSS has the variables of the old framework")
	}
	for _, m := range minWidthRE.FindAllStringSubmatch(css, -1) {
		if px, _ := strconv.Atoi(m[1]); px > 320 {
			t.Errorf("compiled CSS sets min-width:%dpx, want at most 320px", px)
		}
	}
	if strings.Contains(css, "@container") || strings.Contains(css, "subgrid") {
		t.Error("compiled CSS uses container queries or subgrid, which the mock does not use")
	}
}
