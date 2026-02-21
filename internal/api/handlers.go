package api

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kiry163/claw-pliers/internal/config"
	"github.com/kiry163/claw-pliers/internal/database"
	"github.com/kiry163/claw-pliers/internal/file"
	"github.com/kiry163/claw-pliers/internal/utils"
)

type FileHandler struct {
	Config *config.Config
}

func NewFileHandler(cfg *config.Config) *FileHandler {
	return &FileHandler{Config: cfg}
}

func (h *FileHandler) UploadFile(c *gin.Context) {
	uploadedFile, err := c.FormFile("file")
	if err != nil {
		Error(c, http.StatusBadRequest, 10004, "file required")
		return
	}
	maxBytes := h.Config.Upload.MaxSizeMB * 1024 * 1024
	if maxBytes > 0 && uploadedFile.Size > maxBytes {
		Error(c, http.StatusBadRequest, 10004, "file too large")
		return
	}

	folderID := c.Query("folder_id")
	var folderIDPtr *string
	if folderID != "" {
		folderIDPtr = &folderID
	}

	src, err := uploadedFile.Open()
	if err != nil {
		Error(c, http.StatusInternalServerError, 19999, "failed to open file")
		return
	}
	defer src.Close()

	fileID := utils.GenerateFileID()
	saveResult, err := file.FileStorage.Save(c.Request.Context(), src, uploadedFile.Size, fileID, uploadedFile.Filename)
	if err != nil {
		Error(c, http.StatusInternalServerError, 19999, "failed to save file")
		return
	}

	record := database.File{
		FileID:       fileID,
		OriginalName: uploadedFile.Filename,
		ObjectKey:    saveResult.ObjectKey,
		Size:         uploadedFile.Size,
		MimeType:     saveResult.MimeType,
		FolderID:     folderIDPtr,
		CreatedBy:    getUser(c),
		CreatedAt:    database.NowRFC3339(),
		UpdatedAt:    database.NowRFC3339(),
	}

	var dbErr error
	if folderIDPtr != nil {
		dbErr = file.Database.CreateFile(&record)
	} else {
		dbErr = file.Database.CreateFile(&record)
	}

	if dbErr != nil {
		Error(c, http.StatusInternalServerError, 19999, "failed to save record")
		return
	}

	OK(c, gin.H{
		"file_id":       fileID,
		"original_name": uploadedFile.Filename,
		"size":          uploadedFile.Size,
		"mime_type":     saveResult.MimeType,
	})
}

func (h *FileHandler) ListFiles(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	order := c.DefaultQuery("order", "desc")
	keyword := c.Query("keyword")
	folderID := c.Query("folder_id")

	var folderIDPtr *string
	if folderID != "" {
		folderIDPtr = &folderID
	}

	var records []database.File
	var total int64
	var err error

	if folderIDPtr != nil || folderID == "" {
		records, total, err = file.Database.ListFilesByFolder(folderIDPtr, limit, offset, order, keyword)
	} else {
		records, total, err = file.Database.ListFiles(limit, offset, order, keyword)
	}

	if err != nil {
		Error(c, http.StatusInternalServerError, 19999, "failed to list files")
		return
	}

	items := make([]gin.H, 0, len(records))
	for _, r := range records {
		items = append(items, gin.H{
			"file_id":       r.FileID,
			"original_name": r.OriginalName,
			"size":          r.Size,
			"mime_type":     r.MimeType,
			"created_at":    r.CreatedAt,
		})
	}

	OK(c, gin.H{
		"total": total,
		"items": items,
	})
}

func (h *FileHandler) GetFile(c *gin.Context) {
	fileID := c.Param("id")
	record, err := file.Database.GetFile(fileID)
	if err != nil {
		Error(c, http.StatusNotFound, 10002, "file not found")
		return
	}

	OK(c, gin.H{
		"file_id":       record.FileID,
		"original_name": record.OriginalName,
		"size":          record.Size,
		"mime_type":     record.MimeType,
		"created_at":    record.CreatedAt,
	})
}

