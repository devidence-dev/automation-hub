# Graph Report - automation-hub  (2026-08-21)

## Corpus Check
- 24 files · ~9,088 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 232 nodes · 374 edges · 16 communities (14 shown, 2 thin omitted)
- Extraction: 88% EXTRACTED · 11% INFERRED · 1% AMBIGUOUS · INFERRED: 43 edges (avg confidence: 0.81)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `d82ee34a`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Torrent Webhook Handling
- Generic Email Processor
- Processor Manager & Models
- Deployment & Docs
- IMAP Email Client
- CI Security & Quality Tooling
- Telegram Notifications
- Email Message Parsing
- Model Tests
- Root Package
- CLAUDE.md
- NewGenericEmailProcessor
- NewWebhookHandler
- NewIMAPClient
- parseInt64
- BotHandler

## God Nodes (most connected - your core abstractions)
1. `IMAPClient` - 16 edges
2. `GenericEmailProcessor` - 14 edges
3. `NewGenericEmailProcessor()` - 14 edges
4. `Config` - 12 edges
5. `🤖 Automation Hub` - 12 edges
6. `BotHandler` - 10 edges
7. `NewWebhookHandler()` - 9 edges
8. `Email` - 9 edges
9. `NewIMAPClient()` - 9 edges
10. `Manager` - 9 edges

## Surprising Connections (you probably didn't know these)
- `Security & Quality Tooling List (README claims)` --conceptually_related_to--> `CI Job: OSV Scanner (dependency vulnerabilities)`  [AMBIGUOUS]
  README.md → .github/workflows/ci.yml
- `Security & Quality Tooling List (README claims)` --conceptually_related_to--> `CI Job: OWASP Dependency-Check (CVE scanner)`  [AMBIGUOUS]
  README.md → .github/workflows/ci.yml
- `Security & Quality Tooling List (README claims)` --conceptually_related_to--> `Dependabot Go Modules Update Config`  [INFERRED]
  README.md → .github/dependabot.yml
- `Security & Quality Tooling List (README claims)` --conceptually_related_to--> `CI Job: Lint Code (golangci-lint)`  [INFERRED]
  README.md → .github/workflows/ci.yml
- `Security & Quality Tooling List (README claims)` --conceptually_related_to--> `CI Job: SAST (Semgrep SARIF, self-hosted)`  [INFERRED]
  README.md → .github/workflows/ci.yml

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **CI Pipeline Jobs (Tests, Lint, OSV, Dependency-Check, SAST)** — _github_workflows_ci_test_and_build, _github_workflows_ci_lint, _github_workflows_ci_osv_scan, _github_workflows_ci_dependency_check, _github_workflows_ci_sast [EXTRACTED 1.00]
- **Build and Deploy Pipeline Jobs (Version, Build, Deploy)** — _github_workflows_deploy_version, _github_workflows_deploy_build, _github_workflows_deploy_deploy [EXTRACTED 1.00]
- **Dependabot Multi-Ecosystem Update Strategy (gomod, docker, github-actions)** — _github_dependabot_gomod_updates, _github_dependabot_docker_updates, _github_dependabot_github_actions_updates [EXTRACTED 1.00]

## Communities (16 total, 2 thin omitted)

### Community 0 - "Torrent Webhook Handling"
Cohesion: 0.23
Nodes (15): Config, EmailConfig, GitHubConfig, ServerConfig, ServiceConfig, TelegramConfig, WebhookConfig, WorkflowBotConfig (+7 more)

### Community 1 - "Generic Email Processor"
Cohesion: 0.16
Nodes (17): ServiceProcessorConfig, Client, Logger, NewGenericEmailProcessor(), T, TestDecodeQuotedPrintable(), TestExtractCode(), TestExtractPerplexityCode() (+9 more)

### Community 2 - "Processor Manager & Models"
Cohesion: 0.12
Nodes (13): mockNamedProcessor, Client, Context, Logger, NewProcessorManager(), T, TestProcessorManager(), TestProcessorManager_CanceledContextAsync() (+5 more)

### Community 3 - "Deployment & Docs"
Cohesion: 0.12
Nodes (21): Dependabot Docker Update Config (deployments/docker), Dependabot GitHub Actions Update Config, Dependabot Go Modules Update Config, CI Job: OWASP Dependency-Check (CVE scanner), CI Job: Lint Code (golangci-lint), CI Job: OSV Scanner (dependency vulnerabilities), CI Job: SAST (Semgrep SARIF, self-hosted), CI Job: Tests & Build (+13 more)

