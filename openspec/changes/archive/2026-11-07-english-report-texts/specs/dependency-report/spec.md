# Spec Delta

## MODIFIED Requirements

### Requirement: Renovate failure message
If `exit_code` is not `0` or `report` is missing, the system SHALL send one text message: the catalog text `failure.title`, which by default is `❌ Renovate failed (exit {exit}) in pod {pod}` with the exit code and the HTML-escaped pod name, followed by the HTML-escaped `log_tail` inside `<pre>`. The log keeps its last lines that fit within 4096 characters.

#### Scenario: Exit code 1
- **WHEN** `exit_code=1`, `pod=dependency-report-manual-1791763200-x7k2p` and `log_tail` contains `<html>`
- **THEN** the message starts with `❌ Renovate failed (exit 1) in pod dependency-report-manual-1791763200-x7k2p` and shows `&lt;html&gt;`

#### Scenario: Title override
- **WHEN** the config has `messages: {failure: {title: "❌ Renovate failed (exit {exit}) in pod {pod}"}}` and `exit_code=1` with `pod=dependency-report-manual-1791763200-x7k2p`
- **THEN** the message starts with `❌ Renovate failed (exit 1) in pod dependency-report-manual-1791763200-x7k2p`

### Requirement: HTML report document
The system SHALL send the HTML report of a successful run as one Telegram document to the configured chat, captioned with the summary caption, and no other message. The document SHALL be one self-contained HTML file (see "Self-contained document"). The file name and the header timestamp SHALL follow the "Defaults for file name, header and labels" scenario. The `lang` and all labels SHALL come from the message catalog. The system MUST NOT send a document on the Renovate failure path.

#### Scenario: Successful run
- **WHEN** a valid request with `exit_code=0` and a report with `devidence-dev/QuantWarden` having pending updates is received on 2026-10-12 at 09:07 America/Guayaquil
- **THEN** one document named `dependency-report-2026-10-12.html` is sent, whose header shows `2026-10-12 09:07 America/Guayaquil`, whose `lang` is `en`, and whose caption starts with `📦 <b>Dependency report</b> · 2026-10-12`

#### Scenario: Defaults for file name, header and labels
- **WHEN** the config has no `timezone` and no `messages:` section, and a report with a `major` update is received at 2026-10-13 02:30 UTC, which is 2026-10-12 21:30 in America/Guayaquil
- **THEN** the document is named `dependency-report-2026-10-12.html`; its header shows `2026-10-12 21:30 America/Guayaquil`; its `lang` is `en`; its labels are English; and its update chips use the catalog names `major`, `minor` and `patch`

#### Scenario: Self-contained document
- **WHEN** the successful run of the previous scenario is rendered
- **THEN** the document has exactly one `<style>` element with the compiled CSS, no `<link rel="stylesheet">` other than Google Fonts, no `<script src>`, and exactly two inline scripts: the theme script in `<head>` and the report script at the end of `<body>`; the logo and the favicon are inline

#### Scenario: Fonts blocked
- **WHEN** the document, which MAY load Geist and Geist Mono from Google Fonts, is opened in a browser that cannot reach fonts.googleapis.com
- **THEN** the text is shown in the system font stack with no missing glyph for the labels and the emoji

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

### Requirement: Summary caption
The caption SHALL be sent with HTML parse mode and SHALL contain, in this order: the header `caption.header`, a blank line, one line per repository with a listed update (see "Caption repository lines"), `✅ <N> repos up to date` (N = repositories with no pending update) omitted when N is 0, a blank line, the stats lines `📊 <X> outdated · <Y> up to date` and `🐳 Images: <a> · 📚 Packages: <b>`, and `⚠️ <N> scan errors` only when a repository or the report root has a `level >= 50` problem.

#### Scenario: No listed update
- **WHEN** no dependency has a listed update
- **THEN** the caption is the header, a blank line, `✅ All up to date`, a blank line and the stats lines, with no repository line and no `✅ <N> repos up to date` line

