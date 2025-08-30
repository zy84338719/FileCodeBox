package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/dao"
	"github.com/zy84338719/filecodebox/internal/models"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthService 认证服务 - 提供统一的用户认证、密码处理、会话管理等功能
type AuthService struct {
	cfg        *config.Config
	daoManager *dao.DAOManager
}

// NewAuthService 创建认证服务
func NewAuthService(db *gorm.DB, cfg *config.Config) *AuthService {
	return &AuthService{
		cfg:        cfg,
		daoManager: dao.NewDAOManager(db),
	}
}

// AuthClaims JWT声明
type AuthClaims struct {
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	SessionID string `json:"session_id"`
	jwt.RegisteredClaims
}

// PasswordPolicy 密码策略
type PasswordPolicy struct {
	MinLength      int
	MaxLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireNumber  bool
	RequireSpecial bool
	ForbidCommon   bool
	ForbidPersonal bool
}

// UserRegistrationData 用户注册数据
type UserRegistrationData struct {
	Username string
	Email    string
	Password string
	Nickname string
}

// UserLoginData 用户登录数据
type UserLoginData struct {
	Username  string
	Password  string
	IPAddress string
	UserAgent string
}

// PasswordChangeData 密码修改数据
type PasswordChangeData struct {
	UserID      uint
	OldPassword string
	NewPassword string
}

// PasswordResetData 密码重置数据
type PasswordResetData struct {
	Token       string
	NewPassword string
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ================= 密码处理相关方法 =================

// HashPassword 对密码进行哈希处理
func (s *AuthService) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("密码哈希失败: %w", err)
	}
	return string(hashedPassword), nil
}

// VerifyPassword 验证密码是否正确
func (s *AuthService) VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// GetPasswordPolicy 获取当前密码策略
func (s *AuthService) GetPasswordPolicy() PasswordPolicy {
	return PasswordPolicy{
		MinLength:      6,
		MaxLength:      128,
		RequireUpper:   false,
		RequireLower:   false,
		RequireNumber:  false,
		RequireSpecial: false,
		ForbidCommon:   true,
		ForbidPersonal: true,
	}
}

// ValidatePassword 验证密码是否符合策略
func (s *AuthService) ValidatePassword(password string, username string, email string) error {
	policy := s.GetPasswordPolicy()

	// 检查长度
	if len(password) < policy.MinLength {
		return ValidationError{
			Field:   "password",
			Message: fmt.Sprintf("密码长度至少需要%d个字符", policy.MinLength),
		}
	}
	if len(password) > policy.MaxLength {
		return ValidationError{
			Field:   "password",
			Message: fmt.Sprintf("密码长度不能超过%d个字符", policy.MaxLength),
		}
	}

	// 检查字符类型要求
	if policy.RequireUpper && !containsUppercase(password) {
		return ValidationError{
			Field:   "password",
			Message: "密码必须包含至少一个大写字母",
		}
	}
	if policy.RequireLower && !containsLowercase(password) {
		return ValidationError{
			Field:   "password",
			Message: "密码必须包含至少一个小写字母",
		}
	}
	if policy.RequireNumber && !containsNumber(password) {
		return ValidationError{
			Field:   "password",
			Message: "密码必须包含至少一个数字",
		}
	}
	if policy.RequireSpecial && !containsSpecialChar(password) {
		return ValidationError{
			Field:   "password",
			Message: "密码必须包含至少一个特殊字符",
		}
	}

	// 检查是否包含常见密码
	if policy.ForbidCommon && isCommonPassword(password) {
		return ValidationError{
			Field:   "password",
			Message: "密码过于简单，请使用更复杂的密码",
		}
	}

	// 检查是否包含个人信息
	if policy.ForbidPersonal && containsPersonalInfo(password, username, email) {
		return ValidationError{
			Field:   "password",
			Message: "密码不能包含用户名或邮箱信息",
		}
	}

	return nil
}

// ================= 用户输入验证方法 =================

