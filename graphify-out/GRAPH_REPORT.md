# Graph Report - automation-hub  (2026-10-11)

## Corpus Check
- 53 files · ~43,715 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 702 nodes · 1395 edges · 33 communities (26 shown, 7 thin omitted)
- Extraction: 93% EXTRACTED · 7% INFERRED · 0% AMBIGUOUS · INFERRED: 104 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `e718f18b`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Requirement: Document tabs
- testing.T
- telegram_webhook_test.go
- 🤖 Automation Hub
- Config
- IMAPClient
- ADDED Requirements
- context.Context
- NewGenericEmailProcessor
- Requirements
- go.uber.org/zap.Logger
- dependency_report_test.go
- ADDED Requirements
- CLAUDE.md
- automation-hub-network Bridge Network
- Dependabot Docker Update Config (deployments/docker)
- Dependabot GitHub Actions Update Config
- Dependabot Go Modules Update Config
- Requirements
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
1. `Config` - 24 edges
2. `Client` - 24 edges
3. `BotHandler` - 23 edges
4. `IMAPClient` - 16 edges
5. `buildDocumentData()` - 15 edges
6. `GenericEmailProcessor` - 15 edges
7. `NewGenericEmailProcessor()` - 15 edges
8. `setupWorkflowBot()` - 14 edges
9. `NewBotHandler()` - 14 edges
10. `renovateReport` - 13 edges

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

## Communities (33 total, 7 thin omitted)

### Community 0 - "Requirement: Document tabs"
Cohesion: 0.05
Nodes (40): ADDED Requirements, MODIFIED Requirements, REMOVED Requirements, Requirement: All up to date, Requirement: Delivery, Requirement: Document tabs, Requirement: HTML report document, Requirement: Nothing pending document (+32 more)

### Community 1 - "testing.T"
Cohesion: 0.06
Nodes (65): noopSender, Client, TestRegisterRoutesDependencyReport(), mime/multipart.FileHeader, mime/multipart.Form, net/http.HandlerFunc, net/http/httptest.Server, testing.T (+57 more)

### Community 2 - "telegram_webhook_test.go"
Cohesion: 0.21
Nodes (18): sync.Mutex, fakeDispatcher, fakeJobCreator, fakeMessenger, NewBotHandler(), assertScheduledRun(), commandUpdate(), TestBotHandlerDispatchWorkflowFallsBackWhenRunURLIsEmpty() (+10 more)

### Community 3 - "🤖 Automation Hub"
Cohesion: 0.06
Nodes (32): CI Job: Devidence Go CI, CI Pipeline Workflow, Devidence Go CI Reusable Workflow, Deploy Job: Build, Push & Deploy, Deploy Job: Resolve Build Args, Devidence CD Build Deploy Reusable Workflow, Build and Deploy Workflow, README (+24 more)

### Community 4 - "Config"
Cohesion: 0.08
Nodes (43): GitHubConfig, ServerConfig, ServiceConfig, TelegramConfig, WebhookConfig, WorkflowBotConfig, github.com/go-telegram-bot-api/telegram-bot-api/v5.Message, BotHandlerDeps (+35 more)

### Community 5 - "IMAPClient"
Cohesion: 0.08
Nodes (19): IMAPClient, mockNamedProcessor, github.com/emersion/go-imap/client.Client, sync.WaitGroup, imap.Literal, imap.Message, Email, EmailProcessor (+11 more)

### Community 6 - "ADDED Requirements"
Cohesion: 0.09
Nodes (22): ADDED Requirements, Purpose, Requirement: Exact run link, Requirement: Failure notification for scheduled dispatch, Requirement: Fixed local timezone, Requirement: Graceful stop, Requirement: Optional schedule per workflow command, Requirement: Scheduled dispatch uses the command's dispatch path (+14 more)

### Community 7 - "context.Context"
Cohesion: 0.10
Nodes (23): context.Context, encoding/json.RawMessage, github.com/go-telegram-bot-api/telegram-bot-api/v5.Update, net/http.Client, time.Duration, fakeRestarter, Client, NewClient() (+15 more)

### Community 8 - "NewGenericEmailProcessor"
Cohesion: 0.14
Nodes (14): regexp.Regexp, NewGenericEmailProcessor(), TestDecodeQuotedPrintable(), TestExtractCode(), TestExtractCode_IgnoresHexColor(), TestExtractPerplexityCode(), TestNewGenericEmailProcessor_BuiltInPatterns(), TestNewGenericEmailProcessor_CustomPattern() (+6 more)

