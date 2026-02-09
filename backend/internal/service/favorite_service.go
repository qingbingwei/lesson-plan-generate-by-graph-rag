package service

import (
	"context"

	"lesson-plan/backend/internal/model"
	"lesson-plan/backend/internal/repository"

	"github.com/google/uuid"
)

// FavoriteService 收藏服务接口
type FavoriteService interface {
	Add(ctx context.Context, userID, lessonID uuid.UUID) error
	Remove(ctx context.Context, userID, lessonID uuid.UUID) error
	List(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.LessonListItem, int64, error)
	IsFavorited(ctx context.Context, userID, lessonID uuid.UUID) (bool, error)
}

// favoriteService 收藏服务实现
type favoriteService struct {
	favoriteRepo repository.FavoriteRepository
	lessonRepo   repository.LessonRepository
}

// NewFavoriteService 创建收藏服务
func NewFavoriteService(favoriteRepo repository.FavoriteRepository, lessonRepo repository.LessonRepository) FavoriteService {
	return &favoriteService{
		favoriteRepo: favoriteRepo,
		lessonRepo:   lessonRepo,
	}
}

func (s *favoriteService) Add(ctx context.Context, userID, lessonID uuid.UUID) error {
	exists, _ := s.favoriteRepo.Exists(ctx, userID, lessonID)
	if exists {
		return nil
	}

	favorite := &model.Favorite{
		UserID:   userID,
		LessonID: lessonID,
	}

	if err := s.favoriteRepo.Create(ctx, favorite); err != nil {
		return err
	}

	_ = s.lessonRepo.UpdateCounts(ctx, lessonID)
	return nil
}

func (s *favoriteService) Remove(ctx context.Context, userID, lessonID uuid.UUID) error {
	if err := s.favoriteRepo.Delete(ctx, userID, lessonID); err != nil {
		return err
	}

	_ = s.lessonRepo.UpdateCounts(ctx, lessonID)
	return nil
}

func (s *favoriteService) List(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.LessonListItem, int64, error) {
	favorites, total, err := s.favoriteRepo.ListByUserID(ctx, userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	items := make([]model.LessonListItem, 0, len(favorites))
	for _, f := range favorites {
		if f.Lesson != nil {
			item := model.LessonListItem{
				ID:            f.Lesson.ID,
				Title:         f.Lesson.Title,
				Subject:       f.Lesson.Subject,
				Grade:         f.Lesson.Grade,
				Duration:      f.Lesson.Duration,
				Status:        f.Lesson.Status,
				ViewCount:     f.Lesson.ViewCount,
				LikeCount:     f.Lesson.LikeCount,
				FavoriteCount: f.Lesson.FavoriteCount,
				CreatedAt:     f.Lesson.CreatedAt,
				PublishedAt:   f.Lesson.PublishedAt,
			}
			if f.Lesson.User != nil {
				item.AuthorName = f.Lesson.User.FullName
				if item.AuthorName == "" {
					item.AuthorName = f.Lesson.User.Username
				}
				item.AuthorAvatar = f.Lesson.User.AvatarURL
			}
			items = append(items, item)
		}
	}

	return items, total, nil
}

func (s *favoriteService) IsFavorited(ctx context.Context, userID, lessonID uuid.UUID) (bool, error) {
	return s.favoriteRepo.Exists(ctx, userID, lessonID)
}
