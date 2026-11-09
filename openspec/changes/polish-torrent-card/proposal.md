# Proposal

## Why

The first enriched qBittorrent card works but looks plain: OMDb poster URLs carry a 300 px thumbnail that Telegram upscales into a blurry image, the caption is a short list of lines, and the IMDb link is a loose word at the bottom. OMDb already returns more data than the card shows.

## What Changes

- **Sharper poster:** OMDb poster URLs end in `_SX300.jpg`; the hub swaps that width for `metadata.poster_width` (default `1000`; `0` keeps the original URL). A URL that does not match the pattern is used unchanged. If Telegram rejects the resized URL, the original URL is tried before the text fallback.
- **Richer caption:** title becomes the IMDb link; adds IMDb votes, Metacritic, director, main actors and the plot in `<blockquote expandable>` (`metadata.plot_max_chars`, default `200`, `0` hides the plot); one emoji per field. Display order: title, ratings, genre/runtime/rating, tags, director, cast, plot. Under the 1024 limit, lines drop from the end (plot first).
- **Inline button:** a URL button "Open on IMDb" under the photo and under the text fallback.
- All texts, the IMDb URL pattern and the button label come from the message catalog. The loose `torrent.card.link` key is replaced by `imdb_url`, `title_linked` and `button`. **BREAKING** only for a `messages:` override of `torrent.card.link` (none in the shipped config).

## Capabilities

### New Capabilities

None.

### Modified Capabilities
- `media-metadata-lookup`: returns plot, director, actors, IMDb votes and Metacritic; the poster URL is resized; new `poster_width` and `plot_max_chars` settings.
- `torrent-webhook-notification`: new caption layout, caption trimming order (lines, then title text; the title keeps its link), and the link button.

## Impact

- Code: `internal/services/media` (Metadata, OMDb client, poster resizing), `internal/services/processor` (card presenter, processor), `internal/services/telegram` (inline URL button on photo and text sends), `internal/config` (two settings and validation), `internal/messages/default.yaml`, `README.md`, `configs/config.yaml.example`.
- Dependencies: none. External or cross-repo steps: none.
- Each function stays at cognitive complexity 15 or less (CI gate).

## Non-goals

- No TMDB or other providers; no image download, resize or cache by the hub.
- No payload, route or fallback-to-template changes.
- No episode titles or trailers.
