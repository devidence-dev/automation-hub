# Tasks

Docker commands run in the repo root. Go and tools are never installed on the host.

## 1. Config package

- [x] 1.1 Add the documented defaults (`DefaultPollingInterval` 60, `DefaultCodePattern` `\b[a-zA-Z0-9]{4,8}\b`, `DefaultNotFoundText` `Not found`, `DefaultLogPreviewLength` 200, `DefaultWebhookMaxBodyBytes` 1048576) and the fields `Extractor`, `CodeMarkers`, `CodeCandidates`, `NotFoundText`, `StripMIMEHeaders *bool` on `ServiceProcessorConfig`, plus `LogPreviewLength` on `EmailConfig` and `MaxBodyBytes` on `WebhookProcessorConfig`. Add the accessors `Effective*` that return the default when unset. Verify with `TestEffectiveDefaults` (table: zero values give 60, `\b[a-zA-Z0-9]{4,8}\b`, `Not found`, 200, 1048576, `strip_mime_headers` true; set values are returned unchanged).
- [x] 1.2 Validate in `Load`: negative `polling_interval`, negative `log_preview_length`, negative `hook[].config.max_body_bytes`. Verify with `TestLoad_NegativePollingInterval` and `TestLoad_NegativeMaxBodyBytes` (each error names the key), the table `TestValidateEmailServices` case `negative log preview length names the key`, and `TestLoad_EmailDefaults` (no keys: poll 60 s, preview 200).
- [x] 1.3 Validate email services in `Load`: `email_from` must be non-empty on every service (an empty value matches every sender); `extractor` must be `pattern` (or empty) or `marker`; `marker` requires non-empty `code_markers`; `code_pattern` and every `code_candidates` entry must compile with `regexp.Compile`. Verify with the table `TestValidateEmailServices` (missing or empty `email_from` names service `github`; invalid pattern `[invalid regex (` names service `github`; `extractor: regex` names `regex`; marker without markers names `perplexity`; the minimal service from the spec loads).
- [x] 1.4 Remove the webhook lookup from `processor.GetWebhookConfig`. The lookup now happens at route registration (`bot-command-registry`), so `config` has no hook lookup function. Verify with `grep -rn FindWebhook --include='*.go' .` (no output) and the missing-hook route test of task 4.1.
- [x] 1.5 Update `configs/config.yaml.example`: under `cloudflare` add `strip_mime_headers: false` and `code_pattern: "\\b\\d{6}\\b"`; under `perplexity` add the marker keys from task 6.1; add commented examples for `not_found_text`, `log_preview_length`, `max_body_bytes`. Verify with `TestLoad_MigratedEmailServices`, which writes the exact YAML of task 6.1 into a temp `config.yaml` and checks the resolved extractor settings for both services, and `TestLoad_ExampleEmailServices`, which loads the shipped example and checks the same keys.

## 2. Processor package

