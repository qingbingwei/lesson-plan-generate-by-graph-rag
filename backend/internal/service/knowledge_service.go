package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"lesson-plan/backend/internal/config"
	"lesson-plan/backend/internal/model"
	"lesson-plan/backend/internal/repository"
)

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
			Timeout: 600 * time.Second,
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
			RelevanceScore: 0.9 - float64(i)*0.05,
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
