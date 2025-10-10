package router

import (
	"github.com/gin-gonic/gin"
	"github.com/waitform/micro-cloud-storage/internal/api/handler"
	"github.com/waitform/micro-cloud-storage/internal/casbin"
	"github.com/waitform/micro-cloud-storage/internal/middleware"
	"github.com/waitform/micro-cloud-storage/internal/rpc"
	"github.com/waitform/micro-cloud-storage/utils"
)

// RegisterRoutes 注册所有路由
func RegisterRoutes(r *gin.Engine,
	userHandler *handler.UserHandler,
	shareHandler *handler.ShareHandler,
	fileHandler *handler.FileHandler,
	shareClient *rpc.ShareServiceClient,
	ipRateLimiter *utils.IPRateLimiter) {

	// 创建可复用的认证中间件实例
	userAuthMiddleware := middleware.AuthUserMiddleware()

	// 创建分享鉴权中间件实例
	shareAuthMiddleware := middleware.AuthShareMiddleware(shareClient)

	// 创建IP限流中间件实例
	ipRateLimitMiddleware := middleware.IPRateLimitMiddleware(ipRateLimiter)

	//casbin文件鉴权中间件
	casbinMW, err := middleware.NewCasbinMiddleware(casbin.GetEnforcer())
	if err != nil {
		panic(err)
	}
	casbinMW.RequirePermission("file", "file_id", "download")

	// 注册用户相关路由
	userGroup := r.Group("/api/user")
	{
		userGroup.POST("/register", ipRateLimitMiddleware, userHandler.HandleUserRegister)
		userGroup.POST("/login", ipRateLimitMiddleware, userHandler.HandleUserLogin)
		userGroup.GET("/info", userAuthMiddleware, userHandler.HandleGetUserInfo)
	}

	// 注册分享相关路由
	shareGroup := r.Group("/api/share")
	{
		shareGroup.POST("/create", userAuthMiddleware, shareHandler.HandleCreateShare)
		shareGroup.GET("/info", shareHandler.HandleGetShareInfo)
		shareGroup.POST("/validate", shareHandler.HandleValidateAccess)
	}

	// 注册文件相关路由
	fileGroup := r.Group("/api/file")
	// 在路由组上统一应用认证中间件，避免重复调用
	fileGroup.Use(userAuthMiddleware)
	{
		// fileGroup.POST("/direct-upload", fileHandler.HandleDirectUpload)
		fileGroup.POST("/upload/init", fileHandler.HandleInitUpload)
		fileGroup.POST("/upload/part", fileHandler.HandleUploadPart)
		fileGroup.POST("/upload/complete", fileHandler.HandleCompleteUpload)
		fileGroup.GET("/info", casbinMW.RequirePermission("file", "file_id", "read"), fileHandler.HandleGetFileInfo)
		fileGroup.POST("/presigned-url", fileHandler.HandleGeneratePresignedURL)
		fileGroup.GET("/upload/progress", fileHandler.HandleGetUploadProgress)
		fileGroup.POST("/upload/incomplete-parts", fileHandler.HandleGetIncompleteParts)
		fileGroup.POST("/upload/cancel", fileHandler.HandleCancelUpload)
		fileGroup.POST("/delete", fileHandler.HandleDeleteFile)
	}

	// 文件下载路由（支持分享链接访问）
	downloadGroup := r.Group("/api/download")
	{
		downloadGroup.GET("", shareAuthMiddleware, fileHandler.HandleDownloadFile)
	}
}
