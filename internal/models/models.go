// Package models 定义应用程序的数据模型
package models

import (
	"github.com/zy84338719/filecodebox/internal/models/db"
	"github.com/zy84338719/filecodebox/internal/models/dto"
	"github.com/zy84338719/filecodebox/internal/models/mcp"
	"github.com/zy84338719/filecodebox/internal/models/service"
)

// 类型别名，用于向后兼容
type (
	// 数据库模型别名
	FileCode    = db.FileCode
	UploadChunk = db.UploadChunk
	KeyValue    = db.KeyValue
	User        = db.User
	UserSession = db.UserSession

	// 服务模型别名
	BuildInfo = service.BuildInfo

	// 请求结构别名
	ShareFileRequest = service.ShareFileRequest
	ShareTextRequest = service.ShareTextRequest

	// 响应结构别名
	ShareStatsData  = service.ShareStatsData
	ShareUpdateData = service.ShareUpdateData
	ShareTextResult = service.ShareTextResult
	ShareFileResult = service.ShareFileResult

	// 管理员响应结构别名
	DatabaseStats         = service.DatabaseStats
	StorageStatus         = service.StorageStatus
	DiskUsage             = service.DiskUsage
	PerformanceMetrics    = service.PerformanceMetrics
	SystemInfo            = service.SystemInfo
	SecurityScanResult    = service.SecurityScanResult
	PermissionCheckResult = service.PermissionCheckResult
	IntegrityCheckResult  = service.IntegrityCheckResult
	LogStats              = service.LogStats
	RunningTask           = service.RunningTask
	MCPConfig             = service.MCPConfig
	MCPStatus             = service.MCPStatus
	MCPTestResult         = service.MCPTestResult
	StorageTestResult     = service.StorageTestResult
	UserStatsResponse     = service.UserStatsResponse

	// 分块上传响应结构别名
	// Chunk service response structures
	ChunkUploadProgressResponse = service.ChunkUploadProgressResponse
	ChunkUploadStatusResponse   = service.ChunkUploadStatusResponse
	ChunkUploadStatusData       = service.ChunkUploadStatusData
	ChunkVerifyResponse         = service.ChunkVerifyResponse
	ChunkUploadCompleteResponse = service.ChunkUploadCompleteResponse

	// DTO 模型别名
	UserUpdateFields        = dto.UserUpdateFields
	UserProfileUpdateFields = dto.UserProfileUpdateFields
	UserStatsUpdateFields   = dto.UserStatsUpdateFields
	ConfigUpdateFields      = dto.ConfigUpdateFields
	FlatConfigUpdate        = dto.FlatConfigUpdate
	MCPConfigUpdate         = dto.MCPConfigUpdate
	BaseConfigUpdate        = dto.BaseConfigUpdate
	TransferConfigUpdate    = dto.TransferConfigUpdate
	UploadConfigUpdate      = dto.UploadConfigUpdate
	DownloadConfigUpdate    = dto.DownloadConfigUpdate
	UserConfigUpdate        = dto.UserConfigUpdate
	StorageConfigUpdate     = dto.StorageConfigUpdate
	S3ConfigUpdate          = dto.S3ConfigUpdate
	WebDAVConfigUpdate      = dto.WebDAVConfigUpdate
	OneDriveConfigUpdate    = dto.OneDriveConfigUpdate
	NFSConfigUpdate         = dto.NFSConfigUpdate

	// MCP 模型别名
	SystemConfigResponse = mcp.SystemConfigResponse
)

// 全局变量别名
var (
	GoVersion = service.GoVersion
	BuildTime = service.BuildTime
	GitCommit = service.GitCommit
	GitBranch = service.GitBranch
	Version   = service.Version
)

// 函数别名
var GetBuildInfo = service.GetBuildInfo
