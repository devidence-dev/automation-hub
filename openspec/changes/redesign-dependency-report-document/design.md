# Design

## Context

The document is built in `internal/services/depreport/document.go` from a view model of repo groups, file groups and entries, and rendered by templ components (D2). The caption (`caption.go`) reads the same `Summary` from `Aggregate` (`aggregate.go`). The caption and failure goldens in `internal/handlers/testdata/golden/` must stay byte-identical. See `proposal.md` for the motivation and `specs/` for the behavior contract.

Batch 3 ports the liquid glass visual reference the user supplied on 2026-10-12: `reporte-dependencias.html` (removed from the repo after the port; the golden `document.html` now embodies it), with `devidence-logo-light.svg`, `devidence-logo-dark.svg` and `devidence-icon.svg` in the same folder. The mock is the visual source of truth, read only; the folder is the user's untracked work and is not modified. The mock's markup is ported to templ components, its tokens to `input.css`, and its script to `assets/report.js`. The user accepted the mock in full. The hub runs as one binary configured from `CONFIG_YAML`. The Go toolchain is pinned by `.go-version`. Tooling is Docker-first: nothing is installed on the host.

## Goals / Non-Goals

**Goals:**
- The liquid glass look of the mock in light, dark and forced themes, from one stylesheet with tokens.
- Phone-first layout with no horizontal scroll, and the mock's 600px breakpoint.
- Accessible tabs and theme control: ARIA roles, roving focus, labels from the catalog.
- One self-contained HTML file: inline CSS, inline scripts, inline logo, no network resource other than Google Fonts.
- Typed, testable components; markup, styles and view model each have one responsibility.
- Generated files are reproducible, committed, and checked by CI.

**Non-Goals:**
- No change to the Telegram delivery, the caption, the failure message, the webhook or config keys (see `specs/`).
- No hosted page, no second document format, no server-side state.
- No storage of which repo sections are open, and no stored theme choice (D20).
- No TypeScript, no framework, no script beyond the report script (D20).

## Decisions

### D1. One self-contained HTML file per report

The report stays one HTML file sent with `sendDocument`, with the summary caption as its caption. Telegram opens it in the device browser, so inline CSS and inline scripts work, and no backend or hosting is needed. This holds for the logo too: a file sent by Telegram has no sibling files (D21).
- Alternative: a hosted page with a link in the caption. Rejected: needs hosting, access control and retention, none of which the hub has, and it changes the delivery contract.

### D2. templ replaces html/template for the document

Markup is written as typed templ components (`github.com/a-h/templ` v0.3.1070, verified on proxy.golang.org at commit 07dfb6e and approved by the user on 2026-10-12). Each component has one responsibility:

| Component | File | Responsibility |
|---|---|---|
| `ReportDocument` | `report.templ` | document shell: `lang`, head, inline CSS, theme script, report script |
| `Header`, `Logo` | `report.templ`, `logo.templ` | brand logo (inline SVG), scan stamp, theme control |
| `ThemeControl`, `ThemeButton` | `theme.templ` | segmented Automático, Claro, Oscuro control |
| `Tabs`, `TabPanel` | `tabs.templ` | tab list and the two panels (Dependencias first) |
| `StatsPanel`, `DistributionBar`, `Legend` | `stats.templ` | eyebrow, big number, sub line, bar and legend |
| `SummaryLine` | `stats.templ` | the line between the panel and the repo cards |
| `RepoGroup`, `RepoSummary` | `repo.templ` | one `<details>` card and its summary row (dot, name, note, chips, chevron) |
| `TypeChip`, `TypeDot` | `chip.templ` | text chip and dot colored by update type |
| `FileGroup`, `DepRow`, `CopyButton` | `entry.templ` | one package file, one entry row, its copy button |
| `ErrorsCard` | `errors.templ` | scan errors, shown only when there are errors |
| `Icon(name)` | `icon.templ` | the mock's 14px icons from a fixed Go map |

Alternative: keep `html/template` and split it into partials. Rejected: partials are untyped, and the view model still leaks CSS classes. templ escapes text and attributes like `html/template`, so the "Safe document content" requirement holds. `templ.Raw` is used only for the compiled CSS, the two scripts, the logo and the icon SVG markup, all build-time constants.

### D3. Presenter keeps its interface; the view model groups repo, file, entry

