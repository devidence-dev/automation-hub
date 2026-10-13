# Graph Report - automation-hub  (2026-10-12)

## Corpus Check
- 153 files · ~133,565 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 2029 nodes · 3741 edges · 135 communities (121 shown, 9 thin omitted)
- Extraction: 92% EXTRACTED · 8% INFERRED · 0% AMBIGUOUS · INFERRED: 311 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `ef5c2307`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- UpdateType
- Requirements
- Requirement: Document tabs
- 🤖 Automation Hub
- config.go
- Catalog
- ADDED Requirements
- NewCommandSpecs
- generic_test.go
- Requirements
- ADDED Requirements
- github.com/a-h/templ.Component
- document_view_test.go
- ADDED Requirements
- CLAUDE.md
- automation-hub-network Bridge Network
- Dependabot Docker Update Config (deployments/docker)
- Dependabot GitHub Actions Update Config
- Dependabot Go Modules Update Config
- Requirements
- ADDED Requirements
- Graphify Knowledge-Graph Workflow Rules
- dependency_report_test.go
- ADDED Requirements
- Proposal
- Tasks
- Design
- Requirement: Create a Job from a CronJob
- Requirement: Document tabs
- ADDED Requirements
- Proposal
- Load
- Design
- Tasks
- testing.T
- NewTorrentProcessor
- Requirements
- Requirement: Configurable readiness wait
- telegram/client_test.go
- Requirement: Configurable readiness wait
- Requirement: Document tabs
- buildFailure
- kube/client_test.go
- Decisions
- ADDED Requirements
- Requirements
- Design
- Requirements
- Proposal
- documentView
- Tasks
- Proposal
- Tasks
- Requirement: Exact run link
- Decisions
- Decisions
- automation-hub
- imap_test.go
- RetryingSender
- settings_test.go
- main_test.go
- NewWebhookHandler
- go.uber.org/zap.Logger
- Tasks
- Design
- Requirement: Create a Job from a CronJob
- components_test.go
- Proposal
- Proposal
- NewClient
- Requirement: Document tabs
- summaryAt
- context.Context
- Requirement: Configured local timezone
- Email
- Requirement: Scheduled job commands
- Requirement: HTML report document
- Requirement: Caption repository lines
- command
- Requirement: Authenticated endpoint
- Requirement: Webhook settings
- Requirement: Delivery
- Requirement: Request handling
- Requirement: Scan errors in the document
- workflowCommand
- Requirement: Safe HTML and size limit
- jobCommand
- NewScheduledRunner
- kube/client.go
- loadYAML
- Design
- Tasks
- Requirement: HTML report document
- notifier_test.go
- Scheduler
- Requirement: Body size cap
- Requirement: Body size cap
- restartCommand
- Tasks
- document.go
- BotHandlerDeps
- 🐳 Deployment
- Settings
- verify.sh
- biome.json
- retry_test.go
- scheduler_test.go
- styles_test.go
- Send
- Proposal
- Tasks
- redesign-dependency-report-document/proposal.md
- Requirement: Embedded default catalog
- generate-depreport-assets.sh
- ADDED Requirements
- keys_test.go
- StatsPanel
- Requirement: Document theme
- Requirement: Copy buttons
- Requirement: Repo summary row
- Requirement: Summary caption
- � API & Webhooks
- CI Job: Devidence Go CI
- TemplateNotifier
- MODIFIED Requirements
- Requirement: Repo sections
- Requirement: Summary panel
- Requirement: Safe document content
- Requirement: Motion
- Requirement: Header and logo

## God Nodes (most connected - your core abstractions)
1. `loadYAML()` - 33 edges
2. `summaryAt()` - 31 edges
3. `Config` - 28 edges
4. `loadFixtureReport()` - 25 edges
5. `NewCommandSpecs()` - 23 edges
6. `Catalog` - 23 edges
7. `presentDocument()` - 23 edges
8. `Decisions` - 23 edges
9. `DependencyReportHandler` - 20 edges
10. `UpdateType` - 20 edges

## Surprising Connections (you probably didn't know these)
- `🤖 Step 3: Setup Telegram Bot` --references--> `TelegramConfig`  [INFERRED]
  README.md → internal/config/config.go
- `▶️ Trigger GitHub Actions from Telegram` --references--> `WorkflowBotConfig`  [INFERRED]
  README.md → internal/config/config.go
- `main()` --calls--> `Load()`  [EXTRACTED]
  cmd/automation-hub/main.go → internal/config/config.go
- `main()` --calls--> `NewIMAPSource()`  [EXTRACTED]
  cmd/automation-hub/main.go → internal/services/email/imap.go
- `main()` --calls--> `NewPoller()`  [EXTRACTED]
  cmd/automation-hub/main.go → internal/services/email/poller.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Dependabot Multi-Ecosystem Update Strategy (gomod, docker, github-actions)** — _github_dependabot_gomod_updates, _github_dependabot_docker_updates, _github_dependabot_github_actions_updates [EXTRACTED 1.00]