// ValidateUsername 验证用户名
func (s *AuthService) ValidateUsername(username string) error {
	username = strings.TrimSpace(username)

	if len(username) < 3 {
		return ValidationError{
			Field:   "username",
			Message: "用户名长度至少需要3个字符",
		}
	}
	if len(username) > 50 {
		return ValidationError{
			Field:   "username",
			Message: "用户名长度不能超过50个字符",
		}
	}

	// 检查用户名格式 - 只允许字母、数字、下划线、短横线
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !usernameRegex.MatchString(username) {
		return ValidationError{
			Field:   "username",
			Message: "用户名只能包含字母、数字、下划线和短横线",
		}
	}

	return nil
}

// ValidateEmail 验证邮箱格式
func (s *AuthService) ValidateEmail(email string) error {
	email = strings.TrimSpace(email)

	if len(email) < 5 {
		return ValidationError{
			Field:   "email",
			Message: "邮箱格式不正确",
		}
	}

	// 简单的邮箱格式验证
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return ValidationError{
			Field:   "email",
			Message: "邮箱格式不正确",
		}
	}

	return nil
}

// ValidateUserRegistration 验证用户注册数据
func (s *AuthService) ValidateUserRegistration(data UserRegistrationData) error {
	// 验证用户名
	if err := s.ValidateUsername(data.Username); err != nil {
		return err
	}

	// 验证邮箱
	if err := s.ValidateEmail(data.Email); err != nil {
		return err
	}

	// 验证密码
	if err := s.ValidatePassword(data.Password, data.Username, data.Email); err != nil {
		return err
	}

	return nil
}

// ================= 用户注册方法 =================

// RegisterUser 用户注册
func (s *AuthService) RegisterUser(data UserRegistrationData) (*models.User, error) {
	// 检查是否允许注册
	if s.cfg.AllowUserRegistration == 0 {
		return nil, errors.New("用户注册已禁用")
	}

	// 规范化数据
	data.Username = s.NormalizeUsername(data.Username)
	data.Email = s.NormalizeEmail(data.Email)
	if data.Nickname == "" {
		data.Nickname = data.Username
	}

	// 验证注册数据
	if err := s.ValidateUserRegistration(data); err != nil {
		return nil, err
	}

	// 检查用户名和邮箱的唯一性
	if err := s.checkUserUniqueness(data.Username, data.Email); err != nil {
		return nil, err
	}

	// 哈希密码
	hashedPassword, err := s.HashPassword(data.Password)
	if err != nil {
		return nil, err
	}

	// 创建用户
	user := &models.User{
		Username:        data.Username,
		Email:           data.Email,
		PasswordHash:    hashedPassword,
		Nickname:        data.Nickname,
		Role:            "user",
		Status:          "active",
		EmailVerified:   s.cfg.RequireEmailVerify == 0,
		MaxUploadSize:   s.cfg.UserUploadSize,
		MaxStorageQuota: s.cfg.UserStorageQuota,
	}

	if err := s.daoManager.User.Create(user); err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	return user, nil
}

// ================= 用户登录方法 =================

// LoginUser 用户登录
func (s *AuthService) LoginUser(data UserLoginData) (string, *models.User, error) {
	// 查找用户
	user, err := s.daoManager.User.GetByUsernameOrEmail(data.Username)
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
	if err := s.VerifyPassword(user.PasswordHash, data.Password); err != nil {
		return "", nil, errors.New("用户名或密码错误")
	}

	// 检查邮箱验证
	if s.cfg.RequireEmailVerify == 1 && !user.EmailVerified {
		return "", nil, errors.New("请先验证邮箱")
	}

	// 创建会话
	token, err := s.createUserSession(user, data.IPAddress, data.UserAgent)
	if err != nil {
		return "", nil, err
	}

	// 更新登录信息
	s.updateUserLoginInfo(user, data.IPAddress)

	return token, user, nil
}

// ================= 密码修改方法 =================

