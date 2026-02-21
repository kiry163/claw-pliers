## Context

项目当前代码存在以下问题需要重构：

1. **数据库层重复**：存在两套数据库实现（`internal/database/database.go` 使用 GORM，`internal/file/db.go` 使用原始 sql），实际只使用了 GORM 版本
2. **API 响应格式不统一**：handlers.go 中存在多种响应格式
3. **单文件过长**：handlers.go 612 行，包含所有 handler
4. **安全隐患**：使用时间戳生成随机数，不安全
5. **缺少依赖注入**：handler 直接使用全局变量或包级变量
6. **缺少日志**：关键操作没有日志记录

## Goals / Non-Goals

**Goals:**
- 统一数据库层，只使用 GORM
- API 响应格式标准化
- 拆分 handlers.go，按功能模块化
- 修复安全问题，使用 crypto/rand
- 引入依赖注入架构
- 完善日志记录
- 添加 Makefile 方便开发

**Non-Goals:**
- 不添加单元测试（按你之前说的不需要）
- 不修改 API 接口行为
- 不引入新的外部依赖

## Decisions

### D1: 数据库层统一
- **决策**：只保留 `internal/database/database.go` (GORM 版本)
- **理由**：GORM 已是业界标准，API 友好，支持自动迁移
- **替代方案考虑**：保留原始 sql 版本 - 排除，因为维护两套代码成本高

### D2: 响应格式统一
- **决策**：创建 `internal/response/response.go`，提供统一响应函数
- **响应格式**：`{"code": 0, "message": "success", "data": ...}`
- **理由**：前端容易解析，错误码便于定位问题

### D3: Handler 拆分策略
- **决策**：按功能拆分为多个文件
  - `file_handler.go` - 文件操作
  - `folder_handler.go` - 文件夹操作
  - `mail_handler.go` - 邮件操作
- **理由**：单一职责，便于维护

### D4: 依赖注入方式
- **决策**：Handler 和 Service 使用构造函数注入
- **理由**：便于单元测试（可 mock），依赖关系清晰
- **实现**：
  ```go
  type FileHandler struct {
      service  FileService
      storage  Storage
      config   *config.Config
  }
  
  func NewFileHandler(svc FileService, st Storage, cfg *config.Config) *FileHandler
  ```

### D5: 随机数生成
- **决策**：封装统一的随机数生成工具，使用 crypto/rand
- **理由**：crypto/rand 是加密安全的，符合安全要求

### D6: 项目结构
- **决策**：采用以下目录结构
  ```
  internal/
  ├── api/          # HTTP 层
  │   ├── router.go
  │   ├── middleware.go
  │   ├── file_handler.go
  │   ├── folder_handler.go
  │   └── mail_handler.go
  ├── response/     # 响应工具（新增）
  ├── database/     # 数据库层
  ├── service/      # 业务逻辑层（新增）
  ├── storage/      # 存储抽象
  ├── config/       # 配置
  └── logger/       # 日志
  ```

### D7: 日志规范
- **决策**：使用 zerolog，记录操作名称、关键信息、错误详情
- **格式**：
  ```go
  log.Info().
      Str("action", "upload_file").
      Str("file_id", fileID).
      Msg("file uploaded")
  ```

## Risks / Trade-offs

| 风险 | 风险描述 | 缓解措施 |
|------|----------|----------|
| R1 | 重构过程中可能引入 bug | 每次修改后验证功能正常 |
| R2 | 改动量大，影响面广 | 按任务清单顺序逐步执行 |
| R3 | 移除 file/db.go 后原有功能受影响 | 确保 GORM 版本已覆盖所有功能 |

## Migration Plan

1. 创建 Makefile（已完成）
2. 创建 internal/response/ 包
3. 统一数据库层，移除 internal/file/db.go
4. 拆分 handlers.go
5. 添加 service 层，使用依赖注入
6. 修复安全问题
7. 完善日志
8. 验证功能正常
