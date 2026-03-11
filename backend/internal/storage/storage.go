package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/zy84338719/fileCodeBox/internal/repo/db/model"
)

// StorageType 存储类型
type StorageType string

const (
	StorageTypeLocal    StorageType = "local"
	StorageTypeS3       StorageType = "s3"
	StorageTypeWebDAV   StorageType = "webdav"
	StorageTypeOneDrive StorageType = "onedrive"
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

// StorageInterface 存储接口（简化版）
type StorageInterface interface {
	// 基础操作
	SaveFile(ctx context.Context, file *multipart.FileHeader, savePath string) (*FileOperationResult, error)
	DeleteFile(ctx context.Context, filePath string) error
	GetFile(ctx context.Context, filePath string) ([]byte, error)
	FileExists(ctx context.Context, filePath string) bool

	// 分片操作
	SaveChunk(ctx context.Context, uploadID string, chunkIndex int, data []byte) error
	MergeChunks(ctx context.Context, uploadID string, totalChunks int, savePath string) error
	CleanChunks(ctx context.Context, uploadID string) error

	// 工具方法
	GetFileSize(ctx context.Context, filePath string) (int64, error)
	GetFileURL(ctx context.Context, filePath string) (string, error)

	// 流式下载方法
	GetFileReader(ctx context.Context, filePath string) (io.ReadCloser, int64, error)
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type     StorageType
	DataPath string // 本地存储路径
	BaseURL  string // 基础URL

	// S3 配置
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	Region    string

	// WebDAV 配置
	WebDAVURL      string
	WebDAVUsername string
	WebDAVPassword string
}

// StorageService 存储服务
type StorageService struct {
	config *StorageConfig
}

// NewStorageService 创建存储服务
func NewStorageService(config *StorageConfig) *StorageService {
	return &StorageService{config: config}
}

// SaveFile 保存文件
func (s *StorageService) SaveFile(ctx context.Context, file *multipart.FileHeader, savePath string) (*FileOperationResult, error) {
	startTime := time.Now()

	// 确保目录存在
	fullPath := filepath.Join(s.config.DataPath, savePath)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建目录失败: %w", err)
	}

	// 打开上传的文件
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("打开上传文件失败: %w", err)
	}
	defer src.Close()

	// 创建目标文件
	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer dst.Close()

	// 复制文件内容
	written, err := io.Copy(dst, src)
	if err != nil {
		return nil, fmt.Errorf("保存文件失败: %w", err)
	}

	return &FileOperationResult{
		Success:   true,
		Message:   "文件保存成功",
		FilePath:  savePath,
		FileSize:  written,
		Timestamp: startTime,
	}, nil
}

// DeleteFile 删除文件
func (s *StorageService) DeleteFile(ctx context.Context, filePath string) error {
	fullPath := filepath.Join(s.config.DataPath, filePath)

	if !s.FileExists(ctx, filePath) {
		return fmt.Errorf("文件不存在")
	}

	return os.Remove(fullPath)
}

// GetFile 获取文件内容
func (s *StorageService) GetFile(ctx context.Context, filePath string) ([]byte, error) {
	fullPath := filepath.Join(s.config.DataPath, filePath)
	return os.ReadFile(fullPath)
}

// FileExists 检查文件是否存在
func (s *StorageService) FileExists(ctx context.Context, filePath string) bool {
	fullPath := filepath.Join(s.config.DataPath, filePath)
	_, err := os.Stat(fullPath)
	return !os.IsNotExist(err)
}

// SaveChunk 保存分片
func (s *StorageService) SaveChunk(ctx context.Context, uploadID string, chunkIndex int, data []byte) error {
	chunkPath := filepath.Join(s.config.DataPath, "chunks", uploadID, fmt.Sprintf("chunk_%d", chunkIndex))

	// 确保目录存在
	dir := filepath.Dir(chunkPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(chunkPath, data, 0644)
}

// MergeChunks 合并分片
func (s *StorageService) MergeChunks(ctx context.Context, uploadID string, totalChunks int, savePath string) error {
	fullPath := filepath.Join(s.config.DataPath, savePath)

	// 确保目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 创建目标文件
	dst, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	// 按顺序合并所有分片
	for i := 0; i < totalChunks; i++ {
		chunkPath := filepath.Join(s.config.DataPath, "chunks", uploadID, fmt.Sprintf("chunk_%d", i))
		chunkData, err := os.ReadFile(chunkPath)
		if err != nil {
			return fmt.Errorf("读取分片 %d 失败: %w", i, err)
		}

		if _, err := dst.Write(chunkData); err != nil {
			return fmt.Errorf("写入分片 %d 失败: %w", i, err)
		}
	}

	// 清理临时分片
	go s.CleanChunks(context.Background(), uploadID)

	return nil
}

// CleanChunks 清理分片
func (s *StorageService) CleanChunks(ctx context.Context, uploadID string) error {
	chunkDir := filepath.Join(s.config.DataPath, "chunks", uploadID)
	return os.RemoveAll(chunkDir)
}

// GetFileSize 获取文件大小
func (s *StorageService) GetFileSize(ctx context.Context, filePath string) (int64, error) {
	fullPath := filepath.Join(s.config.DataPath, filePath)
	info, err := os.Stat(fullPath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// GetFileURL 获取文件URL
func (s *StorageService) GetFileURL(ctx context.Context, filePath string) (string, error) {
	if s.config.BaseURL == "" {
		return "", fmt.Errorf("base URL not configured")
	}
	return fmt.Sprintf("%s/files/%s", s.config.BaseURL, filePath), nil
}

// GetFileReader 获取文件读取器（用于流式下载）
func (s *StorageService) GetFileReader(ctx context.Context, filePath string) (io.ReadCloser, int64, error) {
	fullPath := filepath.Join(s.config.DataPath, filePath)

	// 检查文件是否存在
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		return nil, 0, fmt.Errorf("文件不存在: %w", err)
	}

	// 打开文件
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, 0, fmt.Errorf("打开文件失败: %w", err)
	}

	return file, fileInfo.Size(), nil
}

// GenerateFilePath 生成文件路径
func (s *StorageService) GenerateFilePath(fileCode *model.FileCode) string {
	now := time.Now()
	return filepath.Join(
		"uploads",
		now.Format("2006"),
		now.Format("01"),
		now.Format("02"),
		fileCode.UUIDFileName,
	)
}
