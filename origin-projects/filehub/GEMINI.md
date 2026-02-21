# FileHub 项目 AI 协作指南

## 1. 项目概述

FileHub 是一个轻量级的文件管理服务，采用 Go 语言开发，并集成了 Vue 3 构建的前端界面。

- **后端技术栈**:
  - Web 框架: `Gin`
  - 数据库: `SQLite`
  - 对象存储: `MinIO` (通过服务代理)

- **前端技术栈**:
  - 框架: `Vue 3`
  - 构建工具: `Vite`
  - UI 组件库: `Element Plus`

- **部署方式**:
  - 单一可执行文件 (前端资源被嵌入)
  - `Docker` (通过 `docker-compose.yaml` 编排)

- **核心功能**:
  - 文件的上传、下载、列表、预览、删除。
  - 文件分享链接。
  - 基于 `JWT` 的 Web UI 认证。
  - 基于 `X-Local-Key` 的 CLI/API 认证。

## 2. 项目结构

- `cmd/filehub/`: 后端服务主程序。
- `cmd/filehub-cli/`: 命令行工具主程序。
- `internal/`: 项目核心业务逻辑。
- `web-ui/`: 前端 Vue 项目源码。
- `web/dist/`: 存放由 `web-ui` 构建出的静态资源，会被 `embed` 到 Go 程序中。
- `config.yaml`: 服务端核心配置文件。
- `docker-compose.yaml`: Docker 编排文件。

## 3. 开发与构建

### 环境准备

- **后端**: Go (版本 >= 1.23)
- **前端**: Node.js (用于构建前端资源)
- **运行环境**: Docker 和 Docker Compose (推荐)

### 本地开发

**启动后端服务:**

```bash
# 首次启动或依赖变更后
go mod tidy

# 启动服务
go run ./cmd/filehub
```

后端服务默认运行在 `http://localhost:8080`。

**启动前端开发服务器:**

```bash
cd web-ui

# 首次启动或依赖变更后
npm install

# 启动开发服务器
npm run dev
```

前端开发服务器通常会运行在 `http://localhost:5173`，并会将 `/api` 请求代理到后端。

### 使用 Docker 运行 (推荐)

这是最简单的启动方式，包含了所有依赖的服务。

```bash
docker-compose up -d
```

启动后，可以通过以下地址访问：
- **Web UI**: `http://localhost:8080`
- **MinIO 控制台**: `http://localhost:9001` (默认账号/密码: `minioadmin`/`minioadmin`)

### 构建生产环境

项目支持将前端构建的静态资源嵌入到 Go 二进制文件中，实现单一文件部署。

1.  **构建前端应用**:

    ```bash
    cd web-ui
    npm install
    npm run build
    ```

2.  **拷贝静态资源**:
    构建产物会生成在 `web-ui/dist` 目录下。需要将其拷贝到 `web/dist`。

    ```bash
    # 在项目根目录执行
    rm -rf web/dist
    cp -r web-ui/dist web/dist
    ```

3.  **构建 Go 应用**:

    ```bash
    # 在项目根目录执行
    go build -o filehub ./cmd/filehub
    ```

    构建成功后，会生成名为 `filehub` 的可执行文件。

## 4. 编码约定与规范

- **Go 后端**:
  - 遵循 Go 社区的通用编码规范。
  - API 设计遵循 RESTful 风格。
  - 配置文件 (`config.yaml`) 中的注释为中文，在修改时请保持。

- **Vue 前端**:
  - 使用 Vue 3 `<script setup>` 语法。
  - API 请求通过 `src/api` 目录下的模块进行封装。
  - 状态管理使用 `pinia` (如 `src/store` 所示)。

- **Git 提交**:
  - 遵循 Conventional Commits 规范，例如使用 `feat:`, `fix:`, `refactor:` 等作为提交信息前缀。

## 5. 命令行工具 (CLI)

项目提供了一个 CLI 工具 `filehub-cli` 进行交互。

**初始化配置**:

```bash
filehub-cli config init \
  --endpoint http://localhost:8080 \
  --local-key filehub-local-key
```

**常用命令**:

```bash
# 上传文件
filehub-cli upload <file_path>

# 文件列表
filehub-cli list

# 下载文件
filehub-cli download filehub://<file_id>
```