package handlers

import (
	"strconv"
	"time"

	"github.com/zy84338719/filecodebox/internal/common"
	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// AdminHandler 管理处理器
type AdminHandler struct {
	service *services.AdminService
	config  *config.Config
}

func NewAdminHandler(service *services.AdminService, config *config.Config) *AdminHandler {
	return &AdminHandler{
		service: service,
		config:  config,
	}
}

// Login 管理员登录
func (h *AdminHandler) Login(c *gin.Context) {
	var loginData struct {
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&loginData); err != nil {
		common.BadRequestResponse(c, "参数错误: "+err.Error())
		return
	}

	// 验证密码
	if loginData.Password != h.config.AdminToken {
		common.UnauthorizedResponse(c, "密码错误")
		return
	}

	// 生成JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"is_admin": true,
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // 24小时过期
	})

	tokenString, err := token.SignedString([]byte(h.config.AdminToken))
	if err != nil {
		common.InternalServerErrorResponse(c, "生成token失败")
		return
	}

	common.SuccessWithToken(c, tokenString, gin.H{
		"token_type": "Bearer",
	})
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
	config, err := h.service.GetConfig()
	if err != nil {
		common.InternalServerErrorResponse(c, "获取配置失败: "+err.Error())
		return
	}

	common.SuccessResponse(c, config)
}

