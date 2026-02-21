# FileHub

FileHub 是一个轻量级文件管理服务，支持 Web UI 和 CLI 操作，专为个人使用和 AI Agent 场景设计。

- **后端**: Go + Gin + SQLite + GORM
- **存储**: MinIO（服务端代理）
- **前端**: Vue3（构建后嵌入 Go 二进制）
- **部署**: 单二进制或 Docker

## 功能特性

- 文件上传 / 列表 / 预览 / 下载 / 删除
- 文件夹管理（创建、重命名、移动、删除、递归删除）
- 移动文件到不同文件夹
- 视频流媒体（Range 请求）
- 认证
  - Web UI: JWT 登录 + Refresh Token
  - CLI/Agent: `X-Local-Key`（无需存储用户名密码）
- 关键操作审计日志

## 快速开始（Docker）

前置条件：Docker + Docker Compose

```bash
docker-compose up -d
```

访问地址：
- Web UI: `http://localhost:8080`
- MinIO 控制台: `http://localhost:9001`

MinIO 生命周期：
- Docker 配置初始化了 1 天后中止未完成分片上传的规则

默认开发配置在 `config.yaml`。

## 一键安装（Docker + CLI）

Linux/amd64，从 GitHub Release 拉取镜像并安装 `filehub-cli`。

```bash
curl -fsSL https://raw.githubusercontent.com/kiry163/filehub/main/scripts/install.sh | bash
```

可选参数：
```bash
curl -fsSL https://raw.githubusercontent.com/kiry163/filehub/main/scripts/install.sh | bash -s -- \
  --port 18080 \
  --version v0.1.0
```

注意：
- 安装脚本会保留已有的配置和数据
- 默认拉取 `kirydocker/filehub:latest`，如失败则回退到最新 Release 标签

## 推荐服务器目录结构

部署在服务器时，将所有运行时状态放在单一目录：

```text
/opt/filehub/
  docker-compose.yml
  config.yaml
  data/
    filehub.db
    minio/
```

这样便于升级和备份。

## 升级（Docker）

如果通过一键脚本安装，默认运行时目录为 `~/.filehub`。

1) 备份（推荐）

```bash
cd ~/.filehub
tar -czf filehub-backup-$(date +%F).tar.gz config.yaml data/
```

2) 拉取新版本并重启

```bash
cd ~/.filehub
docker compose pull
docker compose up -d
```

如果使用版本锁定（推荐），更新 `docker-compose.yml` 中的 `image:` 标签（如 `ghcr.io/kiry163/filehub:v0.1.1`），然后执行上述命令。

## 备份与恢复

需要备份的内容：
- `config.yaml`
- `data/`（SQLite 数据库 + MinIO 数据）

备份：

```bash
cd ~/.filehub
tar -czf filehub-backup-$(date +%F).tar.gz config.yaml data/
```

恢复：

```bash
cd ~/.filehub
docker compose down
rm -rf data config.yaml
tar -xzf filehub-backup-YYYY-MM-DD.tar.gz
docker compose up -d
```

## CLI

初始化配置：

```bash
filehub-cli config init \
  --endpoint http://localhost:8080 \
  --local-key filehub-local-key
```

命令列表：

```bash
# 上传文件
filehub-cli upload ./myfile.zip

# 列出文件
filehub-cli list --limit 10

# 下载文件
filehub-cli download filehub://<id> --output ./downloads

# 删除文件
filehub-cli delete filehub://<id>

# 创建文件夹
filehub-cli mkdir /foldername

# 列出文件夹
filehub-cli ls /foldername

# 移动文件或文件夹
filehub-cli mv filehub://<id> /target/folder

# 删除文件夹
filehub-cli rm /foldername

# 备份（压缩 ~/.filehub/data）
filehub-cli backup

# 备份排除 MinIO 内部元数据 (.minio.sys)
```

## 配置说明

服务端配置文件：`config.yaml`

