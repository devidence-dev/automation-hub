# Design

## Context

The archived change `enrich-torrent-notification` ships: `media.Metadata` (flat struct), the OMDb client (`internal/services/media/omdb.go`), `CardPresenter` and `TorrentProcessor.sendCard` (`internal/services/processor`), `Client.SendPhoto` and `SendHTML` (`internal/services/telegram`), `MetadataConfig` with `applyMetadataDefaults` (`internal/config`), and the `torrent.card.*` catalog keys. `CardPresenter` builds a `card{title, lines, link}`, drops trailing lines, then the link, then truncates the title. CI enforces cognitive complexity 15 per function (gocognit), generated `*_templ.go` excepted. See proposal.md for motivation.

## Goals / Non-Goals

**Goals:**
- A sharper poster, a fuller caption and a link button, all configurable or catalog-driven, with the existing fallbacks intact.

**Non-Goals:**
- Other providers, image handling, payload changes (see proposal).

## Decisions

1. **Poster resizing lives in the media package.** A small pure function `ResizePoster(url string, width int) (resized, original string)` replaces the whole transform segment between `._V1_` and the extension with `SX<width>` (compiled regex, package variable). OMDb serves both `_SX300` and `_QL75_UX380_CR0,0,380,562_` forms; checked against Amazon, `_SX1000` returns the full 1000x1481 image and the bare `_V1_` URL the 4050x6000 original, so the width is the only knob needed. The OMDb client calls it with `poster_width` and fills `Metadata.Poster` (resized) and `Metadata.PosterFallback` (original, empty when nothing changed). Alternative: resize in the processor; rejected, because poster URL shape is a provider detail.
2. **Poster attempts in the processor.** `sendCard` tries `Poster`, then `PosterFallback`, then the text send. The Amazon width trick is undocumented, so a rejected resized URL must not cost the poster. A table of attempts (`posterURLs(m) []string`) keeps the loop flat.
3. **Optional numeric settings are `*int`.** `poster_width` and `plot_max_chars` need "absent = default, 0 = off". A pointer in `MetadataConfig` expresses that; `applyMetadataDefaults` fills nil with constants (`DefaultMetadataPosterWidth = 1000`, `DefaultMetadataPlotMaxChars = 200`); `validateMetadataFields` rejects negatives. The factory passes plain ints onward. Alternative: sentinel `-1`; rejected as less readable in YAML.
4. **`Metadata` grows by `Plot, Director, Actors, IMDbVotes, Metacritic, PosterFallback`.** The OMDb response struct gains `Plot, Director, Actors, ImdbVotes`; Metacritic comes from `Ratings` like Rotten Tomatoes (one shared lookup helper replaces the loop). `N/A` handling stays in the existing `clean`.
5. **Plot truncation is the presenter's job, not the provider's.** The provider returns the full plot; `CardPresenter` gets `PlotMaxChars` (rune-based cut at a word boundary where possible, then catalog `plot_truncated: "{text}…"`). The presenter is built with that setting by the processor. Alternative: truncate in the OMDb client; rejected, because it is a presentation concern.
6. **Caption layout stays a list of lines.** `buildCard` appends lines in this order: ratings (joined by the catalog separator), details, tags, director, cast, plot. `dropLines` removes from the end, so the plot goes first. The `link` field of `card` and its drop step are removed; the title is rendered with the catalog key `title_linked` (`<a href="{url}"><b>{title}</b></a>`) when an IMDb id exists, `title` otherwise. `fitTitle` keeps escaping whole prefixes so entities and the anchor are never split; it now also wraps the shortened title in `title_linked`. Because the cap check counts raw characters including tags, the limit is conservative.
7. **The IMDb URL has one source.** Catalog key `torrent.card.imdb_url` (`https://www.imdb.com/title/{id}/`) feeds both the title link and the button. `CardPresenter.Present` returns a `Card{Caption string, URL string}`; the processor builds the button from `URL` and the catalog key `button`. Removed key: `torrent.card.link`.
8. **Telegram button support is additive.** `type LinkButton struct{ Text, URL string }` in the telegram package; `SendPhoto(chatID, photoURL, caption string, buttons ...LinkButton)` and a new `SendHTMLButtons(chatID, message string, buttons ...LinkButton)`; both set `ReplyMarkup` via `tgbotapi.NewInlineKeyboardMarkup` when buttons are given. `SendHTML` is unchanged (used by other code). The processor's `cardSender` interface lists the two methods it needs.
9. **Complexity budget.** Every new function stays under 15: one function per concern (`ResizePoster`, `ratingValue`, `renderPlot`, `renderPeople`, `posterURLs`, `buttonFor`), table-driven tests with helper runners, no nested branching inside tests.

## Risks / Trade-offs

- [Resized URL does not exist on Amazon] → the processor falls back to the original URL, then to text.
- [Telegram rejects `<blockquote expandable>` on an old server] → the whole caption send fails, so the text path would fail too; mitigated by Telegram's current Bot API supporting it (7.4+). If it ever errors, the response is `500`, as for any Telegram failure.
- [Caption is dense] → plot and cast drop first; `plot_max_chars: 0` removes the plot.
- [`torrent.card.link` removed] → an override of that key fails catalog validation at start-up; none exists in the shipped config. Noted in the README.
- [Larger poster, slower Telegram fetch] → Telegram fetches it, not the hub; default 1000 px stays well under the 5 MB limit.

## Migration Plan

1. Deploy the new version; with no `metadata` block nothing changes, with one the defaults (`1000`, `200`) apply.
2. Rollback: set `poster_width: 0` and `plot_max_chars: 0` to get close to the old card without redeploying, or redeploy the previous image.
