package api

import (
	"context"
	"time"

	"cloud-storage-user-service/internal/service"
	"cloud-storage-user-service/internal/types"
	pb "cloud-storage-user-service/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UserServiceServer 实现proto中定义的UserServiceServer接口
type UserServiceServer struct {
	pb.UnimplementedUserServiceServer
	userService *service.UserService
}

// NewUserServiceServer 创建UserServiceServer实例
func NewUserServiceServer(userService *service.UserService) *UserServiceServer {
	return &UserServiceServer{
		userService: userService,
	}
}

// Register 用户注册
func (s *UserServiceServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	// 转换请求参数
	registerReq := &types.RegisterRequest{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
	}

	// 调用服务层
	resp, err := s.userService.Register(registerReq)
	if err != nil {
		return nil, err
	}

	// 转换响应参数
	var user *pb.User
	if resp.User != nil {
		user = &pb.User{
			Id:         resp.User.ID,
			Username:   resp.User.Username,
			Avatar:     resp.User.Avatar,
			TotalSpace: resp.User.TotalSpace,
			UsedSpace:  resp.User.UsedSpace,
			CreatedAt:  time.Now().Format(time.RFC3339),
			UpdatedAt:  time.Now().Format(time.RFC3339),
		}
	}

	return &pb.RegisterResponse{
		Success: resp.Success,
		Message: resp.Message,
		User:    user,
	}, nil
}

// Login 用户登录
func (s *UserServiceServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	// 转换请求参数
	loginReq := &types.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	}

	// 调用服务层
	resp, err := s.userService.Login(loginReq)
	if err != nil {
		return nil, err
	}

	return &pb.LoginResponse{
		Success: resp.Success,
		Message: resp.Message,
		UserId:  resp.UserID,
		Token:   resp.Token,
	}, nil
}

// GetUserInfo 获取用户信息
func (s *UserServiceServer) GetUserInfo(ctx context.Context, req *pb.GetUserInfoRequest) (*pb.GetUserInfoResponse, error) {
	// 转换请求参数
	getUserInfoReq := &types.GetUserInfoRequest{
		ID: req.UserId,
	}

	// 调用服务层
	resp, err := s.userService.GetUserInfo(getUserInfoReq)
	if err != nil {
		return nil, err
	}

	// 转换响应参数
	var user *pb.User
	if resp.User != nil {
		user = &pb.User{
			Id:         resp.User.ID,
			Username:   resp.User.Username,
			Avatar:     resp.User.Avatar,
			TotalSpace: resp.User.TotalSpace,
			UsedSpace:  resp.User.UsedSpace,
			CreatedAt:  time.Now().Format(time.RFC3339),
			UpdatedAt:  time.Now().Format(time.RFC3339),
		}
	}

	return &pb.GetUserInfoResponse{
		User: user,
	}, nil
}

// UpdateUserInfo 更新用户信息
func (s *UserServiceServer) UpdateUserInfo(ctx context.Context, req *pb.UpdateUserInfoRequest) (*pb.UpdateUserInfoResponse, error) {
	// 转换请求参数
	updateUserReq := &types.UpdateUserRequest{
		ID:     req.UserId,
		Avatar: req.Avatar,
	}

	// 调用服务层
	err := s.userService.UpdateUser(updateUserReq)
	if err != nil {
		return &pb.UpdateUserInfoResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.UpdateUserInfoResponse{
		Success: true,
		Message: "更新成功",
	}, nil
}

// UpdateUsage 更新用户存储使用量
func (s *UserServiceServer) UpdateUsage(ctx context.Context, req *pb.UpdateUsageRequest) (*pb.UpdateUsageResponse, error) {
	// 这个方法在服务层没有实现，暂时返回未实现错误
	return nil, status.Errorf(codes.Unimplemented, "method UpdateUsage not implemented")
}

// UpdateCapacity 更新用户总容量
func (s *UserServiceServer) UpdateCapacity(ctx context.Context, req *pb.UpdateCapacityRequest) (*pb.UpdateCapacityResponse, error) {
	// 转换请求参数
	updateCapacityReq := &types.UpdateCapacityRequest{
		UserID:   req.UserId,
		NewTotal: req.NewTotalSpace,
	}

	// 调用服务层
	err := s.userService.UpdateCapacity(updateCapacityReq)
	if err != nil {
		return &pb.UpdateCapacityResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.UpdateCapacityResponse{
		Success:    true,
		Message:    "更新成功",
		TotalSpace: req.NewTotalSpace,
	}, nil
}

// CheckCapacity 检查用户容量是否足够
func (s *UserServiceServer) CheckCapacity(ctx context.Context, req *pb.CheckCapacityRequest) (*pb.CheckCapacityResponse, error) {
	// 转换请求参数
	checkCapacityReq := &types.CheckCapacityRequest{
		UserID:   req.UserId,
		FileSize: req.FileSize,
	}

	// 调用服务层
	resp, err := s.userService.CheckCapacity(checkCapacityReq)
	if err != nil {
		return nil, err
	}

	return &pb.CheckCapacityResponse{
		Enough: resp.IsEnough,
	}, nil
}
