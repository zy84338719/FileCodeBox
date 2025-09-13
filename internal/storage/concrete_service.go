package storage

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/models"
)

// ConcreteStorageService 具体的存储服务实现
type ConcreteStorageService struct {
	pathManager    *PathManager
	localStrategy  *LocalStorageStrategy
	webdavStrategy *WebDAVStorageStrategy
	s3Strategy     *S3StorageStrategy
	nfsStrategy    *NFSStorageStrategy

	currentType string
	config      *config.ConfigManager
}

// NewConcreteStorageService 创建具体的存储服务
func NewConcreteStorageService(manager *config.ConfigManager) *ConcreteStorageService {
	// 创建 PathManager
	basePath := manager.Storage.StoragePath
	if basePath == "" {
		basePath = filepath.Join(manager.Base.DataPath)
	}
	pathManager := NewPathManager(basePath)

	service := &ConcreteStorageService{
		pathManager: pathManager,
		currentType: manager.Storage.Type,
		config:      manager,
	}

	// 初始化所有策略
	service.initializeStrategies(manager)

	return service
}

// initializeStrategies 初始化所有存储策略
func (css *ConcreteStorageService) initializeStrategies(manager *config.ConfigManager) {
	// 初始化本地存储
	css.localStrategy = NewLocalStorageStrategy(css.pathManager.basePath)

	// 初始化 WebDAV 存储
	if manager.Storage.WebDAV != nil && manager.Storage.WebDAV.Hostname != "" {
		strategy, err := NewWebDAVStorageStrategy(
			manager.Storage.WebDAV.Hostname,
			manager.Storage.WebDAV.RootPath,
			manager.Storage.WebDAV.Username,
			manager.Storage.WebDAV.Password,
		)
		if err == nil {
			css.webdavStrategy = strategy
		}
	}

	// 初始化 S3 存储
	if manager.Storage.S3 != nil && manager.Storage.S3.AccessKeyID != "" &&
		manager.Storage.S3.SecretAccessKey != "" && manager.Storage.S3.BucketName != "" {
		strategy, err := NewS3StorageStrategy(
			manager.Storage.S3.AccessKeyID,
			manager.Storage.S3.SecretAccessKey,
			manager.Storage.S3.BucketName,
			manager.Storage.S3.EndpointURL,
			manager.Storage.S3.RegionName,
			manager.Storage.S3.SessionToken,
			manager.Storage.S3.Hostname,
			manager.Storage.S3.Proxy == 1,
			"filebox_storage",
		)
		if err == nil {
			css.s3Strategy = strategy
		}
	}

	// 初始化 NFS 存储
	if manager.Storage.NFS != nil && manager.Storage.NFS.Server != "" {
		strategy, err := NewNFSStorageStrategy(
			manager.Storage.NFS.Server,
			manager.Storage.NFS.Path,
			manager.Storage.NFS.MountPoint,
			manager.Storage.NFS.Version,
			manager.Storage.NFS.Options,
			manager.Storage.NFS.Timeout,
			manager.Storage.NFS.AutoMount == 1,
			manager.Storage.NFS.RetryCount,
			manager.Storage.NFS.SubPath,
		)
		if err == nil {
			css.nfsStrategy = strategy
		}
	}
}

// getCurrentStrategy 获取当前存储策略
func (css *ConcreteStorageService) getCurrentStrategy() StorageStrategy {
	switch css.currentType {
	case "webdav":
		if css.webdavStrategy != nil {
			return css.webdavStrategy
		}
	case "s3":
		if css.s3Strategy != nil {
			return css.s3Strategy
		}
	case "nfs":
		if css.nfsStrategy != nil {
			return css.nfsStrategy
		}
	default: // "local"
		return css.localStrategy
	}

	// 默认返回本地存储
	return css.localStrategy
}

