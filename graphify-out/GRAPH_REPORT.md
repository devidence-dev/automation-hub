# Graph Report - automation-hub  (2026-10-11)

## Corpus Check
- 39 files · ~19,739 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 402 nodes · 724 edges · 27 communities (20 shown, 7 thin omitted)
- Extraction: 91% EXTRACTED · 9% INFERRED · 0% AMBIGUOUS · INFERRED: 65 edges (avg confidence: 0.86)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `f6c17ae1`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ADDED Requirements
- testing.T
- NewBotHandler
- 🤖 Automation Hub
- config.go
- Email
- IMAPClient
- context.Context
- net/http.Request
- ADDED Requirements
- GenericEmailProcessor
- go.uber.org/zap.Logger
- Scheduler
- ADDED Requirements
- CLAUDE.md
- automation-hub-network Bridge Network
- Dependabot Docker Update Config (deployments/docker)
- Dependabot GitHub Actions Update Config
- Dependabot Go Modules Update Config
- Proposal
- Proposal
- Graphify Knowledge-Graph Workflow Rules
- Design
- Tasks
- Tasks
- Design
- automation-hub

## God Nodes (most connected - your core abstractions)
1. `NewBotHandler()` - 17 edges
2. `Client` - 17 edges
3. `BotHandler` - 16 edges
4. `IMAPClient` - 16 edges
5. `GenericEmailProcessor` - 15 edges
6. `NewGenericEmailProcessor()` - 15 edges
7. `Config` - 14 edges
8. `🤖 Automation Hub` - 12 edges
9. `NewClientWithBaseURL()` - 11 edges
10. `main()` - 10 edges

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

## Communities (27 total, 7 thin omitted)

### Community 0 - "ADDED Requirements"
Cohesion: 0.09
Nodes (22): ADDED Requirements, Purpose, Requirement: All up to date, Requirement: Authenticated endpoint, Requirement: Delivery, Requirement: Per-repository message, Requirement: Renovate failure message, Requirement: Request handling (+14 more)

### Community 1 - "testing.T"
Cohesion: 0.11
Nodes (35): Client, net/http.HandlerFunc, testing.T, TestEmail(), TestTorrentNotification(), assertDispatchWorkflowRequest(), TestDispatchWorkflow(), TestDispatchWorkflowNetworkError() (+27 more)

### Community 2 - "NewBotHandler"
Cohesion: 0.13
Nodes (24): github.com/go-telegram-bot-api/telegram-bot-api/v5.Message, github.com/go-telegram-bot-api/telegram-bot-api/v5.Update, sync.Mutex, BotHandler, deploymentRestarter, fakeMessenger, telegramMessenger, workflowDispatcher (+16 more)

### Community 3 - "🤖 Automation Hub"
Cohesion: 0.07
Nodes (28): Deploy Job: Build, Push & Deploy, Deploy Job: Resolve Build Args, Devidence CD Build Deploy Reusable Workflow, Build and Deploy Workflow, README, Adding New Email Services, Adding New Webhooks, � API & Webhooks (+20 more)

### Community 4 - "config.go"
Cohesion: 0.12
Nodes (25): GitHubConfig, ServerConfig, ServiceConfig, TelegramConfig, WebhookConfig, WorkflowBotConfig, Config, EmailConfig (+17 more)

### Community 5 - "Email"
Cohesion: 0.13
Nodes (9): mockNamedProcessor, sync.WaitGroup, Email, EmailProcessor, TorrentNotification, NewProcessorManager(), TestProcessorManager(), TestProcessorManager_CanceledContextAsync() (+1 more)

### Community 6 - "IMAPClient"
Cohesion: 0.17
Nodes (10): IMAPClient, github.com/emersion/go-imap/client.Client, imap.Literal, imap.Message, NewIMAPClient(), TestExtractTextPlain(), TestHandlePostProcessing(), TestMarkAsReadAndUnreadNilClient() (+2 more)

### Community 7 - "context.Context"
Cohesion: 0.13
Nodes (18): context.Context, time.Duration, fakeDispatcher, fakeRestarter, newClient(), NewClientWithBaseURL(), NewInClusterClient(), assertRestartDeploymentRequest() (+10 more)

### Community 8 - "net/http.Request"
Cohesion: 0.24
Nodes (6): failingRoundTripper, net/http.Request, net/http.Response, net/http.ResponseWriter, failingRoundTripper, failingHTTPClient

