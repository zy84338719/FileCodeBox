package web

// PaginationInfo 分页信息
type PaginationInfo struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
}

// SystemInfoResponse 系统信息响应
type SystemInfoResponse struct {
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Version  string `json:"version"`
	Go       string `json:"go"`
	Uptime   string `json:"uptime"`
	Memory   string `json:"memory"`
	Database string `json:"database"`
}

// APIResponse 通用API响应
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// SuccessResponse 成功响应结构
type SuccessResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// CleanedCountResponse 清理计数响应
type CleanedCountResponse struct {
	CleanedCount int `json:"cleaned_count"`
}

// PaginatedResponse 分页响应
type PaginatedResponse struct {
	List       interface{}    `json:"list"`
	Pagination PaginationInfo `json:"pagination"`
}

// TokenResponse 令牌响应
type TokenResponse struct {
	Token    string      `json:"token"`
	UserInfo interface{} `json:"user_info,omitempty"`
}

// UploadInfoResponse 上传信息响应
type UploadInfoResponse struct {
	ShareCode    string `json:"share_code"`
	DownloadLink string `json:"download_link"`
}
