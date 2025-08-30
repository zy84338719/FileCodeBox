package services

import (
	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/models"
	"github.com/zy84338719/filecodebox/internal/storage"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// AdminService 管理服务
type AdminService struct {
	db             *gorm.DB
	config         *config.Config
	storageManager *storage.StorageManager
}

func NewAdminService(db *gorm.DB, config *config.Config, storageManager *storage.StorageManager) *AdminService {
	return &AdminService{
		db:             db,
		config:         config,
		storageManager: storageManager,
	}
}

// GetStats 获取统计信息
func (s *AdminService) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总文件数（不包括已删除的）
	var totalFiles int64
	s.db.Model(&models.FileCode{}).Count(&totalFiles)
	stats["total_files"] = totalFiles

	// 今日上传数（包括已删除的，因为这是历史统计）
	today := time.Now().Truncate(24 * time.Hour)
	var todayFiles int64
	s.db.Unscoped().Model(&models.FileCode{}).Where("created_at >= ?", today).Count(&todayFiles)
	stats["today_files"] = todayFiles

	// 活跃文件数（未过期且未删除）
	var activeFiles int64
	s.db.Model(&models.FileCode{}).Where("expired_at IS NULL OR expired_at > ? OR expired_count > 0", time.Now()).Count(&activeFiles)
	stats["active_files"] = activeFiles

	// 总大小（不包括已删除的）
	var totalSize int64
	s.db.Model(&models.FileCode{}).Select("COALESCE(SUM(size), 0)").Scan(&totalSize)
	stats["total_size"] = totalSize

	// 系统启动时间
	var sysStart models.KeyValue
	if err := s.db.Where("key = ?", "sys_start").First(&sysStart).Error; err == nil {
		stats["sys_start"] = sysStart.Value
	} else {
		// 如果没有记录，创建一个
		startTime := fmt.Sprintf("%d", time.Now().UnixMilli())
		s.db.Create(&models.KeyValue{
			Key:   "sys_start",
			Value: startTime,
		})
		stats["sys_start"] = startTime
	}

	return stats, nil
}

// GetFiles 获取文件列表
func (s *AdminService) GetFiles(page, pageSize int, search string) ([]models.FileCode, int64, error) {
	var files []models.FileCode
	var total int64

	query := s.db.Model(&models.FileCode{})

	if search != "" {
		query = query.Where("code LIKE ? OR prefix LIKE ? OR suffix LIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// 获取总数
	query.Count(&total)

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&files).Error

	return files, total, err
}

// DeleteFile 删除文件
func (s *AdminService) DeleteFile(id uint) error {
	var fileCode models.FileCode
	if err := s.db.First(&fileCode, id).Error; err != nil {
		return err
	}

	// 删除实际文件
	storageInterface := s.storageManager.GetStorage()
	if err := storageInterface.DeleteFile(&fileCode); err != nil {
		// 记录错误，但不阻止数据库删除
		fmt.Printf("Warning: Failed to delete physical file: %v\n", err)
	}

	return s.db.Delete(&fileCode).Error
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
	var expiredFiles []models.FileCode

	// 查找过期文件
	err := s.db.Where("(expired_at IS NOT NULL AND expired_at < ?) OR expired_count = 0", time.Now()).Find(&expiredFiles).Error
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
		s.db.Delete(&file)
	}

	return count, nil
}

// UpdateFile 更新文件信息
func (s *AdminService) UpdateFile(id uint, code, prefix, suffix string, expiredAt *time.Time, expiredCount *int) error {
	var fileCode models.FileCode
	if err := s.db.First(&fileCode, id).Error; err != nil {
		return err
	}

	updates := make(map[string]interface{})

	if code != "" && code != fileCode.Code {
		// 检查代码是否已存在
		var existingFile models.FileCode
		if err := s.db.Where("code = ? AND id != ?", code, id).First(&existingFile).Error; err == nil {
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
		return s.db.Model(&fileCode).Updates(updates).Error
	}

	return nil
}

// GetFileByID 根据ID获取文件
func (s *AdminService) GetFileByID(id uint) (*models.FileCode, error) {
	var fileCode models.FileCode
	err := s.db.First(&fileCode, id).Error
	if err != nil {
		return nil, err
	}
	return &fileCode, nil
}
