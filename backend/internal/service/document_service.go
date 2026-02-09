package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"lesson-plan/backend/internal/config"
	"lesson-plan/backend/internal/model"
	"lesson-plan/backend/internal/repository"
	"lesson-plan/backend/pkg/logger"
)

// DocumentService 文档服务
type DocumentService struct {
	documentRepo repository.DocumentRepository
	agentConfig  *config.AgentConfig
	httpClient   *http.Client
}

// NewDocumentService 创建文档服务
func NewDocumentService(documentRepo repository.DocumentRepository, agentConfig *config.AgentConfig) *DocumentService {
	return &DocumentService{
		documentRepo: documentRepo,
		agentConfig:  agentConfig,
		httpClient: &http.Client{
			Timeout: 10 * time.Minute, // 长时间处理
		},
	}
}

// CreateDocument 创建文档记录
func (s *DocumentService) CreateDocument(doc *model.KnowledgeDocument) error {
	err := s.documentRepo.CreateDocument(doc)
	if err != nil {
		return err
	}

	// 异步处理文档（带 recover 和超时保护）
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error(fmt.Sprintf("panic in processDocument for doc %s: %v", doc.ID, r))
				s.documentRepo.UpdateDocumentStatus(doc.ID, model.DocStatusFailed, 0, 0, "内部错误: 处理过程异常")
			}
		}()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		s.processDocument(ctx, doc)
	}()

	return nil
}

// processDocument 处理文档，调用Agent构建知识图谱
func (s *DocumentService) processDocument(ctx context.Context, doc *model.KnowledgeDocument) {
	// 更新状态为处理中
	if err := s.documentRepo.UpdateDocumentStatus(doc.ID, model.DocStatusProcessing, 0, 0, ""); err != nil {
		logger.Error("Failed to update document status: " + err.Error())
		return
	}

	// 构建请求
	reqBody := map[string]interface{}{
		"documentId": doc.ID,
		"userId":     doc.UserID,
		"content":    doc.Content,
		"title":      doc.Title,
		"subject":    doc.Subject,
		"grade":      doc.Grade,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		s.documentRepo.UpdateDocumentStatus(doc.ID, model.DocStatusFailed, 0, 0, "JSON编码错误")
		return
	}

	// 调用Agent API（带 context 超时控制）
	agentURL := fmt.Sprintf("%s/api/build-graph", s.agentConfig.URL)
	req, err := http.NewRequestWithContext(ctx, "POST", agentURL, bytes.NewBuffer(jsonData))
	if err != nil {
		s.documentRepo.UpdateDocumentStatus(doc.ID, model.DocStatusFailed, 0, 0, "请求创建失败")
		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		logger.Error("Failed to call agent: " + err.Error())
		s.documentRepo.UpdateDocumentStatus(doc.ID, model.DocStatusFailed, 0, 0, "Agent服务调用失败: "+err.Error())
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		logger.Error("Agent returned error: " + string(body))
		s.documentRepo.UpdateDocumentStatus(doc.ID, model.DocStatusFailed, 0, 0, "Agent处理失败")
		return
	}

	// 解析响应
	var result struct {
		Success     bool   `json:"success"`
		Message     string `json:"message"`
		EntityCount int    `json:"entityCount"`
		RelCount    int    `json:"relationCount"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		s.documentRepo.UpdateDocumentStatus(doc.ID, model.DocStatusFailed, 0, 0, "响应解析失败")
		return
	}

	if !result.Success {
		s.documentRepo.UpdateDocumentStatus(doc.ID, model.DocStatusFailed, 0, 0, result.Message)
		return
	}

	// 更新为完成状态
	s.documentRepo.UpdateDocumentStatus(doc.ID, model.DocStatusCompleted, result.EntityCount, result.RelCount, "")
	logger.Info(fmt.Sprintf("Document %s processed: %d entities, %d relations", doc.ID, result.EntityCount, result.RelCount))
}

// GetDocument 获取文档
func (s *DocumentService) GetDocument(id string, userID string) (*model.KnowledgeDocument, error) {
	return s.documentRepo.GetDocumentByID(id, userID)
}

// ListDocuments 获取文档列表
func (s *DocumentService) ListDocuments(userID string, page, pageSize int) ([]model.KnowledgeDocument, int64, error) {
	return s.documentRepo.ListDocuments(userID, page, pageSize)
}

// DeleteDocument 删除文档
func (s *DocumentService) DeleteDocument(id string, userID string) error {
	// 先获取文档确认权限
	doc, err := s.documentRepo.GetDocumentByID(id, userID)
	if err != nil {
		return err
	}
	if doc == nil {
		return fmt.Errorf("document not found")
	}

	// 调用Agent删除Neo4j中的节点（带 recover 和超时保护）
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error(fmt.Sprintf("panic in deleteDocumentNodes for doc %s: %v", id, r))
			}
		}()
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		s.deleteDocumentNodes(ctx, id)
	}()

	// 删除数据库记录
	return s.documentRepo.DeleteDocument(id, userID)
}

// deleteDocumentNodes 删除Neo4j中的文档节点
func (s *DocumentService) deleteDocumentNodes(ctx context.Context, documentID string) {
	reqBody := map[string]string{
		"documentId": documentID,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return
	}

	agentURL := fmt.Sprintf("%s/api/delete-document-nodes", s.agentConfig.URL)
	req, err := http.NewRequestWithContext(ctx, "POST", agentURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		logger.Error("Failed to delete document nodes: " + err.Error())
		return
	}
	defer resp.Body.Close()
}

// GetDocumentStatus 获取文档状态
func (s *DocumentService) GetDocumentStatus(id string, userID string) (*model.KnowledgeDocument, error) {
	return s.documentRepo.GetDocumentByID(id, userID)
}
