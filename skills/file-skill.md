---
name: claw-file
description: "文件管理服务 CLI，支持上传、下载、列表、分享、删除文件"
metadata:
  {
    "openclaw": {
      "emoji": "📦",
      "requires": { "bins": ["claw-pliers"] }
    }
  }
---

# File Service Skill

文件管理服务，提供文件上传、下载、列表、分享、删除功能。

## 安装

```bash
# 方式1: 使用安装脚本
curl -fsSL https://raw.githubusercontent.com/kiry163/claw-pliers/main/scripts/install.sh | bash

# 方式2: 手动下载
go build -o claw-pliers ./cli/
```

## 配置

初始化配置:
```bash
claw-pliers-cli file config --endpoint http://localhost:8080 --key <local-key>
```

## 命令

### 上传文件
```bash
claw-pliers-cli file upload ./myfile.zip --endpoint http://localhost:8080 --key <local-key>
```

### 列出文件
```bash
claw-pliers-cli file list --endpoint http://localhost:8080 --key <local-key>
```

### 下载文件
```bash
claw-pliers-cli file download <file-id> --output ./downloads/ --endpoint http://localhost:8080 --key <local-key>
```

### 删除文件
```bash
claw-pliers-cli file delete <file-id> --endpoint http://localhost:8080 --key <local-key>
```

## API 端点

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | /api/v1/files | 上传文件 |
| GET | /api/v1/files | 文件列表 |
| GET | /api/v1/files/:id | 文件信息 |
| GET | /api/v1/files/:id/download | 下载文件 |
| DELETE | /api/v1/files/:id | 删除文件 |

## 认证

所有 API 请求需要通过 `X-Local-Key` 头进行认证:
```bash
curl -H "X-Local-Key: <your-key>" http://localhost:8080/api/v1/files
```