## Communities (135 total, 9 thin omitted)

### Community 0 - "UpdateType"
Cohesion: 0.19
Nodes (14): CopyView, CountChip, EntryView, FileGroupView, RepoGroupView, UpdateType, chipsOf(), countKey() (+6 more)

### Community 1 - "Requirements"
Cohesion: 0.09
Nodes (22): bot-command-registry Specification, Purpose, Requirement: Bot texts from the message catalog, Requirement: Command routing by name, Requirement: Failure isolation in the bot, Requirement: No client for unconfigured types, Requirement: Per-entry authorization, Requirement: Single publication of Telegram commands (+14 more)

### Community 2 - "Requirement: Document tabs"
Cohesion: 0.04
Nodes (45): ADDED Requirements, dependency-report Specification, MODIFIED Requirements, Requirement: Authenticated endpoint, Requirement: Document tabs, Requirement: HTML report document, Requirement: Renovate failure message, Requirement: Request handling (+37 more)

### Community 3 - "🤖 Automation Hub"
Cohesion: 0.13
Nodes (15): README, Authors and Acknowledgment, 🤖 Automation Hub, Common Issues, 🔧 External Service Setup, ✨ Features, Gmail Configuration, 📄 License (+7 more)

### Community 4 - "config.go"
Cohesion: 0.07
Nodes (58): GitHubConfig, GitHubHTTPConfig, HTTPConfig, KubernetesConfig, KubernetesHTTPConfig, namedDuration, ServerConfig, TelegramConfig (+50 more)

### Community 5 - "Catalog"
Cohesion: 0.05
Nodes (42): CaptionPresenter, DocumentPresenter, fakeSender, renderer, failingRoundTripper, net/http.Handler, net/http.Request, net/http.Response (+34 more)

### Community 6 - "ADDED Requirements"
Cohesion: 0.09
Nodes (22): ADDED Requirements, Purpose, Requirement: Exact run link, Requirement: Failure notification for scheduled dispatch, Requirement: Fixed local timezone, Requirement: Graceful stop, Requirement: Optional schedule per workflow command, Requirement: Scheduled dispatch uses the command's dispatch path (+14 more)

### Community 7 - "NewCommandSpecs"
Cohesion: 0.10
Nodes (36): go.uber.org/zap/zaptest/observer.ObservedLogs, sync.Mutex, fakeDispatcher, fakeJobCreator, fakeMessenger, NewCommandSpecs(), assertScheduledNotices(), TestAllowList() (+28 more)

### Community 8 - "generic_test.go"
Cohesion: 0.06
Nodes (38): regexp.Regexp, ServiceProcessorConfig, compileAll(), matchCode(), newCodeExtractor(), newMarkerExtractor(), reportResult(), newTestExtractor() (+30 more)

### Community 9 - "Requirements"
Cohesion: 0.07
Nodes (26): Purpose, Requirement: Configured local timezone, Requirement: Exact run link, Requirement: Failure notification for scheduled dispatch, Requirement: Graceful stop, Requirement: Optional schedule per workflow command, Requirement: Scheduled dispatch uses the command's dispatch path, Requirement: Startup schedule log (+18 more)

### Community 10 - "ADDED Requirements"
Cohesion: 0.04
Nodes (46): ADDED Requirements, Purpose, Requirement: Body preparation, Requirement: Extractor configuration validation, Requirement: Malformed message handling, Requirement: Mark-as-read policy, Requirement: Marker extraction, Requirement: Not-found text (+38 more)

### Community 11 - "github.com/a-h/templ.Component"
Cohesion: 0.21
Nodes (14): github.com/a-h/templ.Component, ToneChip(), ToneDot(), TypeChip(), TypeDot(), CopyButton(), DepRow(), FileGroup() (+6 more)

### Community 12 - "document_view_test.go"
Cohesion: 0.53
Nodes (11): groupsOf(), layout(), TestGroupsEntryFields(), TestGroupsLayout(), TestGroupsPerTabCounts(), TestGroupsSectionOrderAndFullName(), TestGroupsSummaryRow(), TestSummaryLinesSingularAndPlural() (+3 more)

### Community 13 - "ADDED Requirements"
Cohesion: 0.09
Nodes (22): ADDED Requirements, Purpose, Requirement: All up to date, Requirement: Authenticated endpoint, Requirement: Delivery, Requirement: Per-repository message, Requirement: Renovate failure message, Requirement: Request handling (+14 more)

### Community 19 - "Requirements"
Cohesion: 0.22
Nodes (8): dependency-report Specification, Purpose, Requirement: Nothing pending document, Requirement: Renovate failure message, Requirements, Scenario: All up to date, Scenario: Exit code 1, Scenario: Title override

### Community 20 - "ADDED Requirements"
Cohesion: 0.09
Nodes (21): ADDED Requirements, Purpose, Requirement: Bot texts from the message catalog, Requirement: Command routing by name, Requirement: Failure isolation in the bot, Requirement: No client for unconfigured types, Requirement: Per-entry authorization, Requirement: Single publication of Telegram commands (+13 more)

