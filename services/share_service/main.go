package main

import (
	"cloud-storage-share-service/database"
	"cloud-storage-share-service/discovery"
	"cloud-storage-share-service/global"
	"cloud-storage-share-service/internal/api"
	"cloud-storage-share-service/internal/config"
	"cloud-storage-share-service/internal/model"
	pb "cloud-storage-share-service/proto"
	"cloud-storage-share-service/utils"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig("configs/config.yml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	if err := utils.InitLogger(cfg.Log.Path, cfg.Log.Level); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer utils.CloseLogger()

	utils.Info("Starting share service...")

	// 初始化数据库
	log.Println("[DEBUG] Connecting to database...")
	db, err := database.NewDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("[DEBUG] Database connected successfully")

	// 创建ShareDAO实例
	log.Println("[DEBUG] Creating ShareDAO instance...")
	shareDAO := model.NewShareDAO(db.DB)
	log.Println("[DEBUG] ShareDAO instance created")

	// 创建gRPC服务器
	log.Printf("[DEBUG] Creating gRPC server on port %d...", cfg.GRPC.Port)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC.Port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	//注册etcd
	globalCfg, err := global.LoadConfig("global/global.yaml")
	if err != nil {
		utils.Warn("Warning: Failed to load global config: %v", err)
	}
	etcdClient, err := discovery.NewEtcdClient(globalCfg.Etcd.Endpoints)
	if err != nil {
		utils.Warn("Warning: Failed to create etcd client: %v", err)
	}
	etcdClient.Register("share-service", fmt.Sprintf("localhost:%d", cfg.GRPC.Port), 5)

	// 注册服务
	log.Println("[DEBUG] Registering share service...")
	shareServer := api.NewShareHandler(shareDAO)
	pb.RegisterShareServiceServer(grpcServer, shareServer)
	log.Println("[DEBUG] Share service registered successfully")

	// 注册反射服务
	log.Println("[DEBUG] Registering reflection service...")
	reflection.Register(grpcServer)
	log.Println("[DEBUG] Reflection service registered successfully")

	log.Printf("Share service started on port %d", cfg.GRPC.Port)

	// 启动服务
	log.Println("[DEBUG] Starting gRPC server...")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
