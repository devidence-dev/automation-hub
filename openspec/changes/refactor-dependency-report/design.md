# Design

## Context

The dependency report is one handler file (`internal/handlers/dependency_report.go`) plus one document file (`dependency_report_document.go`). Parsing, aggregation, two renderers and delivery share types, and the report is aggregated twice (once for the caption, once for the document). Presentation lives in the shared model: `docRow.BadgeClass`, `distItem.Class`. Texts are Go literals in English and Spanish, and the HTML template mixes both languages with `lang="en"`. Timezone is a literal in `main.go` and a second literal, reloaded three times per report, in the document file. `devidence-dev/` is stripped in code. The webhook path, the header, the body limit and the file name are literals.

The hub configuration arrives as one `CONFIG_YAML` value from Infisical (`configs/config.yaml.example`), so every new key must have a default that equals today's behavior. Viper already parses YAML (`go.mod` lists `go.yaml.in/yaml/v3` as indirect); no new module is needed.

See `proposal.md` for the motivation and `specs/` for the behavior contract.

## Goals / Non-Goals

**Goals:**
- Every value the audit listed comes from config with a default equal to today's value.
- One aggregation per report; caption and document read the same `Summary`.
- Each responsibility in its own type: parse, aggregate, present, deliver, orchestrate.
- Texts live in one catalog that `bot-command-registry` can reuse.

**Non-Goals:**
- No new HTTP routes, no change to the Job's form fields or to the 202-then-send order.
- No change to `internal/services/telegram/client.go`; the existing `SendHTML` and `SendDocument` are used as they are.
- No new Go module, no template engine beyond `html/template`.

## Decisions

### D1. Two new packages, the handler stays thin

- `internal/messages`: catalog loader and renderer (generic, reused by bot commands).
- `internal/services/depreport`: domain of the dependency report, split by file:
  - `submission.go`: multipart parser (`ParseSubmission(r, maxBody)`) and the form-field constants `FieldExitCode`, `FieldPod`, `FieldLogTail`, `FieldReport`, the Job↔hub contract.
  - `renovate.go`: Renovate JSON parser (`ParseReport`).
  - `update_type.go`: `UpdateType` enum.
  - `aggregate.go`: `Aggregate(rep, forks) Summary`, called once per report.
  - `presenter.go`: `Presenter` interface, implemented by `caption.go` and `document.go`.
  - `deliver.go`: `ReportSender` interface (the only definition) and `Deliverer`.
  - `failure.go`: the failure message builder with the rune-safe tail cut.
- `internal/handlers/dependency_report.go`: `ServeHTTP` only: token check, parse, choose failure or report path, aggregate, present, answer 202, deliver in the background.

Alternative considered: keep everything in `handlers`. Rejected: the handler would keep parsing, aggregation and rendering, which is the problem being fixed.

### D2. Presentation strategy

`Presenter` has one method, `Present(Summary) ([]byte, error)`. `CaptionPresenter` returns the caption bytes, `DocumentPresenter` returns HTML bytes. The handler holds a slice of presenters and does not know their output types. Both receive the same `Summary` value, so the caption and the document agree by construction.

Alternative considered: a `Renderer` with a generic output type. Rejected: Go generics add complexity for two callers.

### D3. Typed update types and per-tab view models

`UpdateType` is an enum over `major`, `minor`, `patch` and an unknown value (for `digest`, `pin`, and any future type). `Listed()` is true for the first three, `Rank()` orders them, and the string form is the catalog key suffix. Unknown types never reach the aggregate, so the inconsistent defaults (`default:` treated as patch in one place, rank 2 in another) disappear.

`Summary` holds domain data only: rows with repo, file, dep, versions, `UpdateType`, and per-tab totals. CSS classes are gone from it. `DocumentPresenter` builds one `TabView` per tab: `Label`, `Columns` (a column definition list: header text and cell source), `Rows` (cells with an optional class). The template is one `range` over `Tabs`, which renders the columns and cells in order. Images and dependencies differ only by their column lists. The `badge-*` and `dist-*` classes are built in the view model.

### D4. Timezone and dependency injection

`config.Load` validates `timezone` with `time.LoadLocation` and exposes `Config.Location *time.Location` (not from mapstructure). `main.go` passes `cfg.Location` to the scheduler and to the report constructor. The report presenters receive the location in their constructor. `reportTimeZone`, the literal in `main.go` and the repeated loads are removed. The header timestamp layout `2006-01-02 15:04` is a Go format constant, not user-facing text, so it stays in code. `date_layout` (default `2006-01-02`) is used for the file name and the caption date.

Alternative considered: `timezone` as a string passed everywhere. Rejected: each caller would reload and could fail at a different moment.

### D5. Config defaults applied in Go, only when the section exists

