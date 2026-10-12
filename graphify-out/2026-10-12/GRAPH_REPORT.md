# Graph Report - automation-hub  (2026-10-12)

## Corpus Check
- 112 files · ~90,077 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 1522 nodes · 2865 edges · 102 communities (89 shown, 12 thin omitted)
- Extraction: 92% EXTRACTED · 8% INFERRED · 0% AMBIGUOUS · INFERRED: 233 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `91fb4bef`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- summaryAt
- Requirements
- Requirement: Document tabs
- 🤖 Automation Hub
- config.go
- IMAPClient
- ADDED Requirements
- NewCommandSpecs
- NewGenericEmailProcessor
- Requirements
- ADDED Requirements
- NewClient
- dependency_report_test.go
- ADDED Requirements
- CLAUDE.md
- automation-hub-network Bridge Network
- Dependabot Docker Update Config (deployments/docker)
- Dependabot GitHub Actions Update Config
- Dependabot Go Modules Update Config
- Requirements
- ADDED Requirements
- Graphify Knowledge-Graph Workflow Rules
- net/http.Request
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
- NewWebhookHandler
- Decisions
- Requirement: Configurable readiness wait
- telegram/client_test.go
- Requirement: Configurable readiness wait
- Requirement: Body size cap
- buildFailure
- kube/client_test.go
- Decisions
- ADDED Requirements
- Requirements
- Design
- Catalog
- Proposal
- main_test.go
- Tasks
- Proposal
- Tasks
- Requirement: Exact run link
- Proposal
- Tasks
- automation-hub
- Proposal
- document.go
- settings_test.go
- Scheduler
- Design
- Tasks
- Tasks
- Design
- Requirement: Create a Job from a CronJob
- kube/client.go
- defaultCatalog
- Proposal
- aggregate.go
- Requirement: Document tabs
- RouteEnv
- retry_test.go
- Requirement: Configured local timezone
- Client
- Requirement: Scheduled job commands
- Requirement: HTML report document
- Requirement: Summary caption
- command
- Requirement: Authenticated endpoint
- Requirement: Webhook settings
- Requirement: Delivery
- Requirement: Request handling
- Requirement: Scan errors in the document
- Requirement: Renovate failure message
- Requirement: Safe HTML and size limit
- context.Context
- BotHandlerDeps
- NewDeliverer
- go.uber.org/zap.Logger
- DependencyReportHandler
- Summary
- RetryingSender
- dependency_report_golden_test.go
- workflowCommand
- UpdateType
- fakeSender
- restartCommand
- fakeMenu
- models_test.go
- TestParseUpdateType

## God Nodes (most connected - your core abstractions)
1. `Config` - 29 edges
2. `Client` - 29 edges
3. `loadYAML()` - 28 edges
4. `NewCommandSpecs()` - 23 edges
5. `Catalog` - 23 edges
6. `summaryAt()` - 22 edges
7. `DependencyReportHandler` - 20 edges
8. `loadFixtureReport()` - 19 edges
9. `NewClientWithBaseURL()` - 19 edges
10. `Aggregate()` - 18 edges

## Surprising Connections (you probably didn't know these)
- `📝 Step 2: Configure Services` --references--> `ServiceConfig`  [INFERRED]
  README.md → internal/config/config.go
- `🤖 Step 3: Setup Telegram Bot` --references--> `TelegramConfig`  [INFERRED]
  README.md → internal/config/config.go
- `▶️ Trigger GitHub Actions from Telegram` --references--> `WorkflowBotConfig`  [INFERRED]
  README.md → internal/config/config.go
- `main()` --calls--> `Load()`  [EXTRACTED]
  cmd/automation-hub/main.go → internal/config/config.go
- `main()` --calls--> `NewIMAPClient()`  [EXTRACTED]
  cmd/automation-hub/main.go → internal/services/email/imap.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Dependabot Multi-Ecosystem Update Strategy (gomod, docker, github-actions)** — _github_dependabot_gomod_updates, _github_dependabot_docker_updates, _github_dependabot_github_actions_updates [EXTRACTED 1.00]

## Communities (102 total, 12 thin omitted)

