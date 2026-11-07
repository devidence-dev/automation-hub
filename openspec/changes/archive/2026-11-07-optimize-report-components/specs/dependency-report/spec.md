## ADDED Requirements

### Requirement: Document landmarks
The document SHALL put the header outside and the tabs, the panels and the errors card inside exactly one `<main>` element, and SHALL have exactly one `<h1>`. The export dialog and the live region MAY sit outside `<main>`.

#### Scenario: Fixture document
- **WHEN** the document of the fixture report (`devidence-dev/QuantWarden`, `devidence-dev/homelab`, `devidence-dev/discord-tts-bot`) is rendered
- **THEN** it has one `<main>` that contains both `role="tabpanel"` sections and the tablist, one `<h1>` with the text `Dependency report - Devidence`, and the `<header>` is not inside `<main>`

#### Scenario: Nothing pending
- **WHEN** the report has 12 docker deps and 30 non-docker deps, none with a listed update
- **THEN** the document still has one `<main>` with both panels and no export dialog

#### Scenario: Errors card
- **WHEN** `devidence-dev/QuantWarden` has a problem with `level` 50
- **THEN** the errors card is inside the `<main>`, after the panels

### Requirement: Named tablist
The tablist SHALL have an accessible name from the catalog key `document.tabs.label`, whose English default is `Report sections`. `messages:` in the config MAY override it.

#### Scenario: Default name
- **WHEN** the fixture document is rendered with no `messages:` section
- **THEN** the element with `role="tablist"` has `aria-label="Report sections"`

#### Scenario: Overridden name
- **WHEN** the config has `messages: {document: {tabs: {label: "Secciones del informe"}}}`
- **THEN** the tablist has `aria-label="Secciones del informe"` and no other label of the document changes

#### Scenario: Escaped name
- **WHEN** the override is `Sections "A" & <B>`
- **THEN** the attribute holds `Sections &#34;A&#34; &amp; &lt;B&gt;` and no attribute is added to the tablist

### Requirement: Entry lists
The entries of one package file SHALL be the items of one list, so assistive technology announces their count. Each entry SHALL be one list item, and the list SHALL keep `role="list"` because its bullets are hidden.

#### Scenario: File with three entries
- **WHEN** `go.mod` of `devidence-dev/QuantWarden` lists `golang.org/x/net`, `github.com/gorilla/mux` and `go.uber.org/zap`
- **THEN** the card has one `<ul role="list">` for `go.mod` with three `<li>` items, each with one name, one type chip and one copy button

#### Scenario: Two files
- **WHEN** a repo has `go.mod` and `Dockerfile` with listed updates in different tabs
- **THEN** each file group has its own list, and the file location sits after the list, not inside it

#### Scenario: Entry text unchanged
- **WHEN** the entry `golang.org/x/net` from `0.30.0` to `1.0.0` is rendered
- **THEN** it keeps the hidden labels `Current` and `New` for screen readers and the same visible text as before

### Requirement: Live regions
The document SHALL have one polite live region outside the export dialog and one inside it. Each SHALL be an element with `role="status"` and no separate `aria-live` attribute, because that role is already polite.

#### Scenario: Fixture document
- **WHEN** the fixture document is rendered
- **THEN** it has exactly two elements with `role="status"`, one of them inside the `<dialog>`, and none carries `aria-live`

#### Scenario: Copy announcement
- **WHEN** the user copies `golang.org/x/net@1.0.0`
- **THEN** `Copied!` is written to the region outside the dialog, and when the export dialog is open the copy of the YAML writes `Copied!` to the region inside it

#### Scenario: No updates listed
- **WHEN** nothing is listed
- **THEN** the document has one `role="status"` element, outside any dialog, because there is no export dialog

### Requirement: Visible keyboard focus
Every button of the document, the repo card summary rows and the checkboxes of the export dialog SHALL show a focus indicator when focused with the keyboard: a 2px outline in the `--accent` color, 2px away from the element. A mouse click SHALL NOT show it.

#### Scenario: Tab button
- **WHEN** the user moves the focus to the "Images" tab with the keyboard
- **THEN** the tab shows a 2px `--accent` outline offset by 2px

#### Scenario: Theme and copy buttons
- **WHEN** the user focuses the `Dark` theme button and then a copy button with the keyboard
- **THEN** each shows the same 2px `--accent` outline offset by 2px

