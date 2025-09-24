package repository

import (
	"github.com/zy84338719/filecodebox/internal/models"
	"github.com/zy84338719/filecodebox/internal/models/db"
	"gorm.io/gorm"
)

// AdminOperationLogDAO 管理后台运维日志
type AdminOperationLogDAO struct {
	db *gorm.DB
}

func NewAdminOperationLogDAO(db *gorm.DB) *AdminOperationLogDAO {
	return &AdminOperationLogDAO{db: db}
}

func (dao *AdminOperationLogDAO) Create(log *models.AdminOperationLog) error {
	if dao.db == nil {
		return gorm.ErrInvalidDB
	}
	return dao.db.Create(log).Error
}

func (dao *AdminOperationLogDAO) List(query db.AdminOperationLogQuery) ([]models.AdminOperationLog, int64, error) {
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

	dbQuery := dao.db.Model(&models.AdminOperationLog{})

	if query.Action != "" {
		dbQuery = dbQuery.Where("action = ?", query.Action)
	}

	if query.Actor != "" {
		like := "%" + query.Actor + "%"
		dbQuery = dbQuery.Where("actor_name LIKE ?", like)
	}

	if query.Success != nil {
		dbQuery = dbQuery.Where("success = ?", *query.Success)
	}

	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	var logs []models.AdminOperationLog
	if err := dbQuery.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
