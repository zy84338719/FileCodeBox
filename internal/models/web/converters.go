package web

import (
	"time"

	"github.com/zy84338719/filecodebox/internal/models"
	"github.com/zy84338719/filecodebox/internal/models/service"
)

// ConvertUserStatsToWeb 将服务层用户统计数据转换为Web响应
func ConvertUserStatsToWeb(data *service.UserStatsData) *UserStatsResponse {
	result := &UserStatsResponse{
		TotalUploads:    data.TotalUploads,
		TotalDownloads:  data.TotalDownloads,
		TotalStorage:    data.TotalStorage,
		MaxUploadSize:   data.MaxUploadSize,
		MaxStorageQuota: data.MaxStorageQuota,
		CurrentFiles:    data.CurrentFiles,
		FileCount:       data.FileCount,
		LastLoginIP:     data.LastLoginIP,
		EmailVerified:   data.EmailVerified,
		Status:          data.Status,
		Role:            data.Role,
	}

	if data.LastLoginAt != nil {
		result.LastLoginAt = data.LastLoginAt.Format(time.RFC3339)
	}
	if data.LastUploadAt != nil {
		result.LastUploadAt = data.LastUploadAt.Format(time.RFC3339)
	}
	if data.LastDownloadAt != nil {
		result.LastDownloadAt = data.LastDownloadAt.Format(time.RFC3339)
	}

	return result
}

// ConvertShareStatsToWeb 将服务层分享统计数据转换为Web响应
func ConvertShareStatsToWeb(data *service.ShareStatsData) *ShareStatsResponse {
	return &ShareStatsResponse{
		Code:         data.Code,
		FileName:     data.FileName,
		FileSize:     data.FileSize,
		UsedCount:    data.UsedCount,
		ExpiredCount: data.ExpiredCount,
		ExpiredAt:    data.ExpiredAt,
		UploadType:   data.UploadType,
		RequireAuth:  data.RequireAuth,
		IsExpired:    data.IsExpired,
		CreatedAt:    data.CreatedAt,
	}
}

// ConvertFileCodeToFileInfo 将 models.FileCode 转换为 web.FileInfo
func ConvertFileCodeToFileInfo(fileCode *models.FileCode) *FileInfo {
	if fileCode == nil {
		return nil
	}

	// 判断是否过期
	isExpired := false
	if fileCode.ExpiredAt != nil {
		isExpired = time.Now().After(*fileCode.ExpiredAt)
	}

	// 获取文件名
	fileName := fileCode.UUIDFileName
	if fileName == "" && fileCode.FilePath != "" {
		// 如果没有 UUIDFileName，从 FilePath 提取
		fileName = fileCode.FilePath
	}

	return &FileInfo{
		Code:         fileCode.Code,
		FileName:     fileName,
		Size:         fileCode.Size,
		UploadType:   fileCode.UploadType,
		RequireAuth:  fileCode.UserID != nil, // 如果有用户ID，表示需要认证
		UsedCount:    fileCode.UsedCount,
		ExpiredCount: fileCode.ExpiredCount,
		ExpiredAt:    fileCode.ExpiredAt,
		CreatedAt:    fileCode.CreatedAt,
		IsExpired:    isExpired,
	}
}

// ConvertFileCodeSliceToFileInfoSlice 将 []models.FileCode 转换为 []FileInfo
func ConvertFileCodeSliceToFileInfoSlice(fileCodes []models.FileCode) []FileInfo {
	result := make([]FileInfo, len(fileCodes))
	for i, fileCode := range fileCodes {
		if converted := ConvertFileCodeToFileInfo(&fileCode); converted != nil {
			result[i] = *converted
		}
	}
	return result
}
