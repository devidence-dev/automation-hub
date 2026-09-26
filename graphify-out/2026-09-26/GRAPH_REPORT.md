# Graph Report - automation-hub  (2026-09-26)

## Corpus Check
- 24 files · ~12,386 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 320 nodes · 560 edges · 58 communities (14 shown, 44 thin omitted)
- Extraction: 91% EXTRACTED · 9% INFERRED · 0% AMBIGUOUS · INFERRED: 50 edges (avg confidence: 0.86)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- GitHub Client & Workflow Runs
- Webhook Handlers & Testing
- Telegram Bot & Command Dispatcher
- Deployment & Documentation
- Application Configuration
- Generic Email Processor
- IMAP Email Client & Ingestion
- Workflow Dispatcher Interface
- Processor Manager & Models
- Application Entrypoint & CI
- Torrent Webhook Processor
- Telegram API Client
- Webhook Routing & Docs
- Data Model Tests
- Project Architecture Guidelines
- Docker Compose Infrastructure
- Dependabot Docker Updates
- Dependabot GitHub Actions Updates
- Dependabot Go Updates
- External Type: BotAPI
- External Type: BotCommand
- Graphify Rules
- External Type: HandlerFunc
- External Type: HTTPClient
- External Type: T
- External Type: Logger
- External Type: Message
- External Type: Update
- External Type: Context
- External Type: T
- External Type: Update
- External Type: T
- External Type: Client
- External Type: Context
- External Type: Logger
- External Type: Message
- External Type: T
- External Type: Context
- External Type: Logger
- External Type: Request
- External Type: Response
- External Type: T
- External Type: Client
- External Type: Logger
- External Type: T
- External Type: T
- External Type: T
- External Type: Context
- External Type: Logger
- External Type: Update
- External Type: Client
- External Type: Request
- External Type: Response
- External Type: ResponseWriter
- External Type: T
- External Type: Literal
- Go Module Package
- External Type: Regexp

## God Nodes (most connected - your core abstractions)
1. `IMAPClient` - 16 edges
2. `BotHandler` - 15 edges
3. `NewBotHandler()` - 15 edges
4. `GenericEmailProcessor` - 14 edges
5. `NewGenericEmailProcessor()` - 14 edges
6. `Config` - 13 edges
7. `🤖 Automation Hub` - 12 edges
8. `NewClientWithBaseURL()` - 11 edges
9. `NewIMAPClient()` - 10 edges
10. `NewClientWithBaseURL()` - 10 edges

## Surprising Connections (you probably didn't know these)
- `📝 Step 2: Configure Services` --references--> `ServiceConfig`  [INFERRED]
  README.md → internal/config/config.go
- `🤖 Step 3: Setup Telegram Bot` --references--> `TelegramConfig`  [INFERRED]
  README.md → internal/config/config.go
- `▶️ Trigger GitHub Actions from Telegram` --references--> `WorkflowBotConfig`  [INFERRED]
  README.md → internal/config/config.go
- `🛡️ Security & Quality` --conceptually_related_to--> `CI Pipeline Workflow`  [INFERRED]
  README.md → .github/workflows/ci.yml
- `🐳 Deployment` --conceptually_related_to--> `Build and Deploy Workflow`  [INFERRED]
  README.md → .github/workflows/deploy.yml

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Dependabot Multi-Ecosystem Update Strategy (gomod, docker, github-actions)** — _github_dependabot_gomod_updates, _github_dependabot_docker_updates, _github_dependabot_github_actions_updates [EXTRACTED 1.00]

## Communities (58 total, 44 thin omitted)

### Community 0 - "GitHub Client & Workflow Runs"
Cohesion: 0.10
Nodes (33): Client, failingRoundTripper, workflowRun, workflowRunsResponse, go.uber.org/zap.Logger, net/http.Client, net/http.Request, net/http.Response (+25 more)

### Community 1 - "Webhook Handlers & Testing"
Cohesion: 0.10
Nodes (34): Client, net/http.HandlerFunc, net/http.ResponseWriter, testing.T, TestHandleTorrentComplete_InvalidJSON(), TestHandleTorrentComplete_MissingWebhookConfig(), TestHandleTorrentComplete_Success(), TestNewWebhookHandler() (+26 more)

### Community 2 - "Telegram Bot & Command Dispatcher"
Cohesion: 0.14
Nodes (21): github.com/go-telegram-bot-api/telegram-bot-api/v5.Message, github.com/go-telegram-bot-api/telegram-bot-api/v5.Update, sync.Mutex, BotHandler, deploymentRestarter, fakeMessenger, telegramMessenger, workflowDispatcher (+13 more)

