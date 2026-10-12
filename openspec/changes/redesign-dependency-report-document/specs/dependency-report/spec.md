## MODIFIED Requirements

### Requirement: Document tabs
The document SHALL have two tabs, "Dependencias" first and "Imágenes" second, each with a count badge of its listed updates. Clicking a tab, or pressing Enter or Space on it, SHALL show its panel. "Dependencias" lists every dependency whose Renovate `datasource` is not `docker`, "Imágenes" every `docker` one. Only `major`, `minor` or `patch` updates are listed.

#### Scenario: Image dependency
- **WHEN** `devidence-dev/homelab` has the file `kubernetes/apps/media/immich/values.yaml` with dep `ghcr.io/immich-app/immich-server`, datasource `docker`, current `v1.118.0` and a `minor` update to `v1.119.0`
- **THEN** the "Imágenes" panel has, in the section `homelab`, under the file `kubernetes/apps/media/immich/values.yaml`, an entry named `ghcr.io/immich-app/immich-server` from `v1.118.0` to `v1.119.0` with a `minor` chip, and "Dependencias" has no entry for it

#### Scenario: Non-image dependency
- **WHEN** `devidence-dev/QuantWarden` has `go.mod` with dep `golang.org/x/net`, datasource `go`, current `0.30.0` and a `major` update to `1.0.0`
- **THEN** the "Dependencias" panel has, in the section `QuantWarden`, under the file `go.mod`, an entry `golang.org/x/net` from `0.30.0` to `1.0.0` with a `major` chip, and "Imágenes" has no entry for it

#### Scenario: Grouping order
- **WHEN** `devidence-dev/QuantWarden` has the package files `go.mod` and `Dockerfile` with listed updates in "Dependencias" and "Imágenes" respectively
- **THEN** each panel has one section `QuantWarden`, and file groups within a section are sorted by package file path

#### Scenario: Dockerfile image and missing datasource
- **WHEN** `devidence-dev/discord-tts-bot` has a `Dockerfile` dep `golang` with datasource `docker`, and another dep with a `patch` update has no `datasource` field
- **THEN** `golang` is listed in "Imágenes" and the other dep is listed in "Dependencias"

#### Scenario: Digest only
- **WHEN** a dep's only update is of type `digest` or `pin`
- **THEN** it has no entry in either tab

#### Scenario: Several updates for one dep
- **WHEN** `github.com/gorilla/mux` at `1.8.0` has a `major` update to `2.0.0` and a `minor` update to `1.9.0`
- **THEN** "Dependencias" has two entries for it, one with a `major` chip and one with a `minor` chip

#### Scenario: Exact duplicates
- **WHEN** `go.uber.org/zap` at `1.27.0` has the same `patch` update to `1.27.1` listed twice
- **THEN** "Dependencias" has one row for it

#### Scenario: Fork tag
- **WHEN** `devidence-dev/discord-tts-bot` is in `dependency_report.forks` and has a listed update
- **THEN** its repo card shows the short name `discord-tts-bot` followed by the note `fork · solo informativo`, and its expanded card shows the full name `devidence-dev/discord-tts-bot`

#### Scenario: Tab totals
- **WHEN** the report has 3 docker deps, one of them with a `minor` update, and 5 non-docker deps, two of them with a `patch` update and a `major` update
- **THEN** "Dependencias" has the badge `2` and its panel shows 5 total and 3 up to date, and "Imágenes" has the badge `1` and shows 3 total and 2 up to date

#### Scenario: Keyboard
- **WHEN** the focus is on "Dependencias" and the user presses Right
- **THEN** the focus moves to "Imágenes" without changing the panel, and Enter or Space then shows it; the selected tab has `aria-selected="true"` and the other has `false`

#### Scenario: Document fork tag override
- **WHEN** the config has `messages: {fork: {tag_document: "fork"}}` and `devidence-dev/discord-tts-bot` is in `dependency_report.forks` with a listed update
- **THEN** its repo card shows the note `fork`, and the caption line is unchanged

### Requirement: HTML report document
The system SHALL send the HTML report of a successful run as one Telegram document to the configured chat, captioned with the summary caption, and no other message. The document SHALL be one self-contained HTML file (see "Self-contained document"). The file name and the header timestamp SHALL follow the "Defaults for file name, header and labels" scenario. The `lang` and all labels SHALL come from the message catalog. The system MUST NOT send a document on the Renovate failure path.

#### Scenario: Successful run
- **WHEN** a valid request with `exit_code=0` and a report with `devidence-dev/QuantWarden` having pending updates is received on 2026-10-12 at 09:07 America/Guayaquil
- **THEN** one document named `dependency-report-2026-10-12.html` is sent, whose header shows `2026-10-12 09:07 America/Guayaquil`, whose `lang` is `es`, and whose caption starts with `📦 <b>Dependency report</b> · 2026-10-12`

