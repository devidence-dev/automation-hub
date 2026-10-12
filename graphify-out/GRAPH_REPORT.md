# Graph Report - automation-hub  (2026-10-11)

## Corpus Check
- 52 files · ~40,316 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 634 nodes · 1176 edges · 35 communities (28 shown, 7 thin omitted)
- Extraction: 92% EXTRACTED · 8% INFERRED · 0% AMBIGUOUS · INFERRED: 96 edges (avg confidence: 0.86)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `e2c0b5ab`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Requirement: Document tabs
- testing.T
- BotHandler
- 🤖 Automation Hub
- config.go
- IMAPClient
- ADDED Requirements
- go.uber.org/zap.Logger
- dependency_report_test.go
- Requirements
- GenericEmailProcessor
- Client
- Scheduler
- ADDED Requirements
- CLAUDE.md
- automation-hub-network Bridge Network
- Dependabot Docker Update Config (deployments/docker)
- Dependabot GitHub Actions Update Config
- Dependabot Go Modules Update Config
- Requirements
- main
- Graphify Knowledge-Graph Workflow Rules
- dependency_report_document.go
- ADDED Requirements
- Proposal
- Tasks
- Design
- Requirement: Create a Job from a CronJob
- Design
- Proposal
- Proposal
- Tasks
- Design
- Tasks
- automation-hub

## God Nodes (most connected - your core abstractions)
1. `BotHandler` - 21 edges
2. `NewBotHandler()` - 21 edges
3. `Client` - 21 edges
4. `Config` - 18 edges
5. `IMAPClient` - 16 edges
6. `GenericEmailProcessor` - 15 edges
7. `NewGenericEmailProcessor()` - 15 edges
8. `NewClientWithBaseURL()` - 13 edges
9. `DependencyReportHandler` - 12 edges
10. `parseReport()` - 12 edges

## Surprising Connections (you probably didn't know these)
- `📝 Step 2: Configure Services` --references--> `ServiceConfig`  [INFERRED]
  README.md → internal/config/config.go
- `🤖 Step 3: Setup Telegram Bot` --references--> `TelegramConfig`  [INFERRED]
  README.md → internal/config/config.go
- `▶️ Trigger GitHub Actions from Telegram` --references--> `WorkflowBotConfig`  [INFERRED]
  README.md → internal/config/config.go
- `main()` --calls--> `Load()`  [EXTRACTED]
  cmd/automation-hub/main.go → internal/config/config.go
- `main()` --calls--> `NewBotHandler()`  [EXTRACTED]
  cmd/automation-hub/main.go → internal/handlers/telegram_bot.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Dependabot Multi-Ecosystem Update Strategy (gomod, docker, github-actions)** — _github_dependabot_gomod_updates, _github_dependabot_docker_updates, _github_dependabot_github_actions_updates [EXTRACTED 1.00]

## Communities (35 total, 7 thin omitted)

### Community 0 - "Requirement: Document tabs"
Cohesion: 0.07
Nodes (28): ADDED Requirements, MODIFIED Requirements, Requirement: Delivery, Requirement: Document tabs, Requirement: HTML report document, Requirement: Nothing pending document, Requirement: Safe document content, Requirement: Scan errors in the document (+20 more)

### Community 1 - "testing.T"
Cohesion: 0.10
Nodes (43): Client, net/http.HandlerFunc, testing.T, TestEmail(), TestTorrentNotification(), NewClientWithBaseURL(), assertRestartDeploymentRequest(), deploymentJSON() (+35 more)

### Community 2 - "BotHandler"
Cohesion: 0.11
Nodes (27): github.com/go-telegram-bot-api/telegram-bot-api/v5.Message, github.com/go-telegram-bot-api/telegram-bot-api/v5.Update, sync.Mutex, BotHandler, deploymentRestarter, fakeJobCreator, fakeMessenger, jobCreator (+19 more)

### Community 3 - "🤖 Automation Hub"
Cohesion: 0.06
Nodes (32): CI Job: Devidence Go CI, CI Pipeline Workflow, Devidence Go CI Reusable Workflow, Deploy Job: Build, Push & Deploy, Deploy Job: Resolve Build Args, Devidence CD Build Deploy Reusable Workflow, Build and Deploy Workflow, README (+24 more)

### Community 4 - "config.go"
Cohesion: 0.11
Nodes (32): GitHubConfig, ServerConfig, ServiceConfig, TelegramConfig, WebhookConfig, WorkflowBotConfig, Config, DependencyReportConfig (+24 more)

### Community 5 - "IMAPClient"
Cohesion: 0.08
Nodes (19): IMAPClient, mockNamedProcessor, github.com/emersion/go-imap/client.Client, sync.WaitGroup, imap.Literal, imap.Message, Email, EmailProcessor (+11 more)

### Community 6 - "ADDED Requirements"
Cohesion: 0.09
Nodes (22): ADDED Requirements, Purpose, Requirement: Exact run link, Requirement: Failure notification for scheduled dispatch, Requirement: Fixed local timezone, Requirement: Graceful stop, Requirement: Optional schedule per workflow command, Requirement: Scheduled dispatch uses the command's dispatch path (+14 more)

### Community 7 - "go.uber.org/zap.Logger"
Cohesion: 0.11
Nodes (22): Client, context.Context, encoding/json.RawMessage, go.uber.org/zap.Logger, net/http.Client, time.Duration, fakeDispatcher, fakeRestarter (+14 more)

