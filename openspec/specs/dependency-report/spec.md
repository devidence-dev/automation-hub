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
If `exit_code` is not `0` or `report` is missing, the system SHALL send one text message: the catalog text `failure.title`, which by default is `❌ Renovate falló (exit {exit}) en el pod {pod}` with the exit code and the HTML-escaped pod name, followed by the HTML-escaped `log_tail` inside `<pre>`. The log keeps its last lines that fit within 4096 characters.

#### Scenario: Exit code 1
- **WHEN** `exit_code=1`, `pod=dependency-report-manual-1791763200-x7k2p` and `log_tail` contains `<html>`
- **THEN** the message starts with `❌ Renovate falló (exit 1) en el pod dependency-report-manual-1791763200-x7k2p` and shows `&lt;html&gt;`

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
The system SHALL send the HTML report of a successful run as one Telegram document to the configured chat, captioned with the summary caption (see "Summary caption"), and no other message. The file name and the header timestamp SHALL follow the "Defaults for file name, header and labels" scenario. The `lang` and all labels SHALL come from the message catalog. The system MUST NOT send a document on the Renovate failure path.

#### Scenario: Successful run
- **WHEN** a valid request with `exit_code=0` and a report with `devidence-dev/QuantWarden` having pending updates is received on 2026-10-12 at 09:07 America/Guayaquil
- **THEN** one document named `dependency-report-2026-10-12.html` is sent, whose header shows `2026-10-12 09:07 America/Guayaquil`, whose `lang` is `es`, and whose caption starts with `📦 <b>Dependency report</b> · 2026-10-12`

#### Scenario: Defaults for file name, header and labels
- **WHEN** the config has no `timezone` and no `messages:` section, and a report with a `major` update is received at 2026-10-13 02:30 UTC, which is 2026-10-12 21:30 in America/Guayaquil
- **THEN** the document is named `dependency-report-2026-10-12.html`, where `{date}` is the day the hub received the report in the configured `timezone`, formatted with `dependency_report.date_layout`; its header shows the time the hub received it in the form `YYYY-MM-DD HH:MM <timezone>`, here `2026-10-12 21:30 America/Guayaquil`; its `lang` is `es` and its labels are Spanish; and its update badges use Renovate's names `major`, `minor` and `patch`, which are not catalog entries

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
The caption SHALL be sent with HTML parse mode and SHALL contain, in this order: the header `caption.header`, a blank line, one line per repository with a listed update (see "Caption repository lines"), `✅ <N> repos al día` (N = repositories with no pending update) omitted when N is 0, a blank line, the stats lines `📊 <X> desactualizadas · <Y> al día` and `🐳 Imágenes: <a> · 📚 Paquetes: <b>`, and `⚠️ <N> errores en el scan` only when a repository or the report root has a `level >= 50` problem.

#### Scenario: No listed update
- **WHEN** no dependency has a listed update
- **THEN** the caption is the header, a blank line, `✅ Todo al día`, a blank line and the stats lines, with no repository line and no `✅ <N> repos al día` line

#### Scenario: Stats lines across both tabs
- **WHEN** the report of 2026-10-12 has 3 docker deps, one with a `minor` update, and 5 non-docker deps, two of them with a `patch` and a `major` update, and 2 repositories have no pending update
- **THEN** the caption header is the default `📦 <b>Dependency report</b> · {date}` with `{date}` the file-name date, so `📦 <b>Dependency report</b> · 2026-10-12`; the caption has `✅ 2 repos al día` followed by a blank line; the stats count both tabs, so `📊 3 desactualizadas · 5 al día` (1 + 2 with a listed update, 2 + 3 without) and `🐳 Imágenes: 1 · 📚 Paquetes: 2` (1 in "Imágenes", 2 in "Dependencias")

#### Scenario: Error count line
- **WHEN** `devidence-dev/QuantWarden` has a problem with `level` 50, the report root has a problem with `level` 50, and another repository has only a problem with `level` 40
- **THEN** the caption's last line is `⚠️ 2 errores en el scan`, and the `level` 40 problem adds nothing to the caption

#### Scenario: Caption over the limit
- **WHEN** 60 repositories with long names have pending updates and the caption with all its repository lines would exceed 1024 characters
- **THEN** repository lines are dropped from the end, a line `… y <N> repos más (ver reporte)` follows the last kept one, and the header, the `✅` line and the stats lines are kept, so that the caption has at most 1024 characters

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
The document SHALL have two tabs: "Imágenes" lists every dependency whose Renovate `datasource` is `docker`, and "Dependencias" lists every other one. Only `major`, `minor` or `patch` updates are listed, one row per dependency and update, as the update-type scenarios show. Rows show the repo, package file, values and badge. The repo of a fork listed in `dependency_report.forks` carries the catalog text `fork.tag_document`. Each tab SHALL show its totals, as in the "Tab totals" scenario.