func (h *FileHandler) DownloadFile(c *gin.Context) {
	fileID := c.Param("id")
	record, err := file.Database.GetFile(fileID)
	if err != nil {
		Error(c, http.StatusNotFound, 10002, "file not found")
		return
	}

	reader, _, err := file.FileStorage.Get(c.Request.Context(), record.ObjectKey, nil, nil)
	if err != nil {
		Error(c, http.StatusInternalServerError, 19999, "failed to get file")
		return
	}
	defer reader.Close()

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", record.OriginalName))
	c.Header("Content-Type", record.MimeType)
	c.Header("Content-Length", strconv.FormatInt(record.Size, 10))
	c.Status(http.StatusOK)
	io.Copy(c.Writer, reader)
}

func (h *FileHandler) DeleteFile(c *gin.Context) {
	fileID := c.Param("id")
	record, err := file.Database.DeleteFile(fileID)
	if err != nil {
		Error(c, http.StatusNotFound, 10002, "file not found")
		return
	}

	if err := file.FileStorage.Delete(c.Request.Context(), record.ObjectKey); err != nil {
		Error(c, http.StatusInternalServerError, 19999, "failed to delete file")
		return
	}

	Message(c, "file_deleted")
}

func (h *FileHandler) UploadFileByPath(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		Error(c, http.StatusBadRequest, 10004, "path is required")
		return
	}

	path = strings.TrimPrefix(path, "/")

	uploadedFile, err := c.FormFile("file")
	if err != nil {
		Error(c, http.StatusBadRequest, 10004, "file required")
		return
	}

	maxBytes := h.Config.Upload.MaxSizeMB * 1024 * 1024
	if maxBytes > 0 && uploadedFile.Size > maxBytes {
		Error(c, http.StatusBadRequest, 10004, "file too large")
		return
	}

	src, err := uploadedFile.Open()
	if err != nil {
		Error(c, http.StatusInternalServerError, 19999, "failed to open file")
		return
	}
	defer src.Close()

	parts := strings.Split(path, "/")
	var folderID *string
	var fileName string

	if len(parts) > 1 {
		folderPath := strings.Join(parts[:len(parts)-1], "/")
		fileName = parts[len(parts)-1]

		folder, err := file.Database.GetFolderByPath("/" + folderPath)
		if err != nil {
			Error(c, http.StatusNotFound, 10002, "parent folder not found")
			return
		}
		folderID = &folder.FolderID
	} else {
		fileName = parts[0]
	}

	fileID := utils.GenerateFileID()
	saveResult, err := file.FileStorage.Save(c.Request.Context(), src, uploadedFile.Size, fileID, fileName)
	if err != nil {
		Error(c, http.StatusInternalServerError, 19999, "failed to save file")
		return
	}

	record := database.File{
		FileID:       fileID,
		OriginalName: fileName,
		ObjectKey:    saveResult.ObjectKey,
		Size:         uploadedFile.Size,
		MimeType:     saveResult.MimeType,
		FolderID:     folderID,
		CreatedBy:    getUser(c),
		CreatedAt:    database.NowRFC3339(),
		UpdatedAt:    database.NowRFC3339(),
	}

	var dbErr error
	if folderID != nil {
		dbErr = file.Database.CreateFile(&record)
	} else {
		dbErr = file.Database.CreateFile(&record)
	}

	if dbErr != nil {
		Error(c, http.StatusInternalServerError, 19999, "failed to save record")
		return
	}

	OK(c, gin.H{
		"file_id":       fileID,
		"original_name": fileName,
		"path":          "/" + path,
		"size":          uploadedFile.Size,
		"mime_type":     saveResult.MimeType,
	})
}

