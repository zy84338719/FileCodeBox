package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
	"github.com/zy84338719/fileCodeBox/internal/repo/db"
	"github.com/zy84338719/fileCodeBox/internal/repo/db/model"
)

type UserAPIKeyRepository struct {
}

func NewUserAPIKeyRepository() *UserAPIKeyRepository {
	return &UserAPIKeyRepository{}
}

func (r *UserAPIKeyRepository) db() *gorm.DB {
	return db.GetDB()
}

// Create 创建新的 API Key 记录
func (r *UserAPIKeyRepository) Create(ctx context.Context, key *model.UserAPIKey) error {
	return r.db().WithContext(ctx).Create(key).Error
}

// ListByUser 返回某个用户的所有密钥（包含已撤销），按创建时间倒序
func (r *UserAPIKeyRepository) ListByUser(ctx context.Context, userID uint) ([]*model.UserAPIKey, error) {
	var keys []*model.UserAPIKey
	err := r.db().WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&keys).Error
	if err != nil {
		return nil, err
	}
	return keys, nil
}

// GetActiveByHash 根据哈希获取有效密钥（未撤销且未过期）
func (r *UserAPIKeyRepository) GetActiveByHash(ctx context.Context, hash string) (*model.UserAPIKey, error) {
	var key model.UserAPIKey
	err := r.db().WithContext(ctx).Where("key_hash = ? AND revoked = ?", hash, false).First(&key).Error
	if err != nil {
		return nil, err
	}
	if key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now()) {
		return nil, gorm.ErrRecordNotFound
	}
	return &key, nil
}

// TouchLastUsed 更新最后使用时间
func (r *UserAPIKeyRepository) TouchLastUsed(ctx context.Context, id uint) error {
	now := time.Now()
	return r.db().WithContext(ctx).Model(&model.UserAPIKey{}).Where("id = ?", id).Updates(map[string]interface{}{
		"last_used_at": &now,
		"updated_at":   now,
	}).Error
}

// RevokeByID 撤销密钥
func (r *UserAPIKeyRepository) RevokeByID(ctx context.Context, userID, id uint) error {
	now := time.Now()
	res := r.db().WithContext(ctx).Model(&model.UserAPIKey{}).
		Where("id = ? AND user_id = ? AND revoked = ?", id, userID, false).
		Updates(map[string]interface{}{
			"revoked":    true,
			"revoked_at": &now,
			"updated_at": now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// CountActiveByUser 统计用户有效密钥数量
func (r *UserAPIKeyRepository) CountActiveByUser(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db().WithContext(ctx).Model(&model.UserAPIKey{}).
		Where("user_id = ? AND revoked = ?", userID, false).
		Count(&count).Error
	return count, err
}

// GetByID 根据ID获取API Key
func (r *UserAPIKeyRepository) GetByID(ctx context.Context, id uint) (*model.UserAPIKey, error) {
	var key model.UserAPIKey
	err := r.db().WithContext(ctx).First(&key, id).Error
	if err != nil {
		return nil, err
	}
	return &key, nil
}
