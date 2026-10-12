# dependency-report Specification

## ADDED Requirements

### Requirement: Webhook settings
The `dependency_report:` section SHALL accept `path` (default `/webhook/dependency-report`), `token_header` (default `X-Report-Token`), `max_body_bytes` (default `10485760`, which is 10 MiB), `document_name` (default `dependency-report-{date}.html`) and `date_layout` (default `2006-01-02`). Config loading MUST fail when `path` does not start with `/`, when `token_header` is empty or contains a space, when `max_body_bytes` is not a positive integer, or when `document_name` does not contain `{date}` or contains `/` or `\`. The error SHALL name the setting.

#### Scenario: Defaults
- **WHEN** `dependency_report:` sets only `webhook_token`, `telegram_chat_id` and `forks`
- **THEN** the route is `/webhook/dependency-report`, the token header is `X-Report-Token`, the body limit is 10485760 bytes, and a report received on 2026-10-12 is named `dependency-report-2026-10-12.html`

#### Scenario: Invalid path
- **WHEN** the config has `path: "webhook/deps"`
- **THEN** config loading fails with an error that names `dependency_report.path`, and the process exits before it starts the server

#### Scenario: Document name without the date
- **WHEN** the config has `document_name: "report.html"`
- **THEN** config loading fails with an error that names `dependency_report.document_name`

#### Scenario: Zero body limit
- **WHEN** the config has `max_body_bytes: 0`
- **THEN** config loading fails with an error that names `dependency_report.max_body_bytes`

## MODIFIED Requirements

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

### Requirement: HTML report document
The system SHALL send the HTML report of a successful run as one Telegram document to the configured chat, with the summary caption (see "Summary caption") as the document's caption. A successful run sends no other message. The file name is the `dependency_report.document_name` pattern with `{date}` replaced by the day the hub received the report in the configured `timezone` (default `America/Guayaquil`), formatted with `dependency_report.date_layout`. The document's header SHALL show the scan timestamp, defined as the time the hub received the report, as `YYYY-MM-DD HH:MM <timezone>`. The document's `lang` attribute and all its labels SHALL come from the message catalog, which is Spanish by default. The system MUST NOT send a document on the Renovate failure path (`exit_code` not `0`, or `report` missing).

#### Scenario: Successful run
- **WHEN** a valid request with `exit_code=0` and a report with `devidence-dev/QuantWarden` having pending updates is received on 2026-10-12 at 09:07 America/Guayaquil
- **THEN** one document named `dependency-report-2026-10-12.html` is sent, whose header shows `2026-10-12 09:07 America/Guayaquil`, whose `lang` is `es`, and whose caption starts with `📦 <b>Dependency report</b> · 2026-10-12`

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
The caption SHALL be sent with HTML parse mode. For a report with pending updates it SHALL contain, in this order:
1. the header `caption.header`, which by default is `📦 <b>Dependency report</b> · {date}`, with `{date}` as the date of the file name, and a blank line;
2. one line per repository with at least one `major`, `minor` or `patch` update, in the form `<emoji> <name> — <counts>`;
3. `✅ <N> repos al día`, where N is the number of repositories with no pending update, omitted when N is 0;
4. a blank line;
5. `📊 <X> desactualizadas · <Y> al día`, where X is the number of dependencies with a listed update and Y the number without one, both summed over the two tabs of the document;
6. `🐳 Imágenes: <a> · 📚 Paquetes: <b>`, where a and b are the dependencies with a listed update in "Imágenes" and in "Dependencias";
7. `⚠️ <N> errores en el scan`, only when there are problems with `level >= 50` in a repository or at the report root.

The name is the last path segment of the repository name, for example `payments-api` for `acme-corp/payments-api`, followed by the catalog text `fork.tag` (default ` (fork, solo informativo)`) when the repository is in `dependency_report.forks`. The emoji is 🔴 when any of its dependencies has a `major` update, otherwise 🟡. The counts are the non-zero `<N> major`, `<N> minor` and `<N> patch` in that order, joined by ` · `; each dependency counts once, under its highest update type. Repository lines are sorted: repositories with a major first, then by total count descending, then by name. When no dependency has a listed update, the caption has the header, a blank line, `✅ Todo al día`, a blank line and the stats lines (items 5 to 7). All variable text MUST be HTML-escaped.

#### Scenario: Major, minor and patch counts
- **WHEN** `devidence-dev/QuantWarden` has `github.com/gorilla/mux` with a `major` and a `minor` update, `golang.org/x/net` with a `major` update and `go.uber.org/zap` with a `patch` update
- **THEN** its line is `🔴 QuantWarden — 2 major · 1 patch`, because `mux` counts once, as major

#### Scenario: Fork and sort order
- **WHEN** repositories `X` (one `major` and four `minor` deps), `Y` (one `major`), `A` (three `patch`), `B` (three `minor`) and `Z` (two `patch`) have pending updates, and `devidence-dev/discord-tts-bot` is in `dependency_report.forks` with one `minor` update
- **THEN** the repository lines are, in order, `X`, `Y`, `A`, `B`, `Z`, and then `🟡 discord-tts-bot (fork, solo informativo) — 1 minor`

#### Scenario: Owner-agnostic name
- **WHEN** repository `acme-corp/payments-api` has a `patch` update
- **THEN** its line is `🟡 payments-api — 1 patch`, with no owner prefix

#### Scenario: Escaped repository name
- **WHEN** a repository is named `devidence-dev/o<r>` and has a `patch` update
- **THEN** its line shows `o&lt;r&gt;`, and no raw `<r>` appears in the caption

#### Scenario: Caption over the limit
- **WHEN** 60 repositories with long names have pending updates and the caption with all its repository lines would exceed 1024 characters
- **THEN** repository lines are dropped from the end, a line `… y <N> repos más (ver reporte)` follows the last kept one, and the header, the `✅` line and the stats lines are kept, so that the caption has at most 1024 characters
