# Design

## Context

See proposal.md - Why. The report is one self-contained HTML file rendered by `DocumentPresenter` (templ components, `report.js` inlined once, compiled Tailwind CSS inlined). Current state that shapes the approach:

- `buildDocumentView` already produces the ordered repo cards per tab (`TabView.Groups []RepoGroupView`, files and entries in card order). That ordering is what the export must follow.
- Tests assert the document has exactly one `<script` element (`TestReportScriptEmbedded`, `TestDocumentStaticBehavior`). The main spec's "Self-contained document" scenario still says two inline scripts; the code and tests say one. This change keeps one script and does not touch that scenario (pre-existing mismatch, see Open Questions).
- There is no JS test harness; `report.js` is only linted by Biome. Logic that needs tests must live in Go.
- `go.yaml.in/yaml/v3` is only an indirect module; using it directly would be a new direct dependency, which needs user approval.
- The download name follows the existing `dependency_report.document_name` + `date_layout` pattern and `DocumentName()`.

## Goals / Non-Goals

**Goals:** the YAML format is produced and tested in Go; the script stays a thin joiner; the document keeps one script element and no network file.

**Non-Goals:** a general YAML encoder; a JS test harness.

## Decisions

### D1. YAML, every string double-quoted
YAML over JSON: the nested lists cost fewer tokens (no braces, brackets, quoted keys, commas) and read naturally for LLMs. Every string scalar is written with `strconv.Quote`. Go's escapes (`\"`, `\\`, `\n`, `\t`, `\a`, `\v`, `\xNN`, `\uNNNN`, `\UNNNNNNNN`) are all valid YAML 1.2 double-quoted escapes, so any report text stays one scalar. `fork` is a bare `true`/`false`. *Alternatives:* `encoding/json` per value (also valid YAML, but escapes `<>&` as `<`, which hurts readability); `yaml.v3` (new direct dependency, not needed for a fixed shape); plain scalars (unsafe for values like `a: b` or `- x`).

### D2. Fragments rendered in Go, joined in JS
New `export.go` builds, per tab, one YAML fragment per repo card from `RepoGroupView` (`repo`, `fork`, `files[].path`, `files[].updates[]` with `name`, `current`, `new`, `type`). Each fragment is the list item of that repo, indented for its section, ending in `\n`. The header is `scanned_at: <quoted ScanTimestamp>\n`. The final document is:

```
header + for each section with a checked repo: key + ":\n" + checked fragments in order
```

The Go test checks that header + `images:\n` + the homelab fragment equals the spec's "One image" text byte for byte, and covers the quote/newline injection. *Alternative:* embed a JSON model and serialize in JS (untestable here, and needs a JS serializer).

### D3. Data in attributes, not a JSON script
Each repo checkbox carries its fragment in `data-yaml`; each section carries its key (`images` / `dependencies`) in `data-key`; the dialog carries `data-header` and `data-file-name`. templ escapes attribute values, so report text is HTML-safe (spec "Safe export content") and the document keeps one `<script>`. *Alternative:* `templ.JSONScript` (adds a second script element and breaks the existing invariant).

### D4. View model
- `RepoGroupView` gains `Fork bool` (set where `Note` is set), so the export does not parse the note text.
- `documentView` gains `Export ExportView{Header, FileName, Sections []ExportSection}`; `ExportSection{Key, Label, Repos []ExportRepo}`; `ExportRepo{Repo, Short, Note, YAML}`. Sections are built images first, then dependencies, from the already built tabs; empty sections are dropped. No sections means no button and no dialog.
- Field names and section keys are fixed schema constants in `export.go` (a machine format, not catalog text).
- `NewDocumentPresenter(cat, loc, exportName, dateLayout)`; `FileName = DocumentName(s.ReceivedAt, loc, exportName, dateLayout)`. The handler passes `cfg.ExportName` and `cfg.DateLayout`.

### D5. Config
`DependencyReportConfig.ExportName` (`export_name`, default `dependency-updates-{date}.yaml`), defaulted and validated like `document_name` (must contain `{date}`, no `/` or `\`). Documented in `configs/config.yaml.example`.

### D6. Dialog markup and behavior
Native `<dialog>` opened with `showModal()`: modal focus, inert background and Escape come from the browser. Inside: title, one `<fieldset>` per section with the section checkbox in the `<legend>` and the repo checkboxes as labeled list items (short name, fork note, full name in `title`), then the action buttons. `report.js`:
- on open: check every box, sync, `showModal()`; on `close`: focus the `Exportar` button.
- section checkbox `change`: set all its repos; repo `change`: section `checked` = all, `indeterminate` = some.
- after any change: disable both actions when no repo is checked.
- copy: build the text, then the shared copy routine.
- download: `Blob([text], {type: "text/yaml"})`, object URL, temporary `<a download>`, click, revoke.

### D7. Copy routine shared
The body of `onCopyClick` becomes `copyText(text)` (clipboard API, fallback text area) plus `announce`. Entry buttons keep `flash(btn)`; the export copy announces `data-copied-text` of its button. Behavior of entry buttons is unchanged.

### D8. Header layout
The button sits between the stamp and the theme control, styled like the glass stamp. Its text label is hidden up to 600px (`max-[600px]:sr-only` or the existing breakpoint pattern) and `aria-label` keeps the name. New `export` icon in `iconMarkup` (14px stroke, same style).

### D9. Catalog keys
`document.export.button` "Exportar", `document.export.title` "Exportar actualizaciones", `document.export.copy` "Copiar YAML", `document.export.download` "Descargar YAML", `document.export.close` "Cerrar". Section labels reuse `document.tab.images` / `document.tab.deps`; the copied text reuses `document.copy.copied`. Keys added to `keys.go` and `allKeys`.

## Risks / Trade-offs

- [Telegram's in-app viewer may block `a[download]` or blob URLs] → the copy action works there; the file can also be opened in a browser.
- [Fragments duplicate the card data in the HTML] → a few KB per report; far under Telegram's document limit.
- [Hand-written YAML drifts from valid YAML] → fixed shape, exact-text Go test, and every scalar quoted by `strconv.Quote`.
- [Older browsers without `<dialog>`] → the button checks `showModal` and stays hidden without it; the cards and entry copy buttons still work.

## Migration Plan

No data migration. `export_name` is optional with a default, so existing `CONFIG_YAML` keeps working. Rollback is a redeploy of the previous image.

## Open Questions

- The main spec's "Self-contained document" scenario says two inline scripts while code and tests say one. Left as is; fix in a separate spec cleanup.
