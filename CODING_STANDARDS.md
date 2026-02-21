# Claw Pliers 代码规范

本文档定义了 Claw Pliers 项目的代码编写标准，包括通用规范和项目特定规范两部分。

> **注意**：本规范基于 [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md) 和社区最佳实践制定。

---

## 第一部分：通用规范

### 1. 代码格式

#### 1.1 格式化工具

所有代码必须通过 `gofmt` 和 `goimports` 格式化：

```bash
# 格式化所有 Go 文件
gofmt -w .
goimports -w .
```

#### 1.2 导入顺序

导入必须按以下顺序分组，组间用空行分隔：

```go
import (
    // 1. 标准库
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "net/http"
    "strconv"
    "strings"
    "time"

    // 2. 第三方包
    "github.com/gin-gonic/gin"
    "github.com/rs/zerolog"
    "github.com/spf13/cobra"
    "gorm.io/gorm"

    // 3. 项目内部包
    "github.com/kiry163/claw-pliers/internal/config"
    "github.com/kiry163/claw-pliers/internal/database"
    "github.com/kiry163/claw-pliers/internal/service"
)
```

#### 1.3 包命名

- 使用小写字母，简短且有描述性
- 避免使用下划线（除非是为了兼容已有的命名）
- 不要使用复数形式（如使用 `file`，不使用 `files`）

#### 1.4 文件命名

- 使用小写字母和下划线组合
- 命名规则：`{功能}_{类型}.go`
  - `handlers.go` - HTTP 处理器
  - `service.go` - 业务逻辑
  - `models.go` - 数据模型
  - `router.go` - 路由定义

---

### 2. 命名规范

#### 2.1 变量和函数命名

| 类型 | 命名方式 | 示例 |
|------|----------|------|
| 变量 | camelCase | `fileID`, `userName`, `maxSize` |
| 常量 | PascalCase | `MaxUploadSize`, `DefaultPort` |
| 未导出常量 | camelCase | `maxRetries`, `defaultTimeout` |
| 函数 | PascalCase（导出）或 camelCase（未导出） | `GetUser()`, `calculateHash()` |
| 接口 | PascalCase，以 -er 结尾 | `Storage`, `Handler` |

#### 2.2 JSON 标签命名

所有 JSON 序列化字段使用下划线命名：

```go
type File struct {
    FileID       string `json:"file_id"`
    OriginalName string `json:"original_name"`
    ObjectKey    string `json:"object_key"`
    Size         int64  `json:"size"`
    MimeType     string `json:"mime_type"`
    FolderID     string `json:"folder_id,omitempty"`
    CreatedAt    string `json:"created_at"`
}
```

#### 2.3 数据库字段命名

使用下划线命名：

```go
type File struct {
    FileID       string    `gorm:"column:file_id" json:"file_id"`
    OriginalName string    `gorm:"column:original_name" json:"original_name"`
}
```

---

### 3. 函数设计

#### 3.1 函数长度

- 单个函数建议不超过 50 行
- 如果函数过长，考虑拆分

#### 3.2 参数数量

- 建议不超过 4 个参数
- 超过 4 个使用结构体封装

```go
// 不推荐
func CreateFile(ctx context.Context, name string, size int64, mimeType string, folderID string) error

// 推荐
type CreateFileRequest struct {
    Name      string
    Size      int64
    MimeType  string
    FolderID  string
}

func CreateFile(ctx context.Context, req CreateFileRequest) error
```

#### 3.3 返回值

- 多返回值时，必须命名返回值
- 错误必须作为最后一个返回值

```go
// 推荐
func GetUser(id string) (*User, error)

// 不推荐（未命名返回值）
func GetUser(id string) (User, error)
```

---

### 4. 错误处理

#### 4.1 错误检查

- 错误必须显式处理，不能忽略
- 使用哨兵错误（Sentinel Errors）或自定义错误类型

```go
// 推荐
if err != nil {
    return fmt.Errorf("failed to get user: %w", err)
}

// 不推荐
_ = err
```

#### 4.2 错误包装

使用 `fmt.Errorf` 和 `%w` 包装错误，保留错误链：

```go
if err := db.First(&user, id).Error; err != nil {
    return fmt.Errorf("database error: %w", err)
}
```

#### 4.3 避免重复错误信息

```go
// 不推荐
if err != nil {
    return errors.New("error: " + err.Error())
}

// 推荐
if err != nil {
    return fmt.Errorf("failed to process: %w", err)
}
```

---

### 5. 注释规范

#### 5.1 注释要求

每个导出的函数和类型必须有注释：