主要字段：
- `server.port`: HTTP 端口
- `server.log_level`: 日志级别（debug/info/warn/error）
- `database.path`: SQLite 数据库路径
- `auth.admin_username` / `auth.admin_password`: Web 登录凭据
- `auth.jwt_secret`: JWT 签名密钥
- `auth.local_key`: CLI 密钥（`X-Local-Key` 头使用）
- `minio.*`: MinIO 连接配置

环境变量覆盖（示例）：

```bash
FILEHUB_SERVER_PORT=8080
FILEHUB_SERVER_LOG_LEVEL=debug
FILEHUB_DATABASE_PATH=./data/filehub.db
FILEHUB_AUTH_LOCAL_KEY=your-local-key
FILEHUB_MINIO_ENDPOINT=minio:9000
```

## Web 路由

- `/` 文件列表
- `/upload` 上传页面
- `/app/files/:id` 文件详情
- `/app/folders/:id` 文件夹内容
- `/login` 登录页面

Go 服务端提供嵌入的静态资源，非 `/api/*` 路由返回 `index.html`（SPA 模式）。

## API

基础路径：`/api/v1`

**注意**：除 `/auth/*` 外的所有文件/文件夹端点需要认证：
- JWT Token: `Authorization: Bearer <token>` 请求头（Web UI）
- `X-Local-Key`: 请求头（CLI/Agent）

### 认证

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | /auth/login | 用户名密码登录 |
| POST | /auth/refresh | 刷新 Access Token |
| POST | /auth/logout | 登出（撤销所有 Refresh Token） |

### 文件

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | /files | 上传文件 |
| GET | /files | 列出文件（支持 folder_id, keyword, limit, offset） |
| GET | /files/:id | 获取文件元数据 |
| GET | /files/:id/download | 下载文件（支持 Range 流媒体） |
| DELETE | /files/:id | 删除文件 |
| GET | /files/:id/preview | 获取预览流 URL（10 分钟有效期） |
| GET | /files/:id/view | 获取 Web UI 访问 URL |
| PUT | /files/:id/move | 移动文件到文件夹 |
| GET | /files/by-path | 通过路径获取文件 |
| GET | /files/by-path/download | 通过路径下载文件 |

### 文件夹

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | /folders | 创建文件夹 |
| GET | /folders | 列出根文件夹 |
| GET | /folders/by-path | 通过路径获取文件夹 |
| GET | /folders/:id | 获取文件夹信息 |
| GET | /folders/:id/contents | 获取文件夹内容（文件 + 子文件夹） |
| PUT | /folders/:id | 重命名文件夹 |
| PUT | /folders/:id/move | 移动文件夹 |
| DELETE | /folders/:id | 删除文件夹（使用 ?recursive=true 强制删除） |
| GET | /folders/:id/view | 获取 Web UI 访问 URL |

### 查询参数

- `folder_id`: 按文件夹筛选（null 表示根目录）
- `keyword`: 按文件名/文件夹名搜索
- `limit`: 分页限制（默认 20）
- `offset`: 分页偏移（默认 0）
- `order`: 排序方向（asc/desc，默认 desc）
- `recursive`: 删除时使用（true 表示删除非空文件夹）

### 流媒体

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /files/stream?token=... | 通过预览 Token 流式传输文件 |

## 构建

本地开发：

```bash
# 后端
go run ./cmd/filehub

# 前端（开发）
cd web-ui
npm ci
npm run dev
```

生产环境单二进制（嵌入 web-ui）：

```bash
cd web-ui
npm ci
npm run build
rm -rf ../web/dist
cp -r dist ../web/dist

cd ..
go build -o filehub ./cmd/filehub
```

## CI/CD

- CI: `/.github/workflows/ci.yml`
- Release (tag `v*`): 构建并推送 `ghcr.io/kiry163/filehub:<tag>`，创建 GitHub Release 包含二进制文件。
