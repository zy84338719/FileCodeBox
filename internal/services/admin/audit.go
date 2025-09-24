package admin

import (
	"errors"

	"github.com/zy84338719/filecodebox/internal/models"
)

// GetTransferLogs 返回传输日志列表
func (s *Service) GetTransferLogs(page, pageSize int, operation, search string) ([]models.TransferLog, int64, error) {
	if s.repositoryManager == nil || s.repositoryManager.TransferLog == nil {
		return nil, 0, errors.New("传输日志存储未初始化")
	}
	query := models.TransferLogQuery{
		Page:      page,
		PageSize:  pageSize,
		Operation: operation,
		Search:    search,
	}
	return s.repositoryManager.TransferLog.List(query)
}

// RecordOperationLog 记录后台运维操作
func (s *Service) RecordOperationLog(log *models.AdminOperationLog) error {
	if s.repositoryManager == nil || s.repositoryManager.AdminOpLog == nil {
		return errors.New("运维审计存储未初始化")
	}
	if log == nil {
		return errors.New("操作日志为空")
	}
	return s.repositoryManager.AdminOpLog.Create(log)
}

// GetOperationLogs 获取运维审计日志
func (s *Service) GetOperationLogs(page, pageSize int, action, actor string, success *bool) ([]models.AdminOperationLog, int64, error) {
	if s.repositoryManager == nil || s.repositoryManager.AdminOpLog == nil {
		return nil, 0, errors.New("运维审计存储未初始化")
	}
	query := models.AdminOperationLogQuery{
		Action:   action,
		Actor:    actor,
		Success:  success,
		Page:     page,
		PageSize: pageSize,
	}
	return s.repositoryManager.AdminOpLog.List(query)
}
