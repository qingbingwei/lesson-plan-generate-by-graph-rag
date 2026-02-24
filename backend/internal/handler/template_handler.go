package handler

import (
	"errors"
	"net/http"

	"lesson-plan/backend/internal/middleware"
	"lesson-plan/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TemplateHandler 模板库处理器。
type TemplateHandler struct {
	templateService service.TemplateService
}

// NewTemplateHandler 创建模板处理器。
func NewTemplateHandler(templateService service.TemplateService) *TemplateHandler {
	return &TemplateHandler{
		templateService: templateService,
	}
}

func (h *TemplateHandler) resolveUserID(c *gin.Context) (uuid.UUID, bool) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		ErrorWithCode(c, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "未认证", nil)
		return uuid.Nil, false
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		ErrorWithCode(c, http.StatusUnauthorized, "AUTH_INVALID_CLAIMS", "无效的用户标识", nil)
		return uuid.Nil, false
	}
	return uid, true
}

// List 模板列表（内置 + 当前用户私有模板）。
func (h *TemplateHandler) List(c *gin.Context) {
	userID, ok := h.resolveUserID(c)
	if !ok {
		return
	}

	templates, err := h.templateService.List(c.Request.Context(), userID)
	if err != nil {
		Error(c, http.StatusInternalServerError, "获取模板列表失败", err.Error())
		return
	}

	Success(c, templates)
}

// Get 模板详情。
func (h *TemplateHandler) Get(c *gin.Context) {
	userID, ok := h.resolveUserID(c)
	if !ok {
		return
	}

	templateID := c.Param("id")
	template, err := h.templateService.Get(c.Request.Context(), templateID, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTemplateNotFound):
			ErrorWithCode(c, http.StatusNotFound, "TEMPLATE_NOT_FOUND", "模板不存在", nil)
		case errors.Is(err, service.ErrTemplateForbidden):
			ErrorWithCode(c, http.StatusForbidden, "TEMPLATE_FORBIDDEN", "无权访问该模板", nil)
		default:
			Error(c, http.StatusInternalServerError, "获取模板详情失败", err.Error())
		}
		return
	}

	Success(c, template)
}

// Create 创建模板。
func (h *TemplateHandler) Create(c *gin.Context) {
	userID, ok := h.resolveUserID(c)
	if !ok {
		return
	}

	var req service.CreateLessonTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithCode(c, http.StatusBadRequest, "TEMPLATE_INVALID_REQUEST", "模板参数错误", err.Error())
		return
	}

	template, err := h.templateService.Create(c.Request.Context(), userID, &req)
	if err != nil {
		Error(c, http.StatusInternalServerError, "创建模板失败", err.Error())
		return
	}

	c.JSON(http.StatusCreated, Response{
		Success: true,
		Code:    0,
		Message: "模板创建成功",
		Data:    template,
		TraceID: middleware.TraceIDFromGin(c),
	})
}

// Delete 删除模板（仅删除当前用户私有模板）。
func (h *TemplateHandler) Delete(c *gin.Context) {
	userID, ok := h.resolveUserID(c)
	if !ok {
		return
	}

	templateID := c.Param("id")
	if err := h.templateService.Delete(c.Request.Context(), templateID, userID); err != nil {
		switch {
		case errors.Is(err, service.ErrTemplateNotFound):
			ErrorWithCode(c, http.StatusNotFound, "TEMPLATE_NOT_FOUND", "模板不存在", nil)
		case errors.Is(err, service.ErrTemplateForbidden):
			ErrorWithCode(c, http.StatusForbidden, "TEMPLATE_FORBIDDEN", "无权删除该模板", nil)
		default:
			Error(c, http.StatusBadRequest, "删除模板失败", err.Error())
		}
		return
	}

	SuccessWithMessage(c, "模板删除成功", nil)
}

// Apply 复用模板（返回可直接填充生成页的参数）。
func (h *TemplateHandler) Apply(c *gin.Context) {
	userID, ok := h.resolveUserID(c)
	if !ok {
		return
	}

	templateID := c.Param("id")
	payload, err := h.templateService.Apply(c.Request.Context(), templateID, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTemplateNotFound):
			ErrorWithCode(c, http.StatusNotFound, "TEMPLATE_NOT_FOUND", "模板不存在", nil)
		case errors.Is(err, service.ErrTemplateForbidden):
			ErrorWithCode(c, http.StatusForbidden, "TEMPLATE_FORBIDDEN", "无权使用该模板", nil)
		default:
			Error(c, http.StatusBadRequest, "应用模板失败", err.Error())
		}
		return
	}

	Success(c, payload)
}
