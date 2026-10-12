# message-catalog Specification

## Purpose
Holds the user-facing texts of the hub's Telegram and HTML output in one embedded, overridable catalog, so operators can change wording without a code change and every consumer reads the same source.

## ADDED Requirements

### Requirement: Embedded default catalog
The system SHALL embed one default message catalog in the binary, in Spanish, holding every user-facing text of the dependency report: the caption, the failure message, the document labels, the tab labels, the emoji and `lang`. The catalog MUST NOT be read from the host filesystem. Without a `messages:` section, the rendered texts SHALL equal the texts stated in the `dependency-report` requirements. Protocol limits (Telegram's 4096 and 1024 characters) are not catalog entries.

#### Scenario: Defaults render today's caption
- **WHEN** the config has no `messages:` section and the report has no pending update on 2026-10-12 with 42 up-to-date dependencies and no images or packages listed
- **THEN** the caption is exactly `📦 <b>Dependency report</b> · 2026-10-12`, an empty line, `✅ Todo al día`, an empty line, `📊 0 desactualizadas · 42 al día` and `🐳 Imágenes: 0 · 📚 Paquetes: 0`

#### Scenario: Default language
- **WHEN** the config has no `messages:` section
- **THEN** the document's `<html>` element has `lang="es"`

### Requirement: Configured overrides
The `messages:` section SHALL use the same nested keys as the default catalog. Each key it sets SHALL replace the default text of that key only; every other key keeps its default.

#### Scenario: Caption header override
- **WHEN** the config has `messages:` with `caption: {header: "🔔 <b>Deps</b> · {date}"}` and a report with pending updates is received on 2026-10-12
- **THEN** the caption starts with `🔔 <b>Deps</b> · 2026-10-12`, and its repository lines and stats lines are unchanged

#### Scenario: Language attribute override
- **WHEN** the config has `messages:` with `lang: en`
- **THEN** the document's `<html>` element has `lang="en"`, and its labels stay Spanish because only `lang` was overridden

### Requirement: Startup validation of the catalog
Config loading MUST fail, before the process starts the bot, IMAP monitoring, the HTTP server or the scheduler, when an override names a key that does not exist in the default catalog, when an override is not a text where the default is a text, or when an override drops a placeholder that the default text has. The error SHALL name the key and, for a missing placeholder, the placeholder.

#### Scenario: Unknown key
- **WHEN** the config has `messages: {caption: {headr: "x"}}`
- **THEN** config loading fails with an error that names `caption.headr`

#### Scenario: Missing placeholder
- **WHEN** the config has `messages: {caption: {more_repos: "… más"}}`, and the default text has `{count}`
- **THEN** config loading fails with an error that names `caption.more_repos` and `{count}`

#### Scenario: Map where a text is expected
- **WHEN** the config has `messages: {caption: "Dependency report"}`
- **THEN** config loading fails with an error that names `caption`

### Requirement: Placeholders are filled by code
Catalog texts use named placeholders in braces, such as `{date}`, `{count}`, `{exit}` and `{pod}`. The code MUST fill every placeholder it passes, and a placeholder left unfilled SHALL be an error. Values from the report (repository names, dependency names, versions, log lines, pod names) are HTML-escaped before they are inserted into HTML output. Catalog texts are operator configuration and are inserted as written.

#### Scenario: Pod name is escaped
- **WHEN** a failure is received with `pod` set to `<x>` and `exit_code=1`
- **THEN** the failure message starts with `❌ Renovate falló (exit 1) en el pod &lt;x&gt;`
