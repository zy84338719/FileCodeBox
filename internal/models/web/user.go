package web

// UserProfileUpdateRequest 用户资料更新请求
type UserProfileUpdateRequest struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// UserPasswordChangeRequest 密码修改请求
type UserPasswordChangeRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// UserPasswordResetRequest 密码重置请求
type UserPasswordResetRequest struct {
	Email string `json:"email" binding:"required"`
}

// UserProfileResponse 用户资料响应
type UserProfileResponse struct {
	ID            uint   `json:"id"`
	Username      string `json:"username"`
	Email         string `json:"email"`
	Nickname      string `json:"nickname"`
	Avatar        string `json:"avatar"`
	Role          string `json:"role"`
	Status        string `json:"status"`
	EmailVerified bool   `json:"email_verified"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// UserStatsResponse 用户统计响应
type UserStatsResponse struct {
	TotalUploads    int    `json:"total_uploads"`
	TotalDownloads  int    `json:"total_downloads"`
	TotalStorage    int64  `json:"total_storage"`
	MaxUploadSize   int64  `json:"max_upload_size"`
	MaxStorageQuota int64  `json:"max_storage_quota"`
	CurrentFiles    int    `json:"current_files"`
	FileCount       int    `json:"file_count"`
	LastLoginAt     string `json:"last_login_at,omitempty"`
	LastLoginIP     string `json:"last_login_ip"`
	EmailVerified   bool   `json:"email_verified"`
	Status          string `json:"status"`
	Role            string `json:"role"`
	LastUploadAt    string `json:"last_upload_at,omitempty"`
	LastDownloadAt  string `json:"last_download_at,omitempty"`
}

// UserFilesRequest 用户文件列表请求
type UserFilesRequest struct {
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"page_size" form:"page_size"`
	Search   string `json:"search" form:"search"`
}

// UserFilesResponse 用户文件列表响应
type UserFilesResponse struct {
	Files      []FileInfo     `json:"files"`
	Pagination PaginationInfo `json:"pagination"`
}

// UserSystemInfoResponse 用户系统信息响应
type UserSystemInfoResponse struct {
	// 使用整型 0/1 表示开关，以与配置层和前端保持一致
	UserSystemEnabled        int `json:"user_system_enabled"`
	AllowUserRegistration    int `json:"allow_user_registration"`
	RequireEmailVerification int `json:"require_email_verification"`
}

type UserDataRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"omitempty,email"`
	Password string `json:"password" binding:"required"`
	Nickname string `json:"nickname"`
	IsAdmin  bool   `json:"is_admin"`
	IsActive bool   `json:"is_active"`
}