#### Scenario: Defaults for file name, header and labels
- **WHEN** the config has no `timezone` and no `messages:` section, and a report with a `major` update is received at 2026-10-13 02:30 UTC, which is 2026-10-12 21:30 in America/Guayaquil
- **THEN** the document is named `dependency-report-2026-10-12.html`; its header shows `2026-10-12 21:30 America/Guayaquil`; its `lang` is `es`; its labels are Spanish; and its update chips use the catalog names `major`, `minor` and `patch`

#### Scenario: Self-contained document
- **WHEN** the successful run of the previous scenario is rendered
- **THEN** the document has exactly one `<style>` element with the compiled CSS, no `<link rel="stylesheet">` other than Google Fonts, no `<script src>`, and exactly two inline scripts: the theme script in `<head>` and the report script at the end of `<body>`; the logo and the favicon are inline

#### Scenario: Fonts blocked
- **WHEN** the document, which MAY load Geist and Geist Mono from Google Fonts, is opened in a browser that cannot reach fonts.googleapis.com
- **THEN** the text is shown in the system font stack with no missing glyph for the Spanish labels and the emoji

#### Scenario: Date near midnight UTC
- **WHEN** the report is received at 2026-10-13 02:30 UTC, which is 2026-10-12 21:30 in America/Guayaquil
- **THEN** the document is named `dependency-report-2026-10-12.html`

#### Scenario: Configured timezone changes the date and the file name
- **WHEN** the config has `timezone: Europe/Madrid` and the report is received at 2026-10-12 22:30 UTC, which is 2026-10-13 00:30 in Europe/Madrid
- **THEN** the document is named `dependency-report-2026-10-13.html`, its header shows `2026-10-13 00:30 Europe/Madrid`, and its caption starts with `📦 <b>Dependency report</b> · 2026-10-13`

#### Scenario: Renovate failed
- **WHEN** `exit_code=1` and a `report` file is present
- **THEN** only the failure message is sent and no document is sent

#### Scenario: Report file missing
- **WHEN** `exit_code=0` and the `report` file is missing
- **THEN** the failure message is sent and no document is sent

### Requirement: Nothing pending document
When no dependency has a listed update, the system SHALL still send the document, with `✅ Todo al día` in its caption in place of the repository lines, showing zero updates in both tabs, the up-to-date status line of each tab and no repo cards.

#### Scenario: All up to date
- **WHEN** the report has 12 docker deps and 30 non-docker deps, none with a `major`, `minor` or `patch` update
- **THEN** one document is sent whose caption contains `✅ Todo al día` and `📊 0 desactualizadas · 42 al día`, whose tabs show badges `0` and `0`, whose panels show 0 updates, 12 and 30 up to date, the status lines `Todas las dependencias están al día` and `Todas las imágenes están al día`, and no repo cards

### Requirement: Scan errors in the document
When a repo or the report root has problems with `level >= 50`, the document SHALL list each of them in one errors card at the bottom: the heading `document.errors.heading`, a chip with the count, and one row per problem with the repo (`<owner/repo>`) or the root (`document.errors.root`) and the message. The caption's error count uses the same problems. The card is absent when there are none.

#### Scenario: Lookup failure
- **WHEN** `devidence-dev/QuantWarden` has a problem with `level` 50 and `msg` `Failed to look up dependency`
- **THEN** the errors card has a row with `devidence-dev/QuantWarden` and `Failed to look up dependency`

#### Scenario: Root problem
- **WHEN** the report root has a problem with `level` 50 and `msg` `Fatal error outside any repo`
- **THEN** the errors card has a row `raíz` with `Fatal error outside any repo`, and the caption shows the error count including it

#### Scenario: Warning level only
- **WHEN** the only problem of a repo has `level` 40
- **THEN** the document has no errors card

#### Scenario: Three errors
- **WHEN** the report has three problems with `level` 50
- **THEN** the card heading is `Errores del análisis`, its chip reads `3 errores`, and it has three rows

## ADDED Requirements

### Requirement: Repo sections
Each repo group of a tab SHALL be one native `<details>` card with no `name` attribute, so each card opens and closes independently. All cards SHALL be collapsed when the document loads. Opening a card MUST NOT close any other card, and the document MUST NOT store which cards are open.

#### Scenario: All repos collapsed on load
- **WHEN** the document has the cards `QuantWarden`, `homelab` and `discord-tts-bot`
- **THEN** each is a `<details>` element with no `open` attribute and no `name` attribute

#### Scenario: Opening one keeps others open
- **WHEN** the user opens `QuantWarden` and then `homelab`
- **THEN** both are open, `QuantWarden` stays open, and no other card changes state

