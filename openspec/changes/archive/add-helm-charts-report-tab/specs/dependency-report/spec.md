# Spec Delta

## ADDED Requirements

### Requirement: Charts count in the caption
The caption's tabs line SHALL be the text of `caption.tabs` followed by the text of `caption.charts`, whose default is ` · ⛵ Charts: {charts}`. An override of `caption.tabs` written before the "Charts" tab existed SHALL still load. An override of `caption.charts` SHALL keep `{charts}`.

#### Scenario: Existing override of the tabs line
- **WHEN** the config has `messages: {caption: {tabs: "Imágenes: {images} · Paquetes: {packages}"}}` and the report has 1, 2 and 1 deps with listed updates in "Images", "Dependencies" and "Charts"
- **THEN** the hub starts and the line reads `Imágenes: 1 · Paquetes: 2 · ⛵ Charts: 1`

#### Scenario: Override that drops the placeholder
- **WHEN** the config has `messages: {caption: {charts: " · Charts"}}`
- **THEN** config loading fails with an error that names `caption.charts` and `{charts}`, and the hub does not start with that config

### Requirement: Tab list at narrow widths
Up to 420 CSS pixels each tab SHALL show its label above its count badge, and the three tabs SHALL fit in the tab list with no horizontal page scrolling. Above 420 CSS pixels each tab SHALL show its label and its badge on one row.

#### Scenario: Narrowest width
- **WHEN** the fixture document, with the badges `52`, `1` and `0`, is shown in a 320 CSS pixel wide viewport
- **THEN** "Dependencies", "Images" and "Charts" sit on one row inside the tab list, each label is above its badge, and the page does not scroll horizontally

#### Scenario: Wide viewport
- **WHEN** the fixture document is shown in a 760 CSS pixel wide viewport
- **THEN** each tab shows its label and its badge on one row

#### Scenario: Pill follows the selected tab
- **WHEN** the fixture document is shown in a 320 CSS pixel wide viewport and the user selects "Charts"
- **THEN** the pill ends under "Charts" and covers its label and its badge

## MODIFIED Requirements

### Requirement: Summary caption
The caption SHALL be sent with HTML parse mode and SHALL contain, in this order: the header `caption.header`, a blank line, one line per repository with a listed update (see "Caption repository lines"), `✅ <N> repos up to date` (N = repositories with no pending update) omitted when N is 0, a blank line, the stats lines `📊 <X> outdated · <Y> up to date` and `🐳 Images: <a> · 📚 Packages: <b> · ⛵ Charts: <c>`, and `⚠️ <N> scan errors` only when a repository or the report root has a `level >= 50` problem.

#### Scenario: No listed update
- **WHEN** no dependency has a listed update
- **THEN** the caption is the header, a blank line, `✅ All up to date`, a blank line and the stats lines, with no repository line and no `✅ <N> repos up to date` line

#### Scenario: Stats lines across both tabs
- **WHEN** the report of 2026-10-12 has 3 docker deps, one with a `minor` update, 5 deps of other datasources, two of them with a `patch` and a `major` update, and 4 `helm` deps, one with a `major` update, and 2 repositories have no pending update
- **THEN** the caption header is the default `📦 <b>Dependency report</b> · {date}` with `{date}` the file-name date, so `📦 <b>Dependency report</b> · 2026-10-12`; the caption has `✅ 2 repos up to date` followed by a blank line; the stats count the three tabs, so `📊 4 outdated · 8 up to date` (1 + 2 + 1 with a listed update, 2 + 3 + 3 without) and `🐳 Images: 1 · 📚 Packages: 2 · ⛵ Charts: 1` (1 in "Images", 2 in "Dependencies", 1 in "Charts")

#### Scenario: Chart update in a repository line
- **WHEN** `devidence-dev/homelab` has only the `helm` dep `traefik` with a `major` update
- **THEN** its line is `🔴 homelab — 1 major`

