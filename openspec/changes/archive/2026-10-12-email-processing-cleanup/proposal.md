# Proposal

## Why

Code extraction is keyed by service name in code (`cloudflare`, `perplexity`, `default`). The IMAP client mixes transport with dispatch and repeats the processor Manager's loop. Processors depend on `*telegram.Client`. Poll interval and the "Not found" text are hardcoded. Defaults must reproduce today's behavior, and the Infisical CONFIG_YAML must keep working.

## What Changes

- Code extraction becomes a per-service strategy chosen by config: `extractor: pattern` (default, uses `code_pattern`) or `extractor: marker` (`code_markers`, ordered `code_candidates`, `code_pattern` as fallback). Cloudflare and Perplexity behavior becomes explicit entries in `configs/config.yaml.example`. Regexes compile once at startup; an invalid regex fails config loading.
- `GenericEmailProcessor` is split into a matcher, a `CodeExtractor` strategy and a template notifier. Processors depend on a consumer `messageSender` interface, not on `*telegram.Client`.
- The IMAP client is split into a `MailSource` transport (dial, search, fetch, parse, flag) and a poller. Dispatch goes only through the processor Manager.
- `ShouldMarkAsRead` becomes a named optional interface in `models`.
- Dead code is removed: `NewTorrentProcessorLegacy`, `markAsUnread`, `ProcessEmailsConcurrently`, `processEmailAsync`.
- Hardcoded values move to config with defaults: polling interval (60 s), `not_found_text` ("Not found"), `log_preview_length` (200). The torrent hook is resolved once, at route registration (`bot-command-registry`).
- Poll-cycle failures are logged. A nil IMAP envelope no longer panics.
- The torrent webhook JSON body is capped by `max_body_bytes` (default 1 MiB).
- **Config migration:** CONFIG_YAML must add the Cloudflare and Perplexity keys listed in the tasks before rollout.

## Capabilities

### New Capabilities

- `email-code-forwarding`: IMAP polling, sender and subject matching, code extraction strategies, not-found text, Telegram forwarding and the mark-as-read policy. Documents current behavior plus the new configurability.
- `torrent-webhook-notification`: qBittorrent completion notification (body cap, message template, config lookup). Separate because it is a webhook, not email.

### Modified Capabilities

None. No existing spec covers email or webhooks.

## Impact

- Code: `internal/services/email/`, `internal/services/processor/`, `internal/models/models.go`, `internal/config/config.go`, `internal/handlers/webhook.go` (lookup and body cap only), `cmd/automation-hub/main.go` (wiring only).
- Docs: `configs/config.yaml.example`, `README.md`.
- Config: CONFIG_YAML in Infisical (user migration task).
- No new Go dependency.

## Non-goals

- Webhook route registration (owned by bot-command-registry).
- Telegram client changes and the dependency-report message catalog (refactor-dependency-report, http-clients-hardening).
- Moving email message templates out of config (they already live there).
- Changing IMAP search, the mark-as-read default or the template format.
- Replacing go-imap or adding a dependency.
