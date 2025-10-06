package service

import (
	"crypto/md5"
	"fmt"
	"time"

	"cloud-storage-user-service/config"
	"cloud-storage-user-service/internal/model"
	"cloud-storage-user-service/internal/types"

	"github.com/golang-jwt/jwt/v5"
)

type UserService struct {
	userDAO model.UserDAO
	cfg     *config.Config
}

func NewUserService(dao model.UserDAO, cfg *config.Config) *UserService {
	return &UserService{
		userDAO: dao,
		cfg:     cfg,
	}
}

func (s *UserService) Register(req *types.RegisterRequest) (*types.RegisterResponse, error) {
	// 检查用户是否已存在
	existingUser, err := s.userDAO.GetByUsername(req.Username)
	if err != nil {
		return nil, fmt.Errorf("查询用户时出错: %v", err)
	}
	
	if existingUser != nil {
		return &types.RegisterResponse{
			Success: false,
			Message: "用户名已存在",
		}, nil
	}

	// 创建新用户
	hash := fmt.Sprintf("%x", md5.Sum([]byte(req.Password)))
	user := &model.User{
		Username: req.Username,
		Password: hash,
		Email:    req.Email,
	}

	if err := s.userDAO.CreateUser(user); err != nil {
		return &types.RegisterResponse{
			Success: false,
			Message: "注册失败: " + err.Error(),
		}, nil
	}

	return &types.RegisterResponse{
		Success: true,
		Message: "注册成功",
		User: &types.UserInfo{
			ID:         user.ID,
			Username:   user.Username,
			Avatar:     user.Avatar,
			UsedSpace:  user.UsedSpace,
			TotalSpace: user.TotalSpace,
		},
	}, nil
}

func (s *UserService) Login(req *types.LoginRequest) (*types.LoginResponse, error) {
	// 获取用户信息
	user, err := s.userDAO.GetByUsername(req.Username)
	if err != nil {
		return &types.LoginResponse{
			Success: false,
			Message: "数据库查询错误: " + err.Error(),
		}, nil
	}

	if user == nil {
		return &types.LoginResponse{
			Success: false,
			Message: "用户名或密码错误",
		}, nil
	}

	// 验证密码
	hash := fmt.Sprintf("%x", md5.Sum([]byte(req.Password)))
	if user.Password != hash {
		return &types.LoginResponse{
			Success: false,
			Message: "用户名或密码错误",
		}, nil
	}

	// 生成JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // 24小时过期
	})

	tokenString, err := token.SignedString([]byte(s.cfg.JWT.Secret))
	if err != nil {
		return &types.LoginResponse{
			Success: false,
			Message: "生成token失败: " + err.Error(),
		}, nil
	}

	return &types.LoginResponse{
		Success: true,
		Message: "登录成功",
		UserID:  user.ID,
		Token:   tokenString,
	}, nil
}

func (s *UserService) GetUserInfo(req *types.GetUserInfoRequest) (*types.GetUserInfoResponse, error) {
	// 获取用户信息
	user, err := s.userDAO.GetByID(req.ID)
	if err != nil {
		return nil, fmt.Errorf("数据库查询错误: %v", err)
	}

	if user == nil {
		return nil, fmt.Errorf("用户不存在")
	}

	return &types.GetUserInfoResponse{
		User: &types.UserInfo{
			ID:         user.ID,
			Username:   user.Username,
			Avatar:     user.Avatar,
			UsedSpace:  user.UsedSpace,
			TotalSpace: user.TotalSpace,
		},
	}, nil
}

func (s *UserService) UpdateUser(req *types.UpdateUserRequest) error {
	// 获取用户信息
	user, err := s.userDAO.GetByID(req.ID)
	if err != nil {
		return fmt.Errorf("数据库查询错误: %v", err)
	}

	if user == nil {
		return fmt.Errorf("用户不存在")
	}

	// 更新用户头像
	user.Avatar = req.Avatar

	// 保存更新
	if err := s.userDAO.UpdateUser(user); err != nil {
		return fmt.Errorf("更新用户信息失败: %v", err)
	}

	return nil
}

func (s *UserService) UpdateCapacity(req *types.UpdateCapacityRequest) error {
	// 获取用户信息
	user, err := s.userDAO.GetByID(req.UserID)
	if err != nil {
		return fmt.Errorf("数据库查询错误: %v", err)
	}

	if user == nil {
		return fmt.Errorf("用户不存在")
	}

	// 更新用户容量
	if err := s.userDAO.UpdateCapacity(req.UserID, req.NewTotal); err != nil {
		return fmt.Errorf("更新用户容量失败: %v", err)
	}

	return nil
}

func (s *UserService) CheckCapacity(req *types.CheckCapacityRequest) (*types.CheckCapacityResponse, error) {
	// 获取用户信息
	user, err := s.userDAO.GetByID(req.UserID)
	if err != nil {
		return &types.CheckCapacityResponse{
			IsEnough: false,
		}, fmt.Errorf("数据库查询错误: %v", err)
	}

	if user == nil {
		return &types.CheckCapacityResponse{
			IsEnough: false,
		}, fmt.Errorf("用户不存在")
	}

	// 检查容量是否足够
	usedSpace := user.UsedSpace
	totalSpace := user.TotalSpace
	availableSpace := totalSpace - usedSpace

	isEnough := availableSpace >= req.FileSize

	return &types.CheckCapacityResponse{
		IsEnough: isEnough,
	}, nil
}