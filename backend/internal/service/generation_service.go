package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"lesson-plan/backend/internal/config"
	"lesson-plan/backend/internal/model"
	"lesson-plan/backend/internal/repository"

	"github.com/google/uuid"
)

// GenerationService 生成服务接口
type GenerationService interface {
	Generate(ctx context.Context, userID uuid.UUID, req *model.GenerationRequest, keyOverride APIKeyOverride) (*model.GenerationResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Generation, error)
	ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.Generation, int64, error)
	GetStats(ctx context.Context, userID uuid.UUID) (*repository.GenerationStats, error)
	GetLangSmithUsage(ctx context.Context, userID uuid.UUID, page, pageSize int) (*LangSmithUsagePayload, error)
	AskAssistant(ctx context.Context, userID uuid.UUID, req *AssistantChatRequest, keyOverride APIKeyOverride) (*AssistantChatPayload, error)
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

func (s *generationService) Generate(ctx context.Context, userID uuid.UUID, req *model.GenerationRequest, keyOverride APIKeyOverride) (*model.GenerationResponse, error) {
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

	agentResp, err := s.callAgent(ctx, userID, req, keyOverride)
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
		if tokenCount == 0 {
			tokenCount = agentResp.Usage.PromptTokens + agentResp.Usage.CompletionTokens
		}
	}

	resultJSON, _ := json.Marshal(agentResp.Data)
	if err := s.generationRepo.UpdateResult(ctx, generation.ID, string(resultJSON), tokenCount); err != nil {
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

func (s *generationService) GetLangSmithUsage(ctx context.Context, userID uuid.UUID, page, pageSize int) (*LangSmithUsagePayload, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	url := fmt.Sprintf("%s/api/langsmith/token-usage?userId=%s&page=%d&pageSize=%d", s.cfg.URL, userID.String(), page, pageSize)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create langsmith usage request failed: %w", err)
	}

	if s.cfg.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+s.cfg.APIKey)
	}

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("call langsmith usage endpoint failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read langsmith usage response failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("langsmith usage endpoint returned error: %d - %s", resp.StatusCode, string(body))
	}

	var agentResp struct {
		Success bool                  `json:"success"`
		Source  string                `json:"source"`
		Project string                `json:"project"`
		Stats   LangSmithUsageStats   `json:"stats"`
		History LangSmithUsageHistory `json:"history"`
		Error   string                `json:"error,omitempty"`
	}

	if err := json.Unmarshal(body, &agentResp); err != nil {
		return nil, fmt.Errorf("unmarshal langsmith usage response failed: %w", err)
	}

	if !agentResp.Success {
		if agentResp.Error != "" {
			return nil, fmt.Errorf("langsmith usage query failed: %s", agentResp.Error)
		}
		return nil, fmt.Errorf("langsmith usage query failed")
	}

	return &LangSmithUsagePayload{
		Source:  agentResp.Source,
		Project: agentResp.Project,
		Stats:   agentResp.Stats,
		History: agentResp.History,
	}, nil
}

func (s *generationService) AskAssistant(ctx context.Context, userID uuid.UUID, req *AssistantChatRequest, keyOverride APIKeyOverride) (*AssistantChatPayload, error) {
	if req == nil {
		return nil, fmt.Errorf("assistant request is nil")
	}

	req.Question = strings.TrimSpace(req.Question)
	if req.Question == "" {
		return nil, fmt.Errorf("assistant question is required")
	}
	req.UserID = userID.String()

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal assistant request failed: %w", err)
	}

	url := fmt.Sprintf("%s/api/assistant/chat", s.cfg.URL)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create assistant request failed: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if keyOverride.GenerationAPIKey != "" {
		httpReq.Header.Set(HeaderGenerationAPIKey, keyOverride.GenerationAPIKey)
	}
	if s.cfg.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+s.cfg.APIKey)
	}

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("call assistant endpoint failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read assistant response failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("assistant endpoint returned error: %d - %s", resp.StatusCode, string(respBody))
	}

	var agentResp struct {
		Success bool `json:"success"`
		Data    struct {
			Answer      string   `json:"answer"`
			Suggestions []string `json:"suggestions,omitempty"`
		} `json:"data"`
		Usage *TokenUsage `json:"usage,omitempty"`
		Error string      `json:"error,omitempty"`
	}

	if err := json.Unmarshal(respBody, &agentResp); err != nil {
		return nil, fmt.Errorf("unmarshal assistant response failed: %w", err)
	}

	if !agentResp.Success {
		if agentResp.Error != "" {
			return nil, fmt.Errorf("assistant query failed: %s", agentResp.Error)
		}
		return nil, fmt.Errorf("assistant query failed")
	}

	return &AssistantChatPayload{
		Answer:      strings.TrimSpace(agentResp.Data.Answer),
		Suggestions: agentResp.Data.Suggestions,
		Usage:       agentResp.Usage,
	}, nil
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

func (s *generationService) callAgent(ctx context.Context, userID uuid.UUID, req *model.GenerationRequest, keyOverride APIKeyOverride) (*AgentResponse, error) {
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
	if keyOverride.GenerationAPIKey != "" {
		httpReq.Header.Set(HeaderGenerationAPIKey, keyOverride.GenerationAPIKey)
	}
	if keyOverride.EmbeddingAPIKey != "" {
		httpReq.Header.Set(HeaderEmbeddingAPIKey, keyOverride.EmbeddingAPIKey)
	}
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
