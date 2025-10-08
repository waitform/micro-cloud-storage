package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"sync"
	"time"
)

// UploadStat 用于收集上传统计信息
type UploadStat struct {
	filename string
	duration time.Duration
	err      error
}

// DirectUploadResponse 直接上传响应
type DirectUploadResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		File struct {
			ID int64 `json:"id"`
		} `json:"file"`
	} `json:"data"`
}

// DeleteFileRequest 删除文件请求
type DeleteFileRequest struct {
	FileID int64 `json:"file_id"`
}

const (
	defaultBaseURL = "http://localhost:8080/api"
	defaultToken   = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTk5OTIxMDksInVzZXJfaWQiOjEsInVzZXJuYW1lIjoidGVzdHVzZXIifQ.3nKTtB7ojjND2tZ_wkj3gZKyGSnpik-DC4Vcw5VMfe8"
)

var (
	baseURL     = flag.String("url", defaultBaseURL, "API base URL")
	token       = flag.String("token", defaultToken, "Authentication token")
	concurrent  = flag.Int("concurrent", 70, "Number of concurrent uploads")
	fileSizeMB  = flag.Int("filesize", 5, "Size of each file in MB")
	deleteFiles = flag.Bool("delete", true, "Whether to delete files after upload")

	// 全局HTTP客户端，配置连接池和超时
	httpClient *http.Client
)

// generateRandomContent 生成指定大小的随机内容
func generateRandomContent(size int) string {
	// 简单地生成重复的测试内容来模拟文件
	content := make([]byte, size)
	for i := 0; i < size; i++ {
		content[i] = byte('A' + (i % 26))
	}
	return string(content)
}

func init() {
	// 创建带有连接池和超时设置的HTTP客户端
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,              // 最大空闲连接数
		MaxIdleConnsPerHost:   10,               // 每个主机最大空闲连接数
		IdleConnTimeout:       30 * time.Second, // 空闲连接超时时间
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	httpClient = &http.Client{
		Transport: transport,
		Timeout:   60 * time.Second, // 请求超时时间
	}
}

// ... existing code ...

func main() {
	flag.Parse()

	now := time.Now()
	fmt.Printf("开始上传文件... (并发数: %d, 文件大小: %d MB)\n", *concurrent, *fileSizeMB)

	// 创建等待组以并发执行上传任务
	var wg sync.WaitGroup
	fileIDs := make(chan int64, *concurrent)
	stats := make(chan UploadStat, *concurrent)

	// 启动指定数量的协程并发上传文件
	for i := 0; i < *concurrent; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			// 生成随机文件名
			filename := fmt.Sprintf("test_file_%d_%d.txt", time.Now().UnixNano(), index)

			// 记录开始时间
			start := time.Now()

			// 上传文件
			fileID, err := uploadFile(filename)

			// 计算耗时
			duration := time.Since(start)

			// 发送统计信息
			stats <- UploadStat{
				filename: filename,
				duration: duration,
				err:      err,
			}

			if err != nil {
				fmt.Printf("上传文件 %s 失败: %v (耗时: %s)\n", filename, err, duration)
				return
			}

			fmt.Printf("文件 %s 上传成功，ID: %d (耗时: %s)\n", filename, fileID, duration)

			// 将文件ID发送到通道中
			fileIDs <- fileID
		}(i)
	}

	// 等待所有上传完成
	wg.Wait()
	close(fileIDs)
	close(stats)

	// 收集所有文件ID
	var ids []int64
	for id := range fileIDs {
		ids = append(ids, id)
	}

	// 收集统计信息
	var totalDuration time.Duration
	successCount := 0
	failCount := 0

	for stat := range stats {
		if stat.err == nil {
			successCount++
			totalDuration += stat.duration
		} else {
			failCount++
		}
	}

	fmt.Printf("总共上传了 %d 个文件 (成功: %d, 失败: %d)\n", len(ids), successCount, failCount)

	// 计算并显示平均耗时
	if successCount > 0 {
		avgDuration := totalDuration / time.Duration(successCount)
		fmt.Printf("上传成功文件的平均耗时: %s\n", avgDuration)
	}

	// 删除所有上传的文件（如果启用）
	if *deleteFiles && len(ids) > 0 {
		fmt.Println("开始删除文件...")
		for _, id := range ids {
			err := deleteFile(id)
			if err != nil {
				fmt.Printf("删除文件 ID %d 失败: %v\n", id, err)
			} else {
				fmt.Printf("文件 ID %d 删除成功\n", id)
			}
			time.Sleep(10 * time.Millisecond) // 简单限制删除请求频率
		}
	}

	fmt.Println("所有文件已处理完毕")
	elapsed := time.Since(now)
	fmt.Printf("上传文件总耗时: %s\n", elapsed)

	// 显式关闭HTTP客户端的空闲连接
	httpClient.CloseIdleConnections()
}

// ... existing code ...

// uploadFile 上传文件
func uploadFile(filename string) (int64, error) {
	// 创建缓冲区和multipart writer
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 生成指定大小的随机文件内容 (将MB转换为字节)
	fileSizeBytes := *fileSizeMB * 1024 * 1024
	fileContent := generateRandomContent(fileSizeBytes)

	// 创建表单文件字段
	fileWriter, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return 0, err
	}

	// 写入文件内容
	_, err = io.WriteString(fileWriter, fileContent)
	if err != nil {
		return 0, err
	}

	// 添加文件名字段
	err = writer.WriteField("filename", filename)
	if err != nil {
		return 0, err
	}

	// 关闭multipart writer
	err = writer.Close()
	if err != nil {
		return 0, err
	}

	// 创建请求
	req, err := http.NewRequest("POST", *baseURL+"/file/direct-upload", &buf)
	if err != nil {
		return 0, err
	}

	// 设置请求头
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+*token)

	// 发送请求，使用全局HTTP客户端
	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("上传失败，HTTP状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
	}

	// 解析响应
	var uploadResp DirectUploadResponse
	err = json.Unmarshal(respBody, &uploadResp)
	if err != nil {
		return 0, err
	}

	if uploadResp.Code != 200 {
		return 0, fmt.Errorf("上传失败: %s", uploadResp.Message)
	}

	return uploadResp.Data.File.ID, nil
}

// deleteFile 删除文件
func deleteFile(fileID int64) error {
	// 构造删除请求
	deleteReq := DeleteFileRequest{
		FileID: fileID,
	}

	// 序列化请求体
	body, err := json.Marshal(deleteReq)
	if err != nil {
		return err
	}

	// 创建请求
	req, err := http.NewRequest("POST", *baseURL+"/file/delete", bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+*token)

	// 发送请求，使用全局HTTP客户端
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("删除失败，HTTP状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
	}

	// 解析响应
	var deleteResp map[string]interface{}
	err = json.Unmarshal(respBody, &deleteResp)
	if err != nil {
		return err
	}

	code, ok := deleteResp["code"].(float64)
	if !ok || int(code) != 200 {
		return fmt.Errorf("删除失败: %v", deleteResp["message"])
	}

	return nil
}

// ... existing code ...
