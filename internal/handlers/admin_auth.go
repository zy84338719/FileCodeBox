package handlers

import (
	"github.com/zy84338719/filecodebox/internal/common"
	"github.com/zy84338719/filecodebox/internal/models/web"
	"github.com/zy84338719/filecodebox/internal/utils"

	"github.com/gin-gonic/gin"
)

// Login 管理员登录
func (h *AdminHandler) Login(c *gin.Context) {
	var req web.AdminLoginRequest
	if !utils.BindJSONWithValidation(c, &req) {
		return
	}

	tokenString, err := h.service.GenerateTokenForAdmin(req.Username, req.Password)
	if err != nil {
		common.UnauthorizedResponse(c, "认证失败: "+err.Error())
		return
	}

	response := web.AdminLoginResponse{
		Token:     tokenString,
		TokenType: "Bearer",
		ExpiresIn: 24 * 60 * 60,
	}

	common.SuccessWithMessage(c, "登录成功", response)
}

// Dashboard 仪表盘
func (h *AdminHandler) Dashboard(c *gin.Context) {
	stats, err := h.service.GetStats()
	if err != nil {
		common.InternalServerErrorResponse(c, "获取统计信息失败: "+err.Error())
		return
	}

	common.SuccessResponse(c, stats)
}

// GetStats 获取统计信息
func (h *AdminHandler) GetStats(c *gin.Context) {
	stats, err := h.service.GetStats()
	if err != nil {
		common.InternalServerErrorResponse(c, "获取统计信息失败: "+err.Error())
		return
	}

	common.SuccessResponse(c, stats)
}