func (h *FileHandler) ListFilesByPath(c *gin.Context) {
	path := c.Query("path")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	order := c.DefaultQuery("order", "desc")
	keyword := c.Query("keyword")

	var folderID *string
	if path != "" && path != "/" {
		folder, err := file.Database.GetFolderByPath(path)
		if err == nil {
			folderID = &folder.FolderID
		}
	}

	var records []database.File
	var total int64
	var err error
	records, total, err = file.Database.ListFilesByFolder(folderID, limit, offset, order, keyword)
	if err != nil {
		Error(c, http.StatusInternalServerError, 19999, "failed to list files")
		return
	}

	items := make([]gin.H, 0, len(records))
	for _, r := range records {
		var filePath string
		if r.FolderID != nil {
			folderPath, _ := file.Database.GetFolderPath(*r.FolderID)
			filePath = folderPath + "/" + r.OriginalName
		} else {
			filePath = "/" + r.OriginalName
		}

		items = append(items, gin.H{
			"file_id":       r.FileID,
			"original_name": r.OriginalName,
			"path":          filePath,
			"size":          r.Size,
			"mime_type":     r.MimeType,
			"created_at":    r.CreatedAt,
		})
	}

	OK(c, gin.H{
		"total": total,
		"items": items,
	})
}

func (h *FileHandler) GetFileByPath(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		Error(c, http.StatusBadRequest, 10004, "path is required")
		return
	}

	record, err := file.Database.GetFileByPath(path)
	if err != nil {
		Error(c, http.StatusNotFound, 10002, "file not found")
		return
	}

	OK(c, gin.H{
		"file_id":       record.FileID,
		"original_name": record.OriginalName,
		"path":          path,
		"size":          record.Size,
		"mime_type":     record.MimeType,
		"created_at":    record.CreatedAt,
	})
}

func (h *FileHandler) DownloadFileByPath(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		Error(c, http.StatusBadRequest, 10004, "path is required")
		return
	}

	record, err := file.Database.GetFileByPath(path)
	if err != nil {
		Error(c, http.StatusNotFound, 10002, "file not found")
		return
	}

	reader, _, err := file.FileStorage.Get(c.Request.Context(), record.ObjectKey, nil, nil)
	if err != nil {
		Error(c, http.StatusInternalServerError, 19999, "failed to get file")
		return
	}
	defer reader.Close()

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", record.OriginalName))
	c.Header("Content-Type", record.MimeType)
	c.Header("Content-Length", strconv.FormatInt(record.Size, 10))
	c.Status(http.StatusOK)
	io.Copy(c.Writer, reader)
}

func (h *FileHandler) DeleteFileByPath(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		Error(c, http.StatusBadRequest, 10004, "path is required")
		return
	}

	record, err := file.Database.GetFileByPath(path)
	if err != nil {
		Error(c, http.StatusNotFound, 10002, "file not found")
		return
	}

	if err := file.FileStorage.Delete(c.Request.Context(), record.ObjectKey); err != nil {
		Error(c, http.StatusInternalServerError, 19999, "failed to delete file")
		return
	}

	if _, err := file.Database.DeleteFile(record.FileID); err != nil {
		Error(c, http.StatusInternalServerError, 19999, "failed to delete record")
		return
	}

	Message(c, "file_deleted")
}

func (h *FileHandler) MoveFileByPath(c *gin.Context) {
	path := c.Query("path")
	newPath := c.Query("new_path")

	if path == "" || newPath == "" {
		Error(c, http.StatusBadRequest, 10004, "path and new_path are required")
		return
	}

	record, err := file.Database.GetFileByPath(path)
	if err != nil {
		Error(c, http.StatusNotFound, 10002, "file not found")
		return
	}

	newPath = strings.TrimPrefix(newPath, "/")
	parts := strings.Split(newPath, "/")
	var newFolderID *string

	if len(parts) > 1 {
		folderPath := "/" + strings.Join(parts[:len(parts)-1], "/")

		folder, err := file.Database.GetFolderByPath(folderPath)
		if err != nil {
			Error(c, http.StatusNotFound, 10002, "target folder not found")
			return
		}
		newFolderID = &folder.FolderID
	}

	newFileName := parts[len(parts)-1]

	if newFolderID != nil {
		if err := file.Database.UpdateFileFolder(record.FileID, newFolderID); err != nil {
			Error(c, http.StatusInternalServerError, 19999, "failed to move file")
			return
		}
	}

	if newFileName != "" && newFileName != record.OriginalName {
		if err := file.Database.UpdateFileName(record.FileID, newFileName); err != nil {
			Error(c, http.StatusInternalServerError, 19999, "failed to rename file")
			return
		}
	}

	Message(c, "file_moved")
}

