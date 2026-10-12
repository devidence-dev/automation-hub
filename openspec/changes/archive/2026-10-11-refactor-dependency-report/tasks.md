# Tasks

Every behavior task names the test that proves it. Each touched package keeps every function at cognitive complexity 15 or below (`gocognit -over 15`, tests included). Dependencies: this change starts only after `add-dependency-report-html` is archived (task 1.1).

## 1. Baseline

- [x] 1.1 (user) Archive `add-dependency-report-html` with `openspec archive add-dependency-report-html`. Verify: `openspec list --specs` shows `dependency-report`, and `openspec validate --strict refactor-dependency-report` no longer prints the "MODIFIED failed for header HTML report document" info line.
- [x] 1.2 Capture golden outputs from the current code before any change: caption, document, failure message and the all-up-to-date caption for `internal/handlers/testdata/renovate-report.json`, fixed clock `2026-10-12T14:07:00Z`, fork `devidence-dev/discord-tts-bot`. Store them under `internal/handlers/testdata/golden/`. Verify with `TestGoldenBaseline` (new, in `internal/handlers`) passing on the unchanged code.

## 2. Message catalog loader (`internal/messages`)

- [x] 2.1 Implement `messages.Load(overrides map[string]any) (*Catalog, error)` over an internal `load(defaults []byte, overrides map[string]any)`: parse YAML with viper (`viper.New()`, `SetConfigType("yaml")`, `ReadConfig`), flatten to dotted keys, merge overrides leaf by leaf, reject unknown keys, a map where a text is expected, and a dropped placeholder (`\{[a-z_]+\}`). Verify with table tests `TestLoadNestedOverride` (only `caption.header` changes), `TestLoadRejectsUnknownKey` (error names `caption.headr`), `TestLoadRejectsDroppedPlaceholder` (error names `caption.more_repos` and `{count}`) and `TestLoadRejectsMapWhereTextExpected` (error names `caption`).
- [x] 2.2 Implement `(*Catalog).Render(key string, vars map[string]string) (string, error)`: replace `{name}` placeholders, error on an unknown key or a placeholder without a value. Verify with `TestRenderFillsPlaceholders` (`{date}` becomes `2026-10-12`), `TestRenderMissingValueErrors` and `TestRenderUnknownKeyErrors`.

## 3. Catalog content

- [x] 3.1 Write `internal/messages/default.yaml` (each key with the text the hub sends today, Spanish document labels, nested keys, `lang: es`, embedded with `go:embed`) holding every text now in Go and in the HTML template: `failure.title`, `fork.tag_caption` (` (fork)`), `fork.tag_document` (` (fork, solo informativo)`), `caption.*` (header, all-up-to-date, repos up to date, more repos, repo line, stats, tabs, errors, emoji for major and other), `counts.*`, `document.*` (title, every label, table headers, tab labels, errors label, copy labels). Keys are typed constants in `internal/services/depreport/keys.go`. Verify with `TestDefaultsHaveEveryKey` (every key constant exists in the default catalog and is not empty) and `TestDefaultsKeepPlaceholders` (each placeholder in the defaults is one the code passes).
- [x] 3.2 Remove the English labels from the document and set `lang` from the catalog (design D7). Verify with `TestDocumentDefaultsHaveNoEnglishLabels` (rendered default document contains none of the old English labels, listed in the test) and `TestDocumentLangDefault` (`lang="es"`).

## 4. Configuration (`internal/config`)

- [x] 4.1 Add top-level `timezone` (default `America/Guayaquil`) and the exposed `Location` and `Catalog` fields, `mapstructure:"-"`. `Load` calls `time.LoadLocation` and fails with an error naming `timezone`. Verify with `TestLoadDefaultTimezone` (location `America/Guayaquil`), `TestLoadInvalidTimezoneFails` (`Mars/Olympus` names `timezone`) and `TestLoadEuropeMadrid` (`Europe/Madrid`).
- [x] 4.2 Add `path`, `token_header`, `max_body_bytes`, `document_name` and `date_layout` to `DependencyReportConfig`. Fill defaults in Go after unmarshal, only when the section exists (viper `SetDefault` would create the section). Validate per the spec. Verify with `TestDependencyReportDefaults`, `TestDependencyReportAbsentStaysNil` (no section: `DependencyReport` is nil), `TestDependencyReportInvalidPath`, `TestDependencyReportDocumentNameNeedsDate` (`report.html` names `dependency_report.document_name`) and `TestDependencyReportZeroBodyLimit`.
- [x] 4.3 Wire `messages:` into `Load`: `Messages map[string]any` from `messages`, validated through `messages.Load`. Verify with `TestLoadMessagesUnknownKeyFails` (error names `caption.headr`) and `TestLoadMessagesHeaderOverride` (the catalog renders the new header).

