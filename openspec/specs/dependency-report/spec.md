# dependency-report Specification

## Purpose
Receives the Renovate lookup report produced by the in-cluster dependency-report Job and delivers it to a private Telegram chat in a fixed, readable format, without touching any scanned repository.

## Requirements

### Requirement: Authenticated endpoint
`POST` to `dependency_report.path` SHALL exist only when `dependency_report:` is configured. It MUST compare the header named by `dependency_report.token_header` to the configured token in constant time. A missing or wrong token SHALL get 401, and nothing is sent to Telegram.

#### Scenario: Missing token
- **WHEN** a POST arrives without the configured token header
- **THEN** the response is 401 and no Telegram message is sent

#### Scenario: Not configured
- **WHEN** the config has no `dependency_report:` section
- **THEN** `POST /webhook/dependency-report` returns 404

#### Scenario: Custom path
- **WHEN** the config has `path: "/hooks/deps"` and a POST to `/hooks/deps` arrives with a valid token and `exit_code=0` and a `report` file
- **THEN** the response is 202, and a POST to `/webhook/dependency-report` returns 404

#### Scenario: Custom token header
- **WHEN** the config has `token_header: "X-Deps-Token"` and a POST carries only `X-Report-Token` with the correct token
- **THEN** the response is 401, and a POST carrying `X-Deps-Token` with the correct token is accepted

### Requirement: Request handling
The endpoint SHALL accept `multipart/form-data` with fields `exit_code`, `pod`, file `log_tail` and optional file `report`. A body larger than `dependency_report.max_body_bytes` SHALL get 413. A malformed request SHALL get 400. A valid request SHALL get 202 at once, and the Telegram messages are then sent in the background.

#### Scenario: Valid request
- **WHEN** a POST with a valid token, `exit_code=0` and a `report` file arrives
- **THEN** the response is 202 before any Telegram message is sent

#### Scenario: Malformed exit code
- **WHEN** `exit_code` is not an integer
- **THEN** the response is 400 and nothing is sent

#### Scenario: Body over the limit
- **WHEN** the config has `max_body_bytes: 1048576` and a POST body of 2 MiB arrives with a valid token
- **THEN** the response is 413 and nothing is sent

### Requirement: Renovate failure message
If `exit_code` is not `0` or `report` is missing, the system SHALL send one text message: the catalog text `failure.title`, which by default is `❌ Renovate failed (exit {exit}) in pod {pod}` with the exit code and the HTML-escaped pod name, followed by the HTML-escaped `log_tail` inside `<pre>`. The log keeps its last lines that fit within 4096 characters.

#### Scenario: Exit code 1
- **WHEN** `exit_code=1`, `pod=dependency-report-manual-1791763200-x7k2p` and `log_tail` contains `<html>`
- **THEN** the message starts with `❌ Renovate failed (exit 1) in pod dependency-report-manual-1791763200-x7k2p` and shows `&lt;html&gt;`

#### Scenario: Title override
- **WHEN** the config has `messages: {failure: {title: "❌ Renovate failed (exit {exit}) in pod {pod}"}}` and `exit_code=1` with `pod=dependency-report-manual-1791763200-x7k2p`
- **THEN** the message starts with `❌ Renovate failed (exit 1) in pod dependency-report-manual-1791763200-x7k2p`

### Requirement: Safe HTML and size limit
All variable text in the Renovate failure message and in the caption MUST be HTML-escaped (`&`, `<`, `>`), and both SHALL use HTML parse mode. The failure message MUST NOT exceed 4096 characters, and the caption MUST NOT exceed 1024 characters. Content SHALL be cut by whole lines, so a message or caption never ends inside a tag or entity.

#### Scenario: Long escaped content
- **WHEN** a failure log tail with lines containing `<` and `&` would exceed 4096 characters after escaping
- **THEN** the sent message is at most 4096 characters, contains no partial `&lt;` entity or unclosed tag, and Telegram accepts it

#### Scenario: Caption over the limit
- **WHEN** the caption with all its repository lines would exceed 1024 characters
- **THEN** the caption is cut by repository lines as described in "Summary caption", and it has at most 1024 characters with no partial entity or unclosed tag

