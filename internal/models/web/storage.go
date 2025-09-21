package web

import "github.com/zy84338719/filecodebox/internal/config"

// StorageInfoResponse 存储信息响应
type StorageInfoResponse struct {
	Current        string                      `json:"current"`
	Available      []string                    `json:"available"`
	StorageDetails map[string]WebStorageDetail `json:"storage_details"`
	StorageConfig  *config.StorageConfig       `json:"storage_config"`
}

// WebStorageDetail Web API 存储详情
type WebStorageDetail struct {
	Type      string `json:"type"`
	Available bool   `json:"available"`
	Error     string `json:"error,omitempty"`
	// StoragePath 本存储的路径或标识（例如本地目录、S3 bucket）
	StoragePath string `json:"storage_path,omitempty"`
	// UsagePercent 使用率（0-100），如果无法获取则为 nil
	UsagePercent *int `json:"usage_percent,omitempty"`
}

// StorageTestRequest 存储测试请求
type StorageTestRequest struct {
	Type   string                `json:"type" binding:"required"`
	Config *config.StorageConfig `json:"config" binding:"required"`
}

// StorageTestResponse 存储测试响应
type StorageTestResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// StorageSwitchRequest 存储切换请求
type StorageSwitchRequest struct {
	Type   string                `json:"type" binding:"required"`
	Config *config.StorageConfig `json:"config,omitempty"`
}

// StorageSwitchResponse 存储切换响应
type StorageSwitchResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	CurrentType string `json:"current_type"`
}

// StorageConnectionResponse 存储连接响应
type StorageConnectionResponse struct {
	Type   string `json:"type"`
	Status string `json:"status"`
}
