# FileHub 路径格式统一方案

## 背景

当前系统存在两种文件引用方式：
1. **Web端**: 使用 `filehub://file_id` 作为文件URL
2. **CLI**: 正在从 folder_id 方式向路径方式迁移

**问题**：
- `filehub://T40nRdxFqnyG` 难以记忆和操作
- 不适合 AI Agent 管理私人云盘
- CLI 与 Web 端不一致

## 目标

统一使用 **路径格式** 引用文件和文件夹：

```
filehub:/path/to/file
```

### 格式规范

| 类型 | 格式 | 示例 |
|------|------|------|
| 根目录 | `filehub:/` | `filehub:/` |
| 文件夹 | `filehub:/path/` | `filehub:/documents/`, `filehub:/backup/photos/` |
| 文件 | `filehub:/path/file.txt` | `filehub:/documents/report.pdf` |

---

## 实施计划

### Phase 1: 服务端 - 路径API

#### 1.1 数据库层 (internal/db/db.go)

| 方法 | 说明 |
|------|------|
| `GetFilePath(ctx, fileID)` | 根据 fileID 获取完整路径 |
| `GetFolderPath(ctx, folderID)` | 根据 folderID 获取完整路径 |

实现思路：递归查询父文件夹，拼接完整路径。

#### 1.2 API Handlers (internal/api/handlers.go)

| Handler | 路由 | 说明 |
|---------|------|------|
| `GetFileByPath` | `GET /api/v1/files/by-path?path=...` | 通过路径获取文件信息 |
| `DownloadFileByPath` | `GET /api/v1/files/by-path/download?path=...` | 通过路径下载文件 |

#### 1.3 路由 (internal/api/router.go)

```go
files.GET("/by-path", handler.GetFileByPath)
files.GET("/by-path/download", handler.DownloadFileByPath)
```

---

### Phase 2: 服务端 - 返回格式统一

#### 2.1 修改 handlers.go

将所有 `filehub_url` 改为路径格式：

```go
// 修改前
"filehub_url": "filehub://" + record.FileID

// 修改后
"filehub_url": "filehub:/" + filePath  // 例如: filehub:/documents/a.txt
```

需要修改的函数：
- `GetFile` (handlers.go:126)
- `ListFiles` (handlers.go:144)
- `UploadFile` (handlers.go:89)
- `GetFolderContents` (folder_handlers.go:294)

#### 2.2 修改 folder_handlers.go

```go
// 修改前
"filehub_url": "filehub://" + file.FileID

// 修改后
"filehub_url": "filehub:/" + filePath
```

需要修改的函数：
- `GetFolderContents` (folder_handlers.go:294)
- `ShareFile` (folder_handlers.go:669)

#### 2.3 添加路径计算辅助函数

```go
func (h *Handler) buildFilePath(ctx context.Context, file FileRecord) string
func (h *Handler) buildFolderPath(ctx context.Context, folder FolderRecord) string
```

---

### Phase 3: Web端改造

#### 3.1 修改分享链接格式

文件分享时，生成的链接改为路径格式：

```javascript
// 修改前
const url = `filehub://${file.file_id}`

// 修改后
const url = `filehub:/${file.path}`
```

#### 3.2 修改 view_url 字段

```javascript
// 修改前
view_url: `/app/files/${file.file_id}`

// 修改后
view_url: `/app/f${file.path}`
```

#### 3.3 路由改造 (web-ui/src/router/index.js)

支持路径访问：

```javascript
// 新增路由
'/f/*path'  // 文件/文件夹访问
```

#### 3.4 文件详情页改造

根据路径加载文件/文件夹内容。

---

### Phase 4: CLI改造

#### 4.1 路径解析 (internal/cli/commands_helpers.go)

```go
// 修改路径解析函数
func parseFilehubURL(value string) (string, error) {
    // 支持两种格式:
    // 1. filehub:/path/to/file (新格式)
    // 2. filehub://file_id (旧格式，兼容)
    
    if strings.HasPrefix(value, "filehub:/") {
        path := strings.TrimPrefix(value, "filehub:/")
        return resolvePath(value)  // 调用API获取file_id
    }
    
    if strings.HasPrefix(value, "filehub://") {
        // 旧格式，直接返回
        return strings.TrimPrefix(value, "filehub://"), nil
    }
    
    // 裸路径，视为文件路径
    return resolvePath("filehub:/" + value)
}
```

#### 4.2 添加 Client 方法

```go
// GetFileByPath 通过路径获取文件信息
func (c *Client) GetFileByPath(path string) (*FileItem, error)