### Requirement: Delivery
The Renovate failure message SHALL be sent as one text message. A successful report SHALL be sent to `dependency_report.telegram_chat_id` as one document, with the caption as its caption. If the document cannot be sent after retries, the caption SHALL be sent as a text message instead. A failure to send after retries SHALL be logged as an error. A failed send MUST NOT affect other components.

#### Scenario: Telegram down
- **WHEN** Telegram returns errors for every attempt
- **THEN** the hub logs the failure and keeps serving other webhooks and bot commands

#### Scenario: Document rejected
- **WHEN** Telegram returns an error for every `sendDocument` attempt for `dependency-report-2026-10-12.html`
- **THEN** the hub logs an error for the document, sends the caption as a text message so the summary still arrives, and keeps serving other webhooks and bot commands

#### Scenario: Renovate failed
- **WHEN** `exit_code=1` and a `report` file is present
- **THEN** one text message, the failure message, is sent and no document is sent

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

### Requirement: Caption repository lines
Each repository line SHALL be `<emoji> <name><fork tag> — <counts>`. The name is the last path segment, and the fork tag is `fork.tag_caption` for a repository in `dependency_report.forks`. The emoji is 🔴 when any dependency has a `major` update, otherwise 🟡. The counts are the non-zero `<N> major`, `<N> minor` and `<N> patch`, joined by ` · `, each dependency counted once under its highest update type. Lines are sorted by `major` presence, then total count descending, then name.

#### Scenario: Major, minor and patch counts
- **WHEN** `devidence-dev/QuantWarden` has `github.com/gorilla/mux` with a `major` and a `minor` update, `golang.org/x/net` with a `major` update and `go.uber.org/zap` with a `patch` update
- **THEN** its line is `🔴 QuantWarden — 2 major · 1 patch`, because `mux` counts once, as major

#### Scenario: Fork and sort order
- **WHEN** repositories `X` (one `major` and four `minor` deps), `Y` (one `major`), `A` (three `patch`), `B` (three `minor`) and `Z` (two `patch`) have pending updates, and `devidence-dev/discord-tts-bot` is in `dependency_report.forks` with one `minor` update
- **THEN** the repository lines are, in order, `X`, `Y`, `A`, `B`, `Z`, and then `🟡 discord-tts-bot (fork) — 1 minor`

#### Scenario: Owner-agnostic name
- **WHEN** repository `acme-corp/payments-api` has a `patch` update
- **THEN** its line is `🟡 payments-api — 1 patch`, with no owner prefix

#### Scenario: Escaped repository name
- **WHEN** a repository is named `devidence-dev/o<r>` and has a `patch` update
- **THEN** its line shows `o&lt;r&gt;`, and no raw `<r>` appears in the caption

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

### Requirement: Safe document content
All variable text in the document (repo, package file, dependency, versions, error messages) MUST be escaped for HTML, so that report content is shown as text and never interpreted as markup or script.

#### Scenario: Script in a dependency name
- **WHEN** a dep is named `<script>alert(1)</script>` and has a `patch` update from `1.0.0` to `1.0.1`
- **THEN** the document contains `&lt;script&gt;alert(1)&lt;/script&gt;` and contains no `<script>alert(1)</script>`

#### Scenario: Quote in a version
- **WHEN** a current version is `1.0.0" onmouseover="alert(1)`
- **THEN** the copy-button attribute that holds it stays inside its quotes and no `onmouseover` attribute is created