```go
// FileHandler 处理文件相关的 HTTP 请求
type FileHandler struct {
    // ...
}

// UploadFile 处理文件上传请求
// 上传成功后返回文件 ID 和元数据
func (h *FileHandler) UploadFile(c *gin.Context) {
    // ...
}
```

#### 5.2 注释风格

- 使用完整的句子
- 注释以被注释对象的名称开头
- 句末使用句号

```go
// GetFile 根据文件 ID 获取文件信息
func GetFile(fileID string) (*File, error) {
    // ...
}
```

---

### 6. Context 使用

- HTTP 处理程序必须将 context 传递给下层调用
- 使用 `c.Request.Context()` 获取 context
- 不要在 context 中存储值

```go
func (h *FileHandler) GetFile(c *gin.Context) {
    ctx := c.Request.Context()
    
    file, err := h.service.GetFile(ctx, fileID)
    // ...
}
```

---

## 第二部分：项目规范

### 7. 项目结构

```
claw-pliers/
├── cmd/
│   └── claw-pliers/          # 服务入口
│       └── main.go
├── cli/                       # CLI 工具
│   ├── main.go
│   ├── file.go
│   ├── mail.go
│   └── image.go
├── internal/
│   ├── api/                   # HTTP 层
│   │   ├── router.go         # 路由定义
│   │   ├── middleware.go     # 中间件
│   │   ├── file_handler.go  # 文件操作处理器
│   │   ├── folder_handler.go # 文件夹操作处理器
│   │   ├── mail_handler.go  # 邮件操作处理器
│   │   └── response.go     # 响应工具
│   ├── config/               # 配置管理
│   │   └── config.go
│   ├── database/             # 数据库层（GORM）
│   │   ├── models.go        # 数据模型
│   │   └── repository.go    # 数据库操作
│   ├── service/             # 业务逻辑层
│   │   ├── file_service.go
│   │   ├── folder_service.go
│   │   └── mail_service.go
│   ├── storage/             # 存储抽象层
│   │   ├── minio.go         # MinIO 实现
│   │   └── interface.go     # 存储接口定义
│   ├── logger/              # 日志模块
│   │   └── logger.go
│   ├── mail/                # 邮件模块
│   │   └── mail.go
│   └── image/               # 图像模块
│       └── image.go
├── config/                   # 配置文件
│   ├── config.yaml
│   ├── file-config.yaml
│   ├── mail-config.yaml
│   └── image-config.yaml
├── docker-compose.yaml
├── Makefile
└── README.md
```

---

### 8. API 设计规范

#### 8.1 响应格式

所有 API 响应使用统一的 JSON 格式：

```go
// 成功响应
{
    "code": 0,
    "message": "操作成功",
    "data": { ... }
}

// 失败响应
{
    "code": 10001,
    "message": "错误描述"
}

// 列表响应
{
    "code": 0,
    "message": "操作成功",
    "data": {
        "total": 100,
        "items": [...]
    }
}

// 简单消息响应
{
    "code": 0,
    "message": "操作成功"
}
```

#### 8.2 错误码定义

| 错误码 | 含义 |
|--------|------|
| 0 | 成功 |
| 10001 | 未授权 |
| 10002 | 资源不存在 |
| 10003 | 资源已失效 |
| 10004 | 请求参数错误 |
| 19999 | 内部服务器错误 |

#### 8.3 HTTP 状态码使用

| 状态码 | 使用场景 |
|--------|----------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未认证 |
| 403 | 无权限 |
| 404 | 资源不存在 |
| 410 | 资源已失效（如过期链接） |
| 500 | 内部错误 |

#### 8.4 处理器函数模板

```go
// HandlerName 处理某个业务操作
// 1. 解析和验证输入
// 2. 调用服务层
// 3. 记录日志
// 4. 返回响应
func (h *HandlerName) HandlerFunc(c *gin.Context) {
    // 1. 解析请求参数
    param := c.Query("param")
    if param == "" {
        response.Error(c, http.StatusBadRequest, 10004, "param is required")
        return
    }

    // 2. 调用服务层
    ctx := c.Request.Context()
    result, err := h.service.DoSomething(ctx, param)
    if err != nil {
        // 3. 错误处理和日志
        logger.Get().Error().
            Err(err).
            Str("param", param).
            Msg("failed to do something")
        response.Error(c, http.StatusInternalServerError, 19999, "internal error")
        return
    }

    // 4. 返回成功响应
    response.Success(c, result)
}
```

#### 8.5 响应辅助函数

定义统一的响应函数：

