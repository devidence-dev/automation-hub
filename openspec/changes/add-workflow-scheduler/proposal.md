# Proposal

## Why

The scheduled GitHub Actions runs that automation-hub's commands mirror (`ops-github-runner-check-updates.yml`, `ops-github-runner-cleanup.yml`) start 4 to 8 hours late, because GitHub's `schedule` event is best effort. A `workflow_dispatch` sent through the API starts within seconds. Also, the pod sets `TZ=America/Guayaquil`, but the Alpine runtime image has no tzdata, so Go silently falls back to UTC.

## What Changes

- Embed Go's timezone database in the binary. At startup, load `America/Guayaquil` explicitly and exit with an error if it fails.
- Add an optional `schedule` field (5-field cron) to each `workflows:` entry. An invalid expression makes startup fail before any goroutine starts.
- Add an in-process scheduler. It dispatches each scheduled workflow at the configured time in `America/Guayaquil`, through the same dispatch path the Telegram command uses.
- At startup, log the timezone and the next run of each scheduled entry. Log every scheduled dispatch. If a scheduled dispatch fails, notify each of that entry's `allowed_chat_ids`.
- Dispatch with `return_run_details: true`, so GitHub returns the exact run created. Remove the `FindLatestRunURL` polling, which can pick the wrong run when dispatches overlap. `google/go-github` was not adopted (see design).
- Document `schedule` in `configs/config.yaml.example` and the README. Remove the uncommitted `run_dependency_report` workflow entry, which `add-dependency-report-job` replaces.
- New dependency: `github.com/robfig/cron/v3` (needs user approval at apply time; pin a version checked against its source).

## Non-goals

- No catch-up: a run missed while the pod is down is skipped.
- No per-entry timezone. The zone is fixed to `America/Guayaquil`.
- No Kubernetes Job commands or dependency report (`add-dependency-report-job`).
- Removing the `schedule:` triggers and `should-run` gates from the `workflows` repo is listed as a cross-repo task, not code in this repo.
## Capabilities

### New Capabilities
- `scheduled-dispatch`: time-based dispatch of configured bot commands in a fixed local timezone, plus startup validation and failure notification.

### Modified Capabilities
<!-- none: openspec/specs is empty -->

## Impact

- Code: `cmd/automation-hub/main.go` (tzdata import, zone load, scheduler start and stop), `internal/config` (`Schedule` field and validation), `internal/handlers/telegram_bot.go` (extract a chat-independent dispatch core, drop `findRunURL`), `internal/services/github/client.go` (`DispatchWorkflow` returns the run URL; `FindLatestRunURL` removed), new `internal/services/scheduler`.
- Dependency: `github.com/robfig/cron/v3`. The embedded tzdata adds a few hundred KB to the binary.
- Config: `CONFIG_YAML` in Infisical gains `schedule` on `run_check_updates` and `run_cleanup`. Viper ignores unknown keys, so the old image tolerates it.
- Other repos: `workflows` (remove `schedule:` and `should-run` from both host-ops workflows, README).
