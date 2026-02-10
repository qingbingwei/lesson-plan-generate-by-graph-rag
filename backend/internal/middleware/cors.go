package middleware

import (
	"net/http"
	"strconv"
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
			"X-Generation-Api-Key",
			"X-Embedding-Api-Key",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"Content-Type",
			"X-Request-ID",
			"X-Generation-Api-Key",
			"X-Embedding-Api-Key",
		},
		AllowCredentials: true,
		MaxAge:           86400,
	}
}

// CORSMiddleware CORS中间件
func CORSMiddleware(config CORSConfig) gin.HandlerFunc {
	allowMethods := strings.Join(config.AllowMethods, ", ")
	allowHeaders := strings.Join(config.AllowHeaders, ", ")
	exposeHeaders := strings.Join(config.ExposeHeaders, ", ")
	maxAge := strconv.Itoa(config.MaxAge)

	allowAllOrigins := len(config.AllowOrigins) == 1 && config.AllowOrigins[0] == "*"

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 检查origin是否允许
		switch {
		case allowAllOrigins && config.AllowCredentials && origin != "":
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		case allowAllOrigins:
			c.Header("Access-Control-Allow-Origin", "*")
		case origin != "" && isOriginAllowed(config.AllowOrigins, origin):
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		}

		if config.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if exposeHeaders != "" {
			c.Header("Access-Control-Expose-Headers", exposeHeaders)
		}

		if c.Request.Method == http.MethodOptions {
			c.Header("Access-Control-Allow-Methods", allowMethods)
			c.Header("Access-Control-Allow-Headers", allowHeaders)
			if maxAge != "" {
				c.Header("Access-Control-Max-Age", maxAge)
			}
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func isOriginAllowed(allowedOrigins []string, origin string) bool {
	for _, o := range allowedOrigins {
		if o == "*" || o == origin {
			return true
		}
	}
	return false
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
