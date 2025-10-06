package handler

import (
	sharepb "cloud-storage/protos/share/proto"
	"context"
	"time"

	"google.golang.org/grpc"
)

// ShareServiceClient 封装分享服务客户端
type ShareServiceClient struct {
	*ServiceClient
	grpcClient sharepb.ShareServiceClient
}

// NewShareServiceClient 创建分享服务客户端
func NewShareServiceClient(serviceClient *ServiceClient) (*ShareServiceClient, error) {
	// 获取分享服务地址
	addr, err := serviceClient.GetServiceAddr("share-service")
	if err != nil {
		return nil, err
	}

	// 创建gRPC连接
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	// 创建gRPC客户端
	grpcClient := sharepb.NewShareServiceClient(conn)

	return &ShareServiceClient{
		ServiceClient: serviceClient,
		grpcClient:    grpcClient,
	}, nil
}

// CreateShare 创建分享
func (s *ShareServiceClient) CreateShare(ctx context.Context, req *sharepb.CreateShareRequest) (*sharepb.CreateShareResponse, error) {
	// 设置默认超时时间
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}

	return s.grpcClient.CreateShare(ctx, req)
}

// GetShareInfo 获取分享信息
func (s *ShareServiceClient) GetShareInfo(ctx context.Context, req *sharepb.GetShareInfoRequest) (*sharepb.GetShareInfoResponse, error) {
	// 设置默认超时时间
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}

	return s.grpcClient.GetShareInfo(ctx, req)
}

// ValidateAccess 验证访问权限
func (s *ShareServiceClient) ValidateAccess(ctx context.Context, req *sharepb.ValidateAccessRequest) (*sharepb.ValidateAccessResponse, error) {
	// 设置默认超时时间
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}

	return s.grpcClient.ValidateAccess(ctx, req)
}
