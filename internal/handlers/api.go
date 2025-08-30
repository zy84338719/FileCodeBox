package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/zy84338719/filecodebox/internal/common"
)

// APIHandler API处理器
type APIHandler struct{}

func NewAPIHandler() *APIHandler {
	return &APIHandler{}
}

// GetHealth 健康检查
// @Summary 健康检查
// @Description 检查服务器健康状态
// @Tags 系统
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "健康状态信息"
// @Router /health [get]
func (h *APIHandler) GetHealth(c *gin.Context) {
	common.SuccessResponse(c, map[string]interface{}{
		"status":    "ok",
		"timestamp": "2025-08-29",
		"version":   "1.0.0",
	})
}
