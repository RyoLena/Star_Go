// pkg/middleware/rate_limit.go
package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter 简单的内存限流器
type RateLimiter struct {
	ips    map[string][]time.Time
	mu     sync.Mutex
	limit  int           // 时间窗口内允许的最大请求数
	window time.Duration // 时间窗口大小
}

// NewRateLimiter 创建新的限流器
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		ips:    make(map[string][]time.Time),
		limit:  limit,
		window: window,
	}
}

// 清理过期的请求记录
func (rl *RateLimiter) cleanupOldRequests() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for ip, requests := range rl.ips {
		var validRequests []time.Time
		for _, requestTime := range requests {
			if now.Sub(requestTime) <= rl.window {
				validRequests = append(validRequests, requestTime)
			}
		}

		if len(validRequests) > 0 {
			rl.ips[ip] = validRequests
		} else {
			delete(rl.ips, ip)
		}
	}
}

// 检查IP是否超过限制
func (rl *RateLimiter) isAllowed(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// 清理过期请求
	var validRequests []time.Time
	for _, requestTime := range rl.ips[ip] {
		if now.Sub(requestTime) <= rl.window {
			validRequests = append(validRequests, requestTime)
		}
	}

	// 如果请求数量超过限制，拒绝请求
	if len(validRequests) >= rl.limit {
		rl.ips[ip] = validRequests
		return false
	}

	// 记录新请求
	rl.ips[ip] = append(validRequests, now)
	return true
}

// 限流中间件
func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(limit, window)

	// 定期清理过期请求
	go func() {
		ticker := time.NewTicker(window / 2)
		defer ticker.Stop()

		for range ticker.C {
			limiter.cleanupOldRequests()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()

		if !limiter.isAllowed(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "请求过于频繁，请稍后再试",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
