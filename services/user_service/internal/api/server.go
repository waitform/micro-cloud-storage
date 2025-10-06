package api

import (
	"context"

	"cloud-storage/services/user_service/internal/service"
	"cloud-storage/services/user_service/internal/utils"
	pb "cloud-storage/services/user_service/proto"

)

// UserServiceServer 用户服务gRPC服务端
type UserServiceServer struct {
	pb.UnimplementedUserServiceServer
	userService *service.UserService
}

// NewUserServiceServer 创建新的用户服务gRPC服务端实例
func NewUserServiceServer(userService *service.UserService) pb.UserServiceServer {
	return &UserServiceServer{
		userService: userService,
	}
}

// Register 用户注册
func (s *UserServiceServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	utils.Info("开始注册用户: %s", req.Username)
	user, err := s.userService.Register(ctx, req.Username, req.Email, req.Password)
	if err != nil {
		utils.Error("注册用户失败: %s, 错误: %v", req.Username, err)
		return &pb.RegisterResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	utils.Info("用户注册成功: %s", req.Username)
	return &pb.RegisterResponse{
		Success: true,
		Message: "注册成功",
		User: &pb.User{
			Id:        uint64(user.ID),
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.String(),
			UpdatedAt: user.UpdatedAt.String(),
		},
	}, nil
}

// Login 用户登录
func (s *UserServiceServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	utils.Info("用户尝试登录: %s", req.Email)
	token, user, err := s.userService.Login(ctx, req.Email, req.Password)
	if err != nil {
		utils.Error("用户登录失败: %s, 错误: %v", req.Email, err)
		return &pb.LoginResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	utils.Info("用户登录成功: %s", req.Email)
	return &pb.LoginResponse{
		Success: true,
		Message: "登录成功",
		Token:   token,
		User: &pb.User{
			Id:        uint64(user.ID),
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.String(),
			UpdatedAt: user.UpdatedAt.String(),
		},
	}, nil
}

// GetUser 获取用户信息
func (s *UserServiceServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	// 从请求上下文中获取用户ID（由网关解析JWT后传递）
	userID := req.UserId
	
	utils.Info("获取用户信息，用户ID: %d", userID)
	user, err := s.userService.GetUserByID(ctx, uint(userID))
	if err != nil {
		utils.Error("获取用户信息失败，用户ID: %d, 错误: %v", userID, err)
		return &pb.GetUserResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	utils.Info("获取用户信息成功，用户ID: %d", userID)
	return &pb.GetUserResponse{
		Success: true,
		Message: "获取用户信息成功",
		User: &pb.User{
			Id:        uint64(user.ID),
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.String(),
			UpdatedAt: user.UpdatedAt.String(),
		},
	}, nil
}

// UpdateUser 更新用户信息
func (s *UserServiceServer) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	// 注意：UpdateUser操作需要在网关层确保只能更新自己的信息
	// 或者在请求中添加用户ID字段以指定要更新的用户
	utils.Info("更新用户信息")
	
	// 这里我们假设网关会在上下文中传递用户ID
	// 由于.proto文件中没有定义userId字段，我们需要修改.proto文件或者在网关层处理
	// 为了保持.proto文件不变，我们记录此问题，实际使用时需要在网关层处理
	
	utils.Warn("UpdateUser方法缺少用户ID参数，需要在网关层处理用户身份验证")
	
	// 暂时返回错误，直到确定实现方式
	return &pb.UpdateUserResponse{
		Success: false,
		Message: "更新用户信息功能暂未实现",
	}, nil
}
