# Tasks

## 1. Dependency and timezone

- [x] 1.1 (user) Approve `github.com/robfig/cron/v3`. Check the latest stable version against the module proxy and the source (`WithLocation`, `WithParser`, `ParseStandard`), then add it with `go get` in Docker. Verify that `go.mod` pins that exact version.
- [x] 1.2 Add `import _ "time/tzdata"` and `time.LoadLocation("America/Guayaquil")` in `cmd/automation-hub/main.go`, calling `logger.Fatal` before any goroutine starts if loading fails. Verify by building the image and checking the startup log in a container without `/usr/share/zoneinfo`.

## 2. Config

- [x] 2.1 Add `Schedule string` (`mapstructure:"schedule"`) to `WorkflowCommandConfig`, plus a shared cron parser helper. Validate non-empty schedules in `config.Load()` next to `validateBotCommands`. Verify with a table test in `internal/config/config_test.go`: `0 9 * * 1,3,5` accepted, `0 9 * *` rejected with the command name in the error, empty accepted.
- [x] 2.2 Document `schedule` in `configs/config.yaml.example` and the README, and remove the uncommitted `run_dependency_report` workflow entry from both. Verify that the example still loads in `TestLoad*` (or an equivalent config test).

## 3. Dispatch core

- [x] 3.1 Change `github.Client.DispatchWorkflow` to send `{"ref": ..., "return_run_details": true}` and return `(htmlURL string, err error)`: 200 decodes `html_url`, 204 returns `""` with no error, and other statuses return an error as today. Delete `FindLatestRunURL`, `findRunURL`, `runLookupAttempts` and `runLookupDelay`, plus their tests. Verify with `httptest` cases in `client_test.go`: the request body contains `return_run_details: true`, 200 returns `html_url`, 204 returns `""`, and 422 returns an error.
- [x] 3.2 Extract a chat-independent `RunWorkflow(ctx, workflow) (string, error)` from `BotHandler.dispatchWorkflow`, falling back to `https://github.com/<owner>/<repo>/actions` when the URL is empty, and keep the Telegram reply text unchanged. Verify with bot handler tests updated to the new `workflowDispatcher` interface (no `FindLatestRunURL`): the reply contains the `html_url` from the fake, the fallback URL appears when the fake returns `""`, and a direct `RunWorkflow` test uses the fake dispatcher.
- [x] 3.3 Add a scheduled-run function: it calls `RunWorkflow`, logs the command, owner/repo, workflow, result and run URL, and on error sends a failure message to every `allowed_chat_id`. Verify with a test where the fake dispatcher fails and the fake messenger receives one message per allowed chat.

## 4. Scheduler

- [x] 4.1 Create `internal/services/scheduler` (`New(loc, logger)`, `Add(name, spec, fn)`, `Start()`, `Stop()`) on robfig/cron with `WithLocation`, logging the zone and the next run of each entry on `Start`. Verify with a test: for `0 9 * * 1,3,5` from a fixed Sunday instant, the next run is Monday 09:00 America/Guayaquil (14:00 UTC), and later runs fall only on Mon/Wed/Fri.
- [x] 4.2 Verify with a test that a function registered at `* * * * *` (or through an injectable clock or schedule) runs without network access, and that `Stop()` prevents later runs.
- [x] 4.3 Wire the scheduler in `main.go`: build it only when an entry has `schedule`, register each scheduled workflow with the scheduled-run function, start it, and stop it on SIGINT/SIGTERM. Verify by running the binary locally in Docker with a config whose schedule is one minute ahead and a fake GitHub token, and check the startup log line and the dispatch attempt in the log.

## 5. Cross-repo

- [ ] 5.1 (workflows) Remove `schedule:`, the `should-run` job, its `needs:` and its conditions from `ops-github-runner-check-updates.yml` and `ops-github-runner-cleanup.yml`, keeping `workflow_dispatch`. Verify with `actionlint`.
- [ ] 5.2 (workflows) Rewrite the README "Host-ops: github-runner" section to say that automation-hub owns the schedule. Verify by review.
- [ ] 5.3 (user) Add `schedule` to `run_check_updates` (`0 9 * * 1,3,5`) and `run_cleanup` (`0 9 * * 2,5`) in Infisical `CONFIG_YAML`, and confirm `owner`/`repo` point to `devidence-dev/workflows`. Verify after deploy that the startup log shows `America/Guayaquil` and the expected next runs.
- [ ] 5.4 (user) Watch the first scheduled run. Verify that the GitHub run started at 09:00 America/Guayaquil and that `/run_check_updates` still works from Telegram.

## 6. Validation

- [x] 6.1 Run `docker run --rm -v "$PWD":/app -w /app golang:$(cat .go-version) sh -c 'go vet ./... && go test -race ./...'`. The task passes only when both commands succeed on the current code.
