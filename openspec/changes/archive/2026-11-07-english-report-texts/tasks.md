# Tasks

Each task names the test or check that proves it. Tasks marked `(user)` are done by the user. Archive `add-report-export` before this change (design D4).

## 1. Catalog

- [x] 1.1 Replace every value listed in design D2 in `internal/messages/default.yaml`, set `lang: en`, and update the header comment of the file if it names Spanish. Keys, placeholders and emoji stay unchanged.
  **Verify:** `TestDefaultsHaveEveryKey` (depreport/keys_test.go) and the `internal/messages` catalog tests pass; startup validation of the default catalog still passes (`internal/config` tests).

## 2. Tests and goldens

- [x] 2.1 Flip the language guard (design D3): rename `TestDocumentDefaultsHaveNoEnglishLabels` to `TestDocumentDefaultsHaveNoSpanishLabels`; it asserts `lang="en"` and that the rendered fixture contains none of the old Spanish texts of D2. Change the `lang` override test (`TestDocumentLangOverride` and the catalog override test) to override to `es` and assert the labels stay English.
  **Verify:** both tests pass and fail when one Spanish value is put back into default.yaml (checked once by hand, then reverted).

- [x] 2.2 Update every test that asserts a Spanish default text to the D2 English text: `internal/services/depreport/{caption_test.go,components_test.go,document_test.go,document_view_test.go,failure_test.go,assets_test.go,keys_test.go}`, `internal/handlers/{dependency_report_test.go,command_test.go}` and `internal/messages/catalog_test.go`. Leave test data that is report content (dep names, versions) and email-forwarding tests (`internal/services/processor`) alone.
  **Verify:** `git grep -n -E "Todo al día|al día|Imágenes|Dependencias|Copiar|¡Copiado!|Automático|falló|raíz|solo informativo|Reporte de" -- 'internal/**/*_test.go'` prints nothing.

- [x] 2.3 Regenerate the goldens `internal/handlers/testdata/golden/{caption.txt,caption_up_to_date.txt,document.html}` with `UPDATE_GOLDEN=1` and read the diff: only texts and `lang` change.
  **Verify:** `TestGoldenBaseline` passes without `UPDATE_GOLDEN`; `git diff --stat` of the goldens shows text-only changes.

## 3. Comments and docs

- [x] 3.1 Update comments that name Spanish labels in `internal/services/depreport/report.templ` and `theme.templ` (e.g. "Automático" -> "Auto"), then run `sh scripts/generate-depreport-assets.sh`.
  **Verify:** `scripts/verify.sh` reports no asset drift once the generated files are staged.

- [x] 3.2 README: the report description says the caption and labels are English by default (`lang="en"`), the long-list marker is `… and N more repos (see report)`, the report-root error is `root: <msg>`, and Spanish is available through `messages:` overrides.
  **Verify:** `git grep -n -E "Spanish|repos más|raíz" README.md` prints nothing about the report.

- [x] 3.3 In `openspec/config.yaml` context, change "dependency-report message texts are fixed Spanish strings" to state that all user-facing texts, including the dependency report, are English by default.
  **Verify:** `openspec validate --all --strict` passes.

## 4. Validation

- [x] 4.1 Run `docker run --rm -v "$PWD":/app -w /app golang:$(cat .go-version) sh -c 'go vet ./... && go test -race ./...'`, then `scripts/verify.sh`, then `graphify update .`.
  **Verify:** all exit 0.

- [x] 4.2 (user) Check `CONFIG_YAML` in Infisical for a `messages:` section; drop or translate any Spanish override so the report is not mixed.
  **Verify:** the user confirms in the conversation.

- [x] 4.3 (user) Open the next deployed report: caption and document are English, and the longer English lines (panel sub line, status line, header) fit at 320px and on desktop.
  **Verify:** the user confirms in the conversation.
