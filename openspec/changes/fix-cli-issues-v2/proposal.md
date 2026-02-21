# Proposal: fix-cli-issues-v2

## Why

CLI 工具存在以下问题需要修复：
1. 上传文件时允许同一目录下存在重名文件
2. 子目录 ls 显示异常 (claw:/f1 显示 empty，实际有内容)
3. mail list 未正确读取配置文件中的账户
4. 配置文件路径分散，需要统一到 ./data/config
5. 清理旧数据和 ./config 目录

## What Changes

1. **防止文件重名上传** - 上传前检查目标目录是否存在同名文件
2. **修复子目录 ls** - 修复目录遍历逻辑，正确显示子目录内容
3. **修复 mail list** - 从配置文件正确加载账户列表
4. **统一配置路径** - 配置移动到 ./data/config，删除 ./config 目录
5. **清理旧数据** - 删除数据库旧数据避免冲突

## Capabilities

### Modified Capabilities

- `cli-file-upload`: 修复文件上传逻辑，添加重名检查
- `cli-file-list`: 修复子目录遍历问题
- `cli-auth`: 修复配置文件加载路径

## Impact

- 配置文件位置变更: ./config → ./data/config
- 需要删除旧数据库数据
- CLI 需要重新配置认证
