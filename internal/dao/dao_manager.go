package dao

import (
	"gorm.io/gorm"
)

// DAOManager 数据访问对象管理器
type DAOManager struct {
	db *gorm.DB

	// DAO 实例
	FileCode    *FileCodeDAO
	User        *UserDAO
	UserSession *UserSessionDAO
	Chunk       *ChunkDAO
	KeyValue    *KeyValueDAO
}

// NewDAOManager 创建新的 DAO 管理器
func NewDAOManager(db *gorm.DB) *DAOManager {
	return &DAOManager{

		FileCode:    NewFileCodeDAO(db),
		User:        NewUserDAO(db),
		UserSession: NewUserSessionDAO(db),
		Chunk:       NewChunkDAO(db),
		KeyValue:    NewKeyValueDAO(db),
	}
}

// BeginTransaction 开始事务
func (dm *DAOManager) BeginTransaction() *gorm.DB {
	return dm.db.Begin()
}
