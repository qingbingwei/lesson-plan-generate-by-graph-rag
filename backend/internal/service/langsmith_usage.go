package service

// LangSmithUsageStats LangSmith 统计数据。
type LangSmithUsageStats struct {
	TotalCount           int64   `json:"total_count"`
	CompletedCount       int64   `json:"completed_count"`
	FailedCount          int64   `json:"failed_count"`
	TotalTokens          int64   `json:"total_tokens"`
	AvgDurationMs        float64 `json:"avg_duration_ms"`
	ThisMonthGenerations int64   `json:"this_month_generations"`
	TotalLessons         int64   `json:"total_lessons"`
}

// LangSmithUsageHistoryItem LangSmith 历史条目。
type LangSmithUsageHistoryItem struct {
	ID               string `json:"id"`
	Status           string `json:"status"`
	Prompt           string `json:"prompt"`
	TokenCount       int    `json:"token_count"`
	PromptTokens     int    `json:"prompt_tokens"`
	CompletionTokens int    `json:"completion_tokens"`
	DurationMs       int64  `json:"duration_ms"`
	ErrorMsg         string `json:"error_msg,omitempty"`
	CreatedAt        string `json:"created_at"`
	CompletedAt      string `json:"completed_at,omitempty"`
}

// LangSmithUsageHistory LangSmith 历史分页。
type LangSmithUsageHistory struct {
	Items      []LangSmithUsageHistoryItem `json:"items"`
	Total      int64                       `json:"total"`
	Page       int                         `json:"page"`
	PageSize   int                         `json:"pageSize"`
	TotalPages int                         `json:"totalPages"`
}

// LangSmithUsagePayload LangSmith Token 使用量响应。
type LangSmithUsagePayload struct {
	Source  string                `json:"source"`
	Project string                `json:"project,omitempty"`
	Stats   LangSmithUsageStats   `json:"stats"`
	History LangSmithUsageHistory `json:"history"`
}
