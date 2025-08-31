package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/dao"
	"github.com/zy84338719/filecodebox/internal/models"

	"gorm.io/gorm"
)

// UserService 用户服务
type UserService struct {
	cfg         *config.Config
	daoManager  *dao.DAOManager
	authService *AuthService
}

// NewUserService 创建用户服务
func NewUserService(daoManager *dao.DAOManager, cfg *config.Config) *UserService {
	return &UserService{
		cfg:         cfg,
		daoManager:  daoManager,
		authService: NewAuthService(daoManager, cfg),
	}
}

// Register 用户注册
func (s *UserService) Register(username, email, password, nickname string) (*models.User, error) {
	data := UserRegistrationData{
		Username: username,
		Email:    email,
		Password: password,
		Nickname: nickname,
	}
	return s.authService.RegisterUser(data)
}

// Login 用户登录
func (s *UserService) Login(usernameOrEmail, password, ipAddress, userAgent string) (string, *models.User, error) {
	data := UserLoginData{
		Username:  usernameOrEmail,
		Password:  password,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}
	return s.authService.LoginUser(data)
}

// ValidateToken 验证token
func (s *UserService) ValidateToken(tokenString string) (interface{}, error) {
	claims, err := s.authService.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}
	return claims, nil
}

// Logout 用户登出
func (s *UserService) Logout(sessionID string) error {
	return s.authService.LogoutUser(sessionID)
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	data := PasswordChangeData{
		UserID:      userID,
		OldPassword: oldPassword,
		NewPassword: newPassword,
	}
	return s.authService.ChangeUserPassword(data)
}

// ValidateUserInput 验证用户输入
func (s *UserService) ValidateUserInput(username, email, password string) error {
	data := UserRegistrationData{
		Username: username,
		Email:    email,
		Password: password,
	}
	return s.authService.ValidateUserRegistration(data)
}

