package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kiry163/claw-pliers/internal/config"
	"github.com/kiry163/claw-pliers/internal/database"
	"github.com/kiry163/claw-pliers/internal/file"
)

type FolderHandler struct {
	Config *config.Config
}

func NewFolderHandler(cfg *config.Config) *FolderHandler {
	return &FolderHandler{Config: cfg}
}

func (h *FolderHandler) CreateFolder(c *gin.Context) {
	var req struct {
		Name     string  `json:"name" binding:"required"`
		ParentID *string `json:"parent_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 10004, "invalid request")
		return
	}

	if req.Name == "" {
		Error(c, http.StatusBadRequest, 10004, "name is required")
		return
	}

	folderID := generateFolderID()
	record := database.Folder{
		FolderID:  folderID,
		Name:      req.Name,
		ParentID:  req.ParentID,
		CreatedBy: getUser(c),
		CreatedAt: database.NowRFC3339(),
		UpdatedAt: database.NowRFC3339(),
	}

	if err := file.Database.CreateFolder(&record); err != nil {
		Error(c, http.StatusInternalServerError, 19999, "failed to create folder")
		return
	}

	OK(c, gin.H{
		"folder_id":  folderID,
		"name":       req.Name,
		"parent_id":  req.ParentID,
		"created_at": record.CreatedAt,
	})
}

func (h *FolderHandler) ListFolders(c *gin.Context) {
	parentID := c.Query("parent_id")
	var parentIDPtr *string
	if parentID != "" {
		parentIDPtr = &parentID
	}

	folders, err := file.Database.ListFolders(parentIDPtr)
	if err != nil {
		Error(c, http.StatusInternalServerError, 19999, "failed to list folders")
		return
	}

	items := make([]gin.H, 0, len(folders))
	for _, f := range folders {
		items = append(items, gin.H{
			"folder_id":  f.FolderID,
			"name":       f.Name,
			"parent_id":  f.ParentID,
			"created_at": f.CreatedAt,
		})
	}

	OK(c, gin.H{
		"total":   len(folders),
		"folders": items,
	})
}

func (h *FolderHandler) GetFolderByPath(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		path = "/"
	}

	folder, err := file.Database.GetFolderByPath(path)
	if err != nil {
		Error(c, http.StatusNotFound, 10002, "folder not found")
		return
	}

	OK(c, gin.H{
		"folder_id":  folder.FolderID,
		"name":       folder.Name,
		"parent_id":  folder.ParentID,
		"path":       path,
		"created_at": folder.CreatedAt,
	})
}

func (h *FolderHandler) CreateFolderByPath(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		Error(c, http.StatusBadRequest, 10004, "path is required")
		return
	}

	path = strings.TrimPrefix(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		Error(c, http.StatusBadRequest, 10004, "invalid path")
		return
	}

	var currentParentID *string
	for i := 0; i < len(parts)-1; i++ {
		folder, err := file.Database.GetFolderByName(parts[i], currentParentID)
		if err != nil {
			newFolderID := generateFolderID()
			record := database.Folder{
				FolderID:  newFolderID,
				Name:      parts[i],
				ParentID:  currentParentID,
				CreatedBy: getUser(c),
				CreatedAt: database.NowRFC3339(),
				UpdatedAt: database.NowRFC3339(),
			}
			if err := file.Database.CreateFolder(&record); err != nil {
				Error(c, http.StatusInternalServerError, 19999, "failed to create folder")
				return
			}
			currentParentID = &newFolderID
		} else {
			currentParentID = &folder.FolderID
		}
	}

	folderName := parts[len(parts)-1]
	existing, _ := file.Database.GetFolderByName(folderName, currentParentID)
	if existing.FolderID != "" {
		Error(c, http.StatusConflict, 10010, "folder already exists")
		return
	}

	folderID := generateFolderID()
	record := database.Folder{
		FolderID:  folderID,
		Name:      folderName,
		ParentID:  currentParentID,
		CreatedBy: getUser(c),
		CreatedAt: database.NowRFC3339(),
		UpdatedAt: database.NowRFC3339(),
	}

	if err := file.Database.CreateFolder(&record); err != nil {
		Error(c, http.StatusInternalServerError, 19999, "failed to create folder")
		return
	}

	OK(c, gin.H{
		"folder_id":  folderID,
		"name":       folderName,
		"parent_id":  currentParentID,
		"path":       "/" + path,
		"created_at": record.CreatedAt,
	})
}

func (h *FolderHandler) RenameFolderByPath(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		Error(c, http.StatusBadRequest, 10004, "path is required")
		return
	}

	newName := c.Query("new_name")
	if newName == "" {
		Error(c, http.StatusBadRequest, 10004, "new_name is required")
		return
	}

	folder, err := file.Database.GetFolderByPath(path)
	if err != nil {
		Error(c, http.StatusNotFound, 10002, "folder not found")
		return
	}

	if err := file.Database.UpdateFolder(folder.FolderID, newName); err != nil {
		Error(c, http.StatusInternalServerError, 19999, "failed to rename folder")
		return
	}

	Message(c, "folder_renamed")
}

func (h *FolderHandler) DeleteFolderByPath(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		Error(c, http.StatusBadRequest, 10004, "path is required")
		return
	}

	folder, err := file.Database.GetFolderByPath(path)
	if err != nil {
		Error(c, http.StatusNotFound, 10002, "folder not found")
		return
	}

	count, _, err := file.Database.GetFolderItemCount(folder.FolderID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 19999, "failed to check folder")
		return
	}

	if count > 0 {
		Error(c, http.StatusBadRequest, 10011, "folder not empty")
		return
	}

	if err := file.Database.DeleteFolder(folder.FolderID); err != nil {
		Error(c, http.StatusInternalServerError, 19999, "failed to delete folder")
		return
	}

	Message(c, "folder_deleted")
}

func generateFolderID() string {
	return fmt.Sprintf("%d%s", time.Now().Unix(), randomString(8))
}
