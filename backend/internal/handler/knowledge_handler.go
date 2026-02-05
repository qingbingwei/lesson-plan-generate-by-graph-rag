package handler

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"lesson-plan/backend/internal/middleware"
	"lesson-plan/backend/internal/model"
	"lesson-plan/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// KnowledgeHandler 知识库处理器
type KnowledgeHandler struct {
	documentService *service.DocumentService
}

// NewKnowledgeHandler 创建知识库处理器
func NewKnowledgeHandler(documentService *service.DocumentService) *KnowledgeHandler {
	return &KnowledgeHandler{
		documentService: documentService,
	}
}

// UploadDocument 上传知识文档
// POST /api/v1/knowledge/documents
func (h *KnowledgeHandler) UploadDocument(c *gin.Context) {
	// 获取用户ID
	userIDStr, ok := middleware.GetCurrentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, "未授权", nil)
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		Error(c, http.StatusUnauthorized, "无效的用户ID", nil)
		return
	}

	// 获取文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		Error(c, http.StatusBadRequest, "请选择要上传的文件", nil)
		return
	}
	defer file.Close()

	// 验证文件类型
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".txt" && ext != ".md" {
		Error(c, http.StatusBadRequest, "仅支持 .txt 和 .md 格式文件", nil)
		return
	}

	// 验证文件大小（最大 5MB）
	if header.Size > 5*1024*1024 {
		Error(c, http.StatusBadRequest, "文件大小不能超过 5MB", nil)
		return
	}

	// 读取文件内容
	content, err := io.ReadAll(file)
	if err != nil {
		Error(c, http.StatusInternalServerError, "读取文件失败", nil)
		return
	}

	// 获取可选参数
	title := c.PostForm("title")
	if title == "" {
		title = strings.TrimSuffix(header.Filename, ext)
	}
	subject := c.PostForm("subject")
	grade := c.PostForm("grade")

	// 创建文档记录
	doc := &model.KnowledgeDocument{
		UserID:   userID,
		Title:    title,
		FileName: header.Filename,
		FileType: strings.TrimPrefix(ext, "."),
		FileSize: header.Size,
		Content:  string(content),
		Subject:  subject,
		Grade:    grade,
		Status:   model.DocStatusPending,
	}

	// 保存文档并触发处理
	if err := h.documentService.CreateDocument(doc); err != nil {
		Error(c, http.StatusInternalServerError, fmt.Sprintf("保存文档失败: %v", err), nil)
		return
	}

	Success(c, gin.H{
		"id":       doc.ID,
		"title":    doc.Title,
		"fileName": doc.FileName,
		"status":   doc.Status,
		"message":  "文档已上传，正在后台处理中",
	})
}

// ListDocuments 获取用户的知识文档列表
// GET /api/v1/knowledge/documents
func (h *KnowledgeHandler) ListDocuments(c *gin.Context) {
	userIDStr, ok := middleware.GetCurrentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, "未授权", nil)
		return
	}

	docs, _, err := h.documentService.ListDocuments(userIDStr, 1, 100)
	if err != nil {
		Error(c, http.StatusInternalServerError, fmt.Sprintf("获取文档列表失败: %v", err), nil)
		return
	}

	Success(c, docs)
}

// GetDocument 获取文档详情
// GET /api/v1/knowledge/documents/:id
func (h *KnowledgeHandler) GetDocument(c *gin.Context) {
	userIDStr, ok := middleware.GetCurrentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, "未授权", nil)
		return
	}

	docID := c.Param("id")
	if _, err := uuid.Parse(docID); err != nil {
		Error(c, http.StatusBadRequest, "无效的文档ID", nil)
		return
	}

	doc, err := h.documentService.GetDocument(docID, userIDStr)
	if err != nil {
		Error(c, http.StatusNotFound, "文档不存在", nil)
		return
	}

	Success(c, doc)
}

// DeleteDocument 删除文档
// DELETE /api/v1/knowledge/documents/:id
func (h *KnowledgeHandler) DeleteDocument(c *gin.Context) {
	userIDStr, ok := middleware.GetCurrentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, "未授权", nil)
		return
	}

	docID := c.Param("id")
	if _, err := uuid.Parse(docID); err != nil {
		Error(c, http.StatusBadRequest, "无效的文档ID", nil)
		return
	}

	if err := h.documentService.DeleteDocument(docID, userIDStr); err != nil {
		Error(c, http.StatusInternalServerError, fmt.Sprintf("删除文档失败: %v", err), nil)
		return
	}

	Success(c, gin.H{"message": "文档已删除"})
}

// GetDocumentStatus 获取文档处理状态
// GET /api/v1/knowledge/documents/:id/status
func (h *KnowledgeHandler) GetDocumentStatus(c *gin.Context) {
	userIDStr, ok := middleware.GetCurrentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, "未授权", nil)
		return
	}

	docID := c.Param("id")
	if _, err := uuid.Parse(docID); err != nil {
		Error(c, http.StatusBadRequest, "无效的文档ID", nil)
		return
	}

	doc, err := h.documentService.GetDocumentStatus(docID, userIDStr)
	if err != nil {
		Error(c, http.StatusNotFound, "文档不存在", nil)
		return
	}

	Success(c, gin.H{
		"id":            doc.ID,
		"status":        doc.Status,
		"entityCount":   doc.EntityCount,
		"relationCount": doc.RelationCount,
		"errorMsg":      doc.ErrorMsg,
	})
}
