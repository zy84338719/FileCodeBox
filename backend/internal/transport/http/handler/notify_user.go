package handler

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	notifyapp "github.com/zy84338719/fileCodeBox/backend/internal/app/notify"
	"github.com/zy84338719/fileCodeBox/backend/internal/transport/http/middleware"
)

var notifySvc *notifyapp.Service

// SetNotifyService 注入 notify service
func SetNotifyService(s *notifyapp.Service) {
	notifySvc = s
}

func getNotifyService() *notifyapp.Service {
	if notifySvc == nil {
		notifySvc = notifyapp.NewService(nil)
	}
	return notifySvc
}

// ListMyNotifications 我的通知列表（含广播 + 定向）
// GET /api/v1/notifies/mine?page=1&page_size=20
func ListMyNotifications(ctx context.Context, c *app.RequestContext) {
	uid, ok := userIDFromCtx(c)
	if !ok {
		c.JSON(consts.StatusUnauthorized, map[string]interface{}{"code": 401, "message": "未登录"})
		return
	}
	page, _ := strconv.Atoi(string(c.Query("page")))
	pageSize, _ := strconv.Atoi(string(c.Query("page_size")))
	data, err := getNotifyService().ListForUser(ctx, uid, page, pageSize)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]interface{}{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(consts.StatusOK, map[string]interface{}{"code": 200, "message": "ok", "data": data})
}

// UnreadNotifyCount 未读数
// GET /api/v1/notifies/unread-count
func UnreadNotifyCount(ctx context.Context, c *app.RequestContext) {
	uid, ok := userIDFromCtx(c)
	if !ok {
		c.JSON(consts.StatusOK, map[string]interface{}{"code": 200, "message": "ok", "data": map[string]int64{"unread": 0}})
		return
	}
	n, err := getNotifyService().UnreadCountForUser(ctx, uid)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]interface{}{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(consts.StatusOK, map[string]interface{}{"code": 200, "message": "ok", "data": map[string]int64{"unread": n}})
}

// MarkNotifyReq 标记已读请求
type MarkNotifyReq struct {
	All bool `json:"all"` // true=全部已读；忽略下面的 ids
}

// MarkNotifyRead 标记已读
// POST /api/v1/notifies/mark-read
func MarkNotifyRead(ctx context.Context, c *app.RequestContext) {
	uid, ok := userIDFromCtx(c)
	if !ok {
		c.JSON(consts.StatusUnauthorized, map[string]interface{}{"code": 401, "message": "未登录"})
		return
	}
	var req MarkNotifyReq
	// body 可为空，all=true 也行
	_ = c.BindAndValidate(&req)
	// 简化：只支持全部已读（per-id 已读前端基本不用）
	if !req.All {
		c.JSON(consts.StatusBadRequest, map[string]interface{}{"code": 400, "message": "仅支持 all=true"})
		return
	}
	n, err := getNotifyService().MarkAllReadForUser(ctx, uid)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]interface{}{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(consts.StatusOK, map[string]interface{}{"code": 200, "message": "ok", "data": map[string]int64{"marked": n}})
}

// 引用 middleware 包以保留 import
var _ = middleware.ContextKeyUserID