```go
// response/response.go
package response

// Success 返回成功响应
func Success(c *gin.Context, data interface{}) {
    c.JSON(http.StatusOK, gin.H{
        "code":    0,
        "message": "success",
        "data":    data,
    })
}

// SuccessWithMessage 返回带消息的成功响应
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
    c.JSON(http.StatusOK, gin.H{
        "code":    0,
        "message": message,
        "data":    data,
    })
}

// Error 返回错误响应
func Error(c *gin.Context, status int, code int, message string) {
    c.JSON(status, gin.H{
        "code":    code,
        "message": message,
    })
}

// Message 返回简单消息响应
func Message(c *gin.Context, message string) {
    c.JSON(http.StatusOK, gin.H{
        "code":    0,
        "message": message,
    })
}

// Page 返回分页响应
func Page(c *gin.Context, items interface{}, total int64) {
    c.JSON(http.StatusOK, gin.H{
        "code":    0,
        "message": "success",
        "data": gin.H{
            "total": total,
            "items": items,
        },
    })
}
```

---

### 9. 数据库规范

#### 9.1 模型定义

使用 GORM，字段标签简化：

```go
// 文件模型
type File struct {
    ID           uint      `gorm:"primaryKey" json:"-"`
    FileID       string    `gorm:"column:file_id;uniqueIndex" json:"file_id"`
    OriginalName string    `gorm:"column:original_name" json:"original_name"`
    ObjectKey    string    `gorm:"column:object_key" json:"object_key"`
    Size         int64     `gorm:"column:size" json:"size"`
    MimeType     string    `gorm:"column:mime_type" json:"mime_type"`
    FolderID     *string   `gorm:"column:folder_id" json:"folder_id,omitempty"`
    CreatedBy    string    `gorm:"column:created_by" json:"created_by"`
    CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
    UpdatedAt    time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (File) TableName() string {
    return "files"
}
```

> **注意**：模型定义不需要添加 size、not null 等约束。开发阶段删除数据库文件重新生成即可。

#### 9.2 Repository 层

使用 Repository 模式封装数据库操作：

```go
type FileRepository interface {
    Create(ctx context.Context, file *File) error
    GetByID(ctx context.Context, fileID string) (*File, error)
    GetByPath(ctx context.Context, path string) (*File, error)
    List(ctx context.Context, opts ListOptions) ([]File, int64, error)
    Update(ctx context.Context, file *File) error
    Delete(ctx context.Context, fileID string) error
}

type ListOptions struct {
    FolderID *string
    Limit    int
    Offset   int
    Order    string // "asc" or "desc"
    Keyword  string
}
```

#### 9.3 统一数据库入口

只使用一个数据库连接，使用 GORM。开发阶段如需更新表结构，删除数据库文件和 docker 挂载目录后重新生成即可：

```go
// database/database.go
type Database struct {
    *gorm.DB
}

// 全局数据库实例
var DB *Database

func Init(cfg config.DatabaseConfig) error {
    db, err := gorm.Open(sqlite.Open(cfg.Path), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Warn),
    })
    if err != nil {
        return fmt.Errorf("failed to open database: %w", err)
    }

    // 自动迁移（开发阶段使用）
    // 如需更新表结构，删除数据库文件后重新生成
    if err := db.AutoMigrate(
        &File{},
        &Folder{},
        &ShareLink{},
        &AuditLog{},
    ); err != nil {
        return fmt.Errorf("failed to migrate: %w", err)
    }

    DB = &Database{db}
    return nil
}

    // 自动迁移
    if err := db.AutoMigrate(
        &File{},
        &Folder{},
        &ShareLink{},
        &AuditLog{},
    ); err != nil {
        return fmt.Errorf("failed to migrate: %w", err)
    }

    DB = &Database{db}
    return nil
}
```

---

### 10. 依赖注入

#### 10.1 Handler 依赖注入

通过结构体和构造函数注入依赖：

```go
type FileHandler struct {
    service  fileService
    storage  storage.Interface
    config   *config.Config
}

// 构造函数
func NewFileHandler(svc fileService, st storage.Interface, cfg *config.Config) *FileHandler {
    return &FileHandler{
        service: svc,
        storage: st,
        config:  cfg,
    }
}
```

#### 10.2 Service 依赖注入

```go
type FileService struct {
    repo     repository.FileRepository
    storage  storage.Interface
    logger   zerolog.Logger
}

func NewFileService(repo repository.FileRepository, st storage.Interface) *FileService {
    return &FileService{
        repo:    repo,
        storage: st,
        logger:  logger.Get().With().Str("module", "file_service").Logger(),
    }
}
```

#### 10.3 路由中组装依赖

