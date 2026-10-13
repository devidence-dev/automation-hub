# Tasks

Each task ends with a **Verify** clause. Tasks marked `(user)` are done by the user, not by the implementer. Tasks marked `(workflows)` change `devidence-dev/workflows`, not this repo.

## 1. Toolchain

- [x] 1.1 (user) Approve the new Go module `github.com/a-h/templ` v0.3.1070 and the binary downloads for the generation script (templ release, tailwindcss v4.3.3 standalone for linux x64 and arm64). The module is verified on proxy.golang.org (commit 07dfb6e). The downloads run inside an ephemeral Docker container and are checked against the release checksums (design D11). No headless browser is used (design D14).
  **Verify:** the approval is written in the conversation before 1.2 starts. Approved by the user on 2026-10-12.

- [x] 1.2 Add `scripts/generate-depreport-assets.sh`. It runs in `golang:$(cat .go-version)` with `docker run --rm -v "$PWD":/app -w /app`, as the host user (`--user "$(id -u):$(id -g)"`), so generated files are not root-owned. It downloads the pinned templ (v0.3.1070) and tailwindcss (v4.3.3) binaries for the container's architecture (`uname -m`), checks each against the checksum file of its release and stops on mismatch, runs `templ generate`, and builds the minified Tailwind CSS from `internal/services/depreport/assets/input.css` into `depreport.min.css`. Versions and URLs are variables at the top of the script. Nothing is installed on the host. Idempotent. If `input.css` does not exist yet (before 2.1), the CSS step prints that it is skipped and the script still exits 0. (Superseded by 9.1: the skip is removed.)
  **Verify:** the script runs twice and `git diff --stat` is the same before and after. Sourcing the script with `DEPREPORT_SOURCE_ONLY=1` and calling the checksum check with a wrong hash returns non-zero.

- [x] 1.3 Confirm the checksum sources and record them in the script header. Verified 2026-10-12: Tailwind CSS v4.3.3 publishes `sha256sums.txt` in its GitHub release (assets such as `tailwindcss-linux-x64`); templ v0.3.1070 publishes `checksums.txt` (assets such as `templ_Linux_x86_64.tar.gz`). The script downloads the checksum file with each binary and fails if the asset has no line in it. No hash is pinned by hand.
  **Verify:** the script header names both checksum files and the asset each one covers, and the release asset names in the script match the release metadata.

- [x] 1.4 (user) Approve the Biome 2.5.15 standalone binary downloads (`biome-linux-x64` and `biome-linux-arm64`) for `scripts/verify.sh` and the shared `biome` action. The release publishes no checksum file, so each binary is pinned by its SHA-256 (design D16).
  **Verify:** the approval is written in the conversation before 9.4 starts. Approved by the user on 2026-10-12.

## 2. Styles

- [x] 2.1 Add `internal/services/depreport/assets/input.css` with `@import "tailwindcss";`, `@source` lines for `*.templ`, the tokens of design D8 in `:root` and under `prefers-color-scheme: dark`, `@theme inline` mappings, the update-type colors, the Geist font stack and the brand gradient utility.
  **Verify:** `styles_test.go` reads the compiled CSS and checks the dark token block, the three badge pairs and the gradient string.
  **Superseded (batch 3):** by 13.3 (batch 3, 2026-10-12): the palette-A tokens, the `success` tokens and the brand gradient are replaced by the liquid glass tokens; `styles_test.go` is rewritten. The task stays ticked as history.

- [x] 2.2 Commit the compiled, minified CSS as `internal/services/depreport/assets/depreport.min.css` and embed it with `go:embed`.
  **Verify:** `styles_test.go` checks that no `min-width: 520px` and no Bootstrap rule is in the compiled CSS.

## 3. Static prototype (approval before any presenter change)

