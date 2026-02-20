package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	TraceIDHeader   = "X-Trace-ID"
	RequestIDHeader = "X-Request-ID"
)

type traceIDContextKey string

const traceIDCtxKey traceIDContextKey = "trace_id"

// GenerateTraceID 生成全局链路追踪 ID。
func GenerateTraceID() string {
	return uuid.NewString()
}

// ResolveTraceID 从请求头解析 trace_id，优先使用 X-Trace-ID，兼容 X-Request-ID。
func ResolveTraceID(req *http.Request) string {
	if req == nil {
		return GenerateTraceID()
	}

	if traceID := strings.TrimSpace(req.Header.Get(TraceIDHeader)); traceID != "" {
		return traceID
	}
	if requestID := strings.TrimSpace(req.Header.Get(RequestIDHeader)); requestID != "" {
		return requestID
	}

	return GenerateTraceID()
}

// BindTraceID 绑定 trace_id 到 gin context、request context 和响应头。
func BindTraceID(c *gin.Context) string {
	traceID := ResolveTraceID(c.Request)

	c.Set("trace_id", traceID)
	c.Set("request_id", traceID)
	c.Header(TraceIDHeader, traceID)
	c.Header(RequestIDHeader, traceID)
	c.Request = c.Request.WithContext(WithTraceID(c.Request.Context(), traceID))

	return traceID
}

// WithTraceID 将 trace_id 写入 context，供 service 层透传。
func WithTraceID(ctx context.Context, traceID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, traceIDCtxKey, traceID)
}

// TraceIDFromContext 从 context 中读取 trace_id。
func TraceIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	traceID, _ := ctx.Value(traceIDCtxKey).(string)
	return traceID
}

// TraceIDFromGin 从 gin context 中读取 trace_id。
func TraceIDFromGin(c *gin.Context) string {
	if c == nil {
		return ""
	}

	if value, ok := c.Get("trace_id"); ok {
		if traceID, ok := value.(string); ok {
			return traceID
		}
	}

	return ""
}
