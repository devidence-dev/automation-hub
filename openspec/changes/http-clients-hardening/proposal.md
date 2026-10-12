# Proposal

## Why

The three outbound HTTP clients (Telegram, GitHub, Kubernetes) each build their own `http.Client` and transport, cap response bodies with repeated literals, and hardcode timeouts, retry counts and the GitHub API host. The Telegram retry loop retries every error, including permanent 4xx answers, and no value can be tuned without a rebuild. The server timeouts in `main.go` are hardcoded too.

## What Changes

- Add `internal/httpx`, a standard-library package with one transport and client factory (dial, TLS, idle connections, timeouts) and one JSON request helper with a response body cap. It replaces the three client builders and the four hardcoded 1 MiB limits.
- Make every value that is hardcoded today configurable, with today's value as the documented default: `server.read_timeout`, `server.write_timeout`, `server.shutdown_timeout`, `telegram.http.*`, `github.api_url`, `github.http.*` and `kubernetes.http.*`. Invalid values stop startup.
- Move Telegram retries into a `RetryingSender` decorator with an injected backoff policy.
- **BREAKING** (behavior): Telegram retries only network errors, 5xx and 429. A 4xx answer is a single attempt. A 429 waits for the `retry_after` Telegram provides.
- Share the Telegram chat ID parse and nil guard between text and document sends.
- Refactor the GitHub dispatch and Kubernetes restart and Job code (request builder, response decoder, `doJSON`, path and media-type constants, injected clock, typed Job manifest) with no change to wire behavior.

## Capabilities

### New Capabilities

- `http-transport`: configurable timeouts, body caps and API base URL for the outbound clients and the server, and the Telegram retry policy.

### Modified Capabilities

None. No existing requirement changes. `scheduled-dispatch` and `job-commands` keep their wording.

## Impact

- Code: `internal/httpx` (new); `internal/services/{telegram,github,kube}/client.go`; `cmd/automation-hub/main.go` (server timeouts, client construction); `internal/config/config.go`.
- Config and docs: new keys in `configs/config.yaml.example` and `README.md`. No existing key is renamed.
- Dependencies: none new. Standard library plus the existing `telegram-bot-api/v5` v5.5.1, which exposes `*tgbotapi.Error` with `RetryAfter`.
- Signatures: `telegram.NewClient` gains a settings argument, which touches the wiring in `main.go`.

## Non-goals

- No client-go and no new dependency.
- No retry for GitHub or Kubernetes calls.
- No change to Telegram message texts, bot command routing, webhook handlers, timezone handling, the dependency report, IMAP or processors. Those belong to the sibling changes.
- No body cap on responses read inside `tgbotapi`.
- No change to the 200 and 204 handling of `DispatchWorkflow` or to `rolloutComplete`.
