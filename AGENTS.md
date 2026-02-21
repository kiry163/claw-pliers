# AGENTS.md - Claw Pliers 开发指南

> **重要**：在进行任何代码工作之前，必须先阅读 `CODING_STANDARDS.md` 规范文档，确保代码符合项目规范。

## OpenSpec 工作流程

**重要**：所有项目变更必须通过 OpenSpec 管理，禁止直接修改代码。

### 工作流程

1. **探索模式** - 进入探索模式讨论问题
   ```
   /opsx-explore [问题描述]
   ```

2. **创建 Change** - 确认需要实施后，创建正式的 change
   ```
   openspec new change <change-name>
   ```

3. **按 Artifact 执行** - 按照 proposal → design → tasks → implementation 的顺序执行
   ```
   openspec continue-change  # 继续下一个 artifact
   ```

4. **验证** - 实现完成后验证
   ```
   openspec verify-change
   ```

5. **归档** - 合并到主分支后归档
   ```
   openspec archive-change <change-name>
   ```

### Change 命名规范

使用 kebab-case，如：
- `refactor-codebase` - 代码重构
- `add-file-upload` - 新功能
- `fix-auth-bug` - Bug 修复

## 项目规范

遵循 `CODING_STANDARDS.md` 中的规范：

- 响应格式统一
- 依赖注入
- 日志记录
- 错误处理
- 注释要求

## 开发环境规范

### 1. Docker 开发环境

- **不要**在本机上直接运行服务端程序，必须使用 Docker 开发环境
- 使用 `make dev-up` 启动开发环境
- 使用 `make dev-down` 停止开发环境

### 2. Docker 构建规范

- 修改代码后重新构建 Docker 镜像时，**不要**使用 `--no-cache` 参数，这会导致构建变得很慢
- 使用默认的 `docker build` 命令即可

### 3. 网络代理

国内环境部分镜像拉取很慢，需要使用代理：
```bash
export https_proxy=http://127.0.0.1:7890 http_proxy=http://127.0.0.1:7890 all_proxy=socks5://127.0.0.1:7890
```

## 常用命令

```bash
# 启动开发环境
make dev-up

# 停止开发环境
make dev-down

# 查看开发环境日志
make dev-logs

# 重启开发环境
make dev-restart

# 构建项目
make build

# 运行测试
make test
```

## OpenSpec 常用命令

```bash
# 查看变更列表
openspec list

# 查看变更状态
openspec status --change <change-name>
```
