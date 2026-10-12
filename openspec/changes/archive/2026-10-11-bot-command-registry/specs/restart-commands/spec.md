## Purpose
Restarts a Kubernetes Deployment from a Telegram command with a rollout restart, reports when the new pods are ready, and bounds that wait with configurable timeout and poll interval. The bot never blocks on the wait.

## ADDED Requirements

### Requirement: Restart a Deployment from Telegram
When an allowed chat sends a restart command, the system SHALL trigger a rollout restart of the configured Deployment in its namespace. On success the reply SHALL be `🔄 Deployment *<deployment>* restart triggered. Waiting for it to come back online...`. On failure the reply SHALL be `❌ Failed to restart the deployment. Check the hub logs or try again later.`, and no readiness wait starts.

#### Scenario: Restart triggered
- **WHEN** an allowed chat sends `/restart_runner`, which points to Deployment `github-runner` in namespace `infrastructure`
- **THEN** the rollout restart is sent to `infrastructure/github-runner`
- **AND** the chat receives `🔄 Deployment *github-runner* restart triggered. Waiting for it to come back online...`

#### Scenario: Ready again
- **WHEN** the rollout of `github-runner` reports ready within the timeout
- **THEN** the chat receives `✅ Deployment *github-runner* is up and running again.`

#### Scenario: Kubernetes API error
- **WHEN** the Kubernetes API answers the restart with 403
- **THEN** the chat receives `❌ Failed to restart the deployment. Check the hub logs or try again later.`
- **AND** no readiness wait is started and the error is logged

### Requirement: Configurable readiness wait
The top-level `restart_ready` section SHALL set `timeout` (default `3m`) and `poll_interval` (default `5s`) for every restart command. Both MUST be positive durations, and `poll_interval` MUST NOT exceed `timeout`. Invalid values SHALL fail config loading, naming the key.

#### Scenario: Defaults without the section
- **WHEN** the config has no `restart_ready` section
- **THEN** the rollout is polled every `5s` for up to `3m`
- **AND** a timeout is reported as `did not report ready within 3m0s`

#### Scenario: Shorter timeout
- **WHEN** the config sets `restart_ready: {timeout: "1m", poll_interval: "5s"}` and `github-runner` is still not ready after 1 minute
- **THEN** the chat receives `⏱️ Deployment *github-runner* did not report ready within 1m0s — check it manually.`

#### Scenario: Custom poll interval
- **WHEN** the config sets `poll_interval: "2s"`
- **THEN** the rollout is polled every 2 seconds

#### Scenario: Zero timeout rejected
- **WHEN** the config sets `restart_ready.timeout: "0s"`
- **THEN** config loading fails with an error that names `restart_ready.timeout`

#### Scenario: Poll interval longer than timeout rejected
- **WHEN** the config sets `timeout: "5s"` and `poll_interval: "10s"`
- **THEN** config loading fails with an error that names `restart_ready.poll_interval`

#### Scenario: Unparsable duration rejected
- **WHEN** the config sets `timeout: "soon"`
- **THEN** config loading fails with an error that names `restart_ready.timeout`

### Requirement: Wait runs in the background
The readiness wait SHALL run in the background, so the bot handles the next Telegram update while a rollout is pending. A timed-out wait SHALL be logged as a warning.

#### Scenario: Next command during a rollout
- **WHEN** `/restart_runner` is sent and, while its rollout is pending, `/run_cleanup` is sent
- **THEN** `/run_cleanup` is handled before the rollout finishes
