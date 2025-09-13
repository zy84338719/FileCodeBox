package web

import "time"

// FileInfo 文件信息
type FileInfo struct {
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

// ShareTextRequest 分享文本请求
type ShareTextRequest struct {
	Text        string `json:"text" binding:"required"`
	ExpireValue int    `json:"expire_value"`
	ExpireStyle string `json:"expire_style"`
	RequireAuth bool   `json:"require_auth"`
}

// ShareFileRequest 分享文件请求
type ShareFileRequest struct {
	ExpireValue int    `json:"expire_value" form:"expire_value"`
	ExpireStyle string `json:"expire_style" form:"expire_style"`
	RequireAuth bool   `json:"require_auth" form:"require_auth"`
}

// ShareCodeRequest 分享代码请求
type ShareCodeRequest struct {
	Code string `json:"code" form:"code" binding:"required"`
}

// ShareResponse 分享响应
type ShareResponse struct {
	Code      string     `json:"code"`
	ShareURL  string     `json:"share_url"`
	FileName  string     `json:"file_name,omitempty"`
	ExpiredAt *time.Time `json:"expired_at"`
}

// FileInfoResponse 文件信息响应
type FileInfoResponse struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	Text        string `json:"text,omitempty"` // 文本内容或下载链接
	UploadType  string `json:"upload_type"`
	RequireAuth bool   `json:"require_auth"`
}

// ShareStatsResponse 分享统计响应
type ShareStatsResponse struct {
	Code         string     `json:"code"`
	FileName     string     `json:"file_name"`
	FileSize     int64      `json:"file_size"`
	UsedCount    int        `json:"used_count"`
	ExpiredCount int        `json:"expired_count"`
	ExpiredAt    *time.Time `json:"expired_at"`
	UploadType   string     `json:"upload_type"`
	RequireAuth  bool       `json:"require_auth"`
	IsExpired    bool       `json:"is_expired"`
	CreatedAt    time.Time  `json:"created_at"`
}
