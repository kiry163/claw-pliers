package api

import (
	"github.com/gin-gonic/gin"
	"github.com/kiry163/claw-pliers/internal/config"
	"github.com/kiry163/claw-pliers/internal/database"
	"github.com/kiry163/claw-pliers/internal/file"
	"github.com/kiry163/claw-pliers/internal/service"
)

func NewRouter(cfg *config.Config, db *database.DB, version string) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(RequestLogger())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "version": version})
	})

	// Initialize services
	fileService := service.NewFileService(db, file.FileStorage)
	folderService := service.NewFolderService(db)
	mailService := service.NewMailService()

	// Initialize handlers with dependencies
	fileHandler := NewFileHandler(cfg, fileService)
	folderHandler := NewFolderHandler(cfg, folderService)
	mailHandler := NewMailHandler(cfg, mailService)

	api := router.Group("/api/v1")

	// 文件操作 (原有)
	files := api.Group("/files")
	files.Use(AuthMiddleware(cfg))
	files.POST("", fileHandler.UploadFile)
	files.GET("", fileHandler.ListFiles)
	files.GET("/:id", fileHandler.GetFile)
	files.GET("/:id/download", fileHandler.DownloadFile)
	files.DELETE("/:id", fileHandler.DeleteFile)

	// 文件操作 (按路径)
	filesByPath := api.Group("/files/by-path")
	filesByPath.Use(AuthMiddleware(cfg))
	filesByPath.POST("", fileHandler.UploadFileByPath)
	filesByPath.GET("", fileHandler.ListFilesByPath)
	filesByPath.GET("/info", fileHandler.GetFileInfoByPath)
	filesByPath.GET("/share", fileHandler.GenerateShareLinkByPath)
	filesByPath.GET("/download", fileHandler.DownloadFileByPath)
	filesByPath.DELETE("", fileHandler.DeleteFileByPath)
	filesByPath.PUT("", fileHandler.MoveFileByPath)

	// 公开下载链接（无需认证）
	router.GET("/s/:token", fileHandler.DownloadByShareToken)

	// 文件夹操作
	folders := api.Group("/folders")
	folders.Use(AuthMiddleware(cfg))
	folders.POST("", folderHandler.CreateFolder)
	folders.GET("", folderHandler.ListFolders)
	folders.GET("/by-path", folderHandler.GetFolderByPath)

	// 文件夹操作 (按路径)
	foldersByPath := api.Group("/folders/by-path")
	foldersByPath.Use(AuthMiddleware(cfg))
	foldersByPath.POST("", folderHandler.CreateFolderByPath)
	foldersByPath.PUT("", folderHandler.RenameFolderByPath)
	foldersByPath.DELETE("", folderHandler.DeleteFolderByPath)

	// 邮件操作
	mail := api.Group("/mail")
	mail.Use(AuthMiddleware(cfg))
	mail.GET("/test-connection", mailHandler.TestConnection)
	mail.POST("/send", mailHandler.SendMail)
	mail.GET("/latest", mailHandler.GetLatestEmails)
	mail.GET("/accounts", mailHandler.ListAccounts)

	return router
}
