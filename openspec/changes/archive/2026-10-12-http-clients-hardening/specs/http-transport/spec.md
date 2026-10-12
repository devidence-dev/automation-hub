# Spec Delta

## Purpose

Defines the timeouts, body caps, API base URL and Telegram retry policy of the outbound HTTP clients and the HTTP server. Every value has a documented default equal to today's value, so existing deployments behave the same unless they set a key.

## ADDED Requirements

### Requirement: Configurable server timeouts
The system SHALL read the HTTP server's read timeout, write timeout and shutdown timeout from `server.read_timeout`, `server.write_timeout` and `server.shutdown_timeout`. Each MUST be a positive Go duration. When a key is omitted it defaults to 15s, 15s and 30s. Config loading MUST fail on a zero, negative or unparseable value.

#### Scenario: Defaults when keys are omitted
- **WHEN** the config has no `server` timeout keys
- **THEN** the server runs with read timeout 15s, write timeout 15s and shutdown timeout 30s

#### Scenario: Read timeout raised
- **WHEN** the config sets `server.read_timeout: 30s`
- **THEN** the server's read timeout is 30s and the write timeout stays 15s

#### Scenario: Invalid server timeout
- **WHEN** the config sets `server.shutdown_timeout: 0s`
- **THEN** startup fails with an error naming `server.shutdown_timeout`, before the server listens

### Requirement: Duration keys carry a unit
Every duration key of this change, and `restart_ready.timeout` and `restart_ready.poll_interval`, MUST be written as a quoted string with a unit, such as `"30s"`. A bare YAML integer decodes to nanoseconds without an error, so config loading MUST reject it and name the key.

#### Scenario: Bare integer duration
- **WHEN** the config sets `server.read_timeout: 30`
- **THEN** startup fails with an error naming `server.read_timeout` and stating that a unit is required

