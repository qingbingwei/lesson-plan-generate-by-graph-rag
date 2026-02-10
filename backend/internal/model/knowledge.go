package model

import (
	"time"

	"github.com/google/uuid"
)

// 知识点类型
const (
	KnowledgeTypeConcept     = "concept"
	KnowledgeTypePrinciple   = "principle"
	KnowledgeTypeProcedure   = "procedure"
	KnowledgeTypeApplication = "application"
)

// Knowledge 知识点模型（Neo4j）
type Knowledge struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Subject     string    `json:"subject"`
	Grade       string    `json:"grade"`
	Description string    `json:"description"`
	Keywords    []string  `json:"keywords"`
	Embedding   []float64 `json:"embedding,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// KnowledgeRelation 知识点关系
type KnowledgeRelation struct {
	SourceID     string  `json:"source_id"`
	TargetID     string  `json:"target_id"`
	RelationType string  `json:"relation_type"`
	Weight       float64 `json:"weight"`
}

// 关系类型
const (
	RelationPrerequisite = "prerequisite"
	RelationRelated      = "related"
	RelationPartOf       = "part_of"
	RelationSimilar      = "similar"
)

// KnowledgeGraph 知识图谱
type KnowledgeGraph struct {
	Nodes []KnowledgeNode `json:"nodes"`
	Edges []KnowledgeEdge `json:"edges"`
}

// KnowledgeNode 知识图谱节点
type KnowledgeNode struct {
	ID         string  `json:"id"`
	Label      string  `json:"label"`
	Type       string  `json:"type"`
	Subject    string  `json:"subject"`
	Grade      string  `json:"grade"`
	Difficulty string  `json:"difficulty"`
	Importance float64 `json:"importance"`
}

// KnowledgeEdge 知识图谱边
type KnowledgeEdge struct {
	Source string  `json:"source"`
	Target string  `json:"target"`
	Type   string  `json:"type"`
	Weight float64 `json:"weight"`
}

// KnowledgeSearchResult 知识点搜索结果
type KnowledgeSearchResult struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Content        string  `json:"content"`
	RelevanceScore float64 `json:"relevance_score"`
	Source         string  `json:"source"`
}

// Generation 生成记录模型
type Generation struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID  `gorm:"type:uuid;index;not null" json:"user_id"`
	LessonID    *uuid.UUID `gorm:"type:uuid;index" json:"lesson_id,omitempty"`
	Prompt      string     `gorm:"type:text;not null" json:"prompt"`
	Parameters  string     `gorm:"type:jsonb" json:"parameters"`
	Result      string     `gorm:"type:text" json:"result"`
	Status      string     `gorm:"size:20;default:'pending';index" json:"status"`
	TokenCount  int        `gorm:"default:0" json:"token_count"`
	DurationMs  int64      `gorm:"default:0" json:"duration_ms"`
	ErrorMsg    string     `gorm:"type:text" json:"error_msg,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// TableName 表名
func (Generation) TableName() string {
	return "generations"
}

// 生成状态
const (
	GenerationStatusPending    = "pending"
	GenerationStatusProcessing = "processing"
	GenerationStatusCompleted  = "completed"
	GenerationStatusFailed     = "failed"
)

// GenerationRequest 生成请求
type GenerationRequest struct {
	Subject    string   `json:"subject" binding:"required"`
	Grade      string   `json:"grade" binding:"required"`
	Topic      string   `json:"topic" binding:"required"`
	Duration   int      `json:"duration"`
	Objectives []string `json:"objectives"`
	Keywords   []string `json:"keywords"`
	Style      string   `json:"style"`
	Difficulty string   `json:"difficulty"`
}

// GenerationResponse 生成响应
type GenerationResponse struct {
	ID              uuid.UUID `json:"id"`
	Status          string    `json:"status"`
	Title           string    `json:"title,omitempty"`
	Objectives      string    `json:"objectives,omitempty"`
	KeyPoints       string    `json:"key_points,omitempty"`
	DifficultPoints string    `json:"difficult_points,omitempty"`
	TeachingMethods string    `json:"teaching_methods,omitempty"`
	Content         string    `json:"content,omitempty"`
	Activities      string    `json:"activities,omitempty"`
	Assessment      string    `json:"assessment,omitempty"`
	Resources       string    `json:"resources,omitempty"`
	TokenCount      int       `json:"token_count"`
	DurationMs      int64     `json:"duration_ms"`
	ErrorMessage    string    `json:"error_message,omitempty"`
}

// ==================== 知识库文档模型 ====================

// KnowledgeDocument 知识文档模型
type KnowledgeDocument struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID        uuid.UUID `gorm:"type:uuid;not null;index;column:user_id" json:"userId"`
	Title         string    `gorm:"type:varchar(255);not null" json:"title"`
	FileName      string    `gorm:"type:varchar(255);not null;column:file_name" json:"fileName"`
	FileType      string    `gorm:"type:varchar(50);not null;column:file_type" json:"fileType"` // txt, md
	FileSize      int64     `gorm:"not null;column:file_size" json:"fileSize"`
	Content       string    `gorm:"type:text" json:"content"`
	Status        string    `gorm:"type:varchar(50);default:'pending'" json:"status"` // pending, processing, completed, failed
	ErrorMsg      string    `gorm:"type:text;column:error_msg" json:"errorMsg,omitempty"`
	EntityCount   int       `gorm:"default:0;column:entity_count" json:"entityCount"`
	RelationCount int       `gorm:"default:0;column:relation_count" json:"relationCount"`
	Subject       string    `gorm:"type:varchar(100)" json:"subject,omitempty"` // 可选：指定学科
	Grade         string    `gorm:"type:varchar(50)" json:"grade,omitempty"`    // 可选：指定年级
	CreatedAt     time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt     time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

// TableName 知识文档表名
func (KnowledgeDocument) TableName() string {
	return "knowledge_documents"
}

// 文档状态常量
const (
	DocStatusPending    = "pending"
	DocStatusProcessing = "processing"
	DocStatusCompleted  = "completed"
	DocStatusFailed     = "failed"
)