// ChangeUserPassword 修改用户密码
func (s *AuthService) ChangeUserPassword(data PasswordChangeData) error {
	// 获取用户信息
	user, err := s.daoManager.User.GetByID(data.UserID)
	if err != nil {
		return fmt.Errorf("用户不存在: %w", err)
	}

	// 验证旧密码
	if err := s.VerifyPassword(user.PasswordHash, data.OldPassword); err != nil {
		return errors.New("原密码错误")
	}

	// 验证新密码
	if err := s.ValidatePassword(data.NewPassword, user.Username, user.Email); err != nil {
		return err
	}

	// 哈希新密码
	hashedPassword, err := s.HashPassword(data.NewPassword)
	if err != nil {
		return err
	}

	// 更新密码
	if err := s.daoManager.User.UpdatePassword(data.UserID, hashedPassword); err != nil {
		return fmt.Errorf("更新密码失败: %w", err)
	}

	// 可选：使所有其他会话失效
	s.invalidateOtherSessions(data.UserID, "")

	return nil
}

// ResetUserPassword 重置用户密码（管理员操作）
func (s *AuthService) ResetUserPassword(userID uint, newPassword string) error {
	// 获取用户信息
	user, err := s.daoManager.User.GetByID(userID)
	if err != nil {
		return fmt.Errorf("用户不存在: %w", err)
	}

	// 验证新密码
	if err := s.ValidatePassword(newPassword, user.Username, user.Email); err != nil {
		return err
	}

	// 哈希新密码
	hashedPassword, err := s.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// 更新密码
	if err := s.daoManager.User.UpdatePassword(userID, hashedPassword); err != nil {
		return fmt.Errorf("重置密码失败: %w", err)
	}

	// 使所有会话失效
	s.invalidateAllUserSessions(userID)

	return nil
}

// ================= 会话管理方法 =================

// ValidateToken 验证JWT令牌
func (s *AuthService) ValidateToken(tokenString string) (*AuthClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*AuthClaims); ok && token.Valid {
		// 检查会话是否仍然有效
		session, err := s.daoManager.UserSession.GetBySessionID(claims.SessionID)
		if err != nil {
			return nil, errors.New("会话已失效")
		}

		// 检查会话是否过期
		if session.ExpiresAt.Before(time.Now()) {
			if err := s.daoManager.UserSession.UpdateIsActive(session, false); err != nil {
				// 记录错误但继续处理
				log.Printf("Failed to update session active status: %v", err)
			}
			return nil, errors.New("会话已过期")
		}

		return claims, nil
	}

	return nil, errors.New("无效的令牌")
}

// LogoutUser 用户登出
func (s *AuthService) LogoutUser(sessionID string) error {
	session, err := s.daoManager.UserSession.GetBySessionID(sessionID)
	if err != nil {
		return err
	}
	return s.daoManager.UserSession.UpdateIsActive(session, false)
}

// ================= 私有辅助方法 =================