### Community 0 - "summaryAt"
Cohesion: 0.11
Nodes (46): Dep, PackageFile, Problem, RepoReport, Update, Aggregate(), repoJSON(), reportOf() (+38 more)

### Community 1 - "Requirements"
Cohesion: 0.09
Nodes (22): bot-command-registry Specification, Purpose, Requirement: Bot texts from the message catalog, Requirement: Command routing by name, Requirement: Failure isolation in the bot, Requirement: No client for unconfigured types, Requirement: Per-entry authorization, Requirement: Single publication of Telegram commands (+14 more)

### Community 2 - "Requirement: Document tabs"
Cohesion: 0.04
Nodes (45): ADDED Requirements, dependency-report Specification, MODIFIED Requirements, Requirement: Authenticated endpoint, Requirement: Document tabs, Requirement: HTML report document, Requirement: Renovate failure message, Requirement: Request handling (+37 more)

### Community 3 - "🤖 Automation Hub"
Cohesion: 0.07
Nodes (31): CI Job: Devidence Go CI, CI Pipeline Workflow, Devidence Go CI Reusable Workflow, Deploy Job: Build, Push & Deploy, Deploy Job: Resolve Build Args, Devidence CD Build Deploy Reusable Workflow, Build and Deploy Workflow, README (+23 more)

### Community 4 - "config.go"
Cohesion: 0.07
Nodes (56): GitHubConfig, GitHubHTTPConfig, HTTPConfig, KubernetesConfig, KubernetesHTTPConfig, namedDuration, ServerConfig, ServiceConfig (+48 more)

### Community 5 - "IMAPClient"
Cohesion: 0.08
Nodes (19): IMAPClient, mockNamedProcessor, github.com/emersion/go-imap/client.Client, sync.WaitGroup, imap.Literal, imap.Message, Email, EmailProcessor (+11 more)

### Community 6 - "ADDED Requirements"
Cohesion: 0.09
Nodes (22): ADDED Requirements, Purpose, Requirement: Exact run link, Requirement: Failure notification for scheduled dispatch, Requirement: Fixed local timezone, Requirement: Graceful stop, Requirement: Optional schedule per workflow command, Requirement: Scheduled dispatch uses the command's dispatch path (+14 more)

### Community 7 - "NewCommandSpecs"
Cohesion: 0.08
Nodes (46): commandMenu, scheduleAdder, registerCommands(), TestJobSchedulesUseConfiguredZone(), TestRegisterCommandsMenuRejectedReturnsError(), TestRegisterCommandsOneLoop(), go.uber.org/zap/zaptest/observer.ObservedLogs, sync.Mutex (+38 more)

### Community 8 - "NewGenericEmailProcessor"
Cohesion: 0.14
Nodes (14): regexp.Regexp, NewGenericEmailProcessor(), TestDecodeQuotedPrintable(), TestExtractCode(), TestExtractCode_IgnoresHexColor(), TestExtractPerplexityCode(), TestNewGenericEmailProcessor_BuiltInPatterns(), TestNewGenericEmailProcessor_CustomPattern() (+6 more)

### Community 9 - "Requirements"
Cohesion: 0.07
Nodes (26): Purpose, Requirement: Configured local timezone, Requirement: Exact run link, Requirement: Failure notification for scheduled dispatch, Requirement: Graceful stop, Requirement: Optional schedule per workflow command, Requirement: Scheduled dispatch uses the command's dispatch path, Requirement: Startup schedule log (+18 more)

### Community 10 - "ADDED Requirements"
Cohesion: 0.04
Nodes (45): ADDED Requirements, Purpose, Requirement: Body preparation, Requirement: Extractor configuration validation, Requirement: Malformed message handling, Requirement: Mark-as-read policy, Requirement: Marker extraction, Requirement: Not-found text (+37 more)

### Community 11 - "NewClient"
Cohesion: 0.15
Nodes (21): Client, Config, net.Dialer, net/http.Client, net/http.Header, Request, Response, NewClient() (+13 more)

### Community 12 - "dependency_report_test.go"
Cohesion: 0.23
Nodes (25): reportHandlerCase, reportRequest, sentCall, fixture(), kinds(), newReportHandler(), newReportHandlerAt(), onlyDocument() (+17 more)

