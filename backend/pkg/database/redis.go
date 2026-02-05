package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"lesson-plan/backend/internal/config"
	"lesson-plan/backend/pkg/logger"

	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

// InitRedis 初始化Redis连接
func InitRedis(cfg *config.RedisConfig) (*redis.Client, error) {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	logger.Info("Redis connected successfully",
		logger.String("addr", cfg.Addr()),
		logger.Int("db", cfg.DB),
	)

	return redisClient, nil
}

// GetRedis 获取Redis客户端
func GetRedis() *redis.Client {
	return redisClient
}

// CloseRedis 关闭Redis连接
func CloseRedis() error {
	if redisClient != nil {
		return redisClient.Close()
	}
	return nil
}

// Set 设置键值
func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return redisClient.Set(ctx, key, data, expiration).Err()
}

// Get 获取值
func Get(ctx context.Context, key string, dest interface{}) error {
	data, err := redisClient.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// GetString 获取字符串值
func GetString(ctx context.Context, key string) (string, error) {
	return redisClient.Get(ctx, key).Result()
}

// Delete 删除键
func Delete(ctx context.Context, keys ...string) error {
	return redisClient.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在
func Exists(ctx context.Context, key string) (bool, error) {
	n, err := redisClient.Exists(ctx, key).Result()
	return n > 0, err
}

// SetNX 设置键值（如果不存在）
func SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return false, err
	}
	return redisClient.SetNX(ctx, key, data, expiration).Result()
}

// Incr 递增
func Incr(ctx context.Context, key string) (int64, error) {
	return redisClient.Incr(ctx, key).Result()
}

// Decr 递减
func Decr(ctx context.Context, key string) (int64, error) {
	return redisClient.Decr(ctx, key).Result()
}

// Expire 设置过期时间
func Expire(ctx context.Context, key string, expiration time.Duration) error {
	return redisClient.Expire(ctx, key, expiration).Err()
}

// TTL 获取剩余过期时间
func TTL(ctx context.Context, key string) (time.Duration, error) {
	return redisClient.TTL(ctx, key).Result()
}

// HSet 设置哈希字段
func HSet(ctx context.Context, key, field string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return redisClient.HSet(ctx, key, field, data).Err()
}

// HGet 获取哈希字段
func HGet(ctx context.Context, key, field string, dest interface{}) error {
	data, err := redisClient.HGet(ctx, key, field).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// HGetAll 获取所有哈希字段
func HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return redisClient.HGetAll(ctx, key).Result()
}

// HDel 删除哈希字段
func HDel(ctx context.Context, key string, fields ...string) error {
	return redisClient.HDel(ctx, key, fields...).Err()
}

// LPush 列表左侧推入
func LPush(ctx context.Context, key string, values ...interface{}) error {
	return redisClient.LPush(ctx, key, values...).Err()
}

// RPush 列表右侧推入
func RPush(ctx context.Context, key string, values ...interface{}) error {
	return redisClient.RPush(ctx, key, values...).Err()
}

// LRange 获取列表范围
func LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return redisClient.LRange(ctx, key, start, stop).Result()
}

// Keys 获取匹配的键
func Keys(ctx context.Context, pattern string) ([]string, error) {
	return redisClient.Keys(ctx, pattern).Result()
}

// FlushDB 清空数据库（谨慎使用）
func FlushDB(ctx context.Context) error {
	return redisClient.FlushDB(ctx).Err()
}
