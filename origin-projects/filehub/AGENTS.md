# AGENTS.md - FileHub 开发指南

## 项目概述

FileHub 是一个轻量级文件管理服务，专为个人使用和 AI Agent 场景设计：
- **后端**: Go + Gin + SQLite + GORM
- **存储**: MinIO（服务端代理）
- **前端**: Vue3（构建后嵌入 Go 二进制）
- **CLI**: Cobra 编写的命令行工具
- **日志**: zerolog

## 设计目标

1. **单用户私人网盘**：只有一个管理员用户
2. **Agent 友好**：CLI 主要供 AI Agent 使用，通过 `X-Local-Key` 认证
3. **代码优雅**：分层清晰，使用 GORM，日志完善
4. **功能完整**：文件上传/下载、文件夹管理、移动操作、预览等

## 构建命令

### 后端
```bash
# 运行开发服务器
go run ./cmd/filehub

# 构建生产环境二进制
go build -o filehub ./cmd/filehub

# 构建包含嵌入 Web UI 的二进制
cd web-ui && npm ci && npm run build && rm -rf ../web/dist && cp -r dist ../web/dist && cd ..
go build -o filehub ./cmd/filehub
```

### CLI
```bash
# 运行 CLI
go run ./cmd/filehub-cli

# 构建 CLI 二进制
go build -o filehub-cli ./cmd/filehub-cli
```

### 测试
```bash
# 运行所有测试
go test ./...

# 运行单个测试
go test -v ./internal/api/... -run TestName

# 运行带覆盖率的测试
go test -cover ./...
```

### 代码检查与格式化
```bash
# Go vet
go vet ./...

# 格式化代码
gofmt -w .

# 检查并修复导入
goimports -w .
```

### Docker
```bash
# 开发环境
docker-compose up -d

# 构建 Docker 镜像
docker build -t filehub .
```

## 代码风格指南

### 导入顺序

按以下顺序分组导入，组间用空行分隔：
1. 标准库
2. 第三方包
3. 项目内部包

```go
import (
    "context"
    "encoding/json"
    "errors"
    "io"
    "net/http"
    "strconv"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
    "github.com/rs/zerolog"
    "gorm.io/gorm"

    "github.com/kiry163/filehub/internal/config"
    "github.com/kiry163/filehub/internal/db"
    "github.com/kiry163/filehub/internal/service"
)
```

### 格式化

- 使用 `gofmt` 自动格式化（项目使用标准 Go 格式化）
- 使用 `goimports` 自动管理导入
- 单行最大长度：无严格限制，但保持合理
- 使用空行分隔逻辑章节

### 命名规范

- **变量/函数**: camelCase
- **常量**: PascalCase 或未导出时用 camelCase
- **类型/结构体**: PascalCase
- **包**: 小写，简短，描述性
- **文件**: 小写加下划线（如 `folder_handlers.go`）

### 错误处理

- 尽早返回错误，避免嵌套条件
- 使用辅助函数处理 API 响应：
  ```go
  Error(c, http.StatusBadRequest, 10004, "错误信息")
  OK(c, data)
  Message(c, "成功信息")
  ```
- 记录有意义的错误信息
- 使用自定义错误码（如 10004 表示请求错误，19999 表示内部错误）

### HTTP 处理器

处理器函数遵循以下模式：
```go
func (h *Handler) HandlerName(c *gin.Context) {
    // 1. 解析和验证输入
    // 2. 调用服务层
    // 3. 记录审计日志
    // 4. 返回响应
}
```

### 数据库（使用 GORM）

- 使用 GORM 进行数据库操作
- 遵循 `internal/db/models.go` 中的模型定义
- 使用 context 支持取消
- 使用 GORM AutoMigrate 进行迁移
- 示例：
  ```go
  // 创建记录
  result := db.Create(&record)
  
  // 查询记录
  var file File
  db.First(&file, "file_id = ?", fileID)
  
  // 更新记录
  db.Model(&file).Update("name", newName)
  
  // 删除记录
  db.Delete(&file)
  ```

