package depreport

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// The tokens of the liquid glass design (liquid-glass-report, openspec design D8), as the three blocks of
// assets/input.css declare them: light on :root, dark for the system choice and for a stored dark choice.
var lightTokens = map[string]string{
	"bg": "#EDF3F8", "ink": "#0B1B25", "ink-2": "#33475A", "muted": "#56697A", "muted-2": "#5F6F7C", "count-on-bg": "#0A63C2", "count-on-fg": "#FFFFFF", "accent": "#0A63C2",
	"major-fg": "#B93A24", "minor-fg": "#0A63C2", "patch-fg": "#07707F", "err-fg": "#C4141F", "ok": "#07707F",
	"logo-main": "#0B1B25", "logo-sub": "#334C5A", "minor": "#0A84FF", "major": "#FF6B57", "patch": "#16CBE7",
}

var darkTokens = map[string]string{
	"bg": "#04101A", "ink": "#EAF2F8", "ink-2": "#C9D8E2", "muted": "#9FB6C6", "muted-2": "#7D94A4", "accent": "#7FD9EA",
	"major-fg": "#FFA391", "minor-fg": "#8CC8FF", "patch-fg": "#7AF2FF", "err-fg": "#FF8A82", "ok": "#6CF5FF",
	"logo-main": "#F2F7FB", "logo-sub": "#9FB6C6",
}

// textTokens are the text colors, checked against their backgrounds at 4.5:1 (WCAG AA, normal text).
var textTokens = []string{"ink", "ink-2", "muted", "muted-2", "accent", "major-fg", "minor-fg", "patch-fg", "err-fg", "ok"}

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
	checkAdditionalReportCSSTextContrast(t)
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
		".bar-seg{", ".chip{", ".bg-tone{", "details[open]>summary.chev", "rotate(180deg)",
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

// TestReportCSSExportDialogFitsPhone checks the export dialog width: 520px at most, and the viewport less
// the 16px gutters on each side below that.
func TestReportCSSExportDialogFitsPhone(t *testing.T) {
	css := cssCompact(ReportCSS)
	// The minifier writes min() without calc(): min() takes a sum, so both forms are the same width.
	if !strings.Contains(css, ".export-dialog{width:min(520px,100vw-32px)") {
		t.Error("compiled CSS lacks the export dialog width min(520px, 100vw - 32px)")
	}
}

// TestReportCSSExportDialogCentered checks that the dialog keeps the auto margin that centers it, since
// preflight zeroes margins, and that its height is the dynamic viewport less the gutters.
func TestReportCSSExportDialogCentered(t *testing.T) {
	css := cssCompact(ReportCSS)
	if !regexp.MustCompile(`\.export-dialog\{[^}]*[;{]margin:auto[;}]`).MatchString(css) {
		t.Error("the export dialog rule lacks margin:auto")
	}
	if !regexp.MustCompile(`\.export-dialog\{[^}]*[;{]max-height:calc\(100dvh-32px\)[;}]`).MatchString(css) {
		t.Error("the export dialog rule lacks max-height:calc(100dvh - 32px)")
	}
}

// TestReportCSSDialogScrollLock checks that an open export dialog locks the page behind it: the root hides its
// overflow and keeps a stable gutter, and the dialog keeps its overscroll to itself.
func TestReportCSSDialogScrollLock(t *testing.T) {
	css := cssCompact(ReportCSS)
	for _, want := range []string{`html:has\(\.export-dialog\[open\]\)\{(?:[^}]*;)?overflow:hidden[;}]`, `html:has\(\.export-dialog\[open\]\)\{(?:[^}]*;)?scrollbar-gutter:stable[;}]`} {
		if !regexp.MustCompile(want).MatchString(css) {
			t.Errorf("the open export dialog lacks the root rule %s", want)
		}
	}
	if !regexp.MustCompile(`\.export-dialog\{(?:[^}]*;)?overscroll-behavior:contain[;}]`).MatchString(css) {
		t.Error("the export dialog rule lacks overscroll-behavior:contain")
	}
}