### Community 9 - "ADDED Requirements"
Cohesion: 0.09
Nodes (22): ADDED Requirements, Purpose, Requirement: Exact run link, Requirement: Failure notification for scheduled dispatch, Requirement: Fixed local timezone, Requirement: Graceful stop, Requirement: Optional schedule per workflow command, Requirement: Scheduled dispatch uses the command's dispatch path (+14 more)

### Community 10 - "GenericEmailProcessor"
Cohesion: 0.26
Nodes (3): regexp.Regexp, truncateString(), GenericEmailProcessor

### Community 11 - "go.uber.org/zap.Logger"
Cohesion: 0.11
Nodes (29): CI Job: Devidence Go CI, CI Pipeline Workflow, Devidence Go CI Reusable Workflow, main(), Client, github.com/go-telegram-bot-api/telegram-bot-api/v5.BotAPI, github.com/go-telegram-bot-api/telegram-bot-api/v5.BotCommand, github.com/go-telegram-bot-api/telegram-bot-api/v5.HTTPClient (+21 more)

### Community 12 - "Scheduler"
Cohesion: 0.16
Nodes (15): cron.Cron, cron.EntryID, cron.Schedule, cron.ScheduleParser, time.Location, New(), newWithParser(), ParseSpec() (+7 more)

### Community 13 - "ADDED Requirements"
Cohesion: 0.15
Nodes (12): ADDED Requirements, Purpose, Requirement: Create a Job from a CronJob, Requirement: Job command configuration, Requirement: Scheduled job commands, Scenario: Duplicate across types, Scenario: Jobs-only config starts the bot, Scenario: Kubernetes API error (+4 more)

### Community 19 - "Proposal"
Cohesion: 0.22
Nodes (8): Capabilities, Impact, Modified Capabilities, New Capabilities, Non-goals, Proposal, What Changes, Why

### Community 20 - "Proposal"
Cohesion: 0.22
Nodes (8): Capabilities, Impact, Modified Capabilities, New Capabilities, Non-goals, Proposal, What Changes, Why

### Community 22 - "Design"
Cohesion: 0.25
Nodes (7): Context, Decisions, Design, Goals / Non-Goals, Migration Plan, Open Questions, Risks / Trade-offs

### Community 23 - "Tasks"
Cohesion: 0.25
Nodes (7): 1. Pre-checks, 2. Config and wiring, 3. Job commands, 4. Dependency report, 5. Cross-repo, 6. Validation, Tasks

### Community 24 - "Tasks"
Cohesion: 0.25
Nodes (7): 1. Dependency and timezone, 2. Config, 3. Dispatch core, 4. Scheduler, 5. Cross-repo, 6. Validation, Tasks

### Community 25 - "Design"
Cohesion: 0.29
Nodes (6): Context, Decisions, Design, Goals / Non-Goals, Migration Plan, Risks / Trade-offs

## Knowledge Gaps
- **94 isolated node(s):** `automation-hub`, `Context`, `Goals / Non-Goals`, `Decisions`, `Risks / Trade-offs` (+89 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 129 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **7 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `🤖 Automation Hub` connect `🤖 Automation Hub` to `go.uber.org/zap.Logger`, `config.go`?**
  _High betweenness centrality (0.077) - this node is a cross-community bridge._
- **Why does `IMAPClient` connect `IMAPClient` to `go.uber.org/zap.Logger`, `config.go`?**
  _High betweenness centrality (0.051) - this node is a cross-community bridge._
- **Why does `qBittorrent Integration` connect `🤖 Automation Hub` to `go.uber.org/zap.Logger`?**
  _High betweenness centrality (0.048) - this node is a cross-community bridge._
- **Are the 8 inferred relationships involving `NewBotHandler()` (e.g. with `TestBotHandlerDispatchWorkflowFallsBackWhenRunURLIsEmpty()` and `TestBotHandlerDispatchWorkflowLinksToTheDispatchedRun()`) actually correct?**
  _`NewBotHandler()` has 8 INFERRED edges - model-reasoned connections that need verification._
- **What connects `automation-hub`, `Context`, `Goals / Non-Goals` to the rest of the system?**
  _94 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ADDED Requirements` be split into smaller, more focused modules?**
  _Cohesion score 0.08695652173913043 - nodes in this community are weakly interconnected._
- **Should `testing.T` be split into smaller, more focused modules?**
  _Cohesion score 0.10512820512820513 - nodes in this community are weakly interconnected._