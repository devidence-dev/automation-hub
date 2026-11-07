## ADDED Requirements

### Requirement: Header layout by viewport width
Up to 600px the header SHALL have two rows: the logo and the theme control in the first, and the scan stamp and the export button in the second, spanning the full header width. The stamp SHALL take the remaining width and end in an ellipsis when its text does not fit, and the export button SHALL keep its full size. From 601px the header SHALL be one row: the logo, then the stamp, the export button and the theme control.

#### Scenario: Phone width
- **WHEN** the fixture document is shown in a 440 CSS pixel wide viewport
- **THEN** the theme control sits on the same row as the logo, at its right edge, and the stamp `2026-10-12 09:07 America/Guayaquil` and the `Export` button share the row below, with no horizontal page scrolling

#### Scenario: Narrowest width
- **WHEN** the fixture document is shown in a 320 CSS pixel wide viewport
- **THEN** the stamp text ends in an ellipsis, the `Export` button shows its icon and label in full, and the page does not scroll horizontally

#### Scenario: Wide viewport
- **WHEN** the fixture document is shown in a 760 CSS pixel wide viewport
- **THEN** the logo, the stamp, the `Export` button and the theme control sit on one row

#### Scenario: Nothing pending
- **WHEN** the report has 12 docker deps and 30 non-docker deps, none with a listed update, and is shown 440 CSS pixels wide
- **THEN** the header has no export button, the logo and the theme control share the first row, and the stamp sits alone on the second

### Requirement: Machine-readable scan time
The scan stamp SHALL be a `<time>` element whose `datetime` attribute is the scan instant in RFC 3339 form, with the offset of the configured timezone. The visible text SHALL stay as today.

#### Scenario: Fixture scan time
- **WHEN** the report received at `2026-10-12T14:07:00Z` is rendered with the timezone `America/Guayaquil`
- **THEN** the stamp is `<time datetime="2026-10-12T09:07:00-05:00">` with the text `2026-10-12 09:07 America/Guayaquil`

#### Scenario: Timezone with half-hour offset
- **WHEN** the same report is rendered with the timezone `Asia/Kolkata`
- **THEN** the `datetime` attribute is `2026-10-12T19:37:00+05:30`

### Requirement: Export action icons
Each export dialog action SHALL show a decorative icon before its label: the copy icon for `Copy YAML`, a download icon for `Download YAML` and a close icon for `Close`. After a successful copy, `Copy YAML` SHALL show the check icon with its copied text for as long as the entry copy buttons show it, then return to the copy icon. The icons SHALL carry no text and be hidden from assistive technology.

#### Scenario: Icons in the dialog
- **WHEN** the export dialog of the fixture document is rendered
- **THEN** `Copy YAML`, `Download YAML` and `Close` each show one visible `aria-hidden="true"` icon before the label (`Copy YAML` holds the copy and the check icon and shows one at a time), and their accessible names are still `Copy YAML`, `Download YAML` and `Close`

#### Scenario: Copied state
- **WHEN** the user activates `Copy YAML` and the copy succeeds
- **THEN** the button shows the check icon and `Copied!` for 2 seconds, then the copy icon and `Copy YAML` again

#### Scenario: Copy fails
- **WHEN** the clipboard API rejects the write and the fallback copy also fails
- **THEN** the button keeps the copy icon and `Copy YAML`, and nothing is announced

#### Scenario: Disabled actions keep their icons
- **WHEN** no repo is checked
- **THEN** `Copy YAML` and `Download YAML` are disabled and still show their icons

### Requirement: Page scroll locked behind the export dialog
While the export dialog is open, the page behind it SHALL NOT scroll by wheel, touch or keyboard, and the page SHALL NOT shift sideways when the lock starts or ends. Content of the dialog that is taller than the viewport SHALL still scroll inside the dialog. Closing the dialog SHALL leave the page at the scroll position it had when the dialog opened.

#### Scenario: Wheel over the backdrop
- **WHEN** the fixture document is scrolled 400px down, the user opens `Export` and turns the mouse wheel over the backdrop
- **THEN** the page stays at 400px

#### Scenario: Scroll reaches the end of the dialog
- **WHEN** the dialog content is taller than a 640 CSS pixel tall viewport and the user keeps scrolling past its end
- **THEN** the dialog scrolls to its end and the page does not move

