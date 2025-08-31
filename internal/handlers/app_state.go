package handlers

import (
	"sync"

	"github.com/zy84338719/filecodebox/internal/mcp"
)

// AppState 应用状态管理器
type AppState struct {
	mcpManager *mcp.MCPManager
	mu         sync.RWMutex
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