#### Scenario: Stats lines across both tabs
- **WHEN** the report of 2026-10-12 has 3 docker deps, one with a `minor` update, and 5 non-docker deps, two of them with a `patch` and a `major` update, and 2 repositories have no pending update
- **THEN** the caption header is the default `📦 <b>Dependency report</b> · {date}` with `{date}` the file-name date, so `📦 <b>Dependency report</b> · 2026-10-12`; the caption has `✅ 2 repos up to date` followed by a blank line; the stats count both tabs, so `📊 3 outdated · 5 up to date` (1 + 2 with a listed update, 2 + 3 without) and `🐳 Images: 1 · 📚 Packages: 2` (1 in "Images", 2 in "Dependencies")

#### Scenario: Error count line
- **WHEN** `devidence-dev/QuantWarden` has a problem with `level` 50, the report root has a problem with `level` 50, and another repository has only a problem with `level` 40
- **THEN** the caption's last line is `⚠️ 2 scan errors`, and the `level` 40 problem adds nothing to the caption

#### Scenario: Caption over the limit
- **WHEN** 60 repositories with long names have pending updates and the caption with all its repository lines would exceed 1024 characters
- **THEN** repository lines are dropped from the end, a line `… and <N> more repos (see report)` follows the last kept one, and the header, the `✅` line and the stats lines are kept, so that the caption has at most 1024 characters

### Requirement: Document tabs
The document SHALL have two tabs, "Dependencies" first and "Images" second, each with a count badge of its listed updates. Clicking a tab, or pressing Enter or Space on it, SHALL show its panel. "Dependencies" lists every dependency whose Renovate `datasource` is not `docker`, "Images" every `docker` one. Only `major`, `minor` or `patch` updates are listed.

#### Scenario: Image dependency
- **WHEN** `devidence-dev/homelab` has the file `kubernetes/apps/media/immich/values.yaml` with dep `ghcr.io/immich-app/immich-server`, datasource `docker`, current `v1.118.0` and a `minor` update to `v1.119.0`
- **THEN** the "Images" panel has, in the section `homelab`, under the file `kubernetes/apps/media/immich/values.yaml`, an entry named `ghcr.io/immich-app/immich-server` from `v1.118.0` to `v1.119.0` with a `minor` chip, and "Dependencies" has no entry for it

#### Scenario: Non-image dependency
- **WHEN** `devidence-dev/QuantWarden` has `go.mod` with dep `golang.org/x/net`, datasource `go`, current `0.30.0` and a `major` update to `1.0.0`
- **THEN** the "Dependencies" panel has, in the section `QuantWarden`, under the file `go.mod`, an entry `golang.org/x/net` from `0.30.0` to `1.0.0` with a `major` chip, and "Images" has no entry for it

#### Scenario: Grouping order
- **WHEN** `devidence-dev/QuantWarden` has the package files `go.mod` and `Dockerfile` with listed updates in "Dependencies" and "Images" respectively
- **THEN** each panel has one section `QuantWarden`, and file groups within a section are sorted by package file path

#### Scenario: Dockerfile image and missing datasource
- **WHEN** `devidence-dev/discord-tts-bot` has a `Dockerfile` dep `golang` with datasource `docker`, and another dep with a `patch` update has no `datasource` field
- **THEN** `golang` is listed in "Images" and the other dep is listed in "Dependencies"

#### Scenario: Digest only
- **WHEN** a dep's only update is of type `digest` or `pin`
- **THEN** it has no entry in either tab

#### Scenario: Several updates for one dep
- **WHEN** `github.com/gorilla/mux` at `1.8.0` has a `major` update to `2.0.0` and a `minor` update to `1.9.0`
- **THEN** "Dependencies" has two entries for it, one with a `major` chip and one with a `minor` chip

