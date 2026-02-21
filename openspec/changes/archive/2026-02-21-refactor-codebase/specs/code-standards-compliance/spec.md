## ADDED Requirements

### Requirement: 代码格式符合 gofmt 和 goimports 标准
所有 Go 代码必须通过 gofmt 和 goimports 格式化工具的处理，确保代码风格一致。

#### Scenario: 格式化检查
- **WHEN** 开发人员运行 gofmt 或 goimports
- **THEN** 代码无需任何修改即可通过

### Requirement: 导入顺序遵循规范
导入语句必须按标准库、第三方包、项目内部包的顺序排列，组间用空行分隔。

#### Scenario: 导入顺序验证
- **WHEN** 代码通过格式化检查
- **THEN** 导入顺序符合规范要求

### Requirement: 命名规范一致
变量和函数使用驼峰命名，JSON 标签使用下划线命名，数据库字段使用下划线命名。

#### Scenario: 命名验证
- **WHEN** 代码通过 lint 检查
- **THEN** 命名符合规范要求

### Requirement: 导出函数有注释
所有导出的函数和类型必须有注释说明其用途。

#### Scenario: 注释检查
- **WHEN** 代码审查时检查导出函数
- **THEN** 每个导出函数都有对应的注释说明
