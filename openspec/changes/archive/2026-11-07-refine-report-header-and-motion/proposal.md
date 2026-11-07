# Proposal

## Why

At phone widths (440 CSS px) the header wraps by text width, so the theme control drops alone to a second line. Repo cards and the export dialog open and close with no motion, and the theme pill does not visibly slide on a theme change (the page crossfade covers it). The export dialog actions are the only buttons without icons.

## What Changes

- Header layout: up to 600px, row 1 holds the logo and the theme control, and row 2 holds the scan stamp and the export button across the full width. The stamp truncates with an ellipsis when it does not fit. From 601px the header stays one row, as today.
- The scan stamp becomes a machine-readable date (`<time datetime>`).
- The export button shows its label at every width (today it is icon-only up to 600px).
- A repo card animates its height when it opens and when it closes.
- The export dialog and its backdrop fade and scale in on open and out on close.
- The export dialog actions get icons: copy (switching to a check after a copy), download and close.
- The theme control pill slides to the pressed button during the theme change, as the tab pill does.
- All new motion is off under `prefers-reduced-motion: reduce`. Browsers without the needed CSS features keep today's instant behavior.

## Non-goals

- Touch target sizes stay as they are (theme 28×30 px, copy 30×30 px); they meet WCAG 2.5.8 AA.
- No new user-facing text: the action labels stay as they are and the icons are decorative.
- No change to the tab pill, the panel motion, the copy buttons of the entries or the export content.
- No JavaScript animation library and no new dependency.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `dependency-report`: MODIFIED "Export button" (label visible at every width) and "Motion" (card height, dialog open and close, theme pill during the theme change). ADDED "Header layout by viewport width" (two-row layout up to 600px), "Machine-readable scan time" (`<time>` stamp) and "Export action icons".

## Impact

- `internal/services/depreport`: templates, `icon.go`, `document.go`, `assets/input.css`, `assets/report.js` and the generated assets.
- Tests in that package and the golden `internal/handlers/testdata/golden/document.html`.
- No config, catalog, API or deployment change.
