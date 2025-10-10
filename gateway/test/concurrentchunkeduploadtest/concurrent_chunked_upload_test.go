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
	token      = flag.String("token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NjAxNzM0MzAsInVzZXJfaWQiOjEsInVzZXJuYW1lIjoidGVzdHVzZXIifQ.PXr65sH2CABNBa8R43xU7xkjqiJg265YzfP-EBEYopc", "Authentication token")
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
			Timeout:   30 * time.Second,  // 增加连接超时时间
			KeepAlive: 100 * time.Second, // 增加KeepAlive时间
		}).DialContext,
		MaxIdleConns:          200,              // 增加最大空闲连接数
		MaxIdleConnsPerHost:   20,               // 增加每个主机的最大空闲连接数
		IdleConnTimeout:       90 * time.Second, // 增加空闲连接超时时间
		TLSHandshakeTimeout:   10 * time.Second, // 增加TLS握手超时时间
		ExpectContinueTimeout: 5 * time.Second,  // 增加ExpectContinue超时时间
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   60 * time.Second, // 增加整体请求超时时间
	}

	t.Logf("开始高并发分片上传测试... (并发数: %d, 文件大小: %d MB)\n", *concurrent, *fileSizeMB)

	// 创建等待组以并发执行上传任务
	var wg sync.WaitGroup
	var successCount int64
	var failCount int64

	startTime := time.Now()

	// 限制并发数，避免同时创建过多连接
	semaphore := make(chan struct{}, 10) // 增加并发限制到10个goroutine

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
				t.Logf("❌ 分片上传文件 %s 失败: %v", filename, err)
			} else {
				atomic.AddInt64(&successCount, 1)
				t.Logf("✅ 分片上传文件 %s 成功", filename)
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

	// 如果失败率超过30%，则标记测试失败
	if failCount > int64(*concurrent*30/100) {
		t.Errorf("失败率过高: %d/%d", failCount, *concurrent)
	}
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
	var partErrors []error

	// 使用互斥锁保护共享资源
	var mu sync.Mutex

	// 使用WaitGroup并行上传分片
	var partWg sync.WaitGroup
	partErrorChan := make(chan error, fileSizeBytes/partSize+1)

	// 限制分片上传的并发数
	partSemaphore := make(chan struct{}, 5) // 最多同时上传5个分片

	for i := 0; i < fileSizeBytes; i += partSize {
		partWg.Add(1)
		go func(startIndex int) {
			defer partWg.Done()
			partSemaphore <- struct{}{}        // 获取信号量
			defer func() { <-partSemaphore }() // 释放信号量

			// 计算当前分片的大小
			currentPartSize := partSize
			if startIndex+partSize > fileSizeBytes {
				currentPartSize = fileSizeBytes - startIndex
			}

			// 读取分片数据
			partData := fileContent[startIndex : startIndex+currentPartSize]

			// 计算分片MD5
			partHash := md5.Sum([]byte(partData))
			partMD5 := hex.EncodeToString(partHash[:])

			// 上传分片
			err := uploadPart(client, fileID, int32(startIndex/partSize+1), []byte(partData), partMD5)
			if err != nil {
				partErrorChan <- fmt.Errorf("上传分片 %d 失败: %v", startIndex/partSize+1, err)
			}
		}(i)
	}

	// 等待所有分片上传完成
	partWg.Wait()
	close(partErrorChan)

	// 收集所有错误
	for err := range partErrorChan {
		mu.Lock()
		partErrors = append(partErrors, err)
		mu.Unlock()
	}

	// 如果有任何分片上传失败，返回错误
	if len(partErrors) > 0 {
		return fmt.Errorf("分片上传失败: %v", partErrors)
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
		fmt.Printf("⚠️ 警告：删除文件 %d 失败: %v\n", fileID, err)
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 创建请求
	initReqHTTP, err := http.NewRequestWithContext(ctx, "POST", *baseURL+"/file/upload/init", bytes.NewBuffer(initReqBody))
	if err != nil {
		return 0, err
	}

	initReqHTTP.Header.Set("Content-Type", "application/json")
	initReqHTTP.Header.Set("Authorization", "Bearer "+*token)

	// 添加重试机制
	var initResp *http.Response
	for retry := 0; retry < 3; retry++ {
		initResp, err = client.Do(initReqHTTP)
		if err == nil && initResp.StatusCode == http.StatusOK {
			break
		}

		if initResp != nil {
			initResp.Body.Close()
		}

		time.Sleep(time.Duration(retry+1) * time.Second) // 递增重试间隔
	}

	if err != nil {
		return 0, fmt.Errorf("请求失败，已重试3次: %v", err)
	}

	if initResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(initResp.Body)
		initResp.Body.Close()
		return 0, fmt.Errorf("初始化上传失败，HTTP状态码: %d, 响应: %s", initResp.StatusCode, string(body))
	}

	defer initResp.Body.Close()

	initRespBody, err := io.ReadAll(initResp.Body)
	if err != nil {
		return 0, err
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
	timeout := time.Duration(len(data)/1024/1024+10) * time.Second
	if timeout < 30*time.Second {
		timeout = 30 * time.Second
	}
	if timeout > 60*time.Second {
		timeout = 60 * time.Second
	}

	// 添加重试机制
	var resp *http.Response
	var err error

	for retry := 0; retry < 3; retry++ {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)

		// 创建请求
		req, reqErr := http.NewRequestWithContext(ctx, "POST", *baseURL+"/file/upload/part", bytes.NewBuffer(data))
		if reqErr != nil {
			cancel()
			return reqErr
		}

		// 设置请求头
		req.Header.Set("Authorization", "Bearer "+*token)
		req.Header.Set("X-File-Id", fmt.Sprintf("%d", fileID))
		req.Header.Set("X-Part-Number", fmt.Sprintf("%d", partNumber))
		req.Header.Set("X-MD5", md5)

		// 发送请求
		resp, err = client.Do(req)
		cancel()

		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}

		if resp != nil {
			resp.Body.Close()
		}

		time.Sleep(time.Duration(retry+1) * time.Second) // 递增重试间隔
	}

	if err != nil {
		return fmt.Errorf("请求失败，已重试3次: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return fmt.Errorf("上传分片失败，HTTP状态码: %d, 响应: %s", resp.StatusCode, string(body))
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
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second) // 增加超时时间
	defer cancel()

	completeReqHTTP, err := http.NewRequestWithContext(ctx, "POST", *baseURL+"/file/upload/complete", bytes.NewBuffer(completeReqBody))
	if err != nil {
		return err
	}

	completeReqHTTP.Header.Set("Content-Type", "application/json")
	completeReqHTTP.Header.Set("Authorization", "Bearer "+*token)

	// 添加重试机制
	var completeResp *http.Response
	for retry := 0; retry < 3; retry++ {
		completeResp, err = client.Do(completeReqHTTP)
		if err == nil && completeResp.StatusCode == http.StatusOK {
			break
		}

		if completeResp != nil {
			completeResp.Body.Close()
		}

		time.Sleep(time.Duration(retry+1) * time.Second) // 递增重试间隔
	}

	if err != nil {
		return fmt.Errorf("请求失败，已重试3次: %v", err)
	}

	if completeResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(completeResp.Body)
		completeResp.Body.Close()
		return fmt.Errorf("完成上传失败，HTTP状态码: %d, 响应: %s", completeResp.StatusCode, string(body))
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
