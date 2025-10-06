package main

import (
	"log"
	"net"

	"cloud-storage/services/user_service/config"
	"cloud-storage/services/user_service/internal/api"
	"cloud-storage/services/user_service/internal/database"
	"cloud-storage/services/user_service/internal/model"
	"cloud-storage/services/user_service/internal/service"
	"cloud-storage/services/user_service/internal/utils"
	pb "cloud-storage/services/user_service/proto"

	"google.golang.org/grpc"
)

func main() {
	// 加载配置
	config.LoadConfig()

	// 设置JWT密钥
	utils.SetSecret([]byte(config.AppConfig.JWT.Secret))
	utils.InitLogger("", "")

	// 初始化数据库
	db, err := database.NewDB(config.AppConfig)
	if err != nil {
		utils.Error("初始化数据库失败: %v", err)
		log.Fatalf("初始化数据库失败: %v", err)
	}

	// 初始化用户DAO
	userDAO := model.NewUserDAO(db.DB)

	// 初始化用户服务
	userService := service.NewUserService(userDAO)

	// 创建gRPC服务端，不添加JWT拦截器（由网关统一处理鉴权）
	grpcServer := grpc.NewServer()

	// 注册用户服务
	pb.RegisterUserServiceServer(grpcServer, api.NewUserServiceServer(userService))

	// 监听端口
	lis, err := net.Listen("tcp", ":"+config.AppConfig.Server.Port)
	if err != nil {
		utils.Error("监听端口失败: %v", err)
		log.Fatalf("监听端口失败: %v", err)
	}

	utils.Info("用户服务启动成功，监听端口: %s", config.AppConfig.Server.Port)
	log.Printf("用户服务启动成功，监听端口: %s", config.AppConfig.Server.Port)

	// 启动gRPC服务
	if err := grpcServer.Serve(lis); err != nil {
		utils.Error("启动gRPC服务失败: %v", err)
		log.Fatalf("启动gRPC服务失败: %v", err)
	}
}