// UpdateConfig 更新配置
func (h *AdminHandler) UpdateConfig(c *gin.Context) {
	var newConfig map[string]interface{}
	if err := c.ShouldBindJSON(&newConfig); err != nil {
		common.BadRequestResponse(c, "配置参数错误: "+err.Error())
		return
	}

	err := h.service.UpdateConfig(newConfig)
	if err != nil {
		common.InternalServerErrorResponse(c, "更新配置失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, "更新成功", nil)
}

// CleanExpiredFiles 清理过期文件
func (h *AdminHandler) CleanExpiredFiles(c *gin.Context) {
	count, err := h.service.CleanExpiredFiles()
	if err != nil {
		common.InternalServerErrorResponse(c, "清理失败: "+err.Error())
		return
	}

	common.SuccessWithCleanedCount(c, count)
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
		stats = gin.H{
			"total_users":         total,
			"active_users":        total,
			"today_registrations": 0,
			"today_uploads":       0,
		}
	}

	pagination := gin.H{
		"page":      page,
		"page_size": pageSize,
		"total":     total,
		"pages":     (total + int64(pageSize) - 1) / int64(pageSize),
	}

	common.SuccessResponse(c, gin.H{
		"users":      users,
		"stats":      stats,
		"pagination": pagination,
	})
}

// getUsersFromDB 从数据库获取用户列表
func (h *AdminHandler) getUsersFromDB(page, pageSize int, search string) ([]gin.H, int64, error) {
	users, total, err := h.service.GetUsers(page, pageSize, search)
	if err != nil {
		return nil, 0, err
	}

	// 转换为返回格式
	result := make([]gin.H, len(users))
	for i, user := range users {
		lastLoginAt := ""
		if user.LastLoginAt != nil {
			lastLoginAt = user.LastLoginAt.Format("2006-01-02 15:04:05")
		}

		result[i] = gin.H{
			"id":              user.ID,
			"username":        user.Username,
			"email":           user.Email,
			"nickname":        user.Nickname,
			"role":            user.Role,
			"is_admin":        user.Role == "admin",
			"is_active":       user.Status == "active",
			"status":          user.Status,
			"email_verified":  user.EmailVerified,
			"created_at":      user.CreatedAt.Format("2006-01-02 15:04:05"),
			"last_login_at":   lastLoginAt,
			"last_login_ip":   user.LastLoginIP,
			"total_uploads":   user.TotalUploads,
			"total_downloads": user.TotalDownloads,
			"total_storage":   user.TotalStorage,
		}
	}

	return result, total, nil
}

// getUserStats 获取用户统计信息
func (h *AdminHandler) getUserStats() (gin.H, error) {
	stats, err := h.service.GetStats()
	if err != nil {
		return nil, err
	}

	return gin.H{
		"total_users":         stats["total_users"],
		"active_users":        stats["active_users"],
		"today_registrations": stats["today_registrations"],
		"today_uploads":       stats["today_uploads"],
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

	userDetail := gin.H{
		"id":              user.ID,
		"username":        user.Username,
		"email":           user.Email,
		"nickname":        user.Nickname,
		"role":            user.Role,
		"is_admin":        user.Role == "admin",
		"is_active":       user.Status == "active",
		"status":          user.Status,
		"email_verified":  user.EmailVerified,
		"created_at":      user.CreatedAt.Format("2006-01-02 15:04:05"),
		"last_login_at":   lastLoginAt,
		"last_login_ip":   user.LastLoginIP,
		"total_uploads":   user.TotalUploads,
		"total_downloads": user.TotalDownloads,
		"total_storage":   user.TotalStorage,
	}

	common.SuccessResponse(c, userDetail)
}

// CreateUser 创建用户
func (h *AdminHandler) CreateUser(c *gin.Context) {
	var userData struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
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

	// 创建用户
	user, err := h.service.CreateUser(userData.Username, userData.Email, userData.Password, userData.Nickname, role, status)
	if err != nil {
		common.InternalServerErrorResponse(c, "创建用户失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, "用户创建成功", gin.H{
		"id": user.ID,
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
		Email    string `json:"email" binding:"required,email"`
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

	// 更新用户
	err = h.service.UpdateUser(uint(userID64), userData.Email, userData.Password, userData.Nickname, role, status)
	if err != nil {
		common.InternalServerErrorResponse(c, "更新用户失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, "用户更新成功", gin.H{
		"id": uint(userID64),
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

	common.SuccessWithMessage(c, "用户删除成功", gin.H{
		"id": userID,
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

	common.SuccessWithMessage(c, "用户状态更新成功", gin.H{
		"id": userID,
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
	fileList := make([]gin.H, len(files))
	for i, file := range files {
		expiredAt := ""
		if file.ExpiredAt != nil {
			expiredAt = file.ExpiredAt.Format("2006-01-02 15:04:05")
		}

		fileType := "文件"
		if file.Text != "" {
			fileType = "文本"
		}

		fileList[i] = gin.H{
			"id":            file.ID,
			"code":          file.Code,
			"prefix":        file.Prefix,
			"suffix":        file.Suffix,
			"size":          file.Size,
			"type":          fileType,
			"expired_at":    expiredAt,
			"expired_count": file.ExpiredCount,
			"used_count":    file.UsedCount,
			"created_at":    file.CreatedAt.Format("2006-01-02 15:04:05"),
			"require_auth":  file.RequireAuth,
			"upload_type":   file.UploadType,
		}
	}

	common.SuccessResponse(c, gin.H{
		"files":    fileList,
		"username": user.Username,
		"total":    total,
	})
}

// GetMCPConfig 获取 MCP 配置
func (h *AdminHandler) GetMCPConfig(c *gin.Context) {
	mcpConfig := map[string]interface{}{
		"enable_mcp_server": h.config.EnableMCPServer,
		"mcp_port":          h.config.MCPPort,
		"mcp_host":          h.config.MCPHost,
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

	// 重新加载配置
	err = h.config.LoadFromDatabase()
	if err != nil {
		common.InternalServerErrorResponse(c, "重新加载配置失败: "+err.Error())
		return
	}

	// 应用MCP配置更改
	mcpManager := GetMCPManager()
	if mcpManager != nil {
		enableMCP := h.config.EnableMCPServer == 1
		port := h.config.MCPPort

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
		"enabled": h.config.EnableMCPServer == 1,
		"port":    h.config.MCPPort,
		"host":    h.config.MCPHost,
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
	if h.config.EnableMCPServer != 1 {
		common.BadRequestResponse(c, "MCP服务器未启用")
		return
	}

	err := mcpManager.RestartMCPServer(h.config.MCPPort)
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
		if h.config.EnableMCPServer != 1 {
			common.BadRequestResponse(c, "MCP服务器未启用，请先在配置中启用")
			return
		}
		err := mcpManager.StartMCPServer(h.config.MCPPort)
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
