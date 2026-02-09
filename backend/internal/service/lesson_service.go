package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"lesson-plan/backend/internal/model"
	"lesson-plan/backend/internal/repository"

	"github.com/google/uuid"
)

var (
	ErrLessonNotFound  = errors.New("教案不存在")
	ErrUnauthorized    = errors.New("无权操作此教案")
	ErrCommentNotFound = errors.New("评论不存在")
)

// CreateLessonRequest 创建教案请求
type CreateLessonRequest struct {
	Title      string   `json:"title" binding:"required,max=200"`
	Subject    string   `json:"subject" binding:"required,max=50"`
	Grade      string   `json:"grade" binding:"required,max=20"`
	Duration   int      `json:"duration"`
	Objectives string   `json:"objectives"`
	Content    string   `json:"content"`
	Activities string   `json:"activities"`
	Assessment string   `json:"assessment"`
	Resources  string   `json:"resources"`
	Tags       []string `json:"tags"`
}

// UpdateLessonRequest 更新教案请求
type UpdateLessonRequest struct {
	Title      string   `json:"title" binding:"max=200"`
	Subject    string   `json:"subject" binding:"max=50"`
	Grade      string   `json:"grade" binding:"max=20"`
	Duration   int      `json:"duration"`
	Objectives string   `json:"objectives"`
	Content    string   `json:"content"`
	Activities string   `json:"activities"`
	Assessment string   `json:"assessment"`
	Resources  string   `json:"resources"`
	Tags       []string `json:"tags"`
	Status     string   `json:"status"`
}

// LessonService 教案服务接口
type LessonService interface {
	Create(ctx context.Context, userID uuid.UUID, req *CreateLessonRequest) (*model.Lesson, error)
	GetByID(ctx context.Context, id uuid.UUID, currentUserID *uuid.UUID) (*model.LessonDetail, error)
	Update(ctx context.Context, id uuid.UUID, userID uuid.UUID, req *UpdateLessonRequest) (*model.Lesson, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
	List(ctx context.Context, filter repository.LessonFilter, page, pageSize int) ([]model.LessonListItem, int64, error)
	ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.LessonListItem, int64, error)
	Publish(ctx context.Context, id, userID uuid.UUID) error
	Search(ctx context.Context, query string, page, pageSize int) ([]model.LessonListItem, int64, error)
	ListVersions(ctx context.Context, lessonID uuid.UUID, userID uuid.UUID) ([]model.LessonVersion, error)
	GetVersion(ctx context.Context, lessonID uuid.UUID, version int, userID uuid.UUID) (*model.LessonVersion, error)
	RollbackToVersion(ctx context.Context, lessonID uuid.UUID, version int, userID uuid.UUID) (*model.Lesson, error)
}

// lessonService 教案服务实现
type lessonService struct {
	lessonRepo   repository.LessonRepository
	favoriteRepo repository.FavoriteRepository
	likeRepo     repository.LikeRepository
	versionRepo  repository.VersionRepository
}

// NewLessonService 创建教案服务
func NewLessonService(
	lessonRepo repository.LessonRepository,
	favoriteRepo repository.FavoriteRepository,
	likeRepo repository.LikeRepository,
	versionRepo repository.VersionRepository,
) LessonService {
	return &lessonService{
		lessonRepo:   lessonRepo,
		favoriteRepo: favoriteRepo,
		likeRepo:     likeRepo,
		versionRepo:  versionRepo,
	}
}

func (s *lessonService) Create(ctx context.Context, userID uuid.UUID, req *CreateLessonRequest) (*model.Lesson, error) {
	tagsJSON, _ := json.Marshal(req.Tags)

	// 将objectives和content包装为JSON对象字符串（因为数据库是jsonb类型）
	// 使用简单的字符串包装，避免双重编码
	objectivesJSON := fmt.Sprintf(`{"text": %s}`, strconv.Quote(req.Objectives))
	contentJSON := fmt.Sprintf(`{"text": %s}`, strconv.Quote(req.Content))

	lesson := &model.Lesson{
		UserID:     userID,
		Title:      req.Title,
		Subject:    req.Subject,
		Grade:      req.Grade,
		Duration:   req.Duration,
		Objectives: objectivesJSON,
		Content:    contentJSON,
		Activities: req.Activities,
		Assessment: req.Assessment,
		Resources:  req.Resources,
		Tags:       string(tagsJSON),
		Status:     model.LessonStatusDraft,
	}

	if err := s.lessonRepo.Create(ctx, lesson); err != nil {
		return nil, err
	}

	return lesson, nil
}