- [x] 3.1 Write a static HTML prototype with the markup of design D2 and D4 to D6: header with the devidence gradient line, stats, the two tabs (CSS only), repo sections with summary rows, file groups, cards and aligned rows, badges, copy buttons and the errors section. Fill it from `internal/handlers/testdata/renovate-report.json` with the clock fixed at 2026-10-12 09:07 America/Guayaquil (14:07 UTC) and the fork `devidence-dev/discord-tts-bot`, using a throwaway Go program in the session scratchpad, not in the repo. Inline the compiled CSS from 2.2 and load the Geist and Geist Mono fonts from Google Fonts. Write it to `/tmp/claude-1000/-home-px1-repositories-automation-hub/3a94737e-fd46-4738-b128-8b809e4e3f78/scratchpad/prototype/dependency-report-prototype.html`. No headless browser is used (design D14). The HTML comment at the top lists what the user checks in 3.2, including that switching tabs keeps an open section open.
  **Verify:** a throwaway static check in the scratchpad passes: no `<script src>`, no `bootstrap`, `lang="es"`, every repo `<details>` has no `name` and no `open` attribute, the inlined CSS has no fixed `min-width` above 320px, and it contains the `@media (prefers-color-scheme: dark)` block and the `@container (min-width: 40rem)` query.
  **Superseded (batch 3):** by 13.8 (batch 3, 2026-10-12): the tabs are JS-driven (decision f), not CSS-only radios. The prototype stays as history.

- [x] 3.2 (user) Send the prototype to the user's Telegram chat as a file, open it on iOS (phone width), iPadOS and macOS, in light and dark mode. Check the phone for horizontal scrolling, that each repo section opens and closes without closing the others, and that switching tabs keeps an open section open. Report what to change.
  **Verify:** written approval on each platform and scheme; any requested change is applied and 3.1 is repeated before task 4. Result: the user reviewed the prototype on 2026-10-12 and answered "mostly good", which counts as this approval, with two changes applied in batch 2b: (1) the copy buttons are icon-only, named by their catalog label, with Heroicons v2.2.0 20/solid `clipboard-document` and a `check` swap for about 2 seconds after a copy (design D10); (2) each entry has one copy button for the new version only: `<depName>@<new version>` for a dependency, and the latest image reference for an image. The static check of 3.1 is repeated on the changed document in batch 2b, after tasks 4 and 5 were done.
  **Superseded in part (batch 3):** by batch 3 (2026-10-12): the user accepted the liquid glass mock in full, which replaces the batch 2 prototype. The approval above stays as history; the final approval is 12.1.

## 4. View model regrouping

- [x] 4.1 Move the caption's repo comparator into `aggregate.go`, add `pendingOf(rows []Row) []RepoPending` for one tab, and use both from the caption and the document.
  **Verify:** the existing `caption_test.go` and the caption golden tests pass unchanged.

- [x] 4.2 Replace the flat `TabView.Rows` with `Groups`, `Files` and `Entries` built by one pass over the sorted rows. Remove CSS classes from the view model. Each entry carries its name, current and new versions, type, whether it is an image, its copy value and label.
  **Verify:** table-driven `document_view_test.go` covers: repo then file grouping; image and dependency split; two updates of one dep stay two entries; a dep with major and minor counts once as major in the summary; sort order; fork flag; counters only for non-zero types; per-tab counts.
  **Superseded in part (batch 3):** by 13.5 (batch 3): `EntryView.Name` of an image is the image name and its versions are separate, and the summary carries `Dot` and `Chips` instead of emoji counters.

## 5. templ components

- [x] 5.1 Add the `.templ` files of design D2 (`report`, `stats`, `repo`, `entry`, `errors`, `icon`) and generate them with the script from 1.2. `Report` inlines `assets/copy.js` (task 9.2), not a script written in the template. Add `github.com/a-h/templ` to `go.mod` and `go.sum` after the approval in 1.1.
  **Verify:** `go build ./...` and `go vet ./...` pass in Docker.
  **Superseded (batch 3):** by 13.6 (batch 3): the component files are `report`, `logo`, `theme`, `tabs`, `stats`, `chip`, `repo`, `entry`, `errors` and `icon` (design D2).

- [x] 5.2 Add `components_test.go` that renders each component into a buffer and checks its expected markup, table-driven.
  **Verify:** `go test -race ./internal/services/depreport/...` passes.
  **Superseded (batch 3):** by 13.6 (batch 3): `components_test.go` is rewritten for the new components.

- [x] 5.3 Escaping: an entry named `<script>alert(1)</script>` with current version `1.0.0" onmouseover="alert(1)` renders escaped, with no raw `<script>` and no `onmouseover` attribute.
  **Verify:** `TestEntryEscapesValues`.

