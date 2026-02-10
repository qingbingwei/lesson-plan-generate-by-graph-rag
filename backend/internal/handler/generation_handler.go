package handler

import (
	"net/http"
	"strconv"
	"strings"

	"lesson-plan/backend/internal/middleware"
	"lesson-plan/backend/internal/model"
	"lesson-plan/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GenerationHandler 生成处理器
type GenerationHandler struct {
	generationService service.GenerationService
	knowledgeService  service.KnowledgeService
}

// NewGenerationHandler 创建生成处理器
func NewGenerationHandler(
	generationService service.GenerationService,
	knowledgeService service.KnowledgeService,
) *GenerationHandler {
	return &GenerationHandler{
		generationService: generationService,
		knowledgeService:  knowledgeService,
	}
}

// Generate 生成教案
func (h *GenerationHandler) Generate(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, "未认证", nil)
		return
	}

	var req model.GenerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	userUUID, _ := uuid.Parse(userID)
	keyOverride := service.NewAPIKeyOverride(
		c.GetHeader(service.HeaderGenerationAPIKey),
		c.GetHeader(service.HeaderEmbeddingAPIKey),
	)
	resp, err := h.generationService.Generate(c.Request.Context(), userUUID, &req, keyOverride)
	if err != nil {
		Error(c, http.StatusInternalServerError, "生成失败", err.Error())
		return
	}

	Success(c, resp)
}

// GetGeneration 获取生成记录
func (h *GenerationHandler) GetGeneration(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, "无效的ID", nil)
		return
	}

	generation, err := h.generationService.GetByID(c.Request.Context(), id)
	if err != nil {
		Error(c, http.StatusNotFound, "记录不存在", nil)
		return
	}

	Success(c, generation)
}

// ListGenerations 生成历史列表
func (h *GenerationHandler) ListGenerations(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, "未认证", nil)
		return
	}

	page, pageSize := GetPagination(c)
	userUUID, _ := uuid.Parse(userID)

	generations, total, err := h.generationService.ListByUser(c.Request.Context(), userUUID, page, pageSize)
	if err != nil {
		Error(c, http.StatusInternalServerError, "获取列表失败", err.Error())
		return
	}

	Paginated(c, generations, total, page, pageSize)
}

// GetStats 获取生成统计
func (h *GenerationHandler) GetStats(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, "未认证", nil)
		return
	}

	userUUID, _ := uuid.Parse(userID)
	stats, err := h.generationService.GetStats(c.Request.Context(), userUUID)
	if err != nil {
		Error(c, http.StatusInternalServerError, "获取统计失败", err.Error())
		return
	}

	Success(c, stats)
}

// SearchKnowledge 知识搜索
func (h *GenerationHandler) SearchKnowledge(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		Error(c, http.StatusBadRequest, "请输入搜索关键词", nil)
		return
	}

	limit := 10
	keyOverride := service.NewAPIKeyOverride(
		c.GetHeader(service.HeaderGenerationAPIKey),
		c.GetHeader(service.HeaderEmbeddingAPIKey),
	)
	ctx := service.WithAPIKeyOverride(c.Request.Context(), keyOverride)
	results, err := h.knowledgeService.Search(ctx, query, limit)
	if err != nil {
		Error(c, http.StatusInternalServerError, "搜索失败", err.Error())
		return
	}

	Success(c, results)
}

// GetKnowledgeGraph 获取知识图谱
func (h *GenerationHandler) GetKnowledgeGraph(c *gin.Context) {
	subject := c.Query("subject")
	grade := c.Query("grade")
	topic := strings.TrimSpace(c.Query("topic"))
	scope := strings.TrimSpace(c.Query("scope"))
	limit := 50
	if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 && l <= 500 {
		limit = l
	}

	// 获取当前用户ID，只展示用户自己的知识图谱
	userIdStr, _ := middleware.GetCurrentUserID(c)

	graph, err := h.knowledgeService.GetGraph(c.Request.Context(), subject, grade, topic, scope, userIdStr, limit)
	if err != nil {
		Error(c, http.StatusInternalServerError, "获取图谱失败", err.Error())
		return
	}

	Success(c, graph)
}
