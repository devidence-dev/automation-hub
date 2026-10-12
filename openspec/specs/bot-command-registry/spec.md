# bot-command-registry Specification

## Purpose
Routes Telegram commands and scheduled runs for workflow, restart and job entries through one registry of command types, so each type's authorization, dispatch and texts are defined once. Every bot text comes from the shared message catalog, with the current English strings as defaults.

## Requirements

### Requirement: Command routing by name
The system SHALL resolve each Telegram command to exactly one configured entry, whose type (workflow, restart or job) decides how it runs. A command that matches no entry SHALL get the bot text `bot.unknown_command` (default `❓ Unknown command.`) and nothing SHALL run.

#### Scenario: Unknown command
- **WHEN** an allowed chat sends `/deploy_everything`, which no entry defines
- **THEN** the chat receives `❓ Unknown command.`
- **AND** no workflow is dispatched, no deployment is restarted and no Job is created

#### Scenario: Known commands reach their own type
- **WHEN** the config defines `run_cleanup` (workflow), `restart_runner` (restart) and `run_dependency_report` (job)
- **THEN** `/run_cleanup` dispatches the workflow, `/restart_runner` restarts the deployment and `/run_dependency_report` creates a Job, each through its own executor

### Requirement: Per-entry authorization
Before any command runs, the chat ID SHALL be in that entry's `allowed_chat_ids`. A chat that is not allowed SHALL cause no action, SHALL receive the bot text `bot.unauthorized` (default `🚫 You are not authorized to run this command.`), and SHALL produce a warning log with the command, chat ID and username. An entry with an empty `allowed_chat_ids` SHALL deny every chat.

#### Scenario: Unauthorized chat
- **WHEN** chat `555000111` sends `/run_cleanup`, and `run_cleanup` allows only `123456789`
- **THEN** no workflow is dispatched and chat `555000111` receives `🚫 You are not authorized to run this command.`
- **AND** a warning is logged with command `run_cleanup`, chat_id `555000111`

#### Scenario: Empty allow list
- **WHEN** `restart_runner` has `allowed_chat_ids: []` and chat `123456789` sends `/restart_runner`
- **THEN** the deployment is not restarted and the chat receives the unauthorized text

### Requirement: Single publication of Telegram commands
At startup the system SHALL publish every configured workflow, restart and job command to Telegram in one `setMyCommands` call, in config order: workflows, then restarts, then jobs. If Telegram rejects the call, the process MUST exit with an error before it starts the bot.

#### Scenario: Three command types published together
- **WHEN** the config defines `run_cleanup` (workflow), `restart_runner` (restart) and `run_dependency_report` (job)
- **THEN** one `setMyCommands` call publishes those three commands in that order

#### Scenario: Telegram rejects the menu
- **WHEN** `setMyCommands` returns an error at startup
- **THEN** the process exits with an error and does not start polling

### Requirement: Single scheduling path
Every entry with a `schedule` SHALL be registered with the scheduler through the same adapter, whatever its type. Entries without a `schedule` SHALL NOT be registered. Restart entries have no schedule and SHALL never be registered. A scheduled run SHALL use the same executor as its Telegram command, without a chat.

#### Scenario: Mixed schedules
- **WHEN** `run_cleanup` has `schedule: "0 9 * * 2,5"`, `run_dependency_report` has `schedule: "0 9 * * 1"` and `restart_runner` has no schedule
- **THEN** `run_cleanup` and `run_dependency_report` are registered, and `restart_runner` is not

### Requirement: Bot texts from the message catalog
Every bot text SHALL be resolved from the shared message catalog by its key. A `messages:` override in the configuration SHALL replace the default text for that key. A value that does not keep the same number of `%s` placeholders as its default SHALL fail config loading, naming the key.

#### Scenario: Default text
- **WHEN** the config has no `messages:` section and chat `123456789` sends `/deploy_everything`
- **THEN** the reply is `❓ Unknown command.`

#### Scenario: Override changes a text
- **WHEN** the config sets `messages:` with `bot.unknown_command: "❔ Comando no existe."`
- **THEN** the reply to `/deploy_everything` is `❔ Comando no existe.`
- **AND** the unauthorized text stays `🚫 You are not authorized to run this command.`

#### Scenario: Placeholder count mismatch
- **WHEN** the config sets `bot.job_created: "🚀 Job creado."`, which has no `%s` placeholders where the default has two
- **THEN** config loading fails with an error that names `bot.job_created`

### Requirement: Failure isolation in the bot
A failed Telegram reply SHALL be logged and SHALL NOT stop the update loop. Each command type SHALL report its own failures to the chat with its failure text and SHALL NOT crash the bot.

#### Scenario: Reply fails
- **WHEN** Telegram returns an error for the reply to `/run_cleanup`
- **THEN** the error is logged with the chat ID
- **AND** the next update, `/restart_runner`, is still handled

### Requirement: No client for unconfigured types
A Kubernetes or GitHub client SHALL be created only when a configured entry of a type that uses it exists. A missing client SHALL be passed as no client, never as a non-nil client wrapping nothing.

#### Scenario: Workflows only
- **WHEN** the config defines only `workflows:` entries
- **THEN** the process starts without reading in-cluster Kubernetes credentials
