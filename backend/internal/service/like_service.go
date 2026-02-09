package service

import (
	"context"

	"lesson-plan/backend/internal/model"
	"lesson-plan/backend/internal/repository"

	"github.com/google/uuid"
)

// LikeService 点赞服务接口
type LikeService interface {
	Like(ctx context.Context, userID, lessonID uuid.UUID) error
	Unlike(ctx context.Context, userID, lessonID uuid.UUID) error
	IsLiked(ctx context.Context, userID, lessonID uuid.UUID) (bool, error)
}

// likeService 点赞服务实现
type likeService struct {
	likeRepo   repository.LikeRepository
	lessonRepo repository.LessonRepository
}

// NewLikeService 创建点赞服务
func NewLikeService(likeRepo repository.LikeRepository, lessonRepo repository.LessonRepository) LikeService {
	return &likeService{
		likeRepo:   likeRepo,
		lessonRepo: lessonRepo,
	}
}

func (s *likeService) Like(ctx context.Context, userID, lessonID uuid.UUID) error {
	exists, _ := s.likeRepo.Exists(ctx, userID, lessonID)
	if exists {
		return nil
	}

	like := &model.Like{
		UserID:   userID,
		LessonID: lessonID,
	}

	if err := s.likeRepo.Create(ctx, like); err != nil {
		return err
	}

	_ = s.lessonRepo.UpdateCounts(ctx, lessonID)
	return nil
}

func (s *likeService) Unlike(ctx context.Context, userID, lessonID uuid.UUID) error {
	if err := s.likeRepo.Delete(ctx, userID, lessonID); err != nil {
		return err
	}

	_ = s.lessonRepo.UpdateCounts(ctx, lessonID)
	return nil
}

func (s *likeService) IsLiked(ctx context.Context, userID, lessonID uuid.UUID) (bool, error) {
	return s.likeRepo.Exists(ctx, userID, lessonID)
}