### Requirement: Webhook settings
The `dependency_report:` section SHALL accept `path`, `token_header`, `max_body_bytes`, `document_name` and `date_layout`, with the defaults in the "Defaults" scenario. Config loading MUST fail when `path` does not start with `/`, when `token_header` is empty or contains a space, when `max_body_bytes` is not a positive integer, or when `document_name` lacks `{date}` or contains `/` or `\`. The error SHALL name the setting.

#### Scenario: Defaults
- **WHEN** `dependency_report:` sets only `webhook_token`, `telegram_chat_id` and `forks`
- **THEN** the route is `/webhook/dependency-report`, the token header is `X-Report-Token`, the body limit is 10485760 bytes (10 MiB), the date layout is `2006-01-02`, and a report received on 2026-10-12 is named `dependency-report-2026-10-12.html`

#### Scenario: Invalid path
- **WHEN** the config has `path: "webhook/deps"`
- **THEN** config loading fails with an error that names `dependency_report.path`, and the process exits before it starts the server

#### Scenario: Document name without the date
- **WHEN** the config has `document_name: "report.html"`
- **THEN** config loading fails with an error that names `dependency_report.document_name`

#### Scenario: Zero body limit
- **WHEN** the config has `max_body_bytes: 0`
- **THEN** config loading fails with an error that names `dependency_report.max_body_bytes`

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
- **THEN** its summary row shows `discord-tts-bot`, the note `fork · informational only` and the chip `1 minor`

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
The document SHALL use three color schemes from one set of tokens: light by default, dark when the system prefers dark and Auto is selected, and dark when Dark is selected for the page view. The tokens SHALL be declared on `:root`, redefined under `@media (prefers-color-scheme: dark)` for `:root:not([data-theme=light])`, and redefined for `:root[data-theme=dark]`. Every text color SHALL reach a contrast of at least 4.5:1 against its background in all three schemes.

#### Scenario: Light tokens
- **WHEN** the compiled CSS is read
- **THEN** the `:root` block defines the background as `#EDF3F8`, the glass as `rgba(255,255,255,.55)` and the ink as `#0B1B25`

#### Scenario: Dark tokens in both dark blocks
- **WHEN** the compiled CSS is read
- **THEN** the media block for `:root:not([data-theme=light])` and the block for `:root[data-theme=dark]` each define the background as `#04101A` and the ink as `#EAF2F8`, and the rendered document has no `dark:` variant class

#### Scenario: Text contrast
- **WHEN** each text token of the light and dark blocks is read, including the secondary text of the file folder (`--muted-2`)
- **THEN** its contrast against the page background and against the card is at least 4.5:1, with `--muted-2` at `#5F6F7C` in light and `#7D94A4` in dark

#### Scenario: Selected tab count
- **WHEN** the "Dependencies" tab is selected
- **THEN** its count badge shows white `#FFFFFF` text on `#0A63C2` in light and in dark, a contrast of 5.87:1

#### Scenario: Chip text on its tint
- **WHEN** the chips of the types `major`, `minor` and `patch` and of the errors card are read in the light and the dark scheme
- **THEN** each text color against its tinted background, blended over the card, is at least 4.5:1

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

### Requirement: Motion
The document SHALL animate the pill of the selected tab and of the theme control, the panel that appears, the chevron of an open card and the theme change, and SHALL turn all of these off under `prefers-reduced-motion: reduce`.

#### Scenario: Reduced motion
- **WHEN** the user prefers reduced motion
- **THEN** the pills move without transition, the panel appears without animation, the chevron turns without transition and the theme changes without a view transition

#### Scenario: Panel enters
- **WHEN** the user shows the other tab without reduced motion
- **THEN** the panel animates from 24px to the side in its direction of travel, with opacity from 0 to 1 and blur from 4px to 0, over 420 ms

### Requirement: Header and logo
The header SHALL show the devidence logo as one inline SVG with `role="img"` and the label `devidence.dev`, the scan timestamp in a stamp, the export button (see "Export button") and the theme control. The logo SHALL take its wordmark colors from the tokens of the active scheme, and the document SHALL NOT reference any image file.

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

### Requirement: Export button
The header SHALL show an export button, labeled by `document.export.button`, when at least one update is listed in either tab. It SHALL open the export dialog as a modal. Up to 600px the button SHALL show its icon only and keep the label as its accessible name. When no update is listed, the document SHALL have no export button and no export dialog.

#### Scenario: Updates listed
- **WHEN** the document of a report where `devidence-dev/homelab` has the image `ghcr.io/immich-app/immich-server` from `v1.118.0` to `v1.119.0` is rendered
- **THEN** the header has one button named `Export`, and activating it opens the export dialog as a modal with focus inside it

#### Scenario: Nothing listed
- **WHEN** the report has 12 docker deps and 30 non-docker deps, none with a `major`, `minor` or `patch` update
- **THEN** the document has no `Export` button and no export dialog

#### Scenario: Phone width
- **WHEN** the document is shown in a 320 CSS pixel wide viewport
- **THEN** the header shows the export button as an icon, named `Export`, with no horizontal page scrolling

