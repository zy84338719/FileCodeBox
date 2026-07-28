// Package handler 提供定制 HTTP handlers（不走 thrift IDL 生成）。
package handler

import (
	"context"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"github.com/zy84338719/fileCodeBox/backend/internal/app/share"
	"github.com/zy84338719/fileCodeBox/backend/internal/repo/db/dao"
	"github.com/zy84338719/fileCodeBox/backend/internal/transport/http/middleware"
)

var shareSvc *share.Service

// SetShareService 注入 share service（bootstrap 调用）
func SetShareService(s *share.Service) {
	shareSvc = s
}

func getShareService() *share.Service {
	if shareSvc == nil {
		shareSvc = share.NewService("", nil)
	}
	return shareSvc
}

// userIDFromCtx 提取 userID，未登录返回 0
func userIDFromCtx(c *app.RequestContext) (uint, bool) {
	v, ok := c.Get(middleware.ContextKeyUserID)
	if !ok {
		return 0, false
	}
	id, ok := v.(uint)
	return id, ok
}

// ==================== 我的分享列表 ====================

// UserSharesResp 列表响应
type UserSharesResp struct {
	Code    int                          `json:"code"`
	Message string                       `json:"message"`
	Data    *UserSharesListData          `json:"data,omitempty"`
}

type UserSharesListData struct {
	Items      []*share.UserShareListItem `json:"items"`
	Total      int64                      `json:"total"`
	Page       int                        `json:"page"`
	PageSize   int                        `json:"page_size"`
	TotalPages int64                      `json:"total_pages"`
	HasNext    bool                       `json:"has_next"`
	HasPrev    bool                       `json:"has_prev"`
}

// ListUserShares 我的分享列表
// GET /api/v1/user/shares?status=active&search=&page=1&page_size=20
func ListUserShares(ctx context.Context, c *app.RequestContext) {
	uid, ok := userIDFromCtx(c)
	if !ok {
		c.JSON(consts.StatusUnauthorized, map[string]interface{}{
			"code": 401, "message": "未登录",
		})
		return
	}

	status := string(c.Query("status"))
	search := string(c.Query("search"))
	page, _ := strconv.Atoi(string(c.Query("page")))
	pageSize, _ := strconv.Atoi(string(c.Query("page_size")))

	items, total, err := getShareService().ListUserShares(ctx, uid, dao.UserShareFilter{
		Status:   status,
		Search:   search,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]interface{}{
			"code": 500, "message": "获取分享列表失败: " + err.Error(),
		})
		return
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)
	if totalPages < 1 {
		totalPages = 1
	}

	c.JSON(consts.StatusOK, UserSharesResp{
		Code: 200, Message: "ok",
		Data: &UserSharesListData{
			Items:      items,
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
			HasNext:    int64(page) < totalPages,
			HasPrev:    int64(page) > 1,
		},
	})
}

// ==================== 批量删除 ====================

// BatchDeleteUserSharesReq 批量删除请求体
type BatchDeleteUserSharesReq struct {
	Codes []string `json:"codes"`
}

// BatchDeleteUserShares 批量软删除我的分享
// POST /api/v1/user/shares/batch-delete
func BatchDeleteUserShares(ctx context.Context, c *app.RequestContext) {
	uid, ok := userIDFromCtx(c)
	if !ok {
		c.JSON(consts.StatusUnauthorized, map[string]interface{}{"code": 401, "message": "未登录"})
		return
	}
	var req BatchDeleteUserSharesReq
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(consts.StatusBadRequest, map[string]interface{}{"code": 400, "message": err.Error()})
		return
	}
	if len(req.Codes) == 0 {
		c.JSON(consts.StatusBadRequest, map[string]interface{}{"code": 400, "message": "codes 不能为空"})
		return
	}
	n, err := getShareService().BatchDeleteUserShares(ctx, uid, req.Codes)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]interface{}{"code": 500, "message": "删除失败: " + err.Error()})
		return
	}
	c.JSON(consts.StatusOK, map[string]interface{}{
		"code": 200, "message": "ok", "data": map[string]int{"deleted": n},
	})
}

// ==================== 批量延期 ====================

// BatchExtendUserSharesReq 批量延期请求
type BatchExtendUserSharesReq struct {
	Codes   []string `json:"codes"`
	Hours   int      `json:"hours"`   // 延长小时数（>0），0 = 永久
	Forever bool     `json:"forever"` // true = 设为永久
}

// BatchExtendUserShares 批量延期我的分享
// POST /api/v1/user/shares/batch-extend
func BatchExtendUserShares(ctx context.Context, c *app.RequestContext) {
	uid, ok := userIDFromCtx(c)
	if !ok {
		c.JSON(consts.StatusUnauthorized, map[string]interface{}{"code": 401, "message": "未登录"})
		return
	}
	var req BatchExtendUserSharesReq
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(consts.StatusBadRequest, map[string]interface{}{"code": 400, "message": err.Error()})
		return
	}
	if len(req.Codes) == 0 {
		c.JSON(consts.StatusBadRequest, map[string]interface{}{"code": 400, "message": "codes 不能为空"})
		return
	}
	var newExpire *time.Time
	if req.Forever {
		// nil 表示永久（清空 expired_at 字段）
		newExpire = nil
	} else if req.Hours > 0 {
		t := time.Now().Add(time.Duration(req.Hours) * time.Hour)
		newExpire = &t
	} else {
		c.JSON(consts.StatusBadRequest, map[string]interface{}{"code": 400, "message": "hours 必须 > 0 或 forever=true"})
		return
	}
	n, err := getShareService().BatchExtendUserShares(ctx, uid, req.Codes, newExpire)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]interface{}{"code": 500, "message": "延期失败: " + err.Error()})
		return
	}
	c.JSON(consts.StatusOK, map[string]interface{}{
		"code": 200, "message": "ok", "data": map[string]int{"extended": n},
	})
}

// ==================== 恢复 + 永久删除 ====================

// RestoreUserShare 恢复软删除的分享
// POST /api/v1/user/shares/:code/restore
func RestoreUserShare(ctx context.Context, c *app.RequestContext) {
	uid, ok := userIDFromCtx(c)
	if !ok {
		c.JSON(consts.StatusUnauthorized, map[string]interface{}{"code": 401, "message": "未登录"})
		return
	}
	code := c.Param("code")
	if code == "" {
		c.JSON(consts.StatusBadRequest, map[string]interface{}{"code": 400, "message": "code 必填"})
		return
	}
	if err := getShareService().RestoreUserShare(ctx, uid, code); err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]interface{}{"code": 500, "message": "恢复失败: " + err.Error()})
		return
	}
	c.JSON(consts.StatusOK, map[string]interface{}{"code": 200, "message": "ok"})
}

// HardDeleteUserShare 永久删除（仅已软删除的）
// DELETE /api/v1/user/shares/:code/hard
func HardDeleteUserShare(ctx context.Context, c *app.RequestContext) {
	uid, ok := userIDFromCtx(c)
	if !ok {
		c.JSON(consts.StatusUnauthorized, map[string]interface{}{"code": 401, "message": "未登录"})
		return
	}
	code := c.Param("code")
	if code == "" {
		c.JSON(consts.StatusBadRequest, map[string]interface{}{"code": 400, "message": "code 必填"})
		return
	}
	if err := getShareService().HardDeleteUserShare(ctx, uid, code); err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]interface{}{"code": 500, "message": "永久删除失败: " + err.Error()})
		return
	}
	c.JSON(consts.StatusOK, map[string]interface{}{"code": 200, "message": "ok"})
}
