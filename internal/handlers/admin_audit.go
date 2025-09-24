package handlers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/zy84338719/filecodebox/internal/common"
	"github.com/zy84338719/filecodebox/internal/models"
	"github.com/zy84338719/filecodebox/internal/models/web"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *AdminHandler) GetOperationLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	action := c.Query("action")
	actor := c.Query("actor")
	successParam := c.Query("success")

	var success *bool
	if successParam != "" {
		switch successParam {
		case "true", "1":
			v := true
			success = &v
		case "false", "0":
			v := false
			success = &v
		}
	}

	logs, total, err := h.service.GetOperationLogs(page, pageSize, action, actor, success)
	if err != nil {
		common.InternalServerErrorResponse(c, "获取运维日志失败: "+err.Error())
		return
	}

	items := make([]web.AdminOperationLogItem, 0, len(logs))
	for _, logEntry := range logs {
		item := web.AdminOperationLogItem{
			ID:        logEntry.ID,
			Action:    logEntry.Action,
			Target:    logEntry.Target,
			Success:   logEntry.Success,
			Message:   logEntry.Message,
			ActorName: logEntry.ActorName,
			IP:        logEntry.IP,
			LatencyMs: logEntry.LatencyMs,
			CreatedAt: logEntry.CreatedAt.Format(time.RFC3339),
		}
		if logEntry.ActorID != nil {
			id := *logEntry.ActorID
			item.ActorID = &id
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

	response := web.AdminOperationLogListResponse{
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

func (h *AdminHandler) recordOperationLog(c *gin.Context, action, target string, success bool, msg string, started time.Time) {
	if h == nil || h.service == nil {
		return
	}

	var actorID *uint
	if rawID, exists := c.Get("user_id"); exists {
		switch v := rawID.(type) {
		case uint:
			id := v
			actorID = &id
		case uint64:
			id := uint(v)
			actorID = &id
		case int:
			if v >= 0 {
				id := uint(v)
				actorID = &id
			}
		}
	}

	actorName := ""
	if v, exists := c.Get("username"); exists {
		actorName = fmt.Sprint(v)
	}
	if actorName == "" {
		actorName = "<unknown>"
	}

	entry := &models.AdminOperationLog{
		Action:    action,
		Target:    target,
		Success:   success,
		Message:   msg,
		ActorName: actorName,
		IP:        c.ClientIP(),
	}
	if actorID != nil {
		entry.ActorID = actorID
	}
	if !started.IsZero() {
		entry.LatencyMs = time.Since(started).Milliseconds()
	}

	if err := h.service.RecordOperationLog(entry); err != nil {
		logrus.WithError(err).Warn("record operation log failed")
	}
}
