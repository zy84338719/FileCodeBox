package user

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/zy84338719/fileCodeBox/internal/pkg/auth"
	"github.com/zy84338719/fileCodeBox/internal/repo/db/dao"
	"github.com/zy84338719/fileCodeBox/internal/repo/db/model"
	usermodel "github.com/zy84338719/fileCodeBox/biz/model/user"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type CreateUserReq struct {
	Username string
	Email    string
	Password string
	Nickname string
}

type UpdateUserReq struct {
	Nickname string
	Avatar   string
	Role     string
	Status   string
}

type Service struct {
	repo *dao.UserRepository
	apiKeyRepo *dao.UserAPIKeyRepository
}

func NewService() *Service {
	// 延迟初始化 repository，确保数据库已经准备好
	return &Service{
		repo: nil,        // 延迟初始化
		apiKeyRepo: nil,  // 延迟初始化
	}
}

// ensureRepository 确保repository已初始化
func (s *Service) ensureRepository() {
	if s.repo == nil {
		s.repo = dao.NewUserRepository()
	}
	if s.apiKeyRepo == nil {
		s.apiKeyRepo = dao.NewUserAPIKeyRepository()
	}
}

func (s *Service) Create(ctx context.Context, req *CreateUserReq) (*model.UserResp, error) {
	s.ensureRepository()
	s.ensureRepository()

	existing, err := s.repo.GetByUsername(ctx, req.Username)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("username already exists")
	}

	existing, err = s.repo.GetByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Nickname:     req.Nickname,
		Status:       "active",
		Role:         "user",
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user.ToResp(), nil
}

func (s *Service) GetByID(ctx context.Context, id uint) (*model.UserResp, error) {
	s.ensureRepository()
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user.ToResp(), nil
}

func (s *Service) Update(ctx context.Context, id uint, req *UpdateUserReq) (*model.UserResp, error) {
	s.ensureRepository()
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	if req.Role != "" {
		user.Role = req.Role
	}
	if req.Status != "" {
		user.Status = req.Status
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user.ToResp(), nil
}

func (s *Service) Delete(ctx context.Context, id uint) error {
	s.ensureRepository()
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context, page, pageSize int) ([]*model.UserResp, int64, error) {
	s.ensureRepository()
	users, total, err := s.repo.List(ctx, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	resps := make([]*model.UserResp, len(users))
	for i, user := range users {
		resps[i] = user.ToResp()
	}

	return resps, total, nil
}

// Login 用户登录
func (s *Service) Login(ctx context.Context, username, password string) (*model.UserResp, string, error) {
	s.ensureRepository()
	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", errors.New("用户名或密码错误")
		}
		return nil, "", err
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", errors.New("用户名或密码错误")
	}

	// 检查用户状态
	if user.Status != "active" {
		return nil, "", errors.New("用户账号已被禁用")
	}

	// 生成 JWT Token
	token, err := auth.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, "", errors.New("生成令牌失败")
	}

	return user.ToResp(), token, nil
}

// LoginByEmail 通过邮箱登录
func (s *Service) LoginByEmail(ctx context.Context, email, password string) (*model.UserResp, string, error) {
	s.ensureRepository()
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", errors.New("邮箱或密码错误")
		}
		return nil, "", err
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", errors.New("邮箱或密码错误")
	}

	// 检查用户状态
	if user.Status != "active" {
		return nil, "", errors.New("用户账号已被禁用")
	}

	// 生成 JWT Token
	token, err := auth.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, "", errors.New("生成令牌失败")
	}

	return user.ToResp(), token, nil
}

// ChangePassword 修改密码
func (s *Service) ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error {
	s.ensureRepository()
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("旧密码错误")
	}

	// 生成新密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hashedPassword)
	return s.repo.Update(ctx, user)
}

// UpdateUserStats 更新用户统计信息
func (s *Service) UpdateUserStats(userID uint, statsType string, value int64) error {
	s.ensureRepository()
	ctx := context.Background()
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	switch statsType {
	case "upload", "uploads":
		user.TotalUploads += int(value)
	case "download", "downloads":
		user.TotalDownloads += int(value)
	case "storage":
		user.TotalStorage += value
	}

	return s.repo.Update(ctx, user)
}

// GetStats 获取用户统计信息
func (s *Service) GetStats(ctx context.Context, userID uint) (*model.UserStats, error) {
	s.ensureRepository()
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &model.UserStats{
		UserID:         user.ID,
		TotalUploads:   user.TotalUploads,
		TotalDownloads: user.TotalDownloads,
		TotalStorage:   user.TotalStorage,
		FileCount:      0, // TODO: 从 FileCode 表统计
	}, nil
}

// ==================== API Key 相关常量和类型 ====================

const (
	apiKeyPrefix = "fcb_sk_"
	maxUserAPIKeys = 5  // 每个用户最多保留的有效密钥数量
)

// CreateAPIKeyReq 创建 API Key 请求
type CreateAPIKeyReq struct {
	Name         string
	ExpiresAt    *time.Time
	ExpiresInDays *int64
}

// CreateAPIKeyResp 创建 API Key 响应
type CreateAPIKeyResp struct {
	Key    string
	APIKey *APIKeyData
}

// APIKeyData API Key 数据
type APIKeyData struct {
	ID         uint
	Name       string
	Prefix     string
	LastUsedAt *time.Time
	ExpiresAt  *time.Time
	CreatedAt  *time.Time
	Revoked    bool
}

// ==================== API Key 方法 ====================