### Community 22 - "dependency_report_test.go"
Cohesion: 0.14
Nodes (33): fakeSender, gatedSender, goldenSends, reportHandlerCase, reportRequest, sentCall, goldenOutputs(), readGolden() (+25 more)

### Community 23 - "ADDED Requirements"
Cohesion: 0.15
Nodes (12): ADDED Requirements, Purpose, Requirement: Create a Job from a CronJob, Requirement: Job command configuration, Requirement: Scheduled job commands, Scenario: Duplicate across types, Scenario: Jobs-only config starts the bot, Scenario: Kubernetes API error (+4 more)

### Community 24 - "Proposal"
Cohesion: 0.22
Nodes (8): Capabilities, Impact, Modified Capabilities, New Capabilities, Non-goals, Proposal, What Changes, Why

### Community 25 - "Tasks"
Cohesion: 0.25
Nodes (7): 1. Dependency and timezone, 2. Config, 3. Dispatch core, 4. Scheduler, 5. Cross-repo, 6. Validation, Tasks

### Community 26 - "Design"
Cohesion: 0.29
Nodes (6): Context, Decisions, Design, Goals / Non-Goals, Migration Plan, Risks / Trade-offs

### Community 27 - "Requirement: Create a Job from a CronJob"
Cohesion: 0.14
Nodes (13): job-commands Specification, Purpose, Requirement: Create a Job from a CronJob, Requirement: Job command configuration, Requirement: Scheduled job commands, Requirements, Scenario: Duplicate across types, Scenario: Jobs-only config starts the bot (+5 more)

### Community 28 - "Requirement: Document tabs"
Cohesion: 0.05
Nodes (40): ADDED Requirements, MODIFIED Requirements, REMOVED Requirements, Requirement: All up to date, Requirement: Delivery, Requirement: Document tabs, Requirement: HTML report document, Requirement: Nothing pending document (+32 more)

### Community 29 - "ADDED Requirements"
Cohesion: 0.05
Nodes (38): ADDED Requirements, Purpose, Requirement: Bounded response bodies, Requirement: Configurable GitHub API base URL, Requirement: Configurable server timeouts, Requirement: Duration keys carry a unit, Requirement: Shared outbound HTTP settings, Requirement: Telegram long-poll timeouts (+30 more)

### Community 30 - "Proposal"
Cohesion: 0.22
Nodes (8): Capabilities, Impact, Modified Capabilities, New Capabilities, Non-goals, Proposal, What Changes, Why

### Community 31 - "Load"
Cohesion: 0.15
Nodes (22): TestBotTextDefaults(), applyGroup(), applyOverrides(), checkPlaceholders(), flatten(), hasGroup(), join(), Load() (+14 more)

### Community 32 - "Design"
Cohesion: 0.25
Nodes (7): Context, Decisions, Design, Goals / Non-Goals, Migration Plan, Open Questions, Risks / Trade-offs

### Community 33 - "Tasks"
Cohesion: 0.25
Nodes (7): 1. Pre-checks, 2. Config and wiring, 3. Job commands, 4. Dependency report, 5. Cross-repo, 6. Validation, Tasks

### Community 34 - "testing.T"
Cohesion: 0.13
Nodes (23): testing.T, assertJobsAndReportLoad(), assertScheduleError(), Config, loadExampleConfig(), TestLoadDefaults(), TestLoadErrorMissingConfig(), TestLoadExampleConfig() (+15 more)

### Community 35 - "NewTorrentProcessor"
Cohesion: 0.38
Nodes (5): messageSender, NewTorrentProcessor(), TestNewTorrentProcessor(), TestTorrentProcessor_Process(), TorrentProcessor

### Community 36 - "Requirements"
Cohesion: 0.04
Nodes (48): email-code-forwarding Specification, Purpose, Requirement: Body preparation, Requirement: Email from configuration, Requirement: Extractor configuration validation, Requirement: Malformed message handling, Requirement: Mark-as-read policy, Requirement: Marker extraction (+40 more)

### Community 37 - "Requirement: Configurable readiness wait"
Cohesion: 0.12
Nodes (16): Purpose, Requirement: Configurable readiness wait, Requirement: Restart a Deployment from Telegram, Requirement: Wait runs in the background, Requirements, restart-commands Specification, Scenario: Custom poll interval, Scenario: Defaults without the section (+8 more)

### Community 38 - "telegram/client_test.go"
Cohesion: 0.09
Nodes (41): fakeMenu, Client, github.com/go-telegram-bot-api/telegram-bot-api/v5.BotAPI, github.com/go-telegram-bot-api/telegram-bot-api/v5.BotCommand, github.com/go-telegram-bot-api/telegram-bot-api/v5.HTTPClient, mime/multipart.FileHeader, mime/multipart.Form, net/http.HandlerFunc (+33 more)

