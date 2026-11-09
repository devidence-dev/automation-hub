# torrent-webhook-notification Specification

## Purpose
Forwards qBittorrent download-completion webhooks to Telegram. The hub reads the torrent name and save path from a JSON body, fills the hook's message template and sends it to the hook's chat. Covers the body size cap and the hook configuration lookup. Route registration is owned by the bot-command-registry change.

## Requirements

### Requirement: Completion notification
The `qbittorrent` hook SHALL accept a JSON body with `torrent_name` and `save_path`. When the title is not identified (no `metadata` block, release not identifiable, lookup not found or failed), it MUST fill its `telegram_message` with the name as the first `%s` and the path as the second, and send the result to its `telegram_chat_id`. On success the response MUST be `200` with body `{"status":"success"}`. A body that is not valid JSON MUST get `400` with `Invalid request` and send nothing.

#### Scenario: Completed download
- **WHEN** the body is `{"torrent_name":"Dune.2021.1080p","save_path":"/downloads/movies"}`, the hook has no `metadata` block and the template is `Download finished: %s at %s`
- **THEN** the chat receives `Download finished: Dune.2021.1080p at /downloads/movies` and the response is `200` with `{"status":"success"}`

#### Scenario: Invalid JSON
- **WHEN** the body is `{"torrent_name":`
- **THEN** the response is `400` with `Invalid request` and no Telegram message is sent

#### Scenario: Telegram failure
- **WHEN** the Telegram send fails
- **THEN** the response is `500` with `Processing failed` and the failure is logged

#### Scenario: Title not identified
- **WHEN** the hook has a `metadata` block, OMDb answers `Movie not found!` for `Unknown.Thing.2031.1080p`
- **THEN** the chat receives the filled `telegram_message` and the response is `200`

### Requirement: Body size cap
The request body SHALL be read up to the hook's `max_body_bytes` (under `hook[].config`). When the key is absent or `0`, the cap MUST be 1048576 bytes (1 MiB). A body larger than the cap MUST get `413` and send nothing to Telegram. A negative `max_body_bytes` MUST fail config loading with an error that names `hook.qbittorrent.config.max_body_bytes`.

#### Scenario: Body within the default cap
- **WHEN** the hook has no `max_body_bytes` and the JSON body is exactly 1048576 bytes
- **THEN** the body is accepted and processed normally

#### Scenario: Body over the default cap
- **WHEN** the hook has no `max_body_bytes` and the JSON body is 1048577 bytes
- **THEN** the response is `413` and no Telegram message is sent

#### Scenario: Configured cap
- **WHEN** the hook has `max_body_bytes: 1024` and the body is 2048 bytes
- **THEN** the response is `413` and no Telegram message is sent

#### Scenario: Negative cap
- **WHEN** the hook has `max_body_bytes: -1`
- **THEN** config loading fails with an error that names `hook.qbittorrent.config.max_body_bytes`

### Requirement: Hook configuration lookup
The `qbittorrent` hook configuration SHALL be resolved once, when the route is registered, and handed to the handler, so the body cap is known before the body is read. When the `hook:` list has no `qbittorrent` entry, no route is mounted and a `POST` to its path MUST get `404`, regardless of the body.

#### Scenario: Hook missing
- **WHEN** the config has only a `sonarr` hook and a `POST` arrives at `/webhook/qbittorrent` with an invalid body
- **THEN** the response is `404` and no Telegram message is sent

### Requirement: Enriched notification
When the lookup identifies the title, the hub SHALL send ONE photo message whose source is the poster URL (the hub MUST NOT download the image). The HTML caption holds, one per line in this order: the title linked to `https://www.imdb.com/title/<id>/` with year (and `SxxExx` for a series); ratings (IMDb with votes, Rotten Tomatoes, Metacritic); genre, runtime and content rating; quality tags; director; actors; and the plot in `<blockquote expandable>`, cut to `plot_max_chars` with an ellipsis. Each field has its own emoji. Absent values are omitted with their label, the save path is never shown, and every text, including the IMDb URL pattern, comes from the message catalog.

#### Scenario: Full card
- **WHEN** the body is `{"torrent_name":"Coyote.Vs.Acme.2026.WEB-DL.4k.HDR-Dual-Lat","save_path":"/downloads/complete/Coyote.Vs.Acme.2026.WEB-DL.4k.HDR-Dual-Lat"}` and OMDb identifies `tt1756855` with rating `7.6`, votes `12,345`, Metacritic `62/100`, director `Dave Green` and a 400-character plot
- **THEN** the chat receives one photo whose caption starts with the title linked to `https://www.imdb.com/title/tt1756855/`, contains `7.6`, `12,345`, `62/100`, `Dave Green`, `4k`, `HDR` and `WEB-DL`, holds a plot of at most 200 characters inside `<blockquote expandable>`, and does not contain `/downloads/complete`

#### Scenario: Hostile title
- **WHEN** OMDb returns the title `<b>Tom & Jerry`
- **THEN** the caption shows it as text (`&lt;b&gt;Tom &amp; Jerry`), not as markup

#### Scenario: Plot hidden
- **WHEN** `plot_max_chars` is `0`
- **THEN** the caption has no `<blockquote` element

#### Scenario: Long caption
- **WHEN** the provider values make the caption exceed 1024 characters
- **THEN** the sent caption is at most 1024 characters, and lines are dropped from the end (plot first, then cast, director, tags) while the title line stays

### Requirement: Caption safety and delivery fallback
The caption MUST escape HTML in every provider value and MUST stay within 1024 characters, trimming in this order until it fits: lines from the end, then the title text (the title keeps its IMDb link). The limit MUST hold even when a catalog override makes a card text long. The poster is tried at its resized URL, then at the original URL when different; if there is no poster or every attempt fails, the same caption is sent as a text message. The response is `200` once any of these sends succeeds.

#### Scenario: No poster
- **WHEN** OMDb identifies the title with `Poster: N/A`
- **THEN** the chat receives the caption as a text message and no photo

#### Scenario: Resized poster rejected
- **WHEN** Telegram rejects the `_SX1000` poster URL and accepts the original `_SX300` URL
- **THEN** the chat receives the photo from the original URL with the same caption, and no text message

#### Scenario: Photo rejected
- **WHEN** Telegram rejects every poster URL and accepts the text send
- **THEN** the chat receives the caption as text and the response is `200`

#### Scenario: Both sends fail
- **WHEN** Telegram rejects the photo and the text fallback
- **THEN** the response is `500` with `Processing failed`

#### Scenario: Overlong override
- **WHEN** the catalog overrides `torrent.card.imdb_url` with a 1200-character text
- **THEN** the sent caption is at most 1024 characters and no error is returned

### Requirement: IMDb link button
Every enriched message, photo or text fallback, SHALL carry one inline URL button that opens `https://www.imdb.com/title/<id>/`. The label and the URL pattern come from the message catalog (default label `Open on IMDb`). A title without an IMDb id gets no button.

#### Scenario: Button on photo
- **WHEN** OMDb identifies `tt1756855` and the photo send succeeds
- **THEN** the photo message has one inline button labelled `Open on IMDb` that opens `https://www.imdb.com/title/tt1756855/`

#### Scenario: Button on text fallback
- **WHEN** the poster is missing and the caption is sent as text
- **THEN** the text message carries the same button

#### Scenario: Label override
- **WHEN** `messages.torrent.card.button` is `Ver en IMDb`
- **THEN** the button label is `Ver en IMDb`
