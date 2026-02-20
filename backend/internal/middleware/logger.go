package middleware

import (
	"time"

	"lesson-plan/backend/internal/observability"
	"lesson-plan/backend/pkg/logger"

	"github.com/gin-gonic/gin"
)

// LoggerMiddleware 日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		traceID := BindTraceID(c)

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

		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}

		observability.RecordHTTPRequest(method, route, statusCode, latency)

		// 根据状态码选择日志级别
		logFunc := logger.Info
		if statusCode >= 400 && statusCode < 500 {
			logFunc = logger.Warn
		} else if statusCode >= 500 {
			logFunc = logger.Error
		}

		logFunc("HTTP request",
			logger.String("trace_id", traceID),
			logger.String("client_ip", clientIP),
			logger.String("method", method),
			logger.String("path", path),
			logger.String("route", route),
			logger.Int("status", statusCode),
			logger.Int("body_size", bodySize),
			logger.Duration("latency", latency),
		)
	}
}
