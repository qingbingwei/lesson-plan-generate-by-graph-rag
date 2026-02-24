package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"lesson-plan/backend/internal/middleware"
	"lesson-plan/backend/internal/model"
	"lesson-plan/backend/internal/repository"
	"lesson-plan/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// LessonHandler 教案处理器
type LessonHandler struct {
	lessonService   service.LessonService
	favoriteService service.FavoriteService
	likeService     service.LikeService
	commentService  service.CommentService
}

type exportLayoutOption struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// NewLessonHandler 创建教案处理器
func NewLessonHandler(
	lessonService service.LessonService,
	favoriteService service.FavoriteService,
	likeService service.LikeService,
	commentService service.CommentService,
) *LessonHandler {
	return &LessonHandler{
		lessonService:   lessonService,
		favoriteService: favoriteService,
		likeService:     likeService,
		commentService:  commentService,
	}
}

func availableExportLayouts() []exportLayoutOption {
	return []exportLayoutOption{
		{
			ID:          "standard",
			Name:        "标准模板",
			Description: "通用教学文档结构，适合日常备课提交。",
		},
		{
			ID:          "compact",
			Name:        "紧凑模板",
			Description: "更简洁的版面，强调关键结论与执行步骤。",
		},
		{
			ID:          "research",
			Name:        "教研模板",
			Description: "适合教研分享，补充设计说明与反思提示。",
		},
	}
}

func isValidExportLayout(layout string) bool {
	for _, item := range availableExportLayouts() {
		if item.ID == layout {
			return true
		}
	}
	return false
}

// List 教案列表
func (h *LessonHandler) List(c *gin.Context) {
	page, pageSize := GetPagination(c)

	filter := repository.LessonFilter{
		Subject: c.Query("subject"),
		Grade:   c.Query("grade"),
		Status:  c.Query("status"),
		Keyword: c.Query("keyword"),
	}

	// 只显示当前用户的教案
	if userID, ok := middleware.GetCurrentUserID(c); ok {
		uid, _ := uuid.Parse(userID)
		filter.UserID = &uid
	}

	lessons, total, err := h.lessonService.List(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		Error(c, http.StatusInternalServerError, "获取列表失败", err.Error())
		return
	}

	Paginated(c, lessons, total, page, pageSize)
}

// GetByID 获取教案详情
func (h *LessonHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的ID", nil)
		return
	}

	var currentUserID *uuid.UUID
	if userID, ok := middleware.GetCurrentUserID(c); ok {
		uid, _ := uuid.Parse(userID)
		currentUserID = &uid
	}

	lesson, err := h.lessonService.GetByID(c.Request.Context(), id, currentUserID)
	if err != nil {
		Error(c, http.StatusNotFound, "教案不存在", nil)
		return
	}

	Success(c, lesson)
}

// Create 创建教案
func (h *LessonHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, "未认证", nil)
		return
	}

	var req service.CreateLessonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	userUUID, _ := uuid.Parse(userID)
	lesson, err := h.lessonService.Create(c.Request.Context(), userUUID, &req)
	if err != nil {
		Error(c, http.StatusInternalServerError, "创建失败", err.Error())
		return
	}

	c.JSON(http.StatusCreated, Response{
		Code:    0,
		Message: "创建成功",
		Data:    lesson,
	})
}

// Update 更新教案
func (h *LessonHandler) Update(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, "未认证", nil)
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的ID", nil)
		return
	}

	var req service.UpdateLessonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	userUUID, _ := uuid.Parse(userID)
	lesson, err := h.lessonService.Update(c.Request.Context(), id, userUUID, &req)
	if err != nil {
		Error(c, http.StatusInternalServerError, "更新失败", err.Error())
		return
	}

	Success(c, lesson)
}

// Delete 删除教案
func (h *LessonHandler) Delete(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, "未认证", nil)
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的ID", nil)
		return
	}

	userUUID, _ := uuid.Parse(userID)
	if err := h.lessonService.Delete(c.Request.Context(), id, userUUID); err != nil {
		Error(c, http.StatusInternalServerError, "删除失败", err.Error())
		return
	}

	SuccessWithMessage(c, "删除成功", nil)
}

// Publish 发布教案
func (h *LessonHandler) Publish(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, "未认证", nil)
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的ID", nil)
		return
	}

	userUUID, _ := uuid.Parse(userID)
	if err := h.lessonService.Publish(c.Request.Context(), id, userUUID); err != nil {
		Error(c, http.StatusInternalServerError, "发布失败", err.Error())
		return
	}

	SuccessWithMessage(c, "发布成功", nil)
}

