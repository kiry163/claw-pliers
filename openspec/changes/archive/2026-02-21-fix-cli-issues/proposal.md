# Proposal: fix-cli-issues

## Why

CLI 工具存在以下问题需要修复：

1. **file put 上传失败** - 返回 404 Not Found，但通过 curl 直接调用 API 正常
2. **mail.go 认证密钥硬编码错误** - 使用 `test-local-key-change-me`，实际应为 `change-me-in-production`
3. **image 命令是存根** - convert 和 ocr 未实现

## What Changes

1. **修复 file put 上传问题**
   - 排查 CLI 端请求构造问题
   - 确保 endpoint 和 key 正确传递

2. **修复 mail 认证密钥**
   - 移除硬编码的密钥
   - 从配置文件或环境变量加载

3. **image 命令说明**
   - 保持存根状态（当前无需求）
   - 或移除未实现的命令

## Capabilities

### Modified Capabilities

- `cli-file-upload`: 修复文件上传功能
- `cli-auth`: 修复认证密钥问题

## Impact

- **API 接口**: 无变化
- **依赖**: 无新增依赖
- **配置**: 统一使用配置文件中的认证密钥
