// Package config MCP配置模块
package config

import (
	"fmt"
	"strconv"
	"strings"
)

// MCPConfig MCP服务器配置
type MCPConfig struct {
	EnableMCPServer int    `json:"enable_mcp_server"` // 是否启用 MCP 服务器
	MCPPort         string `json:"mcp_port"`          // MCP 服务器端口
	MCPHost         string `json:"mcp_host"`          // MCP 服务器绑定地址
}

// NewMCPConfig 创建MCP配置
func NewMCPConfig() *MCPConfig {
	return &MCPConfig{
		EnableMCPServer: 0,         // 默认禁用
		MCPPort:         "8081",    // 默认端口
		MCPHost:         "0.0.0.0", // 默认绑定所有IP
	}
}

// Validate 验证MCP配置
func (mc *MCPConfig) Validate() error {
	var errors []string

	// 验证端口号
	if port, err := strconv.Atoi(mc.MCPPort); err != nil {
		errors = append(errors, "MCP端口号必须是有效数字")
	} else {
		if port < 1 || port > 65535 {
			errors = append(errors, "MCP端口号必须在1-65535之间")
		}
	}

	// 验证主机地址
	if strings.TrimSpace(mc.MCPHost) == "" {
		errors = append(errors, "MCP主机地址不能为空")
	}

	if len(errors) > 0 {
		return fmt.Errorf("MCP配置验证失败: %s", strings.Join(errors, "; "))
	}

	return nil
}

// IsMCPEnabled 判断是否启用MCP服务器
func (mc *MCPConfig) IsMCPEnabled() bool {
	return mc.EnableMCPServer == 1
}

// Update 更新配置
// Clone 克隆配置
func (mc *MCPConfig) Clone() *MCPConfig {
	return &MCPConfig{
		EnableMCPServer: mc.EnableMCPServer,
		MCPPort:         mc.MCPPort,
		MCPHost:         mc.MCPHost,
	}
}
