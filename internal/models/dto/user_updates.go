package dto

import "time"

// UserUpdateFields 用户更新字段结构体
type UserUpdateFields struct {
	Email          *string    `json:"email,omitempty"`
	PasswordHash   *string    `json:"password_hash,omitempty"`
	Nickname       *string    `json:"nickname,omitempty"`
	Avatar         *string    `json:"avatar,omitempty"`
	Role           *string    `json:"role,omitempty"`
	Status         *string    `json:"status,omitempty"`
	EmailVerified  *bool      `json:"email_verified,omitempty"`
	LastLoginAt    *time.Time `json:"last_login_at,omitempty"`
	LastLoginIP    *string    `json:"last_login_ip,omitempty"`
	TotalUploads   *int       `json:"total_uploads,omitempty"`
	TotalDownloads *int       `json:"total_downloads,omitempty"`
	TotalStorage   *int64     `json:"total_storage,omitempty"`
}

// ToMap 将结构体转换为 map，只包含非空字段
func (u *UserUpdateFields) ToMap() map[string]interface{} {
	updates := make(map[string]interface{})

	if u.Email != nil {
		updates["email"] = *u.Email
	}
	if u.PasswordHash != nil {
		updates["password_hash"] = *u.PasswordHash
	}
	if u.Nickname != nil {
		updates["nickname"] = *u.Nickname
	}
	if u.Avatar != nil {
		updates["avatar"] = *u.Avatar
	}
	if u.Role != nil {
		updates["role"] = *u.Role
	}
	if u.Status != nil {
		updates["status"] = *u.Status
	}
	if u.EmailVerified != nil {
		updates["email_verified"] = *u.EmailVerified
	}
	if u.LastLoginAt != nil {
		updates["last_login_at"] = *u.LastLoginAt
	}
	if u.LastLoginIP != nil {
		updates["last_login_ip"] = *u.LastLoginIP
	}
	if u.TotalUploads != nil {
		updates["total_uploads"] = *u.TotalUploads
	}
	if u.TotalDownloads != nil {
		updates["total_downloads"] = *u.TotalDownloads
	}
	if u.TotalStorage != nil {
		updates["total_storage"] = *u.TotalStorage
	}

	return updates
}

// HasUpdates 检查是否有任何更新字段
func (u *UserUpdateFields) HasUpdates() bool {
	return u.Email != nil || u.PasswordHash != nil || u.Nickname != nil ||
		u.Avatar != nil || u.Role != nil || u.Status != nil ||
		u.EmailVerified != nil || u.LastLoginAt != nil || u.LastLoginIP != nil ||
		u.TotalUploads != nil || u.TotalDownloads != nil || u.TotalStorage != nil
}

// UserProfileUpdateFields 用户资料更新字段（用户自己更新）
type UserProfileUpdateFields struct {
	Email        *string `json:"email,omitempty"`
	Nickname     *string `json:"nickname,omitempty"`
	Avatar       *string `json:"avatar,omitempty"`
	PasswordHash *string `json:"password_hash,omitempty"`
}

// ToMap 将用户资料更新字段转换为 map
func (u *UserProfileUpdateFields) ToMap() map[string]interface{} {
	updates := make(map[string]interface{})

	if u.Email != nil {
		updates["email"] = *u.Email
	}
	if u.Nickname != nil {
		updates["nickname"] = *u.Nickname
	}
	if u.Avatar != nil {
		updates["avatar"] = *u.Avatar
	}
	if u.PasswordHash != nil {
		updates["password_hash"] = *u.PasswordHash
	}

	return updates
}

// HasUpdates 检查是否有任何更新字段
func (u *UserProfileUpdateFields) HasUpdates() bool {
	return u.Email != nil || u.Nickname != nil || u.Avatar != nil || u.PasswordHash != nil
}

// UserStatsUpdateFields 用户统计信息更新字段
type UserStatsUpdateFields struct {
	TotalUploads   *int       `json:"total_uploads,omitempty"`
	TotalDownloads *int       `json:"total_downloads,omitempty"`
	TotalStorage   *int64     `json:"total_storage,omitempty"`
	LastLoginAt    *time.Time `json:"last_login_at,omitempty"`
	LastLoginIP    *string    `json:"last_login_ip,omitempty"`
}

// ToMap 将用户统计更新字段转换为 map
func (u *UserStatsUpdateFields) ToMap() map[string]interface{} {
	updates := make(map[string]interface{})

	if u.TotalUploads != nil {
		updates["total_uploads"] = *u.TotalUploads
	}
	if u.TotalDownloads != nil {
		updates["total_downloads"] = *u.TotalDownloads
	}
	if u.TotalStorage != nil {
		updates["total_storage"] = *u.TotalStorage
	}
	if u.LastLoginAt != nil {
		updates["last_login_at"] = *u.LastLoginAt
	}
	if u.LastLoginIP != nil {
		updates["last_login_ip"] = *u.LastLoginIP
	}

	return updates
}

// HasUpdates 检查是否有任何更新字段
func (u *UserStatsUpdateFields) HasUpdates() bool {
	return u.TotalUploads != nil || u.TotalDownloads != nil || u.TotalStorage != nil ||
		u.LastLoginAt != nil || u.LastLoginIP != nil
}