### Community 9 - "Requirements"
Cohesion: 0.09
Nodes (22): Purpose, Requirement: Exact run link, Requirement: Failure notification for scheduled dispatch, Requirement: Fixed local timezone, Requirement: Graceful stop, Requirement: Optional schedule per workflow command, Requirement: Scheduled dispatch uses the command's dispatch path, Requirement: Startup schedule log (+14 more)

### Community 11 - "go.uber.org/zap.Logger"
Cohesion: 0.08
Nodes (43): reportSender, main(), newGitHubClient(), newHTTPServer(), newKubeClient(), registerBotCommands(), registerRoutes(), registerSchedules() (+35 more)

### Community 12 - "dependency_report_test.go"
Cohesion: 0.09
Nodes (42): failingRoundTripper, net/http.Request, net/http.Response, net/http.ResponseWriter, DependencyReportHandler, fakeSender, packageFile, problem (+34 more)

### Community 13 - "ADDED Requirements"
Cohesion: 0.09
Nodes (22): ADDED Requirements, Purpose, Requirement: All up to date, Requirement: Authenticated endpoint, Requirement: Delivery, Requirement: Per-repository message, Requirement: Renovate failure message, Requirement: Request handling (+14 more)

### Community 19 - "Requirements"
Cohesion: 0.09
Nodes (22): dependency-report Specification, Purpose, Requirement: All up to date, Requirement: Authenticated endpoint, Requirement: Delivery, Requirement: Per-repository message, Requirement: Renovate failure message, Requirement: Request handling (+14 more)

### Community 22 - "dependency_report_document.go"
Cohesion: 0.11
Nodes (45): time.Time, depKey, distItem, docRow, documentData, renovateReport, repoPending, reportTab (+37 more)

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
Cohesion: 0.20
Nodes (9): Context, Decisions, Decisions (resolved), Design, Goals / Non-Goals, Implementation notes, Migration Plan, Renovate 44.151.0 findings (read from source) (+1 more)

### Community 29 - "Proposal"
Cohesion: 0.22
Nodes (8): Capabilities, Impact, Modified Capabilities, New Capabilities, Non-goals, Proposal, What Changes, Why

### Community 30 - "Proposal"
Cohesion: 0.22
Nodes (8): Capabilities, Impact, Modified Capabilities, New Capabilities, Non-goals, Proposal, What Changes, Why

### Community 31 - "Tasks"
Cohesion: 0.22
Nodes (8): 1. Pre-checks, 2. Telegram client, 3. Report document, 4. Handler wiring and docs, 5. Cross-repo: scan homelab and retire icr, 6. Caption and single message, 7. Validation, Tasks

### Community 32 - "Design"
Cohesion: 0.25
Nodes (7): Context, Decisions, Design, Goals / Non-Goals, Migration Plan, Open Questions, Risks / Trade-offs

### Community 33 - "Tasks"
Cohesion: 0.25
Nodes (7): 1. Pre-checks, 2. Config and wiring, 3. Job commands, 4. Dependency report, 5. Cross-repo, 6. Validation, Tasks

## Knowledge Gaps
- **176 isolated node(s):** `automation-hub`, `Context`, `Renovate 44.151.0 findings (read from source)`, `Goals / Non-Goals`, `Decisions` (+171 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 225 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **7 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `🤖 Automation Hub` connect `🤖 Automation Hub` to `Config`?**
  _High betweenness centrality (0.040) - this node is a cross-community bridge._
- **Why does `Config` connect `Config` to `testing.T`, `go.uber.org/zap.Logger`, `dependency_report_test.go`?**
  _High betweenness centrality (0.030) - this node is a cross-community bridge._
- **Why does `IMAPClient` connect `IMAPClient` to `go.uber.org/zap.Logger`, `Config`?**
  _High betweenness centrality (0.027) - this node is a cross-community bridge._
- **What connects `automation-hub`, `Context`, `Renovate 44.151.0 findings (read from source)` to the rest of the system?**
  _176 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Requirement: Document tabs` be split into smaller, more focused modules?**
  _Cohesion score 0.04878048780487805 - nodes in this community are weakly interconnected._
- **Should `testing.T` be split into smaller, more focused modules?**
  _Cohesion score 0.06293965198074787 - nodes in this community are weakly interconnected._
- **Should `🤖 Automation Hub` be split into smaller, more focused modules?**
  _Cohesion score 0.0625 - nodes in this community are weakly interconnected._