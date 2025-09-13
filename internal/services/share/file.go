package share

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zy84338719/filecodebox/internal/models"
	"github.com/zy84338719/filecodebox/internal/models/service"
)

// GetFileByCode 通过代码获取文件
func (s *Service) GetFileByCode(code string) (*models.FileCode, error) {
	fileCode, err := s.repositoryManager.FileCode.GetByCode(code)
	if err != nil {
		return nil, err
	}

	// 检查文件是否过期
	if fileCode.IsExpired() {
		return nil, errors.New("file has expired")
	}

	return fileCode, nil
}

// DownloadFile 下载文件
func (s *Service) DownloadFile(c *gin.Context, code string) error {
	fileCode, err := s.GetFileByCode(code)
	if err != nil {
		return err
	}

	// 增加使用次数
	if fileCode.ExpiredCount > 0 {
		fileCode.ExpiredCount--
		err = s.repositoryManager.FileCode.Update(fileCode)
		if err != nil {
			return err
		}
	}

	// 增加使用统计
	fileCode.UsedCount++
	err = s.repositoryManager.FileCode.Update(fileCode)
	if err != nil {
		return err
	}

	// 使用存储服务提供文件下载
	return s.storageService.GetFileResponse(c, fileCode)
}

// GetFileInfo 获取文件信息（不下载）
func (s *Service) GetFileInfo(code string) (*models.FileCode, error) {
	return s.GetFileByCode(code)
}

// CheckFileExists 检查文件是否存在
func (s *Service) CheckFileExists(code string) bool {
	_, err := s.GetFileByCode(code)
	return err == nil
}

// CreateFileShare 创建文件分享
func (s *Service) CreateFileShare(
	code, prefix, suffix, uuidFileName, filePath, text string,
	fileSize int64,
	expiredAt *time.Time,
	expiredCount int,
	userID *uint,
	requireAuth bool,
	ownerIP string,
) (*models.FileCode, error) {
	fileCode := &models.FileCode{
		Code:         code,
		Prefix:       prefix,
		Suffix:       suffix,
		UUIDFileName: uuidFileName,
		FilePath:     filePath,
		Size:         fileSize,
		Text:         text,
		ExpiredAt:    expiredAt,
		ExpiredCount: expiredCount,
		UserID:       userID,
		RequireAuth:  requireAuth,
		OwnerIP:      ownerIP,
	}

	// 设置上传类型
	if userID != nil {
		fileCode.UploadType = "authenticated"
	} else {
		fileCode.UploadType = "anonymous"
	}

	err := s.repositoryManager.FileCode.Create(fileCode)
	if err != nil {
		return nil, err
	}

	// 如果是已认证用户上传，更新用户统计信息
	if userID != nil {
		if s.userService != nil {
			// 更新用户上传统计：增加上传次数和存储使用量
			err = s.userService.UpdateUserStats(*userID, "upload", fileSize)
			if err != nil {
				// 记录警告但不中断流程
				fmt.Printf("Warning: Failed to update user stats: %v\n", err)
			}
		}
	}

	return fileCode, nil
}

// UpdateFileShare 更新文件分享信息
func (s *Service) UpdateFileShare(code string, updates map[string]interface{}) error {
	fileCode, err := s.GetFileByCode(code)
	if err != nil {
		return err
	}

	// 更新字段
	for key, value := range updates {
		switch key {
		case "expired_at":
			if expiredAt, ok := value.(*time.Time); ok {
				fileCode.ExpiredAt = expiredAt
			}
		case "expired_count":
			if expiredCount, ok := value.(int); ok {
				fileCode.ExpiredCount = expiredCount
			}
		case "require_auth":
			if requireAuth, ok := value.(bool); ok {
				fileCode.RequireAuth = requireAuth
			}
		}
	}

	return s.repositoryManager.FileCode.Update(fileCode)
}

// DeleteFileShare 删除文件分享
func (s *Service) DeleteFileShare(code string) error {
	fileCode, err := s.GetFileByCode(code)
	if err != nil {
		return err
	}

	// 删除实际文件
	result := s.storageService.DeleteFileWithResult(fileCode)
	if !result.Success {
		fmt.Printf("Warning: Failed to delete physical file: %v\n", result.Error)
	}

	// 如果是已认证用户的文件，更新用户统计信息
	if fileCode.UserID != nil && s.userService != nil {
		// 减少用户存储使用量（注意：这里使用负值来减少）
		err = s.userService.UpdateUserStats(*fileCode.UserID, "delete", -fileCode.Size)
		if err != nil {
			// 记录警告但不中断流程
			fmt.Printf("Warning: Failed to update user stats for deletion: %v\n", err)
		}
	}

	// 删除数据库记录
	return s.repositoryManager.FileCode.DeleteByFileCode(fileCode)
}

