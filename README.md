# 🤖 Automation Hub

![Go Version](https://img.shields.io/badge/Go-1.27.2-00ADD8?logo=go)
![Docker](https://img.shields.io/badge/Docker-Required-2496ED?logo=docker&logoColor=white)
![Platform](https://img.shields.io/badge/Platform-ARM64%20%7C%20AMD64-green)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=devidence-dev_automation-hub&metric=alert_status&token=<SONARCLOUD_BADGE_TOKEN>)](https://sonarcloud.io/summary/new_code?id=devidence-dev_automation-hub)
[![Reliability Rating](https://sonarcloud.io/api/project_badges/measure?project=devidence-dev_automation-hub&metric=reliability_rating&token=<SONARCLOUD_BADGE_TOKEN>)](https://sonarcloud.io/summary/new_code?id=devidence-dev_automation-hub)
[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=devidence-dev_automation-hub&metric=security_rating&token=<SONARCLOUD_BADGE_TOKEN>)](https://sonarcloud.io/summary/new_code?id=devidence-dev_automation-hub)
[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=devidence-dev_automation-hub&metric=sqale_rating&token=<SONARCLOUD_BADGE_TOKEN>)](https://sonarcloud.io/summary/new_code?id=devidence-dev_automation-hub)

> **A powerful, configurable automation hub for monitoring emails and handling webhooks with Telegram notifications**

## 🛡️ Security & Quality

This project implements comprehensive **security and code quality** measures:

- 🔍 **Static Code Analysis** - golangci-lint (errcheck, govet, staticcheck, unused, misspell, gosec) in CI; gocognit in the local gate
- 🔒 **Vulnerability Scanning** - govulncheck in CI
- 📊 **SonarCloud** - coverage from `go test`, with a blocking Quality Gate
- 🧬 **Generated Assets Check** - CI regenerates the templ and Tailwind assets and fails on any drift
- 🤖 **Automated Dependency Updates** - Dependabot for Go modules, the Docker base image and GitHub Actions
- 🚀 **CI/CD Pipeline** - `ci.yml` (shared Go CI) and `deploy.yml` (build, push and deploy to k3s)
- ✅ **Local gate** - `scripts/verify.sh` runs the same checks in Docker before you push

## ✨ Features

- 📧 **Real-time email monitoring** - IMAP-based email processing with configurable polling
- 🔧 **Dynamic service configuration** - Add new email processors without code changes
- 🤖 **Telegram notifications** - Organized notifications with custom formatting per service
- 🔗 **Configurable webhook support** - Handles qBittorrent and other webhook integrations with custom messages
- ▶️ **GitHub Actions commands** - Authorized Telegram chats can dispatch configured workflows
- 🏗️ **Modular architecture** - Clean, extensible, and maintainable codebase
- 🚀 **Container ready** - Alpine-based image that runs as uid 1000, deployed to k3s

## 🎯 Quick Start

### 📋 Prerequisites

| Component | Version | Notes |
|-----------|---------|--------|
| **Go** | 1.27.2 (`.go-version`) | For local development and tests |
| **Docker** | Current release | Runs the local gate and the asset generator, and builds the image |

**Build targets:** ARM64 (CI builds the `arm64` binary) and AMD64 (the Dockerfile follows `TARGETARCH`)

---

## ⚙️ Configuration

### 🔐 Step 1: Setup Configuration

```bash
# Copy example configuration
cp configs/config.yaml.example configs/config.yaml
```

### 📝 Step 2: Configure Services

The new **dynamic service system** allows you to add email processors through configuration only:

```yaml
server:
  address: ":8080"

email:
  host: "imap.gmail.com"
  port: 993
  username: "your-email@gmail.com"
  password: "your-app-password"
  polling_interval: 30  # Seconds between email checks
  
  # 🚀 Dynamic service configuration - Add any email processor here!
  services:
    - name: "cloudflare"
      config:
        email_from: "noreply@notify.cloudflare.com"
        email_subject:
          - "your-domain.com"
        telegram_chat_id: "YOUR_CLOUDFLARE_CHAT_ID"
        telegram_message: "🛡️ Cloudflare Code: ```%s```"
        code_pattern: "\\b\\d{6}\\b"  # the 6-digit code
        strip_mime_headers: false  # the code is in the header block
        
    - name: "perplexity"
      config:
        email_from: "team@mail.perplexity.ai"
        email_subject:
          - "Sign in to Perplexity"
        telegram_chat_id: "YOUR_PERPLEXITY_CHAT_ID"
        telegram_message: "🔮 Perplexity Code: ```%s```"
        extractor: marker  # the code follows one of the markers
        code_markers: ["directly:", "directamente:"]
        code_candidates: ["\\b(\\d{5,6})\\b", "\\b([a-zA-Z0-9]+-[a-zA-Z0-9]+)\\b"]
        code_pattern: "(?:\\d{5,6}|[a-zA-Z0-9]+-[a-zA-Z0-9]+)"
        
    # ➕ Add more services here without touching code!
    # Each service's extraction comes from its keys; see "Email extraction settings" below.

# 🔗 Webhook Configuration
hook:
  - name: "qbittorrent"
    path: "/webhook/qbittorrent"
    config:
      telegram_chat_id: "YOUR_QBITTORRENT_CHAT_ID"
      telegram_message: "📥 **Download completed successfully!** 🎬 \n🔍 **Name:**  \n%s\n📍 **Path:**  \n%s"

telegram:
  bot_token: "YOUR_BOT_TOKEN"
```

### 🤖 Step 3: Setup Telegram Bot

1. **Create a bot**: Message [@BotFather](https://t.me/botfather) → `/newbot`
2. **Get bot token**: Save the token from BotFather
3. **Get chat IDs**: 
   - Add your bot to the desired chats
   - Send a message, then visit: `https://api.telegram.org/bot<TOKEN>/getUpdates`
   - Find the `chat.id` values

### ▶️ Trigger GitHub Actions from Telegram

Configure `workflow_bot.bot_token` with the existing bot token used by `github-runner` to notify workflow results. This is intentionally distinct from the notification bot in `telegram`, which automation-hub uses for IMAP and qBittorrent notifications. The hub uses Telegram long polling, so it requires no public endpoint or ingress. Configure a fine-grained GitHub PAT with **Actions: Read and write** access limited to the target repository, then provide it as `AUTOMATION_GITHUB_TOKEN` (or set `github.token` in the configuration).

The example config registers these commands:

- `/run_cleanup` dispatches `ops-github-runner-cleanup.yml` in `devidence-dev/workflows`.

An entry can also carry an optional `schedule`, a 5-field cron expression (minute, hour, day of month, month, day of week) evaluated in the top-level `timezone` (default `America/Guayaquil`). For example, `schedule: "0 9 * * 1,3,5"` dispatches the workflow every Monday, Wednesday and Friday at 09:00. Scheduled dispatches use the same path as the Telegram command, and the log records the run URL GitHub returns. If a scheduled dispatch fails, every chat in that entry's `allowed_chat_ids` gets a failure message. The timezone is embedded in the binary, so it does not depend on the container image. An invalid `schedule` stops startup. Entries without `schedule` run only from Telegram.

Each command has its own `allowed_chat_ids`; requests from other chats are rejected before GitHub is called. To add another ad-hoc workflow, add an entry under `workflows` with a unique lowercase `command` (letters, digits, underscores; 1–32 characters), its repository details, branch (`ref`), and allowlist. Restarting the hub synchronizes the Telegram command menu and restarts long polling. Only one process may consume updates for this bot token; a second `getUpdates` consumer or webhook would cause Telegram to return a `409 Conflict`.

---

## 🐳 Deployment

### 🚀 Production: k3s through GitHub Actions

Production runs on a single-node k3s cluster. `.github/workflows/deploy.yml` is started by hand (`workflow_dispatch`) with `bump` (`patch`, `minor`, `major` or `rebuild`), an optional `custom_tag` and `prerelease`:

1. `deploy.yml` reads `.go-version` into the `GO_VERSION` build argument and calls `devidence-dev/workflows/.github/workflows/cd-build-deploy-semver.yml@v1`.
2. The shared workflow calculates the SemVer tag, builds `deployments/docker/Dockerfile` and pushes `ghcr.io/devidence-dev/automation-hub` to GHCR.
3. Its `deploy` job runs under `environment: production` and bumps the tag and digest in `kubernetes/internal-apps/automation-hub/values.yaml` in the homelab repo. ArgoCD then syncs the Deployment.

The Dockerfile is multi-stage: it builds a static binary with `CGO_ENABLED=0`, then copies it into `alpine` as uid 1000 (`appuser`). The container listens on port 8080.

**Configuration.** The hub reads one `config.yaml`. In the cluster, the `CONFIG_YAML` secret comes from Infisical, and an ExternalSecret (`automation-hub-config`, in the homelab repo) renders it into `/app/config.yaml`. `configs/config.yaml.example` documents every key.

### 💻 Run locally

Use the Go version in `.go-version` (see Prerequisites). From the repository root:

```bash
# The hub reads config.yaml from the first of /app, ./configs, /app/configs or . that has one
cp configs/config.yaml.example configs/config.yaml   # then fill in your values

# Run the application
go run cmd/automation-hub/main.go

# Run tests
go test ./...
```

Without Go on the host, run the tests in the image that matches `.go-version`:

```bash
docker run --rm -v "$PWD":/app -w /app golang:1.27.2 go test ./...
```

`configs/config.yaml` is tracked in this repository, so keep real secrets out of it.

### 🗂️ Run Kubernetes Jobs from Telegram

A `jobs` entry creates a one-off Job from a suspended CronJob template, like `kubectl create job --from=cronjob/<name>`. The Job is named `<cronjob>-manual-<unix seconds>` and copies the template as-is (labels, annotations, spec); the hub does not inject any environment into it. Command names share one namespace with `workflows` and `restarts`, and `schedule` works as it does for workflows. The hub needs an in-cluster ServiceAccount allowed to `get` the CronJob and `create` Jobs in its namespace (RBAC lives in the homelab repo). Example:

```yaml
jobs:
  - command: "run_dependency_report"
    description: "Run the dependency report (Renovate lookup)"
    namespace: "internal-apps"
    cronjob: "dependency-report"
    allowed_chat_ids: ["123456789"]
    schedule: "0 9 * * 1"
```

### ⚙️ Bot settings

All bot values are optional. The defaults keep the current behavior:

```yaml
github:
  web_url: "https://github.com"   # base of the run link when GitHub returns no run details
restart_ready:
  timeout: "3m"                   # how long a restart waits for the new pods to report ready (positive)
  poll_interval: "5s"             # how often it checks; positive and not longer than timeout
messages:                         # overrides single bot texts; keep the {placeholders} of the default
  bot:
    unknown_command: "❔ Comando no existe."
```

The bot texts are the `bot.*` keys of `internal/messages/default.yaml`. A `messages:` override that drops or adds a `{placeholder}` fails config loading and names the key.

### 🌐 HTTP settings

All HTTP keys are optional. The defaults keep the current behavior; the commented block in `configs/config.yaml.example` lists every key.

```yaml
server:
  read_timeout: "15s"        # webhook request read limit (positive)
  write_timeout: "15s"       # webhook response write limit (positive)
  shutdown_timeout: "30s"    # graceful stop on SIGINT/SIGTERM (positive)
github:
  api_url: "https://api.github.com"  # https only; trailing slash ignored; GHES uses its /api/v3 base
  http:
    timeout: "30s"           # whole request; the other transport keys below are also positive
    dial_timeout: "30s"
    tls_handshake_timeout: "10s"
    keep_alive: "30s"
    max_idle_conns: 100
    idle_conn_timeout: "90s"
    max_response_bytes: 1048576  # 1 MiB cap on every response body read
kubernetes:
  http:                      # same keys as github.http; timeout defaults to 15s, dial_timeout to 10s
    max_response_bytes: 1048576
telegram:
  http:                      # same transport keys; timeout defaults to poll_timeout + 10s
    poll_timeout: "30s"      # long-poll window sent to Telegram
    response_header_timeout: "35s"  # defaults to poll_timeout + 5s; must exceed poll_timeout
    expect_continue_timeout: "1s"
    max_attempts: 3          # sends per message, minimum 1; only network errors, 5xx and 429 are retried
    retry_backoff: "1s"      # doubled after each failed attempt
    max_retry_after: "30s"   # a 429 asking for a longer wait fails at once
```

Telegram's timeout and response_header_timeout must each be greater than poll_timeout. A value that is zero, negative or unparseable stops startup and names the key.

### 📦 Dependency report webhook

When `dependency_report:` is configured, the hub exposes `POST` on `dependency_report.path` (default `/webhook/dependency-report`). The Job's reporter container posts a `multipart/form-data` body (`exit_code`, `pod`, file `log_tail`, optional file `report`) with the token in the header `dependency_report.token_header` (default `X-Report-Token`). The hub answers `202` and sends one Telegram message to `telegram_chat_id`, through the notifications bot (`telegram.bot_token`): the HTML document `dependency-report-<YYYY-MM-DD>.html`, named from `dependency_report.document_name` with the day the hub received the report in the top-level `timezone` (default `America/Guayaquil`), formatted with `dependency_report.date_layout`, with a short Spanish summary as its caption. Its header shows the scan time in the same zone. The caption lists the repos with pending updates (🔴 when one has a major update), with the count per update type, then how many repos are up to date, the totals of the two tabs and the scan errors. Telegram caps captions at 1024 characters (the hub also keeps messages under 4096), so when the list is too long the hub drops repo lines from the end and adds `… y N repos más (ver reporte)`. If Telegram rejects the document, the caption is sent as a text message instead. A wrong token gets `401`, a malformed request `400`, and a body over `dependency_report.max_body_bytes` (default 10 MiB) gets `413`. A Renovate failure (`exit_code` not `0`, or no `report`) sends only the failure message, never a document.

The document has two tabs:

- **Imágenes** lists dependencies whose Renovate `datasource` is `docker` (container images), with the current and latest image and a copy button for the latest image reference.
- **Dependencias** lists every other dependency, including those without a `datasource`, with the current and new version and a copy button for `name@new version`.

Each tab shows its totals and a count per update type (`patch`, `minor`, `major`). Only those three update types are listed (`digest` and `pin` are not), and a fork in `dependency_report.forks` is tagged in its rows. Problems with level 50 or more appear in an Errors section at the bottom (report-level problems as `raíz: <msg>`). The document is one self-contained HTML file: inline Tailwind CSS, one copy script, and the Geist fonts from Google Fonts as its only external request (the system font is the fallback). Its tabs and repo sections need no JavaScript for switching. It is built from the templ components in `internal/services/depreport` (`*.templ`, generated `*_templ.go`), so report text is always escaped. Its `lang` and every label are Spanish by default (`lang="es"`), from the message catalog.

Caption and document texts come from one catalog: the embedded defaults are in `internal/messages/default.yaml`. Override single keys under `messages:` with the same nested keys; an unknown key, a map where a text is expected, or a changed placeholder (such as `{date}` or `{count}`) stops startup. The two fork tags are catalog keys too: `fork.tag_caption` (default ` (fork)`) for caption lines and `fork.tag_document` (default ` (fork, solo informativo)`) for document rows. Protocol limits (1024 characters for captions, 4096 for messages) are not configurable.

```yaml
timezone: "America/Guayaquil"                    # schedules, document date and header

dependency_report:
  webhook_token: "${DEPENDENCY_REPORT_TOKEN}"   # Infisical reference inside CONFIG_YAML
  telegram_chat_id: "123456789"
  # Do not quote this reference. It holds a JSON array, so unquoted YAML reads a
  # list. A quoted reference is a string and config loading fails.
  forks: ${DEPENDENCY_REPORT_FORKS}              # e.g. ["devidence-dev/discord-tts-bot"]
  ignore:                                        # optional: package files left out of the report
    - repo: "devidence-dev/homelab"              # a path is a file or a folder, relative to the repo root
      paths: ["kubernetes/apps/disabled-app"]
  path: "/webhook/dependency-report"             # route; must start with /
  token_header: "X-Report-Token"                 # header carrying webhook_token
  max_body_bytes: 10485760                       # 10 MiB, must be positive
  document_name: "dependency-report-{date}.html" # must contain {date}, no / or \
  date_layout: "2006-01-02"                      # Go time layout of {date}

messages:                                        # optional: override single texts
  caption:
    header: "🔔 <b>Deps</b> · {date}"
```

Changing `timezone` moves the schedules and the report date. Without a `messages:` section, the texts are the defaults.

---

### 🧱 Report assets and the local gate

The report components (`*.templ`, generated into `*_templ.go`) and the compiled stylesheet `internal/services/depreport/assets/depreport.min.css` are committed generated files. After changing a `.templ` file or `assets/input.css`, regenerate them:

```sh
scripts/generate-depreport-assets.sh
```

The script runs in an ephemeral `golang:<.go-version>` container, so nothing is installed on the host. It downloads templ v0.3.1070 and Tailwind CSS v4.3.3 for the container architecture, checks each one against its release's checksum file, and stops on a mismatch. A second run changes nothing.

Before pushing, run the local gate. It needs only Docker and runs the CI checks in containers:

```sh
scripts/verify.sh          # gofmt, golangci-lint, go test -race with 80% coverage, gocognit, asset drift, biome ci, openspec validate
scripts/verify.sh --fix    # first runs gofmt -w, biome check --write and the asset generation
```

Pinned tools: golangci-lint v2.14.0, gocognit v1.2.0, Biome 2.5.15 (standalone binary, checked by SHA-256 before it runs), OpenSpec 1.14.1 and the two generation tools above. CI runs the same drift check in the `generated-assets` job of this repo's `.github/workflows/ci.yml`: it regenerates the assets and fails on a diff or on an untracked generated file.

## 🔌 API & Webhooks

### 📡 Available Endpoints

| Endpoint | Method | Description |
|----------|---------|-------------|
| `/webhook/qbittorrent` | POST | qBittorrent completion notifications (the `path` of the `qbittorrent` hook; the example config uses this path) |
| `/webhook/dependency-report` | POST | Renovate report from the dependency-report Job (only if `dependency_report:` is set; the default, changed with `dependency_report.path`) |

### 📦 qBittorrent Integration

Configure qBittorrent to send webhooks on completion. The webhook supports **custom messages** configured in your `config.yaml`:

**Tools** → **Options** → **Downloads** → **Run external program on torrent completion:**

```bash
curl -X POST http://your-server:8080/webhook/qbittorrent \
  -H "Content-Type: application/json" \
  -d '{"torrent_name":"%N","save_path":"%F"}'
```

**📝 Custom Message Configuration:**

You can customize the notification message in your `config.yaml`:

```yaml
hook:
  - name: "qbittorrent"
    path: "/webhook/qbittorrent"
    config:
      telegram_chat_id: "YOUR_CHAT_ID"
      telegram_message: "🎬 Your custom message! \n📁 File: %s\n📂 Location: %s"
```

The `%s` placeholders will be replaced with:
1. First `%s` → Torrent name
2. Second `%s` → Save path

**📏 Body size cap:** the request body is read up to `max_body_bytes` (under the hook's `config`, default `1048576`, 1 MiB). A larger body gets `413` and no Telegram message is sent. A negative value stops startup.

```yaml
hook:
  - name: "qbittorrent"
    config:
      telegram_chat_id: "YOUR_CHAT_ID"
      max_body_bytes: 1048576  # optional
```

### 🆕 Webhook Types

Each `hook` entry is a route: its `name` selects the handler and its `path` is the URL. The handler for each name is registered in `internal/handlers/webhook_registry.go`. Today the registry has `qbittorrent` and the dependency report (`dependency-report`, configured only in its own `dependency_report:` section). An entry with any other name is logged as an unknown webhook type and skipped, so a new type needs code in the registry. The `qbittorrent` hook is configured like this:

```yaml
hook:
  - name: "qbittorrent"
    path: "/webhook/qbittorrent"
    config:
      telegram_chat_id: "TORRENT_CHAT_ID"
      telegram_message: "📥 Download complete: %s in %s"
```

Each `qbittorrent` hook can have:
- ✅ **Custom chat destination**
- ✅ **Personalized message format**
- ✅ **Two parameter placeholders**

### 🔄 Adding New Email Services

The **magic** ✨ of this system is that you can add new email processors without writing any code:

```yaml
# Just add to your config.yaml:
email:
  services:
    - name: "github"
      config:
        email_from: "noreply@github.com"
        email_subject: ["verification code"]
        telegram_chat_id: "YOUR_CHAT_ID"
        telegram_message: "🐙 GitHub Code: ```%s```"
        code_pattern: "\\b\\d{6}\\b"  # Custom regex for 6-digit codes
        mark_as_read: true  # Optional: mark the email as read after processing (default: true)
```

**That's it!** The system will automatically:
- ✅ Monitor emails from the specified sender
- ✅ Match subject patterns
- ✅ Extract codes using the pattern
- ✅ Send formatted Telegram notifications
- ✅ Mark the email as read (unless `mark_as_read: false`), so it isn't re-forwarded on the next poll

### 🔎 Email extraction settings

Each service's behavior comes from its keys, not from its name. All keys are optional:

| Key | Default | Meaning |
|-----|---------|---------|
| `extractor` | `pattern` | `pattern` takes the first match of `code_pattern`; `marker` finds the code after one of `code_markers` |
| `code_markers` | none | Phrases that precede the code, tried in order and case-insensitively (required by `marker`) |
| `code_candidates` | none | Regexes tried on the text after the marker, in order; the first capture group is the code |
| `code_pattern` | `\b[a-zA-Z0-9]{4,8}\b` | The code pattern; with `marker` it is the fallback after the candidates, and absent means no fallback |
| `strip_mime_headers` | `true` | Removes the header block before extraction; set `false` when the code is in the headers |
| `not_found_text` | `Not found` | Sent in place of the code when none is found |
| `log_preview_length` (under `email:`) | `200` | Characters of the searched text shown in not-found warnings; zero or absent means `200`, a negative value stops startup |
| `polling_interval` (under `email:`) | `60` | Seconds between mailbox polls |

`email_from` is the one required key: every service needs a non-empty value, because an empty one would match every sender. An empty or missing `email_from`, an invalid regex, an unknown `extractor`, or `marker` without `code_markers` stops startup, and the error names the service.

---

## 🔧 External Service Setup

### 📧 Gmail Configuration

1. **Enable 2FA**: Go to Google Account Security
2. **Generate App Password**: 
   - Security → 2-Step Verification → App passwords
   - Select "Mail" and generate password
3. **Use App Password**: Use the generated password in `config.yaml`

### 🏴‍☠️ qBittorrent Setup

1. **Tools** → **Options** → **Downloads**
2. **Run external program on torrent completion:**
   ```bash
   curl -X POST http://localhost:8080/webhook/qbittorrent \
     -H "Content-Type: application/json" \
     -d '{"torrent_name":"%N","save_path":"%F"}'
   ```

**📝 Message Customization:**
- Edit `telegram_message` in your `config.yaml` 
- Use `%s` placeholders for torrent name and save path
- Supports Markdown formatting for rich notifications
- All example messages have been translated to English; customize freely.

---

## 📊 Monitoring & Logs

### 🔍 Viewing Logs

```bash
# Production (k3s): the Deployment in the internal-apps namespace
kubectl -n internal-apps logs deploy/automation-hub -f

# Local development
# Logs output to stdout with structured JSON format
```

### 📁 Log Locations

- **Production**: stdout of the pod, JSON format for easy parsing and monitoring
- **Local**: Standard output with structured logging (JSON, the same logger)

---

## 🛠️ Troubleshooting

### 🚨 Common Issues

<details>
<summary><strong>📧 IMAP Connection Failed</strong></summary>

```yaml
# Check your config.yaml:
email:
  host: "imap.gmail.com"  # Correct IMAP server
  port: 993               # Correct port (usually 993 for SSL)
  username: "your-email@gmail.com"
  password: "app-password"  # Use app password, not regular password!
```

**Solutions:**
- ✅ Use app password for Gmail (not your regular password)
- ✅ Enable 2FA and generate app password
- ✅ Check firewall settings
- ✅ Verify IMAP is enabled in email provider
</details>

<details>
<summary><strong>🤖 Telegram Notifications Not Working</strong></summary>

**Check the basics:**
- ✅ Bot token is correct
- ✅ Chat IDs are correct (including negative sign for groups)
- ✅ Bot has permission to send messages
- ✅ Bot is added to the target chat/group

**Get chat ID:**
```bash
# Send a message to your bot, then:
curl https://api.telegram.org/bot<BOT_TOKEN>/getUpdates
```
</details>

<details>
<summary><strong>🔗 Webhooks Not Received</strong></summary>

**Debugging steps:**
- ✅ Check if port 8080 is accessible
- ✅ Verify JSON payload format matches expected: `{"torrent_name":"...", "save_path":"..."}`
- ✅ Verify webhook configuration exists in `config.yaml`
- ✅ Check application logs for errors
- ✅ Test with curl manually:
  ```bash
  curl -X POST http://localhost:8080/webhook/qbittorrent \
    -H "Content-Type: application/json" \
    -d '{"torrent_name":"Test Movie","save_path":"/downloads/movies/"}'
  ```
</details>

<details>
<summary><strong>🎨 Webhook Message Not Formatting</strong></summary>

**Check your config:**
```yaml
hook:
  - name: "qbittorrent"
    path: "/webhook/qbittorrent"
    config:
      telegram_chat_id: "123456789"  # Correct chat ID
      telegram_message: "Download: %s in %s"  # Two %s placeholders
```

**Common issues:**
- ✅ Ensure you have exactly 2 `%s` placeholders for qBittorrent
- ✅ Check that the webhook name matches exactly (`"qbittorrent"`)
- ✅ Verify chat ID is correct (including negative sign for groups)
</details>


---

## Authors and acknowledgment 🛡

PX1 - devidence.dev ©

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

<div align="center">

**Made with ❤️ for automation enthusiasts**

![Stars](https://img.shields.io/github/stars/devidence-dev/automation-hub?style=social)
![Forks](https://img.shields.io/github/forks/devidence-dev/automation-hub?style=social)

</div>
