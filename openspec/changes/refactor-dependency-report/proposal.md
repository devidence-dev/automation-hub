# Proposal

## Why

The dependency report hardcodes its timezone, an owner prefix, the webhook path and header, its size limit and about 40 Spanish strings. Parsing, aggregation, rendering and delivery share one handler file, and the report is aggregated twice. Operators cannot adapt the hub without a code change, and the bot-command change needs the same texts. This change makes those values configurable, with defaults equal to today's behavior, and splits the code by responsibility.

## Dependency

Requires `add-dependency-report-html` to be archived first. Its requirements are the baseline for the MODIFIED blocks here. It is in progress (20 of 30 tasks).

## What Changes

- New top-level `timezone:` (default `America/Guayaquil`), validated with `time.LoadLocation` at startup. It drives the scheduler, caption date, file name and document header.
- `dependency_report:` gains `path`, `token_header`, `max_body_bytes`, `document_name` and `date_layout`, defaulting to today's values.
- New `messages:` catalog: an embedded Spanish default in `internal/messages` holds every report text, emoji and HTML label, and `lang`. Config overrides single keys; unknown keys and missing placeholders fail at startup.
- One fork tag, `fork.tag`, default ` (fork, solo informativo)`. **Visible change**: caption fork lines gain ", solo informativo". Set the key to ` (fork)` to keep the old caption.
- The `devidence-dev/` strip is removed; the short name is the last path segment.
- Internal split: parser, single aggregation, caption and document presenters behind one interface, deliverer, typed `UpdateType`, per-tab view models, one column-driven tab template, named form-field constants, one `ReportSender`.

## Capabilities

### New Capabilities

- `message-catalog`: embedded default texts, config overrides, startup validation.

### Modified Capabilities

- `dependency-report`: configurable path, header, body limit, file name and timezone; catalog texts; fork tag.
- `scheduled-dispatch`: schedules use `timezone:`.
- `job-commands`: scheduled jobs use `timezone:`.

## Impact

- Code: `internal/messages` (new), `internal/config`, `internal/handlers/dependency_report*.go`, a report package (parser, aggregator, presenters, deliverer), `cmd/automation-hub/main.go`, the HTML template, tests.
- Docs: `configs/config.yaml.example`, `README.md`.
- No new Go dependency: the catalog is parsed with viper, already in `go.mod`.
- No cross-repo change: Job field names, token header and path keep their defaults.

## Non-goals

- No redesign of the document or caption layout.
- No change to the Job-to-hub contract or the Renovate scan scope.
- No catalog file path; overrides live in `messages:`.
- No pluralization, locale switching or per-chat languages.
- No bot-command registry; that belongs to `bot-command-registry`.
