package handlers

import (
	"sync"

	"github.com/zy84338719/filecodebox/internal/mcp"
)

// AppState 应用状态管理器
type AppState struct {
	mcpManager   *mcp.MCPManager
	adminHandler *AdminHandler
	mu           sync.RWMutex
}

var appState = &AppState{}

// SetMCPManager 设置 MCP 管理器
func SetMCPManager(manager *mcp.MCPManager) {
	appState.mu.Lock()
	defer appState.mu.Unlock()
	appState.mcpManager = manager
}

// GetMCPManager 获取 MCP 管理器
func GetMCPManager() *mcp.MCPManager {
	appState.mu.RLock()
	defer appState.mu.RUnlock()
	return appState.mcpManager
}

// SetInjectedAdminHandler 注入 AdminHandler，供占位路由委派使用
func SetInjectedAdminHandler(h *AdminHandler) {
	appState.mu.Lock()
	defer appState.mu.Unlock()
	appState.adminHandler = h
}

// GetInjectedAdminHandler 获取注入的 AdminHandler
func GetInjectedAdminHandler() *AdminHandler {
	appState.mu.RLock()
	defer appState.mu.RUnlock()
	return appState.adminHandler
}
