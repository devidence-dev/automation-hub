# Design

## Context

See `proposal.md` for the motivation and `specs/dependency-report/spec.md` for the behavior.

What Renovate already does (checked against the source of the pinned version, 44.223.1):

- The Job's `RENOVATE_ARGOCD` patterns match `**/application.yaml` and `**/applicationset.yaml`.
- `lib/modules/manager/argocd/extract.ts` reads `spec.source`, `spec.sources` and the ApplicationSet `spec.template.spec`. A source with `chart` and an HTTP(S) `repoURL` becomes a dep with `depName: <chart>` and `datasource: helm`.
- A chart in an OCI registry gets `datasource: docker` (`lib/modules/manager/helmv3/oci.ts`).

So the report that reaches the hub already holds the chart updates of the homelab. `addFileDeps` files them under "Dependencies" because their datasource is not `docker`. Every chart in the homelab uses an HTTPS repo, and the `dependency-report` Job already has egress on 443, so no NetworkPolicy changes.

No real report was inspected for this design, and the test fixture has no `helm` dep.

Limits that shape the change:

- `internal/messages/catalog.go` (`checkPlaceholders`) rejects an override that lacks a placeholder of the default text, and that error stops the hub at startup.
- Compiled CSS is at 21,373 of its 22,000 byte budget. The fixture document is at 107,709 of 130,000 bytes.
- A tab today needs about 144 CSS pixels for "Dependencies" with a two-digit badge. Three tabs do not fit the 264 pixels available at a 320 pixel viewport.

## Goals / Non-Goals

**Goals:**

- Show charts as a third tab, built by the same code path as the other two.
- Leave existing `messages:` overrides valid.

**Non-Goals:**

- The non-goals of the proposal.
- A generic, data-driven list of tabs. Three fixed tabs do not need it.

## Decisions

**D1. No change to Renovate's lookup.** The manager already extracts the charts, so the change is in the hub. The alternative, a custom manager or a separate Helm scan, would duplicate what `argocd` gives for free. The only Renovate-side edit is the file pattern of D7.

**D2. Classify by datasource alone.** `helm` goes to "Charts", `docker` to "Images", anything else to "Dependencies". A `helmDatasource` constant sits next to `dockerDatasource`, since datasource ids are Renovate's protocol values.

Alternative: treat a `docker` dep from the `argocd` manager as a chart, to catch OCI charts. Rejected: that manager also emits `docker` deps for Kustomize images of Git sources, so manager name does not identify a chart, and the homelab has no OCI chart. The limit gets a `shortcut:` comment at the classification.

**D3. A third `tabBuilder` and `Summary.Charts`.** `addFileDeps` receives a small `tabBuilders` struct with a `forDatasource` method, instead of three positional arguments of the same type. The classification then has one owner, and the arguments cannot be swapped. `pendingByRepo` is already variadic, so repo lines count charts with one more argument.

**D4. `tabSpec` holds the catalog keys of its status lines.** `docBuilder.summary` branches on `spec.Images` to choose its texts. A third branch would repeat that, so `tabSpec` gets `RestKey`, `RestOneKey` and `UpToDateKey`, and `summary` reads them. `spec.Images` stays only for the copy format, where charts take the `name@version` path of a dependency.

**D5. Separate `caption.charts` key.** Adding `{charts}` to `caption.tabs` would make any existing override of that key fail at startup (see Context). The new key is appended to the rendered line. Its default starts with a space, like `fork.tag_caption`.

**D6. Export gets a `charts` key, last.** Each tab maps to one export section, so the dialog logic in `report.js` needs no change. `buildExport` takes the third tab. `images` and `dependencies` keep their position.

**D7. (homelab) Scan the minimal Applications.** `kubernetes/minimal/traefik-application.yaml` and `adguardhome-application.yaml` do not match `**/application.yaml`, so their charts are unknown to Renovate. Their traefik is at `41.4.0`, against `42.3.0` in the main Application. `RENOVATE_ARGOCD` gets `**/*-application.yaml`.

**D8. Stack label over badge up to 420 pixels.** One reusable Tailwind variant `narrow` is declared in `assets/input.css` as `@custom-variant narrow (@media (max-width: 420px))`. `TabButton` uses it as `narrow:flex-col`, `narrow:gap-1` and `narrow:px-1`. The arbitrary variant `max-[420px]:` is not used: Tailwind compiles it to `not all and (min-width:420px)`, which excludes exactly 420px and adds a second `min-width` to the compiled CSS. That goes against the "up to 420" scenario and the Responsive layout rule that 601px is the only `min-width`. The named variant is inclusive and leaves that rule true. The pill reads `offsetLeft` and `offsetWidth` and its height comes from `top` and `bottom`, so `movePill` needs no change.

Alternatives: a scrolling tab list hides a tab at the width where the user most needs to see it; shorter labels would need overrides of existing keys; hiding the badge breaks the requirement that each tab has one.

**D9. Keep the fixture, regenerate the goldens.** `renovate-report.json` stays as it is, so the 54-row weight test holds. New behavior is tested with small reports through the existing `reportOf` helper. The goldens change only by the empty third tab and the caption line, and the diff is reviewed after `UPDATE_GOLDEN=1`. The `panes` helper in `document_test.go` must cut the images pane at the charts pane.

## Risks / Trade-offs

- [The empty third tab adds bytes to every document] → about 1 to 2 KB, within budget; `TestDocumentWeight` guards it.
- [CSS budget has 627 bytes left] → D8 adds only a few variants. If the budget test fails, raise the limit as a deliberate spec change instead of removing rules.
- [Export consumers lose chart updates from `dependencies`] → the `deps-update` skill is the only known consumer, and it handles the `charts` section.
- [A chart repo that fails lookup may count as up to date] → the errors card only shows problems of level 50 or more, and the level Renovate uses for this failure was not verified. Check it on the first real run; widening the error filter is out of scope.
- [The first real report may differ from the design] → the design rests on Renovate's source, not on a real report. The final (user) task checks the deployed report.

## Migration Plan

1. Deploy the hub. The third tab fills on the next report, with no config change.
2. (homelab) Add the file pattern of D7. It works with or without step 1.

Rollback: redeploy the previous image. No state or config is involved.