func (s *lessonService) GetByID(ctx context.Context, id uuid.UUID, currentUserID *uuid.UUID) (*model.LessonDetail, error) {
	lesson, err := s.lessonRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrLessonNotFound
	}

	// 增加浏览量
	_ = s.lessonRepo.IncrementViewCount(ctx, id)

	detail := &model.LessonDetail{
		ID:            lesson.ID,
		UserID:        lesson.UserID,
		Title:         lesson.Title,
		Subject:       lesson.Subject,
		Grade:         lesson.Grade,
		Duration:      lesson.Duration,
		Objectives:    lesson.Objectives,
		Content:       lesson.Content,
		Activities:    lesson.Activities,
		Assessment:    lesson.Assessment,
		Resources:     lesson.Resources,
		Status:        lesson.Status,
		Version:       lesson.Version,
		ViewCount:     lesson.ViewCount + 1,
		LikeCount:     lesson.LikeCount,
		FavoriteCount: lesson.FavoriteCount,
		CommentCount:  lesson.CommentCount,
		CreatedAt:     lesson.CreatedAt,
		PublishedAt:   lesson.PublishedAt,
	}

	// 解析标签
	if lesson.Tags != "" {
		_ = json.Unmarshal([]byte(lesson.Tags), &detail.Tags)
	}

	// 作者信息
	if lesson.User != nil {
		detail.AuthorName = lesson.User.FullName
		if detail.AuthorName == "" {
			detail.AuthorName = lesson.User.Username
		}
		detail.AuthorAvatar = lesson.User.AvatarURL
	}

	// 检查是否已收藏/点赞
	if currentUserID != nil {
		detail.IsFavorited, _ = s.favoriteRepo.Exists(ctx, *currentUserID, id)
		detail.IsLiked, _ = s.likeRepo.Exists(ctx, *currentUserID, id)
	}

	return detail, nil
}

func (s *lessonService) Update(ctx context.Context, id uuid.UUID, userID uuid.UUID, req *UpdateLessonRequest) (*model.Lesson, error) {
	lesson, err := s.lessonRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrLessonNotFound
	}

	if lesson.UserID != userID {
		return nil, ErrUnauthorized
	}

	// 保存当前版本快照
	if s.versionRepo != nil {
		// 将当前教案内容打包为 JSON
		contentSnapshot, _ := json.Marshal(map[string]interface{}{
			"title":      lesson.Title,
			"objectives": lesson.Objectives,
			"content":    lesson.Content,
			"activities": lesson.Activities,
			"assessment": lesson.Assessment,
			"resources":  lesson.Resources,
			"subject":    lesson.Subject,
			"grade":      lesson.Grade,
			"duration":   lesson.Duration,
		})
		snapshot := &model.LessonVersion{
			LessonID:  lesson.ID,
			Content:   string(contentSnapshot),
			CreatedBy: &userID,
		}
		_ = s.versionRepo.Create(ctx, snapshot)
	}

	// 递增版本号
	lesson.Version++

	if req.Title != "" {
		lesson.Title = req.Title
	}
	if req.Subject != "" {
		lesson.Subject = req.Subject
	}
	if req.Grade != "" {
		lesson.Grade = req.Grade
	}
	if req.Duration > 0 {
		lesson.Duration = req.Duration
	}
	if req.Objectives != "" {
		// 检查是否已经是有效的 JSON
		if strings.HasPrefix(strings.TrimSpace(req.Objectives), "{") {
			lesson.Objectives = req.Objectives
		} else {
			// 包装为 JSON 对象
			lesson.Objectives = fmt.Sprintf(`{"text": %s}`, strconv.Quote(req.Objectives))
		}
	}
	if req.Content != "" {
		// 检查是否已经是有效的 JSON
		if strings.HasPrefix(strings.TrimSpace(req.Content), "{") {
			lesson.Content = req.Content
		} else {
			// 包装为 JSON 对象
			lesson.Content = fmt.Sprintf(`{"text": %s}`, strconv.Quote(req.Content))
		}
	}
	if req.Activities != "" {
		lesson.Activities = req.Activities
	}
	if req.Assessment != "" {
		lesson.Assessment = req.Assessment
	}
	if req.Resources != "" {
		lesson.Resources = req.Resources
	}
	if len(req.Tags) > 0 {
		tagsJSON, _ := json.Marshal(req.Tags)
		lesson.Tags = string(tagsJSON)
	}
	if req.Status != "" {
		lesson.Status = req.Status
	}

	if err := s.lessonRepo.Update(ctx, lesson); err != nil {
		return nil, err
	}

	return lesson, nil
}

func (s *lessonService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	lesson, err := s.lessonRepo.GetByID(ctx, id)
	if err != nil {
		return ErrLessonNotFound
	}

	if lesson.UserID != userID {
		return ErrUnauthorized
	}

	return s.lessonRepo.Delete(ctx, id)
}

