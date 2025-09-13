package service

import "time"

// ChunkUploadInitData 分片上传初始化数据
type ChunkUploadInitData struct {
	UploadID      string  `json:"upload_id"`
	TotalChunks   int     `json:"total_chunks"`
	ChunkSize     int     `json:"chunk_size"`
	FileName      string  `json:"file_name"`
	FileSize      int64   `json:"file_size"`
	FileHash      string  `json:"file_hash"`
	UploadedCount int     `json:"uploaded_count"`
	Progress      float64 `json:"progress"`
	Status        string  `json:"status"`
}

// ChunkUploadStatusData 分片上传状态数据
type ChunkUploadStatusData struct {
	UploadID      string    `json:"upload_id"`
	Status        string    `json:"status"`
	Progress      float64   `json:"progress"`
	TotalChunks   int       `json:"total_chunks"`
	UploadedCount int       `json:"uploaded_count"`
	FileName      string    `json:"file_name"`
	FileSize      int64     `json:"file_size"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ChunkUploadCompleteData 分片上传完成数据
type ChunkUploadCompleteData struct {
	Code      string     `json:"code"`
	ShareURL  string     `json:"share_url"`
	FileName  string     `json:"file_name"`
	FileSize  int64      `json:"file_size"`
	ExpiredAt *time.Time `json:"expired_at"`
}

// ChunkUploadProgressResponse 分块上传进度响应
type ChunkUploadProgressResponse struct {
	UploadID string `json:"upload_id"`
	Progress int    `json:"progress"`
	Status   string `json:"status"`
}

// ChunkUploadStatusResponse 分块上传状态响应
type ChunkUploadStatusResponse struct {
	UploadID       string `json:"upload_id"`
	Status         string `json:"status"`
	UploadedChunks []int  `json:"uploaded_chunks"`
	TotalChunks    int    `json:"total_chunks"`
	Progress       int    `json:"progress"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

// ChunkVerifyResponse 分块验证响应
type ChunkVerifyResponse struct {
	Verified bool   `json:"verified"`
	Message  string `json:"message"`
}

// ChunkUploadCompleteResponse 分块上传完成响应
type ChunkUploadCompleteResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	FileCode string `json:"file_code,omitempty"`
	ShareURL string `json:"share_url,omitempty"`
}
