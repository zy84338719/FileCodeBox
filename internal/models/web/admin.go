package web

import "github.com/zy84338719/filecodebox/internal/config"

// AdminLoginRequest 管理员登录请求
type AdminLoginRequest struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password" binding:"required"`
}

// AdminLoginResponse 管理员登录响应
type AdminLoginResponse struct {
	Token     string `json:"token"`
	TokenType string `json:"token_type"`
	ExpiresIn int64  `json:"expires_in"`
}

// AdminStatsResponse 管理员统计响应
type AdminStatsResponse struct {
	TotalUsers         int64  `json:"total_users"`
	ActiveUsers        int64  `json:"active_users"`
	TodayRegistrations int64  `json:"today_registrations"`
	TodayUploads       int64  `json:"today_uploads"`
	TotalFiles         int64  `json:"total_files"`
	ActiveFiles        int64  `json:"active_files"`
	TotalSize          int64  `json:"total_size"`
	SysStart           string `json:"sys_start"`
}

// AdminSystemInfoResponse 系统信息响应
type AdminSystemInfoResponse struct {
	GoVersion          string `json:"go_version"`
	BuildTime          string `json:"build_time"`
	GitCommit          string `json:"git_commit"`
	OSInfo             string `json:"os_info"`
	CPUCores           int    `json:"cpu_cores"`
	MemoryUsage        int64  `json:"memory_usage"`
	DiskUsage          int64  `json:"disk_usage"`
	Uptime             string `json:"uptime"`
	FileCodeBoxVersion string `json:"filecodebox_version"`
}

// AdminStorageStatusResponse 存储状态响应
type AdminStorageStatusResponse struct {
	StorageType    string  `json:"storage_type"`
	TotalSpace     int64   `json:"total_space"`
	UsedSpace      int64   `json:"used_space"`
	AvailableSpace int64   `json:"available_space"`
	UsagePercent   float64 `json:"usage_percent"`
	Status         string  `json:"status"`
	IsHealthy      bool    `json:"is_healthy"`
}

// AdminPerformanceMetricsResponse 性能指标响应
type AdminPerformanceMetricsResponse struct {
	CPUUsage          float64 `json:"cpu_usage"`
	MemoryUsage       int64   `json:"memory_usage"`
	DiskUsage         int64   `json:"disk_usage"`
	NetworkIn         int64   `json:"network_in"`
	NetworkOut        int64   `json:"network_out"`
	RequestsPerSecond int64   `json:"requests_per_second"`
	ActiveConnections int     `json:"active_connections"`
	ResponseTime      float64 `json:"response_time"`
}

// AdminDatabaseStatsResponse 数据库统计响应
type AdminDatabaseStatsResponse struct {
	DatabaseSize         int64   `json:"database_size"`
	TableCount           int     `json:"table_count"`
	RecordCount          int64   `json:"record_count"`
	DatabaseType         string  `json:"database_type"`
	DatabaseVersion      string  `json:"database_version"`
	LastBackupTime       string  `json:"last_backup_time"`
	FragmentationPercent float64 `json:"fragmentation_percent"`
}

// AdminFilesRequest 管理员文件列表请求
type AdminFilesRequest struct {
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"page_size" form:"page_size"`
	Search   string `json:"search" form:"search"`
}

// AdminFilesResponse 管理员文件列表响应
type AdminFilesResponse struct {
	Files      []AdminFileInfo `json:"files"`
	Pagination PaginationInfo  `json:"pagination"`
}

// AdminFileInfo 管理员文件信息
type AdminFileInfo struct {
	FileInfo
	UserID   *uint  `json:"user_id"`
	Username string `json:"username,omitempty"`
	OwnerIP  string `json:"owner_ip"`
}

// AdminUsersRequest 管理员用户列表请求
type AdminUsersRequest struct {
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"page_size" form:"page_size"`
	Search   string `json:"search" form:"search"`
	Status   string `json:"status" form:"status"`
}

// AdminUsersResponse 管理员用户列表响应
type AdminUsersResponse struct {
	Users      []AdminUserInfo `json:"users"`
	Pagination PaginationInfo  `json:"pagination"`
}

// AdminUserInfo 管理员用户信息
type AdminUserInfo struct {
	UserInfo
	TotalUploads   int    `json:"total_uploads"`
	TotalDownloads int    `json:"total_downloads"`
	TotalStorage   int64  `json:"total_storage"`
	LastLoginAt    string `json:"last_login_at,omitempty"`
	LastLoginIP    string `json:"last_login_ip"`
}

// AdminUserCreateRequest 管理员创建用户请求
type AdminUserCreateRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Nickname string `json:"nickname"`
	Role     string `json:"role"`
}

// AdminUserUpdateRequest 管理员更新用户请求
type AdminUserUpdateRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
	Role     string `json:"role"`
	Status   string `json:"status"`
}

// AdminUserStatusRequest 管理员用户状态请求
type AdminUserStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// AdminConfigResponse 管理员配置响应
type AdminConfigResponse struct {
	AdminConfigRequest
}

