package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kiry163/claw-pliers/internal/config"
	"github.com/kiry163/claw-pliers/internal/database"
	"github.com/kiry163/claw-pliers/internal/file"
	"github.com/kiry163/claw-pliers/internal/response"
	"github.com/kiry163/claw-pliers/internal/service"
)

type FolderHandler struct {
	Config  *config.Config
	Service *service.FolderService
}

func NewFolderHandler(cfg *config.Config, svc *service.FolderService) *FolderHandler {
	return &FolderHandler{Config: cfg, Service: svc}
}

func (h *FolderHandler) CreateFolder(c *gin.Context) {
	var req struct {
		Name     string  `json:"name" binding:"required"`
		ParentID *string `json:"parent_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 10004, "invalid request")
		return
	}

	if req.Name == "" {
		response.Error(c, http.StatusBadRequest, 10004, "name is required")
		return
	}

	metadata, err := h.Service.CreateFolder(c.Request.Context(), req.Name, "", getUser(c))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 19999, "failed to create folder")
		return
	}

	response.Success(c, gin.H{
		"folder_id":  metadata.FolderID,
		"name":       metadata.Name,
		"parent_id":  metadata.ParentID,
		"created_at": metadata.CreatedAt,
	})
}

func (h *FolderHandler) ListFolders(c *gin.Context) {
	parentID := c.Query("parent_id")
	var parentIDPtr *string
	if parentID != "" {
		parentIDPtr = &parentID
	}

	folders, err := h.Service.ListFolders(c.Request.Context(), parentIDPtr)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 19999, "failed to list folders")
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

	response.Success(c, gin.H{
		"total":   len(folders),
		"folders": items,
	})
}

func (h *FolderHandler) GetFolderByPath(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		path = "/"
	}

	metadata, err := h.Service.GetFolderByPath(c.Request.Context(), path)
	if err != nil {
		response.Error(c, http.StatusNotFound, 10002, "folder not found")
		return
	}

	response.Success(c, gin.H{
		"folder_id":  metadata.FolderID,
		"name":       metadata.Name,
		"parent_id":  metadata.ParentID,
		"path":       path,
		"created_at": metadata.CreatedAt,
	})
}

func (h *FolderHandler) CreateFolderByPath(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		response.Error(c, http.StatusBadRequest, 10004, "path is required")
		return
	}

	path = strings.TrimPrefix(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		response.Error(c, http.StatusBadRequest, 10004, "invalid path")
		return
	}

	var currentParentID *string
	for i := 0; i < len(parts)-1; i++ {
		folder, err := h.Service.GetFolderByPath(c.Request.Context(), "/"+parts[i])
		if err != nil {
			newFolderID := h.Service.GenerateFolderID()
			record := database.Folder{
				FolderID:  newFolderID,
				Name:      parts[i],
				ParentID:  currentParentID,
				CreatedBy: getUser(c),
				CreatedAt: database.NowRFC3339(),
				UpdatedAt: database.NowRFC3339(),
			}
			if err := file.Database.CreateFolder(&record); err != nil {
				response.Error(c, http.StatusInternalServerError, 19999, "failed to create folder")
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
		response.Error(c, http.StatusConflict, 10010, "folder already exists")
		return
	}

	metadata, err := h.Service.CreateFolder(c.Request.Context(), folderName, "", getUser(c))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 19999, "failed to create folder")
		return
	}

	response.Success(c, gin.H{
		"folder_id":  metadata.FolderID,
		"name":       metadata.Name,
		"parent_id":  metadata.ParentID,
		"path":       "/" + path,
		"created_at": metadata.CreatedAt,
	})
}

func (h *FolderHandler) RenameFolderByPath(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		response.Error(c, http.StatusBadRequest, 10004, "path is required")
		return
	}

	newName := c.Query("new_name")
	if newName == "" {
		response.Error(c, http.StatusBadRequest, 10004, "new_name is required")
		return
	}

	metadata, err := h.Service.GetFolderByPath(c.Request.Context(), path)
	if err != nil {
		response.Error(c, http.StatusNotFound, 10002, "folder not found")
		return
	}

	if err := h.Service.RenameFolder(c.Request.Context(), metadata.FolderID, newName); err != nil {
		response.Error(c, http.StatusInternalServerError, 19999, "failed to rename folder")
		return
	}

	response.Message(c, "folder_renamed")
}

func (h *FolderHandler) DeleteFolderByPath(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		response.Error(c, http.StatusBadRequest, 10004, "path is required")
		return
	}

	metadata, err := h.Service.GetFolderByPath(c.Request.Context(), path)
	if err != nil {
		response.Error(c, http.StatusNotFound, 10002, "folder not found")
		return
	}

	count, _, err := h.Service.GetFolderItemCount(c.Request.Context(), metadata.FolderID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 19999, "failed to check folder")
		return
	}

	if count > 0 {
		response.Error(c, http.StatusBadRequest, 10011, "folder not empty")
		return
	}

	if err := h.Service.DeleteFolder(c.Request.Context(), metadata.FolderID); err != nil {
		response.Error(c, http.StatusInternalServerError, 19999, "failed to delete folder")
		return
	}

	response.Message(c, "folder_deleted")
}
