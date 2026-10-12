# dependency-report Specification

## Purpose
Receives the Renovate lookup report produced by the in-cluster dependency-report Job and delivers it to a private Telegram chat in a fixed, readable format, without touching any scanned repository.

## Requirements

### Requirement: Authenticated endpoint
`POST /webhook/dependency-report` SHALL exist only when `dependency_report:` is configured. It MUST compare the `X-Report-Token` header to the configured token in constant time. A missing or wrong token SHALL get 401, and nothing is sent to Telegram.

#### Scenario: Missing token
- **WHEN** a POST arrives without `X-Report-Token`
- **THEN** the response is 401 and no Telegram message is sent

#### Scenario: Not configured
- **WHEN** the config has no `dependency_report:` section
- **THEN** `POST /webhook/dependency-report` returns 404

### Requirement: Request handling
The endpoint SHALL accept `multipart/form-data` with fields `exit_code`, `pod`, file `log_tail` and optional file `report`, and SHALL enforce a maximum body size. A malformed request SHALL get 400. A valid request SHALL get 202 at once, and the Telegram messages are then sent in the background.

#### Scenario: Valid request
- **WHEN** a POST with a valid token, `exit_code=0` and a `report` file arrives
- **THEN** the response is 202 before any Telegram message is sent

#### Scenario: Malformed exit code
- **WHEN** `exit_code` is not an integer
- **THEN** the response is 400 and nothing is sent

### Requirement: Renovate failure message
If `exit_code` is not `0` or `report` is missing, the system SHALL send one message, `❌ Renovate falló (exit <N>) en el pod <pod>`, followed by the HTML-escaped `log_tail` inside `<pre>`. The log keeps its last lines that fit within 4096 characters.

#### Scenario: Exit code 1
- **WHEN** `exit_code=1`, `pod=dependency-report-manual-1791763200-x7k2p` and `log_tail` contains `<html>`
- **THEN** the message starts with `❌ Renovate falló (exit 1) en el pod dependency-report-manual-1791763200-x7k2p` and shows `&lt;html&gt;`

### Requirement: Per-repository message
For each repository in the report, the system SHALL list every dependency that has at least one update of type `major`, `minor` or `patch` (`digest` and `pin` are ignored). Each line looks like `<emoji> <code>depName</code> <current> → <new> (<type>), …`, with 🔴 when any of those updates is `major` and 🟡 otherwise. Duplicate lines are removed. The message SHALL start with `<b>📦 <owner/repo></b>`, followed by ` (fork, solo informativo)` when the repo is in the configured `dependency_report.forks`. When `libYearsWithStatus` is present, it SHALL then show `<outdated>/<total> desactualizadas · <libyears rounded to 1 decimal, without a trailing .0> libyears`, then `⚠️ <N> errores en el scan` when the repo has N problems with `level >= 50`, then a blank line and at most 40 dependency lines, then `… y <M> más`. A repo with no listed dependencies and no errors SHALL produce no message.

#### Scenario: Major update
- **WHEN** `devidence-dev/QuantWarden` has dep `golang.org/x/net` at `0.30.0` with a `major` update to `1.0.0`, and `dependencyStatus` is `{outdated: 3, total: 20}` with `libYears.total` `2.345`
- **THEN** its message contains `<b>📦 devidence-dev/QuantWarden</b>`, `3/20 desactualizadas · 2.3 libyears` and a 🔴 line for `golang.org/x/net`

#### Scenario: Fork
- **WHEN** `devidence-dev/discord-tts-bot` is in `dependency_report.forks` and has a `minor` update
- **THEN** its header ends with ` (fork, solo informativo)` and the line uses 🟡

#### Scenario: More than 40 lines
- **WHEN** a repo has 45 dependency lines
- **THEN** the message shows 40 lines followed by `… y 5 más`

#### Scenario: Digest only
- **WHEN** a repo's only updates are of type `digest`
- **THEN** that repo produces no message

### Requirement: All up to date
If no repository produces a message, the system SHALL send `✅ Dependency report: todo al día.`

#### Scenario: Nothing pending
- **WHEN** the report has no `major`, `minor` or `patch` updates and no problems with `level >= 50`
- **THEN** exactly one message, `✅ Dependency report: todo al día.`, is sent

### Requirement: Safe HTML and size limit
All variable text MUST be HTML-escaped (`&`, `<`, `>`), and messages SHALL use HTML parse mode. No message may exceed 4096 characters. Content SHALL be cut by whole lines before escaping, so a message never ends inside a tag or entity.

#### Scenario: Long escaped content
- **WHEN** a repo's lines with `<` and `&` would exceed 4096 characters after escaping
- **THEN** the sent message is at most 4096 characters, contains no partial `&lt;` entity or unclosed tag, and Telegram accepts it

### Requirement: Delivery
Messages SHALL be sent to `dependency_report.telegram_chat_id`. A Telegram failure after retries SHALL be logged as an error. It MUST NOT affect other components.

#### Scenario: Telegram down
- **WHEN** Telegram returns errors for every attempt
- **THEN** the hub logs the failure and keeps serving other webhooks and bot commands
