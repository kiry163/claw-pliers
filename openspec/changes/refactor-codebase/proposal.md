# Proposal: refactor-codebase

## Why

项目当前代码存在多个问题：数据库层重复实现、API 响应格式不统一、单文件代码过长、安全隐患（不安全的随机数）、缺少依赖注入和日志记录。按照已制定的 `CODING_STANDARDS.md` 规范重构代码，提升代码可维护性和安全性。

## What Changes

1. **统一数据库层** - 移除 `internal/file/db.go`，保留 `internal/database/database.go` (GORM 版本)
2. **创建响应格式统一模块** - 新建 `internal/response/` 包
3. **拆分 handlers** - 按功能拆分为 `file_handler.go`, `folder_handler.go`, `mail_handler.go`
4. **修复安全问题** - 使用 `crypto/rand` 替换不安全的随机数生成
5. **创建 Service 层** - 新建 `internal/service/` 业务逻辑层
6. **重构依赖注入** - Handler 使用构造函数依赖注入
7. **完善日志记录** - 添加更详细的操作日志
8. **添加 .gitignore** - 忽略构建产物和临时文件

## Capabilities

### New Capabilities

- `code-standards-compliance`: 代码符合 CODING_STANDARDS.md 规范
- `unified-database-layer`: 统一的 GORM 数据库层
- `standardized-api-response`: 标准化的 API 响应格式
- `secure-random-generation`: 安全的随机数生成
- `dependency-injection`: 依赖注入架构

### Modified Capabilities

- (无 - 新功能不涉及现有需求变更)

## Impact

- **代码结构**: 目录结构调整，新增 `internal/service/`、`internal/response/` 目录
- **API 接口**: 响应格式统一为 `{"code": 0, "message": "success", "data": ...}`
- **依赖**: 无新增依赖，使用现有库
- **配置**: 新增 `Makefile`，开发流程更便捷
