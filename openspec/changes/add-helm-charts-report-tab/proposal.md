# Proposal

## Why

The report does not show Helm chart updates as their own group. Renovate already finds them through its `argocd` manager, but the hub files every non-`docker` dep under "Dependencies", so chart updates are mixed with Go and npm packages. A homelab chart bump is a different kind of work from a package bump, and it deserves its own view.

## What Changes

- The document gets a third tab, "Charts", after "Dependencies" and "Images". It lists every dep whose Renovate `datasource` is `helm`, with the same badge, panel, repo cards, chips and copy button as the other tabs.
- "Dependencies" no longer lists `helm` deps.
- The caption stats count the three tabs and the tabs line adds `⛵ Charts: <c>`.
- The YAML export gets a `charts` section, after `dependencies`. **BREAKING** for consumers of the export: chart updates move from `dependencies` to `charts`.
- At narrow widths the tab list stacks each label over its badge, so three tabs fit in 320 CSS pixels.
- New catalog keys hold the texts of the tab. `caption.tabs` keeps its placeholders, so existing overrides still load.

## Non-goals

- No change to Renovate's lookup mode: the report stays read-only, with no issues, PRs or branches.
- Charts in OCI registries are not moved to "Charts". Renovate gives them the `docker` datasource, so they stay in "Images". The homelab has none today.
- No handling of `Chart.yaml` dependencies, `helmfile` or Kustomize `helmCharts` beyond what the `helm` datasource already brings.
- No applying of chart updates from the hub.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `dependency-report`: a third tab for Helm charts, with its caption count, export section and narrow layout.

## Impact

- Code: `internal/services/depreport` (aggregation, document builder, caption, export, keys, tabs component, stylesheet) and `internal/messages/default.yaml`. Generated assets and golden files are regenerated.
- No new dependency and no new config setting.
- (homelab) `dependency-report.yaml`: `RENOVATE_ARGOCD` must also match `kubernetes/minimal/*-application.yaml`, whose charts are not scanned today.
- The `deps-update` skill, the only known consumer of the export, handles the `charts` section.
