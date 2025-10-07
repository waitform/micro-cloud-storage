package rpc

import (
	filepb "cloud-storage/protos/file/proto"
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// FileServiceClient 封装文件服务客户端
type FileServiceClient struct {
	*ServiceClient
	grpcClient filepb.FileServiceClient
}

// NewFileServiceClient 创建文件服务客户端
func NewFileServiceClient(serviceClient *ServiceClient) (*FileServiceClient, error) {
	// 获取文件服务地址
	addr, err := serviceClient.GetServiceAddr("file-service")
	if err != nil {
		return nil, err
	}

	// 创建gRPC连接
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	// 创建gRPC客户端
	grpcClient := filepb.NewFileServiceClient(conn)

	return &FileServiceClient{
		ServiceClient: serviceClient,
		grpcClient:    grpcClient,
	}, nil
}

// InitUpload 初始化上传
func (f *FileServiceClient) InitUpload(ctx context.Context, req *filepb.InitUploadRequest) (*filepb.InitUploadResponse, error) {
	// 设置默认超时时间
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}

	return f.grpcClient.InitUpload(ctx, req)
}

// UploadPart 上传分片
func (f *FileServiceClient) UploadPart(ctx context.Context, req *filepb.UploadPartRequest) (*emptypb.Empty, error) {
	// 设置默认超时时间
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}

	return f.grpcClient.UploadPart(ctx, req)
}

// CompleteUpload 完成上传
func (f *FileServiceClient) CompleteUpload(ctx context.Context, req *filepb.CompleteUploadRequest) (*emptypb.Empty, error) {
	// 设置默认超时时间
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}

	return f.grpcClient.CompleteUpload(ctx, req)
}

// GetFileInfo 获取文件信息
func (f *FileServiceClient) GetFileInfo(ctx context.Context, req *filepb.GetFileInfoRequest) (*filepb.GetFileInfoResponse, error) {
	// 设置默认超时时间
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}

	return f.grpcClient.GetFileInfo(ctx, req)
}

// GeneratePresignedURL 生成预签名URL
func (f *FileServiceClient) GeneratePresignedURL(ctx context.Context, req *filepb.GeneratePresignedURLRequest) (*filepb.GeneratePresignedURLResponse, error) {
	// 设置默认超时时间
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}

	return f.grpcClient.GeneratePresignedURL(ctx, req)
}
