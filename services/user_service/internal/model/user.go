package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"not null;uniqueIndex"`
	Email     string    `json:"email" gorm:"not null;uniqueIndex"`
	Password  string    `json:"password" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// -------------------- DAO 接口 --------------------

// UserDAO 定义用户数据访问接口
type UserDAO interface {
	CreateUser(user *User) error
	GetUserByEmail(email string) (*User, error)
	GetUserByID(id uint) (*User, error)
	UpdateUser(user *User) error
}

// -------------------- DAO 实现 --------------------

type userDAOImpl struct {
	db *gorm.DB
}

// NewUserDAO 创建 DAO 实例
func NewUserDAO(db *gorm.DB) *userDAOImpl {
	// 自动迁移表
	db.AutoMigrate(&User{})
	return &userDAOImpl{db: db}
}

// CreateUser 创建用户
func (dao *userDAOImpl) CreateUser(user *User) error {
	return dao.db.Create(user).Error
}

// GetUserByEmail 根据邮箱获取用户
func (dao *userDAOImpl) GetUserByEmail(email string) (*User, error) {
	var user User
	err := dao.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID 根据ID获取用户
func (dao *userDAOImpl) GetUserByID(id uint) (*User, error) {
	var user User
	err := dao.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser 更新用户信息
func (dao *userDAOImpl) UpdateUser(user *User) error {
	return dao.db.Save(user).Error
}