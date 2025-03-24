package cache

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

type SMSCodeCache interface {
	Set(ctx context.Context, biz, phone, code string) error
	Verify(ctx context.Context, biz, phone, code string) (bool, error)
	Remove(ctx context.Context, biz, phone string) error
}

type RedisCodeCache struct {
	client      *redis.Client //Redis客户端
	expiration  time.Duration //验证码过期时间
	maxAttempts int           //最大重试次数
}

func NewRedisCodeCache() *RedisCodeCache {
	return &RedisCodeCache{
		client:      redis.NewClient(&redis.Options{}),
		expiration:  10 * time.Second,
		maxAttempts: 3,
	}
}

func (r *RedisCodeCache) Set(ctx context.Context, biz, phone, code string) error {
	key := generateSMSKey(biz, phone)
	attemptsKey := generateAttemptsKey(biz, phone)

	//用管道批量执行命令
	pipe := r.client.Pipeline()
	pipe.Set(ctx, key, code, r.expiration)
	pipe.Set(ctx, attemptsKey, 0, r.expiration) // 重置尝试次数
	_, err := pipe.Exec(ctx)

	return err
}

func (r *RedisCodeCache) Verify(ctx context.Context, biz, phone, code string) (bool, error) {
	key := generateSMSKey(biz, phone)
	attemptsKey := generateAttemptsKey(biz, phone)

	//获取当前的尝试次数
	attempts, err := r.client.Get(ctx, attemptsKey).Int()
	if err != nil && !errors.Is(err, redis.Nil) {
		return false, err
	}
	if attempts > r.maxAttempts {
		return false, fmt.Errorf("验证次数过多，请重新获取")
	}

	//增加尝试次数
	r.client.Incr(ctx, attemptsKey)

	//获取存储的验证码
	storeCode, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, fmt.Errorf("验证码失效")
		}
		return false, err
	}
	//验证码匹配

	if storeCode != code {
		return false, fmt.Errorf("验证码不匹配，请重新输入")
	}
	//获取验证码并删除验证码以及尝试次数
	pipe := r.client.Pipeline()
	pipe.Del(ctx, key)
	pipe.Del(ctx, attemptsKey)
	_, err = pipe.Exec(ctx)
	return true, err
}

func (r *RedisCodeCache) Remove(ctx context.Context, biz, phone string) error {
	key := generateSMSKey(biz, phone)
	attemptsKey := generateAttemptsKey(biz, phone)

	pipe := r.client.Pipeline()
	pipe.Del(ctx, key)
	pipe.Del(ctx, attemptsKey)
	_, err := pipe.Exec(ctx)
	return err
}

// 生成缓存键
func generateSMSKey(biz, phone string) string {
	return fmt.Sprintf("sms:%s:%s", biz, phone)
}

// 生成尝试次数键
func generateAttemptsKey(biz, phone string) string {
	return fmt.Sprintf("sms:attempts:%s:%s", biz, phone)
}
