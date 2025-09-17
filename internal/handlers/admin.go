package handlers

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/zy84338719/filecodebox/internal/common"
	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/models"
	"github.com/zy84338719/filecodebox/internal/models/web"
	"github.com/zy84338719/filecodebox/internal/services"

	"github.com/gin-gonic/gin"
)

// AdminHandler 管理处理器
type AdminHandler struct {
	service *services.AdminService
	config  *config.ConfigManager
}

func NewAdminHandler(service *services.AdminService, config *config.ConfigManager) *AdminHandler {
	return &AdminHandler{
		service: service,
		config:  config,
	}
}

// Login 管理员登录
func (h *AdminHandler) Login(c *gin.Context) {
	var req web.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequestResponse(c, "参数错误: "+err.Error())
		return
	}

	// 使用 AdminService 进行管理员凭据验证并生成 token
	tokenString, err := h.service.GenerateTokenForAdmin(req.Username, req.Password)
	if err != nil {
		common.UnauthorizedResponse(c, "认证失败: "+err.Error())
		return
	}

	response := web.AdminLoginResponse{
		Token:     tokenString,
		TokenType: "Bearer",
		ExpiresIn: 24 * 60 * 60, // 24小时，单位秒
	}

	common.SuccessWithMessage(c, "登录成功", response)
}

// Dashboard 仪表盘
func (h *AdminHandler) Dashboard(c *gin.Context) {
	stats, err := h.service.GetStats()
	if err != nil {
		common.InternalServerErrorResponse(c, "获取统计信息失败: "+err.Error())
		return
	}

	common.SuccessResponse(c, stats)
}

// GetStats 获取统计信息
func (h *AdminHandler) GetStats(c *gin.Context) {
	stats, err := h.service.GetStats()
	if err != nil {
		common.InternalServerErrorResponse(c, "获取统计信息失败: "+err.Error())
		return
	}

	common.SuccessResponse(c, stats)
}

// GetFiles 获取文件列表
func (h *AdminHandler) GetFiles(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")
	search := c.Query("search")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	files, total, err := h.service.GetFiles(page, pageSize, search)
	if err != nil {
		common.InternalServerErrorResponse(c, "获取文件列表失败: "+err.Error())
		return
	}

	common.SuccessWithPagination(c, files, int(total), page, pageSize)
}

// DeleteFile 删除文件
func (h *AdminHandler) DeleteFile(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		common.BadRequestResponse(c, "文件代码不能为空")
		return
	}

	err := h.service.DeleteFileByCode(code)
	if err != nil {
		common.InternalServerErrorResponse(c, "删除失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, "删除成功", nil)
}

// GetFile 获取单个文件信息
func (h *AdminHandler) GetFile(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		common.BadRequestResponse(c, "文件代码不能为空")
		return
	}

	fileCode, err := h.service.GetFileByCode(code)
	if err != nil {
		common.NotFoundResponse(c, "文件不存在")
		return
	}

	common.SuccessResponse(c, fileCode)
}

// GetConfig 获取配置
func (h *AdminHandler) GetConfig(c *gin.Context) {
	config := h.service.GetFullConfig()
	common.SuccessResponse(c, config)
}