### Community 39 - "Requirement: Configurable readiness wait"
Cohesion: 0.12
Nodes (15): ADDED Requirements, Purpose, Requirement: Configurable readiness wait, Requirement: Restart a Deployment from Telegram, Requirement: Wait runs in the background, Scenario: Custom poll interval, Scenario: Defaults without the section, Scenario: Kubernetes API error (+7 more)

### Community 40 - "Requirement: Document tabs"
Cohesion: 0.17
Nodes (12): Requirement: Document tabs, Scenario: Digest only, Scenario: Dockerfile image and missing datasource, Scenario: Document fork tag override, Scenario: Exact duplicates, Scenario: Fork tag, Scenario: Grouping order, Scenario: Image dependency (+4 more)

### Community 41 - "buildFailure"
Cohesion: 0.24
Nodes (13): BuildFailure(), escapedSuffix(), keptLines(), logTailBlock(), assertKeepsLastLines(), buildFailure(), TestBuildFailureMessageEmptyTail(), TestBuildFailureMessageEscapesPod() (+5 more)

### Community 42 - "kube/client_test.go"
Cohesion: 0.19
Nodes (20): net/http/httptest.Server, NewClientWithBaseURL(), assertPostedJob(), assertRestartDeploymentRequest(), deploymentJSON(), newCronJobAPIServer(), runCreateJobCase(), TestCreateJobFromCronJob() (+12 more)

### Community 43 - "Decisions"
Cohesion: 0.12
Nodes (15): Context, D1. Two new packages, the handler stays thin, D2. Presentation strategy, D3. Typed update types and per-tab view models, D4. Timezone and dependency injection, D5. Config defaults applied in Go, only when the section exists, D6. Message catalog, D7. Visible default change (+7 more)

### Community 44 - "ADDED Requirements"
Cohesion: 0.12
Nodes (15): ADDED Requirements, message-catalog Specification, Purpose, Requirement: Configured overrides, Requirement: Embedded default catalog, Requirement: Placeholders are filled by code, Requirement: Startup validation of the catalog, Scenario: Caption header override (+7 more)

### Community 45 - "Requirements"
Cohesion: 0.12
Nodes (16): message-catalog Specification, Purpose, Requirement: Configured overrides, Requirement: Embedded default catalog, Requirement: Placeholders are filled by code, Requirement: Startup validation of the catalog, Requirements, Scenario: Caption header override (+8 more)

### Community 46 - "Design"
Cohesion: 0.20
Nodes (9): Context, Decisions, Decisions (resolved), Design, Goals / Non-Goals, Implementation notes, Migration Plan, Renovate 44.151.0 findings (read from source) (+1 more)

### Community 47 - "Requirements"
Cohesion: 0.05
Nodes (40): http-transport Specification, Purpose, Requirement: Bounded response bodies, Requirement: Configurable GitHub API base URL, Requirement: Configurable server timeouts, Requirement: Duration keys carry a unit, Requirement: Shared outbound HTTP settings, Requirement: Telegram long-poll timeouts (+32 more)

### Community 48 - "Proposal"
Cohesion: 0.20
Nodes (9): Capabilities, Dependency, Impact, Modified Capabilities, New Capabilities, Non-goals, Proposal, What Changes (+1 more)

### Community 49 - "documentView"
Cohesion: 0.15
Nodes (14): documentView, SummaryLine, TabView, TestFaviconIsDataURI(), labelSet, faviconHref(), Logo(), Header() (+6 more)

### Community 50 - "Tasks"
Cohesion: 0.15
Nodes (12): 10. Cross-repo and user steps, 11. Validation, 1. Baseline, 2. Message catalog loader (`internal/messages`), 3. Catalog content, 4. Configuration (`internal/config`), 5. Parsers (`internal/services/depreport`), 6. Domain (`internal/services/depreport`) (+4 more)

### Community 51 - "Proposal"
Cohesion: 0.22
Nodes (8): Capabilities, Impact, Modified Capabilities, New Capabilities, Non-goals, Proposal, What Changes, Why

### Community 52 - "Tasks"
Cohesion: 0.22
Nodes (8): 1. Pre-checks, 2. Telegram client, 3. Report document, 4. Handler wiring and docs, 5. Cross-repo: scan homelab and retire icr, 6. Caption and single message, 7. Validation, Tasks

### Community 53 - "Requirement: Exact run link"
Cohesion: 0.20
Nodes (9): MODIFIED Requirements, Requirement: Exact run link, Scenario: Custom GitHub web URL, Scenario: Invalid GitHub web URL, Scenario: No run details, Scenario: Not a scheduled day, Scenario: Overlapping dispatches, Scenario: Run details returned (+1 more)

### Community 54 - "Decisions"
Cohesion: 0.06
Nodes (30): Context, D10. Copy buttons and handler (amended), D11. Generated files are committed and checked, D12. Catalog keys (amended), D13. Errors and empty state, D14. Checks without a browser (amended), D15. Local gate `scripts/verify.sh`, D16. Biome as a standalone binary (+22 more)

