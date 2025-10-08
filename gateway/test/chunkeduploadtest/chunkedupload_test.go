package chunkeduploadtest

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
)

// TestChunkedUpload 测试分片上传功能
func TestChunkedUpload(t *testing.T) {
	// 生成测试文件内容
	fileSizeBytes := 5 * 1024 * 1024 // 5MB
	fileContent := generateRandomContent(fileSizeBytes)
	filename := fmt.Sprintf("test_file_%d.txt", time.Now().UnixNano())

	// 计算整个文件的MD5
	hash := md5.Sum([]byte(fileContent))
	fileMD5 := hex.EncodeToString(hash[:])

	// 1. 初始化上传
	fileID, err := initUpload(filename, int64(fileSizeBytes), fileMD5)
	if err != nil {
		t.Fatalf("初始化上传失败: %v", err)
	}
	t.Logf("初始化上传成功，文件ID: %d", fileID)

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
		err := uploadPart(fileID, int32(i/partSize+1), []byte(partData), partMD5)
		if err != nil {
			t.Fatalf("上传分片 %d 失败: %v", i/partSize+1, err)
		}
		t.Logf("上传分片 %d 成功", i/partSize+1)
	}

	// 3. 完成上传
	err = completeUpload(fileID)
	if err != nil {
		t.Fatalf("完成上传失败: %v", err)
	}
	t.Log("完成上传成功")

	// 4. 删除文件
	err = deleteFile(fileID)
	if err != nil {
		t.Fatalf("删除文件失败: %v", err)
	}
	t.Log("删除文件成功")
}

// initUpload 初始化上传
func initUpload(filename string, size int64, md5 string) (int64, error) {
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

	initReqHTTP, err := http.NewRequest("POST", "http://localhost:8080/api/file/upload/init", bytes.NewBuffer(initReqBody))
	if err != nil {
		return 0, err
	}

	initReqHTTP.Header.Set("Content-Type", "application/json")
	initReqHTTP.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTk5OTIxMDksInVzZXJfaWQiOjEsInVzZXJuYW1lIjoidGVzdHVzZXIifQ.3nKTtB7ojjND2tZ_wkj3gZKyGSnpik-DC4Vcw5VMfe8")

	client := &http.Client{}
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
func uploadPart(fileID int64, partNumber int32, data []byte, md5 string) error {
	// 创建请求
	req, err := http.NewRequest("POST", "http://localhost:8080/api/file/upload/part", bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	// 设置请求头
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTk5OTIxMDksInVzZXJfaWQiOjEsInVzZXJuYW1lIjoidGVzdHVzZXIifQ.3nKTtB7ojjND2tZ_wkj3gZKyGSnpik-DC4Vcw5VMfe8")
	req.Header.Set("X-File-Id", fmt.Sprintf("%d", fileID))
	req.Header.Set("X-Part-Number", fmt.Sprintf("%d", partNumber))
	req.Header.Set("X-MD5", md5)

	// 发送请求
	client := &http.Client{}
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
func completeUpload(fileID int64) error {
	// 构造请求
	completeReq := map[string]interface{}{
		"file_id": fileID,
	}

	completeReqBody, err := json.Marshal(completeReq)
	if err != nil {
		return err
	}

	completeReqHTTP, err := http.NewRequest("POST", "http://localhost:8080/api/file/upload/complete", bytes.NewBuffer(completeReqBody))
	if err != nil {
		return err
	}

	completeReqHTTP.Header.Set("Content-Type", "application/json")
	completeReqHTTP.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTk5OTIxMDksInVzZXJfaWQiOjEsInVzZXJuYW1lIjoidGVzdHVzZXIifQ.3nKTtB7ojjND2tZ_wkj3gZKyGSnpik-DC4Vcw5VMfe8")

	client := &http.Client{}
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
func deleteFile(fileID int64) error {
	// 构造删除请求
	deleteReq := map[string]interface{}{
		"file_id": fileID,
	}

	// 序列化请求体
	body, err := json.Marshal(deleteReq)
	if err != nil {
		return err
	}

	// 创建请求
	req, err := http.NewRequest("POST", "http://localhost:8080/api/file/delete", bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTk5OTIxMDksInVzZXJfaWQiOjEsInVzZXJuYW1lIjoidGVzdHVzZXIifQ.3nKTtB7ojjND2tZ_wkj3gZKyGSnpik-DC4Vcw5VMfe8")

	// 发送请求
	client := &http.Client{}
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
