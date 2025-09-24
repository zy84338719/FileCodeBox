package mcp

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/models"
	"github.com/zy84338719/filecodebox/internal/repository"
	"github.com/zy84338719/filecodebox/internal/services"
	"github.com/zy84338719/filecodebox/internal/storage"
)

// MCPStatus MCP服务器状态
type MCPStatus struct {
	Running    bool       `json:"running"`
	Timestamp  string     `json:"timestamp"`
	ServerInfo ServerInfo `json:"server_info,omitempty"`
}

// MCPManager MCP 服务器管理器
type MCPManager struct {
	manager        *config.ConfigManager
	daoManager     *repository.RepositoryManager
	storageManager *storage.StorageManager
	shareService   *services.ShareService
	adminService   *services.AdminService
	userService    *services.UserService

	server    *FileCodeBoxMCPServer
	running   bool
	mu        sync.RWMutex
	cancelCtx context.CancelFunc
}

// NewMCPManager 创建新的 MCP 管理器
func NewMCPManager(
	manager *config.ConfigManager,
	daoManager *repository.RepositoryManager,
	storageManager *storage.StorageManager,
	shareService *services.ShareService,
	adminService *services.AdminService,
	userService *services.UserService,
) *MCPManager {
	return &MCPManager{
		manager:        manager,
		daoManager:     daoManager,
		storageManager: storageManager,
		shareService:   shareService,
		adminService:   adminService,
		userService:    userService,
		running:        false,
	}
}

// StartMCPServer 启动 MCP 服务器
func (m *MCPManager) StartMCPServer(port string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("MCP 服务器已在运行")
	}

	// 创建 MCP 服务器
	m.server = NewFileCodeBoxMCPServer(
		m.manager,
		m.daoManager,
		m.storageManager,
		m.shareService,
		m.adminService,
		m.userService,
	)

	// 创建取消上下文
	_, cancel := context.WithCancel(context.Background())
	m.cancelCtx = cancel

	// 在后台启动 MCP 服务器
	started := make(chan error, 1)
	go func() {
		logrus.Infof("启动 MCP 服务器，监听端口: %s", port)
		if err := m.server.ServeTCP(":" + port); err != nil {
			logrus.Errorf("MCP 服务器启动失败: %v", err)
			started <- err
			m.mu.Lock()
			m.running = false
			m.mu.Unlock()
			return
		}
		started <- nil
	}()

	// 等待启动结果
	select {
	case err := <-started:
		if err != nil {
			// 如果是端口已被占用错误，检查是否是我们自己的服务器
			if strings.Contains(err.Error(), "address already in use") {
				// 尝试连接测试
				if m.testMCPConnection(port) {
					m.running = true
					logrus.Info("检测到 MCP 服务器已在运行")
					return nil
				}
			}
			return err
		}
		m.running = true
		logrus.Info("MCP 服务器启动成功")
		return nil
	case <-time.After(2 * time.Second):
		// 超时情况下，测试连接
		if m.testMCPConnection(port) {
			m.running = true
			logrus.Info("MCP 服务器启动成功（通过连接测试确认）")
			return nil
		}
		return fmt.Errorf("MCP 服务器启动超时")
	}
}

// StopMCPServer 停止 MCP 服务器
func (m *MCPManager) StopMCPServer() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return fmt.Errorf("MCP 服务器未在运行")
	}

	// 取消上下文
	if m.cancelCtx != nil {
		m.cancelCtx()
	}

	// 这里应该有一个更优雅的关闭方式
	// 但由于当前的 MCP 服务器实现没有提供 Shutdown 方法
	// 我们只能标记为停止状态
	m.running = false
	m.server = nil

	logrus.Info("MCP 服务器已停止")
	return nil
}

// RestartMCPServer 重启 MCP 服务器
func (m *MCPManager) RestartMCPServer(port string) error {
	if m.IsRunning() {
		if err := m.StopMCPServer(); err != nil {
			return fmt.Errorf("停止 MCP 服务器失败: %v", err)
		}
		// 等待一段时间确保服务器完全停止
		time.Sleep(500 * time.Millisecond)
	}
	return m.StartMCPServer(port)
}

// IsRunning 检查 MCP 服务器是否在运行
func (m *MCPManager) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}

// GetStatus 获取 MCP 服务器状态
func (m *MCPManager) GetStatus() MCPStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := MCPStatus{
		Running:   m.running,
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	}

	if m.running && m.server != nil {
		status.ServerInfo = ServerInfo{
			Name:    "FileCodeBox MCP Server",
			Version: models.Version,
		}
	}

	return status
}

// ApplyConfig 应用配置更改
func (m *MCPManager) ApplyConfig(enableMCP bool, port string) error {
	if enableMCP {
		if m.IsRunning() {
			// 如果已经在运行，重启以应用新配置
			return m.RestartMCPServer(port)
		} else {
			// 如果未运行，启动服务器
			return m.StartMCPServer(port)
		}
	} else {
		if m.IsRunning() {
			// 如果在运行但配置为禁用，停止服务器
			return m.StopMCPServer()
		}
		// 如果未运行且配置为禁用，无需操作
		return nil
	}
}

// testMCPConnection 测试 MCP 连接是否可用
func (m *MCPManager) testMCPConnection(port string) bool {
	conn, err := net.DialTimeout("tcp", "localhost:"+port, 3*time.Second)
	if err != nil {
		return false
	}
	defer func() {
		if err := conn.Close(); err != nil {
			// 在这种测试场景下，关闭连接的错误不是关键问题，只记录一下
			logrus.WithError(err).Warn("failed to close MCP test connection")
		}
	}()

	// 简单的连接测试，如果能连接就认为服务器在运行
	return true
}
