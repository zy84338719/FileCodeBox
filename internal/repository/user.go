package repository

import (
	"errors"
	"time"

	"github.com/zy84338719/filecodebox/internal/models"
	"gorm.io/gorm"
)

// UserDAO 用户数据访问对象
type UserDAO struct {
	db *gorm.DB
}

// NewUserDAO 创建新的用户DAO
func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{db: db}
}

// Create 创建新用户
func (dao *UserDAO) Create(user *models.User) error {
	return dao.db.Create(user).Error
}

// GetByID 根据ID获取用户
func (dao *UserDAO) GetByID(id uint) (*models.User, error) {
	var user models.User
	err := dao.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (dao *UserDAO) GetByUsername(username string) (*models.User, error) {
	var user models.User
	err := dao.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail 根据邮箱获取用户
func (dao *UserDAO) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := dao.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUsernameOrEmail 根据用户名或邮箱获取用户
func (dao *UserDAO) GetByUsernameOrEmail(usernameOrEmail string) (*models.User, error) {
	var user models.User
	err := dao.db.Where("username = ? OR email = ?", usernameOrEmail, usernameOrEmail).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update 更新用户信息
func (dao *UserDAO) Update(user *models.User) error {
	return dao.db.Save(user).Error
}

// UpdateColumns 更新指定字段
func (dao *UserDAO) UpdateColumns(id uint, updates map[string]interface{}) error {
	return dao.db.Model(&models.User{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateUserFields 更新用户字段（结构化方式）
func (dao *UserDAO) UpdateUserFields(id uint, user models.User) error {
	// 直接使用结构体进行更新，GORM 会自动处理非零值字段
	result := dao.db.Model(&models.User{}).Where("id = ?", id).Updates(user)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("没有需要更新的字段或用户不存在")
	}
	return nil
}

// UpdateUserProfile 更新用户资料（用户自己更新）
func (dao *UserDAO) UpdateUserProfile(id uint, user *models.User) error {
	if user == nil {
		return errors.New("用户信息不能为空")
	}

	// 直接使用结构体进行更新，GORM 会自动处理非零值字段
	return dao.db.Model(&models.User{}).Where("id = ?", id).Updates(user).Error
}

// UpdatePassword 更新用户密码
func (dao *UserDAO) UpdatePassword(id uint, passwordHash string) error {
	return dao.db.Model(&models.User{}).Where("id = ?", id).Update("password_hash", passwordHash).Error
}

// UpdateStatus 更新用户状态
func (dao *UserDAO) UpdateStatus(id uint, status string) error {
	return dao.db.Model(&models.User{}).Where("id = ?", id).Update("status", status).Error
}

// Delete 删除用户
func (dao *UserDAO) Delete(tx *gorm.DB, user *models.User) error {
	return tx.Delete(user).Error
}

// CheckExists 检查用户是否存在（用户名或邮箱）
func (dao *UserDAO) CheckExists(username, email string) (*models.User, error) {
	var existingUser models.User
	err := dao.db.Where("username = ? OR email = ?", username, email).First(&existingUser).Error
	if err != nil {
		return nil, err
	}
	return &existingUser, nil
}

// CheckEmailExists 检查邮箱是否存在（排除指定ID）
func (dao *UserDAO) CheckEmailExists(email string, excludeID uint) (*models.User, error) {
	var existingUser models.User
	err := dao.db.Where("email = ? AND id != ?", email, excludeID).First(&existingUser).Error
	if err != nil {
		return nil, err
	}
	return &existingUser, nil
}

// Count 统计用户总数
func (dao *UserDAO) Count() (int64, error) {
	var count int64
	err := dao.db.Model(&models.User{}).Count(&count).Error
	return count, err
}

// CountActive 统计活跃用户数
func (dao *UserDAO) CountActive() (int64, error) {
	var count int64
	err := dao.db.Model(&models.User{}).Where("status = ?", "active").Count(&count).Error
	return count, err
}

// CountTodayRegistrations 统计今天注册的用户数
func (dao *UserDAO) CountTodayRegistrations() (int64, error) {
	var count int64
	today := time.Now().Format("2006-01-02")
	err := dao.db.Model(&models.User{}).Where("created_at >= ?", today).Count(&count).Error
	return count, err
}

// CountAdminUsers 统计管理员用户数量
func (dao *UserDAO) CountAdminUsers() (int64, error) {
	var count int64
	err := dao.db.Model(&models.User{}).Where("role = ?", "admin").Count(&count).Error
	return count, err
}

// List 分页获取用户列表
func (dao *UserDAO) List(page, pageSize int, search string) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := dao.db.Model(&models.User{})

	// 搜索条件
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("username LIKE ? OR email LIKE ? OR nickname LIKE ?",
			searchPattern, searchPattern, searchPattern)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&users).Error

	return users, total, err
}

// GetAllUsers 获取所有用户（管理员用途）
func (dao *UserDAO) GetAllUsers(page, pageSize int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	// 获取总数
	if err := dao.db.Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := dao.db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&users).Error

	return users, total, err
}