// GenerateRandomKey 生成32位随机字符串
func GenerateRandomKey() (string, error) {
	bytes := make([]byte, 16) // 16字节 = 32位十六进制字符
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// HashAPIKey 计算密钥的 SHA256 哈希
func HashAPIKey(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}

// GetKeyPrefix 获取密钥的前缀（用于展示）
func GetKeyPrefix(key string) string {
	if len(key) <= 12 {
		return key
	}
	return key[:12]
}

// CreateAPIKey 为用户创建新的 API Key
func (s *Service) CreateAPIKey(ctx context.Context, userID uint, req *CreateAPIKeyReq) (*CreateAPIKeyResp, error) {
	s.ensureRepository()

	// 限制有效密钥数量
	count, err := s.apiKeyRepo.CountActiveByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("统计用户 API Key 失败: %w", err)
	}
	if count >= maxUserAPIKeys {
		return nil, fmt.Errorf("最多只能保留 %d 个有效 API Key，请先删除旧的密钥", maxUserAPIKeys)
	}

	// 处理过期时间
	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		if req.ExpiresAt.After(time.Now().Add(-1 * time.Minute)) {
			t := req.ExpiresAt.UTC()
			expiresAt = &t
		}
	} else if req.ExpiresInDays != nil && *req.ExpiresInDays > 0 {
		t := time.Now().UTC().Add(time.Duration(*req.ExpiresInDays) * 24 * time.Hour)
		expiresAt = &t
	}

	// 生成随机密钥
	randomPart, err := GenerateRandomKey()
	if err != nil {
		return nil, fmt.Errorf("生成随机密钥失败: %w", err)
	}
	plainKey := apiKeyPrefix + randomPart
	keyHash := HashAPIKey(plainKey)
	keyPrefix := GetKeyPrefix(plainKey)

	// 归一化名称
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = "API Key"
	}
	if len(name) > 100 {
		name = name[:100]
	}

	// 创建记录
	record := &model.UserAPIKey{
		UserID:    userID,
		Name:      name,
		Prefix:    keyPrefix,
		KeyHash:   keyHash,
		ExpiresAt: expiresAt,
		Revoked:   false,
	}

	if err := s.apiKeyRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("保存 API Key 失败: %w", err)
	}

	return &CreateAPIKeyResp{
		Key: plainKey,
		APIKey: &APIKeyData{
			ID:         record.ID,
			Name:       record.Name,
			Prefix:     record.Prefix,
			LastUsedAt: record.LastUsedAt,
			ExpiresAt:  record.ExpiresAt,
			CreatedAt:  &record.CreatedAt,
			Revoked:    record.Revoked,
		},
	}, nil
}

// ListAPIKeys 获取用户的全部 API Key 列表（包含已撤销）
func (s *Service) ListAPIKeys(ctx context.Context, userID uint) ([]*APIKeyData, error) {
	s.ensureRepository()

	keys, err := s.apiKeyRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]*APIKeyData, len(keys))
	for i, key := range keys {
		result[i] = &APIKeyData{
			ID:         key.ID,
			Name:       key.Name,
			Prefix:     key.Prefix,
			LastUsedAt: key.LastUsedAt,
			ExpiresAt:  key.ExpiresAt,
			CreatedAt:  &key.CreatedAt,
			Revoked:    key.Revoked,
		}
	}

	return result, nil
}

// DeleteAPIKey 撤销指定的 API Key
func (s *Service) DeleteAPIKey(ctx context.Context, userID, id uint) error {
	s.ensureRepository()

	err := s.apiKeyRepo.RevokeByID(ctx, userID, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("API Key 不存在或已撤销")
	}
	if err != nil {
		return fmt.Errorf("撤销 API Key 失败: %w", err)
	}
	return nil
}

// ==================== 用户文件列表方法 ====================

// GetUserFiles 获取用户文件列表（支持分页）
func (s *Service) GetUserFiles(ctx context.Context, userID uint, page, pageSize int) (*usermodel.UserFileList, error) {
	fileCodeRepo := dao.NewFileCodeRepository()

	// 获取文件列表
	files, total, err := fileCodeRepo.GetFilesByUserIDWithPagination(ctx, userID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("获取用户文件列表失败: %w", err)
	}

	// 转换为响应格式
	fileItems := make([]*usermodel.UserFileItem, len(files))
	for i, file := range files {
		// 提取文件名（从 UUIDFileName 或 FilePath 中获取）
		fileName := file.UUIDFileName
		if fileName == "" && file.FilePath != "" {
			// 从路径中提取文件名
			parts := strings.Split(file.FilePath, "/")
			if len(parts) > 0 {
				fileName = parts[len(parts)-1]
			}
		}

		fileItems[i] = &usermodel.UserFileItem{
			Id:          uint32(file.ID),
			Code:        file.Code,
			Prefix:      file.Prefix,
			Suffix:      file.Suffix,
			FileName:    fileName,
			FilePath:    file.FilePath,
			Size:        file.Size,
			ExpiredCount: int32(file.ExpiredCount),
			UsedCount:   int32(file.UsedCount),
			CreatedAt:   file.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   file.UpdatedAt.Format("2006-01-02 15:04:05"),
		}

		// 格式化过期时间
		if file.ExpiredAt != nil {
			fileItems[i].ExpiredAt = file.ExpiredAt.Format("2006-01-02 15:04:05")
		}
	}

	// 计算分页信息
	totalPages := int32((total + int64(pageSize) - 1) / int64(pageSize))
	if totalPages < 1 {
		totalPages = 1
	}

	return &usermodel.UserFileList{
		Files: fileItems,
		Pagination: &usermodel.UserFilePagination{
			Page:       int32(page),
			PageSize:   int32(pageSize),
			Total:      total,
			TotalPages: totalPages,
			HasNext:    int32(page) < totalPages,
			HasPrev:    int32(page) > 1,
		},
	}, nil
}
