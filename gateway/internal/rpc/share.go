package rpc

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc/keepalive"

	sharepb "github.com/waitform/micro-cloud-storage/protos/share/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials/insecure"
)

// ShareServiceClient 封装分享服务客户端
type ShareServiceClient struct {
	*ServiceClient
	grpcClient sharepb.ShareServiceClient
	conn       *grpc.ClientConn
}

// NewShareServiceClient 创建分享服务客户端
func NewShareServiceClient(serviceClient *ServiceClient) (*ShareServiceClient, error) {
	// 使用etcd解析器创建连接
	target := fmt.Sprintf("etcd:///%s", "share-service")
	conn, err := grpc.Dial(target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, roundrobin.Name)),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                20 * time.Second, // 每20秒发送一次ping
			Timeout:             3 * time.Second,  // ping超时时间
			PermitWithoutStream: true,             // 允许在没有活跃流时发送ping
		}),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial share-service: %w", err)
	}

	// 创建gRPC客户端
	grpcClient := sharepb.NewShareServiceClient(conn)

	return &ShareServiceClient{
		ServiceClient: serviceClient,
		grpcClient:    grpcClient,
		conn:          conn,
	}, nil
}

// Close 关闭gRPC连接
func (s *ShareServiceClient) Close() error {
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
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
