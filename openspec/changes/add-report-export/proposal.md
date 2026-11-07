# Proposal

## Why

The HTML dependency report lists every pending image and dependency update, but handing them to an LLM (to plan or apply the bumps) means copying entries one by one. An export of the selected updates as one LLM-friendly text block removes that step.

## What Changes

- The report header gets an **Exportar** button, shown only when at least one update is listed.
- The button opens a modal dialog with two sections, **Imágenes** and **Dependencias**, each with a section checkbox and one checkbox per repo with pending updates in that tab. All start checked; a section checkbox selects or clears its repos and shows the mixed state when only some are checked. A section without pending repos is not shown.
- The export is **YAML**: the scan timestamp, then `images:` and `dependencies:` lists of the checked repos, each with its fork flag, package files and updates (name, current, new, type). YAML carries the same structure as JSON with fewer tokens (no braces, no quoted keys). Every string value is double-quoted, so report content cannot break the structure.
- Two actions, disabled while nothing is checked: **Copiar YAML** (clipboard, with the existing fallback and announcement) and **Descargar YAML** (a `.yaml` file named from a catalog pattern with the scan date).
- The YAML fragments are built in Go at render time, so the format is unit-tested; the script only joins the checked fragments.
- New Spanish labels come from the message catalog.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `dependency-report`: the header requirement gains the export button; new requirements for the export dialog, the YAML content and its safety.

## Non-goals

- No JSON or other format toggle.
- No per-file or per-update selection; the finest grain is one repo in one section.
- Scan errors and up-to-date counts are not exported.
- No change to the caption, delivery or webhook.
- No server-side export endpoint.
- No saved selection: every open starts with all repos checked.

## Impact

- `internal/services/depreport`: export builder and YAML writer, `ExportDialog` component, header change, `report.js` (dialog, checkboxes, copy refactor, download), `input.css`, regenerated `*_templ.go` and `depreport.min.css`, new `export` icon.
- `internal/messages/default.yaml` and `keys.go`: new `document.export.*` keys.
- No new Go module or tool: YAML is written with the standard library and the dialog is the native `<dialog>` element.
