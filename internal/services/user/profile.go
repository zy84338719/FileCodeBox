package user

import (
	"errors"
	"fmt"
	"strings"

	"github.com/zy84338719/filecodebox/internal/models"
	"github.com/zy84338719/filecodebox/internal/models/service"
)

// GetProfile 获取用户资料
func (s *Service) GetProfile(userID uint) (*models.User, error) {
	return s.repositoryManager.User.GetByID(userID)
}

// UpdateProfile 更新用户资料 - 使用结构化更新
func (s *Service) UpdateProfile(userID uint, updates map[string]interface{}) error {
	// 验证更新字段
	allowedFields := map[string]bool{
		"nickname": true,
		"email":    true,
		"avatar":   true,
	}

	for key := range updates {
		if !allowedFields[key] {
			return errors.New("field not allowed to update: " + key)
		}
	}

	user, err := s.repositoryManager.User.GetByID(userID)
	if err != nil {
		return err
	}

	// 准备结构化更新字段
	profileFields := &models.UserProfileUpdateFields{}

	// 检查邮箱是否已被其他用户使用
	if email, ok := updates["email"]; ok {
		emailStr := email.(string)
		if emailStr != user.Email {
			existingUser, err := s.repositoryManager.User.GetByEmail(emailStr)
			if err == nil && existingUser.ID != userID {
				return errors.New("email already in use")
			}
		}
		profileFields.Email = &emailStr
	}

	if nickname, ok := updates["nickname"]; ok {
		nicknameStr := nickname.(string)
		profileFields.Nickname = &nicknameStr
	}

	if avatar, ok := updates["avatar"]; ok {
		avatarStr := avatar.(string)
		profileFields.Avatar = &avatarStr
	}

	// 使用结构化更新
	return s.repositoryManager.User.UpdateUserProfile(userID, profileFields)
}

// ChangePassword 修改密码
func (s *Service) ChangePassword(userID uint, oldPassword, newPassword string) error {
	return s.authService.ChangePassword(userID, oldPassword, newPassword)
}

// GetUserStats 获取用户统计信息
func (s *Service) GetUserStats(userID uint) (*service.UserStatsData, error) {
	user, err := s.repositoryManager.User.GetByID(userID)
	if err != nil {
		return nil, err
	}

	// 获取用户上传的文件数量
	fileCount, err := s.repositoryManager.FileCode.CountByUserID(userID)
	if err != nil {
		return nil, err
	}

	return &service.UserStatsData{
		TotalUploads:    user.TotalUploads,
		TotalDownloads:  user.TotalDownloads,
		TotalStorage:    user.TotalStorage,
		MaxUploadSize:   user.MaxUploadSize,
		MaxStorageQuota: user.MaxStorageQuota,
		CurrentFiles:    int(fileCount),
		FileCount:       int(fileCount),
		LastLoginAt:     user.LastLoginAt,
		LastLoginIP:     user.LastLoginIP,
		EmailVerified:   user.EmailVerified,
		Status:          user.Status,
		Role:            user.Role,
		LastUploadAt:    nil, // TODO: 从文件记录中获取
		LastDownloadAt:  nil, // TODO: 从下载记录中获取
	}, nil
}

// UpdateUserStats 更新用户统计信息
func (s *Service) UpdateUserStats(userID uint, statsType string, value int64) error {
	user, err := s.repositoryManager.User.GetByID(userID)
	if err != nil {
		return err
	}

	switch statsType {
	case "upload":
		user.TotalUploads++
		user.TotalStorage += value
	case "download":
		user.TotalDownloads++
	case "delete":
		// 删除文件时，减少存储使用量（value应该是负数）
		user.TotalStorage += value
		// 确保存储使用量不会变成负数
		if user.TotalStorage < 0 {
			user.TotalStorage = 0
		}
	default:
		return errors.New("invalid stats type")
	}

	return s.repositoryManager.User.Update(user)
}

// CheckUserQuota 检查用户配额
func (s *Service) CheckUserQuota(userID uint, fileSize int64) error {
	user, err := s.repositoryManager.User.GetByID(userID)
	if err != nil {
		return err
	}

	// 检查单次上传大小限制
	if user.MaxUploadSize > 0 && fileSize > user.MaxUploadSize {
		return errors.New("file size exceeds user upload limit")
	}

	// 检查存储配额
	if user.MaxStorageQuota > 0 {
		if user.TotalStorage+fileSize > user.MaxStorageQuota {
			return errors.New("storage quota exceeded")
		}
	}

	return nil
}

// DeleteAccount 删除用户账户
func (s *Service) DeleteAccount(userID uint) error {
	// 首先获取用户对象
	user, err := s.repositoryManager.User.GetByID(userID)
	if err != nil {
		return err
	}

	// 删除用户的所有文件
	files, err := s.repositoryManager.FileCode.GetFilesByUserID(userID)
	if err != nil {
		return err
	}

	for _, file := range files {
		// 这里应该删除实际文件，但需要存储服务支持
		// 暂时只删除数据库记录
		err = s.repositoryManager.FileCode.DeleteByFileCode(&file)
		if err != nil {
			return err
		}
	}

	// 删除用户记录
	tx := s.repositoryManager.BeginTransaction()
	return s.repositoryManager.User.Delete(tx, user)
}