#### Scenario: Mouse click
- **WHEN** the user clicks the `Copy name@version` button with a mouse
- **THEN** the button shows no outline

#### Scenario: Disabled export action
- **WHEN** no repo is checked in the export dialog
- **THEN** the `Copy YAML` and `Download YAML` buttons are disabled, cannot take focus, and show the default cursor and no hover surface

### Requirement: Native controls follow the theme
The document SHALL set `color-scheme` to the active scheme, so that scrollbars, checkboxes and the dialog backdrop match the theme: `light` when Light is chosen, `dark` when Dark is chosen, and the system scheme (`light dark`) on Auto.

#### Scenario: Light on a dark system
- **WHEN** the system prefers a dark color scheme and the user presses `Light`
- **THEN** `<html>` has `data-theme="light"` and its computed `color-scheme` is `light`

#### Scenario: Dark on a light system
- **WHEN** the system prefers a light color scheme and the user presses `Dark`
- **THEN** its computed `color-scheme` is `dark`

#### Scenario: Auto
- **WHEN** the document is opened and the user has not pressed a theme button
- **THEN** `<html>` has no `data-theme` and its computed `color-scheme` is `light dark`

#### Scenario: Compiled CSS
- **WHEN** the compiled CSS is read
- **THEN** it declares `color-scheme` for the default `:root`, for `:root[data-theme=light]` and for `:root[data-theme=dark]`

### Requirement: Document weight
The fixture document SHALL NOT exceed 130,000 bytes, its compiled CSS SHALL NOT exceed 22,000 bytes, and an entry row SHALL NOT average more than 1,000 bytes of markup. These budgets are checked by a test, so growth is a deliberate change.

#### Scenario: Fixture within budget
- **WHEN** the fixture report (54 entries across 5 repo cards) is rendered with the default catalog
- **THEN** the document is at most 130,000 bytes, the stylesheet at most 22,000 bytes, and the markup of its 54 entry rows at most 54,000 bytes

#### Scenario: Nothing pending is smaller
- **WHEN** the report has 12 docker deps and 30 non-docker deps, none with a listed update
- **THEN** the document is at most 60,000 bytes and has no entry rows

#### Scenario: Budget exceeded
- **WHEN** a change adds markup that takes the fixture document over 130,000 bytes
- **THEN** the budget test fails and names the measured size and the budget

## MODIFIED Requirements

### Requirement: Document theme
The document SHALL use three color schemes from one set of tokens: light by default, dark when the system prefers dark and Auto is selected, and dark when Dark is selected for the page view. The tokens SHALL be declared on `:root`, redefined under `@media (prefers-color-scheme: dark)` for `:root:not([data-theme=light])`, and redefined for `:root[data-theme=dark]`. Every text color SHALL reach a contrast of at least 4.5:1 against its background in all three schemes.

#### Scenario: Light tokens
- **WHEN** the compiled CSS is read
- **THEN** the `:root` block defines the background as `#EDF3F8`, the glass as `rgba(255,255,255,.55)` and the ink as `#0B1B25`

#### Scenario: Dark tokens in both dark blocks
- **WHEN** the compiled CSS is read
- **THEN** the media block for `:root:not([data-theme=light])` and the block for `:root[data-theme=dark]` each define the background as `#04101A` and the ink as `#EAF2F8`, and the rendered document has no `dark:` variant class

#### Scenario: Text contrast
- **WHEN** each text token of the light and dark blocks is read, including the secondary text of the file folder (`--muted-2`)
- **THEN** its contrast against the page background and against the card is at least 4.5:1, with `--muted-2` at `#5F6F7C` in light and `#7D94A4` in dark

#### Scenario: Selected tab count
- **WHEN** the "Dependencies" tab is selected
- **THEN** its count badge shows white `#FFFFFF` text on `#0A63C2` in light and in dark, a contrast of 5.87:1

#### Scenario: Chip text on its tint
- **WHEN** the chips of the types `major`, `minor` and `patch` and of the errors card are read in the light and the dark scheme
- **THEN** each text color against its tinted background, blended over the card, is at least 4.5:1

#### Scenario: Chip colors of one type
- **WHEN** the `major` chip is read in the light block
- **THEN** its text is `#B93A24` on the tinted background of the mock, and in the dark block its text is `#FFA391`
