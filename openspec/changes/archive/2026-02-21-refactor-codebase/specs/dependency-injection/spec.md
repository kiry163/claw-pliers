## ADDED Requirements

### Requirement: Handler 依赖注入
HTTP Handler 通过构造函数注入依赖，不使用全局变量。

#### Scenario: 创建 Handler
- **WHEN** 创建新的 Handler 实例
- **THEN** 通过构造函数传入 service、storage、config 等依赖

### Requirement: Service 层依赖注入
Service 层通过构造函数注入 Repository 和其他依赖。

#### Scenario: 创建 Service
- **WHEN** 创建新的 Service 实例
- **THEN** 通过构造函数传入 repository、storage 等依赖

### Requirement: 移除全局变量
业务逻辑代码中不使用全局变量存储状态。

#### Scenario: 代码审查
- **WHEN** 检查业务代码
- **THEN** 不存在用于存储业务状态的全局变量（配置、日志等除外）

### Requirement: Service 层创建
创建 internal/service/ 目录，包含业务逻辑层。

#### Scenario: 项目结构
- **WHEN** 查看 internal/ 目录
- **THEN** 存在 internal/service/ 目录，包含 file_service.go、folder_service.go 等

### Requirement: 依赖组装
在路由初始化时组装所有依赖。

#### Scenario: 应用启动
- **WHEN** 应用启动并初始化路由
- **THEN** 所有依赖被正确组装并注入到 Handler 中
