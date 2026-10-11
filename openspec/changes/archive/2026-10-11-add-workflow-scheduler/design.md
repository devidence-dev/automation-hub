# Design

## Context

- `main.go` builds the workflow bot, `githubClient` and `BotHandler` only inside `if len(cfg.Workflows) > 0 || len(cfg.Restarts) > 0`. `kube.NewInClusterClient` is created only when `restarts:` is set.
- `BotHandler.dispatchWorkflow(commandName, chatID, workflow)` mixes three things: the dispatch, the run-URL lookup (`findRunURL`, which polls `FindLatestRunURL` up to `runLookupAttempts` times and takes the first `workflow_dispatch` run created after the dispatch, so it can return the wrong run when two dispatches overlap) and the Telegram reply. The `telegramMessenger` sends through the workflow bot.
- `validateBotCommands` runs inside `config.Load()` and is the existing place where config errors stop startup.
- The runtime image is `alpine:3.24.2` without tzdata, and the pod sets `TZ=America/Guayaquil`.

## Goals / Non-Goals

**Goals:**
- Telegram commands and schedules share one dispatch function with no chat dependency.
- Invalid schedules fail at config load, which the tests can cover.

**Non-Goals:**
- A generic job framework. The scheduler only runs funcs that are registered at startup.

## Decisions

- **`import _ "time/tzdata"` + `time.LoadLocation("America/Guayaquil")` in `main.go`.** This makes the binary independent of the image. Alternative: `apk add tzdata` in the Dockerfile. It was rejected because a later base-image change could silently bring back the UTC fallback.
- **`github.com/robfig/cron/v3` with `cron.WithLocation(loc)` and a standard 5-field parser.** Day-of-week lists and DST-safe local time are already solved there. Alternative: a hand-rolled `time.Ticker` loop with our own parsing. It was rejected because that means more code and more edge cases. The library is a new dependency, so the user approves it and its exact version is checked against the module proxy at apply time. Before use, confirm `WithLocation`, `WithParser` and `ParseStandard` exist in that version's source.
- **Validation lives in `internal/config`.** It parses each non-empty `schedule` with the same parser config the scheduler uses, next to `validateBotCommands`. The error includes the command name. The parser is exposed as a small function so config and scheduler cannot drift apart.
- **`return_run_details: true` on `workflow_dispatch`, with no run polling.** Since the GitHub changelog of 2026-02-19, the dispatch endpoint accepts this optional flag and answers `200` with `{workflow_run_id, run_url, html_url}` (REST docs, API version 2022-11-28). Without the flag it still answers `204`. `DispatchWorkflow` sends the flag and returns `html_url`. On `204` it returns an empty URL and no error, and the caller falls back to `https://github.com/<owner>/<repo>/actions`. `FindLatestRunURL`, `findRunURL`, `runLookupAttempts` and `runLookupDelay` are deleted. Alternative: `google/go-github`. It was rejected because the client covers 2 endpoints in ~120 lines with httptest coverage, go-github's frequent major versions (`/vNN` import path) mean bump churn, and it would not fix the overlapping-dispatch race, which comes from the API pattern and not from the HTTP code.
- **Extract a dispatch core from `BotHandler`.** Proposed: `RunWorkflow(ctx, workflow) (runURL string, err error)`, which dispatches and resolves the URL with the fallback above. The Telegram path calls it and formats the reply. The scheduler calls it and logs. Failure notifications reuse `sendMessage` for each `allowed_chat_id`.
- **New package `internal/services/scheduler`.** It wraps cron with `Add(name, spec, fn)`, `Start()` and `Stop()`, and logs the next run of each entry at start. `main.go` builds it only when at least one entry has `schedule`. `Stop()` runs on the existing SIGINT/SIGTERM path: it blocks new runs and waits for a running job to finish. Jobs use `context.Background()`, so a dispatch in flight during shutdown is bounded by the GitHub client's 30 s HTTP timeout, not cancelled.
- **Wiring stays inside the existing bot block.** Scheduled entries are `workflows:` entries, so the block that builds `githubClient` already runs whenever a schedule exists. No client construction is duplicated.

## Risks / Trade-offs

- [Invalid `CONFIG_YAML` takes down the whole hub, including IMAP] → This is intended fail-fast behavior. Tests cover parsing, and `configs/config.yaml.example` documents the format.
- [A run is lost if the pod restarts in the scheduled minute] → Accepted. No catch-up.
- [GitHub Enterprise Server older than 3.21 rejects `return_run_details` with 400] → Not applicable: the hub only targets github.com. If that changes, drop the flag and accept the Actions-page link.
- [Window with nothing scheduled while the `workflows` repo drops its `schedule:`] → Deploy the hub first. Viper ignores unknown keys, so `schedule` can go into `CONFIG_YAML` before the new image lands.

## Migration Plan

1. Paste `schedule` into `CONFIG_YAML` (user). The old image ignores it.
2. Deploy the new image. Check the startup log for the zone and the next runs.
3. Remove the `schedule:` triggers and `should-run` gates from the `workflows` repo.
4. Rollback: revert to the previous image and revert the `workflows` commit.
