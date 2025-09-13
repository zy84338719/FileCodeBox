package admin

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zy84338719/filecodebox/internal/models"
	"github.com/zy84338719/filecodebox/internal/models/dto"
)

// GenerateToken 生成管理员JWT令牌
func (s *Service) GenerateToken() (string, error) {
	// 创建JWT claims
	claims := jwt.MapClaims{
		"is_admin": true,
		"exp":      time.Now().Add(24 * time.Hour).Unix(), // 24小时过期
	}

	// 创建token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名token
	tokenString, err := token.SignedString([]byte(s.manager.AdminToken))
	if err != nil {
		return "", fmt.Errorf("生成token失败: %w", err)
	}

	return tokenString, nil
}

// ValidateToken 验证管理员JWT令牌
func (s *Service) ValidateToken(tokenString string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 确保签名方法是HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.manager.AdminToken), nil
	})

	if err != nil {
		return err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// 检查是否是管理员token
		if isAdmin, exists := claims["is_admin"]; !exists || !isAdmin.(bool) {
			return errors.New("不是管理员token")
		}
		return nil
	}

	return errors.New("无效的token")
}

// ResetUserPassword 重置用户密码 - 使用统一的认证服务
func (s *Service) ResetUserPassword(userID uint, newPassword string) error {
	hashedPassword, err := s.authService.HashPassword(newPassword)
	if err != nil {
		return err
	}

	updateFields := &dto.UserUpdateFields{
		PasswordHash: &hashedPassword,
	}
	return s.repositoryManager.User.UpdateUserFields(userID, updateFields)
}

// GetUserStats 获取用户统计信息
func (s *Service) GetUserStats(userID uint) (*models.UserStatsResponse, error) {
	user, err := s.repositoryManager.User.GetByID(userID)
	if err != nil {
		return nil, err
	}

	// 获取文件数量
	fileCount, err := s.repositoryManager.FileCode.CountByUserID(userID)
	if err != nil {
		return nil, err
	}

	// 获取今日上传数量
	files, err := s.repositoryManager.FileCode.GetFilesByUserID(userID)
	if err != nil {
		return nil, err
	}

	// 计算今日上传数量
	today := time.Now().Truncate(24 * time.Hour)
	var todayUploads int64
	for _, file := range files {
		if file.CreatedAt.After(today) {
			todayUploads++
		}
	}

	return &models.UserStatsResponse{
		UserID:            userID,
		TotalUploads:      int64(user.TotalUploads),
		TotalDownloads:    int64(user.TotalDownloads),
		TotalStorage:      user.TotalStorage,
		TotalFiles:        fileCount,
		TodayUploads:      todayUploads,
		MaxUploadSize:     user.MaxUploadSize,
		MaxStorageQuota:   user.MaxStorageQuota,
		StorageUsage:      user.TotalStorage,
		StoragePercentage: float64(user.TotalStorage) / float64(user.MaxStorageQuota) * 100,
	}, nil
}

// UpdateUserStatus 更新用户状态
func (s *Service) UpdateUserStatus(userID uint, isActive bool) error {
	status := "inactive"
	if isActive {
		status = "active"
	}

	updateFields := &dto.UserUpdateFields{
		Status: &status,
	}
	return s.repositoryManager.User.UpdateUserFields(userID, updateFields)
}

// GetUserFiles 获取用户文件列表
func (s *Service) GetUserFiles(userID uint, page, limit int) ([]models.FileCode, int64, error) {
	offset := (page - 1) * limit

	// 计算总数
	total, err := s.repositoryManager.FileCode.CountByUserID(userID)
	if err != nil {
		return nil, 0, err
	}

	// 获取文件列表
	files, err := s.repositoryManager.FileCode.GetFilesByUserID(userID)
	if err != nil {
		return nil, 0, err
	}

	// 手动分页
	start := offset
	end := start + limit
	if start > len(files) {
		return []models.FileCode{}, total, nil
	}
	if end > len(files) {
		end = len(files)
	}

	return files[start:end], total, nil
}
