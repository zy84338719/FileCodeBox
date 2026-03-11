package dao

import (
	"context"
	"time"

	"github.com/zy84338719/fileCodeBox/internal/repo/db"
	"github.com/zy84338719/fileCodeBox/internal/repo/db/model"
	"gorm.io/gorm"
)

type FileCodeRepository struct {
}

func NewFileCodeRepository() *FileCodeRepository {
	return &FileCodeRepository{}
}

func (r *FileCodeRepository) db() *gorm.DB {
	return db.GetDB()
}

func (r *FileCodeRepository) Create(ctx context.Context, fileCode *model.FileCode) error {
	return r.db().WithContext(ctx).Create(fileCode).Error
}

func (r *FileCodeRepository) GetByID(ctx context.Context, id uint) (*model.FileCode, error) {
	var fileCode model.FileCode
	err := r.db().WithContext(ctx).First(&fileCode, id).Error
	if err != nil {
		return nil, err
	}
	return &fileCode, nil
}

func (r *FileCodeRepository) GetByCode(ctx context.Context, code string) (*model.FileCode, error) {
	var fileCode model.FileCode
	err := r.db().WithContext(ctx).Where("code = ?", code).First(&fileCode).Error
	if err != nil {
		return nil, err
	}
	return &fileCode, nil
}

func (r *FileCodeRepository) GetByHashAndSize(ctx context.Context, fileHash string, size int64) (*model.FileCode, error) {
	var fileCode model.FileCode
	err := r.db().WithContext(ctx).Where("file_hash = ? AND size = ? AND deleted_at IS NULL", fileHash, size).First(&fileCode).Error
	if err != nil {
		return nil, err
	}
	return &fileCode, nil
}

func (r *FileCodeRepository) Update(ctx context.Context, fileCode *model.FileCode) error {
	return r.db().WithContext(ctx).Save(fileCode).Error
}

func (r *FileCodeRepository) UpdateColumns(ctx context.Context, id uint, updates map[string]interface{}) error {
	return r.db().WithContext(ctx).Model(&model.FileCode{}).Where("id = ?", id).Updates(updates).Error
}

func (r *FileCodeRepository) Delete(ctx context.Context, id uint) error {
	return r.db().WithContext(ctx).Delete(&model.FileCode{}, id).Error
}

func (r *FileCodeRepository) DeleteByFileCode(ctx context.Context, fileCode *model.FileCode) error {
	return r.db().WithContext(ctx).Delete(fileCode).Error
}

func (r *FileCodeRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db().WithContext(ctx).Model(&model.FileCode{}).Count(&count).Error
	return count, err
}

func (r *FileCodeRepository) CountToday(ctx context.Context) (int64, error) {
	var count int64
	today := time.Now().Format("2006-01-02")
	err := r.db().WithContext(ctx).Unscoped().Model(&model.FileCode{}).Where("created_at >= ?", today).Count(&count).Error
	return count, err
}

func (r *FileCodeRepository) CountActive(ctx context.Context) (int64, error) {
	var count int64
	err := r.db().WithContext(ctx).Model(&model.FileCode{}).
		Where("expired_at IS NULL OR expired_at > ? OR expired_count > 0", time.Now()).
		Count(&count).Error
	return count, err
}

func (r *FileCodeRepository) GetTotalSize(ctx context.Context) (int64, error) {
	var totalSize int64
	err := r.db().WithContext(ctx).Model(&model.FileCode{}).Select("COALESCE(SUM(size), 0)").Scan(&totalSize).Error
	return totalSize, err
}

func (r *FileCodeRepository) List(ctx context.Context, page, pageSize int, search string) ([]*model.FileCode, int64, error) {
	var files []*model.FileCode
	var total int64

	query := r.db().WithContext(ctx).Model(&model.FileCode{})

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

func (r *FileCodeRepository) GetExpiredFiles(ctx context.Context) ([]*model.FileCode, error) {
	var expiredFiles []*model.FileCode
	err := r.db().WithContext(ctx).Where("(expired_at IS NOT NULL AND expired_at < ?) OR expired_count = 0", time.Now()).
		Find(&expiredFiles).Error
	return expiredFiles, err
}

func (r *FileCodeRepository) DeleteExpiredFiles(ctx context.Context, expiredFiles []*model.FileCode) (int, error) {
	if len(expiredFiles) == 0 {
		return 0, nil
	}

	count := 0
	for _, file := range expiredFiles {
		if err := r.db().WithContext(ctx).Delete(file).Error; err != nil {
			continue // 记录错误但继续处理其他文件
		}
		count++
	}
	return count, nil
}

func (r *FileCodeRepository) CheckCodeExists(ctx context.Context, code string, excludeID uint) (bool, error) {
	var existingFile model.FileCode
	err := r.db().WithContext(ctx).Where("code = ? AND id != ?", code, excludeID).First(&existingFile).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *FileCodeRepository) GetByHash(ctx context.Context, fileHash string, fileSize int64) (*model.FileCode, error) {
	var existingFile model.FileCode
	err := r.db().WithContext(ctx).Where("file_hash = ? AND size = ? AND deleted_at IS NULL", fileHash, fileSize).
		First(&existingFile).Error
	if err != nil {
		return nil, err
	}
	return &existingFile, nil
}

func (r *FileCodeRepository) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db().WithContext(ctx).Model(&model.FileCode{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

func (r *FileCodeRepository) GetTotalSizeByUserID(ctx context.Context, userID uint) (int64, error) {
	var totalSize int64
	err := r.db().WithContext(ctx).Model(&model.FileCode{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(size), 0)").
		Scan(&totalSize).Error
	return totalSize, err
}

func (r *FileCodeRepository) GetByUserID(ctx context.Context, userID uint, fileID uint) (*model.FileCode, error) {
	var fileCode model.FileCode
	err := r.db().WithContext(ctx).Where("id = ? AND user_id = ?", fileID, userID).First(&fileCode).Error
	if err != nil {
		return nil, err
	}
	return &fileCode, nil
}

func (r *FileCodeRepository) GetFilesByUserID(ctx context.Context, userID uint) ([]*model.FileCode, error) {
	var files []*model.FileCode
	err := r.db().WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&files).Error
	return files, err
}

func (r *FileCodeRepository) GetFilesByUserIDWithPagination(ctx context.Context, userID uint, page, pageSize int) ([]*model.FileCode, int64, error) {
	var files []*model.FileCode
	var total int64

	// 构建查询条件
	query := r.db().WithContext(ctx).Model(&model.FileCode{}).Where("user_id = ?", userID)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&files).Error

	return files, total, err
}

func (r *FileCodeRepository) DeleteByUserID(ctx context.Context, userID uint) error {
	return r.db().WithContext(ctx).Where("user_id = ?", userID).Delete(&model.FileCode{}).Error
}

func (r *FileCodeRepository) CountTodayUploads(ctx context.Context) (int64, error) {
	var count int64
	today := time.Now().Format("2006-01-02")
	err := r.db().WithContext(ctx).Model(&model.FileCode{}).Where("created_at >= ?", today).Count(&count).Error
	return count, err
}
