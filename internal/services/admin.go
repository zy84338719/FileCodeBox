package services

import (
	"fmt"
	"time"

	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/dao"
	"github.com/zy84338719/filecodebox/internal/models"
	"github.com/zy84338719/filecodebox/internal/storage"

	"gorm.io/gorm"
)

// AdminService 管理服务
type AdminService struct {
	config         *config.Config
	storageManager *storage.StorageManager
	daoManager     *dao.DAOManager
}

func NewAdminService(db *gorm.DB, config *config.Config, storageManager *storage.StorageManager) *AdminService {
	return &AdminService{
		config:         config,
		storageManager: storageManager,
		daoManager:     dao.NewDAOManager(db),
	}
}

// GetStats 获取统计信息
func (s *AdminService) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总文件数（不包括已删除的）
	totalFiles, err := s.daoManager.FileCode.Count()
	if err != nil {
		return nil, err
	}
	stats["total_files"] = totalFiles

	// 今日上传数（包括已删除的，因为这是历史统计）
	todayFiles, err := s.daoManager.FileCode.CountToday()
	if err != nil {
		return nil, err
	}
	stats["today_files"] = todayFiles

	// 活跃文件数（未过期且未删除）
	activeFiles, err := s.daoManager.FileCode.CountActive()
	if err != nil {
		return nil, err
	}
	stats["active_files"] = activeFiles

	// 总大小（不包括已删除的）
	totalSize, err := s.daoManager.FileCode.GetTotalSize()
	if err != nil {
		return nil, err
	}
	stats["total_size"] = totalSize

	// 系统启动时间
	sysStart, err := s.daoManager.KeyValue.GetByKey("sys_start")
	if err == nil {
		stats["sys_start"] = sysStart.Value
	} else {
		// 如果没有记录，创建一个
		startTime := fmt.Sprintf("%d", time.Now().UnixMilli())
		err := s.daoManager.KeyValue.SetValue("sys_start", startTime)
		if err != nil {
			return nil, err
		}
		stats["sys_start"] = startTime
	}

	return stats, nil
}

// GetFiles 获取文件列表
func (s *AdminService) GetFiles(page, pageSize int, search string) ([]models.FileCode, int64, error) {
	return s.daoManager.FileCode.List(page, pageSize, search)
}

// DeleteFile 删除文件
func (s *AdminService) DeleteFile(id uint) error {
	fileCode, err := s.daoManager.FileCode.GetByID(id)
	if err != nil {
		return err
	}

	// 删除实际文件
	storageInterface := s.storageManager.GetStorage()
	if err := storageInterface.DeleteFile(fileCode); err != nil {
		// 记录错误，但不阻止数据库删除
		fmt.Printf("Warning: Failed to delete physical file: %v\n", err)
	}

	return s.daoManager.FileCode.DeleteByFileCode(fileCode)
}

// GetConfig 获取配置
func (s *AdminService) GetConfig() (*config.Config, error) {
	return s.config, nil
}

// UpdateConfig 更新配置
func (s *AdminService) UpdateConfig(newConfig map[string]interface{}) error {
	// 更新内存中的配置
	if name, ok := newConfig["name"].(string); ok {
		s.config.Name = name
	}
	if description, ok := newConfig["description"].(string); ok {
		s.config.Description = description
	}
	if keywords, ok := newConfig["keywords"].(string); ok {
		s.config.Keywords = keywords
	}
	if uploadSize, ok := newConfig["upload_size"].(float64); ok {
		s.config.UploadSize = int64(uploadSize) // 前端已转换为字节，无需再次转换
	}
	if uploadSizeInt, ok := newConfig["upload_size"].(int); ok {
		s.config.UploadSize = int64(uploadSizeInt) // 前端已转换为字节，无需再次转换
	}
	if adminToken, ok := newConfig["admin_token"].(string); ok && adminToken != "" {
		s.config.AdminToken = adminToken
	}
	if pageExplain, ok := newConfig["page_explain"].(string); ok {
		s.config.PageExplain = pageExplain
	}
	if notifyTitle, ok := newConfig["notify_title"].(string); ok {
		s.config.NotifyTitle = notifyTitle
	}
	if notifyContent, ok := newConfig["notify_content"].(string); ok {
		s.config.NotifyContent = notifyContent
	}
	if openUpload, ok := newConfig["open_upload"].(float64); ok {
		s.config.OpenUpload = int(openUpload)
	}
	if enableChunk, ok := newConfig["enable_chunk"].(float64); ok {
		s.config.EnableChunk = int(enableChunk)
	}

	// 保存配置（会同时保存到文件和数据库）
	return s.config.Save()
}

// CleanExpiredFiles 清理过期文件
func (s *AdminService) CleanExpiredFiles() (int, error) {
	// 查找过期文件
	expiredFiles, err := s.daoManager.FileCode.GetExpiredFiles()
	if err != nil {
		return 0, err
	}

	count := len(expiredFiles)
	storageInterface := s.storageManager.GetStorage()

	// 删除过期文件
	for _, file := range expiredFiles {
		// 删除实际文件
		if err := storageInterface.DeleteFile(&file); err != nil {
			fmt.Printf("Warning: Failed to delete physical file %s: %v\n", file.Code, err)
		}
		// 删除数据库记录
		s.daoManager.FileCode.DeleteByFileCode(&file)
	}

	return count, nil
}

