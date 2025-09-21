package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/zy84338719/filecodebox/internal/common"
)

// 新版 handler 响应示例，直接序列化 ConfigManager

func (h *AdminHandler) GetConfigExample(c *gin.Context) {
	cfg := h.service.GetFullConfig()
	common.SuccessResponse(c, cfg)
}
