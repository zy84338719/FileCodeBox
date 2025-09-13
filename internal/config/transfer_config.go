// Package config 上传下载配置模块
package config

import (
	"fmt"
	"strconv"
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

// ToMap 上传配置转换为map格式
func (uc *UploadConfig) ToMap() map[string]string {
	return map[string]string{
		"open_upload":      fmt.Sprintf("%d", uc.OpenUpload),
		"upload_size":      fmt.Sprintf("%d", uc.UploadSize),
		"enable_chunk":     fmt.Sprintf("%d", uc.EnableChunk),
		"chunk_size":       fmt.Sprintf("%d", uc.ChunkSize),
		"max_save_seconds": fmt.Sprintf("%d", uc.MaxSaveSeconds),
	}
}

// ToMap 下载配置转换为map格式
func (dc *DownloadConfig) ToMap() map[string]string {
	return map[string]string{
		"enable_concurrent_download": fmt.Sprintf("%d", dc.EnableConcurrentDownload),
		"max_concurrent_downloads":   fmt.Sprintf("%d", dc.MaxConcurrentDownloads),
		"download_timeout":           fmt.Sprintf("%d", dc.DownloadTimeout),
	}
}

// ToMap 传输配置转换为map格式
func (tc *TransferConfig) ToMap() map[string]string {
	result := make(map[string]string)

	// 合并上传配置
	for k, v := range tc.Upload.ToMap() {
		result[k] = v
	}

	// 合并下载配置
	for k, v := range tc.Download.ToMap() {
		result[k] = v
	}

	return result
}

// FromMap 从map加载上传配置
func (uc *UploadConfig) FromMap(data map[string]string) error {
	if val, ok := data["open_upload"]; ok {
		if v, err := strconv.Atoi(val); err == nil {
			uc.OpenUpload = v
		}
	}
	if val, ok := data["upload_size"]; ok {
		if v, err := strconv.ParseInt(val, 10, 64); err == nil {
			uc.UploadSize = v
		}
	}
	if val, ok := data["enable_chunk"]; ok {
		if v, err := strconv.Atoi(val); err == nil {
			uc.EnableChunk = v
		}
	}
	if val, ok := data["chunk_size"]; ok {
		if v, err := strconv.ParseInt(val, 10, 64); err == nil {
			uc.ChunkSize = v
		}
	}
	if val, ok := data["max_save_seconds"]; ok {
		if v, err := strconv.Atoi(val); err == nil {
			uc.MaxSaveSeconds = v
		}
	}

	return uc.Validate()
}

// FromMap 从map加载下载配置
func (dc *DownloadConfig) FromMap(data map[string]string) error {
	if val, ok := data["enable_concurrent_download"]; ok {
		if v, err := strconv.Atoi(val); err == nil {
			dc.EnableConcurrentDownload = v
		}
	}
	if val, ok := data["max_concurrent_downloads"]; ok {
		if v, err := strconv.Atoi(val); err == nil {
			dc.MaxConcurrentDownloads = v
		}
	}
	if val, ok := data["download_timeout"]; ok {
		if v, err := strconv.Atoi(val); err == nil {
			dc.DownloadTimeout = v
		}
	}

	return dc.Validate()
}

// FromMap 从map加载传输配置
func (tc *TransferConfig) FromMap(data map[string]string) error {
	if err := tc.Upload.FromMap(data); err != nil {
		return err
	}
	return tc.Download.FromMap(data)
}

// Update 更新上传配置
func (uc *UploadConfig) Update(updates map[string]interface{}) error {
	if openUpload, ok := updates["open_upload"].(int); ok {
		uc.OpenUpload = openUpload
	}
	if uploadSize, ok := updates["upload_size"].(int64); ok {
		uc.UploadSize = uploadSize
	}
	if enableChunk, ok := updates["enable_chunk"].(int); ok {
		uc.EnableChunk = enableChunk
	}
	if chunkSize, ok := updates["chunk_size"].(int64); ok {
		uc.ChunkSize = chunkSize
	}
	if maxSaveSeconds, ok := updates["max_save_seconds"].(int); ok {
		uc.MaxSaveSeconds = maxSaveSeconds
	}

	return uc.Validate()
}

// Update 更新下载配置
func (dc *DownloadConfig) Update(updates map[string]interface{}) error {
	if enableConcurrent, ok := updates["enable_concurrent_download"].(int); ok {
		dc.EnableConcurrentDownload = enableConcurrent
	}
	if maxConcurrent, ok := updates["max_concurrent_downloads"].(int); ok {
		dc.MaxConcurrentDownloads = maxConcurrent
	}
	if timeout, ok := updates["download_timeout"].(int); ok {
		dc.DownloadTimeout = timeout
	}

	return dc.Validate()
}

// Update 更新传输配置
func (tc *TransferConfig) Update(updates map[string]interface{}) error {
	if err := tc.Upload.Update(updates); err != nil {
		return err
	}
	return tc.Download.Update(updates)
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