#### Scenario: Exact duplicates
- **WHEN** `go.uber.org/zap` at `1.27.0` has the same `patch` update to `1.27.1` listed twice
- **THEN** "Dependencies" has one row for it

#### Scenario: Fork tag
- **WHEN** `devidence-dev/discord-tts-bot` is in `dependency_report.forks` and has a listed update
- **THEN** its repo card shows the short name `discord-tts-bot` followed by the note `fork · informational only`, and its expanded card shows the full name `devidence-dev/discord-tts-bot`

#### Scenario: Tab totals
- **WHEN** the report has 3 docker deps, one of them with a `minor` update, and 5 non-docker deps, two of them with a `patch` update and a `major` update
- **THEN** "Dependencies" has the badge `2` and its panel shows 5 total and 3 up to date, and "Images" has the badge `1` and shows 3 total and 2 up to date

#### Scenario: Keyboard
- **WHEN** the focus is on "Dependencies" and the user presses Right
- **THEN** the focus moves to "Images" without changing the panel, and Enter or Space then shows it; the selected tab has `aria-selected="true"` and the other has `false`

#### Scenario: Document fork tag override
- **WHEN** the config has `messages: {fork: {tag_document: "fork"}}` and `devidence-dev/discord-tts-bot` is in `dependency_report.forks` with a listed update
- **THEN** its repo card shows the note `fork`, and the caption line is unchanged

### Requirement: Scan errors in the document
When a repo or the report root has problems with `level >= 50`, the document SHALL list each of them in one errors card at the bottom: the heading `document.errors.heading`, a chip with the count, and one row per problem with the repo (`<owner/repo>`) or the root (`document.errors.root`) and the message. The caption's error count uses the same problems. The card is absent when there are none.

#### Scenario: Lookup failure
- **WHEN** `devidence-dev/QuantWarden` has a problem with `level` 50 and `msg` `Failed to look up dependency`
- **THEN** the errors card has a row with `devidence-dev/QuantWarden` and `Failed to look up dependency`

#### Scenario: Root problem
- **WHEN** the report root has a problem with `level` 50 and `msg` `Fatal error outside any repo`
- **THEN** the errors card has a row `root` with `Fatal error outside any repo`, and the caption shows the error count including it

#### Scenario: Warning level only
- **WHEN** the only problem of a repo has `level` 40
- **THEN** the document has no errors card

#### Scenario: Three errors
- **WHEN** the report has three problems with `level` 50
- **THEN** the card heading is `Scan errors`, its chip reads `3 errors`, and it has three rows

### Requirement: Nothing pending document
When no dependency has a listed update, the system SHALL still send the document, with `✅ All up to date` in its caption in place of the repository lines, showing zero updates in both tabs, the up-to-date status line of each tab and no repo cards.

#### Scenario: All up to date
- **WHEN** the report has 12 docker deps and 30 non-docker deps, none with a `major`, `minor` or `patch` update
- **THEN** one document is sent whose caption contains `✅ All up to date` and `📊 0 outdated · 42 up to date`, whose tabs show badges `0` and `0`, whose panels show 0 updates, 12 and 30 up to date, the status lines `All dependencies are up to date` and `All images are up to date`, and no repo cards

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
- **THEN** its summary row shows `discord-tts-bot`, the note `fork · informational only` and the chip `1 minor`

#### Scenario: Summary counts match the caption
- **WHEN** `devidence-dev/QuantWarden` has `github.com/gorilla/mux` with a `major` and a `minor` update, `golang.org/x/net` with a `major` update and `go.uber.org/zap` with a `patch` update
- **THEN** the chips are `2 major` and `1 patch`, the same counts as the caption line `🔴 QuantWarden — 2 major · 1 patch`, with `github.com/gorilla/mux` counted once, under `major`