#### Scenario: Error count line
- **WHEN** `devidence-dev/QuantWarden` has a problem with `level` 50, the report root has a problem with `level` 50, and another repository has only a problem with `level` 40
- **THEN** the caption's last line is `⚠️ 2 scan errors`, and the `level` 40 problem adds nothing to the caption

#### Scenario: Caption over the limit
- **WHEN** 60 repositories with long names have pending updates and the caption with all its repository lines would exceed 1024 characters
- **THEN** repository lines are dropped from the end, a line `… and <N> more repos (see report)` follows the last kept one, and the header, the `✅` line and the stats lines are kept, so that the caption has at most 1024 characters

### Requirement: Document tabs
The document SHALL have three tabs, "Dependencies" first, "Images" second and "Charts" third, each with a count badge of its listed updates. Clicking a tab, or pressing Enter or Space on it, SHALL show its panel. "Images" lists every dependency whose Renovate `datasource` is `docker`, "Charts" every `helm` one, and "Dependencies" every other one. Only `major`, `minor` or `patch` updates are listed.

#### Scenario: Image dependency
- **WHEN** `devidence-dev/homelab` has the file `kubernetes/apps/media/immich/values.yaml` with dep `ghcr.io/immich-app/immich-server`, datasource `docker`, current `v1.118.0` and a `minor` update to `v1.119.0`
- **THEN** the "Images" panel has, in the section `homelab`, under the file `kubernetes/apps/media/immich/values.yaml`, an entry named `ghcr.io/immich-app/immich-server` from `v1.118.0` to `v1.119.0` with a `minor` chip, and "Dependencies" has no entry for it

#### Scenario: Non-image dependency
- **WHEN** `devidence-dev/QuantWarden` has `go.mod` with dep `golang.org/x/net`, datasource `go`, current `0.30.0` and a `major` update to `1.0.0`
- **THEN** the "Dependencies" panel has, in the section `QuantWarden`, under the file `go.mod`, an entry `golang.org/x/net` from `0.30.0` to `1.0.0` with a `major` chip, and "Images" has no entry for it

#### Scenario: Chart dependency
- **WHEN** `devidence-dev/homelab` has the file `kubernetes/infrastructure/traefik/application.yaml` with dep `traefik`, datasource `helm`, current `42.3.0` and a `major` update to `43.0.0`
- **THEN** the "Charts" panel has, in the section `homelab`, under that file, an entry `traefik` from `42.3.0` to `43.0.0` with a `major` chip, and neither "Dependencies" nor "Images" has an entry for it

#### Scenario: Chart in an OCI registry
- **WHEN** `devidence-dev/homelab` has the dep `ghcr.io/acme/charts/api` with datasource `docker` in `kubernetes/infrastructure/api/application.yaml`
- **THEN** it is listed in "Images" and "Charts" has no entry for it

#### Scenario: Same chart in several files
- **WHEN** `app-template` at `5.3.0`, datasource `helm`, has a `minor` update to `5.4.0` in `kubernetes/applicationset.yaml` and in `kubernetes/infrastructure/redis/application.yaml`
- **THEN** the card `homelab` of "Charts" has one entry for it under each file, the tab badge is `2`, and the panel counts 2 dependencies

#### Scenario: Chart lookup failed
- **WHEN** `devidence-dev/homelab` has the `helm` dep `vault` with no updates and a problem with `level` 50 and `msg` `Failed to look up helm package vault`
- **THEN** "Charts" has no entry for `vault` and counts it as up to date, and the errors card has a row with `devidence-dev/homelab` and that message

#### Scenario: Grouping order
- **WHEN** `devidence-dev/QuantWarden` has the package files `go.mod` and `Dockerfile` with listed updates in "Dependencies" and "Images" respectively
- **THEN** each panel has one section `QuantWarden`, and file groups within a section are sorted by package file path