### Community 13 - "ADDED Requirements"
Cohesion: 0.09
Nodes (22): ADDED Requirements, Purpose, Requirement: All up to date, Requirement: Authenticated endpoint, Requirement: Delivery, Requirement: Per-repository message, Requirement: Renovate failure message, Requirement: Request handling (+14 more)

### Community 19 - "Requirements"
Cohesion: 0.22
Nodes (8): dependency-report Specification, Purpose, Requirement: Nothing pending document, Requirement: Safe document content, Requirements, Scenario: All up to date, Scenario: Quote in a version, Scenario: Script in a dependency name

### Community 20 - "ADDED Requirements"
Cohesion: 0.09
Nodes (21): ADDED Requirements, Purpose, Requirement: Bot texts from the message catalog, Requirement: Command routing by name, Requirement: Failure isolation in the bot, Requirement: No client for unconfigured types, Requirement: Per-entry authorization, Requirement: Single publication of Telegram commands (+13 more)

### Community 22 - "net/http.Request"
Cohesion: 0.17
Nodes (14): failingRoundTripper, net/http.Request, net/http.Response, Submission, ParseSubmission(), readFile(), multipartRequest(), TestParseSubmissionBadExitCode() (+6 more)

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
Cohesion: 0.14
Nodes (23): defaultCatalog(), BotHandlerDeps, TestBotTextDefaults(), applyGroup(), applyOverrides(), checkPlaceholders(), flatten(), hasGroup() (+15 more)

### Community 32 - "Design"
Cohesion: 0.25
Nodes (7): Context, Decisions, Design, Goals / Non-Goals, Migration Plan, Open Questions, Risks / Trade-offs

### Community 33 - "Tasks"
Cohesion: 0.25
Nodes (7): 1. Pre-checks, 2. Config and wiring, 3. Job commands, 4. Dependency report, 5. Cross-repo, 6. Validation, Tasks

### Community 34 - "testing.T"
Cohesion: 0.14
Nodes (33): testing.T, assertJobsAndReportLoad(), assertScheduleError(), Config, loadYAML(), TestLoadDefaults(), TestLoadErrorMissingConfig(), TestLoadExampleConfig() (+25 more)

### Community 35 - "NewWebhookHandler"
Cohesion: 0.17
Nodes (14): WebhookHandler, WebhookProcessorConfig, NewWebhookHandler(), TestHandleTorrentComplete_InvalidJSON(), TestHandleTorrentComplete_Success(), TestNewWebhookHandler(), GetWebhookConfig(), NewTorrentProcessor() (+6 more)

### Community 36 - "Decisions"
Cohesion: 0.11
Nodes (17): Context, Coordination with sibling changes, D1. Extraction is a strategy chosen by `extractor`, with two implementations, D2. Built-in name defaults are removed; configs carry the values, D3. Defaults live in `config`, applied through accessors, D4. Regexes are validated at load and compiled once at construction, D5. `messageSender` and `TemplateNotifier`, D6. `GenericEmailProcessor` is a composition (+9 more)

### Community 37 - "Requirement: Configurable readiness wait"
Cohesion: 0.12
Nodes (16): Purpose, Requirement: Configurable readiness wait, Requirement: Restart a Deployment from Telegram, Requirement: Wait runs in the background, Requirements, restart-commands Specification, Scenario: Custom poll interval, Scenario: Defaults without the section (+8 more)

### Community 38 - "telegram/client_test.go"
Cohesion: 0.15
Nodes (29): Client, mime/multipart.FileHeader, mime/multipart.Form, net/http.HandlerFunc, checkDocumentRequest(), checkFileBytes(), checkFormValue(), Config (+21 more)

### Community 39 - "Requirement: Configurable readiness wait"
Cohesion: 0.12
Nodes (15): ADDED Requirements, Purpose, Requirement: Configurable readiness wait, Requirement: Restart a Deployment from Telegram, Requirement: Wait runs in the background, Scenario: Custom poll interval, Scenario: Defaults without the section, Scenario: Kubernetes API error (+7 more)

