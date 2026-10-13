## Why

The dependency report document is sent to Telegram as one HTML file. Its first redesign (batches 1 to 2b) used Tailwind and templ with a flat palette and CSS-only tabs. On 2026-10-12 the user supplied a "liquid glass" visual reference (now in the golden `document.html`) and accepted it in full. The report must match it and stay one self-contained file that Telegram opens with no sibling files.

## What Changes

- Port the liquid glass design into the templ and Tailwind report: tokens for three color schemes, glass surfaces, the summary panel, tabs, repo cards, entry rows, the errors card and the motion of the mock.
- Replace CSS-only tabs with JS-driven tabs (`role=tab`, `aria-selected`, hidden panels, Left/Right to move focus). Dependencias and Imágenes are the tabs, each with its count of listed updates.
- Add a theme control (Automático, Claro, Oscuro). It starts on Automático on every open. Claro and Oscuro apply to the page view only and are not stored (user decision, 2026-10-12).
- Inline the logo as one SVG whose wordmark colors are CSS custom properties. The favicon is the stripped icon as a data URI.
- Summary row and distribution: text chips colored by type and a colored dot. The emoji counters and the repo emoji are removed; their catalog keys stay, unused.
- The copy check uses the mock's teal `--ok` color. The copy values do not change.
- Every visible text comes from the catalog, and the bar widths are computed in Go.
- Add a local `generated-assets` job to `.github/workflows/ci.yml`. It regenerates the templ and Tailwind assets and fails on any drift. The SonarCloud Quality Gate already covers coverage, complexity and lint, so the shared-CI extension is dropped (design D17).

## Non-goals

- No change to the Telegram caption, the failure message, the webhook, the config keys or the delivery. The caption golden is byte-identical.
- No hosted page, no second document format, no server-side state, and no stored open or closed state of the repo sections.
- No new Go or npm dependency. Tailwind 4.3.3 and templ v0.3.1070 stay the build tools.
- No change to `internal/messages` beyond the default catalog and key constants.

## Superseded decisions

The earlier decisions D4, D5, D6, D8, D10 and D14 are superseded by batch 3, and D17 by the 2026-10-12 CI decision. `design.md` records each one and the reason.