// SaveFileWithResult 保存文件并返回结果
func (css *ConcreteStorageService) SaveFileWithResult(file *multipart.FileHeader, savePath string) *FileOperationResult {
	result := &FileOperationResult{
		Timestamp: time.Now(),
		FilePath:  savePath,
		FileSize:  file.Size,
	}

	strategy := css.getCurrentStrategy()
	operator := NewStorageOperator(strategy, css.pathManager)

	err := operator.SaveFile(file, savePath)
	if err != nil {
		result.Success = false
		result.Error = err
		result.Message = fmt.Sprintf("保存文件失败: %v", err)
		return result
	}

	// 计算文件哈希
	if hash, err := CalculateFileHash(file); err == nil {
		result.FileHash = hash
	}

	result.Success = true
	result.Message = "文件保存成功"
	result.Metadata = map[string]interface{}{
		"storage_type": css.currentType,
		"strategy":     fmt.Sprintf("%T", strategy),
	}

	return result
}

// DeleteFileWithResult 删除文件并返回结果
func (css *ConcreteStorageService) DeleteFileWithResult(fileCode *models.FileCode) *FileOperationResult {
	result := &FileOperationResult{
		Timestamp: time.Now(),
		FilePath:  fileCode.FilePath,
	}

	strategy := css.getCurrentStrategy()
	operator := NewStorageOperator(strategy, css.pathManager)

	err := operator.DeleteFile(fileCode)
	if err != nil {
		result.Success = false
		result.Error = err
		result.Message = fmt.Sprintf("删除文件失败: %v", err)
		return result
	}

	result.Success = true
	result.Message = "文件删除成功"
	result.Metadata = map[string]interface{}{
		"storage_type": css.currentType,
		"file_code":    fileCode.Code,
	}

	return result
}

// GetFileDownloadInfo 获取文件下载信息
func (css *ConcreteStorageService) GetFileDownloadInfo(fileCode *models.FileCode) (*FileDownloadInfo, error) {
	strategy := css.getCurrentStrategy()
	operator := NewStorageOperator(strategy, css.pathManager)

	// 尝试获取直接URL
	url, err := operator.GetFileURL(fileCode)

	info := &FileDownloadInfo{
		FilePath:     fileCode.FilePath,
		FileName:     fileCode.Prefix + fileCode.Suffix,
		FileSize:     fileCode.Size,
		ContentType:  "application/octet-stream", // 可以根据文件扩展名推断
		DirectAccess: err == nil && url != "",
		DownloadURL:  url,
		Metadata: map[string]interface{}{
			"storage_type": css.currentType,
			"upload_time":  fileCode.CreatedAt,
		},
	}

	return info, nil
}

// SaveChunkWithResult 保存分片并返回结果
func (css *ConcreteStorageService) SaveChunkWithResult(uploadID string, chunkIndex int, data []byte, chunkHash string) *ChunkOperationResult {
	result := &ChunkOperationResult{
		Timestamp:  time.Now(),
		UploadID:   uploadID,
		ChunkIndex: chunkIndex,
		ChunkHash:  chunkHash,
		ChunkSize:  len(data),
	}

	strategy := css.getCurrentStrategy()
	operator := NewStorageOperator(strategy, css.pathManager)

	err := operator.SaveChunk(uploadID, chunkIndex, data, chunkHash)
	if err != nil {
		result.Success = false
		result.Error = err
		result.Message = fmt.Sprintf("保存分片失败: %v", err)
		return result
	}

	result.Success = true
	result.Message = "分片保存成功"
	result.Metadata = map[string]interface{}{
		"storage_type": css.currentType,
	}

	return result
}

// MergeChunksWithResult 合并分片并返回结果
func (css *ConcreteStorageService) MergeChunksWithResult(uploadID string, chunk *models.UploadChunk, savePath string) *FileOperationResult {
	result := &FileOperationResult{
		Timestamp: time.Now(),
		FilePath:  savePath,
	}

	strategy := css.getCurrentStrategy()
	operator := NewStorageOperator(strategy, css.pathManager)

	err := operator.MergeChunks(uploadID, chunk, savePath)
	if err != nil {
		result.Success = false
		result.Error = err
		result.Message = fmt.Sprintf("合并分片失败: %v", err)
		return result
	}

	result.Success = true
	result.Message = "分片合并成功"
	result.FileSize = chunk.FileSize
	result.Metadata = map[string]interface{}{
		"storage_type":  css.currentType,
		"total_chunks":  chunk.TotalChunks,
		"original_name": chunk.FileName,
	}

	return result
}

