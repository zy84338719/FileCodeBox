package storage

import (
	"context"
	"fmt"
	"os"

	"github.com/zy84338719/fileCodeBox/backend/internal/conf"
)

// Service 存储服务
type Service struct {
	config *conf.AppConfiguration
}

// NewService 创建存储服务
func NewService() *Service {
	return &Service{
		config: conf.GetGlobalConfig(),
	}
}

// GetStorageInfo 获取存储信息
func (s *Service) GetStorageInfo(ctx context.Context) (*StorageInfo, error) {
	// 获取可用存储类型
	availableStorages := []string{"local"}
	if s.config.Storage.Type != "" {
		availableStorages = append(availableStorages, "local")
	}

	// 获取各存储类型的详细信息
	storageDetails := make(map[string]*StorageDetail)

	// 本地存储详情
	storageDetails["local"] = &StorageDetail{
		Type:         "local",
		Available:    true,
		StoragePath:  s.getStoragePath(),
		UsagePercent: s.getDiskUsage(),
	}

	// 当前存储类型
	currentType := s.config.Storage.Type
	if currentType == "" {
		currentType = "local"
	}

	// 存储配置
	storageConfig := s.getStorageConfig()

	return &StorageInfo{
		Current:        currentType,
		Available:      availableStorages,
		StorageDetails: storageDetails,
		StorageConfig:  storageConfig,
	}, nil
}

// SwitchStorage 切换存储类型
func (s *Service) SwitchStorage(ctx context.Context, storageType string) error {
	if storageType != "local" && storageType != "s3" && storageType != "webdav" && storageType != "nfs" {
		return fmt.Errorf("不支持的存储类型: %s", storageType)
	}
	// TODO: 实现实际的存储切换逻辑
	// 这里需要重新初始化存储管理器
	return nil
}

// TestStorageConnection 测试存储连接
func (s *Service) TestStorageConnection(ctx context.Context, storageType string) error {
	if storageType == "" {
		return fmt.Errorf("存储类型不能为空")
	}

	switch storageType {
	case "local":
		path := s.getStoragePath()
		if path == "" {
			return fmt.Errorf("存储路径未配置")
		}
		// 检查目录是否存在并可写
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return fmt.Errorf("存储路径不存在: %s", path)
		}
		// 测试可写
		testFile := path + "/.test_write"
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			return fmt.Errorf("存储路径不可写: %s", err)
		}
		os.Remove(testFile)
		return nil

	case "s3", "webdav", "nfs":
		// TODO: 实现各存储类型的连接测试
		return fmt.Errorf("存储类型 %s 连接测试暂未实现", storageType)

	default:
		return fmt.Errorf("不支持的存储类型: %s", storageType)
	}
}

// UpdateStorageConfig 更新存储配置
func (s *Service) UpdateStorageConfig(ctx context.Context, req *UpdateConfigRequest) error {
	// TODO: 实现配置更新和持久化
	// 这里需要更新配置文件并重新加载

	switch req.Type {
	case "local":
		if req.Config.StoragePath != "" {
			s.config.Storage.StoragePath = req.Config.StoragePath
		}
	case "webdav", "s3", "nfs":
		// TODO: 实现各存储类型的配置更新
		return fmt.Errorf("存储类型 %s 配置更新暂未实现", req.Type)

	default:
		return fmt.Errorf("不支持的存储类型: %s", req.Type)
	}

	return nil
}

// getStoragePath 获取存储路径
func (s *Service) getStoragePath() string {
	if s.config.Storage.StoragePath != "" {
		return s.config.Storage.StoragePath
	}
	if s.config.App.DataPath != "" {
		return s.config.App.DataPath
	}
	return "./data"
}

// getDiskUsage 获取磁盘使用率
func (s *Service) getDiskUsage() int32 {
	// TODO: 实现磁盘使用率计算
	// 可以使用 syscall.Statfs 获取磁盘信息
	return 0
}

// getStorageConfig 获取存储配置
func (s *Service) getStorageConfig() *StorageConfig {
	return &StorageConfig{
		Type:        s.config.Storage.Type,
		StoragePath: s.getStoragePath(),
		WebDAV:      s.getWebDAVConfig(),
		S3:          s.getS3Config(),
		NFS:         s.getNFSConfig(),
	}
}

// getWebDAVConfig 获取 WebDAV 配置
func (s *Service) getWebDAVConfig() *WebDAVConfig {
	// TODO: 从配置中读取 WebDAV 配置
	return &WebDAVConfig{
		Hostname: "",
		Username: "",
		Password: "",
		RootPath: "",
		URL:      "",
	}
}

// getS3Config 获取 S3 配置
func (s *Service) getS3Config() *S3Config {
	// TODO: 从配置中读取 S3 配置
	return &S3Config{
		AccessKeyID:     "",
		SecretAccessKey: "",
		BucketName:      "",
		EndpointURL:     "",
		RegionName:      "",
		Hostname:        "",
		Proxy:           "",
	}
}

// getNFSConfig 获取 NFS 配置
func (s *Service) getNFSConfig() *NFSConfig {
	// TODO: 从配置中读取 NFS 配置
	return &NFSConfig{
		Server:     "",
		Path:       "",
		MountPoint: "",
		Version:    "",
		Options:    "",
		Timeout:    0,
		AutoMount:  0,
		RetryCount: 0,
		SubPath:    "",
	}
}

// ==================== 响应模型 ====================

type StorageInfo struct {
	Current        string                    `json:"current"`
	Available      []string                  `json:"available"`
	StorageDetails map[string]*StorageDetail `json:"storage_details"`
	StorageConfig  *StorageConfig            `json:"storage_config"`
}

type StorageDetail struct {
	Type         string `json:"type"`
	Available    bool   `json:"available"`
	StoragePath  string `json:"storage_path"`
	UsagePercent int32  `json:"usage_percent"`
	Error        string `json:"error,omitempty"`
}

type StorageConfig struct {
	Type        string        `json:"type"`
	StoragePath string        `json:"storage_path"`
	WebDAV      *WebDAVConfig `json:"webdav,omitempty"`
	S3          *S3Config     `json:"s3,omitempty"`
	NFS         *NFSConfig    `json:"nfs,omitempty"`
}

type WebDAVConfig struct {
	Hostname string `json:"hostname"`
	Username string `json:"username"`
	Password string `json:"password"`
	RootPath string `json:"root_path"`
	URL      string `json:"url"`
}

type S3Config struct {
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	BucketName      string `json:"bucket_name"`
	EndpointURL     string `json:"endpoint_url"`
	RegionName      string `json:"region_name"`
	Hostname        string `json:"hostname"`
	Proxy           string `json:"proxy"`
}

type NFSConfig struct {
	Server     string `json:"server"`
	Path       string `json:"path"`
	MountPoint string `json:"mount_point"`
	Version    string `json:"version"`
	Options    string `json:"options"`
	Timeout    int32  `json:"timeout"`
	AutoMount  int32  `json:"auto_mount"`
	RetryCount int32  `json:"retry_count"`
	SubPath    string `json:"sub_path"`
}

// UpdateConfigRequest 更新配置请求
type UpdateConfigRequest struct {
	Type   string `json:"type"`
	Config struct {
		StoragePath string        `json:"storage_path"`
		WebDAV      *WebDAVConfig `json:"webdav"`
		S3          *S3Config     `json:"s3"`
		NFS         *NFSConfig    `json:"nfs"`
	} `json:"config"`
}
