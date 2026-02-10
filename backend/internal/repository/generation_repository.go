package repository

import (
	"context"

	"lesson-plan/backend/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GenerationRepository 生成记录仓库接口
type GenerationRepository interface {
	Create(ctx context.Context, generation *model.Generation) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Generation, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateResult(ctx context.Context, id uuid.UUID, result string, tokenCount int) error
	UpdateError(ctx context.Context, id uuid.UUID, errorMsg string) error
	ListByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.Generation, int64, error)
	GetStats(ctx context.Context, userID uuid.UUID) (*GenerationStats, error)
}

// GenerationStats 生成统计
type GenerationStats struct {
	TotalCount           int64   `json:"total_count"`
	CompletedCount       int64   `json:"completed_count"`
	FailedCount          int64   `json:"failed_count"`
	TotalTokens          int64   `json:"total_tokens"`
	AvgDurationMs        float64 `json:"avg_duration_ms"`
	ThisMonthGenerations int64   `json:"this_month_generations"`
	TotalLessons         int64   `json:"total_lessons"`
}

type generationRepository struct {
	db *gorm.DB
}

// NewGenerationRepository 创建生成记录仓库
func NewGenerationRepository(db *gorm.DB) GenerationRepository {
	return &generationRepository{db: db}
}

func (r *generationRepository) Create(ctx context.Context, generation *model.Generation) error {
	return r.db.WithContext(ctx).Create(generation).Error
}

func (r *generationRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Generation, error) {
	var generation model.Generation
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&generation).Error
	if err != nil {
		return nil, err
	}
	return &generation, nil
}

func (r *generationRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	return r.db.WithContext(ctx).Model(&model.Generation{}).Where("id = ?", id).
		Update("status", status).Error
}

func (r *generationRepository) UpdateResult(ctx context.Context, id uuid.UUID, result string, tokenCount int) error {
	return r.db.WithContext(ctx).Model(&model.Generation{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"result":      result,
			"token_count": tokenCount, "completed_at": gorm.Expr("NOW()"),
			"duration_ms": gorm.Expr("EXTRACT(EPOCH FROM (NOW() - created_at)) * 1000"),
			"status":      model.GenerationStatusCompleted,
		}).Error
}

func (r *generationRepository) UpdateError(ctx context.Context, id uuid.UUID, errorMsg string) error {
	return r.db.WithContext(ctx).Model(&model.Generation{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"error_msg":    errorMsg,
			"status":       model.GenerationStatusFailed,
			"completed_at": gorm.Expr("NOW()"),
			"duration_ms":  gorm.Expr("EXTRACT(EPOCH FROM (NOW() - created_at)) * 1000"),
		}).Error
}

func (r *generationRepository) ListByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.Generation, int64, error) {
	var generations []model.Generation
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Generation{}).Where("user_id = ?", userID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&generations).Error; err != nil {
		return nil, 0, err
	}

	return generations, total, nil
}

func (r *generationRepository) GetStats(ctx context.Context, userID uuid.UUID) (*GenerationStats, error) {
	var stats GenerationStats

	// 获取生成统计
	err := r.db.WithContext(ctx).Model(&model.Generation{}).
		Where("user_id = ?", userID).
		Select(`
			COUNT(*) as total_count,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_count,
			COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_count,
			COALESCE(SUM(token_count), 0) as total_tokens,
			COALESCE(AVG(duration_ms), 0) as avg_duration_ms,
			COUNT(CASE WHEN created_at >= date_trunc('month', CURRENT_DATE) THEN 1 END) as this_month_generations
		`).
		Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	// 获取教案总数
	var lessonCount int64
	r.db.WithContext(ctx).Model(&model.Lesson{}).Where("user_id = ?", userID).Count(&lessonCount)
	stats.TotalLessons = lessonCount

	return &stats, nil
}
