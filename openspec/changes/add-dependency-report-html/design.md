# Design

## Context

- `DependencyReportHandler` (`internal/handlers/dependency_report.go`) parses the Renovate report into `renovateReport`, builds the caption and the HTML document from it (`buildCaption`, `buildReportDocument`), answers 202 and sends them from a goroutine. Its sender is the consumer-defined interface `reportSender` (`SendHTML` + `SendDocument`); `main.go` declares a matching one.
- `telegram.Client` has `SendMessage` and `SendHTML`, both going through one private `send` with a 3-attempt retry loop and `tgbotapi.NewMessage`. The library `github.com/go-telegram-bot-api/telegram-bot-api/v5` already provides `tgbotapi.NewDocument(chatID, tgbotapi.FileBytes{Name, Bytes})`, so no new dependency. The HTTP client timeout is 40 s per attempt.
- `reportDep` today reads `depName`, `currentVersion`, `currentValue` and `updates[]`. It does not read `datasource`, nor the `packageFile` of each package file entry.
- The synthetic fixture `internal/handlers/testdata/renovate-report.json` has 54 deps, all with `"datasource": "go"`, including the ones in `Dockerfile` files.
- icr (`/home/px1/repositories/image-container-reporter`) has the template `internal/report/assets/report_template.html` (583 lines, Bootstrap 5.3.2 and bootstrap-icons 1.11.1 from `cdn.jsdelivr.net`, CSS inline, copy buttons in a small inline script) and `html_formatter.go` (`html/template`, `go:embed`, distribution items with colors, badge classes). Its labels are in English.
- The hub binary embeds timezone data (see `scheduled-dispatch`), so `America/Guayaquil` loads without OS tzdata.
- See proposal.md for motivation.

## Renovate 44.151.0 findings (read from source)

Checked at `https://raw.githubusercontent.com/renovatebot/renovate/44.151.0/...`:

- `lib/modules/manager/types.ts`: `interface PackageDependency` (declared at line 190) has `datasource?: string;` (line 230). It is optional, so a dep may lack it; the spec sends those to "Dependencias". `PackageFileContent` has its own `datasource?: string` (line 106), which the format ignores.
- `lib/modules/manager/kubernetes/index.ts` and `.../argocd/index.ts`: `defaultConfig = { managerFilePatterns: [] }`, so both match no file until patterns are given. `.../helm-values/index.ts`: `managerFilePatterns: ['/(^|/)values\\.ya?ml$/']`, so `values.yaml` files work without config. The kubernetes manager supports datasources `docker` and `kubernetes-api`; argocd supports `docker`, `git-tags` and `helm`.
- `lib/config/options/index.ts`: the option is `managerFilePatterns` (type array, `env: false`, patterns are `/regex/` or globs). `fileMatch` no longer exists as an option: `lib/config/migrations/custom/file-match-migration.ts` is a deprecated migration that rewrites it to `managerFilePatterns`.
- The same file registers one object option per manager (`name: 'kubernetes'`, `type: 'object'`), and `lib/config/options/env.ts` derives the env name `RENOVATE_<UPPER_SNAKE>`; `lib/workers/global/config/parse/env.ts` parses object values with JSON5. So the expected form is `RENOVATE_KUBERNETES='{"managerFilePatterns":["/^kubernetes\\/infrastructure\\/cloudflared\\/deployment\\.yaml$/"]}'`, and the same for `RENOVATE_ARGOCD`. `RENOVATE_CONFIG` (JSON) is the alternative.

These are source reads, confirmed by the runs below.

Task 1.1 (run, `ghcr.io/renovatebot/renovate:44.151.0`, `RENOVATE_PLATFORM=local`, `RENOVATE_REPORT_TYPE=file`, on a read-only mount of this repo):
- All 30 deps have a `datasource` key. Dockerfile image deps (`FROM`) show `docker`. Dockerfile `RUN apk add` packages show `apk`, not `docker`, so they go to "Dependencias" under the datasource rule.
- Stage references (`FROM builder`) and docker-compose images have `depName` null, so the text fallback `?` applies to images too.
- `PackageDependency.datasource` is `datasource?: string` at `lib/modules/manager/types.ts:230`, as recorded above. The shape is unchanged.

