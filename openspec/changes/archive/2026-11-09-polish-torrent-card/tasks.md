# Tasks

## 1. Metadata and poster

- [x] 1.1 Add `Plot, Director, Actors, IMDbVotes, Metacritic, PosterFallback` to `media.Metadata`, parse them in the OMDb client (`N/A` is empty; Metacritic from `Ratings` through one shared helper). Verify: `omdb_test.go` table cases for found, missing fields, `Plot: N/A`.
- [x] 1.2 Add `ResizePoster(url, width)` and use it in the OMDb client (the factory and `NewOMDb` take `posterWidth`). Verify: `poster_test.go` for resized, unknown pattern, width `0`, empty URL; `omdb_test.go` asserts poster and fallback.

- [x] 1.3 Fix `ResizePoster` to replace everything between `._V1_` and the file extension, so `_QL75_UX380_CR0,0,380,562_` URLs are resized too. Verify: `poster_test.go` cases for `_SX300`, `_QL75_UX380_CR0,0,380,562_`, a URL without `._V1_`, width `0`, and the real URL `https://m.media-amazon.com/images/M/MV5BMDg5NDFkNTctZWE2ZC00MDYzLThlYjItZTljZDY3NTcwNjA3XkEyXkFqcGc@._V1_QL75_UX380_CR0,0,380,562_.jpg` -> `..._V1_SX1000.jpg`.

## 2. Configuration

- [x] 2.1 Add `PosterWidth` and `PlotMaxChars` (`*int`) to `MetadataConfig`, default constants, defaults in `applyMetadataDefaults`, negative check in `validateMetadataFields` naming `hook.<name>.config.metadata.<key>`. Verify: `metadata_test.go` cases for defaults, explicit `0`, negative width, negative plot length.
- [x] 2.2 Pass the settings through `media.NewProvider` and `ConfigureMetadata`/the route factory. Verify: `webhook_test.go` end-to-end with `httptest` OMDb checks the requested poster URL width.

## 3. Telegram

- [x] 3.1 Add `LinkButton`, extend `SendPhoto` with `...LinkButton`, add `SendHTMLButtons`. Verify: telegram tests against the fake API assert `reply_markup` with one URL button on `sendPhoto` and `sendMessage`, and none when no button is passed.

## 4. Messages

- [x] 4.1 In `default.yaml` replace `torrent.card.link` with `imdb_url`, `title_linked`, `button`, `plot_truncated`; add emoji labels for votes, Metacritic, director, cast and plot (`plot` wraps `<blockquote expandable>`). Verify: `catalog_test.go` and the keys test pass; an override of `button` is honored.

## 5. Presenter and processor

- [x] 5.1 Update `CardPresenter`: new line order, linked title, plot truncation by `PlotMaxChars`, remove the link line, return `Card{Caption, URL}`. Keep every function at complexity 15 or less. Verify: `card_test.go` for the full card, missing fields, plot `0`, plot truncation, hostile title, long caption drops plot first, overlong `imdb_url` override.
- [x] 5.2 Update `TorrentProcessor.sendCard`: poster attempts (resized, original, text), button on photo and on text. Verify: `torrent_test.go` for resized rejected then original accepted, no poster, all rejected then text, both sends fail returns an error, button present on both paths, no button without an IMDb id.

## 6. Docs

- [x] 6.1 Document `poster_width`, `plot_max_chars`, the card layout, the button and the removed `torrent.card.link` key in `README.md` and `configs/config.yaml.example`. Verify: `git diff` shows both files and the example still parses in the config tests.

## 7. Validation

- [x] 7.1 Run `docker run --rm -v "$PWD":/app -w /app golang:$(cat .go-version) sh -c 'go vet ./... && go test -race ./...'`. Verify: exit code 0.
- [x] 7.2 Run gocognit as CI does, in Docker: `go run github.com/uudashr/gocognit/cmd/gocognit@v1.2.0 -over 15 .` and ignore `*_templ.go`. Verify: no function listed outside generated files.
- [x] 7.3 Run `graphify update .`. Verify: command succeeds.
