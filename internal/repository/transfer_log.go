package repository

import (
	"github.com/zy84338719/filecodebox/internal/models"
	"github.com/zy84338719/filecodebox/internal/models/db"
	"gorm.io/gorm"
)

// TransferLogDAO 负责上传/下载日志的持久化
// 目前仅提供简单的写入接口，便于后续扩展查询统计等功能

type TransferLogDAO struct {
	db *gorm.DB
}

func NewTransferLogDAO(db *gorm.DB) *TransferLogDAO {
	return &TransferLogDAO{db: db}
}

func (dao *TransferLogDAO) Create(log *models.TransferLog) error {
	return dao.db.Create(log).Error
}

func (dao *TransferLogDAO) WithDB(db *gorm.DB) *TransferLogDAO {
	return &TransferLogDAO{db: db}
}

// List 返回传输日志，支持基本筛选和分页
func (dao *TransferLogDAO) List(query db.TransferLogQuery) ([]models.TransferLog, int64, error) {
	if dao.db == nil {
		return nil, 0, gorm.ErrInvalidDB
	}

	page := query.Page
	if page < 1 {
		page = 1
	}

	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}

	dbQuery := dao.db.Model(&models.TransferLog{})

	if query.Operation != "" {
		dbQuery = dbQuery.Where("operation = ?", query.Operation)
	}

	if query.UserID != nil {
		dbQuery = dbQuery.Where("user_id = ?", *query.UserID)
	}

	if query.Search != "" {
		like := "%" + query.Search + "%"
		dbQuery = dbQuery.Where(
			"file_code LIKE ? OR file_name LIKE ? OR username LIKE ? OR ip LIKE ?",
			like, like, like, like,
		)
	}

	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	var logs []models.TransferLog
	if err := dbQuery.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
