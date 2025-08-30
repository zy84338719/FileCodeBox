package services

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"log"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/dao"
	"github.com/zy84338719/filecodebox/internal/models"
	"github.com/zy84338719/filecodebox/internal/storage"
	"gorm.io/gorm"
)

// ShareService 分享服务
type ShareService struct {
	storage     *storage.StorageManager
	config      *config.Config
	userService *UserService
	daoManager  *dao.DAOManager
}

func NewShareService(db *gorm.DB, storageManager *storage.StorageManager, config *config.Config, userService *UserService) *ShareService {
	return &ShareService{
		storage:     storageManager,
		config:      config,
		userService: userService,
		daoManager:  dao.NewDAOManager(db),
	}
}

// ShareTextRequest 分享文本请求
type ShareTextRequest struct {
	Text        string
	ExpireValue int
	ExpireStyle string
	UserID      *uint  // 用户ID，nil表示匿名上传
	RequireAuth bool   // 是否需要登录才能下载
	ClientIP    string // 客户端IP
}

// ShareFileRequest 分享文件请求
type ShareFileRequest struct {
	File        *multipart.FileHeader
	ExpireValue int
	ExpireStyle string
	UserID      *uint  // 用户ID，nil表示匿名上传
	RequireAuth bool   // 是否需要登录才能下载
	ClientIP    string // 客户端IP
}

// ShareText 分享文本
func (s *ShareService) ShareText(text string, expireValue int, expireStyle string) (*models.FileCode, error) {
	return s.ShareTextWithAuth(ShareTextRequest{
		Text:        text,
		ExpireValue: expireValue,
		ExpireStyle: expireStyle,
		UserID:      nil, // 匿名上传
		RequireAuth: false,
		ClientIP:    "",
	})
}

// ShareTextWithAuth 带认证的分享文本
func (s *ShareService) ShareTextWithAuth(req ShareTextRequest) (*models.FileCode, error) {
	textSize := len([]byte(req.Text))
	maxTextSize := 222 * 1024 // 222KB

	if textSize > maxTextSize {
		return nil, fmt.Errorf("内容过多，建议采用文件形式")
	}

	code := s.generateCode()
	expiredAt, expiredCount, usedCount := s.parseExpireInfo(req.ExpireValue, req.ExpireStyle)

	// 确定上传类型
	uploadType := "anonymous"
	if req.UserID != nil {
		uploadType = "authenticated"
	}

	fileCode := &models.FileCode{
		Code:         code,
		Text:         req.Text,
		ExpiredAt:    expiredAt,
		ExpiredCount: expiredCount,
		UsedCount:    usedCount,
		Size:         int64(textSize),
		UserID:       req.UserID,
		UploadType:   uploadType,
		RequireAuth:  req.RequireAuth,
		OwnerIP:      req.ClientIP,
		Prefix:       "Text",
	}

	if err := s.daoManager.FileCode.Create(fileCode); err != nil {
		return nil, err
	}

	// 如果是认证用户上传，更新用户统计信息
	if req.UserID != nil {
		if err := s.userService.UpdateUserUploadStats(*req.UserID, int64(textSize)); err != nil {
			// 记录错误但不影响上传成功
			log.Printf("Failed to update user upload stats: %v", err)
		}
	}

	return fileCode, nil
}

// ShareFile 分享文件
func (s *ShareService) ShareFile(file *multipart.FileHeader, expireValue int, expireStyle string) (*models.FileCode, error) {
	return s.ShareFileWithAuth(ShareFileRequest{
		File:        file,
		ExpireValue: expireValue,
		ExpireStyle: expireStyle,
		UserID:      nil, // 匿名上传
		RequireAuth: false,
		ClientIP:    "",
	})
}

