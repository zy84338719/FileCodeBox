package admin

import (
	"fmt"
	"strings"

	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/models/web"
)

// GetConfig 获取配置信息
func (s *Service) GetConfig() *config.ConfigManager {
	return s.manager
}

// UpdateConfig 更新配置 - 已弃用，保留向后兼容
func (s *Service) UpdateConfig(configData map[string]interface{}) error {
	// 这个方法保留用于向后兼容，但不再建议使用
	// 新代码应该使用 UpdateConfigFromRequest
	return fmt.Errorf("deprecated: use UpdateConfigFromRequest instead")
}

// UpdateConfigFromRequest 从结构化请求更新配置
func (s *Service) UpdateConfigFromRequest(configRequest *web.AdminConfigRequest) error {
	// 直接更新配置管理器的各个模块，不使用 map 转换
	ensureUI := func() *config.UIConfig {
		if s.manager.UI == nil {
			s.manager.UI = &config.UIConfig{}
		}
		return s.manager.UI
	}

	// 处理基础配置
	if configRequest.Base != nil {
		baseConfig := configRequest.Base
		if baseConfig.Name != "" {
			s.manager.Base.Name = baseConfig.Name
		}
		if baseConfig.Description != "" {
			s.manager.Base.Description = baseConfig.Description
		}
		if baseConfig.Keywords != "" {
			s.manager.Base.Keywords = baseConfig.Keywords
		}
		if baseConfig.Port != 0 {
			s.manager.Base.Port = baseConfig.Port
		}
		if baseConfig.Host != "" {
			s.manager.Base.Host = baseConfig.Host
		}
		if baseConfig.DataPath != "" {
			s.manager.Base.DataPath = baseConfig.DataPath
		}
		s.manager.Base.Production = baseConfig.Production
	}

	// 处理数据库配置
	if configRequest.Database != nil {
		dbConfig := configRequest.Database
		if dbConfig.Type != "" {
			s.manager.Database.Type = dbConfig.Type
		}
		if dbConfig.Host != "" {
			s.manager.Database.Host = dbConfig.Host
		}
		if dbConfig.Port != 0 {
			s.manager.Database.Port = dbConfig.Port
		}
		if dbConfig.Name != "" {
			s.manager.Database.Name = dbConfig.Name
		}
		if dbConfig.User != "" {
			s.manager.Database.User = dbConfig.User
		}
		if dbConfig.Pass != "" {
			s.manager.Database.Pass = dbConfig.Pass
		}
		if dbConfig.SSL != "" {
			s.manager.Database.SSL = dbConfig.SSL
		}
	}

	// 处理传输配置
	if configRequest.Transfer != nil {
		if configRequest.Transfer.Upload != nil {
			uploadConfig := configRequest.Transfer.Upload
			s.manager.Transfer.Upload.OpenUpload = uploadConfig.OpenUpload
			s.manager.Transfer.Upload.UploadSize = uploadConfig.UploadSize
			s.manager.Transfer.Upload.EnableChunk = uploadConfig.EnableChunk
			s.manager.Transfer.Upload.ChunkSize = uploadConfig.ChunkSize
			s.manager.Transfer.Upload.MaxSaveSeconds = uploadConfig.MaxSaveSeconds
		}

		if configRequest.Transfer.Download != nil {
			downloadConfig := configRequest.Transfer.Download
			s.manager.Transfer.Download.EnableConcurrentDownload = downloadConfig.EnableConcurrentDownload
			s.manager.Transfer.Download.MaxConcurrentDownloads = downloadConfig.MaxConcurrentDownloads
			s.manager.Transfer.Download.DownloadTimeout = downloadConfig.DownloadTimeout
		}
	}

	// 处理存储配置
	if configRequest.Storage != nil {
		storageConfig := configRequest.Storage
		if storageConfig.Type != "" {
			s.manager.Storage.Type = storageConfig.Type
		}
		if storageConfig.StoragePath != "" {
			s.manager.Storage.StoragePath = storageConfig.StoragePath
		}
		if storageConfig.S3 != nil {
			s.manager.Storage.S3 = storageConfig.S3
		}
		if storageConfig.WebDAV != nil {
			s.manager.Storage.WebDAV = storageConfig.WebDAV
		}
		if storageConfig.OneDrive != nil {
			s.manager.Storage.OneDrive = storageConfig.OneDrive
		}
		if storageConfig.NFS != nil {
			s.manager.Storage.NFS = storageConfig.NFS
		}
	}

	// 处理用户系统配置
	if configRequest.User != nil {
		userConfig := configRequest.User
		s.manager.User.AllowUserRegistration = userConfig.AllowUserRegistration
		s.manager.User.RequireEmailVerify = userConfig.RequireEmailVerify
		if userConfig.UserStorageQuota != 0 {
			s.manager.User.UserStorageQuota = userConfig.UserStorageQuota
		}
		if userConfig.UserUploadSize != 0 {
			s.manager.User.UserUploadSize = userConfig.UserUploadSize
		}
		if userConfig.SessionExpiryHours != 0 {
			s.manager.User.SessionExpiryHours = userConfig.SessionExpiryHours
		}
		if userConfig.MaxSessionsPerUser != 0 {
			s.manager.User.MaxSessionsPerUser = userConfig.MaxSessionsPerUser
		}
		if userConfig.JWTSecret != "" {
			s.manager.User.JWTSecret = userConfig.JWTSecret
		}
	}

	// 处理 MCP 配置
	if configRequest.MCP != nil {
		mcpConfig := configRequest.MCP
		s.manager.MCP.EnableMCPServer = mcpConfig.EnableMCPServer
		if mcpConfig.MCPPort != "" {
			s.manager.MCP.MCPPort = mcpConfig.MCPPort
		}
		if mcpConfig.MCPHost != "" {
			s.manager.MCP.MCPHost = mcpConfig.MCPHost
		}
	}

	// 处理 UI 配置
	if configRequest.UI != nil {
		uiConfig := configRequest.UI
		ui := ensureUI()
		if strings.TrimSpace(uiConfig.ThemesSelect) != "" {
			ui.ThemesSelect = uiConfig.ThemesSelect
		}
		ui.PageExplain = uiConfig.PageExplain
		ui.Opacity = uiConfig.Opacity
	}

	// 顶层通知字段
	if configRequest.NotifyTitle != nil {
		s.manager.NotifyTitle = *configRequest.NotifyTitle
	}
	if configRequest.NotifyContent != nil {
		s.manager.NotifyContent = *configRequest.NotifyContent
	}

	// 处理系统运行时字段
	if configRequest.SysStart != nil {
		s.manager.SysStart = *configRequest.SysStart
	}

	if err := s.manager.PersistYAML(); err != nil {
		return fmt.Errorf("persist config: %w", err)
	}
	return nil
}

// GetFullConfig 获取完整配置 - 返回配置管理器结构体
func (s *Service) GetFullConfig() *config.ConfigManager {
	// 直接返回配置管理器的克隆，保护原始配置不被修改
	return s.manager.Clone()
}

// GetStorageConfig 获取存储配置
func (s *Service) GetStorageConfig() *config.StorageConfig {
	return s.manager.Storage
}

// GetUserConfig 获取用户配置
func (s *Service) GetUserConfig() *config.UserSystemConfig {
	return s.manager.User
}

// GetMCPConfig 获取MCP配置
func (s *Service) GetMCPConfig() *config.MCPConfig {
	return s.manager.MCP
}

// ValidateConfig 验证配置
func (s *Service) ValidateConfig() error {
	return s.manager.Validate()
}

// ReloadConfig 重新加载配置
func (s *Service) ReloadConfig() error {
	return s.manager.ReloadConfig()
}