`DocumentPresenter.Present(Summary) ([]byte, error)` keeps its signature. Its body builds `documentView` and renders `ReportDocument` into a buffer.

The view model (`document.go`, `groups.go`, `stats.go`) has:
- `TabView`: `ID`, `Label`, `Count` (listed updates, the badge), `Selected`, `Images`, `Stats StatsView`, `Summary SummaryLine`, `Groups []RepoGroupView`.
- `StatsView`: `Eyebrow`, `Big` (the listed updates), `Sub` (pendientes de total · up to date), `Segments []BarSegment` (patch, minor, major; zero segments omitted), `Legend []LegendItem` (patch, minor, major, zero kept).
- `RepoGroupView`: `Repo`, `Short`, `Note` (fork note, empty when not a fork), `Dot UpdateType` (major if any, else minor, else patch), `Chips []CountChip` (non-zero types), `Files []FileGroupView`.
- `FileGroupView`: `Path` and `Entries []EntryView`.
- `EntryView`: `Name`, `Current`, `New` (versions only, for both deps and images), `Type`, `TypeLabel`, `CurrentLabel`, `NewLabel`, `Copy CopyView`.

`Aggregate` already sorts rows by repo, package file, dependency and type, so grouping is one pass over consecutive rows. Per-tab repo totals reuse `RepoPending` and the caption's comparator, which lives in `aggregate.go`. CSS classes leave the view model: `tone.go` maps `UpdateType` to a tone class in the templ layer.

### D4. Repo sections are native `<details>` without `name` (amended)

Each repo group is a `<details>` element with no `name` attribute and no `open` attribute, so each section toggles on its own and none closes the others. The state is not stored; reopening the file starts collapsed. All cards start closed. The mock shows QuantWarden open for illustration only; the document does not.
- Alternative: a shared `name` (exclusive accordion). Rejected: opening one closes the others.
- Alternative: a script that tracks open sections. Rejected: no state is stored (see D20).

Amended on 2026-10-12 (batch 3): the summary row shows a colored dot, the short name with the fork note, one text chip per non-zero type (`2 major`, `45 minor`, `3 patch`, colored by type), and a chevron. The dot is major when the section has a major update, otherwise minor, otherwise patch. The repo emoji and the emoji counters of batch 2 are removed from the document (decision a). The caption keeps its emoji.

**Counting rule:** each dependency counts once under its highest type, as in the caption, so `2 major · 1 patch` matches the caption line `🔴 QuantWarden — 2 major · 1 patch`. Counts are per tab: a repo in both tabs has one section in each, with the counts of that tab.

### D5. Tabs with JavaScript (superseded)

Superseded by batch 3 (decision f). The batch 2 radio tabs are replaced by buttons with `role="tab"` in a `role="tablist"`, `aria-selected`, roving `tabindex`, and hidden panels (`hidden`). Dependencias is the first tab and selected by default, Imágenes the second. Each tab shows its count of listed updates. Mouse and touch: a click activates the tab. Keyboard: Enter and Space activate the focused tab (native button behavior); Left and Right move focus between tabs and wrap around. Activation is manual, as the brief specifies. The sliding pill follows the selected tab, and the newly shown panel animates in (D20).

Switching tabs hides panels with the `hidden` attribute. An open repo section in a hidden panel keeps its state, since `hidden` does not reset the `details` element; the user checks this on the phone (task 12.1).

### D6. Entry layout by viewport width (superseded)

Superseded by batch 3: the mock uses the viewport, not a container query, and no subgrid. Each entry row is its own grid with the same column template as every other row, so the columns align within a file group without subgrid: `minmax(0,1fr) auto 58px 30px` (name, versions, type chip, copy button). Below 600px the row stacks into a two-column grid: name and versions in the first row, chip and copy button in the second. The breakpoint is the viewport: the stacked layout is the default and applies up to 600px, and `min-[601px]:` restores the row from 601px, which is the mock's `max-width: 600px` written mobile first.

### D7. Tailwind CSS 4.3.3, compiled offline in Docker