- [x] 5.4 Badges and summary row: `UpdateBadge` gives distinct classes per type that match the D8 tokens; `RepoSummary` shows 🔴 with a major update and 🟡 without, and only the non-zero counters.
  **Verify:** `TestUpdateBadgePerType` and `TestRepoSummary`.
  **Superseded (batch 3):** by 13.5 and 13.6 (batch 3): the repo emoji (🔴/🟡) and the emoji counters are removed (decision a); the chips and the dot replace them.

## 6. Migrate DocumentPresenter

- [x] 6.1 Keep `DocumentPresenter.Present(Summary) ([]byte, error)`. Its body builds the new view and renders `Report` into a buffer. Remove the `html/template` use.
  **Verify:** `TestDocumentPresenterPresent` checks that the output starts with `<!doctype html>` (templ writes the doctype in lower case) and has `lang="es"`.

- [x] 6.2 Regenerate the golden document by running the golden test in `internal/handlers/` with `UPDATE_GOLDEN=1` (the switch in `dependency_report_golden_test.go`). Review the diff of `document.html` by eye.
  **Verify:** `git diff --stat -- internal/handlers/testdata/golden/` lists only `document.html`; `go test ./internal/handlers/...` passes without `UPDATE_GOLDEN`.

- [x] 6.3 Static behavior: the document has exactly one `<script>` (the copy handler), no `<link rel="stylesheet">` other than Google Fonts, no bootstrap, no `details` with a `name` attribute, and every repo `details` without `open`.
  **Verify:** `TestDocumentStaticBehavior`.
  **Superseded in part (batch 3):** by 13.7 (batch 3): the document has two inline scripts (theme and report), not one; no `name` or `open` on `details` still holds.

## 7. Catalog keys

- [x] 7.1 Add to `internal/messages/default.yaml`: `document.summary.emoji.major` (🚨), `document.summary.emoji.minor` (⚠️), `document.summary.emoji.patch` (♻️), `document.summary.counter` (`{emoji} {count}`), `document.copy.new_version`. Add the constants and list them in `allKeys` in `keys.go`. Keep the unused `document.columns.*` and `document.copy.current_image` keys with a comment.
  **Verify:** the existing key-existence test in `keys_test.go` passes with the new keys.
  **Superseded (batch 3):** by 13.4 (batch 3): the emoji summary keys stay in the catalog, unused (decision a); the new keys are added in 13.4.

- [x] 7.2 Overrides: `messages: {document: {summary: {emoji: {major: "🔥"}}}}` changes the summary emoji; an override of `document.columns.dependency` still passes startup.
  **Verify:** `TestOverrideSummaryEmoji` and `TestUnusedKeyOverrideValidates`.
  **Superseded (batch 3):** by 13.4 (batch 3): `TestOverrideSummaryEmoji` now asserts that the override does not change the document, since the emoji are not rendered.

## 8. Remove Bootstrap and the old template

- [x] 8.1 Delete `internal/services/depreport/assets/dependency_report.html` and every Bootstrap or bootstrap-icons reference in the repo, outside `openspec/changes/archive` and `graphify-out`.
  **Verify:** `grep -rni "bootstrap" . --exclude-dir=openspec --exclude-dir=graphify-out --exclude-dir=.git` returns nothing; `go test ./...` passes.

## 9. Local gate, Biome and CI wiring (this repo)

Order: 9.1, 9.2 and 9.3 are prerequisites of 5.1 and of the Biome checks in 9.5, so they are done before 5.1. 9.7 needs no release: its drift job is local (design D17). Numbering follows the split of the work, not the order.

- [x] 9.1 Remove the CSS skip from `scripts/generate-depreport-assets.sh` (it supersedes the skip described in 1.2). If `internal/services/depreport/assets/input.css` is missing, the script exits non-zero and prints the path, so a missing input fails the script, the drift check and CI.
  **Verify:** in a scratch copy of the tree without `input.css`, the script exits non-zero and names the path; with the file present it exits 0, and two runs leave `git diff --stat` unchanged.

- [x] 9.2 Move the copy-to-clipboard script out of the `.templ` into `internal/services/depreport/assets/copy.js` (design D10). Embed it with `go:embed`, and have `Report` inline it in one `<script>` element.
  **Verify:** `TestCopyScriptEmbedded` in `components_test.go` checks that the rendered document has exactly one `<script>` and that its body equals the embedded `copy.js`; `biome ci` (9.5) passes on the file.
  **Superseded (batch 3):** by 13.2 (batch 3): `copy.js` is renamed to `report.js` (with the tabs, theme and motion), and `theme.js` is added.