```go
func NewRouter(cfg *config.Config) *gin.Engine {
    // 初始化依赖
    db, _ := database.Init(cfg.Database)
    repo := repository.NewFileRepository(db)
    st := minio.NewStorage(cfg.Minio)
    svc := service.NewFileService(repo, st)
    handler := handler.NewFileHandler(svc, st, cfg)

    // 注册路由
    // ...
}
```

---

### 11. 日志规范

#### 11.1 日志级别

| 级别 | 使用场景 |
|------|----------|
| Debug | 调试信息，仅开发环境输出 |
| Info | 正常业务流程记录 |
| Warn | 可恢复的错误，需要关注 |
| Error | 不可恢复的错误 |

#### 11.2 日志内容要求

每个日志条目应包含：

1. **操作名称**：使用 `[模块名]` 前缀
2. **关键信息**：如 ID、路径、用户
3. **错误详情**：错误发生时记录具体原因

```go
// 常规操作日志
log.Info().
    Str("action", "upload_file").
    Str("file_id", fileID).
    Str("filename", filename).
    Int64("size", size).
    Msg("file uploaded successfully")

// 错误日志
log.Error().
    Err(err).
    Str("action", "delete_file").
    Str("file_id", fileID).
    Msg("failed to delete file")

// 带上下文的日志
log.WithContext(ctx).Info().
    Str("user", getUser(c)).
    Str("ip", c.ClientIP()).
    Msg("request processed")
```

#### 11.3 禁止事项

- **禁止**使用 `fmt.Println` 或标准库 `log`
- **禁止**记录敏感信息（密码、密钥、Token）
- **禁止**记录完整的请求体（除非调试需要）

---

### 12. 配置规范

#### 12.1 配置文件格式

使用 YAML 格式，配置项使用下划线命名：

```yaml
server:
  port: 8080
  log_level: info
  public_endpoint: ""

database:
  path: "./data/claw-pliers.db"

auth:
  local_key: "change-me-in-production"
  jwt_secret: ""
  admin_username: "admin"
  admin_password: ""

upload:
  max_size_mb: 1024

minio:
  endpoint: "localhost:9000"
  access_key: "minioadmin"
  secret_key: "minioadmin"
  bucket: "claw-pliers"
  use_ssl: false

logger:
  level: info
  format: console
  output_path: "./logs"
  enable_file: true
```

#### 12.2 配置加载

配置文件加载顺序（按优先级）：

1. 项目目录配置文件 `./config/config.yaml`
2. 用户目录配置文件 `~/.config/claw-pliers/config.yaml`
3. 默认值

> **注意**：本项目不考虑使用环境变量覆盖配置，减少编码复杂度。

---

### 13. 安全规范

#### 13.1 敏感信息

- 禁止在代码中硬编码密钥、密码
- 使用配置文件或环境变量
- 日志中禁止记录敏感信息

#### 13.2 随机数生成

使用加密安全的随机数生成器：

```go
import "crypto/rand"

// 推荐
func generateToken(length int) string {
    b := make([]byte, length)
    if _, err := rand.Read(b); err != nil {
        // 处理错误
    }
    return hex.EncodeToString(b)
}

// 不推荐（不安全）
func randomString(n int) string {
    // 使用时间戳 - 不安全
}
```

#### 13.3 输入验证

- 所有用户输入必须验证
- 使用白名单验证方式

---

### 14. CLI 规范

#### 14.1 命令结构

使用 Cobra，遵循以下结构：

```
claw-pliers <command> <subcommand> [flags]
```

示例：
```bash
claw-pliers file put <local> [remote]
claw-pliers file list [--limit N]
claw-pliers mail send --to user@example.com --subject "Subject"
```

#### 14.2 配置加载优先级

1. 命令行 flags
2. 环境变量
3. 配置文件（项目目录）
4. 配置文件（用户目录）
5. 默认值

---

## 附录

### A. 代码检查常用命令

```bash
# 格式化代码
gofmt -w .
goimports -w .

# 代码检查
go vet ./...

# 运行测试
go test -v ./...
```

### B. golangci-lint 配置（可选）

如果需要更严格的代码检查，可以使用 golangci-lint：

```yaml
run:
  timeout: 5m

linters:
  enable:
    - gofmt
    - goimports
    - golint
    - errcheck
    - staticcheck
    - unused

linters-settings:
  errcheck:
    check-type-assertions: true

issues:
  exclude-use-default: false
```

---

## 变更历史

| 版本 | 日期 | 变更说明 |
|------|------|----------|
| 1.0.0 | 2026-02-21 | 初始版本 |