#### Scenario: Image dependency
- **WHEN** `devidence-dev/homelab` has the file `kubernetes/apps/media/immich/values.yaml` with dep `ghcr.io/immich-app/immich-server`, datasource `docker`, current `v1.118.0` and a `minor` update to `v1.119.0`
- **THEN** the "Imágenes" tab has a row with `devidence-dev/homelab`, `kubernetes/apps/media/immich/values.yaml`, `ghcr.io/immich-app/immich-server:v1.118.0`, `ghcr.io/immich-app/immich-server:v1.119.0` and a `minor` badge, and the "Dependencias" tab has no row for it

#### Scenario: Non-image dependency
- **WHEN** `devidence-dev/QuantWarden` has `go.mod` with dep `golang.org/x/net`, datasource `go`, current `0.30.0` and a `major` update to `1.0.0`
- **THEN** the "Dependencias" tab has a row with `devidence-dev/QuantWarden`, `go.mod`, `golang.org/x/net`, `0.30.0`, `1.0.0` and a `major` badge, and the "Imágenes" tab has no row for it

#### Scenario: Dockerfile image and missing datasource
- **WHEN** `devidence-dev/discord-tts-bot` has a `Dockerfile` dep `golang` with datasource `docker`, and another dep with a `patch` update has no `datasource` field
- **THEN** `golang` is listed in "Imágenes" and the other dep is listed in "Dependencias"

#### Scenario: Digest only
- **WHEN** a dep's only update is of type `digest` or `pin`
- **THEN** it has no row in either tab

#### Scenario: Several updates for one dep
- **WHEN** `github.com/gorilla/mux` at `1.8.0` has a `major` update to `2.0.0` and a `minor` update to `1.9.0`
- **THEN** "Dependencias" has two rows for it, one with a `major` badge and one with a `minor` badge

#### Scenario: Exact duplicates
- **WHEN** `go.uber.org/zap` at `1.27.0` has the same `patch` update to `1.27.1` listed twice
- **THEN** "Dependencias" has one row for it

#### Scenario: Fork tag
- **WHEN** `devidence-dev/discord-tts-bot` is in `dependency_report.forks` and has a listed update
- **THEN** its rows show the repo as `devidence-dev/discord-tts-bot (fork, solo informativo)`

#### Scenario: Tab totals
- **WHEN** the report has 3 docker deps, one of them with a `minor` update, and 5 non-docker deps, two of them with a `patch` update and a `major` update
- **THEN** "Imágenes" shows 3 total, 1 with updates, 2 up to date and `minor` 1; "Dependencias" shows 5 total, 2 with updates, 3 up to date, `patch` 1 and `major` 1

#### Scenario: Document fork tag override
- **WHEN** the config has `messages: {fork: {tag_document: " (fork)"}}` and `devidence-dev/discord-tts-bot` is in `dependency_report.forks` with a listed update
- **THEN** its rows show the repo as `devidence-dev/discord-tts-bot (fork)`, and the caption line is unchanged

### Requirement: Scan errors in the document
When a repo or the report root has problems with `level >= 50`, the document SHALL list each of them in an errors section: `<owner/repo>: <msg>` for a repo, and `raíz: <msg>` for the report root (`raíz` is the catalog text `document.errors.root`). The caption's error count uses the same problems. The section is absent when there are none.

#### Scenario: Lookup failure
- **WHEN** `devidence-dev/QuantWarden` has a problem with `level` 50 and `msg` `Failed to look up dependency`
- **THEN** the document's errors section contains `devidence-dev/QuantWarden: Failed to look up dependency`

#### Scenario: Root problem
- **WHEN** the report root has a problem with `level` 50 and `msg` `Fatal error outside any repo`
- **THEN** the document's errors section contains `raíz: Fatal error outside any repo`, and the caption shows the error count including it

#### Scenario: Warning level only
- **WHEN** the only problem of a repo has `level` 40
- **THEN** the document has no errors section

### Requirement: Nothing pending document
When no dependency has a listed update, the system SHALL still send the document, with `✅ Todo al día` in its caption in place of the repository lines, showing zero updates in both tabs.

#### Scenario: All up to date
- **WHEN** the report has 12 docker deps and 30 non-docker deps, none with a `major`, `minor` or `patch` update
- **THEN** one document is sent whose caption contains `✅ Todo al día` and `📊 0 desactualizadas · 42 al día`, and whose tabs show `0` updates, 12 and 30 up to date, and no table rows

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
