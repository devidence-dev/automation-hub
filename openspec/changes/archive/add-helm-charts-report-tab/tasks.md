# Tasks

## 1. Catalog

- [x] 1.1 Add to `internal/messages/default.yaml` the keys `caption.charts` (` · ⛵ Charts: {charts}`), `document.tab.charts` (`Charts`), `document.eyebrow.charts` (`Chart report`), `document.status.charts_rest` (`across {deps} charts`), `document.status.charts_rest_one` (`across {deps} chart`) and `document.status.charts_up_to_date` (`All charts are up to date`). Add their constants and `allKeys` entries in `keys.go`, and `charts` to `placeholderNames`. Proved by the existing `keys_test.go` catalog checks.
- [x] 1.2 Add `TestCaptionChartsOverride` to `caption_test.go`: an override of `caption.tabs` without `{charts}` loads and renders with the charts text appended, and an override of `caption.charts` without `{charts}` fails naming the key and the placeholder.

## 2. Aggregation

- [x] 2.1 Add `helmDatasource`, `Summary.Charts` and a third builder to `Aggregate`, and pass it to `pendingByRepo`. Pass `addFileDeps` a small `tabBuilders` struct and classify through its `forDatasource` method, with a `shortcut:` comment on the OCI limit (design D2, D3). Proved by `TestAggregateChartsTab` in `aggregate_test.go`: a `helm` dep lands in `Charts` only, a `docker` dep stays in `Images`, a dep without a datasource stays in `Deps`, repo pending counts include charts, and the error carries the repo `devidence-dev/homelab`.

## 3. Document

- [x] 3.1 Replace the `spec.Images` branches of `docBuilder.summary` with `RestKey`, `RestOneKey` and `UpToDateKey` on `tabSpec` (design D4), and build the charts tab third, with `Tabs` holding deps, images and charts. Proved by a charts case in the summary-line test of `document_view_test.go`: `3 updates`, `across 2 charts`, and `All charts are up to date`. Also proved by `TestTabTotalsThreeTabs` in `document_view_test.go` and `TestNothingPendingWithCharts` in `document_test.go`.
- [x] 3.2 Update the comments that say two tabs in `tabs.templ`, `document.go` and `aggregate.go`, and fix the `panes` helper of `document_test.go` so the images pane ends at the charts pane. Proved by `TestDocumentChartRowOnlyInCharts` in `document_test.go`, which also checks the copy value `traefik@43.0.0`, the label `Copy name@version`, the card `homelab`, the file location, `42.3.0` and the `major` chip in the charts pane.
- [x] 3.3 Declare the `narrow` custom variant in `assets/input.css` and use `narrow:flex-col`, `narrow:gap-1` and `narrow:px-1` in place of `max-[420px]:` on `TabButton` in `tabs.templ` (design D8), then run `scripts/generate-depreport-assets.sh`. Proved by `TestTabsMarkup` in `components_test.go` checking the variants on each tab, by `TestNarrowVariantIsMaxWidth` in `styles_test.go`, and by `TestDocumentWeight` staying within its budgets.

## 4. Caption

- [x] 4.1 Count the charts tab in `CaptionPresenter.stats`, appending the `caption.charts` text to the tabs line. Proved by a `TestCaptionStatsThreeTabs` case in `caption_test.go` with the values of the spec scenario: `📊 4 outdated · 8 up to date` and `🐳 Images: 1 · 📚 Packages: 2 · ⛵ Charts: 1`, and the chart-only repo line `🔴 homelab — 1 major` in `caption_test.go`.

## 5. Export

- [x] 5.1 Add `yamlCharts` and pass the charts tab to `buildExport` as its third section. Proved by `TestExportChartsSection` in `export_test.go`, against the exact YAML of the spec scenario "One chart", and a case with only a chart listed that still produces the export button. Also proved by `TestExportDialogThreeSections` in `document_view_test.go`.

## 6. Goldens and docs

- [x] 6.1 Regenerate the goldens with `UPDATE_GOLDEN=1 go test ./internal/handlers/ -run TestGoldenBaseline` in Docker. Review the diff: it may show only the empty third tab and panel, and the caption's charts text. Proved by `TestGoldenBaseline` passing without the variable.
- [x] 6.2 Update `README.md` ("The document has two tabs", "the totals of the two tabs") to describe the "Charts" tab and the OCI limit. Verify by reading the rendered section.

## 7. Other repos and user steps

- [x] 7.1 (homelab) Add `**/*-application.yaml` to `RENOVATE_ARGOCD` in `kubernetes/internal-apps/automation-hub/dependency-report.yaml`. Verified by the diff of that file; its effect on a real report is checked in 7.2.
- [ ] 7.2 (user) After deploying, trigger a report and check the "Charts" tab against the homelab's `targetRevision` values, that `kubernetes/minimal/traefik-application.yaml` is listed, the tab row at 320 CSS pixels, whether a chart lookup failure reaches the errors card, and that `/automation-hub:deps-update` applies an export with one chart update.

## 8. Validation

- [x] 8.1 Run `scripts/verify.sh`, then `docker run --rm -v "$PWD":/app -w /app golang:$(cat .go-version) sh -c 'go vet ./... && go test -race ./...'`, and `graphify update .`. All must pass, and the generated assets must show no drift.
