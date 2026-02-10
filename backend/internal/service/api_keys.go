package service

import (
	"context"
	"strings"
)

const (
	HeaderGenerationAPIKey = "X-Generation-Api-Key"
	HeaderEmbeddingAPIKey  = "X-Embedding-Api-Key"
)

// APIKeyOverride 用户请求级 API Key 覆盖配置。
// 为空时使用服务端默认配置。
type APIKeyOverride struct {
	GenerationAPIKey string
	EmbeddingAPIKey  string
}

func NewAPIKeyOverride(generationAPIKey, embeddingAPIKey string) APIKeyOverride {
	return APIKeyOverride{
		GenerationAPIKey: strings.TrimSpace(generationAPIKey),
		EmbeddingAPIKey:  strings.TrimSpace(embeddingAPIKey),
	}
}

type apiKeyOverrideContextKey string

const requestAPIKeyOverrideKey apiKeyOverrideContextKey = "request_api_key_override"

func WithAPIKeyOverride(ctx context.Context, override APIKeyOverride) context.Context {
	return context.WithValue(ctx, requestAPIKeyOverrideKey, override)
}

func APIKeyOverrideFromContext(ctx context.Context) APIKeyOverride {
	if ctx == nil {
		return APIKeyOverride{}
	}

	override, ok := ctx.Value(requestAPIKeyOverrideKey).(APIKeyOverride)
	if !ok {
		return APIKeyOverride{}
	}

	return override
}
