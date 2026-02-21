## ADDED Requirements

### Requirement: 配置路径变更
配置文件从 ./config 移动到 ./data/config

#### Scenario: 加载配置文件
- **WHEN** CLI 启动时
- **THEN** 按顺序搜索: ./data/config/config.yaml → ~/.config/claw-pliers/config.yaml

### Requirement: 删除旧配置
删除 ./config 目录，不提交到 git

#### Scenario: 清理旧配置
- **WHEN** 项目根目录存在 ./config
- **THEN** 删除该目录
