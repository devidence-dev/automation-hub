# Spec Delta

## MODIFIED Requirements

### Requirement: Header and logo
The header SHALL show the devidence logo as one inline SVG with `role="img"` and the label `devidence.dev`, the scan timestamp in a stamp, the export button (see "Export button") and the theme control. The logo SHALL take its wordmark colors from the tokens of the active scheme, and the document SHALL NOT reference any image file.

#### Scenario: Logo follows the theme
- **WHEN** the theme is dark, from the system or from the control
- **THEN** the wordmark fills resolve to `#F2F7FB` and `#9FB6C6`, and in light to `#0B1B25` and `#334C5A`

#### Scenario: Logo and icon are clean
- **WHEN** the embedded `assets/logo.svg` and `assets/icon.svg` are read
- **THEN** neither contains a `<script`, a `<metadata` element or an `on*=` attribute, and the favicon is a `data:image/svg+xml` URI

## ADDED Requirements

### Requirement: Export button
The header SHALL show an export button, labeled by `document.export.button`, when at least one update is listed in either tab. It SHALL open the export dialog as a modal. Up to 600px the button SHALL show its icon only and keep the label as its accessible name. When no update is listed, the document SHALL have no export button and no export dialog.

#### Scenario: Updates listed
- **WHEN** the document of a report where `devidence-dev/homelab` has the image `ghcr.io/immich-app/immich-server` from `v1.118.0` to `v1.119.0` is rendered
- **THEN** the header has one button named `Exportar`, and activating it opens the export dialog as a modal with focus inside it

#### Scenario: Nothing listed
- **WHEN** the report has 12 docker deps and 30 non-docker deps, none with a `major`, `minor` or `patch` update
- **THEN** the document has no `Exportar` button and no export dialog

#### Scenario: Phone width
- **WHEN** the document is shown in a 320 CSS pixel wide viewport
- **THEN** the header shows the export button as an icon, named `Exportar`, with no horizontal page scrolling

### Requirement: Export selection
The export dialog SHALL have one section for images (`document.tab.images`) and one for dependencies (`document.tab.deps`), in that order, each with a section checkbox and one checkbox per repo with listed updates in that tab, in the order of that tab's repo cards. A section without listed updates SHALL be left out. Every checkbox SHALL start checked on each open.

#### Scenario: Both sections
- **WHEN** images are listed for `devidence-dev/homelab` and dependencies for `devidence-dev/QuantWarden` and `devidence-dev/homelab`
- **THEN** the dialog shows `Imágenes` with the repo `homelab`, then `Dependencias` with `QuantWarden` and `homelab` in card order, all checked

#### Scenario: Section toggles its repos
- **WHEN** the user clears the `Dependencias` section checkbox
- **THEN** every repo checkbox of `Dependencias` is cleared and the `Imágenes` checkboxes keep their state

#### Scenario: Mixed section
- **WHEN** `Dependencias` lists `QuantWarden` and `homelab` and the user clears only `homelab`
- **THEN** the `Dependencias` section checkbox shows the mixed (indeterminate) state, and checking it again checks both repos

#### Scenario: Only dependencies listed
- **WHEN** no image update is listed and `devidence-dev/QuantWarden` has a dependency update
- **THEN** the dialog shows only the `Dependencias` section

#### Scenario: Reopened dialog
- **WHEN** the user clears `homelab`, closes the dialog with Escape and opens it again
- **THEN** every checkbox is checked again

### Requirement: Export actions
The dialog SHALL offer a copy action (`document.export.copy`) and a download action (`document.export.download`), both disabled while no repo is checked, and a close button (`document.export.close`). The close button and Escape SHALL close the dialog.

#### Scenario: Nothing checked
- **WHEN** the user clears both section checkboxes
- **THEN** `Copiar YAML` and `Descargar YAML` are disabled, and checking `homelab` again enables both

#### Scenario: Close
- **WHEN** the dialog is open and the user presses Escape or activates `Cerrar`
- **THEN** the dialog closes and focus returns to the `Exportar` button

### Requirement: Export copy
The copy action SHALL put the export on the clipboard, falling back to a hidden text area when the clipboard API is missing or rejects the write, and SHALL announce `document.copy.copied` in the polite live region.

