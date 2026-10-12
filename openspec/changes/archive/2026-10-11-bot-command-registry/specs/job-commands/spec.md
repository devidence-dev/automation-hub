# Spec Delta

## MODIFIED Requirements

### Requirement: Create a Job from a CronJob
When an allowed chat sends a job command, the system SHALL read the configured CronJob and create a Job in the same namespace from its job template. The Job SHALL be named `<cronjob>-manual-<unix seconds>`, keep the template's labels and annotations, and carry the annotation `cronjob.kubernetes.io/instantiate: manual`. The reply SHALL be the bot text `bot.job_created` from the message catalog, whose default is `🚀 Job *<name>* created from *<cronjob>*.` A failed creation SHALL get the bot text `bot.job_failed`, whose default is `❌ Failed to create the job. Check the hub logs or try again later.`

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
