## MODIFIED Requirements

### Requirement: Embedded default catalog
The system SHALL embed one default message catalog in the binary, holding every user-facing report text (see "Catalog contents"). Each default SHALL keep today's caption text. Keys the document no longer renders stay in the catalog. The catalog MUST NOT be read from the host filesystem. Telegram's limits are not catalog entries.

#### Scenario: Defaults render today's caption
- **WHEN** the config has no `messages:` section and the report has no pending update on 2026-10-12 with 42 up-to-date dependencies and no images or packages listed
- **THEN** the caption is exactly `📦 <b>Dependency report</b> · 2026-10-12`, an empty line, `✅ Todo al día`, an empty line, `📊 0 desactualizadas · 42 al día` and `🐳 Imágenes: 0 · 📚 Paquetes: 0`

#### Scenario: Default language
- **WHEN** the config has no `messages:` section
- **THEN** the document's `<html>` element has `lang="es"`

#### Scenario: Default fork note and chips
- **WHEN** the config has no `messages:` section and a document shows the repo `QuantWarden` with a `major` and a `patch` update
- **THEN** its dot is major, its chips read `1 major` and `1 patch`, and its fork note, when it is a fork, reads `fork · solo informativo`

#### Scenario: Catalog contents
- **WHEN** the default catalog is loaded
- **THEN** it holds the caption, the failure message, the fork tags, the document and tab labels, the eyebrows, the panel and status texts, the update type names, the theme labels, the errors texts, the copy labels and `lang`

#### Scenario: Unused key override still validates
- **WHEN** the config has `messages: {document: {columns: {dependency: "Paquete"}}}`
- **THEN** config loading succeeds, the key stays in the catalog, and the document does not render that text

#### Scenario: Unused emoji key override still validates
- **WHEN** the config has `messages: {document: {summary: {emoji: {major: "🔥"}}}}`
- **THEN** config loading succeeds, the key stays in the catalog, and the document shows no emoji for the repo
