package service

import "time"

// AdminStatsData 管理员统计数据
type AdminStatsData struct {
	TotalFiles     int64 `json:"total_files"`
	TotalUsers     int64 `json:"total_users"`
	TotalStorage   int64 `json:"total_storage"`
	TodayUploads   int64 `json:"today_uploads"`
	TodayDownloads int64 `json:"today_downloads"`
	ActiveUsers    int64 `json:"active_users"`
	ExpiredFiles   int64 `json:"expired_files"`
	SystemUptime   int64 `json:"system_uptime"`
}

// AdminFileData 管理员文件数据
type AdminFileData struct {
	Code         string     `json:"code"`
	FileName     string     `json:"file_name"`
	Size         int64      `json:"size"`
	UploadType   string     `json:"upload_type"`
	RequireAuth  bool       `json:"require_auth"`
	UsedCount    int        `json:"used_count"`
	ExpiredCount int        `json:"expired_count"`
	ExpiredAt    *time.Time `json:"expired_at"`
	CreatedAt    time.Time  `json:"created_at"`
	IsExpired    bool       `json:"is_expired"`
	UserID       *uint      `json:"user_id"`
	Username     string     `json:"username,omitempty"`
	OwnerIP      string     `json:"owner_ip"`
}

// AdminFilesResult 管理员文件查询结果
type AdminFilesResult struct {
	Files []AdminFileData `json:"files"`
	Total int64           `json:"total"`
}

// AdminUserData 管理员用户数据
type AdminUserData struct {
	ID             uint       `json:"id"`
	Username       string     `json:"username"`
	Email          string     `json:"email"`
	Nickname       string     `json:"nickname"`
	Avatar         string     `json:"avatar"`
	Role           string     `json:"role"`
	Status         string     `json:"status"`
	EmailVerified  bool       `json:"email_verified"`
	TotalUploads   int        `json:"total_uploads"`
	TotalDownloads int        `json:"total_downloads"`
	TotalStorage   int64      `json:"total_storage"`
	LastLoginAt    *time.Time `json:"last_login_at"`
	LastLoginIP    string     `json:"last_login_ip"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// AdminUsersResult 管理员用户查询结果
type AdminUsersResult struct {
	Users []AdminUserData `json:"users"`
	Total int64           `json:"total"`
}

// SystemStatusData 系统状态数据
type SystemStatusData struct {
	Status   string                 `json:"status"`
	Uptime   int64                  `json:"uptime"`
	Memory   map[string]interface{} `json:"memory"`
	Storage  map[string]interface{} `json:"storage"`
	Database map[string]interface{} `json:"database"`
	Services map[string]interface{} `json:"services"`
	Config   map[string]interface{} `json:"config"`
}

// DatabaseStats 数据库统计信息
type DatabaseStats struct {
	TotalFiles   int64  `json:"total_files"`
	TotalUsers   int64  `json:"total_users"`
	TotalSize    int64  `json:"total_size"`
	DatabaseSize string `json:"database_size"`
}

// StorageStatus 存储状态信息
type StorageStatus struct {
	Type      string            `json:"type"`
	Status    string            `json:"status"`
	Available bool              `json:"available"`
	Details   map[string]string `json:"details"`
}

// DiskUsage 磁盘使用情况
type DiskUsage struct {
	TotalSpace     int64   `json:"total_space"`
	UsedSpace      int64   `json:"used_space"`
	AvailableSpace int64   `json:"available_space"`
	UsagePercent   float64 `json:"usage_percent"`
	StorageType    string  `json:"storage_type"`
	Success        bool    `json:"success"`
	Error          *string `json:"error"`
}

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	MemoryUsage   string    `json:"memory_usage"`
	CPUUsage      string    `json:"cpu_usage"`
	ResponseTime  string    `json:"response_time"`
	LastUpdated   time.Time `json:"last_updated"`
	DatabaseStats string    `json:"database_stats"`
}

// SystemInfo 系统信息
type SystemInfo struct {
	OS           string    `json:"os"`
	Architecture string    `json:"architecture"`
	GoVersion    string    `json:"go_version"`
	StartTime    time.Time `json:"start_time"`
	Uptime       string    `json:"uptime"`
}

// SecurityScanResult 安全扫描结果
type SecurityScanResult struct {
	Status      string   `json:"status"`
	Issues      []string `json:"issues"`
	LastScanned string   `json:"last_scanned"`
	Passed      bool     `json:"passed"`
	Suggestions []string `json:"suggestions"`
}

// PermissionCheckResult 权限检查结果
type PermissionCheckResult struct {
	Status      string            `json:"status"`
	Permissions map[string]string `json:"permissions"`
	Issues      []string          `json:"issues"`
}

// IntegrityCheckResult 完整性检查结果
type IntegrityCheckResult struct {
	Status       string   `json:"status"`
	CheckedFiles int      `json:"checked_files"`
	CorruptFiles int      `json:"corrupt_files"`
	MissingFiles int      `json:"missing_files"`
	Issues       []string `json:"issues"`
}

// LogStats 日志统计
type LogStats struct {
	TotalLogs   int    `json:"total_logs"`
	ErrorLogs   int    `json:"error_logs"`
	WarningLogs int    `json:"warning_logs"`
	InfoLogs    int    `json:"info_logs"`
	LastLogTime string `json:"last_log_time"`
	LogSize     string `json:"log_size"`
}

// RunningTask 运行中的任务
type RunningTask struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	Progress    int    `json:"progress"`
	StartTime   string `json:"start_time"`
	Description string `json:"description"`
}

// MCPConfig MCP配置信息
type MCPConfig struct {
	Enabled bool   `json:"enabled"`
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Status  string `json:"status"`
}

// MCPStatus MCP状态信息
type MCPStatus struct {
	Config      MCPConfig `json:"config"`
	IsRunning   bool      `json:"is_running"`
	LastStarted string    `json:"last_started"`
	Version     string    `json:"version"`
}

// MCPTestResult MCP测试结果
type MCPTestResult struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	ConnectedAt string `json:"connected_at"`
}

// StorageTestResult 存储测试结果
type StorageTestResult struct {
	Type    string `json:"type"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// UserStatsResponse 用户统计响应
type UserStatsResponse struct {
	UserID            uint    `json:"user_id"`
	TotalUploads      int64   `json:"total_uploads"`
	TotalDownloads    int64   `json:"total_downloads"`
	TotalStorage      int64   `json:"total_storage"`
	TotalFiles        int64   `json:"total_files"`
	TodayUploads      int64   `json:"today_uploads"`
	MaxUploadSize     int64   `json:"max_upload_size"`
	MaxStorageQuota   int64   `json:"max_storage_quota"`
	StorageUsage      int64   `json:"storage_usage"`
	StoragePercentage float64 `json:"storage_percentage"`
}
