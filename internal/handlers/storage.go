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
	storageConfig := map[string]interface{}{
		"local":  sh.storageConfig,
		"webdav": sh.storageConfig.WebDAV,
		"nfs":    sh.storageConfig.NFS,
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

	common.SuccessResponse(c, web.StorageSwitchResponse{
		Success:     true,
		Message:     "存储切换成功",
		CurrentType: req.Type,
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
		if val, ok := req.Config["storage_path"]; ok {
			if storagePath, ok := val.(string); ok {
				sh.storageConfig.StoragePath = storagePath
			} else {
				common.BadRequestResponse(c, "storage_path 必须是字符串类型")
				return
			}
		}

	case "webdav":
		if val, ok := req.Config["hostname"]; ok {
			if hostname, ok := val.(string); ok {
				sh.storageConfig.WebDAV.Hostname = hostname
			} else {
				common.BadRequestResponse(c, "hostname 必须是字符串类型")
				return
			}
		}
		if val, ok := req.Config["username"]; ok {
			if username, ok := val.(string); ok {
				sh.storageConfig.WebDAV.Username = username
			} else {
				common.BadRequestResponse(c, "username 必须是字符串类型")
				return
			}
		}
		if val, ok := req.Config["password"]; ok {
			if password, ok := val.(string); ok && password != "" {
				sh.storageConfig.WebDAV.Password = password
			} else if !ok {
				common.BadRequestResponse(c, "password 必须是字符串类型")
				return
			}
		}
		if val, ok := req.Config["root_path"]; ok {
			if rootPath, ok := val.(string); ok {
				sh.storageConfig.WebDAV.RootPath = rootPath
			} else {
				common.BadRequestResponse(c, "root_path 必须是字符串类型")
				return
			}
		}
		if val, ok := req.Config["url"]; ok {
			if url, ok := val.(string); ok {
				sh.storageConfig.WebDAV.URL = url
			} else {
				common.BadRequestResponse(c, "url 必须是字符串类型")
				return
			}
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
		if val, ok := req.Config["access_key_id"]; ok {
			if accessKeyID, ok := val.(string); ok {
				sh.storageConfig.S3.AccessKeyID = accessKeyID
			} else {
				common.BadRequestResponse(c, "access_key_id 必须是字符串类型")
				return
			}
		}
		if val, ok := req.Config["secret_access_key"]; ok {
			if secretAccessKey, ok := val.(string); ok && secretAccessKey != "" {
				sh.storageConfig.S3.SecretAccessKey = secretAccessKey
			} else if !ok {
				common.BadRequestResponse(c, "secret_access_key 必须是字符串类型")
				return
			}
		}
		if val, ok := req.Config["bucket_name"]; ok {
			if bucketName, ok := val.(string); ok {
				sh.storageConfig.S3.BucketName = bucketName
			} else {
				common.BadRequestResponse(c, "bucket_name 必须是字符串类型")
				return
			}
		}
		if val, ok := req.Config["endpoint_url"]; ok {
			if endpointURL, ok := val.(string); ok {
				sh.storageConfig.S3.EndpointURL = endpointURL
			} else {
				common.BadRequestResponse(c, "endpoint_url 必须是字符串类型")
				return
			}
		}
		if val, ok := req.Config["region_name"]; ok {
			if regionName, ok := val.(string); ok {
				sh.storageConfig.S3.RegionName = regionName
			} else {
				common.BadRequestResponse(c, "region_name 必须是字符串类型")
				return
			}
		}
		if val, ok := req.Config["hostname"]; ok {
			if hostname, ok := val.(string); ok {
				sh.storageConfig.S3.Hostname = hostname
			} else {
				common.BadRequestResponse(c, "hostname 必须是字符串类型")
				return
			}
		}
		if val, ok := req.Config["proxy"]; ok {
			if proxy, ok := val.(bool); ok {
				if proxy {
					sh.storageConfig.S3.Proxy = 1
				} else {
					sh.storageConfig.S3.Proxy = 0
				}
			} else {
				common.BadRequestResponse(c, "proxy 必须是布尔类型")
				return
			}
		}

	case "nfs":
		if val, ok := req.Config["server"]; ok {
			if server, ok := val.(string); ok {
				sh.storageConfig.NFS.Server = server
			} else {
				common.BadRequestResponse(c, "server 必须是字符串类型")
				return
			}
		}
		if val, ok := req.Config["nfs_path"]; ok {
			if nfsPath, ok := val.(string); ok {
				sh.storageConfig.NFS.Path = nfsPath
			} else {
				common.BadRequestResponse(c, "nfs_path 必须是字符串类型")
				return
			}
		}
		if val, ok := req.Config["mount_point"]; ok {
			if mountPoint, ok := val.(string); ok {
				sh.storageConfig.NFS.MountPoint = mountPoint
			} else {
				common.BadRequestResponse(c, "mount_point 必须是字符串类型")
				return
			}
		}
		if val, ok := req.Config["version"]; ok {
			if version, ok := val.(string); ok {
				sh.storageConfig.NFS.Version = version
			} else {
				common.BadRequestResponse(c, "version 必须是字符串类型")
				return
			}
		}
		if val, ok := req.Config["options"]; ok {
			if options, ok := val.(string); ok {
				sh.storageConfig.NFS.Options = options
			} else {
				common.BadRequestResponse(c, "options 必须是字符串类型")
				return
			}
		}
		if val, ok := req.Config["timeout"]; ok {
			if timeout, ok := val.(float64); ok {
				sh.storageConfig.NFS.Timeout = int(timeout)
			} else {
				common.BadRequestResponse(c, "timeout 必须是数值类型")
				return
			}
		}
		if val, ok := req.Config["auto_mount"]; ok {
			if autoMount, ok := val.(bool); ok {
				if autoMount {
					sh.storageConfig.NFS.AutoMount = 1
				} else {
					sh.storageConfig.NFS.AutoMount = 0
				}
			} else {
				common.BadRequestResponse(c, "auto_mount 必须是布尔类型")
				return
			}
		}
		if val, ok := req.Config["retry_count"]; ok {
			if retryCount, ok := val.(float64); ok {
				sh.storageConfig.NFS.RetryCount = int(retryCount)
			} else {
				common.BadRequestResponse(c, "retry_count 必须是数值类型")
				return
			}
		}
		if val, ok := req.Config["sub_path"]; ok {
			if subPath, ok := val.(string); ok {
				sh.storageConfig.NFS.SubPath = subPath
			} else {
				common.BadRequestResponse(c, "sub_path 必须是字符串类型")
				return
			}
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
