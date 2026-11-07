# Tasks

Each task names the test or check that proves it. Tasks marked `(user)` are done by the user. Generated files (`*_templ.go`, `assets/depreport.min.css`) are rebuilt with `scripts/generate-depreport-assets.sh` after any `.templ` or `input.css` change, never edited by hand.

## 1. Config

- [x] 1.1 Add `ExportName string \`mapstructure:"export_name"\`` to the dependency report config, with `defaultExportName = "dependency-updates-{date}.yaml"` applied when `dependency_report.export_name` is not set, and the same validation as `document_name` (must contain `{date}`, no `/` or `\`), error naming `dependency_report.export_name`. Document the key, commented with its default, next to `document_name` in `configs/config.yaml.example` and in the README config block.
  **Verify:** `internal/config/settings_test.go`: `TestDependencyReportDefaults` asserts the default; new `TestDependencyReportExportNameNeedsDate` covers `updates.yaml` and `a/{date}.yaml`; `TestDependencyReportCustomSettings` reads `export_name: "renovate-{date}.yaml"`.

## 2. Catalog

- [x] 2.1 Add `document.export.button` "Exportar", `document.export.title` "Exportar actualizaciones", `document.export.copy` "Copiar YAML", `document.export.download` "Descargar YAML", `document.export.close` "Cerrar" to `internal/messages/default.yaml`; add the `KeyDocExport*` constants to `keys.go` and to `allKeys`.
  **Verify:** `TestDefaultsHaveEveryKey` (keys_test.go) passes with the new keys.

## 3. YAML fragments (Go, no markup yet)

- [x] 3.1 Add `Fork bool` to `RepoGroupView`, set in `groupTab` where `Note` is set.
  **Verify:** `TestDocumentForkTag` extended: the `devidence-dev/discord-tts-bot` group has `Fork == true`, a non-fork group `false`.

- [x] 3.2 New `export.go`: fixed schema constants (`scanned_at`, `images`, `dependencies`, `repo`, `fork`, `files`, `path`, `updates`, `name`, `current`, `new`, `type`), `exportHeader(scanTimestamp string) string` and `repoFragment(g RepoGroupView) string`, writing every string with `strconv.Quote` and the indentation of design D2. Each fragment ends in `\n`.
  **Verify:** new `export_test.go`, table-driven: `TestExportOneImage` checks header + `images:\n` + the homelab fragment equals the spec "One image" text byte for byte; `TestExportFork` checks `fork: true`; `TestExportQuotedScalars` checks `1.0.0"\n  injected: true` becomes `"1.0.0\"\n  injected: true"` and no line starts with `injected:`, and `<script>alert(1)</script>` is kept as `"<script>alert(1)</script>"`.

- [x] 3.3 Add `ExportView`, `ExportSection`, `ExportRepo` (design D4) to `documentView`, built in `buildDocumentView` from the built tabs: images section first, then dependencies, repos in card order, empty sections dropped. `NewDocumentPresenter(cat, loc, exportName, dateLayout)` sets `FileName` with `DocumentName`. Update the handler wiring (`internal/handlers/dependency_report.go`) and every test caller of `NewDocumentPresenter`.
  **Verify:** `TestExportSections` in `export_test.go` (fixture report: section order, repo order equals card order, a deps-only report has one section, an all-up-to-date report has none) and `TestExportFileName` (2026-10-13 02:30 UTC in America/Guayaquil gives `dependency-updates-2026-10-12.yaml`; `renovate-{date}.yaml` gives `renovate-2026-10-12.yaml`).

## 4. Markup

- [x] 4.1 Add the `export` icon to `iconMarkup` (14px stroke style). Add `ExportButton` to `Header` between the stamp and `ThemeControl`, rendered only when `Export.Sections` is not empty: glass pill, icon plus label, label visually hidden up to 600px, `aria-label` and `aria-haspopup="dialog"`, `data-export-open`.
  **Verify:** `document_view_test.go`: `TestExportButton` finds one button named `Exportar` for the fixture; `TestDocumentNothingPending` extended to assert no `data-export-open` and no `<dialog`.

- [x] 4.2 New `export.templ` `ExportDialog(view)`: `<dialog>` with `data-header`, `data-file-name`, `aria-labelledby` the title; one `<fieldset data-key=…>` per section with the section checkbox in its `<legend>` and one labeled checkbox per repo (`data-yaml`, short name, fork note, full name in `title`), all `checked`; buttons `Copiar YAML` (`data-export-copy`, `data-copied-text`), `Descargar YAML` (`data-export-download`) and `Cerrar` in a `<form method="dialog">`. Render it in `ReportDocument` after the tabs, only with sections. Regenerate assets.
  **Verify:** `TestExportDialogMarkup`: section labels `Imágenes` then `Dependencias`, every repo checkbox `checked`, `data-file-name="dependency-updates-2026-10-12.yaml"`; `TestDocumentScriptEscaped` and a new case for a `"` in a version confirm `data-yaml` holds `&lt;script&gt;` / `&#34;` and adds no attribute; `TestReportScriptEmbedded` still counts one `<script`.

- [x] 4.3 Style the dialog in `input.css` with the existing glass tokens (backdrop, panel, checkbox accent, disabled buttons), no fixed `min-width` above 320px, width capped at the viewport minus the 16px gutters. Regenerate assets.
  **Verify:** `styles_test.go` responsive check (no `min-width` above 320px) still passes; `scripts/verify.sh` reports no asset drift.

## 5. Behavior (report.js)

- [x] 5.1 Extract `copyText(text)` from `onCopyClick` (clipboard API, fallback text area on missing API or rejection). Entry copy buttons keep their current behavior through it.
  **Verify:** Biome passes (`scripts/verify.sh`); `TestReportScriptEmbedded` passes; manual check in 6.2 that an entry copy still shows the check mark.

- [x] 5.2 Wire the dialog (design D6): open checks every box, syncs and calls `showModal()`; close returns focus to the button; section and repo checkbox sync with `indeterminate`; both actions disabled when nothing is checked; copy builds header + `key:\n` + checked fragments and announces `data-copied-text`; download saves a `text/yaml` blob as `data-file-name`. Without `HTMLDialogElement.prototype.showModal`, the button is hidden. Every query guarded, nothing stored.
  **Verify:** Biome passes; manual checks in 6.2.

## 6. Validation

- [x] 6.1 Run `docker run --rm -v "$PWD":/app -w /app golang:$(cat .go-version) sh -c 'go vet ./... && go test -race ./...'`, then `scripts/verify.sh` (asset drift, Biome, openspec validate), then `graphify update .`.
  **Verify:** all three exit 0.

- [ ] 6.2 (user) Open the fixture document rendered from `internal/handlers/testdata/renovate-report.json` in a desktop browser and on a phone (320px): `Exportar` opens the dialog; clearing `Dependencias` clears its repos; clearing one repo shows the mixed state; nothing checked disables both actions; Escape and `Cerrar` close and focus returns to `Exportar`; `Copiar YAML` output pasted into an LLM parses as the expected YAML; `Descargar YAML` saves `dependency-updates-<date>.yaml`; an entry copy button still works.
  **Verify:** the user confirms in the conversation.
