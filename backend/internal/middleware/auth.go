package middleware

import (
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
			abortWithError(c, 401, "AUTH_MISSING_HEADER", "缺少认证头", nil)
			return
		}

		fields := strings.Fields(authHeader)
		if len(fields) < 2 {
			abortWithError(c, 401, "AUTH_INVALID_HEADER", "认证头格式错误", nil)
			return
		}

		authType := fields[0]
		if !strings.EqualFold(authType, AuthorizationTypeBearer) {
			abortWithError(c, 401, "AUTH_UNSUPPORTED_TYPE", "不支持的认证类型", nil)
			return
		}

		accessToken := fields[1]
		claims, err := jwtManager.ValidateToken(accessToken)
		if err != nil {
			abortWithError(c, 401, "AUTH_INVALID_TOKEN", "无效的令牌", err.Error())
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
			abortWithError(c, 401, "AUTH_UNAUTHORIZED", "未认证", nil)
			return
		}

		userClaims, ok := claims.(*jwt.Claims)
		if !ok {
			abortWithError(c, 401, "AUTH_INVALID_CLAIMS", "无效的令牌载荷", nil)
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
			abortWithError(c, 403, "AUTH_FORBIDDEN", "权限不足", nil)
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
