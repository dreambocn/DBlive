// 健康检查接口
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Health(c *gin.Context) {
	// 用于探活与监控
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
