package utils

import (
	"context"
	"log"
	"sync"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClient struct {
	Client *minio.Client
}

var (
	minioOnce   sync.Once
	globalMinio *MinioClient
)

// InitMinio 初始化 MinIO 客户端（只初始化一次）
func InitMinio(endpoint, accessKey, secretKey string, useSSL bool) *MinioClient {
	minioOnce.Do(func() {
		client, err := minio.New(endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
			Secure: useSSL,
		})
		if err != nil {
			log.Fatalf("❌ 初始化 MinIO 客户端失败: %v", err)
		}
		globalMinio = &MinioClient{Client: client}
		log.Println("✅ MinIO 客户端初始化成功")
	})
	return globalMinio
}

// GetMinio 获取全局客户端
func GetMinio() *MinioClient {
	if globalMinio == nil {
		log.Fatal("MinIO 未初始化，请先调用 InitMinio()")
	}
	return globalMinio
}

// CreateBucket 创建桶（如果不存在）
func (m *MinioClient) CreateBucket(bucketName string) error {
	ctx := context.Background()
	exists, err := m.Client.BucketExists(ctx, bucketName)
	if err != nil {
		return err
	}
	if !exists {
		err = m.Client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return err
		}
		log.Printf("🪣 创建桶成功: %s\n", bucketName)
	}
	return nil
}

// UploadFile 上传本地文件
func (m *MinioClient) UploadFile(bucketName, objectName, filePath string) error {
	ctx := context.Background()
	_, err := m.Client.FPutObject(ctx, bucketName, objectName, filePath, minio.PutObjectOptions{})
	if err != nil {
		return err
	}
	log.Printf("📤 上传成功: %s -> %s\n", filePath, bucketName+"/"+objectName)
	return nil
}

// DownloadFile 下载文件到本地
func (m *MinioClient) DownloadFile(bucketName, objectName, filePath string) error {
	ctx := context.Background()
	return m.Client.FGetObject(ctx, bucketName, objectName, filePath, minio.GetObjectOptions{})
}