- [x] 2.1 Add `CodeExtractor` (`Extract(text) (code string, found bool)`), `patternExtractor` (capture group 1 when non-empty, else whole match) and `newCodeExtractor(cfg, logPreviewLength, logger)` that selects by `extractor`. The regexes are compiled here, once. Verify with `TestPatternExtractor` (table: `[^#]\b(\d{6})\b` on `color:#202123; padding: 0\nIngresa este codigo:\n662575\n` gives `662575`, `\b\d{6}\b` on `Code: 123456` gives `123456`, default pattern on `Use 4821 now` gives `4821`, no match gives not found), and keep `TestExtractCode_IgnoresHexColor` with its body and assertion unchanged.
- [x] 2.2 Add `markerExtractor`: markers in configured order (case-insensitive), then candidates in order on the text after the marker, then `code_pattern` fallback. Verify with `TestMarkerExtractor` (table: `Inicie sesión directamente:\n123456` gives `123456`; `Log in directly:\naw9s5-y1zoy` gives `aw9s5-y1zoy`; `directamente: 111111 ... directly: 222222` gives `222222`; `Welcome to Perplexity. Click here to confirm.` gives not found; `Log in directly: !!!` gives not found). Port the Perplexity cases of `TestExtractPerplexityCode` to this table.
- [x] 2.3 Add the consumer interface `messageSender { SendMessage(chatID, message string) error }` and `TemplateNotifier{sender, chatID, template}` with `Notify(args ...any) error`. Verify with `TestTemplateNotifier` using a fake sender that records chat ID and text: one argument for the email template, two for the torrent template.
- [x] 2.4 Rewrite `GenericEmailProcessor` as a composition: matcher (existing `ShouldProcess` semantics), body preparation (quoted-printable decode, then MIME header strip unless `strip_mime_headers` is false), `CodeExtractor`, not-found substitution with the warning capped at `log_preview_length`, and `TemplateNotifier`. Constructor is `NewGenericEmailProcessor(name, cfg, sender messageSender, logger) (*GenericEmailProcessor, error)`. Verify with `TestGenericEmailProcessor_Process` (table: the Cloudflare example `Use this code directly: 123456` produces the Cloudflare template with `123456` in its code block through a fake sender; configured `not_found_text` `No code in email` is sent; sender error is returned) and keep `TestShouldProcess`, `TestStripMIMEHeaders`, `TestDecodeQuotedPrintable`, `TestTruncateString` passing.
- [x] 2.5 Confirm there is no name-keyed behavior: remove `defaultPatterns`, the `cloudflare`/`perplexity` branches and the `NotFoundCode` constant. Verify with `TestNoNameDefaults`: a service named `cloudflare` with no keys, given the text `Subject: Code 111111\n\nYour code is 222222`, strips the headers and applies the generic pattern, so it extracts `Your` (the first 4-8 character word), not `222222`. Rewrite `TestNewGenericEmailProcessor_BuiltInPatterns` and the Cloudflare case of `TestExtractCode` with explicit config; expectations unchanged.
- [x] 2.6 Add a log test for the preview cap: `TestPatternExtractor_LogPreviewCapped` uses `zaptest/observer` (part of the existing zap module) with `log_preview_length` 10 and checks that the warning carries at most 10 characters plus `...`.
- [x] 2.7 Change `Manager`: `NewProcessorManager(emailCfg, sender messageSender, logger) (*Manager, error)`, `Senders() []string`, `Dispatch(email) (models.EmailProcessor, error)` (first match in config order, logs the outcome, returns the matched processor or nil). Delete `ProcessEmailsConcurrently` and `processEmailAsync`. Verify with `TestManagerDispatch` (first matching service handles the email, the second is not called, unmatched returns nil and no error, a failing Process returns its error) and `TestManagerSenders`. Rewrite `TestProcessorManager` and `TestProcessorManager_CanceledContextAsync` on these, deleting the context-cancel case.
- [x] 2.8 Change `TorrentProcessor` to `NewTorrentProcessor(sender messageSender, cfg *config.WebhookProcessorConfig, logger)` built on `TemplateNotifier`, and delete `NewTorrentProcessorLegacy` with its hardcoded Markdown template. Verify with `TestTorrentProcessor_Process` (template `Download finished: %s at %s` with `Dune.2021.1080p` and `/downloads/movies` gives the expected text). Delete `TestNewTorrentProcessorLegacy`; update `TestNewTorrentProcessor` to the new fields.

## 3. Email package

- [x] 3.1 Add the `ReadMarker` named interface (`ShouldMarkAsRead() bool`) to `internal/models`, and replace the anonymous assertion in the IMAP code. Verify with `TestPollerMarkAsReadPolicy` (task 3.3).
- [x] 3.2 Split the go-imap code into `IMAPSource` implementing `MailSource { Unread(senders, handle func(models.Email) bool) error }`, behind a small `imapSession` interface that `*client.Client` satisfies. Delete `markAsUnread`. Fix the nil envelope check in `parseMessage`. Log per-sender search failures and return an error when every sender search fails. Verify with `TestUnreadOneSenderFails` and `TestUnreadAllSendersFail` (fake session), `TestParseMessageNilEnvelope` (body `123456`, no panic, empty subject and sender), and keep `TestNewIMAPClient`, `TestExtractTextPlain` and `TestMarkAsReadNilClient` passing under the new names.
- [x] 3.3 Add `Poller{source MailSource, dispatcher Dispatcher, interval time.Duration, logger}` with `Dispatcher` (`Senders`, `Dispatch`) as a consumer interface. `Run(ctx)` polls on a ticker and logs each returned error once. Mark-as-read: flag only when the matched processor is a `ReadMarker` returning true and dispatch succeeded. Verify with `TestPollerMarkAsReadPolicy` (table: `mark_as_read` true marks, false does not, non-marker does not, dispatch error does not, no match does not), `TestPollerLogsPollError` (observer, one error entry with the host), and `TestPollerStopsOnCancel`.
- [x] 3.4 Rewire `cmd/automation-hub/main.go`: handle the error from `NewProcessorManager`, build `IMAPSource` and `Poller` with `EffectivePollingInterval()`, and start `Poller.Run` in place of `StartMonitoring`. This is wiring only; `cmd/` is excluded from coverage, and task 7.4 proves it builds.
- [x] 3.5 Update `README.md` email section: the new keys (`extractor`, `code_markers`, `code_candidates`, `strip_mime_headers`, `not_found_text`, `log_preview_length`), the defaults, and the note that invalid regexes now stop startup. Verify by reading the section against the spec (no behavior).