- [x] 9.3 Add `biome.json` at the repo root (design D16): formatter and linter on, `css.parser.tailwindDirectives: true`, `files.includes` limited to `internal/services/depreport/assets/**` and excluding `depreport.min.css`. No `package.json` and no `node_modules`.
  **Verify:** `biome ci` (9.5) passes on the tree; in a scratch copy, a deliberate lint error in `copy.js` makes it exit non-zero, and the Tailwind directives in `input.css` parse without an error.
  **Superseded in part (batch 3):** by 13.2 (batch 3): `files.includes` covers `theme.js` and `report.js`; the SVG sources are not linted.

- [x] 9.4 Pin Biome 2.5.15 by SHA-256. In an ephemeral container, download `biome-linux-x64` and `biome-linux-arm64` at implementation time, compute each SHA-256, and commit the hashes in `scripts/verify.sh`. The release publishes no checksum file (design D16), so the script checks the binary against the pinned hash before running it, and a mismatch stops the script.
  **Verify:** each committed hash equals the `sha256sum` of a fresh download and matches the GitHub asset digest recorded in design D16; in a scratch copy, one changed hex digit makes the script fail before Biome runs.

- [x] 9.5 Add `scripts/verify.sh`, modelled on voraz-bot's `scripts/validate.sh` (design D15). Steps in order: `gofmt -l .` must be empty; golangci-lint v2.14.0 with `--enable=errcheck,govet,staticcheck,unused,misspell,gosec`; `go test -race` with coverage after the `sonar.coverage.exclusions` filter, minimum `MIN_COVERAGE` (default 80); gocognit v1.2.0 with `-over 15 -ignore '_templ\.go$' .`; asset drift (the generation script, then `git diff --exit-code` and a check for untracked files); `biome ci` on the CSS and JS sources; `openspec validate --all --strict`. Everything runs in Docker with `--user "$(id -u):$(id -g)"` and shared caches in `~/.cache`. The Go image comes from `.go-version`, and the coverage filter takes the module path from `go.mod`, not from validate.sh's `voraz-bot/` prefix. `--fix` runs `gofmt -w`, `biome check --write` and regenerates the assets. No pre-push hook. Biome is not in this list of Go tools; `openspec validate` runs with `@fission-ai/openspec@1.14.1` (npm latest, verified 2026-10-12) in a pinned `node:<LTS>-slim` image whose tag is verified at implementation time.
  **Verify:** `scripts/verify.sh` exits 0 on the tree; in a scratch copy, an unformatted Go file, `MIN_COVERAGE=99`, and a stale generated file each make it exit non-zero; `--fix` on a scratch copy leaves `gofmt -l .` empty.
  **Batch 3:** this gate is unchanged; 13.12 runs it on the new assets.

- [x] 9.6 Add `**/*_templ.go` and `internal/services/depreport/assets/depreport.min.css` to `sonar.exclusions` in `sonar-project.properties`.
  **Verify:** `grep -n "_templ.go\|depreport.min.css" sonar-project.properties` shows both.

- [x] 9.7 Add the local `generated-assets` drift job to `.github/workflows/ci.yml`. It sits beside the `ci` job, which is unchanged. It checks out the repo, runs `scripts/generate-depreport-assets.sh`, then fails on `git diff --exit-code` or on any untracked file (design D11, D17).
  **Verify:** the YAML is valid (actionlint 1.7.12 in Docker, with the `ubuntu-26.04` runner label declared in a temporary config); the job runs the script and fails on a diff or on an untracked file.

## 9b. Shared CI extension (workflows)

- [x] 9b.1 Dropped (2026-10-12): the Sonar Quality Gate covers this; see design D17. (workflows) Add input `coverage_min` (default `0`, off) to `ci/go/steps/test/action.yml`. When above 0, read `sonar.coverage.exclusions` from the caller's `sonar-project.properties`, drop those files from `coverage.out`, and fail the step if total coverage is below the minimum. The output `coverage_file` is unchanged.
  **Verify:** a local run of the step's bash in Docker against this repo with `coverage_min` 99 exits non-zero, and with 0 the step behaves as `v1.2.4`.

