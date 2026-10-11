# scheduled-dispatch Specification

## Purpose
Lets automation-hub decide when each operational workflow runs. It dispatches configured bot commands on a cron schedule in the America/Guayaquil timezone, so runs no longer depend on GitHub's best-effort scheduler.

## Requirements

### Requirement: Fixed local timezone
The system SHALL evaluate every schedule in the `America/Guayaquil` timezone, using timezone data embedded in the binary. It SHALL NOT depend on the host or container image. If the timezone cannot be loaded, the process MUST exit with an error before it starts the bot, IMAP monitoring, the HTTP server or the scheduler.

#### Scenario: Image without tzdata
- **WHEN** automation-hub starts in a container that has no `/usr/share/zoneinfo`
- **THEN** the startup log reports timezone `America/Guayaquil`
- **AND** a schedule of `0 9 * * 1` next fires on a Monday at 09:00 -05:00 (14:00 UTC)

### Requirement: Optional schedule per workflow command
A `workflows:` entry MAY declare a `schedule` with a standard 5-field cron expression (minute, hour, day of month, month, day of week). An entry without `schedule` SHALL behave exactly as before: it runs only on a Telegram command.

#### Scenario: Valid schedule accepted
- **WHEN** the config has `run_check_updates` with `schedule: "0 9 * * 1,3,5"`
- **THEN** startup succeeds and the entry is scheduled

#### Scenario: Invalid schedule rejected
- **WHEN** the config has an entry with `schedule: "0 9 * *"`
- **THEN** config loading fails with an error that names the command and the invalid expression
- **AND** the process exits before it starts the bot or the scheduler

#### Scenario: Entry without schedule
- **WHEN** `run_cleanup` has no `schedule`
- **THEN** it is not scheduled, and `/run_cleanup` still dispatches it from Telegram

### Requirement: Scheduled dispatch uses the command's dispatch path
At each scheduled time, the system SHALL send a `workflow_dispatch` for the entry's `owner`, `repo`, `workflow_file` and `ref`, through the same dispatch behavior as the Telegram command. The run does not depend on any Telegram chat or update.

#### Scenario: Dispatch at the agreed time
- **WHEN** `run_check_updates` has `schedule: "0 9 * * 1,3,5"` and the clock reaches Wednesday 09:00 America/Guayaquil
- **THEN** a `workflow_dispatch` is sent for `devidence-dev/workflows` `ops-github-runner-check-updates.yml`
- **AND** the log records the command, owner/repo, workflow file, the result and the HTML URL of the exact run that GitHub created

### Requirement: Exact run link
Every successful dispatch, whether scheduled or from Telegram, SHALL identify the run it created from GitHub's dispatch response. It SHALL NOT pick the run by searching recent runs. If GitHub answers without run details, the link SHALL fall back to the repository's Actions page.

#### Scenario: Run details returned
- **WHEN** GitHub answers the dispatch with 200 and `html_url` `https://github.com/devidence-dev/workflows/actions/runs/123456`
- **THEN** the Telegram reply and the log use that URL

#### Scenario: Overlapping dispatches
- **WHEN** a scheduled and a manual `run_check_updates` are dispatched within the same second
- **THEN** each one reports the URL of its own run

#### Scenario: No run details
- **WHEN** GitHub answers the dispatch with 204 No Content
- **THEN** the dispatch counts as successful and the link is `https://github.com/devidence-dev/workflows/actions`

#### Scenario: Not a scheduled day
- **WHEN** the clock reaches Tuesday 09:00 America/Guayaquil for the same entry
- **THEN** no dispatch is sent

### Requirement: Startup schedule log
At startup, the system SHALL log the timezone in use and the next run time of every scheduled entry.

#### Scenario: Next runs logged
- **WHEN** automation-hub starts on Sunday with `run_check_updates` (`0 9 * * 1,3,5`) and `run_cleanup` (`0 9 * * 2,5`)
- **THEN** the log shows `America/Guayaquil`, the next `run_check_updates` run on Monday at 09:00, and the next `run_cleanup` run on Tuesday at 09:00

### Requirement: Failure notification for scheduled dispatch
When a scheduled dispatch fails, the system SHALL log the error and send a failure message to every chat in that entry's `allowed_chat_ids`. A failed dispatch MUST NOT stop the scheduler, the bot or any other component.

#### Scenario: GitHub API error
- **WHEN** the scheduled dispatch of `run_cleanup` gets HTTP 500 from GitHub
- **THEN** each chat in `run_cleanup`'s `allowed_chat_ids` receives a failure message naming the command
- **AND** the next scheduled run of every entry still happens

### Requirement: Graceful stop
The scheduler SHALL stop on SIGINT or SIGTERM together with the rest of the process, and SHALL NOT start new dispatches after shutdown begins.

#### Scenario: Shutdown
- **WHEN** the process receives SIGTERM
- **THEN** the scheduler stops and the process exits without starting new dispatches
