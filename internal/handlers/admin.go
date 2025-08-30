package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/models"
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
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 验证密码
	if loginData.Password != h.config.AdminToken {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "密码错误",
		})
		return
	}

	// 生成JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"is_admin": true,
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // 24小时过期
	})

	tokenString, err := token.SignedString([]byte(h.config.AdminToken))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "生成token失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "登录成功",
		"detail": gin.H{
			"token":      tokenString,
			"token_type": "Bearer",
		},
	})
}

// Dashboard 仪表盘
func (h *AdminHandler) Dashboard(c *gin.Context) {
	stats, err := h.service.GetStats()
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

// GetStats 获取统计信息
func (h *AdminHandler) GetStats(c *gin.Context) {
	stats, err := h.service.GetStats()
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
			"files":     files,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// DeleteFile 删除文件
func (h *AdminHandler) DeleteFile(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "文件ID错误",
		})
		return
	}

	err = h.service.DeleteFile(uint(id64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

// GetFile 获取单个文件信息
func (h *AdminHandler) GetFile(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "文件ID错误",
		})
		return
	}

	fileCode, err := h.service.GetFileByID(uint(id64))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "文件不存在",
		})
		return
	}

	c.JSON(http.StatusOK, fileCode)
}

// GetConfig 获取配置
func (h *AdminHandler) GetConfig(c *gin.Context) {
	config, err := h.service.GetConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取配置失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"detail":  config,
	})
}

// UpdateConfig 更新配置
func (h *AdminHandler) UpdateConfig(c *gin.Context) {
	var newConfig map[string]interface{}
	if err := c.ShouldBindJSON(&newConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "配置参数错误: " + err.Error(),
		})
		return
	}

	err := h.service.UpdateConfig(newConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新配置失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
	})
}

// CleanExpiredFiles 清理过期文件
func (h *AdminHandler) CleanExpiredFiles(c *gin.Context) {
	count, err := h.service.CleanExpiredFiles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "清理失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"detail": gin.H{
			"cleaned_count": count,
		},
	})
}

// UpdateFile 更新文件信息
func (h *AdminHandler) UpdateFile(c *gin.Context) {
	// 从URL参数获取ID
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "文件ID错误",
		})
		return
	}

	var updateData struct {
		Code         string     `json:"code"`
		Text         string     `json:"text"`
		ExpiredAt    *time.Time `json:"expired_at"`
		ExpiredCount *int       `json:"expired_count"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 获取现有文件信息
	fileCode, err := h.service.GetFileByID(uint(id64))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "文件不存在",
		})
		return
	}

	// 更新字段
	var expiredAt *time.Time
	if updateData.ExpiredAt != nil {
		expiredAt = updateData.ExpiredAt
	}

	// 保存更新 - 使用现有的UpdateFile方法
	err = h.service.UpdateFile(uint(id64), updateData.Code, fileCode.Prefix, fileCode.Suffix, expiredAt, updateData.ExpiredCount)
	if err != nil {
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

// DownloadFile 下载文件（管理员）
func (h *AdminHandler) DownloadFile(c *gin.Context) {
	idStr := c.Query("id")
	id64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "文件ID错误",
		})
		return
	}

	fileCode, err := h.service.GetFileByID(uint(id64))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "文件不存在",
		})
		return
	}

	if fileCode.Text != "" {
		// 文本内容
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "success",
			"detail":  fileCode.Text,
		})
		return
	}

	// 文件下载
	filePath := fileCode.GetFilePath()
	if filePath == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "文件路径为空",
		})
		return
	}

	fileName := fileCode.Prefix + fileCode.Suffix
	c.Header("Content-Disposition", "attachment; filename=\""+fileName+"\"")
	c.File(filePath)
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取用户列表失败: " + err.Error(),
		})
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

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"detail": gin.H{
			"users":      users,
			"stats":      stats,
			"pagination": pagination,
		},
	})
}

// getUsersFromDB 从数据库获取用户列表
func (h *AdminHandler) getUsersFromDB(page, pageSize int, search string) ([]gin.H, int64, error) {
	var users []models.User
	var total int64

	// 构建查询
	query := h.service.GetDB().Model(&models.User{})

	if search != "" {
		query = query.Where("username LIKE ? OR email LIKE ? OR nickname LIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
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
	var totalUsers int64
	var activeUsers int64
	var todayRegistrations int64
	var todayUploads int64

	db := h.service.GetDB()

	// 总用户数
	if err := db.Model(&models.User{}).Count(&totalUsers).Error; err != nil {
		return nil, err
	}

	// 活跃用户数
	if err := db.Model(&models.User{}).Where("status = ?", "active").Count(&activeUsers).Error; err != nil {
		return nil, err
	}

	// 今日注册数
	today := time.Now().Truncate(24 * time.Hour)
	if err := db.Model(&models.User{}).Where("created_at >= ?", today).Count(&todayRegistrations).Error; err != nil {
		return nil, err
	}

	// 今日上传数
	if err := db.Model(&models.FileCode{}).Where("created_at >= ?", today).Count(&todayUploads).Error; err != nil {
		return nil, err
	}

	return gin.H{
		"total_users":         totalUsers,
		"active_users":        activeUsers,
		"today_registrations": todayRegistrations,
		"today_uploads":       todayUploads,
	}, nil
}

// GetUser 获取单个用户
func (h *AdminHandler) GetUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID64, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "用户ID错误",
		})
		return
	}

	var user models.User
	if err := h.service.GetDB().First(&user, uint(userID64)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "用户不存在",
		})
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

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"detail":  userDetail,
	})
}

// CreateUser 创建用户
func (h *AdminHandler) CreateUser(c *gin.Context) {
	var userData struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
		Nickname string `json:"nickname"`
		IsAdmin  bool   `json:"is_admin"`
		IsActive bool   `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&userData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 检查用户名和邮箱是否已存在
	var existingUser models.User
	if err := h.service.GetDB().Where("username = ? OR email = ?", userData.Username, userData.Email).First(&existingUser).Error; err == nil {
		if existingUser.Username == userData.Username {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "用户名已存在",
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "邮箱已存在",
		})
		return
	}

	// 创建新用户
	role := "user"
	if userData.IsAdmin {
		role = "admin"
	}

	status := "active"
	if !userData.IsActive {
		status = "inactive"
	}

	// 这里需要密码哈希，暂时使用原密码（在实际环境中应该使用bcrypt）
	user := models.User{
		Username:     userData.Username,
		Email:        userData.Email,
		PasswordHash: userData.Password, // 在实际环境中应该进行哈希
		Nickname:     userData.Nickname,
		Role:         role,
		Status:       status,
	}

	if err := h.service.GetDB().Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建用户失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "用户创建成功",
		"detail": gin.H{
			"id": user.ID,
		},
	})
}

