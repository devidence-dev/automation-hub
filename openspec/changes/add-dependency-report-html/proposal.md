# Proposal

## Why

Two tools report outdated versions: Renovate (dependency report, text messages) and image-container-reporter (icr), which scans running pod images and has the HTML report the user likes. Renovate can cover the images too if it also scans `devidence-dev/homelab`, where all cluster images are declared. One tool, one chat, and the HTML report survives the merge.

## What Changes

- The hub generates ONE HTML document per successful run and sends it to `dependency_report.telegram_chat_id` with `sendDocument`, after the per-repo messages (or after "todo al día"). The Renovate failure path sends no document.
- The document is migrated from icr's template (Bootstrap, dark theme, copy buttons, summary cards) to `html/template`, embedded with `go:embed`, with two tabs: "Imágenes" (Renovate `datasource` `docker`) and "Dependencias" (everything else). Only `major`, `minor` and `patch` updates are listed.
- `telegram.Client` gains `SendDocument`, sharing the retry loop. No new dependency.
- Renovate scans `devidence-dev/homelab`; its `kubernetes` and `argocd` managers get file patterns for the manifests that `helm-values` does not read.
- icr is retired: its workflow, `/run_check_updates` command, `extra-images.yml` mount and host file are removed (cross-repo tasks).

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `dependency-report`: adds the HTML document (content, tabs, escaping, file name, failure path) and extends the Delivery requirement to the document.

## Impact

- Code: `internal/handlers/dependency_report.go` (+ new template and builder), `internal/services/telegram/client.go`, `cmd/automation-hub/main.go` (sender interface), test fixture.
- Docs: `configs/config.yaml.example` and `README.md` drop `run_check_updates`.
- Other repos: homelab (Job env, runner values), workflows, github-runner, Infisical, GitHub (user).
- No new Go dependency; the document needs internet to load Bootstrap from `cdn.jsdelivr.net` when opened.

## Non-goals

- Redesigning the look and feel. The template is migrated as it is; a redesign is a later change.
- Showing `digest` or `pin` updates, or a historical archive of reports.
- Serving the report over HTTP or storing it on disk.
- Changing the text messages, the 4096-character limits or the failure message.
- Scanning pods or nodes in the cluster. Only declared images are covered.
- Replacing Bootstrap CDN links with embedded assets.