// UpdateFile 更新文件信息
func (s *AdminService) UpdateFile(id uint, code, prefix, suffix string, expiredAt *time.Time, expiredCount *int) error {
	fileCode, err := s.daoManager.FileCode.GetByID(id)
	if err != nil {
		return err
	}

	updates := make(map[string]interface{})

	if code != "" && code != fileCode.Code {
		// 检查代码是否已存在
		exists, err := s.daoManager.FileCode.CheckCodeExists(code, id)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("代码已存在")
		}
		updates["code"] = code
	}

	if prefix != "" && prefix != fileCode.Prefix {
		updates["prefix"] = prefix
	}

	if suffix != "" && suffix != fileCode.Suffix {
		updates["suffix"] = suffix
	}

	if expiredAt != nil {
		updates["expired_at"] = expiredAt
	}

	if expiredCount != nil {
		updates["expired_count"] = *expiredCount
	}

	if len(updates) > 0 {
		return s.daoManager.FileCode.UpdateColumns(id, updates)
	}

	return nil
}

// GetFileByID 根据ID获取文件
func (s *AdminService) GetFileByID(id uint) (*models.FileCode, error) {
	return s.daoManager.FileCode.GetByID(id)
}

// ========== 用户管理相关方法 ==========

// GetUsers 获取用户列表
func (s *AdminService) GetUsers(page, pageSize int, search string) ([]models.User, int64, error) {
	return s.daoManager.User.List(page, pageSize, search)
}

// GetUserStats 获取用户统计信息
func (s *AdminService) GetUserStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总用户数
	totalUsers, err := s.daoManager.User.Count()
	if err != nil {
		return nil, err
	}
	stats["total_users"] = totalUsers

	// 活跃用户数
	activeUsers, err := s.daoManager.User.CountActive()
	if err != nil {
		return nil, err
	}
	stats["active_users"] = activeUsers

	// 今日注册数
	todayRegistrations, err := s.daoManager.User.CountTodayRegistrations()
	if err != nil {
		return nil, err
	}
	stats["today_registrations"] = todayRegistrations

	// 今日上传数
	todayUploads, err := s.daoManager.FileCode.CountToday()
	if err != nil {
		return nil, err
	}
	stats["today_uploads"] = todayUploads

	return stats, nil
}

// GetUserByID 根据ID获取用户
func (s *AdminService) GetUserByID(id uint) (*models.User, error) {
	return s.daoManager.User.GetByID(id)
}

// CreateUser 创建用户
func (s *AdminService) CreateUser(username, email, password, nickname, role, status string) (*models.User, error) {
	// 检查用户名和邮箱是否已存在
	existingUser, err := s.daoManager.User.CheckExists(username, email)
	if err == nil && existingUser != nil {
		if existingUser.Username == username {
			return nil, fmt.Errorf("用户名已存在")
		}
		return nil, fmt.Errorf("邮箱已存在")
	}

	user := &models.User{
		Username:     username,
		Email:        email,
		PasswordHash: password, // 在实际环境中应该进行哈希
		Nickname:     nickname,
		Role:         role,
		Status:       status,
	}

	err = s.daoManager.User.Create(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUser 更新用户
func (s *AdminService) UpdateUser(id uint, email, password, nickname, role, status string) error {
	// 检查用户是否存在
	_, err := s.daoManager.User.GetByID(id)
	if err != nil {
		return err
	}

	// 检查邮箱是否被其他用户使用
	existingUser, err := s.daoManager.User.CheckExists("", email)
	if err == nil && existingUser != nil && existingUser.ID != id {
		return fmt.Errorf("邮箱已被其他用户使用")
	}

	updates := map[string]interface{}{
		"email":    email,
		"nickname": nickname,
		"role":     role,
		"status":   status,
	}

	// 如果提供了密码，更新密码
	if password != "" {
		updates["password_hash"] = password
	}

	return s.daoManager.User.UpdateColumns(id, updates)
}

// DeleteUser 删除用户
func (s *AdminService) DeleteUser(id uint) error {
	// 检查用户是否存在
	user, err := s.daoManager.User.GetByID(id)
	if err != nil {
		return err
	}

	// 不允许删除管理员账户
	if user.Role == "admin" {
		return fmt.Errorf("不能删除管理员账户")
	}

	// 开始事务
	tx := s.daoManager.BeginTransaction()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除用户的所有文件
	err = s.daoManager.FileCode.DeleteByUserID(tx, id)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("删除用户文件失败: %v", err)
	}

	// 删除用户的会话
	err = s.daoManager.UserSession.DeleteByUserID(tx, id)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("删除用户会话失败: %v", err)
	}

	// 删除用户
	err = s.daoManager.User.Delete(tx, user)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("删除用户失败: %v", err)
	}

	// 提交事务
	return tx.Commit().Error
}

// UpdateUserStatus 更新用户状态
func (s *AdminService) UpdateUserStatus(id uint, isActive bool) error {
	// 检查用户是否存在
	user, err := s.daoManager.User.GetByID(id)
	if err != nil {
		return err
	}

	// 不允许禁用管理员账户
	if user.Role == "admin" && !isActive {
		return fmt.Errorf("不能禁用管理员账户")
	}

	status := "active"
	if !isActive {
		status = "inactive"
	}

	err = s.daoManager.User.UpdateColumns(id, map[string]interface{}{
		"status": status,
	})
	if err != nil {
		return err
	}

	// 如果禁用用户，同时禁用其所有会话
	if !isActive {
		err = s.daoManager.UserSession.DeactivateUserSessions(id)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetUserFiles 获取用户文件
func (s *AdminService) GetUserFiles(userID uint) ([]models.FileCode, error) {
	// 检查用户是否存在
	_, err := s.daoManager.User.GetByID(userID)
	if err != nil {
		return nil, err
	}

	return s.daoManager.FileCode.GetFilesByUserID(userID)
}
