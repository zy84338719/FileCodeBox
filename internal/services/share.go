package services

import (
	"crypto/md5"
	"crypto/rand"
	"filecodebox/internal/config"
	"filecodebox/internal/models"
	"filecodebox/internal/storage"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"time"

	"gorm.io/gorm"
)

// ShareService 分享服务
type ShareService struct {
	db      *gorm.DB
	storage *storage.StorageManager
	config  *config.Config
}

func NewShareService(db *gorm.DB, storageManager *storage.StorageManager, config *config.Config) *ShareService {
	return &ShareService{
		db:      db,
		storage: storageManager,
		config:  config,
	}
}

// ShareText 分享文本
func (s *ShareService) ShareText(text string, expireValue int, expireStyle string) (*models.FileCode, error) {
	textSize := len([]byte(text))
	maxTextSize := 222 * 1024 // 222KB

	if textSize > maxTextSize {
		return nil, fmt.Errorf("内容过多，建议采用文件形式")
	}

	code := s.generateCode()
	expiredAt, expiredCount, usedCount := s.parseExpireInfo(expireValue, expireStyle)

	fileCode := &models.FileCode{
		Code:         code,
		Text:         text,
		ExpiredAt:    expiredAt,
		ExpiredCount: expiredCount,
		UsedCount:    usedCount,
		Size:         int64(textSize),
		Prefix:       "Text",
	}

	if err := s.db.Create(fileCode).Error; err != nil {
		return nil, err
	}

	return fileCode, nil
}

// ShareFile 分享文件
func (s *ShareService) ShareFile(file *multipart.FileHeader, expireValue int, expireStyle string) (*models.FileCode, error) {
	// 验证文件大小
	if err := storage.ValidateFileSize(file, s.config.UploadSize); err != nil {
		return nil, err
	}

	// 生成文件路径信息
	uploadID := s.generateUploadID()
	path, suffix, prefix, uuidFileName := storage.GenerateFileInfo(file.Filename, uploadID)

	// 构建完整的保存路径 - 使用绝对路径
	basePath := s.config.StoragePath
	if basePath == "" {
		basePath = filepath.Join(s.config.DataPath, "share", "data")
	}
	savePath := filepath.Join(basePath, path, uuidFileName)

	// 保存文件
	storageInterface := s.storage.GetStorage()
	if err := storageInterface.SaveFile(file, savePath); err != nil {
		return nil, fmt.Errorf("保存文件失败: %v", err)
	}

	// 计算文件哈希
	fileHash, err := storage.CalculateFileHash(file)
	if err != nil {
		return nil, fmt.Errorf("计算文件哈希失败: %v", err)
	}

	code := s.generateCode()
	expiredAt, expiredCount, usedCount := s.parseExpireInfo(expireValue, expireStyle)

	fileCode := &models.FileCode{
		Code:         code,
		Prefix:       prefix,
		Suffix:       suffix,
		UUIDFileName: uuidFileName,
		FilePath:     path,
		Size:         file.Size,
		ExpiredAt:    expiredAt,
		ExpiredCount: expiredCount,
		UsedCount:    usedCount,
		FileHash:     fileHash,
	}

	if err := s.db.Create(fileCode).Error; err != nil {
		return nil, err
	}

	return fileCode, nil
}

// GetFileByCode 根据代码获取文件
func (s *ShareService) GetFileByCode(code string, checkExpire bool) (*models.FileCode, error) {
	var fileCode models.FileCode
	if err := s.db.Where("code = ?", code).First(&fileCode).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("文件不存在")
		}
		return nil, err
	}

	if checkExpire && fileCode.IsExpired() {
		return nil, fmt.Errorf("文件已过期")
	}

	return &fileCode, nil
}

// UpdateFileUsage 更新文件使用次数
func (s *ShareService) UpdateFileUsage(fileCode *models.FileCode) error {
	fileCode.UsedCount++
	if fileCode.ExpiredCount > 0 {
		fileCode.ExpiredCount--
	}
	return s.db.Save(fileCode).Error
}

// generateCode 生成随机代码
func (s *ShareService) generateCode() string {
	bytes := make([]byte, 6)
	rand.Read(bytes)
	return fmt.Sprintf("%x", bytes)[:12]
}

// generateUploadID 生成上传ID
func (s *ShareService) generateUploadID() string {
	data := fmt.Sprintf("%d", time.Now().UnixNano())
	return fmt.Sprintf("%x", md5.Sum([]byte(data)))
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
