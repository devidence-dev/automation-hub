# Design

## Context

Behavior is specified in `specs/email-code-forwarding/spec.md` and `specs/torrent-webhook-notification/spec.md`; motivation is in `proposal.md`.

Current code, verified by reading the sources:

- `processor/generic.go` builds three regexes keyed by service name, compiles the custom pattern and the Perplexity regexes per call, branches on the name for MIME stripping (`cloudflare` keeps headers) and for the Perplexity marker cascade, and holds a `*telegram.Client`.
- `processor/manager.go` has a first-match dispatch loop (`processEmailAsync`) that only tests use, plus `ProcessEmailsConcurrently`. `cmd/automation-hub/main.go` passes processors to `StartMonitoring`, which dispatches again in `email/imap.go` (`processMessage`). Mark-as-read is decided by an anonymous interface assertion in `imap.go`.
- `imap.go` dereferences `msg.Envelope` before its nil check (`parseMessage`).
- `processor/torrent.go` holds `GetWebhookConfig` and `NewTorrentProcessorLegacy`. `handlers/webhook.go` reads the body without a limit.
- Dead code, confirmed by grep: `NewTorrentProcessorLegacy` (only `torrent_test.go`), `markAsUnread` (only `imap_test.go`), `ProcessEmailsConcurrently` and `processEmailAsync` (only `manager_test.go`).
- The existing `openai` hex-color test (`TestExtractCode_IgnoresHexColor`, commit f6c17ae) relies on a capture group in `code_pattern`. It must keep passing.

## Goals / Non-Goals

**Goals:**
- Every service's extraction, body preparation, not-found text and mark-as-read come from its config entry. No name-keyed branch remains in code.
- One dispatch path (the Manager). The IMAP code is a transport plus a poller.
- Processors depend on consumer interfaces only.
- Every function stays at cognitive complexity 15 or below.

**Non-Goals:**
- Webhook route registration, the Telegram client, the dependency-report catalog (see proposal).
- A plugin registry, a general template engine, or a generic message bus.

## Decisions

### D1. Extraction is a strategy chosen by `extractor`, with two implementations
`CodeExtractor` has one method, `Extract(text) (code, found)`. Two implementations exist:
- `pattern` (default): first match of `code_pattern`; capture group 1 wins when it matched a non-empty value.
- `marker`: the first marker in `code_markers` in configured order (not text position, as today) that occurs, case-insensitively; then `code_candidates` in order on the text after it; then `code_pattern` as fallback.

A factory selects the implementation from `extractor`. Alternatives considered: a registry of named extractors (speculative, rejected), and a single regex with an optional marker prefix (cannot express the Perplexity cascade in order, rejected).

### D2. Built-in name defaults are removed; configs carry the values
Today's cloudflare and perplexity behavior is written out in `configs/config.yaml.example` and in the migration YAML of task 6.1. A service with no optional keys uses the generic defaults (`pattern`, `\b[a-zA-Z0-9]{4,8}\b`, `Not found`, headers stripped). This is the only way to satisfy "no name-keyed code", but it means a CONFIG_YAML that is not migrated silently changes Cloudflare extraction. Mitigations are in Risks.

`strip_mime_headers` is not in the audit list. It is required because the name branch that keeps Cloudflare headers is otherwise lost.

### D3. Defaults live in `config`, applied through accessors
The config package defines documented defaults (`DefaultPollingInterval`, `DefaultCodePattern`, `DefaultNotFoundText`, `DefaultLogPreviewLength`, `DefaultWebhookMaxBodyBytes`) and accessor methods (`EmailConfig.EffectivePollingInterval()`, `ServiceProcessorConfig.EffectiveCodePattern()` and so on). Accessors are used instead of `viper.SetDefault` alone because processor and webhook tests build structs directly, and both paths must see the same defaults. `Load` validates values (negative intervals and caps, regex syntax, extractor enum) so that errors surface at startup.

### D4. Regexes are validated at load and compiled once at construction
`config.Load` calls `regexp.Compile` on `code_pattern` and each `code_candidates` entry, so a bad regex stops startup with the service name in the error. Today it logs a warning and falls back. The processor compiles again once, in its constructor, and stores the result. No compilation happens per email.

### D5. `messageSender` and `TemplateNotifier`
Processors and the torrent processor depend on `messageSender { SendMessage(chatID, message string) error }`, which `*telegram.Client` already satisfies (the same pattern as `handlers/telegram_bot.go`). `TemplateNotifier{sender, chatID, template}.Notify(args ...any)` does `fmt.Sprintf` plus send. Both torrent and email use it, which removes the duplicated flow.

### D6. `GenericEmailProcessor` is a composition
- matcher: sender and subject containment (unchanged semantics);
- body preparation: quoted-printable decode, then MIME header strip unless disabled;
- `CodeExtractor`;
- not-found substitution, which logs a warning capped at `log_preview_length`;
- `TemplateNotifier`.

`Process` reads as those five steps. Each step is a small private function that can be tested alone.

