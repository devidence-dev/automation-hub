# torrent-webhook-notification Specification

## MODIFIED Requirements

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

## ADDED Requirements

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
