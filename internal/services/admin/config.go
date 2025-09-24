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
	return s.manager.UpdateTransaction(func(draft *config.ConfigManager) error {
		ensureUI := func(cfg *config.ConfigManager) *config.UIConfig {
			if cfg.UI == nil {
				cfg.UI = &config.UIConfig{}
			}
			return cfg.UI
		}

		if configRequest.Base != nil {
			baseConfig := configRequest.Base
			if baseConfig.Name != "" {
				draft.Base.Name = baseConfig.Name
			}
			if baseConfig.Description != "" {
				draft.Base.Description = baseConfig.Description
			}
			if baseConfig.Keywords != "" {
				draft.Base.Keywords = baseConfig.Keywords
			}
			if baseConfig.Port != 0 {
				draft.Base.Port = baseConfig.Port
			}
			if baseConfig.Host != "" {
				draft.Base.Host = baseConfig.Host
			}
			if baseConfig.DataPath != "" {
				draft.Base.DataPath = baseConfig.DataPath
			}
			draft.Base.Production = baseConfig.Production
		}

		if configRequest.Database != nil {
			dbConfig := configRequest.Database
			if dbConfig.Type != "" {
				draft.Database.Type = dbConfig.Type
			}
			if dbConfig.Host != "" {
				draft.Database.Host = dbConfig.Host
			}
			if dbConfig.Port != 0 {
				draft.Database.Port = dbConfig.Port
			}
			if dbConfig.Name != "" {
				draft.Database.Name = dbConfig.Name
			}
			if dbConfig.User != "" {
				draft.Database.User = dbConfig.User
			}
			if dbConfig.Pass != "" {
				draft.Database.Pass = dbConfig.Pass
			}
			if dbConfig.SSL != "" {
				draft.Database.SSL = dbConfig.SSL
			}
		}

		if configRequest.Transfer != nil {
			if configRequest.Transfer.Upload != nil {
				uploadConfig := configRequest.Transfer.Upload
				draft.Transfer.Upload.OpenUpload = uploadConfig.OpenUpload
				draft.Transfer.Upload.UploadSize = uploadConfig.UploadSize
				draft.Transfer.Upload.EnableChunk = uploadConfig.EnableChunk
				draft.Transfer.Upload.ChunkSize = uploadConfig.ChunkSize
				draft.Transfer.Upload.MaxSaveSeconds = uploadConfig.MaxSaveSeconds
				draft.Transfer.Upload.RequireLogin = uploadConfig.RequireLogin
			}

			if configRequest.Transfer.Download != nil {
				downloadConfig := configRequest.Transfer.Download
				draft.Transfer.Download.EnableConcurrentDownload = downloadConfig.EnableConcurrentDownload
				draft.Transfer.Download.MaxConcurrentDownloads = downloadConfig.MaxConcurrentDownloads
				draft.Transfer.Download.DownloadTimeout = downloadConfig.DownloadTimeout
				draft.Transfer.Download.RequireLogin = downloadConfig.RequireLogin
			}
		}

		if configRequest.Storage != nil {
			storageConfig := configRequest.Storage
			if storageConfig.Type != "" {
				draft.Storage.Type = storageConfig.Type
			}
			if storageConfig.StoragePath != "" {
				draft.Storage.StoragePath = storageConfig.StoragePath
			}
			if storageConfig.S3 != nil {
				draft.Storage.S3 = storageConfig.S3
			}
			if storageConfig.WebDAV != nil {
				draft.Storage.WebDAV = storageConfig.WebDAV
			}
			if storageConfig.OneDrive != nil {
				draft.Storage.OneDrive = storageConfig.OneDrive
			}
			if storageConfig.NFS != nil {
				draft.Storage.NFS = storageConfig.NFS
			}
		}

		if configRequest.User != nil {
			userConfig := configRequest.User
			draft.User.AllowUserRegistration = userConfig.AllowUserRegistration
			draft.User.RequireEmailVerify = userConfig.RequireEmailVerify
			if userConfig.UserStorageQuota != 0 {
				draft.User.UserStorageQuota = userConfig.UserStorageQuota
			}
			if userConfig.UserUploadSize != 0 {
				draft.User.UserUploadSize = userConfig.UserUploadSize
			}
			if userConfig.SessionExpiryHours != 0 {
				draft.User.SessionExpiryHours = userConfig.SessionExpiryHours
			}
			if userConfig.MaxSessionsPerUser != 0 {
				draft.User.MaxSessionsPerUser = userConfig.MaxSessionsPerUser
			}
			if userConfig.JWTSecret != "" {
				draft.User.JWTSecret = userConfig.JWTSecret
			}
		}

		if configRequest.MCP != nil {
			mcpConfig := configRequest.MCP
			draft.MCP.EnableMCPServer = mcpConfig.EnableMCPServer
			if mcpConfig.MCPPort != "" {
				draft.MCP.MCPPort = mcpConfig.MCPPort
			}
			if mcpConfig.MCPHost != "" {
				draft.MCP.MCPHost = mcpConfig.MCPHost
			}
		}

		if configRequest.UI != nil {
			uiConfig := configRequest.UI
			ui := ensureUI(draft)
			if strings.TrimSpace(uiConfig.ThemesSelect) != "" {
				ui.ThemesSelect = uiConfig.ThemesSelect
			}
			ui.PageExplain = uiConfig.PageExplain
			ui.Opacity = uiConfig.Opacity
		}

		if configRequest.NotifyTitle != nil {
			draft.NotifyTitle = *configRequest.NotifyTitle
		}
		if configRequest.NotifyContent != nil {
			draft.NotifyContent = *configRequest.NotifyContent
		}

		if configRequest.SysStart != nil {
			draft.SysStart = *configRequest.SysStart
		}

		return nil
	})
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
