package api

import (
	"cloud-storage/internal/api/handler"
	"cloud-storage/internal/router"
	"cloud-storage/internal/rpc"
	"cloud-storage/utils"

	"github.com/gin-gonic/gin"
)

// GatewayServer 网关服务器结构
type GatewayServer struct {
	UserHandler  *handler.UserHandler
	ShareHandler *handler.ShareHandler
	FileHandler  *handler.FileHandler
}

// NewGatewayServer 创建网关服务器实例
func NewGatewayServer(
	userClient *rpc.UserServiceClient,
	shareClient *rpc.ShareServiceClient,
	fileClient *rpc.FileServiceClient,
) *GatewayServer {
	// 创建处理器实例
	userHandler := handler.NewUserHandler(userClient)
	shareHandler := handler.NewShareHandler(shareClient)
	fileHandler := handler.NewFileHandler(fileClient)

	return &GatewayServer{
		UserHandler:  userHandler,
		ShareHandler: shareHandler,
		FileHandler:  fileHandler,
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

	// 注册路由，直接传递handler实例
	router.RegisterRoutes(r, s.UserHandler, s.ShareHandler, s.FileHandler)

	utils.Info("HTTP server starting on %s", addr)
	return r.Run(addr)
}