- [x] 9b.2 Dropped (2026-10-12): the Sonar Quality Gate covers this; see design D17. (workflows) Add composite `ci/go/steps/cognitive` with required input `max` and optional input `ignore` (a regexp, default empty): installs gocognit v1.2.0 and runs `gocognit -over <max> [-ignore <ignore>] .`, failing on any output. gocognit v1.2.0 has the `-ignore` flag (`cmd/gocognit/main.go:155`). Confirm the module path against the module proxy before writing the install line.
  **Verify:** a local run in Docker against this repo with `max` 15 prints nothing and exits 0; with `max` 1 it prints findings and exits non-zero.

- [x] 9b.3 Dropped (2026-10-12): the Sonar Quality Gate covers this; see design D17. (workflows) Add composite `ci/steps/biome`: downloads the standalone Biome 2.5.15 binary for the runner architecture, checks it against the SHA-256 pinned in the action (the hashes of 9.4), and runs `biome ci` with the caller's `biome.json`. No Node or npm step.
  **Verify:** the step passes on the branch run in 9b.7; in a local run, a wrong pinned hash fails before the binary runs.

- [x] 9b.4 Dropped (2026-10-12): the Sonar Quality Gate covers this; see design D17. (workflows) Add composite `ci/steps/generated-drift` with input `command`: runs the command, runs `git diff --exit-code`, and fails if `git status --porcelain --untracked-files=all` lists anything (`git diff` does not see new files; design D17).
  **Verify:** the step passes on the branch run in 9b.7; a local run where the command edits a generated file fails, and one where it creates an untracked file also fails.

- [x] 9b.5 Dropped (2026-10-12): the Sonar Quality Gate covers this; see design D17. (workflows) Add optional inputs to `ci-go.yml`: `coverage_min` (default `0`), `cognitive_max` (default `0`), `cognitive_ignore` (default empty, passed to the `cognitive` step as `ignore`), `biome` (default `false`) and `generate_command` (default empty). Pass `coverage_min` to the `test` job's test step. Add the jobs `cognitive`, `biome` and `generated-drift`, each parallel and gated with `if:` on its input. Existing jobs, inputs and defaults are unchanged.
  **Verify:** `actionlint` in Docker passes on `ci-go.yml`; the branch run in 9b.7 shows the new jobs; a run with the defaults shows only the previous jobs, the new ones skipped.

- [x] 9b.6 Dropped (2026-10-12): the Sonar Quality Gate covers this; see design D17. (user) Commit the workflows changes (9b.1 to 9b.5) on a feature branch in `devidence-dev/workflows` and push it. Branch name: `feat/go-quality-gates`.
  **Verify:** `git ls-remote` shows the branch at the commit that holds 9b.1 to 9b.5.

- [x] 9b.7 Dropped (2026-10-12): the Sonar Quality Gate covers this; see design D17. (workflows) Branch test: point this repo's `.github/workflows/ci.yml` at the branch ref of `ci-go.yml`, with the four inputs. The change is pushed as a branch in this repo by the user, and reverted by 9.7.
  **Verify:** the CI run on that branch is green, and its logs show the coverage minimum enforced, `cognitive`, `biome` and `generated-drift` each run.

- [x] 9b.8 Dropped (2026-10-12): the Sonar Quality Gate covers this; see design D17. (user) Tag `v1.3.0` in `devidence-dev/workflows` and move the `v1` tag to it. Do this only after 9b.7 is green, and after checking the `@v1` callers found on 2026-10-12: automation-hub, devidence-home, home-ctl, dc-sentinel, QuantWarden (`ci.yml`) and voraz-bot (`release.yml`). All new inputs default off, so they are unaffected.
  **Verify:** `git ls-remote --tags` shows `v1.3.0` and `v1` at the same commit.

## 10. README

- [x] 10.1 Add a section to `README.md` after "Dependency report webhook" on regenerating the report assets and running the local gate: the generation script, when to run it, the pinned versions, `scripts/verify.sh`, and the CI drift check.
  **Verify:** the commands in the section match the script's entry point (read-only check).

## 11. Final gate (Docker)

- [x] 11.1 Run `scripts/verify.sh` with no flags. It covers gofmt, golangci-lint (with `govet`), `go test -race` with the 80% coverage minimum, gocognit, asset drift, `biome ci` and `openspec validate --all --strict`. The static checks of 2.1, 2.2 and 6.3 run inside `go test` (`TestDocumentStaticBehavior`, `styles_test.go`), with no browser.
  **Verify:** exit 0, every step prints its ok line, and the coverage line reports 80% or more.
  **Batch 3:** re-run in 13.12.

