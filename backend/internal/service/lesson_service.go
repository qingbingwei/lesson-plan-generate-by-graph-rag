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
}

// FavoriteService 收藏服务接口
type FavoriteService interface {
	Add(ctx context.Context, userID, lessonID uuid.UUID) error
	Remove(ctx context.Context, userID, lessonID uuid.UUID) error
	List(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.LessonListItem, int64, error)
	IsFavorited(ctx context.Context, userID, lessonID uuid.UUID) (bool, error)
}

// LikeService 点赞服务接口
type LikeService interface {
	Like(ctx context.Context, userID, lessonID uuid.UUID) error
	Unlike(ctx context.Context, userID, lessonID uuid.UUID) error
	IsLiked(ctx context.Context, userID, lessonID uuid.UUID) (bool, error)
}

// CommentService 评论服务接口
type CommentService interface {
	Create(ctx context.Context, userID, lessonID uuid.UUID, content string, parentID *uuid.UUID) (*model.Comment, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
	List(ctx context.Context, lessonID uuid.UUID, page, pageSize int) ([]model.Comment, int64, error)
}

// lessonService 教案服务实现
type lessonService struct {
	lessonRepo   repository.LessonRepository
	favoriteRepo repository.FavoriteRepository
	likeRepo     repository.LikeRepository
}

// NewLessonService 创建教案服务
func NewLessonService(
	lessonRepo repository.LessonRepository,
	favoriteRepo repository.FavoriteRepository,
	likeRepo repository.LikeRepository,
) LessonService {
	return &lessonService{
		lessonRepo:   lessonRepo,
		favoriteRepo: favoriteRepo,
		likeRepo:     likeRepo,
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
