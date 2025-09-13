package storage

import (
	"mime/multipart"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zy84338719/filecodebox/internal/models"
)

// FileOperationResult 文件操作结果
type FileOperationResult struct {
	Success   bool                   `json:"success"`
	Message   string                 `json:"message,omitempty"`
	Error     error                  `json:"-"`
	FilePath  string                 `json:"file_path,omitempty"`
	FileSize  int64                  `json:"file_size,omitempty"`
	FileHash  string                 `json:"file_hash,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// ChunkOperationResult 分片操作结果
type ChunkOperationResult struct {
	Success     bool                   `json:"success"`
	Message     string                 `json:"message,omitempty"`
	Error       error                  `json:"-"`
	UploadID    string                 `json:"upload_id"`
	ChunkIndex  int                    `json:"chunk_index,omitempty"`
	ChunkHash   string                 `json:"chunk_hash,omitempty"`
	ChunkSize   int                    `json:"chunk_size,omitempty"`
	TotalChunks int                    `json:"total_chunks,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// FileDownloadInfo 文件下载信息
type FileDownloadInfo struct {
	FilePath     string                 `json:"file_path"`
	FileName     string                 `json:"file_name"`
	FileSize     int64                  `json:"file_size"`
	ContentType  string                 `json:"content_type"`
	DownloadURL  string                 `json:"download_url,omitempty"`
	DirectAccess bool                   `json:"direct_access"` // 是否支持直接访问
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// StorageInfo 存储信息
type StorageInfo struct {
	Type        string                 `json:"type"`
	Available   bool                   `json:"available"`
	Connected   bool                   `json:"connected"`
	TotalSpace  int64                  `json:"total_space,omitempty"`
	UsedSpace   int64                  `json:"used_space,omitempty"`
	FreeSpace   int64                  `json:"free_space,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
	LastChecked time.Time              `json:"last_checked"`
}

// FileGenerationInfo 文件生成信息
type FileGenerationInfo struct {
	Path         string `json:"path"`          // 相对路径(如: 2024/01/15)
	Suffix       string `json:"suffix"`        // 文件扩展名
	Prefix       string `json:"prefix"`        // 原始文件名(不含扩展名)
	UUIDFileName string `json:"uuid_filename"` // 生成的唯一文件名
	SavePath     string `json:"save_path"`     // 完整保存路径
	FullPath     string `json:"full_path"`     // 绝对路径
}

// StorageOperations 存储操作接口(保留必要的interface，但简化)
type StorageOperations interface {
	// 文件操作
	SaveFileWithResult(file *multipart.FileHeader, savePath string) *FileOperationResult
	DeleteFileWithResult(fileCode *models.FileCode) *FileOperationResult
	GetFileDownloadInfo(fileCode *models.FileCode) (*FileDownloadInfo, error)

	// 分片操作
	SaveChunkWithResult(uploadID string, chunkIndex int, data []byte, chunkHash string) *ChunkOperationResult
	MergeChunksWithResult(uploadID string, chunk *models.UploadChunk, savePath string) *FileOperationResult
	CleanChunksWithResult(uploadID string) *FileOperationResult

	// 连接测试
	TestConnectionWithResult() *StorageInfo

	// 传统兼容方法(逐步废弃)
	GetFileResponse(c *gin.Context, fileCode *models.FileCode) error
}
