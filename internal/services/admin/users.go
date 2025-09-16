package admin

import (
	"errors"
	"fmt"

	"github.com/zy84338719/filecodebox/internal/models"
	"gorm.io/gorm"
)

// GetUsers 获取用户列表
func (s *Service) GetUsers(page, pageSize int, search string) ([]models.User, int64, error) {
	return s.repositoryManager.User.List(page, pageSize, search)
}

// GetUser 获取用户信息
func (s *Service) GetUser(id uint) (*models.User, error) {
	return s.repositoryManager.User.GetByID(id)
}

// CreateUser 创建用户 - 使用统一的认证服务
func (s *Service) CreateUser(username, email, password, nickname, role, status string) (*models.User, error) {
	// 哈希密码
	hashedPassword, err := s.authService.HashPassword(password)
	if err != nil {
		return nil, err
	}

	// 检查用户名唯一性
	existingUser, err := s.repositoryManager.User.GetByUsername(username)
	if err == nil && existingUser != nil {
		return nil, errors.New("用户名已存在")
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("检查用户名唯一性失败: %w", err)
	}

	// 检查邮箱唯一性（如果提供了邮箱）
	if email != "" {
		existingUser, err := s.repositoryManager.User.GetByEmail(email)
		if err == nil && existingUser != nil {
			return nil, errors.New("该邮箱已被使用")
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("检查邮箱唯一性失败: %w", err)
		}
	}

	user := &models.User{
		Username:        username,
		Email:           email,
		PasswordHash:    hashedPassword,
		Nickname:        nickname,
		Role:            role,
		Status:          status,
		EmailVerified:   true, // 管理员创建的用户默认已验证
		MaxUploadSize:   s.manager.User.UserUploadSize,
		MaxStorageQuota: s.manager.User.UserStorageQuota,
	}

	if err := s.repositoryManager.User.Create(user); err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	return user, nil
}

// UpdateUser 更新用户 - 使用结构化更新
func (s *Service) UpdateUser(id uint, email, password, nickname, role, status string) error {
	// 准备更新字段
	updateFields := &models.UserUpdateFields{}

	if email != "" {
		// 检查邮箱是否已被其他用户使用
		existingUser, err := s.repositoryManager.User.CheckEmailExists(email, id)
		if err == nil && existingUser != nil {
			return errors.New("该邮箱已被其他用户使用")
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("检查邮箱唯一性失败: %w", err)
		}
		updateFields.Email = &email
	}

	if password != "" {
		hashedPassword, err := s.authService.HashPassword(password)
		if err != nil {
			return err
		}
		updateFields.PasswordHash = &hashedPassword
	}

	if nickname != "" {
		updateFields.Nickname = &nickname
	}
	if role != "" {
		updateFields.Role = &role
	}
	if status != "" {
		updateFields.Status = &status
	}

	return s.repositoryManager.User.UpdateUserFields(id, updateFields)
}

// DeleteUser 删除用户
func (s *Service) DeleteUser(id uint) error {
	user, err := s.repositoryManager.User.GetByID(id)
	if err != nil {
		return err
	}
	// 开始事务
	tx := s.repositoryManager.BeginTransaction()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	err = s.repositoryManager.User.Delete(tx, user)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// ToggleUserStatus 切换用户状态
func (s *Service) ToggleUserStatus(id uint) error {
	user, err := s.repositoryManager.User.GetByID(id)
	if err != nil {
		return err
	}

	// 切换状态
	var newStatus string
	if user.Status == "active" {
		newStatus = "inactive"
	} else {
		newStatus = "active"
	}

	updateFields := &models.UserUpdateFields{
		Status: &newStatus,
	}

	return s.repositoryManager.User.UpdateUserFields(id, updateFields)
}

// GetUserByID 根据ID获取用户 (兼容性方法)
func (s *Service) GetUserByID(id uint) (*models.User, error) {
	return s.GetUser(id)
}

// BatchUpdateUserStatus 批量更新用户状态：enable=true 表示启用(active)，false 表示禁用(inactive)
func (s *Service) BatchUpdateUserStatus(userIDs []uint, enable bool) error {
	if len(userIDs) == 0 {
		return nil
	}

	status := "inactive"
	if enable {
		status = "active"
	}

	tx := s.repositoryManager.BeginTransaction()
	if tx == nil {
		return errors.New("无法开始数据库事务")
	}

	if err := tx.Model(&models.User{}).Where("id IN ?", userIDs).Update("status", status).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// BatchDeleteUsers 批量删除用户
func (s *Service) BatchDeleteUsers(userIDs []uint) error {
	if len(userIDs) == 0 {
		return nil
	}

	tx := s.repositoryManager.BeginTransaction()
	if tx == nil {
		return errors.New("无法开始数据库事务")
	}

	if err := tx.Where("id IN ?", userIDs).Delete(&models.User{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
