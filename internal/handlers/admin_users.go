package handlers

import (
	"fmt"
	"strconv"

	"github.com/zy84338719/filecodebox/internal/common"
	"github.com/zy84338719/filecodebox/internal/models/web"
	"github.com/zy84338719/filecodebox/internal/services"
	"github.com/zy84338719/filecodebox/internal/utils"

	"github.com/gin-gonic/gin"
)

// GetUsers 获取用户列表
func (h *AdminHandler) GetUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	users, total, err := h.getUsersFromDB(page, pageSize, search)
	if err != nil {
		common.InternalServerErrorResponse(c, "获取用户列表失败: "+err.Error())
		return
	}

	stats, err := h.getUserStats()
	if err != nil {
		stats = &web.AdminUserStatsResponse{
			TotalUsers:         total,
			ActiveUsers:        total,
			TodayRegistrations: 0,
			TodayUploads:       0,
		}
	}

	pagination := web.PaginationResponse{
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		Pages:    (total + int64(pageSize) - 1) / int64(pageSize),
	}

	common.SuccessResponse(c, web.AdminUsersListResponse{
		Users:      users,
		Stats:      *stats,
		Pagination: pagination,
	})
}

func (h *AdminHandler) getUsersFromDB(page, pageSize int, search string) ([]web.AdminUserDetail, int64, error) {
	users, total, err := h.service.GetUsers(page, pageSize, search)
	if err != nil {
		return nil, 0, err
	}

	result := make([]web.AdminUserDetail, len(users))
	for i, user := range users {
		lastLoginAt := ""
		if user.LastLoginAt != nil {
			lastLoginAt = user.LastLoginAt.Format("2006-01-02 15:04:05")
		}

		result[i] = web.AdminUserDetail{
			ID:             user.ID,
			Username:       user.Username,
			Email:          user.Email,
			Nickname:       user.Nickname,
			Role:           user.Role,
			IsAdmin:        user.Role == "admin",
			IsActive:       user.Status == "active",
			Status:         user.Status,
			EmailVerified:  user.EmailVerified,
			CreatedAt:      user.CreatedAt.Format("2006-01-02 15:04:05"),
			LastLoginAt:    lastLoginAt,
			LastLoginIP:    user.LastLoginIP,
			TotalUploads:   user.TotalUploads,
			TotalDownloads: user.TotalDownloads,
			TotalStorage:   user.TotalStorage,
		}
	}

	return result, total, nil
}

func (h *AdminHandler) getUserStats() (*web.AdminUserStatsResponse, error) {
	stats, err := h.service.GetStats()
	if err != nil {
		return nil, err
	}

	return &web.AdminUserStatsResponse{
		TotalUsers:         stats.TotalUsers,
		ActiveUsers:        stats.ActiveUsers,
		TodayRegistrations: stats.TodayRegistrations,
		TodayUploads:       stats.TodayUploads,
	}, nil
}

// GetUser 获取单个用户
func (h *AdminHandler) GetUser(c *gin.Context) {
	userID, ok := utils.ParseUserIDFromParam(c, "id")
	if !ok {
		return
	}

	user, err := h.service.GetUserByID(userID)
	if err != nil {
		common.NotFoundResponse(c, "用户不存在")
		return
	}

	lastLoginAt := ""
	if user.LastLoginAt != nil {
		lastLoginAt = user.LastLoginAt.Format("2006-01-02 15:04:05")
	}

	userDetail := web.AdminUserDetail{
		ID:             user.ID,
		Username:       user.Username,
		Email:          user.Email,
		Nickname:       user.Nickname,
		Role:           user.Role,
		IsAdmin:        user.Role == "admin",
		IsActive:       user.Status == "active",
		Status:         user.Status,
		EmailVerified:  user.EmailVerified,
		CreatedAt:      user.CreatedAt.Format("2006-01-02 15:04:05"),
		LastLoginAt:    lastLoginAt,
		LastLoginIP:    user.LastLoginIP,
		TotalUploads:   user.TotalUploads,
		TotalDownloads: user.TotalDownloads,
		TotalStorage:   user.TotalStorage,
	}

	common.SuccessResponse(c, userDetail)
}

