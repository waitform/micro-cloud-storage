package concurrentchunkeduploadtest

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var (
	baseURL    = flag.String("url", "http://localhost:8080/api", "API base URL")
	token      = flag.String("token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTk5OTIxMDksInVzZXJfaWQiOjEsInVzZXJuYW1lIjoidGVzdHVzZXIifQ.3nKTtB7ojjND2tZ_wkj3gZKyGSnpik-DC4Vcw5VMfe8", "Authentication token")
	concurrent = flag.Int("concurrent", 100, "Number of concurrent uploads")
	fileSizeMB = flag.Int("filesize", 10, "Size of each file in MB")
)

// TestConcurrentChunkedUpload 测试高并发分片上传功能
func TestConcurrentChunkedUpload(t *testing.T) {
	flag.Parse()
	// 初始化HTTP客户端，配置连接池参数
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          50,
		MaxIdleConnsPerHost:   5,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	t.Logf("开始高并发分片上传测试... (并发数: %d, 文件大小: %d MB)\n", *concurrent, *fileSizeMB)

	// 创建等待组以并发执行上传任务
	var wg sync.WaitGroup
	var successCount int64
	var failCount int64

	startTime := time.Now()

	// 限制并发数，避免同时创建过多连接
	semaphore := make(chan struct{}, 5) // 最多同时运行5个goroutine

	// 启动指定数量的协程并发上传文件
	for i := 0; i < *concurrent; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			semaphore <- struct{}{}        // 获取信号量
			defer func() { <-semaphore }() // 释放信号量

			// 生成随机文件名
			filename := fmt.Sprintf("concurrent_test_file_%d_%d.txt", time.Now().UnixNano(), index)

			// 上传文件
			err := uploadFileChunked(client, filename)
			if err != nil {
				atomic.AddInt64(&failCount, 1)
				t.Logf("分片上传文件 %s 失败: %v", filename, err)
			} else {
				atomic.AddInt64(&successCount, 1)
				t.Logf("分片上传文件 %s 成功", filename)
			}
		}(i)
	}

	// 等待所有上传完成
	wg.Wait()

	// 显式关闭HTTP客户端的空闲连接
	transport.CloseIdleConnections()

	elapsed := time.Since(startTime)
	t.Logf("总共上传了 %d 个文件 (成功: %d, 失败: %d)", *concurrent, successCount, failCount)
	t.Logf("总耗时: %s, 平均每个文件耗时: %s", elapsed, elapsed/time.Duration(*concurrent))
}

// uploadFileChunked 使用分片上传方式上传文件
func uploadFileChunked(client *http.Client, filename string) error {
	// 生成指定大小的随机文件内容 (将MB转换为字节)
	fileSizeBytes := *fileSizeMB * 1024 * 1024
	fileContent := generateRandomContent(fileSizeBytes)

	// 计算整个文件的MD5
	hash := md5.Sum([]byte(fileContent))
	fileMD5 := hex.EncodeToString(hash[:])

	// 1. 初始化上传
	fileID, err := initUpload(client, filename, int64(fileSizeBytes), fileMD5)
	if err != nil {
		return fmt.Errorf("初始化上传失败: %v", err)
	}

	// 2. 分片上传文件
	partSize := 5 * 1024 * 1024 // 5MB per part
	for i := 0; i < fileSizeBytes; i += partSize {
		// 计算当前分片的大小
		currentPartSize := partSize
		if i+partSize > fileSizeBytes {
			currentPartSize = fileSizeBytes - i
		}

		// 读取分片数据
		partData := fileContent[i : i+currentPartSize]

		// 计算分片MD5
		partHash := md5.Sum([]byte(partData))
		partMD5 := hex.EncodeToString(partHash[:])

		// 上传分片
		err := uploadPart(client, fileID, int32(i/partSize+1), []byte(partData), partMD5)
		if err != nil {
			return fmt.Errorf("上传分片 %d 失败: %v", i/partSize+1, err)
		}
	}

	// 3. 完成上传
	err = completeUpload(client, fileID)
	if err != nil {
		return fmt.Errorf("完成上传失败: %v", err)
	}

	// 4. 删除文件以清理资源
	err = deleteFile(client, fileID)
	if err != nil {
		// 这里我们只记录错误，但不返回错误，因为上传本身是成功的
		fmt.Printf("警告：删除文件 %d 失败: %v\n", fileID, err)
	}

	return nil
}