// AdminConfigRequest 管理员配置更新请求
type AdminConfigRequest struct {
	// 配置模块，直接使用各配置结构体避免字段重复
	Base     *config.BaseConfig       `json:"base,omitempty"`
	Database *config.DatabaseConfig   `json:"database,omitempty"`
	Transfer *config.TransferConfig   `json:"transfer,omitempty"`
	Storage  *config.StorageConfig    `json:"storage,omitempty"`
	User     *config.UserSystemConfig `json:"user,omitempty"`
	MCP      *config.MCPConfig        `json:"mcp,omitempty"`
	UI       *config.UIConfig         `json:"ui,omitempty"`

	// 系统运行时特有字段（不属于配置模块的字段）
	SysStart *string `json:"sys_start,omitempty"`

	// 顶层通知字段保留与历史配置结构兼容
	NotifyTitle   *string `json:"notify_title,omitempty"`
	NotifyContent *string `json:"notify_content,omitempty"`
}

// CountResponse 通用计数响应
type CountResponse struct {
	Count int `json:"count"`
}

// BackupPathResponse 备份路径响应
type BackupPathResponse struct {
	BackupPath string `json:"backup_path"`
}

// LogPathResponse 日志路径响应
type LogPathResponse struct {
	LogPath string `json:"log_path"`
}

// AdminUserStatsResponse 管理员用户统计响应
type AdminUserStatsResponse struct {
	TotalUsers         int64 `json:"total_users"`
	ActiveUsers        int64 `json:"active_users"`
	TodayRegistrations int64 `json:"today_registrations"`
	TodayUploads       int64 `json:"today_uploads"`
}

// AdminUsersListResponse 管理员用户列表响应
type AdminUsersListResponse struct {
	Users      []AdminUserDetail      `json:"users"`
	Stats      AdminUserStatsResponse `json:"stats"`
	Pagination PaginationResponse     `json:"pagination"`
}

// AdminUserDetail 管理员用户详细信息
type AdminUserDetail struct {
	ID             uint   `json:"id"`
	Username       string `json:"username"`
	Email          string `json:"email"`
	Nickname       string `json:"nickname"`
	Role           string `json:"role"`
	IsAdmin        bool   `json:"is_admin"`
	IsActive       bool   `json:"is_active"`
	Status         string `json:"status"`
	EmailVerified  bool   `json:"email_verified"`
	CreatedAt      string `json:"created_at"`
	LastLoginAt    string `json:"last_login_at"`
	LastLoginIP    string `json:"last_login_ip"`
	TotalUploads   int    `json:"total_uploads"`
	TotalDownloads int    `json:"total_downloads"`
	TotalStorage   int64  `json:"total_storage"`
}

// PaginationResponse 分页响应
type PaginationResponse struct {
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
	Pages    int64 `json:"pages"`
}

// IDResponse 通用ID响应
type IDResponse struct {
	ID uint `json:"id"`
}

// AdminUserFilesResponse 管理员用户文件响应
type AdminUserFilesResponse struct {
	Files    []AdminFileDetail `json:"files"`
	Username string            `json:"username"`
	Total    int64             `json:"total"`
}

// AdminFileDetail 管理员文件详细信息
type AdminFileDetail struct {
	ID           uint   `json:"id"`
	Code         string `json:"code"`
	Prefix       string `json:"prefix"`
	Suffix       string `json:"suffix"`
	Size         int64  `json:"size"`
	Type         string `json:"type"`
	ExpiredAt    string `json:"expired_at"`
	ExpiredCount int    `json:"expired_count"`
	UsedCount    int    `json:"used_count"`
	CreatedAt    string `json:"created_at"`
	RequireAuth  bool   `json:"require_auth"`
	UploadType   string `json:"upload_type"`
}

// TransferLogItem 审计日志单条记录
type TransferLogItem struct {
	ID         uint    `json:"id"`
	Operation  string  `json:"operation"`
	FileCode   string  `json:"file_code"`
	FileName   string  `json:"file_name"`
	FileSize   int64   `json:"file_size"`
	UserID     *uint   `json:"user_id,omitempty"`
	Username   string  `json:"username"`
	IP         string  `json:"ip"`
	DurationMs int64   `json:"duration_ms"`
	CreatedAt  string  `json:"created_at"`
}

// TransferLogListResponse 审计日志列表响应
type TransferLogListResponse struct {
	Logs       []TransferLogItem `json:"logs"`
	Pagination PaginationResponse `json:"pagination"`
}

// MCPStatusResponse MCP状态响应
type MCPStatusResponse struct {
	Status string            `json:"status"`
	Config *config.MCPConfig `json:"config"`
}

// MCPTestResponse MCP连接测试响应
type MCPTestResponse struct {
	MCPStatusResponse
}

// LogsResponse 日志响应
type LogsResponse struct {
	Logs  []string `json:"logs"`
	Total int      `json:"total"`
}

// TasksResponse 任务响应
type TasksResponse struct {
	Tasks interface{} `json:"tasks"` // 使用 interface{} 以兼容现有类型
	Total int         `json:"total"`
}
