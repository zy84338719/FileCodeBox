package handlers

import (
	"filecodebox/internal/config"
	"filecodebox/internal/services"
	"net/http"
	"strconv"
	"time"

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
