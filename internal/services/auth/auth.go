package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zy84338719/filecodebox/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

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

// ValidationError 验证错误
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// HashPassword 加密密码
func (s *Service) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("密码哈希失败: %w", err)
	}
	return string(hashedPassword), nil
}

// VerifyPassword 验证密码是否正确
func (s *Service) VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// CheckPassword 验证密码
func (s *Service) CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GetPasswordPolicy 获取当前密码策略
func (s *Service) GetPasswordPolicy() PasswordPolicy {
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
func (s *Service) ValidatePassword(password string, username string, email string) error {
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

// ValidateUsername 验证用户名
func (s *Service) ValidateUsername(username string) error {
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
func (s *Service) ValidateEmail(email string) error {
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

// NormalizeUsername 规范化用户名
func (s *Service) NormalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

// NormalizeEmail 规范化邮箱
func (s *Service) NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// Login 用户登录
func (s *Service) Login(username, password string) (*models.User, error) {
	user, err := s.repositoryManager.User.GetByUsername(username)
	if err != nil {
		return nil, err
	}

	if !s.CheckPassword(password, user.PasswordHash) {
		return nil, errors.New("invalid password")
	}

	return user, nil
}

// Register 用户注册
func (s *Service) Register(username, password, email string) (*models.User, error) {
	// 检查用户名是否已存在
	if _, err := s.repositoryManager.User.GetByUsername(username); err == nil {
		return nil, errors.New("username already exists")
	}

	// 加密密码
	hashedPassword, err := s.HashPassword(password)
	if err != nil {
		return nil, err
	}

	// 创建用户
	user := &models.User{
		Username:     username,
		PasswordHash: hashedPassword,
		Email:        email,
	}

	err = s.repositoryManager.User.Create(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GenerateRandomToken 生成随机令牌
func (s *Service) GenerateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// ChangePassword 修改密码
func (s *Service) ChangePassword(userID uint, oldPassword, newPassword string) error {
	user, err := s.repositoryManager.User.GetByID(userID)
	if err != nil {
		return err
	}

	if !s.CheckPassword(oldPassword, user.PasswordHash) {
		return errors.New("old password is incorrect")
	}

	hashedPassword, err := s.HashPassword(newPassword)
	if err != nil {
		return err
	}

	return s.repositoryManager.User.UpdatePassword(userID, hashedPassword)
}

// CreateUserSession 创建用户会话
func (s *Service) CreateUserSession(user *models.User, ipAddress, userAgent string) (string, error) {
	// 生成会话ID
	sessionID, err := s.GenerateRandomToken(32)
	if err != nil {
		return "", err
	}

	// 创建会话记录
	session := &models.UserSession{
		UserID:    user.ID,
		SessionID: sessionID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		ExpiresAt: time.Now().Add(time.Hour * time.Duration(s.manager.User.SessionExpiryHours)),
		IsActive:  true,
	}

	err = s.repositoryManager.UserSession.Create(session)
	if err != nil {
		return "", err
	}

	// 生成JWT token
	claims := AuthClaims{
		UserID:    user.ID,
		Username:  user.Username,
		Role:      user.Role,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(session.ExpiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.manager.User.JWTSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken 验证JWT令牌
func (s *Service) ValidateToken(tokenString string) (*AuthClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.manager.User.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*AuthClaims); ok && token.Valid {
		// 验证会话是否仍然有效
		session, err := s.repositoryManager.UserSession.GetBySessionID(claims.SessionID)
		if err != nil || !session.IsActive || session.ExpiresAt.Before(time.Now()) {
			return nil, errors.New("session expired or invalid")
		}

		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// LogoutUser 用户登出
func (s *Service) LogoutUser(sessionID string) error {
	// 简化实现，目前只返回nil
	// 实际应该删除或标记会话为无效
	return nil
}

// CheckUserExists 检查用户是否存在
func (s *Service) CheckUserExists(username, email string) error {
	existingUser, err := s.repositoryManager.User.CheckExists(username, email)
	if err == nil {
		if existingUser.Username == username {
			return errors.New("用户名已存在")
		}
		return errors.New("邮箱已存在")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("检查用户唯一性失败: %w", err)
	}
	return nil
}

// 辅助函数
func containsUppercase(s string) bool {
	for _, r := range s {
		if unicode.IsUpper(r) {
			return true
		}
	}
	return false
}

func containsLowercase(s string) bool {
	for _, r := range s {
		if unicode.IsLower(r) {
			return true
		}
	}
	return false
}

func containsNumber(s string) bool {
	for _, r := range s {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func containsSpecialChar(s string) bool {
	for _, r := range s {
		if unicode.IsPunct(r) || unicode.IsSymbol(r) {
			return true
		}
	}
	return false
}

func isCommonPassword(password string) bool {
	commonPasswords := []string{
		"123456", "password", "123456789", "12345678", "12345",
		"1234567", "1234567890", "qwerty", "abc123", "111111",
	}

	lowerPassword := strings.ToLower(password)
	for _, common := range commonPasswords {
		if lowerPassword == common {
			return true
		}
	}
	return false
}

func containsPersonalInfo(password, username, email string) bool {
	lowerPassword := strings.ToLower(password)
	lowerUsername := strings.ToLower(username)

	if strings.Contains(lowerPassword, lowerUsername) {
		return true
	}

	if email != "" {
		emailParts := strings.Split(email, "@")
		if len(emailParts) > 0 {
			emailUsername := strings.ToLower(emailParts[0])
			if strings.Contains(lowerPassword, emailUsername) {
				return true
			}
		}
	}

	return false
}
