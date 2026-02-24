package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTemplateNotFound  = errors.New("模板不存在")
	ErrTemplateForbidden = errors.New("无权访问该模板")
)

// LessonTemplate 教案模板实体。
type LessonTemplate struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description,omitempty"`
	Category       string    `json:"category,omitempty"`
	Subject        string    `json:"subject,omitempty"`
	Grade          string    `json:"grade,omitempty"`
	Duration       int       `json:"duration,omitempty"`
	TopicHint      string    `json:"topic_hint,omitempty"`
	Style          string    `json:"style,omitempty"`
	Requirements   string    `json:"requirements,omitempty"`
	Objectives     string    `json:"objectives,omitempty"`
	ContentOutline string    `json:"content_outline,omitempty"`
	Activities     string    `json:"activities,omitempty"`
	Assessment     string    `json:"assessment,omitempty"`
	Resources      string    `json:"resources,omitempty"`
	Tags           []string  `json:"tags,omitempty"`
	BuiltIn        bool      `json:"built_in"`
	OwnerID        string    `json:"owner_id,omitempty"`
	UsageCount     int       `json:"usage_count"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// CreateLessonTemplateRequest 创建模板请求。
type CreateLessonTemplateRequest struct {
	Name           string   `json:"name" binding:"required,max=120"`
	Description    string   `json:"description" binding:"max=500"`
	Category       string   `json:"category" binding:"max=50"`
	Subject        string   `json:"subject" binding:"max=50"`
	Grade          string   `json:"grade" binding:"max=20"`
	Duration       int      `json:"duration"`
	TopicHint      string   `json:"topic_hint" binding:"max=200"`
	Style          string   `json:"style" binding:"max=50"`
	Requirements   string   `json:"requirements"`
	Objectives     string   `json:"objectives"`
	ContentOutline string   `json:"content_outline"`
	Activities     string   `json:"activities"`
	Assessment     string   `json:"assessment"`
	Resources      string   `json:"resources"`
	Tags           []string `json:"tags"`
}

// AppliedTemplatePayload 模板复用返回结构（可直接用于生成页表单）。
type AppliedTemplatePayload struct {
	TemplateID   string   `json:"template_id"`
	Name         string   `json:"name"`
	Subject      string   `json:"subject,omitempty"`
	Grade        string   `json:"grade,omitempty"`
	Duration     int      `json:"duration,omitempty"`
	Topic        string   `json:"topic,omitempty"`
	Style        string   `json:"style,omitempty"`
	Requirements string   `json:"requirements,omitempty"`
	Tags         []string `json:"tags,omitempty"`
}

// TemplateService 模板库服务。
type TemplateService interface {
	List(ctx context.Context, userID uuid.UUID) ([]LessonTemplate, error)
	Get(ctx context.Context, id string, userID uuid.UUID) (*LessonTemplate, error)
	Create(ctx context.Context, userID uuid.UUID, req *CreateLessonTemplateRequest) (*LessonTemplate, error)
	Delete(ctx context.Context, id string, userID uuid.UUID) error
	Apply(ctx context.Context, id string, userID uuid.UUID) (*AppliedTemplatePayload, error)
}

type templateStore struct {
	Templates []LessonTemplate `json:"templates"`
}

type templateService struct {
	mu sync.RWMutex

	storePath string
	templates map[string]LessonTemplate
}

// NewTemplateService 创建模板库服务。
func NewTemplateService(storePath string) TemplateService {
	svc := &templateService{
		storePath: storePath,
		templates: make(map[string]LessonTemplate),
	}

	now := time.Now().UTC()
	for _, tpl := range defaultBuiltInTemplates(now) {
		svc.templates[tpl.ID] = tpl
	}

	_ = svc.loadFromDisk()
	return svc
}

func defaultBuiltInTemplates(now time.Time) []LessonTemplate {
	return []LessonTemplate{
		{
			ID:           "builtin-primary-inquiry",
			Name:         "小学探究课模板",
			Description:  "适用于小学阶段，以提问-探究-展示为主线。",
			Category:     "探究课",
			Subject:      "科学",
			Grade:        "五年级",
			Duration:     40,
			TopicHint:    "围绕一个生活化问题展开课堂探究",
			Style:        "interactive",
			Requirements: "包含小组合作、实验观察与课堂展示环节。",
			Tags:         []string{"小学", "探究", "合作学习"},
			BuiltIn:      true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           "builtin-junior-reading",
			Name:         "初中阅读课模板",
			Description:  "适用于语文/英语阅读课，强调目标分层与任务驱动。",
			Category:     "阅读课",
			Subject:      "语文",
			Grade:        "七年级",
			Duration:     45,
			TopicHint:    "围绕文本理解与表达训练设计教学活动",
			Style:        "lecture",
			Requirements: "包含导入、精读、迁移练习和形成性评价。",
			Tags:         []string{"初中", "阅读", "任务驱动"},
			BuiltIn:      true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           "builtin-senior-problem-solving",
			Name:         "高中问题解决模板",
			Description:  "适用于数理化课程，强调例题剖析与分层训练。",
			Category:     "习题课",
			Subject:      "数学",
			Grade:        "高一",
			Duration:     45,
			TopicHint:    "围绕核心概念设计问题链与分层练习",
			Style:        "project",
			Requirements: "包含错因分析、变式训练、课堂小结。",
			Tags:         []string{"高中", "问题解决", "分层训练"},
			BuiltIn:      true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}
}

func (s *templateService) loadFromDisk() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	raw, err := os.ReadFile(s.storePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	var payload templateStore
	if err := json.Unmarshal(raw, &payload); err != nil {
		return err
	}

	for _, tpl := range payload.Templates {
		if tpl.ID == "" {
			continue
		}
		// 内置模板以代码为准，防止外部覆盖。
		if existing, ok := s.templates[tpl.ID]; ok && existing.BuiltIn {
			continue
		}
		s.templates[tpl.ID] = tpl
	}

	return nil
}

func (s *templateService) persist() error {
	payload := templateStore{
		Templates: make([]LessonTemplate, 0, len(s.templates)),
	}
	for _, tpl := range s.templates {
		payload.Templates = append(payload.Templates, tpl)
	}

	sort.Slice(payload.Templates, func(i, j int) bool {
		if payload.Templates[i].BuiltIn != payload.Templates[j].BuiltIn {
			return payload.Templates[i].BuiltIn
		}
		return payload.Templates[i].UpdatedAt.After(payload.Templates[j].UpdatedAt)
	})

	if err := os.MkdirAll(filepath.Dir(s.storePath), 0755); err != nil {
		return err
	}

	body, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}

	tempPath := s.storePath + ".tmp"
	if err := os.WriteFile(tempPath, body, 0644); err != nil {
		return err
	}

	return os.Rename(tempPath, s.storePath)
}

func (s *templateService) visibleToUser(tpl LessonTemplate, userID uuid.UUID) bool {
	if tpl.BuiltIn {
		return true
	}
	return strings.EqualFold(tpl.OwnerID, userID.String())
}

func (s *templateService) List(_ context.Context, userID uuid.UUID) ([]LessonTemplate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]LessonTemplate, 0, len(s.templates))
	for _, tpl := range s.templates {
		if s.visibleToUser(tpl, userID) {
			result = append(result, tpl)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].BuiltIn != result[j].BuiltIn {
			return result[i].BuiltIn
		}
		if result[i].UsageCount != result[j].UsageCount {
			return result[i].UsageCount > result[j].UsageCount
		}
		return result[i].UpdatedAt.After(result[j].UpdatedAt)
	})

	return result, nil
}

func (s *templateService) Get(_ context.Context, id string, userID uuid.UUID) (*LessonTemplate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tpl, ok := s.templates[strings.TrimSpace(id)]
	if !ok {
		return nil, ErrTemplateNotFound
	}
	if !s.visibleToUser(tpl, userID) {
		return nil, ErrTemplateForbidden
	}

	copyTpl := tpl
	return &copyTpl, nil
}

func (s *templateService) Create(_ context.Context, userID uuid.UUID, req *CreateLessonTemplateRequest) (*LessonTemplate, error) {
	if req == nil {
		return nil, errors.New("请求不能为空")
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("模板名称不能为空")
	}

	now := time.Now().UTC()
	tpl := LessonTemplate{
		ID:             "tpl-" + uuid.NewString(),
		Name:           name,
		Description:    strings.TrimSpace(req.Description),
		Category:       strings.TrimSpace(req.Category),
		Subject:        strings.TrimSpace(req.Subject),
		Grade:          strings.TrimSpace(req.Grade),
		Duration:       req.Duration,
		TopicHint:      strings.TrimSpace(req.TopicHint),
		Style:          strings.TrimSpace(req.Style),
		Requirements:   strings.TrimSpace(req.Requirements),
		Objectives:     strings.TrimSpace(req.Objectives),
		ContentOutline: strings.TrimSpace(req.ContentOutline),
		Activities:     strings.TrimSpace(req.Activities),
		Assessment:     strings.TrimSpace(req.Assessment),
		Resources:      strings.TrimSpace(req.Resources),
		Tags:           req.Tags,
		BuiltIn:        false,
		OwnerID:        userID.String(),
		UsageCount:     0,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if tpl.Duration <= 0 {
		tpl.Duration = 45
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.templates[tpl.ID] = tpl
	if err := s.persist(); err != nil {
		delete(s.templates, tpl.ID)
		return nil, fmt.Errorf("保存模板失败: %w", err)
	}

	copyTpl := tpl
	return &copyTpl, nil
}

func (s *templateService) Delete(_ context.Context, id string, userID uuid.UUID) error {
	templateID := strings.TrimSpace(id)
	if templateID == "" {
		return ErrTemplateNotFound
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	tpl, ok := s.templates[templateID]
	if !ok {
		return ErrTemplateNotFound
	}
	if tpl.BuiltIn {
		return errors.New("内置模板不允许删除")
	}
	if !strings.EqualFold(tpl.OwnerID, userID.String()) {
		return ErrTemplateForbidden
	}

	delete(s.templates, templateID)
	if err := s.persist(); err != nil {
		s.templates[templateID] = tpl
		return fmt.Errorf("删除模板失败: %w", err)
	}
	return nil
}

func (s *templateService) Apply(ctx context.Context, id string, userID uuid.UUID) (*AppliedTemplatePayload, error) {
	tpl, err := s.Get(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	// 仅用法计数需要写锁。
	s.mu.Lock()
	if current, ok := s.templates[tpl.ID]; ok {
		current.UsageCount++
		current.UpdatedAt = time.Now().UTC()
		s.templates[tpl.ID] = current
		_ = s.persist()
	}
	s.mu.Unlock()

	return &AppliedTemplatePayload{
		TemplateID:   tpl.ID,
		Name:         tpl.Name,
		Subject:      tpl.Subject,
		Grade:        tpl.Grade,
		Duration:     tpl.Duration,
		Topic:        tpl.TopicHint,
		Style:        tpl.Style,
		Requirements: tpl.Requirements,
		Tags:         tpl.Tags,
	}, nil
}