## 5. Parsers (`internal/services/depreport`)

- [x] 5.1 Move the multipart reading into `submission.go`: constants `FieldExitCode = "exit_code"`, `FieldPod = "pod"`, `FieldLogTail = "log_tail"`, `FieldReport = "report"`, and `ParseSubmission(r *http.Request, maxBody int64) (Submission, error)` returning `ErrTooLarge` or `ErrInvalid`. Verify with `TestParseSubmissionValid`, `TestParseSubmissionTooLarge` (`errors.Is(err, ErrTooLarge)`), `TestParseSubmissionBadExitCode` (`ErrInvalid`) and `TestParseSubmissionMissingReport` (no error, `HasReport` false).
- [x] 5.2 Move the Renovate JSON types and `ParseReport` into `renovate.go`, with a fixture test. Verify with `TestParseReportFixture` (repo count and one dep from `testdata/renovate-report.json`) and `TestParseReportInvalidJSON` (error).

## 6. Domain (`internal/services/depreport`)

- [x] 6.1 Add `UpdateType` (`UpdateUnknown` zero value, `UpdateMajor`, `UpdateMinor`, `UpdatePatch`) with `ParseUpdateType`, `Listed()`, `Rank()` and `String()`. Verify with `TestParseUpdateType` (table: `major`, `minor`, `patch`, `digest`, `pin`, empty) and `TestUpdateTypeRank` (major before minor before patch).
- [x] 6.2 Add `Aggregate(rep Report, forks []string, receivedAt time.Time) Summary` (both tabs, pending repos in caption order, repo count, errors) and `ShortName`. Verify with `TestAggregateFixtureTotals` (tab totals of the fixture), `TestPendingOrder` (the spec "Fork and sort order" case, lines `X`, `Y`, `A`, `B`, `Z`), `TestAggregateDigestOnlyRepoHasNoPending`, `TestScanErrorsIncludeRoot` (level 50 listed, level 40 not), and `TestShortName` (`acme-corp/payments-api` gives `payments-api`, a name with no slash is kept).

## 7. Presenters and file name (`internal/services/depreport`)

- [x] 7.1 Add `Presenter` (`Present(Summary) ([]byte, error)`) and `CaptionPresenter` (catalog texts, `fork.tag_caption` from the catalog, `date_layout`, location, fit to 1024 runes). Verify with `TestCaptionFixture` (golden text of the fixture, unchanged from task 1.2), `TestCaptionNothingPending` (the exact text in the message-catalog spec scenario), `TestCaptionRepoLines` (table: emoji, counts, fork tag), `TestCaptionEscapesNames` (`o&lt;r&gt;`), `TestCaptionOverflow` (60 repos: at most 1024 runes, kept lines plus N equal 60), `TestCaptionHeaderOverride` and `TestCaptionDateLayout`.
- [x] 7.2 Add `DocumentPresenter` with per-tab view models (`TabView` with `Label`, `Columns` and `Rows` of cells with an optional class) and one template, `internal/services/depreport/assets/dependency_report.html` (moved with `git mv`), whose tab is one `range` over the column definitions. `lang`, all labels and `fork.tag_document` come from the catalog. Verify with `TestDocumentTabColumns` (images 5 columns, dependencies 6 columns), `TestDocumentImageRowOnlyInImages` and `TestDocumentNonImageRowOnlyInDependencies` (spec scenarios), `TestDocumentScriptEscaped`, `TestDocumentQuoteInVersionNoAttribute`, `TestDocumentErrorsSection` (level 50 listed, level 40 absent), `TestDocumentNothingPending` (0 updates, no rows, 12 and 30 up to date), `TestDocumentLangOverride` (`lang="en"`, labels Spanish), `TestDocumentForkTag` (default and overridden `fork.tag_document`), and `TestDocumentTimestampZone` (`2026-10-12 09:07 America/Guayaquil`).
- [x] 7.3 Add `DocumentName(scannedAt, loc, pattern, layout)`. Verify with `TestDocumentNameDefault` (`2026-10-12T14:07:00Z` gives `dependency-report-2026-10-12.html`), `TestDocumentNameNearMidnight` (`2026-10-13T02:30:00Z` gives `dependency-report-2026-10-12.html`), `TestDocumentNameConfiguredZone` (`2026-10-12T22:30:00Z` in `Europe/Madrid` gives `dependency-report-2026-10-13.html`) and `TestDocumentNameCustomPattern` (`report-{date}.html` with layout `20060102` gives `report-20261012.html`).
- [x] 7.4 Verify that both presenters read one `Summary`: `TestPresentersShareSummary` (caption `📊` totals equal the sum of the two tab totals of the same `Summary`).