- [x] 11.2 Run `graphify update .` after the code changes. This is a project rule (`openspec/config.yaml`), not part of the gate.
  **Verify:** exit 0.

## 12. Deploy and check (user)

- [x] 12.1 (user) Deploy through the existing pipeline and let one real report arrive in Telegram. Check it on iPhone, iPad and Mac, in light and dark mode: no horizontal scroll on the phone, each repo section opens and closes without closing the others, copy works, and the tabs switch.
  **Verify:** written confirmation from the user for each platform and scheme.

## 13. Batch 3: liquid glass port (2026-10-12)

The user's mock `reporte-dependencias.html` (removed from the repo after the port; the golden `document.html` now embodies it) is the visual reference (read only; the folder is not modified). Decisions a to f of design D20 and D21 are the user's, accepted on 2026-10-12. Tasks are done in order; each ticks when its Verify passes.

- [x] 13.1 Stripped logo and icon. Write `internal/services/depreport/assets/logo.svg` (the two logo sources merged into one: the wordmark main fill as `style="fill:var(--logo-main)"`, the sub fill as `style="fill:var(--logo-sub)"`, `role="img"`, `aria-label="devidence.dev"`, no `<metadata>`, `<title>`, `<desc>`, `xmlns:c2pa`) and `assets/icon.svg` (the icon without metadata, title and desc). Embed both in `logo.go` with `go:embed`.
  **Verify:** `TestLogoAssetsAreClean` passes (no `<script`, no `<metadata`, no `on*=`, both `role="img"`), and the sizes before and after are recorded in the batch 3 report (13290 and 9191 bytes before; 5345 and 1288 bytes after).
- [x] 13.2 Scripts. Rename `assets/copy.js` to `assets/report.js` with `git mv` and extend it with the theme control, tabs, pills, panel motion and the batch 2 copy handler (D20). Add `assets/theme.js` for the head script. `script.go` embeds both.
  **Verify:** `TestThemeScriptEmbedded` and `TestReportScriptEmbedded` (each inline script equals its embedded file, one `<script>` each); `biome ci` passes on both files in 13.9.
  **Superseded in part (2026-10-12):** by 13.14: `theme.js` is deleted, and `report.js` uses no storage.
- [x] 13.3 Tokens and styles. Rewrite `assets/input.css`: the mock's tokens in three blocks (`:root`, `@media (prefers-color-scheme: dark) { :root:not([data-theme=light]) }`, `:root[data-theme=dark]`), `@theme inline` color mappings, `@utility glass`, `@utility pill`, `@utility app-bg` and the tone utilities `tone-major`, `tone-minor`, `tone-patch`, `tone-err`, and a `@layer components` block for the motion rules (chevron, `html.no-anim`, view transition, reduced motion). Remove the palette-A `--c-*` tokens, the `success` tokens and the brand gradient.
  **Verify:** `styles_test.go` checks the mock's values in each block, that no `--c-` token and no `success` token remain, that each `@utility` (glass, pill, app-bg) and each component class (`tone-major`, `tone-minor`, `tone-patch`, `tone-err`) appears in the compiled CSS, and that the reduced-motion rule is present.
- [x] 13.4 Catalog. Update `internal/messages/default.yaml` for D12 (added keys, changed defaults, unused keys commented), add the constants and `allKeys` in `keys.go`, and add `total` and `up_to_date` to `placeholderNames`.
  **Verify:** `TestDefaultsHaveEveryKey`, `TestDefaultsKeepPlaceholders` and `TestUnusedKeyOverrideValidates` pass; a new `TestNewKeysRender` renders each added key with its placeholders filled.
- [x] 13.5 View model. Add `StatsView`, `BarSegment` (type and percent, zero omitted), `LegendItem` (every type), `RepoGroupView.Dot`, `Note` and `Chips`, `CountChip`, `EntryView.Name` as the package name of an image with its versions separate, `TypeLabel`, `TabView.Count` and `Selected`, and `tone.go` (type to tone class). Remove `distItem`, `DistPill` and the emoji counters.
  **Verify:** table-driven `document_view_test.go` covers: deps tab first and selected; badge counts; bar percentages (7.7, 88.5, 3.8) and the omitted zero segment; the legend with zeros; dot precedence major over minor over patch; chips for non-zero types only; the fork note; an image entry named by its image; summary `2 actualizaciones` / `para 2 dependencias` and `1 actualización` / `para 1 imagen`.
