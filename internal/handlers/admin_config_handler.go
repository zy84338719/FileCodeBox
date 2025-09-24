package handlers

import (
	"github.com/zy84338719/filecodebox/internal/common"
	"github.com/zy84338719/filecodebox/internal/models/web"

	"github.com/gin-gonic/gin"
)

// GetConfig 获取配置
func (h *AdminHandler) GetConfig(c *gin.Context) {
	cfg := h.service.GetFullConfig()
	resp := web.AdminConfigResponse{
		AdminConfigRequest: web.AdminConfigRequest{
			Base:     cfg.Base,
			Database: cfg.Database,
			Transfer: cfg.Transfer,
			Storage:  cfg.Storage,
			User:     cfg.User,
			MCP:      cfg.MCP,
			UI:       cfg.UI,
			SysStart: &cfg.SysStart,
		},
	}

	resp.NotifyTitle = &cfg.NotifyTitle
	resp.NotifyContent = &cfg.NotifyContent
	common.SuccessResponse(c, resp)
}

// UpdateConfig 更新配置
func (h *AdminHandler) UpdateConfig(c *gin.Context) {
	var req web.AdminConfigRequest
	if err := c.ShouldBind(&req); err != nil {
		common.BadRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}

	if err := h.service.UpdateConfigFromRequest(&req); err != nil {
		common.InternalServerErrorResponse(c, "更新配置失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, "更新成功", nil)
}
