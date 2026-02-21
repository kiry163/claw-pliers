package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	sql *sql.DB
}

// FolderRecord 文件夹记录
type FolderRecord struct {
	FolderID  string
	Name      string
	ParentID  *string // nil 表示根目录
	CreatedBy string
	CreatedAt string
	UpdatedAt string
}

type FileRecord struct {
	FileID       string
	OriginalName string
	ObjectKey    string
	Size         int64
	MimeType     string
	FolderID     *string // nil 表示根目录
	CreatedBy    string
	CreatedAt    string
	UpdatedAt    string
}

type RefreshToken struct {
	Token     string
	ExpiresAt string
	IsRevoked bool
}

type ShareLink struct {
	Token     string
	FileID    string
	ExpiresAt string
	CreatedAt string
	CreatedBy string
	Status    string
}

func Open(path string) (*DB, error) {
	handle, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	handle.SetMaxOpenConns(1)
	db := &DB{sql: handle}
	if err := db.migrate(); err != nil {
		return nil, err
	}
	return db, nil
}

func (db *DB) migrate() error {
	statements := []string{
		// 文件夹表
		`CREATE TABLE IF NOT EXISTS folders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			folder_id VARCHAR(12) UNIQUE NOT NULL,
			name VARCHAR(255) NOT NULL,
			parent_id VARCHAR(12),
			created_by VARCHAR(64) NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			UNIQUE(name, parent_id)
		);`,
		// 文件表（添加 folder_id 字段）
		`CREATE TABLE IF NOT EXISTS files (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_id VARCHAR(12) UNIQUE NOT NULL,
			original_name VARCHAR(255) NOT NULL,
			object_key VARCHAR(512) NOT NULL,
			size BIGINT NOT NULL,
			mime_type VARCHAR(100),
			folder_id VARCHAR(12),
			created_by VARCHAR(64) NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			metadata JSON
		);`,
		`CREATE TABLE IF NOT EXISTS refresh_tokens (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			token VARCHAR(128) UNIQUE NOT NULL,
			expires_at DATETIME NOT NULL,
			is_revoked BOOLEAN DEFAULT FALSE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS audit_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			action VARCHAR(50) NOT NULL,
			file_id VARCHAR(32),
			actor VARCHAR(64) NOT NULL,
			ip_address VARCHAR(45),
			status VARCHAR(20),
			message TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS share_links (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			token VARCHAR(64) UNIQUE NOT NULL,
			file_id VARCHAR(32) NOT NULL,
			expires_at DATETIME NOT NULL,
			created_at DATETIME NOT NULL,
			created_by VARCHAR(64) NOT NULL,
			status VARCHAR(20) NOT NULL
		);`,
		// 索引
		`CREATE INDEX IF NOT EXISTS idx_share_links_file_id ON share_links(file_id);`,
		`CREATE INDEX IF NOT EXISTS idx_share_links_token ON share_links(token);`,
		`CREATE INDEX IF NOT EXISTS idx_files_created_by ON files(created_by);`,
		`CREATE INDEX IF NOT EXISTS idx_files_created_at ON files(created_at);`,
		`CREATE INDEX IF NOT EXISTS idx_files_folder_id ON files(folder_id);`,
		`CREATE INDEX IF NOT EXISTS idx_folders_parent_id ON folders(parent_id);`,
	}

	for _, stmt := range statements {
		if _, err := db.sql.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) CreateFile(ctx context.Context, record FileRecord) error {
	_, err := db.sql.ExecContext(
		ctx,
		`INSERT INTO files (file_id, original_name, object_key, size, mime_type, created_by, created_at, updated_at)
     VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		record.FileID,
		record.OriginalName,
		record.ObjectKey,
		record.Size,
		record.MimeType,
		record.CreatedBy,
		record.CreatedAt,
		record.UpdatedAt,
	)
	return err
}

func (db *DB) GetFile(ctx context.Context, fileID string) (FileRecord, error) {
	var record FileRecord
	row := db.sql.QueryRowContext(ctx, `
    SELECT file_id, original_name, object_key, size, mime_type, folder_id, created_by, created_at, updated_at
    FROM files WHERE file_id = ?`, fileID)
	if err := row.Scan(
		&record.FileID,
		&record.OriginalName,
		&record.ObjectKey,
		&record.Size,
		&record.MimeType,
		&record.FolderID,
		&record.CreatedBy,
		&record.CreatedAt,
		&record.UpdatedAt,
	); err != nil {
		return FileRecord{}, err
	}
	return record, nil
}

// GetFileByName 通过文件名和文件夹ID获取文件
func (db *DB) GetFileByName(ctx context.Context, name string, folderID *string) (FileRecord, error) {
	var record FileRecord
	var row *sql.Row
	if folderID == nil {
		row = db.sql.QueryRowContext(ctx, `
			SELECT file_id, original_name, object_key, size, mime_type, folder_id, created_by, created_at, updated_at
			FROM files WHERE original_name = ? AND folder_id IS NULL`, name)
	} else {
		row = db.sql.QueryRowContext(ctx, `
			SELECT file_id, original_name, object_key, size, mime_type, folder_id, created_by, created_at, updated_at
			FROM files WHERE original_name = ? AND folder_id = ?`, name, *folderID)
	}
	if err := row.Scan(
		&record.FileID,
		&record.OriginalName,
		&record.ObjectKey,
		&record.Size,
		&record.MimeType,
		&record.FolderID,
		&record.CreatedBy,
		&record.CreatedAt,
		&record.UpdatedAt,
	); err != nil {
		return FileRecord{}, err
	}
	return record, nil
}

func (db *DB) ListFiles(ctx context.Context, limit, offset int, order, keyword string) ([]FileRecord, int, error) {
	if order != "asc" {
		order = "desc"
	}
	var total int
	countQuery := "SELECT COUNT(1) FROM files"
	args := []interface{}{}
	if keyword != "" {
		countQuery += " WHERE original_name LIKE ?"
		args = append(args, "%"+keyword+"%")
	}
	if err := db.sql.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
    SELECT file_id, original_name, object_key, size, mime_type, folder_id, created_by, created_at, updated_at
    FROM files`
	args = []interface{}{}
	if keyword != "" {
		query += " WHERE original_name LIKE ?"
		args = append(args, "%"+keyword+"%")
	}
	query += fmt.Sprintf(" ORDER BY created_at %s LIMIT ? OFFSET ?", order)
	args = append(args, limit, offset)
	rows, err := db.sql.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	records := make([]FileRecord, 0)
	for rows.Next() {
		var record FileRecord
		if err := rows.Scan(
			&record.FileID,
			&record.OriginalName,
			&record.ObjectKey,
			&record.Size,
			&record.MimeType,
			&record.FolderID,
			&record.CreatedBy,
			&record.CreatedAt,
			&record.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		records = append(records, record)
	}
	return records, total, nil
}

func (db *DB) AddAuditLog(ctx context.Context, action, fileID, actor, ipAddress, status, message string) error {
	_, err := db.sql.ExecContext(
		ctx,
		`INSERT INTO audit_logs (action, file_id, actor, ip_address, status, message) VALUES (?, ?, ?, ?, ?, ?)`,
		action,
		fileID,
		actor,
		ipAddress,
		status,
		message,
	)
	return err
}

func (db *DB) DeleteFile(ctx context.Context, fileID string) (FileRecord, error) {
	record, err := db.GetFile(ctx, fileID)
	if err != nil {
		return FileRecord{}, err
	}
	_, err = db.sql.ExecContext(ctx, `DELETE FROM files WHERE file_id = ?`, fileID)
	if err != nil {
		return FileRecord{}, err
	}
	return record, nil
}

func (db *DB) CreateRefreshToken(ctx context.Context, token string, expiresAt string) error {
	_, err := db.sql.ExecContext(
		ctx,
		`INSERT INTO refresh_tokens (token, expires_at, is_revoked) VALUES (?, ?, false)`,
		token,
		expiresAt,
	)
	return err
}

func (db *DB) GetRefreshToken(ctx context.Context, token string) (RefreshToken, error) {
	var record RefreshToken
	row := db.sql.QueryRowContext(ctx, `
    SELECT token, expires_at, is_revoked
    FROM refresh_tokens WHERE token = ?`, token)
	if err := row.Scan(&record.Token, &record.ExpiresAt, &record.IsRevoked); err != nil {
		return RefreshToken{}, err
	}
	return record, nil
}

func (db *DB) RevokeRefreshToken(ctx context.Context, token string) error {
	_, err := db.sql.ExecContext(ctx, `UPDATE refresh_tokens SET is_revoked = true WHERE token = ?`, token)
	return err
}

func (db *DB) RevokeAllRefreshTokens(ctx context.Context) error {
	_, err := db.sql.ExecContext(ctx, `UPDATE refresh_tokens SET is_revoked = true`)
	return err
}

func (db *DB) CreateShareLink(ctx context.Context, link ShareLink) error {
	_, err := db.sql.ExecContext(
		ctx,
		`INSERT INTO share_links (token, file_id, expires_at, created_at, created_by, status)
	 VALUES (?, ?, ?, ?, ?, ?)`,
		link.Token,
		link.FileID,
		link.ExpiresAt,
		link.CreatedAt,
		link.CreatedBy,
		link.Status,
	)
	return err
}

func (db *DB) GetShareLink(ctx context.Context, token string) (ShareLink, error) {
	var link ShareLink
	row := db.sql.QueryRowContext(ctx, `
    SELECT token, file_id, expires_at, created_at, created_by, status
    FROM share_links WHERE token = ?`, token)
	if err := row.Scan(
		&link.Token,
		&link.FileID,
		&link.ExpiresAt,
		&link.CreatedAt,
		&link.CreatedBy,
		&link.Status,
	); err != nil {
		return ShareLink{}, err
	}
	return link, nil
}

func (db *DB) GetActiveShareLink(ctx context.Context, fileID, nowRFC3339 string) (ShareLink, error) {
	var link ShareLink
	row := db.sql.QueryRowContext(ctx, `
    SELECT token, file_id, expires_at, created_at, created_by, status
    FROM share_links
    WHERE file_id = ? AND status = 'active' AND expires_at > ?
    ORDER BY created_at DESC
    LIMIT 1`, fileID, nowRFC3339)
	if err := row.Scan(
		&link.Token,
		&link.FileID,
		&link.ExpiresAt,
		&link.CreatedAt,
		&link.CreatedBy,
		&link.Status,
	); err != nil {
		return ShareLink{}, err
	}
	return link, nil
}

func NowRFC3339() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// ==================== 文件夹操作 ====================

// CreateFolder 创建文件夹
func (db *DB) CreateFolder(ctx context.Context, record FolderRecord) error {
	_, err := db.sql.ExecContext(
		ctx,
		`INSERT INTO folders (folder_id, name, parent_id, created_by, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		record.FolderID,
		record.Name,
		record.ParentID,
		record.CreatedBy,
		record.CreatedAt,
		record.UpdatedAt,
	)
	return err
}

// GetFolder 获取文件夹
func (db *DB) GetFolder(ctx context.Context, folderID string) (FolderRecord, error) {
	var record FolderRecord
	row := db.sql.QueryRowContext(ctx, `
		SELECT folder_id, name, parent_id, created_by, created_at, updated_at
		FROM folders WHERE folder_id = ?`, folderID)
	if err := row.Scan(
		&record.FolderID,
		&record.Name,
		&record.ParentID,
		&record.CreatedBy,
		&record.CreatedAt,
		&record.UpdatedAt,
	); err != nil {
		return FolderRecord{}, err
	}
	return record, nil
}

// GetFolderByName 通过名称和父文件夹获取文件夹
func (db *DB) GetFolderByName(ctx context.Context, name string, parentID *string) (FolderRecord, error) {
	var record FolderRecord
	var row *sql.Row
	if parentID == nil {
		row = db.sql.QueryRowContext(ctx, `
			SELECT folder_id, name, parent_id, created_by, created_at, updated_at
			FROM folders WHERE name = ? AND parent_id IS NULL`, name)
	} else {
		row = db.sql.QueryRowContext(ctx, `
			SELECT folder_id, name, parent_id, created_by, created_at, updated_at
			FROM folders WHERE name = ? AND parent_id = ?`, name, *parentID)
	}
	if err := row.Scan(
		&record.FolderID,
		&record.Name,
		&record.ParentID,
		&record.CreatedBy,
		&record.CreatedAt,
		&record.UpdatedAt,
	); err != nil {
		return FolderRecord{}, err
	}
	return record, nil
}

// GetFolderByPath 通过路径获取文件夹
func (db *DB) GetFolderByPath(ctx context.Context, path string) (FolderRecord, error) {
	if path == "" || path == "/" {
		return FolderRecord{}, fmt.Errorf("empty path")
	}

	parts := splitPath(path)
	if len(parts) == 0 {
		return FolderRecord{}, fmt.Errorf("invalid path")
	}

	var currentParentID *string
	var record FolderRecord
	var err error

	for _, part := range parts {
		record, err = db.GetFolderByName(ctx, part, currentParentID)
		if err != nil {
			return FolderRecord{}, fmt.Errorf("folder '%s' not found in path '%s'", part, path)
		}
		currentParentID = &record.FolderID
	}

	return record, nil
}

func splitPath(path string) []string {
	if path == "" {
		return nil
	}
	if path[0] == '/' {
		path = path[1:]
	}
	if path == "" {
		return nil
	}

	var parts []string
	start := 0
	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			if i > start {
				parts = append(parts, path[start:i])
			}
			start = i + 1
		}
	}
	if start < len(path) {
		parts = append(parts, path[start:])
	}
	return parts
}

// GetFolderPath 根据 folderID 获取完整路径
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

// GetFileByPath 通过路径获取文件
func (db *DB) GetFileByPath(ctx context.Context, path string) (FileRecord, error) {
	if path == "" || path == "/" {
		return FileRecord{}, fmt.Errorf("invalid path")
	}

	path = strings.TrimSuffix(path, "/")
	parts := splitPath(path)
	if len(parts) == 0 {
		return FileRecord{}, fmt.Errorf("invalid path")
	}

	fileName := parts[len(parts)-1]
	folderPath := ""
	if len(parts) > 1 {
		folderPath = "/" + strings.Join(parts[:len(parts)-1], "/")
	}

	var folderID *string
	if folderPath != "" {
		folder, err := db.GetFolderByPath(ctx, folderPath)
		if err != nil {
			return FileRecord{}, fmt.Errorf("folder not found: %s", folderPath)
		}
		folderID = &folder.FolderID
	}

	return db.GetFileByName(ctx, fileName, folderID)
}

// GetFilePath 根据 fileID 获取完整路径
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

// ListFolders 列出文件夹
func (db *DB) ListFolders(ctx context.Context, parentID *string) ([]FolderRecord, error) {
	var rows *sql.Rows
	var err error
	if parentID == nil {
		rows, err = db.sql.QueryContext(ctx, `
			SELECT folder_id, name, parent_id, created_by, created_at, updated_at
			FROM folders WHERE parent_id IS NULL ORDER BY name ASC`)
	} else {
		rows, err = db.sql.QueryContext(ctx, `
			SELECT folder_id, name, parent_id, created_by, created_at, updated_at
			FROM folders WHERE parent_id = ? ORDER BY name ASC`, *parentID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]FolderRecord, 0)
	for rows.Next() {
		var record FolderRecord
		if err := rows.Scan(
			&record.FolderID,
			&record.Name,
			&record.ParentID,
			&record.CreatedBy,
			&record.CreatedAt,
			&record.UpdatedAt,
		); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

// UpdateFolder 更新文件夹（重命名）
func (db *DB) UpdateFolder(ctx context.Context, folderID, name string) error {
	_, err := db.sql.ExecContext(ctx,
		`UPDATE folders SET name = ?, updated_at = ? WHERE folder_id = ?`,
		name, NowRFC3339(), folderID)
	return err
}

// MoveFolder 移动文件夹
func (db *DB) MoveFolder(ctx context.Context, folderID string, parentID *string) error {
	_, err := db.sql.ExecContext(ctx,
		`UPDATE folders SET parent_id = ?, updated_at = ? WHERE folder_id = ?`,
		parentID, NowRFC3339(), folderID)
	return err
}

// DeleteFolder 删除文件夹
func (db *DB) DeleteFolder(ctx context.Context, folderID string) error {
	_, err := db.sql.ExecContext(ctx, `DELETE FROM folders WHERE folder_id = ?`, folderID)
	return err
}

// GetFolderDepth 获取文件夹深度（从根目录开始的层级）
// 根目录的子文件夹深度为 1，孙文件夹深度为 2，以此类推
func (db *DB) GetFolderDepth(ctx context.Context, folderID string) (int, error) {
	depth := 0
	currentID := folderID
	for currentID != "" {
		var parentID *string
		err := db.sql.QueryRowContext(ctx,
			`SELECT parent_id FROM folders WHERE folder_id = ?`, currentID).Scan(&parentID)
		if err != nil {
			return 0, err
		}
		if parentID == nil {
			break
		}
		currentID = *parentID
		depth++         // 向上移动一层，增加深度
		if depth > 20 { // 防止循环引用导致死循环
			return 0, fmt.Errorf("possible circular reference")
		}
	}
	return depth, nil
}

// IsDescendant 检查 targetID 是否是 folderID 的子孙
func (db *DB) IsDescendant(ctx context.Context, folderID, targetID string) (bool, error) {
	if folderID == targetID {
		return true, nil
	}
	currentID := targetID
	for currentID != "" {
		var parentID *string
		err := db.sql.QueryRowContext(ctx,
			`SELECT parent_id FROM folders WHERE folder_id = ?`, currentID).Scan(&parentID)
		if err != nil {
			return false, err
		}
		if parentID == nil {
			break
		}
		if *parentID == folderID {
			return true, nil
		}
		currentID = *parentID
	}
	return false, nil
}

// GetFolderItemCount 获取文件夹内项目数量（文件+子文件夹）
func (db *DB) GetFolderItemCount(ctx context.Context, folderID string) (int, error) {
	var fileCount, folderCount int
	// 文件数
	if err := db.sql.QueryRowContext(ctx,
		`SELECT COUNT(1) FROM files WHERE folder_id = ?`, folderID).Scan(&fileCount); err != nil {
		return 0, err
	}
	// 子文件夹数
	if err := db.sql.QueryRowContext(ctx,
		`SELECT COUNT(1) FROM folders WHERE parent_id = ?`, folderID).Scan(&folderCount); err != nil {
		return 0, err
	}
	return fileCount + folderCount, nil
}

// GetFolderStats 获取文件夹统计信息
func (db *DB) GetFolderStats(ctx context.Context, folderID string) (fileCount int, totalSize int64, err error) {
	row := db.sql.QueryRowContext(ctx, `
		SELECT COUNT(1), COALESCE(SUM(size), 0) 
		FROM files WHERE folder_id = ?`, folderID)
	err = row.Scan(&fileCount, &totalSize)
	return
}

// ==================== 文件操作（更新） ====================

// CreateFileWithFolder 创建文件（支持文件夹）
func (db *DB) CreateFileWithFolder(ctx context.Context, record FileRecord) error {
	_, err := db.sql.ExecContext(
		ctx,
		`INSERT INTO files (file_id, original_name, object_key, size, mime_type, folder_id, created_by, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		record.FileID,
		record.OriginalName,
		record.ObjectKey,
		record.Size,
		record.MimeType,
		record.FolderID,
		record.CreatedBy,
		record.CreatedAt,
		record.UpdatedAt,
	)
	return err
}

// UpdateFileFolder 移动文件到文件夹
func (db *DB) UpdateFileFolder(ctx context.Context, fileID string, folderID *string) error {
	_, err := db.sql.ExecContext(ctx,
		`UPDATE files SET folder_id = ?, updated_at = ? WHERE file_id = ?`,
		folderID, NowRFC3339(), fileID)
	return err
}

// ListFilesByFolder 列出指定文件夹的文件
func (db *DB) ListFilesByFolder(ctx context.Context, folderID *string, limit, offset int, order, keyword string) ([]FileRecord, int, error) {
	if order != "asc" {
		order = "desc"
	}

	// 构建 WHERE 条件
	whereClause := ""
	args := []interface{}{}

	if folderID == nil {
		whereClause = "WHERE folder_id IS NULL"
	} else {
		whereClause = "WHERE folder_id = ?"
		args = append(args, *folderID)
	}

	if keyword != "" {
		whereClause += " AND original_name LIKE ?"
		args = append(args, "%"+keyword+"%")
	}

	// 统计总数
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(1) FROM files %s", whereClause)
	if err := db.sql.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// 查询列表
	query := fmt.Sprintf(`
		SELECT file_id, original_name, object_key, size, mime_type, folder_id, created_by, created_at, updated_at
		FROM files %s ORDER BY created_at %s LIMIT ? OFFSET ?`, whereClause, order)
	args = append(args, limit, offset)

	rows, err := db.sql.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	records := make([]FileRecord, 0)
	for rows.Next() {
		var record FileRecord
		if err := rows.Scan(
			&record.FileID,
			&record.OriginalName,
			&record.ObjectKey,
			&record.Size,
			&record.MimeType,
			&record.FolderID,
			&record.CreatedBy,
			&record.CreatedAt,
			&record.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		records = append(records, record)
	}
	return records, total, nil
}

// GetFileWithFolder 获取文件（包含 folder_id）
func (db *DB) GetFileWithFolder(ctx context.Context, fileID string) (FileRecord, error) {
	var record FileRecord
	row := db.sql.QueryRowContext(ctx, `
		SELECT file_id, original_name, object_key, size, mime_type, folder_id, created_by, created_at, updated_at
		FROM files WHERE file_id = ?`, fileID)
	if err := row.Scan(
		&record.FileID,
		&record.OriginalName,
		&record.ObjectKey,
		&record.Size,
		&record.MimeType,
		&record.FolderID,
		&record.CreatedBy,
		&record.CreatedAt,
		&record.UpdatedAt,
	); err != nil {
		return FileRecord{}, err
	}
	return record, nil
}
