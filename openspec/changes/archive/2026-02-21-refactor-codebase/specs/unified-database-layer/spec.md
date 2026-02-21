## ADDED Requirements

### Requirement: 统一使用 GORM 数据库层
项目只使用一个数据库连接，全部通过 GORM 进行操作。

#### Scenario: 数据库初始化
- **WHEN** 应用启动时
- **THEN** 使用 GORM 初始化数据库连接，并自动迁移模型

### Requirement: 移除重复的数据库实现
移除 internal/file/db.go 中的原始 sql 实现，只保留 GORM 版本。

#### Scenario: 代码清理
- **WHEN** 重构完成后
- **THEN** internal/file/db.go 文件已被删除，所有数据库操作通过 internal/database/ 进行

### Requirement: 数据库模型简化
GORM 模型定义不添加 size、not null 等约束，依赖自动迁移。

#### Scenario: 模型定义
- **WHEN** 定义新的数据模型
- **THEN** 使用简化的 GORM 标签，如 `gorm:"column:file_id;uniqueIndex"`

### Requirement: 数据库操作使用 Repository 模式
数据库操作封装在 Repository 接口中，便于测试和替换实现。

#### Scenario: 数据访问
- **WHEN** 业务代码需要访问数据库
- **THEN** 通过 Repository 接口进行，不直接操作 gorm.DB
