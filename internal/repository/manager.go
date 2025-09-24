package repository

import (
	"gorm.io/gorm"
)

// RepositoryManager 数据访问管理器
type RepositoryManager struct {
	db          *gorm.DB
	User        *UserDAO
	FileCode    *FileCodeDAO
	Chunk       *ChunkDAO
	UserSession *UserSessionDAO
	Upload      *ChunkDAO
	TransferLog *TransferLogDAO
	AdminOpLog  *AdminOperationLogDAO
}

// NewRepositoryManager 创建新的数据访问管理器
func NewRepositoryManager(db *gorm.DB) *RepositoryManager {
	return &RepositoryManager{
		db:          db,
		User:        NewUserDAO(db),
		FileCode:    NewFileCodeDAO(db),
		Chunk:       NewChunkDAO(db),
		UserSession: NewUserSessionDAO(db),
		Upload:      NewChunkDAO(db), // 别名
		TransferLog: NewTransferLogDAO(db),
		AdminOpLog:  NewAdminOperationLogDAO(db),
	}
}

// BeginTransaction 开始事务
func (m *RepositoryManager) BeginTransaction() *gorm.DB {
	return m.db.Begin()
}

// DB 返回底层 gorm.DB 引用（只读）
func (m *RepositoryManager) DB() *gorm.DB {
	return m.db
}
