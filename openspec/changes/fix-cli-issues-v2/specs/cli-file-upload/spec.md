## ADDED Requirements

### Requirement: 上传前检查重名文件
上传文件前应检查目标目录是否已存在同名文件。

#### Scenario: 目标目录存在同名文件
- **WHEN** 上传文件时，目标目录已存在同名文件
- **THEN** 返回错误提示用户

#### Scenario: 目标目录无同名文件
- **WHEN** 上传文件时，目标目录无同名文件
- **THEN** 正常上传

### Requirement: 统一配置路径
CLI 配置文件路径统一到 ./data/config

#### Scenario: 使用项目配置
- **WHEN** 存在 ./data/config/config.yaml
- **THEN** 使用该配置文件

#### Scenario: 使用用户配置
- **WHEN** 不存在项目配置但存在 ~/.config/claw-pliers/config.yaml
- **THEN** 使用用户配置文件
