package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"cloud-storage-file-service/config"
	"cloud-storage-file-service/database"
	"cloud-storage-file-service/discovery"
	"cloud-storage-file-service/global"
	"cloud-storage-file-service/internal/api"
	"cloud-storage-file-service/internal/model"
	"cloud-storage-file-service/internal/service"
	pb "cloud-storage-file-service/proto"
	"cloud-storage-file-service/utils"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"google.golang.org/grpc"
)

var (
	cfg            *config.Config
	minioClient    *minio.Client
	fileDAO        model.FileDAO
	storageService *service.StorageService
	grpcServer     *grpc.Server
	db             *database.DB
)

func main() {
	// 初始化配置
	initConfig()

	// 初始化日志
	initLogger()

	// 初始化数据库
	initDatabase()

	// 初始化MinIO
	initMinIO()

	// 初始化DAO
	initDAO()

	// 初始化Service
	initService()

	// 初始化gRPC服务
	initGRPC()

	// 启动所有服务
	startServices()

	// 等待中断信号以优雅地关闭服务器
	select {}
}

// initConfig 初始化配置
func initConfig() {
	var err error
	cfg, err = config.LoadConfig("config/config.yaml")
	if err != nil {
		panic(fmt.Errorf("failed to load config: %w", err))
	}

	utils.Info("Config loaded successfully")
}

// initLogger 初始化日志
func initLogger() {
	err := utils.InitLogger(cfg.Log.Path, cfg.Log.Level)
	if err != nil {
		panic(fmt.Errorf("failed to initialize logger: %w", err))
	}
	utils.Info("Logger initialized successfully")
}

// initDatabase 初始化数据库
func initDatabase() {
	var err error
	db, err = database.NewDB(cfg)
	if err != nil {
		panic(fmt.Errorf("failed to initialize database: %w", err))
	}
	utils.Info("Database initialized successfully")
}

// initMinIO 初始化MinIO客户端
func initMinIO() {
	// 使用服务配置中的MinIO配置
	minioConfig := cfg.Minio

	// 初始化MinIO客户端
	var err error
	minioClient, err = minio.New(minioConfig.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(minioConfig.AccessKey, minioConfig.SecretKey, ""),
		Secure: minioConfig.UseSSL,
	})

	if err != nil {
		panic(fmt.Errorf("failed to create minio client: %w", err))
	}

	// 检查存储桶是否存在，不存在则创建
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	exists, err := minioClient.BucketExists(ctx, "cloud-storage")
	if err != nil {
		panic(fmt.Errorf("failed to check bucket: %w", err))
	}

	if !exists {
		utils.Info("Creating bucket: cloud-storage")
		err = minioClient.MakeBucket(ctx, "cloud-storage", minio.MakeBucketOptions{})
		if err != nil {
			panic(fmt.Errorf("failed to create bucket: %w", err))
		}
	}

	utils.Info("MinIO initialized successfully")
}

// initDAO 初始化数据访问对象
func initDAO() {
	fileDAO = model.NewFileDAO(db.DB)
	utils.Info("DAO initialized successfully")
}

// initService 初始化业务服务
func initService() {
	storageService = service.NewStorageService(minioClient, "cloud-storage", fileDAO)
	// 设置分片大小
	storageService.SetPartSize(cfg.Storage.PartSize)
	utils.Info("Service initialized successfully")
}

// initGRPC 初始化gRPC服务
func initGRPC() {
	// 增加gRPC消息大小限制
	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(10 * 1024 * 1024), // 10MB
		grpc.MaxSendMsgSize(10 * 1024 * 1024), // 10MB
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC.Port))
	if err != nil {
		panic(fmt.Errorf("failed to listen on port %d: %w", cfg.GRPC.Port, err))
	}

	grpcServer = grpc.NewServer(opts...)
	fileServiceServer := api.NewFileServiceServer(storageService)
	pb.RegisterFileServiceServer(grpcServer, fileServiceServer)

	// 在后台启动gRPC服务
	go func() {
		utils.Info("gRPC server starting on port %d", cfg.GRPC.Port)
		if err := grpcServer.Serve(lis); err != nil {
			utils.Error("Failed to serve gRPC: %v", err)
		}
	}()

	utils.Info("gRPC initialized successfully")
}

// startServices 启动所有服务
func startServices() {
	//注册etcd
	globalCfg, err := global.LoadConfig("global/global.yaml")
	if err != nil {
		utils.Warn("Warning: Failed to load global config: %v", err)
	}
	etcdClient, err := discovery.NewEtcdClient(globalCfg.Etcd.Endpoints)
	if err != nil {
		utils.Warn("Warning: Failed to create etcd client: %v", err)
	}
	etcdClient.Register("file-service", fmt.Sprintf("localhost:%d", cfg.GRPC.Port), 5)

	// 启动HTTP健康检查服务
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// 在后台启动HTTP服务
	go func() {
		utils.Info("HTTP server starting on port %d", cfg.Server.Port)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Server.Port), nil); err != nil {
			utils.Error("Failed to start HTTP server: %v", err)
		}
	}()

	utils.Info("All services started successfully")
}