#### Scenario: Dockerfile image and missing datasource
- **WHEN** `devidence-dev/discord-tts-bot` has a `Dockerfile` dep `golang` with datasource `docker`, and another dep with a `patch` update has no `datasource` field
- **THEN** `golang` is listed in "Images" and the other dep is listed in "Dependencies"

#### Scenario: Digest only
- **WHEN** a dep's only update is of type `digest` or `pin`
- **THEN** it has no entry in any tab

#### Scenario: Several updates for one dep
- **WHEN** `github.com/gorilla/mux` at `1.8.0` has a `major` update to `2.0.0` and a `minor` update to `1.9.0`
- **THEN** "Dependencies" has two entries for it, one with a `major` chip and one with a `minor` chip

#### Scenario: Exact duplicates
- **WHEN** `go.uber.org/zap` at `1.27.0` has the same `patch` update to `1.27.1` listed twice
- **THEN** "Dependencies" has one row for it

#### Scenario: Fork tag
- **WHEN** `devidence-dev/discord-tts-bot` is in `dependency_report.forks` and has a listed update
- **THEN** its repo card shows the short name `discord-tts-bot` followed by the note `fork · informational only`, and its expanded card shows the full name `devidence-dev/discord-tts-bot`

#### Scenario: Tab totals
- **WHEN** the report has 3 docker deps, one of them with a `minor` update, 5 deps of other datasources, two of them with a `patch` update and a `major` update, and 4 `helm` deps, one of them with a `major` update
- **THEN** "Dependencies" has the badge `2` and its panel shows 5 total and 3 up to date, "Images" has the badge `1` and shows 3 total and 2 up to date, and "Charts" has the badge `1` and shows 4 total and 3 up to date

#### Scenario: Keyboard
- **WHEN** the focus is on "Images" and the user presses Right
- **THEN** the focus moves to "Charts" without changing the panel, and Enter or Space then shows it; the selected tab has `aria-selected="true"` and the others have `false`

#### Scenario: Document fork tag override
- **WHEN** the config has `messages: {fork: {tag_document: "fork"}}` and `devidence-dev/discord-tts-bot` is in `dependency_report.forks` with a listed update
- **THEN** its repo card shows the note `fork`, and the caption line is unchanged

### Requirement: Nothing pending document
When no dependency has a listed update, the system SHALL still send the document, with `✅ All up to date` in its caption in place of the repository lines, showing zero updates in the three tabs, the up-to-date status line of each tab and no repo cards.

#### Scenario: All up to date
- **WHEN** the report has 12 docker deps, 30 deps of other datasources and 4 `helm` deps, none with a `major`, `minor` or `patch` update
- **THEN** one document is sent whose caption contains `✅ All up to date` and `📊 0 outdated · 46 up to date`, whose tabs show badges `0`, `0` and `0`, whose panels show 0 updates, with 30, 12 and 4 up to date, the status lines `All dependencies are up to date`, `All images are up to date` and `All charts are up to date`, and no repo cards

### Requirement: Summary panel
Each tab SHALL show a panel with the eyebrow `document.eyebrow.deps`, `document.eyebrow.images` or `document.eyebrow.charts`, the big number of listed updates, the sub line `document.panel.sub`, a stacked bar of the update types and a legend in the order patch, minor, major. The bar SHALL omit zero segments; the legend SHALL keep every type, with zero counts.

#### Scenario: Distribution bar of the deps tab
- **WHEN** the "Dependencies" tab has 4 patch, 46 minor and 2 major updates
- **THEN** the bar has three segments of widths 7.7%, 88.5% and 3.8%, in patch, minor and major order, and the legend reads `patch 4`, `minor 46`, `major 2`

#### Scenario: Zero segment omitted
- **WHEN** the "Images" tab has 1 minor update and no patch or major update
- **THEN** the bar has one segment of width 100% and the legend still reads `patch 0`, `minor 1` and `major 0`