### Community 40 - "Requirement: Body size cap"
Cohesion: 0.13
Nodes (14): ADDED Requirements, Purpose, Requirement: Body size cap, Requirement: Completion notification, Requirement: Hook configuration lookup, Scenario: Body over the default cap, Scenario: Body within the default cap, Scenario: Completed download (+6 more)

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
Nodes (15): message-catalog Specification, Purpose, Requirement: Configured overrides, Requirement: Embedded default catalog, Requirement: Placeholders are filled by code, Requirement: Startup validation of the catalog, Requirements, Scenario: Caption header override (+7 more)

### Community 46 - "Design"
Cohesion: 0.20
Nodes (9): Context, Decisions, Decisions (resolved), Design, Goals / Non-Goals, Implementation notes, Migration Plan, Renovate 44.151.0 findings (read from source) (+1 more)

### Community 47 - "Catalog"
Cohesion: 0.22
Nodes (9): CaptionPresenter, DocumentPresenter, renderer, time.Location, NewDependencyReportHandler(), Catalog, countsText(), NewCaptionPresenter() (+1 more)

### Community 48 - "Proposal"
Cohesion: 0.20
Nodes (9): Capabilities, Dependency, Impact, Modified Capabilities, New Capabilities, Non-goals, Proposal, What Changes (+1 more)

### Community 49 - "main_test.go"
Cohesion: 0.18
Nodes (13): fakeAdder, noopSender, jobPost(), routesFor(), routesWith(), TestNewGitHubClientNilWhenUnconfigured(), TestNewHTTPServerUsesConfiguredTimeouts(), TestNewKubeClientNilWhenUnconfigured() (+5 more)

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

### Community 54 - "Proposal"
Cohesion: 0.22
Nodes (8): Capabilities, Impact, Modified Capabilities, New Capabilities, Non-goals, Proposal, What Changes, Why

### Community 55 - "Tasks"
Cohesion: 0.22
Nodes (8): 1. Config package, 2. Processor package, 3. Email package, 4. Torrent webhook, 5. Quality gates, 6. User and other-repo steps, 7. Final validation, Tasks

### Community 57 - "Proposal"
Cohesion: 0.22
Nodes (8): Capabilities, Impact, Modified Capabilities, New Capabilities, Non-goals, Proposal, What Changes, Why

### Community 58 - "document.go"
Cohesion: 0.22
Nodes (21): cell, columnSpec, columnView, distItem, docBuilder, documentView, labelSet, Row (+13 more)

### Community 59 - "settings_test.go"
Cohesion: 0.23
Nodes (13): expectSettingError(), loadZone(), TestDependencyReportDefaults(), TestDependencyReportDocumentNameNeedsDate(), TestDependencyReportInvalidPath(), TestDependencyReportNonNumericBodyLimit(), TestDependencyReportTokenHeaderAndLayout(), TestDependencyReportZeroBodyLimit() (+5 more)

### Community 60 - "Scheduler"
Cohesion: 0.15
Nodes (14): cron.Cron, cron.EntryID, cron.Schedule, cron.ScheduleParser, Scheduler, New(), newWithParser(), ParseSpec() (+6 more)

### Community 61 - "Design"
Cohesion: 0.29
Nodes (6): Context, Decisions, Design, Goals / Non-Goals, Migration Plan, Risks / Trade-offs

### Community 62 - "Tasks"
Cohesion: 0.29
Nodes (6): 1. Shared HTTP package and configuration, 2. Telegram client, 3. GitHub client, 4. Kubernetes client, 5. Final validation, Tasks

### Community 63 - "Tasks"
Cohesion: 0.22
Nodes (8): 1. Prerequisite gate, 2. Configuration, 3. Message texts, 4. Command registry, 5. Scheduling and menu wiring, 6. Webhook registry, 7. Validation, Tasks

### Community 65 - "Design"
Cohesion: 0.25
Nodes (7): Context, Decisions, Design, Goals / Non-Goals, Migration Plan, Open Questions, Risks / Trade-offs

### Community 66 - "Requirement: Create a Job from a CronJob"
Cohesion: 0.29
Nodes (6): MODIFIED Requirements, Requirement: Create a Job from a CronJob, Scenario: Kubernetes API error, Scenario: Successful creation, Scenario: Unauthorized chat, Spec Delta

