package main

import (
	"cloud-storage/discovery"
	"cloud-storage/gateway/global"
	"cloud-storage/gateway/internal/handler"
	"cloud-storage/gateway/internal/server"
	"cloud-storage/utils"
	"log"
)

var (
	globalCfg     *global.GlobalConfig
	etcdClient    *discovery.EtcdClient
	userClient    *handler.UserServiceClient
	shareClient   *handler.ShareServiceClient
	fileClient    *handler.FileServiceClient
	gatewayServer *server.GatewayServer
)

// 初始化日志
func InitLogger() {
	utils.InitLogger("", "")
}

// 加载全局配置
func LoadGlobalCfg() {
	var err error
	globalCfg, err = global.LoadConfig("/home/haobin/桌面/test/go_test/cloud-storage/gateway/global/global.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
}

// 初始化 etcd 客户端
func initETCD() {
	var err error
	etcdClient, err = discovery.NewEtcdClient(globalCfg.Etcd.Endpoints)
	if err != nil {
		log.Fatalf("failed to init etcd: %v", err)
	}
}

// 初始化服务客户端
func initServiceClients() {
	// 创建基础服务客户端
	serviceClient := handler.NewServiceClient(etcdClient, globalCfg)

	// 初始化用户服务客户端
	var err error
	userClient, err = handler.NewUserServiceClient(serviceClient)
	if err != nil {
		utils.Error("failed to init user service client: %v", err)
	} else {
		utils.Info("user service client initialized")
	}

	// 初始化分享服务客户端
	shareClient, err = handler.NewShareServiceClient(serviceClient)
	if err != nil {
		utils.Error("failed to init share service client: %v", err)
	} else {
		utils.Info("share service client initialized")
	}

	// 初始化文件服务客户端
	fileClient, err = handler.NewFileServiceClient(serviceClient)
	if err != nil {
		utils.Error("failed to init file service client: %v", err)
	} else {
		utils.Info("file service client initialized")
	}
}

// 初始化网关服务器
func initGatewayServer() {
	gatewayServer = server.NewGatewayServer(userClient, shareClient, fileClient)
	utils.Info("gateway server initialized")
}

// 启动HTTP服务器
func startHTTPServer() {
	// 在单独的goroutine中启动HTTP服务器
	go func() {
		err := gatewayServer.StartHTTPServer(":8080")
		if err != nil {
			utils.Error("failed to start HTTP server: %v", err)
		}
	}()
	utils.Info("HTTP server started on :8080")
}

// 统一入口
func main() {
	InitLogger()
	LoadGlobalCfg()
	initETCD()
	utils.Info("etcd client initialized")

	// 初始化服务客户端
	initServiceClients()

	// 初始化网关服务器
	initGatewayServer()

	// 启动HTTP服务器
	startHTTPServer()
	
	utils.Info("gateway started successfully")
	
	// 阻塞主进程
	select {}
}