// UpdateUser 更新用户
func (h *AdminHandler) UpdateUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID64, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "用户ID错误",
		})
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
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 检查用户是否存在
	var user models.User
	if err := h.service.GetDB().First(&user, uint(userID64)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "用户不存在",
		})
		return
	}

	// 检查邮箱是否被其他用户使用
	var existingUser models.User
	if err := h.service.GetDB().Where("email = ? AND id != ?", userData.Email, uint(userID64)).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "邮箱已被其他用户使用",
		})
		return
	}

	// 准备更新数据
	updateData := map[string]interface{}{
		"email":    userData.Email,
		"nickname": userData.Nickname,
	}

	if userData.IsAdmin {
		updateData["role"] = "admin"
	} else {
		updateData["role"] = "user"
	}

	if userData.IsActive {
		updateData["status"] = "active"
	} else {
		updateData["status"] = "inactive"
	}

	// 如果提供了密码，更新密码（在实际环境中应该进行哈希）
	if userData.Password != "" {
		updateData["password_hash"] = userData.Password
	}

	// 更新用户
	if err := h.service.GetDB().Model(&user).Updates(updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新用户失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "用户更新成功",
		"detail": gin.H{
			"id": uint(userID64),
		},
	})
}

// DeleteUser 删除用户
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID64, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "用户ID错误",
		})
		return
	}

	userID := uint(userID64)

	// 检查用户是否存在
	var user models.User
	if err := h.service.GetDB().First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "用户不存在",
		})
		return
	}

	// 不允许删除管理员账户
	if user.Role == "admin" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "不能删除管理员账户",
		})
		return
	}

	// 开始事务
	tx := h.service.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除用户的所有文件
	if err := tx.Where("user_id = ?", userID).Delete(&models.FileCode{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除用户文件失败: " + err.Error(),
		})
		return
	}

	// 删除用户的会话
	if err := tx.Where("user_id = ?", userID).Delete(&models.UserSession{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除用户会话失败: " + err.Error(),
		})
		return
	}

	// 删除用户
	if err := tx.Delete(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除用户失败: " + err.Error(),
		})
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "提交事务失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "用户删除成功",
		"detail": gin.H{
			"id": userID,
		},
	})
}

// UpdateUserStatus 更新用户状态
func (h *AdminHandler) UpdateUserStatus(c *gin.Context) {
	userIDStr := c.Param("id")
	userID64, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "用户ID错误",
		})
		return
	}

	var statusData struct {
		IsActive bool `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&statusData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	userID := uint(userID64)

	// 检查用户是否存在
	var user models.User
	if err := h.service.GetDB().First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "用户不存在",
		})
		return
	}

	// 不允许禁用管理员账户
	if user.Role == "admin" && !statusData.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "不能禁用管理员账户",
		})
		return
	}

	// 更新状态
	status := "active"
	if !statusData.IsActive {
		status = "inactive"
	}

	if err := h.service.GetDB().Model(&user).Update("status", status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新用户状态失败: " + err.Error(),
		})
		return
	}

	// 如果禁用用户，同时禁用其所有会话
	if !statusData.IsActive {
		h.service.GetDB().Model(&models.UserSession{}).Where("user_id = ?", userID).Update("is_active", false)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "用户状态更新成功",
		"detail": gin.H{
			"id": userID,
		},
	})
}

// GetUserFiles 获取用户文件
func (h *AdminHandler) GetUserFiles(c *gin.Context) {
	userIDStr := c.Param("id")
	userID64, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "用户ID错误",
		})
		return
	}

	userID := uint(userID64)

	// 检查用户是否存在
	var user models.User
	if err := h.service.GetDB().First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "用户不存在",
		})
		return
	}

	// 获取用户的文件列表
	var files []models.FileCode
	if err := h.service.GetDB().Where("user_id = ?", userID).Order("created_at DESC").Find(&files).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取用户文件失败: " + err.Error(),
		})
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

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"detail": gin.H{
			"files":    fileList,
			"username": user.Username,
			"total":    len(fileList),
		},
	})
}
