# Proposal

## Why

`BotHandler.Handle` repeats the same allow-check and dispatch for workflows, restarts and jobs. `main.go` loops over the command types twice (Telegram menu and scheduler), and webhook routes switch on the hook name. Bot texts, the restart wait (3m, 5s) and the GitHub web URL are hardcoded, against the project rule that user-facing texts and limits come from config.

Defaults equal current behavior, so nothing visible changes unless an operator sets the new keys.

## Dependency

**Depends on `refactor-dependency-report`, which must land first (and that change needs `add-dependency-report-html` archived).** It creates the shared message catalog (`internal/messages`, embedded default YAML, overridable from `messages:`). This change adds its bot texts as catalog keys with the current English defaults. It creates no second catalog.

## What Changes

- Command registry: `BotHandler.Handle` resolves a `Command` (`Authorize`, `Run`) from a map keyed by command name. Workflow, restart and job executors implement it. Scheduled runs call the same executors without a chat.
- One adapter interface lets `main.go` publish Telegram commands and register schedules in a single loop.
- Webhook registry: name to factory replaces `switch hook.Name`. The dependency-report route registers through it.
- Bot texts move to the catalog; defaults are the current English strings.
- New config keys with defaults: `github.web_url` (`https://github.com`), `restart_ready.timeout` (`3m`), `restart_ready.poll_interval` (`5s`), plus catalog `messages:` overrides.
- Clients are built only when configured; absent clients are passed as explicit nil interfaces.

No breaking change.

## Capabilities

### New Capabilities
- `bot-command-registry`: authorization, dispatch and scheduling of command types; bot texts from the catalog.
- `restart-commands`: Telegram restart behavior and its configurable readiness wait.

### Modified Capabilities
- `job-commands`: the job reply text becomes overridable through the catalog.
- `scheduled-dispatch`: the run-link fallback uses the configurable GitHub web URL.

## Impact

- `internal/handlers/telegram_bot.go` (split), `webhook.go`, `cmd/automation-hub/main.go`. `internal/services/processor` is not touched; `email-processing-cleanup` owns it.
- `internal/config/config.go`: new keys, defaults, fail-fast validation.
- `internal/messages`: new bot keys only.
- Tests: table-driven with fakes; existing handler tests pass unchanged.
- Docs: `configs/config.yaml.example`, `README.md`.
- Coordination: `add-dependency-report-html` and `refactor-dependency-report` also edit `main.go` and the dependency-report handler. This change touches only route registration and the bot wiring.

## Non-goals

- No new commands, Telegram behavior or dependencies.
- No change to dependency-report texts or behavior, and no second catalog.
- No change to HTTP client transports (`http-clients-hardening`) or email processing (`email-processing-cleanup`).
- No per-command restart timeouts, multi-language messages or hot reload.
