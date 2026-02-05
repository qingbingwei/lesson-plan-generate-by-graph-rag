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
			Timeout: 600 * time.Second, // 10分钟，生成教案需要较长时间
		},
	}
}

func (s *generationService) Generate(ctx context.Context, userID uuid.UUID, req *model.GenerationRequest) (*model.GenerationResponse, error) {
	// 构建prompt
	prompt := s.buildPrompt(req)

	// 序列化参数
	paramsJSON, _ := json.Marshal(req)

	// 创建生成记录
	generation := &model.Generation{
		UserID:     userID,
		Prompt:     prompt,
		Parameters: string(paramsJSON),
		Status:     model.GenerationStatusPending,
	}

	if err := s.generationRepo.Create(ctx, generation); err != nil {
		return nil, err
	}

	// 更新状态为处理中
	_ = s.generationRepo.UpdateStatus(ctx, generation.ID, model.GenerationStatusProcessing)

	// 调用Agent服务（传递userID用于过滤个人知识库）
	agentResp, err := s.callAgent(ctx, userID, req)
	if err != nil {
		_ = s.generationRepo.UpdateError(ctx, generation.ID, err.Error())
		return &model.GenerationResponse{
			ID:           generation.ID,
			Status:       model.GenerationStatusFailed,
			ErrorMessage: err.Error(),
		}, nil
	}

	// 提取token数量
	tokenCount := 0
	if agentResp.Usage != nil {
		tokenCount = agentResp.Usage.TotalTokens
	}

	// 计算成本（示例：0.001元/token）
	cost := float64(tokenCount) * 0.001

	// 将完整的生成数据序列化存储
	resultJSON, _ := json.Marshal(agentResp.Data)
	if err := s.generationRepo.UpdateResult(ctx, generation.ID, string(resultJSON), tokenCount, cost); err != nil {
		return nil, err
	}

	// 将复杂结构转换为简单字符串格式返回
	var objectives, keyPoints, difficultPoints, teachingMethods, content, activities, assessment, resources string
	if agentResp.Data != nil {
		// 转换教学目标
		objectives = formatObjectives(agentResp.Data.Objectives)

		// 转换教学重点、难点、方法
		keyPoints = formatStringList(agentResp.Data.KeyPoints)
		difficultPoints = formatStringList(agentResp.Data.DifficultPoints)
		teachingMethods = formatStringList(agentResp.Data.TeachingMethods)

		// 转换内容和活动
		content = formatSections(agentResp.Data.Content.Sections)
		activities = formatActivities(agentResp.Data.Content.Sections)

		// 评价和资源
		assessment = agentResp.Data.Evaluation
		if agentResp.Data.Content.Homework != "" {
			assessment += "\n\n## 课后作业\n" + agentResp.Data.Content.Homework
		}
		resources = formatMaterials(agentResp.Data.Content.Materials)
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

// 辅助函数：格式化教学目标
func formatObjectives(obj LessonObjectives) string {
	result := ""
	if obj.Knowledge != "" {
		result += "【知识与技能】\n" + obj.Knowledge + "\n\n"
	}
	if obj.Process != "" {
		result += "【过程与方法】\n" + obj.Process + "\n\n"
	}
	if obj.Emotion != "" {
		result += "【情感态度价值观】\n" + obj.Emotion
	}
	return result
}

// 辅助函数：格式化字符串列表
func formatStringList(items []string) string {
	if len(items) == 0 {
		return ""
	}
	result := ""
	for i, item := range items {
		if i > 0 {
			result += "\n"
		}
		result += fmt.Sprintf("%d. %s", i+1, item)
	}
	return result
}

// 辅助函数：格式化教学环节
func formatSections(sections []LessonSection) string {
	result := ""
	for i, section := range sections {
		if i > 0 {
			result += "\n\n"
		}
		result += fmt.Sprintf("## %s (%d分钟)\n\n", section.Title, section.Duration)
		if section.Content != "" {
			result += section.Content + "\n\n"
		}
		if section.TeacherActivity != "" {
			result += "**教师活动：**\n" + section.TeacherActivity + "\n\n"
		}
		if section.StudentActivity != "" {
			result += "**学生活动：**\n" + section.StudentActivity
		}
	}
	return result
}

// 辅助函数：格式化学生活动
func formatActivities(sections []LessonSection) string {
	result := ""
	for i, section := range sections {
		if section.StudentActivity != "" {
			if i > 0 {
				result += "\n\n"
			}
			result += fmt.Sprintf("**%s阶段：**\n%s", section.Title, section.StudentActivity)
		}
	}
	return result
}

// 辅助函数：格式化教学材料
func formatMaterials(materials []string) string {
	result := ""
	for i, material := range materials {
		if i > 0 {
			result += "\n"
		}
		result += fmt.Sprintf("%d. %s", i+1, material)
	}
	return result
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

// AgentRequest Agent请求
type AgentRequest struct {
	Subject    string   `json:"subject"`
	Grade      string   `json:"grade"`
	Topic      string   `json:"topic"`
	Duration   int      `json:"duration"`
	Objectives []string `json:"objectives"`
	Keywords   []string `json:"keywords"`
	Style      string   `json:"style"`
	Difficulty string   `json:"difficulty"`
	UserId     string   `json:"userId"` // 用户ID，用于过滤个人知识库
}

// AgentResponse Agent响应
type AgentResponse struct {
	Success bool                 `json:"success"`
	Data    *GeneratedLessonData `json:"data"`
	Error   string               `json:"error,omitempty"`
	Usage   *TokenUsage          `json:"usage,omitempty"`
}

// GeneratedLessonData 生成的教案数据
type GeneratedLessonData struct {
	Title           string           `json:"title"`
	Objectives      LessonObjectives `json:"objectives"`
	KeyPoints       []string         `json:"keyPoints"`
	DifficultPoints []string         `json:"difficultPoints"`
	TeachingMethods []string         `json:"teachingMethods"`
	Content         LessonContent    `json:"content"`
	Evaluation      string           `json:"evaluation"`
	Reflection      string           `json:"reflection,omitempty"`
}

// LessonObjectives 教学目标
type LessonObjectives struct {
	Knowledge string `json:"knowledge"`
	Process   string `json:"process"`
	Emotion   string `json:"emotion"`
}

// LessonContent 教学内容
type LessonContent struct {
	Sections  []LessonSection `json:"sections"`
	Materials []string        `json:"materials"`
	Homework  string          `json:"homework"`
}

// LessonSection 教学环节
type LessonSection struct {
	Title           string `json:"title"`
	Duration        int    `json:"duration"`
	TeacherActivity string `json:"teacherActivity"`
	StudentActivity string `json:"studentActivity"`
	Content         string `json:"content"`
	DesignIntent    string `json:"designIntent,omitempty"`
}

// TokenUsage Token使用情况
type TokenUsage struct {
	PromptTokens     int `json:"promptTokens"`
	CompletionTokens int `json:"completionTokens"`
	TotalTokens      int `json:"totalTokens"`
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
		UserId:     userID.String(), // 传递用户ID用于过滤个人知识库
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

// KnowledgeService 知识服务接口
type KnowledgeService interface {
	Search(ctx context.Context, query string, limit int) ([]model.KnowledgeSearchResult, error)
	GetGraph(ctx context.Context, subject, grade, userId string, limit int) (*model.KnowledgeGraph, error)
	GetEmbedding(ctx context.Context, text string) ([]float64, error)
}

// knowledgeService 知识服务实现
type knowledgeService struct {
	knowledgeRepo repository.KnowledgeRepository
	cfg           *config.AgentConfig
	httpClient    *http.Client
}

// NewKnowledgeService 创建知识服务
func NewKnowledgeService(
	knowledgeRepo repository.KnowledgeRepository,
	cfg *config.AgentConfig,
) KnowledgeService {
	return &knowledgeService{
		knowledgeRepo: knowledgeRepo,
		cfg:           cfg,
		httpClient: &http.Client{
			Timeout: 600 * time.Second, // 10分钟
		},
	}
}

func (s *knowledgeService) Search(ctx context.Context, query string, limit int) ([]model.KnowledgeSearchResult, error) {
	// 获取查询的embedding
	embedding, err := s.GetEmbedding(ctx, query)
	if err != nil {
		// 如果embedding失败，回退到文本搜索
		knowledges, err := s.knowledgeRepo.Search(ctx, query, limit)
		if err != nil {
			return nil, err
		}

		results := make([]model.KnowledgeSearchResult, len(knowledges))
		for i, k := range knowledges {
			results[i] = model.KnowledgeSearchResult{
				ID:             k.ID,
				Name:           k.Name,
				Content:        k.Description,
				RelevanceScore: 1.0,
				Source:         "text_search",
			}
		}
		return results, nil
	}

	// 使用embedding进行向量搜索
	knowledges, err := s.knowledgeRepo.SearchByEmbedding(ctx, embedding, limit)
	if err != nil {
		return nil, err
	}

	results := make([]model.KnowledgeSearchResult, len(knowledges))
	for i, k := range knowledges {
		results[i] = model.KnowledgeSearchResult{
			ID:             k.ID,
			Name:           k.Name,
			Content:        k.Description,
			RelevanceScore: 0.9 - float64(i)*0.05, // 简单的相关性评分
			Source:         "vector_search",
		}
	}

	return results, nil
}

func (s *knowledgeService) GetGraph(ctx context.Context, subject, grade, userId string, limit int) (*model.KnowledgeGraph, error) {
	return s.knowledgeRepo.GetGraph(ctx, subject, grade, userId, limit)
}

func (s *knowledgeService) GetEmbedding(ctx context.Context, text string) ([]float64, error) {
	reqBody := map[string]interface{}{
		"text": text,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/embedding", s.cfg.URL)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if s.cfg.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+s.cfg.APIKey)
	}

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedding API returned status: %d", resp.StatusCode)
	}

	var result struct {
		Embedding []float64 `json:"embedding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Embedding, nil
}