### Community 55 - "Decisions"
Cohesion: 0.11
Nodes (17): Context, Coordination with sibling changes, D1. Extraction is a strategy chosen by `extractor`, with two implementations, D2. Built-in name defaults are removed; configs carry the values, D3. Defaults live in `config`, applied through accessors, D4. Regexes are validated at load and compiled once at construction, D5. `messageSender` and `TemplateNotifier`, D6. `GenericEmailProcessor` is a composition (+9 more)

### Community 57 - "imap_test.go"
Cohesion: 0.13
Nodes (20): fakeSession, imap.FetchItem, imap.MailboxStatus, imap.SeqSet, imap.StoreItem, imap.Message, imap.SearchCriteria, inSeqSet() (+12 more)

### Community 58 - "RetryingSender"
Cohesion: 0.39
Nodes (5): github.com/go-telegram-bot-api/telegram-bot-api/v5.Chattable, NewRetryingSender(), Delayer, RetryingSender, sender

### Community 59 - "settings_test.go"
Cohesion: 0.18
Nodes (16): expectSettingError(), loadZone(), TestDependencyReportAbsentStaysNil(), TestDependencyReportCustomSettings(), TestDependencyReportDefaults(), TestDependencyReportDocumentNameNeedsDate(), TestDependencyReportInvalidPath(), TestDependencyReportNonNumericBodyLimit() (+8 more)

### Community 60 - "main_test.go"
Cohesion: 0.14
Nodes (18): fakeAdder, noopSender, routeEntry, mountRoute(), registerRoutes(), jobPost(), routesFor(), routesWith() (+10 more)

### Community 61 - "NewWebhookHandler"
Cohesion: 0.22
Nodes (11): messageSender, WebhookHandler, webhookSender, WebhookProcessorConfig, NewWebhookHandler(), TestHandleTorrentComplete_InvalidJSON(), TestHandleTorrentComplete_Success(), TestHandleTorrentCompleteProcessingError() (+3 more)

### Community 62 - "go.uber.org/zap.Logger"
Cohesion: 0.35
Nodes (13): menuPublisher, scheduleAdder, main(), newGitHubClient(), newHTTPServer(), newKubeClient(), newTelegramClient(), registerCommands() (+5 more)

### Community 63 - "Tasks"
Cohesion: 0.22
Nodes (8): 1. Prerequisite gate, 2. Configuration, 3. Message texts, 4. Command registry, 5. Scheduling and menu wiring, 6. Webhook registry, 7. Validation, Tasks

### Community 65 - "Design"
Cohesion: 0.25
Nodes (7): Context, Decisions, Design, Goals / Non-Goals, Migration Plan, Open Questions, Risks / Trade-offs

### Community 66 - "Requirement: Create a Job from a CronJob"
Cohesion: 0.29
Nodes (6): MODIFIED Requirements, Requirement: Create a Job from a CronJob, Scenario: Kubernetes API error, Scenario: Successful creation, Scenario: Unauthorized chat, Spec Delta

### Community 67 - "components_test.go"
Cohesion: 0.29
Nodes (18): expectContains(), fixtureView(), muxEntry(), renderOf(), TestCopyButtonStates(), TestDepRowColumnsAndCopy(), TestEntryEscapesValues(), TestErrorsCardRows() (+10 more)

### Community 68 - "Proposal"
Cohesion: 0.22
Nodes (8): Capabilities, Impact, Modified Capabilities, New Capabilities, Non-goals, Proposal, What Changes, Why

### Community 69 - "Proposal"
Cohesion: 0.20
Nodes (9): Capabilities, Dependency, Impact, Modified Capabilities, New Capabilities, Non-goals, Proposal, What Changes (+1 more)

### Community 70 - "NewClient"
Cohesion: 0.26
Nodes (14): Client, Config, net/http.Client, NewClient(), NewClient(), newClient(), NewClientWithBaseURL(), assertDispatchWorkflowRequest() (+6 more)

### Community 71 - "Requirement: Document tabs"
Cohesion: 0.20
Nodes (10): Requirement: Document tabs, Scenario: Digest only, Scenario: Dockerfile image and missing datasource, Scenario: Document fork tag override, Scenario: Exact duplicates, Scenario: Fork tag, Scenario: Image dependency, Scenario: Non-image dependency (+2 more)

### Community 72 - "summaryAt"
Cohesion: 0.06
Nodes (76): Dep, depKey, ErrorView, PackageFile, Problem, RepoPending, RepoReport, Row (+68 more)

### Community 73 - "context.Context"
Cohesion: 0.31
Nodes (5): context.Context, blockingRestarter, resourceURL(), Client, deploymentStatus

### Community 74 - "Requirement: Configured local timezone"
Cohesion: 0.25
Nodes (7): MODIFIED Requirements, RENAMED Requirements, Requirement: Configured local timezone, Scenario: Configured zone, Scenario: Image without tzdata, Scenario: Invalid zone, scheduled-dispatch Specification

### Community 75 - "Email"
Cohesion: 0.05
Nodes (35): Dispatcher, fakeDispatcher, fakeSource, imapSession, IMAPSource, markingProcessor, plainProcessor, Poller (+27 more)

