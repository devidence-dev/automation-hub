# Proposal

## Why

The dependency report (HTML document, Telegram caption and Renovate failure message) ships Spanish default texts and `lang="es"`, while the rest of the hub (bot replies, code, docs) is English. The user wants the report in English by default, so the whole hub speaks one language.

## What Changes

- **BREAKING (output text):** every default text of the dependency report in `internal/messages/default.yaml` becomes English, and `lang` becomes `en`. This covers the HTML document (title, tabs, eyebrows, panel and status lines, theme labels, copy and export labels, errors card), the caption (`✅ All up to date`, stats, tabs, more repos, errors) and the failure title. The caption header is already English.
- Catalog keys, placeholders and the override mechanism do not change, so an existing `messages:` override keeps working, and Spanish stays available through overrides.
- Spec scenarios that quote Spanish texts are restated with the English defaults; the "defaults have no English labels" check becomes "defaults have no Spanish labels".
- The unarchived `add-report-export` delta is updated to the English export labels (`Export`, `Copy YAML`, `Download YAML`, `Close`), so both changes agree.
- Tests, golden files, README and the project rule in `openspec/config.yaml` that names the report texts as Spanish are updated.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `dependency-report`: requirements whose scenarios quote default texts (document, caption, failure message, tabs, theme, copy buttons, errors, nothing pending, summary panel, repo summary row) now quote the English defaults and `lang="en"`.
- `message-catalog`: the default language scenario becomes `lang="en"`; the scenarios quoting Spanish defaults are restated in English, and the `lang` override scenario now overrides to `es`.

## Non-goals

- No second built-in catalog or language switch at runtime; Spanish is an override, not a bundled locale.
- No change to catalog keys, placeholders, emojis, update type names (`major`, `minor`, `patch`) or the YAML export field names.
- No change to email code forwarding, whose Spanish markers match incoming mail.
- No migration of existing `CONFIG_YAML` overrides; a Spanish override in Infisical keeps winning.

## Impact

- `internal/messages/default.yaml` (texts and `lang`), comments in `report.templ` and `theme.templ` that name Spanish labels, regenerated `*_templ.go`.
- Tests in `internal/services/depreport`, `internal/handlers` and `internal/messages` that assert Spanish strings; golden files `caption.txt`, `caption_up_to_date.txt`, `document.html`.
- README (report description), `openspec/config.yaml` context rule, `openspec/changes/add-report-export` spec delta.
- Recipients see English captions and documents from the next report on.
