## ADDED Requirements

### Requirement: file put 命令能够成功上传文件
file put 命令应该能够成功上传文件到服务器。

#### Scenario: 正常上传
- **WHEN** 执行 `claw-pliers-cli file put <local> claw:/<path>`
- **THEN** 文件成功上传，返回上传成功信息

#### Scenario: 上传失败
- **WHEN** 服务器返回错误
- **THEN** 显示详细的错误信息

### Requirement: 上传进度显示
上传文件时应该显示进度百分比。

#### Scenario: 显示进度
- **WHEN** 上传文件过程中
- **THEN** 显示进度百分比（如 "Progress: 50%"）