#### Scenario: Reopening the file starts collapsed
- **WHEN** the user opens `QuantWarden`, closes the document and opens the same file again
- **THEN** all three cards are collapsed

### Requirement: Repo summary row
The summary row of a card SHALL show, in order: a dot, the short repo name with the fork note for a fork, one text chip per update type present in the card, and a chevron. The dot is major when any dependency has a `major` update, otherwise minor, otherwise patch. Each chip shows the count and the type name from the catalog, in the color of its type, and no emoji is shown.

#### Scenario: Summary row with a major update
- **WHEN** `devidence-dev/QuantWarden` has `golang.org/x/net` with a `major` update and `go.uber.org/zap` with a `patch` update
- **THEN** its summary row has a major dot, the name `QuantWarden`, the chips `1 major` and `1 patch`, and a chevron

#### Scenario: Summary row without a major update
- **WHEN** `devidence-dev/homelab` has two dependencies with a `minor` update and none with `major` or `patch`
- **THEN** its summary row has a minor dot and the single chip `2 minor`

#### Scenario: Summary row of a fork
- **WHEN** `devidence-dev/discord-tts-bot` is a fork with one `minor` update
- **THEN** its summary row shows `discord-tts-bot`, the note `fork · solo informativo` and the chip `1 minor`

#### Scenario: Summary counts match the caption
- **WHEN** `devidence-dev/QuantWarden` has `github.com/gorilla/mux` with a `major` and a `minor` update, `golang.org/x/net` with a `major` update and `go.uber.org/zap` with a `patch` update
- **THEN** the chips are `2 major` and `1 patch`, the same counts as the caption line `🔴 QuantWarden — 2 major · 1 patch`, with `github.com/gorilla/mux` counted once, under `major`

### Requirement: Responsive layout
The document SHALL be usable from a 320 CSS pixel wide viewport up, with no horizontal page scrolling. No element and no compiled CSS rule MUST set a fixed `min-width` above 320px. No element sets a fixed min-width above 320px; the only min-width is the 601px breakpoint. Checked statically (design D14), approved on the deployed report (task 12.1).

#### Scenario: Phone width
- **WHEN** the compiled CSS and the rendered document of the fixture are searched for `min-width`
- **THEN** no declaration sets a fixed `min-width` above 320px, and no `<table>` element has a fixed `width` or `min-width`

### Requirement: Entry layout by viewport width
Each dependency entry SHALL be one aligned row, with name, versions, type chip and copy button in the same columns as the other rows of its file group, when the viewport is wider than 600px. Up to 600px the entry SHALL stack into two rows: name and versions, then type chip and copy button.

#### Scenario: Aligned rows on a wide viewport
- **WHEN** the compiled CSS is read
- **THEN** the entry row declares the column template `minmax(0,1fr) auto 58px 30px` inside a `min-width: 601px` media query

#### Scenario: Stacked rows up to 600px
- **WHEN** the compiled CSS is read
- **THEN** the default entry row has two columns outside any media query, so the chip and copy button sit in the second row up to 600px

### Requirement: Document theme
The document SHALL use three color schemes, from one set of tokens: the light values by default, the dark values when the system prefers a dark color scheme and Automático is selected, and the dark values when Oscuro is selected for the page view. The tokens SHALL be declared on `:root`, redefined under `@media (prefers-color-scheme: dark)` for `:root:not([data-theme=light])`, and redefined for `:root[data-theme=dark]`.

#### Scenario: Light tokens
- **WHEN** the compiled CSS is read
- **THEN** the `:root` block defines the background as `#EDF3F8`, the glass as `rgba(255,255,255,.55)` and the ink as `#0B1B25`

#### Scenario: Dark tokens in both dark blocks
- **WHEN** the compiled CSS is read
- **THEN** the media block for `:root:not([data-theme=light])` and the block for `:root[data-theme=dark]` each define the background as `#04101A` and the ink as `#EAF2F8`, and the rendered document has no `dark:` variant class

#### Scenario: Text contrast
- **WHEN** each text token of the light and dark blocks is read
- **THEN** its contrast against its base background is at least 4.5:1, except the white count on the selected tab, which is recorded in design D8 as an open question

#### Scenario: Chip colors of one type
- **WHEN** the `major` chip is read in the light block
- **THEN** its text is `#B93A24` on the tinted background of the mock, and in the dark block its text is `#FFA391`

### Requirement: Theme control
The header SHALL show a segmented control with three buttons, Automático, Claro and Oscuro, in a group labeled `document.theme.group`. The pressed button SHALL carry `aria-pressed="true"`. The control SHALL start on Automático on every open, and the choice SHALL NOT be stored: Claro or Oscuro applies to the current page view only, and `report.js` SHALL NOT use `localStorage` or `sessionStorage`. Rationale: the user's request of 2026-10-12, and no storage errors on `file://`.

