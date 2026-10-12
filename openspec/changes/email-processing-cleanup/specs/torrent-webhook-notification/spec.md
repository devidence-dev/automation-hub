# Spec Delta

## Purpose

Forwards qBittorrent download-completion webhooks to Telegram. The hub reads the torrent name and save path from a JSON body, fills the hook's message template and sends it to the hook's chat. Covers the body size cap and the hook configuration lookup. Route registration is owned by the bot-command-registry change.

## ADDED Requirements

### Requirement: Completion notification
The `qbittorrent` hook SHALL accept a JSON body with `torrent_name` and `save_path`, fill its `telegram_message` with the name as the first `%s` and the path as the second, and send the result to its `telegram_chat_id`. On success the response MUST be `200` with body `{"status":"success"}`. A body that is not valid JSON MUST get `400` with `Invalid request` and send nothing.

#### Scenario: Completed download
- **WHEN** the body is `{"torrent_name":"Dune.2021.1080p","save_path":"/downloads/movies"}` and the hook template is `Download finished: %s at %s`
- **THEN** the chat receives `Download finished: Dune.2021.1080p at /downloads/movies` and the response is `200` with `{"status":"success"}`

#### Scenario: Invalid JSON
- **WHEN** the body is `{"torrent_name":`
- **THEN** the response is `400` with `Invalid request` and no Telegram message is sent

#### Scenario: Telegram failure
- **WHEN** the Telegram send fails
- **THEN** the response is `500` with `Processing failed` and the failure is logged

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
The hook named `qbittorrent` SHALL be looked up by name in the `hook:` list before the body is read. When no such hook exists, the response MUST be `500` with `Webhook configuration not found`, regardless of the body.

#### Scenario: Hook missing
- **WHEN** the config has only a `sonarr` hook and a `POST` arrives at `/webhook/qbittorrent` with an invalid body
- **THEN** the response is `500` with `Webhook configuration not found` and no Telegram message is sent
