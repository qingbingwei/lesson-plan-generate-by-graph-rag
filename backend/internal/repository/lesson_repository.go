package repository

import (
	"context"

	"lesson-plan/backend/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LessonRepository 教案仓库接口
type LessonRepository interface {
	Create(ctx context.Context, lesson *model.Lesson) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Lesson, error)
	Update(ctx context.Context, lesson *model.Lesson) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter LessonFilter, page, pageSize int) ([]model.Lesson, int64, error)
	ListByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.Lesson, int64, error)
	IncrementViewCount(ctx context.Context, id uuid.UUID) error
	UpdateCounts(ctx context.Context, id uuid.UUID) error
	Search(ctx context.Context, query string, page, pageSize int) ([]model.Lesson, int64, error)
}

// LessonFilter 教案过滤器
type LessonFilter struct {
	Subject string
	Grade   string
	Status  string
	UserID  *uuid.UUID
	Keyword string
}

type lessonRepository struct {
	db *gorm.DB
}

// NewLessonRepository 创建教案仓库
func NewLessonRepository(db *gorm.DB) LessonRepository {
	return &lessonRepository{db: db}
}

func (r *lessonRepository) Create(ctx context.Context, lesson *model.Lesson) error {
	return r.db.WithContext(ctx).Create(lesson).Error
}

func (r *lessonRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Lesson, error) {
	var lesson model.Lesson
	err := r.db.WithContext(ctx).Preload("User").Where("id = ?", id).First(&lesson).Error
	if err != nil {
		return nil, err
	}
	return &lesson, nil
}

func (r *lessonRepository) Update(ctx context.Context, lesson *model.Lesson) error {
	return r.db.WithContext(ctx).Save(lesson).Error
}

func (r *lessonRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Lesson{}, "id = ?", id).Error
}

func (r *lessonRepository) List(ctx context.Context, filter LessonFilter, page, pageSize int) ([]model.Lesson, int64, error) {
	var lessons []model.Lesson
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Lesson{}).Preload("User")

	if filter.Subject != "" {
		db = db.Where("subject = ?", filter.Subject)
	}
	if filter.Grade != "" {
		db = db.Where("grade = ?", filter.Grade)
	}
	if filter.Status != "" {
		db = db.Where("status = ?", filter.Status)
	}
	if filter.UserID != nil {
		db = db.Where("user_id = ?", *filter.UserID)
	}
	if filter.Keyword != "" {
		db = db.Where("title ILIKE ? OR content ILIKE ?", "%"+filter.Keyword+"%", "%"+filter.Keyword+"%")
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&lessons).Error; err != nil {
		return nil, 0, err
	}

	return lessons, total, nil
}

func (r *lessonRepository) ListByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.Lesson, int64, error) {
	return r.List(ctx, LessonFilter{UserID: &userID}, page, pageSize)
}

func (r *lessonRepository) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&model.Lesson{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

func (r *lessonRepository) UpdateCounts(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Exec(`
		UPDATE lessons SET
			like_count = (SELECT COUNT(*) FROM lesson_likes WHERE lesson_id = lessons.id),
			favorite_count = (SELECT COUNT(*) FROM lesson_favorites WHERE lesson_id = lessons.id),
			comment_count = (SELECT COUNT(*) FROM lesson_comments WHERE lesson_id = lessons.id AND deleted_at IS NULL)
		WHERE id = ?
	`, id).Error
}

func (r *lessonRepository) Search(ctx context.Context, query string, page, pageSize int) ([]model.Lesson, int64, error) {
	return r.List(ctx, LessonFilter{Keyword: query, Status: model.LessonStatusPublished}, page, pageSize)
}

// CommentRepository 评论仓库接口
type CommentRepository interface {
	Create(ctx context.Context, comment *model.Comment) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Comment, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ListByLessonID(ctx context.Context, lessonID uuid.UUID, page, pageSize int) ([]model.Comment, int64, error)
}

type commentRepository struct {
	db *gorm.DB
}

// NewCommentRepository 创建评论仓库
func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &commentRepository{db: db}
}

func (r *commentRepository) Create(ctx context.Context, comment *model.Comment) error {
	return r.db.WithContext(ctx).Create(comment).Error
}

func (r *commentRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Comment, error) {
	var comment model.Comment
	err := r.db.WithContext(ctx).Preload("User").Where("id = ?", id).First(&comment).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *commentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Comment{}, "id = ?", id).Error
}

func (r *commentRepository) ListByLessonID(ctx context.Context, lessonID uuid.UUID, page, pageSize int) ([]model.Comment, int64, error) {
	var comments []model.Comment
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Comment{}).
		Preload("User").
		Preload("Replies.User").
		Where("lesson_id = ? AND parent_id IS NULL", lessonID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&comments).Error; err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

// FavoriteRepository 收藏仓库接口
type FavoriteRepository interface {
	Create(ctx context.Context, favorite *model.Favorite) error
	Delete(ctx context.Context, userID, lessonID uuid.UUID) error
	Exists(ctx context.Context, userID, lessonID uuid.UUID) (bool, error)
	ListByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.Favorite, int64, error)
	CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
}

type favoriteRepository struct {
	db *gorm.DB
}

// NewFavoriteRepository 创建收藏仓库
func NewFavoriteRepository(db *gorm.DB) FavoriteRepository {
	return &favoriteRepository{db: db}
}

func (r *favoriteRepository) Create(ctx context.Context, favorite *model.Favorite) error {
	return r.db.WithContext(ctx).Create(favorite).Error
}

func (r *favoriteRepository) Delete(ctx context.Context, userID, lessonID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND lesson_id = ?", userID, lessonID).
		Delete(&model.Favorite{}).Error
}

func (r *favoriteRepository) Exists(ctx context.Context, userID, lessonID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Favorite{}).
		Where("user_id = ? AND lesson_id = ?", userID, lessonID).Count(&count).Error
	return count > 0, err
}

func (r *favoriteRepository) ListByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.Favorite, int64, error) {
	var favorites []model.Favorite
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Favorite{}).
		Preload("Lesson.User").
		Where("user_id = ?", userID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&favorites).Error; err != nil {
		return nil, 0, err
	}

	return favorites, total, nil
}

func (r *favoriteRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Favorite{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

// LikeRepository 点赞仓库接口
type LikeRepository interface {
	Create(ctx context.Context, like *model.Like) error
	Delete(ctx context.Context, userID, lessonID uuid.UUID) error
	Exists(ctx context.Context, userID, lessonID uuid.UUID) (bool, error)
}

type likeRepository struct {
	db *gorm.DB
}

// NewLikeRepository 创建点赞仓库
func NewLikeRepository(db *gorm.DB) LikeRepository {
	return &likeRepository{db: db}
}

func (r *likeRepository) Create(ctx context.Context, like *model.Like) error {
	return r.db.WithContext(ctx).Create(like).Error
}

func (r *likeRepository) Delete(ctx context.Context, userID, lessonID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND lesson_id = ?", userID, lessonID).
		Delete(&model.Like{}).Error
}

func (r *likeRepository) Exists(ctx context.Context, userID, lessonID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Like{}).
		Where("user_id = ? AND lesson_id = ?", userID, lessonID).Count(&count).Error
	return count > 0, err
}
