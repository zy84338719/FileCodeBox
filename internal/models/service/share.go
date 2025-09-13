package service

import (
	"mime/multipart"
	"time"
)

// ShareFileRequest 文件分享请求结构
type ShareFileRequest struct {
	ExpireValue int                   `json:"expire_value" form:"expire_value"`
	ExpireStyle string                `json:"expire_style" form:"expire_style"`
	RequireAuth bool                  `json:"require_auth" form:"require_auth"`
	File        *multipart.FileHeader `json:"-" form:"-"` // 文件内容
	ClientIP    string                `json:"client_ip" form:"client_ip"`
	UserID      *uint                 `json:"user_id" form:"user_id"`
}

// ShareTextRequest 文本分享请求结构
type ShareTextRequest struct {
	Text        string `json:"text" binding:"required"`
	ExpireValue int    `json:"expire_value"`
	ExpireStyle string `json:"expire_style"`
	RequireAuth bool   `json:"require_auth"`
	ClientIP    string `json:"client_ip"`
	UserID      *uint  `json:"user_id"`
}

// ShareStatsData 分享统计数据
type ShareStatsData struct {
	Code         string     `json:"code"`
	UsedCount    int        `json:"used_count"`
	ExpiredCount int        `json:"expired_count"`
	ExpiredAt    *time.Time `json:"expired_at"`
	IsExpired    bool       `json:"is_expired"`
	CreatedAt    time.Time  `json:"created_at"`
	FileSize     int64      `json:"file_size"`
	FileName     string     `json:"file_name"`
	UploadType   string     `json:"upload_type"`
	RequireAuth  bool       `json:"require_auth"`
}

// ShareUpdateData 分享更新数据
type ShareUpdateData struct {
	ExpiredAt    *time.Time `json:"expired_at"`
	ExpiredCount int        `json:"expired_count"`
	RequireAuth  bool       `json:"require_auth"`
}

// ShareTextResult 分享文本结果
type ShareTextResult struct {
	Code      string     `json:"code"`
	ShareURL  string     `json:"share_url"`
	ExpiredAt *time.Time `json:"expired_at"`
}

// ShareFileResult 分享文件结果
type ShareFileResult struct {
	Code      string     `json:"code"`
	ShareURL  string     `json:"share_url"`
	FileName  string     `json:"file_name"`
	ExpiredAt *time.Time `json:"expired_at"`
}
