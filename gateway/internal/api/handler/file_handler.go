package handler

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/http"
	"strconv"

	pack "github.com/waitform/micro-cloud-storage/internal/pack"
	"github.com/waitform/micro-cloud-storage/internal/rpc"
	filepb "github.com/waitform/micro-cloud-storage/protos/file/proto"
	utils "github.com/waitform/micro-cloud-storage/utils"

	"github.com/gin-gonic/gin"
)

type FileHandler struct {
	fileClient *rpc.FileServiceClient
}

func NewFileHandler(fileClient *rpc.FileServiceClient) *FileHandler {
	return &FileHandler{
		fileClient: fileClient,
	}
}

// HandleInitUpload 处理初始化上传请求
func (h *FileHandler) HandleInitUpload(c *gin.Context) {
	var req filepb.InitUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pack.WriteError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := context.Background()
	resp, err := h.fileClient.InitUpload(ctx, &req)
	if err != nil {
		utils.Error("Failed to init upload: %v", err)
		pack.WriteError(c, http.StatusInternalServerError, "Failed to init upload")
		return
	}

	pack.WriteJSON(c, http.StatusOK, "Upload initialized successfully", resp)
}

// HandleUploadPart 处理上传分片请求
func (h *FileHandler) HandleUploadPart(c *gin.Context) {
	var req filepb.UploadPartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pack.WriteError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := context.Background()
	resp, err := h.fileClient.UploadPart(ctx, &req)
	if err != nil {
		utils.Error("Failed to upload part: %v", err)
		pack.WriteError(c, http.StatusInternalServerError, "Failed to upload part")
		return
	}

	pack.WriteJSON(c, http.StatusOK, "Part uploaded successfully", resp)
}

// HandleCompleteUpload 处理完成上传请求
func (h *FileHandler) HandleCompleteUpload(c *gin.Context) {
	var req filepb.CompleteUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pack.WriteError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := context.Background()
	resp, err := h.fileClient.CompleteUpload(ctx, &req)
	if err != nil {
		utils.Error("Failed to complete upload: %v", err)
		pack.WriteError(c, http.StatusInternalServerError, "Failed to complete upload")
		return
	}

	pack.WriteJSON(c, http.StatusOK, "Upload completed successfully", resp)
}

// HandleGetFileInfo 处理获取文件信息请求
func (h *FileHandler) HandleGetFileInfo(c *gin.Context) {
	// 首先尝试从查询参数获取文件ID
	fileIDStr := c.Query("file_id")

	// 如果查询参数中没有file_id，则尝试从上下文中获取（来自分享鉴权中间件）
	if fileIDStr == "" {
		if fileIDVal, exists := c.Get("file_id"); exists {
			// 从分享鉴权中间件中获取的file_id
			fileIDStr = strconv.FormatInt(fileIDVal.(int64), 10)
		}
	}

	// 如果仍然没有file_id，则返回错误
	if fileIDStr == "" {
		pack.WriteError(c, http.StatusBadRequest, "Missing file_id parameter")
		return
	}

	fileID, err := strconv.ParseInt(fileIDStr, 10, 64)
	if err != nil {
		pack.WriteError(c, http.StatusBadRequest, "Invalid file_id parameter")
		return
	}

	req := &filepb.GetFileInfoRequest{
		FileId: fileID,
	}

	ctx := context.Background()
	resp, err := h.fileClient.GetFileInfo(ctx, req)
	if err != nil {
		utils.Error("Failed to get file info: %v", err)
		pack.WriteError(c, http.StatusInternalServerError, "Failed to get file info")
		return
	}

	pack.WriteJSON(c, http.StatusOK, "File info retrieved successfully", resp.GetFile())
}

