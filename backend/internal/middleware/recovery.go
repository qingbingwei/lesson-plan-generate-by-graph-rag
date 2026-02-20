package middleware

import (
	"fmt"
	"runtime/debug"

	"lesson-plan/backend/pkg/logger"

	"github.com/gin-gonic/gin"
)

// RecoveryMiddleware 恢复中间件
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 获取堆栈信息
				stack := debug.Stack()

				logger.Error("Panic recovered",
					logger.String("error", fmt.Sprintf("%v", err)),
					logger.String("stack", string(stack)),
					logger.String("path", c.Request.URL.Path),
					logger.String("method", c.Request.Method),
					logger.String("trace_id", TraceIDFromGin(c)),
				)

				abortWithError(c, 500, "INTERNAL_SERVER_ERROR", "服务器内部错误", nil)
			}
		}()

		c.Next()
	}
}

// GinRecovery 返回Gin框架使用的Recovery中间件
func GinRecovery() gin.HandlerFunc {
	return RecoveryMiddleware()
}
