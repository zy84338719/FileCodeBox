package repository

import (
	"time"

	"github.com/zy84338719/filecodebox/internal/models"
	"gorm.io/gorm"
)

// UserAPIKeyDAO 管理用户 API Key 的持久化
// 所有查询默认只返回未撤销的记录，除非特别说明

type UserAPIKeyDAO struct {
	db *gorm.DB
}

// NewUserAPIKeyDAO 创建 DAO
func NewUserAPIKeyDAO(db *gorm.DB) *UserAPIKeyDAO {
	return &UserAPIKeyDAO{db: db}
}

// Create 创建新的 API Key 记录
func (dao *UserAPIKeyDAO) Create(key *models.UserAPIKey) error {
	return dao.db.Create(key).Error
}

// ListByUser 返回某个用户的所有密钥（包含已撤销），按创建时间倒序
func (dao *UserAPIKeyDAO) ListByUser(userID uint) ([]models.UserAPIKey, error) {
	var keys []models.UserAPIKey
	err := dao.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&keys).Error
	if err != nil {
		return nil, err
	}
	return keys, nil
}

// GetActiveByHash 根据哈希获取有效密钥（未撤销且未过期）
func (dao *UserAPIKeyDAO) GetActiveByHash(hash string) (*models.UserAPIKey, error) {
	var key models.UserAPIKey
	err := dao.db.Where("key_hash = ? AND revoked = ?", hash, false).First(&key).Error
	if err != nil {
		return nil, err
	}
	if key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now()) {
		return nil, gorm.ErrRecordNotFound
	}
	return &key, nil
}

// TouchLastUsed 更新最后使用时间
func (dao *UserAPIKeyDAO) TouchLastUsed(id uint) error {
	now := time.Now()
	return dao.db.Model(&models.UserAPIKey{}).Where("id = ?", id).Updates(map[string]interface{}{
		"last_used_at": &now,
		"updated_at":   now,
	}).Error
}

// RevokeByID 撤销密钥
func (dao *UserAPIKeyDAO) RevokeByID(userID, id uint) error {
	now := time.Now()
	res := dao.db.Model(&models.UserAPIKey{}).
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
func (dao *UserAPIKeyDAO) CountActiveByUser(userID uint) (int64, error) {
	var count int64
	err := dao.db.Model(&models.UserAPIKey{}).
		Where("user_id = ? AND revoked = ?", userID, false).
		Count(&count).Error
	return count, err
}