// IsUserSystemEnabled 检查用户系统是否启用 - 始终返回true
func (s *Service) IsUserSystemEnabled() bool {
	return true
}

// IsSystemInitialized 检测系统是否已初始化（是否有管理员用户）
func (s *Service) IsSystemInitialized() (bool, error) {
	adminCount, err := s.repositoryManager.User.CountAdminUsers()
	if err != nil {
		return false, fmt.Errorf("检查管理员用户失败: %w", err)
	}
	return adminCount > 0, nil
}

// ValidateUserInput 验证用户输入
func (s *Service) ValidateUserInput(username, email, password string) error {
	if len(username) < 3 {
		return fmt.Errorf("用户名长度至少3个字符")
	}
	if len(password) < 6 {
		return fmt.Errorf("密码长度至少6个字符")
	}
	// 简单的邮箱验证
	if !strings.Contains(email, "@") {
		return fmt.Errorf("邮箱格式无效")
	}
	return nil
}

// NormalizeUsername 规范化用户名
func (s *Service) NormalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

// Register 用户注册
func (s *Service) Register(username, email, password, nickname string) (*models.User, error) {
	// 验证输入
	if err := s.ValidateUserInput(username, email, password); err != nil {
		return nil, err
	}

	// 规范化用户名
	username = s.NormalizeUsername(username)

	// 检查是否已存在
	existingUser, _ := s.repositoryManager.User.GetByUsernameOrEmail(username)
	if existingUser != nil {
		return nil, fmt.Errorf("用户名或邮箱已存在")
	}

	// 加密密码
	hashedPassword, err := s.authService.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 设置默认昵称
	if nickname == "" {
		nickname = username
	}

	// 创建用户
	user := &models.User{
		Username:     username,
		Email:        email,
		PasswordHash: hashedPassword,
		Nickname:     nickname,
		Status:       "active",
		Role:         "user",
	}

	err = s.repositoryManager.User.Create(user)
	return user, err
}

// Login 用户登录并返回token
func (s *Service) Login(username, password, ipAddress, userAgent string) (string, *models.User, error) {
	// 通过认证服务验证用户
	user, err := s.authService.Login(username, password)
	if err != nil {
		return "", nil, fmt.Errorf("用户名或密码错误")
	}

	// 创建用户会话并生成JWT令牌
	token, err := s.authService.CreateUserSession(user, ipAddress, userAgent)
	if err != nil {
		return "", nil, err
	}

	return token, user, nil
}

// Logout 用户登出 (兼容性方法)
func (s *Service) Logout(userID uint) error {
	// 这里可以实现登出逻辑，比如清除会话等
	// 目前只是一个占位符
	return nil
}

// GetUserByID 根据ID获取用户 (兼容性方法)
func (s *Service) GetUserByID(userID uint) (*models.User, error) {
	return s.repositoryManager.User.GetByID(userID)
}

// CheckPasswordHash 验证密码哈希
func (s *Service) CheckPasswordHash(password, hash string) bool {
	// 这里应该使用实际的密码验证逻辑
	// 假设使用 bcrypt 或类似的哈希算法
	return password == hash // 这是简化实现，实际应该使用哈希比较
}

// UpdateUserProfile 更新用户资料 (兼容性方法)
func (s *Service) UpdateUserProfile(userID uint, nickname, avatar string) error {
	user, err := s.GetUserByID(userID)
	if err != nil {
		return err
	}

	user.Nickname = nickname
	user.Avatar = avatar

	return s.repositoryManager.User.Update(user)
}

// GetUserFiles 获取用户文件列表 (兼容性方法)
func (s *Service) GetUserFiles(userID uint, page, limit int) (interface{}, int64, error) {
	// 使用仓库层的分页查询获取用户文件
	files, total, err := s.repositoryManager.FileCode.GetFilesByUserIDWithPagination(userID, page, limit)
	if err != nil {
		return nil, 0, err
	}

	// 直接返回files而不是转换为interface{}切片
	return files, total, nil
}

// IsRegistrationAllowed 检查是否允许注册 (兼容性方法)
func (s *Service) IsRegistrationAllowed() bool {
	// 检查用户系统是否启用
	return s.manager.IsUserSystemEnabled()
}

// DeleteUserFileByCode 根据代码删除用户文件 (兼容性方法)
func (s *Service) DeleteUserFileByCode(userID uint, code string) error {
	// 这是一个占位符实现
	return nil
}

// ValidateToken 验证令牌 (兼容性方法)
func (s *Service) ValidateToken(token string) (interface{}, error) {
	// 这应该委托给 auth 服务
	return s.authService.ValidateToken(token)
}
