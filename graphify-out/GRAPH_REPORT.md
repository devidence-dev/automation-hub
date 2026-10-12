# Graph Report - automation-hub  (2026-10-11)

## Corpus Check
- 97 files · ~75,127 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 1243 nodes · 2239 edges · 59 communities (51 shown, 7 thin omitted)
- Extraction: 92% EXTRACTED · 8% INFERRED · 0% AMBIGUOUS · INFERRED: 172 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `5ca5b96c`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- document.go
- testing.T
- 🤖 Automation Hub
- Config
- IMAPClient
- ADDED Requirements
- telegram_webhook_test.go
- NewGenericEmailProcessor
- Requirements
- ADDED Requirements
- time.Location
- dependency_report_test.go
- ADDED Requirements
- CLAUDE.md
- automation-hub-network Bridge Network
- Dependabot Docker Update Config (deployments/docker)
- Dependabot GitHub Actions Update Config
- Dependabot Go Modules Update Config
- Requirements
- Requirement: Document tabs
- Graphify Knowledge-Graph Workflow Rules
- DependencyReportHandler
- ADDED Requirements
- Proposal
- Tasks
- Design
- Requirement: Create a Job from a CronJob
- Requirement: Document tabs
- ADDED Requirements
- Proposal
- Catalog
- Design
- Tasks
- ADDED Requirements
- go.uber.org/zap.Logger
- Decisions
- Requirement: Configurable readiness wait
- Decisions
- ADDED Requirements
- Requirement: Body size cap
- Tasks
- Design
- Proposal
- Requirement: Exact run link
- Proposal
- Proposal
- Tasks
- Tasks
- Proposal
- Tasks
- automation-hub
- Proposal
- Design
- Requirement: Configured local timezone
- Requirement: Create a Job from a CronJob
- Design
- Tasks
- Requirement: Scheduled job commands

## God Nodes (most connected - your core abstractions)
1. `Config` - 29 edges
2. `Client` - 24 edges
3. `BotHandler` - 23 edges
4. `summaryAt()` - 22 edges
5. `DependencyReportHandler` - 20 edges
6. `loadFixtureReport()` - 19 edges
7. `Catalog` - 18 edges
8. `Aggregate()` - 18 edges
9. `newReportHandler()` - 17 edges
10. `Row` - 16 edges

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

## Communities (59 total, 7 thin omitted)

### Community 0 - "document.go"
Cohesion: 0.07
Nodes (52): CaptionPresenter, cell, columnSpec, columnView, Dep, depKey, distItem, docBuilder (+44 more)

### Community 1 - "testing.T"
Cohesion: 0.06
Nodes (96): Client, mime/multipart.FileHeader, mime/multipart.Form, net/http.HandlerFunc, testing.T, loadYAML(), TestLoadExampleConfig(), expectSettingError() (+88 more)

### Community 3 - "🤖 Automation Hub"
Cohesion: 0.07
Nodes (31): CI Job: Devidence Go CI, CI Pipeline Workflow, Devidence Go CI Reusable Workflow, Deploy Job: Build, Push & Deploy, Deploy Job: Resolve Build Args, Devidence CD Build Deploy Reusable Workflow, Build and Deploy Workflow, README (+23 more)

### Community 4 - "Config"
Cohesion: 0.08
Nodes (46): GitHubConfig, ServerConfig, ServiceConfig, TelegramConfig, WebhookConfig, WorkflowBotConfig, github.com/go-telegram-bot-api/telegram-bot-api/v5.Message, BotHandlerDeps (+38 more)

### Community 5 - "IMAPClient"
Cohesion: 0.09
Nodes (16): IMAPClient, mockNamedProcessor, github.com/emersion/go-imap/client.Client, sync.WaitGroup, imap.Literal, imap.Message, Email, EmailProcessor (+8 more)

### Community 6 - "ADDED Requirements"
Cohesion: 0.09
Nodes (22): ADDED Requirements, Purpose, Requirement: Exact run link, Requirement: Failure notification for scheduled dispatch, Requirement: Fixed local timezone, Requirement: Graceful stop, Requirement: Optional schedule per workflow command, Requirement: Scheduled dispatch uses the command's dispatch path (+14 more)

