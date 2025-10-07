package main

import (
	"fmt"
	"log"
	"net"

	"cloud-storage-user-service/config"
	"cloud-storage-user-service/discovery"
	"cloud-storage-user-service/global"
	"cloud-storage-user-service/internal/api"
	"cloud-storage-user-service/internal/database"
	"cloud-storage-user-service/internal/model"
	"cloud-storage-user-service/internal/service"
	"cloud-storage-user-service/proto"
	"cloud-storage-user-service/utils"

	"google.golang.org/grpc"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库
	db, err := database.NewDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移 User 模型
	if err := db.AutoMigrate(&model.User{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 初始化 UserDAO
	userDAO := model.NewUserDAO(db.DB)

	// 初始化 UserService
	userService := service.NewUserService(userDAO, cfg)

	// 初始化 gRPC 服务端
	grpcServer := grpc.NewServer()
	userServer := api.NewUserServiceServer(userService)
	proto.RegisterUserServiceServer(grpcServer, userServer)

	//注册etcd
	utils.Info(" Registering user service...")
	globalCfg, err := global.LoadConfig("global/global.yaml")
	if err != nil {
		utils.Warn("Warning: Failed to load global config: %v", err)
	}
	etcdClient, err := discovery.NewEtcdClient(globalCfg.Etcd.Endpoints)
	if err != nil {
		utils.Warn("Warning: Failed to create etcd client: %v", err)
	}
	etcdClient.Register("user-service", fmt.Sprintf("localhost:%d", cfg.GRPC.Port), 5)
	utils.Info(" user-service registered successfully")
	// 监听端口
	lis, err := net.Listen("tcp", ":"+cfg.Server.Port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", cfg.Server.Port, err)
	}

	fmt.Printf("User service is running on port %s\n", cfg.Server.Port)

	// 启动 gRPC 服务
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC: %v", err)
	}
}
