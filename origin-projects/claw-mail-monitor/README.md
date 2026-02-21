# Claw Mail Monitor

Multi-account email monitoring via IMAP with HTTP and CLI integration.

## Features

- 📧 Monitor multiple email accounts (QQ, 163/NetEase, Gmail)
- 🔔 Real-time email notifications via webhook to OpenClaw
- 📱 Feishu notifications
- 🔄 Smart fallback: IMAP IDLE → Polling mode
- 🌐 HTTP API and CLI tooling

## Quick Start

### Configuration

Edit `~/.config/claw-mail-monitor/config.yaml`:

```yaml
accounts:
  - provider: 163
    email: your-email@163.com
    auth_token: your-auth-token
    enabled: true

webhook:
  url: http://127.0.0.1:18789/hooks/agent
  token: your-webhook-token
  to: ou_your-feishu-openid
  session_key: ""
  custom_payload: "{\"text\": \"📧 收到新邮件通知\", \"mode\": \"now\"}"
  enable: true

logging:
  level: info
  file: ~/.cache/claw-mail-monitor/mail-monitor.log

monitoring:
  poll_interval: 30s
```

### Installation

```bash
cd claw-mail-monitor
go build -o bin/claw-mail-monitor ./cmd/claw-mail-monitor/
sudo cp bin/claw-mail-monitor /usr/local/bin/
```

### Install from Release

Download a prebuilt binary:

```bash
curl -fsSL https://github.com/kiry163/claw-mail-monitor/releases/latest/download/claw-mail-monitor_darwin_arm64 \
  -o /usr/local/bin/claw-mail-monitor
chmod +x /usr/local/bin/claw-mail-monitor
```

Or use the install script (auto-detect OS/arch):

```bash
curl -fsSL https://raw.githubusercontent.com/kiry163/claw-mail-monitor/master/install-release.sh | bash
```

Set a specific version:

```bash
VERSION=v0.1.0 curl -fsSL https://raw.githubusercontent.com/kiry163/claw-mail-monitor/master/install-release.sh | bash
```

### Run

```bash
claw-mail-monitor serve --listen 127.0.0.1:14630 --poll-interval 30s
```

### Service Install (macOS/Linux)

```bash
sudo ./install.sh
```


### HTTP API

Default base URL: `http://127.0.0.1:14630`

- `GET /health`
- `GET /status`
- `GET /accounts`
- `POST /accounts`
- `DELETE /accounts/{email}`
- `POST /send`
- `POST /test-connection`
- `GET /latest?email=...&count=1`

#### HTTP Examples

```bash
# Health
curl http://127.0.0.1:14630/health

# Status
curl http://127.0.0.1:14630/status

# List accounts
curl http://127.0.0.1:14630/accounts

# Add account
curl -X POST http://127.0.0.1:14630/accounts \
  -H "Content-Type: application/json" \
  -d '{"provider":"163","email":"your@email.com","auth_token":"your-auth-token"}'

# Remove account
curl -X DELETE "http://127.0.0.1:14630/accounts/your%40email.com"

# Send email
curl -X POST http://127.0.0.1:14630/send \
  -H "Content-Type: application/json" \
  -d '{"to":["to@example.com"],"subject":"Hello","body":"Hi there"}'

# Test connection
curl -X POST http://127.0.0.1:14630/test-connection \
  -H "Content-Type: application/json" \
  -d '{"email":"your@email.com"}'

# Latest emails
curl "http://127.0.0.1:14630/latest?email=your%40email.com&count=1"
```

### CLI

- `claw-mail-monitor serve`
- `claw-mail-monitor status`
- `claw-mail-monitor accounts list|add|remove`
- `claw-mail-monitor send`
- `claw-mail-monitor test-connection`
- `claw-mail-monitor latest`
- `claw-mail-monitor version`

#### CLI Examples

```bash
# Start server
claw-mail-monitor serve --listen 127.0.0.1:14630 --poll-interval 30s

# Status
claw-mail-monitor status

# Status output example
{
  "config_path": "/Users/you/.config/claw-mail-monitor/config.yaml",
  "log_file": "/Users/you/.cache/claw-mail-monitor/mail-monitor.log",
  "monitoring": 1,
  "status": "running",
  "total": 1,
  "version": "0.1.0"
}

# Accounts
claw-mail-monitor accounts list
claw-mail-monitor accounts add --provider 163 --email your@email.com --auth-token your-auth-token
claw-mail-monitor accounts remove --email your@email.com

# Send mail
claw-mail-monitor send --to to@example.com --subject "Hello" --body "Hi there"

# Test connection
claw-mail-monitor test-connection --email your@email.com

# Latest
claw-mail-monitor latest --email your@email.com --count 1
claw-mail-monitor latest --count 3
claw-mail-monitor latest --since 1m --count 1

# Version
claw-mail-monitor version
```

## License

MIT
