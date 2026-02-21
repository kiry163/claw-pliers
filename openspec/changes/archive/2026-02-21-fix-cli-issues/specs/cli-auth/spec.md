## ADDED Requirements

### Requirement: 从配置文件加载认证密钥
CLI 应该从配置文件加载认证密钥，而不是硬编码。

#### Scenario: 使用配置文件
- **WHEN** CLI 启动时
- **THEN** 从 `./config/config.yaml` 或 `~/.config/claw-pliers/config.yaml` 加载密钥

#### Scenario: 使用环境变量
- **WHEN** 设置了 `CLAWPLIERS_AUTH_LOCAL_KEY` 环境变量
- **THEN** 使用环境变量中的密钥（优先级最高）

### Requirement: 移除硬编码密钥
移除所有硬编码的默认密钥。

#### Scenario: 无有效密钥
- **WHEN** 没有配置文件且没有设置环境变量
- **THEN** 返回错误提示用户配置密钥

### Requirement: mail 命令使用正确的认证
所有 mail 命令应该使用正确的认证密钥。

#### Scenario: mail send
- **WHEN** 执行 `claw-pliers-cli mail send`
- **THEN** 使用正确的认证密钥进行请求