### Community 4 - "IMAP Email Client"
Cohesion: 0.26
Nodes (5): IMAPClient, Client, Context, Message, Literal

### Community 5 - "CI Security & Quality Tooling"
Cohesion: 0.22
Nodes (7): fakeDispatcher, fakeMessenger, commandUpdate(), Context, T, Update, TestBotHandlerHandle()

### Community 6 - "Telegram Notifications"
Cohesion: 0.07
Nodes (29): 🔄 Adding New Email Services, 🆕 Adding New Webhooks, � API & Webhooks, Authors and acknowledgment 🛡, 🤖 Automation Hub, 📡 Available Endpoints, 🚨 Common Issues, ⚙️ Configuration (+21 more)

### Community 7 - "Email Message Parsing"
Cohesion: 0.23
Nodes (13): WebhookProcessorConfig, GetWebhookConfig(), Client, Logger, NewTorrentProcessor(), NewTorrentProcessorLegacy(), T, TestGetWebhookConfig() (+5 more)

### Community 8 - "Model Tests"
Cohesion: 0.67
Nodes (3): T, TestEmail(), TestTorrentNotification()

### Community 11 - "NewGenericEmailProcessor"
Cohesion: 0.29
Nodes (7): Client, Context, Logger, NewClient(), NewClientWithBaseURL(), T, TestDispatchWorkflow()

### Community 12 - "NewWebhookHandler"
Cohesion: 0.35
Nodes (9): WebhookHandler, Client, Logger, NewWebhookHandler(), T, TestHandleTorrentComplete_InvalidJSON(), TestHandleTorrentComplete_MissingWebhookConfig(), TestHandleTorrentComplete_Success() (+1 more)

### Community 13 - "NewIMAPClient"
Cohesion: 0.38
Nodes (8): Logger, NewIMAPClient(), T, TestExtractTextPlain(), TestHandlePostProcessing(), TestMarkAsReadAndUnreadNilClient(), TestNewIMAPClient(), TestParseMessage()

### Community 14 - "parseInt64"
Cohesion: 0.15
Nodes (12): BotAPI, BotCommand, Context, Logger, Update, NewClient(), parseInt64(), T (+4 more)

### Community 15 - "BotHandler"
Cohesion: 0.28
Nodes (10): WorkflowCommandConfig, BotHandler, telegramMessenger, workflowDispatcher, Logger, Message, Update, isAllowed() (+2 more)

## Ambiguous Edges - Review These
- `CI Job: OSV Scanner (dependency vulnerabilities)` → `Security & Quality Tooling List (README claims)`  [AMBIGUOUS]
  README.md · relation: conceptually_related_to
- `CI Job: OWASP Dependency-Check (CVE scanner)` → `Security & Quality Tooling List (README claims)`  [AMBIGUOUS]
  README.md · relation: conceptually_related_to

## Knowledge Gaps
- **27 isolated node(s):** `automation-hub`, `graphify`, `🛡️ Security & Quality`, `✨ Features`, `📋 Prerequisites` (+22 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **2 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **What is the exact relationship between `CI Job: OSV Scanner (dependency vulnerabilities)` and `Security & Quality Tooling List (README claims)`?**
  _Edge tagged AMBIGUOUS (relation: conceptually_related_to) - confidence is low._
- **What is the exact relationship between `CI Job: OWASP Dependency-Check (CVE scanner)` and `Security & Quality Tooling List (README claims)`?**
  _Edge tagged AMBIGUOUS (relation: conceptually_related_to) - confidence is low._
- **Why does `EmailConfig` connect `Torrent Webhook Handling` to `Processor Manager & Models`, `IMAP Email Client`, `NewIMAPClient`?**
  _High betweenness centrality (0.151) - this node is a cross-community bridge._
- **Why does `Config` connect `Torrent Webhook Handling` to `NewWebhookHandler`, `Email Message Parsing`, `BotHandler`?**
  _High betweenness centrality (0.122) - this node is a cross-community bridge._
- **Why does `WorkflowCommandConfig` connect `BotHandler` to `Torrent Webhook Handling`?**
  _High betweenness centrality (0.111) - this node is a cross-community bridge._
- **Are the 9 inferred relationships involving `NewGenericEmailProcessor()` (e.g. with `TestDecodeQuotedPrintable()` and `TestExtractCode()`) actually correct?**
  _`NewGenericEmailProcessor()` has 9 INFERRED edges - model-reasoned connections that need verification._
- **What connects `automation-hub`, `graphify`, `🛡️ Security & Quality` to the rest of the system?**
  _27 weakly-connected nodes found - possible documentation gaps or missing edges._