### Requirement: Export selection
The export dialog SHALL have one section for images (`document.tab.images`) and one for dependencies (`document.tab.deps`), in that order, each with a section checkbox and one checkbox per repo with listed updates in that tab, in the order of that tab's repo cards. A section without listed updates SHALL be left out. Every checkbox SHALL start checked on each open.

#### Scenario: Both sections
- **WHEN** images are listed for `devidence-dev/homelab` and dependencies for `devidence-dev/QuantWarden` and `devidence-dev/homelab`
- **THEN** the dialog shows `Images` with the repo `homelab`, then `Dependencies` with `QuantWarden` and `homelab` in card order, all checked

#### Scenario: Section toggles its repos
- **WHEN** the user clears the `Dependencies` section checkbox
- **THEN** every repo checkbox of `Dependencies` is cleared and the `Images` checkboxes keep their state

#### Scenario: Mixed section
- **WHEN** `Dependencies` lists `QuantWarden` and `homelab` and the user clears only `homelab`
- **THEN** the `Dependencies` section checkbox shows the mixed (indeterminate) state, and checking it again checks both repos

#### Scenario: Only dependencies listed
- **WHEN** no image update is listed and `devidence-dev/QuantWarden` has a dependency update
- **THEN** the dialog shows only the `Dependencies` section

#### Scenario: Reopened dialog
- **WHEN** the user clears `homelab`, closes the dialog with Escape and opens it again
- **THEN** every checkbox is checked again

### Requirement: Export actions
The dialog SHALL offer a copy action (`document.export.copy`) and a download action (`document.export.download`), both disabled while no repo is checked, and a close button (`document.export.close`). The close button and Escape SHALL close the dialog.

#### Scenario: Nothing checked
- **WHEN** the user clears both section checkboxes
- **THEN** `Copy YAML` and `Download YAML` are disabled, and checking `homelab` again enables both

#### Scenario: Close
- **WHEN** the dialog is open and the user presses Escape or activates `Close`
- **THEN** the dialog closes and focus returns to the `Export` button

### Requirement: Export copy
The copy action SHALL put the export on the clipboard, falling back to a hidden text area when the clipboard API is missing or rejects the write, and SHALL announce `document.copy.copied` in the polite live region.

#### Scenario: Copy
- **WHEN** the scan was received on 2026-10-12 at 09:07 America/Guayaquil and the user activates `Copy YAML` with every repo checked
- **THEN** the clipboard holds the export of every listed update and `Copied!` is announced

#### Scenario: Clipboard API blocked
- **WHEN** `navigator.clipboard` rejects the write
- **THEN** the export is copied through a hidden text area and `Copied!` is announced

### Requirement: Export download
The download action SHALL save the export as a `text/yaml` file named by `dependency_report.export_name` (default `dependency-updates-{date}.yaml`), with `{date}` replaced by the scan date in the configured timezone, formatted with `dependency_report.date_layout`. `export_name` SHALL follow the same validation as `document_name`.

#### Scenario: Download
- **WHEN** the scan was received on 2026-10-12 at 09:07 America/Guayaquil and the user activates `Download YAML` with every repo checked
- **THEN** the browser saves `dependency-updates-2026-10-12.yaml` with the same content as `Copy YAML` puts on the clipboard

#### Scenario: Date near midnight UTC
- **WHEN** the report is received at 2026-10-13 02:30 UTC, which is 2026-10-12 21:30 in America/Guayaquil
- **THEN** the download is named `dependency-updates-2026-10-12.yaml`

#### Scenario: Configured export name
- **WHEN** the config has `export_name: "renovate-{date}.yaml"` and the scan was received on 2026-10-12 at 09:07 America/Guayaquil
- **THEN** the download is named `renovate-2026-10-12.yaml`

#### Scenario: Invalid export name
- **WHEN** the config has `export_name: "updates.yaml"` (no `{date}`) or `export_name: "a/{date}.yaml"`
- **THEN** loading the config fails with an error naming `dependency_report.export_name`, and the hub does not start with that config

### Requirement: Export content
The export SHALL be a YAML document: `scanned_at` with the header timestamp, then `images` and `dependencies`, each present only when one of its repos is checked. Each is a list of the checked repos in card order, each with `repo` (full name), `fork`, and `files` in card order, each with `path` and `updates` in card order, each with `name`, `current`, `new` and `type` (`major`, `minor` or `patch`). Field names SHALL be fixed and not translated.