func (s *lessonService) List(ctx context.Context, filter repository.LessonFilter, page, pageSize int) ([]model.LessonListItem, int64, error) {
	lessons, total, err := s.lessonRepo.List(ctx, filter, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	items := make([]model.LessonListItem, len(lessons))
	for i, l := range lessons {
		items[i] = s.toListItem(l)
	}

	return items, total, nil
}

func (s *lessonService) ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.LessonListItem, int64, error) {
	lessons, total, err := s.lessonRepo.ListByUserID(ctx, userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	items := make([]model.LessonListItem, len(lessons))
	for i, l := range lessons {
		items[i] = s.toListItem(l)
	}

	return items, total, nil
}

func (s *lessonService) Publish(ctx context.Context, id, userID uuid.UUID) error {
	lesson, err := s.lessonRepo.GetByID(ctx, id)
	if err != nil {
		return ErrLessonNotFound
	}

	if lesson.UserID != userID {
		return ErrUnauthorized
	}

	lesson.Status = model.LessonStatusPublished
	return s.lessonRepo.Update(ctx, lesson)
}

func (s *lessonService) Search(ctx context.Context, query string, page, pageSize int) ([]model.LessonListItem, int64, error) {
	lessons, total, err := s.lessonRepo.Search(ctx, query, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	items := make([]model.LessonListItem, len(lessons))
	for i, l := range lessons {
		items[i] = s.toListItem(l)
	}

	return items, total, nil
}

func (s *lessonService) toListItem(l model.Lesson) model.LessonListItem {
	item := model.LessonListItem{
		ID:            l.ID,
		Title:         l.Title,
		Subject:       l.Subject,
		Grade:         l.Grade,
		Duration:      l.Duration,
		Status:        l.Status,
		Version:       l.Version,
		ViewCount:     l.ViewCount,
		LikeCount:     l.LikeCount,
		FavoriteCount: l.FavoriteCount,
		CreatedAt:     l.CreatedAt,
		PublishedAt:   l.PublishedAt,
	}

	if l.User != nil {
		item.AuthorName = l.User.FullName
		if item.AuthorName == "" {
			item.AuthorName = l.User.Username
		}
		item.AuthorAvatar = l.User.AvatarURL
	}

	return item
}

func (s *lessonService) ListVersions(ctx context.Context, lessonID uuid.UUID, userID uuid.UUID) ([]model.LessonVersion, error) {
	lesson, err := s.lessonRepo.GetByID(ctx, lessonID)
	if err != nil {
		return nil, ErrLessonNotFound
	}
	if lesson.UserID != userID {
		return nil, ErrUnauthorized
	}
	if s.versionRepo == nil {
		return nil, nil
	}
	return s.versionRepo.ListByLessonID(ctx, lessonID)
}

func (s *lessonService) GetVersion(ctx context.Context, lessonID uuid.UUID, version int, userID uuid.UUID) (*model.LessonVersion, error) {
	lesson, err := s.lessonRepo.GetByID(ctx, lessonID)
	if err != nil {
		return nil, ErrLessonNotFound
	}
	if lesson.UserID != userID {
		return nil, ErrUnauthorized
	}
	if s.versionRepo == nil {
		return nil, errors.New("版本功能未启用")
	}
	return s.versionRepo.GetByVersion(ctx, lessonID, version)
}

func (s *lessonService) RollbackToVersion(ctx context.Context, lessonID uuid.UUID, version int, userID uuid.UUID) (*model.Lesson, error) {
	lesson, err := s.lessonRepo.GetByID(ctx, lessonID)
	if err != nil {
		return nil, ErrLessonNotFound
	}
	if lesson.UserID != userID {
		return nil, ErrUnauthorized
	}
	if s.versionRepo == nil {
		return nil, errors.New("版本功能未启用")
	}
	v, err := s.versionRepo.GetByVersion(ctx, lessonID, version)
	if err != nil {
		return nil, errors.New("版本不存在")
	}

	// 先快照当前版本
	contentSnapshot, _ := json.Marshal(map[string]interface{}{
		"title":      lesson.Title,
		"objectives": lesson.Objectives,
		"content":    lesson.Content,
		"activities": lesson.Activities,
		"assessment": lesson.Assessment,
		"resources":  lesson.Resources,
		"subject":    lesson.Subject,
		"grade":      lesson.Grade,
		"duration":   lesson.Duration,
	})
	snapshot := &model.LessonVersion{
		LessonID:      lesson.ID,
		Content:       string(contentSnapshot),
		ChangeSummary: fmt.Sprintf("回滚至版本 %d", version),
		CreatedBy:     &userID,
	}
	_ = s.versionRepo.Create(ctx, snapshot)

	// 从 JSON 中恢复各字段
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(v.Content), &data); err == nil {
		if title, ok := data["title"].(string); ok && title != "" {
			lesson.Title = title
		}
		if obj, ok := data["objectives"].(string); ok {
			lesson.Objectives = obj
		}
		if content, ok := data["content"].(string); ok {
			lesson.Content = content
		}
		if act, ok := data["activities"].(string); ok {
			lesson.Activities = act
		}
		if assess, ok := data["assessment"].(string); ok {
			lesson.Assessment = assess
		}
		if res, ok := data["resources"].(string); ok {
			lesson.Resources = res
		}
	}

	lesson.Version++
	if err := s.lessonRepo.Update(ctx, lesson); err != nil {
		return nil, err
	}
	return lesson, nil
}
