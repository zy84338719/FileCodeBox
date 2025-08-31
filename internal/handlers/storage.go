package handlers

import (
	"github.com/zy84338719/filecodebox/internal/common"
	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/storage"

	"github.com/gin-gonic/gin"
)

// StorageHandler 存储管理处理器
type StorageHandler struct {
	storageManager *storage.StorageManager
	config         *config.Config
}

// NewStorageHandler 创建存储处理器
func NewStorageHandler(sm *storage.StorageManager, cfg *config.Config) *StorageHandler {
	return &StorageHandler{
		storageManager: sm,
		config:         cfg,
	}
}

// GetStorageInfo 获取存储信息
func (sh *StorageHandler) GetStorageInfo(c *gin.Context) {
	availableStorages := sh.storageManager.GetAvailableStorages()
	currentStorage := sh.storageManager.GetCurrentStorage()

	// 获取各存储类型的详细信息
	storageDetails := make(map[string]interface{})

	for _, storageType := range availableStorages {
		details := map[string]interface{}{
			"type":      storageType,
			"available": true,
		}

		// 测试连接状态
		if err := sh.storageManager.TestStorage(storageType); err != nil {
			details["available"] = false
			details["error"] = err.Error()
		}

		storageDetails[storageType] = details
	}

	common.SuccessResponse(c, gin.H{
		"current":         currentStorage,
		"available":       availableStorages,
		"storage_details": storageDetails,
		"storage_config": gin.H{
			"local": gin.H{
				"storage_path": sh.config.StoragePath,
			},
			"webdav": gin.H{
				"hostname":  sh.config.WebDAVHostname,
				"username":  sh.config.WebDAVUsername,
				"root_path": sh.config.WebDAVRootPath,
				"url":       sh.config.WebDAVURL,
			},
			"nfs": gin.H{
				"server":      sh.config.NFSServer,
				"nfs_path":    sh.config.NFSPath,
				"mount_point": sh.config.NFSMountPoint,
				"version":     sh.config.NFSVersion,
				"options":     sh.config.NFSOptions,
				"timeout":     sh.config.NFSTimeout,
				"auto_mount":  sh.config.NFSAutoMount,
				"retry_count": sh.config.NFSRetryCount,
				"sub_path":    sh.config.NFSSubPath,
			},
		},
	})
}

// SwitchStorage 切换存储类型
func (sh *StorageHandler) SwitchStorage(c *gin.Context) {
	var req struct {
		StorageType string `json:"storage_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequestResponse(c, "参数错误: "+err.Error())
		return
	}

	// 切换存储
	if err := sh.storageManager.SwitchStorage(req.StorageType); err != nil {
		common.BadRequestResponse(c, err.Error())
		return
	}

	// 更新配置
	sh.config.FileStorage = req.StorageType
	if err := sh.config.Save(); err != nil {
		common.InternalServerErrorResponse(c, "保存配置失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, "存储切换成功", gin.H{
		"current": req.StorageType,
	})
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

	common.SuccessWithMessage(c, "连接测试成功", gin.H{
		"type":   storageType,
		"status": "connected",
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
			sh.config.StoragePath = storagePath
		}

	case "webdav":
		if hostname, ok := req.Config["hostname"].(string); ok {
			sh.config.WebDAVHostname = hostname
		}
		if username, ok := req.Config["username"].(string); ok {
			sh.config.WebDAVUsername = username
		}
		if password, ok := req.Config["password"].(string); ok && password != "" {
			sh.config.WebDAVPassword = password
		}
		if rootPath, ok := req.Config["root_path"].(string); ok {
			sh.config.WebDAVRootPath = rootPath
		}
		if url, ok := req.Config["url"].(string); ok {
			sh.config.WebDAVURL = url
		}

		// 重新创建 WebDAV 存储以应用新配置
		// 由于使用了策略模式，我们需要重新创建存储实例
		if err := sh.storageManager.ReconfigureWebDAV(
			sh.config.WebDAVHostname,
			sh.config.WebDAVUsername,
			sh.config.WebDAVPassword,
			sh.config.WebDAVRootPath,
		); err != nil {
			common.InternalServerErrorResponse(c, "重新配置WebDAV存储失败: "+err.Error())
			return
		}

	case "s3":
		if accessKeyID, ok := req.Config["access_key_id"].(string); ok {
			sh.config.S3AccessKeyID = accessKeyID
		}
		if secretAccessKey, ok := req.Config["secret_access_key"].(string); ok && secretAccessKey != "" {
			sh.config.S3SecretAccessKey = secretAccessKey
		}
		if bucketName, ok := req.Config["bucket_name"].(string); ok {
			sh.config.S3BucketName = bucketName
		}
		if endpointURL, ok := req.Config["endpoint_url"].(string); ok {
			sh.config.S3EndpointURL = endpointURL
		}
		if regionName, ok := req.Config["region_name"].(string); ok {
			sh.config.S3RegionName = regionName
		}
		if hostname, ok := req.Config["hostname"].(string); ok {
			sh.config.S3Hostname = hostname
		}
		if proxy, ok := req.Config["proxy"].(bool); ok {
			if proxy {
				sh.config.S3Proxy = 1
			} else {
				sh.config.S3Proxy = 0
			}
		}

	case "nfs":
		if server, ok := req.Config["server"].(string); ok {
			sh.config.NFSServer = server
		}
		if nfsPath, ok := req.Config["nfs_path"].(string); ok {
			sh.config.NFSPath = nfsPath
		}
		if mountPoint, ok := req.Config["mount_point"].(string); ok {
			sh.config.NFSMountPoint = mountPoint
		}
		if version, ok := req.Config["version"].(string); ok {
			sh.config.NFSVersion = version
		}
		if options, ok := req.Config["options"].(string); ok {
			sh.config.NFSOptions = options
		}
		if timeout, ok := req.Config["timeout"].(float64); ok {
			sh.config.NFSTimeout = int(timeout)
		}
		if autoMount, ok := req.Config["auto_mount"].(bool); ok {
			if autoMount {
				sh.config.NFSAutoMount = 1
			} else {
				sh.config.NFSAutoMount = 0
			}
		}
		if retryCount, ok := req.Config["retry_count"].(float64); ok {
			sh.config.NFSRetryCount = int(retryCount)
		}
		if subPath, ok := req.Config["sub_path"].(string); ok {
			sh.config.NFSSubPath = subPath
		}

		// 重新创建 NFS 存储以应用新配置
		if err := sh.storageManager.ReconfigureNFS(
			sh.config.NFSServer,
			sh.config.NFSPath,
			sh.config.NFSMountPoint,
			sh.config.NFSVersion,
			sh.config.NFSOptions,
			sh.config.NFSTimeout,
			sh.config.NFSAutoMount == 1,
			sh.config.NFSRetryCount,
			sh.config.NFSSubPath,
		); err != nil {
			common.InternalServerErrorResponse(c, "重新配置NFS存储失败: "+err.Error())
			return
		}

	default:
		common.BadRequestResponse(c, "不支持的存储类型: "+req.StorageType)
		return
	}

	// 保存配置（会同时保存到文件和数据库）
	if err := sh.config.Save(); err != nil {
		common.InternalServerErrorResponse(c, "保存配置失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, "存储配置更新成功", nil)
}
