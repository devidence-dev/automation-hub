# Design

## Context

See proposal.md for the motivation. Current state:

- `Header` in `report.templ` is a `flex flex-wrap justify-between` row of the logo and one `flex-wrap` group with the stamp, the export button and the theme control. Where it wraps depends on text width.
- The stamp text is built in `document.go` from `Summary.ReceivedAt` in the configured timezone, so the instant is available for a `datetime` value with no new input.
- `ExportButton` hides its label up to 600px (`sr-only min-[601px]:not-sr-only`).
- `exportAction` in `export.templ` renders text children only. `CopyButton` already swaps the copy and check icons through `group-data-[state=copied]`, and `report.js` sets `data-state="copied"` on any flashed button, the dialog copy action included.
- Repo cards are native `<details>`; only the chevron has a transition.
- The export dialog is a native `<dialog>` opened with `showModal()` and closed by a `method="dialog"` form or Escape.
- `chooseTheme` in `report.js` calls `movePill` and then `document.startViewTransition(applyTheme)`. The root crossfade (0.45 s) shows the old snapshot over the live page, so the 0.45 s pill transition is mostly hidden.
- Budgets: the CSS is 19,804 of 22,000 bytes and the fixture document 104,468 of 130,000 bytes (`TestDocumentWeight`).

## Goals / Non-Goals

**Goals:**
- CSS-first motion with progressive enhancement: unsupported browsers keep today's instant behavior.
- Keep the DOM order a sensible focus order.

**Non-Goals:**
- No JavaScript height or dialog animation (no `animate()` and no measuring).
- No polyfill for `interpolate-size`, `@starting-style` or the `overlay` transition.

## Decisions

### D1. Header as a CSS grid with named areas

The header becomes `grid` with two columns (`1fr auto`) up to 600px: `logo theme` / `meta meta`. `meta` is a flex row with the stamp (`min-w-0 flex-1`, text `truncate`) and the export button (`shrink-0`). From 601px the grid is one row, `logo meta theme`, with `meta` sized to its content and aligned right.
DOM order: logo, theme control, meta. This matches the phone layout, where the header is most crowded. From 601px the theme control shows last while it comes before Export in the focus order. Two independent header controls, so the reading order keeps its meaning (WCAG 2.4.3).
Alternative: keep `flex-wrap` and add `order` classes. Rejected: wrapping still depends on text width.
Alternative: one layout at every width. Rejected: a stamp stretched across 720px looks empty.

### D2. `<time>` stamp

`HeaderView` gains `ScanDateTime string`, set in `document.go` as `ReceivedAt.In(loc).Format(time.RFC3339)`. The template renders `<time datetime={...}>` around the existing text. When the text is truncated, the full value stays in the DOM, and a `title` with the stamp text shows it on hover.

### D3. Export button label always visible

Drop `sr-only min-[601px]:not-sr-only` from the label span. `aria-label` and `title` stay. The `Export` text comes from the catalog as today.

### D4. Card height motion

```
@supports (interpolate-size: allow-keywords) {
  :root { interpolate-size: allow-keywords; }
  details::details-content { height: 0; overflow: clip;
    transition: height .3s var(--ease), content-visibility .3s allow-discrete; }
  details[open]::details-content { height: auto; }
}
```
The rules sit in a `prefers-reduced-motion: no-preference` block, as the panel motion does. The `@supports` guard keeps browsers that know `::details-content` but cannot interpolate `auto` on today's behavior.
Alternative: JavaScript that measures `scrollHeight`. Rejected: more script, and it fights the native toggle and find-in-page.

### D5. Dialog motion

`.export-dialog` and `::backdrop` transition `opacity`, `transform` (scale .96 to 1), `overlay` and `display` (both `allow-discrete`) over 200 ms. The entry state comes from `@starting-style`. Browsers without `overlay` (Firefox today) animate the open and close the dialog at once. The rules sit in the reduced-motion `no-preference` block. While the close transition runs, the dialog no longer has `open` and the page is no longer inert, so `.export-dialog:not([open])` gets `pointer-events: none` to keep clicks from landing on the fading dialog.
Risk handled: `transform` on the dialog makes it the containing block for `position: fixed`. The copy fallback text area appended inside the open dialog is invisible (`opacity: 0`), so its position does not matter.

### D6. Icons on the dialog actions

`icon.go` gets `download` (arrow down into a tray, the mirror of `export`) and `close` (two strokes on a 14px grid). `exportAction` keeps its children API: each call site passes `@Icon(...)` before the label. Copy passes both `copy` (`group-data-[state=copied]:hidden`) and `check` (`hidden group-data-[state=copied]:block`), like `CopyButton`. The button class gains `inline-flex items-center gap-1.5`. No new catalog key: the icons are `aria-hidden` and the labels already exist.

### D7. Theme pill inside the view transition

The theme pill gets `view-transition-name: theme-pill`, so the browser captures it apart from the root and animates its box from the old to the new position. In `chooseTheme`, when a view transition runs, the pill moves inside the update callback with the `no-anim` class on its group, so the captured new state is the final position. The group animation then slides it over the root's 0.45 s with `--ease`. Without view transitions, or under reduced motion, `movePill` runs as today, and the pill's own CSS transition slides it. The tab pill gets no name, because it does not move during a theme change.
Alternative: drop the view transition. Rejected: the crossfade of the theme colors is wanted (Motion requirement).

### D8. Scroll lock with `:has()`

The root gets `overflow: hidden` and `scrollbar-gutter: stable` through `html:has(.export-dialog[open])`, so the page does not scroll while the dialog is open, and no script runs. The gutter stays reserved while the scrollbar is hidden, so the page does not shift sideways. The dialog gets `overscroll-behavior: contain`, so scrolling inside it does not chain to the page. The rule is not motion, so it sits outside the reduced-motion blocks. The lock lifts as soon as `open` is removed, at the start of the close transition, which is acceptable.
Alternative: toggle a class on the root from `report.js` on open and close. Rejected: more script, and it must handle Escape and the form close as well.

### D9. Corner close

An icon-only X sits beside the dialog title, in a `method="dialog"` form. Activating it closes the dialog natively, so no script runs: `report.js` already returns focus to `Export` on the `close` event. The X reuses the catalog `document.export.close` label for its accessible name and tooltip. It reuses the 30x30 icon button style, renamed from `copy-btn` to `icon-btn` because it is no longer copy-specific. The bottom `Close` action stays for users who reach the end of the form. With the X first in the dialog, `showModal()` puts the initial focus on it, which is acceptable: the WAI-ARIA dialog pattern allows focusing the close control.

## Risks / Trade-offs

- [The CSS budget of 22,000 bytes] → about 1 KB is expected. `TestDocumentWeight` reports the measured size; raise the budget only on purpose, in this change's spec.
- [Browsers differ: Firefox and Safari lack `interpolate-size`, and Firefox lacks `overlay`] → they show the instant behavior, which the spec allows.
- [The theme pill's view transition depends on the capture timing] → the static tests cover the name and the script contract; the slide itself is a (user) browser check.
- [The header focus order differs from the visual order from 601px] → accepted (D1).

## Migration Plan

Only the generated document changes. Rollback is a revert of the commit. Old reports already sent are unaffected.
