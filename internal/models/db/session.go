package db

import (
	"time"

	"gorm.io/gorm"
)

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

// SessionQuery 会话查询条件
type SessionQuery struct {
	gorm.Model
	UserID    *uint      `json:"user_id"`
	SessionID string     `json:"session_id"`
	IPAddress string     `json:"ip_address"`
	IsActive  *bool      `json:"is_active"`
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
	Limit     int        `json:"limit"`
	Offset    int        `json:"offset"`
}

// SessionUpdate 会话更新数据
type SessionUpdate struct {
	gorm.Model
	IsActive  *bool      `json:"is_active"`
	ExpiresAt *time.Time `json:"expires_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

// KeyValueQuery 键值对查询条件
type KeyValueQuery struct {
	gorm.Model
	Key       string `json:"key"`
	KeyPrefix string `json:"key_prefix"` // 前缀匹配
	Limit     int    `json:"limit"`
	Offset    int    `json:"offset"`
}

// KeyValueUpdate 键值对更新数据
type KeyValueUpdate struct {
	gorm.Model
	Value     *string    `json:"value"`
	UpdatedAt *time.Time `json:"updated_at"`
}
