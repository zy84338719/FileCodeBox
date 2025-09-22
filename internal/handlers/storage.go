package handlers

import (
	"github.com/zy84338719/filecodebox/internal/common"
	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/models/web"
	"github.com/zy84338719/filecodebox/internal/storage"
	"github.com/zy84338719/filecodebox/internal/utils"

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
	storageDetails := make(map[string]web.WebStorageDetail)

	for _, storageType := range availableStorages {
		detail := web.WebStorageDetail{
			Type:      storageType,
			Available: true,
		}

		// 测试连接状态
		if err := sh.storageManager.TestStorage(storageType); err != nil {
			detail.Available = false
			detail.Error = err.Error()
		}

		// 尝试附加路径与使用率信息
		switch storageType {
		case "local":
			// 本地存储使用配置中的 StoragePath，如果未配置则回退到数据目录
			path := sh.storageConfig.StoragePath
			if path == "" {
				path = sh.configManager.Base.DataPath
			}
			detail.StoragePath = path

			// 尝试读取磁盘使用率（若可用）
			if path != "" {
				if usagePercent, err := utils.GetUsagePercent(path); err == nil {
					val := int(usagePercent)
					detail.UsagePercent = &val
				}
			}
		case "s3":
			// S3 使用 bucket 名称作为标识
			if sh.storageConfig.S3 != nil {
				detail.StoragePath = sh.storageConfig.S3.BucketName
			}
		case "webdav":
			if sh.storageConfig.WebDAV != nil {
				detail.StoragePath = sh.storageConfig.WebDAV.Hostname
			}
		case "nfs":
			if sh.storageConfig.NFS != nil {
				detail.StoragePath = sh.storageConfig.NFS.MountPoint
			}
		}

		storageDetails[storageType] = detail
	}

	// 为前端创建适配的存储配置
	adaptedStorageConfig := sh.createAdaptedStorageConfig()

	response := web.StorageInfoResponse{
		Current:        currentStorage,
		Available:      availableStorages,
		StorageDetails: storageDetails,
		StorageConfig:  adaptedStorageConfig,
	}

	common.SuccessResponse(c, response)
}

// SwitchStorage 切换存储类型
func (sh *StorageHandler) SwitchStorage(c *gin.Context) {
	var req web.StorageSwitchRequest
	if !utils.BindJSONWithValidation(c, &req) {
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
	var req web.StorageTestRequest
	if !utils.BindJSONWithValidation(c, &req) {
		return
	}

	// 根据存储类型更新配置
	switch req.Type {
	case "local":
		if req.Config != nil && req.Config.StoragePath != "" {
			sh.storageConfig.StoragePath = req.Config.StoragePath
		}

	case "webdav":
		if req.Config != nil && req.Config.WebDAV != nil {
			if sh.storageConfig.WebDAV == nil {
				sh.storageConfig.WebDAV = &config.WebDAVConfig{}
			}
			if req.Config.WebDAV.Hostname != "" {
				sh.storageConfig.WebDAV.Hostname = req.Config.WebDAV.Hostname
			}
			if req.Config.WebDAV.Username != "" {
				sh.storageConfig.WebDAV.Username = req.Config.WebDAV.Username
			}
			if req.Config.WebDAV.Password != "" {
				sh.storageConfig.WebDAV.Password = req.Config.WebDAV.Password
			}
			if req.Config.WebDAV.RootPath != "" {
				sh.storageConfig.WebDAV.RootPath = req.Config.WebDAV.RootPath
			}
			if req.Config.WebDAV.URL != "" {
				sh.storageConfig.WebDAV.URL = req.Config.WebDAV.URL
			}

			// 重新创建 WebDAV 存储以应用新配置
			if err := sh.storageManager.ReconfigureWebDAV(
				sh.storageConfig.WebDAV.Hostname,
				sh.storageConfig.WebDAV.Username,
				sh.storageConfig.WebDAV.Password,
				sh.storageConfig.WebDAV.RootPath,
			); err != nil {
				common.InternalServerErrorResponse(c, "重新配置WebDAV存储失败: "+err.Error())
				return
			}
		}

	case "s3":
		if req.Config != nil && req.Config.S3 != nil {
			if sh.storageConfig.S3 == nil {
				sh.storageConfig.S3 = &config.S3Config{}
			}
			if req.Config.S3.AccessKeyID != "" {
				sh.storageConfig.S3.AccessKeyID = req.Config.S3.AccessKeyID
			}
			if req.Config.S3.SecretAccessKey != "" {
				sh.storageConfig.S3.SecretAccessKey = req.Config.S3.SecretAccessKey
			}
			if req.Config.S3.BucketName != "" {
				sh.storageConfig.S3.BucketName = req.Config.S3.BucketName
			}
			if req.Config.S3.EndpointURL != "" {
				sh.storageConfig.S3.EndpointURL = req.Config.S3.EndpointURL
			}
			if req.Config.S3.RegionName != "" {
				sh.storageConfig.S3.RegionName = req.Config.S3.RegionName
			}
			if req.Config.S3.Hostname != "" {
				sh.storageConfig.S3.Hostname = req.Config.S3.Hostname
			}
			// Proxy 字段直接赋值
			sh.storageConfig.S3.Proxy = req.Config.S3.Proxy
		}

	case "nfs":
		if req.Config != nil && req.Config.NFS != nil {
			if sh.storageConfig.NFS == nil {
				sh.storageConfig.NFS = &config.NFSConfig{}
			}
			if req.Config.NFS.Server != "" {
				sh.storageConfig.NFS.Server = req.Config.NFS.Server
			}
			if req.Config.NFS.Path != "" {
				sh.storageConfig.NFS.Path = req.Config.NFS.Path
			}
			if req.Config.NFS.MountPoint != "" {
				sh.storageConfig.NFS.MountPoint = req.Config.NFS.MountPoint
			}
			if req.Config.NFS.Version != "" {
				sh.storageConfig.NFS.Version = req.Config.NFS.Version
			}
			if req.Config.NFS.Options != "" {
				sh.storageConfig.NFS.Options = req.Config.NFS.Options
			}
			if req.Config.NFS.Timeout > 0 {
				sh.storageConfig.NFS.Timeout = req.Config.NFS.Timeout
			}
			if req.Config.NFS.SubPath != "" {
				sh.storageConfig.NFS.SubPath = req.Config.NFS.SubPath
			}
			// 直接赋值的字段
			sh.storageConfig.NFS.AutoMount = req.Config.NFS.AutoMount
			sh.storageConfig.NFS.RetryCount = req.Config.NFS.RetryCount

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
		}

	default:
		common.BadRequestResponse(c, "不支持的存储类型: "+req.Type)
		return
	}

	// 保存配置（会同时保存到文件和数据库）
	if err := sh.configManager.Save(); err != nil {
		common.InternalServerErrorResponse(c, "保存配置失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, "存储配置更新成功", nil)
}

// createAdaptedStorageConfig 创建适配前端的存储配置
func (sh *StorageHandler) createAdaptedStorageConfig() *config.StorageConfig {
	adapted := sh.storageConfig.Clone()

	// 确保Type字段正确设置
	if adapted.Type == "" {
		adapted.Type = "local"
	}

	// 设置存储路径的默认值
	if adapted.StoragePath == "" {
		adapted.StoragePath = "./data"
	}

	return adapted
}

// ...existing code...
