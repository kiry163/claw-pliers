---
name: claw-mail-monitor
description: "Manage claw-mail-monitor service and accounts via CLI. Send and query emails."
---

# Claw Mail Monitor Skill

Use `claw-mail-monitor` CLI to manage the service, accounts, and email actions.

## ⚡️ Priority: Use CLI

Always prefer CLI over HTTP API when using OpenClaw skill.

## 🚀 Quick Start

Install and auto-start the service:

```bash
curl -fsSL https://raw.githubusercontent.com/kiry163/claw-mail-monitor/master/install-release.sh | bash
```

Check status before use:

```bash
claw-mail-monitor service status
```

Default install runs in background. Use `status` to confirm and view config/log paths.

Set a specific version:

```bash
VERSION=v0.1.0 bash -c "$(curl -fsSL https://raw.githubusercontent.com/kiry163/claw-mail-monitor/master/install-release.sh)"
```

---

## 🛠️ Service Management

```bash
# Foreground
claw-mail-monitor serve --listen 127.0.0.1:14630 --poll-interval 30s

# Install system service
sudo ./install.sh

# Service control
claw-mail-monitor service install
claw-mail-monitor service start
claw-mail-monitor service status
claw-mail-monitor service stop
claw-mail-monitor service restart
claw-mail-monitor service uninstall
```

---

## 🔧 Account Management

```bash
claw-mail-monitor accounts list

claw-mail-monitor accounts add --provider 163 --email your@163.com --auth-token XXXXXXXX
claw-mail-monitor accounts add --provider qq --email your@qq.com --auth-token your-auth-token

claw-mail-monitor accounts remove --email your@163.com
```

---

## 📧 Query Emails

```bash
claw-mail-monitor latest
claw-mail-monitor latest --count 5
claw-mail-monitor latest --email kiry_zen@163.com --count 1
claw-mail-monitor latest --count 3
claw-mail-monitor latest --since 1m --count 1
```

---

## ✉️ Send Email

```bash
claw-mail-monitor send --to to@example.com --subject "Hello" --body "Hi there"
```

---

## 📊 Status

```bash
claw-mail-monitor status
claw-mail-monitor test-connection --email your@163.com

claw-mail-monitor version
```

---

## 🔌 Supported Providers

| Provider | Value | IMAP Host | SMTP Host |
|----------|-------|-----------|-----------|
| 163/网易邮箱 | `163` | imap.163.com:993 | smtp.163.com:465 |
| QQ邮箱 | `qq` | imap.qq.com:993 | smtp.qq.com:465 |
| Gmail | `gmail` | imap.gmail.com:993 | smtp.gmail.com:465 |

## 🔐 Auth Token Notes

- **163邮箱**：使用登录密码或授权码
- **QQ邮箱**：需要开启 IMAP/SMTP 并获取授权码
- **Gmail**：需要使用 App Password

---

## 📝 Notes

- Service runs at `http://127.0.0.1:14630`
- CLI commands don't need `--base-url` or `--config`
- New emails trigger webhook notifications to Feishu automatically
- Output is JSON format, use `jq` for filtering if needed
