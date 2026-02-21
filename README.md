# Claw Pliers

统一 CLI 工具，用于管理文件、邮件和图像服务。

## 功能概览

| 模块 | 状态 | 说明 |
|------|------|------|
| File | ✅ 已实现 | 文件上传、下载、列表、删除 |
| Mail | ⚠️ Stub | 邮件管理（待实现） |
| Image | ⚠️ Stub | 图像处理（待实现） |

## 架构

- **服务端**: Go + Gin + SQLite + MinIO
- **CLI**: Cobra
- **端口**: 8080
- **认证**: X-Local-Key / Bearer Token

## 快速开始

### 启动服务

```bash
# 使用 docker-compose 开发环境
docker-compose -f docker-compose.dev.yaml up -d

# 或直接运行
go run cmd/claw-pliers/main.go
```

### 认证方式

所有 API 请求需要通过以下方式认证：

1. **X-Local-Key** (推荐)
   ```bash
   curl -H "X-Local-Key: change-me-in-production" http://localhost:8080/api/v1/files
   ```

2. **Bearer Token**
   ```bash
   curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/files
   ```

---

## File 模块

### API 端点

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/files` | 上传文件 |
| GET | `/api/v1/files` | 获取文件列表 |
| GET | `/api/v1/files/:id` | 获取文件信息 |
| GET | `/api/v1/files/:id/download` | 下载文件 |
| DELETE | `/api/v1/files/:id` | 删除文件 |

### 上传文件

```bash
curl -X POST http://localhost:8080/api/v1/files \
  -H "X-Local-Key: change-me-in-production" \
  -F "file=@/path/to/file.txt"
```

### 获取文件列表

```bash
curl http://localhost:8080/api/v1/files \
  -H "X-Local-Key: change-me-in-production"
```

查询参数：
- `limit`: 返回数量（默认 50）
- `offset`: 偏移量（默认 0）
- `order`: 排序方式 (`asc` / `desc`)
- `keyword`: 关键词搜索
- `folder_id`: 文件夹 ID

### 获取文件信息

```bash
curl http://localhost:8080/api/v1/files/{file_id} \
  -H "X-Local-Key: change-me-in-production"
```

### 下载文件

```bash
curl -O http://localhost:8080/api/v1/files/{file_id}/download \
  -H "X-Local-Key: change-me-in-production"
```

### 删除文件

```bash
curl -X DELETE http://localhost:8080/api/v1/files/{file_id} \
  -H "X-Local-Key: change-me-in-production"
```

---

## Mail 模块

⚠️ **状态**: Stub 实现

当前仅提供配置结构，实际功能待实现。

### CLI 命令 (Stub)

```bash
# 发送邮件
claw-pliers mail send --to user@example.com --subject "Subject" --body "Body"

# 列出邮件账户
claw-pliers mail list
```

---

## Image 模块

⚠️ **状态**: Stub 实现

当前仅提供配置结构，实际功能待实现。

### CLI 命令 (Stub)

```bash
# 转换图像格式
claw-pliers image convert input.jpg output.png

# OCR 文字识别
claw-pliers image ocr image.png
```

---

## CLI 工具

### 安装

```bash
go build -o claw-pliers ./cli
```

### 配置说明

CLI 复用服务端的配置文件，自动加载以下位置：
- **项目目录**: `./config/config.yaml`（优先）
- **用户目录**: `~/.config/claw-pliers/config.yaml`

也可通过环境变量覆盖：
```bash
export CLAWPLIERS_ENDPOINT=http://localhost:8080
export CLAWPLIERS_AUTH_LOCAL_KEY=change-me-in-production
```

### File 命令

#### 上传文件

```bash
claw-pliers file put <file>
claw-pliers file put /path/to/file.txt

# 指定端点和密钥
claw-pliers file put /path/to/file.txt --endpoint http://localhost:8080 --key change-me-in-production
```

输出示例：
```
Uploading file.txt (1.2 MB)...
Progress: 100%
✓ Uploaded: file.txt (ID: 1771427558V8f5SDqd)
```

#### 列出文件

```bash
claw-pliers file list
claw-pliers file list --limit 10 --offset 0
```

输出示例：
```
NAME                           SIZE         TYPE                           DATE
-----------------------------------------------------------------------------------------------
document.pdf                   256.5 KB     application/pdf                2026-02-18
photo.jpg                      1.2 MB       image/jpeg                     2026-02-17

Total: 2 files
```

#### 下载文件

```bash
claw-pliers file get <file_id>
claw-pliers file get <file_id> /path/to/save

# 示例
claw-pliers file get 1771427558V8f5SDqd ./downloads/
```

输出示例：
```
Downloading 1771427558V8f5SDqd...
Progress: 100%
✓ Saved to: ./downloads/file.txt
```

#### 删除文件

```bash
claw-pliers file delete <file_id>