// MyLessons 我的教案
func (h *LessonHandler) MyLessons(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, "未认证", nil)
		return
	}

	page, pageSize := GetPagination(c)
	userUUID, _ := uuid.Parse(userID)

	lessons, total, err := h.lessonService.ListByUser(c.Request.Context(), userUUID, page, pageSize)
	if err != nil {
		Error(c, http.StatusInternalServerError, "获取列表失败", err.Error())
		return
	}

	Paginated(c, lessons, total, page, pageSize)
}

// AddFavorite 添加收藏
func (h *LessonHandler) AddFavorite(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, "未认证", nil)
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的ID", nil)
		return
	}

	userUUID, _ := uuid.Parse(userID)
	if err := h.favoriteService.Add(c.Request.Context(), userUUID, id); err != nil {
		Error(c, http.StatusInternalServerError, "收藏失败", err.Error())
		return
	}

	SuccessWithMessage(c, "收藏成功", nil)
}

// RemoveFavorite 取消收藏
func (h *LessonHandler) RemoveFavorite(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, "未认证", nil)
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的ID", nil)
		return
	}

	userUUID, _ := uuid.Parse(userID)
	if err := h.favoriteService.Remove(c.Request.Context(), userUUID, id); err != nil {
		Error(c, http.StatusInternalServerError, "取消收藏失败", err.Error())
		return
	}

	SuccessWithMessage(c, "取消收藏成功", nil)
}

// MyFavorites 我的收藏
func (h *LessonHandler) MyFavorites(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, "未认证", nil)
		return
	}

	page, pageSize := GetPagination(c)
	userUUID, _ := uuid.Parse(userID)

	lessons, total, err := h.favoriteService.List(c.Request.Context(), userUUID, page, pageSize)
	if err != nil {
		Error(c, http.StatusInternalServerError, "获取列表失败", err.Error())
		return
	}

	Paginated(c, lessons, total, page, pageSize)
}

// Like 点赞
func (h *LessonHandler) Like(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, "未认证", nil)
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的ID", nil)
		return
	}

	userUUID, _ := uuid.Parse(userID)
	if err := h.likeService.Like(c.Request.Context(), userUUID, id); err != nil {
		Error(c, http.StatusInternalServerError, "点赞失败", err.Error())
		return
	}

	SuccessWithMessage(c, "点赞成功", nil)
}

// Unlike 取消点赞
func (h *LessonHandler) Unlike(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, "未认证", nil)
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的ID", nil)
		return
	}

	userUUID, _ := uuid.Parse(userID)
	if err := h.likeService.Unlike(c.Request.Context(), userUUID, id); err != nil {
		Error(c, http.StatusInternalServerError, "取消点赞失败", err.Error())
		return
	}

	SuccessWithMessage(c, "取消点赞成功", nil)
}

// ListComments 评论列表
func (h *LessonHandler) ListComments(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的ID", nil)
		return
	}

	page, pageSize := GetPagination(c)

	comments, total, err := h.commentService.List(c.Request.Context(), id, page, pageSize)
	if err != nil {
		Error(c, http.StatusInternalServerError, "获取评论失败", err.Error())
		return
	}

	Paginated(c, comments, total, page, pageSize)
}

