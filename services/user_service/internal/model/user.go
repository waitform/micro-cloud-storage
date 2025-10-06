package model

import "time"

// User 表结构（对应数据库）
type User struct {
	ID         int64     `gorm:"primaryKey;autoIncrement"`
	Username   string    `gorm:"uniqueIndex;size:64;not null"`
	Password   string    `gorm:"size:128;not null"`
	Email      string    `gorm:"size:128"`
	Avatar     string    `gorm:"size:255"`
	TotalSpace int64     `gorm:"not null;default:10737418240"` // 默认10GB
	UsedSpace  int64     `gorm:"not null;default:0"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime"`
}