### Community 7 - "telegram_webhook_test.go"
Cohesion: 0.07
Nodes (46): context.Context, encoding/json.RawMessage, github.com/go-telegram-bot-api/telegram-bot-api/v5.Update, net/http/httptest.Server, sync.Mutex, time.Duration, fakeDispatcher, fakeJobCreator (+38 more)

### Community 8 - "NewGenericEmailProcessor"
Cohesion: 0.14
Nodes (14): regexp.Regexp, NewGenericEmailProcessor(), TestDecodeQuotedPrintable(), TestExtractCode(), TestExtractCode_IgnoresHexColor(), TestExtractPerplexityCode(), TestNewGenericEmailProcessor_BuiltInPatterns(), TestNewGenericEmailProcessor_CustomPattern() (+6 more)

### Community 9 - "Requirements"
Cohesion: 0.09
Nodes (22): Purpose, Requirement: Exact run link, Requirement: Failure notification for scheduled dispatch, Requirement: Fixed local timezone, Requirement: Graceful stop, Requirement: Optional schedule per workflow command, Requirement: Scheduled dispatch uses the command's dispatch path, Requirement: Startup schedule log (+14 more)

### Community 10 - "ADDED Requirements"
Cohesion: 0.04
Nodes (45): ADDED Requirements, Purpose, Requirement: Body preparation, Requirement: Extractor configuration validation, Requirement: Malformed message handling, Requirement: Mark-as-read policy, Requirement: Marker extraction, Requirement: Not-found text (+37 more)

### Community 11 - "time.Location"
Cohesion: 0.15
Nodes (17): cron.Cron, cron.EntryID, cron.Schedule, cron.ScheduleParser, DocumentPresenter, time.Location, NewDocumentPresenter(), Scheduler (+9 more)

### Community 12 - "dependency_report_test.go"
Cohesion: 0.12
Nodes (35): countingAggregate, fakeSender, gatedSender, goldenSends, reportHandlerCase, reportRequest, sentCall, goldenOutputs() (+27 more)

### Community 13 - "ADDED Requirements"
Cohesion: 0.09
Nodes (22): ADDED Requirements, Purpose, Requirement: All up to date, Requirement: Authenticated endpoint, Requirement: Delivery, Requirement: Per-repository message, Requirement: Renovate failure message, Requirement: Request handling (+14 more)

### Community 19 - "Requirements"
Cohesion: 0.04
Nodes (45): dependency-report Specification, Purpose, Requirement: Authenticated endpoint, Requirement: Delivery, Requirement: Document tabs, Requirement: HTML report document, Requirement: Nothing pending document, Requirement: Renovate failure message (+37 more)

### Community 20 - "Requirement: Document tabs"
Cohesion: 0.04
Nodes (45): ADDED Requirements, dependency-report Specification, MODIFIED Requirements, Requirement: Authenticated endpoint, Requirement: Document tabs, Requirement: HTML report document, Requirement: Renovate failure message, Requirement: Request handling (+37 more)

### Community 22 - "DependencyReportHandler"
Cohesion: 0.08
Nodes (30): fakeSender, failingRoundTripper, net/http.Request, net/http.Response, net/http.ResponseWriter, DependencyReportHandler, Deliverer, Delivery (+22 more)

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
Cohesion: 0.15
Nodes (12): job-commands Specification, Purpose, Requirement: Create a Job from a CronJob, Requirement: Job command configuration, Requirement: Scheduled job commands, Requirements, Scenario: Duplicate across types, Scenario: Jobs-only config starts the bot (+4 more)

### Community 28 - "Requirement: Document tabs"
Cohesion: 0.05
Nodes (40): ADDED Requirements, MODIFIED Requirements, REMOVED Requirements, Requirement: All up to date, Requirement: Delivery, Requirement: Document tabs, Requirement: HTML report document, Requirement: Nothing pending document (+32 more)

### Community 29 - "ADDED Requirements"
Cohesion: 0.06
Nodes (35): ADDED Requirements, Purpose, Requirement: Bounded response bodies, Requirement: Configurable GitHub API base URL, Requirement: Configurable server timeouts, Requirement: Shared outbound HTTP settings, Requirement: Telegram long-poll timeouts, Requirement: Telegram retry policy (+27 more)

### Community 30 - "Proposal"
Cohesion: 0.22
Nodes (8): Capabilities, Impact, Modified Capabilities, New Capabilities, Non-goals, Proposal, What Changes, Why