`input.css` contains `@import "tailwindcss" source(none);`, `@source` for `*.templ`, the tokens of D8, `@theme inline` mappings, and `@utility` definitions for the glass surface, the pill and the mesh background, and a `@layer components` block for the type tones, the bar segments, the chevron and the motion rules (see D8, D20). The tones are plain classes because the Go code adds them (`tone.go`), and Tailwind scans only the `.templ` files. Components use utility classes in the `.templ` files. The standalone CLI compiles and minifies the CSS, the result is embedded with `go:embed`, and `ReportDocument` inlines it in a `<style>` element.
- Alternative: Bootstrap 5.3 kept locally. Rejected: it clashes with Tailwind. Bootstrap 6 exists only as `6.0.0-alpha.1`.
- Alternative: hand-written CSS. Rejected: the user chose Tailwind; the tokens and utilities are easier to keep consistent.

Tailwind drops an unused `@utility`. Each utility the document uses (`glass`, `pill`, `app-bg`) and each class of the component layer (`tone-*`, `bar-seg`, `swatch`) must appear in the compiled CSS, and `styles_test.go` checks that.

### D8. Liquid glass tokens and three color schemes

The tokens are the mock's custom properties, with its names, in `input.css`. They are declared in three blocks, exactly like the mock:
- `:root`: light values.
- `@media (prefers-color-scheme: dark) { :root:not([data-theme=light]) { … } }`: dark values when the system prefers dark and Automático is selected.
- `:root[data-theme=dark] { … }`: dark values for Oscuro, chosen for the page view.

`:root[data-theme=light]` takes the light values from `:root`, and `data-theme` is removed for Automático. `@theme inline` maps the color tokens to utilities (`--color-ink: var(--ink)`, `--color-major-fg: var(--major-fg)` and so on), so one stylesheet serves all three schemes and no `dark:` variant is used.

