package handlers

import (
	"github.com/zy84338719/filecodebox/internal/common"
	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/models/web"
	"github.com/zy84338719/filecodebox/internal/storage"

	"github.com/gin-gonic/gin"
)

// StorageHandler 存储管理处理器
type StorageHandler struct {
	storageManager *storage.StorageManager
	storageConfig  *config.StorageConfig
	configManager  *config.ConfigManager
}

// NewStorageHandler 创建存储处理器
func NewStorageHandler(sm *storage.StorageManager, storageConfig *config.StorageConfig, configManager *config.ConfigManager) *StorageHandler {
	return &StorageHandler{
		storageManager: sm,
		storageConfig:  storageConfig,
		configManager:  configManager,
	}
}

// GetStorageInfo 获取存储信息
func (sh *StorageHandler) GetStorageInfo(c *gin.Context) {
	availableStorages := sh.storageManager.GetAvailableStorages()
	currentStorage := sh.storageManager.GetCurrentStorage()

	// 获取各存储类型的详细信息
	storageDetails := make(map[string]web.StorageDetail)

	for _, storageType := range availableStorages {
		detail := web.StorageDetail{
			Type:      storageType,
			Available: true,
		}

		// 测试连接状态
		if err := sh.storageManager.TestStorage(storageType); err != nil {
			detail.Available = false
			detail.Error = err.Error()
		}

		storageDetails[storageType] = detail
	}

	// 构建存储配置信息
	storageConfig := map[string]map[string]interface{}{
		"local": {
			"storage_path": sh.storageConfig.StoragePath,
		},
		"webdav": {
			"hostname":  sh.storageConfig.WebDAV.Hostname,
			"username":  sh.storageConfig.WebDAV.Username,
			"root_path": sh.storageConfig.WebDAV.RootPath,
			"url":       sh.storageConfig.WebDAV.URL,
		},
		"nfs": {
			"server":      sh.storageConfig.NFS.Server,
			"nfs_path":    sh.storageConfig.NFS.Path,
			"mount_point": sh.storageConfig.NFS.MountPoint,
			"version":     sh.storageConfig.NFS.Version,
			"options":     sh.storageConfig.NFS.Options,
			"timeout":     sh.storageConfig.NFS.Timeout,
			"auto_mount":  sh.storageConfig.NFS.AutoMount,
			"retry_count": sh.storageConfig.NFS.RetryCount,
			"sub_path":    sh.storageConfig.NFS.SubPath,
		},
	}

	response := web.StorageInfoResponse{
		Current:        currentStorage,
		Available:      availableStorages,
		StorageDetails: storageDetails,
		StorageConfig:  storageConfig,
	}

	common.SuccessResponse(c, response)
}

// SwitchStorage 切换存储类型
func (sh *StorageHandler) SwitchStorage(c *gin.Context) {
	var req web.StorageSwitchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequestResponse(c, "参数错误: "+err.Error())
		return
	}

	// 切换存储
	if err := sh.storageManager.SwitchStorage(req.Type); err != nil {
		common.BadRequestResponse(c, err.Error())
		return
	}

	// 更新配置
	sh.storageConfig.Type = req.Type
	if err := sh.configManager.Save(); err != nil {
		common.InternalServerErrorResponse(c, "保存配置失败: "+err.Error())
		return
	}

	response := web.StorageSwitchResponse{
		Success:     true,
		Message:     "存储切换成功",
		CurrentType: req.Type,
	}
	common.SuccessResponse(c, response)
}

// TestStorageConnection 测试存储连接
func (sh *StorageHandler) TestStorageConnection(c *gin.Context) {
	storageType := c.Param("type")
	if storageType == "" {
		common.BadRequestResponse(c, "存储类型不能为空")
		return
	}

	err := sh.storageManager.TestStorage(storageType)
	if err != nil {
		common.BadRequestResponse(c, "连接测试失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, "连接测试成功", web.StorageConnectionResponse{
		Type:   storageType,
		Status: "connected",
	})
}

