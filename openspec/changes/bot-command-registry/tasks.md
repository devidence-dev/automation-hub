# Tasks

## 1. Prerequisite gate

- [x] 1.1 Confirm that `refactor-dependency-report` has landed `internal/messages` (embedded default YAML, config override). Record its lookup signature and the override section name in design.md Open Questions, and update decision 5 if they differ. Verify with `go doc ./internal/messages` showing the lookup used below, and with that change's catalog tests passing in Docker.

## 2. Configuration

- [x] 2.1 Add `GitHubConfig.WebURL` (`web_url`), `RestartReadyConfig` (`restart_ready`, `timeout` and `poll_interval` as `time.Duration`) and the `messages` override map to `internal/config/config.go`, with defaults `https://github.com`, `3m` and `5s` set through `viper.SetDefault`. Verify with table test `TestLoadDefaults` in `config_test.go`: a config without these sections yields the three defaults, and an override yields the configured value.
- [x] 2.2 Add `validateGitHubWebURL` (absolute http or https URL, trailing slash trimmed) and `validateRestartReady` (timeout and poll interval positive, poll interval not longer than timeout) next to the existing validators. Verify with table tests `TestValidateGitHubWebURL` and `TestValidateRestartReady`: `github.com` fails naming `github.web_url`, `timeout: "0s"` fails naming `restart_ready.timeout`, `timeout: "5s"` with `poll_interval: "10s"` fails naming `restart_ready.poll_interval`, and `timeout: "soon"` fails naming `restart_ready.timeout`.
- [x] 2.3 Check `messages:` overrides against the catalog defaults at load, so a placeholder mismatch fails config loading. Verify with `TestLoadMessagesPlaceholderMismatch`: `bot.job_created: "🚀 Job creado."` fails with an error naming `bot.job_created`, and a same-shape override loads.
- [x] 2.4 Document `github.web_url`, `restart_ready`, and `messages:` with their defaults in `configs/config.yaml.example` and the README configuration section. Verify with `TestLoadExampleConfig` passing and `grep -n "restart_ready\|web_url" README.md configs/config.yaml.example` showing both files.

## 3. Message texts

- [x] 3.1 Register the bot keys from design.md decision 5 in the shared catalog's default YAML, with the current English strings. Remove the Go string literals for these texts from `telegram_bot.go`. Verify with `TestBotTextDefaults`: each key resolves to the exact current string (including `❓ Unknown command.` and `⏱️ Deployment *%s* did not report ready within %s — check it manually.`), and an override in `messages:` replaces only its key.

## 4. Command registry

- [x] 4.1 Define the `CommandSpec` interface and the shared `allowList` type in `internal/handlers/command.go`. Verify with table test `TestAllowList`: an empty list denies, a listed chat is allowed, and an unlisted chat is denied.
- [x] 4.2 Implement the workflow executor in `command_workflow.go`: `Run` (dispatch and reply), `RunScheduled` (returns an error), and the run-URL fallback built from `github.web_url`. Verify with `TestWorkflowCommandRun`: the reply uses `bot.workflow_dispatched`, the fallback is `https://github.example.com/devidence-dev/workflows/actions` for a configured `web_url`, and a dispatch error sends `bot.dispatch_failed`.
- [x] 4.3 Implement the job executor in `command_job.go`. Verify with `TestJobCommandRun`: success replies `🚀 Job *dependency-report-manual-1791763200* created from *dependency-report*.`, a 403 replies with the `bot.job_failed` text, and `TestJobRunScheduled` checks that the error is returned for the scheduled path.
- [x] 4.4 Implement the restart executor in `command_restart.go`, taking its `RestartReady` values by injection and running the readiness wait in the background. Verify with table test `TestRestartCommand`: triggered reply, ready reply, 403 failure with no wait started, timeout reply `within 1m0s` for a 1m timeout and `within 3m0s` by default, and a fake rollout that blocks until its context ends does not block the next `Run`.
- [x] 4.5 Add `ScheduledRunner` in `scheduled.go`, which logs each scheduled outcome and sends `bot.scheduled_dispatch_failed` or `bot.scheduled_job_failed` to each `allowed_chat_ids` entry. Verify with `TestScheduledRunnerNotifiesEveryChat`: a failing executor sends one message per allowed chat, and the success path logs `result=dispatched` with a zap observer.
- [x] 4.6 Replace the three allow and dispatch blocks in `BotHandler.Handle` with a name lookup over the registry. Keep the unknown and unauthorized replies and the warning log in the router. Verify with the `telegram_webhook_test.go` cases moved onto `Handle` and `ScheduledRunner` with the same assertions, plus `TestHandleUnknownCommand` and `TestHandleUnauthorizedLogsWarning` (zap observer, command, chat ID and username fields).
- [x] 4.7 Add `NewCommandSpecs(deps)`, which returns specs in config order (workflows, restarts, jobs) and leaves unused client fields as untyped nil interfaces. Verify with `TestNewCommandSpecsOrder` and `TestNewCommandSpecsWorkflowsOnlyLeavesKubeNil`.

## 5. Scheduling and menu wiring

- [x] 5.1 Replace `registerBotCommands` and `registerSchedules` in `cmd/automation-hub/main.go` with one loop over the specs that publishes the menu and registers schedules. Verify with `TestRegisterCommandsOneLoop` in `main_test.go`, using a fake menu publisher and a fake scheduler adapter: the menu has three entries in config order, and only entries with a schedule are added.
- [x] 5.2 Make `newGitHubClient` and `newKubeClient` return untyped nil when no matching entry is configured, and assign their results to the interface fields only after that check. Verify with `TestNewKubeClientNilWhenUnconfigured` and `TestNewGitHubClientNilWhenUnconfigured`: the returned interface compares equal to nil.

## 6. Webhook registry

- [x] 6.1 Change `NewWebhookHandler` to take `(telegramClient *telegram.Client, hook config.WebhookProcessorConfig, logger)`, with no per-request hook lookup and no `*config.Config`. Verify with the updated `webhook_test.go` cases: a valid POST returns 200 with `{"status":"success"}`, and a wrong method is rejected by the route. The processing-error case (500 `Processing failed`) needs a fake sender, which exists only once `email-processing-cleanup` gives `NewTorrentProcessor` its `messageSender` interface, so that change tests it (its task 6.3).
- [x] 6.2 Add `internal/handlers/webhook_registry.go` with a name to factory map (`qbittorrent`, `dependency-report`) and make `registerRoutes` a single loop over the configured hooks and the dependency report section. Verify with `TestRegisterRoutesDependencyReport` (updated call), `TestRegisterRoutesUnknownHookWarns` (observer logger: warning logged, no route, 404), and `TestRegisterRoutesQbittorrent` (POST routed, GET rejected).

## 7. Validation

- [x] 7.1 Run `docker run --rm -v "$PWD":/app -w /app golang:$(cat .go-version) sh -c 'go vet ./... && go test -race ./...'`. The task passes only when both commands succeed on the current code.
- [x] 7.2 Run golangci-lint v2.14.0 in Docker with `--enable=errcheck,govet,staticcheck,unused,misspell,gosec`, and gocognit v1.2.0 in Docker with `-over 15 -test` on `internal/handlers`, `internal/config` and `cmd/automation-hub`. Verify that golangci-lint reports no issues and gocognit lists no function above 15.
- [x] 7.3 Run `openspec validate --strict bot-command-registry` and confirm it reports the change as valid.
