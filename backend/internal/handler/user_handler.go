package handler

import (
	"net/http"

	"lesson-plan/backend/internal/middleware"
	"lesson-plan/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService service.UserService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetProfile 获取用户资料
func (h *UserHandler) GetProfile(c *gin.Context) {
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

// UpdateProfile 更新用户资料
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, "未认证", nil)
		return
	}

	var req service.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	userUUID, _ := uuid.Parse(userID)
	user, err := h.userService.UpdateProfile(c.Request.Context(), userUUID, &req)
	if err != nil {
		Error(c, http.StatusInternalServerError, "更新失败", err.Error())
		return
	}

	Success(c, user.ToProfile())
}

// UploadAvatar 上传头像
func (h *UserHandler) UploadAvatar(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, "未认证", nil)
		return
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		Error(c, http.StatusBadRequest, "请上传文件", nil)
		return
	}

	// 简化处理：实际应保存到对象存储
	avatarURL := "/uploads/avatars/" + file.Filename

	userUUID, _ := uuid.Parse(userID)
	_, err = h.userService.UpdateProfile(c.Request.Context(), userUUID, &service.UpdateUserRequest{
		AvatarURL: avatarURL,
	})
	if err != nil {
		Error(c, http.StatusInternalServerError, "上传失败", err.Error())
		return
	}

	Success(c, gin.H{"avatar_url": avatarURL})
}