// UpdateStorageConfig 更新存储配置
func (sh *StorageHandler) UpdateStorageConfig(c *gin.Context) {
	var req struct {
		StorageType string                 `json:"storage_type" binding:"required"`
		Config      map[string]interface{} `json:"config" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequestResponse(c, "参数错误: "+err.Error())
		return
	}

	// 根据存储类型更新配置
	switch req.StorageType {
	case "local":
		if storagePath, ok := req.Config["storage_path"].(string); ok {
			sh.storageConfig.StoragePath = storagePath
		}

	case "webdav":
		if hostname, ok := req.Config["hostname"].(string); ok {
			sh.storageConfig.WebDAV.Hostname = hostname
		}
		if username, ok := req.Config["username"].(string); ok {
			sh.storageConfig.WebDAV.Username = username
		}
		if password, ok := req.Config["password"].(string); ok && password != "" {
			sh.storageConfig.WebDAV.Password = password
		}
		if rootPath, ok := req.Config["root_path"].(string); ok {
			sh.storageConfig.WebDAV.RootPath = rootPath
		}
		if url, ok := req.Config["url"].(string); ok {
			sh.storageConfig.WebDAV.URL = url
		}

		// 重新创建 WebDAV 存储以应用新配置
		// 由于使用了策略模式，我们需要重新创建存储实例
		if err := sh.storageManager.ReconfigureWebDAV(
			sh.storageConfig.WebDAV.Hostname,
			sh.storageConfig.WebDAV.Username,
			sh.storageConfig.WebDAV.Password,
			sh.storageConfig.WebDAV.RootPath,
		); err != nil {
			common.InternalServerErrorResponse(c, "重新配置WebDAV存储失败: "+err.Error())
			return
		}

	case "s3":
		if accessKeyID, ok := req.Config["access_key_id"].(string); ok {
			sh.storageConfig.S3.AccessKeyID = accessKeyID
		}
		if secretAccessKey, ok := req.Config["secret_access_key"].(string); ok && secretAccessKey != "" {
			sh.storageConfig.S3.SecretAccessKey = secretAccessKey
		}
		if bucketName, ok := req.Config["bucket_name"].(string); ok {
			sh.storageConfig.S3.BucketName = bucketName
		}
		if endpointURL, ok := req.Config["endpoint_url"].(string); ok {
			sh.storageConfig.S3.EndpointURL = endpointURL
		}
		if regionName, ok := req.Config["region_name"].(string); ok {
			sh.storageConfig.S3.RegionName = regionName
		}
		if hostname, ok := req.Config["hostname"].(string); ok {
			sh.storageConfig.S3.Hostname = hostname
		}
		if proxy, ok := req.Config["proxy"].(bool); ok {
			if proxy {
				sh.storageConfig.S3.Proxy = 1
			} else {
				sh.storageConfig.S3.Proxy = 0
			}
		}

	case "nfs":
		if server, ok := req.Config["server"].(string); ok {
			sh.storageConfig.NFS.Server = server
		}
		if nfsPath, ok := req.Config["nfs_path"].(string); ok {
			sh.storageConfig.NFS.Path = nfsPath
		}
		if mountPoint, ok := req.Config["mount_point"].(string); ok {
			sh.storageConfig.NFS.MountPoint = mountPoint
		}
		if version, ok := req.Config["version"].(string); ok {
			sh.storageConfig.NFS.Version = version
		}
		if options, ok := req.Config["options"].(string); ok {
			sh.storageConfig.NFS.Options = options
		}
		if timeout, ok := req.Config["timeout"].(float64); ok {
			sh.storageConfig.NFS.Timeout = int(timeout)
		}
		if autoMount, ok := req.Config["auto_mount"].(bool); ok {
			if autoMount {
				sh.storageConfig.NFS.AutoMount = 1
			} else {
				sh.storageConfig.NFS.AutoMount = 0
			}
		}
		if retryCount, ok := req.Config["retry_count"].(float64); ok {
			sh.storageConfig.NFS.RetryCount = int(retryCount)
		}
		if subPath, ok := req.Config["sub_path"].(string); ok {
			sh.storageConfig.NFS.SubPath = subPath
		}

		// 重新创建 NFS 存储以应用新配置
		if err := sh.storageManager.ReconfigureNFS(
			sh.storageConfig.NFS.Server,
			sh.storageConfig.NFS.Path,
			sh.storageConfig.NFS.MountPoint,
			sh.storageConfig.NFS.Version,
			sh.storageConfig.NFS.Options,
			sh.storageConfig.NFS.Timeout,
			sh.storageConfig.NFS.AutoMount == 1,
			sh.storageConfig.NFS.RetryCount,
			sh.storageConfig.NFS.SubPath,
		); err != nil {
			common.InternalServerErrorResponse(c, "重新配置NFS存储失败: "+err.Error())
			return
		}

	default:
		common.BadRequestResponse(c, "不支持的存储类型: "+req.StorageType)
		return
	}

	// 保存配置（会同时保存到文件和数据库）
	if err := sh.configManager.Save(); err != nil {
		common.InternalServerErrorResponse(c, "保存配置失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, "存储配置更新成功", nil)
}
