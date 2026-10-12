# job-commands Specification

## Purpose
Lets Telegram commands and schedules start a one-off Kubernetes Job from a suspended CronJob template, so heavy scans run in the cluster and outside the hub process.

## Requirements

### Requirement: Job command configuration
The system SHALL accept `jobs:` entries with `command`, `description`, `namespace`, `cronjob`, an optional `schedule` and `allowed_chat_ids`. Command names MUST match `^[a-z0-9_]{1,32}$` and MUST be unique across `workflows`, `restarts` and `jobs`. Config loading MUST fail on a violation, or on an invalid `schedule`.

#### Scenario: Duplicate across types
- **WHEN** `workflows:` and `jobs:` both define `run_check_updates`
- **THEN** config loading fails with an error that names `run_check_updates`

#### Scenario: Jobs-only config starts the bot
- **WHEN** the config has one `jobs:` entry and no `workflows:` or `restarts:`
- **THEN** the bot starts, registers the command and can create the Job

### Requirement: Create a Job from a CronJob
When an allowed chat sends a job command, the system SHALL read the configured CronJob and create a Job in the same namespace from its job template. The Job SHALL be named `<cronjob>-manual-<unix seconds>`, keep the template's labels and annotations, and carry the annotation `cronjob.kubernetes.io/instantiate: manual`. The reply SHALL be `🚀 Job *<name>* created from *<cronjob>*.`

#### Scenario: Successful creation
- **WHEN** an allowed chat sends `/run_dependency_report` at Unix time 1791763200 and the entry points to CronJob `dependency-report` in `internal-apps`
- **THEN** Job `dependency-report-manual-1791763200` is created in `internal-apps`
- **AND** the chat receives `🚀 Job *dependency-report-manual-1791763200* created from *dependency-report*.`

#### Scenario: Kubernetes API error
- **WHEN** the API returns 403 when the Job is created
- **THEN** the chat receives `❌ Failed to create the job. Check the hub logs or try again later.` and the error is logged

#### Scenario: Unauthorized chat
- **WHEN** a chat outside `allowed_chat_ids` sends `/run_dependency_report`
- **THEN** no Job is created and the chat receives the existing unauthorized message

### Requirement: Scheduled job commands
A `jobs:` entry with `schedule` SHALL be scheduled like a scheduled workflow, in the zone named by the top-level `timezone` key (default `America/Guayaquil`). A failure SHALL be logged and notified to each of the entry's `allowed_chat_ids`.

#### Scenario: Weekly report
- **WHEN** `run_dependency_report` has `schedule: "0 9 * * 1"` and the clock reaches Monday 09:00 America/Guayaquil
- **THEN** a Job is created from CronJob `dependency-report`

#### Scenario: Weekly report in a configured zone
- **WHEN** the config has `timezone: Europe/Madrid` and `run_dependency_report` has `schedule: "0 9 * * 1"`, and the clock reaches Monday 2026-10-19 09:00 Europe/Madrid
- **THEN** a Job is created from CronJob `dependency-report`