// TestReportCSSBackdropFilterUnprefixed checks that the glass card and the dialog backdrop keep the unprefixed
// backdrop-filter. The minifier drops it when a hand-written -webkit- copy sits beside it, so the source
// declares only the unprefixed property and the build adds the prefix. The [;{] before the name rejects
// -webkit-backdrop-filter.
func TestReportCSSBackdropFilterUnprefixed(t *testing.T) {
	css := cssCompact(ReportCSS)
	for _, sel := range []string{".glass{", ".export-dialog::backdrop{"} {
		if !regexp.MustCompile(regexp.QuoteMeta(sel) + `[^}]*[;{]backdrop-filter:`).MatchString(css) {
			t.Errorf("compiled %s rule lacks the unprefixed backdrop-filter", sel)
		}
	}
}

// TestReportCSSDisabledActionHover checks that the action hover applies only to enabled buttons, so a disabled
// Copy or Download keeps its surface. The variant compiles to the :enabled:hover selector.
func TestReportCSSDisabledActionHover(t *testing.T) {
	if css := cssCompact(ReportCSS); !strings.Contains(css, `.enabled\:hover\:bg-btn-hover:enabled:hover{`) {
		t.Error("compiled CSS lacks the enabled-only hover selector of the export actions")
	}
}

// TestReportCSSButtonBase checks the shared button base: every button has the pointer cursor, and a disabled one
// the default cursor. The btn utility compiles to .btn and its nested :disabled rule.
func TestReportCSSButtonBase(t *testing.T) {
	css := cssCompact(ReportCSS)
	if !strings.Contains(css, ".btn{cursor:pointer;") {
		t.Error("compiled CSS lacks the btn pointer cursor, so buttons get no pointer cursor")
	}
	if !regexp.MustCompile(`\.btn:disabled\{[^}]*cursor:default[;}]`).MatchString(css) {
		t.Error("compiled CSS lacks the .btn:disabled rule with cursor:default")
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

// blendHex composites a compiled #rrggbbaa or #rgba color over an opaque #rrggbb base.
func blendHex(fg, bg string) string {
	fg = strings.TrimPrefix(fg, "#")
	if len(fg) == 4 {
		fg = string(fg[0]) + string(fg[0]) + string(fg[1]) + string(fg[1]) + string(fg[2]) + string(fg[2]) + string(fg[3]) + string(fg[3])
	}
	if len(fg) == 3 {
		fg = string(fg[0]) + string(fg[0]) + string(fg[1]) + string(fg[1]) + string(fg[2]) + string(fg[2])
	}
	alpha := 1.0
	if len(fg) == 8 {
		a, _ := strconv.ParseUint(fg[6:], 16, 8)
		alpha = float64(a) / 255
	}
	bg = strings.TrimPrefix(canonHex(bg), "#")
	rgb := "#"
	for i := 0; i < 6; i += 2 {
		f, _ := strconv.ParseUint(fg[i:i+2], 16, 8)
		b, _ := strconv.ParseUint(bg[i:i+2], 16, 8)
		rgb += fmt.Sprintf("%02x", int(math.Round(float64(f)*alpha+float64(b)*(1-alpha))))
	}
	return rgb
}

func rawColorToken(css, name string) string {
	re := regexp.MustCompile(`--` + regexp.QuoteMeta(name) + `:(#[0-9a-fA-F]+)`)
	m := re.FindStringSubmatch(css)
	if m == nil {
		return ""
	}
	return m[1]
}

type contrastPair struct{ name, fg, bg string }

func additionalContrastPairs(t *testing.T) []contrastPair {
	t.Helper()
	light, dark, _ := schemeBlocks(t)
	var pairs []contrastPair
	for scheme, css := range map[string]string{"light": light, "dark": dark} {
		token := func(name string) string {
			v := rawColorToken(css, name)
			if v == "" {
				t.Fatalf("%s token --%s missing", scheme, name)
			}
			return v
		}
		bg := canonHex(token("bg"))
		card := blendHex(token("glass"), bg)
		pairs = append(pairs, contrastPair{scheme + " muted-2 on page", canonHex(token("muted-2")), bg}, contrastPair{scheme + " muted-2 on card", canonHex(token("muted-2")), card})
		for _, tone := range []string{"major", "minor", "patch", "err"} {
			pairs = append(pairs, contrastPair{scheme + " " + tone + " chip", canonHex(token(tone + "-fg")), blendHex(token(tone+"-bg"), card)})
		}
	}
	lt := cssTokens(light)
	pairs = append(pairs, contrastPair{"selected tab count", lt["count-on-fg"], lt["count-on-bg"]})
	return pairs
}

func checkAdditionalReportCSSTextContrast(t *testing.T) {
	for _, p := range additionalContrastPairs(t) {
		if ratio := contrastRatio(p.fg, p.bg); ratio < 4.5 {
			t.Errorf("%s = %.2f:1, want at least 4.5:1", p.name, ratio)
		}
	}
}

func TestReportCSSStrongGlassShadow(t *testing.T) {
	css := cssCompact(ReportCSS)
	if strings.Contains(css, "--shadow-lg") {
		t.Error("compiled CSS retains --shadow-lg")
	}
	for _, want := range []string{"--shadow-strong:", "--glass-shadow:var(--shadow-strong)", "var(--glass-shadow,var(--shadow))"} {
		if !strings.Contains(css, want) {
			t.Errorf("compiled CSS lacks %q", want)
		}
	}
}

func TestReportCSSToneColors(t *testing.T) {
	css := cssCompact(ReportCSS)
	for _, want := range []string{".bg-tone{background-color:var(--tone)}", ".text-tone{color:var(--tone)}", "background:var(--tone-bg)", "color:var(--tone-fg)", ".bar-seg{min-width:6px}"} {
		if !strings.Contains(css, want) {
			t.Errorf("compiled CSS lacks %q", want)
		}
	}
	if strings.Contains(css, ".swatch{") {
		t.Error("unused swatch selector remains")
	}
}

func TestReportCSSButtonFocus(t *testing.T) {
	css := cssCompact(ReportCSS)
	if !regexp.MustCompile(`\.btn:focus-visible\{[^}]*outline:2pxsolidvar\(--accent\);outline-offset:2px`).MatchString(css) {
		t.Error("button focus outline missing")
	}
	if !strings.Contains(css, ".icon-btn[data-state=copied]{color:var(--ok)}") {
		t.Error("copied-state color missing")
	}
}

func TestReportCSSColorScheme(t *testing.T) {
	css := cssCompact(ReportCSS)
	for _, pattern := range []string{`:root\{[^}]*color-scheme:lightdark`, `:root\[data-theme=light\]\{color-scheme:light`, `:root\[data-theme=dark\]\{[^}]*color-scheme:dark`, `:root:not\(\[data-theme=light\]\)\{[^}]*color-scheme:dark`} {
		if !regexp.MustCompile(pattern).MatchString(css) {
			t.Errorf("missing color scheme: %s", pattern)
		}
	}
}
func TestReportCSSFocusVisible(t *testing.T) {
	css := cssCompact(ReportCSS)
	for _, sel := range []string{"summary:focus-visible", ".export-dialoginput[type=checkbox]:focus-visible"} {
		i := strings.Index(css, sel)
		if i < 0 {
			t.Errorf("missing %s", sel)
			continue
		}
		end := strings.Index(css[i:], "}")
		if end < 0 || !strings.Contains(css[i:i+end], "outline:2pxsolidvar(--accent)") || !strings.Contains(css[i:i+end], "outline-offset:2px") {
			t.Errorf("focus outline missing for %s", sel)
		}
	}
}

func TestReportCSSPanelMotion(t *testing.T) {
	css := cssCompact(ReportCSS)
	for _, want := range []string{"@keyframespanel-in", "blur(4px)", "24px", "-24px", ".42s", "@media(prefers-reduced-motion:no-preference)", "[data-panel][data-dir]:not([hidden])"} {
		if !strings.Contains(css, want) {
			t.Errorf("panel animation lacks %q", want)
		}
	}
	block := cssBlock(css, "@media(prefers-reduced-motion:no-preference){")
	if !strings.Contains(block, "[data-panel][data-dir]:not([hidden])") || !strings.Contains(block, "animation:panel-in") {
		t.Error("panel animation must be inside the no-preference media block")
	}
	markup := presentDocument(t, nil, summaryAt(loadFixtureReport(t)))
	markup = regexp.MustCompile(`(?s)<style.*?</style>|<script.*?</script>`).ReplaceAllString(markup, "")
	if strings.Contains(markup, "data-dir=") {
		t.Error("panel animates on first render")
	}
}

func cssBlock(css, selector string) string {
	start := strings.Index(css, selector)
	if start < 0 {
		return ""
	}
	open := strings.IndexByte(css[start:], '{') + start
	depth := 0
	for i := open; i < len(css); i++ {
		switch css[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return css[open:i]
			}
		}
	}
	return ""
}

func TestReportCSSDetailsMotion(t *testing.T) {
	css := cssCompact(ReportCSS)
	motion := cssBlock(css, "@media(prefers-reduced-motion:no-preference){")
	support := cssBlock(motion, "@supports(interpolate-size:allow-keywords){")
	for _, want := range []string{":root{interpolate-size:allow-keywords}", "details::details-content{height:0", "height.3svar(--ease)", "content-visibility.3sallow-discrete", "details[open]::details-content{height:auto}"} {
		if !strings.Contains(support, want) {
			t.Errorf("details motion lacks %q", want)
		}
	}
	if strings.Count(css, "details::details-content") != strings.Count(motion, "details::details-content") {
		t.Error("details motion outside no-preference")
	}
}

func TestReportCSSDialogMotion(t *testing.T) {
	css := cssCompact(ReportCSS)
	motion := cssBlock(css, "@media(prefers-reduced-motion:no-preference){")
	for _, want := range []string{".export-dialog{", "opacity.2svar(--ease)", "transform.2svar(--ease)", "overlay.2sallow-discrete", "display.2sallow-discrete", ".export-dialog[open]{", ".export-dialog:not([open]){pointer-events:none}", ".export-dialog::backdrop{", ".export-dialog[open]::backdrop{", "@starting-style{"} {
		if !strings.Contains(motion, want) {
			t.Errorf("dialog motion lacks %q", want)
		}
	}
	start := cssBlock(motion, "@starting-style{")
	if !strings.Contains(start, ".export-dialog[open]") || !strings.Contains(start, ".export-dialog[open]::backdrop") || !strings.Contains(start, "scale(.96)") {
		t.Error("dialog entry states missing")
	}
	if strings.Contains(strings.Replace(css, motion, "", 1), "overlay.2sallow-discrete") {
		t.Error("dialog transition outside no-preference")
	}
}

func TestReportCSSThemePillTransition(t *testing.T) {
	css := cssCompact(ReportCSS)
	if !strings.Contains(css, ".theme-control>.pill{view-transition-name:theme-pill}") || strings.Count(css, "view-transition-name:theme-pill") != 1 {
		t.Error("theme pill transition name missing or not unique")
	}
	group := cssBlock(css, "::view-transition-group(theme-pill){")
	if !strings.Contains(group, "animation-duration:.45s") || !strings.Contains(group, "animation-timing-function:var(--ease)") {
		t.Error("theme pill timing missing")
	}
	reduce := cssBlock(css, "@media(prefers-reduced-motion:reduce){")
	for _, want := range []string{"::view-transition-group(theme-pill)", "::view-transition-old(theme-pill)", "::view-transition-new(theme-pill)", "animation:none"} {
		if !strings.Contains(reduce, want) {
			t.Errorf("reduced motion lacks %s", want)
		}
	}
}
