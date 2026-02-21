package database

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	Path string
}

type DB struct {
	*gorm.DB
}

type Folder struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	FolderID  string    `gorm:"column:folder_id;uniqueIndex" json:"folder_id"`
	Name      string    `gorm:"column:name" json:"name"`
	ParentID  *string   `gorm:"column:parent_id" json:"parent_id"`
	CreatedBy string    `gorm:"column:created_by" json:"created_by"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (Folder) TableName() string {
	return "folders"
}

type File struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	FileID       string    `gorm:"column:file_id;uniqueIndex" json:"file_id"`
	OriginalName string    `gorm:"column:original_name" json:"original_name"`
	ObjectKey    string    `gorm:"column:object_key" json:"object_key"`
	Size         int64     `gorm:"column:size" json:"size"`
	MimeType     string    `gorm:"column:mime_type" json:"mime_type"`
	FolderID     *string   `gorm:"column:folder_id" json:"folder_id"`
	CreatedBy    string    `gorm:"column:created_by" json:"created_by"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updated_at"`
	Metadata     string    `gorm:"column:metadata;type:json" json:"metadata"`
}

func (File) TableName() string {
	return "files"
}

type RefreshToken struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Token     string    `gorm:"column:token;uniqueIndex" json:"token"`
	ExpiresAt time.Time `gorm:"column:expires_at" json:"expires_at"`
	IsRevoked bool      `gorm:"column:is_revoked;default:false" json:"is_revoked"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

type AuditLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Action    string    `gorm:"column:action" json:"action"`
	FileID    *string   `gorm:"column:file_id" json:"file_id"`
	Actor     string    `gorm:"column:actor" json:"actor"`
	IPAddress *string   `gorm:"column:ip_address" json:"ip_address"`
	Status    string    `gorm:"column:status" json:"status"`
	Message   string    `gorm:"column:message;type:text" json:"message"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}

