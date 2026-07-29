package dao

import (
	"context"
	"time"

	"github.com/zy84338719/fileCodeBox/backend/internal/repo/db"
	"github.com/zy84338719/fileCodeBox/backend/internal/repo/db/model"
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

// UserShareFilter 用户分享列表筛选条件
type UserShareFilter struct {
	Status   string // all / active / expired / text / file / deleted
	Search   string // 模糊搜索 code / 文件名
	Page     int
	PageSize int
}

// GetUserSharesWithFilter 获取用户的分享列表（带筛选）
//   - "active": 未过期且有剩余次数
//   - "expired": 时间过期 或 次数用尽
//   - "text": 文本分享（Text != ""）
//   - "file": 文件分享（Text == ""）
//   - "deleted": 软删除的（deleted_at != null）
//   - "all" / "": 不过滤状态
func (r *FileCodeRepository) GetUserSharesWithFilter(ctx context.Context, userID uint, filter UserShareFilter) ([]*model.FileCode, int64, error) {
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	now := time.Now()
	q := r.db().WithContext(ctx).Model(&model.FileCode{}).Where("user_id = ?", userID)

	switch filter.Status {
	case "deleted":
		// 只看已删除
		q = q.Unscoped().Where("deleted_at IS NOT NULL")
	case "active":
		q = q.Where("(expired_at IS NULL OR expired_at > ?) AND (expired_count <> 0)", now)
	case "expired":
		q = q.Where("((expired_at IS NOT NULL AND expired_at <= ?) OR expired_count = 0)")
	case "text":
		q = q.Where("text <> ''")
	case "file":
		q = q.Where("(text = '' OR text IS NULL)")
	}

	if filter.Search != "" {
		like := "%" + filter.Search + "%"
		q = q.Where("code LIKE ? OR prefix LIKE ? OR suffix LIKE ? OR uuid_file_name LIKE ?",
			like, like, like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	var files []*model.FileCode
	if err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&files).Error; err != nil {
		return nil, 0, err
	}
	return files, total, nil
}

// BatchSoftDeleteByCodes 按 code 列表软删除（限定 userID 防止越权）
// 返回 (受影响行数, error)
func (r *FileCodeRepository) BatchSoftDeleteByCodes(ctx context.Context, userID uint, codes []string) (int, error) {
	if len(codes) == 0 {
		return 0, nil
	}
	res := r.db().WithContext(ctx).Model(&model.FileCode{}).
		Where("user_id = ? AND code IN ?", userID, codes).
		Update("deleted_at", time.Now())
	if res.Error != nil {
		return 0, res.Error
	}
	return int(res.RowsAffected), nil
}

// BatchExtendByCodes 批量延期
// newExpireAt 为 nil 表示永久（清空 expired_at 字段）
func (r *FileCodeRepository) BatchExtendByCodes(ctx context.Context, userID uint, codes []string, newExpireAt *time.Time) (int, error) {
	if len(codes) == 0 {
		return 0, nil
	}
	updates := map[string]interface{}{}
	if newExpireAt == nil {
		updates["expired_at"] = nil
		updates["expired_count"] = -1
	} else {
		updates["expired_at"] = *newExpireAt
		// 次数设为 -1（无限）让延期后能继续取
		updates["expired_count"] = -1
	}
	res := r.db().WithContext(ctx).Model(&model.FileCode{}).
		Where("user_id = ? AND code IN ?", userID, codes).
		Updates(updates)
	if res.Error != nil {
		return 0, res.Error
	}
	return int(res.RowsAffected), nil
}

// RestoreByCode 恢复软删除的分享（仅 owner）
func (r *FileCodeRepository) RestoreByCode(ctx context.Context, userID uint, code string) error {
	return r.db().WithContext(ctx).Unscoped().Model(&model.FileCode{}).
		Where("user_id = ? AND code = ? AND deleted_at IS NOT NULL", userID, code).
		Update("deleted_at", nil).Error
}

// HardDeleteByCode 永久删除（仅 owner，已软删除的）
func (r *FileCodeRepository) HardDeleteByCode(ctx context.Context, userID uint, code string) error {
	return r.db().WithContext(ctx).Unscoped().Model(&model.FileCode{}).
		Where("user_id = ? AND code = ? AND deleted_at IS NOT NULL", userID, code).
		Delete(&model.FileCode{}).Error
}

// DecrementExpiredCount 原子扣减剩余次数。
// ExpiredCount 语义：-1=无限(只 +used_count), 0=已耗尽(拒绝), >0=剩余(扣减)
// 返回 ok=true 表示扣减成功；ok=false 表示已耗尽（未扣减）。
// 用单条 UPDATE 的 WHERE 条件保证原子性，避免并发超卖。
func (r *FileCodeRepository) DecrementExpiredCount(ctx context.Context, code string) (bool, error) {
	res := r.db().WithContext(ctx).Model(&model.FileCode{}).
		Where("code = ? AND (expired_count = -1 OR expired_count > 0)", code).
		UpdateColumns(map[string]interface{}{
			"expired_count": gorm.Expr("CASE WHEN expired_count > 0 THEN expired_count - 1 ELSE expired_count END"),
			"used_count":    gorm.Expr("used_count + 1"),
		})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

// UpdateViewer 记录取件人信息（IP + 时间 + 累计 +1）
func (r *FileCodeRepository) UpdateViewer(ctx context.Context, code, viewerIP string) error {
	now := time.Now()
	// 用 SQL 原子自增 viewer_count
	return r.db().WithContext(ctx).Model(&model.FileCode{}).
		Where("code = ?", code).
		Updates(map[string]interface{}{
			"viewer_ip":    viewerIP,
			"viewer_at":    now,
			"viewer_count": gorm.Expr("viewer_count + 1"),
		}).Error
}
