# Tasks

## 1. Shared HTTP package and configuration

- [ ] 1.1 Add `internal/httpx` with `Settings` and `NewClient(Settings)` covering dial, keep-alive, TLS handshake and TLS config, idle connections, idle timeout, and optional response-header and expect-continue timeouts; verify `TestNewClientAppliesSettings` checks every transport field and the client timeout.
- [ ] 1.2 Add `httpx.Send` (one request, status and body read up to `maxBytes`) and `DefaultMaxResponseBytes` = 1 MiB; verify `TestSendCapsBodyAtLimit` (2000-byte body, limit 1024, 1024 bytes returned, no error) and `TestSendReturnsStatusAndBody`.
- [ ] 1.3 Add `server.read_timeout`, `server.write_timeout` (default 15s) and `server.shutdown_timeout` (default 30s), positive only; verify `TestServerTimeoutDefaults` and `TestServerTimeoutRejectsZero` in `internal/config`.
- [ ] 1.4 Use the server timeouts in `newHTTPServer` and `shutdownServer` in `cmd/automation-hub/main.go`; verify `TestNewHTTPServerUsesConfiguredTimeouts` in `cmd/automation-hub/main_test.go`.
- [ ] 1.5 Add the `telegram.http`, `github.http` and `kubernetes.http` sections with the per-client defaults from the spec and positive-only validation; verify `TestHTTPSettingsDefaults` (table over the three sections) and `TestHTTPSettingsRejectNonPositive`.
- [ ] 1.6 Add `github.api_url` (default `https://api.github.com`, trailing slash trimmed, `https` required); verify `TestAPIURLValidation` (default, GHES with trailing slash, `http://` rejected).
- [ ] 1.7 Add `telegram.http.poll_timeout` (30s), `timeout` (poll + 10s), `response_header_timeout` (poll + 5s) and `expect_continue_timeout` (1s), rejecting a `timeout` or `response_header_timeout` not greater than `poll_timeout`; verify `TestTelegramPollBoundsDerived` (poll 50s gives timeout 60s and header 55s) and `TestTelegramTimeoutNotAbovePoll` (poll 30s and timeout 30s fail).
- [ ] 1.8 Add `telegram.http.max_attempts` (3, minimum 1), `retry_backoff` (1s) and `max_retry_after` (30s); verify `TestRetryConfigDefaultsAndMinimum` (`max_attempts: 0` fails).
- [ ] 1.9 Add `github.http.max_response_bytes` and `kubernetes.http.max_response_bytes` (default `httpx.DefaultMaxResponseBytes`, positive only); verify `TestMaxResponseBytesValidation`.
- [ ] 1.10 Document every new key with its default in `configs/config.yaml.example` and in the README configuration section; verify `TestConfigExampleDocumentsNewKeys`, which reads the example file and asserts each key path from the spec is present.

## 2. Telegram client

- [ ] 2.1 Change `telegram.NewClient` to take a `Config` (token, logger, HTTP settings, poll timeout, retry policy) built on `httpx.NewClient`, and make `StartPolling` use the configured poll timeout; verify `TestNewClientUsesConfiguredPollTimeout` and the updated `TestNewClient`.
- [ ] 2.2 Add `resolveChat` and a `ready()` nil guard shared by `sendText` and `SendDocument`; verify `TestSendRejectsInvalidChatIDWithoutRequest` (both entry points, zero requests received) and `TestNilClientSendsNothing`.
- [ ] 2.3 Add the `Backoff` strategy (exponential from `retry_backoff`, honouring `*tgbotapi.Error.ResponseParameters.RetryAfter`, capped by `max_retry_after`) and the `RetryingSender` decorator with an injected `sleep`, retrying only 429, 5xx and non-API errors; verify table test `TestRetryingSenderClassification` (400, 429, 500, 502, network error, `max_attempts: 1`) and `TestRetryingSenderBackoffSequence` (502, 502, 200 records waits of 1s and 2s).
- [ ] 2.4 Verify the `retry_after` branch: `TestRetryingSenderHonoursRetryAfter` (429 with `retry_after: 5` waits 5s) and `TestRetryingSenderStopsAboveMaxRetryAfter` (`retry_after: 120` makes one attempt and waits 0s).
- [ ] 2.5 Replace the old `send` loop in `sendText` and `SendDocument` with the decorator; verify `TestSendMessageRetriesGatewayError` (502 then 200 succeeds) and the existing `TestSendMessageSuccess` and `TestSendParseMode`.
- [ ] 2.6 Rename `TestSendDocumentReturnsErrorAfterRetries` to `TestSendDocumentDoesNotRetryBadRequest`: the fake answers `error_code: 400` and the test expects one attempt, which is the spec's intended change; verify it passes.
- [ ] 2.7 Confirm `SetMyCommands` and `StartPolling` keep their behavior; verify `TestSetMyCommands`, `TestSetMyCommandsReturnsTelegramError` and `TestStartPollingReceivesUpdatesUntilContextIsCancelled` pass unchanged.

