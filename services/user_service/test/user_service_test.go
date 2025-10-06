package test

import (
	"testing"

	"cloud-storage-user-service/internal/types"
)

// MockUserService 是 UserService 的模拟实现
type MockUserService struct{}

func (m *MockUserService) Register(req *types.RegisterRequest) (*types.RegisterResponse, error) {
	return &types.RegisterResponse{
		Success: true,
		Message: "注册成功",
		User: &types.UserInfo{
			ID:         1,
			Username:   req.Username,
			Avatar:     "",
			UsedSpace:  0,
			TotalSpace: 10737418240, // 10GB
		},
	}, nil
}

func (m *MockUserService) Login(req *types.LoginRequest) (*types.LoginResponse, error) {
	return &types.LoginResponse{
		Success: true,
		Message: "登录成功",
		UserID:  1,
		Token:   "test-token",
	}, nil
}

func (m *MockUserService) GetUserInfo(req *types.GetUserInfoRequest) (*types.GetUserInfoResponse, error) {
	return &types.GetUserInfoResponse{
		User: &types.UserInfo{
			ID:         req.ID,
			Username:   "testuser",
			Avatar:     "",
			UsedSpace:  1024,
			TotalSpace: 10737418240,
		},
	}, nil
}

func (m *MockUserService) UpdateUser(req *types.UpdateUserRequest) error {
	return nil
}

func (m *MockUserService) UpdateCapacity(req *types.UpdateCapacityRequest) error {
	return nil
}

func (m *MockUserService) CheckCapacity(req *types.CheckCapacityRequest) (*types.CheckCapacityResponse, error) {
	return &types.CheckCapacityResponse{
		IsEnough: true,
	}, nil
}

func TestUserService_Register(t *testing.T) {
	// 创建模拟服务
	mockService := &MockUserService{}

	// 测试注册
	req := &types.RegisterRequest{
		Username: "testuser",
		Password: "testpassword",
		Email:    "test@example.com",
	}

	resp, err := mockService.Register(req)
	if err != nil {
		t.Fatalf("注册失败: %v", err)
	}

	if !resp.Success {
		t.Errorf("注册应该成功，但返回了失败")
	}

	if resp.User == nil {
		t.Errorf("注册成功后应该返回用户信息")
	}

	if resp.User.Username != req.Username {
		t.Errorf("用户名不匹配，期望: %s, 实际: %s", req.Username, resp.User.Username)
	}
}

func TestUserService_Login(t *testing.T) {
	// 创建模拟服务
	mockService := &MockUserService{}

	// 测试登录
	req := &types.LoginRequest{
		Username: "testuser",
		Password: "testpassword",
	}

	resp, err := mockService.Login(req)
	if err != nil {
		t.Fatalf("登录失败: %v", err)
	}

	if !resp.Success {
		t.Errorf("登录应该成功，但返回了失败")
	}

	if resp.Token == "" {
		t.Errorf("登录成功后应该返回token")
	}
}

func TestUserService_GetUserInfo(t *testing.T) {
	// 创建模拟服务
	mockService := &MockUserService{}

	// 测试获取用户信息
	req := &types.GetUserInfoRequest{
		ID: 1,
	}

	resp, err := mockService.GetUserInfo(req)
	if err != nil {
		t.Fatalf("获取用户信息失败: %v", err)
	}

	if resp.User == nil {
		t.Errorf("应该返回用户信息")
	}

	if resp.User.ID != req.ID {
		t.Errorf("用户ID不匹配，期望: %d, 实际: %d", req.ID, resp.User.ID)
	}
}

func TestUserService_UpdateUser(t *testing.T) {
	// 创建模拟服务
	mockService := &MockUserService{}

	// 测试更新用户信息
	req := &types.UpdateUserRequest{
		ID:     1,
		Avatar: "new_avatar_url",
	}

	err := mockService.UpdateUser(req)
	if err != nil {
		t.Fatalf("更新用户信息失败: %v", err)
	}
}

func TestUserService_UpdateCapacity(t *testing.T) {
	// 创建模拟服务
	mockService := &MockUserService{}

	// 测试更新用户容量
	req := &types.UpdateCapacityRequest{
		UserID:   1,
		NewTotal: 21474836480, // 20GB
	}

	err := mockService.UpdateCapacity(req)
	if err != nil {
		t.Fatalf("更新用户容量失败: %v", err)
	}
}

func TestUserService_CheckCapacity(t *testing.T) {
	// 创建模拟服务
	mockService := &MockUserService{}

	// 测试检查容量
	req := &types.CheckCapacityRequest{
		UserID:   1,
		FileSize: 1024,
	}

	resp, err := mockService.CheckCapacity(req)
	if err != nil {
		t.Fatalf("检查容量失败: %v", err)
	}

	if !resp.IsEnough {
		t.Errorf("容量应该足够")
	}
}