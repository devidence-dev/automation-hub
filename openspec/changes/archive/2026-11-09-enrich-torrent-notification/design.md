# Design

## Context

Today `WebhookHandler.HandleTorrentComplete` decodes `{torrent_name, save_path}` and builds a `TorrentProcessor` per request, which fills `telegram_message` through `TemplateNotifier` and calls `Client.SendMessage` (Markdown). `newTorrentRoute` in `webhook_registry.go` builds the handler from `RouteEnv`. The Telegram client already has `SendHTML` and `SendDocument` on a shared retrying sender (`send(chattable)`), and the dependency report already uses the "caption, fall back to text" pattern with the 1024-char cap. Texts come from the message catalog; HTTP clients come from `httpx` (see the GitHub client for the config-to-`httpx.Settings` mapping). See proposal.md for motivation.

## Goals / Non-Goals

**Goals:**
- Identified title → one photo message with a rich caption; everything else → today's message, byte for byte.
- A lookup problem never fails the webhook and never delays it past the configured timeout.
- Provider is swappable without touching the processor.

**Non-Goals:**
- Providers other than OMDb, caching, image handling, payload changes (see proposal).

## Decisions

1. **OMDb as the only provider.** One call by title+year returns poster, IMDb rating, Rotten Tomatoes score, genre, runtime and id; the free tier (1,000/day) far exceeds a homelab's downloads. Alternatives: TMDB (better matching, but no IMDb/RT ratings, so a second source would be needed); scraping IMDb or RT (fragile, against terms); the IMDb link preview from the screenshot (zero API, but the card is chosen by Telegram: no RT score, no quality tags, no control over fields).
2. **`sendPhoto` with the poster URL plus caption, not a link preview.** Telegram fetches the URL itself, so the hub moves no image bytes; the caption is ours. Trade-off: Telegram's fetch can fail (URL gone, over 5 MB), so the text fallback is mandatory.
3. **Package `internal/services/media`** with three small parts (single responsibility): `ParseRelease(name) Release` (pure function, table-tested); `Provider` interface (`Lookup(ctx, Release) (*Metadata, error)`, `ErrNotFound`) with the `omdb` implementation (strategy; built by a factory keyed on `provider`, like `routeFactories`); `Metadata` as a plain struct with empty fields for absent values.
4. **Processor depends on consumer-defined interfaces.** `TorrentProcessor` gets a `lookup` (nil when `metadata` is absent) and a `cardSender` (`SendPhoto`, `SendHTML`); `*telegram.Client` satisfies both. Flow: parse → lookup (any error or not found → fallback) → build caption → photo → on error or no poster, `SendHTML(caption)` → on error return it. The fallback is the unchanged `TemplateNotifier`. A small `CardPresenter` builds the caption from the catalog (like `depreport.CaptionPresenter`), escaping every value with `html.EscapeString` and trimming optional lines, then the link line, then the title until it fits 1024 characters.
5. **Telegram `SendPhoto(chatID, photoURL, caption)`** next to `SendDocument`, using `tgbotapi.NewPhoto` with `FileURL` and `ModeHTML` through the same `send`/retry path. 1024 is Telegram's protocol limit, so it is a constant (as the project context allows).
6. **Config under `hook[].config.metadata`**: `provider`, `api_key`, `base_url`, `timeout`, with documented defaults as constants in `config`; `validateHooks` gains the new checks and keeps the existing message format (`hook.<name>.config.metadata.<key> ...`). The HTTP client is built with `httpx.NewClient` and `httpx.Send` (capped body), mapping `timeout` to `Settings.Timeout`. The OMDb client sets only `Settings.Timeout` and relies on httpx/stdlib defaults for dial, TLS and idle settings, because one request per download needs no pooling tuning.
7. **Wiring in `newTorrentRoute`**: it builds the provider from `hook.Config.Metadata` once at registration, passes it to `NewWebhookHandler`, and the handler passes it to the processor. A nil provider means today's path. The handler keeps its fail-closed behavior: processor error → `500`.
8. **Texts in `messages/default.yaml`** under `torrent.card.*` (line templates, labels, separators), overridable under `messages:`. The catalog's key validation (`Keys`) must know the new keys.
9. **Matching rules.** Query `t=<title>` plus `y=<year>` when present, `type` left unset so series match too. A wrong match is possible for ambiguous titles; the caption shows the resolved title and year so it is visible.

## Risks / Trade-offs

- [Wrong title match] → caption shows the matched title, year and IMDb link; year in the query narrows it.
- [qBittorrent waits for the hook] → lookup timeout defaults to `5s`, the fallback is immediate.
- [API key leaks via the URL] → key only in the outbound query string; never logged (errors wrapped without the URL).
- [Telegram cannot fetch the poster] → text fallback with the same caption.
- [OMDb outage or daily quota (1,000)] → falls back to today's message; logged at warn.
- [Release name with unusual tokens parses badly] → the parser is conservative (title ends at year, `SxxExx` or a known quality token); unknown names fall back to the simple message when OMDb finds nothing.
- [New egress] → NetworkPolicy in the homelab repo must allow `www.omdbapi.com:443`; without it every lookup times out and the fallback is used.

## Migration Plan

1. Deploy the code with no `metadata` block: behavior unchanged.
2. (user) Create an OMDb key, store it in Infisical, add `metadata` to the `qbittorrent` hook in `CONFIG_YAML`; (homelab) allow the egress.
3. Rollback: remove the `metadata` block (no redeploy of code needed).

## Open Questions

- Whether to show the episode title for series (OMDb needs a second call with `Season`/`Episode`); deferred, series show the show title plus `SxxExx`.