// HandleGeneratePresignedURL 处理生成预签名URL请求
func (h *FileHandler) HandleGeneratePresignedURL(c *gin.Context) {
	var req filepb.GeneratePresignedURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pack.WriteError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := context.Background()
	resp, err := h.fileClient.GeneratePresignedURL(ctx, &req)
	if err != nil {
		utils.Error("Failed to generate presigned URL: %v", err)
		pack.WriteError(c, http.StatusInternalServerError, "Failed to generate presigned URL")
		return
	}

	pack.WriteJSON(c, http.StatusOK, "Presigned URL generated successfully", resp)
}

// HandleDirectUpload 处理直接上传文件请求（通过表单）
func (h *FileHandler) HandleDirectUpload(c *gin.Context) {
	// 从表单获取文件
	file, err := c.FormFile("file")
	if err != nil {
		pack.WriteError(c, http.StatusBadRequest, "Failed to get file from form")
		return
	}

	// 获取其他表单参数
	filename := c.PostForm("filename")
	if filename == "" {
		filename = file.Filename
	}

	// 打开文件
	src, err := file.Open()
	if err != nil {
		pack.WriteError(c, http.StatusInternalServerError, "Failed to open file")
		return
	}
	defer src.Close()

	// globalLimiter, _ := c.Get("global_limiter")
	// userLimiter, _ := c.Get("user_limiter")

	// limiterReader := utils.NewFairRateLimitedReader(src, globalLimiter.(*rate.Limiter), userLimiter.(*rate.Limiter))
	limiterReader := src
	// 读取文件内容并计算MD5
	fileData, err := io.ReadAll(limiterReader)
	if err != nil {
		pack.WriteError(c, http.StatusInternalServerError, "Failed to read file")
		return
	}

	// 计算文件MD5
	hash := md5.Sum(fileData)
	md5Str := hex.EncodeToString(hash[:])

	// 初始化上传
	initReq := &filepb.InitUploadRequest{
		FileName: filename,
		Size:     file.Size,
		Md5:      md5Str,
	}

	ctx := context.Background()
	initResp, err := h.fileClient.InitUpload(ctx, initReq)
	if err != nil {
		utils.Error("Failed to init upload: %v", err)
		pack.WriteError(c, http.StatusInternalServerError, "Failed to init upload")
		return
	}

	fileID := initResp.File.Id
	partSize := int64(5 * 1024 * 1024) // 5MB per part
	var partNumber int32 = 1

	// 分片上传文件
	for i := int64(0); i < file.Size; i += partSize {
		// 计算当前分片的大小
		currentPartSize := partSize
		if i+partSize > file.Size {
			currentPartSize = file.Size - i
		}

		// 读取分片数据
		partData := fileData[i : i+currentPartSize]

		// 计算分片MD5
		partHash := md5.Sum(partData)
		partMD5 := hex.EncodeToString(partHash[:])

		// 上传分片
		uploadPartReq := &filepb.UploadPartRequest{
			FileId:     fileID,
			PartNumber: partNumber,
			Data:       partData,
			Md5:        partMD5, // 添加MD5校验
		}

		_, err := h.fileClient.UploadPart(ctx, uploadPartReq)
		if err != nil {
			utils.Error("Failed to upload part %d: %v", partNumber, err)
			pack.WriteError(c, http.StatusInternalServerError, "Failed to upload part")
			return
		}

		partNumber++
	}

	// 完成上传
	completeReq := &filepb.CompleteUploadRequest{
		FileId: fileID,
	}

	completeResp, err := h.fileClient.CompleteUpload(ctx, completeReq)
	if err != nil {
		utils.Error("Failed to complete upload: %v", err)
		pack.WriteError(c, http.StatusInternalServerError, "Failed to complete upload")
		return
	}

	pack.WriteJSON(c, http.StatusOK, "File uploaded successfully", completeResp)
}

