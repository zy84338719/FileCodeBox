package dao

import (
	"context"

	"github.com/zy84338719/fileCodeBox/backend/internal/repo/db"
	"github.com/zy84338719/fileCodeBox/backend/internal/repo/db/model"
	"gorm.io/gorm"
)

type AdminOperationLogRepository struct {
}

func NewAdminOperationLogRepository() *AdminOperationLogRepository {
	return &AdminOperationLogRepository{}
}

func (r *AdminOperationLogRepository) db() *gorm.DB {
	return db.GetDB()
}

func (r *AdminOperationLogRepository) Create(ctx context.Context, log *model.AdminOperationLog) error {
	return r.db().WithContext(ctx).Create(log).Error
}

func (r *AdminOperationLogRepository) List(ctx context.Context, query model.AdminOperationLogQuery) ([]*model.AdminOperationLog, int64, error) {
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

	dbQuery := r.db().WithContext(ctx).Model(&model.AdminOperationLog{})

	if query.Action != "" {
		dbQuery = dbQuery.Where("action = ?", query.Action)
	}

	if query.Actor != "" {
		dbQuery = dbQuery.Where("actor_name LIKE ?", "%"+query.Actor+"%")
	}

	if query.Success != nil {
		dbQuery = dbQuery.Where("success = ?", *query.Success)
	}

	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	var logs []*model.AdminOperationLog
	if err := dbQuery.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
