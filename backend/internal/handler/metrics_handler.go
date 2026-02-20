package handler

import (
	"lesson-plan/backend/internal/observability"

	"github.com/gin-gonic/gin"
)

// Metrics 返回服务运行指标快照。
func Metrics(c *gin.Context) {
	Success(c, observability.SnapshotMetrics())
}
