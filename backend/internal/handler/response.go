package handler

import (
	"net/http"
	"strconv"

	"lesson-plan/backend/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// APIError 统一错误结构。
type APIError struct {
	Code    string      `json:"code"`
	Details interface{} `json:"details,omitempty"`
}

// Response 标准响应结构
type Response struct {
	Success bool        `json:"success"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
	TraceID string      `json:"trace_id,omitempty"`
}

// PaginatedResponse 分页响应
type PaginatedResponse struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Code:    0,
		Message: "success",
		Data:    data,
		TraceID: middleware.TraceIDFromGin(c),
	})
}

// SuccessWithMessage 带消息的成功响应
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Code:    0,
		Message: message,
		Data:    data,
		TraceID: middleware.TraceIDFromGin(c),
	})
}

// Error 错误响应
func Error(c *gin.Context, statusCode int, message string, data interface{}) {
	ErrorWithCode(c, statusCode, defaultErrorCode(statusCode), message, data)
}

// ErrorWithCode 带业务错误码的错误响应。
func ErrorWithCode(c *gin.Context, statusCode int, errorCode, message string, details interface{}) {
	errPayload := &APIError{
		Code:    errorCode,
		Details: details,
	}
	if details == nil {
		errPayload.Details = nil
	}

	c.JSON(statusCode, Response{
		Success: false,
		Code:    statusCode,
		Message: message,
		Error:   errPayload,
		TraceID: middleware.TraceIDFromGin(c),
	})
}

// BadRequest 400错误
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message, nil)
}

// Unauthorized 401错误
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, message, nil)
}

// Forbidden 403错误
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, message, nil)
}

// NotFound 404错误
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message, nil)
}

// InternalError 500错误
func InternalError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message, nil)
}

// Paginated 分页响应
func Paginated(c *gin.Context, items interface{}, total int64, page, pageSize int) {
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Code:    0,
		Message: "success",
		Data: PaginatedResponse{
			Items:      items,
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		},
		TraceID: middleware.TraceIDFromGin(c),
	})
}

// GetPagination 获取分页参数
func GetPagination(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	return page, pageSize
}

// ParseUUID 解析UUID
func ParseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

// HealthCheck 健康检查
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func defaultErrorCode(statusCode int) string {
	switch statusCode {
	case http.StatusBadRequest:
		return "BAD_REQUEST"
	case http.StatusUnauthorized:
		return "UNAUTHORIZED"
	case http.StatusForbidden:
		return "FORBIDDEN"
	case http.StatusNotFound:
		return "NOT_FOUND"
	case http.StatusConflict:
		return "CONFLICT"
	case http.StatusTooManyRequests:
		return "RATE_LIMITED"
	case http.StatusGatewayTimeout:
		return "GATEWAY_TIMEOUT"
	default:
		if statusCode >= 500 {
			return "INTERNAL_SERVER_ERROR"
		}
		return "UNKNOWN_ERROR"
	}
}