### 配置管理

- 使用 `config.yaml` 配合环境变量覆盖
- 环境变量遵循模式：`FILEHUB_<模块>_<键>`
- 示例：`FILEHUB_SERVER_PORT`、`FILEHUB_AUTH_LOCAL_KEY`

### 响应格式

使用 `gin.H` 构建 JSON 响应：
```go
OK(c, gin.H{
    "field1": value1,
    "field2": value2,
})

// 列表响应：
OK(c, gin.H{
    "total": totalCount,
    "items": items,
})
```

### 审计日志

使用 audit 函数记录关键操作：
```go
h.audit(c, "action_name", "file_id", "user", "success", "message")
```

### Context 使用

始终传递 context 到服务和数据库调用：
```go
h.Service.Operation(c.Request.Context(), args...)
```

### 日志规范（zerolog）

项目使用 zerolog 进行日志记录，所有日志通过全局 logger 输出。

#### 日志级别

- `zerolog.DebugLevel`: 调试信息，仅在 `log_level: debug` 时输出
- `zerolog.InfoLevel`: 常规操作信息（默认）
- `zerolog.WarnLevel`: 警告信息
- `zerolog.ErrorLevel`: 错误信息

#### 日志内容要求

每个日志条目应包含：
1. **操作名称**：使用 `[模块名]` 前缀，如 `[Upload]`, `[DeleteFolder]`
2. **关键信息**：如文件 ID、路径、用户、操作结果
3. **错误详情**：错误发生时记录具体原因

#### 日志示例

```go
// 常规操作日志
log.Info().Str("action", "upload").
    Str("file_id", fileID).
    Str("filename", filename).
    Int64("size", size).
    Msg("file uploaded successfully")

// 错误日志
log.Error().Err(err).
    Str("action", "delete_file").
    Str("file_id", fileID).
    Msg("failed to delete file")

// 带上下文的日志
log.WithContext(ctx).Info().
    Str("user", getUser(c)).
    Str("ip", c.ClientIP()).
    Msg("request processed")
```

#### 禁止事项

- **禁止**使用 `fmt.Println` 或标准库 `log`
- **禁止**记录敏感信息（密码、密钥、Token）
- **禁止**记录完整的请求体（除非调试需要）

## 项目结构

```
filehub/
├── cmd/
│   ├── filehub/          # 主服务器二进制
│   └── filehub-cli/      # CLI 二进制
├── internal/
│   ├── api/              # HTTP 处理器和路由
│   ├── cli/              # CLI 命令
│   ├── config/           # 配置加载
│   ├── db/               # 数据库操作（GORM）
│   │   └── models.go     # GORM 模型定义
│   ├── service/          # 业务逻辑
│   ├── storage/          # MinIO 存储抽象
│   └── version/          # 版本信息
├── web/                  # 嵌入的 Web UI
└── web-ui/               # Vue3 源码
```

## 常见任务

### 添加新的 API 端点

1. 在 `internal/api/handlers.go` 添加处理器
2. 在 `internal/api/router.go` 注册路由
3. 如需要，在 `internal/service/service.go` 添加服务方法
4. 如需要，在 `internal/db/models.go` 添加 GORM 模型

### 添加新的 CLI 命令

1. 在 `internal/cli/commands_<name>.go` 创建文件
2. 在 root.go 或相关 init 函数中注册命令

### 数据库迁移

- 数据库为 SQLite，路径在配置中定义
- 使用 GORM AutoMigrate 自动迁移
- 模型定义在 `internal/db/models.go`

## 测试指南

- 在被测试代码同目录创建 `*_test.go` 文件
- 使用描述性测试名称：`TestHandlerName_期望行为`
- 对多个测试用例使用表驱动测试
- 适当模拟外部依赖
