package dao

import (
	"time"

	"github.com/zy84338719/filecodebox/internal/models"
	"gorm.io/gorm"
)

// UserSessionDAO 用户会话数据访问对象
type UserSessionDAO struct {
	db *gorm.DB
}

// NewUserSessionDAO 创建新的用户会话DAO
func NewUserSessionDAO(db *gorm.DB) *UserSessionDAO {
	return &UserSessionDAO{db: db}
}

// Create 创建新会话
func (dao *UserSessionDAO) Create(session *models.UserSession) error {
	return dao.db.Create(session).Error
}

// GetBySessionID 根据会话ID获取会话
func (dao *UserSessionDAO) GetBySessionID(sessionID string) (*models.UserSession, error) {
	var session models.UserSession
	err := dao.db.Where("session_id = ? AND is_active = true", sessionID).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// CountActiveSessionsByUserID 统计用户的活跃会话数
func (dao *UserSessionDAO) CountActiveSessionsByUserID(userID uint) (int64, error) {
	var count int64
	err := dao.db.Model(&models.UserSession{}).Where("user_id = ? AND is_active = true", userID).Count(&count).Error
	return count, err
}

// GetOldestSessionByUserID 获取用户最老的会话
func (dao *UserSessionDAO) GetOldestSessionByUserID(userID uint) (*models.UserSession, error) {
	var session models.UserSession
	err := dao.db.Where("user_id = ? AND is_active = true", userID).Order("created_at ASC").First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// UpdateIsActive 更新会话活跃状态
func (dao *UserSessionDAO) UpdateIsActive(session *models.UserSession, isActive bool) error {
	return dao.db.Model(session).Update("is_active", isActive).Error
}

// UpdateIsActiveByID 根据ID更新会话活跃状态
func (dao *UserSessionDAO) UpdateIsActiveByID(id uint, isActive bool) error {
	return dao.db.Model(&models.UserSession{}).Where("id = ?", id).Update("is_active", isActive).Error
}

// DeactivateUserSessions 停用用户的所有会话
func (dao *UserSessionDAO) DeactivateUserSessions(userID uint) error {
	return dao.db.Model(&models.UserSession{}).Where("user_id = ?", userID).Update("is_active", false).Error
}

// DeleteByUserID 删除用户的所有会话
func (dao *UserSessionDAO) DeleteByUserID(tx *gorm.DB, userID uint) error {
	return tx.Where("user_id = ?", userID).Delete(&models.UserSession{}).Error
}

// CleanExpiredSessions 清理过期会话
func (dao *UserSessionDAO) CleanExpiredSessions() error {
	return dao.db.Model(&models.UserSession{}).
		Where("expires_at < ? AND is_active = true", time.Now()).
		Update("is_active", false).Error
}

// GetUserSessions 获取用户的会话列表
func (dao *UserSessionDAO) GetUserSessions(userID uint, page, pageSize int) ([]models.UserSession, int64, error) {
	var sessions []models.UserSession
	var total int64

	// 获取总数
	if err := dao.db.Model(&models.UserSession{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := dao.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&sessions).Error

	return sessions, total, err
}
