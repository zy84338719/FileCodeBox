package db

import (
	"time"

	"gorm.io/gorm"
)

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

// UserQuery 用户查询条件
type UserQuery struct {
	gorm.Model
	ID       *uint  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Status   string `json:"status"`
	Search   string `json:"search"` // 模糊搜索用户名或邮箱
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
}

// UserUpdate 用户更新数据
type UserUpdate struct {
	gorm.Model
	Username      *string    `json:"username"`
	Email         *string    `json:"email"`
	PasswordHash  *string    `json:"password_hash"`
	Nickname      *string    `json:"nickname"`
	Avatar        *string    `json:"avatar"`
	Role          *string    `json:"role"`
	Status        *string    `json:"status"`
	EmailVerified *bool      `json:"email_verified"`
	LastLoginAt   *time.Time `json:"last_login_at"`
	LastLoginIP   *string    `json:"last_login_ip"`
}

// UserStats 用户统计查询结果
type UserStats struct {
	gorm.Model
	UserID         uint  `json:"user_id"`
	TotalUploads   int   `json:"total_uploads"`
	TotalDownloads int   `json:"total_downloads"`
	TotalStorage   int64 `json:"total_storage"`
	FileCount      int   `json:"file_count"`
}
