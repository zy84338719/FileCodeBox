package handlers

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/zy84338719/filecodebox/internal/common"
	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/models/web"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GetMCPConfig 获取 MCP 配置
func (h *AdminHandler) GetMCPConfig(c *gin.Context) {
	common.SuccessResponse(c, h.config.MCP)
}

// UpdateMCPConfig 更新 MCP 配置
func (h *AdminHandler) UpdateMCPConfig(c *gin.Context) {
	var mcpConfig struct {
		EnableMCPServer *int    `json:"enable_mcp_server"`
		MCPPort         *string `json:"mcp_port"`
		MCPHost         *string `json:"mcp_host"`
	}

	if err := c.ShouldBindJSON(&mcpConfig); err != nil {
		common.BadRequestResponse(c, "MCP配置参数错误: "+err.Error())
		return
	}

	start := time.Now()
	if err := h.config.UpdateTransaction(func(draft *config.ConfigManager) error {
		if mcpConfig.EnableMCPServer != nil {
			draft.MCP.EnableMCPServer = *mcpConfig.EnableMCPServer
		}
		if mcpConfig.MCPPort != nil {
			draft.MCP.MCPPort = *mcpConfig.MCPPort
		}
		if mcpConfig.MCPHost != nil {
			draft.MCP.MCPHost = *mcpConfig.MCPHost
		}
		return nil
	}); err != nil {
		h.recordOperationLog(c, "mcp.update_config", "config", false, err.Error(), start)
		common.InternalServerErrorResponse(c, "保存MCP配置失败: "+err.Error())
		return
	}

	mcpManager := GetMCPManager()
	if mcpManager != nil {
		enableMCP := h.config.MCP.EnableMCPServer == 1
		port := h.config.MCP.MCPPort

		if err := mcpManager.ApplyConfig(enableMCP, port); err != nil {
			h.recordOperationLog(c, "mcp.update_config", "config", false, err.Error(), start)
			common.InternalServerErrorResponse(c, "应用MCP配置失败: "+err.Error())
			return
		}
	}

	h.recordOperationLog(c, "mcp.update_config", "config", true, "MCP配置更新成功", start)
	common.SuccessWithMessage(c, "MCP配置更新成功", nil)
}

// GetMCPStatus 获取 MCP 服务器状态
func (h *AdminHandler) GetMCPStatus(c *gin.Context) {
	mcpManager := GetMCPManager()
	if mcpManager == nil {
		common.InternalServerErrorResponse(c, "MCP管理器未初始化")
		return
	}

	status := mcpManager.GetStatus()

	statusText := "inactive"
	if status.Running {
		statusText = "active"
	}

	response := web.MCPStatusResponse{
		Status: statusText,
		Config: h.config.MCP,
	}

	common.SuccessResponse(c, response)
}

// RestartMCPServer 重启 MCP 服务
func (h *AdminHandler) RestartMCPServer(c *gin.Context) {
	mcpManager := GetMCPManager()
	if mcpManager == nil {
		common.InternalServerErrorResponse(c, "MCP管理器未初始化")
		return
	}

	if h.config.MCP.EnableMCPServer != 1 {
		common.BadRequestResponse(c, "MCP服务器未启用")
		return
	}

	start := time.Now()
	if err := mcpManager.RestartMCPServer(h.config.MCP.MCPPort); err != nil {
		h.recordOperationLog(c, "mcp.restart", "mcp", false, err.Error(), start)
		common.InternalServerErrorResponse(c, "重启MCP服务器失败: "+err.Error())
		return
	}

	h.recordOperationLog(c, "mcp.restart", "mcp", true, "MCP服务器重启成功", start)
	common.SuccessWithMessage(c, "MCP服务器重启成功", nil)
}

// ControlMCPServer 控制 MCP 服务的启停
func (h *AdminHandler) ControlMCPServer(c *gin.Context) {
	var controlData struct {
		Action string `json:"action" binding:"required"`
	}

	if err := c.ShouldBindJSON(&controlData); err != nil {
		common.BadRequestResponse(c, "参数错误: "+err.Error())
		return
	}

	mcpManager := GetMCPManager()
	if mcpManager == nil {
		common.InternalServerErrorResponse(c, "MCP管理器未初始化")
		return
	}

	start := time.Now()
	switch controlData.Action {
	case "start":
		if h.config.MCP.EnableMCPServer != 1 {
			common.BadRequestResponse(c, "MCP服务器未启用，请先在配置中启用")
			return
		}
		if err := mcpManager.StartMCPServer(h.config.MCP.MCPPort); err != nil {
			h.recordOperationLog(c, "mcp.start", "mcp", false, err.Error(), start)
			common.InternalServerErrorResponse(c, "启动MCP服务器失败: "+err.Error())
			return
		}
		h.recordOperationLog(c, "mcp.start", "mcp", true, "MCP服务器启动成功", start)
		common.SuccessWithMessage(c, "MCP服务器启动成功", nil)
	case "stop":
		if err := mcpManager.StopMCPServer(); err != nil {
			h.recordOperationLog(c, "mcp.stop", "mcp", false, err.Error(), start)
			common.InternalServerErrorResponse(c, "停止MCP服务器失败: "+err.Error())
			return
		}
		h.recordOperationLog(c, "mcp.stop", "mcp", true, "MCP服务器停止成功", start)
		common.SuccessWithMessage(c, "MCP服务器停止成功", nil)
	default:
		common.BadRequestResponse(c, "无效的操作，只支持 start 或 stop")
	}
}

// TestMCPConnection 测试 MCP 服务连接
func (h *AdminHandler) TestMCPConnection(c *gin.Context) {
	var testData struct {
		Port string `json:"port"`
		Host string `json:"host"`
	}

	if err := c.ShouldBindJSON(&testData); err != nil {
		common.BadRequestResponse(c, "参数错误: "+err.Error())
		return
	}

	address, _, _, err := normalizeMCPAddress(testData.Host, testData.Port, h.config)
	if err != nil {
		common.BadRequestResponse(c, "参数错误: "+err.Error())
		return
	}

	if err := tcpProbe(address, 3*time.Second); err != nil {
		common.ErrorResponse(c, 400, fmt.Sprintf("连接测试失败: %s，端口可能未开放或MCP服务器未启动", err.Error()))
		return
	}

	response := web.MCPTestResponse{
		MCPStatusResponse: web.MCPStatusResponse{
			Status: "连接正常",
			Config: h.config.MCP,
		},
	}

	common.SuccessWithMessage(c, "MCP连接测试成功", response)
}

func normalizeMCPAddress(reqHost, reqPort string, cfg *config.ConfigManager) (address, host, port string, err error) {
	port = reqPort
	if port == "" {
		port = cfg.MCP.MCPPort
	}
	if port == "" {
		port = "8081"
	}

	pnum, perr := strconv.Atoi(port)
	if perr != nil || pnum < 1 || pnum > 65535 {
		err = fmt.Errorf("无效端口号: %s", port)
		return
	}

	host = reqHost
	if host == "" {
		host = cfg.MCP.MCPHost
	}
	if host == "" {
		host = "0.0.0.0"
	}

	address = host + ":" + port
	if host == "0.0.0.0" {
		address = "127.0.0.1:" + port
	}
	return
}

func tcpProbe(address string, timeout time.Duration) error {
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return err
	}
	defer func() {
		if err := conn.Close(); err != nil {
			logrus.WithError(err).Warn("关闭 TCP 连接失败")
		}
	}()
	return nil
}
