package db

import (
	"time"

	"gorm.io/gorm"
)

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

// ChunkQuery 分片查询条件
type ChunkQuery struct {
	gorm.Model
	UploadID   string `json:"upload_id"`
	ChunkIndex *int   `json:"chunk_index"`
	Status     string `json:"status"`
	Completed  *bool  `json:"completed"`
	Limit      int    `json:"limit"`
	Offset     int    `json:"offset"`
}

// ChunkUpdate 分片更新数据
type ChunkUpdate struct {
	gorm.Model
	ChunkHash  *string    `json:"chunk_hash"`
	Completed  *bool      `json:"completed"`
	RetryCount *int       `json:"retry_count"`
	LastError  *string    `json:"last_error"`
	Status     *string    `json:"status"`
	UpdatedAt  *time.Time `json:"updated_at"`
}

// ChunkStats 分片统计查询结果
type ChunkStats struct {
	gorm.Model
	TotalChunks     int64 `json:"total_chunks"`
	CompletedChunks int64 `json:"completed_chunks"`
	PendingChunks   int64 `json:"pending_chunks"`
	FailedChunks    int64 `json:"failed_chunks"`
}
