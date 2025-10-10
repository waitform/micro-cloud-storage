package main

import (
	"context"
	"crypto/md5"
	"fmt"
	"log"

	filepb "cloud-storage-file-service/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// 连接到gRPC服务
	conn, err := grpc.Dial("localhost:35001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("无法连接到gRPC服务: %v", err)
	}
	defer conn.Close()

	// 创建gRPC客户端
	client := filepb.NewFileServiceClient(conn)

	// 测试文件上传
	err = testFileUpload(client)
	if err != nil {
		log.Fatalf("文件上传测试失败: %v", err)
	}

	fmt.Println("文件上传测试成功完成！")
}

func testFileUpload(client filepb.FileServiceClient) error {
	ctx := context.Background()

	// 1. 初始化上传
	fmt.Println("步骤1: 初始化上传...")
	// 创建一个12MB的测试文件内容
	fileContent := make([]byte, 12*1024*1024)
	for i := range fileContent {
		fileContent[i] = byte(i % 256)
	}

	fileMD5 := fmt.Sprintf("%x", md5.Sum(fileContent))
	fileSize := int64(len(fileContent))

	fmt.Printf("文件大小: %d 字节 (%.2f MB)\n", fileSize, float64(fileSize)/(1024*1024))
	fmt.Printf("文件MD5: %s\n", fileMD5)

	initReq := &filepb.InitUploadRequest{
		FileName: "test-file-12mb.bin",
		Size:     fileSize,
		Md5:      fileMD5,
		UserID:   1,
	}

	initResp, err := client.InitUpload(ctx, initReq)
	if err != nil {
		return fmt.Errorf("初始化上传失败: %v", err)
	}

	fileID := initResp.File.Id
	fmt.Printf("上传初始化成功，文件ID: %d\n", fileID)
	fmt.Printf("文件状态: %d\n", initResp.File.Status)

	// 2. 上传分片
	fmt.Println("步骤2: 上传分片...")
	// 将文件分成5MB的分片
	partSize := 5 * 1024 * 1024 // 5MB
	partCount := (len(fileContent) + partSize - 1) / partSize

	fmt.Printf("文件将被分成 %d 个分片，每个分片最大 %d 字节\n", partCount, partSize)

	for i := 0; i < partCount; i++ {
		start := i * partSize
		end := start + partSize
		if end > len(fileContent) {
			end = len(fileContent)
		}

		partData := fileContent[start:end]
		partMD5 := fmt.Sprintf("%x", md5.Sum(partData))

		fmt.Printf("上传第 %d 个分片，大小: %d 字节\n", i+1, len(partData))
		err = uploadPart(client, fileID, int32(i+1), partData, partMD5)
		if err != nil {
			return fmt.Errorf("上传第 %d 个分片失败: %v", i+1, err)
		}
	}

	fmt.Println("所有分片上传成功")

	// 3. 完成上传
	fmt.Println("步骤3: 完成上传...")
	completeReq := &filepb.CompleteUploadRequest{
		FileId: fileID,
	}

	completeResp, err := client.CompleteUpload(ctx, completeReq)
	if err != nil {
		return fmt.Errorf("完成上传失败: %v", err)
	}

	fmt.Printf("上传完成，文件状态: %d\n", completeResp.File.Status)

	// 4. 验证文件信息
	fmt.Println("步骤4: 验证文件信息...")
	infoReq := &filepb.GetFileInfoRequest{
		FileId: fileID,
	}

	infoResp, err := client.GetFileInfo(ctx, infoReq)
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %v", err)
	}

	fmt.Printf("文件信息验证成功:\n")
	fmt.Printf("  文件名: %s\n", infoResp.File.Name)
	fmt.Printf("  文件大小: %d\n", infoResp.File.Size)
	fmt.Printf("  文件MD5: %s\n", infoResp.File.Md5)
	fmt.Printf("  文件状态: %d\n", infoResp.File.Status)

	return nil
}

func uploadPart(client filepb.FileServiceClient, fileID int64, partNumber int32, data []byte, md5Str string) error {
	// 创建上传流
	stream, err := client.UploadPart(context.Background())
	if err != nil {
		return fmt.Errorf("创建上传流失败: %v", err)
	}

	// 发送元数据
	metaReq := &filepb.UploadPartRequest{
		PartData: &filepb.UploadPartRequest_PartMetadata{
			PartMetadata: &filepb.PartMetadata{
				FileId:     fileID,
				PartNumber: int64(partNumber),
				Size:       int64(len(data)),
				Md5:        md5Str,
			},
		},
	}

	fmt.Printf("发送第 %d 个分片元数据: fileID=%d, partNumber=%d, size=%d, md5=%s\n",
		partNumber, fileID, partNumber, len(data), md5Str)

	err = stream.Send(metaReq)
	if err != nil {
		return fmt.Errorf("发送元数据失败: %v", err)
	}

	// 分块发送数据，每次发送1KB
	bufferSize := 1024
	totalSent := 0
	for i := 0; i < len(data); i += bufferSize {
		end := i + bufferSize
		if end > len(data) {
			end = len(data)
		}

		dataReq := &filepb.UploadPartRequest{
			PartData: &filepb.UploadPartRequest_PartContent{
				PartContent: &filepb.PartContent{
					Data: data[i:end],
				},
			},
		}

		err = stream.Send(dataReq)
		if err != nil {
			return fmt.Errorf("发送数据失败: %v", err)
		}

		totalSent += len(data[i:end])
	}

	fmt.Printf("第 %d 个分片总共发送数据: %d 字节\n", partNumber, totalSent)

	// 关闭流并接收响应
	_, err = stream.CloseAndRecv()
	if err != nil {
		return fmt.Errorf("关闭流失败: %v", err)
	}

	return nil
}
