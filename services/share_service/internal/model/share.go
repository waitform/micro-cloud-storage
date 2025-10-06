package model

import (
	"errors"
	"time"

	"github.com/jinzhu/gorm"
)

type Share struct {
	gorm.Model
	FileID   int64     `gorm:"not null;index"` // 被分享的文件
	OwnerID  int64     `gorm:"not null;index"` // 文件所有者
	ShareID  string    `gorm:"not null;index"`
	Password string    `gorm:"size:255"`
	ExpireAt time.Time `gorm:"not null"`
}
type ShareDAO struct {
	db *gorm.DB
}

func NewShareDAO(db *gorm.DB) *ShareDAO {
	return &ShareDAO{db: db}
}

// 创建分享
func (dao *ShareDAO) Create(share *Share) error {
	return dao.db.Create(share).Error
}

// 根据 share_id 查找
func (dao *ShareDAO) GetByID(id int64) (*Share, error) {
	var s Share
	if err := dao.db.First(&s, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}
func (dao *ShareDAO) GetByShareID(share_id string) (*Share, error) {
	var s Share
	if err := dao.db.Where("share_id = ?", share_id).First(&s).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

// 检查分享是否过期
func (dao *ShareDAO) IsExpired(id int64) (bool, error) {
	s, err := dao.GetByID(id)
	if err != nil {
		return false, err
	}
	if s == nil {
		return true, nil // 不存在视为过期
	}
	return time.Now().After(s.ExpireAt), nil
}
