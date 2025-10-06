package api

import (
	"cloud-storage-file-service/internal/service"
	"cloud-storage-file-service/proto"
	"context"

	"google.golang.org/protobuf/types/known/emptypb"
)

type FileServiceServer struct {
	proto.UnimplementedFileServiceServer
	storage *service.StorageService
}

func NewFileServiceServer(storage *service.StorageService) proto.FileServiceServer {
	return &FileServiceServer{
		storage: storage,
	}
}
func (s *FileServiceServer) InitUpload(ctx context.Context, req *proto.InitUploadRequest) (*proto.InitUploadResponse, error) {
	file, err := s.storage.InitUpload(ctx, req.FileName, req.Size, req.Md5, req.UserID)
	if err != nil {
		return nil, err
	}

	return &proto.InitUploadResponse{
		File: &proto.FileInfo{
			Id:     file.ID,
			Name:   file.FileName,
			UserID: file.UserID,
			Size:   file.Size,
			Md5:    file.Md5,
			Status: int32(file.Status),
		},
	}, nil
}
func (s *FileServiceServer) UploadPart(ctx context.Context, req *proto.UploadPartRequest) (*emptypb.Empty, error) {
	err := s.storage.UploadPart(ctx, req.FileId, int(req.PartNumber), req.Data, req.Md5)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
func (s *FileServiceServer) CompleteUpload(ctx context.Context, req *proto.CompleteUploadRequest) (*emptypb.Empty, error) {
	err := s.storage.UploadComplete(ctx, req.FileId)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
func (s *FileServiceServer) DownloadPart(req *proto.DownloadRequest, stream proto.FileService_DownloadPartServer) error {
	chunkData, md5, err := s.storage.DownloadChunk(stream.Context(), req.FileId, int(req.PartNumber), 0)
	if err != nil {
		return err
	}

	resp := &proto.DownloadResponse{
		Data: chunkData,
		Md5:  md5,
	}

	return stream.Send(resp)
}
func (s *FileServiceServer) GetFileInfo(ctx context.Context, req *proto.GetFileInfoRequest) (*proto.GetFileInfoResponse, error) {
	file, err := s.storage.GetFileInfo(ctx, req.FileId)
	if err != nil {
		return nil, err
	}

	return &proto.GetFileInfoResponse{
		File: &proto.FileInfo{
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
func (s *FileServiceServer) GeneratePresignedURL(ctx context.Context, req *proto.GeneratePresignedURLRequest) (*proto.GeneratePresignedURLResponse, error) {
	url, expireAt, err := s.storage.GeneratePresignedURL(ctx, req.FileId, req.ExpireSeconds)
	if err != nil {
		return nil, err
	}
	//返回url和过期时间
	return &proto.GeneratePresignedURLResponse{
		Url:      url,
		ExpireAt: expireAt,
	}, nil
}

// 删除文件
func (s *FileServiceServer) DeleteFile(ctx context.Context, req *proto.DeleteRequest) (*emptypb.Empty, error) {
	err := s.storage.DeleteFile(ctx, req.FileId)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
