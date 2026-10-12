# Design

## Context

`internal/handlers/telegram_bot.go` holds one `BotHandler` that routes, authorizes, dispatches, replies, runs scheduled commands and runs the restart wait. `cmd/automation-hub/main.go` loops over the three command lists twice (Telegram menu in `registerBotCommands`, schedules in `registerSchedules`) and routes webhooks with `switch hook.Name`. `WebhookHandler` keeps the whole `*config.Config` and looks up its hook on every request. `processor.NewTorrentProcessor` takes a concrete `*telegram.Client`. See proposal.md for the motivation.

## Goals / Non-Goals

**Goals:** one place per command type for authorization, dispatch and texts; one loop over command specs; defaults equal current behavior; no client is built or passed as a typed nil.

**Non-Goals:** no change to the consumer-defined `workflowDispatcher`, `deploymentRestarter`, `jobCreator` and `telegramMessenger` interfaces; no change to the HTTP clients; no change to `DependencyReportHandler` internals.

## Decisions

**1. Command objects behind one interface, routed by a map.** `CommandSpec` is the single adapter. It exposes `Name()`, `Description()`, `Schedule()`, `Authorize(chatID)`, `Run(ctx, chatID)` (interactive reply) and `RunScheduled(ctx) error`. Map lookup by name replaces the three `if … exists` blocks. Authorization is shared through an embedded `allowList` type, so the three copies of `isAllowed` disappear. Alternative rejected: a type switch in `Handle`, which keeps routing coupled to every type.

**2. Executors per type, files per type.** `command_workflow.go`, `command_restart.go` and `command_job.go` each implement `CommandSpec` with the same consumer-defined interfaces as today. Each executor sends its own replies through the injected `telegramMessenger`. Restart keeps the background wait, now given its `RestartReady` values by injection.

**3. Separate interactive router and scheduled runner.** `BotHandler` shrinks to routing: parse the command, look it up, unknown and unauthorized replies, then call `Run`. A new `ScheduledRunner` (`scheduled.go`) wraps `RunScheduled` with logging, the failure notification to `AllowedChatIDs` and the shared `messages` text. `main.go` passes `spec.RunScheduled` through the runner, so no executor needs a scheduler reference.

**4. One adapter loop in main.** `handlers.NewCommandSpecs(deps)` returns the specs in config order (workflows, restarts, jobs). `main.go` iterates once to publish `setMyCommands` and once to register schedules, both over the same slice. The loop bodies are the only places that read `Name`, `Description` and `Schedule`. The two loops stay in `main.go`, because the scheduler and Telegram client are wired there.

**5. Texts from the catalog.** Every reply uses a key of the shared catalog from `internal/messages` (defined by refactor-dependency-report). This change adds keys `bot.unknown_command`, `bot.unauthorized`, `bot.workflow_dispatched`, `bot.dispatch_failed`, `bot.scheduled_dispatch_failed`, `bot.job_created`, `bot.job_failed`, `bot.scheduled_job_failed`, `bot.restart_triggered`, `bot.restart_failed`, `bot.restart_ready`, `bot.restart_timeout`, with the current strings as defaults. Placeholder counts are checked when the catalog loads (see the spec).

**6. Configurable values.** `github.web_url` (default `https://github.com`, trailing slash trimmed, must be an absolute http or https URL) is used only for the run-link fallback in `RunWorkflow`. `restart_ready.timeout` (default `3m`) and `restart_ready.poll_interval` (default `5s`) are `time.Duration` fields. Viper's default decode hooks parse them. Defaults are set with `viper.SetDefault` in `config` package, one table, not in code paths. `validateRestartReady` and `validateGitHubWebURL` are fail-fast validators next to the existing ones.

**7. No typed-nil clients.** `newGitHubClient` and `newKubeClient` return the interface type and `return nil` when unconfigured, so `BotHandlerDeps` fields are nil interfaces. The current `*github.Client` and `*kube.Client` return values are assigned to interfaces only after the nil check.

**8. Webhook registry.** A `map[string]RouteFactory` in `internal/handlers/webhook_registry.go` maps a route name to a factory `func(hook config.WebhookConfig, env RouteEnv) (Route, error)`, where `Route` holds the path and the `http.Handler`. `qbittorrent` is registered there. `dependency-report` is registered in the same map, and its factory reads the `dependency_report:` section and uses the fixed path. `registerRoutes` becomes one loop: for each configured hook, and for `dependency-report` when its section is set, look up the factory, log unknown names as warnings, and mount the route. Alternative rejected: a separate special case for dependency-report, which is the code the change removes.

**9. Webhook handler gets its inputs at registration.** `NewWebhookHandler(telegramClient *telegram.Client, hook config.WebhookProcessorConfig, logger)` receives the hook config from the registry factory. It no longer looks up the hook per request, so the 500 "configuration not found" branch goes away with it. The processor side is out of scope (decided 2026-10-11): `email-processing-cleanup` owns `processor.NewTorrentProcessor`, its sender interface (`messageSender`) and `processor.GetWebhookConfig`. That change lands later and replaces the concrete client with its interface.

## Risks / Trade-offs

- [Behavior drift in the reply texts] → Defaults are copied verbatim; table tests assert every default string, and `TestLoadExampleConfig` keeps the example valid.
- [Catalog API unknown until refactor-dependency-report lands] → Coding waits for its API. If it lacks a formatted lookup with args, this change stops and asks, rather than adding a second catalog.
- [Restart goroutine after shutdown] → Unchanged from today: the wait uses `context.Background()` and a timeout. Documented as a known limit; no new guarantee is added.
- [Test churn] → `registerRoutes`, `NewWebhookHandler` and the handler tests change signature. Behavior assertions stay the same.

## Migration Plan

No config migration: every new key is optional with a default. Rollback is reverting the commit. Deploy order is refactor-dependency-report first, then this change, then the README and config example in the same commit.

## Open Questions

- Language of the catalog. Resolved on 2026-10-11: each key keeps its current text, so `bot.*` defaults are English. `refactor-dependency-report` states the same rule.
- Exact catalog API and the name of the override section (`messages:` assumed) come from refactor-dependency-report. Gate: task 1.1 confirms them before any other task starts.
- Whether `restart_ready` should also be settable per restart entry. Current proposal: global only.
