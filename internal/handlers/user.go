package handlers

import (
	"strconv"

	"github.com/zy84338719/filecodebox/internal/common"
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

	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
		Nickname string `json:"nickname"`
	}

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

	common.SuccessWithMessage(c, "注册成功", gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"nickname": user.Nickname,
		"role":     user.Role,
	})
}

// Login 用户登录
func (h *UserHandler) Login(c *gin.Context) {
	if !h.userService.IsUserSystemEnabled() {
		common.ForbiddenResponse(c, "用户系统未启用")
		return
	}

	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

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

	common.SuccessWithMessage(c, "登录成功", gin.H{
		"token":      token,
		"token_type": "Bearer",
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"nickname": user.Nickname,
			"role":     user.Role,
			"avatar":   user.Avatar,
		},
	})
}

// Logout 用户登出
func (h *UserHandler) Logout(c *gin.Context) {
	// 从上下文获取会话ID
	sessionID, exists := c.Get("session_id")
	if !exists {
		common.BadRequestResponse(c, "会话信息不存在")
		return
	}

	if err := h.userService.Logout(sessionID.(string)); err != nil {
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

	common.SuccessResponse(c, gin.H{
		"id":             user.ID,
		"username":       user.Username,
		"email":          user.Email,
		"nickname":       user.Nickname,
		"avatar":         user.Avatar,
		"role":           user.Role,
		"status":         user.Status,
		"email_verified": user.EmailVerified,
		"created_at":     user.CreatedAt,
		"last_login_at":  user.LastLoginAt,
	})
}

// UpdateProfile 更新用户资料
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		common.UnauthorizedResponse(c, "用户未登录")
		return
	}

	var req struct {
		Nickname string `json:"nickname"`
		Avatar   string `json:"avatar"`
	}

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

	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}

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
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	files, total, err := h.userService.GetUserFiles(userID.(uint), page, limit)
	if err != nil {
		common.InternalServerErrorResponse(c, "获取文件列表失败: "+err.Error())
		return
	}

	common.SuccessResponse(c, gin.H{
		"files": files,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
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

	common.SuccessResponse(c, stats)
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

	common.SuccessResponse(c, gin.H{
		"id":             user.ID,
		"username":       user.Username,
		"email":          user.Email,
		"role":           user.Role,
		"status":         user.Status,
		"email_verified": user.EmailVerified,
	})
}

// GetSystemInfo 获取系统信息（公开接口）
func (h *UserHandler) GetSystemInfo(c *gin.Context) {
	common.SuccessResponse(c, gin.H{
		"enable_user_system":      h.userService.IsUserSystemEnabled(),
		"allow_user_registration": h.userService.IsRegistrationAllowed(),
	})
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
