---
name: claw-mail
description: "邮件管理服务 CLI，支持邮件账户管理、发送、接收功能"
metadata:
  {
    "openclaw": {
      "emoji": "📧",
      "requires": { "bins": ["claw-pliers"] }
    }
  }
---

# Mail Service Skill

邮件管理服务，提供邮件账户管理、发送、接收功能。

## 安装

```bash
go build -o claw-pliers ./cli/
```

## 配置

配置邮件账户:
```bash
claw-pliers-cli mail config --endpoint http://localhost:8080 --key <local-key>
```

## 命令

### 发送邮件
```bash
claw-pliers-cli mail send --to example@example.com --subject "Hello" --body "Content"
```

### 列出账户
```bash
claw-pliers-cli mail account list
```

### 添加账户
```bash
claw-pliers-cli mail account add --provider 163 --email xxx@163.com --auth-token <token>
```

## API 端点

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/v1/mail/accounts | 账户列表 |
| POST | /api/v1/mail/accounts | 添加账户 |
| DELETE | /api/v1/mail/accounts/:email | 删除账户 |
| POST | /api/v1/mail/send | 发送邮件 |
| GET | /api/v1/mail/latest | 最新邮件 |

## 认证

所有 API 请求需要通过 `X-Local-Key` 头进行认证。