### Community 31 - "Catalog"
Cohesion: 0.12
Nodes (24): applyGroup(), applyOverrides(), checkPlaceholders(), flatten(), Catalog, hasGroup(), join(), Load() (+16 more)

### Community 32 - "Design"
Cohesion: 0.25
Nodes (7): Context, Decisions, Design, Goals / Non-Goals, Migration Plan, Open Questions, Risks / Trade-offs

### Community 33 - "Tasks"
Cohesion: 0.25
Nodes (7): 1. Pre-checks, 2. Config and wiring, 3. Job commands, 4. Dependency report, 5. Cross-repo, 6. Validation, Tasks

### Community 34 - "ADDED Requirements"
Cohesion: 0.09
Nodes (21): ADDED Requirements, Purpose, Requirement: Bot texts from the message catalog, Requirement: Command routing by name, Requirement: Failure isolation in the bot, Requirement: No client for unconfigured types, Requirement: Per-entry authorization, Requirement: Single publication of Telegram commands (+13 more)

### Community 35 - "go.uber.org/zap.Logger"
Cohesion: 0.06
Nodes (56): noopSender, main(), newGitHubClient(), newHTTPServer(), newKubeClient(), registerBotCommands(), registerRoutes(), registerSchedules() (+48 more)

### Community 36 - "Decisions"
Cohesion: 0.11
Nodes (17): Context, Coordination with sibling changes, D1. Extraction is a strategy chosen by `extractor`, with two implementations, D2. Built-in name defaults are removed; configs carry the values, D3. Defaults live in `config`, applied through accessors, D4. Regexes are validated at load and compiled once at construction, D5. `messageSender` and `TemplateNotifier`, D6. `GenericEmailProcessor` is a composition (+9 more)

### Community 37 - "Requirement: Configurable readiness wait"
Cohesion: 0.12
Nodes (15): ADDED Requirements, Purpose, Requirement: Configurable readiness wait, Requirement: Restart a Deployment from Telegram, Requirement: Wait runs in the background, Scenario: Custom poll interval, Scenario: Defaults without the section, Scenario: Kubernetes API error (+7 more)

### Community 38 - "Decisions"
Cohesion: 0.12
Nodes (15): Context, D1. Two new packages, the handler stays thin, D2. Presentation strategy, D3. Typed update types and per-tab view models, D4. Timezone and dependency injection, D5. Config defaults applied in Go, only when the section exists, D6. Message catalog, D7. Visible default change (+7 more)

### Community 39 - "ADDED Requirements"
Cohesion: 0.12
Nodes (15): ADDED Requirements, message-catalog Specification, Purpose, Requirement: Configured overrides, Requirement: Embedded default catalog, Requirement: Placeholders are filled by code, Requirement: Startup validation of the catalog, Scenario: Caption header override (+7 more)

### Community 40 - "Requirement: Body size cap"
Cohesion: 0.13
Nodes (14): ADDED Requirements, Purpose, Requirement: Body size cap, Requirement: Completion notification, Requirement: Hook configuration lookup, Scenario: Body over the default cap, Scenario: Body within the default cap, Scenario: Completed download (+6 more)

### Community 42 - "Tasks"
Cohesion: 0.15
Nodes (12): 10. Cross-repo and user steps, 11. Validation, 1. Baseline, 2. Message catalog loader (`internal/messages`), 3. Catalog content, 4. Configuration (`internal/config`), 5. Parsers (`internal/services/depreport`), 6. Domain (`internal/services/depreport`) (+4 more)

### Community 46 - "Design"
Cohesion: 0.20
Nodes (9): Context, Decisions, Decisions (resolved), Design, Goals / Non-Goals, Implementation notes, Migration Plan, Renovate 44.151.0 findings (read from source) (+1 more)

### Community 47 - "Proposal"
Cohesion: 0.20
Nodes (9): Capabilities, Dependency, Impact, Modified Capabilities, New Capabilities, Non-goals, Proposal, What Changes (+1 more)

### Community 48 - "Requirement: Exact run link"
Cohesion: 0.20
Nodes (9): MODIFIED Requirements, Requirement: Exact run link, Scenario: Custom GitHub web URL, Scenario: Invalid GitHub web URL, Scenario: No run details, Scenario: Not a scheduled day, Scenario: Overlapping dispatches, Scenario: Run details returned (+1 more)

