# Tasks

Every task that edits a `.templ` file or `assets/input.css` regenerates the assets with `sh scripts/generate-depreport-assets.sh` before running its test.

## 1. Header layout and stamp

- [x] 1.1 Add `ScanDateTime` to `HeaderView`, set from `ReceivedAt.In(loc).Format(time.RFC3339)` in `document.go`, and render the stamp text inside `<time datetime>` with a `title` (design D2). Proof: a new table test `TestHeaderTimeElement` in `components_test.go` covers `America/Guayaquil` (`2026-10-12T09:07:00-05:00`) and `Asia/Kolkata` (`2026-10-12T19:37:00+05:30`), and the visible text is unchanged.
- [x] 1.2 Rebuild `Header` as a grid (design D1): DOM order logo, theme control, meta (stamp, then export); `grid-cols-[1fr_auto]` with areas `logo theme` / `meta meta` up to 600px, and one row `logo meta theme` from 601px; the stamp gets `min-w-0 flex-1 truncate` and the export button `shrink-0`. Proof: new `TestHeaderOrderAndAreas` in `components_test.go` checks the DOM order and the grid classes, and `TestReportCSSHasNoOldFrameworkOrWideMinWidth` and `class_coverage_test.go` still pass.
- [x] 1.3 Show the export button label at every width: remove `sr-only min-[601px]:not-sr-only` from its span (design D3). Proof: new `TestExportButtonLabelVisible` in `components_test.go` checks that the span has no `sr-only`, and the `aria-label` is still `Export`.
- [x] 1.4 (user) Browser check of the golden document, with and without updates: at 320 and 440 CSS px the theme control sits on the logo row and the stamp and Export share the second row (the stamp ends in an ellipsis at 320 px); at 640 and 760 CSS px logo, stamp, Export and theme control sit on one row with no overlap (at 640 the stamp may end in an ellipsis); the page never scrolls horizontally. Proof: user reply.

## 2. Export action icons

- [x] 2.1 Add the `download` and `close` icons to `iconMarkup` in `icon.go` on the 14px grid (design D6). Proof: `TestIconsAreFixedAndInert` lists and checks both new names.
- [x] 2.2 Pass the icons in the export actions: `copy` and `check` with the `group-data-[state=copied]` swap on `Copy YAML`, `download` on `Download YAML`, `close` on `Close`, and add `inline-flex items-center gap-1.5` to `exportAction`. Proof: new `TestExportActionIcons` in `components_test.go` checks one `aria-hidden` icon before each label, and that the accessible names are unchanged.
- [x] 2.3 (user) In a browser, open the export dialog: the icons align with the labels in light and dark, `Copy YAML` shows the check and `Copied!` for 2 s, disabled actions keep their icons, and with clipboard access blocked (e.g. DevTools > site settings, clipboard denied) and the fallback copy failing, `Copy YAML` keeps its copy icon and label and nothing is announced. Proof: user reply.

## 3. Card height motion

- [x] 3.1 Add the `::details-content` height transition with `interpolate-size: allow-keywords`, inside `@supports (interpolate-size: allow-keywords)` and the `prefers-reduced-motion: no-preference` block (design D4). Proof: new `TestReportCSSDetailsMotion` in `styles_test.go` checks the `@supports` guard, the 300 ms height transition with `content-visibility ... allow-discrete`, and that no details transition exists outside the `no-preference` block.
- [x] 3.2 (user) Open and close `QuantWarden` in Chrome (animates) and in Firefox or Safari (instant, content complete), and with reduced motion on (instant). Proof: user reply.

## 4. Dialog motion

- [x] 4.1 Add the 200 ms open and close transition to `.export-dialog` and its `::backdrop` (`opacity`, scale `.96` to `1`, `overlay` and `display` with `allow-discrete`, entry state in `@starting-style`), inside the `no-preference` block (design D5). Proof: new `TestReportCSSDialogMotion` in `styles_test.go` checks the `@starting-style` rule, the transitioned properties and the reduced-motion guard; `TestReportCSSExportDialogCentered` still passes.
- [x] 4.2 (user) Open the dialog with `Export` and close it with `Close` and with Escape: it fades and scales in and out, the backdrop fades, focus enters the dialog and returns to `Export`, and the copy still works inside the dialog. With reduced motion on, it opens and closes at once. Proof: user reply.

## 5. Theme pill during the theme change

- [x] 5.1 Give the theme control pill `view-transition-name: theme-pill` and its group animation the `--ease` timing (design D7). Proof: new `TestReportCSSThemePillTransition` in `styles_test.go` checks the name on the theme pill only, and that reduced motion turns the view transition animations off.
- [x] 5.2 In `report.js` `chooseTheme`, move the pill inside the view transition update callback with `no-anim` on its group, and keep today's `movePill` path without view transitions or under reduced motion. Proof: `TestReportScriptContract` in `js_contract_test.go` is extended for the new order, and `gate biome` passes.
- [x] 5.3 (user) In Chrome, press Auto, Light and Dark: the pill slides to the pressed button while the colors crossfade. In a browser without view transitions, the pill still slides. With reduced motion, it moves at once. Proof: user reply.

## 6. Integration

- [x] 6.1 Regenerate the assets and the golden with `UPDATE_GOLDEN=1 go test ./internal/handlers/...`, review its diff (header grid, `<time>`, label, icons), and check that `TestDocumentWeight` passes within the current budgets (CSS ≤ 22,000 bytes, document ≤ 130,000 bytes). Proof: the golden diff shows only these changes, and the budget test passes.
- [x] 6.2 Run `docker run --rm -v "$PWD":/app -w /app golang:$(cat .go-version) sh -c 'go vet ./... && go test -race ./...'` and `scripts/verify.sh` (with the generated files staged). Proof: both pass.

## 7. Scroll lock behind the dialog

- [x] 7.1 Add `html:has(.export-dialog[open]) { overflow: hidden; scrollbar-gutter: stable; }` and `overscroll-behavior: contain` on `.export-dialog` in `assets/input.css`, outside the reduced-motion blocks (design D8). Proof: new `TestReportCSSDialogScrollLock` in `styles_test.go` checks both rules in the compiled CSS, and the golden is regenerated.
- [x] 7.2 (user) Scroll the golden document 400px down, open `Export`: wheel, touch and arrow keys over the backdrop do not move the page, the dialog content still scrolls, the page does not shift sideways, and after `Close` or Escape the page is at the same position and scrolls again. Proof: user reply.

## 8. Corner close

- [x] 8.1 Add the corner close button to `ExportDialog` and rename the `copy-btn` utility to `icon-btn` (design D9). Proof: new `TestExportDialogCornerClose` in `components_test.go` passes, and the renamed assertions in `styles_test.go` and `components_test.go` pass.
- [x] 8.2 (user) Open `Export`: the X sits at the top right beside the title in light and dark, shows the focus ring by keyboard, closes the dialog with focus back on `Export`, and initial focus on open is acceptable. Proof: user reply.

## Workflow follow-up

- Verify the change with `/opsx:verify`, then archive it after the user checks 1.4, 2.3, 3.2, 4.2, 5.3, 7.2 and 8.2.