// CreateComment 创建评论
func (h *LessonHandler) CreateComment(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, "未认证", nil)
		return
	}

	lessonID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的教案ID", nil)
		return
	}

	var req struct {
		Content  string  `json:"content" binding:"required,max=1000"`
		ParentID *string `json:"parent_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	var parentID *uuid.UUID
	if req.ParentID != nil && *req.ParentID != "" {
		pid, err := uuid.Parse(*req.ParentID)
		if err == nil {
			parentID = &pid
		}
	}

	userUUID, _ := uuid.Parse(userID)
	comment, err := h.commentService.Create(c.Request.Context(), userUUID, lessonID, req.Content, parentID)
	if err != nil {
		Error(c, http.StatusInternalServerError, "创建评论失败", err.Error())
		return
	}

	c.JSON(http.StatusCreated, Response{
		Code:    0,
		Message: "评论成功",
		Data:    comment,
	})
}

// DeleteComment 删除评论
func (h *LessonHandler) DeleteComment(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, "未认证", nil)
		return
	}

	commentID, err := uuid.Parse(c.Param("commentId"))
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的评论ID", nil)
		return
	}

	userUUID, _ := uuid.Parse(userID)
	if err := h.commentService.Delete(c.Request.Context(), commentID, userUUID); err != nil {
		Error(c, http.StatusInternalServerError, "删除评论失败", err.Error())
		return
	}

	SuccessWithMessage(c, "删除成功", nil)
}

// Search 搜索教案
func (h *LessonHandler) Search(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		Error(c, http.StatusBadRequest, "请输入搜索关键词", nil)
		return
	}

	page, pageSize := GetPagination(c)

	lessons, total, err := h.lessonService.Search(c.Request.Context(), query, page, pageSize)
	if err != nil {
		Error(c, http.StatusInternalServerError, "搜索失败", err.Error())
		return
	}

	Paginated(c, lessons, total, page, pageSize)
}

// ListVersions 获取教案版本历史
func (h *LessonHandler) ListVersions(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, "未认证", nil)
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的ID", nil)
		return
	}

	userUUID, _ := uuid.Parse(userID)
	versions, err := h.lessonService.ListVersions(c.Request.Context(), id, userUUID)
	if err != nil {
		Error(c, http.StatusInternalServerError, "获取版本列表失败", err.Error())
		return
	}

	Success(c, versions)
}

// GetVersion 获取指定版本
func (h *LessonHandler) GetVersion(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, "未认证", nil)
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的ID", nil)
		return
	}

	var version int
	if _, err := fmt.Sscanf(c.Param("version"), "%d", &version); err != nil {
		Error(c, http.StatusBadRequest, "无效的版本号", nil)
		return
	}

	userUUID, _ := uuid.Parse(userID)
	v, err := h.lessonService.GetVersion(c.Request.Context(), id, version, userUUID)
	if err != nil {
		Error(c, http.StatusNotFound, "版本不存在", err.Error())
		return
	}

	Success(c, v)
}

// RollbackToVersion 回滚到指定版本
func (h *LessonHandler) RollbackToVersion(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, "未认证", nil)
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的ID", nil)
		return
	}

	var version int
	if _, err := fmt.Sscanf(c.Param("version"), "%d", &version); err != nil {
		Error(c, http.StatusBadRequest, "无效的版本号", nil)
		return
	}

	userUUID, _ := uuid.Parse(userID)
	lesson, err := h.lessonService.RollbackToVersion(c.Request.Context(), id, version, userUUID)
	if err != nil {
		Error(c, http.StatusInternalServerError, "回滚失败", err.Error())
		return
	}

	Success(c, lesson)
}

// QualityReview 教案质量评分与自动审查。
func (h *LessonHandler) QualityReview(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, "未认证", nil)
		return
	}

	lessonID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的ID", nil)
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		Error(c, http.StatusUnauthorized, "无效的用户标识", nil)
		return
	}

	report, err := h.lessonService.ReviewQuality(c.Request.Context(), lessonID, userUUID)
	if err != nil {
		Error(c, http.StatusInternalServerError, "质量审查失败", err.Error())
		return
	}

	Success(c, report)
}

// DiffVersions 比较两个版本的差异。
func (h *LessonHandler) DiffVersions(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, "未认证", nil)
		return
	}

	lessonID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的ID", nil)
		return
	}

	fromVersion := strings.TrimSpace(c.Query("from"))
	toVersion := strings.TrimSpace(c.Query("to"))
	if fromVersion == "" {
		Error(c, http.StatusBadRequest, "缺少 from 版本号", nil)
		return
	}
	if toVersion == "" {
		toVersion = "current"
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		Error(c, http.StatusUnauthorized, "无效的用户标识", nil)
		return
	}

	diff, err := h.lessonService.CompareVersions(c.Request.Context(), lessonID, userUUID, fromVersion, toVersion)
	if err != nil {
		Error(c, http.StatusBadRequest, "版本对比失败", err.Error())
		return
	}

	Success(c, diff)
}

// ExportLayouts 返回可用导出模板。
func (h *LessonHandler) ExportLayouts(c *gin.Context) {
	Success(c, availableExportLayouts())
}

// Export 导出教案
func (h *LessonHandler) Export(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的ID", nil)
		return
	}

	format := c.Query("format")
	if format == "" {
		format = "md"
	}
	layout := strings.TrimSpace(c.Query("layout"))
	if layout == "" {
		layout = "standard"
	}

	// 验证格式
	validFormats := map[string]bool{"md": true, "pdf": true, "docx": true}
	if !validFormats[format] {
		Error(c, http.StatusBadRequest, "不支持的格式，请使用 md、pdf 或 docx", nil)
		return
	}
	if !isValidExportLayout(layout) {
		Error(c, http.StatusBadRequest, "不支持的模板，请使用 standard、compact 或 research", nil)
		return
	}

	var currentUserID *uuid.UUID
	if userID, ok := middleware.GetCurrentUserID(c); ok {
		uid, _ := uuid.Parse(userID)
		currentUserID = &uid
	}

	lesson, err := h.lessonService.GetByID(c.Request.Context(), id, currentUserID)
	if err != nil {
		Error(c, http.StatusNotFound, "教案不存在", nil)
		return
	}

	// 生成 Markdown 内容（模板化版式）
	mdContent := h.generateMarkdown(lesson, layout)

	// 如果是 md 格式，直接返回
	if format == "md" {
		c.Header("Content-Type", "text/markdown; charset=utf-8")
		// 使用 RFC 5987 编码处理中文文件名
		encodedFilename := url.PathEscape(lesson.Title + ".md")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", encodedFilename))
		c.String(http.StatusOK, mdContent)
		return
	}

	// 使用 pandoc 转换
	outputFile, err := h.convertWithPandoc(mdContent, lesson.Title, format, layout)
	if err != nil {
		Error(c, http.StatusInternalServerError, "转换失败: "+err.Error(), nil)
		return
	}
	defer os.Remove(outputFile)

	// 设置响应头
	var contentType string
	var ext string
	switch format {
	case "pdf":
		contentType = "application/pdf"
		ext = "pdf"
	case "docx":
		contentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
		ext = "docx"
	}

	c.Header("Content-Type", contentType)
	// 使用 RFC 5987 编码处理中文文件名
	encodedFilename := url.PathEscape(lesson.Title + "." + ext)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", encodedFilename))
	c.File(outputFile)
}

func extractLessonText(raw string) string {
	if raw == "" || raw == "{}" {
		return ""
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &data); err == nil {
		if text, ok := data["text"].(string); ok {
			return strings.TrimSpace(text)
		}
		return formatJSONContent(data)
	}
	return strings.TrimSpace(raw)
}

func sanitizeMarkdown(raw string) string {
	value := extractLessonText(raw)
	value = strings.ReplaceAll(value, "\n---\n", "\n\n---\n\n")

	lines := strings.Split(value, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) == "---" {
			lines[i] = "* * *"
		}
	}
	return strings.Join(lines, "\n")
}

// generateMarkdown 生成模板化 Markdown 内容。
func (h *LessonHandler) generateMarkdown(lesson *model.LessonDetail, layout string) string {
	objectives := sanitizeMarkdown(lesson.Objectives)
	content := sanitizeMarkdown(lesson.Content)
	activities := sanitizeMarkdown(lesson.Activities)
	assessment := sanitizeMarkdown(lesson.Assessment)
	resources := sanitizeMarkdown(lesson.Resources)

	var sb strings.Builder

	switch layout {
	case "compact":
		sb.WriteString(fmt.Sprintf("# %s\n\n", lesson.Title))
		sb.WriteString("| 学科 | 年级 | 课时 |\n| --- | --- | --- |\n")
		sb.WriteString(fmt.Sprintf("| %s | %s | %d 分钟 |\n\n", lesson.Subject, lesson.Grade, lesson.Duration))

		if objectives != "" {
			sb.WriteString("## 目标速览\n\n")
			sb.WriteString(objectives + "\n\n")
		}
		if content != "" {
			sb.WriteString("## 内容主线\n\n")
			sb.WriteString(content + "\n\n")
		}
		if activities != "" {
			sb.WriteString("## 课堂执行步骤\n\n")
			sb.WriteString(activities + "\n\n")
		}
		if assessment != "" {
			sb.WriteString("## 达成检测\n\n")
			sb.WriteString(assessment + "\n\n")
		}
		if resources != "" {
			sb.WriteString("## 资源清单\n\n")
			sb.WriteString(resources + "\n\n")
		}

	case "research":
		sb.WriteString(fmt.Sprintf("# %s（教研版）\n\n", lesson.Title))
		sb.WriteString(fmt.Sprintf("**学科：** %s  \n", lesson.Subject))
		sb.WriteString(fmt.Sprintf("**年级：** %s  \n", lesson.Grade))
		sb.WriteString(fmt.Sprintf("**课时：** %d 分钟  \n", lesson.Duration))
		sb.WriteString(fmt.Sprintf("**版本：** v%d  \n\n", lesson.Version))

		sb.WriteString("## 一、教学目标\n\n")
		sb.WriteString(orFallback(objectives, "待补充目标说明。"))
		sb.WriteString("\n\n")

		sb.WriteString("## 二、教学内容与重难点\n\n")
		sb.WriteString(orFallback(content, "待补充内容说明。"))
		sb.WriteString("\n\n")

		sb.WriteString("## 三、教学活动设计\n\n")
		sb.WriteString(orFallback(activities, "待补充活动设计。"))
		sb.WriteString("\n\n")

		sb.WriteString("## 四、评价方案\n\n")
		sb.WriteString(orFallback(assessment, "待补充评价方案。"))
		sb.WriteString("\n\n")

		sb.WriteString("## 五、资源与保障\n\n")
		sb.WriteString(orFallback(resources, "待补充资源配置。"))
		sb.WriteString("\n\n")

		sb.WriteString("## 六、教学设计说明（自动生成）\n\n")
		sb.WriteString("- 建议围绕“目标-活动-评价”闭环进行同伴评审。\n")
		sb.WriteString("- 建议在课堂后补充执行反馈与改进建议。\n\n")

	default:
		sb.WriteString(fmt.Sprintf("# %s\n\n", lesson.Title))
		sb.WriteString(fmt.Sprintf("**学科：** %s  \n", lesson.Subject))
		sb.WriteString(fmt.Sprintf("**年级：** %s  \n", lesson.Grade))
		sb.WriteString(fmt.Sprintf("**课时：** %d分钟  \n\n", lesson.Duration))

		if objectives != "" {
			sb.WriteString("## 教学目标\n\n")
			sb.WriteString(objectives + "\n\n")
		}
		if content != "" {
			sb.WriteString("## 教学内容\n\n")
			sb.WriteString(content + "\n\n")
		}
		if activities != "" {
			sb.WriteString("## 教学活动\n\n")
			sb.WriteString(activities + "\n\n")
		}
		if assessment != "" {
			sb.WriteString("## 教学评价\n\n")
			sb.WriteString(assessment + "\n\n")
		}
		if resources != "" {
			sb.WriteString("## 教学资源\n\n")
			sb.WriteString(resources + "\n\n")
		}
	}

	return sb.String()
}

func orFallback(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

// convertWithPandoc 使用 pandoc 转换文件
func (h *LessonHandler) convertWithPandoc(mdContent, title, format, layout string) (string, error) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "lesson-export-")
	if err != nil {
		return "", fmt.Errorf("创建临时目录失败: %v", err)
	}

	// 写入 Markdown 文件
	mdFile := filepath.Join(tmpDir, "lesson.md")
	if err := os.WriteFile(mdFile, []byte(mdContent), 0644); err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("写入文件失败: %v", err)
	}

	// 确定输出文件
	var outputFile string
	var args []string

	// 基础参数：禁用 YAML 元数据解析
	baseArgs := []string{
		"--from", "markdown-yaml_metadata_block",
		mdFile,
	}

	switch format {
	case "pdf":
		outputFile = filepath.Join(tmpDir, title+".pdf")
		// 使用 weasyprint 以支持中文
		args = append(baseArgs,
			"-o", outputFile,
			"--pdf-engine=weasyprint",
		)
		cssFile := filepath.Join("templates", "export", layout+".css")
		if _, err := os.Stat(cssFile); err == nil {
			args = append(args, "--css", cssFile)
		}
	case "docx":
		outputFile = filepath.Join(tmpDir, title+".docx")
		args = append(baseArgs,
			"-o", outputFile,
		)
	}

	// 执行 pandoc
	cmd := exec.Command("pandoc", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("pandoc 转换失败: %v, 输出: %s", err, string(output))
	}

	// 删除临时 md 文件，保留输出文件
	os.Remove(mdFile)

	return outputFile, nil
}

// formatJSONContent 将 JSON 对象格式化为可读文本
func formatJSONContent(data map[string]interface{}) string {
	var sb strings.Builder
	for key, value := range data {
		switch v := value.(type) {
		case string:
			sb.WriteString(fmt.Sprintf("**%s：** %s\n\n", key, v))
		case []interface{}:
			sb.WriteString(fmt.Sprintf("**%s：**\n", key))
			for _, item := range v {
				sb.WriteString(fmt.Sprintf("- %v\n", item))
			}
			sb.WriteString("\n")
		case map[string]interface{}:
			sb.WriteString(fmt.Sprintf("**%s：**\n", key))
			sb.WriteString(formatJSONContent(v))
		default:
			sb.WriteString(fmt.Sprintf("**%s：** %v\n\n", key, v))
		}
	}
	return sb.String()
}
