package admin

import (
	"encoding/json"
	"fmt"

	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/models/web"
)

// GetConfig 获取配置信息
func (s *Service) GetConfig() *config.ConfigManager {
	return s.manager
}

// UpdateConfig 更新配置
func (s *Service) UpdateConfig(configData map[string]interface{}) error {
	// 过滤掉端口和管理员密码配置，这些不应该通过API更新
	filteredConfigData := make(map[string]interface{})
	for key, value := range configData {
		// 跳过端口和管理员密码配置
		if key == "port" || key == "admin_token" {
			continue
		}
		filteredConfigData[key] = value
	}

	// 扁平化配置数据
	flattenedConfig := make(map[string]interface{})
	for key, value := range filteredConfigData {
		if err := s.flattenConfig(key, value, flattenedConfig); err != nil {
			return fmt.Errorf("处理配置数据失败: %w", err)
		}
	}

	// 保存扁平化的配置
	for key, value := range flattenedConfig {
		// 将value转换为字符串
		var valueStr string
		switch v := value.(type) {
		case string:
			valueStr = v
		case int, int32, int64:
			valueStr = fmt.Sprintf("%d", v)
		case float32, float64:
			valueStr = fmt.Sprintf("%g", v)
		case bool:
			if v {
				valueStr = "1"
			} else {
				valueStr = "0"
			}
		default:
			// 对于复杂类型，序列化为JSON
			jsonBytes, err := json.Marshal(v)
			if err != nil {
				return fmt.Errorf("序列化配置值失败: %w", err)
			}
			valueStr = string(jsonBytes)
		}

		if err := s.repositoryManager.KeyValue.SetValue(key, valueStr); err != nil {
			return fmt.Errorf("保存配置失败: %w", err)
		}
	}

	// 配置保存成功后，执行热重载
	if err := s.manager.ReloadConfig(); err != nil {
		return fmt.Errorf("热重载配置失败: %w", err)
	}

	return nil
}

// UpdateConfigFromRequest 从结构化请求更新配置
func (s *Service) UpdateConfigFromRequest(configRequest *web.AdminConfigRequest) error {
	// 构建配置更新数据
	configUpdates := make(map[string]interface{})

	// 处理基础配置
	if configRequest.Base != nil {
		base := make(map[string]interface{})
		if configRequest.Base.Name != nil {
			base["name"] = *configRequest.Base.Name
		}
		if configRequest.Base.Description != nil {
			base["description"] = *configRequest.Base.Description
		}
		if configRequest.Base.Keywords != nil {
			base["keywords"] = *configRequest.Base.Keywords
		}
		if len(base) > 0 {
			configUpdates["base"] = base
		}
	}

	// 处理传输配置
	if configRequest.Transfer != nil {
		transfer := make(map[string]interface{})

		if configRequest.Transfer.Upload != nil {
			upload := make(map[string]interface{})
			uploadConfig := configRequest.Transfer.Upload
			if uploadConfig.OpenUpload != nil {
				upload["open_upload"] = *uploadConfig.OpenUpload
			}
			if uploadConfig.UploadSize != nil {
				upload["upload_size"] = *uploadConfig.UploadSize
			}
			if uploadConfig.EnableChunk != nil {
				upload["enable_chunk"] = *uploadConfig.EnableChunk
			}
			if uploadConfig.ChunkSize != nil {
				upload["chunk_size"] = *uploadConfig.ChunkSize
			}
			if uploadConfig.MaxSaveSeconds != nil {
				upload["max_save_seconds"] = *uploadConfig.MaxSaveSeconds
			}
			if len(upload) > 0 {
				transfer["upload"] = upload
			}
		}

		if configRequest.Transfer.Download != nil {
			download := make(map[string]interface{})
			downloadConfig := configRequest.Transfer.Download
			if downloadConfig.EnableConcurrentDownload != nil {
				download["enable_concurrent_download"] = *downloadConfig.EnableConcurrentDownload
			}
			if downloadConfig.MaxConcurrentDownloads != nil {
				download["max_concurrent_downloads"] = *downloadConfig.MaxConcurrentDownloads
			}
			if downloadConfig.DownloadTimeout != nil {
				download["download_timeout"] = *downloadConfig.DownloadTimeout
			}
			if len(download) > 0 {
				transfer["download"] = download
			}
		}

		if len(transfer) > 0 {
			configUpdates["transfer"] = transfer
		}
	}

	// 处理用户配置
	if configRequest.User != nil {
		user := make(map[string]interface{})
		userConfig := configRequest.User
		if userConfig.AllowUserRegistration != nil {
			user["allow_user_registration"] = *userConfig.AllowUserRegistration
		}
		if userConfig.RequireEmailVerify != nil {
			user["require_email_verify"] = *userConfig.RequireEmailVerify
		}
		if userConfig.UserUploadSize != nil {
			user["user_upload_size"] = *userConfig.UserUploadSize
		}
		if userConfig.UserStorageQuota != nil {
			user["user_storage_quota"] = *userConfig.UserStorageQuota
		}
		if userConfig.SessionExpiryHours != nil {
			user["session_expiry_hours"] = *userConfig.SessionExpiryHours
		}
		if userConfig.MaxSessionsPerUser != nil {
			user["max_sessions_per_user"] = *userConfig.MaxSessionsPerUser
		}
		if len(user) > 0 {
			configUpdates["user"] = user
		}
	}

	// 处理其他配置
	if configRequest.NotifyTitle != nil {
		configUpdates["notify_title"] = *configRequest.NotifyTitle
	}
	if configRequest.NotifyContent != nil {
		configUpdates["notify_content"] = *configRequest.NotifyContent
	}
	if configRequest.PageExplain != nil {
		configUpdates["page_explain"] = *configRequest.PageExplain
	}
	if configRequest.Opacity != nil {
		configUpdates["opacity"] = *configRequest.Opacity
	}
	if configRequest.ThemesSelect != nil {
		configUpdates["themes_select"] = *configRequest.ThemesSelect
	}

	// 调用原有的更新方法
	return s.UpdateConfig(configUpdates)
}

// flattenConfig 扁平化配置数据
func (s *Service) flattenConfig(prefix string, value interface{}, result map[string]interface{}) error {
	switch v := value.(type) {
	case map[string]interface{}:
		// 对于嵌套的对象，递归处理
		for key, val := range v {
			newKey := key
			if prefix != "" {
				newKey = prefix + "." + key
			}
			if err := s.flattenConfig(newKey, val, result); err != nil {
				return err
			}
		}
	default:
		// 直接的值
		if prefix != "" {
			result[prefix] = value
		}
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