#### Scenario: Summary line
- **WHEN** the "Dependencies" tab has 52 listed updates for 52 dependencies
- **THEN** the line under the panel shows `52 updates` in bold and `across 52 dependencies`

#### Scenario: Summary line of the charts tab
- **WHEN** the "Charts" tab has 3 listed updates for 2 charts
- **THEN** its eyebrow is `Chart report` and the line under the panel shows `3 updates` in bold and `across 2 charts`

### Requirement: Copy buttons
Every entry SHALL have one icon-only copy button, named by its catalog label. It copies the new value only: `<depName>@<new version>` for a dependency or a chart (`document.copy.new_version`) and `<image>:<new version>` for an image (`document.copy.latest_image`). A successful copy SHALL show a check mark in the `--ok` color for about 2 seconds and announce `document.copy.copied` in a polite live region.

#### Scenario: Image copy
- **WHEN** the user copies the image `ghcr.io/immich-app/immich-server` from `v1.118.0` to `v1.119.0`
- **THEN** the clipboard holds `ghcr.io/immich-app/immich-server:v1.119.0`, the button is named `Copy new image`, and the check mark shows for about 2 seconds while `Copied!` is announced

#### Scenario: Image entry has one copy button
- **WHEN** the image entry of the previous scenario is rendered
- **THEN** it has exactly one copy button, and no copy value is `ghcr.io/immich-app/immich-server:v1.118.0`

#### Scenario: Dependency copy
- **WHEN** the user copies `golang.org/x/net` from `0.30.0` to `1.0.0`
- **THEN** the clipboard holds `golang.org/x/net@1.0.0`, the button is named `Copy name@version`, and `Copied!` is announced

#### Scenario: Chart copy
- **WHEN** the user copies the chart `traefik` from `42.3.0` to `43.0.0`
- **THEN** the clipboard holds `traefik@43.0.0`, the button is named `Copy name@version`, and `Copied!` is announced

#### Scenario: Clipboard API blocked
- **WHEN** `navigator.clipboard` rejects the write
- **THEN** the text is copied through a hidden text area and the check mark shows while `Copied!` is announced

### Requirement: Export button
The header SHALL show an export button, labeled by `document.export.button`, when at least one update is listed in any tab. It SHALL open the export dialog as a modal. The button SHALL show its icon and its label at every viewport width. When no update is listed, the document SHALL have no export button and no export dialog.

#### Scenario: Updates listed
- **WHEN** the document of a report where `devidence-dev/homelab` has the image `ghcr.io/immich-app/immich-server` from `v1.118.0` to `v1.119.0` is rendered
- **THEN** the header has one button named `Export`, and activating it opens the export dialog as a modal with focus inside it

#### Scenario: Only a chart listed
- **WHEN** the only listed update of the report is the `helm` dep `traefik` from `42.3.0` to `43.0.0`
- **THEN** the header has one button named `Export`

#### Scenario: Nothing listed
- **WHEN** the report has 12 docker deps and 30 non-docker deps, none with a `major`, `minor` or `patch` update
- **THEN** the document has no `Export` button and no export dialog

#### Scenario: Phone width
- **WHEN** the document is shown in a 320 CSS pixel wide viewport
- **THEN** the header shows the export button with its icon and the visible label `Export`, with no horizontal page scrolling

### Requirement: Export selection
The export dialog SHALL have one section for images (`document.tab.images`), one for dependencies (`document.tab.deps`) and one for charts (`document.tab.charts`), in that order, each with a section checkbox and one checkbox per repo with listed updates in that tab, in the order of that tab's repo cards. A section without listed updates SHALL be left out. Every checkbox SHALL start checked on each open.

#### Scenario: Both sections
- **WHEN** images are listed for `devidence-dev/homelab` and dependencies for `devidence-dev/QuantWarden` and `devidence-dev/homelab`
- **THEN** the dialog shows `Images` with the repo `homelab`, then `Dependencies` with `QuantWarden` and `homelab` in card order, all checked

