package dao_preview

import (
	"context"
	"github.com/zy84338719/fileCodeBox/internal/repo/db"
	"github.com/zy84338719/fileCodeBox/internal/repo/db/model"
	"gorm.io/gorm"
)

// FilePreviewRepository 预览信息仓库
type FilePreviewRepository struct {
}

// NewFilePreviewRepository 创建预览仓库
func NewFilePreviewRepository() *FilePreviewRepository {
	return &FilePreviewRepository{}
}

func (r *FilePreviewRepository) db() *gorm.DB {
	return db.GetDB()
}

// Create 创建预览信息
func (r *FilePreviewRepository) Create(ctx context.Context, preview *model.FilePreview) error {
	return r.db().WithContext(ctx).Create(preview).Error
}

// GetByFileCodeID 根据文件ID获取预览信息
func (r *FilePreviewRepository) GetByFileCodeID(ctx context.Context, fileCodeID uint) (*model.FilePreview, error) {
	var preview model.FilePreview
	err := r.db().WithContext(ctx).Where("file_code_id = ?", fileCodeID).First(&preview).Error
	if err != nil {
		return nil, err
	}
	return &preview, nil
}

// Update 更新预览信息
func (r *FilePreviewRepository) Update(ctx context.Context, preview *model.FilePreview) error {
	return r.db().WithContext(ctx).Save(preview).Error
}

// Delete 删除预览信息
func (r *FilePreviewRepository) Delete(ctx context.Context, fileCodeID uint) error {
	return r.db().WithContext(ctx).Where("file_code_id = ?", fileCodeID).Delete(&model.FilePreview{}).Error
}
