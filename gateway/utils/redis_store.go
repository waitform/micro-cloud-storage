package utils

import (
	"context"
	"time"

	base64Captcha "github.com/mojocn/base64Captcha"
	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	Redis  *redis.Client
	Expire time.Duration
}

// NewRedisStore 创建新的Redis存储实例
func NewRedisStore(redisClient *redis.Client) *RedisStore {
	return &RedisStore{
		Redis:  redisClient,
		Expire: time.Minute * 5, // 默认5分钟过期时间
	}
}

// Set 存储验证码
func (rs *RedisStore) Set(id string, value string) error {
	ctx := context.Background()
	return rs.Redis.Set(ctx, "captcha:"+id, value, rs.Expire).Err()
}

// Get 获取验证码
func (rs *RedisStore) Get(id string, clear bool) string {
	ctx := context.Background()
	key := "captcha:" + id
	val, err := rs.Redis.Get(ctx, key).Result()
	if err != nil {
		return ""
	}
	if clear {
		rs.Redis.Del(ctx, key)
	}
	return val
}

// Verify 验证验证码
func (rs *RedisStore) Verify(id, answer string, clear bool) bool {
	v := rs.Get(id, clear)
	return v == answer
}

// 确保RedisStore实现了base64Captcha.Store接口
var _ base64Captcha.Store = (*RedisStore)(nil)