// DownloadFileByPath 通过路径下载文件
func (c *Client) DownloadFileByPath(path, outputPath string) error
```

#### 4.3 更新命令支持

所有命令自动支持路径格式：

```bash
# 上传
filehub put local.txt filehub:/documents/
filehub put local.txt filehub:/documents/a.txt

# 下载
filehub get filehub:/documents/a.txt
filehub get filehub:/documents/a.txt ./local.txt

# 列出
filehub ls filehub:/documents/

# 查看
filehub info filehub:/documents/a.txt
filehub info filehub:/documents/

# 移动
filehub mv filehub:/docs/a.txt filehub:/backup/
filehub mv filehub:/docs/a.txt filehub:/backup/b.txt

# 删除
filehub rm filehub:/docs/a.txt
filehub rm filehub:/docs/ -r
```

---

## API 设计

### 通过路径获取文件信息

**请求**
```
GET /api/v1/files/by-path?path=/documents/a.txt
Authorization: Bearer <token>
```

**响应**
```json
{
  "code": 0,
  "data": {
    "file_id": "T40nRdxFqnyG",
    "original_name": "a.txt",
    "path": "/documents/a.txt",
    "filehub_url": "filehub:/documents/a.txt",
    "size": 1024,
    "mime_type": "text/plain",
    "created_at": "2026-02-13T12:00:00Z",
    "download_url": "http://localhost:8080/api/v1/files/by-path/download?path=/documents/a.txt"
  }
}
```

### 通过路径下载文件

**请求**
```
GET /api/v1/files/by-path/download?path=/documents/a.txt
Authorization: Bearer <token>
```

**响应**: 文件二进制流

---

## 数据库变更

无需新增字段，路径通过实时计算获取：

```go
// 根据 fileID 计算路径
func (db *DB) GetFilePath(ctx context.Context, fileID string) (string, error) {
    file, err := db.GetFile(ctx, fileID)
    if err != nil {
        return "", err
    }
    
    if file.FolderID == nil {
        return "/" + file.OriginalName, nil
    }
    
    folderPath, err := db.GetFolderPath(ctx, *file.FolderID)
    if err != nil {
        return "", err
    }
    
    return folderPath + "/" + file.OriginalName, nil
}

// 根据 folderID 计算路径
func (db *DB) GetFolderPath(ctx context.Context, folderID string) (string, error) {
    folder, err := db.GetFolder(ctx, folderID)
    if err != nil {
        return "", err
    }
    
    if folder.ParentID == nil {
        return "/" + folder.Name, nil
    }
    
    parentPath, err := db.GetFolderPath(ctx, *folder.ParentID)
    if err != nil {
        return "", err
    }
    
    return parentPath + "/" + folder.Name, nil
}
```

---

## 兼容性考虑

1. **旧格式兼容**: 保留 `filehub://file_id` 格式支持
2. **数据库**: 无需迁移，现有数据继续使用
3. **Web端**: 渐进式改造，先改API返回格式，再改前端

---

## 测试计划

### 单元测试
- [ ] `GetFilePath` 路径计算正确性
- [ ] `GetFolderPath` 路径计算正确性
- [ ] 路径解析兼容性

### 集成测试
- [ ] 通过路径上传文件
- [ ] 通过路径下载文件
- [ ] 通过路径查看文件信息
- [ ] 通过路径移动/删除文件
- [ ] CLI 路径命令测试

---

## 进度追踪

| 阶段 | 状态 | 负责人 |
|------|------|--------|
| Phase 1: 服务端路径API | TODO | - |
| Phase 2: 返回格式统一 | TODO | - |
| Phase 3: Web端改造 | TODO | - |
| Phase 4: CLI改造 | TODO | - |

---

## 参考

- 当前系统设计: `filehub-design.md`
- CLI命令文档: 运行 `filehub-cli --help`