#### Scenario: Three sections
- **WHEN** images, dependencies and charts are listed for `devidence-dev/homelab`
- **THEN** the dialog shows `Images`, `Dependencies` and `Charts` in that order, each with the repo `homelab`, all checked

#### Scenario: Section toggles its repos
- **WHEN** the user clears the `Dependencies` section checkbox
- **THEN** every repo checkbox of `Dependencies` is cleared and the `Images` checkboxes keep their state

#### Scenario: Mixed section
- **WHEN** `Dependencies` lists `QuantWarden` and `homelab` and the user clears only `homelab`
- **THEN** the `Dependencies` section checkbox shows the mixed (indeterminate) state, and checking it again checks both repos

#### Scenario: Only dependencies listed
- **WHEN** no image or chart update is listed and `devidence-dev/QuantWarden` has a dependency update
- **THEN** the dialog shows only the `Dependencies` section

#### Scenario: Reopened dialog
- **WHEN** the user clears `homelab`, closes the dialog with Escape and opens it again
- **THEN** every checkbox is checked again

### Requirement: Export content
The export SHALL be a YAML document: `scanned_at` with the header timestamp, then `images`, `dependencies` and `charts`, each present only when one of its repos is checked. Each is a list of the checked repos in card order, each with `repo` (full name), `fork`, and `files` in card order, each with `path` and `updates` in card order, each with `name`, `current`, `new` and `type` (`major`, `minor` or `patch`). Field names SHALL be fixed and not translated.

#### Scenario: One image
- **WHEN** only `devidence-dev/homelab` is checked under `Images`, with `ghcr.io/immich-app/immich-server` in `kubernetes/apps/immich/values.yaml` from `v1.118.0` to `v1.119.0`, type `minor`, scanned 2026-10-12 09:07 America/Guayaquil
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

#### Scenario: One chart
- **WHEN** only `devidence-dev/homelab` is checked under `Charts`, with `traefik` in `kubernetes/infrastructure/traefik/application.yaml` from `42.3.0` to `43.0.0`, type `major`, scanned 2026-10-12 09:07 America/Guayaquil
- **THEN** the export is exactly:
  ```yaml
  scanned_at: "2026-10-12 09:07 America/Guayaquil"
  charts:
    - repo: "devidence-dev/homelab"
      fork: false
      files:
        - path: "kubernetes/infrastructure/traefik/application.yaml"
          updates:
            - name: "traefik"
              current: "42.3.0"
              new: "43.0.0"
              type: "major"
  ```

#### Scenario: Fork repo
- **WHEN** `devidence-dev/discord-tts-bot` is a configured fork and is checked under `Dependencies`
- **THEN** its entry has `fork: true`

#### Scenario: Unchecked section left out
- **WHEN** every `Images` repo is cleared and `QuantWarden` is checked under `Dependencies`
- **THEN** the export has `scanned_at` and `dependencies` with only `devidence-dev/QuantWarden`, and no `images` key

### Requirement: Document landmarks
The document SHALL put the header outside and the tabs, the panels and the errors card inside exactly one `<main>` element, and SHALL have exactly one `<h1>`. The export dialog and the live region MAY sit outside `<main>`.

#### Scenario: Fixture document
- **WHEN** the document of the fixture report (`devidence-dev/QuantWarden`, `devidence-dev/homelab`, `devidence-dev/discord-tts-bot`) is rendered
- **THEN** it has one `<main>` that contains the three `role="tabpanel"` sections and the tablist, one `<h1>` with the text `Dependency report - Devidence`, and the `<header>` is not inside `<main>`

#### Scenario: Nothing pending
- **WHEN** the report has 12 docker deps and 30 non-docker deps, none with a listed update
- **THEN** the document still has one `<main>` with the three panels and no export dialog

#### Scenario: Errors card
- **WHEN** `devidence-dev/QuantWarden` has a problem with `level` 50
- **THEN** the errors card is inside the `<main>`, after the panels