### Community 76 - "Requirement: Scheduled job commands"
Cohesion: 0.33
Nodes (5): job-commands Specification, MODIFIED Requirements, Requirement: Scheduled job commands, Scenario: Weekly report, Scenario: Weekly report in a configured zone

### Community 77 - "Requirement: HTML report document"
Cohesion: 0.29
Nodes (7): Requirement: HTML report document, Scenario: Configured timezone changes the date and the file name, Scenario: Date near midnight UTC, Scenario: Defaults for file name, header and labels, Scenario: Renovate failed, Scenario: Report file missing, Scenario: Successful run

### Community 78 - "Requirement: Caption repository lines"
Cohesion: 0.40
Nodes (5): Requirement: Caption repository lines, Scenario: Escaped repository name, Scenario: Fork and sort order, Scenario: Major, minor and patch counts, Scenario: Owner-agnostic name

### Community 79 - "command"
Cohesion: 0.17
Nodes (7): allowList, command, defaultCatalog(), BotHandlerDeps, BotHandlerDeps, newCommand(), newReplier()

### Community 80 - "Requirement: Authenticated endpoint"
Cohesion: 0.40
Nodes (5): Requirement: Authenticated endpoint, Scenario: Custom path, Scenario: Custom token header, Scenario: Missing token, Scenario: Not configured

### Community 81 - "Requirement: Webhook settings"
Cohesion: 0.40
Nodes (5): Requirement: Webhook settings, Scenario: Defaults, Scenario: Document name without the date, Scenario: Invalid path, Scenario: Zero body limit

### Community 82 - "Requirement: Delivery"
Cohesion: 0.50
Nodes (4): Requirement: Delivery, Scenario: Document rejected, Scenario: Renovate failed, Scenario: Telegram down

### Community 83 - "Requirement: Request handling"
Cohesion: 0.50
Nodes (4): Requirement: Request handling, Scenario: Body over the limit, Scenario: Malformed exit code, Scenario: Valid request

### Community 84 - "Requirement: Scan errors in the document"
Cohesion: 0.50
Nodes (4): Requirement: Scan errors in the document, Scenario: Lookup failure, Scenario: Root problem, Scenario: Warning level only

### Community 85 - "workflowCommand"
Cohesion: 0.43
Nodes (3): workflowCommand, BotHandlerDeps, newWorkflowCommand()

### Community 86 - "Requirement: Safe HTML and size limit"
Cohesion: 0.67
Nodes (3): Requirement: Safe HTML and size limit, Scenario: Caption over the limit, Scenario: Long escaped content

### Community 87 - "jobCommand"
Cohesion: 0.43
Nodes (3): jobCommand, BotHandlerDeps, newJobCommand()

### Community 88 - "NewScheduledRunner"
Cohesion: 0.27
Nodes (7): TestJobSchedulesUseConfiguredZone(), TestRegisterCommandsMenuRejectedReturnsError(), TestRegisterCommandsOneLoop(), CommandSpec, BotHandlerDeps, ScheduledRunner, NewScheduledRunner()

### Community 89 - "kube/client.go"
Cohesion: 0.22
Nodes (14): encoding/json.RawMessage, jobFromTemplate(), newClient(), NewInClusterClient(), TestInClusterSettingsUseCA(), TestNewInClusterClientOutsideCluster(), withServiceAccountCA(), apiCall (+6 more)

### Community 90 - "loadYAML"
Cohesion: 0.14
Nodes (23): loadYAML(), TestEffectiveDefaults(), TestLoad_EmailDefaults(), TestLoad_MigratedEmailServices(), TestLoad_NegativeMaxBodyBytes(), TestLoad_NegativePollingInterval(), TestShouldMarkAsRead(), TestValidateEmailServices() (+15 more)

### Community 91 - "Design"
Cohesion: 0.29
Nodes (6): Context, Decisions, Design, Goals / Non-Goals, Migration Plan, Risks / Trade-offs

### Community 92 - "Tasks"
Cohesion: 0.29
Nodes (6): 1. Shared HTTP package and configuration, 2. Telegram client, 3. GitHub client, 4. Kubernetes client, 5. Final validation, Tasks

### Community 93 - "Requirement: HTML report document"
Cohesion: 0.22
Nodes (9): Requirement: HTML report document, Scenario: Configured timezone changes the date and the file name, Scenario: Date near midnight UTC, Scenario: Defaults for file name, header and labels, Scenario: Fonts blocked, Scenario: Renovate failed, Scenario: Report file missing, Scenario: Self-contained document (+1 more)

### Community 95 - "Scheduler"
Cohesion: 0.23
Nodes (7): lazyScheduler, cron.Cron, cron.EntryID, cron.ScheduleParser, Scheduler, New(), newWithParser()

### Community 96 - "Requirement: Body size cap"
Cohesion: 0.13
Nodes (14): ADDED Requirements, Purpose, Requirement: Body size cap, Requirement: Completion notification, Requirement: Hook configuration lookup, Scenario: Body over the default cap, Scenario: Body within the default cap, Scenario: Completed download (+6 more)

