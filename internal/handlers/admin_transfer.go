package handlers

import (
	"strconv"
	"strings"
	"time"

	"github.com/zy84338719/filecodebox/internal/common"
	"github.com/zy84338719/filecodebox/internal/models/web"

	"github.com/gin-gonic/gin"
)

// GetTransferLogs 获取上传/下载审计日志
func (h *AdminHandler) GetTransferLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	} else if pageSize > 200 {
		pageSize = 200
	}

	operation := strings.TrimSpace(c.DefaultQuery("operation", ""))
	search := strings.TrimSpace(c.DefaultQuery("search", ""))

	logs, total, err := h.service.GetTransferLogs(page, pageSize, operation, search)
	if err != nil {
		common.InternalServerErrorResponse(c, "获取传输日志失败: "+err.Error())
		return
	}

	items := make([]web.TransferLogItem, 0, len(logs))
	for _, record := range logs {
		item := web.TransferLogItem{
			ID:         record.ID,
			Operation:  record.Operation,
			FileCode:   record.FileCode,
			FileName:   record.FileName,
			FileSize:   record.FileSize,
			Username:   record.Username,
			IP:         record.IP,
			DurationMs: record.DurationMs,
			CreatedAt:  record.CreatedAt.Format(time.RFC3339),
		}
		if record.UserID != nil {
			id := *record.UserID
			item.UserID = &id
		}
		items = append(items, item)
	}

	pages := int64(0)
	if pageSize > 0 {
		pages = (total + int64(pageSize) - 1) / int64(pageSize)
	}
	if pages == 0 {
		pages = 1
	}

	response := web.TransferLogListResponse{
		Logs: items,
		Pagination: web.PaginationResponse{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
			Pages:    pages,
		},
	}

	common.SuccessResponse(c, response)
}