// UpdateConfig 更新配置
func (h *AdminHandler) UpdateConfig(c *gin.Context) {

	// 绑定为 AdminConfigRequest 并使用服务层处理（服务会构建 map 并持久化）
	var req web.AdminConfigRequest
	if err := c.ShouldBind(&req); err != nil {
		common.BadRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}

	if err := h.service.UpdateConfigFromRequest(&req); err != nil {
		common.InternalServerErrorResponse(c, "更新配置失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, "更新成功", nil)
}

// CleanExpiredFiles 清理过期文件
func (h *AdminHandler) CleanExpiredFiles(c *gin.Context) {
	// TODO: 修复服务方法调用
	//count, err := h.service.CleanupExpiredFiles()
	//if err != nil {
	//	common.InternalServerErrorResponse(c, "清理失败: "+err.Error())
	//	return
	//}

	// 临时返回
	common.SuccessWithMessage(c, "清理完成", web.CleanedCountResponse{CleanedCount: 0})
}

// CleanTempFiles 清理临时文件
func (h *AdminHandler) CleanTempFiles(c *gin.Context) {
	count, err := h.service.CleanTempFiles()
	if err != nil {
		common.InternalServerErrorResponse(c, "清理失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, fmt.Sprintf("清理了 %d 个临时文件", count), web.CountResponse{Count: count})
}

// CleanInvalidRecords 清理无效记录
func (h *AdminHandler) CleanInvalidRecords(c *gin.Context) {
	count, err := h.service.CleanInvalidRecords()
	if err != nil {
		common.InternalServerErrorResponse(c, "清理失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, fmt.Sprintf("清理了 %d 个无效记录", count), web.CountResponse{Count: count})
}

// OptimizeDatabase 优化数据库
func (h *AdminHandler) OptimizeDatabase(c *gin.Context) {
	err := h.service.OptimizeDatabase()
	if err != nil {
		common.InternalServerErrorResponse(c, "优化失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, "数据库优化完成", nil)
}

// AnalyzeDatabase 分析数据库
func (h *AdminHandler) AnalyzeDatabase(c *gin.Context) {
	stats, err := h.service.AnalyzeDatabase()
	if err != nil {
		common.InternalServerErrorResponse(c, "分析失败: "+err.Error())
		return
	}

	common.SuccessResponse(c, stats)
}

// BackupDatabase 备份数据库
func (h *AdminHandler) BackupDatabase(c *gin.Context) {
	backupPath, err := h.service.BackupDatabase()
	if err != nil {
		common.InternalServerErrorResponse(c, "备份失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, "数据库备份完成", web.BackupPathResponse{BackupPath: backupPath})
}

// ClearSystemCache 清理系统缓存
func (h *AdminHandler) ClearSystemCache(c *gin.Context) {
	err := h.service.ClearSystemCache()
	if err != nil {
		common.InternalServerErrorResponse(c, "清理失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, "系统缓存清理完成", nil)
}

// ClearUploadCache 清理上传缓存
func (h *AdminHandler) ClearUploadCache(c *gin.Context) {
	err := h.service.ClearUploadCache()
	if err != nil {
		common.InternalServerErrorResponse(c, "清理失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, "上传缓存清理完成", nil)
}

// ClearDownloadCache 清理下载缓存
func (h *AdminHandler) ClearDownloadCache(c *gin.Context) {
	err := h.service.ClearDownloadCache()
	if err != nil {
		common.InternalServerErrorResponse(c, "清理失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, "下载缓存清理完成", nil)
}

// GetSystemInfo 获取系统信息
func (h *AdminHandler) GetSystemInfo(c *gin.Context) {
	info, err := h.service.GetSystemInfo()
	if err != nil {
		common.InternalServerErrorResponse(c, "获取系统信息失败: "+err.Error())
		return
	}

	common.SuccessResponse(c, info)
}

// GetStorageStatus 获取存储状态
func (h *AdminHandler) GetStorageStatus(c *gin.Context) {
	status, err := h.service.GetStorageStatus()
	if err != nil {
		common.InternalServerErrorResponse(c, "获取存储状态失败: "+err.Error())
		return
	}

	common.SuccessResponse(c, status)
}

// GetPerformanceMetrics 获取性能指标
func (h *AdminHandler) GetPerformanceMetrics(c *gin.Context) {
	metrics, err := h.service.GetPerformanceMetrics()
	if err != nil {
		common.InternalServerErrorResponse(c, "获取性能指标失败: "+err.Error())
		return
	}

	common.SuccessResponse(c, metrics)
}

// ScanSecurity 安全扫描
func (h *AdminHandler) ScanSecurity(c *gin.Context) {
	result, err := h.service.ScanSecurity()
	if err != nil {
		common.InternalServerErrorResponse(c, "安全扫描失败: "+err.Error())
		return
	}

	common.SuccessResponse(c, result)
}

// CheckPermissions 检查权限
func (h *AdminHandler) CheckPermissions(c *gin.Context) {
	result, err := h.service.CheckPermissions()
	if err != nil {
		common.InternalServerErrorResponse(c, "权限检查失败: "+err.Error())
		return
	}

	common.SuccessResponse(c, result)
}

// CheckIntegrity 检查完整性
func (h *AdminHandler) CheckIntegrity(c *gin.Context) {
	result, err := h.service.CheckIntegrity()
	if err != nil {
		common.InternalServerErrorResponse(c, "完整性检查失败: "+err.Error())
		return
	}

	common.SuccessResponse(c, result)
}

// ClearSystemLogs 清理系统日志
func (h *AdminHandler) ClearSystemLogs(c *gin.Context) {
	count, err := h.service.ClearSystemLogs()
	if err != nil {
		common.InternalServerErrorResponse(c, "清理失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, fmt.Sprintf("清理了 %d 条系统日志", count), web.CountResponse{Count: count})
}

// ClearAccessLogs 清理访问日志
func (h *AdminHandler) ClearAccessLogs(c *gin.Context) {
	count, err := h.service.ClearAccessLogs()
	if err != nil {
		common.InternalServerErrorResponse(c, "清理失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, fmt.Sprintf("清理了 %d 条访问日志", count), web.CountResponse{Count: count})
}

// ClearErrorLogs 清理错误日志
func (h *AdminHandler) ClearErrorLogs(c *gin.Context) {
	count, err := h.service.ClearErrorLogs()
	if err != nil {
		common.InternalServerErrorResponse(c, "清理失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, fmt.Sprintf("清理了 %d 条错误日志", count), web.CountResponse{Count: count})
}

// ExportLogs 导出日志
func (h *AdminHandler) ExportLogs(c *gin.Context) {
	logType := c.DefaultQuery("type", "system")

	logPath, err := h.service.ExportLogs(logType)
	if err != nil {
		common.InternalServerErrorResponse(c, "导出失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, "日志导出完成", web.LogPathResponse{LogPath: logPath})
}

// GetLogStats 获取日志统计
func (h *AdminHandler) GetLogStats(c *gin.Context) {
	stats, err := h.service.GetLogStats()
	if err != nil {
		common.InternalServerErrorResponse(c, "获取日志统计失败: "+err.Error())
		return
	}

	common.SuccessResponse(c, stats)
}

// UpdateFile 更新文件信息
func (h *AdminHandler) UpdateFile(c *gin.Context) {
	// 从URL参数获取code
	code := c.Param("code")
	if code == "" {
		common.BadRequestResponse(c, "文件代码不能为空")
		return
	}

	var updateData struct {
		Code         string     `json:"code"`
		Text         string     `json:"text"`
		ExpiredAt    *time.Time `json:"expired_at"`
		ExpiredCount *int       `json:"expired_count"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		common.BadRequestResponse(c, "参数错误: "+err.Error())
		return
	}

	// 获取现有文件信息
	_, err := h.service.GetFileByCode(code)
	if err != nil {
		common.NotFoundResponse(c, "文件不存在")
		return
	}

	// 更新字段
	var expiredAt *time.Time
	if updateData.ExpiredAt != nil {
		expiredAt = updateData.ExpiredAt
	}

	// 保存更新 - 使用UpdateFileByCode方法
	var expTime time.Time
	if expiredAt != nil {
		expTime = *expiredAt
	}
	err = h.service.UpdateFileByCode(code, updateData.Code, "", expTime)
	if err != nil {
		common.InternalServerErrorResponse(c, "更新失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, "更新成功", nil)
}

// DownloadFile 下载文件（管理员）
func (h *AdminHandler) DownloadFile(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		common.BadRequestResponse(c, "文件代码不能为空")
		return
	}

	fileCode, err := h.service.GetFileByCode(code)
	if err != nil {
		common.NotFoundResponse(c, "文件不存在")
		return
	}

	if fileCode.Text != "" {
		// 文本内容
		fileName := fileCode.Prefix + ".txt"
		c.Header("Content-Disposition", "attachment; filename=\""+fileName+"\"")
		c.Header("Content-Type", "text/plain")
		c.String(200, fileCode.Text)
		return
	}

	// 文件下载 - 通过存储管理器
	filePath := fileCode.GetFilePath()
	if filePath == "" {
		common.NotFoundResponse(c, "文件路径为空")
		return
	}

	err = h.service.ServeFile(c, fileCode)
	if err != nil {
		common.InternalServerErrorResponse(c, "文件下载失败: "+err.Error())
		return
	}
}

// ========== 用户管理相关方法 ==========

// GetUsers 获取用户列表
func (h *AdminHandler) GetUsers(c *gin.Context) {
	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 从数据库获取用户列表
	users, total, err := h.getUsersFromDB(page, pageSize, search)
	if err != nil {
		common.InternalServerErrorResponse(c, "获取用户列表失败: "+err.Error())
		return
	}

	// 计算统计信息
	stats, err := h.getUserStats()
	if err != nil {
		// 如果统计失败，使用默认值但不阻止接口
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

// getUsersFromDB 从数据库获取用户列表
func (h *AdminHandler) getUsersFromDB(page, pageSize int, search string) ([]web.AdminUserDetail, int64, error) {
	users, total, err := h.service.GetUsers(page, pageSize, search)
	if err != nil {
		return nil, 0, err
	}

	// 转换为返回格式
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

// getUserStats 获取用户统计信息
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
	userIDStr := c.Param("id")
	userID64, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		common.BadRequestResponse(c, "用户ID错误")
		return
	}

	user, err := h.service.GetUserByID(uint(userID64))
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
	if err := c.ShouldBindJSON(&userData); err != nil {
		common.BadRequestResponse(c, "参数错误: "+err.Error())
		return
	}

	// 准备参数
	role := "user"
	if userData.IsAdmin {
		role = "admin"
	}

	status := "active"
	if !userData.IsActive {
		status = "inactive"
	}

	// 创建用户
	user, err := h.service.CreateUser(userData.Username, userData.Email, userData.Password, userData.Nickname, role, status)
	if err != nil {
		common.InternalServerErrorResponse(c, "创建用户失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, "用户创建成功", web.IDResponse{
		ID: user.ID,
	})
}

// UpdateUser 更新用户
func (h *AdminHandler) UpdateUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID64, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		common.BadRequestResponse(c, "用户ID错误")
		return
	}

	var userData struct {
		Email    string `json:"email" binding:"omitempty,email"`
		Password string `json:"password"`
		Nickname string `json:"nickname"`
		IsAdmin  bool   `json:"is_admin"`
		IsActive bool   `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&userData); err != nil {
		common.BadRequestResponse(c, "参数错误: "+err.Error())
		return
	}

	// 准备参数
	role := "user"
	if userData.IsAdmin {
		role = "admin"
	}

	status := "active"
	if !userData.IsActive {
		status = "inactive"
	}

	// 更新用户：构建 models.User 并调用服务方法
	// 构建更新用的 models.User：只填充仓库 UpdateUserFields 所需字段
	user := models.User{}
	// ID will be used by UpdateUserFields via repository; ensure repository method uses provided ID
	// NOTE: models.User uses gorm.Model embed; set via zero-value and pass id to repository
	user.Email = userData.Email
	if userData.Password != "" {
		// Hashing handled inside service layer; here we pass raw password in a convention used elsewhere
		user.PasswordHash = userData.Password
	}
	user.Nickname = userData.Nickname
	user.Role = role
	user.Status = status

	err = h.service.UpdateUser(user)
	if err != nil {
		common.InternalServerErrorResponse(c, "更新用户失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, "用户更新成功", web.IDResponse{
		ID: uint(userID64),
	})
}

// DeleteUser 删除用户
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID64, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		common.BadRequestResponse(c, "用户ID错误")
		return
	}

	userID := uint(userID64)

	// 删除用户
	err = h.service.DeleteUser(userID)
	if err != nil {
		common.InternalServerErrorResponse(c, "删除用户失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, "用户删除成功", web.IDResponse{
		ID: userID,
	})
}

// UpdateUserStatus 更新用户状态
func (h *AdminHandler) UpdateUserStatus(c *gin.Context) {
	userIDStr := c.Param("id")
	userID64, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		common.BadRequestResponse(c, "用户ID错误")
		return
	}

	var statusData struct {
		IsActive bool `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&statusData); err != nil {
		common.BadRequestResponse(c, "参数错误: "+err.Error())
		return
	}

	userID := uint(userID64)

	// 更新用户状态
	err = h.service.UpdateUserStatus(userID, statusData.IsActive)
	if err != nil {
		common.InternalServerErrorResponse(c, "更新用户状态失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, "用户状态更新成功", web.IDResponse{
		ID: userID,
	})
}

// GetUserFiles 获取用户文件
func (h *AdminHandler) GetUserFiles(c *gin.Context) {
	userIDStr := c.Param("id")
	userID64, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		common.BadRequestResponse(c, "用户ID错误")
		return
	}

	userID := uint(userID64)

	// 获取用户信息
	user, err := h.service.GetUserByID(userID)
	if err != nil {
		common.NotFoundResponse(c, "用户不存在")
		return
	}

	// 获取用户的文件列表
	page := 1   // 默认第一页
	limit := 20 // 默认每页20条

	files, total, err := h.service.GetUserFiles(userID, page, limit)
	if err != nil {
		common.InternalServerErrorResponse(c, "获取用户文件失败: "+err.Error())
		return
	}

	// 转换为返回格式
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
	// 获取所有用户数据
	users, _, err := h.getUsersFromDB(1, 10000, "") // 获取大量用户数据用于导出
	if err != nil {
		common.InternalServerErrorResponse(c, "获取用户数据失败: "+err.Error())
		return
	}

	// 生成CSV内容
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
			float64(user.TotalStorage)/(1024*1024), // 转换为MB
		)
	}

	// 添加UTF-8 BOM以确保Excel正确显示中文
	bomContent := "\xEF\xBB\xBF" + csvContent

	// 设置响应头（Content-Length 使用实际发送的字节长度）
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=users_export.csv")
	c.Header("Content-Length", strconv.Itoa(len([]byte(bomContent))))

	// 使用 Write 写入原始字节，避免框架对长度的二次处理
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

// GetMCPConfig 获取 MCP 配置
func (h *AdminHandler) GetMCPConfig(c *gin.Context) {
	mcpConfig := map[string]interface{}{
		"enable_mcp_server": h.config.MCP.EnableMCPServer,
		"mcp_port":          h.config.MCP.MCPPort,
		"mcp_host":          h.config.MCP.MCPHost,
	}

	common.SuccessResponse(c, mcpConfig)
}

// UpdateMCPConfig 更新 MCP 配置
func (h *AdminHandler) UpdateMCPConfig(c *gin.Context) {
	var mcpConfig struct {
		EnableMCPServer *int    `json:"enable_mcp_server"`
		MCPPort         *string `json:"mcp_port"`
		MCPHost         *string `json:"mcp_host"`
	}

	if err := c.ShouldBindJSON(&mcpConfig); err != nil {
		common.BadRequestResponse(c, "MCP配置参数错误: "+err.Error())
		return
	}

	// 构建配置更新映射
	configUpdates := make(map[string]interface{})

	if mcpConfig.EnableMCPServer != nil {
		configUpdates["enable_mcp_server"] = *mcpConfig.EnableMCPServer
	}
	if mcpConfig.MCPPort != nil {
		configUpdates["mcp_port"] = *mcpConfig.MCPPort
	}
	if mcpConfig.MCPHost != nil {
		configUpdates["mcp_host"] = *mcpConfig.MCPHost
	}

	// 更新配置
	err := h.service.UpdateConfig(configUpdates)
	if err != nil {
		common.InternalServerErrorResponse(c, "更新MCP配置失败: "+err.Error())
		return
	}

	// 重新加载配置（从 config.yaml 与环境变量）
	err = h.config.ReloadConfig()
	if err != nil {
		common.InternalServerErrorResponse(c, "重新加载配置失败: "+err.Error())
		return
	}

	// 应用MCP配置更改
	mcpManager := GetMCPManager()
	if mcpManager != nil {
		enableMCP := h.config.MCP.EnableMCPServer == 1
		port := h.config.MCP.MCPPort

		err = mcpManager.ApplyConfig(enableMCP, port)
		if err != nil {
			common.InternalServerErrorResponse(c, "应用MCP配置失败: "+err.Error())
			return
		}
	}

	common.SuccessWithMessage(c, "MCP配置更新成功", nil)
}

// GetMCPStatus 获取 MCP 服务器状态
func (h *AdminHandler) GetMCPStatus(c *gin.Context) {
	mcpManager := GetMCPManager()
	if mcpManager == nil {
		common.InternalServerErrorResponse(c, "MCP管理器未初始化")
		return
	}

	status := mcpManager.GetStatus()
	status["config"] = map[string]interface{}{
		"enabled": h.config.MCP.EnableMCPServer == 1,
		"port":    h.config.MCP.MCPPort,
		"host":    h.config.MCP.MCPHost,
	}

	common.SuccessResponse(c, status)
}

// RestartMCPServer 重启 MCP 服务器
func (h *AdminHandler) RestartMCPServer(c *gin.Context) {
	mcpManager := GetMCPManager()
	if mcpManager == nil {
		common.InternalServerErrorResponse(c, "MCP管理器未初始化")
		return
	}

	// 检查是否启用了MCP服务器
	if h.config.MCP.EnableMCPServer != 1 {
		common.BadRequestResponse(c, "MCP服务器未启用")
		return
	}

	err := mcpManager.RestartMCPServer(h.config.MCP.MCPPort)
	if err != nil {
		common.InternalServerErrorResponse(c, "重启MCP服务器失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, "MCP服务器重启成功", nil)
}

// ControlMCPServer 控制 MCP 服务器（启动/停止）
func (h *AdminHandler) ControlMCPServer(c *gin.Context) {
	var controlData struct {
		Action string `json:"action" binding:"required"` // "start" 或 "stop"
	}

	if err := c.ShouldBindJSON(&controlData); err != nil {
		common.BadRequestResponse(c, "参数错误: "+err.Error())
		return
	}

	mcpManager := GetMCPManager()
	if mcpManager == nil {
		common.InternalServerErrorResponse(c, "MCP管理器未初始化")
		return
	}

	switch controlData.Action {
	case "start":
		if h.config.MCP.EnableMCPServer != 1 {
			common.BadRequestResponse(c, "MCP服务器未启用，请先在配置中启用")
			return
		}
		err := mcpManager.StartMCPServer(h.config.MCP.MCPPort)
		if err != nil {
			common.InternalServerErrorResponse(c, "启动MCP服务器失败: "+err.Error())
			return
		}
		common.SuccessWithMessage(c, "MCP服务器启动成功", nil)

	case "stop":
		err := mcpManager.StopMCPServer()
		if err != nil {
			common.InternalServerErrorResponse(c, "停止MCP服务器失败: "+err.Error())
			return
		}
		common.SuccessWithMessage(c, "MCP服务器停止成功", nil)

	default:
		common.BadRequestResponse(c, "无效的操作，只支持 start 或 stop")
	}
}

// TestMCPConnection 测试 MCP 连接
func (h *AdminHandler) TestMCPConnection(c *gin.Context) {
	var testData struct {
		Port string `json:"port"`
		Host string `json:"host"`
	}

	if err := c.ShouldBindJSON(&testData); err != nil {
		common.BadRequestResponse(c, "参数错误: "+err.Error())
		return
	}

	// 使用提供的端口或默认配置
	port := testData.Port
	if port == "" {
		port = h.config.MCP.MCPPort
	}
	if port == "" {
		port = "8081"
	}

	host := testData.Host
	if host == "" {
		host = h.config.MCP.MCPHost
	}
	if host == "" {
		host = "0.0.0.0"
	}

	// 进行简单的端口连通性测试
	address := host + ":" + port
	if host == "0.0.0.0" {
		address = "127.0.0.1:" + port
	}

	// 尝试连接端口
	conn, err := net.DialTimeout("tcp", address, time.Second*3)
	if err != nil {
		// 端口未开放或连接失败
		common.ErrorResponse(c, 400, fmt.Sprintf("连接测试失败: %s，端口可能未开放或MCP服务器未启动", err.Error()))
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("关闭连接失败: %v", err)
		}
	}()

	common.SuccessWithMessage(c, "MCP连接测试成功", map[string]interface{}{
		"address": address,
		"status":  "连接正常",
		"port":    port,
		"host":    host,
	})
}

// GetSystemLogs 获取系统日志
func (h *AdminHandler) GetSystemLogs(c *gin.Context) {
	level := c.DefaultQuery("level", "")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	logs, err := h.service.GetSystemLogs(limit)
	if err != nil {
		common.InternalServerErrorResponse(c, "获取日志失败: "+err.Error())
		return
	}

	// 如果指定了日志级别，则过滤日志
	if level != "" {
		filteredLogs := make([]string, 0)
		for _, log := range logs {
			// 简单的级别过滤，实际应该解析日志格式
			if len(log) > 0 && (level == "" || len(log) > 10) {
				filteredLogs = append(filteredLogs, log)
			}
		}
		logs = filteredLogs
	}

	common.SuccessResponse(c, map[string]interface{}{
		"logs":  logs,
		"total": len(logs),
	})
}

// GetRunningTasks 获取运行中的任务
func (h *AdminHandler) GetRunningTasks(c *gin.Context) {
	tasks, err := h.service.GetRunningTasks()
	if err != nil {
		common.InternalServerErrorResponse(c, "获取运行任务失败: "+err.Error())
		return
	}

	common.SuccessResponse(c, map[string]interface{}{
		"tasks": tasks,
		"total": len(tasks),
	})
}

// CancelTask 取消任务
func (h *AdminHandler) CancelTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		common.BadRequestResponse(c, "任务ID不能为空")
		return
	}

	err := h.service.CancelTask(taskID)
	if err != nil {
		common.InternalServerErrorResponse(c, "取消任务失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, "任务已取消", nil)
}

// RetryTask 重试任务
func (h *AdminHandler) RetryTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		common.BadRequestResponse(c, "任务ID不能为空")
		return
	}

	err := h.service.RetryTask(taskID)
	if err != nil {
		common.InternalServerErrorResponse(c, "重试任务失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, "任务已重新启动", nil)
}

// RestartSystem 重启系统
func (h *AdminHandler) RestartSystem(c *gin.Context) {
	err := h.service.RestartSystem()
	if err != nil {
		common.InternalServerErrorResponse(c, "重启系统失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, "系统重启指令已发送", nil)
}
