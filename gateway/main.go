package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/waitform/micro-cloud-storage/discovery"
	"github.com/waitform/micro-cloud-storage/discovery/resolver"
	"github.com/waitform/micro-cloud-storage/global"
	"github.com/waitform/micro-cloud-storage/internal/api"
	"github.com/waitform/micro-cloud-storage/internal/casbin"
	"github.com/waitform/micro-cloud-storage/internal/rpc"
	"github.com/waitform/micro-cloud-storage/utils"
)

var (
	globalCfg     *global.GlobalConfig
	etcdClient    *discovery.EtcdClient
	userClient    *rpc.UserServiceClient
	shareClient   *rpc.ShareServiceClient
	fileClient    *rpc.FileServiceClient
	gatewayServer *api.GatewayServer
	serviceClient *rpc.ServiceClient
)

// 初始化日志
func InitLogger() {
	utils.InitLogger("", "")
}

// 加载全局配置
func LoadGlobalCfg() {
	var err error
	globalCfg, err = global.LoadConfig()
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

	// 注册etcd解析器
	resolver.RegisterEtcdResolver(etcdClient)
}

// 初始化服务客户端
func initServiceClients() {
	// 创建基础服务客户端
	serviceClient = rpc.NewServiceClient(etcdClient, globalCfg)

	// 初始化用户服务客户端
	var err error
	userClient, err = rpc.NewUserServiceClient(serviceClient)
	if err != nil {
		utils.Error("failed to init user service client: %v", err)
	} else {
		utils.Info("user service client initialized")
	}

	// 初始化分享服务客户端
	shareClient, err = rpc.NewShareServiceClient(serviceClient)
	if err != nil {
		utils.Error("failed to init share service client: %v", err)
	} else {
		utils.Info("share service client initialized")
	}

	// 初始化文件服务客户端
	fileClient, err = rpc.NewFileServiceClient(serviceClient)
	if err != nil {
		utils.Error("failed to init file service client: %v", err)
	} else {
		utils.Info("file service client initialized")
	}
}

// 关闭服务客户端
func closeServiceClients() {
	if userClient != nil {
		userClient.Close()
	}
	if shareClient != nil {
		shareClient.Close()
	}
	if fileClient != nil {
		fileClient.Close()
	}
	if serviceClient != nil {
		serviceClient.Close()
	}
	utils.Info("service clients closed")
}

// 初始化网关服务器
func initGatewayServer() {
	gatewayServer = api.NewGatewayServer(userClient, shareClient, fileClient)
	utils.Info("gateway server initialized")
}
func initCasbin() {
	casbin.InitCasbin()
}

// 启动HTTP服务器
func startHTTPServer() {
	// 在单独的goroutine中启动HTTP服务器
	go func() {
		err := gatewayServer.StartHTTPServer("0.0.0.0:8080")
		if err != nil {
			utils.Error("failed to start HTTP server: %v", err)
		}
	}()
	utils.Info("HTTP server started on 0.0.0.0:8080")
}
func startPprofServer() {
	go func() {
		utils.Info("starting pprof server on 0.0.0.0:6060")
		if err := http.ListenAndServe(":6060", nil); err != nil {
			utils.Error("failed to start pprof server: %v", err)
		}
	}()
}

// 统一入口
func main() {
	InitLogger()
	LoadGlobalCfg()
	initETCD()
	utils.Info("etcd client initialized")

	// 初始化服务客户端
	initServiceClients()
	defer closeServiceClients()

	// 初始化网关服务器
	initGatewayServer()

	// 启动HTTP服务器
	startHTTPServer()

	startPprofServer()

	// 等待中断信号以优雅关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	utils.Info("shutting down server...")

	// 关闭服务客户端
	closeServiceClients()
	utils.Info("server exited")
}
