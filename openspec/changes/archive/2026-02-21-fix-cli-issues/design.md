## Context

CLI 工具测试发现以下问题需要修复：

1. file put 返回 404 但 curl 正常
2. mail.go 硬编码了错误的密钥

## Goals / Non-Goals

**Goals:**
- 修复 file put 上传问题
- 修复 mail 认证密钥问题
- 确保 CLI 行为一致

**Non-Goals:**
- 不实现 image 功能（当前无需求）
- 不修改 API 服务端

## Decisions

### D1: file put 问题排查

**分析**:
- curl 直接调用 API 正常
- file get 下载正常
- file ls 正常

**可能原因**:
- 请求 URL 构造有问题
- 需要添加调试日志排查

**决策**: 添加调试日志，输出实际请求的 URL 和参数

### D2: mail 密钥问题

**位置**: `cli/mail.go:467`

**当前代码**:
```go
localKey := "test-local-key-change-me"
```

**问题**: 与服务器配置 `change-me-in-production` 不匹配

**决策**: 
- 移除硬编码
- 使用与 file.go 相同的配置加载逻辑

### D3: image 命令处理

**当前状态**: stub 实现

**决策**: 保持现状，后续有需求再实现

## Implementation Plan

1. **file put 调试**
   - 添加请求日志
   - 检查 URL 构造
   - 验证参数传递

2. **修复 mail 密钥**
   - 复用 file.go 的 loadConfig() 函数
   - 或统一配置加载逻辑

3. **验证**
   - 测试 file put 上传
   - 测试 mail 命令认证