#### Scenario: Opened after a dark choice
- **WHEN** the user chose Oscuro in an earlier page view and the file is opened again
- **THEN** the document has no `data-theme` on `<html>`, the Automático button is pressed, and the page shows the system scheme

#### Scenario: Choice lasts for the page view
- **WHEN** the user presses Oscuro and then reloads the page
- **THEN** the reloaded page starts on Automático with no `data-theme`, and nothing was written to storage

#### Scenario: No storage access
- **WHEN** `report.js` is read
- **THEN** it contains no `localStorage`, no `sessionStorage` and no `dv-theme` key

#### Scenario: Theme change animates
- **WHEN** the user presses Claro and the browser supports view transitions without reduced motion
- **THEN** the change runs in a view transition and the pill slides to the Claro button

### Requirement: Summary panel
Each tab SHALL show a panel with the eyebrow `document.eyebrow.deps` or `document.eyebrow.images`, the big number of listed updates, the sub line `document.panel.sub`, a stacked bar of the update types and a legend in the order patch, minor, major. The bar SHALL omit zero segments; the legend SHALL keep every type, with zero counts.

#### Scenario: Distribution bar of the deps tab
- **WHEN** the "Dependencias" tab has 4 patch, 46 minor and 2 major updates
- **THEN** the bar has three segments of widths 7.7%, 88.5% and 3.8%, in patch, minor and major order, and the legend reads `patch 4`, `minor 46`, `major 2`

#### Scenario: Zero segment omitted
- **WHEN** the "Imágenes" tab has 1 minor update and no patch or major update
- **THEN** the bar has one segment of width 100% and the legend still reads `patch 0`, `minor 1` and `major 0`

#### Scenario: Summary line
- **WHEN** the "Dependencias" tab has 52 listed updates for 52 dependencies
- **THEN** the line under the panel shows `52 actualizaciones` in bold and `para 52 dependencias`

### Requirement: Motion
The document SHALL animate the pill of the selected tab and of the theme control, the panel that appears, the chevron of an open card and the theme change, and SHALL turn all of these off under `prefers-reduced-motion: reduce`.

#### Scenario: Reduced motion
- **WHEN** the user prefers reduced motion
- **THEN** the pills move without transition, the panel appears without animation, the chevron turns without transition and the theme changes without a view transition

#### Scenario: Panel enters
- **WHEN** the user shows the other tab without reduced motion
- **THEN** the panel animates from 24px to the side in its direction of travel, with opacity from 0 to 1 and blur from 4px to 0, over 420 ms

### Requirement: Header and logo
The header SHALL show the devidence logo as one inline SVG with `role="img"` and the label `devidence.dev`, the scan timestamp in a stamp, and the theme control. The logo SHALL take its wordmark colors from the tokens of the active scheme, and the document SHALL NOT reference any image file.

#### Scenario: Logo follows the theme
- **WHEN** the theme is dark, from the system or from the control
- **THEN** the wordmark fills resolve to `#F2F7FB` and `#9FB6C6`, and in light to `#0B1B25` and `#334C5A`

#### Scenario: Logo and icon are clean
- **WHEN** the embedded `assets/logo.svg` and `assets/icon.svg` are read
- **THEN** neither contains a `<script`, a `<metadata` element or an `on*=` attribute, and the favicon is a `data:image/svg+xml` URI

### Requirement: Copy buttons
Every entry SHALL have one icon-only copy button, named by its catalog label. It copies the new value only: `<depName>@<new version>` for a dependency (`document.copy.new_version`) and `<image>:<new version>` for an image (`document.copy.latest_image`). A successful copy SHALL show a check mark in the `--ok` color for about 2 seconds and announce `document.copy.copied` in a polite live region.

#### Scenario: Image copy
- **WHEN** the user copies the image `ghcr.io/immich-app/immich-server` from `v1.118.0` to `v1.119.0`
- **THEN** the clipboard holds `ghcr.io/immich-app/immich-server:v1.119.0`, the button is named `Copiar imagen nueva`, and the check mark shows for about 2 seconds while `¡Copiado!` is announced

#### Scenario: Image entry has one copy button
- **WHEN** the image entry of the previous scenario is rendered
- **THEN** it has exactly one copy button, and no copy value is `ghcr.io/immich-app/immich-server:v1.118.0`

#### Scenario: Dependency copy
- **WHEN** the user copies `golang.org/x/net` from `0.30.0` to `1.0.0`
- **THEN** the clipboard holds `golang.org/x/net@1.0.0`, the button is named `Copiar nombre@versión`, and `¡Copiado!` is announced

#### Scenario: Clipboard API blocked
- **WHEN** `navigator.clipboard` rejects the write
- **THEN** the text is copied through a hidden text area and the check mark shows while `¡Copiado!` is announced
