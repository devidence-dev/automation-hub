# 🤖 Automation Hub

![Go Version](https://img.shields.io/badge/Go-1.24.5-00ADD8?logo=go)
![Docker](https://img.shields.io/badge/Docker-20.10+-2496ED?logo=docker&logoColor=white)
![Platform](https://img.shields.io/badge/Platform-ARM64%20%7C%20AMD64-green)
[![Quality Gate Status](https://sonarqube.devidence.dev/api/project_badges/measure?project=automation-hub&metric=alert_status&token=sqb_7d178597a22d329b0e50efb39a6ae16d00f64db4)](https://sonarqube.devidence.dev/dashboard?id=automation-hub)
[![Reliability Rating](https://sonarqube.devidence.dev/api/project_badges/measure?project=automation-hub&metric=software_quality_reliability_rating&token=sqb_7d178597a22d329b0e50efb39a6ae16d00f64db4)](https://sonarqube.devidence.dev/dashboard?id=automation-hub)
[![Security Rating](https://sonarqube.devidence.dev/api/project_badges/measure?project=automation-hub&metric=software_quality_security_rating&token=sqb_7d178597a22d329b0e50efb39a6ae16d00f64db4)](https://sonarqube.devidence.dev/dashboard?id=automation-hub)
[![Maintainability Rating](https://sonarqube.devidence.dev/api/project_badges/measure?project=automation-hub&metric=software_quality_maintainability_rating&token=sqb_7d178597a22d329b0e50efb39a6ae16d00f64db4)](https://sonarqube.devidence.dev/dashboard?id=automation-hub)

> **A powerful, configurable automation hub for monitoring emails and handling webhooks with Telegram notifications**

## 🛡️ Security & Quality

This project implements comprehensive **security and code quality** measures:

- 🔍 **Static Code Analysis** - golangci-lint, gosec, CodeQL
- 🕵️ **Secret Detection** - TruffleHog, GitLeaks, Semgrep
- 🔒 **Vulnerability Scanning** - govulncheck, Nancy, Trivy
- 🤖 **Automated Dependency Updates** - Dependabot
- 📊 **Code Coverage** - Codecov integration
- 🚀 **CI/CD Pipeline** - Automated testing and security checks

## ✨ Features

- 📧 **Real-time email monitoring** - IMAP-based email processing with configurable polling
- 🔧 **Dynamic service configuration** - Add new email processors without code changes
- 🤖 **Telegram notifications** - Organized notifications with custom formatting per service
- 🔗 **Configurable webhook support** - Handles qBittorrent and other webhook integrations with custom messages
- 🏗️ **Modular architecture** - Clean, extensible, and maintainable codebase
- 🚀 **Docker ready** - Optimized for Raspberry Pi 5 and cloud deployment

## 🎯 Quick Start

### 📋 Prerequisites

| Component | Version | Notes |
|-----------|---------|--------|
| **Go** | 1.24.5+ | For local development |
| **Docker** | 20.10+ | Required for deployment |
| **Docker Compose** | 2.0+ | Orchestration |

**Supported Platforms:** ARM64 (Raspberry Pi 5), AMD64

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
        # code_pattern: "\\b\\d{6}\\b"  # Optional: custom regex
        
    - name: "perplexity"
      config:
        email_from: "team@mail.perplexity.ai"
        email_subject:
          - "Sign in to Perplexity"
        telegram_chat_id: "YOUR_PERPLEXITY_CHAT_ID"
        telegram_message: "🔮 Perplexity Code: ```%s```"
        
    # ➕ Add more services here without touching code!

# 🔗 Webhook Configuration
hook:
  - name: "qbittorrent"
    config:
      telegram_chat_id: "YOUR_QBITTORRENT_CHAT_ID"
  telegram_message: "📥 **Download completed successfully!** 🎬 \n🔍 **Name:**  \n%s\n📍 **Path:**  \n%s"
  # Add more webhooks here:
  # - name: "sonarr"
  #   config:
  #     telegram_chat_id: "YOUR_SONARR_CHAT_ID"
  #     telegram_message: "📺 Serie descargada: %s"

telegram:
  bot_token: "YOUR_BOT_TOKEN"
  chat_ids:
    torrent: "YOUR_TORRENT_CHAT_ID"
```

### 🤖 Step 3: Setup Telegram Bot

1. **Create a bot**: Message [@BotFather](https://t.me/botfather) → `/newbot`
2. **Get bot token**: Save the token from BotFather
3. **Get chat IDs**: 
   - Add your bot to the desired chats
   - Send a message, then visit: `https://api.telegram.org/bot<TOKEN>/getUpdates`
   - Find the `chat.id` values

---

## 🐳 Deployment

### 🚀 Option 1: Docker Compose (Recommended)

```bash
# Clone repository
git clone <your-repo-url>
cd automation-hub

# Configure
cp configs/config.yaml.example configs/config.yaml
# Edit configs/config.yaml with your credentials

# Deploy
docker-compose up -d

# View logs
docker-compose logs -f automation-hub
```

### 🔧 Option 2: Manual Docker

```bash
# Build for your platform
docker build -f deployments/docker/Dockerfile -t automation-hub .

# Run with volume mount
docker run -d \
  --name automation-hub \
  --restart unless-stopped \
  -v $(pwd)/configs/config.yaml:/root/configs/config.yaml:ro \
  automation-hub
```

### � Option 3: Local Development

```bash
# Install dependencies
go mod download

# Run application
go run cmd/automation-hub/main.go

# Run tests
go test ./...
```

---

## � API & Webhooks

### 📡 Available Endpoints

| Endpoint | Method | Description |
|----------|---------|-------------|
| `/webhook/qbitorrent` | POST | qBittorrent completion notifications |

### 📦 qBittorrent Integration

Configure qBittorrent to send webhooks on completion. The webhook now supports **custom messages** configured in your `config.yaml`:

**Tools** → **Options** → **Downloads** → **Run external program on torrent completion:**

```bash
curl -X POST http://your-server:8080/webhook/qbitorrent \
  -H "Content-Type: application/json" \
  -d '{"torrent_name":"%N","save_path":"%F"}'
```

**📝 Custom Message Configuration:**

You can customize the notification message in your `config.yaml`:

```yaml
hook:
  - name: "qbittorrent"
    config:
      telegram_chat_id: "YOUR_CHAT_ID"
  telegram_message: "🎬 Your custom message! \n📁 File: %s\n📂 Location: %s"
```

The `%s` placeholders will be replaced with:
1. First `%s` → Torrent name
2. Second `%s` → Save path

### 🆕 Adding New Webhooks

The system now supports **configurable webhooks**! Add new webhook handlers without coding:

```yaml
hook:
  - name: "qbittorrent"
    config:
      telegram_chat_id: "TORRENT_CHAT_ID"
  telegram_message: "📥 Download complete: %s in %s"
  - name: "sonarr"
    config:
      telegram_chat_id: "SONARR_CHAT_ID" 
  telegram_message: "📺 New series: %s"
  - name: "radarr"
    config:
      telegram_chat_id: "RADARR_CHAT_ID"
  telegram_message: "🎬 New movie: %s"
```

Each webhook can have:
- ✅ **Custom chat destination**
- ✅ **Personalized message format**
- ✅ **Multiple parameter placeholders**

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
```

**That's it!** The system will automatically:
- ✅ Monitor emails from the specified sender
- ✅ Match subject patterns
- ✅ Extract codes using the pattern
- ✅ Send formatted Telegram notifications

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
   curl -X POST http://localhost:8080/webhook/qbitorrent \
     -H "Content-Type: application/json" \
     -d '{"torrent_name":"%N","save_path":"%F"}'
   ```

**📝 Message Customization:**
- Edit `telegram_message` in your `config.yaml` 
- Use `%s` placeholders for torrent name and save path
- Supports Markdown formatting for rich notifications
- All example messages have been translated to English; customize freely.

---

## � Monitoring & Logs

### 🔍 Viewing Logs

```bash
# Docker Compose
docker-compose logs -f automation-hub

# Docker
docker logs -f automation-hub

# Local development
# Logs output to stdout with structured JSON format
```

### 📁 Log Locations

- **Docker**: Logs rotate automatically with size and time limits
- **Local**: Standard output with structured logging
- **Production**: JSON format for easy parsing and monitoring

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
  curl -X POST http://localhost:8080/webhook/qbitorrent \
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