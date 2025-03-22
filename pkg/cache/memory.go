// Package cache pkg/cache/memory.go
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"star-go/pkg/config"
	"star-go/pkg/logger"
	"sync"
	"time"

	"go.uber.org/zap"
)

// 缓存项结构
type cacheItem struct {
	Value      []byte
	Expiration int64
}

// Expired 判断缓存项是否过期
func (item cacheItem) Expired() bool {
	if item.Expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > item.Expiration
}

// 内存缓存实现
type memoryCache struct {
	items      map[string]cacheItem
	mu         sync.RWMutex
	prefix     string
	defaultTTL time.Duration
	enableLog  bool
	janitor    *janitor
}

// 清理器
type janitor struct {
	Interval time.Duration
	stop     chan bool
}

// 初始化内存缓存
func initMemoryCache() error {
	cfg := config.GetConfig().Cache

	// 创建内存缓存实例
	cache := &memoryCache{
		items:      make(map[string]cacheItem),
		prefix:     cfg.Prefix,
		defaultTTL: cfg.DefaultTTL * time.Second,
		enableLog:  cfg.EnableLog,
	}

	// 启动定期清理过期项的协程
	cache.janitor = &janitor{
		Interval: time.Minute,
		stop:     make(chan bool),
	}

	go cache.janitor.Run(cache)

	globalCache = cache

	logger.GetLogger().Info("内存缓存初始化成功")

	return nil
}

// Run 运行清理器
func (j *janitor) Run(c *memoryCache) {
	ticker := time.NewTicker(j.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.DeleteExpired()
		case <-j.stop:
			return
		}
	}
}

// Stop 停止清理器
func (j *janitor) Stop() {
	j.stop <- true
}

// 删除过期项
func (m *memoryCache) DeleteExpired() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().UnixNano()
	for k, v := range m.items {
		if v.Expiration > 0 && now > v.Expiration {
			delete(m.items, k)
			if m.enableLog {
				logger.GetLogger().Debug("内存缓存过期删除", zap.String("键", k))
			}
		}
	}
}

// 生成带前缀的键
func (m *memoryCache) prefixKey(key string) string {
	return m.prefix + key
}

// 记录操作日志
func (m *memoryCache) logOperation(operation, key string, err error) {
	if !m.enableLog {
		return
	}

	if err != nil {
		logger.GetLogger().Debug("内存缓存操作",
			zap.String("操作", operation),
			zap.String("键", key),
			zap.Error(err))
	} else {
		logger.GetLogger().Debug("内存缓存操作",
			zap.String("操作", operation),
			zap.String("键", key),
			zap.String("结果", "成功"))
	}
}

// Set 设置缓存
func (m *memoryCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	prefixedKey := m.prefixKey(key)

	// 如果未指定过期时间，使用默认值
	if expiration == 0 {
		expiration = m.defaultTTL
	}

	// 序列化值为JSON
	data, err := json.Marshal(value)
	if err != nil {
		m.logOperation("SET", key, err)
		return fmt.Errorf("序列化缓存值失败: %w", err)
	}

	// 计算过期时间
	var exp int64
	if expiration > 0 {
		exp = time.Now().Add(expiration).UnixNano()
	}

	m.mu.Lock()
	m.items[prefixedKey] = cacheItem{
		Value:      data,
		Expiration: exp,
	}
	m.mu.Unlock()

	m.logOperation("SET", key, nil)
	return nil
}

// Get 获取缓存
func (m *memoryCache) Get(ctx context.Context, key string, dest interface{}) error {
	prefixedKey := m.prefixKey(key)

	m.mu.RLock()
	item, found := m.items[prefixedKey]
	if !found {
		m.mu.RUnlock()
		err := fmt.Errorf("缓存键不存在: %s", key)
		m.logOperation("GET", key, err)
		return err
	}

	// 检查是否过期
	if item.Expired() {
		m.mu.RUnlock()
		m.mu.Lock()
		delete(m.items, prefixedKey)
		m.mu.Unlock()
		err := fmt.Errorf("缓存键已过期: %s", key)
		m.logOperation("GET", key, err)
		return err
	}

	data := item.Value
	m.mu.RUnlock()

	// 反序列化JSON到目标结构
	if err := json.Unmarshal(data, dest); err != nil {
		m.logOperation("GET", key, err)
		return fmt.Errorf("反序列化缓存值失败: %w", err)
	}

	m.logOperation("GET", key, nil)
	return nil
}

// Delete 删除缓存
func (m *memoryCache) Delete(ctx context.Context, key string) error {
	prefixedKey := m.prefixKey(key)

	m.mu.Lock()
	delete(m.items, prefixedKey)
	m.mu.Unlock()

	m.logOperation("DEL", key, nil)
	return nil
}

// Exists 检查键是否存在
func (m *memoryCache) Exists(ctx context.Context, key string) (bool, error) {
	prefixedKey := m.prefixKey(key)

	m.mu.RLock()
	item, found := m.items[prefixedKey]
	m.mu.RUnlock()

	if !found {
		m.logOperation("EXISTS", key, nil)
		return false, nil
	}

	// 检查是否过期
	if item.Expired() {
		m.mu.Lock()
		delete(m.items, prefixedKey)
		m.mu.Unlock()
		m.logOperation("EXISTS", key, nil)
		return false, nil
	}

	m.logOperation("EXISTS", key, nil)
	return true, nil
}

// Expire 设置过期时间
func (m *memoryCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	prefixedKey := m.prefixKey(key)

	m.mu.Lock()
	defer m.mu.Unlock()

	item, found := m.items[prefixedKey]
	if !found {
		err := fmt.Errorf("缓存键不存在: %s", key)
		m.logOperation("EXPIRE", key, err)
		return err
	}

	// 计算新的过期时间
	var exp int64
	if expiration > 0 {
		exp = time.Now().Add(expiration).UnixNano()
	}

	// 更新过期时间
	item.Expiration = exp
	m.items[prefixedKey] = item

	m.logOperation("EXPIRE", key, nil)
	return nil
}

// FlushDB 清空当前数据库
func (m *memoryCache) FlushDB(ctx context.Context) error {
	m.mu.Lock()
	m.items = make(map[string]cacheItem)
	m.mu.Unlock()

	m.logOperation("FLUSHDB", "*", nil)
	return nil
}

// Close 关闭缓存连接
func (m *memoryCache) Close() error {
	if m.janitor != nil {
		m.janitor.Stop()
	}

	m.mu.Lock()
	m.items = nil
	m.mu.Unlock()

	logger.GetLogger().Info("内存缓存已关闭")
	return nil
}

// GetClient 获取原始客户端
func (m *memoryCache) GetClient() interface{} {
	return m.items
}
