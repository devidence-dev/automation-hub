# Proposal

## Why

The qBittorrent completion message only echoes the release name and save path, which says nothing the user cannot already see. The hub can identify the title from the release name and send a richer message: poster, ratings, genre and runtime, plus the quality tags already in the name.

## What Changes

- Parse the release name (`Coyote.Vs.Acme.2026.WEB-DL.4k.HDR-Dual-Lat`) into title, year, optional season/episode and quality tags.
- Look the title up in OMDb (one call returns poster, IMDb rating, Rotten Tomatoes score, genre, runtime, rating and IMDb id).
- Send one Telegram photo message: the poster is passed to Telegram as a URL (the hub downloads no image), with an HTML caption: title, year, ⭐ IMDb, 🍅 Rotten Tomatoes, genre, runtime, tags and an IMDb link. The save path is dropped from this message.
- Fallbacks, in order: poster missing or rejected by Telegram → caption as text; title not identified or lookup fails → today's simple `telegram_message` template, unchanged.
- New optional `metadata:` block under `hook[].config` (provider, API key, base URL, timeout). Absent block → behavior identical to today.
- All caption texts come from the message catalog, overridable under `messages:`.

## Capabilities

### New Capabilities
- `media-metadata-lookup`: release-name parsing, provider lookup of title metadata, its configuration and failure isolation.

### Modified Capabilities
- `torrent-webhook-notification`: the notification becomes a metadata card when the title is identified; the template message becomes the fallback.

## Impact

- Code: new `internal/services/media` package (parser, provider interface, OMDb client); `internal/services/telegram` gains `SendPhoto`; `internal/services/processor` TorrentProcessor, `internal/handlers/webhook_registry.go` wiring, `internal/config` (`metadata` block and validation), `internal/messages/default.yaml`.
- Dependencies: none (stdlib and the existing httpx/telegram stack).
- External: OMDb account and free API key (1,000 calls/day) (user); egress from `internal-apps` to `www.omdbapi.com` (homelab NetworkPolicy); API key into Infisical/`CONFIG_YAML` (user).
- Docs: README and `configs/config.yaml.example`.

## Non-goals

- No other providers (TMDB, TVDB) now; only the interface leaves room for them.
- No image download, resize or caching by the hub; no response cache.
- No plot, cast or trailer in the caption.
- No new fields in the qBittorrent payload (size, category); the body stays `torrent_name` and `save_path`.
- No change to the other webhooks, email forwarding or the dependency report.