### Community 8 - "dependency_report_test.go"
Cohesion: 0.07
Nodes (56): failingRoundTripper, net/http.Request, net/http.Response, net/http.ResponseWriter, time.Time, DependencyReportHandler, fakeHTMLSender, packageFile (+48 more)

### Community 9 - "Requirements"
Cohesion: 0.09
Nodes (22): Purpose, Requirement: Exact run link, Requirement: Failure notification for scheduled dispatch, Requirement: Fixed local timezone, Requirement: Graceful stop, Requirement: Optional schedule per workflow command, Requirement: Scheduled dispatch uses the command's dispatch path, Requirement: Startup schedule log (+14 more)

### Community 10 - "GenericEmailProcessor"
Cohesion: 0.23
Nodes (4): regexp.Regexp, TestTruncateString(), truncateString(), GenericEmailProcessor

### Community 11 - "Client"
Cohesion: 0.14
Nodes (15): github.com/go-telegram-bot-api/telegram-bot-api/v5.BotAPI, github.com/go-telegram-bot-api/telegram-bot-api/v5.BotCommand, github.com/go-telegram-bot-api/telegram-bot-api/v5.Chattable, github.com/go-telegram-bot-api/telegram-bot-api/v5.HTTPClient, NewTorrentProcessor(), NewTorrentProcessorLegacy(), TestGetWebhookConfig(), TestNewTorrentProcessor() (+7 more)

### Community 12 - "Scheduler"
Cohesion: 0.16
Nodes (15): cron.Cron, cron.EntryID, cron.Schedule, cron.ScheduleParser, time.Location, New(), newWithParser(), ParseSpec() (+7 more)

### Community 13 - "ADDED Requirements"
Cohesion: 0.09
Nodes (22): ADDED Requirements, Purpose, Requirement: All up to date, Requirement: Authenticated endpoint, Requirement: Delivery, Requirement: Per-repository message, Requirement: Renovate failure message, Requirement: Request handling (+14 more)

### Community 19 - "Requirements"
Cohesion: 0.09
Nodes (22): dependency-report Specification, Purpose, Requirement: All up to date, Requirement: Authenticated endpoint, Requirement: Delivery, Requirement: Per-repository message, Requirement: Renovate failure message, Requirement: Request handling (+14 more)

### Community 20 - "main"
Cohesion: 0.16
Nodes (12): noopSender, reportSender, main(), registerRoutes(), TestRegisterRoutesDependencyReport(), github.com/gorilla/mux.Router, WebhookHandler, NewWebhookHandler() (+4 more)

### Community 22 - "dependency_report_document.go"
Cohesion: 0.24
Nodes (14): depKey, distItem, docRow, documentData, reportTab, reportView, tabBuilder, buildDocumentData() (+6 more)

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

### Community 28 - "Design"
Cohesion: 0.22
Nodes (8): Context, Decisions, Design, Goals / Non-Goals, Migration Plan, Open Questions, Renovate 44.151.0 findings (read from source), Risks / Trade-offs

### Community 29 - "Proposal"
Cohesion: 0.22
Nodes (8): Capabilities, Impact, Modified Capabilities, New Capabilities, Non-goals, Proposal, What Changes, Why

### Community 30 - "Proposal"
Cohesion: 0.22
Nodes (8): Capabilities, Impact, Modified Capabilities, New Capabilities, Non-goals, Proposal, What Changes, Why

### Community 31 - "Tasks"
Cohesion: 0.25
Nodes (7): 1. Pre-checks, 2. Telegram client, 3. Report document, 4. Handler wiring and docs, 5. Cross-repo: scan homelab and retire icr, 6. Validation, Tasks

### Community 32 - "Design"
Cohesion: 0.25
Nodes (7): Context, Decisions, Design, Goals / Non-Goals, Migration Plan, Open Questions, Risks / Trade-offs

### Community 33 - "Tasks"
Cohesion: 0.25
Nodes (7): 1. Pre-checks, 2. Config and wiring, 3. Job commands, 4. Dependency report, 5. Cross-repo, 6. Validation, Tasks

## Knowledge Gaps
- **165 isolated node(s):** `automation-hub`, `Context`, `Renovate 44.151.0 findings (read from source)`, `Goals / Non-Goals`, `Decisions` (+160 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 213 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **7 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `🤖 Automation Hub` connect `🤖 Automation Hub` to `config.go`?**
  _High betweenness centrality (0.044) - this node is a cross-community bridge._
- **Why does `IMAPClient` connect `IMAPClient` to `config.go`, `go.uber.org/zap.Logger`?**
  _High betweenness centrality (0.030) - this node is a cross-community bridge._
- **Why does `NewWebhookHandler()` connect `main` to `Client`, `config.go`, `go.uber.org/zap.Logger`?**
  _High betweenness centrality (0.029) - this node is a cross-community bridge._
- **Are the 10 inferred relationships involving `NewBotHandler()` (e.g. with `TestBotHandlerDispatchWorkflowFallsBackWhenRunURLIsEmpty()` and `TestBotHandlerDispatchWorkflowLinksToTheDispatchedRun()`) actually correct?**
  _`NewBotHandler()` has 10 INFERRED edges - model-reasoned connections that need verification._
- **What connects `automation-hub`, `Context`, `Renovate 44.151.0 findings (read from source)` to the rest of the system?**
  _165 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Requirement: Document tabs` be split into smaller, more focused modules?**
  _Cohesion score 0.06896551724137931 - nodes in this community are weakly interconnected._
- **Should `testing.T` be split into smaller, more focused modules?**
  _Cohesion score 0.1026827012025902 - nodes in this community are weakly interconnected._