package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter 限流器接口
type RateLimiter interface {
	Allow(c *gin.Context) bool
}

// TokenBucketLimiter 令牌桶限流器
type TokenBucketLimiter struct {
	rate       float64
	bucketSize int
	tokens     float64
	lastTime   time.Time
	mu         sync.Mutex
}

// NewTokenBucketLimiter 创建令牌桶限流器
func NewTokenBucketLimiter(rate float64, bucketSize int) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		rate:       rate,
		bucketSize: bucketSize,
		tokens:     float64(bucketSize),
		lastTime:   time.Now(),
	}
}

// Allow 检查是否允许请求
func (l *TokenBucketLimiter) Allow(c *gin.Context) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(l.lastTime).Seconds()
	l.lastTime = now

	// 添加令牌
	l.tokens += elapsed * l.rate
	if l.tokens > float64(l.bucketSize) {
		l.tokens = float64(l.bucketSize)
	}

	if l.tokens >= 1 {
		l.tokens--
		return true
	}

	return false
}

// IPRateLimiter IP限流器
type IPRateLimiter struct {
	limiters map[string]*TokenBucketLimiter
	rate     float64
	size     int
	mu       sync.Mutex
}

// NewIPRateLimiter 创建IP限流器
func NewIPRateLimiter(rate float64, bucketSize int) *IPRateLimiter {
	return &IPRateLimiter{
		limiters: make(map[string]*TokenBucketLimiter),
		rate:     rate,
		size:     bucketSize,
	}
}

// Allow 检查是否允许请求
func (l *IPRateLimiter) Allow(c *gin.Context) bool {
	ip := c.ClientIP()

	l.mu.Lock()
	limiter, exists := l.limiters[ip]
	if !exists {
		limiter = NewTokenBucketLimiter(l.rate, l.size)
		l.limiters[ip] = limiter
	}
	l.mu.Unlock()

	return limiter.Allow(c)
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(limiter RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !limiter.Allow(c) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "请求过于频繁，请稍后再试",
			})
			return
		}
		c.Next()
	}
}

// NewRateLimitMiddleware 创建限流中间件
func NewRateLimitMiddleware(rate float64, bucketSize int) gin.HandlerFunc {
	limiter := NewIPRateLimiter(rate, bucketSize)
	return RateLimitMiddleware(limiter)
}
