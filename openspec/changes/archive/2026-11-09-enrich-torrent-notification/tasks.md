# Tasks

## 1. Release name parser

- [x] 1.1 Add `internal/services/media` with `Release` and `ParseRelease(name)` (title, year, `SxxExx`, ordered quality tags, extension dropped), with the quality-token list as a package variable. Verify: table test `release_test.go` covers the four spec scenarios plus a name with underscores and a name with two years (`Blade.Runner.2049.2017.2160p` → title `Blade Runner 2049`, year `2017`).

## 2. Provider and OMDb client

- [x] 2.1 Add `Metadata`, the `Provider` interface and `ErrNotFound`. Verify: compiles; used by 2.2 and 3.x tests.
- [x] 2.2 Implement the OMDb client (query by `t`/`y`/`apikey`, `N/A` → empty, Rotten Tomatoes from `Ratings`, capped body via `httpx.Send`, errors without the URL). Verify: `httptest` table test for found, missing fields, `Response: False`, `401`, invalid JSON, timeout; assert the key never appears in the returned error.
- [x] 2.3 Add a provider factory keyed on `provider` (`omdb`). Verify: test for `omdb` and an unknown name.

## 3. Telegram photo

- [x] 3.1 Add `Client.SendPhoto(chatID, photoURL, caption)` (HTML caption, URL source, same retry path; nil/uninitialized client sends nothing). Verify: telegram package test against the existing fake API server asserts `sendPhoto` with `photo` URL and `caption`.

## 4. Config and messages

- [x] 4.1 Add `MetadataConfig` under `WebhookProcessorConfig`, default constants (`base_url`, `5s`) and validation in `validateHooks` naming `hook.<name>.config.metadata.<key>`. Verify: config tests for absent block, missing `api_key`, unknown `provider`, negative `timeout`, defaults applied.
- [x] 4.2 Add `torrent.card.*` keys to `internal/messages/default.yaml` (title line, ratings, genre/runtime line, tags, link, labels). Verify: `catalog_test.go` passes and an override of one key is honored.

## 5. Notification flow

- [x] 5.1 Add `CardPresenter` (caption from catalog, HTML escaping, omit absent values, 1024 cap, no save path). Verify: tests for full card, missing fields, hostile title `<b>Tom & Jerry`, over-long caption.
- [x] 5.2 Update `TorrentProcessor`: parse → lookup → photo → text fallback → template fallback; nil provider keeps today's path. Verify: `torrent_test.go` cases: no provider (existing test unchanged), not found, lookup error, no poster, photo rejected then text ok, both rejected returns error, full card.
- [x] 5.3 Wire `newTorrentRoute` and `NewWebhookHandler` to build the provider once from `hook.Config.Metadata`. Verify: `webhook_test.go` end-to-end with `httptest` OMDb and fake sender: `200` for the full card and for the fallback, `500` when sends fail, `404` unchanged when the hook is missing.
- [x] 5.4 Enforce the 1024 limit in CardPresenter when a catalog override is long (trim optional lines, then the link line, then the title). Verify: regression test in card_test.go with a 1200-character torrent.card.link override asserts a caption of at most 1024 runes.

## 6. Docs and cross-repo

- [x] 6.1 Document the `metadata` block and the new message in `README.md` and `configs/config.yaml.example` (commented block, defaults shown). Verify: `git diff` shows both files and the example parses in the config test that loads it, if one exists.
- [x] 6.2 (user) Create the OMDb API key and add it to Infisical and `CONFIG_YAML`.
- [x] 6.3 (homelab) Allow egress from `internal-apps` to `www.omdbapi.com:443` in the NetworkPolicy. No change needed: `allow-internet-egress` in the homelab baseline already allows TCP 443 to any destination, and `allow-dns-egress` covers DNS.

## 7. Validation

- [x] 7.1 Run `docker run --rm -v "$PWD":/app -w /app golang:$(cat .go-version) sh -c 'go vet ./... && go test -race ./...'`. Verify: exit code 0.
- [x] 7.2 Run `graphify update .` after the code changes. Verify: command succeeds.
