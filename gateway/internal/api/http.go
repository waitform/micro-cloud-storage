package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/waitform/micro-cloud-storage/internal/api/handler"
	"github.com/waitform/micro-cloud-storage/internal/router"
	"github.com/waitform/micro-cloud-storage/internal/rpc"
	"github.com/waitform/micro-cloud-storage/utils"
	"golang.org/x/time/rate"
)

// GatewayServer 网关服务器结构
type GatewayServer struct {
	UserHandler   *handler.UserHandler
	ShareHandler  *handler.ShareHandler
	FileHandler   *handler.FileHandler
	ShareClient   *rpc.ShareServiceClient
	IPRateLimiter *utils.IPRateLimiter
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

	// 创建IP限流器 (每秒10个请求，突发20个)
	ipRateLimiter := utils.NewIPRateLimiter(rate.Limit(10), 20)

	return &GatewayServer{
		UserHandler:   userHandler,
		ShareHandler:  shareHandler,
		FileHandler:   fileHandler,
		ShareClient:   shareClient,
		IPRateLimiter: ipRateLimiter,
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
	server := &http.Server{
		Addr:           addr,
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    30 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}
	// 注册路由，直接传递handler实例
	router.RegisterRoutes(r, s.UserHandler, s.ShareHandler, s.FileHandler, s.ShareClient, s.IPRateLimiter)

	utils.Info("HTTP server starting on %s", addr)
	return server.ListenAndServe()
}
