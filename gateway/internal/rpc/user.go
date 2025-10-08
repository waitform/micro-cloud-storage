package rpc

import (
	"context"
	"fmt"
	"time"

	userpb "github.com/waitform/micro-cloud-storage/protos/user/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials/insecure"
)

// UserServiceClient 封装用户服务客户端
type UserServiceClient struct {
	*ServiceClient
	grpcClient userpb.UserServiceClient
	conn       *grpc.ClientConn
}

// NewUserServiceClient 创建用户服务客户端
func NewUserServiceClient(serviceClient *ServiceClient) (*UserServiceClient, error) {
	// 使用etcd解析器创建连接
	target := fmt.Sprintf("etcd:///%s", "user-service")
	conn, err := grpc.Dial(target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, roundrobin.Name)),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial user-service: %w", err)
	}

	// 创建gRPC客户端
	grpcClient := userpb.NewUserServiceClient(conn)

	return &UserServiceClient{
		ServiceClient: serviceClient,
		grpcClient:    grpcClient,
		conn:          conn,
	}, nil
}

// Close 关闭gRPC连接
func (u *UserServiceClient) Close() error {
	if u.conn != nil {
		return u.conn.Close()
	}
	return nil
}

// Register 用户注册
func (u *UserServiceClient) Register(ctx context.Context, req *userpb.RegisterRequest) (*userpb.RegisterResponse, error) {
	// 设置默认超时时间
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}

	return u.grpcClient.Register(ctx, req)
}

// Login 用户登录
func (u *UserServiceClient) Login(ctx context.Context, req *userpb.LoginRequest) (*userpb.LoginResponse, error) {
	// 设置默认超时时间
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}

	return u.grpcClient.Login(ctx, req)
}

// GetUserInfo 获取用户信息
func (u *UserServiceClient) GetUserInfo(ctx context.Context, req *userpb.GetUserInfoRequest) (*userpb.GetUserInfoResponse, error) {
	// 设置默认超时时间
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}

	return u.grpcClient.GetUserInfo(ctx, req)
}