# 示例
claw-pliers file delete 1771427558V8f5SDqd
```

输出示例：
```
✓ Deleted: 1771427558V8f5SDqd
```

#### 查看文件详情

```bash
claw-pliers file info <file_id>

# 示例
claw-pliers file info 1771427558V8f5SDqd
```

输出示例：
```
File ID:       1771427558V8f5SDqd
Name:          file.txt
Size:          1.2 MB (1258291 bytes)
Type:          text/plain
Created:       2026-02-18T15:12:38Z
```

### 邮件命令 (Stub)

```bash
# 发送邮件
claw-pliers mail send --to user@example.com --subject "Subject" --body "Body"

# 列出邮件账户
claw-pliers mail list
```

### 图像命令 (Stub)

```bash
# 转换图像格式
claw-pliers image convert input.jpg output.png

# OCR 文字识别
claw-pliers image ocr image.png
```

---

## 配置文件

### 主配置 (config/config.yaml)

```yaml
server:
  port: 8080
  log_level: info

auth:
  local_key: "change-me-in-production"

includes:
  - name: file
    path: "./file-config.yaml"
  - name: mail
    path: "./mail-config.yaml"
  - name: image
    path: "./image-config.yaml"
```

### File 配置 (config/file-config.yaml)

```yaml
database:
  path: "./data/claw-pliers.db"

upload:
  max_size_mb: 1024

minio:
  endpoint: "localhost:9000"
  access_key: "minioadmin"
  secret_key: "minioadmin"
  bucket: "claw-pliers"
  use_ssl: false
```

### Mail 配置 (config/mail-config.yaml)

```yaml
accounts: []

webhook:
  url: "http://127.0.0.1:18789/hooks/agent"
  enable: false

monitoring:
  poll_interval: "30s"
```

### Image 配置 (config/image-config.yaml)

```yaml
libvips:
  path: ""

ocr:
  api_key: ""

vision:
  api_key: ""

image_generation:
  api_key: ""
```

---

## 环境变量

### 服务端

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `MINIO_API_PORT` | 9000 | MinIO API 端口 |
| `MINIO_CONSOLE_PORT` | 9001 | MinIO 控制台端口 |
| `MINIO_ROOT_USER` | minioadmin | MinIO 用户名 |
| `MINIO_ROOT_PASSWORD` | minioadmin | MinIO 密码 |
| `CLAWPLIERS_PORT` | 8080 | 服务端口 |
| `CLAWPLIERS_AUTH_LOCAL_KEY` | change-me-in-production | 认证密钥 |

### CLI

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `CLAWPLIERS_ENDPOINT` | http://localhost:8080 | API 端点 |
| `CLAWPLIERS_AUTH_LOCAL_KEY` | (从配置文件读取) | 认证密钥 |
| `CLAWPLIERS_CONFIG_DIR` | ~/.config/claw-pliers | 配置目录 |

---

## 测试结果

### File 模块 API ✅

| 测试项 | 状态 | 说明 |
|--------|------|------|
| 健康检查 | ✅ 通过 | `GET /health` 返回 `{"status":"ok","version":"dev"}` |
| 认证检查 | ✅ 通过 | 未认证请求返回 `{"code":10001,"message":"unauthorized"}` |
| 文件上传 | ✅ 通过 | 成功上传文件，返回 file_id |
| 文件列表 | ✅ 通过 | 返回文件列表，包含分页参数 |
| 文件信息 | ✅ 通过 | 返回指定文件详情 |
| 文件下载 | ✅ 通过 | 成功下载文件内容 |
| 文件删除 | ✅ 通过 | 返回 `{"message":"file_deleted"}` |

### File 模块 CLI ✅

| 测试项 | 状态 | 说明 |
|--------|------|------|
| file put | ✅ 通过 | 上传文件，显示进度条 |
| file list | ✅ 通过 | 表格形式输出 |
| file get | ✅ 通过 | 下载文件，显示进度条 |
| file delete | ✅ 通过 | 删除成功提示 |
| file info | ✅ 通过 | 显示文件详情 |
| 配置加载 | ✅ 通过 | 自动读取项目配置 |

### Mail 模块 ⚠️

| 测试项 | 状态 | 说明 |
|--------|------|------|
| API 路由 | ❌ 未实现 | 返回 `404 page not found` |

### Image 模块 ⚠️

| 测试项 | 状态 | 说明 |
|--------|------|------|
| API 路由 | ❌ 未实现 | 无相关路由注册 |

---

## 开发环境端口

开发环境使用以下端口（可通过环境变量覆盖）：

| 服务 | 端口 |
|------|------|
| API 服务 | 18080 |
| MinIO API | 19000 |
| MinIO Console | 19001 |

启动命令：
```bash
MINIO_API_PORT=19000 MINIO_CONSOLE_PORT=19001 CLAWPLIERS_PORT=18080 docker-compose -f docker-compose.dev.yaml up -d
```

```bash
curl http://localhost:8080/health
```

响应：
```json
{
  "status": "ok",
  "version": "1.0.0"
}
```
