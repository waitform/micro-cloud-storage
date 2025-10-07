package utils

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient Redis客户端包装类
type RedisClient struct {
	client *redis.Client
}

// RedisConfig Redis配置结构
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// NewRedisClient 创建新的Redis客户端实例
func NewRedisClient(config RedisConfig) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})

	// 测试连接
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	return &RedisClient{client: client}, nil
}

// Set 设置键值对
func (r *RedisClient) Set(key, value string, expiration time.Duration) error {
	return r.client.Set(context.Background(), key, value, expiration).Err()
}

// Get 获取键对应的值
func (r *RedisClient) Get(key string) (string, error) {
	return r.client.Get(context.Background(), key).Result()
}

// Del 删除键
func (r *RedisClient) Del(keys ...string) error {
	return r.client.Del(context.Background(), keys...).Err()
}

// Exists 检查键是否存在
func (r *RedisClient) Exists(keys ...string) (int64, error) {
	return r.client.Exists(context.Background(), keys...).Result()
}

// GetClient 获取底层redis.Client实例（仅供内部使用）
func (r *RedisClient) GetClient() *redis.Client {
	return r.client
}

