## ADDED Requirements

### Requirement: 使用加密安全的随机数生成器
生成令牌、ID 等使用 crypto/rand 而不是时间戳。

#### Scenario: 生成随机令牌
- **WHEN** 需要生成随机令牌（如分享链接 token）
- **THEN** 使用 crypto/rand 库生成加密安全的随机数

### Requirement: 文件 ID 生成
文件 ID 使用安全的随机数生成，确保唯一性。

#### Scenario: 生成文件 ID
- **WHEN** 上传新文件需要生成 file_id
- **THEN** 使用 crypto/rand 生成安全的唯一 ID

### Requirement: 禁止使用不安全的随机数方法
禁止使用 math/rand 或基于时间戳的随机数生成。

#### Scenario: 代码审查
- **WHEN** 检查代码中的随机数生成
- **THEN** 不存在使用 time.Now().UnixNano() 生成随机数的情况

### Requirement: 随机数生成函数封装
提供统一的随机数生成工具函数。

#### Scenario: 使用工具函数
- **WHEN** 需要生成随机字符串
- **THEN** 调用统一的工具函数，如 generateToken()