## 3. GitHub client

- [ ] 3.1 Change `github.NewClient` to take a `Config` (token, API URL, HTTP settings, max response bytes) built on `httpx.NewClient`; verify `TestNewClientUsesAPIURL` (an `httptest` server stands in for GHES and receives the `/repos/.../dispatches` path) and the updated `TestNewClientDispatchWorkflow` and `TestNewClientWithBaseURLUsesDefaults`.
- [ ] 3.2 Split `DispatchWorkflow` into `newDispatchRequest` and `decodeDispatch`, and send through `httpx.Send`, keeping `return_run_details: true` and the 200 and 204 handling; verify the existing `TestDispatchWorkflow` and `TestDispatchWorkflowNetworkError` pass unchanged.
- [ ] 3.3 Verify that an oversized 200 body is cut at the cap: `TestDispatchWorkflowCapsOversizedBody` (cap 1024, body 2000 bytes) returns an empty URL with no error.

## 4. Kubernetes client

- [ ] 4.1 Change `NewInClusterClient` to take `httpx.Settings` and the max response bytes, mapping the service account CA into `Settings.TLSConfig`; keep the outside-cluster error; verify `TestNewInClusterClientOutsideCluster` and a new `TestInClusterSettingsUseCA` that checks the RootCAs pool and the settings mapping.
- [ ] 4.2 Add `resourceURL`, named path-format and media-type constants, and route `doJSON` through `httpx.Send` with the cap; verify `TestDoJSONCapsErrorBody` (4096-byte error body with cap 1024 gives a bounded error) and the existing `TestCreateJobFromCronJobGetError`.
- [ ] 4.3 Route `RestartDeployment` through `doJSON` and take `restartedAt` from `c.now()`; verify `TestRestartDeployment` with a fixed clock asserting `2026-10-11T09:07:00Z` and the strategic-merge content type, and `TestRestartDeploymentErrorStatus`.
- [ ] 4.4 Route `getDeployment` through `doJSON`; verify `TestWaitForRolloutReturnsOnceReady` and `TestWaitForRolloutTimesOutWhileNotReady` pass unchanged.
- [ ] 4.5 Replace the `map[string]any` manifest with `job`, `objectMeta`, `ownerReference` and `jobFromTemplate`, keeping the template spec as `json.RawMessage`; verify `TestCreateJobFromCronJob` and `assertPostedJob` (labels, annotations, owner reference, spec forwarded) and that the fixed clock yields `ops-cleanup-manual-1791709620`.
- [ ] 4.6 Verify that an oversized CronJob body is rejected: `TestCreateJobFromCronJobCapsCronJobBody` (4096-byte body, cap 1024) returns an error wrapping `decode response` and sends no POST.
- [ ] 4.7 (homelab) If `github.api_url` points outside `api.github.com`, add an egress NetworkPolicy from the automation-hub pods to that host in the homelab repo; verify the dispatch from the pod succeeds in the cluster.

## 5. Final validation

- [ ] 5.1 Run `docker run --rm -v "$PWD":/app -w /app golang:1.27.2 sh -c 'go vet ./... && go test -race ./...'` and verify exit code 0; then run `graphify update .`.
- [ ] 5.2 Run golangci-lint v2.14.0 in Docker with `--enable=errcheck,govet,staticcheck,unused,misspell,gosec` over the module and verify zero findings.
- [ ] 5.3 Run gocognit v1.2.0 in Docker (`go install github.com/uudashr/gocognit/cmd/gocognit@v1.2.0` inside the container) with `-over 15` over all packages, tests included, and verify no function is reported.
- [ ] 5.4 Run `openspec validate --strict http-clients-hardening` and verify it passes.
