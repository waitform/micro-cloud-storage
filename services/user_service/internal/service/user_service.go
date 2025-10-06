package service

import (
	"context"
	"errors"

	"cloud-storage/services/user_service/internal/model"
	"cloud-storage/services/user_service/internal/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService 用户服务结构
type UserService struct {
	userDAO model.UserDAO
}

// NewUserService 创建新的用户服务实例
func NewUserService(userDAO model.UserDAO) *UserService {
	return &UserService{
		userDAO: userDAO,
	}
}

// Register 用户注册
func (s *UserService) Register(ctx context.Context, username, email, password string) (*model.User, error) {
	// 检查密码强度
	if len(password) < 6 {
		return nil, errors.New("密码长度不能少于6位")
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		utils.Error("密码加密失败: %v", err)
		return nil, errors.New("服务器内部错误")
	}

	// 创建用户对象
	user := &model.User{
		Username: username,
		Email:    email,
		Password: string(hashedPassword),
	}

	// 保存到数据库
	if err := s.userDAO.CreateUser(user); err != nil {
		// 检查是否是唯一性约束违反错误
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, errors.New("用户已存在")
		}
		utils.Error("创建用户失败: %v", err)
		return nil, errors.New("服务器内部错误")
	}

	// 清除密码字段，避免返回给客户端
	user.Password = ""

	return user, nil
}

// Login 用户登录
func (s *UserService) Login(ctx context.Context, email, password string) (string, *model.User, error) {
	// 根据邮箱查找用户
	user, err := s.userDAO.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, errors.New("用户不存在")
		}
		utils.Error("获取用户失败: %v", err)
		return "", nil, errors.New("服务器内部错误")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, errors.New("密码错误")
	}

	// 生成JWT token
	token, err := utils.GenerateToken(user.ID, user.Username)
	if err != nil {
		utils.Error("生成token失败: %v", err)
		return "", nil, errors.New("服务器内部错误")
	}

	// 清除密码字段，避免返回给客户端
	user.Password = ""

	return token, user, nil
}

// GetUserByID 根据ID获取用户信息
func (s *UserService) GetUserByID(ctx context.Context, userID uint) (*model.User, error) {
	user, err := s.userDAO.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		utils.Error("获取用户失败: %v", err)
		return nil, errors.New("服务器内部错误")
	}

	// 清除密码字段，避免返回给客户端
	user.Password = ""

	return user, nil
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(ctx context.Context, userID uint, username, email string) (*model.User, error) {
	// 获取用户信息
	user, err := s.userDAO.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		utils.Error("获取用户失败: %v", err)
		return nil, errors.New("服务器内部错误")
	}

	// 更新用户信息
	user.Username = username
	user.Email = email

	// 保存到数据库
	if err := s.userDAO.UpdateUser(user); err != nil {
		// 检查是否是唯一性约束违反错误
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, errors.New("邮箱已被使用")
		}
		utils.Error("更新用户失败: %v", err)
		return nil, errors.New("服务器内部错误")
	}

	// 清除密码字段，避免返回给客户端
	user.Password = ""

	return user, nil
}