### Requirement: Document theme
The document SHALL use three color schemes, from one set of tokens: the light values by default, the dark values when the system prefers a dark color scheme and Auto is selected, and the dark values when Dark is selected for the page view. The tokens SHALL be declared on `:root`, redefined under `@media (prefers-color-scheme: dark)` for `:root:not([data-theme=light])`, and redefined for `:root[data-theme=dark]`.

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
The header SHALL show a segmented control with three buttons, Auto, Light and Dark, in a group labeled `document.theme.group`. The pressed button SHALL carry `aria-pressed="true"`. The control SHALL start on Auto on every open, and the choice SHALL NOT be stored: Light or Dark applies to the current page view only, and `report.js` SHALL NOT use `localStorage` or `sessionStorage`. Rationale: the user's request of 2026-10-12, and no storage errors on `file://`.

#### Scenario: Opened after a dark choice
- **WHEN** the user chose Dark in an earlier page view and the file is opened again
- **THEN** the document has no `data-theme` on `<html>`, the Auto button is pressed, and the page shows the system scheme

#### Scenario: Choice lasts for the page view
- **WHEN** the user presses Dark and then reloads the page
- **THEN** the reloaded page starts on Auto with no `data-theme`, and nothing was written to storage

#### Scenario: No storage access
- **WHEN** `report.js` is read
- **THEN** it contains no `localStorage`, no `sessionStorage` and no `dv-theme` key

#### Scenario: Theme change animates
- **WHEN** the user presses Light and the browser supports view transitions without reduced motion
- **THEN** the change runs in a view transition and the pill slides to the Light button

### Requirement: Summary panel
Each tab SHALL show a panel with the eyebrow `document.eyebrow.deps` or `document.eyebrow.images`, the big number of listed updates, the sub line `document.panel.sub`, a stacked bar of the update types and a legend in the order patch, minor, major. The bar SHALL omit zero segments; the legend SHALL keep every type, with zero counts.

#### Scenario: Distribution bar of the deps tab
- **WHEN** the "Dependencies" tab has 4 patch, 46 minor and 2 major updates
- **THEN** the bar has three segments of widths 7.7%, 88.5% and 3.8%, in patch, minor and major order, and the legend reads `patch 4`, `minor 46`, `major 2`

#### Scenario: Zero segment omitted
- **WHEN** the "Images" tab has 1 minor update and no patch or major update
- **THEN** the bar has one segment of width 100% and the legend still reads `patch 0`, `minor 1` and `major 0`

#### Scenario: Summary line
- **WHEN** the "Dependencies" tab has 52 listed updates for 52 dependencies
- **THEN** the line under the panel shows `52 updates` in bold and `across 52 dependencies`

### Requirement: Copy buttons
Every entry SHALL have one icon-only copy button, named by its catalog label. It copies the new value only: `<depName>@<new version>` for a dependency (`document.copy.new_version`) and `<image>:<new version>` for an image (`document.copy.latest_image`). A successful copy SHALL show a check mark in the `--ok` color for about 2 seconds and announce `document.copy.copied` in a polite live region.

#### Scenario: Image copy
- **WHEN** the user copies the image `ghcr.io/immich-app/immich-server` from `v1.118.0` to `v1.119.0`
- **THEN** the clipboard holds `ghcr.io/immich-app/immich-server:v1.119.0`, the button is named `Copy new image`, and the check mark shows for about 2 seconds while `Copied!` is announced

#### Scenario: Image entry has one copy button
- **WHEN** the image entry of the previous scenario is rendered
- **THEN** it has exactly one copy button, and no copy value is `ghcr.io/immich-app/immich-server:v1.118.0`

#### Scenario: Dependency copy
- **WHEN** the user copies `golang.org/x/net` from `0.30.0` to `1.0.0`
- **THEN** the clipboard holds `golang.org/x/net@1.0.0`, the button is named `Copy name@version`, and `Copied!` is announced

#### Scenario: Clipboard API blocked
- **WHEN** `navigator.clipboard` rejects the write
- **THEN** the text is copied through a hidden text area and the check mark shows while `Copied!` is announced