type ShareLink struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Token     string    `gorm:"column:token;uniqueIndex" json:"token"`
	FileID    string    `gorm:"column:file_id" json:"file_id"`
	ExpiresAt time.Time `gorm:"column:expires_at" json:"expires_at"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	CreatedBy string    `gorm:"column:created_by" json:"created_by"`
	Status    string    `gorm:"column:status" json:"status"`
}

func (ShareLink) TableName() string {
	return "share_links"
}

func Open(cfg Config) (*DB, error) {
	db, err := gorm.Open(sqlite.Open(cfg.Path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}
	sqlDB.SetMaxOpenConns(1)

	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return &DB{db}, nil
}

func migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&Folder{},
		&File{},
		&RefreshToken{},
		&AuditLog{},
		&ShareLink{},
	)
}

func (db *DB) CreateFile(record *File) error {
	return db.Create(record).Error
}

func (db *DB) GetFile(fileID string) (File, error) {
	var file File
	err := db.Where("file_id = ?", fileID).First(&file).Error
	return file, err
}

func (db *DB) GetFileByName(name string, folderID *string) (File, error) {
	var file File
	query := db.Where("original_name = ?", name)
	if folderID == nil {
		query = query.Where("folder_id IS NULL")
	} else {
		query = query.Where("folder_id = ?", *folderID)
	}
	err := query.First(&file).Error
	return file, err
}

func (db *DB) ListFiles(limit, offset int, order, keyword string) ([]File, int64, error) {
	var files []File
	var total int64

	query := db.Model(&File{})
	if keyword != "" {
		query = query.Where("original_name LIKE ?", "%"+keyword+"%")
	}
	query.Count(&total)

	if order != "asc" {
		order = "desc"
	}
	query = query.Order("created_at " + order).Limit(limit).Offset(offset)

	err := query.Find(&files).Error
	return files, total, err
}

func (db *DB) DeleteFile(fileID string) (File, error) {
	file, err := db.GetFile(fileID)
	if err != nil {
		return file, err
	}
	err = db.Delete(&file).Error
	return file, err
}

func (db *DB) CreateRefreshToken(record *RefreshToken) error {
	return db.Create(record).Error
}

func (db *DB) GetRefreshToken(token string) (RefreshToken, error) {
	var tokenRecord RefreshToken
	err := db.Where("token = ?", token).First(&tokenRecord).Error
	return tokenRecord, err
}

func (db *DB) RevokeRefreshToken(token string) error {
	return db.Model(&RefreshToken{}).Where("token = ?", token).Update("is_revoked", true).Error
}

func (db *DB) RevokeAllRefreshTokens() error {
	return db.Model(&RefreshToken{}).Where("1=1").Update("is_revoked", true).Error
}

func (db *DB) CreateShareLink(link *ShareLink) error {
	return db.Create(link).Error
}

func (db *DB) GetShareLink(token string) (ShareLink, error) {
	var link ShareLink
	err := db.Where("token = ?", token).First(&link).Error
	return link, err
}

func (db *DB) GetActiveShareLink(fileID string, now time.Time) (ShareLink, error) {
	var link ShareLink
	err := db.Where("file_id = ? AND status = ? AND expires_at > ?", fileID, "active", now).
		Order("created_at DESC").
		First(&link).Error
	return link, err
}

func (db *DB) CreateFolder(record *Folder) error {
	return db.Create(record).Error
}

func (db *DB) GetFolder(folderID string) (Folder, error) {
	var folder Folder
	err := db.Where("folder_id = ?", folderID).First(&folder).Error
	return folder, err
}

func (db *DB) GetFolderByName(name string, parentID *string) (Folder, error) {
	var folder Folder
	query := db.Where("name = ?", name)
	if parentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", *parentID)
	}
	err := query.First(&folder).Error
	return folder, err
}

func (db *DB) GetFolderPath(folderID string) (string, error) {
	folder, err := db.GetFolder(folderID)
	if err != nil {
		return "", err
	}

	if folder.ParentID == nil {
		return "/" + folder.Name, nil
	}

	parentPath, err := db.GetFolderPath(*folder.ParentID)
	if err != nil {
		return "", err
	}

	return parentPath + "/" + folder.Name, nil
}

func (db *DB) GetFilePath(fileID string) (string, error) {
	file, err := db.GetFile(fileID)
	if err != nil {
		return "", err
	}

	if file.FolderID == nil {
		return "/" + file.OriginalName, nil
	}

	folderPath, err := db.GetFolderPath(*file.FolderID)
	if err != nil {
		return "", err
	}

	return folderPath + "/" + file.OriginalName, nil
}

func (db *DB) ListFolders(parentID *string) ([]Folder, error) {
	var folders []Folder
	query := db.Model(&Folder{})
	if parentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", *parentID)
	}
	err := query.Order("name ASC").Find(&folders).Error
	return folders, err
}

func (db *DB) UpdateFolder(folderID, name string) error {
	return db.Model(&Folder{}).Where("folder_id = ?", folderID).Update("name", name).Error
}

func (db *DB) MoveFolder(folderID string, parentID *string) error {
	return db.Model(&Folder{}).Where("folder_id = ?", folderID).Update("parent_id", parentID).Error
}

func (db *DB) DeleteFolder(folderID string) error {
	return db.Where("folder_id = ?", folderID).Delete(&Folder{}).Error
}

func (db *DB) GetFolderItemCount(folderID string) (int64, int64, error) {
	var fileCount, folderCount int64
	db.Model(&File{}).Where("folder_id = ?", folderID).Count(&fileCount)
	db.Model(&Folder{}).Where("parent_id = ?", folderID).Count(&folderCount)
	return fileCount, folderCount, nil
}

func (db *DB) GetFolderStats(folderID string) (int64, int64, error) {
	var fileCount int64
	var totalSize int64
	db.Model(&File{}).Where("folder_id = ?", folderID).Count(&fileCount)
	db.Model(&File{}).Where("folder_id = ?", folderID).Select("COALESCE(SUM(size), 0)").Scan(&totalSize)
	return fileCount, totalSize, nil
}

func (db *DB) UpdateFileFolder(fileID string, folderID *string) error {
	return db.Model(&File{}).Where("file_id = ?", fileID).Update("folder_id", folderID).Error
}

func (db *DB) UpdateFileName(fileID string, newName string) error {
	return db.Model(&File{}).Where("file_id = ?", fileID).Update("original_name", newName).Error
}

func (db *DB) ListFilesByFolder(folderID *string, limit, offset int, order, keyword string) ([]File, int64, error) {
	var files []File
	var total int64

	query := db.Model(&File{})
	if folderID == nil {
		query = query.Where("folder_id IS NULL")
	} else {
		query = query.Where("folder_id = ?", *folderID)
	}

	if keyword != "" {
		query = query.Where("original_name LIKE ?", "%"+keyword+"%")
	}

	query.Count(&total)

	if order != "asc" {
		order = "desc"
	}
	query = query.Order("created_at " + order).Limit(limit).Offset(offset)

	err := query.Find(&files).Error
	return files, total, err
}

func (db *DB) GetFolderByPath(path string) (Folder, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 {
		return Folder{}, fmt.Errorf("invalid path")
	}

	name := parts[0]
	var parentID *string

	folder, err := db.GetFolderByName(name, parentID)
	if err != nil {
		return folder, err
	}

	for i := 1; i < len(parts); i++ {
		parentID = &folder.FolderID
		folder, err = db.GetFolderByName(parts[i], parentID)
		if err != nil {
			return folder, err
		}
	}

	return folder, nil
}

func (db *DB) GetFileByPath(path string) (File, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 {
		return File{}, fmt.Errorf("invalid path")
	}

	fileName := parts[len(parts)-1]
	var folderID *string

	if len(parts) > 1 {
		folderPath := "/" + strings.Join(parts[:len(parts)-1], "/")
		folder, err := db.GetFolderByPath(folderPath)
		if err != nil {
			return File{}, err
		}
		folderID = &folder.FolderID
	}

	file, err := db.GetFileByName(fileName, folderID)
	return file, err
}

func (db *DB) AddAuditLog(action, fileID, actor, ipAddress, status, message string) error {
	record := &AuditLog{
		Action:    action,
		FileID:    &fileID,
		Actor:     actor,
		IPAddress: &ipAddress,
		Status:    status,
		Message:   message,
	}
	return db.Create(record).Error
}

func NowRFC3339() time.Time {
	return time.Now().UTC()
}
