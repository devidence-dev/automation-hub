# job-commands Specification

## MODIFIED Requirements

### Requirement: Scheduled job commands
A `jobs:` entry with `schedule` SHALL be scheduled like a scheduled workflow, in the zone named by the top-level `timezone` key (default `America/Guayaquil`). A failure SHALL be logged and notified to each of the entry's `allowed_chat_ids`.

#### Scenario: Weekly report
- **WHEN** `run_dependency_report` has `schedule: "0 9 * * 1"` and the clock reaches Monday 09:00 America/Guayaquil
- **THEN** a Job is created from CronJob `dependency-report`

#### Scenario: Weekly report in a configured zone
- **WHEN** the config has `timezone: Europe/Madrid` and `run_dependency_report` has `schedule: "0 9 * * 1"`, and the clock reaches Monday 2026-10-19 09:00 Europe/Madrid
- **THEN** a Job is created from CronJob `dependency-report`
