package database

import (
	"fmt"
	"time"

	"cloud-storage/services/user_service/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 结构体用于封装数据库连接
type DB struct {
	*gorm.DB
}

// NewDB 创建新的数据库连接实例
func NewDB(cfg *config.Config) (*DB, error) {
	// 构造DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name)

	// 配置GORM
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)                 // 空闲连接池中连接的最大数量
	sqlDB.SetMaxOpenConns(100)                // 打开数据库连接的最大数量
	sqlDB.SetConnMaxLifetime(time.Hour)       // 连接可复用的最大时间

	return &DB{db}, nil
}