### Community 97 - "Requirement: Body size cap"
Cohesion: 0.13
Nodes (14): Purpose, Requirement: Body size cap, Requirement: Completion notification, Requirement: Hook configuration lookup, Requirements, Scenario: Body over the default cap, Scenario: Body within the default cap, Scenario: Completed download (+6 more)

### Community 98 - "restartCommand"
Cohesion: 0.32
Nodes (4): go.uber.org/zap.Field, restartCommand, BotHandlerDeps, newRestartCommand()

### Community 99 - "Tasks"
Cohesion: 0.12
Nodes (15): 10. README, 11. Final gate (Docker), 12. Deploy and check (user), 13. Batch 3: liquid glass port (2026-10-12), 1. Toolchain, 2. Styles, 3. Static prototype (approval before any presenter change), 4. View model regrouping (+7 more)

### Community 100 - "document.go"
Cohesion: 0.21
Nodes (13): BarSegment, docBuilder, LegendItem, StatsView, Tab, tabSpec, labelSet, newLabels() (+5 more)

### Community 101 - "BotHandlerDeps"
Cohesion: 0.20
Nodes (12): github.com/go-telegram-bot-api/telegram-bot-api/v5.Message, github.com/go-telegram-bot-api/telegram-bot-api/v5.Update, DeploymentRestarter, JobCreator, replier, TelegramMessenger, BotHandler, BotHandlerDeps (+4 more)

### Community 102 - "🐳 Deployment"
Cohesion: 0.29
Nodes (8): Deploy Job: Build, Push & Deploy, Deploy Job: Resolve Build Args, Devidence CD Build Deploy Reusable Workflow, Build and Deploy Workflow, 🐳 Deployment, 🚀 Option 1: Docker Compose (Recommended), 🔧 Option 2: Manual Docker, � Option 3: Local Development

### Community 103 - "Settings"
Cohesion: 0.29
Nodes (5): crypto/tls.Config, net.Dialer, Settings, newDialer(), TestNewClientAppliesSettings()

### Community 104 - "verify.sh"
Cohesion: 0.31
Nodes (12): check_biome(), check_cognitive(), check_drift(), check_gofmt(), check_lint(), check_openspec(), check_tests(), fix() (+4 more)

### Community 105 - "biome.json"
Cohesion: 0.18
Nodes (10): css, parser, files, includes, formatter, enabled, linter, enabled (+2 more)

### Community 106 - "retry_test.go"
Cohesion: 0.26
Nodes (14): TestNewClientWithoutEnvProxyOrHTTP2(), TestSendCapsBodyAtLimit(), TestSendReturnsStatusAndBody(), apiError(), rateLimited(), retryTest(), testBackoff(), TestBackoffNoOverflowAtHighAttempts() (+6 more)

### Community 107 - "scheduler_test.go"
Cohesion: 0.31
Nodes (8): cron.Schedule, ParseSpec(), mustLocation(), TestAddRejectsInvalidSpec(), TestNextRunIsInGuayaquil(), TestParseSpec(), TestSchedulerReportsNextRunInConfiguredZone(), TestSchedulerRunsUntilStopped()

### Community 108 - "styles_test.go"
Cohesion: 0.35
Nodes (12): canonHex(), contrastRatio(), cssCompact(), cssTokens(), relLuminance(), schemeBlocks(), TestReportCSSDarkTokensInBothBlocks(), TestReportCSSHasNoOldFrameworkOrWideMinWidth() (+4 more)

### Community 109 - "Send"
Cohesion: 0.39
Nodes (4): net/http.Header, Request, Response, Send()

### Community 110 - "Proposal"
Cohesion: 0.22
Nodes (8): Capabilities, Impact, Modified Capabilities, New Capabilities, Non-goals, Proposal, What Changes, Why

### Community 111 - "Tasks"
Cohesion: 0.22
Nodes (8): 1. Config package, 2. Processor package, 3. Email package, 4. Torrent webhook, 5. Quality gates, 6. User and other-repo steps, 7. Final validation, Tasks

### Community 112 - "redesign-dependency-report-document/proposal.md"
Cohesion: 0.40
Nodes (4): Non-goals, Superseded decisions, What Changes, Why

### Community 113 - "Requirement: Embedded default catalog"
Cohesion: 0.22
Nodes (8): MODIFIED Requirements, Requirement: Embedded default catalog, Scenario: Catalog contents, Scenario: Default fork note and chips, Scenario: Default language, Scenario: Defaults render today's caption, Scenario: Unused emoji key override still validates, Scenario: Unused key override still validates

### Community 114 - "generate-depreport-assets.sh"
Cohesion: 0.52
Nodes (6): arch(), fetch(), generate(), main(), generate-depreport-assets.sh script, verify_sha256()

