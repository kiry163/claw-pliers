## ADDED Requirements

### Requirement: 子目录列表显示
列出子目录内容时应正确显示所有文件和子目录。

#### Scenario: 列出子目录
- **WHEN** 执行 `claw-pliers-cli file ls claw:/f1` 其中 f1 包含子目录 f2
- **THEN** 显示 f2 子目录，不显示 empty
