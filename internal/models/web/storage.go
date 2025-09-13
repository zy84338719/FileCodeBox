package web

// StorageInfoResponse 存储信息响应
type StorageInfoResponse struct {
	Current        string                   `json:"current"`
	Available      []string                 `json:"available"`
	StorageDetails map[string]StorageDetail `json:"storage_details"`
	StorageConfig  map[string]interface{}   `json:"storage_config"`
}

// StorageDetail 存储详情
type StorageDetail struct {
	Type      string `json:"type"`
	Available bool   `json:"available"`
	Error     string `json:"error,omitempty"`
}

// StorageTestRequest 存储测试请求
type StorageTestRequest struct {
	Type   string                 `json:"type" binding:"required"`
	Config map[string]interface{} `json:"config" binding:"required"`
}

// StorageTestResponse 存储测试响应
type StorageTestResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// StorageSwitchRequest 存储切换请求
type StorageSwitchRequest struct {
	Type   string                 `json:"type" binding:"required"`
	Config map[string]interface{} `json:"config" binding:"required"`
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