#### Scenario: One image
- **WHEN** only `devidence-dev/homelab` is checked under `Images`, with `ghcr.io/immich-app/immich-server` in `kubernetes/apps/immich/values.yaml` from `v1.118.0` to `v1.119.0`, type `minor`, scanned 2026-10-12 09:07 America/Guayaquil
- **THEN** the export is exactly:
  ```yaml
  scanned_at: "2026-10-12 09:07 America/Guayaquil"
  images:
    - repo: "devidence-dev/homelab"
      fork: false
      files:
        - path: "kubernetes/apps/immich/values.yaml"
          updates:
            - name: "ghcr.io/immich-app/immich-server"
              current: "v1.118.0"
              new: "v1.119.0"
              type: "minor"
  ```

#### Scenario: Fork repo
- **WHEN** `devidence-dev/discord-tts-bot` is a configured fork and is checked under `Dependencies`
- **THEN** its entry has `fork: true`

#### Scenario: Unchecked section left out
- **WHEN** every `Images` repo is cleared and `QuantWarden` is checked under `Dependencies`
- **THEN** the export has `scanned_at` and `dependencies` with only `devidence-dev/QuantWarden`, and no `images` key

### Requirement: Safe export content
Every string value of the export SHALL be a YAML double-quoted scalar with `"`, `\` and control characters escaped, so report content stays one value and never adds keys, list items or documents. The export data in the document SHALL be escaped for HTML like any other report text.

#### Scenario: Quote and newline in a version
- **WHEN** a dep has the current version `1.0.0"\n  injected: true`
- **THEN** its `current` value is the one double-quoted scalar `"1.0.0\"\n  injected: true"` and the export has no `injected` key

#### Scenario: Script in a dependency name
- **WHEN** a dep is named `<script>alert(1)</script>` and has a `patch` update from `1.0.0` to `1.0.1`
- **THEN** the document contains no `<script>alert(1)</script>`, and the copied export holds the name `"<script>alert(1)</script>"` as text

### Requirement: Document landmarks
The document SHALL put the header outside and the tabs, the panels and the errors card inside exactly one `<main>` element, and SHALL have exactly one `<h1>`. The export dialog and the live region MAY sit outside `<main>`.

#### Scenario: Fixture document
- **WHEN** the document of the fixture report (`devidence-dev/QuantWarden`, `devidence-dev/homelab`, `devidence-dev/discord-tts-bot`) is rendered
- **THEN** it has one `<main>` that contains both `role="tabpanel"` sections and the tablist, one `<h1>` with the text `Dependency report - Devidence`, and the `<header>` is not inside `<main>`

#### Scenario: Nothing pending
- **WHEN** the report has 12 docker deps and 30 non-docker deps, none with a listed update
- **THEN** the document still has one `<main>` with both panels and no export dialog

#### Scenario: Errors card
- **WHEN** `devidence-dev/QuantWarden` has a problem with `level` 50
- **THEN** the errors card is inside the `<main>`, after the panels

### Requirement: Named tablist
The tablist SHALL have an accessible name from the catalog key `document.tabs.label`, whose English default is `Report sections`. `messages:` in the config MAY override it.

#### Scenario: Default name
- **WHEN** the fixture document is rendered with no `messages:` section
- **THEN** the element with `role="tablist"` has `aria-label="Report sections"`

#### Scenario: Overridden name
- **WHEN** the config has `messages: {document: {tabs: {label: "Secciones del informe"}}}`
- **THEN** the tablist has `aria-label="Secciones del informe"` and no other label of the document changes

#### Scenario: Escaped name
- **WHEN** the override is `Sections "A" & <B>`
- **THEN** the attribute holds `Sections &#34;A&#34; &amp; &lt;B&gt;` and no attribute is added to the tablist

### Requirement: Entry lists
The entries of one package file SHALL be the items of one list, so assistive technology announces their count. Each entry SHALL be one list item, and the list SHALL keep `role="list"` because its bullets are hidden.

#### Scenario: File with three entries
- **WHEN** `go.mod` of `devidence-dev/QuantWarden` lists `golang.org/x/net`, `github.com/gorilla/mux` and `go.uber.org/zap`
- **THEN** the card has one `<ul role="list">` for `go.mod` with three `<li>` items, each with one name, one type chip and one copy button