## 4. Torrent webhook

- [x] 4.1 In `handlers/webhook.go`, read the cap from the hook config the handler receives at registration (`bot-command-registry` moved the lookup there, so the cap is known before decoding), wrap `r.Body` with `http.MaxBytesReader(w, r.Body, EffectiveMaxBodyBytes())`, and map `*http.MaxBytesError` to `413`. Verify with `TestTorrentCompleteBodyCap` (table: default cap accepts 1048576 bytes and rejects 1048577 with `413`; `max_body_bytes: 1024` rejects 2048 bytes; in every rejected case the fake sender records no message) and `TestRegisterRoutesQbittorrentMissing` in `cmd/automation-hub/main_test.go` (only `sonarr` configured: a `POST` to `/webhook/qbittorrent` gives `404`). Existing webhook tests stay green.
- [x] 4.2 Add the body-cap section to `README.md` next to the qBittorrent hook. Verify by reading it against `specs/torrent-webhook-notification/spec.md`.

## 5. Quality gates

- [x] 5.1 Before using them, confirm the tool versions exist in the image: `docker run --rm golang:$(cat .go-version) sh -c 'go list -m -versions github.com/uudashr/gocognit'` must list `v1.2.0`, and `docker run --rm golangci/golangci-lint:v2.14.0 golangci-lint version` must print v2.14.0. If either is missing, stop and report it to the user.
- [x] 5.2 Run gocognit v1.2.0 in Docker over the changed packages: `docker run --rm -v "$PWD":/app -w /app golang:$(cat .go-version) sh -c 'go run github.com/uudashr/gocognit/cmd/gocognit@v1.2.0 -over 15 ./internal ./cmd'` must print nothing. Test functions count too. Verify by an empty output, and fix any function above 15 by splitting it, not by raising the limit.
- [x] 5.3 Run golangci-lint v2.14.0 in Docker: `docker run --rm -v "$PWD":/app -w /app golangci/golangci-lint:v2.14.0 golangci-lint run --enable=errcheck,govet,staticcheck,unused,misspell,gosec ./...`. Verify zero issues; no `//nolint` added unless the user approves it.

## 6. User and other-repo steps

- [x] 6.1 (user, applied in Infisical 2026-10-12) Before deploying this version, add the following to the Infisical `CONFIG_YAML` `email.services` entries. Keep every other key. Service names are matched case-insensitively today, as the code did, so if your entry is named differently, apply the same keys to it. Existing `code_pattern` values stay as they are.

```yaml
email:
  services:
    - name: "cloudflare"
      config:
        # ...existing keys unchanged...
        strip_mime_headers: false
        code_pattern: "\\b\\d{6}\\b"
    - name: "perplexity"
      config:
        # ...existing keys unchanged...
        extractor: marker
        code_markers: ["directly:", "directamente:"]
        code_candidates: ["\\b(\\d{5,6})\\b", "\\b([a-zA-Z0-9]+-[a-zA-Z0-9]+)\\b"]
        code_pattern: "(?:\\d{5,6}|[a-zA-Z0-9]+-[a-zA-Z0-9]+)"
```

The current binary ignores unknown keys, so this can be applied before the deploy with no behavior change. Verify after the deploy by checking the startup log of each service (task 6.2).

- [x] 6.2 (user) After deploying, check the startup log: each "Loaded email processor" entry shows the extractor, and neither cloudflare nor perplexity logs the generic default pattern warning. Send one test Cloudflare code and one Perplexity code and confirm the forwarded values match the previous behavior. Roll back by redeploying the previous image; the migrated keys are harmless to it.

- [x] 6.3 Check the overlap with `bot-command-registry`, resolved on 2026-10-11: this change owns `NewTorrentProcessor`, its `messageSender` interface and `GetWebhookConfig`, and `bot-command-registry` lands first without touching them. Rebase on it, keep its registration-time hook config in `NewWebhookHandler`, and verify by checking that `NewTorrentProcessor` has exactly one interface parameter after both changes are merged, that `go vet` passes (task 7.1), and with `TestHandleTorrentCompleteProcessingError` in `webhook_test.go`: a fake `messageSender` that fails makes the handler answer 500 with `Processing failed` (moved here from `bot-command-registry` task 6.1).

## 7. Final validation

- [x] 7.1 Run the full gate in Docker and fix any failure before finishing: `docker run --rm -v "$PWD":/app -w /app golang:$(cat .go-version) sh -c 'go vet ./... && go test -race ./...'` must pass. Then run tasks 5.2 and 5.3 again, and `openspec validate --strict email-processing-cleanup` must report valid. Verify with all of these commands passing on the final code.