#### Scenario: Copy
- **WHEN** the scan was received on 2026-10-12 at 09:07 America/Guayaquil and the user activates `Copiar YAML` with every repo checked
- **THEN** the clipboard holds the export of every listed update and `¡Copiado!` is announced

#### Scenario: Clipboard API blocked
- **WHEN** `navigator.clipboard` rejects the write
- **THEN** the export is copied through a hidden text area and `¡Copiado!` is announced

### Requirement: Export download
The download action SHALL save the export as a `text/yaml` file named by `dependency_report.export_name` (default `dependency-updates-{date}.yaml`), with `{date}` replaced by the scan date in the configured timezone, formatted with `dependency_report.date_layout`. `export_name` SHALL follow the same validation as `document_name`.

#### Scenario: Download
- **WHEN** the scan was received on 2026-10-12 at 09:07 America/Guayaquil and the user activates `Descargar YAML` with every repo checked
- **THEN** the browser saves `dependency-updates-2026-10-12.yaml` with the same content as `Copiar YAML` puts on the clipboard

#### Scenario: Date near midnight UTC
- **WHEN** the report is received at 2026-10-13 02:30 UTC, which is 2026-10-12 21:30 in America/Guayaquil
- **THEN** the download is named `dependency-updates-2026-10-12.yaml`

#### Scenario: Configured export name
- **WHEN** the config has `export_name: "renovate-{date}.yaml"` and the scan was received on 2026-10-12 at 09:07 America/Guayaquil
- **THEN** the download is named `renovate-2026-10-12.yaml`

#### Scenario: Invalid export name
- **WHEN** the config has `export_name: "updates.yaml"` (no `{date}`) or `export_name: "a/{date}.yaml"`
- **THEN** loading the config fails with an error naming `dependency_report.export_name`, and the hub does not start with that config

### Requirement: Export content
The export SHALL be a YAML document: `scanned_at` with the header timestamp, then `images` and `dependencies`, each present only when one of its repos is checked. Each is a list of the checked repos in card order, each with `repo` (full name), `fork`, and `files` in card order, each with `path` and `updates` in card order, each with `name`, `current`, `new` and `type` (`major`, `minor` or `patch`). Field names SHALL be fixed and not translated.

#### Scenario: One image
- **WHEN** only `devidence-dev/homelab` is checked under `Imágenes`, with `ghcr.io/immich-app/immich-server` in `kubernetes/apps/immich/values.yaml` from `v1.118.0` to `v1.119.0`, type `minor`, scanned 2026-10-12 09:07 America/Guayaquil
- **THEN** the export is exactly:
  ```yaml
  scanned_at: "2026-10-12 09:07 America/Guayaquil"
  images:
    - repo: "devidence-dev/homelab"
      fork: false
      files:
        - path: "kubernetes/apps/immich/values.yaml"
          updates:
            - name: "ghcr.io/immich-app/immich-server"
              current: "v1.118.0"
              new: "v1.119.0"
              type: "minor"
  ```

#### Scenario: Fork repo
- **WHEN** `devidence-dev/discord-tts-bot` is a configured fork and is checked under `Dependencias`
- **THEN** its entry has `fork: true`

#### Scenario: Unchecked section left out
- **WHEN** every `Imágenes` repo is cleared and `QuantWarden` is checked under `Dependencias`
- **THEN** the export has `scanned_at` and `dependencies` with only `devidence-dev/QuantWarden`, and no `images` key

### Requirement: Safe export content
Every string value of the export SHALL be a YAML double-quoted scalar with `"`, `\` and control characters escaped, so report content stays one value and never adds keys, list items or documents. The export data in the document SHALL be escaped for HTML like any other report text.

#### Scenario: Quote and newline in a version
- **WHEN** a dep has the current version `1.0.0"\n  injected: true`
- **THEN** its `current` value is the one double-quoted scalar `"1.0.0\"\n  injected: true"` and the export has no `injected` key

#### Scenario: Script in a dependency name
- **WHEN** a dep is named `<script>alert(1)</script>` and has a `patch` update from `1.0.0` to `1.0.1`
- **THEN** the document contains no `<script>alert(1)</script>`, and the copied export holds the name `"<script>alert(1)</script>"` as text
