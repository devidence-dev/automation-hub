# Graph Report - .  (2026-07-14)

## Corpus Check
- Corpus is ~5,370 words - fits in a single context window. You may not need a graph.

## Summary
- 103 nodes · 182 edges · 10 communities (8 shown, 2 thin omitted)
- Extraction: 92% EXTRACTED · 7% INFERRED · 1% AMBIGUOUS · INFERRED: 13 edges (avg confidence: 0.83)
- Token cost: 50,993 input · 0 output

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

## God Nodes (most connected - your core abstractions)
1. `Client` - 22 edges
2. `IMAPClient` - 16 edges
3. `GenericEmailProcessor` - 14 edges
4. `Config` - 9 edges
5. `Manager` - 9 edges
6. `Email` - 7 edges
7. `TorrentProcessor` - 7 edges
8. `EmailConfig` - 6 edges
9. `WebhookHandler` - 6 edges
10. `NewGenericEmailProcessor()` - 6 edges

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

## Communities (10 total, 2 thin omitted)

### Community 0 - "Torrent Webhook Handling"
Cohesion: 0.19
Nodes (16): Config, ServerConfig, TelegramConfig, WebhookConfig, WebhookProcessorConfig, WebhookHandler, Load(), Logger (+8 more)

### Community 1 - "Generic Email Processor"
Cohesion: 0.22
Nodes (7): ServiceConfig, ServiceProcessorConfig, Logger, NewGenericEmailProcessor(), truncateString(), GenericEmailProcessor, Regexp

### Community 2 - "Processor Manager & Models"
Cohesion: 0.20
Nodes (8): Context, Logger, NewProcessorManager(), Email, EmailProcessor, TorrentNotification, Manager, WaitGroup

### Community 3 - "Deployment & Docs"
Cohesion: 0.18
Nodes (13): Dependabot Docker Update Config (deployments/docker), Deploy Job: Build & Push Image to Zot Registry, Deploy Job: Deploy to Production (homelab repo), Deploy Job: Calculate semver Version, Graphify Knowledge-Graph Workflow Rules, automation-hub-network Bridge Network, automation-hub Docker Compose Service, Automation Hub Project Overview (+5 more)

### Community 4 - "IMAP Email Client"
Cohesion: 0.29
Nodes (5): EmailConfig, IMAPClient, Context, Logger, NewIMAPClient()

### Community 5 - "CI Security & Quality Tooling"
Cohesion: 0.36
Nodes (8): Dependabot GitHub Actions Update Config, Dependabot Go Modules Update Config, CI Job: OWASP Dependency-Check (CVE scanner), CI Job: Lint Code (golangci-lint), CI Job: OSV Scanner (dependency vulnerabilities), CI Job: SAST (Semgrep SARIF, self-hosted), CI Job: Tests & Build, Security & Quality Tooling List (README claims)

### Community 6 - "Telegram Notifications"
Cohesion: 0.36
Nodes (5): BotAPI, Logger, NewClient(), parseInt64(), Client

### Community 8 - "Model Tests"
Cohesion: 0.67
Nodes (3): TestEmail(), TestTorrentNotification(), T

## Ambiguous Edges - Review These
- `CI Job: OSV Scanner (dependency vulnerabilities)` → `Security & Quality Tooling List (README claims)`  [AMBIGUOUS]
  README.md · relation: conceptually_related_to
- `CI Job: OWASP Dependency-Check (CVE scanner)` → `Security & Quality Tooling List (README claims)`  [AMBIGUOUS]
  README.md · relation: conceptually_related_to

## Knowledge Gaps
- **5 isolated node(s):** `automation-hub`, `Dependabot Go Modules Update Config`, `Dependabot Docker Update Config (deployments/docker)`, `Docker Compose Deployment Instructions`, `automation-hub-network Bridge Network`
  These have ≤1 connection - possible missing edges or undocumented components.
- **2 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **What is the exact relationship between `CI Job: OSV Scanner (dependency vulnerabilities)` and `Security & Quality Tooling List (README claims)`?**
  _Edge tagged AMBIGUOUS (relation: conceptually_related_to) - confidence is low._
- **What is the exact relationship between `CI Job: OWASP Dependency-Check (CVE scanner)` and `Security & Quality Tooling List (README claims)`?**
  _Edge tagged AMBIGUOUS (relation: conceptually_related_to) - confidence is low._
- **Why does `Client` connect `Telegram Notifications` to `Torrent Webhook Handling`, `Generic Email Processor`, `Processor Manager & Models`, `IMAP Email Client`, `Email Message Parsing`?**
  _High betweenness centrality (0.303) - this node is a cross-community bridge._
- **Why does `GenericEmailProcessor` connect `Generic Email Processor` to `Processor Manager & Models`, `Telegram Notifications`?**
  _High betweenness centrality (0.131) - this node is a cross-community bridge._
- **Why does `IMAPClient` connect `IMAP Email Client` to `Telegram Notifications`, `Email Message Parsing`?**
  _High betweenness centrality (0.097) - this node is a cross-community bridge._
- **What connects `automation-hub`, `Dependabot Go Modules Update Config`, `Dependabot Docker Update Config (deployments/docker)` to the rest of the system?**
  _5 weakly-connected nodes found - possible documentation gaps or missing edges._