// ListUserFiles 列出用户的文件
func (s *Service) ListUserFiles(userID uint, page, pageSize int) ([]models.FileCode, int64, error) {
	// 获取用户的所有文件（暂时不支持分页）
	files, err := s.repositoryManager.FileCode.GetFilesByUserID(userID)
	if err != nil {
		return nil, 0, err
	}

	// 简单的分页逻辑
	total := int64(len(files))
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= len(files) {
		return []models.FileCode{}, total, nil
	}
	if end > len(files) {
		end = len(files)
	}

	return files[start:end], total, nil
}

// GetShareStats 获取分享统计信息
func (s *Service) GetShareStats(code string) (*service.ShareStatsData, error) {
	fileCode, err := s.GetFileByCode(code)
	if err != nil {
		return nil, err
	}

	return &service.ShareStatsData{
		Code:         fileCode.Code,
		UsedCount:    fileCode.UsedCount,
		ExpiredCount: fileCode.ExpiredCount,
		ExpiredAt:    fileCode.ExpiredAt,
		IsExpired:    fileCode.IsExpired(),
		CreatedAt:    fileCode.CreatedAt,
		FileSize:     fileCode.Size,
		FileName:     fileCode.Prefix + fileCode.Suffix,
		UploadType:   fileCode.UploadType,
		RequireAuth:  fileCode.RequireAuth,
	}, nil
}

// ShareText 分享文本内容
func (s *Service) ShareText(text string, expireValue int, expireStyle string) (*models.ShareTextResult, error) {
	// 生成唯一的代码
	code := generateRandomCode()

	// 计算过期时间
	var expiredAt *time.Time
	if expireValue > 0 {
		var duration time.Duration
		switch expireStyle {
		case "minute":
			duration = time.Duration(expireValue) * time.Minute
		case "hour":
			duration = time.Duration(expireValue) * time.Hour
		case "day":
			duration = time.Duration(expireValue) * 24 * time.Hour
		default:
			duration = 24 * time.Hour // 默认1天
		}
		expTime := time.Now().Add(duration)
		expiredAt = &expTime
	}

	// 创建文件记录
	fileCode := &models.FileCode{
		Code:         code,
		Text:         text,
		ExpiredAt:    expiredAt,
		UsedCount:    0,
		ExpiredCount: 1, // 文本分享默认只能使用一次
	}

	err := s.repositoryManager.FileCode.Create(fileCode)
	if err != nil {
		return nil, fmt.Errorf("创建文本分享失败: %v", err)
	}

	return &models.ShareTextResult{
		Code:      code,
		ShareURL:  fmt.Sprintf("/s/%s", code),
		ExpiredAt: expiredAt,
	}, nil
}

// generateRandomCode 生成随机代码
func generateRandomCode() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 8

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// ShareTextWithAuth 带认证的文本分享 (兼容性方法)
func (s *Service) ShareTextWithAuth(text string, expireValue int, expireStyle string, userID *uint) (*models.ShareTextResult, error) {
	// 生成唯一的代码
	code := generateRandomCode()

	// 计算过期时间
	var expiredAt *time.Time
	if expireValue > 0 {
		var duration time.Duration
		switch expireStyle {
		case "minute":
			duration = time.Duration(expireValue) * time.Minute
		case "hour":
			duration = time.Duration(expireValue) * time.Hour
		case "day":
			duration = time.Duration(expireValue) * 24 * time.Hour
		default:
			duration = 24 * time.Hour // 默认1天
		}
		expTime := time.Now().Add(duration)
		expiredAt = &expTime
	}

	// 创建文件记录
	fileCode := &models.FileCode{
		Code:         code,
		Text:         text,
		ExpiredAt:    expiredAt,
		UsedCount:    0,
		ExpiredCount: 1,      // 文本分享默认只能使用一次
		UserID:       userID, // 关联用户
	}

	err := s.repositoryManager.FileCode.Create(fileCode)
	if err != nil {
		return nil, fmt.Errorf("创建文本分享失败: %v", err)
	}

	return &models.ShareTextResult{
		Code:      code,
		ShareURL:  fmt.Sprintf("/s/%s", code),
		ExpiredAt: expiredAt,
	}, nil
}