func (h *FileHandler) GetFileInfoByPath(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		Error(c, http.StatusBadRequest, 10004, "path is required")
		return
	}

	record, err := file.Database.GetFileByPath(path)
	if err != nil {
		Error(c, http.StatusNotFound, 10002, "file not found")
		return
	}

	shareLink, _ := file.Database.GetActiveShareLink(record.FileID, database.NowRFC3339())

	var downloadLink string
	var expiresAt time.Time
	publicURL := h.Config.Server.PublicEndpoint
	if publicURL == "" {
		publicURL = fmt.Sprintf("http://localhost:%d", h.Config.Server.Port)
	}

	if shareLink.Token == "" {
		now := time.Now().UTC()
		expiresAtVal := now.Add(7 * 24 * time.Hour)
		token := utils.GenerateShareToken()

		newShareLink := database.ShareLink{
			Token:     token,
			FileID:    record.FileID,
			ExpiresAt: expiresAtVal,
			CreatedAt: now,
			CreatedBy: getUser(c),
			Status:    "active",
		}
		_ = file.Database.CreateShareLink(&newShareLink)
		shareLink.Token = token
		shareLink.ExpiresAt = newShareLink.ExpiresAt
	}

	downloadLink = fmt.Sprintf("%s/s/%s", publicURL, shareLink.Token)
	expiresAt = shareLink.ExpiresAt

	OK(c, gin.H{
		"file_id":       record.FileID,
		"original_name": record.OriginalName,
		"path":          path,
		"size":          record.Size,
		"mime_type":     record.MimeType,
		"created_at":    record.CreatedAt,
		"download_link": downloadLink,
		"expires_at":    expiresAt,
	})
}

func (h *FileHandler) GenerateShareLinkByPath(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		Error(c, http.StatusBadRequest, 10004, "path is required")
		return
	}

	record, err := file.Database.GetFileByPath(path)
	if err != nil {
		Error(c, http.StatusNotFound, 10002, "file not found")
		return
	}

	now := time.Now().UTC()
	expiresAt := now.Add(7 * 24 * time.Hour)
	token := utils.GenerateShareToken()

	shareLink := database.ShareLink{
		Token:     token,
		FileID:    record.FileID,
		ExpiresAt: expiresAt,
		CreatedAt: now,
		CreatedBy: getUser(c),
		Status:    "active",
	}

	if err := file.Database.CreateShareLink(&shareLink); err != nil {
		Error(c, http.StatusInternalServerError, 19999, "failed to create share link")
		return
	}

	publicURL := h.Config.Server.PublicEndpoint
	if publicURL == "" {
		publicURL = fmt.Sprintf("http://localhost:%d", h.Config.Server.Port)
	}
	downloadLink := fmt.Sprintf("%s/s/%s", publicURL, token)

	OK(c, gin.H{
		"token":        token,
		"download_url": downloadLink,
		"expires_at":   expiresAt.Format(time.RFC3339),
	})
}

func (h *FileHandler) DownloadByShareToken(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		Error(c, http.StatusBadRequest, 10004, "token is required")
		return
	}

	link, err := file.Database.GetShareLink(token)
	if err != nil {
		Error(c, http.StatusNotFound, 10002, "link not found")
		return
	}

	if link.Status != "active" {
		Error(c, http.StatusGone, 10003, "link has been revoked")
		return
	}

	if time.Now().UTC().After(link.ExpiresAt) {
		Error(c, http.StatusGone, 10003, "link has expired")
		return
	}

	record, err := file.Database.GetFile(link.FileID)
	if err != nil {
		Error(c, http.StatusNotFound, 10002, "file not found")
		return
	}

	reader, _, err := file.FileStorage.Get(c.Request.Context(), record.ObjectKey, nil, nil)
	if err != nil {
		Error(c, http.StatusInternalServerError, 19999, "failed to get file")
		return
	}
	defer reader.Close()

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", record.OriginalName))
	c.Header("Content-Type", record.MimeType)
	c.Header("Content-Length", strconv.FormatInt(record.Size, 10))
	c.Status(http.StatusOK)
	io.Copy(c.Writer, reader)
}
