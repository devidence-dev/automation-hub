# Graph Report - automation-hub  (2026-10-11)

## Corpus Check
- 43 files · ~28,771 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 489 nodes · 931 edges · 28 communities (21 shown, 7 thin omitted)
- Extraction: 91% EXTRACTED · 9% INFERRED · 0% AMBIGUOUS · INFERRED: 81 edges (avg confidence: 0.86)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `8ff5cee9`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ADDED Requirements
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
- Proposal
- main
- Graphify Knowledge-Graph Workflow Rules
- Design
- Tasks
- Proposal
- Tasks
- Design
- automation-hub

## God Nodes (most connected - your core abstractions)
1. `BotHandler` - 21 edges
2. `NewBotHandler()` - 21 edges
3. `Client` - 19 edges
4. `Config` - 18 edges
5. `IMAPClient` - 16 edges
6. `GenericEmailProcessor` - 15 edges
7. `NewGenericEmailProcessor()` - 15 edges
8. `NewClientWithBaseURL()` - 13 edges
9. `Client` - 12 edges
10. `🤖 Automation Hub` - 12 edges

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

## Communities (28 total, 7 thin omitted)

### Community 0 - "ADDED Requirements"
Cohesion: 0.09
Nodes (22): ADDED Requirements, Purpose, Requirement: All up to date, Requirement: Authenticated endpoint, Requirement: Delivery, Requirement: Per-repository message, Requirement: Renovate failure message, Requirement: Request handling (+14 more)

### Community 1 - "testing.T"
Cohesion: 0.10
Nodes (37): Client, net/http.HandlerFunc, testing.T, TestEmail(), TestTorrentNotification(), NewIMAPClient(), TestExtractTextPlain(), TestHandlePostProcessing() (+29 more)

### Community 2 - "BotHandler"
Cohesion: 0.11
Nodes (28): github.com/go-telegram-bot-api/telegram-bot-api/v5.Message, github.com/go-telegram-bot-api/telegram-bot-api/v5.Update, sync.Mutex, BotHandler, deploymentRestarter, fakeJobCreator, fakeMessenger, jobCreator (+20 more)

### Community 3 - "🤖 Automation Hub"
Cohesion: 0.06
Nodes (32): CI Job: Devidence Go CI, CI Pipeline Workflow, Devidence Go CI Reusable Workflow, Deploy Job: Build, Push & Deploy, Deploy Job: Resolve Build Args, Devidence CD Build Deploy Reusable Workflow, Build and Deploy Workflow, README (+24 more)

### Community 4 - "config.go"
Cohesion: 0.11
Nodes (30): GitHubConfig, ServerConfig, ServiceConfig, TelegramConfig, WebhookConfig, WorkflowBotConfig, Config, EmailConfig (+22 more)

### Community 5 - "IMAPClient"
Cohesion: 0.11
Nodes (10): IMAPClient, mockNamedProcessor, github.com/emersion/go-imap/client.Client, sync.WaitGroup, imap.Literal, imap.Message, Email, EmailProcessor (+2 more)

### Community 6 - "ADDED Requirements"
Cohesion: 0.09
Nodes (22): ADDED Requirements, Purpose, Requirement: Exact run link, Requirement: Failure notification for scheduled dispatch, Requirement: Fixed local timezone, Requirement: Graceful stop, Requirement: Optional schedule per workflow command, Requirement: Scheduled dispatch uses the command's dispatch path (+14 more)

### Community 7 - "go.uber.org/zap.Logger"
Cohesion: 0.08
Nodes (34): Client, context.Context, encoding/json.RawMessage, go.uber.org/zap.Logger, net/http.Client, time.Duration, time.Time, fakeDispatcher (+26 more)

### Community 8 - "dependency_report_test.go"
Cohesion: 0.09
Nodes (36): failingRoundTripper, net/http.Request, net/http.Response, net/http.ResponseWriter, DependencyReportHandler, fakeHTMLSender, htmlSender, renovateReport (+28 more)

### Community 9 - "Requirements"
Cohesion: 0.09
Nodes (22): Purpose, Requirement: Exact run link, Requirement: Failure notification for scheduled dispatch, Requirement: Fixed local timezone, Requirement: Graceful stop, Requirement: Optional schedule per workflow command, Requirement: Scheduled dispatch uses the command's dispatch path, Requirement: Startup schedule log (+14 more)

