package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/dao"
	"github.com/zy84338719/filecodebox/internal/models"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService 用户服务
type UserService struct {
	cfg        *config.Config
	daoManager *dao.DAOManager
}

// NewUserService 创建用户服务
func NewUserService(db *gorm.DB, cfg *config.Config) *UserService {
	return &UserService{
		cfg:        cfg,
		daoManager: dao.NewDAOManager(db),
	}
}

// UserClaims JWT claims
type UserClaims struct {
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	SessionID string `json:"session_id"`
	jwt.RegisteredClaims
}

// Register 用户注册
func (s *UserService) Register(username, email, password, nickname string) (*models.User, error) {
	// 检查是否允许注册
	if s.cfg.AllowUserRegistration == 0 {
		return nil, errors.New("用户注册已禁用")
	}

	// 验证用户名和邮箱的唯一性
	existingUser, err := s.daoManager.User.CheckExists(username, email)
	if err == nil {
		if existingUser.Username == username {
			return nil, errors.New("用户名已存在")
		}
		return nil, errors.New("邮箱已存在")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("检查用户唯一性失败: %w", err)
	}

	// 密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码哈希失败: %w", err)
	}

	// 创建用户
	user := &models.User{
		Username:        username,
		Email:           email,
		PasswordHash:    string(hashedPassword),
		Nickname:        nickname,
		Role:            "user",
		Status:          "active",
		EmailVerified:   s.cfg.RequireEmailVerify == 0, // 如果不需要验证邮箱，则直接设为已验证
		MaxUploadSize:   s.cfg.UserUploadSize,
		MaxStorageQuota: s.cfg.UserStorageQuota,
	}

	if err := s.daoManager.User.Create(user); err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	return user, nil
}

// Login 用户登录
func (s *UserService) Login(usernameOrEmail, password, ipAddress, userAgent string) (string, *models.User, error) {
	// 查找用户
	user, err := s.daoManager.User.GetByUsernameOrEmail(usernameOrEmail)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, errors.New("用户名或密码错误")
		}
		return "", nil, fmt.Errorf("查询用户失败: %w", err)
	}

	// 检查用户状态
	if user.Status != "active" {
		return "", nil, errors.New("用户账号已被禁用")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", nil, errors.New("用户名或密码错误")
	}

	// 检查邮箱验证
	if s.cfg.RequireEmailVerify == 1 && !user.EmailVerified {
		return "", nil, errors.New("请先验证邮箱")
	}

	// 清理过期会话
	if err := s.cleanExpiredSessions(user.ID); err != nil {
		// 记录错误但不阻止登录
		fmt.Printf("清理过期会话失败: %v\n", err)
	}

	// 检查会话数量限制
	sessionCount, err := s.daoManager.UserSession.CountActiveSessionsByUserID(user.ID)
	if err != nil {
		return "", nil, fmt.Errorf("检查会话数量失败: %w", err)
	}

	if sessionCount >= int64(s.cfg.MaxSessionsPerUser) {
		// 删除最旧的会话
		oldestSession, err := s.daoManager.UserSession.GetOldestSessionByUserID(user.ID)
		if err == nil {
			s.daoManager.UserSession.UpdateIsActive(oldestSession, false)
		}
	}

	// 生成会话ID
	sessionID, err := s.generateSessionID()
	if err != nil {
		return "", nil, fmt.Errorf("生成会话ID失败: %w", err)
	}

	// 创建JWT token
	expirationTime := time.Now().Add(time.Duration(s.cfg.SessionExpiryHours) * time.Hour)
	claims := &UserClaims{
		UserID:    user.ID,
		Username:  user.Username,
		Role:      user.Role,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", nil, fmt.Errorf("生成token失败: %w", err)
	}

	// 保存会话
	session := &models.UserSession{
		UserID:    user.ID,
		SessionID: sessionID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		ExpiresAt: expirationTime,
		IsActive:  true,
	}

	if err := s.daoManager.UserSession.Create(session); err != nil {
		return "", nil, fmt.Errorf("保存会话失败: %w", err)
	}

	// 更新用户最后登录信息
	now := time.Now()
	user.LastLoginAt = &now
	user.LastLoginIP = ipAddress
	if err := s.daoManager.User.Update(user); err != nil {
		// 记录错误但不阻止登录
		fmt.Printf("更新用户登录信息失败: %v\n", err)
	}

	return tokenString, user, nil
}

// ValidateToken 验证token
func (s *UserService) ValidateToken(tokenString string) (interface{}, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		// 检查会话是否仍然有效
		session, err := s.daoManager.UserSession.GetBySessionID(claims.SessionID)
		if err != nil {
			return nil, errors.New("会话已失效")
		}

		// 检查会话是否过期
		if session.ExpiresAt.Before(time.Now()) {
			// 标记会话为无效
			s.daoManager.UserSession.UpdateIsActive(session, false)
			return nil, errors.New("会话已过期")
		}

		return claims, nil
	}

	return nil, errors.New("无效的token")
}

// Logout 用户登出
func (s *UserService) Logout(sessionID string) error {
	// 获取会话并标记为无效
	session, err := s.daoManager.UserSession.GetBySessionID(sessionID)
	if err != nil {
		return err
	}
	return s.daoManager.UserSession.UpdateIsActive(session, false)
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

// ChangePassword 修改密码
func (s *UserService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	user, err := s.daoManager.User.GetByID(userID)
	if err != nil {
		return fmt.Errorf("用户不存在: %w", err)
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("原密码错误")
	}

	// 哈希新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码哈希失败: %w", err)
	}

	// 更新密码
	return s.daoManager.User.UpdatePassword(userID, string(hashedPassword))
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

// generateSessionID 生成会话ID
func (s *UserService) generateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// cleanExpiredSessions 清理过期会话
func (s *UserService) cleanExpiredSessions(userID uint) error {
	return s.daoManager.UserSession.CleanExpiredSessions()
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

// NormalizeUsername 规范化用户名
func (s *UserService) NormalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

// ValidateUserInput 验证用户输入
func (s *UserService) ValidateUserInput(username, email, password string) error {
	if len(username) < 3 || len(username) > 50 {
		return errors.New("用户名长度必须在3-50个字符之间")
	}

	if len(password) < 6 {
		return errors.New("密码长度至少6个字符")
	}

	// 简单的邮箱格式验证
	if !strings.Contains(email, "@") || len(email) < 5 {
		return errors.New("邮箱格式不正确")
	}

	return nil
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
