# Graph Report - automation-hub  (2026-09-26)

## Corpus Check
- 28 files · ~12,484 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 273 nodes · 574 edges · 19 communities (12 shown, 7 thin omitted)
- Extraction: 90% EXTRACTED · 10% INFERRED · 0% AMBIGUOUS · INFERRED: 57 edges (avg confidence: 0.86)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `fcb89d26`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- go.uber.org/zap.Logger
- testing.T
- BotHandler
- 🤖 Automation Hub
- config.go
- GenericEmailProcessor
- IMAPClient
- context.Context
- net/http.Request
- NewIMAPClient
- Client
- NewWebhookHandler
- CLAUDE.md
- automation-hub-network Bridge Network
- Dependabot Docker Update Config (deployments/docker)
- Dependabot GitHub Actions Update Config
- Dependabot Go Modules Update Config
- Graphify Knowledge-Graph Workflow Rules
- automation-hub

## God Nodes (most connected - your core abstractions)
1. `Client` - 17 edges
2. `IMAPClient` - 16 edges
3. `BotHandler` - 15 edges
4. `NewBotHandler()` - 15 edges
5. `GenericEmailProcessor` - 15 edges
6. `NewGenericEmailProcessor()` - 14 edges
7. `Config` - 13 edges
8. `🤖 Automation Hub` - 12 edges
9. `NewClientWithBaseURL()` - 11 edges
10. `NewWebhookHandler()` - 10 edges

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

## Communities (19 total, 7 thin omitted)

### Community 0 - "go.uber.org/zap.Logger"
Cohesion: 0.24
Nodes (15): Client, workflowRun, workflowRunsResponse, go.uber.org/zap.Logger, net/http.Client, NewClient(), newClient(), NewClientWithBaseURL() (+7 more)

### Community 1 - "testing.T"
Cohesion: 0.14
Nodes (27): Client, net/http.HandlerFunc, net/http.ResponseWriter, testing.T, TestEmail(), TestTorrentNotification(), NewGenericEmailProcessor(), TestDecodeQuotedPrintable() (+19 more)

### Community 2 - "BotHandler"
Cohesion: 0.14
Nodes (21): github.com/go-telegram-bot-api/telegram-bot-api/v5.Message, github.com/go-telegram-bot-api/telegram-bot-api/v5.Update, sync.Mutex, BotHandler, deploymentRestarter, fakeMessenger, telegramMessenger, workflowDispatcher (+13 more)

### Community 3 - "🤖 Automation Hub"
Cohesion: 0.08
Nodes (25): CI Pipeline Workflow, Deploy Job: Build, Push & Deploy, Deploy Job: Resolve Build Args, Devidence CD Build Deploy Reusable Workflow, Build and Deploy Workflow, README, Authors and Acknowledgment, 🤖 Automation Hub (+17 more)

### Community 4 - "config.go"
Cohesion: 0.15
Nodes (20): GitHubConfig, ServerConfig, ServiceConfig, TelegramConfig, WebhookConfig, WorkflowBotConfig, Config, EmailConfig (+12 more)

### Community 5 - "GenericEmailProcessor"
Cohesion: 0.09
Nodes (12): mockNamedProcessor, regexp.Regexp, sync.WaitGroup, Email, EmailProcessor, TorrentNotification, truncateString(), NewProcessorManager() (+4 more)

### Community 6 - "IMAPClient"
Cohesion: 0.28
Nodes (4): IMAPClient, github.com/emersion/go-imap/client.Client, imap.Literal, imap.Message

### Community 7 - "context.Context"
Cohesion: 0.11
Nodes (19): context.Context, time.Duration, time.Time, fakeDispatcher, fakeRestarter, newClient(), NewClientWithBaseURL(), NewInClusterClient() (+11 more)

### Community 8 - "net/http.Request"
Cohesion: 0.32
Nodes (5): failingRoundTripper, net/http.Request, net/http.Response, failingRoundTripper, failingHTTPClient

### Community 9 - "NewIMAPClient"
Cohesion: 0.23
Nodes (9): CI Job: Devidence Go CI, Devidence Go CI Reusable Workflow, main(), NewIMAPClient(), TestExtractTextPlain(), TestHandlePostProcessing(), TestMarkAsReadAndUnreadNilClient(), TestNewIMAPClient() (+1 more)

### Community 11 - "Client"
Cohesion: 0.14
Nodes (17): github.com/go-telegram-bot-api/telegram-bot-api/v5.BotAPI, github.com/go-telegram-bot-api/telegram-bot-api/v5.BotCommand, github.com/go-telegram-bot-api/telegram-bot-api/v5.HTTPClient, WebhookProcessorConfig, GetWebhookConfig(), NewTorrentProcessor(), NewTorrentProcessorLegacy(), TestGetWebhookConfig() (+9 more)

### Community 12 - "NewWebhookHandler"
Cohesion: 0.21
Nodes (11): WebhookHandler, NewWebhookHandler(), TestHandleTorrentComplete_InvalidJSON(), TestHandleTorrentComplete_MissingWebhookConfig(), TestHandleTorrentComplete_Success(), TestNewWebhookHandler(), Adding New Email Services, Adding New Webhooks (+3 more)

## Knowledge Gaps
- **26 isolated node(s):** `automation-hub`, `graphify`, `✨ Features`, `📄 License`, `🚀 Option 1: Docker Compose (Recommended)` (+21 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 45 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **7 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `🤖 Automation Hub` connect `🤖 Automation Hub` to `NewWebhookHandler`, `config.go`?**
  _High betweenness centrality (0.154) - this node is a cross-community bridge._
- **Why does `IMAPClient` connect `IMAPClient` to `go.uber.org/zap.Logger`, `NewIMAPClient`, `config.go`?**
  _High betweenness centrality (0.104) - this node is a cross-community bridge._
- **What connects `automation-hub`, `graphify`, `✨ Features` to the rest of the system?**
  _26 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `testing.T` be split into smaller, more focused modules?**
  _Cohesion score 0.14482758620689656 - nodes in this community are weakly interconnected._
- **Should `BotHandler` be split into smaller, more focused modules?**
  _Cohesion score 0.14015151515151514 - nodes in this community are weakly interconnected._
- **Should `🤖 Automation Hub` be split into smaller, more focused modules?**
  _Cohesion score 0.08333333333333333 - nodes in this community are weakly interconnected._
- **Should `config.go` be split into smaller, more focused modules?**
  _Cohesion score 0.14624505928853754 - nodes in this community are weakly interconnected._