### D7. One dispatcher, two consumer interfaces in `email`
`Manager.Dispatch(email) (models.EmailProcessor, error)` tries services in config order, processes the first match, logs the outcome, and returns the matched processor (nil when none matched). `Manager.Senders()` returns the `email_from` values.

The poller in `email` depends on a `Dispatcher` interface with those two methods, not on `*processor.Manager`. Mark-as-read is decided in the poller: the matched processor is checked against the named optional interface `models.ReadMarker { ShouldMarkAsRead() bool }`. A processor without it, a failed dispatch and a nil match all leave the email unread.

`processEmailAsync` and `ProcessEmailsConcurrently` are deleted. Nothing in production uses them.

### D8. Transport is separate from polling
- `MailSource { Unread(senders []string, handle func(models.Email) (markRead bool)) error }`. The go-imap adapter (`IMAPSource`) dials, logs in, selects INBOX, searches per sender (union), fetches, parses and calls `handle`. When `handle` returns true, it stores `\Seen` while the session is still open. It returns errors (wrapped with the host) and does not log poll failures itself.
- `Poller` runs the ticker, calls `Unread`, and logs each returned error once. Its `handle` is the mark policy from D7.

The go-imap client sits behind a small `imapSession` interface (Login, Select, Search, Fetch, Store, Logout), which `*client.Client` satisfies. Search and error handling are testable with a fake, with no network.

### D9. Torrent webhook: cap in the handler, processor only formats and sends
The hook is read once at registration (`bot-command-registry` D9), so the handler already holds its config and no per-request lookup is added. `handlers/webhook.go` reads the body through `http.MaxBytesReader` with the hook's `EffectiveMaxBodyBytes()`, then decodes it, so a body over the cap is rejected whole. A `MaxBytesError` maps to `413`. The processor only formats and sends. The cap sits on the hook's config because it applies to this route only. The dependency-report upload has its own limit and is out of scope.

## Coordination with sibling changes

- `processor.NewTorrentProcessor` and `processor.GetWebhookConfig` belong to this change (decided 2026-10-11). `bot-command-registry` no longer touches them; it lands first and passes `*telegram.Client` to the handler, which satisfies `messageSender`.
- `bot-command-registry` design item 9 moves the hook lookup to route registration and passes the hook config to the handler. The cap in D9 reads the same config either way, so whichever change lands second adapts the lookup only.
- `GetWebhookConfig` moves to `config.FindWebhook` in task 1.4 here; `bot-command-registry` leaves it in place.
- `main.go` edits are on different lines from `http-clients-hardening` (server timeouts, client construction) and `refactor-dependency-report` (`registerRoutes`, sender interface). Whichever lands later rebases the wiring.
- `http-clients-hardening` changes outbound limits and server timeouts. Its `httpx` request caps are not used by the processors or the torrent handler here.

## Risks / Trade-offs

- **Unmigrated CONFIG_YAML changes Cloudflare extraction.** Without `strip_mime_headers: false` and `code_pattern`, Cloudflare would use the generic pattern on headers. → Mitigation: the migration YAML in task 6.1 is applied before deploy. It is backward compatible, because the current binary ignores unknown keys (mapstructure default), so applying it first is safe. A startup warning fires when a service uses the generic default pattern, and the "Loaded email processor" log gains `extractor` and effective pattern fields.
- **Invalid regexes now stop startup** (previously a warning and a fallback). → Mitigation: stated in the proposal, tested in task 1.3. It is a deliberate tightening.
- **Negative `polling_interval` is rejected** (today `time.NewTicker` panics). → Mitigation: explicit error at load.
- **Logs change:** the Perplexity search-text preview drops from 300 to the configured 200 characters. Logs only, no behavior change.
- **No request-time hook lookup.** A hook absent from `hook:` mounts no route, so a POST to its path gets the mux's 404 and not the spec's 500 `Webhook configuration not found`. Producing that 500 would need a per-request lookup, which D9 rules out. → Open question below.
- **Shared wiring in `cmd/automation-hub/main.go`.** Sibling changes also edit it. → Keep the edit to the constructor and poller lines, and rebase carefully.
- **Test churn.** Tests keyed on the removed name defaults (`BuiltInPatterns`, the Cloudflare and Perplexity cases in `TestExtractCode` and `TestExtractPerplexityCode`) get explicit config with the same expectations. `TestExtractCode_IgnoresHexColor` keeps its body and assertion.

## Migration Plan

1. Merge the code, example config and README (tasks 1 to 5).
2. Apply the CONFIG_YAML migration (task 6.1) in Infisical. The running binary ignores the new keys, so behavior does not change.
3. Deploy the new image. Check the startup log: `extractor=pattern`/`marker` per service, and no "generic default pattern" warning for cloudflare or perplexity.
4. Rollback: redeploy the previous image. The migrated keys are harmless to it.

## Open Questions

- Hook configuration lookup (`specs/torrent-webhook-notification`): the spec requires `500 Webhook configuration not found` for a missing `qbittorrent` hook. With the hook fixed at registration, that case is a missing route (404). Decide: amend the spec to the 404, or accept a per-request lookup against D9.

Possible later work (not in this change): `IDLE` instead of polling, and per-service body caps for other webhooks.
