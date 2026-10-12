## Why

The dependency report document is sent to Telegram as one HTML file. Its first redesign (batches 1 to 2b) used Tailwind and templ with a flat palette and CSS-only tabs. On 2026-10-12 the user supplied a visual reference, `reporte-dependencias.html` (removed from the repo after the port; the golden `document.html` now embodies it), with a "liquid glass" look: a mesh background, translucent glass surfaces, sliding pills for the tabs and the theme control, and a light and a dark theme. The user accepted the mock in full ("acepto todo lo del mock"). The report must match it, and the document must stay one self-contained file that Telegram can open with no sibling files.

## What Changes

- Port the liquid glass design into the real templ and Tailwind report: tokens for three color schemes, glass surfaces, the summary panel, tabs, repo cards, entry rows, the errors card and the motion of the mock.
- Replace CSS-only tabs with JS-driven tabs (`role=tab`, `aria-selected`, hidden panels, Left/Right to move focus). Dependencias is the first tab, Imágenes the second, and each tab shows its count of listed updates.
- Add a theme control (Automático, Claro, Oscuro). It starts on Automático on every open. Claro and Oscuro apply to the page view only and are not stored (user decision, 2026-10-12: the mock and the report shared the `dv-theme` key on `file://`, so the report opened in dark).
- Inline the logo as one SVG whose wordmark colors are CSS custom properties. The stripped SVG sources live in `assets/` and are embedded. The favicon is the stripped icon as a data URI.
- Summary row and distribution: text chips colored by type and a colored dot. The emoji counters and the repo emoji are removed from the document. Their catalog keys stay, unused.
- The copy check uses the mock's teal `--ok` color. The copy values do not change.
- Every visible text comes from the catalog, and the bar widths are computed in Go.

## Non-goals

- No change to the Telegram caption, the failure message, the webhook, the config keys or the delivery. The caption golden is byte-identical.
- No hosted page, no second document format, no server-side state, and no stored open or closed state of the repo sections.
- No new Go or npm dependency. Tailwind 4.3.3 and templ v0.3.1070 stay the build tools.
- No change to the code of `internal/messages` beyond the default catalog and key constants.

## Superseded decisions

The earlier decisions D4 (emoji and summary counters), D5 (CSS-only tabs), D6 (container queries), D8 (palette A), D10 (Heroicons and green copy) and D14 (contrast values) are superseded by batch 3. `design.md` records each one and the reason.
