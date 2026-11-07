# Spec Delta

## MODIFIED Requirements

### Requirement: Embedded default catalog
The system SHALL embed one default message catalog in the binary, holding every user-facing report text (see "Catalog contents"). Each default SHALL keep today's caption text. Keys the document no longer renders stay in the catalog. The catalog MUST NOT be read from the host filesystem. Telegram's limits are not catalog entries.

#### Scenario: Defaults render today's caption
- **WHEN** the config has no `messages:` section and the report has no pending update on 2026-10-12 with 42 up-to-date dependencies and no images or packages listed
- **THEN** the caption is exactly `📦 <b>Dependency report</b> · 2026-10-12`, an empty line, `✅ All up to date`, an empty line, `📊 0 outdated · 42 up to date` and `🐳 Images: 0 · 📚 Packages: 0`

#### Scenario: Default language
- **WHEN** the config has no `messages:` section
- **THEN** the document's `<html>` element has `lang="en"`

#### Scenario: Default fork note and chips
- **WHEN** the config has no `messages:` section and a document shows the repo `QuantWarden` with a `major` and a `patch` update
- **THEN** its dot is major, its chips read `1 major` and `1 patch`, and its fork note, when it is a fork, reads `fork · informational only`

#### Scenario: Catalog contents
- **WHEN** the default catalog is loaded
- **THEN** it holds the caption, the failure message, the fork tags, the document and tab labels, the eyebrows, the panel and status texts, the update type names, the theme labels, the errors texts, the copy labels and `lang`

#### Scenario: Unused key override still validates
- **WHEN** the config has `messages: {document: {columns: {dependency: "Paquete"}}}`
- **THEN** config loading succeeds, the key stays in the catalog, and the document does not render that text

#### Scenario: Unused emoji key override still validates
- **WHEN** the config has `messages: {document: {summary: {emoji: {major: "🔥"}}}}`
- **THEN** config loading succeeds, the key stays in the catalog, and the document shows no emoji for the repo

### Requirement: Configured overrides
The `messages:` section SHALL use the same nested keys as the default catalog. Each key it sets SHALL replace the default text of that key only; every other key keeps its default.

#### Scenario: Caption header override
- **WHEN** the config has `messages:` with `caption: {header: "🔔 <b>Deps</b> · {date}"}` and a report with pending updates is received on 2026-10-12
- **THEN** the caption starts with `🔔 <b>Deps</b> · 2026-10-12`, and its repository lines and stats lines are unchanged

#### Scenario: Language attribute override
- **WHEN** the config has `messages:` with `lang: es`
- **THEN** the document's `<html>` element has `lang="es"`, and its labels stay English because only `lang` was overridden

### Requirement: Placeholders are filled by code
Catalog texts use named placeholders in braces, such as `{date}`, `{count}`, `{exit}` and `{pod}`. The code MUST fill every placeholder it passes, and a placeholder left unfilled SHALL be an error. Values from the report (repository names, dependency names, versions, log lines, pod names) are HTML-escaped before they are inserted into HTML output. Catalog texts are operator configuration and are inserted as written.

#### Scenario: Pod name is escaped
- **WHEN** a failure is received with `pod` set to `<x>` and `exit_code=1`
- **THEN** the failure message starts with `❌ Renovate failed (exit 1) in pod &lt;x&gt;`
