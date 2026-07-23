# Graph Report - automation-hub  (2026-07-22)

## Corpus Check
- 13 files · ~5,501 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 134 nodes · 211 edges · 11 communities (9 shown, 2 thin omitted)
- Extraction: 93% EXTRACTED · 6% INFERRED · 1% AMBIGUOUS · INFERRED: 13 edges (avg confidence: 0.83)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `39badec2`
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

## God Nodes (most connected - your core abstractions)
1. `Client` - 22 edges
2. `IMAPClient` - 16 edges
3. `GenericEmailProcessor` - 14 edges
4. `🤖 Automation Hub` - 12 edges
5. `Config` - 9 edges
6. `Manager` - 9 edges
7. `Email` - 7 edges
8. `TorrentProcessor` - 7 edges
9. `EmailConfig` - 6 edges
10. `WebhookHandler` - 6 edges

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

## Communities (11 total, 2 thin omitted)

### Community 0 - "Torrent Webhook Handling"
Cohesion: 0.28
Nodes (11): Config, EmailConfig, ServerConfig, ServiceConfig, ServiceProcessorConfig, TelegramConfig, WebhookConfig, WebhookHandler (+3 more)

### Community 1 - "Generic Email Processor"
Cohesion: 0.26
Nodes (5): Logger, NewGenericEmailProcessor(), truncateString(), GenericEmailProcessor, Regexp

### Community 2 - "Processor Manager & Models"
Cohesion: 0.20
Nodes (8): Context, Logger, NewProcessorManager(), Email, EmailProcessor, TorrentNotification, Manager, WaitGroup

### Community 3 - "Deployment & Docs"
Cohesion: 0.18
Nodes (13): Dependabot Docker Update Config (deployments/docker), Deploy Job: Build & Push Image to Zot Registry, Deploy Job: Deploy to Production (homelab repo), Deploy Job: Calculate semver Version, Graphify Knowledge-Graph Workflow Rules, automation-hub-network Bridge Network, automation-hub Docker Compose Service, Automation Hub Project Overview (+5 more)

### Community 4 - "IMAP Email Client"
Cohesion: 0.15
Nodes (11): BotAPI, IMAPClient, Context, Logger, NewIMAPClient(), Logger, NewClient(), parseInt64() (+3 more)

### Community 5 - "CI Security & Quality Tooling"
Cohesion: 0.36
Nodes (8): Dependabot GitHub Actions Update Config, Dependabot Go Modules Update Config, CI Job: OWASP Dependency-Check (CVE scanner), CI Job: Lint Code (golangci-lint), CI Job: OSV Scanner (dependency vulnerabilities), CI Job: SAST (Semgrep SARIF, self-hosted), CI Job: Tests & Build, Security & Quality Tooling List (README claims)

### Community 6 - "Telegram Notifications"
Cohesion: 0.07
Nodes (28): 🔄 Adding New Email Services, 🆕 Adding New Webhooks, � API & Webhooks, Authors and acknowledgment 🛡, 🤖 Automation Hub, 📡 Available Endpoints, 🚨 Common Issues, ⚙️ Configuration (+20 more)

### Community 7 - "Email Message Parsing"
Cohesion: 0.36
Nodes (8): WebhookProcessorConfig, GetWebhookConfig(), Logger, NewTorrentProcessor(), NewTorrentProcessorLegacy(), TorrentProcessor, Request, ResponseWriter

### Community 8 - "Model Tests"
Cohesion: 0.67
Nodes (3): TestEmail(), TestTorrentNotification(), T

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
- **Why does `Client` connect `IMAP Email Client` to `Torrent Webhook Handling`, `Generic Email Processor`, `Processor Manager & Models`, `Email Message Parsing`?**
  _High betweenness centrality (0.178) - this node is a cross-community bridge._
- **Why does `GenericEmailProcessor` connect `Generic Email Processor` to `Torrent Webhook Handling`, `Processor Manager & Models`, `IMAP Email Client`?**
  _High betweenness centrality (0.077) - this node is a cross-community bridge._
- **Why does `IMAPClient` connect `IMAP Email Client` to `Torrent Webhook Handling`?**
  _High betweenness centrality (0.057) - this node is a cross-community bridge._
- **What connects `automation-hub`, `graphify`, `🛡️ Security & Quality` to the rest of the system?**
  _26 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Telegram Notifications` be split into smaller, more focused modules?**
  _Cohesion score 0.06896551724137931 - nodes in this community are weakly interconnected._