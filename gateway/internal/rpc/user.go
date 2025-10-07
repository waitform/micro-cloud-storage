package rpc

import (
	userpb "cloud-storage/protos/user/proto"
	"context"
	"time"

	"google.golang.org/grpc"
)

// UserServiceClient 封装用户服务客户端
type UserServiceClient struct {
	*ServiceClient
	grpcClient userpb.UserServiceClient
}

// NewUserServiceClient 创建用户服务客户端
func NewUserServiceClient(serviceClient *ServiceClient) (*UserServiceClient, error) {
	// 获取用户服务地址
	addr, err := serviceClient.GetServiceAddr("user-service")
	if err != nil {
		return nil, err
	}

	// 创建gRPC连接
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	// 创建gRPC客户端
	grpcClient := userpb.NewUserServiceClient(conn)

	return &UserServiceClient{
		ServiceClient: serviceClient,
		grpcClient:    grpcClient,
	}, nil
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

// UpdateUserInfo 更新用户信息
func (u *UserServiceClient) UpdateUserInfo(ctx context.Context, req *userpb.UpdateUserInfoRequest) (*userpb.UpdateUserInfoResponse, error) {
	// 设置默认超时时间
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}

	return u.grpcClient.UpdateUserInfo(ctx, req)
}

// CheckCapacity 检查用户容量
func (u *UserServiceClient) CheckCapacity(ctx context.Context, req *userpb.CheckCapacityRequest) (*userpb.CheckCapacityResponse, error) {
	// 设置默认超时时间
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}

	return u.grpcClient.CheckCapacity(ctx, req)
}