Main tokens (light / dark):
- background `--bg` `#EDF3F8` / `#04101A`, mesh `--mesh` (three radial gradients, the mock's values per scheme).
- ink `--ink` `#0B1B25` / `#EAF2F8`; ink-2 `#33475A` / `#C9D8E2`; muted `--muted` `#56697A` / `#9FB6C6`.
- accent and file `#0A63C2` / `#7FD9EA`; minor `#0A84FF` in both schemes.
- glass `--glass` `rgba(255,255,255,.55)` / `.05`; glass-strong and glass-soft as in the mock; edge `--edge`; highlight `--hi`; shadows `--shadow`, `--shadow-lg`; saturation `--sat` 180% / 160%.
- type tones (dot, chip text, chip background): major dot `#FF6B57`, fg `#B93A24`, bg `rgba(255,99,71,.13)` in light, and dot `#FF6B57`, fg `#FFA391`, bg `rgba(255,110,90,.18)` in dark; minor dot `#0A84FF`, fg `#0A63C2` / `#8CC8FF`, bg `rgba(10,132,255,.12)` / `.22`; patch dot `#16CBE7`, fg `#07707F` / `#7AF2FF`, bg `rgba(22,203,231,.16)` / `.18`; errors `--err` `#FF3B30` / `#FF453A`, fg `#C4141F` / `#FF8A82`, bg `rgba(255,59,48,.12)` / `.16`.
- `--ok`: `#07707F` light, `#6CF5FF` dark (copy done state).
- logo wordmark: `--logo-main` `#0B1B25` / `#F2F7FB` and `--logo-sub` `#334C5A` / `#9FB6C6` (D21).

Contrast (WCAG 2.x, computed 2026-10-12 against the solid page background, `#FFFFFF` and `#EDF3F8` in light, `#04101A` in dark): the text tokens are 5.08:1 or more in light (major 5.08, minor 5.25, patch 5.17, errors 5.42, muted 5.08) and 8.4:1 or more in dark. One pair fails AA: the white count on the selected tab badge (`#FFFFFF` on `#0A84FF`) is 3.65:1. The mock uses it; it is kept and listed as an open question.

The batch 2 palette A (`--c-*` tokens and the `success` tokens) is removed. Its contrast test is replaced by a test of the mock token values.

### D9. Fonts

Geist (300, 400, 500, 600) and Geist Mono (400, 500) load from Google Fonts. Weight 300 is added for the big number. The font stack falls back to `ui-sans-serif, system-ui, -apple-system, "Segoe UI", sans-serif` and `ui-monospace, "SF Mono", Menlo, monospace`. The Google Fonts URL is exactly the mock's `family=Geist:wght@300;400;500;600&family=Geist+Mono:wght@400;500&display=swap`.
- Alternative: embed the font files as data URIs. Deferred: the file would grow by the size of the font files, and the user accepted a CDN.

### D10. Copy buttons and handler (amended)

Each entry has one icon-only copy button. Its accessible name (`aria-label`) and tooltip (`title`) are the catalog label. The icons are the mock's 14px stroke icons: a clipboard frame and a check, drawn in `icon.templ`. The button is 30px, with the mock's `--btn-bg`, `--btn-edge` and `--btn-hover`. After a copy the handler sets `data-state="copied"` for about 2 seconds, and the button shows the check in the `--ok` color. The swap is CSS only, through `group-data-[state=copied]:` variants.

Amended on 2026-10-12 (batch 3): the Heroicons 20/solid icons of batch 2 are replaced by the mock's 14px icons, and the green `success` copy color by the mock's teal `--ok` (decision b). The copy values are unchanged: a dependency copies `<depName>@<new version>` (`document.copy.new_version`), an image copies only its latest reference `<image>:<new version>` (`document.copy.latest_image`).

The handler lives in `assets/report.js` (D20). It copies `data-copy` values with `navigator.clipboard`, and falls back to a hidden text area when the API is missing or rejects the write. The fallback keeps the batch 2 iOS-safe calls: the text area is not `readonly`, and it is focused and selected with `select()` and `setSelectionRange`, because iOS WebKit does not copy a readonly text area's selection. After a copy it writes the `data-copied-text` text into the one `role="status"` live region (`copy-status`), and it sets `title` to `data-copied-before` on a second copy.

### D11. Generated files are committed and checked

`*_templ.go` and the compiled CSS are committed, like dc-sentinel. A generation script runs `templ generate` and Tailwind inside Docker. The shared CI runs it through the `generated-drift` step (`generate_command` input) and then `git diff --exit-code`, and `scripts/verify.sh` runs the same check. `git diff` does not see new untracked files, so the step also fails when the command creates one (D17). The generated Go files carry templ's `Code generated ... DO NOT EDIT` header, so golangci-lint skips them. `sonar-project.properties` excludes `**/*_templ.go` and the compiled CSS.

Binaries are downloaded only inside an ephemeral Docker container (`docker run --rm`), and each one is checked against the checksum file its release publishes before it runs. Verified 2026-10-12:
- Tailwind CSS v4.3.3 publishes `sha256sums.txt` in its GitHub release; the script downloads the asset for the container architecture and its line in that file.
- templ v0.3.1070 publishes `checksums.txt` in its GitHub release; the script downloads the archive for the architecture and its line in that file.
The script fails when the asset has no line in the checksum file or the hash differs. A missing `input.css` fails the script (task 9.1).

### D12. Catalog keys (amended)

Existing keys are kept where the meaning is the same. Batch 3 adds the keys below, changes some defaults, and marks others unused. Every key stays in `default.yaml` and has a constant, so an existing override still passes startup validation (see `specs/message-catalog`).

Added (defaults):
- `document.eyebrow.deps` = "Reporte de dependencias"; `document.eyebrow.images` = "Reporte de imágenes".
- `document.panel.sub` = "pendientes de {total} · {up_to_date} al día".
- `document.status.updates_bold_one` = "{updates} actualización"; `document.status.updates_rest_one` = "para {deps} dependencia"; `document.status.images_rest` = "para {deps} imágenes"; `document.status.images_rest_one` = "para {deps} imagen"; `document.status.images_up_to_date` = "Todas las imágenes están al día".
- `document.type.patch` = "patch", `document.type.minor` = "minor", `document.type.major` = "major" (legend and row chips).
- `document.errors.count` = "{count} errores"; `document.errors.count_one` = "{count} error".
- `document.theme.group` = "Tema"; `document.theme.system` = "Automático"; `document.theme.light` = "Claro"; `document.theme.dark` = "Oscuro".

Changed defaults (the meaning stays):
- `document.status.updates_bold` = "{updates} actualizaciones" (was "… disponibles", which the mock omits).
- `document.errors.heading` = "Errores del análisis" (was "Errores").
- `fork.tag_document` = "fork · solo informativo" (was " (fork, solo informativo)"); the view trims it and shows it as the note of the summary row. The caption's `fork.tag_caption` is unchanged.
- `document.copy.new_version` and `document.copy.latest_image` keep their texts.

Unused in the document, kept in the default catalog so existing overrides still validate: `document.brand` (the eyebrow replaces it), `document.total`, `document.pending`, `document.up_to_date`, `document.distribution` (the panel texts replace them), `document.summary.emoji.*` and `document.summary.counter` (decision a), `document.errors.item` (the errors card shows where and message in two cells), `document.copy.current_image`, `document.available`, and the `document.columns.*` keys except `current` and `new`, which label the versions for screen readers. `counts.*` keeps the caption and the summary chips (`counts.major` = "{count} major").

Placeholders the code fills are listed in `placeholderNames` (`keys.go`): `total` and `up_to_date` are added.

### D13. Errors and empty state

The errors card keeps its semantics and moves to the bottom through `ErrorsCard`. It has a red dot, the heading `document.errors.heading`, a chip with `document.errors.count` (or `count_one`), and one row per error: the repo or root in muted mono and the message. It is shown only when there are errors. The nothing-pending state keeps `✅ Todo al día` in the caption; the document shows zero updates in both tabs, the status line `document.status.up_to_date` (or `images_up_to_date`) and no repo cards.

### D14. Checks without a browser (amended)

No headless browser is used anywhere in this change. Each layout and theme requirement is split in two:
- **Static (Go tests, run in Docker):** the compiled CSS and the rendered markup are checked for no fixed `min-width` above 320px (the only `min-width` is the 601px media breakpoint), no `<table>`, the three token blocks with the mock's values, the 601px breakpoint, the `@utility` definitions in the stylesheet, the contrast of the text tokens (D8, minimum 4.5:1 for each text token against its base), and the rendered attributes (ARIA, `lang`, `color-scheme`, fonts link).
- **Visual (user):** how the layout looks and scrolls at phone, iPad and Mac width, in light, dark and both forced themes, is approved on the deployed report (task 12.1).

Amended on 2026-10-12: the batch 2 badge contrast values (palette A) are replaced by the D8 check.

### D15. Local gate `scripts/verify.sh`

One script runs the CI checks in Docker, in this order: `gofmt -l .`; golangci-lint v2.14.0 with `--enable=errcheck,govet,staticcheck,unused,misspell,gosec`; `go test -race` with coverage after the `sonar.coverage.exclusions` filter, minimum `MIN_COVERAGE` (default 80, D18); gocognit v1.2.0 with `-over 15 -ignore '_templ\.go$' .`; asset drift (the generation script, then `git diff --exit-code` plus the untracked-file check of D17); `biome ci` on the CSS and JS sources (D16); `openspec validate --all --strict`. Modelled on voraz-bot's `scripts/validate.sh`, with these differences:
- The Go image comes from `.go-version`, as the config and `ci.yml` do, not from the Dockerfile.
- The coverage filter takes the module path from `go.mod`, not validate.sh's `voraz-bot/` prefix.
- Every container runs with `--user "$(id -u):$(id -g)"`, and caches live in `~/.cache`.
- `--fix` runs `gofmt -w`, `biome check --write` and regenerates the assets.
- openspec runs in a pinned Node container, not on the host.

Alternatives: a pre-push hook. Deferred: the user runs the script by hand before pushing.

### D16. Biome as a standalone binary

Biome 2.5.15 (latest on npm, 2026-09-30) is run as its standalone binary inside Docker. There is no npm, no `package.json` and no `node_modules`. npm was rejected: it needs Node, a `package.json` and a lockfile, and the repo has no Node toolchain.

**Checksum finding (verified 2026-10-12):** the GitHub release `@biomejs/biome@2.5.15` publishes no checksum file. The GitHub API reports a SHA-256 digest for each binary, which is metadata and not the trust anchor. The binary is pinned by its SHA-256, computed from a download and committed in `verify.sh` and in the shared `biome` action.

Configuration in `biome.json`:
- Formatter and linter on, `css.parser.tailwindDirectives: true` (the Tailwind 4 directives in `input.css` parse; task 9.3).
- `files.includes` limited to `internal/services/depreport/assets/**` (`input.css` and `report.js`), excluding the generated `depreport.min.css` (D11). The SVG sources are not linted.

### D17. Shared CI extension (workflows)

The reusable `ci-go.yml` gains five optional inputs, all off by default: `coverage_min` (0), `cognitive_max` (0), `cognitive_ignore` (empty), `biome` (false) and `generate_command` (empty). Their jobs run in parallel, like the existing ones, and each is gated with `if:`, so existing callers see no change. The release order and the `v1` callers are unchanged from batch 2 (tasks 9b and 9.7).

### D18. Coverage minimum 80%

The minimum is enforced twice, in `verify.sh` (`MIN_COVERAGE`, default 80) and in CI (`coverage_min: 80`), both after excluding `cmd/automation-hub/main.go` as `sonar-project.properties` does. Measured 2026-10-12: 93.5% total, and the lowest package is `internal/services/kube` at 85.4%.

### D19. Tool version pins

gocognit v1.2.0, Biome 2.5.15, golangci-lint v2.14.0, templ v0.3.1070 and tailwindcss v4.3.3 are pinned in the scripts (`verify.sh`, the generation script) and in the workflow inputs and actions.

### D20. Liquid glass behavior: theme, tabs, motion, scripts (amended 2026-10-12)

Batch 3 ports the mock's script into one inline script, `assets/report.js`, embedded with `go:embed` (no network, decision c). It runs at the end of `<body>` and holds the theme control, the tabs, the pills, the panel motion and the copy handler. Every DOM query is guarded. It reads and writes no storage.

Amended 2026-10-12 (user decision): the theme choice is not stored. The early head script `assets/theme.js` is deleted, since nothing is restored before paint. Every open of the file starts on Automático, and Claro or Oscuro applies to the current page view only. Rationale: the user's request, and no storage errors on `file://`.

Behavior, from the mock:
- **Theme control.** Three buttons (`data-theme-set` system, light, dark) in a `role="group"` labeled `document.theme.group`, with `aria-pressed`. Automático is pressed on load and removes `data-theme`; Claro and Oscuro set `data-theme` on `<html>` for the page view only, and nothing is stored. A change runs inside `document.startViewTransition` when the API exists and reduced motion is not requested, so the page cross-fades; the pill slides to the new button.
- **Tabs.** Described in D5. The panel that appears animates with the Web Animations API: from `translateX(±24px)` with `blur(4px)` and opacity 0 to rest over 420 ms, in the direction of travel. Reduced motion skips the animation.
- **Pills.** The tab pill and the theme pill are placed with `width` and `translateX` from the selected button's `offsetWidth` and `offsetLeft`. The placement runs on load, on resize and after `document.fonts.ready`, with transitions off for those two frames (`html.no-anim`).
- **Chevron.** The chevron of an open card rotates 180 degrees with a 200 ms transition, through CSS on `details[open]`.
- **Reduced motion.** `prefers-reduced-motion: reduce` turns off the pill transitions, the view transition animation, the panel animation and the chevron transition.
- **Copy.** D10.

The scripts are linted by Biome (D16) and tested as follows: `TestThemeScriptEmbedded` and `TestReportScriptEmbedded` check that each inline script equals its embedded file and that exactly one `<script>` element holds each. The rendered document has no other script.

### D21. Logos and icon are inline

A file sent by Telegram has no sibling files, so no logo is an `<img src>`:
- **Logo.** `devidence-logo-light.svg` and `devidence-logo-dark.svg` differ only in two wordmark fills. They are merged into one source, `assets/logo.svg`, with the wordmark main fill as `style="fill:var(--logo-main)"` and the sub fill as `style="fill:var(--logo-sub)"`. The light and dark values are the tokens of D8, so the logo follows the theme chosen in the control, through the same `data-theme` switch. The gradients of the mock stay in the file, with their ids.
- **Stripping.** The `<metadata>` element (the C2PA manifest, about 7.7 KB per file), the `xmlns:c2pa` attribute, the `<desc>` and `<title>` elements and the `aria-labelledby` attributes are removed. The logo keeps `role="img"` and `aria-label="devidence.dev"`. The icon keeps its paths, gradients and `role="img"`.
- **Test.** `TestLogoAssetsAreClean` checks that `assets/logo.svg` and `assets/icon.svg` contain no `<script`, no `<metadata` and no `on*=` attribute. The Go code renders them with `templ.Raw` only after that check.
- **Favicon.** `<link rel="icon">` uses `data:image/svg+xml;base64,…` built in Go from `assets/icon.svg` at render time (base64 avoids escaping).

### D22. Sizes and what changes for the tests

The stripped logo is 5345 bytes (from 13290), and the stripped icon 1288 bytes (from 9191). The batch 2 tests of the palette (`TestReportCSSTokens`, `TestReportCSSBadgeContrast`, `TestReportCSSBrandGradient`) and the copy script test (`TestCopyScriptEmbedded`, one `<script>`) are rewritten for D8, D14 and D20. The old `copy.js` is renamed to `report.js` (`git mv`), and the batch 2 tests of Heroicons paths are removed with `icon.go`.

## Superseded decisions (batch 3, 2026-10-12)

| Decision | Batch 2 | Batch 3 | Reason |
|---|---|---|---|
| D4 | Repo emoji and emoji counters | Colored dot and text chips | User decision a |
| D5 | CSS-only radio tabs | JS tabs, ARIA roles, Left/Right focus | User decision f and the mock |
| D6 | Container query at 40rem, subgrid | Viewport breakpoint at 600px, same column template per row | The mock |
| D8 | Palette A tokens, `--c-*` names, `success` tokens | Mock tokens, three schemes, `--ok` | The mock and user decision b |
| D10 | Heroicons 20/solid, green copy | Mock 14px icons, teal `--ok` | The mock and user decision b |
| D12 | `document.summary.emoji.*` in use | Kept, unused | User decision a |
| D20 theme | Choice stored in `localStorage` `dv-theme`, applied by `theme.js` before paint | Not stored; every open starts on Automático; one script | User decision, 2026-10-12 |
| D14 | Palette A contrast ratios | Mock tokens, 5.08:1 minimum in light | The mock |
| Brand gradient | Header accent line | Removed; the logo carries the brand | The mock has no accent line |

## Risks / Trade-offs

- [Old iOS or Telegram browser lacks `backdrop-filter`] → The glass background is 55% white (light) or 5% white (dark) over the mesh, so the text stays readable without blur. The user checks iOS, iPadOS and macOS (task 12.1).
- [Google Fonts request reveals the reader's IP to Google] → Accepted by the user; the system font fallback keeps the report readable when the request fails.
- [Inline scripts and `templ.Raw` could carry unsafe content] → Only build-time constants are raw. The scripts are linted, and the tests check the embedded bytes.
- [The white count on the selected tab badge is 3.65:1] → Kept as in the mock; recorded in D8 and in the open questions.
- [A stored theme could open the report in the wrong scheme, as the shared `dv-theme` key did on `file://`] → Nothing is stored, and every open starts on Automático (D20).
- [templ is pre-1.0; minor releases can change generated code] → The version is pinned in `go.mod` and in the generation script.
- [Generated code drifts from the templates] → The shared `generated-drift` step and `verify.sh` (D11, D17).
- [gocognit does not skip generated `*_templ.go` files by default] → `-ignore '_templ\.go$'` in `verify.sh` and the shared `cognitive` step.
- [Biome has no published checksum file] → The hash is computed once and checked against the API digest (D16).
- [Counts per dependency do not match the number of entries in a section] → Documented in the spec; the entries show every update.
- [The mock's big number and its sub line disagree] → The mock shows 53 pending and "pendientes de 53" with 52 updates in its summary line. The document uses the listed updates for the big number, the badge and the summary line, and the total dependencies in the sub line.

## Migration Plan

No data, config or API migration. The change ships as one commit: regenerated assets, new components and scripts, updated goldens, removed Bootstrap and palette-A files, `verify.sh`, `biome.json` and the liquid glass files. Rollback is `git revert` of that commit. The workflows changes ship first (D17).

## Open Questions

- The white count on the selected tab (3.65:1 on `#0A84FF`) is below AA for 11px text. The user decides whether to darken the tab count there, or keep the mock's value.
- Whether to embed the Geist fonts later (D9).
- Whether the summary line should show the owner for repos with the same short name in one tab. The expanded card already shows the full name, so this can wait for a real case.

Resolved on 2026-10-12:
- gocognit v1.2.0 has `-ignore <regexp>`; generated `*_templ.go` files are skipped with `'_templ\.go$'`.
- openspec in `verify.sh`: `@fission-ai/openspec@1.14.1` in a pinned `node:<LTS>-slim` image.
- Workflows branch: `feat/go-quality-gates`.
- `@v1` callers: automation-hub, devidence-home, home-ctl, dc-sentinel, QuantWarden (`ci.yml`) and voraz-bot (`release.yml`). New inputs default off.
- The untracked-file check in `generated-drift` and `verify.sh` is kept.
- Batch 3 decisions a to f were accepted by the user as "acepto todo lo del mock" on 2026-10-12.
