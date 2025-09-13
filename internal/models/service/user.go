package service

import "time"

// UserStatsData 用户统计数据
type UserStatsData struct {
	TotalUploads    int        `json:"total_uploads"`
	TotalDownloads  int        `json:"total_downloads"`
	TotalStorage    int64      `json:"total_storage"`
	MaxUploadSize   int64      `json:"max_upload_size"`
	MaxStorageQuota int64      `json:"max_storage_quota"`
	CurrentFiles    int        `json:"current_files"`
	FileCount       int        `json:"file_count"`
	LastLoginAt     *time.Time `json:"last_login_at"`
	LastLoginIP     string     `json:"last_login_ip"`
	EmailVerified   bool       `json:"email_verified"`
	Status          string     `json:"status"`
	Role            string     `json:"role"`
	LastUploadAt    *time.Time `json:"last_upload_at"`
	LastDownloadAt  *time.Time `json:"last_download_at"`
}

// UserProfileUpdateData 用户资料更新数据
type UserProfileUpdateData struct {
	Nickname      string `json:"nickname"`
	Avatar        string `json:"avatar"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}

// UserFileData 用户文件数据
type UserFileData struct {
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
}

// UserFilesResult 用户文件查询结果
type UserFilesResult struct {
	Files []UserFileData `json:"files"`
	Total int64          `json:"total"`
}
