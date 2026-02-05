package handler

import (
	"errors"
	"net/http"

	"lesson-plan/backend/internal/middleware"
	"lesson-plan/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService service.AuthService
	userService service.UserService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authService service.AuthService, userService service.UserService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userService: userService,
	}
}

// Register 注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	user, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrUserExists) {
			Error(c, http.StatusConflict, "用户已存在", nil)
			return
		}
		Error(c, http.StatusInternalServerError, "注册失败", err.Error())
		return
	}

	Success(c, user.ToProfile())
}

// Login 登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			Error(c, http.StatusUnauthorized, "用户名或密码错误", nil)
			return
		}
		if errors.Is(err, service.ErrUserInactive) {
			Error(c, http.StatusForbidden, "用户已被禁用", nil)
			return
		}
		Error(c, http.StatusInternalServerError, "登录失败", err.Error())
		return
	}

	Success(c, resp)
}

// RefreshToken 刷新Token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	resp, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		Error(c, http.StatusUnauthorized, "Token无效或已过期", err.Error())
		return
	}

	Success(c, resp)
}

// Logout 登出
func (h *AuthHandler) Logout(c *gin.Context) {
	// 简单登出，客户端清除token即可
	SuccessWithMessage(c, "登出成功", nil)
}

// ChangePassword 修改密码
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, "未认证", nil)
		return
	}

	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	userUUID, _ := uuid.Parse(userID)
	if err := h.userService.ChangePassword(c.Request.Context(), userUUID, req.OldPassword, req.NewPassword); err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			Error(c, http.StatusBadRequest, "旧密码错误", nil)
			return
		}
		Error(c, http.StatusInternalServerError, "修改密码失败", err.Error())
		return
	}

	SuccessWithMessage(c, "密码修改成功", nil)
}

// GetCurrentUser 获取当前用户信息
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, "未认证", nil)
		return
	}

	userUUID, _ := uuid.Parse(userID)
	profile, err := h.userService.GetProfile(c.Request.Context(), userUUID)
	if err != nil {
		Error(c, http.StatusNotFound, "用户不存在", nil)
		return
	}

	Success(c, profile)
}