// CreateUser 创建用户
func (h *AdminHandler) CreateUser(c *gin.Context) {
	var userData web.UserDataRequest
	if !utils.BindJSONWithValidation(c, &userData) {
		return
	}

	role := "user"
	if userData.IsAdmin {
		role = "admin"
	}

	status := "active"
	if !userData.IsActive {
		status = "inactive"
	}

	user, err := h.service.CreateUser(userData.Username, userData.Email, userData.Password, userData.Nickname, role, status)
	if err != nil {
		common.InternalServerErrorResponse(c, "创建用户失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, "用户创建成功", web.IDResponse{ID: user.ID})
}

// UpdateUser 更新用户
func (h *AdminHandler) UpdateUser(c *gin.Context) {
	userID, ok := utils.ParseUserIDFromParam(c, "id")
	if !ok {
		return
	}

	var userData struct {
		Email    string `json:"email" binding:"omitempty,email"`
		Password string `json:"password"`
		Nickname string `json:"nickname"`
		IsAdmin  bool   `json:"is_admin"`
		IsActive bool   `json:"is_active"`
	}

	if !utils.BindJSONWithValidation(c, &userData) {
		return
	}

	params := services.AdminUserUpdateParams{}
	if userData.Email != "" {
		email := userData.Email
		params.Email = &email
	}
	if userData.Password != "" {
		password := userData.Password
		params.Password = &password
	}
	if userData.Nickname != "" {
		nickname := userData.Nickname
		params.Nickname = &nickname
	}
	isAdmin := userData.IsAdmin
	params.IsAdmin = &isAdmin
	isActive := userData.IsActive
	params.IsActive = &isActive

	if err := h.service.UpdateUserWithParams(userID, params); err != nil {
		common.InternalServerErrorResponse(c, "更新用户失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, "用户更新成功", web.IDResponse{ID: userID})
}

// DeleteUser 删除用户
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	userID, ok := utils.ParseUserIDFromParam(c, "id")
	if !ok {
		return
	}

	if err := h.service.DeleteUser(userID); err != nil {
		common.InternalServerErrorResponse(c, "删除用户失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, "用户删除成功", web.IDResponse{ID: userID})
}

// UpdateUserStatus 更新用户状态
func (h *AdminHandler) UpdateUserStatus(c *gin.Context) {
	userID, ok := utils.ParseUserIDFromParam(c, "id")
	if !ok {
		return
	}

	var statusData struct {
		IsActive bool `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&statusData); err != nil {
		common.BadRequestResponse(c, "参数错误: "+err.Error())
		return
	}

	if err := h.service.UpdateUserStatus(userID, statusData.IsActive); err != nil {
		common.InternalServerErrorResponse(c, "更新用户状态失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, "用户状态更新成功", web.IDResponse{ID: userID})
}

// GetUserFiles 获取用户文件
func (h *AdminHandler) GetUserFiles(c *gin.Context) {
	userID, ok := utils.ParseUserIDFromParam(c, "id")
	if !ok {
		return
	}

	user, err := h.service.GetUserByID(userID)
	if err != nil {
		common.NotFoundResponse(c, "用户不存在")
		return
	}

	page := 1
	limit := 20

	files, total, err := h.service.GetUserFiles(userID, page, limit)
	if err != nil {
		common.InternalServerErrorResponse(c, "获取用户文件失败: "+err.Error())
		return
	}

	fileList := make([]web.AdminFileDetail, len(files))
	for i, file := range files {
		expiredAt := ""
		if file.ExpiredAt != nil {
			expiredAt = file.ExpiredAt.Format("2006-01-02 15:04:05")
		}

		fileType := "文件"
		if file.Text != "" {
			fileType = "文本"
		}

		fileList[i] = web.AdminFileDetail{
			ID:           file.ID,
			Code:         file.Code,
			Prefix:       file.Prefix,
			Suffix:       file.Suffix,
			Size:         file.Size,
			Type:         fileType,
			ExpiredAt:    expiredAt,
			ExpiredCount: file.ExpiredCount,
			UsedCount:    file.UsedCount,
			CreatedAt:    file.CreatedAt.Format("2006-01-02 15:04:05"),
			RequireAuth:  file.RequireAuth,
			UploadType:   file.UploadType,
		}
	}

	common.SuccessResponse(c, web.AdminUserFilesResponse{
		Files:    fileList,
		Username: user.Username,
		Total:    total,
	})
}

// ExportUsers 导出用户列表为CSV
func (h *AdminHandler) ExportUsers(c *gin.Context) {
	users, _, err := h.getUsersFromDB(1, 10000, "")
	if err != nil {
		common.InternalServerErrorResponse(c, "获取用户数据失败: "+err.Error())
		return
	}

	csvContent := "用户名,邮箱,昵称,状态,注册时间,最后登录,上传次数,下载次数,存储大小(MB)\n"
	for _, user := range users {
		status := "正常"
		if user.Status == "disabled" || user.Status == "inactive" {
			status = "禁用"
		}

		lastLoginTime := "从未登录"
		if user.LastLoginAt != "" {
			lastLoginTime = user.LastLoginAt
		}

		csvContent += fmt.Sprintf("%s,%s,%s,%s,%s,%s,%d,%d,%.2f\n",
			user.Username,
			user.Email,
			user.Nickname,
			status,
			user.CreatedAt,
			lastLoginTime,
			user.TotalUploads,
			user.TotalDownloads,
			float64(user.TotalStorage)/(1024*1024),
		)
	}

	bomContent := "\xEF\xBB\xBF" + csvContent

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=users_export.csv")
	c.Header("Content-Length", strconv.Itoa(len([]byte(bomContent))))

	c.Writer.WriteHeader(200)
	_, _ = c.Writer.Write([]byte(bomContent))
}

// BatchEnableUsers 批量启用用户
func (h *AdminHandler) BatchEnableUsers(c *gin.Context) {
	var req struct {
		UserIDs []uint `json:"user_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequestResponse(c, "参数错误: "+err.Error())
		return
	}

	if len(req.UserIDs) == 0 {
		common.BadRequestResponse(c, "user_ids 不能为空")
		return
	}

	if err := h.service.BatchUpdateUserStatus(req.UserIDs, true); err != nil {
		common.InternalServerErrorResponse(c, "批量启用用户失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, "批量启用成功", nil)
}

// BatchDisableUsers 批量禁用用户
func (h *AdminHandler) BatchDisableUsers(c *gin.Context) {
	var req struct {
		UserIDs []uint `json:"user_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequestResponse(c, "参数错误: "+err.Error())
		return
	}

	if len(req.UserIDs) == 0 {
		common.BadRequestResponse(c, "user_ids 不能为空")
		return
	}

	if err := h.service.BatchUpdateUserStatus(req.UserIDs, false); err != nil {
		common.InternalServerErrorResponse(c, "批量禁用用户失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, "批量禁用成功", nil)
}

// BatchDeleteUsers 批量删除用户
func (h *AdminHandler) BatchDeleteUsers(c *gin.Context) {
	var req struct {
		UserIDs []uint `json:"user_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequestResponse(c, "参数错误: "+err.Error())
		return
	}

	if len(req.UserIDs) == 0 {
		common.BadRequestResponse(c, "user_ids 不能为空")
		return
	}

	if err := h.service.BatchDeleteUsers(req.UserIDs); err != nil {
		common.InternalServerErrorResponse(c, "批量删除用户失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, "批量删除成功", nil)
}
