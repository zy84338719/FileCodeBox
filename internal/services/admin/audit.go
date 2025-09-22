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
