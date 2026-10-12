# Spec Delta

## ADDED Requirements

### Requirement: HTML report document
After the per-repository messages (or after `✅ Dependency report: todo al día.`), the system SHALL send exactly one HTML document to the configured chat as a Telegram document. The file name is `dependency-report-<YYYY-MM-DD>.html`, where the date is the day the hub received the report in America/Guayaquil. The document's header SHALL show the scan timestamp, defined as the time the hub received the report, as `YYYY-MM-DD HH:MM America/Guayaquil`. The system MUST NOT send a document on the Renovate failure path (`exit_code` not `0`, or `report` missing).

#### Scenario: Successful run
- **WHEN** a valid request with `exit_code=0` and a report with two repos having updates is received on 2026-10-12 at 09:07 America/Guayaquil
- **THEN** two repo messages are sent first, then one document named `dependency-report-2026-10-12.html` whose header shows `2026-10-12 09:07 America/Guayaquil`

#### Scenario: Date near midnight UTC
- **WHEN** the report is received at 2026-10-13 02:30 UTC, which is 2026-10-12 21:30 in America/Guayaquil
- **THEN** the document is named `dependency-report-2026-10-12.html`

#### Scenario: Renovate failed
- **WHEN** `exit_code=1` and a `report` file is present
- **THEN** only the `❌ Renovate falló (exit 1) en el pod <pod>` message is sent and no document is sent

#### Scenario: Report file missing
- **WHEN** `exit_code=0` and the `report` file is missing
- **THEN** the failure message is sent and no document is sent

### Requirement: Document tabs
The document SHALL have two tabs. "Imágenes" lists every dependency whose Renovate `datasource` is `docker`. "Dependencias" lists every other dependency, including one with no `datasource`. Only updates of type `major`, `minor` or `patch` are listed (`digest` and `pin` are ignored), one row per dependency and update, with exact duplicates removed. A row in "Imágenes" shows the repo, the package file, the current image (`<depName>:<current version>`), the latest image (`<depName>:<new version>`) and the update type badge. A row in "Dependencias" shows the repo, the package file, the dependency name, the current version, the new version and the update type badge. The repo of a fork listed in `dependency_report.forks` carries the tag ` (fork, solo informativo)`. Each tab SHALL show its totals: dependencies in the tab, dependencies with a listed update, dependencies up to date, and the count of `patch`, `minor` and `major` rows.

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
- **WHEN** a dep's only update is of type `digest`
- **THEN** it has no row in either tab

#### Scenario: Several updates for one dep
- **WHEN** `github.com/gorilla/mux` at `1.8.0` has a `major` update to `2.0.0` and a `minor` update to `1.9.0`
- **THEN** "Dependencias" has two rows for it, one with a `major` badge and one with a `minor` badge

#### Scenario: Fork tag
- **WHEN** `devidence-dev/discord-tts-bot` is in `dependency_report.forks` and has a listed update
- **THEN** its rows show the repo as `devidence-dev/discord-tts-bot (fork, solo informativo)`

#### Scenario: Tab totals
- **WHEN** the report has 3 docker deps, one of them with a `minor` update, and 5 non-docker deps, two of them with a `patch` update and a `major` update
- **THEN** "Imágenes" shows 3 total, 1 with updates, 2 up to date and `minor` 1; "Dependencias" shows 5 total, 2 with updates, 3 up to date, `patch` 1 and `major` 1

### Requirement: Scan errors in the document
When a repo has problems with `level >= 50`, the document SHALL list each of them in an errors section as `<owner/repo>: <msg>`. The section is absent when there are none.

#### Scenario: Lookup failure
- **WHEN** `devidence-dev/QuantWarden` has a problem with `level` 50 and `msg` `Failed to look up dependency`
- **THEN** the document's errors section contains `devidence-dev/QuantWarden: Failed to look up dependency`

#### Scenario: Warning level only
- **WHEN** the only problem of a repo has `level` 40
- **THEN** the document has no errors section

### Requirement: Nothing pending document
When no dependency has a listed update, the system SHALL still send the document after `✅ Dependency report: todo al día.`, showing zero updates in both tabs.

#### Scenario: All up to date
- **WHEN** the report has 12 docker deps and 30 non-docker deps, none with a `major`, `minor` or `patch` update
- **THEN** `✅ Dependency report: todo al día.` is sent, then the document, whose tabs show `0` updates, 12 and 30 up to date, and no table rows

### Requirement: Safe document content
All variable text in the document (repo, package file, dependency, versions, error messages) MUST be escaped for HTML, so that report content is shown as text and never interpreted as markup or script.

#### Scenario: Script in a dependency name
- **WHEN** a dep is named `<script>alert(1)</script>` and has a `patch` update from `1.0.0` to `1.0.1`
- **THEN** the document contains `&lt;script&gt;alert(1)&lt;/script&gt;` and contains no `<script>alert(1)</script>`

#### Scenario: Quote in a version
- **WHEN** a current version is `1.0.0" onmouseover="alert(1)`
- **THEN** the copy-button attribute that holds it stays inside its quotes and no `onmouseover` attribute is created

## MODIFIED Requirements

### Requirement: Delivery
Messages and the HTML document SHALL be sent to `dependency_report.telegram_chat_id`, the messages first. A failure to send a message, or the document, after retries SHALL be logged as an error. A failed message MUST NOT stop the remaining messages or the document, and a failed document MUST NOT affect the messages already sent. A Telegram failure MUST NOT affect other components.

#### Scenario: Telegram down
- **WHEN** Telegram returns errors for every attempt
- **THEN** the hub logs the failure and keeps serving other webhooks and bot commands

#### Scenario: Document rejected
- **WHEN** the two repo messages are delivered and Telegram returns an error for every `sendDocument` attempt for `dependency-report-2026-10-12.html`
- **THEN** the hub logs an error for the document, the two messages stay delivered, and the hub keeps serving other webhooks and bot commands

#### Scenario: Message fails, document still sent
- **WHEN** the first of two repo messages fails after retries and Telegram accepts the second message and the document
- **THEN** the failure is logged, and the second message and the document are still sent