### Community 10 - "GenericEmailProcessor"
Cohesion: 0.23
Nodes (4): regexp.Regexp, TestTruncateString(), truncateString(), GenericEmailProcessor

### Community 11 - "Client"
Cohesion: 0.15
Nodes (14): github.com/go-telegram-bot-api/telegram-bot-api/v5.BotAPI, github.com/go-telegram-bot-api/telegram-bot-api/v5.BotCommand, github.com/go-telegram-bot-api/telegram-bot-api/v5.HTTPClient, NewTorrentProcessor(), NewTorrentProcessorLegacy(), TestGetWebhookConfig(), TestNewTorrentProcessor(), TestNewTorrentProcessorLegacy() (+6 more)

### Community 12 - "Scheduler"
Cohesion: 0.16
Nodes (15): cron.Cron, cron.EntryID, cron.Schedule, cron.ScheduleParser, time.Location, New(), newWithParser(), ParseSpec() (+7 more)

### Community 13 - "ADDED Requirements"
Cohesion: 0.15
Nodes (12): ADDED Requirements, Purpose, Requirement: Create a Job from a CronJob, Requirement: Job command configuration, Requirement: Scheduled job commands, Scenario: Duplicate across types, Scenario: Jobs-only config starts the bot, Scenario: Kubernetes API error (+4 more)

### Community 19 - "Proposal"
Cohesion: 0.22
Nodes (8): Capabilities, Impact, Modified Capabilities, New Capabilities, Non-goals, Proposal, What Changes, Why

### Community 20 - "main"
Cohesion: 0.24
Nodes (10): reportSender, main(), registerRoutes(), github.com/gorilla/mux.Router, WebhookHandler, NewWebhookHandler(), TestHandleTorrentComplete_InvalidJSON(), TestHandleTorrentComplete_MissingWebhookConfig() (+2 more)

### Community 22 - "Design"
Cohesion: 0.25
Nodes (7): Context, Decisions, Design, Goals / Non-Goals, Migration Plan, Open Questions, Risks / Trade-offs

### Community 23 - "Tasks"
Cohesion: 0.25
Nodes (7): 1. Pre-checks, 2. Config and wiring, 3. Job commands, 4. Dependency report, 5. Cross-repo, 6. Validation, Tasks

### Community 24 - "Proposal"
Cohesion: 0.22
Nodes (8): Capabilities, Impact, Modified Capabilities, New Capabilities, Non-goals, Proposal, What Changes, Why

### Community 25 - "Tasks"
Cohesion: 0.25
Nodes (7): 1. Dependency and timezone, 2. Config, 3. Dispatch core, 4. Scheduler, 5. Cross-repo, 6. Validation, Tasks

### Community 26 - "Design"
Cohesion: 0.29
Nodes (6): Context, Decisions, Design, Goals / Non-Goals, Migration Plan, Risks / Trade-offs

## Knowledge Gaps
- **107 isolated node(s):** `automation-hub`, `Context`, `Goals / Non-Goals`, `Decisions`, `Risks / Trade-offs` (+102 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 146 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **7 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `🤖 Automation Hub` connect `🤖 Automation Hub` to `config.go`?**
  _High betweenness centrality (0.063) - this node is a cross-community bridge._
- **Why does `IMAPClient` connect `IMAPClient` to `testing.T`, `config.go`, `go.uber.org/zap.Logger`?**
  _High betweenness centrality (0.043) - this node is a cross-community bridge._
- **Why does `NewBotHandler()` connect `BotHandler` to `main`, `go.uber.org/zap.Logger`?**
  _High betweenness centrality (0.041) - this node is a cross-community bridge._
- **Are the 10 inferred relationships involving `NewBotHandler()` (e.g. with `TestBotHandlerDispatchWorkflowFallsBackWhenRunURLIsEmpty()` and `TestBotHandlerDispatchWorkflowLinksToTheDispatchedRun()`) actually correct?**
  _`NewBotHandler()` has 10 INFERRED edges - model-reasoned connections that need verification._
- **What connects `automation-hub`, `Context`, `Goals / Non-Goals` to the rest of the system?**
  _107 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ADDED Requirements` be split into smaller, more focused modules?**
  _Cohesion score 0.08695652173913043 - nodes in this community are weakly interconnected._
- **Should `testing.T` be split into smaller, more focused modules?**
  _Cohesion score 0.10188261351052048 - nodes in this community are weakly interconnected._