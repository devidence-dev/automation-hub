# Design

## Context

See proposal.md - Why. All report texts already come from one embedded catalog, `internal/messages/default.yaml`, validated at startup against the key constants and placeholder names; `messages:` in the config overrides any key. So the switch is a data change plus the tests, goldens, specs and docs that quote the old texts. No code path chooses a language.

## Goals / Non-Goals

**Goals:** one English default catalog; every quoted default in specs and tests matches it; overrides behave as before.

**Non-Goals:** bundled locales, runtime language selection, renaming keys.

## Decisions

### D1. Translate in place, keep keys and placeholders
Only values change. Keys, placeholder sets and emoji stay, so startup validation and every existing override still pass. *Alternative:* a second `default.es.yaml` selected by config (rejected: a feature nobody asked for; an override already gives Spanish back).

### D2. Translation table (authoritative for code, tests and specs)

| Key | Spanish (old) | English (new) |
|---|---|---|
| `lang` | `es` | `en` |
| `failure.title` | `❌ Renovate falló (exit {exit}) en el pod {pod}` | `❌ Renovate failed (exit {exit}) in pod {pod}` |
| `fork.tag_document` | `fork · solo informativo` | `fork · informational only` |
| `caption.all_up_to_date` | `✅ Todo al día` | `✅ All up to date` |
| `caption.repos_up_to_date` | `✅ {count} repos al día` | `✅ {count} repos up to date` |
| `caption.more_repos` | `… y {count} repos más (ver reporte)` | `… and {count} more repos (see report)` |
| `caption.stats` | `📊 {updated} desactualizadas · {current} al día` | `📊 {updated} outdated · {current} up to date` |
| `caption.tabs` | `🐳 Imágenes: {images} · 📚 Paquetes: {packages}` | `🐳 Images: {images} · 📚 Packages: {packages}` |
| `caption.errors` | `⚠️ {count} errores en el scan` | `⚠️ {count} scan errors` |
| `document.title` | `Reporte de dependencias - Devidence` | `Dependency report - Devidence` |
| `document.brand` | `Reporte de dependencias` | `Dependency report` |
| `document.eyebrow.deps` | `Reporte de dependencias` | `Dependency report` |
| `document.eyebrow.images` | `Reporte de imágenes` | `Image report` |
| `document.tab.images` | `Imágenes` | `Images` |
| `document.tab.deps` | `Dependencias` | `Dependencies` |
| `document.total` | `Total` | `Total` |
| `document.pending` | `Actualizaciones pendientes` | `Pending updates` |
| `document.up_to_date` | `Al día` | `Up to date` |
| `document.distribution` | `Distribución de actualizaciones` | `Update distribution` |
| `document.panel.sub` | `pendientes de {total} · {up_to_date} al día` | `pending out of {total} · {up_to_date} up to date` |
| `document.status.updates_bold` | `{updates} actualizaciones` | `{updates} updates` |
| `document.status.updates_bold_one` | `{updates} actualización` | `{updates} update` |
| `document.status.updates_rest` | `para {deps} dependencias` | `across {deps} dependencies` |
| `document.status.updates_rest_one` | `para {deps} dependencia` | `across {deps} dependency` |
| `document.status.images_rest` | `para {deps} imágenes` | `across {deps} images` |
| `document.status.images_rest_one` | `para {deps} imagen` | `across {deps} image` |
| `document.status.up_to_date` | `Todas las dependencias están al día` | `All dependencies are up to date` |
| `document.status.images_up_to_date` | `Todas las imágenes están al día` | `All images are up to date` |
| `document.theme.group` | `Tema` | `Theme` |
| `document.theme.system` | `Automático` | `Auto` |
| `document.theme.light` | `Claro` | `Light` |
| `document.theme.dark` | `Oscuro` | `Dark` |
| `document.available` | `Actualizaciones disponibles` | `Available updates` |
| `document.columns.package_file` | `Archivo de paquete` | `Package file` |
| `document.columns.repository` | `Repositorio` | `Repository` |
| `document.columns.current_image` | `Imagen actual` | `Current image` |
| `document.columns.latest_image` | `Imagen nueva` | `New image` |
| `document.columns.dependency` | `Dependencia` | `Dependency` |
| `document.columns.current` | `Actual` | `Current` |
| `document.columns.new` | `Nueva` | `New` |
| `document.columns.type` | `Tipo` | `Type` |
| `document.copy.current_image` | `Copiar imagen actual` | `Copy current image` |
| `document.copy.latest_image` | `Copiar imagen nueva` | `Copy new image` |
| `document.copy.copied` | `¡Copiado!` | `Copied!` |
| `document.copy.copied_before` | `Copiar (copiado antes)` | `Copy (copied before)` |
| `document.copy.new_version` | `Copiar nombre@versión` | `Copy name@version` |
| `document.export.button` | `Exportar` | `Export` |
| `document.export.title` | `Exportar actualizaciones` | `Export updates` |
| `document.export.copy` | `Copiar YAML` | `Copy YAML` |
| `document.export.download` | `Descargar YAML` | `Download YAML` |
| `document.export.close` | `Cerrar` | `Close` |
| `document.errors.heading` | `Errores del análisis` | `Scan errors` |
| `document.errors.count` | `{count} errores` | `{count} errors` |
| `document.errors.count_one` | `{count} error` | `{count} error` |
| `document.errors.root` | `raíz` | `root` |

Unchanged: `fork.tag_caption`, `caption.header`, `caption.repo_line`, `caption.emoji.*`, `counts.*`, `document.type.*`, `document.summary.*`, `document.errors.item`, all `bot.*`.

### D3. Language guard test flips
`TestDocumentDefaultsHaveNoEnglishLabels` becomes `TestDocumentDefaultsHaveNoSpanishLabels`: the rendered fixture document has `lang="en"` and contains none of the old Spanish texts of D2 (and no `á é í ó ú ñ ¿ ¡` outside report data). The `lang` override test now overrides to `es` and checks that the labels stay English.

### D4. Spec deltas restate whole requirements
Every requirement whose scenarios quote a D2 text is copied whole under MODIFIED with only the quoted texts (and `lang`) changed; wording that says "Spanish" changes to "English". The `add-report-export` delta (not archived yet) is edited in place to the English export labels; it archives first, then this change.

## Risks / Trade-offs

- [A `messages:` override in Infisical pins some Spanish texts, giving a mixed report] → task (user): check `CONFIG_YAML` for `messages:` and drop or translate it.
- [English texts are longer in places, e.g. `pending out of 283 · 255 up to date`] → layout already wraps; the user checks the deployed report.
- [Missed Spanish string in a test or golden] → the flipped language guard test fails on any leftover default text.

## Migration Plan

Deploy as usual. Rollback: redeploy the previous image, or set Spanish texts under `messages:` without a redeploy of code.
