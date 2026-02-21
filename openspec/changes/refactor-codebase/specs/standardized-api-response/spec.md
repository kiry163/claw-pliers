## ADDED Requirements

### Requirement: 统一 API 响应格式
所有 API 响应使用统一的 JSON 格式，包含 code、message、data 字段。

#### Scenario: 成功响应格式
- **WHEN** API 请求成功
- **THEN** 返回 `{"code": 0, "message": "success", "data": {...}}`

#### Scenario: 错误响应格式
- **WHEN** API 请求失败
- **THEN** 返回 `{"code": <错误码>, "message": "<错误消息>"}`

### Requirement: 响应辅助函数
提供统一的响应函数处理成功、错误、消息、分页等场景。

#### Scenario: 使用响应函数
- **WHEN** 处理器返回响应
- **THEN** 使用 response.Success、response.Error、response.Message、response.Page 等函数

### Requirement: 响应包位置
响应辅助函数位于 internal/response/ 包中。

#### Scenario: 响应包结构
- **WHEN** 查看项目结构
- **THEN** 存在 internal/response/response.go 文件

### Requirement: 错误码定义
定义基础错误码体系，用于区分不同类型的错误。

#### Scenario: 错误码使用
- **WHEN** 返回错误响应
- **THEN** 使用预定义的错误码（如 10001=未授权，10004=参数错误）