#### Scenario: Two files
- **WHEN** a repo has `go.mod` and `Dockerfile` with listed updates in different tabs
- **THEN** each file group has its own list, and the file location sits after the list, not inside it

#### Scenario: Entry text unchanged
- **WHEN** the entry `golang.org/x/net` from `0.30.0` to `1.0.0` is rendered
- **THEN** it keeps the hidden labels `Current` and `New` for screen readers and the same visible text as before

### Requirement: Live regions
The document SHALL have one polite live region outside the export dialog and one inside it. Each SHALL be an element with `role="status"` and no separate `aria-live` attribute, because that role is already polite.

#### Scenario: Fixture document
- **WHEN** the fixture document is rendered
- **THEN** it has exactly two elements with `role="status"`, one of them inside the `<dialog>`, and none carries `aria-live`

#### Scenario: Copy announcement
- **WHEN** the user copies `golang.org/x/net@1.0.0`
- **THEN** `Copied!` is written to the region outside the dialog, and when the export dialog is open the copy of the YAML writes `Copied!` to the region inside it

#### Scenario: No updates listed
- **WHEN** nothing is listed
- **THEN** the document has one `role="status"` element, outside any dialog, because there is no export dialog

### Requirement: Visible keyboard focus
Every button of the document, the repo card summary rows and the checkboxes of the export dialog SHALL show a focus indicator when focused with the keyboard: a 2px outline in the `--accent` color, 2px away from the element. A mouse click SHALL NOT show it.

#### Scenario: Tab button
- **WHEN** the user moves the focus to the "Images" tab with the keyboard
- **THEN** the tab shows a 2px `--accent` outline offset by 2px

#### Scenario: Theme and copy buttons
- **WHEN** the user focuses the `Dark` theme button and then a copy button with the keyboard
- **THEN** each shows the same 2px `--accent` outline offset by 2px

#### Scenario: Mouse click
- **WHEN** the user clicks the `Copy name@version` button with a mouse
- **THEN** the button shows no outline

#### Scenario: Disabled export action
- **WHEN** no repo is checked in the export dialog
- **THEN** the `Copy YAML` and `Download YAML` buttons are disabled, cannot take focus, and show the default cursor and no hover surface

### Requirement: Native controls follow the theme
The document SHALL set `color-scheme` to the active scheme, so that scrollbars, checkboxes and the dialog backdrop match the theme: `light` when Light is chosen, `dark` when Dark is chosen, and the system scheme (`light dark`) on Auto.

#### Scenario: Light on a dark system
- **WHEN** the system prefers a dark color scheme and the user presses `Light`
- **THEN** `<html>` has `data-theme="light"` and its computed `color-scheme` is `light`

#### Scenario: Dark on a light system
- **WHEN** the system prefers a light color scheme and the user presses `Dark`
- **THEN** its computed `color-scheme` is `dark`

#### Scenario: Auto
- **WHEN** the document is opened and the user has not pressed a theme button
- **THEN** `<html>` has no `data-theme` and its computed `color-scheme` is `light dark`

#### Scenario: Compiled CSS
- **WHEN** the compiled CSS is read
- **THEN** it declares `color-scheme` for the default `:root`, for `:root[data-theme=light]` and for `:root[data-theme=dark]`

### Requirement: Document weight
The fixture document SHALL NOT exceed 130,000 bytes, its compiled CSS SHALL NOT exceed 22,000 bytes, and an entry row SHALL NOT average more than 1,000 bytes of markup. These budgets are checked by a test, so growth is a deliberate change.

#### Scenario: Fixture within budget
- **WHEN** the fixture report (54 entries across 5 repo cards) is rendered with the default catalog
- **THEN** the document is at most 130,000 bytes, the stylesheet at most 22,000 bytes, and the markup of its 54 entry rows at most 54,000 bytes

#### Scenario: Nothing pending is smaller
- **WHEN** the report has 12 docker deps and 30 non-docker deps, none with a listed update
- **THEN** the document is at most 60,000 bytes and has no entry rows

#### Scenario: Budget exceeded
- **WHEN** a change adds markup that takes the fixture document over 130,000 bytes
- **THEN** the budget test fails and names the measured size and the budget
