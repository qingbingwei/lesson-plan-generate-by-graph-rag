package service

import (
	"context"
	"errors"
	"strings"

	"lesson-plan/backend/internal/model"
	"lesson-plan/backend/internal/repository"

	"github.com/google/uuid"
)

// CommentService 评论服务接口
type CommentService interface {
	Create(ctx context.Context, userID, lessonID uuid.UUID, content string, parentID *uuid.UUID) (*model.Comment, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
	List(ctx context.Context, lessonID uuid.UUID, page, pageSize int) ([]model.Comment, int64, error)
}

// commentService 评论服务实现
type commentService struct {
	commentRepo repository.CommentRepository
	lessonRepo  repository.LessonRepository
}

// NewCommentService 创建评论服务
func NewCommentService(commentRepo repository.CommentRepository, lessonRepo repository.LessonRepository) CommentService {
	return &commentService{
		commentRepo: commentRepo,
		lessonRepo:  lessonRepo,
	}
}

func (s *commentService) Create(ctx context.Context, userID, lessonID uuid.UUID, content string, parentID *uuid.UUID) (*model.Comment, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, errors.New("评论内容不能为空")
	}

	comment := &model.Comment{
		LessonID: lessonID,
		UserID:   userID,
		ParentID: parentID,
		Content:  content,
	}

	if err := s.commentRepo.Create(ctx, comment); err != nil {
		return nil, err
	}

	_ = s.lessonRepo.UpdateCounts(ctx, lessonID)
	return comment, nil
}

func (s *commentService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	comment, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		return ErrCommentNotFound
	}

	if comment.UserID != userID {
		return ErrUnauthorized
	}

	if err := s.commentRepo.Delete(ctx, id); err != nil {
		return err
	}

	_ = s.lessonRepo.UpdateCounts(ctx, comment.LessonID)
	return nil
}

func (s *commentService) List(ctx context.Context, lessonID uuid.UUID, page, pageSize int) ([]model.Comment, int64, error) {
	return s.commentRepo.ListByLessonID(ctx, lessonID, page, pageSize)
}