### Community 115 - "ADDED Requirements"
Cohesion: 0.18
Nodes (11): ADDED Requirements, Requirement: Entry layout by viewport width, Requirement: Responsive layout, Requirement: Theme control, Scenario: Aligned rows on a wide viewport, Scenario: Choice lasts for the page view, Scenario: No storage access, Scenario: Opened after a dark choice (+3 more)

### Community 117 - "keys_test.go"
Cohesion: 0.67
Nodes (5): defaultCatalog(), ownKeys(), TestDefaultsHaveEveryKey(), TestDefaultsKeepPlaceholders(), TestNewKeysRender()

### Community 118 - "StatsPanel"
Cohesion: 0.53
Nodes (4): DistributionBar(), Legend(), StatsPanel(), toneOf()

### Community 119 - "Requirement: Document theme"
Cohesion: 0.40
Nodes (5): Requirement: Document theme, Scenario: Chip colors of one type, Scenario: Dark tokens in both dark blocks, Scenario: Light tokens, Scenario: Text contrast

### Community 124 - "Requirement: Copy buttons"
Cohesion: 0.40
Nodes (5): Requirement: Copy buttons, Scenario: Clipboard API blocked, Scenario: Dependency copy, Scenario: Image copy, Scenario: Image entry has one copy button

### Community 125 - "Requirement: Repo summary row"
Cohesion: 0.40
Nodes (5): Requirement: Repo summary row, Scenario: Summary counts match the caption, Scenario: Summary row of a fork, Scenario: Summary row with a major update, Scenario: Summary row without a major update

### Community 126 - "Requirement: Summary caption"
Cohesion: 0.40
Nodes (5): Requirement: Summary caption, Scenario: Caption over the limit, Scenario: Error count line, Scenario: No listed update, Scenario: Stats lines across both tabs

### Community 127 - "� API & Webhooks"
Cohesion: 0.40
Nodes (5): Adding New Email Services, Adding New Webhooks, � API & Webhooks, Available Endpoints, qBittorrent Integration

### Community 128 - "CI Job: Devidence Go CI"
Cohesion: 0.50
Nodes (4): CI Job: Devidence Go CI, CI Pipeline Workflow, Devidence Go CI Reusable Workflow, 🛡️ Security & Quality

### Community 130 - "MODIFIED Requirements"
Cohesion: 0.22
Nodes (8): MODIFIED Requirements, Requirement: Nothing pending document, Requirement: Scan errors in the document, Scenario: All up to date, Scenario: Lookup failure, Scenario: Root problem, Scenario: Three errors, Scenario: Warning level only

### Community 131 - "Requirement: Repo sections"
Cohesion: 0.50
Nodes (4): Requirement: Repo sections, Scenario: All repos collapsed on load, Scenario: Opening one keeps others open, Scenario: Reopening the file starts collapsed

### Community 132 - "Requirement: Summary panel"
Cohesion: 0.50
Nodes (4): Requirement: Summary panel, Scenario: Distribution bar of the deps tab, Scenario: Summary line, Scenario: Zero segment omitted

### Community 133 - "Requirement: Safe document content"
Cohesion: 0.67
Nodes (3): Requirement: Safe document content, Scenario: Quote in a version, Scenario: Script in a dependency name

### Community 134 - "Requirement: Motion"
Cohesion: 0.67
Nodes (3): Requirement: Motion, Scenario: Panel enters, Scenario: Reduced motion

### Community 138 - "Requirement: Header and logo"
Cohesion: 0.67
Nodes (3): Requirement: Header and logo, Scenario: Logo and icon are clean, Scenario: Logo follows the theme

## Knowledge Gaps
- **681 isolated node(s):** `$schema`, `enabled`, `enabled`, `tailwindDirectives`, `includes` (+676 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 817 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **9 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `NewCommandSpecs()` connect `NewCommandSpecs` to `restartCommand`, `command`, `workflowCommand`, `jobCommand`, `NewScheduledRunner`, `go.uber.org/zap.Logger`?**
  _High betweenness centrality (0.020) - this node is a cross-community bridge._
- **Why does `Config` connect `config.go` to `main_test.go`, `Catalog`, `go.uber.org/zap.Logger`?**
  _High betweenness centrality (0.019) - this node is a cross-community bridge._
- **Why does `Summary` connect `summaryAt` to `document.go`, `Catalog`?**
  _High betweenness centrality (0.011) - this node is a cross-community bridge._
- **Are the 26 inferred relationships involving `loadYAML()` (e.g. with `loadFrom()` and `TestLoad_EmailDefaults()`) actually correct?**
  _`loadYAML()` has 26 INFERRED edges - model-reasoned connections that need verification._
- **Are the 28 inferred relationships involving `summaryAt()` (e.g. with `TestFaviconLinkIsNotSanitized()` and `TestReportScriptEmbedded()`) actually correct?**
  _`summaryAt()` has 28 INFERRED edges - model-reasoned connections that need verification._
- **What connects `$schema`, `enabled`, `enabled` to the rest of the system?**
  _681 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Requirements` be split into smaller, more focused modules?**
  _Cohesion score 0.08695652173913043 - nodes in this community are weakly interconnected._