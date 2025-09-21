// Package config 用户系统配置模块
package config

import (
	"fmt"
	"strings"
	"time"
)

// UserSystemConfig 用户系统配置
type UserSystemConfig struct {
	AllowUserRegistration int    `json:"allow_user_registration"` // 是否允许用户注册
	RequireEmailVerify    int    `json:"require_email_verify"`    // 是否需要邮箱验证
	UserUploadSize        int64  `json:"user_upload_size"`        // 用户上传文件大小限制
	UserStorageQuota      int64  `json:"user_storage_quota"`      // 用户存储配额
	SessionExpiryHours    int    `json:"session_expiry_hours"`    // 用户会话过期时间
	MaxSessionsPerUser    int    `json:"max_sessions_per_user"`   // 每个用户最大会话数
	JWTSecret             string `json:"jwt_secret"`              // JWT签名密钥
}

// NewUserSystemConfig 创建用户系统配置
func NewUserSystemConfig() *UserSystemConfig {
	return &UserSystemConfig{
		AllowUserRegistration: 1,                    // 允许注册
		RequireEmailVerify:    0,                    // 不要求邮箱验证
		UserUploadSize:        50 * 1024 * 1024,     // 50MB
		UserStorageQuota:      1024 * 1024 * 1024,   // 1GB
		SessionExpiryHours:    24 * 7,               // 7天
		MaxSessionsPerUser:    5,                    // 最多5个会话
		JWTSecret:             "FileCodeBox2025JWT", // 默认密钥
	}
}

// Validate 验证用户系统配置
func (usc *UserSystemConfig) Validate() error {
	var errors []string

	// 验证上传大小限制
	if usc.UserUploadSize < 0 {
		errors = append(errors, "用户上传文件大小限制不能为负数")
	}
	if usc.UserUploadSize > 10*1024*1024*1024 { // 10GB
		errors = append(errors, "用户上传文件大小限制不能超过10GB")
	}

	// 验证存储配额
	if usc.UserStorageQuota < 0 {
		errors = append(errors, "用户存储配额不能为负数")
	}
	if usc.UserStorageQuota > 100*1024*1024*1024 { // 100GB
		errors = append(errors, "用户存储配额不能超过100GB")
	}

	// 验证会话过期时间
	if usc.SessionExpiryHours < 1 {
		errors = append(errors, "会话过期时间不能小于1小时")
	}
	if usc.SessionExpiryHours > 24*365 { // 1年
		errors = append(errors, "会话过期时间不能超过1年")
	}

	// 验证最大会话数
	if usc.MaxSessionsPerUser < 1 {
		errors = append(errors, "每用户最大会话数不能小于1")
	}
	if usc.MaxSessionsPerUser > 50 {
		errors = append(errors, "每用户最大会话数不能超过50")
	}

	// 验证JWT密钥
	if strings.TrimSpace(usc.JWTSecret) == "" {
		errors = append(errors, "JWT密钥不能为空")
	}
	if len(usc.JWTSecret) < 16 {
		errors = append(errors, "JWT密钥长度不能小于16个字符")
	}

	if len(errors) > 0 {
		return fmt.Errorf("用户系统配置验证失败: %s", strings.Join(errors, "; "))
	}

	return nil
}

// IsUserSystemEnabled 判断是否启用用户系统 - 始终返回true
func (usc *UserSystemConfig) IsUserSystemEnabled() bool {
	return true
}

// IsRegistrationAllowed 判断是否允许用户注册
func (usc *UserSystemConfig) IsRegistrationAllowed() bool {
	return usc.AllowUserRegistration == 1
}

// IsEmailVerifyRequired 判断是否需要邮箱验证
func (usc *UserSystemConfig) IsEmailVerifyRequired() bool {
	return usc.RequireEmailVerify == 1
}

// GetUserUploadSizeMB 获取用户上传大小限制（MB）
func (usc *UserSystemConfig) GetUserUploadSizeMB() float64 {
	return float64(usc.UserUploadSize) / (1024 * 1024)
}

// GetUserStorageQuotaGB 获取用户存储配额（GB）
func (usc *UserSystemConfig) GetUserStorageQuotaGB() float64 {
	if usc.UserStorageQuota == 0 {
		return 0 // 无限制
	}
	return float64(usc.UserStorageQuota) / (1024 * 1024 * 1024)
}

// GetSessionDuration 获取会话持续时间
func (usc *UserSystemConfig) GetSessionDuration() time.Duration {
	return time.Duration(usc.SessionExpiryHours) * time.Hour
}

// GetSessionExpiryDays 获取会话过期天数
func (usc *UserSystemConfig) GetSessionExpiryDays() float64 {
	return float64(usc.SessionExpiryHours) / 24
}

// IsStorageQuotaUnlimited 判断存储配额是否无限制
func (usc *UserSystemConfig) IsStorageQuotaUnlimited() bool {
	return usc.UserStorageQuota == 0
}

// Clone 克隆配置
func (usc *UserSystemConfig) Clone() *UserSystemConfig {
	return &UserSystemConfig{
		AllowUserRegistration: usc.AllowUserRegistration,
		RequireEmailVerify:    usc.RequireEmailVerify,
		UserUploadSize:        usc.UserUploadSize,
		UserStorageQuota:      usc.UserStorageQuota,
		SessionExpiryHours:    usc.SessionExpiryHours,
		MaxSessionsPerUser:    usc.MaxSessionsPerUser,
		JWTSecret:             usc.JWTSecret,
	}
}

// EnableRegistration 启用用户注册
func (usc *UserSystemConfig) EnableRegistration() {
	usc.AllowUserRegistration = 1
}

// DisableRegistration 禁用用户注册
func (usc *UserSystemConfig) DisableRegistration() {
	usc.AllowUserRegistration = 0
}

// EnableEmailVerify 启用邮箱验证
func (usc *UserSystemConfig) EnableEmailVerify() {
	usc.RequireEmailVerify = 1
}

// DisableEmailVerify 禁用邮箱验证
func (usc *UserSystemConfig) DisableEmailVerify() {
	usc.RequireEmailVerify = 0
}

// SetUserUploadSizeMB 设置用户上传大小限制（MB）
func (usc *UserSystemConfig) SetUserUploadSizeMB(sizeMB int64) error {
	usc.UserUploadSize = sizeMB * 1024 * 1024
	return usc.Validate()
}

// SetUserStorageQuotaGB 设置用户存储配额（GB）
func (usc *UserSystemConfig) SetUserStorageQuotaGB(quotaGB int64) error {
	if quotaGB == 0 {
		usc.UserStorageQuota = 0 // 无限制
	} else {
		usc.UserStorageQuota = quotaGB * 1024 * 1024 * 1024
	}
	return usc.Validate()
}

// SetSessionExpiryDays 设置会话过期天数
func (usc *UserSystemConfig) SetSessionExpiryDays(days int) error {
	usc.SessionExpiryHours = days * 24
	return usc.Validate()
}

// GenerateRandomJWTSecret 生成随机JWT密钥
func (usc *UserSystemConfig) GenerateRandomJWTSecret() error {
	// 这里可以实现随机密钥生成逻辑
	// 暂时使用时间戳作为简单实现
	usc.JWTSecret = fmt.Sprintf("FileCodeBox%d", time.Now().Unix())
	return usc.Validate()
}