// CleanChunksWithResult 清理分片并返回结果
func (css *ConcreteStorageService) CleanChunksWithResult(uploadID string) *FileOperationResult {
	result := &FileOperationResult{
		Timestamp: time.Now(),
	}

	strategy := css.getCurrentStrategy()
	operator := NewStorageOperator(strategy, css.pathManager)

	err := operator.CleanChunks(uploadID)
	if err != nil {
		result.Success = false
		result.Error = err
		result.Message = fmt.Sprintf("清理分片失败: %v", err)
		return result
	}

	result.Success = true
	result.Message = "分片清理成功"
	result.Metadata = map[string]interface{}{
		"storage_type": css.currentType,
		"upload_id":    uploadID,
	}

	return result
}

// TestConnectionWithResult 测试连接并返回结果
func (css *ConcreteStorageService) TestConnectionWithResult() *StorageInfo {
	info := &StorageInfo{
		Type:        css.currentType,
		LastChecked: time.Now(),
	}

	strategy := css.getCurrentStrategy()

	err := strategy.TestConnection()
	info.Connected = err == nil
	info.Available = err == nil

	if err != nil {
		info.Config = map[string]interface{}{
			"error": err.Error(),
		}
	} else {
		info.Config = map[string]interface{}{
			"status": "healthy",
		}
	}

	return info
}

// GetFileResponse 兼容旧接口
func (css *ConcreteStorageService) GetFileResponse(c *gin.Context, fileCode *models.FileCode) error {
	strategy := css.getCurrentStrategy()
	operator := NewStorageOperator(strategy, css.pathManager)
	return operator.GetFileResponse(c, fileCode)
}

// SwitchStorageType 切换存储类型
func (css *ConcreteStorageService) SwitchStorageType(storageType string) error {
	// 验证存储类型是否可用
	switch storageType {
	case "local":
		// 本地存储总是可用
	case "webdav":
		if css.webdavStrategy == nil {
			return fmt.Errorf("WebDAV存储未配置")
		}
	case "s3":
		if css.s3Strategy == nil {
			return fmt.Errorf("S3存储未配置")
		}
	case "nfs":
		if css.nfsStrategy == nil {
			return fmt.Errorf("NFS存储未配置")
		}
	default:
		return fmt.Errorf("不支持的存储类型: %s", storageType)
	}

	css.currentType = storageType
	return nil
}

// GetAvailableStorageTypes 获取可用的存储类型
func (css *ConcreteStorageService) GetAvailableStorageTypes() []string {
	types := []string{"local"} // 本地存储总是可用

	if css.webdavStrategy != nil {
		types = append(types, "webdav")
	}
	if css.s3Strategy != nil {
		types = append(types, "s3")
	}
	if css.nfsStrategy != nil {
		types = append(types, "nfs")
	}

	return types
}

// GetCurrentStorageType 获取当前存储类型
func (css *ConcreteStorageService) GetCurrentStorageType() string {
	return css.currentType
}

// GenerateFileInfo 生成文件信息
func (css *ConcreteStorageService) GenerateFileInfo(fileName string, uploadID string) *FileGenerationInfo {
	path, suffix, prefix, uuidFileName := GenerateFileInfo(fileName, uploadID)
	savePath := css.pathManager.GetDateBasedPath(uuidFileName)
	fullPath := filepath.Join(css.pathManager.basePath, savePath)

	return &FileGenerationInfo{
		Path:         path,
		Suffix:       suffix,
		Prefix:       prefix,
		UUIDFileName: uuidFileName,
		SavePath:     savePath,
		FullPath:     fullPath,
	}
}
