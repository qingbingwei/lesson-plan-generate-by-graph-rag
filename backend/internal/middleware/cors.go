package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSConfig CORS配置
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig 默认CORS配置
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			"Accept",
			"Accept-Language",
			"Content-Type",
			"Content-Language",
			"Authorization",
			"Origin",
			"X-Requested-With",
			"X-Request-ID",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"Content-Type",
			"X-Request-ID",
		},
		AllowCredentials: true,
		MaxAge:           86400,
	}
}

// CORSMiddleware CORS中间件
func CORSMiddleware(config CORSConfig) gin.HandlerFunc {
	allowOrigins := strings.Join(config.AllowOrigins, ", ")
	allowMethods := strings.Join(config.AllowMethods, ", ")
	allowHeaders := strings.Join(config.AllowHeaders, ", ")
	exposeHeaders := strings.Join(config.ExposeHeaders, ", ")

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 检查origin是否允许
		if len(config.AllowOrigins) == 1 && config.AllowOrigins[0] == "*" {
			c.Header("Access-Control-Allow-Origin", "*")
		} else {
			for _, o := range config.AllowOrigins {
				if o == origin {
					c.Header("Access-Control-Allow-Origin", origin)
					break
				}
			}
		}

		if config.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if c.Request.Method == http.MethodOptions {
			c.Header("Access-Control-Allow-Methods", allowMethods)
			c.Header("Access-Control-Allow-Headers", allowHeaders)
			c.Header("Access-Control-Expose-Headers", exposeHeaders)
			c.Header("Access-Control-Max-Age", "86400")
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Header("Access-Control-Allow-Origin", allowOrigins)
		c.Header("Access-Control-Expose-Headers", exposeHeaders)

		c.Next()
	}
}

// NewCORSMiddleware 创建CORS中间件
func NewCORSMiddleware(allowOrigins []string, allowCredentials bool) gin.HandlerFunc {
	config := DefaultCORSConfig()
	if len(allowOrigins) > 0 {
		config.AllowOrigins = allowOrigins
	}
	config.AllowCredentials = allowCredentials
	return CORSMiddleware(config)
}
