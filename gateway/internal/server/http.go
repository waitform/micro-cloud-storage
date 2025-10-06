package server

import (
	"cloud-storage/gateway/internal/handler"
	"cloud-storage/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GatewayServer 网关服务器结构
type GatewayServer struct {
	UserClient  *handler.UserServiceClient
	ShareClient *handler.ShareServiceClient
	FileClient  *handler.FileServiceClient
}

// NewGatewayServer 创建网关服务器实例
func NewGatewayServer(
	userClient *handler.UserServiceClient,
	shareClient *handler.ShareServiceClient,
	fileClient *handler.FileServiceClient,
) *GatewayServer {
	return &GatewayServer{
		UserClient:  userClient,
		ShareClient: shareClient,
		FileClient:  fileClient,
	}
}

// JSONResponse 标准JSON响应格式
type JSONResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// writeJSON 写入JSON响应
func (s *GatewayServer) writeJSON(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(http.StatusOK, JSONResponse{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

// writeError 写入错误响应
func (s *GatewayServer) writeError(c *gin.Context, code int, message string) {
	utils.Error("HTTP error: %s", message)
	c.JSON(code, JSONResponse{
		Code:    code,
		Message: message,
	})
}

// AuthMiddleware 鉴权中间件
func (s *GatewayServer) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 这里应该实现JWT验证逻辑
		// 为简化示例，暂时跳过验证
		c.Next()
	}
}

// StartHTTPServer 启动HTTP服务器
func (s *GatewayServer) StartHTTPServer(addr string) error {
	// 设置Gin为发布模式
	gin.SetMode(gin.ReleaseMode)
	
	// 创建Gin路由器
	r := gin.New()
	
	// 添加日志和恢复中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	
	// 注册用户相关路由
	userGroup := r.Group("/api/user")
	{
		userGroup.POST("/register", s.handleUserRegister)
		userGroup.POST("/login", s.handleUserLogin)
		userGroup.GET("/info", s.AuthMiddleware(), s.handleGetUserInfo)
	}
	
	// 注册分享相关路由
	shareGroup := r.Group("/api/share")
	{
		shareGroup.POST("/create", s.AuthMiddleware(), s.handleCreateShare)
		shareGroup.GET("/info", s.handleGetShareInfo)
		shareGroup.POST("/validate", s.handleValidateAccess)
	}
	
	// 注册文件相关路由
	fileGroup := r.Group("/api/file")
	{
		fileGroup.POST("/upload/init", s.AuthMiddleware(), s.handleInitUpload)
		fileGroup.POST("/upload/part", s.AuthMiddleware(), s.handleUploadPart)
		fileGroup.POST("/upload/complete", s.AuthMiddleware(), s.handleCompleteUpload)
		fileGroup.GET("/info", s.AuthMiddleware(), s.handleGetFileInfo)
		fileGroup.POST("/presigned-url", s.AuthMiddleware(), s.handleGeneratePresignedURL)
	}
	
	utils.Info("HTTP server starting on %s", addr)
	return r.Run(addr)
}