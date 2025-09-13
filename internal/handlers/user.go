package handlers

import (
	"strconv"
	"time"

	"github.com/zy84338719/filecodebox/internal/common"
	"github.com/zy84338719/filecodebox/internal/models"
	"github.com/zy84338719/filecodebox/internal/models/web"
	"github.com/zy84338719/filecodebox/internal/services"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService *services.UserService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// Register 用户注册
func (h *UserHandler) Register(c *gin.Context) {
	if !h.userService.IsUserSystemEnabled() {
		common.ForbiddenResponse(c, "用户系统未启用")
		return
	}

	var req web.AuthRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}

	// 验证输入
	if err := h.userService.ValidateUserInput(req.Username, req.Email, req.Password); err != nil {
		common.BadRequestResponse(c, err.Error())
		return
	}

	// 规范化用户名
	req.Username = h.userService.NormalizeUsername(req.Username)

	// 设置默认昵称
	if req.Nickname == "" {
		req.Nickname = req.Username
	}

	// 注册用户
	user, err := h.userService.Register(req.Username, req.Email, req.Password, req.Nickname)
	if err != nil {
		common.BadRequestResponse(c, err.Error())
		return
	}

	response := web.AuthRegisterResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Nickname: user.Nickname,
		Role:     user.Role,
	}

	common.SuccessWithMessage(c, "注册成功", response)
}

// Login 用户登录
func (h *UserHandler) Login(c *gin.Context) {
	if !h.userService.IsUserSystemEnabled() {
		common.ForbiddenResponse(c, "用户系统未启用")
		return
	}

	var req web.AuthLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}

	// 获取客户端信息
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// 用户登录
	token, user, err := h.userService.Login(req.Username, req.Password, ipAddress, userAgent)
	if err != nil {
		common.UnauthorizedResponse(c, err.Error())
		return
	}

	response := web.AuthLoginResponse{
		User: &web.UserInfo{
			ID:            user.ID,
			Username:      user.Username,
			Email:         user.Email,
			Nickname:      user.Nickname,
			Role:          user.Role,
			Avatar:        user.Avatar,
			Status:        user.Status,
			EmailVerified: user.EmailVerified,
		},
		Token:     token,
		TokenType: "Bearer",
	}

	common.SuccessWithMessage(c, "登录成功", response)
}

// Logout 用户登出
func (h *UserHandler) Logout(c *gin.Context) {
	// 获取用户ID用于登出
	userID, exists := c.Get("user_id")
	if !exists {
		common.UnauthorizedResponse(c, "用户未登录")
		return
	}

	if err := h.userService.Logout(userID.(uint)); err != nil {
		common.InternalServerErrorResponse(c, "登出失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, "登出成功", nil)
}

// GetProfile 获取用户资料
func (h *UserHandler) GetProfile(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		common.UnauthorizedResponse(c, "用户未登录")
		return
	}

	user, err := h.userService.GetUserByID(userID.(uint))
	if err != nil {
		common.NotFoundResponse(c, "用户不存在")
		return
	}

	// 构建用户信息响应
	userInfo := &web.UserInfo{
		ID:            user.ID,
		Username:      user.Username,
		Email:         user.Email,
		Nickname:      user.Nickname,
		Avatar:        user.Avatar,
		Role:          user.Role,
		Status:        user.Status,
		EmailVerified: user.EmailVerified,
	}

	// 添加格式化的日期字段
	if !user.CreatedAt.IsZero() {
		userInfo.CreatedAt = user.CreatedAt.Format(time.RFC3339)
	}
	if user.LastLoginAt != nil && !user.LastLoginAt.IsZero() {
		userInfo.LastLoginAt = user.LastLoginAt.Format(time.RFC3339)
	}

	common.SuccessResponse(c, userInfo)
}

// UpdateProfile 更新用户资料
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		common.UnauthorizedResponse(c, "用户未登录")
		return
	}

	var req web.UserProfileUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}

	if err := h.userService.UpdateUserProfile(userID.(uint), req.Nickname, req.Avatar); err != nil {
		common.InternalServerErrorResponse(c, "更新失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, "更新成功", nil)
}

// ChangePassword 修改密码
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		common.UnauthorizedResponse(c, "用户未登录")
		return
	}

	var req web.UserPasswordChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}

	if err := h.userService.ChangePassword(userID.(uint), req.OldPassword, req.NewPassword); err != nil {
		common.BadRequestResponse(c, err.Error())
		return
	}

	common.SuccessWithMessage(c, "密码修改成功", nil)
}

