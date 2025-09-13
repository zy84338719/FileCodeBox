package repository

import (
	"github.com/zy84338719/filecodebox/internal/models"
	"gorm.io/gorm"
)

// KeyValueDAO 键值对数据访问对象
type KeyValueDAO struct {
	db *gorm.DB
}

// NewKeyValueDAO 创建新的键值对DAO
func NewKeyValueDAO(db *gorm.DB) *KeyValueDAO {
	return &KeyValueDAO{db: db}
}

// Create 创建新的键值对
func (dao *KeyValueDAO) Create(kv *models.KeyValue) error {
	return dao.db.Create(kv).Error
}

// GetByKey 根据键获取值
func (dao *KeyValueDAO) GetByKey(key string) (*models.KeyValue, error) {
	var kv models.KeyValue
	err := dao.db.Where("key = ?", key).First(&kv).Error
	if err != nil {
		return nil, err
	}
	return &kv, nil
}

// Update 更新键值对
func (dao *KeyValueDAO) Update(kv *models.KeyValue) error {
	return dao.db.Save(kv).Error
}

// Delete 删除键值对
func (dao *KeyValueDAO) Delete(key string) error {
	return dao.db.Where("key = ?", key).Delete(&models.KeyValue{}).Error
}

// GetAll 获取所有键值对
func (dao *KeyValueDAO) GetAll() ([]models.KeyValue, error) {
	var kvs []models.KeyValue
	err := dao.db.Find(&kvs).Error
	return kvs, err
}

// GetByKeys 根据多个键获取值
func (dao *KeyValueDAO) GetByKeys(keys []string) ([]models.KeyValue, error) {
	var kvs []models.KeyValue
	err := dao.db.Where("key IN ?", keys).Find(&kvs).Error
	return kvs, err
}

// SetValue 设置键值对（如果存在则更新，不存在则创建）
func (dao *KeyValueDAO) SetValue(key, value string) error {
	var kv models.KeyValue
	err := dao.db.Where("key = ?", key).First(&kv).Error

	if err == gorm.ErrRecordNotFound {
		// 不存在，创建新记录
		kv = models.KeyValue{Key: key, Value: value}
		return dao.db.Create(&kv).Error
	} else if err != nil {
		return err
	}

	// 存在，更新值
	kv.Value = value
	return dao.db.Save(&kv).Error
}

// BatchSet 批量设置键值对
func (dao *KeyValueDAO) BatchSet(kvMap map[string]string) error {
	return dao.db.Transaction(func(tx *gorm.DB) error {
		for key, value := range kvMap {
			var kv models.KeyValue
			err := tx.Where("key = ?", key).First(&kv).Error

			if err == gorm.ErrRecordNotFound {
				// 不存在，创建新记录
				kv = models.KeyValue{Key: key, Value: value}
				if err := tx.Create(&kv).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			} else {
				// 存在，更新值
				kv.Value = value
				if err := tx.Save(&kv).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// Search 搜索键值对
func (dao *KeyValueDAO) Search(searchKey string) ([]models.KeyValue, error) {
	var kvs []models.KeyValue
	searchPattern := "%" + searchKey + "%"
	err := dao.db.Where("key LIKE ? OR value LIKE ?", searchPattern, searchPattern).Find(&kvs).Error
	return kvs, err
}

// Count 统计键值对数量
func (dao *KeyValueDAO) Count() (int64, error) {
	var count int64
	err := dao.db.Model(&models.KeyValue{}).Count(&count).Error
	return count, err
}
