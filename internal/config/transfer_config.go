// Package config 上传下载配置模块
package config

import (
	"fmt"
	"strings"
)

// UploadConfig 上传配置
type UploadConfig struct {
	OpenUpload     int   `json:"open_upload"`      // 是否开启上传 0-禁用 1-启用
	UploadSize     int64 `json:"upload_size"`      // 上传文件大小限制（字节）
	EnableChunk    int   `json:"enable_chunk"`     // 是否启用分片上传 0-禁用 1-启用
	ChunkSize      int64 `json:"chunk_size"`       // 分片大小（字节）
	MaxSaveSeconds int   `json:"max_save_seconds"` // 最大保存时间（秒）
}

// DownloadConfig 下载配置
type DownloadConfig struct {
	EnableConcurrentDownload int `json:"enable_concurrent_download"` // 是否启用并发下载
	MaxConcurrentDownloads   int `json:"max_concurrent_downloads"`   // 最大并发下载数
	DownloadTimeout          int `json:"download_timeout"`           // 下载超时时间(秒)
}

// TransferConfig 文件传输配置（包含上传和下载）
type TransferConfig struct {
	Upload   *UploadConfig   `json:"upload"`
	Download *DownloadConfig `json:"download"`
}

// NewUploadConfig 创建上传配置
func NewUploadConfig() *UploadConfig {
	return &UploadConfig{
		OpenUpload:     1,
		UploadSize:     10 * 1024 * 1024, // 10MB
		EnableChunk:    0,
		ChunkSize:      2 * 1024 * 1024, // 2MB
		MaxSaveSeconds: 0,               // 0表示不限制
	}
}

// NewDownloadConfig 创建下载配置
func NewDownloadConfig() *DownloadConfig {
	return &DownloadConfig{
		EnableConcurrentDownload: 1,   // 默认启用
		MaxConcurrentDownloads:   10,  // 最大10个并发
		DownloadTimeout:          300, // 5分钟超时
	}
}

// NewTransferConfig 创建传输配置
func NewTransferConfig() *TransferConfig {
	return &TransferConfig{
		Upload:   NewUploadConfig(),
		Download: NewDownloadConfig(),
	}
}

// Validate 验证上传配置
func (uc *UploadConfig) Validate() error {
	var errors []string

	// 验证上传大小限制
	if uc.UploadSize < 0 {
		errors = append(errors, "上传文件大小限制不能为负数")
	}
	if uc.UploadSize > 10*1024*1024*1024 { // 10GB
		errors = append(errors, "上传文件大小限制不能超过10GB")
	}

	// 验证分片大小
	if uc.EnableChunk == 1 {
		if uc.ChunkSize < 1024*1024 { // 1MB最小分片
			errors = append(errors, "分片大小不能小于1MB")
		}
		if uc.ChunkSize > 100*1024*1024 { // 100MB最大分片
			errors = append(errors, "分片大小不能超过100MB")
		}
		if uc.ChunkSize > uc.UploadSize {
			errors = append(errors, "分片大小不能超过上传文件大小限制")
		}
	}

	// 验证保存时间
	if uc.MaxSaveSeconds < 0 {
		errors = append(errors, "最大保存时间不能为负数")
	}

	if len(errors) > 0 {
		return fmt.Errorf("上传配置验证失败: %s", strings.Join(errors, "; "))
	}

	return nil
}

// Validate 验证下载配置
func (dc *DownloadConfig) Validate() error {
	var errors []string

	// 验证并发下载数
	if dc.MaxConcurrentDownloads < 1 {
		errors = append(errors, "最大并发下载数必须大于0")
	}
	if dc.MaxConcurrentDownloads > 100 {
		errors = append(errors, "最大并发下载数不能超过100")
	}

	// 验证下载超时时间
	if dc.DownloadTimeout < 30 {
		errors = append(errors, "下载超时时间不能小于30秒")
	}
	if dc.DownloadTimeout > 3600 {
		errors = append(errors, "下载超时时间不能超过1小时")
	}

	if len(errors) > 0 {
		return fmt.Errorf("下载配置验证失败: %s", strings.Join(errors, "; "))
	}

	return nil
}

// Validate 验证传输配置
func (tc *TransferConfig) Validate() error {
	if err := tc.Upload.Validate(); err != nil {
		return err
	}
	return tc.Download.Validate()
}

// IsUploadEnabled 判断是否启用上传
func (uc *UploadConfig) IsUploadEnabled() bool {
	return uc.OpenUpload == 1
}

// IsChunkEnabled 判断是否启用分片上传
func (uc *UploadConfig) IsChunkEnabled() bool {
	return uc.EnableChunk == 1
}

// GetUploadSizeMB 获取上传大小限制（MB）
func (uc *UploadConfig) GetUploadSizeMB() float64 {
	return float64(uc.UploadSize) / (1024 * 1024)
}

// GetChunkSizeMB 获取分片大小（MB）
func (uc *UploadConfig) GetChunkSizeMB() float64 {
	return float64(uc.ChunkSize) / (1024 * 1024)
}

// GetMaxSaveHours 获取最大保存时间（小时）
func (uc *UploadConfig) GetMaxSaveHours() float64 {
	if uc.MaxSaveSeconds == 0 {
		return 0 // 不限制
	}
	return float64(uc.MaxSaveSeconds) / 3600
}

// IsDownloadConcurrentEnabled 判断是否启用并发下载
func (dc *DownloadConfig) IsDownloadConcurrentEnabled() bool {
	return dc.EnableConcurrentDownload == 1
}

// GetDownloadTimeoutMinutes 获取下载超时时间（分钟）
func (dc *DownloadConfig) GetDownloadTimeoutMinutes() float64 {
	return float64(dc.DownloadTimeout) / 60
}

// Clone 克隆上传配置
func (uc *UploadConfig) Clone() *UploadConfig {
	return &UploadConfig{
		OpenUpload:     uc.OpenUpload,
		UploadSize:     uc.UploadSize,
		EnableChunk:    uc.EnableChunk,
		ChunkSize:      uc.ChunkSize,
		MaxSaveSeconds: uc.MaxSaveSeconds,
	}
}

// Clone 克隆下载配置
func (dc *DownloadConfig) Clone() *DownloadConfig {
	return &DownloadConfig{
		EnableConcurrentDownload: dc.EnableConcurrentDownload,
		MaxConcurrentDownloads:   dc.MaxConcurrentDownloads,
		DownloadTimeout:          dc.DownloadTimeout,
	}
}

// Clone 克隆传输配置
func (tc *TransferConfig) Clone() *TransferConfig {
	return &TransferConfig{
		Upload:   tc.Upload.Clone(),
		Download: tc.Download.Clone(),
	}
}