// ShareFileWithAuth 带认证的分享文件
func (s *ShareService) ShareFileWithAuth(req ShareFileRequest) (*models.FileCode, error) {
	// 确定上传大小限制
	uploadSizeLimit := s.config.UploadSize
	if req.UserID != nil {
		// 用户上传，使用用户专属限制
		uploadSizeLimit = s.config.UserUploadSize
	}

	// 验证文件大小
	if err := storage.ValidateFileSize(req.File, uploadSizeLimit); err != nil {
		return nil, err
	}

	// 生成文件路径信息
	uploadID := s.generateUploadID()
	path, suffix, prefix, uuidFileName := storage.GenerateFileInfo(req.File.Filename, uploadID)

	// 构建完整的保存路径 - 使用绝对路径
	basePath := s.config.StoragePath
	if basePath == "" {
		basePath = filepath.Join(s.config.DataPath)
	}
	savePath := filepath.Join(basePath, path, uuidFileName)

	// 保存文件
	storageInterface := s.storage.GetStorage()
	if err := storageInterface.SaveFile(req.File, savePath); err != nil {
		return nil, fmt.Errorf("保存文件失败: %v", err)
	}

	// 计算文件哈希
	fileHash, err := storage.CalculateFileHash(req.File)
	if err != nil {
		return nil, fmt.Errorf("计算文件哈希失败: %v", err)
	}

	code := s.generateCode()
	expiredAt, expiredCount, usedCount := s.parseExpireInfo(req.ExpireValue, req.ExpireStyle)

	// 确定上传类型
	uploadType := "anonymous"
	if req.UserID != nil {
		uploadType = "authenticated"
	}

	fileCode := &models.FileCode{
		Code:         code,
		Prefix:       prefix,
		Suffix:       suffix,
		UUIDFileName: uuidFileName,
		FilePath:     path,
		Size:         req.File.Size,
		ExpiredAt:    expiredAt,
		ExpiredCount: expiredCount,
		UsedCount:    usedCount,
		FileHash:     fileHash,
		UserID:       req.UserID,
		UploadType:   uploadType,
		RequireAuth:  req.RequireAuth,
		OwnerIP:      req.ClientIP,
	}

	if err := s.daoManager.FileCode.Create(fileCode); err != nil {
		return nil, err
	}

	// 如果是认证用户上传，更新用户统计信息
	if req.UserID != nil {
		if err := s.userService.UpdateUserUploadStats(*req.UserID, req.File.Size); err != nil {
			// 记录错误但不影响上传成功
			log.Printf("Failed to update user upload stats: %v", err)
		}
	}

	return fileCode, nil
}

// GetFileByCode 根据代码获取文件
func (s *ShareService) GetFileByCode(code string, checkExpire bool) (*models.FileCode, error) {
	return s.GetFileByCodeWithAuth(code, checkExpire, nil)
}

// GetFileByCodeWithAuth 根据代码获取文件（带认证检查）
func (s *ShareService) GetFileByCodeWithAuth(code string, checkExpire bool, userID *uint) (*models.FileCode, error) {
	fileCode, err := s.daoManager.FileCode.GetByCode(code)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("文件不存在")
		}
		return nil, err
	}

	if checkExpire && fileCode.IsExpired() {
		return nil, fmt.Errorf("文件已过期")
	}

	// 检查是否需要登录才能访问
	if fileCode.RequireAuth {
		if userID == nil {
			// 如果文件是匿名上传的（user_id为空），但设置了require_auth，
			// 这可能是一个配置错误，我们应该允许访问
			if fileCode.UserID == nil {
				// 匿名上传的文件，即使设置了require_auth也允许访问
				// 这是一个兼容性处理
			} else {
				return nil, fmt.Errorf("该文件需要登录后才能访问")
			}
		}

		// 如果是用户自己上传的文件，允许访问
		// 如果不是，也允许访问（公开的认证文件）
		// 这里可以根据业务需求调整权限控制逻辑
	}

	return fileCode, nil
}

// UpdateFileUsage 更新文件使用次数
func (s *ShareService) UpdateFileUsage(fileCode *models.FileCode) error {
	fileCode.UsedCount++
	if fileCode.ExpiredCount > 0 {
		fileCode.ExpiredCount--
	}
	return s.daoManager.FileCode.Update(fileCode)
}

// generateCode 生成随机代码
func (s *ShareService) generateCode() string {
	bytes := make([]byte, 6)
	if _, err := rand.Read(bytes); err != nil {
		// 如果随机数生成失败，使用时间戳作为备选方案
		return fmt.Sprintf("%x", time.Now().UnixNano())[:12]
	}
	return fmt.Sprintf("%x", bytes)[:12]
}

// generateUploadID 生成上传ID
func (s *ShareService) generateUploadID() string {
	data := fmt.Sprintf("%d", time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)[:32] // 截取前32位
}

// parseExpireInfo 解析过期信息
func (s *ShareService) parseExpireInfo(expireValue int, expireStyle string) (*time.Time, int, int) {
	var expiredAt *time.Time
	expiredCount := 0
	usedCount := 0

	switch expireStyle {
	case "day":
		t := time.Now().Add(time.Duration(expireValue) * 24 * time.Hour)
		expiredAt = &t
		expiredCount = -1
	case "hour":
		t := time.Now().Add(time.Duration(expireValue) * time.Hour)
		expiredAt = &t
		expiredCount = -1
	case "minute":
		t := time.Now().Add(time.Duration(expireValue) * time.Minute)
		expiredAt = &t
		expiredCount = -1
	case "count":
		expiredCount = expireValue
	case "forever":
		// 永不过期
		expiredCount = -1
	default:
		// 默认1天
		t := time.Now().Add(24 * time.Hour)
		expiredAt = &t
		expiredCount = -1
	}

	return expiredAt, expiredCount, usedCount
}

// GetStorageInterface 获取当前存储接口
func (s *ShareService) GetStorageInterface() storage.StorageInterface {
	return s.storage.GetStorage()
}
