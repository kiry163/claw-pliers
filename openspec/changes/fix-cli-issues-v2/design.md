# Design: fix-cli-issues-v2

## Context

CLI 工具存在多个问题需要修复，包括文件上传、目录遍历、配置加载等。

## Goals

- 修复文件重名上传问题
- 修复子目录 ls 显示
- 修复 mail list 配置读取
- 统一配置路径到 ./data/config

## Decisions

### D1: 配置路径变更

**当前**: `./config/config.yaml`
**目标**: `./data/config/config.yaml`

**实现**:
1. 修改 cli/file.go 中 loadConfig() 函数
2. 搜索顺序: `./data/config/config.yaml` → `~/.config/claw-pliers/config.yaml` → 默认

### D2: 文件重名检查

**问题**: 上传文件时未检查目标目录是否已存在同名文件

**实现**:
1. 上传前调用 API 获取目录文件列表
2. 检查目标目录是否包含同名文件
3. 如存在，返回错误提示用户选择其他名称或删除旧文件

### D3: 子目录 ls 修复

**问题**: `claw-pliers-cli file ls claw:/f1` 显示 empty，但实际有子目录 f2

**分析**: URL 构造问题，`/f1/` vs `/f1` 导致 API 返回不同结果

**实现**:
1. 规范化路径处理
2. 确保目录路径不带尾部斜杠时能正确查询

### D4: mail list 配置读取

**问题**: mail list 显示 "No accounts configured"，但配置文件有账户

**分析**: 可能是 API 调用问题或配置传递问题

**实现**:
1. 检查 API 调用是否正确传递认证
2. 验证服务器端账户列表接口

## Implementation Plan

1. 修改 loadConfig() 配置路径
2. 添加文件重名检查逻辑
3. 修复 ls 路径处理
4. 调试 mail list 接口
5. 全面测试
