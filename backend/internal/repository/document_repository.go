package repository

import (
	"lesson-plan/backend/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DocumentRepository 知识文档仓库（PostgreSQL）
type DocumentRepository struct {
	db *gorm.DB
}

// NewDocumentRepository 创建文档仓库
func NewDocumentRepository(db *gorm.DB) *DocumentRepository {
	return &DocumentRepository{db: db}
}

// CreateDocument 创建文档
func (r *DocumentRepository) CreateDocument(doc *model.KnowledgeDocument) error {
	return r.db.Create(doc).Error
}

// GetDocumentByID 根据ID和用户ID获取文档
func (r *DocumentRepository) GetDocumentByID(docID string, userID string) (*model.KnowledgeDocument, error) {
	var doc model.KnowledgeDocument
	err := r.db.
		Where("id = ? AND user_id = ?", docID, userID).
		First(&doc).Error
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

// ListDocuments 获取用户的文档列表
func (r *DocumentRepository) ListDocuments(userID string, page, pageSize int) ([]model.KnowledgeDocument, int64, error) {
	var docs []model.KnowledgeDocument
	var total int64

	offset := (page - 1) * pageSize

	// 获取总数
	if err := r.db.Model(&model.KnowledgeDocument{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := r.db.
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&docs).Error

	return docs, total, err
}

// UpdateDocumentStatus 更新文档状态
func (r *DocumentRepository) UpdateDocumentStatus(docID uuid.UUID, status string, entityCount, relCount int, errorMsg string) error {
	updates := map[string]interface{}{
		"status":         status,
		"error_msg":      errorMsg,
		"entity_count":   entityCount,
		"relation_count": relCount,
	}
	return r.db.
		Model(&model.KnowledgeDocument{}).
		Where("id = ?", docID).
		Updates(updates).Error
}

// DeleteDocument 删除文档
func (r *DocumentRepository) DeleteDocument(docID string, userID string) error {
	return r.db.
		Where("id = ? AND user_id = ?", docID, userID).
		Delete(&model.KnowledgeDocument{}).Error
}