### Community 3 - "Deployment & Documentation"
Cohesion: 0.09
Nodes (23): Deploy Job: Build, Push & Deploy, Deploy Job: Resolve Build Args, Devidence CD Build Deploy Reusable Workflow, Build and Deploy Workflow, README, Authors and Acknowledgment, 🤖 Automation Hub, Common Issues (+15 more)

### Community 4 - "Application Configuration"
Cohesion: 0.16
Nodes (20): Config, GitHubConfig, ServerConfig, ServiceConfig, TelegramConfig, WebhookConfig, WorkflowBotConfig, EmailConfig (+12 more)

### Community 5 - "Generic Email Processor"
Cohesion: 0.15
Nodes (6): mockNamedProcessor, automation-hub/internal/models.Email, regexp.Regexp, TestTruncateString(), truncateString(), GenericEmailProcessor

### Community 6 - "IMAP Email Client & Ingestion"
Cohesion: 0.26
Nodes (5): IMAPClient, automation-hub/internal/models.EmailProcessor, github.com/emersion/go-imap/client.Client, imap.Literal, imap.Message

### Community 7 - "Workflow Dispatcher Interface"
Cohesion: 0.17
Nodes (6): context.Context, time.Duration, time.Time, fakeDispatcher, fakeRestarter, deploymentStatus

### Community 8 - "Processor Manager & Models"
Cohesion: 0.21
Nodes (9): Client, Context, Logger, NewProcessorManager(), Email, EmailProcessor, TorrentNotification, Manager (+1 more)

### Community 9 - "Application Entrypoint & CI"
Cohesion: 0.19
Nodes (11): CI Job: Devidence Go CI, CI Pipeline Workflow, Devidence Go CI Reusable Workflow, main(), NewIMAPClient(), TestExtractTextPlain(), TestHandlePostProcessing(), TestMarkAsReadAndUnreadNilClient() (+3 more)

### Community 10 - "Torrent Webhook Processor"
Cohesion: 0.35
Nodes (9): WebhookProcessorConfig, Request, ResponseWriter, GetWebhookConfig(), Client, Logger, NewTorrentProcessor(), NewTorrentProcessorLegacy() (+1 more)

### Community 11 - "Telegram API Client"
Cohesion: 0.27
Nodes (8): github.com/go-telegram-bot-api/telegram-bot-api/v5.BotAPI, github.com/go-telegram-bot-api/telegram-bot-api/v5.BotCommand, github.com/go-telegram-bot-api/telegram-bot-api/v5.HTTPClient, Client, NewClient(), newClient(), newHTTPClient(), parseInt64()

### Community 12 - "Webhook Routing & Docs"
Cohesion: 0.27
Nodes (9): WebhookHandler, Client, Logger, NewWebhookHandler(), Adding New Email Services, Adding New Webhooks, � API & Webhooks, Available Endpoints (+1 more)

### Community 13 - "Data Model Tests"
Cohesion: 0.67
Nodes (3): T, TestEmail(), TestTorrentNotification()

## Knowledge Gaps
- **26 isolated node(s):** `graphify`, `README`, `✨ Features`, `📋 Prerequisites`, `🔐 Step 1: Setup Configuration` (+21 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 84 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **44 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `🤖 Automation Hub` connect `Deployment & Documentation` to `Application Entrypoint & CI`, `Webhook Routing & Docs`, `Application Configuration`?**
  _High betweenness centrality (0.119) - this node is a cross-community bridge._
- **Why does `NewGenericEmailProcessor()` connect `Webhook Handlers & Testing` to `GitHub Client & Workflow Runs`, `Application Configuration`, `Generic Email Processor`, `Processor Manager & Models`, `Telegram API Client`?**
  _High betweenness centrality (0.111) - this node is a cross-community bridge._
- **Why does `⚙️ Configuration` connect `Application Configuration` to `Deployment & Documentation`?**
  _High betweenness centrality (0.104) - this node is a cross-community bridge._
- **Are the 6 inferred relationships involving `NewBotHandler()` (e.g. with `TestBotHandlerDispatchWorkflowFallsBackWhenRunLookupFails()` and `TestBotHandlerDispatchWorkflowLinksToTheDispatchedRun()`) actually correct?**
  _`NewBotHandler()` has 6 INFERRED edges - model-reasoned connections that need verification._
- **What connects `graphify`, `README`, `✨ Features` to the rest of the system?**
  _26 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `GitHub Client & Workflow Runs` be split into smaller, more focused modules?**
  _Cohesion score 0.10384615384615385 - nodes in this community are weakly interconnected._
- **Should `Webhook Handlers & Testing` be split into smaller, more focused modules?**
  _Cohesion score 0.1039136302294197 - nodes in this community are weakly interconnected._