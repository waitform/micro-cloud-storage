package test

import (
	pb "cloud-storage-file-service/proto"
	"cloud-storage-file-service/utils"
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// 集成测试：测试完整的文件操作流程（上传->下载->生成预签名URL->删除）
func TestGRPCFileOperations(t *testing.T) {
	// 初始化日志
	utils.InitLogger("./logs", "")

	// 连接到gRPC服务器
	conn, err := grpc.Dial("localhost:35001",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(10*1024*1024), // 10MB
			grpc.MaxCallSendMsgSize(10*1024*1024), // 10MB
		),
	)
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	// 创建客户端
	client := pb.NewFileServiceClient(conn)

	// 生成测试文件数据
	fileName := "integration_test_file.txt"
	fileSize := int64(5 * 1024 * 1024) // 5MB
	fileData := make([]byte, fileSize)
	for i := range fileData {
		fileData[i] = byte(i % 256)
	}

	// 计算整个文件的MD5
	hash := md5.Sum(fileData)
	fileMD5 := fmt.Sprintf("%x", hash)

	var fileID int64

	// 1. 上传文件
	t.Run("UploadFile", func(t *testing.T) {
		// 初始化上传
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		initResp, err := client.InitUpload(ctx, &pb.InitUploadRequest{
			FileName: fileName,
			Size:     fileSize,
			Md5:      fileMD5,
		})
		if err != nil {
			t.Fatalf("初始化上传失败: %v", err)
		}

		utils.Info("初始化上传成功，文件ID: %d", initResp.File.Id)

		// 上传分片
		fileID = initResp.File.Id
		partData := fileData[:fileSize] // 使用整个文件作为单个分片

		partHash := md5.Sum(partData)
		partMD5 := fmt.Sprintf("%x", partHash)

		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		_, err = client.UploadPart(ctx, &pb.UploadPartRequest{
			FileId:     fileID,
			PartNumber: 1,
			Data:       partData,
			Md5:        partMD5,
		})
		if err != nil {
			t.Fatalf("上传分片失败: %v", err)
		}

		utils.Info("上传分片成功")

		// 完成上传
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err = client.CompleteUpload(ctx, &pb.CompleteUploadRequest{
			FileId: fileID,
		})
		if err != nil {
			t.Fatalf("完成上传失败: %v", err)
		}

		utils.Info("完成上传成功，文件ID: %d", fileID)
	})

	// 确保文件ID已设置
	if fileID == 0 {
		t.Fatal("文件上传失败，未获取到文件ID")
	}

	// 2. 下载文件
	t.Run("DownloadFile", func(t *testing.T) {
		// 下载文件分片
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		stream, err := client.DownloadPart(ctx, &pb.DownloadRequest{
			FileId:     fileID,
			PartNumber: 1,
		})
		if err != nil {
			t.Fatalf("初始化下载失败: %v", err)
		}

		// 接收数据
		totalBytes := int64(0)
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("接收数据失败: %v", err)
			}

			totalBytes += int64(len(resp.Data))
		}

		utils.Info("下载完成，总字节数: %d", totalBytes)
	})

	// 3. 生成预签名URL
	t.Run("GeneratePresignedURL", func(t *testing.T) {
		// 生成预签名URL
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		resp, err := client.GeneratePresignedURL(ctx, &pb.GeneratePresignedURLRequest{
			FileId:        fileID,
			ExpireSeconds: 3600, // 1小时过期
		})
		if err != nil {
			t.Fatalf("生成预签名URL失败: %v", err)
		}

		utils.Info("生成预签名URL成功: %s, 过期时间: %d", resp.Url, resp.ExpireAt)
	})

	// 4. 删除文件
	t.Run("DeleteFile", func(t *testing.T) {
		// 删除文件
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err = client.DeleteFile(ctx, &pb.DeleteRequest{
			FileId: fileID,
		})
		if err != nil {
			t.Fatalf("删除文件失败: %v", err)
		}

		utils.Info("删除文件成功")
	})
}
