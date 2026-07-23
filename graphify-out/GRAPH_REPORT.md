# Graph Report - automation-hub  (2026-07-22)

## Corpus Check
- 20 files · ~7,540 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 182 nodes · 303 edges · 15 communities (13 shown, 2 thin omitted)
- Extraction: 86% EXTRACTED · 13% INFERRED · 1% AMBIGUOUS · INFERRED: 40 edges (avg confidence: 0.81)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `e149ae47`
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

## God Nodes (most connected - your core abstractions)
1. `Client` - 22 edges
2. `IMAPClient` - 16 edges
3. `GenericEmailProcessor` - 14 edges
4. `NewGenericEmailProcessor()` - 14 edges
5. `🤖 Automation Hub` - 12 edges
6. `Config` - 9 edges
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

## Communities (15 total, 2 thin omitted)

### Community 0 - "Torrent Webhook Handling"
Cohesion: 0.26
Nodes (12): Config, EmailConfig, ServerConfig, ServiceConfig, ServiceProcessorConfig, TelegramConfig, WebhookConfig, Load() (+4 more)

### Community 1 - "Generic Email Processor"
Cohesion: 0.26
Nodes (4): Logger, truncateString(), GenericEmailProcessor, Regexp

### Community 2 - "Processor Manager & Models"
Cohesion: 0.12
Nodes (12): mockNamedProcessor, Context, Logger, NewProcessorManager(), T, TestProcessorManager(), TestProcessorManager_CanceledContextAsync(), Email (+4 more)

### Community 3 - "Deployment & Docs"
Cohesion: 0.18
Nodes (13): Dependabot Docker Update Config (deployments/docker), Deploy Job: Build & Push Image to Zot Registry, Deploy Job: Deploy to Production (homelab repo), Deploy Job: Calculate semver Version, Graphify Knowledge-Graph Workflow Rules, automation-hub-network Bridge Network, automation-hub Docker Compose Service, Automation Hub Project Overview (+5 more)

### Community 4 - "IMAP Email Client"
Cohesion: 0.20
Nodes (8): BotAPI, IMAPClient, Context, Logger, NewClient(), Literal, Message, Client

### Community 5 - "CI Security & Quality Tooling"
Cohesion: 0.36
Nodes (8): Dependabot GitHub Actions Update Config, Dependabot Go Modules Update Config, CI Job: OWASP Dependency-Check (CVE scanner), CI Job: Lint Code (golangci-lint), CI Job: OSV Scanner (dependency vulnerabilities), CI Job: SAST (Semgrep SARIF, self-hosted), CI Job: Tests & Build, Security & Quality Tooling List (README claims)

### Community 6 - "Telegram Notifications"
Cohesion: 0.07
Nodes (28): 🔄 Adding New Email Services, 🆕 Adding New Webhooks, � API & Webhooks, Authors and acknowledgment 🛡, 🤖 Automation Hub, 📡 Available Endpoints, 🚨 Common Issues, ⚙️ Configuration (+20 more)

### Community 7 - "Email Message Parsing"
Cohesion: 0.24
Nodes (12): WebhookProcessorConfig, GetWebhookConfig(), Logger, NewTorrentProcessor(), NewTorrentProcessorLegacy(), T, TestGetWebhookConfig(), TestNewTorrentProcessor() (+4 more)

### Community 8 - "Model Tests"
Cohesion: 0.67
Nodes (3): T, TestEmail(), TestTorrentNotification()

### Community 11 - "NewGenericEmailProcessor"
Cohesion: 0.39
Nodes (11): NewGenericEmailProcessor(), T, TestDecodeQuotedPrintable(), TestExtractCode(), TestExtractPerplexityCode(), TestNewGenericEmailProcessor_BuiltInPatterns(), TestNewGenericEmailProcessor_CustomPattern(), TestNewGenericEmailProcessor_InvalidCustomPatternFallback() (+3 more)

### Community 12 - "NewWebhookHandler"
Cohesion: 0.38
Nodes (8): WebhookHandler, Logger, NewWebhookHandler(), T, TestHandleTorrentComplete_InvalidJSON(), TestHandleTorrentComplete_MissingWebhookConfig(), TestHandleTorrentComplete_Success(), TestNewWebhookHandler()

### Community 13 - "NewIMAPClient"
Cohesion: 0.38
Nodes (8): Logger, NewIMAPClient(), T, TestExtractTextPlain(), TestHandlePostProcessing(), TestMarkAsReadAndUnreadNilClient(), TestNewIMAPClient(), TestParseMessage()

### Community 14 - "parseInt64"
Cohesion: 0.38
Nodes (5): parseInt64(), T, TestParseInt64(), TestSendMessageInvalidChatID(), TestSendMessageNilClientOrBot()

## Ambiguous Edges - Review These
- `CI Job: OSV Scanner (dependency vulnerabilities)` → `Security & Quality Tooling List (README claims)`  [AMBIGUOUS]
  README.md · relation: conceptually_related_to
- `CI Job: OWASP Dependency-Check (CVE scanner)` → `Security & Quality Tooling List (README claims)`  [AMBIGUOUS]
  README.md · relation: conceptually_related_to

## Knowledge Gaps
- **26 isolated node(s):** `automation-hub`, `graphify`, `🛡️ Security & Quality`, `✨ Features`, `📋 Prerequisites` (+21 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **2 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **What is the exact relationship between `CI Job: OSV Scanner (dependency vulnerabilities)` and `Security & Quality Tooling List (README claims)`?**
  _Edge tagged AMBIGUOUS (relation: conceptually_related_to) - confidence is low._
- **What is the exact relationship between `CI Job: OWASP Dependency-Check (CVE scanner)` and `Security & Quality Tooling List (README claims)`?**
  _Edge tagged AMBIGUOUS (relation: conceptually_related_to) - confidence is low._
- **Why does `Client` connect `IMAP Email Client` to `Generic Email Processor`, `Processor Manager & Models`, `Email Message Parsing`, `NewGenericEmailProcessor`, `NewWebhookHandler`, `parseInt64`?**
  _High betweenness centrality (0.240) - this node is a cross-community bridge._
- **Why does `NewGenericEmailProcessor()` connect `NewGenericEmailProcessor` to `Torrent Webhook Handling`, `Generic Email Processor`, `Processor Manager & Models`, `IMAP Email Client`?**
  _High betweenness centrality (0.087) - this node is a cross-community bridge._
- **Why does `GenericEmailProcessor` connect `Generic Email Processor` to `Torrent Webhook Handling`, `Processor Manager & Models`, `NewGenericEmailProcessor`, `IMAP Email Client`?**
  _High betweenness centrality (0.073) - this node is a cross-community bridge._
- **Are the 9 inferred relationships involving `NewGenericEmailProcessor()` (e.g. with `TestDecodeQuotedPrintable()` and `TestExtractCode()`) actually correct?**
  _`NewGenericEmailProcessor()` has 9 INFERRED edges - model-reasoned connections that need verification._
- **What connects `automation-hub`, `graphify`, `🛡️ Security & Quality` to the rest of the system?**
  _26 weakly-connected nodes found - possible documentation gaps or missing edges._