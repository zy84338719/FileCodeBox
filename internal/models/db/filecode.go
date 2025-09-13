package db

import (
	"path/filepath"
	"time"

	"gorm.io/gorm"
)

// FileCode 文件代码模型
type FileCode struct {
	gorm.Model
	Code         string     `gorm:"uniqueIndex;size:255" json:"code"`
	Prefix       string     `gorm:"size:255" json:"prefix"`
	Suffix       string     `gorm:"size:255" json:"suffix"`
	UUIDFileName string     `gorm:"size:255" json:"uuid_file_name"`
	FilePath     string     `gorm:"size:255" json:"file_path"`
	Size         int64      `gorm:"default:0" json:"size"`
	Text         string     `gorm:"type:text" json:"text"`
	ExpiredAt    *time.Time `json:"expired_at"`
	ExpiredCount int        `gorm:"default:0" json:"expired_count"`
	UsedCount    int        `gorm:"default:0" json:"used_count"`

	FileHash  string `gorm:"size:64" json:"file_hash"`
	IsChunked bool   `gorm:"default:false" json:"is_chunked"`
	UploadID  string `gorm:"size:36" json:"upload_id"`

	// 新增：用户认证相关字段
	UserID      *uint  `gorm:"index" json:"user_id"`                           // 上传用户ID，为null表示匿名上传
	UploadType  string `gorm:"size:20;default:'anonymous'" json:"upload_type"` // anonymous, authenticated
	RequireAuth bool   `gorm:"default:false" json:"require_auth"`              // 是否需要登录才能下载
	OwnerIP     string `gorm:"size:45" json:"owner_ip"`                        // 上传者IP地址
}

// IsExpired 检查是否过期
func (f *FileCode) IsExpired() bool {
	// 检查时间过期
	if f.ExpiredAt != nil && f.ExpiredAt.Before(time.Now()) {
		return true
	}

	// 检查次数过期
	// ExpiredCount = -1 表示无限制次数，不过期
	// ExpiredCount = 0 表示已用完所有次数，过期
	// ExpiredCount > 0 表示剩余次数，不过期
	if f.ExpiredCount == 0 {
		return true
	}

	return false
}

// GetFilePath 获取文件路径
func (f *FileCode) GetFilePath() string {
	if f.FilePath == "" || f.UUIDFileName == "" {
		return ""
	}
	// 使用相对路径，让存储管理器来处理具体的基础路径
	return filepath.Join(f.FilePath, f.UUIDFileName)
}

// FileCodeQuery 文件代码查询条件
type FileCodeQuery struct {
	gorm.Model
	Code        string     `json:"code"`
	UserID      *uint      `json:"user_id"`
	UploadType  string     `json:"upload_type"`
	RequireAuth *bool      `json:"require_auth"`
	IsExpired   *bool      `json:"is_expired"`
	Search      string     `json:"search"` // 模糊搜索文件名
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
	Limit       int        `json:"limit"`
	Offset      int        `json:"offset"`
}

// FileCodeUpdate 文件代码更新数据
type FileCodeUpdate struct {
	gorm.Model
	ExpiredAt    *time.Time `json:"expired_at"`
	ExpiredCount *int       `json:"expired_count"`
	UsedCount    *int       `json:"used_count"`
	RequireAuth  *bool      `json:"require_auth"`
	OwnerIP      *string    `json:"owner_ip"`
}

// FileCodeStats 文件统计查询结果
type FileCodeStats struct {
	gorm.Model
	TotalFiles     int64 `json:"total_files"`
	TotalSize      int64 `json:"total_size"`
	TodayUploads   int64 `json:"today_uploads"`
	TodayDownloads int64 `json:"today_downloads"`
	ExpiredFiles   int64 `json:"expired_files"`
	AnonymousFiles int64 `json:"anonymous_files"`
	UserFiles      int64 `json:"user_files"`
}