// initUpload 初始化上传
func initUpload(client *http.Client, filename string, size int64, md5 string) (int64, error) {
	// 构造请求
	initReq := map[string]interface{}{
		"file_name": filename,
		"size":      size,
		"md5":       md5,
	}

	initReqBody, err := json.Marshal(initReq)
	if err != nil {
		return 0, err
	}

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 创建请求
	initReqHTTP, err := http.NewRequestWithContext(ctx, "POST", *baseURL+"/file/upload/init", bytes.NewBuffer(initReqBody))
	if err != nil {
		return 0, err
	}

	initReqHTTP.Header.Set("Content-Type", "application/json")
	initReqHTTP.Header.Set("Authorization", "Bearer "+*token)

	initResp, err := client.Do(initReqHTTP)
	if err != nil {
		return 0, err
	}
	defer initResp.Body.Close()

	initRespBody, err := io.ReadAll(initResp.Body)
	if err != nil {
		return 0, err
	}

	if initResp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("初始化上传失败，HTTP状态码: %d, 响应: %s", initResp.StatusCode, string(initRespBody))
	}

	var initUploadResp map[string]interface{}
	err = json.Unmarshal(initRespBody, &initUploadResp)
	if err != nil {
		return 0, err
	}

	code, ok := initUploadResp["code"].(float64)
	if !ok || int(code) != 200 {
		return 0, fmt.Errorf("初始化上传失败: %v", initUploadResp["message"])
	}

	data, ok := initUploadResp["data"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("响应格式错误")
	}

	file, ok := data["file"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("响应格式错误")
	}

	fileID, ok := file["id"].(float64)
	if !ok {
		return 0, fmt.Errorf("无法获取文件ID")
	}

	return int64(fileID), nil
}

// uploadPart 上传单个分片
func uploadPart(client *http.Client, fileID int64, partNumber int32, data []byte, md5 string) error {
	// 创建带超时的上下文，根据数据大小动态调整超时时间
	timeout := time.Duration(len(data)/1024/1024+5) * time.Second
	if timeout < 10*time.Second {
		timeout = 10 * time.Second
	}
	if timeout > 30*time.Second {
		timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "POST", *baseURL+"/file/upload/part", bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	// 设置请求头
	req.Header.Set("Authorization", "Bearer "+*token)
	req.Header.Set("X-File-Id", fmt.Sprintf("%d", fileID))
	req.Header.Set("X-Part-Number", fmt.Sprintf("%d", partNumber))
	req.Header.Set("X-MD5", md5)

	// 发送请求
	resp, err := client.Do(req)
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
		return fmt.Errorf("上传分片失败，HTTP状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
	}

	// 解析响应
	var uploadResp map[string]interface{}
	err = json.Unmarshal(respBody, &uploadResp)
	if err != nil {
		return err
	}

	code, ok := uploadResp["code"].(float64)
	if !ok || int(code) != 200 {
		return fmt.Errorf("上传分片失败: %v", uploadResp["message"])
	}

	return nil
}

// completeUpload 完成上传
func completeUpload(client *http.Client, fileID int64) error {
	// 构造请求
	completeReq := map[string]interface{}{
		"file_id": fileID,
	}

	completeReqBody, err := json.Marshal(completeReq)
	if err != nil {
		return err
	}

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	completeReqHTTP, err := http.NewRequestWithContext(ctx, "POST", *baseURL+"/file/upload/complete", bytes.NewBuffer(completeReqBody))
	if err != nil {
		return err
	}

	completeReqHTTP.Header.Set("Content-Type", "application/json")
	completeReqHTTP.Header.Set("Authorization", "Bearer "+*token)

	completeResp, err := client.Do(completeReqHTTP)
	if err != nil {
		return err
	}
	defer completeResp.Body.Close()

	completeRespBody, err := io.ReadAll(completeResp.Body)
	if err != nil {
		return err
	}

	if completeResp.StatusCode != http.StatusOK {
		return fmt.Errorf("完成上传失败，HTTP状态码: %d, 响应: %s", completeResp.StatusCode, string(completeRespBody))
	}

	return nil
}

// deleteFile 删除文件
func deleteFile(client *http.Client, fileID int64) error {
	// 构造删除请求
	deleteReq := map[string]interface{}{
		"file_id": fileID,
	}

	// 序列化请求体
	body, err := json.Marshal(deleteReq)
	if err != nil {
		return err
	}

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "POST", *baseURL+"/file/delete", bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+*token)

	// 发送请求
	resp, err := client.Do(req)
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

// generateRandomContent 生成指定大小的随机内容
func generateRandomContent(size int) string {
	// 简单地生成重复的测试内容来模拟文件
	content := make([]byte, size)
	for i := 0; i < size; i++ {
		content[i] = byte('A' + (i % 26))
	}
	return string(content)
}
