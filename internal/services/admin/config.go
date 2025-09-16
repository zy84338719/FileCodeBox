package admin

import (
	"encoding/json"
	"fmt"

	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/models"
	"github.com/zy84338719/filecodebox/internal/models/web"
)

// GetConfig 获取配置信息
func (s *Service) GetConfig() *config.ConfigManager {
	return s.manager
}

// UpdateConfig 更新配置 - 使用结构化DTO
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

	// 转换为结构化配置更新
	configUpdates := s.convertMapToConfigUpdate(filteredConfigData)

	// 保存配置更新
	return s.SaveConfigUpdate(configUpdates)
}

// UpdateConfigWithDTO 使用DTO更新配置
func (s *Service) UpdateConfigWithDTO(configUpdate *models.ConfigUpdateFields) error {
	return s.SaveConfigUpdate(configUpdate)
}

// UpdateConfigWithFlatDTO 使用平面化DTO更新配置
func (s *Service) UpdateConfigWithFlatDTO(flatUpdate *models.FlatConfigUpdate) error {
	configUpdate := s.convertFlatDTOToNested(flatUpdate)
	return s.SaveConfigUpdate(configUpdate)
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
		if userConfig.JWTSecret != nil {
			user["jwt_secret"] = *userConfig.JWTSecret
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

// convertMapToConfigUpdate 将map转换为配置更新DTO
func (s *Service) convertMapToConfigUpdate(data map[string]interface{}) *models.ConfigUpdateFields {
	configUpdate := &models.ConfigUpdateFields{}

	// 用户配置字段映射
	userFields := map[string]bool{
		"allow_user_registration": true,
		"require_email_verify":    true,
		"user_upload_size":        true,
		"user_storage_quota":      true,
		"session_expiry_hours":    true,
		"max_sessions_per_user":   true,
		"jwt_secret":              true,
	}

	// 传输配置字段映射
	transferUploadFields := map[string]bool{
		"open_upload":      true,
		"upload_size":      true,
		"enable_chunk":     true,
		"chunk_size":       true,
		"max_save_seconds": true,
	}

	transferDownloadFields := map[string]bool{
		"enable_concurrent_download": true,
		"max_concurrent_downloads":   true,
		"download_timeout":           true,
	}

	// 基础配置字段映射
	baseFields := map[string]bool{
		"name":        true,
		"description": true,
		"keywords":    true,
		"port":        true,
		"host":        true,
		"data_path":   true,
		"production":  true,
	}

	// MCP配置字段映射
	mcpFields := map[string]bool{
		"enable_mcp_server": true,
		"mcp_port":          true,
		"mcp_host":          true,
	}

	// 分类处理字段
	var userConfig *models.UserConfigUpdate
	var uploadConfig *models.UploadConfigUpdate
	var downloadConfig *models.DownloadConfigUpdate
	var baseConfig *models.BaseConfigUpdate
	var mcpConfig *models.MCPConfigUpdate

	for key, value := range data {
		if userFields[key] {
			if userConfig == nil {
				userConfig = &models.UserConfigUpdate{}
			}
			s.setUserConfigField(userConfig, key, value)
		} else if transferUploadFields[key] {
			if uploadConfig == nil {
				uploadConfig = &models.UploadConfigUpdate{}
			}
			s.setUploadConfigField(uploadConfig, key, value)
		} else if transferDownloadFields[key] {
			if downloadConfig == nil {
				downloadConfig = &models.DownloadConfigUpdate{}
			}
			s.setDownloadConfigField(downloadConfig, key, value)
		} else if baseFields[key] {
			if baseConfig == nil {
				baseConfig = &models.BaseConfigUpdate{}
			}
			s.setBaseConfigField(baseConfig, key, value)
		} else if mcpFields[key] {
			if mcpConfig == nil {
				mcpConfig = &models.MCPConfigUpdate{}
			}
			s.setMCPConfigField(mcpConfig, key, value)
		} else {
			// 其他字段直接设置
			s.setOtherConfigField(configUpdate, key, value)
		}
	}

	// 构建嵌套结构
	if userConfig != nil {
		configUpdate.User = userConfig
	}

	if uploadConfig != nil || downloadConfig != nil {
		transferConfig := &models.TransferConfigUpdate{}
		if uploadConfig != nil {
			transferConfig.Upload = uploadConfig
		}
		if downloadConfig != nil {
			transferConfig.Download = downloadConfig
		}
		configUpdate.Transfer = transferConfig
	}

	if baseConfig != nil {
		configUpdate.Base = baseConfig
	}

	if mcpConfig != nil {
		configUpdate.MCP = mcpConfig
	}

	return configUpdate
}

// convertFlatDTOToNested 将平面化DTO转换为嵌套DTO
func (s *Service) convertFlatDTOToNested(flatUpdate *models.FlatConfigUpdate) *models.ConfigUpdateFields {
	configUpdate := &models.ConfigUpdateFields{}

	// 基础配置
	if hasBaseConfig(flatUpdate) {
		configUpdate.Base = &models.BaseConfigUpdate{
			Name:        flatUpdate.Name,
			Description: flatUpdate.Description,
			Keywords:    flatUpdate.Keywords,
			Port:        flatUpdate.Port,
			Host:        flatUpdate.Host,
			DataPath:    flatUpdate.DataPath,
			Production:  flatUpdate.Production,
		}
	}

	// 传输配置
	if hasTransferConfig(flatUpdate) {
		transferConfig := &models.TransferConfigUpdate{}

		if hasUploadConfig(flatUpdate) {
			transferConfig.Upload = &models.UploadConfigUpdate{
				OpenUpload:     flatUpdate.OpenUpload,
				UploadSize:     flatUpdate.UploadSize,
				EnableChunk:    flatUpdate.EnableChunk,
				ChunkSize:      flatUpdate.ChunkSize,
				MaxSaveSeconds: flatUpdate.MaxSaveSeconds,
			}
		}

		if hasDownloadConfig(flatUpdate) {
			transferConfig.Download = &models.DownloadConfigUpdate{
				EnableConcurrentDownload: flatUpdate.EnableConcurrentDownload,
				MaxConcurrentDownloads:   flatUpdate.MaxConcurrentDownloads,
				DownloadTimeout:          flatUpdate.DownloadTimeout,
			}
		}

		configUpdate.Transfer = transferConfig
	}

	// 用户配置
	if hasUserConfig(flatUpdate) {
		configUpdate.User = &models.UserConfigUpdate{
			AllowUserRegistration: flatUpdate.AllowUserRegistration,
			RequireEmailVerify:    flatUpdate.RequireEmailVerify,
			UserUploadSize:        flatUpdate.UserUploadSize,
			UserStorageQuota:      flatUpdate.UserStorageQuota,
			SessionExpiryHours:    flatUpdate.SessionExpiryHours,
			MaxSessionsPerUser:    flatUpdate.MaxSessionsPerUser,
			JWTSecret:             flatUpdate.JWTSecret,
		}
	}

	// MCP配置
	if hasMCPConfig(flatUpdate) {
		configUpdate.MCP = &models.MCPConfigUpdate{
			EnableMCPServer: flatUpdate.EnableMCPServer,
			MCPPort:         flatUpdate.MCPPort,
			MCPHost:         flatUpdate.MCPHost,
		}
	}

	// 其他配置
	configUpdate.NotifyTitle = flatUpdate.NotifyTitle
	configUpdate.NotifyContent = flatUpdate.NotifyContent
	configUpdate.PageExplain = flatUpdate.PageExplain
	configUpdate.Opacity = flatUpdate.Opacity
	configUpdate.ThemesSelect = flatUpdate.ThemesSelect

	return configUpdate
}

// SaveConfigUpdate 保存配置更新
func (s *Service) SaveConfigUpdate(configUpdate *models.ConfigUpdateFields) error {
	// 转换为map格式
	configMap := configUpdate.ToMap()

	// 扁平化配置数据
	flattenedConfig := make(map[string]interface{})
	for key, value := range configMap {
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

		if err := s.manager.UpdateKeyValue(key, valueStr); err != nil {
			return fmt.Errorf("保存配置失败: %w", err)
		}
	}

	// 配置保存成功后，执行热重载
	if err := s.manager.ReloadConfig(); err != nil {
		return fmt.Errorf("热重载配置失败: %w", err)
	}

	return nil
}

// 辅助方法：设置用户配置字段
func (s *Service) setUserConfigField(config *models.UserConfigUpdate, key string, value interface{}) {
	switch key {
	case "allow_user_registration":
		if v, ok := value.(int); ok {
			config.AllowUserRegistration = &v
		}
	case "require_email_verify":
		if v, ok := value.(int); ok {
			config.RequireEmailVerify = &v
		}
	case "user_upload_size":
		if v, ok := value.(int64); ok {
			config.UserUploadSize = &v
		}
	case "user_storage_quota":
		if v, ok := value.(int64); ok {
			config.UserStorageQuota = &v
		}
	case "session_expiry_hours":
		if v, ok := value.(int); ok {
			config.SessionExpiryHours = &v
		}
	case "max_sessions_per_user":
		if v, ok := value.(int); ok {
			config.MaxSessionsPerUser = &v
		}
	case "jwt_secret":
		if v, ok := value.(string); ok {
			config.JWTSecret = &v
		}
	}
}

// 辅助方法：设置上传配置字段
func (s *Service) setUploadConfigField(config *models.UploadConfigUpdate, key string, value interface{}) {
	switch key {
	case "open_upload":
		if v, ok := value.(int); ok {
			config.OpenUpload = &v
		}
	case "upload_size":
		if v, ok := value.(int64); ok {
			config.UploadSize = &v
		}
	case "enable_chunk":
		if v, ok := value.(int); ok {
			config.EnableChunk = &v
		}
	case "chunk_size":
		if v, ok := value.(int64); ok {
			config.ChunkSize = &v
		}
	case "max_save_seconds":
		if v, ok := value.(int); ok {
			config.MaxSaveSeconds = &v
		}
	}
}

// 辅助方法：设置下载配置字段
func (s *Service) setDownloadConfigField(config *models.DownloadConfigUpdate, key string, value interface{}) {
	switch key {
	case "enable_concurrent_download":
		if v, ok := value.(int); ok {
			config.EnableConcurrentDownload = &v
		}
	case "max_concurrent_downloads":
		if v, ok := value.(int); ok {
			config.MaxConcurrentDownloads = &v
		}
	case "download_timeout":
		if v, ok := value.(int); ok {
			config.DownloadTimeout = &v
		}
	}
}

// 辅助方法：设置基础配置字段
func (s *Service) setBaseConfigField(config *models.BaseConfigUpdate, key string, value interface{}) {
	switch key {
	case "name":
		if v, ok := value.(string); ok {
			config.Name = &v
		}
	case "description":
		if v, ok := value.(string); ok {
			config.Description = &v
		}
	case "keywords":
		if v, ok := value.(string); ok {
			config.Keywords = &v
		}
	case "port":
		if v, ok := value.(int); ok {
			config.Port = &v
		}
	case "host":
		if v, ok := value.(string); ok {
			config.Host = &v
		}
	case "data_path":
		if v, ok := value.(string); ok {
			config.DataPath = &v
		}
	case "production":
		if v, ok := value.(bool); ok {
			config.Production = &v
		}
	}
}

// 辅助方法：设置MCP配置字段
func (s *Service) setMCPConfigField(config *models.MCPConfigUpdate, key string, value interface{}) {
	switch key {
	case "enable_mcp_server":
		if v, ok := value.(int); ok {
			config.EnableMCPServer = &v
		}
	case "mcp_port":
		if v, ok := value.(string); ok {
			config.MCPPort = &v
		}
	case "mcp_host":
		if v, ok := value.(string); ok {
			config.MCPHost = &v
		}
	}
}

// 辅助方法：设置其他配置字段
func (s *Service) setOtherConfigField(config *models.ConfigUpdateFields, key string, value interface{}) {
	switch key {
	case "notify_title":
		if v, ok := value.(string); ok {
			config.NotifyTitle = &v
		}
	case "notify_content":
		if v, ok := value.(string); ok {
			config.NotifyContent = &v
		}
	case "page_explain":
		if v, ok := value.(string); ok {
			config.PageExplain = &v
		}
	case "opacity":
		if v, ok := value.(int); ok {
			config.Opacity = &v
		}
	case "themes_select":
		if v, ok := value.(string); ok {
			config.ThemesSelect = &v
		}
	}
}

// 辅助方法：检查是否有基础配置
func hasBaseConfig(flat *models.FlatConfigUpdate) bool {
	return flat.Name != nil || flat.Description != nil || flat.Keywords != nil ||
		flat.Port != nil || flat.Host != nil || flat.DataPath != nil || flat.Production != nil
}

// 辅助方法：检查是否有传输配置
func hasTransferConfig(flat *models.FlatConfigUpdate) bool {
	return hasUploadConfig(flat) || hasDownloadConfig(flat)
}

// 辅助方法：检查是否有上传配置
func hasUploadConfig(flat *models.FlatConfigUpdate) bool {
	return flat.OpenUpload != nil || flat.UploadSize != nil || flat.EnableChunk != nil ||
		flat.ChunkSize != nil || flat.MaxSaveSeconds != nil
}

// 辅助方法：检查是否有下载配置
func hasDownloadConfig(flat *models.FlatConfigUpdate) bool {
	return flat.EnableConcurrentDownload != nil || flat.MaxConcurrentDownloads != nil ||
		flat.DownloadTimeout != nil
}

// 辅助方法：检查是否有用户配置
func hasUserConfig(flat *models.FlatConfigUpdate) bool {
	return flat.AllowUserRegistration != nil || flat.RequireEmailVerify != nil ||
		flat.UserUploadSize != nil || flat.UserStorageQuota != nil ||
		flat.SessionExpiryHours != nil || flat.MaxSessionsPerUser != nil ||
		flat.JWTSecret != nil
}

// 辅助方法：检查是否有MCP配置
func hasMCPConfig(flat *models.FlatConfigUpdate) bool {
	return flat.EnableMCPServer != nil || flat.MCPPort != nil || flat.MCPHost != nil
}