### Community 67 - "kube/client.go"
Cohesion: 0.16
Nodes (17): encoding/json.RawMessage, jobFromTemplate(), newClient(), NewInClusterClient(), resourceURL(), TestInClusterSettingsUseCA(), TestNewInClusterClientOutsideCluster(), withServiceAccountCA() (+9 more)

### Community 68 - "defaultCatalog"
Cohesion: 0.80
Nodes (4): defaultCatalog(), ownKeys(), TestDefaultsHaveEveryKey(), TestDefaultsKeepPlaceholders()

### Community 69 - "Proposal"
Cohesion: 0.20
Nodes (9): Capabilities, Dependency, Impact, Modified Capabilities, New Capabilities, Non-goals, Proposal, What Changes (+1 more)

### Community 70 - "aggregate.go"
Cohesion: 0.22
Nodes (14): depKey, RepoPending, ScanError, tabBuilder, addFileDeps(), compareRows(), firstNonNil(), newTabBuilder() (+6 more)

### Community 71 - "Requirement: Document tabs"
Cohesion: 0.22
Nodes (9): Requirement: Document tabs, Scenario: Digest only, Scenario: Dockerfile image and missing datasource, Scenario: Document fork tag override, Scenario: Fork tag, Scenario: Image dependency, Scenario: Non-image dependency, Scenario: Several updates for one dep (+1 more)

### Community 72 - "RouteEnv"
Cohesion: 0.28
Nodes (12): routeEntry, mountRoute(), registerRoutes(), github.com/gorilla/mux.Router, net/http.Handler, Route, RouteFactory, WebhookConfig (+4 more)

### Community 73 - "retry_test.go"
Cohesion: 0.26
Nodes (14): TestNewClientWithoutEnvProxyOrHTTP2(), TestSendCapsBodyAtLimit(), TestSendReturnsStatusAndBody(), apiError(), rateLimited(), retryTest(), testBackoff(), TestBackoffNoOverflowAtHighAttempts() (+6 more)

### Community 74 - "Requirement: Configured local timezone"
Cohesion: 0.25
Nodes (7): MODIFIED Requirements, RENAMED Requirements, Requirement: Configured local timezone, Scenario: Configured zone, Scenario: Image without tzdata, Scenario: Invalid zone, scheduled-dispatch Specification

### Community 75 - "Client"
Cohesion: 0.26
Nodes (10): github.com/go-telegram-bot-api/telegram-bot-api/v5.BotAPI, github.com/go-telegram-bot-api/telegram-bot-api/v5.HTTPClient, Client, NewClient(), newClient(), newClientWithBot(), parseInt64(), resolveChat() (+2 more)

### Community 76 - "Requirement: Scheduled job commands"
Cohesion: 0.33
Nodes (5): job-commands Specification, MODIFIED Requirements, Requirement: Scheduled job commands, Scenario: Weekly report, Scenario: Weekly report in a configured zone

### Community 77 - "Requirement: HTML report document"
Cohesion: 0.33
Nodes (6): Requirement: HTML report document, Scenario: Configured timezone changes the date and the file name, Scenario: Date near midnight UTC, Scenario: Renovate failed, Scenario: Report file missing, Scenario: Successful run

### Community 78 - "Requirement: Summary caption"
Cohesion: 0.33
Nodes (6): Requirement: Summary caption, Scenario: Caption over the limit, Scenario: Escaped repository name, Scenario: Fork and sort order, Scenario: Major, minor and patch counts, Scenario: Owner-agnostic name

### Community 79 - "command"
Cohesion: 0.20
Nodes (6): allowList, command, replier, BotHandlerDeps, newCommand(), newReplier()

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

### Community 85 - "Requirement: Renovate failure message"
Cohesion: 0.67
Nodes (3): Requirement: Renovate failure message, Scenario: Exit code 1, Scenario: Title override

### Community 86 - "Requirement: Safe HTML and size limit"
Cohesion: 0.67
Nodes (3): Requirement: Safe HTML and size limit, Scenario: Caption over the limit, Scenario: Long escaped content

