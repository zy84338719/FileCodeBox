// Package models 定义应用程序的数据模型
package models

import (
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"gorm.io/gorm"
)

// 编译时通过 -ldflags 传入的全局变量
var (
	// GoVersion 编译使用的 Go 版本
	GoVersion = "unknown"

	// BuildTime 编译时间，格式为 ISO8601
	BuildTime = "unknown"

	// GitCommit Git 提交哈希值
	GitCommit = "unknown"

	// GitBranch Git 分支名称
	GitBranch = "unknown"

	// Version 应用版本号
	Version = "0.0.1"
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

	// 检查次数过期 (当次数为0时过期，-1表示无限制)
	if f.ExpiredCount >= 0 && f.ExpiredCount <= 0 {
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

// UploadChunk 上传分片模型
type UploadChunk struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	UploadID    string    `gorm:"index;size:36" json:"upload_id"`
	ChunkIndex  int       `json:"chunk_index"`
	ChunkHash   string    `gorm:"size:64" json:"chunk_hash"`
	TotalChunks int       `json:"total_chunks"`
	FileSize    int64     `json:"file_size"`
	ChunkSize   int       `json:"chunk_size"`
	FileName    string    `gorm:"size:255" json:"file_name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Completed   bool      `gorm:"default:false" json:"completed"`
	RetryCount  int       `gorm:"default:0" json:"retry_count"`            // 重试次数
	LastError   string    `gorm:"type:text" json:"last_error"`             // 最后错误信息
	Status      string    `gorm:"size:20;default:'pending'" json:"status"` // pending, uploading, completed, failed
}

// GetUploadProgress 获取上传进度
func (u *UploadChunk) GetUploadProgress(db *gorm.DB) (float64, error) {
	if u.ChunkIndex != -1 {
		return 0, fmt.Errorf("只能从控制记录获取进度")
	}

	var completedCount int64
	err := db.Model(&UploadChunk{}).
		Where("upload_id = ? AND completed = true AND chunk_index >= 0", u.UploadID).
		Count(&completedCount).Error
	if err != nil {
		return 0, err
	}

	if u.TotalChunks == 0 {
		return 0, nil
	}

	return float64(completedCount) / float64(u.TotalChunks) * 100, nil
}

// IsComplete 检查上传是否完成
func (u *UploadChunk) IsComplete(db *gorm.DB) (bool, error) {
	progress, err := u.GetUploadProgress(db)
	if err != nil {
		return false, err
	}
	return progress >= 100.0, nil
}

// KeyValue 键值对模型
type KeyValue struct {
	gorm.Model
	Key   string `gorm:"uniqueIndex;size:255" json:"key"`
	Value string `gorm:"type:text" json:"value"`
}

// User 用户模型
type User struct {
	gorm.Model
	Username      string     `gorm:"uniqueIndex;size:50" json:"username"`
	Email         string     `gorm:"uniqueIndex;size:100" json:"email"`
	PasswordHash  string     `gorm:"size:128" json:"-"`                      // 密码哈希，不返回给前端
	Nickname      string     `gorm:"size:50" json:"nickname"`                // 用户昵称
	Avatar        string     `gorm:"size:255" json:"avatar"`                 // 头像URL
	Role          string     `gorm:"size:20;default:'user'" json:"role"`     // admin, user
	Status        string     `gorm:"size:20;default:'active'" json:"status"` // active, inactive, banned
	EmailVerified bool       `gorm:"default:false" json:"email_verified"`    // 邮箱是否验证
	LastLoginAt   *time.Time `json:"last_login_at"`                          // 最后登录时间
	LastLoginIP   string     `gorm:"size:45" json:"last_login_ip"`           // 最后登录IP

	// 用户上传统计
	TotalUploads    int   `gorm:"default:0" json:"total_uploads"`     // 总上传次数
	TotalDownloads  int   `gorm:"default:0" json:"total_downloads"`   // 总下载次数
	TotalStorage    int64 `gorm:"default:0" json:"total_storage"`     // 总存储大小(字节)
	MaxUploadSize   int64 `gorm:"default:0" json:"max_upload_size"`   // 最大单次上传大小(字节)，0表示使用系统默认
	MaxStorageQuota int64 `gorm:"default:0" json:"max_storage_quota"` // 最大存储配额(字节)，0表示无限制
}

// UserSession 用户会话模型
type UserSession struct {
	gorm.Model
	UserID    uint      `gorm:"index" json:"user_id"`
	SessionID string    `gorm:"uniqueIndex;size:128" json:"session_id"` // JWT token ID
	IPAddress string    `gorm:"size:45" json:"ip_address"`              // 登录IP
	UserAgent string    `gorm:"size:500" json:"user_agent"`             // 用户代理
	ExpiresAt time.Time `json:"expires_at"`                             // 过期时间

	IsActive bool `gorm:"default:true" json:"is_active"` // 是否活跃
}

// BuildInfo 构建信息结构体
type BuildInfo struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	GitBranch string `json:"git_branch"`
	BuildTime string `json:"build_time"`
	GoVersion string `json:"go_version"`
	Arch      string `json:"arch"`
	OS        string `json:"os"`
}

// GetBuildInfo 获取应用构建信息
func GetBuildInfo() *BuildInfo {
	return &BuildInfo{
		Version:   Version,
		GitCommit: GitCommit,
		GitBranch: GitBranch,
		BuildTime: BuildTime,
		GoVersion: runtime.Version(), // 运行时获取真实的Go版本
		Arch:      runtime.GOARCH,
		OS:        runtime.GOOS,
	}
}
