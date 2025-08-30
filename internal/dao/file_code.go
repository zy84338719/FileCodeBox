package dao

import (
	"time"

	"github.com/zy84338719/filecodebox/internal/models"
	"gorm.io/gorm"
)

// FileCodeDAO 文件代码数据访问对象
type FileCodeDAO struct {
	db *gorm.DB
}

// NewFileCodeDAO 创建新的文件代码DAO
func NewFileCodeDAO(db *gorm.DB) *FileCodeDAO {
	return &FileCodeDAO{db: db}
}

// Create 创建新的文件记录
func (dao *FileCodeDAO) Create(fileCode *models.FileCode) error {
	return dao.db.Create(fileCode).Error
}

// GetByID 根据ID获取文件记录
func (dao *FileCodeDAO) GetByID(id uint) (*models.FileCode, error) {
	var fileCode models.FileCode
	err := dao.db.First(&fileCode, id).Error
	if err != nil {
		return nil, err
	}
	return &fileCode, nil
}

// GetByCode 根据代码获取文件记录
func (dao *FileCodeDAO) GetByCode(code string) (*models.FileCode, error) {
	var fileCode models.FileCode
	err := dao.db.Where("code = ?", code).First(&fileCode).Error
	if err != nil {
		return nil, err
	}
	return &fileCode, nil
}

// GetByHashAndSize 根据文件哈希和大小获取文件记录
func (dao *FileCodeDAO) GetByHashAndSize(fileHash string, size int64) (*models.FileCode, error) {
	var fileCode models.FileCode
	err := dao.db.Where("file_hash = ? AND size = ? AND deleted_at IS NULL", fileHash, size).First(&fileCode).Error
	if err != nil {
		return nil, err
	}
	return &fileCode, nil
}

// Update 更新文件记录
func (dao *FileCodeDAO) Update(fileCode *models.FileCode) error {
	return dao.db.Save(fileCode).Error
}

// UpdateColumns 更新指定字段
func (dao *FileCodeDAO) UpdateColumns(id uint, updates map[string]interface{}) error {
	return dao.db.Model(&models.FileCode{}).Where("id = ?", id).Updates(updates).Error
}

// Delete 删除文件记录
func (dao *FileCodeDAO) Delete(id uint) error {
	return dao.db.Delete(&models.FileCode{}, id).Error
}

// DeleteByFileCode 删除文件记录
func (dao *FileCodeDAO) DeleteByFileCode(fileCode *models.FileCode) error {
	return dao.db.Delete(fileCode).Error
}

// Count 统计文件总数
func (dao *FileCodeDAO) Count() (int64, error) {
	var count int64
	err := dao.db.Model(&models.FileCode{}).Count(&count).Error
	return count, err
}

// CountToday 统计今天的文件数量
func (dao *FileCodeDAO) CountToday() (int64, error) {
	var count int64
	today := time.Now().Format("2006-01-02")
	err := dao.db.Unscoped().Model(&models.FileCode{}).Where("created_at >= ?", today).Count(&count).Error
	return count, err
}

// CountActive 统计活跃文件数量
func (dao *FileCodeDAO) CountActive() (int64, error) {
	var count int64
	err := dao.db.Model(&models.FileCode{}).
		Where("expired_at IS NULL OR expired_at > ? OR expired_count > 0", time.Now()).
		Count(&count).Error
	return count, err
}

// GetTotalSize 获取文件总大小
func (dao *FileCodeDAO) GetTotalSize() (int64, error) {
	var totalSize int64
	err := dao.db.Model(&models.FileCode{}).Select("COALESCE(SUM(size), 0)").Scan(&totalSize).Error
	return totalSize, err
}

// List 分页获取文件列表
func (dao *FileCodeDAO) List(page, pageSize int, search string) ([]models.FileCode, int64, error) {
	var files []models.FileCode
	var total int64

	query := dao.db.Model(&models.FileCode{})

	// 搜索条件
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("code LIKE ? OR prefix LIKE ? OR suffix LIKE ?",
			searchPattern, searchPattern, searchPattern)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&files).Error

	return files, total, err
}

// GetExpiredFiles 获取过期文件
func (dao *FileCodeDAO) GetExpiredFiles() ([]models.FileCode, error) {
	var expiredFiles []models.FileCode
	err := dao.db.Where("(expired_at IS NOT NULL AND expired_at < ?) OR expired_count = 0", time.Now()).
		Find(&expiredFiles).Error
	return expiredFiles, err
}

// CheckCodeExists 检查代码是否存在（排除指定ID）
func (dao *FileCodeDAO) CheckCodeExists(code string, excludeID uint) (bool, error) {
	var existingFile models.FileCode
	err := dao.db.Where("code = ? AND id != ?", code, excludeID).First(&existingFile).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetByHash 根据文件哈希获取文件记录
func (dao *FileCodeDAO) GetByHash(fileHash string, fileSize int64) (*models.FileCode, error) {
	var existingFile models.FileCode
	err := dao.db.Where("file_hash = ? AND size = ? AND deleted_at IS NULL", fileHash, fileSize).
		First(&existingFile).Error
	if err != nil {
		return nil, err
	}
	return &existingFile, nil
}

// CountByUserID 统计用户上传的文件数量
func (dao *FileCodeDAO) CountByUserID(userID uint) (int64, error) {
	var count int64
	err := dao.db.Model(&models.FileCode{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

// GetTotalSizeByUserID 获取用户文件总大小
func (dao *FileCodeDAO) GetTotalSizeByUserID(userID uint) (int64, error) {
	var totalSize int64
	err := dao.db.Model(&models.FileCode{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(size), 0)").
		Scan(&totalSize).Error
	return totalSize, err
}

// GetByUserID 获取用户的文件记录
func (dao *FileCodeDAO) GetByUserID(userID uint, fileID uint) (*models.FileCode, error) {
	var fileCode models.FileCode
	err := dao.db.Where("id = ? AND user_id = ?", fileID, userID).First(&fileCode).Error
	if err != nil {
		return nil, err
	}
	return &fileCode, nil
}

// GetFilesByUserID 获取用户的所有文件
func (dao *FileCodeDAO) GetFilesByUserID(userID uint) ([]models.FileCode, error) {
	var files []models.FileCode
	err := dao.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&files).Error
	return files, err
}

// DeleteByUserID 删除用户的所有文件
func (dao *FileCodeDAO) DeleteByUserID(tx *gorm.DB, userID uint) error {
	return tx.Where("user_id = ?", userID).Delete(&models.FileCode{}).Error
}

// CountTodayUploads 统计今天的上传数量
func (dao *FileCodeDAO) CountTodayUploads() (int64, error) {
	var count int64
	today := time.Now().Format("2006-01-02")
	err := dao.db.Model(&models.FileCode{}).Where("created_at >= ?", today).Count(&count).Error
	return count, err
}
