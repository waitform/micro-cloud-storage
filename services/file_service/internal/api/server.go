package api

import (
	"context"
	"io"

	"cloud-storage-file-service/internal/service"
	filepb "cloud-storage-file-service/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

type FileServiceServer struct {
	filepb.UnimplementedFileServiceServer
	storage *service.StorageService
}

func NewFileServiceServer(storage *service.StorageService) filepb.FileServiceServer {
	return &FileServiceServer{
		storage: storage,
	}
}

func (s *FileServiceServer) InitUpload(ctx context.Context, req *filepb.InitUploadRequest) (*filepb.InitUploadResponse, error) {
	file, err := s.storage.InitUpload(ctx, req.FileName, req.Size, req.Md5, req.UserID)
	if err != nil {
		return nil, err
	}

	return &filepb.InitUploadResponse{
		File: &filepb.FileInfo{
			Id:     file.ID,
			Name:   file.FileName,
			UserID: file.UserID,
			Size:   file.Size,
			Md5:    file.Md5,
			Status: int32(file.Status),
		},
	}, nil
}

func (s *FileServiceServer) UploadPart(stream filepb.FileService_UploadPartServer) error {
	var fileID int64
	var partNumber int32
	var md5Str string
	var receivedData []byte
	
	// 接收流中的数据
	firstMessage := true
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		
		// 处理第一个消息（包含元数据）
		if firstMessage {
			fileID = req.FileId
			partNumber = req.PartNumber
			md5Str = req.Md5
			firstMessage = false
			
			// 如果这是仅包含元数据的消息（没有实际数据），则跳过
			if len(req.Data) == 0 {
				continue
			}
			// 如果消息中包含数据，则处理数据
		}
		
		// 累积数据
		receivedData = append(receivedData, req.Data...)
	}
	
	// 调用存储服务上传分片
	err := s.storage.UploadPart(stream.Context(), fileID, int(partNumber), receivedData, md5Str)
	if err != nil {
		return err
	}
	
	// 发送响应
	return stream.SendAndClose(&emptypb.Empty{})
}

func (s *FileServiceServer) CompleteUpload(ctx context.Context, req *filepb.CompleteUploadRequest) (*filepb.CompleteUploadResponse, error) {
	err := s.storage.UploadComplete(ctx, req.FileId)
	if err != nil {
		return nil, err
	}
	
	// 获取文件信息用于返回
	file, err := s.storage.GetFileInfo(ctx, req.FileId)
	if err != nil {
		return nil, err
	}
	
	return &filepb.CompleteUploadResponse{
		File: &filepb.FileInfo{
			Id:     file.ID,
			Name:   file.FileName,
			Size:   file.Size,
			UserID: file.UserID,
			Md5:    file.Md5,
			Status: int32(file.Status),
		},
	}, nil
}

func (s *FileServiceServer) DownloadPart(req *filepb.DownloadRequest, stream filepb.FileService_DownloadPartServer) error {
	chunkData, md5, err := s.storage.DownloadChunk(stream.Context(), req.FileId, int(req.PartNumber), 0)
	if err != nil {
		return err
	}

	resp := &filepb.DownloadResponse{
		Data: chunkData,
		Md5:  md5,
	}

	return stream.Send(resp)
}

func (s *FileServiceServer) GetFileInfo(ctx context.Context, req *filepb.GetFileInfoRequest) (*filepb.GetFileInfoResponse, error) {
	file, err := s.storage.GetFileInfo(ctx, req.FileId)
	if err != nil {
		return nil, err
	}

	return &filepb.GetFileInfoResponse{
		File: &filepb.FileInfo{
			Id:     file.ID,
			Name:   file.FileName,
			Size:   file.Size,
			UserID: file.UserID,
			Md5:    file.Md5,
			Status: int32(file.Status),
		},
	}, nil
}

// 生成预签名URL
func (s *FileServiceServer) GeneratePresignedURL(ctx context.Context, req *filepb.GeneratePresignedURLRequest) (*filepb.GeneratePresignedURLResponse, error) {
	url, expireAt, err := s.storage.GeneratePresignedURL(ctx, req.FileId, req.ExpireSeconds)
	if err != nil {
		return nil, err
	}
	//返回url和过期时间
	return &filepb.GeneratePresignedURLResponse{
		Url:      url,
		ExpireAt: expireAt,
	}, nil
}

// 删除文件
func (s *FileServiceServer) DeleteFile(ctx context.Context, req *filepb.DeleteRequest) (*emptypb.Empty, error) {
	err := s.storage.DeleteFile(ctx, req.FileId)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// 获取上传进度
func (s *FileServiceServer) GetUploadProgress(ctx context.Context, req *filepb.GetUploadProgressRequest) (*filepb.GetUploadProgressResponse, error) {
	uploadedSize, totalSize, err := s.storage.GetUploadProgress(req.FileId)
	if err != nil {
		return nil, err
	}

	// 计算进度百分比
	var progress float64
	if totalSize > 0 {
		progress = float64(uploadedSize) / float64(totalSize) * 100
	}

	return &filepb.GetUploadProgressResponse{
		UploadedSize: uploadedSize,
		TotalSize:    totalSize,
		Progress:     progress,
	}, nil
}

// 获取未完成的分片
func (s *FileServiceServer) GetIncompleteParts(ctx context.Context, req *filepb.GetIncompletePartsRequest) (*filepb.GetIncompletePartsResponse, error) {
	missingParts, err := s.storage.GetIncompleteParts(ctx, req.FileId, int(req.TotalParts))
	if err != nil {
		return nil, err
	}

	// 转换为int32切片
	missingParts32 := make([]int32, len(missingParts))
	for i, part := range missingParts {
		missingParts32[i] = int32(part)
	}

	return &filepb.GetIncompletePartsResponse{
		MissingParts: missingParts32,
	}, nil
}

// 取消上传
func (s *FileServiceServer) CancelUpload(ctx context.Context, req *filepb.CancelUploadRequest) (*emptypb.Empty, error) {
	err := s.storage.DeleteFile(ctx, req.FileId)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}