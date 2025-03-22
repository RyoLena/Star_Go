// pkg/cache/redis.go
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"star-go/pkg/config"
	"star-go/pkg/logger"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Redis缓存实现
type redisCache struct {
	client     *redis.Client
	prefix     string
	defaultTTL time.Duration
	enableLog  bool
}

// 初始化Redis缓存
func initRedisCache() error {
	cfg := config.GetConfig().Cache

	// 创建Redis客户端
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  cfg.DialTimeout * time.Second,
		ReadTimeout:  cfg.ReadTimeout * time.Second,
		WriteTimeout: cfg.WriteTimeout * time.Second,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("连接Redis失败: %w", err)
	}

	// 创建Redis缓存实例
	globalCache = &redisCache{
		client:     client,
		prefix:     cfg.Prefix,
		defaultTTL: cfg.DefaultTTL * time.Second,
		enableLog:  cfg.EnableLog,
	}

	logger.GetLogger().Info("Redis缓存连接成功",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.Int("db", cfg.DB))

	return nil
}

// 生成带前缀的键
func (r *redisCache) prefixKey(key string) string {
	return r.prefix + key
}

// 记录操作日志
func (r *redisCache) logOperation(operation, key string, err error) {
	if !r.enableLog {
		return
	}

	if err != nil {
		logger.GetLogger().Debug("Redis操作",
			zap.String("操作", operation),
			zap.String("键", key),
			zap.Error(err))
	} else {
		logger.GetLogger().Debug("Redis操作",
			zap.String("操作", operation),
			zap.String("键", key),
			zap.String("结果", "成功"))
	}
}

// Set 设置缓存
func (r *redisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	prefixedKey := r.prefixKey(key)

	// 如果未指定过期时间，使用默认值
	if expiration == 0 {
		expiration = r.defaultTTL
	}

	// 序列化值为JSON
	data, err := json.Marshal(value)
	if err != nil {
		r.logOperation("SET", key, err)
		return fmt.Errorf("序列化缓存值失败: %w", err)
	}

	// 设置缓存
	err = r.client.Set(ctx, prefixedKey, data, expiration).Err()
	r.logOperation("SET", key, err)

	if err != nil {
		return fmt.Errorf("设置缓存失败: %w", err)
	}

	return nil
}

// Get 获取缓存
func (r *redisCache) Get(ctx context.Context, key string, dest interface{}) error {
	prefixedKey := r.prefixKey(key)

	// 获取缓存
	data, err := r.client.Get(ctx, prefixedKey).Bytes()
	r.logOperation("GET", key, err)

	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("缓存键不存在: %s", key)
		}
		return fmt.Errorf("获取缓存失败: %w", err)
	}

	// 反序列化JSON到目标结构
	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("反序列化缓存值失败: %w", err)
	}

	return nil
}

// Delete 删除缓存
func (r *redisCache) Delete(ctx context.Context, key string) error {
	prefixedKey := r.prefixKey(key)

	// 删除缓存
	err := r.client.Del(ctx, prefixedKey).Err()
	r.logOperation("DEL", key, err)

	if err != nil {
		return fmt.Errorf("删除缓存失败: %w", err)
	}

	return nil
}

// Exists 检查键是否存在
func (r *redisCache) Exists(ctx context.Context, key string) (bool, error) {
	prefixedKey := r.prefixKey(key)

	// 检查键是否存在
	result, err := r.client.Exists(ctx, prefixedKey).Result()
	r.logOperation("EXISTS", key, err)

	if err != nil {
		return false, fmt.Errorf("检查缓存键是否存在失败: %w", err)
	}

	return result > 0, nil
}

// Expire 设置过期时间
func (r *redisCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	prefixedKey := r.prefixKey(key)

	// 设置过期时间
	err := r.client.Expire(ctx, prefixedKey, expiration).Err()
	r.logOperation("EXPIRE", key, err)

	if err != nil {
		return fmt.Errorf("设置缓存过期时间失败: %w", err)
	}

	return nil
}

// FlushDB 清空当前数据库
func (r *redisCache) FlushDB(ctx context.Context) error {
	// 清空数据库
	err := r.client.FlushDB(ctx).Err()
	r.logOperation("FLUSHDB", "*", err)

	if err != nil {
		return fmt.Errorf("清空缓存数据库失败: %w", err)
	}

	return nil
}

// Close 关闭缓存连接
func (r *redisCache) Close() error {
	if r.client != nil {
		err := r.client.Close()
		if err != nil {
			return fmt.Errorf("关闭Redis连接失败: %w", err)
		}
		logger.GetLogger().Info("Redis缓存连接已关闭")
	}
	return nil
}

// GetClient 获取原始客户端
func (r *redisCache) GetClient() interface{} {
	return r.client
}
