package middleware

import (
	"time"

	"lesson-plan/backend/pkg/logger"

	"github.com/gin-gonic/gin"
)

// LoggerMiddleware 日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 生成请求ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()

		// 计算延迟
		latency := time.Since(start)

		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()

		if raw != "" {
			path = path + "?" + raw
		}

		// 根据状态码选择日志级别
		logFunc := logger.Info
		if statusCode >= 400 && statusCode < 500 {
			logFunc = logger.Warn
		} else if statusCode >= 500 {
			logFunc = logger.Error
		}

		logFunc("HTTP request",
			logger.String("request_id", requestID),
			logger.String("client_ip", clientIP),
			logger.String("method", method),
			logger.String("path", path),
			logger.Int("status", statusCode),
			logger.Int("body_size", bodySize),
			logger.Duration("latency", latency),
		)
	}
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 16)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