- [x] 13.6 Components. Write the templ files of D2 (`report`, `logo`, `theme`, `tabs`, `stats`, `chip`, `repo`, `entry`, `errors`, `icon`) and regenerate them. The copy button uses the mock's 14px icons and `--btn-*` colors, and the done state uses `--ok`.
  **Verify:** `components_test.go` renders each component table-driven and checks its markup; `TestEntryEscapesValues` (task 5.3) still passes with the new `DepRow`; `TestToneClasses` checks one tone class per type.
- [x] 13.7 Document shell. `ReportDocument` has `lang="es"`, the viewport meta, `<meta name="color-scheme" content="light dark">`, exactly the Google Fonts stylesheet link (preconnect links allowed), the favicon as a `data:image/svg+xml;base64` URI, the theme script in `<head>`, the report script at the end of `<body>`, an sr-only `<h1>`, and the `copy-status` live region.
  **Verify:** `TestDocumentStaticBehavior` (updated) checks these (the head script and the two-script count are superseded by 13.14), that there is no `<script src>`, no `<img`, no `name=` and no `open` on `details`, and that the body has the `app-bg` class.
- [x] 13.8 Tabs, theme control and pills markup. Tabs are `role="tablist"` with `role="tab"` buttons (`aria-selected`, `aria-controls`, roving `tabindex`) and `role="tabpanel"` sections with `hidden` on the non-selected one. The theme control is a `role="group"` with three `aria-pressed` buttons.
  **Verify:** `TestTabsMarkup` and `TestThemeControlMarkup` check the roles, the ARIA states, the order (Dependencias first) and the `dv-theme` key appears only in the two scripts (superseded by 13.14: no script uses it).
- [x] 13.9 Gate for the scripts and the CSS. Run `biome ci` on the assets through `scripts/verify.sh` steps, and fix what Biome reports in `report.js`, `theme.js` and `input.css` (`theme.js` is deleted by 13.14).
  **Verify:** `biome ci internal/services/depreport/assets` exits 0 in Docker, as the gate runs it.
- [x] 13.10 Regenerate. Run `scripts/generate-depreport-assets.sh`, then again.
  **Verify:** the second run leaves `git diff --stat` unchanged, and `depreport.min.css` contains the three token blocks and the four tone utilities.
- [x] 13.11 Goldens. Regenerate `internal/handlers/testdata/golden/document.html` with `UPDATE_GOLDEN=1`.
  **Verify:** `git diff --stat -- internal/handlers/testdata/golden/` lists only `document.html`; `caption.txt`, `caption_up_to_date.txt` and `failure.txt` are byte-identical to the index; `go test ./internal/handlers/...` passes without `UPDATE_GOLDEN`.
- [x] 13.12 Final gate. Stage the changes (not `graphify-out/` or the temporary review copies), run `scripts/verify.sh`, run `graphify update .`, then copy the golden to `dependency-report-prototype.html` for the user.
  **Verify:** `scripts/verify.sh` exits 0 with every step printing its ok line; `openspec validate --all --strict` is part of it; the copy exists and is not staged.
- [x] 13.13 (user) Compare the deployed report and `dependency-report-prototype.html` with the mock in light, dark, forced Claro and forced Oscuro, at phone and desktop width, and report any difference. Mock removed from the repo; the user verified the deployed report on Mac, iPhone and iPad on 2026-10-12.
  **Verify:** written confirmation from the user, which closes 12.1 for the liquid glass design.
- [x] 13.14 Theme not stored (2026-10-12, user decision). Delete `assets/theme.js`, its `go:embed` and its `<script>` in `<head>`. `report.js` reads and writes no storage: the control starts on Automático on every open, and Claro or Oscuro sets `data-theme` for the page view only. The rationale is the user's request, and no storage errors on `file://`. Rewrite D20, the theme requirement and its scenarios, and regenerate the assets and the golden.
  **Verify:** `TestReportScriptHasNoStorage` (no `localStorage`, `sessionStorage` or `dv-theme` in `report.js`), `TestDocumentStartsOnAutomatic` (no `data-theme` on `<html>`, Automático pressed), `TestReportScriptEmbedded` and `TestDocumentStaticBehavior` (one inline `<script>`); `openspec validate --strict redesign-dependency-report-document` passes; the caption and failure goldens are byte-identical to the index.
