package handlers

import (
	"net/http"
	"strconv"

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
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "用户系统未启用",
		})
		return
	}

	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
		Nickname string `json:"nickname"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 验证输入
	if err := h.userService.ValidateUserInput(req.Username, req.Email, req.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
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
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "注册成功",
		"detail": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"nickname": user.Nickname,
			"role":     user.Role,
		},
	})
}

// Login 用户登录
func (h *UserHandler) Login(c *gin.Context) {
	if !h.userService.IsUserSystemEnabled() {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "用户系统未启用",
		})
		return
	}

	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 获取客户端信息
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// 用户登录
	token, user, err := h.userService.Login(req.Username, req.Password, ipAddress, userAgent)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "登录成功",
		"detail": gin.H{
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
		},
	})
}

// Logout 用户登出
func (h *UserHandler) Logout(c *gin.Context) {
	// 从上下文获取会话ID
	sessionID, exists := c.Get("session_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "会话信息不存在",
		})
		return
	}

	if err := h.userService.Logout(sessionID.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "登出失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "登出成功",
	})
}

// GetProfile 获取用户资料
func (h *UserHandler) GetProfile(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "用户未登录",
		})
		return
	}

	user, err := h.userService.GetUserByID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "用户不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"detail": gin.H{
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
		},
	})
}

// UpdateProfile 更新用户资料
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "用户未登录",
		})
		return
	}

	var req struct {
		Nickname string `json:"nickname"`
		Avatar   string `json:"avatar"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	if err := h.userService.UpdateUserProfile(userID.(uint), req.Nickname, req.Avatar); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
	})
}

// ChangePassword 修改密码
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "用户未登录",
		})
		return
	}

	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	if len(req.NewPassword) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "新密码长度至少6个字符",
		})
		return
	}

	if err := h.userService.ChangePassword(userID.(uint), req.OldPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "密码修改成功",
	})
}

// GetUserFiles 获取用户文件列表
func (h *UserHandler) GetUserFiles(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "用户未登录",
		})
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取文件列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"detail": gin.H{
			"files": files,
			"pagination": gin.H{
				"page":  page,
				"limit": limit,
				"total": total,
				"pages": (total + int64(limit) - 1) / int64(limit),
			},
		},
	})
}

// GetUserStats 获取用户统计信息
func (h *UserHandler) GetUserStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "用户未登录",
		})
		return
	}

	stats, err := h.userService.GetUserStats(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取统计信息失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"detail":  stats,
	})
}

// CheckAuth 检查认证状态
func (h *UserHandler) CheckAuth(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "用户未登录",
		})
		return
	}

	username, _ := c.Get("username")
	role, _ := c.Get("role")

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "已登录",
		"detail": gin.H{
			"user_id":  userID,
			"username": username,
			"role":     role,
		},
	})
}

// GetSystemInfo 获取用户系统信息
func (h *UserHandler) GetSystemInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"detail": gin.H{
			"enable_user_system":      h.userService.IsUserSystemEnabled(),
			"allow_user_registration": h.userService.IsRegistrationAllowed(),
		},
	})
}

// DeleteFile 删除用户文件
func (h *UserHandler) DeleteFile(c *gin.Context) {
	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "用户未登录",
		})
		return
	}

	// 获取文件ID
	fileIDStr := c.Param("id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "文件ID无效",
		})
		return
	}

	// 删除文件（只允许删除用户自己的文件）
	err = h.userService.DeleteUserFile(userID.(uint), uint(fileID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除文件失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "文件删除成功",
	})
}