#### Scenario: Close restores
- **WHEN** the user closes the dialog with `Close` or Escape
- **THEN** the page is still at 400px, scrolls again, and its content did not shift sideways

#### Scenario: Compiled CSS
- **WHEN** the compiled CSS is read
- **THEN** it hides the root overflow and keeps a stable scrollbar gutter while an export dialog is open, and the dialog contains its overscroll

### Requirement: Export dialog corner close
The export dialog SHALL show an icon-only close button at the top right, beside its title, named and titled by `document.export.close`. Activating it SHALL close the dialog like the `Close` action and return focus to the `Export` button.

#### Scenario: Corner close in the fixture
- **WHEN** the export dialog of the fixture document is rendered
- **THEN** it has one icon-only button named `Close` before the first section, and its close icon is hidden from assistive technology

#### Scenario: Close from the corner
- **WHEN** the user opens `Export` and activates the corner button
- **THEN** the dialog closes, focus returns to `Export`, and the page scrolls again

#### Scenario: Overridden label
- **WHEN** the config sets `messages: {document: {export: {close: "Cerrar"}}}`
- **THEN** the corner button and the bottom action are both named `Cerrar`

## MODIFIED Requirements

### Requirement: Export button
The header SHALL show an export button, labeled by `document.export.button`, when at least one update is listed in either tab. It SHALL open the export dialog as a modal. The button SHALL show its icon and its label at every viewport width. When no update is listed, the document SHALL have no export button and no export dialog.

#### Scenario: Updates listed
- **WHEN** the document of a report where `devidence-dev/homelab` has the image `ghcr.io/immich-app/immich-server` from `v1.118.0` to `v1.119.0` is rendered
- **THEN** the header has one button named `Export`, and activating it opens the export dialog as a modal with focus inside it

#### Scenario: Nothing listed
- **WHEN** the report has 12 docker deps and 30 non-docker deps, none with a `major`, `minor` or `patch` update
- **THEN** the document has no `Export` button and no export dialog

#### Scenario: Phone width
- **WHEN** the document is shown in a 320 CSS pixel wide viewport
- **THEN** the header shows the export button with its icon and the visible label `Export`, with no horizontal page scrolling

### Requirement: Motion
The document SHALL animate the pill of the selected tab and of the theme control, the panel that appears, the chevron of an open card, the height of a card as it opens and closes, the export dialog and its backdrop as they open and close, and the theme change, during which the theme pill SHALL slide to the pressed button. All of these SHALL be off under `prefers-reduced-motion: reduce`. A browser without the needed CSS features SHALL show the instant change.

#### Scenario: Reduced motion
- **WHEN** the user prefers reduced motion
- **THEN** the pills move without transition, the panel appears without animation, the chevron turns without transition, a card opens and closes at once, the export dialog appears and closes at once, and the theme changes without a view transition

#### Scenario: Panel enters
- **WHEN** the user shows the other tab without reduced motion
- **THEN** the panel animates from 24px to the side in its direction of travel, with opacity from 0 to 1 and blur from 4px to 0, over 420 ms

#### Scenario: Card opens and closes
- **WHEN** the user opens the `QuantWarden` card and then closes it, without reduced motion
- **THEN** its content grows from height 0 to its full height, and shrinks back to 0 before it is hidden, each over 300 ms

#### Scenario: Card in a browser without height interpolation
- **WHEN** the browser does not support animating to an automatic height
- **THEN** the `QuantWarden` card opens and closes at once, and its content is complete and usable

#### Scenario: Dialog opens and closes
- **WHEN** the user activates `Export` and later `Close`, without reduced motion
- **THEN** the dialog fades and scales from 96% to 100% on open and back on close, its backdrop fades in and out, each over 200 ms, and focus still moves into the dialog and back to `Export`

#### Scenario: Theme pill during the theme change
- **WHEN** the theme is Auto and the user presses `Dark`, without reduced motion
- **THEN** the theme pill slides from `Auto` to `Dark` while the page changes theme, over the 450 ms of the theme change, and ends under `Dark`

#### Scenario: Theme pill without view transitions
- **WHEN** the browser has no view transitions and the user presses `Light`
- **THEN** the theme changes at once and the pill slides to `Light` with its own 450 ms transition
