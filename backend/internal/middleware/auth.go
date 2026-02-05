package middleware

import (
	"net/http"
	"strings"

	"lesson-plan/backend/pkg/jwt"

	"github.com/gin-gonic/gin"
)

const (
	AuthorizationHeaderKey  = "Authorization"
	AuthorizationTypeBearer = "Bearer"
	AuthorizationPayloadKey = "authorization_payload"
)

// AuthMiddleware 认证中间件
func AuthMiddleware(jwtManager *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeaderKey)
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "缺少认证头",
			})
			return
		}

		fields := strings.Fields(authHeader)
		if len(fields) < 2 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "认证头格式错误",
			})
			return
		}

		authType := fields[0]
		if !strings.EqualFold(authType, AuthorizationTypeBearer) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "不支持的认证类型",
			})
			return
		}

		accessToken := fields[1]
		claims, err := jwtManager.ValidateToken(accessToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效的令牌: " + err.Error(),
			})
			return
		}

		c.Set(AuthorizationPayloadKey, claims)
		c.Next()
	}
}

// OptionalAuthMiddleware 可选认证中间件
func OptionalAuthMiddleware(jwtManager *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeaderKey)
		if authHeader == "" {
			c.Next()
			return
		}

		fields := strings.Fields(authHeader)
		if len(fields) < 2 {
			c.Next()
			return
		}

		authType := fields[0]
		if !strings.EqualFold(authType, AuthorizationTypeBearer) {
			c.Next()
			return
		}

		accessToken := fields[1]
		claims, err := jwtManager.ValidateToken(accessToken)
		if err != nil {
			c.Next()
			return
		}

		c.Set(AuthorizationPayloadKey, claims)
		c.Next()
	}
}

// RoleMiddleware 角色中间件
func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get(AuthorizationPayloadKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未认证",
			})
			return
		}

		userClaims, ok := claims.(*jwt.Claims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效的令牌载荷",
			})
			return
		}

		allowed := false
		for _, role := range allowedRoles {
			if userClaims.Role == role {
				allowed = true
				break
			}
		}

		if !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "权限不足",
			})
			return
		}

		c.Next()
	}
}

// GetCurrentUserID 获取当前用户ID
func GetCurrentUserID(c *gin.Context) (string, bool) {
	claims, exists := c.Get(AuthorizationPayloadKey)
	if !exists {
		return "", false
	}

	userClaims, ok := claims.(*jwt.Claims)
	if !ok {
		return "", false
	}

	return userClaims.UserID, true
}

// GetCurrentClaims 获取当前用户声明
func GetCurrentClaims(c *gin.Context) (*jwt.Claims, bool) {
	claims, exists := c.Get(AuthorizationPayloadKey)
	if !exists {
		return nil, false
	}

	cl, ok := claims.(*jwt.Claims)
	return cl, ok
}