Task 1.2 (run on a read-only mount of `homelab`, `RENOVATE_PLATFORM=local`, `RENOVATE_REPORT_TYPE=file`):
- Without env: 31 `helm-values` `values.yaml` files and `ansible/requirements.yml`. No `kubernetes` or `argocd` package file; the cloudflared deployment, `dependency-report.yaml` and every `application.yaml` are absent.
- Option name `managerFilePatterns`, set per manager through the object option `RENOVATE_KUBERNETES` / `RENOVATE_ARGOCD` (JSON5, the same mechanism as the other object options). Globs work; no regex escaping is needed.
- With the env below: `kubernetes` lists `kubernetes/infrastructure/cloudflared/deployment.yaml` (`cloudflare/cloudflared` docker `2026.10.0`) and `kubernetes/internal-apps/automation-hub/dependency-report.yaml` (`ghcr.io/renovatebot/renovate` docker `44.151.0`, `patch` update; `curlimages/curl` docker `8.22.0`, no update). `argocd` lists 42 `application.yaml` files with Helm chart deps (`datasource` `helm`), for example `headlamp` `0.43.0` (`minor`), `vault` `0.30.0` (`minor`), `zot` `0.1.116` (`patch`) and `rancher` `2.15.0-rc1` (`patch`). `helm-values` stays at 31 files.
- Exact values for task 5.1:
  - `RENOVATE_KUBERNETES={"managerFilePatterns":["kubernetes/infrastructure/cloudflared/deployment.yaml","kubernetes/internal-apps/automation-hub/dependency-report.yaml"]}`
  - `RENOVATE_ARGOCD={"managerFilePatterns":["**/application.yaml"]}`
  - `RENOVATE_CONFIG` was not needed.

## Goals / Non-Goals

**Goals:**
- One Telegram message per successful run: the document with its caption. Both are built from the same parsed report and the same data model, so they always agree.
- Look and feel identical to icr's report; all variable text escaped by `html/template`.
- A Telegram document failure falls back to the caption as text and never changes other components.

**Non-Goals:**
- Any change to the failure message or to who triggers the Job.
- Generating the document in the Job or storing it anywhere. The hub builds it in memory.

## Decisions

