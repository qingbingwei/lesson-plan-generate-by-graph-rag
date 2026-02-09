package repository

import (
	"context"

	"lesson-plan/backend/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// VersionRepository 版本仓库接口
type VersionRepository interface {
	Create(ctx context.Context, version *model.LessonVersion) error
	ListByLessonID(ctx context.Context, lessonID uuid.UUID) ([]model.LessonVersion, error)
	GetByVersion(ctx context.Context, lessonID uuid.UUID, version int) (*model.LessonVersion, error)
}

type versionRepository struct {
	db *gorm.DB
}

// NewVersionRepository 创建版本仓库
func NewVersionRepository(db *gorm.DB) VersionRepository {
	return &versionRepository{db: db}
}

func (r *versionRepository) Create(ctx context.Context, version *model.LessonVersion) error {
	return r.db.WithContext(ctx).Create(version).Error
}

func (r *versionRepository) ListByLessonID(ctx context.Context, lessonID uuid.UUID) ([]model.LessonVersion, error) {
	var versions []model.LessonVersion
	err := r.db.WithContext(ctx).
		Where("lesson_id = ?", lessonID).
		Order("version_number DESC").
		Find(&versions).Error
	return versions, err
}

func (r *versionRepository) GetByVersion(ctx context.Context, lessonID uuid.UUID, version int) (*model.LessonVersion, error) {
	var v model.LessonVersion
	err := r.db.WithContext(ctx).
		Where("lesson_id = ? AND version_number = ?", lessonID, version).
		First(&v).Error
	if err != nil {
		return nil, err
	}
	return &v, nil
}
