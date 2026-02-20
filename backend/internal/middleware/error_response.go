package middleware

import "github.com/gin-gonic/gin"

type middlewareErrorPayload struct {
	Code    string      `json:"code"`
	Details interface{} `json:"details,omitempty"`
}

type middlewareErrorResponse struct {
	Success bool                    `json:"success"`
	Code    int                     `json:"code"`
	Message string                  `json:"message"`
	Error   *middlewareErrorPayload `json:"error,omitempty"`
	TraceID string                  `json:"trace_id,omitempty"`
}

func abortWithError(c *gin.Context, statusCode int, errorCode, message string, details interface{}) {
	payload := &middlewareErrorPayload{
		Code:    errorCode,
		Details: details,
	}
	if details == nil {
		payload.Details = nil
	}

	c.AbortWithStatusJSON(statusCode, middlewareErrorResponse{
		Success: false,
		Code:    statusCode,
		Message: message,
		Error:   payload,
		TraceID: TraceIDFromGin(c),
	})
}