- **Document built by a pure function.** `buildReportDocument(rep renovateReport, forks []string, scannedAt time.Time) ([]byte, error)` returns the HTML bytes, and a small `documentName(scannedAt)` returns the file name. Tests call them without HTTP or Telegram. `scannedAt` comes from an injectable clock on the handler (`now func() time.Time`, default `time.Now`) converted to `America/Guayaquil` with `time.LoadLocation`. Alternative: use a timestamp from the Renovate report. Not used because no timestamp was verified in the report; receipt time is within seconds of the scan end.
- **Caption.** `buildCaption(rep, forks, scannedAt) (string, error)` calls `buildDocumentData`, the same function the document uses, so the caption and the document share the tab totals, the repo counts with pending updates and the error list. Each repo with pending updates gets one line; a dep with several updates counts once, under its highest type. The caption is HTML with parse mode HTML and must hold at most 1024 runes after escaping, tags included (a conservative bound for Telegram's 1024 visible characters). When it is longer, repo lines are dropped from the end and `… y N repos más (ver reporte)` is added after the last kept one, until it fits; the header, the `✅` lines and the stats are always kept. Alternative: a separate text message for the summary. Rejected: the user decided on one message.
- **Classification by `datasource`.** `reportDep` gains `Datasource *string`. `docker` goes to the images list; anything else (including nil) goes to the dependencies list. The package file name comes from `packageFiles.<manager>[].packageFile`. Alternative: classify by manager (`dockerfile`, `helm-values`, `kubernetes`, ...). Rejected because `github-actions` and `argocd` mix datasources, and the user decided on `datasource`.
- **One row per dep and update.** A dep with a `major` and a `minor` update gets two rows, each with its own badge. Rows are deduplicated on (repo, package file, dep, current, new, type) and sorted by repo, package file, dep name, then `major` before `minor` before `patch`. Alternative: one row with several badges. Rejected as it needs a different cell layout than icr's single badge column.
- **Image strings.** Current image is `<depName>:<current>` and latest is `<depName>:<new>`, with the fallback order `currentVersion` then `currentValue`, and `newVersion` then `newValue`; `?` when absent. `digest` is not shown.
- **First column shows the package file.** icr shows a service name from Compose; Renovate has none, so the first "Imágenes" column shows the package file in the same style and the repo gets its own column. This follows the user's "service or package file".
- **Totals per tab.** `total` counts distinct deps (repo, package file, dep name) in the tab, including those without updates; `with updates` counts deps with at least one listed row; `up to date` is the difference; the distribution counts rows by type. icr's three cards map to Total, Pending Updates and Up to Date. The layout is icr's: each tab pane has its sidebar of cards and its content area; the navbar carries the title and the scan timestamp.
- **Template.** `internal/handlers/assets/dependency_report.html`, embedded with `//go:embed`, parsed once at package init (`template.Must`). It is icr's file with: title and navbar renamed to "Dependency Report", Bootstrap `nav-tabs` + `tab-content` added, the table section duplicated per tab with the columns above, and the errors list kept. Inline `style="color: {{.Color}}"` is replaced by CSS classes (`.dist-patch`, `.dist-minor`, `.dist-major`) so `html/template`'s CSS filter cannot turn a value into `ZgotmplZ`. icr's theme toggle button is dropped because it has no handler in icr (dead UI); the CDN links, SRI hash on the Bootstrap bundle, copy buttons and CSS are unchanged. Tab names are Spanish as decided; other labels stay in icr's English. Alternative: `text/template` or string building. Rejected; auto-escaping is the safety property in the spec.
- **Telegram client.** Rename the private `send` loop to take a `tgbotapi.Chattable` and add `SendDocument(chatID, filename, caption string, data []byte) error`. It builds `tgbotapi.NewDocument(chatIDInt, tgbotapi.FileBytes{Name: filename, Bytes: data})` and sets the document's `Caption` and `ParseMode` to HTML. `SendMessage` and `SendHTML` call the same loop. Retry and logging stay as they are. Alternative: a second copy of the loop. Rejected as duplication.
- **Interfaces.** The handler's consumer-defined interface is `SendHTML` + `SendDocument` (still two small methods, one `reportSender` in `main.go`). The fake in `dependency_report_test.go` records both kinds of call in order.
- **Delivery order and errors.** `deliver` sends the document with the caption as its caption. If the document fails after the client's retries, the error is logged and the caption goes as a text message (`SendHTML`), so the summary still arrives. A caption build error sends nothing; a document build error sends the caption as text, and the error is logged. The failure path sends the failure message as text and never builds a document. The build happens before 202. The 2-minute background context is dropped: a request makes one or two sends now, and the client's retry loop bounds them.
- **Renovate scope.** The Job gets `devidence-dev/homelab` in `DEPENDENCY_REPORT_REPOSITORIES`, plus `RENOVATE_KUBERNETES` and `RENOVATE_ARGOCD` in the CronJob template. Candidates, to be confirmed by task 1.2 against real files: `kubernetes/infrastructure/cloudflared/deployment.yaml` and `kubernetes/internal-apps/automation-hub/dependency-report.yaml` for `kubernetes`, and `kubernetes/**/application.yaml` for `argocd`. `values.yaml` is already read by `helm-values`.
- **Retiring icr.** The runner workflow, command, mount and host file go in cross-repo tasks. Nothing is lost: the Dockerfiles listed in `extra-images.yml` (discord-tts-bot, devidence-home, github-runner, voraz-bot, automation-hub) are in repos Renovate already scans. The `scheduled-dispatch` and `job-commands` main specs use `run_check_updates` only as an example value, not as a requirement, so they need no delta.

## Risks / Trade-offs

- [The document loads Bootstrap from `cdn.jsdelivr.net`] → Unstyled and without tabs when opened offline. Accepted, as in icr. Embedding the assets is a non-goal.
- [Telegram caps documents at 50 MB] → Expected size is under 500 KB (about 6 KB of CSS and script plus about 400 bytes per row); the 10 MiB request cap bounds the input. No size guard is added.
- [Renovate's updateType for non-semver image tags is odd (for example a `4.0.9.2244` build tag)] → Types are shown as Renovate reports them. If noisy, tune `packageRules` or versioning in a later change.
- [Adding homelab makes the report larger and the Job slower] → `activeDeadlineSeconds: 1800` stays; check the first run's duration and the `report.json` size against the 10 MiB cap (task 5.3).
- [`kubernetes` and `argocd` patterns may match nothing, silently] → Tasks 1.2 and 5.3 assert that a lookup lists at least one dep from each file; a missing file shows as absent in the "Imágenes" tab.
- [Tabs need JavaScript] → With scripts blocked only the first tab shows. Accepted; the document is opened in a browser.
- [Template regressions are visual] → Tests check structure and escaping, and one manual render of the fixture is opened in a browser (task 3.5).

## Migration Plan

1. Deploy the hub change. It sends the document on the next report; no config change is needed.
2. Update the Job template and Infisical so Renovate scans homelab, then run `/run_dependency_report` and check both tabs.
3. Retire icr: remove the workflow and command, then the mounts and host file, then archive the repo.
4. Rollback: revert the hub image to stop the document, and remove `devidence-dev/homelab` from `DEPENDENCY_REPORT_REPOSITORIES` to stop scanning it. The icr removal is reversible through the git history of each repo.

## Decisions (resolved)

- Labels: icr's English labels stay; only the tab names are Spanish ("Imágenes", "Dependencias").
- The inert theme toggle is dropped.
- The Errors section stays. It lists every problem with `level >= 50` of a repo, so `level` 60 problems appear too, and it lists the report root's problems as `root: <msg>`.
- One message per successful run (user decision, final): the document with its caption. The per-repository messages and the "todo al día" message are removed. The Renovate failure path is unchanged.

## Implementation notes

- The handler's consumer-defined interface is `reportSender` (`SendHTML` + `SendDocument`); it replaces `htmlSender`.
- The caption's repo and error counts come from `buildDocumentData` (`Pending`, `Errors`), not from a second pass over the report, so they match the document.
- Deps without a `depName` (Renovate's null for Dockerfile stage refs) are keyed by their position in the file, so each one counts once in the totals.
- Each tab has an `Updates` count (listed rows). The status badge reads "N updates available for M dependencies", or "All dependencies are up to date" when N is 0.
- `TestLoadExampleConfig` now expects one workflow (`run_cleanup`) instead of two, because `run_check_updates` is removed from the example.
