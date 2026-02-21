## 1. 基础设施

- [x] 1.1 创建 Makefile（docker 开发环境）
- [x] 1.2 添加 .gitignore 文件

## 2. 响应格式统一

- [x] 2.1 创建 internal/response/response.go 统一响应函数
- [ ] 2.2 更新现有 handlers 使用新的响应函数

## 3. 数据库层统一

- [x] 3.1 确认 internal/database/database.go 功能完整
- [x] 3.2 移除 internal/file/db.go
- [x] 3.3 更新 internal/file/init.go 使用统一数据库

## 4. Handlers 拆分

- [x] 4.1 创建 internal/api/file_handler.go（从 handlers.go 拆分）
- [x] 4.2 创建 internal/api/folder_handler.go
- [x] 4.3 创建 internal/api/mail_handler.go
- [ ] 4.4 移除原 internal/api/handlers.go（更新使用 response 包后）

## 5. 安全修复

- [x] 5.1 创建 internal/utils/random.go 安全的随机数生成工具
- [ ] 5.2 替换 handlers 中的不安全随机数生成

## 6. Service 层创建

- [ ] 6.1 创建 internal/service/file_service.go
- [ ] 6.2 创建 internal/service/folder_service.go
- [ ] 6.3 创建 internal/service/mail_service.go

## 7. 依赖注入重构

- [ ] 7.1 重构 FileHandler 使用依赖注入
- [ ] 7.2 重构 FolderHandler 使用依赖注入
- [ ] 7.3 重构 MailHandler 使用依赖注入
- [ ] 7.4 更新 router.go 组装依赖

## 8. 日志完善

- [ ] 8.1 在关键操作添加日志记录
- [ ] 8.2 统一日志格式

## 9. 验证

- [ ] 9.1 运行 make build 确保编译通过
- [ ] 9.2 运行 make run 测试服务启动
- [ ] 9.3 测试基本 API 功能