### Requirement: Shared outbound HTTP settings
The Telegram, GitHub and Kubernetes clients SHALL build their HTTP transport and client from the same settings: `timeout` (whole request), `dial_timeout`, `tls_handshake_timeout`, `keep_alive`, `max_idle_conns` and `idle_conn_timeout`. Each client has its own section (`telegram.http`, `github.http`, `kubernetes.http`). Defaults: `timeout` 30s for GitHub and 15s for Kubernetes; `dial_timeout` 30s for GitHub (today's Go default transport) and 10s for Telegram and Kubernetes; `tls_handshake_timeout` 10s, `keep_alive` 30s, `max_idle_conns` 100 and `idle_conn_timeout` 90s for all three. Config loading MUST fail on a zero or negative duration or count.

#### Scenario: Kubernetes timeout raised
- **WHEN** the config sets `kubernetes.http.timeout: 20s` and omits every other key
- **THEN** the Kubernetes client uses a 20s request timeout and the defaults for the other settings

#### Scenario: Negative idle connection count
- **WHEN** the config sets `github.http.max_idle_conns: -1`
- **THEN** startup fails with an error naming `github.http.max_idle_conns`

### Requirement: Telegram long-poll timeouts
The Telegram client SHALL use `telegram.http.poll_timeout` (default 30s) as the long-polling window sent to Telegram. The request `timeout` MUST be greater than `poll_timeout` and defaults to `poll_timeout + 10s`. `response_header_timeout` MUST be greater than `poll_timeout` and defaults to `poll_timeout + 5s`. `expect_continue_timeout` defaults to 1s and MUST be positive. Config loading MUST fail when either bound is not greater than `poll_timeout`, and when `expect_continue_timeout` is zero or negative.

#### Scenario: Poll timeout raised without explicit bounds
- **WHEN** the config sets `telegram.http.poll_timeout: 50s` and omits `timeout` and `response_header_timeout`
- **THEN** the polling request is sent with `timeout: 50` seconds, the client timeout is 60s and the response header timeout is 55s

#### Scenario: Request timeout not above the poll window
- **WHEN** the config sets `telegram.http.poll_timeout: 30s` and `telegram.http.timeout: 30s`
- **THEN** startup fails with an error naming `telegram.http.timeout`

### Requirement: Telegram retry policy
The Telegram client SHALL retry a failed send only when the failure is a network or transport error, an answer with status 5xx, or an answer with status 429. It SHALL NOT retry any other 4xx answer. The number of attempts is `telegram.http.max_attempts` (default 3, minimum 1). Between attempts it SHALL wait `telegram.http.retry_backoff` (default 1s) doubled after each failed attempt, and never longer than `telegram.http.max_retry_after` (default 30s). When a 429 answer carries `retry_after`, the wait SHALL be that many seconds instead of the backoff. If that wait is longer than `telegram.http.max_retry_after`, the client SHALL stop and return the error without waiting. `retry_backoff` and `max_retry_after` MUST be positive; config loading MUST fail otherwise.

#### Scenario: Bad request is not retried
- **WHEN** Telegram answers `sendDocument` with 400 `Bad Request: can't parse entities`
- **THEN** exactly one attempt is made, no wait happens, and the error is returned

#### Scenario: Gateway errors then success
- **WHEN** Telegram answers `sendMessage` with 502 twice and then 200
- **THEN** three attempts are made, waiting 1s before the second and 2s before the third, and the send succeeds

#### Scenario: Backoff capped at max_retry_after
- **WHEN** the config sets `telegram.http.retry_backoff: 10s`, `telegram.http.max_retry_after: 30s` and `telegram.http.max_attempts: 4`, and Telegram answers `sendMessage` with 502 three times and then 200
- **THEN** the waits before the second, third and fourth attempts are 10s, 20s and 30s, and the send succeeds

#### Scenario: Attempts exhausted
- **WHEN** Telegram answers every `sendMessage` attempt with 502 and `max_attempts` is 3
- **THEN** exactly three attempts are made and the error reads `failed to send message after 3 attempts`

#### Scenario: Rate limit with retry_after
- **WHEN** Telegram answers `sendMessage` with 429 and `parameters.retry_after: 5`, then 200
- **THEN** the second attempt is made after a 5s wait, not the 1s backoff, and the send succeeds

#### Scenario: Rate limit beyond the retry ceiling
- **WHEN** Telegram answers with 429 and `retry_after: 120`
- **THEN** no further attempt is made, the error is returned, and the wait is 0s

#### Scenario: Network failure
- **WHEN** the connection to Telegram is reset on the first attempt and succeeds on the second
- **THEN** two attempts are made with a 1s wait between them

#### Scenario: Retries disabled
- **WHEN** the config sets `telegram.http.max_attempts: 1` and Telegram answers 502
- **THEN** one attempt is made and the error is returned

#### Scenario: Zero attempts rejected
- **WHEN** the config sets `telegram.http.max_attempts: 0`
- **THEN** startup fails with an error naming `telegram.http.max_attempts`

### Requirement: Configurable GitHub API base URL
The GitHub client SHALL send its requests to `github.api_url`, default `https://api.github.com`. A trailing slash is ignored. The value MUST be an absolute `https` URL; config loading MUST fail otherwise.

#### Scenario: GitHub Enterprise Server
- **WHEN** the config sets `github.api_url: https://ghe.example.com/api/v3/` and a `workflows:` entry has owner `devidence-dev`, repo `workflows` and workflow file `ops-github-runner-cleanup.yml`
- **THEN** the dispatch is sent to `https://ghe.example.com/api/v3/repos/devidence-dev/workflows/actions/workflows/ops-github-runner-cleanup.yml/dispatches`

#### Scenario: Plain HTTP base URL rejected
- **WHEN** the config sets `github.api_url: http://ghe.example.com/api/v3`
- **THEN** startup fails with an error naming `github.api_url`

### Requirement: Bounded response bodies
The GitHub and Kubernetes clients SHALL read at most `github.http.max_response_bytes` and `kubernetes.http.max_response_bytes` bytes of any response body, including error bodies and successful JSON bodies. Both default to 1048576 (1 MiB). Config loading MUST fail on a value of zero or less.

#### Scenario: GitHub dispatch answers 200 with an oversized body
- **WHEN** `github.http.max_response_bytes` is 1024 and GitHub answers 200 with a 2000-byte body
- **THEN** the body is cut at 1024 bytes, the decode fails, the warning is logged, and the dispatch counts as successful with an empty run URL

#### Scenario: Kubernetes CronJob response too large
- **WHEN** `kubernetes.http.max_response_bytes` is 1024 and the CronJob GET answers with a 4096-byte body
- **THEN** `CreateJobFromCronJob` returns an error wrapping `decode response` and no Job is created

### Requirement: Telegram send entry points
Every Telegram send (text, HTML and document) SHALL first parse the chat ID as a base-10 int64. If the parse fails, the call SHALL return an error beginning `invalid chat ID` without any request to Telegram. A nil or uninitialized client SHALL return nil without any request.

#### Scenario: Non-numeric chat ID
- **WHEN** `SendDocument` is called with chat ID `dependency-report`
- **THEN** the error begins with `invalid chat ID` and the fake Telegram server receives no request

#### Scenario: Uninitialized client
- **WHEN** `SendMessage` is called on a nil client
- **THEN** it returns nil and sends nothing

### Requirement: Unchanged GitHub and Kubernetes wire behavior
The refactor SHALL NOT change any request or response handled by the GitHub and Kubernetes clients. The GitHub dispatch SHALL keep `return_run_details: true` and the 200 and 204 handling. The Kubernetes restart SHALL send a strategic merge patch with annotation `kubectl.kubernetes.io/restartedAt` set from the injected clock in UTC, RFC 3339. The Job name SHALL be `<cronjob>-manual-<unix seconds>` from the injected clock.

#### Scenario: Restart uses the injected clock
- **WHEN** the injected clock returns 2026-10-11 09:07:00 UTC and `RestartDeployment` runs for `internal-apps` deployment `github-runner`
- **THEN** the PATCH body carries `restartedAt` `2026-10-11T09:07:00Z` and the request has content type `application/strategic-merge-patch+json`

#### Scenario: Job named from the injected clock
- **WHEN** the injected clock returns 2026-10-11 09:07:00 UTC and the CronJob `ops-cleanup` is triggered
- **THEN** the created Job is named `ops-cleanup-manual-1791709620`

#### Scenario: Dispatch answers 204
- **WHEN** GitHub answers the dispatch with 204 No Content
- **THEN** the dispatch succeeds with an empty run URL and no error
