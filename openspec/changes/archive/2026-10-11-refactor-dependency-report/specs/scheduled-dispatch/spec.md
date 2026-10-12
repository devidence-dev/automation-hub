# scheduled-dispatch Specification

## RENAMED Requirements

- FROM: `### Requirement: Fixed local timezone`
- TO: `### Requirement: Configured local timezone`

## MODIFIED Requirements

### Requirement: Configured local timezone
The system SHALL evaluate every schedule in the zone named by the top-level `timezone` key, which defaults to `America/Guayaquil`, using timezone data embedded in the binary. It SHALL NOT depend on the host or container image. If `timezone` is not a valid IANA zone name, the process MUST exit with an error before it starts the bot, IMAP monitoring, the HTTP server or the scheduler.

#### Scenario: Image without tzdata
- **WHEN** automation-hub starts in a container that has no `/usr/share/zoneinfo` and `timezone` is not set
- **THEN** the startup log reports timezone `America/Guayaquil`
- **AND** a schedule of `0 9 * * 1` next fires on a Monday at 09:00 -05:00 (14:00 UTC)

#### Scenario: Configured zone
- **WHEN** the config has `timezone: Europe/Madrid` and a workflow with `schedule: "0 9 * * 1"`, and the clock is 2026-10-13 00:00 in Madrid
- **THEN** the startup log reports timezone `Europe/Madrid`, and the next run is Monday 2026-10-19 at 09:00 +02:00 (07:00 UTC)

#### Scenario: Invalid zone
- **WHEN** the config has `timezone: Mars/Olympus`
- **THEN** config loading fails with an error that names `timezone`
- **AND** the process exits before it starts the bot, IMAP monitoring, the HTTP server or the scheduler