// ShareFileWithAuth 带认证的文件分享
func (s *Service) ShareFileWithAuth(req models.ShareFileRequest) (*models.ShareFileResult, error) {
	// 获取文件
	fileHeader := req.File
	if fileHeader == nil {
		return nil, fmt.Errorf("no file provided")
	}

	// 打开文件
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("关闭文件失败: %v\n", err)
		}
	}()

	// 生成文件代码
	code := generateRandomCode()

	// 生成文件路径
	uuidFileName := fmt.Sprintf("%s-%s", code, fileHeader.Filename)
	directoryPath := "uploads"
	fullFilePath := fmt.Sprintf("%s/%s", directoryPath, uuidFileName)

	// 保存文件到存储系统
	result := s.storageService.SaveFileWithResult(fileHeader, fullFilePath)
	if !result.Success {
		return nil, fmt.Errorf("failed to save file: %v", result.Error)
	}

	// 计算过期时间和次数限制
	var expiredAt *time.Time
	var expiredCount int

	switch req.ExpireStyle {
	case "forever":
		// 永久有效，不设置过期时间和次数限制
		expiredAt = nil
		expiredCount = -1 // -1 表示无限制次数
	case "count":
		// 按次数限制
		expiredAt = nil                // 不设置时间过期
		expiredCount = req.ExpireValue // 设置剩余次数
	case "minute":
		// 按分钟限制
		if req.ExpireValue > 0 {
			t := time.Now().Add(time.Duration(req.ExpireValue) * time.Minute)
			expiredAt = &t
		}
		expiredCount = -1 // 时间限制时，不限制次数
	case "hour":
		// 按小时限制
		if req.ExpireValue > 0 {
			t := time.Now().Add(time.Duration(req.ExpireValue) * time.Hour)
			expiredAt = &t
		}
		expiredCount = -1 // 时间限制时，不限制次数
	case "day":
		// 按天限制
		if req.ExpireValue > 0 {
			t := time.Now().Add(time.Duration(req.ExpireValue) * 24 * time.Hour)
			expiredAt = &t
		}
		expiredCount = -1 // 时间限制时，不限制次数
	default:
		// 默认永久有效
		expiredAt = nil
		expiredCount = 0
	}

	// 创建文件分享记录
	fileCode, err := s.CreateFileShare(
		code,
		"",            // prefix
		"",            // suffix
		uuidFileName,  // UUID化的文件名
		directoryPath, // 目录路径
		"",            // text - 文件分享不需要文本内容
		fileHeader.Size,
		expiredAt,
		expiredCount, // 使用计算出的过期次数
		req.UserID,
		req.RequireAuth,
		req.ClientIP,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create file share: %w", err)
	}

	return &models.ShareFileResult{
		Code:      fileCode.Code,
		ShareURL:  fmt.Sprintf("/s/%s", fileCode.Code),
		FileName:  fileCode.UUIDFileName,
		ExpiredAt: fileCode.ExpiredAt,
	}, nil
}

// GetFileByCodeWithAuth 带认证获取文件 (兼容性方法)
func (s *Service) GetFileByCodeWithAuth(code string, userID *uint) (*models.FileCode, error) {
	fileCode, err := s.GetFileByCode(code)
	if err != nil {
		return nil, err
	}

	// 检查文件是否需要认证下载
	if fileCode.RequireAuth && userID == nil {
		return nil, fmt.Errorf("该文件需要登录后才能下载")
	}

	// 注意：这里不再检查 UserID 匹配，因为文件分享应该允许其他用户下载
	// UserID 字段只是用来标识文件的上传者，而不是限制下载权限
	// RequireAuth 字段才是控制下载权限的标志

	return fileCode, nil
}

// UpdateFileUsage 更新文件使用情况 (兼容性方法)
func (s *Service) UpdateFileUsage(code string) error {
	fileCode, err := s.GetFileByCode(code)
	if err != nil {
		return err
	}

	// 增加使用次数
	fileCode.UsedCount++

	// 减少可用次数
	if fileCode.ExpiredCount > 0 {
		fileCode.ExpiredCount--
	}

	return s.repositoryManager.FileCode.Update(fileCode)
}

// GetStorageService 获取存储服务 (兼容性方法)
func (s *Service) GetStorageService() interface{} {
	return s.storageService
}