// NormalizeUsername 规范化用户名
func (s *UserService) NormalizeUsername(username string) string {
	return s.authService.NormalizeUsername(username)
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(userID uint) (*models.User, error) {
	return s.daoManager.User.GetByID(userID)
}

// UpdateUserProfile 更新用户资料
func (s *UserService) UpdateUserProfile(userID uint, nickname, avatar string) error {
	return s.daoManager.User.UpdateColumns(userID, map[string]interface{}{
		"nickname": nickname,
		"avatar":   avatar,
	})
}

// GetUserFiles 获取用户上传的文件
func (s *UserService) GetUserFiles(userID uint, page, limit int) ([]models.FileCode, int64, error) {
	var total int64

	offset := (page - 1) * limit

	// 计算总数
	total, err := s.daoManager.FileCode.CountByUserID(userID)
	if err != nil {
		return nil, 0, err
	}

	// 获取文件列表 - 注意：这里需要添加分页支持到 DAO
	files, err := s.daoManager.FileCode.GetFilesByUserID(userID)
	if err != nil {
		return nil, 0, err
	}

	// 手动分页（后续可以优化到 DAO 层）
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

// UpdateUserStats 更新用户统计信息
func (s *UserService) UpdateUserStats(userID uint, uploadCount, downloadCount int, storageSize int64) error {
	return s.daoManager.User.UpdateColumns(userID, map[string]interface{}{
		"total_uploads":   gorm.Expr("total_uploads + ?", uploadCount),
		"total_downloads": gorm.Expr("total_downloads + ?", downloadCount),
		"total_storage":   gorm.Expr("total_storage + ?", storageSize),
	})
}

// CheckStorageQuota 检查用户存储配额
func (s *UserService) CheckStorageQuota(userID uint, additionalSize int64) error {
	user, err := s.daoManager.User.GetByID(userID)
	if err != nil {
		return fmt.Errorf("用户不存在: %w", err)
	}

	// 如果配额为0，表示无限制
	if user.MaxStorageQuota == 0 {
		return nil
	}

	if user.TotalStorage+additionalSize > user.MaxStorageQuota {
		return errors.New("存储空间不足")
	}

	return nil
}

// GetUserStats 获取用户统计信息
func (s *UserService) GetUserStats(userID uint) (map[string]interface{}, error) {
	user, err := s.daoManager.User.GetByID(userID)
	if err != nil {
		return nil, err
	}

	// 获取文件数量
	fileCount, err := s.daoManager.FileCode.CountByUserID(userID)
	if err != nil {
		return nil, err
	}

	// 获取今日上传数量 - 使用用户的文件数量作为近似值
	files, err := s.daoManager.FileCode.GetFilesByUserID(userID)
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

	stats := map[string]interface{}{
		"total_uploads":      user.TotalUploads,
		"total_downloads":    user.TotalDownloads,
		"total_storage":      user.TotalStorage,
		"total_files":        fileCount,
		"today_uploads":      todayUploads,
		"max_upload_size":    user.MaxUploadSize,
		"max_storage_quota":  user.MaxStorageQuota,
		"storage_usage":      user.TotalStorage,                                                // 实际使用的字节数
		"storage_percentage": float64(user.TotalStorage) / float64(user.MaxStorageQuota) * 100, // 使用百分比
	}

	return stats, nil
}

// IsUserSystemEnabled 检查用户系统是否启用
func (s *UserService) IsUserSystemEnabled() bool {
	return s.cfg.EnableUserSystem == 1
}

// IsRegistrationAllowed 检查是否允许用户注册
func (s *UserService) IsRegistrationAllowed() bool {
	return s.cfg.AllowUserRegistration == 1
}

// UpdateUserUploadStats 更新用户上传统计信息
func (s *UserService) UpdateUserUploadStats(userID uint, fileSize int64) error {
	return s.daoManager.User.UpdateColumns(userID, map[string]interface{}{
		"total_uploads": gorm.Expr("total_uploads + ?", 1),
		"total_storage": gorm.Expr("total_storage + ?", fileSize),
	})
}

// UpdateUserDownloadStats 更新用户下载统计信息
func (s *UserService) UpdateUserDownloadStats(userID uint) error {
	return s.daoManager.User.UpdateColumns(userID, map[string]interface{}{
		"total_downloads": gorm.Expr("total_downloads + ?", 1),
	})
}

// DeleteUserFile 删除用户文件
func (s *UserService) DeleteUserFile(userID uint, fileID uint) error {
	// 首先检查文件是否属于该用户
	fileCode, err := s.daoManager.FileCode.GetByUserID(userID, fileID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("文件不存在或您没有权限删除该文件")
		}
		return fmt.Errorf("查询文件失败: %w", err)
	}

	// 开始事务
	tx := s.daoManager.BeginTransaction()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除文件记录
	if err := s.daoManager.FileCode.DeleteByUserID(tx, userID); err != nil {
		tx.Rollback()
		return fmt.Errorf("删除文件记录失败: %w", err)
	}

	// 更新用户存储统计（减去删除的文件大小）
	if err := tx.Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"total_storage": gorm.Expr("total_storage - ?", fileCode.Size),
		}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新用户存储统计失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	// TODO: 删除实际的物理文件
	// 这里应该调用存储管理器来删除物理文件
	// 但为了简化，暂时只删除数据库记录

	return nil
}

// DeleteUserFileByCode 根据code删除用户文件
func (s *UserService) DeleteUserFileByCode(userID uint, code string) error {
	// 首先根据code查找文件
	fileCode, err := s.daoManager.FileCode.GetByCode(code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("文件不存在")
		}
		return fmt.Errorf("查询文件失败: %w", err)
	}

	// 检查文件是否属于该用户
	if fileCode.UserID == nil || *fileCode.UserID != userID {
		return errors.New("文件不存在或您没有权限删除该文件")
	}

	// 调用DeleteUserFile方法
	return s.DeleteUserFile(userID, fileCode.ID)
}
