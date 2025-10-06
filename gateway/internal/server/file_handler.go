package server

import (
	filepb "cloud-storage/protos/file/proto"
	"context"
	"net/http"
	"strconv"

	"cloud-storage/utils"
	"github.com/gin-gonic/gin"
)

// handleInitUpload 处理初始化上传请求
func (s *GatewayServer) handleInitUpload(c *gin.Context) {
	var req filepb.InitUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.writeError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := context.Background()
	resp, err := s.FileClient.InitUpload(ctx, &req)
	if err != nil {
		utils.Error("Failed to init upload: %v", err)
		s.writeError(c, http.StatusInternalServerError, "Failed to init upload")
		return
	}

	s.writeJSON(c, http.StatusOK, "Upload initialized successfully", resp)
}

// handleUploadPart 处理上传分片请求
func (s *GatewayServer) handleUploadPart(c *gin.Context) {
	var req filepb.UploadPartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.writeError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := context.Background()
	_, err := s.FileClient.UploadPart(ctx, &req)
	if err != nil {
		utils.Error("Failed to upload part: %v", err)
		s.writeError(c, http.StatusInternalServerError, "Failed to upload part")
		return
	}

	s.writeJSON(c, http.StatusOK, "Part uploaded successfully", nil)
}

// handleCompleteUpload 处理完成上传请求
func (s *GatewayServer) handleCompleteUpload(c *gin.Context) {
	var req filepb.CompleteUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.writeError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := context.Background()
	_, err := s.FileClient.CompleteUpload(ctx, &req)
	if err != nil {
		utils.Error("Failed to complete upload: %v", err)
		s.writeError(c, http.StatusInternalServerError, "Failed to complete upload")
		return
	}

	s.writeJSON(c, http.StatusOK, "Upload completed successfully", nil)
}

// handleGetFileInfo 处理获取文件信息请求
func (s *GatewayServer) handleGetFileInfo(c *gin.Context) {
	// 从查询参数获取文件ID
	fileIDStr := c.Query("file_id")
	if fileIDStr == "" {
		s.writeError(c, http.StatusBadRequest, "Missing file_id parameter")
		return
	}

	fileID, err := strconv.ParseInt(fileIDStr, 10, 64)
	if err != nil {
		s.writeError(c, http.StatusBadRequest, "Invalid file_id parameter")
		return
	}

	req := &filepb.GetFileInfoRequest{
		FileId: fileID,
	}

	ctx := context.Background()
	resp, err := s.FileClient.GetFileInfo(ctx, req)
	if err != nil {
		utils.Error("Failed to get file info: %v", err)
		s.writeError(c, http.StatusInternalServerError, "Failed to get file info")
		return
	}

	s.writeJSON(c, http.StatusOK, "File info retrieved successfully", resp.GetFile())
}

// handleGeneratePresignedURL 处理生成预签名URL请求
func (s *GatewayServer) handleGeneratePresignedURL(c *gin.Context) {
	var req filepb.GeneratePresignedURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.writeError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := context.Background()
	resp, err := s.FileClient.GeneratePresignedURL(ctx, &req)
	if err != nil {
		utils.Error("Failed to generate presigned URL: %v", err)
		s.writeError(c, http.StatusInternalServerError, "Failed to generate presigned URL")
		return
	}

	s.writeJSON(c, http.StatusOK, "Presigned URL generated successfully", resp)
}