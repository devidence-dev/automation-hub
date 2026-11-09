# media-metadata-lookup Specification

## MODIFIED Requirements

### Requirement: OMDb lookup
When the hook has a `metadata` block with `provider: omdb`, the hub SHALL query OMDb by title, adding the year when the name has one, using the configured `api_key`. A positive answer MUST yield: title, year, IMDb id, IMDb rating, IMDb votes, Rotten Tomatoes score, Metacritic score, genre, runtime, content rating, director, actors, plot and poster URL. A field OMDb reports as `N/A` MUST be treated as absent, never shown. The API key MUST NOT appear in logs. Each lookup MUST end within `metadata.timeout` (default `5s`).

#### Scenario: Found
- **WHEN** OMDb answers `Response: True` for `Coyote vs. Acme` with `imdbRating: 7.6`, `imdbVotes: 12,345`, a `Rotten Tomatoes` rating of `85%`, a `Metacritic` rating of `62/100`, director `Dave Green`, actors `Will Forte, Lana Condor, John Cena`, a plot and a poster URL
- **THEN** the lookup returns those values and the IMDb id as `tt1756855`

#### Scenario: Missing fields
- **WHEN** OMDb answers with `Poster: N/A`, `Plot: N/A` and no Rotten Tomatoes or Metacritic rating
- **THEN** the lookup returns the rest, with no poster, no plot and no scores for those two sources

#### Scenario: Not found
- **WHEN** OMDb answers `Response: False` with `Error: Movie not found!`
- **THEN** the lookup reports not found, with no error logged above debug level

#### Scenario: Provider failure
- **WHEN** OMDb answers `401`, answers invalid JSON, or does not answer within `metadata.timeout`
- **THEN** the lookup returns an error that the caller treats as not found, and the failure is logged at warn level without the API key

### Requirement: Metadata configuration
`hook[].config.metadata` SHALL accept `provider`, `api_key`, `base_url` (default `https://www.omdbapi.com`), `timeout` (default `5s`), `poster_width` (default `1000`) and `plot_max_chars` (default `200`). For `poster_width` and `plot_max_chars`, an absent key means the default and `0` means off. When the block is absent, no lookup is ever made. Config loading MUST fail, with an error naming the key, when `provider` is not `omdb`, when `api_key` is empty, or when `timeout`, `poster_width` or `plot_max_chars` is negative.

#### Scenario: Block absent
- **WHEN** the `qbittorrent` hook has no `metadata` block
- **THEN** no request leaves the hub toward any metadata provider

#### Scenario: Missing key
- **WHEN** the block has `provider: omdb` and an empty `api_key`
- **THEN** config loading fails with an error that names `hook.qbittorrent.config.metadata.api_key`

#### Scenario: Unknown provider
- **WHEN** the block has `provider: imdb`
- **THEN** config loading fails with an error that names `hook.qbittorrent.config.metadata.provider`

#### Scenario: Defaults
- **WHEN** the block has only `provider: omdb` and `api_key: k`
- **THEN** the effective `poster_width` is `1000` and `plot_max_chars` is `200`

#### Scenario: Explicit zero
- **WHEN** the block has `poster_width: 0` and `plot_max_chars: 0`
- **THEN** the effective values are `0`, so the poster is not resized and no plot is shown

#### Scenario: Negative width
- **WHEN** the block has `poster_width: -1`
- **THEN** config loading fails with an error that names `hook.qbittorrent.config.metadata.poster_width`

## ADDED Requirements

### Requirement: Poster resolution
The lookup SHALL return the poster URL with its image transform segment, everything between `._V1_` and the file extension (for example `SX300` or `QL75_UX380_CR0,0,380,562_`), replaced by `SX<poster_width>`, when `poster_width` is above `0` and the URL contains `._V1_`. Otherwise the URL MUST be returned unchanged. When the URL was resized, the lookup MUST also return the original URL as the fallback poster.

#### Scenario: Resized thumbnail
- **WHEN** `poster_width` is `1000` and OMDb returns `https://m.media-amazon.com/images/M/MV5BNzk@._V1_SX300.jpg`
- **THEN** the poster is `https://m.media-amazon.com/images/M/MV5BNzk@._V1_SX1000.jpg` and the fallback poster is the original URL

#### Scenario: Resized cropped thumbnail
- **WHEN** `poster_width` is `1000` and OMDb returns `https://m.media-amazon.com/images/M/MV5BNzk@._V1_QL75_UX380_CR0,0,380,562_.jpg`
- **THEN** the poster is `https://m.media-amazon.com/images/M/MV5BNzk@._V1_SX1000.jpg` and the fallback poster is the original URL

#### Scenario: Unknown pattern
- **WHEN** OMDb returns `https://example.com/poster.jpg`
- **THEN** the poster is that URL and there is no fallback poster

#### Scenario: Resizing off
- **WHEN** `poster_width` is `0`
- **THEN** the poster is the original URL and there is no fallback poster
