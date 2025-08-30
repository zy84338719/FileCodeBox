// Package models 定义应用程序的数据模型
package models

import (
	"fmt"
	"path/filepath"
	"time"

	"gorm.io/gorm"
)

// FileCode 文件代码模型
type FileCode struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	Code         string         `gorm:"uniqueIndex;size:255" json:"code"`
	Prefix       string         `gorm:"size:255" json:"prefix"`
	Suffix       string         `gorm:"size:255" json:"suffix"`
	UUIDFileName string         `gorm:"size:255" json:"uuid_file_name"`
	FilePath     string         `gorm:"size:255" json:"file_path"`
	Size         int64          `gorm:"default:0" json:"size"`
	Text         string         `gorm:"type:text" json:"text"`
	ExpiredAt    *time.Time     `json:"expired_at"`
	ExpiredCount int            `gorm:"default:0" json:"expired_count"`
	UsedCount    int            `gorm:"default:0" json:"used_count"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	FileHash     string         `gorm:"size:64" json:"file_hash"`
	IsChunked    bool           `gorm:"default:false" json:"is_chunked"`
	UploadID     string         `gorm:"size:36" json:"upload_id"`
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
	ID        uint           `gorm:"primarykey" json:"id"`
	Key       string         `gorm:"uniqueIndex;size:255" json:"key"`
	Value     string         `gorm:"type:text" json:"value"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}