All defaults are filled in Go after unmarshal. Viper `SetDefault` on `dependency_report.*` would create the section and turn "not configured" into "configured", breaking the 404 requirement. The defaults for `dependency_report:` apply only when the section is present.

### D6. Message catalog

- Storage: `internal/messages/default.yaml`, embedded with `go:embed`, nested keys (`caption.header`, `caption.emoji.major`, `counts.major`, `document.tab.images`, ...). Each key keeps the text the hub sends today: the report texts are Spanish and the caption header keeps `Dependency report`. The only exception is the document labels (D7). `bot-command-registry` adds `bot.*` keys with their current English texts. The top-level `lang` is `es`, the document language.
- Loading: `messages.Load(overrides map[string]any) (*Catalog, error)` parses the embedded YAML with viper (`viper.New()`, `SetConfigType("yaml")`, `ReadConfig`), flattens both trees to dotted keys, merges overrides leaf by leaf, and validates: unknown keys fail, a non-text where a text is expected fails, and a dropped placeholder fails. Placeholder names are found with `\{[a-z_]+\}`.
- Rendering: `Render(key string, vars map[string]string) string` replaces `{name}` placeholders. A key that the code asks for and that is not in the catalog is a programming error caught by a test that renders every key of every call site.
- Overrides: `messages:` in `CONFIG_YAML`. A file path was considered and rejected: Infisical delivers one value, so a file would need a mount (a cross-repo change) for no benefit.
- Placeholders instead of `%d`: named placeholders let validation check that a translated text keeps its values, which `%d` cannot.
- Escaping: report values are escaped by presenters before insertion. Catalog text is trusted operator config.
- Shared use: `bot-command-registry` adds its keys under `bot.` in `default.yaml`. Its overrides are valid only after this change lands, which is the ordering this change owns.

### D7. Visible default change

One default changes visible text. The user approved it on 2026-10-11:
- Document labels: the template mixes Spanish tab labels with English labels and `lang="en"`. Under the "no es/en mix" rule every label moves to the catalog in Spanish, and `lang` is `es`. The Spanish labels are drafted in task 3.1 and approved visually in task 10.1.

The fork tags do not change visible text. Two entries keep today's texts: `fork.tag_caption` = ` (fork)` for caption lines and `fork.tag_document` = ` (fork, solo informativo)` for document rows.

The HTML `<title>` keeps the brand word through the catalog (`document.title`), so no owner literal is in code.

### D8. Submission parser and limits

`ParseSubmission(r, maxBody)` wraps the body in `http.MaxBytesReader`, calls `ParseMultipartForm(maxBody)`, and returns typed errors (`ErrTooLarge` for 413, `ErrInvalid` for 400). The memory threshold is the body limit itself, as today. `exit_code` parsing and the `report` presence check stay in the parser. The token check stays in the handler, constant time, with the header name from config.

Protocol limits `telegramLimit = 4096` and `captionLimit = 1024` remain constants in `depreport`.

## Risks / Trade-offs

- [Visible text change (D7)] → The label change is explicit in the spec scenarios and listed in the README. Override keys restore the old values.
- [Override typos silently changing wording] → Startup validation rejects unknown keys and dropped placeholders; a test renders every default key.
- [Viper key case and nesting] → Catalog keys are lowercase snake case; the loader test covers a nested override and the lowercasing.
- [Section-gated defaults (D5)] → A test asserts that a config without `dependency_report:` leaves `DependencyReport` nil and the route returns 404.
- [Timezone change moves cron times] → Documented in the README and in the scheduled-dispatch scenario. Defaults keep today's zone.
- [Refactor breaks golden output] → The fixture caption and document are compared with golden strings captured from the current code before the move (task 1.2).
- [Cognitive complexity drift] → Every task names `gocognit -over 15` on the touched package; the final validation runs it on the whole module.

## Migration Plan

1. `add-dependency-report-html` was archived on 2026-10-11, so the MODIFIED blocks apply to its requirements.
2. Capture golden outputs from the current code (caption, document, failure message) before changing anything (task 1.2).
3. Implement the catalog, config keys and the split in the order of `tasks.md`; each step keeps the golden outputs green, except the D7 label change, which is applied in one dedicated step.
4. Deploy with no config change: the defaults equal today's values except the D7 labels. Rollback is a binary revert: the new keys are unknown to the old binary and ignored by viper.

## Open Questions

None. Resolved on 2026-10-11:
1. Fork tag: two keys, `fork.tag_caption` and `fork.tag_document`, with today's texts, so neither output changes (D7).
2. Document labels: Spanish, with `lang="es"` (D7). The exact wording is approved visually in task 10.1.
3. Catalog language: each key keeps its current text (D6).
4. `add-dependency-report-html` is archived; the MODIFIED blocks match its requirements in `openspec/specs/`.
