# Proposal

## Why

The dependency report front-end repeats the same markup thousands of times. In the fixture document (170,638 bytes), the 54 entry rows take 98.8 KB (1,857 B each): 117 inline icons repeat nine attributes, and the copy button, chip and row classes repeat 290, 146 and 148 characters per row. The compiled CSS (21,474 B) carries rules nobody uses: Tailwind scans comment words ("ring", "inline") and the shadow and numeric utilities pull in 19 extra `@property` rules. An audit also found real accessibility gaps: no `<main>`, an unnamed tablist, two text colors below 4.5:1, and native controls that ignore a forced theme. Tests pin class strings, which blocks cleanup.

## What Changes

- Write safety-net tests first: class coverage both ways, a JS-to-markup contract, semantic checks, contrast pairs and size budgets.
- Remove CSS leaks: reword comment words, plain utilities for the dot glow and tabular numbers, rename `--shadow-lg`.
- Move icon attributes into one `icon` utility; add `copy-btn`, `chip` (tone tokens), `dep-row` and `glass-pill` utilities; put the focus ring into `btn`; drop dead wrappers and markers.
- Move per-row constants (copied texts, Current/New labels) to document level; split `labelSet`; add `ErrorsView` and `HeaderView`.
- Accessibility: `<main>`, named tablist (new catalog key `document.tabs.label`, default `Report sections`), list rows, no redundant `aria-live`, `color-scheme` rules, contrast fixes.
- Clean `report.js`: dead guards, redundant init, `copyAndFlash`, `data-pill-group`, `ResizeObserver` for pills, panel animation in CSS.
- Visible changes: focus ring on tab and theme buttons, two slightly darker/different colors, native controls follow the theme.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `dependency-report`: ADDED landmarks and list semantics, named tablist, visible keyboard focus, native controls following the theme, document weight budgets. MODIFIED "Document theme" (all text tokens meet 4.5:1, closing the open question about the selected tab count).

## Impact

- `internal/services/depreport`: `*.templ`, `groups.go`, `document.go`, `keys.go`, `assets/input.css`, `assets/report.js`, generated `*_templ.go` and `depreport.min.css`, tests.
- `internal/messages/default.yaml`: one new key.
- `internal/handlers/testdata/golden/document.html`: regenerated at each step.
- No new dependency, config change, or change to caption, delivery or export format.

## Non-goals

- Tokenizing text sizes, radii, spacing and the 601px breakpoint (`sm:`), naming and file layout, hand-written preflight, `light-dark()` dedupe: possible follow-ups.
- An SVG sprite, removing `data-yaml`, a JS minifier, client-side row rendering, CSS-only tabs or theme radios, Invoker Commands, a JS test framework.
- Dropping Google Fonts or adding a no-JS fallback.
