# torrent-webhook-notification Specification

## MODIFIED Requirements

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

## ADDED Requirements

### Requirement: Enriched notification
When the lookup identifies the title, the hub SHALL send ONE photo message whose source is the poster URL (the hub MUST NOT download the image). The HTML caption holds the title with year (and `SxxExx` for a series), IMDb rating, Rotten Tomatoes score, genre, runtime, quality tags and the link `https://www.imdb.com/title/<id>/`. Absent values are omitted with their label, the save path is never shown, and the texts come from the message catalog.

#### Scenario: Full card
- **WHEN** the body is `{"torrent_name":"Coyote.Vs.Acme.2026.WEB-DL.4k.HDR-Dual-Lat","save_path":"/downloads/complete/Coyote.Vs.Acme.2026.WEB-DL.4k.HDR-Dual-Lat"}` and OMDb identifies `tt1756855` with rating `7.3`
- **THEN** the chat receives one photo with the OMDb poster URL, and a caption that contains `Coyote vs. Acme (2026)`, `7.3`, `85%`, `4k`, `HDR`, `WEB-DL` and `https://www.imdb.com/title/tt1756855/`, and does not contain `/downloads/complete`

#### Scenario: Hostile title
- **WHEN** OMDb returns the title `<b>Tom & Jerry`
- **THEN** the caption shows it as text (`&lt;b&gt;Tom &amp; Jerry`), not as markup

#### Scenario: Long caption
- **WHEN** the provider values make the caption exceed 1024 characters
- **THEN** the sent caption is at most 1024 characters; the title and IMDb link are retained when they fit after removing optional lines, and trimmed in the specified order otherwise

### Requirement: Caption safety and delivery fallback
The caption MUST escape HTML in every provider value and MUST stay within 1024 characters, trimming in this order until it fits: optional lines, then the IMDb link line, then the title. The limit MUST hold even when a catalog override makes a card text long. If there is no poster, or Telegram rejects the photo, the same caption MUST be sent as a text message. The response is `200` once any of these sends succeeds.

#### Scenario: No poster
- **WHEN** OMDb identifies the title with `Poster: N/A`
- **THEN** the chat receives the caption as a text message and no photo

#### Scenario: Photo rejected
- **WHEN** Telegram answers an error to the photo send and accepts the text send
- **THEN** the chat receives the caption as text and the response is `200`

#### Scenario: Both sends fail
- **WHEN** Telegram rejects the photo and the text fallback
- **THEN** the response is `500` with `Processing failed`

#### Scenario: Overlong override
- **WHEN** the catalog overrides `torrent.card.link` with a 1200-character text
- **THEN** the sent caption is at most 1024 characters and no error is returned