// checkUserUniqueness 检查用户名和邮箱的唯一性
func (s *AuthService) checkUserUniqueness(username, email string) error {
	existingUser, err := s.daoManager.User.CheckExists(username, email)
	if err == nil {
		if existingUser.Username == username {
			return ValidationError{
				Field:   "username",
				Message: "用户名已存在",
			}
		}
		return ValidationError{
			Field:   "email",
			Message: "邮箱已存在",
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("检查用户唯一性失败: %w", err)
	}
	return nil
}

// createUserSession 创建用户会话
func (s *AuthService) createUserSession(user *models.User, ipAddress, userAgent string) (string, error) {
	// 清理过期会话
	s.cleanExpiredSessions(user.ID)

	// 检查会话数量限制
	sessionCount, err := s.daoManager.UserSession.CountActiveSessionsByUserID(user.ID)
	if err != nil {
		return "", fmt.Errorf("检查会话数量失败: %w", err)
	}

	if sessionCount >= int64(s.cfg.MaxSessionsPerUser) {
		// 删除最旧的会话
		oldestSession, err := s.daoManager.UserSession.GetOldestSessionByUserID(user.ID)
		if err == nil {
			if err := s.daoManager.UserSession.UpdateIsActive(oldestSession, false); err != nil {
				log.Printf("Failed to update oldest session active status: %v", err)
			}
		}
	}

	// 生成会话ID
	sessionID, err := s.generateSessionID()
	if err != nil {
		return "", fmt.Errorf("生成会话ID失败: %w", err)
	}

	// 创建JWT令牌
	expirationTime := time.Now().Add(time.Duration(s.cfg.SessionExpiryHours) * time.Hour)
	claims := &AuthClaims{
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
		return "", fmt.Errorf("生成令牌失败: %w", err)
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
		return "", fmt.Errorf("保存会话失败: %w", err)
	}

	return tokenString, nil
}

// updateUserLoginInfo 更新用户登录信息
func (s *AuthService) updateUserLoginInfo(user *models.User, ipAddress string) {
	now := time.Now()
	user.LastLoginAt = &now
	user.LastLoginIP = ipAddress
	if err := s.daoManager.User.Update(user); err != nil {
		fmt.Printf("更新用户登录信息失败: %v\n", err)
	}
}

// generateSessionID 生成会话ID
func (s *AuthService) generateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// cleanExpiredSessions 清理过期会话
func (s *AuthService) cleanExpiredSessions(userID uint) {
	if err := s.daoManager.UserSession.CleanExpiredSessions(); err != nil {
		log.Printf("Failed to clean expired sessions: %v", err)
	}
}

// invalidateOtherSessions 使其他会话失效
func (s *AuthService) invalidateOtherSessions(userID uint, currentSessionID string) {
	// 这里可以实现使除当前会话外的所有会话失效
	// 暂时留空，可以根据需要实现
}

// invalidateAllUserSessions 使用户的所有会话失效
func (s *AuthService) invalidateAllUserSessions(userID uint) {
	// 这里可以实现使用户的所有会话失效
	// 暂时留空，可以根据需要实现
}

// NormalizeUsername 规范化用户名
func (s *AuthService) NormalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

// NormalizeEmail 规范化邮箱
func (s *AuthService) NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// ================= 密码验证辅助函数 =================

// containsUppercase 检查是否包含大写字母
func containsUppercase(s string) bool {
	for _, r := range s {
		if unicode.IsUpper(r) {
			return true
		}
	}
	return false
}

// containsLowercase 检查是否包含小写字母
func containsLowercase(s string) bool {
	for _, r := range s {
		if unicode.IsLower(r) {
			return true
		}
	}
	return false
}

// containsNumber 检查是否包含数字
func containsNumber(s string) bool {
	for _, r := range s {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

// containsSpecialChar 检查是否包含特殊字符
func containsSpecialChar(s string) bool {
	for _, r := range s {
		if unicode.IsPunct(r) || unicode.IsSymbol(r) {
			return true
		}
	}
	return false
}

// isCommonPassword 检查是否是常见密码
func isCommonPassword(password string) bool {
	commonPasswords := []string{
		"123456", "password", "12345678", "qwerty", "123456789",
		"12345", "1234", "111111", "1234567", "dragon",
		"123123", "baseball", "abc123", "football", "monkey",
		"letmein", "696969", "shadow", "master", "666666",
		"qwertyuiop", "123321", "mustang", "1234567890",
	}

	lowerPassword := strings.ToLower(password)
	for _, common := range commonPasswords {
		if lowerPassword == common {
			return true
		}
	}
	return false
}

// containsPersonalInfo 检查是否包含个人信息
func containsPersonalInfo(password, username, email string) bool {
	password = strings.ToLower(password)
	username = strings.ToLower(username)
	email = strings.ToLower(email)

	// 检查是否包含用户名
	if len(username) >= 3 && strings.Contains(password, username) {
		return true
	}

	// 检查是否包含邮箱的用户名部分
	if atIndex := strings.Index(email, "@"); atIndex > 0 {
		emailUsername := email[:atIndex]
		if len(emailUsername) >= 3 && strings.Contains(password, emailUsername) {
			return true
		}
	}

	return false
}
