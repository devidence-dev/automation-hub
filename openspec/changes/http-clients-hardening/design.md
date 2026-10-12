# Design

## Context

Three clients build HTTP stacks on their own: `telegram.newHTTPClient` (custom transport, timeouts derived from `pollTimeout`), `github.NewClient` (a 30s client on the default transport) and `kube.NewInClusterClient` (CA-pinned transport, 15s client). Response reads use `io.LimitReader(…, 1<<20)` in four places. Telegram's `send` retries every error with `time.Sleep`. The `tgbotapi` v5.5.1 library returns `*tgbotapi.Error{Code, Message, ResponseParameters{RetryAfter}}` for Telegram error answers (`MakeRequest` in `bot.go`), and a raw `encoding/json` error when the body is not JSON. It does not expose the HTTP status otherwise. See proposal.md for motivation and specs/http-transport/spec.md for the requirements.

## Goals / Non-Goals

**Goals:**
- One place that builds transports, clients and capped JSON reads.
- Every value in the spec comes from config, with defaults equal to today's values.
- Retry is a decorator that can be tested without sleeping.

**Non-Goals:**
- No change to which requests are made, beyond the GitHub base URL and the Telegram retry rule in the spec.
- No generic retry for GitHub or Kubernetes.

## Decisions

**1. `internal/httpx` is a stdlib package with two functions.** `httpx.Settings` holds the transport values (timeout, dial, TLS handshake, keep-alive, idle connections, idle timeout, optional response-header and expect-continue timeouts, optional `*tls.Config`). `httpx.NewClient(Settings) *http.Client` is the factory. `httpx.Send(ctx, client, Request, maxBytes) (Response, error)` sends one request and reads its body up to `maxBytes`. `Response` carries the status code, status text and the capped body. The caller decides what a status means. Alternative considered: a generic `DoJSON[T]` with a status and out parameter. It was rejected because GitHub's 200 path tolerates an undecodable body while Kubernetes does not, and a single generic helper would need flags for that. `httpx.Send` is the shared part; the decoding stays in each client.

**2. Config values are plain structs filled in `internal/config`.** Each section maps to a `Settings` value, or to the client's own `Config` struct (`telegram.Config`, `github.Config`, `kube` options). Config defaults are one table set with `viper.SetDefault` in `internal/config`, the same mechanism `bot-command-registry` uses for `github.web_url`, and duration keys rely on viper's default decode hook. Clients never import `config`, which keeps the dependency direction one way. The default for `max_response_bytes` is the exported constant `httpx.DefaultMaxResponseBytes`, so the test constructors keep working without a magic number.

**3. Telegram retry is a decorator with a strategy.** `RetryingSender` wraps a `sender` interface (`Send(tgbotapi.Chattable) (tgbotapi.Message, error)`, which `*tgbotapi.BotAPI` already satisfies). The delay comes from a `Backoff` interface, `Delay(attempt int, err error) (time.Duration, bool)`. The production policy reads `retry_after` from `*tgbotapi.Error` when present and otherwise doubles `retry_backoff`. The wait is an injected `sleep func(time.Duration)`, so tests record the sequence instead of sleeping. Retryable means: any error that is not a `*tgbotapi.Error` (network or undecodable body), `Code >= 500`, or `Code == 429`. The client keeps `*tgbotapi.BotAPI` for `GetUpdatesChan`, `StopReceivingUpdates` and `Request` (`SetMyCommands`), which are not retried today and stay unchanged.

**4. Telegram sends share one chat-ID helper.** `resolveChat(chatID string) (int64, error)` plus a `ready()` nil guard. `sendText` and `SendDocument` both call them, then build the chattable and hand it to the sender. The public methods keep their signatures.

**5. Kubernetes and GitHub keep their seams.** `newClient`, `NewClientWithBaseURL` and the injected `now` stay. The Job manifest becomes typed structs (`job`, `objectMeta`, `ownerReference`) built by `jobFromTemplate(cj, name)`. The `spec` stays `json.RawMessage` so unknown template fields are still forwarded. Path formats and media types are named constants; `resourceURL(base, namespace, name)` replaces the two inline `fmt.Sprintf` calls. `RestartDeployment` and `getDeployment` both use `doJSON`, which uses `httpx.Send`.

**6. GitHub dispatch splits into a builder and a decoder.** `newDispatchRequest(owner, repo, file, ref)` builds the request, and `decodeDispatch(resp)` maps 204 to empty, 200 to the decoded `html_url` (with the existing warning on an undecodable body), and anything else to the existing error text. The wire format is unchanged.

**7. Server timeouts extend the existing `server:` section.** `server.read_timeout`, `server.write_timeout` and `server.shutdown_timeout` sit next to `server.address`, not under a new top-level `http:` key. `newHTTPServer` and `shutdownServer` receive the values from the config.

## Risks / Trade-offs

- **[Risk] A transient 4xx (for example, a 400 caused by an upstream proxy) is no longer retried.** → Mitigation: 4xx is reported immediately, which is what the spec requires. The log line keeps the attempt count so it stays visible.
- **[Risk] A network error after the request reached Telegram can duplicate a message.** → Mitigation: this is today's behavior too. It is kept on purpose and documented, not changed silently.
- **[Risk] An undecodable body (for example, an HTML page from a proxy) is classified as transport and retried.** → Mitigation: `tgbotapi` hides the HTTP status for that case, so the classifier cannot do better. Attempts are bounded by `max_attempts`.
- **[Risk] A long `retry_after` delays the caller.** The dependency-report webhook answers 202 and sends from a goroutine (see `add-dependency-report-html`), so it is not affected. Command replies and the torrent webhook send synchronously. → Mitigation: waits above `max_retry_after` fail fast. Otherwise the worst case is `max_attempts - 1` waits of up to `max_retry_after`, which can exceed `server.write_timeout` for a synchronous caller. The process stays up. This is accepted and documented.
- **[Trade-off] `github.api_url` (this change, API base, https only) and `github.web_url` (`bot-command-registry`, run-link fallback, http or https) are separate keys.** They must not be merged: a GHES deployment sets both, and the two validators differ on purpose.
- **[Trade-off] Config keys are added to the Telegram section, but the same `telegram.http` settings apply to both the notifications bot and the workflow bot.** The workflow bot has no separate section, which keeps the config small.
- **[Trade-off] Telegram gains a 90s idle timeout, and Kubernetes gains a 30s keep-alive and a 90s idle timeout, which they lack today.** Both are already bounded by the request timeout, so the effect is limited to reaping idle sockets.

## Migration Plan

No config migration is required. Every new key is optional and defaults to today's value. Rollback is a binary revert: an older binary ignores the unknown keys.
