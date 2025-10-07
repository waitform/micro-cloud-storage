package router

import (
	"cloud-storage/internal/api/handler"
	"cloud-storage/internal/api/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册所有路由
func RegisterRoutes(r *gin.Engine,
	userHandler *handler.UserHandler,
	shareHandler *handler.ShareHandler,
	fileHandler *handler.FileHandler) {

	// 注册用户相关路由
	userGroup := r.Group("/api/user")
	{
		userGroup.POST("/register", userHandler.HandleUserRegister)
		userGroup.POST("/login", userHandler.HandleUserLogin)
		userGroup.GET("/info", middleware.AuthMiddleware(), userHandler.HandleGetUserInfo)
	}

	// 注册分享相关路由
	shareGroup := r.Group("/api/share")
	{
		shareGroup.POST("/create", middleware.AuthMiddleware(), shareHandler.HandleCreateShare)
		shareGroup.GET("/info", shareHandler.HandleGetShareInfo)
		shareGroup.POST("/validate", shareHandler.HandleValidateAccess)
	}

	// 注册文件相关路由
	fileGroup := r.Group("/api/file")
	{
		fileGroup.POST("/direct-upload", middleware.AuthMiddleware(), fileHandler.HandleDirectUpload)
		fileGroup.POST("/upload/init", middleware.AuthMiddleware(), fileHandler.HandleInitUpload)
		fileGroup.POST("/upload/part", middleware.AuthMiddleware(), fileHandler.HandleUploadPart)
		fileGroup.POST("/upload/complete", middleware.AuthMiddleware(), fileHandler.HandleCompleteUpload)
		fileGroup.GET("/info", middleware.AuthMiddleware(), fileHandler.HandleGetFileInfo)
		fileGroup.POST("/presigned-url", middleware.AuthMiddleware(), fileHandler.HandleGeneratePresignedURL)
	}
}