## 8. Delivery, failure path and handler

- [x] 8.1 Move the failure builder to `failure.go` (`BuildFailure(cat, exitCode, pod, logTail)`), keeping the rune-safe tail cut and the 4096 limit as a protocol constant. Verify with the moved tests `TestBuildFailureMessage*`, `TestEscapedSuffix*` and `TestFailureTitleOverride` (message-catalog spec scenario).
- [x] 8.2 Add `ReportSender` (the only definition of this interface, with `SendHTML` and `SendDocument`) and `Deliverer` (document with caption; on failure the caption as text; errors logged). Verify with `TestDeliverSendsDocumentWithCaption`, `TestDeliverFallsBackToCaptionText`, `TestDeliverFailureMessageIsTextOnly` (no document) and `TestDeliverLogsSenderError` (observer logger, no panic).
- [x] 8.3 Reduce `internal/handlers/dependency_report.go` to orchestration: token check with the configured header (constant time), `ParseSubmission`, failure or report branch, one `Aggregate` call through a function field (`aggregate`), presenters, 202, then `go deliver`. Verify with `TestAuthRejectsMissingToken` (401, no send), `TestCustomTokenHeader`, `TestBodyTooLargeGets413`, `TestMalformedExitCodeGets400`, `TestAcceptsBeforeSending` (202 before any send), `TestAggregatesOnce` (the fake aggregate counts one call per report).
- [x] 8.4 Update `cmd/automation-hub/main.go`: use `cfg.Location` for the scheduler, `cfg.Catalog` for the handler, `cfg.DependencyReport.Path` for the route, and `depreport.ReportSender` in `registerRoutes` (its local interface is removed). Verify with `TestRegisterRoutesDependencyReport` (custom path returns 202, old path returns 404) and `TestRegisterRoutesNotConfigured` (404 when the section is absent), plus `go build ./...`.
- [x] 8.5 Delete the old code: `dependency_report_document.go`, `reportTimeZone`, the repeated `LoadLocation` calls, `shortRepoName`'s owner strip, and the `forkTag` literal. Verify with `grep -rn "devidence-dev/\|America/Guayaquil\|X-Report-Token\|reportTimeZone" internal cmd --include=*.go`, which must list only `internal/config` defaults and test fixtures.

## 9. Docs

- [x] 9.1 Update `configs/config.yaml.example`: top-level `timezone` with its default, the new `dependency_report` keys commented with their defaults, and a commented `messages:` example that overrides `caption.header`. Verify with `TestLoadExampleConfig` (the example still loads) and `grep -n "timezone\|token_header\|messages:" configs/config.yaml.example`.
- [x] 9.2 Update the dependency report section of `README.md`: timezone, path and header settings, the catalog overrides, the two fork tags, the Spanish labels, and the 1024 and 4096 protocol limits. Verify with `grep -n "per repository\|one per repo" README.md` returning nothing and `grep -n "messages:" README.md` returning the new section.

## 10. Cross-repo and user steps

- [x] 10.1 (user) Approve visually the Spanish document labels drafted in `default.yaml` (design D7) before task 3.1 is applied. The fork keys and `lang="es"` were decided on 2026-10-11. Verify: the user replies with approval or changes, and `default.yaml` matches them.
- [x] 10.2 (homelab) Confirm the Job still posts to `/webhook/dependency-report` with header `X-Report-Token` and the four field names, so no homelab change is needed. Verify: a grep of `kubernetes/internal-apps/automation-hub/dependency-report.yaml` in `/home/px1/repositories/homelab` shows the same path, header and fields.
- [x] 10.3 (user) Deploy the hub with the default config and run `/run_dependency_report`. Verify: one Telegram document named `dependency-report-<today>.html` with the caption header unchanged, the fork line showing ` (fork)`, and the document header in `America/Guayaquil`.

## 11. Validation

- [x] 11.1 Run in Docker: `docker run --rm -v "$PWD":/app -w /app golang:$(cat .go-version) sh -c 'go vet ./... && go test -race ./...'`; `docker run --rm -v "$PWD":/app -w /app golangci/golangci-lint:v2.14.0 golangci-lint run --enable=errcheck,govet,staticcheck,unused,misspell,gosec ./...`; and gocognit v1.2.0 (`go install github.com/uudashr/gocognit/cmd/gocognit@v1.2.0`) with `gocognit -over 15 .`, which must print nothing with tests included. Run `graphify update .` and `openspec validate --strict refactor-dependency-report`. Verify: every command exits 0 on the current code. Do not commit; summarize the changed files and propose a commit message.
