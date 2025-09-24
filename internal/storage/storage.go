package storage

import (
	"crypto/sha256"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// StorageInterface 存储接口
type StorageInterface interface {
	SaveFile(file *multipart.FileHeader, savePath string) error
	SaveChunk(uploadID string, chunkIndex int, data []byte, chunkHash string) error
	MergeChunks(uploadID string, chunk *models.UploadChunk, savePath string) error
	CleanChunks(uploadID string) error
	GetFileResponse(c *gin.Context, fileCode *models.FileCode) error
	GetFileURL(fileCode *models.FileCode) (string, error)
	DeleteFile(fileCode *models.FileCode) error
}

// StorageManager 存储管理器
type StorageManager struct {
	storages map[string]StorageInterface
	current  string
}

func NewStorageManager(manager *config.ConfigManager) *StorageManager {
	sm := &StorageManager{
		storages: make(map[string]StorageInterface),
		current:  manager.Storage.Type,
	}

	// 如果配置中的存储类型为空，默认使用本地存储
	if sm.current == "" {
		sm.current = "local"
	}

	// 创建 PathManager
	basePath := manager.Storage.StoragePath
	if basePath == "" {
		basePath = filepath.Join(manager.Base.DataPath)
	}
	pathManager := NewPathManager(basePath)

	// 注册本地存储 - 使用新的策略模式
	localStrategy := NewLocalStorageStrategy(basePath)
	sm.storages["local"] = NewStrategyBasedStorage(localStrategy, pathManager)

	// 注册 WebDAV 存储 - 使用新的策略模式
	if manager.Storage.WebDAV != nil && manager.Storage.WebDAV.Hostname != "" {
		webdavStrategy, err := NewWebDAVStorageStrategy(
			manager.Storage.WebDAV.Hostname,
			manager.Storage.WebDAV.RootPath,
			manager.Storage.WebDAV.Username,
			manager.Storage.WebDAV.Password,
		)
		if err == nil {
			sm.storages["webdav"] = NewStrategyBasedStorage(webdavStrategy, pathManager)
		}
	}

	// 注册 S3 存储 - 使用新的策略模式
	if manager.Storage.S3 != nil && manager.Storage.S3.AccessKeyID != "" && manager.Storage.S3.SecretAccessKey != "" && manager.Storage.S3.BucketName != "" {
		s3Strategy, err := NewS3StorageStrategy(
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
			sm.storages["s3"] = NewStrategyBasedStorage(s3Strategy, pathManager)
		}
	}

	// 注册 NFS 存储 - 使用新的策略模式
	if manager.Storage.NFS != nil && manager.Storage.NFS.Server != "" {
		nfsStrategy, err := NewNFSStorageStrategy(
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
			sm.storages["nfs"] = NewStrategyBasedStorage(nfsStrategy, pathManager)
		}
	}

	return sm
}

func (sm *StorageManager) GetStorage() StorageInterface {
	storage, exists := sm.storages[sm.current]
	if !exists {
		return sm.storages["local"] // 默认返回本地存储
	}
	return storage
}

// SwitchStorage 切换存储后端
func (sm *StorageManager) SwitchStorage(storageType string) error {
	if _, exists := sm.storages[storageType]; !exists {
		return fmt.Errorf("存储类型 %s 未注册", storageType)
	}
	sm.current = storageType
	return nil
}

// GetAvailableStorages 获取可用的存储类型
func (sm *StorageManager) GetAvailableStorages() []string {
	var types []string
	for storageType := range sm.storages {
		types = append(types, storageType)
	}
	return types
}

// GetCurrentStorage 获取当前存储类型
func (sm *StorageManager) GetCurrentStorage() string {
	return sm.current
}

// TestStorage 测试存储连接
func (sm *StorageManager) TestStorage(storageType string) error {
	storage, exists := sm.storages[storageType]
	if !exists {
		return fmt.Errorf("存储类型 %s 未注册", storageType)
	}

	// 如果存储实现了测试连接接口
	if tester, ok := storage.(interface{ TestConnection() error }); ok {
		return tester.TestConnection()
	}

	return nil // 默认认为连接正常
}

// ReconfigureWebDAV 重新配置 WebDAV 存储
func (sm *StorageManager) ReconfigureWebDAV(hostname, username, password, rootPath string) error {
	// 获取现有的存储适配器中的 PathManager
	var pathManager *PathManager
	if existingStorage, exists := sm.storages["webdav"]; exists {
		if strategyBased, ok := existingStorage.(*StrategyBasedStorage); ok {
			pathManager = strategyBased.operator.pathManager
		}
	}

	// 如果没有找到 PathManager，创建一个默认的
	if pathManager == nil {
		pathManager = NewPathManager("./data")
	}

	// 创建新的 WebDAV 策略
	webdavStrategy, err := NewWebDAVStorageStrategy(hostname, username, password, rootPath)
	if err != nil {
		return fmt.Errorf("创建 WebDAV 策略失败: %v", err)
	}

	// 重新注册 WebDAV 存储
	sm.storages["webdav"] = NewStrategyBasedStorage(webdavStrategy, pathManager)

	return nil
}

// ReconfigureNFS 重新配置 NFS 存储
func (sm *StorageManager) ReconfigureNFS(server, nfsPath, mountPoint, version, options string, timeout int, autoMount bool, retryCount int, subPath string) error {
	// 获取现有的存储适配器中的 PathManager
	var pathManager *PathManager
	if existingStorage, exists := sm.storages["nfs"]; exists {
		if strategyBased, ok := existingStorage.(*StrategyBasedStorage); ok {
			pathManager = strategyBased.operator.pathManager
		}
	}

	// 如果没有找到 PathManager，创建一个默认的
	if pathManager == nil {
		pathManager = NewPathManager("./data")
	}

	// 创建新的 NFS 策略
	nfsStrategy, err := NewNFSStorageStrategy(server, nfsPath, mountPoint, version, options, timeout, autoMount, retryCount, subPath)
	if err != nil {
		return fmt.Errorf("创建 NFS 策略失败: %v", err)
	}

	// 重新注册 NFS 存储
	sm.storages["nfs"] = NewStrategyBasedStorage(nfsStrategy, pathManager)

	return nil
}

// GetStorageInstance 获取指定类型的存储实例
func (sm *StorageManager) GetStorageInstance(storageType string) (StorageInterface, bool) {
	storage, exists := sm.storages[storageType]
	return storage, exists
}

// 工具函数

// GenerateFilePath 生成文件路径 (兼容旧代码)
func GenerateFilePath(fileName string, uploadID string, pathManager *PathManager) (path, suffix, prefix, uuidFileName, savePath string) {
	path, suffix, prefix, uuidFileName = GenerateFileInfo(fileName, uploadID)
	savePath = pathManager.GetDateBasedPath(uuidFileName)
	return
}

// GenerateFileInfo 生成文件信息 (不依赖PathManager)
func GenerateFileInfo(fileName string, uploadID string) (path, suffix, prefix, uuidFileName string) {
	// 计算文件哈希作为文件名
	hash := sha256.Sum256([]byte(fileName + uploadID))
	hashStr := fmt.Sprintf("%x", hash)

	// 提取文件扩展名
	ext := filepath.Ext(fileName)
	nameWithoutExt := fileName[:len(fileName)-len(ext)]

	// 生成按日期分组的路径 (YYYY/MM/DD)
	now := time.Now()
	path = filepath.Join(
		fmt.Sprintf("%04d", now.Year()),
		fmt.Sprintf("%02d", now.Month()),
		fmt.Sprintf("%02d", now.Day()),
	)

	suffix = ext
	prefix = nameWithoutExt
	uuidFileName = hashStr + ext

	return
}

func CalculateFileHash(file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer func() {
		if cerr := src.Close(); cerr != nil {
			logrus.WithError(cerr).Warn("storage: failed to close source file during hash calculation")
		}
	}()

	hash := sha256.New()
	if _, err := io.Copy(hash, src); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func ValidateFileSize(file *multipart.FileHeader, maxSize int64) error {
	if file.Size > maxSize {
		maxSizeMB := float64(maxSize) / (1024 * 1024)
		return fmt.Errorf("文件大小超过限制，最大为%.2fMB", maxSizeMB)
	}
	return nil
}

func ParseExpireInfo(expireValue int, expireStyle string) (expiredAt *int64, expiredCount int, usedCount int) {
	// 根据过期样式计算过期时间
	expiredCount = expireValue
	usedCount = 0

	// 如果是时间相关的过期样式，计算具体时间
	// TODO: 实现具体的时间计算逻辑

	return nil, expiredCount, usedCount
}
