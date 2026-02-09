package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"lesson-plan/backend/internal/config"
	"lesson-plan/backend/internal/model"
	"lesson-plan/backend/internal/repository"

	"github.com/google/uuid"
)

// GenerationService 生成服务接口
type GenerationService interface {
	Generate(ctx context.Context, userID uuid.UUID, req *model.GenerationRequest) (*model.GenerationResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Generation, error)
	ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.Generation, int64, error)
	GetStats(ctx context.Context, userID uuid.UUID) (*repository.GenerationStats, error)
}

// generationService 生成服务实现
type generationService struct {
	generationRepo repository.GenerationRepository
	lessonRepo     repository.LessonRepository
	cfg            *config.AgentConfig
	httpClient     *http.Client
}

// NewGenerationService 创建生成服务
func NewGenerationService(
	generationRepo repository.GenerationRepository,
	lessonRepo repository.LessonRepository,
	cfg *config.AgentConfig,
) GenerationService {
	return &generationService{
		generationRepo: generationRepo,
		lessonRepo:     lessonRepo,
		cfg:            cfg,
		httpClient: &http.Client{
			Timeout: 600 * time.Second,
		},
	}
}

func (s *generationService) Generate(ctx context.Context, userID uuid.UUID, req *model.GenerationRequest) (*model.GenerationResponse, error) {
	prompt := s.buildPrompt(req)
	paramsJSON, _ := json.Marshal(req)

	generation := &model.Generation{
		UserID:     userID,
		Prompt:     prompt,
		Parameters: string(paramsJSON),
		Status:     model.GenerationStatusPending,
	}

	if err := s.generationRepo.Create(ctx, generation); err != nil {
		return nil, err
	}

	_ = s.generationRepo.UpdateStatus(ctx, generation.ID, model.GenerationStatusProcessing)

	agentResp, err := s.callAgent(ctx, userID, req)
	if err != nil {
		_ = s.generationRepo.UpdateError(ctx, generation.ID, err.Error())
		return &model.GenerationResponse{
			ID:           generation.ID,
			Status:       model.GenerationStatusFailed,
			ErrorMessage: err.Error(),
		}, nil
	}

	tokenCount := 0
	if agentResp.Usage != nil {
		tokenCount = agentResp.Usage.TotalTokens
	}

	cost := float64(tokenCount) * 0.001

	resultJSON, _ := json.Marshal(agentResp.Data)
	if err := s.generationRepo.UpdateResult(ctx, generation.ID, string(resultJSON), tokenCount, cost); err != nil {
		return nil, err
	}

	var objectives, keyPoints, difficultPoints, teachingMethods, content, activities, assessment, resources string
	if agentResp.Data != nil {
		objectives = FormatObjectives(agentResp.Data.Objectives)
		keyPoints = FormatStringList(agentResp.Data.KeyPoints)
		difficultPoints = FormatStringList(agentResp.Data.DifficultPoints)
		teachingMethods = FormatStringList(agentResp.Data.TeachingMethods)
		content = FormatSections(agentResp.Data.Content.Sections)
		activities = FormatActivities(agentResp.Data.Content.Sections)
		assessment = agentResp.Data.Evaluation
		if agentResp.Data.Content.Homework != "" {
			assessment += "\n\n## 课后作业\n" + agentResp.Data.Content.Homework
		}
		resources = FormatMaterials(agentResp.Data.Content.Materials)
	}

	return &model.GenerationResponse{
		ID:              generation.ID,
		Status:          model.GenerationStatusCompleted,
		Title:           agentResp.Data.Title,
		Objectives:      objectives,
		KeyPoints:       keyPoints,
		DifficultPoints: difficultPoints,
		TeachingMethods: teachingMethods,
		Content:         content,
		Activities:      activities,
		Assessment:      assessment,
		Resources:       resources,
		TokenCount:      tokenCount,
	}, nil
}

func (s *generationService) GetByID(ctx context.Context, id uuid.UUID) (*model.Generation, error) {
	return s.generationRepo.GetByID(ctx, id)
}

func (s *generationService) ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.Generation, int64, error) {
	return s.generationRepo.ListByUserID(ctx, userID, page, pageSize)
}

func (s *generationService) GetStats(ctx context.Context, userID uuid.UUID) (*repository.GenerationStats, error) {
	return s.generationRepo.GetStats(ctx, userID)
}

func (s *generationService) buildPrompt(req *model.GenerationRequest) string {
	prompt := fmt.Sprintf(`请生成一份%s学科%s年级的教案，主题是：%s。

要求：
- 课时时长：%d分钟
- 难度：%s
- 教学风格：%s
`,
		req.Subject,
		req.Grade,
		req.Topic,
		req.Duration,
		req.Difficulty,
		req.Style,
	)

	if len(req.Objectives) > 0 {
		prompt += "\n教学目标：\n"
		for _, obj := range req.Objectives {
			prompt += fmt.Sprintf("- %s\n", obj)
		}
	}

	if len(req.Keywords) > 0 {
		prompt += "\n关键知识点：\n"
		for _, kw := range req.Keywords {
			prompt += fmt.Sprintf("- %s\n", kw)
		}
	}

	return prompt
}

func (s *generationService) callAgent(ctx context.Context, userID uuid.UUID, req *model.GenerationRequest) (*AgentResponse, error) {
	agentReq := &AgentRequest{
		Subject:    req.Subject,
		Grade:      req.Grade,
		Topic:      req.Topic,
		Duration:   req.Duration,
		Objectives: req.Objectives,
		Keywords:   req.Keywords,
		Style:      req.Style,
		Difficulty: req.Difficulty,
		UserId:     userID.String(),
	}

	body, err := json.Marshal(agentReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}

	url := fmt.Sprintf("%s/api/generate", s.cfg.URL)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if s.cfg.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+s.cfg.APIKey)
	}

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("call agent failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("agent returned error: %d - %s", resp.StatusCode, string(respBody))
	}

	var agentResp AgentResponse
	if err := json.Unmarshal(respBody, &agentResp); err != nil {
		return nil, fmt.Errorf("unmarshal response failed: %w", err)
	}

	if !agentResp.Success {
		return nil, fmt.Errorf("generation failed: %s", agentResp.Error)
	}

	return &agentResp, nil
}