### Community 49 - "Proposal"
Cohesion: 0.20
Nodes (9): Capabilities, Dependency, Impact, Modified Capabilities, New Capabilities, Non-goals, Proposal, What Changes (+1 more)

### Community 51 - "Proposal"
Cohesion: 0.22
Nodes (8): Capabilities, Impact, Modified Capabilities, New Capabilities, Non-goals, Proposal, What Changes, Why

### Community 52 - "Tasks"
Cohesion: 0.22
Nodes (8): 1. Pre-checks, 2. Telegram client, 3. Report document, 4. Handler wiring and docs, 5. Cross-repo: scan homelab and retire icr, 6. Caption and single message, 7. Validation, Tasks

### Community 53 - "Tasks"
Cohesion: 0.22
Nodes (8): 1. Prerequisite gate, 2. Configuration, 3. Message texts, 4. Command registry, 5. Scheduling and menu wiring, 6. Webhook registry, 7. Validation, Tasks

### Community 54 - "Proposal"
Cohesion: 0.22
Nodes (8): Capabilities, Impact, Modified Capabilities, New Capabilities, Non-goals, Proposal, What Changes, Why

### Community 55 - "Tasks"
Cohesion: 0.22
Nodes (8): 1. Config package, 2. Processor package, 3. Email package, 4. Torrent webhook, 5. Quality gates, 6. User and other-repo steps, 7. Final validation, Tasks

### Community 57 - "Proposal"
Cohesion: 0.22
Nodes (8): Capabilities, Impact, Modified Capabilities, New Capabilities, Non-goals, Proposal, What Changes, Why

### Community 58 - "Design"
Cohesion: 0.25
Nodes (7): Context, Decisions, Design, Goals / Non-Goals, Migration Plan, Open Questions, Risks / Trade-offs

### Community 59 - "Requirement: Configured local timezone"
Cohesion: 0.25
Nodes (7): MODIFIED Requirements, RENAMED Requirements, Requirement: Configured local timezone, Scenario: Configured zone, Scenario: Image without tzdata, Scenario: Invalid zone, scheduled-dispatch Specification

### Community 60 - "Requirement: Create a Job from a CronJob"
Cohesion: 0.29
Nodes (6): MODIFIED Requirements, Requirement: Create a Job from a CronJob, Scenario: Kubernetes API error, Scenario: Successful creation, Scenario: Unauthorized chat, Spec Delta

### Community 61 - "Design"
Cohesion: 0.29
Nodes (6): Context, Decisions, Design, Goals / Non-Goals, Migration Plan, Risks / Trade-offs

### Community 62 - "Tasks"
Cohesion: 0.29
Nodes (6): 1. Shared HTTP package and configuration, 2. Telegram client, 3. GitHub client, 4. Kubernetes client, 5. Final validation, Tasks

### Community 63 - "Requirement: Scheduled job commands"
Cohesion: 0.33
Nodes (5): job-commands Specification, MODIFIED Requirements, Requirement: Scheduled job commands, Scenario: Weekly report, Scenario: Weekly report in a configured zone

## Knowledge Gaps
- **438 isolated node(s):** `automation-hub`, `Context`, `Renovate 44.151.0 findings (read from source)`, `Goals / Non-Goals`, `Decisions` (+433 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 518 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **7 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `🤖 Automation Hub` connect `🤖 Automation Hub` to `Config`?**
  _High betweenness centrality (0.025) - this node is a cross-community bridge._
- **Why does `Config` connect `Config` to `time.Location`, `testing.T`, `go.uber.org/zap.Logger`, `Catalog`?**
  _High betweenness centrality (0.022) - this node is a cross-community bridge._
- **Why does `⚙️ Configuration` connect `Config` to `🤖 Automation Hub`?**
  _High betweenness centrality (0.016) - this node is a cross-community bridge._
- **What connects `automation-hub`, `Context`, `Renovate 44.151.0 findings (read from source)` to the rest of the system?**
  _438 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `document.go` be split into smaller, more focused modules?**
  _Cohesion score 0.0656140350877193 - nodes in this community are weakly interconnected._
- **Should `testing.T` be split into smaller, more focused modules?**
  _Cohesion score 0.055860805860805864 - nodes in this community are weakly interconnected._
- **Should `🤖 Automation Hub` be split into smaller, more focused modules?**
  _Cohesion score 0.06666666666666667 - nodes in this community are weakly interconnected._