// HandleGetUploadProgress 处理获取上传进度请求
func (h *FileHandler) HandleGetUploadProgress(c *gin.Context) {
	fileIDStr := c.Query("file_id")
	if fileIDStr == "" {
		pack.WriteError(c, http.StatusBadRequest, "Missing file_id parameter")
		return
	}

	fileID, err := strconv.ParseInt(fileIDStr, 10, 64)
	if err != nil {
		pack.WriteError(c, http.StatusBadRequest, "Invalid file_id parameter")
		return
	}

	req := &filepb.GetUploadProgressRequest{
		FileId: fileID,
	}

	ctx := context.Background()
	resp, err := h.fileClient.GetUploadProgress(ctx, req)
	if err != nil {
		utils.Error("Failed to get upload progress: %v", err)
		pack.WriteError(c, http.StatusInternalServerError, "Failed to get upload progress")
		return
	}

	pack.WriteJSON(c, http.StatusOK, "Upload progress retrieved successfully", resp)
}

// HandleGetIncompleteParts 处理获取未完成分片请求
func (h *FileHandler) HandleGetIncompleteParts(c *gin.Context) {
	var req filepb.GetIncompletePartsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pack.WriteError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := context.Background()
	resp, err := h.fileClient.GetIncompleteParts(ctx, &req)
	if err != nil {
		utils.Error("Failed to get incomplete parts: %v", err)
		pack.WriteError(c, http.StatusInternalServerError, "Failed to get incomplete parts")
		return
	}

	pack.WriteJSON(c, http.StatusOK, "Incomplete parts retrieved successfully", resp)
}

// HandleCancelUpload 处理取消上传请求
func (h *FileHandler) HandleCancelUpload(c *gin.Context) {
	var req filepb.CancelUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pack.WriteError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := context.Background()
	_, err := h.fileClient.CancelUpload(ctx, &req)
	if err != nil {
		utils.Error("Failed to cancel upload: %v", err)
		pack.WriteError(c, http.StatusInternalServerError, "Failed to cancel upload")
		return
	}

	pack.WriteJSON(c, http.StatusOK, "Upload cancelled successfully", nil)
}

// HandleDownloadFile 处理文件下载请求
func (h *FileHandler) HandleDownloadFile(c *gin.Context) {
	// 首先尝试从查询参数获取文件ID
	fileIDStr := c.Query("file_id")

	// 如果查询参数中没有file_id，则尝试从上下文中获取（来自分享鉴权中间件）
	if fileIDStr == "" {
		if fileIDVal, exists := c.Get("file_id"); exists {
			// 从分享鉴权中间件中获取的file_id
			fileIDStr = strconv.FormatInt(fileIDVal.(int64), 10)
		}
	}

	// 如果仍然没有file_id，则返回错误
	if fileIDStr == "" {
		pack.WriteError(c, http.StatusBadRequest, "Missing file_id parameter")
		return
	}

	fileID, err := strconv.ParseInt(fileIDStr, 10, 64)
	if err != nil {
		pack.WriteError(c, http.StatusBadRequest, "Invalid file_id parameter")
		return
	}

	// 生成预签名URL用于下载
	req := &filepb.GeneratePresignedURLRequest{
		FileId:        fileID,
		ExpireSeconds: 3600, // 1小时过期时间
	}

	ctx := context.Background()
	resp, err := h.fileClient.GeneratePresignedURL(ctx, req)
	if err != nil {
		utils.Error("Failed to generate presigned URL: %v", err)
		pack.WriteError(c, http.StatusInternalServerError, "Failed to generate download link")
		return
	}

	// 重定向到预签名URL
	c.Redirect(http.StatusFound, resp.GetUrl())
}

// HandleDeleteFile 处理删除文件请求
func (h *FileHandler) HandleDeleteFile(c *gin.Context) {
	var req filepb.DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pack.WriteError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := context.Background()
	_, err := h.fileClient.DeleteFile(ctx, &req)
	if err != nil {
		utils.Error("Failed to delete file: %v", err)
		pack.WriteError(c, http.StatusInternalServerError, "Failed to delete file")
		return
	}

	pack.WriteJSON(c, http.StatusOK, "File deleted successfully", nil)
}