### Community 87 - "context.Context"
Cohesion: 0.22
Nodes (6): context.Context, go.uber.org/zap.Field, blockingRestarter, jobCommand, BotHandlerDeps, newJobCommand()

### Community 88 - "BotHandlerDeps"
Cohesion: 0.24
Nodes (11): github.com/go-telegram-bot-api/telegram-bot-api/v5.Message, github.com/go-telegram-bot-api/telegram-bot-api/v5.Update, DeploymentRestarter, JobCreator, TelegramMessenger, BotHandler, BotHandlerDeps, KubeClient (+3 more)

### Community 89 - "NewDeliverer"
Cohesion: 0.23
Nodes (8): fakeSender, Deliverer, ReportSender, NewDeliverer(), TestDeliverFailureMessageIsTextOnly(), TestDeliverFallsBackToCaptionText(), TestDeliverLogsSenderError(), TestDeliverSendsDocumentWithCaption()

### Community 90 - "go.uber.org/zap.Logger"
Cohesion: 0.42
Nodes (11): lazyScheduler, main(), newGitHubClient(), newHTTPServer(), newKubeClient(), newTelegramClient(), setupWorkflowBot(), shutdownServer() (+3 more)

### Community 91 - "DependencyReportHandler"
Cohesion: 0.30
Nodes (4): net/http.ResponseWriter, DependencyReportHandler, Delivery, Presenter

### Community 92 - "Summary"
Cohesion: 0.24
Nodes (7): time.Time, countingAggregate, Summary, buildDocumentView(), DocumentName(), TestDocumentName(), fixedClock()

### Community 93 - "RetryingSender"
Cohesion: 0.33
Nodes (5): github.com/go-telegram-bot-api/telegram-bot-api/v5.Chattable, NewRetryingSender(), Backoff, RetryingSender, sender

### Community 94 - "dependency_report_golden_test.go"
Cohesion: 0.48
Nodes (6): goldenSends, goldenOutputs(), readGolden(), sendAll(), TestGoldenBaseline(), upToDateFixture()

### Community 95 - "workflowCommand"
Cohesion: 0.43
Nodes (3): workflowCommand, BotHandlerDeps, newWorkflowCommand()

### Community 98 - "restartCommand"
Cohesion: 0.47
Nodes (3): restartCommand, BotHandlerDeps, newRestartCommand()

## Knowledge Gaps
- **490 isolated node(s):** `automation-hub`, `BotHandlerDeps`, `Context`, `Renovate 44.151.0 findings (read from source)`, `Goals / Non-Goals` (+485 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 586 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **12 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Config` connect `config.go` to `NewWebhookHandler`, `RouteEnv`, `Catalog`, `main_test.go`, `go.uber.org/zap.Logger`?**
  _High betweenness centrality (0.028) - this node is a cross-community bridge._
- **Why does `Catalog` connect `Catalog` to `summaryAt`, `config.go`, `defaultCatalog`, `RouteEnv`, `buildFailure`, `command`, `BotHandlerDeps`, `DependencyReportHandler`, `Load`?**
  _High betweenness centrality (0.026) - this node is a cross-community bridge._
- **Why does `DependencyReportHandler` connect `DependencyReportHandler` to `summaryAt`, `config.go`, `dependency_report_test.go`, `Catalog`, `NewDeliverer`, `go.uber.org/zap.Logger`, `Summary`?**
  _High betweenness centrality (0.023) - this node is a cross-community bridge._
- **Are the 21 inferred relationships involving `loadYAML()` (e.g. with `loadFrom()` and `TestAPIURLValidation()`) actually correct?**
  _`loadYAML()` has 21 INFERRED edges - model-reasoned connections that need verification._
- **What connects `automation-hub`, `BotHandlerDeps`, `Context` to the rest of the system?**
  _490 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `summaryAt` be split into smaller, more focused modules?**
  _Cohesion score 0.10935143288084465 - nodes in this community are weakly interconnected._
- **Should `Requirements` be split into smaller, more focused modules?**
  _Cohesion score 0.08695652173913043 - nodes in this community are weakly interconnected._