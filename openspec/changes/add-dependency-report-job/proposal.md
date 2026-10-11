# Proposal

## Why

The dependency report PoC runs Renovate inside GitHub Actions in the public `workflows` repo. Its logs and artifacts are public, so it needs leak safeguards. Running the scan as a Kubernetes Job and sending the result through automation-hub keeps the report private, and keeps automation-hub as the one place that decides when ops tasks run. The scan stays out of the hub's process, so a failing scan cannot take down the Telegram bot.

## What Changes

- New `jobs:` bot command type: a command creates a Kubernetes Job from a suspended CronJob template (like `kubectl create job --from=cronjob/...`). It can be triggered from Telegram or by a `schedule` (reusing `add-workflow-scheduler`).
- New `dependency_report:` config (`webhook_token`, `telegram_chat_id`), plus a new `POST /webhook/dependency-report` endpoint that receives the Renovate report from the Job, formats it and sends it to Telegram in the agreed format.
- Telegram client: support for HTML parse mode, with no change to the existing Markdown messages.
- `main.go`: build the bot and the in-cluster kube client when `jobs:` is configured, and register the new route only when `dependency_report:` is configured.
- Command names stay unique across `workflows`, `restarts` and `jobs`.

## Non-goals

- Running Renovate or icr inside the hub process.
- Persisting reports. They are only sent to Telegram.
- Moving icr into the cluster (phase 3). In this repo that is a config-only change that reuses `jobs:`.
- Restricting Job `create` by name in RBAC. Kubernetes does not support it; the risk is accepted and documented in the ChangePlan.
- Retrying failed Telegram sends after the webhook has answered.

## Capabilities

### New Capabilities
- `job-commands`: Telegram and scheduled commands that create a Kubernetes Job from a CronJob template.
- `dependency-report`: authenticated webhook that turns a Renovate report into Telegram messages.

### Modified Capabilities
<!-- none: openspec/specs is empty; scheduling of jobs is specified inside job-commands -->

## Impact

- Depends on `add-workflow-scheduler` (scheduler, dispatch core, `schedule` validation).
- Code: `internal/config`, `cmd/automation-hub/main.go`, `internal/services/kube` (`CreateJobFromCronJob`), `internal/services/telegram` (HTML send), `internal/handlers` (job command and new `dependency_report.go`).
- API: new HTTP endpoint `POST /webhook/dependency-report` (multipart, header `X-Report-Token`).
- Other repos and the user: homelab (ExternalSecret, suspended CronJob, Role/RoleBinding, two mandatory NetworkPolicies), Infisical secrets, and deleting the interim `ops-dependency-report.yml` from `workflows`.
