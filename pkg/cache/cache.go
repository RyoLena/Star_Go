// Package cache pkg/cache/cache.go
package cache

import (
	"context"
	"fmt"
	"star-go/pkg/config"
	"star-go/pkg/logger"
	"time"

	"go.uber.org/zap"
)

// Cache 定义缓存接口
type Cache interface {
	// Set 设置缓存
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error

	// Get 获取缓存
	Get(ctx context.Context, key string, dest interface{}) error

	// Delete 删除缓存
	Delete(ctx context.Context, key string) error

	// Exists 检查键是否存在
	Exists(ctx context.Context, key string) (bool, error)

	// Expire 设置过期时间
	Expire(ctx context.Context, key string, expiration time.Duration) error

	// FlushDB 清空当前数据库
	FlushDB(ctx context.Context) error

	// Close 关闭缓存连接
	Close() error

	// GetClient 获取原始客户端
	GetClient() interface{}
}

// 全局缓存实例
var globalCache Cache

// InitCache 初始化缓存
func InitCache() error {
	cfg := config.GetConfig().Cache

	// 根据配置选择缓存实现
	var err error
	switch cfg.Type {
	case "redis":
		err = initRedisCache()
	case "memory":
		err = initMemoryCache()
	default:
		logger.GetLogger().Warn("未知的缓存类型，使用内存缓存作为默认实现", zap.String("type", cfg.Type))
		err = initMemoryCache()
	}

	if err != nil {
		return fmt.Errorf("初始化缓存失败: %w", err)
	}

	logger.GetLogger().Info("缓存系统初始化成功", zap.String("type", cfg.Type))
	return nil
}

// GetCache 获取缓存实例
func GetCache() Cache {
	return globalCache
}

// CloseCache 关闭缓存连接
func CloseCache() error {
	if globalCache != nil {
		return globalCache.Close()
	}
	return nil
}