// GetUserFiles 获取用户文件列表
func (h *UserHandler) GetUserFiles(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		common.UnauthorizedResponse(c, "用户未登录")
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 兼容老版本的limit参数
	if pageSize == 20 {
		if limitParam := c.Query("limit"); limitParam != "" {
			pageSize, _ = strconv.Atoi(limitParam)
		}
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	files, total, err := h.userService.GetUserFiles(userID.(uint), page, pageSize)
	if err != nil {
		common.InternalServerErrorResponse(c, "获取文件列表失败: "+err.Error())
		return
	}

	// 临时处理：因为 userService.GetUserFiles 返回 interface{}，我们需要处理类型转换
	var fileInfos []web.FileInfo

	// 如果是 []models.FileCode 类型，进行转换
	if fileCodes, ok := files.([]models.FileCode); ok {
		fileInfos = web.ConvertFileCodeSliceToFileInfoSlice(fileCodes)
	} else {
		// 如果是其他类型，返回空列表
		fileInfos = []web.FileInfo{}
	}

	totalPages := (int(total) + pageSize - 1) / pageSize
	response := web.UserFilesResponse{
		Files: fileInfos,
		Pagination: web.PaginationInfo{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
			HasNext:    page < totalPages,
			HasPrev:    page > 1,
		},
	}

	common.SuccessResponse(c, response)
}

// GetUserStats 获取用户统计信息
func (h *UserHandler) GetUserStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		common.UnauthorizedResponse(c, "用户未登录")
		return
	}

	stats, err := h.userService.GetUserStats(userID.(uint))
	if err != nil {
		common.InternalServerErrorResponse(c, "获取用户统计信息失败: "+err.Error())
		return
	}

	// 使用转换函数将 service 层数据转换为 web 层响应
	response := web.ConvertUserStatsToWeb(stats)
	common.SuccessResponse(c, response)
}

// CheckAuth 检查用户认证状态
func (h *UserHandler) CheckAuth(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		common.UnauthorizedResponse(c, "用户未登录")
		return
	}

	user, err := h.userService.GetUserByID(userID.(uint))
	if err != nil {
		common.UnauthorizedResponse(c, "用户信息获取失败")
		return
	}

	response := &web.UserInfo{
		ID:            user.ID,
		Username:      user.Username,
		Email:         user.Email,
		Nickname:      user.Nickname,
		Avatar:        user.Avatar,
		Role:          user.Role,
		Status:        user.Status,
		EmailVerified: user.EmailVerified,
	}

	common.SuccessResponse(c, response)
}

// GetSystemInfo 获取系统信息（公开接口）
func (h *UserHandler) GetSystemInfo(c *gin.Context) {
	response := &web.UserSystemInfoResponse{
		UserSystemEnabled:        h.userService.IsUserSystemEnabled(),
		AllowUserRegistration:    h.userService.IsRegistrationAllowed(),
		RequireEmailVerification: false, // 这里需要从配置获取
	}

	common.SuccessResponse(c, response)
}

// CheckSystemInitialization 检查系统初始化状态（公开接口）
func (h *UserHandler) CheckSystemInitialization(c *gin.Context) {
	initialized, err := h.userService.IsSystemInitialized()
	if err != nil {
		common.InternalServerErrorResponse(c, "检查系统初始化状态失败")
		return
	}

	response := map[string]interface{}{
		"initialized": initialized,
	}

	common.SuccessResponse(c, response)
}

// IsSystemInitialized 内部方法，直接返回初始化状态
func (h *UserHandler) IsSystemInitialized() (bool, error) {
	return h.userService.IsSystemInitialized()
}

// DeleteFile 删除用户文件
func (h *UserHandler) DeleteFile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		common.UnauthorizedResponse(c, "用户未登录")
		return
	}

	code := c.Param("code")
	if code == "" {
		common.BadRequestResponse(c, "文件代码不能为空")
		return
	}

	err := h.userService.DeleteUserFileByCode(userID.(uint), code)
	if err != nil {
		if err.Error() == "文件不存在或您没有权限删除该文件" {
			common.NotFoundResponse(c, err.Error())
			return
		}
		common.InternalServerErrorResponse(c, "删除文件失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, "文件删除成功", nil)
}
