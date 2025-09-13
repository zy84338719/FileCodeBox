package handlers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/zy84338719/filecodebox/internal/common"
	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/models"
	"github.com/zy84338719/filecodebox/internal/repository"
	"github.com/zy84338719/filecodebox/internal/services/auth"
)

// SetupHandler 系统初始化处理器
type SetupHandler struct {
	daoManager *repository.RepositoryManager
	manager    *config.ConfigManager
}

// NewSetupHandler 创建系统初始化处理器
func NewSetupHandler(daoManager *repository.RepositoryManager, manager *config.ConfigManager) *SetupHandler {
	return &SetupHandler{
		daoManager: daoManager,
		manager:    manager,
	}
}

// SetupRequest 初始化请求结构
type SetupRequest struct {
	Database DatabaseConfig `json:"database"`
	Admin    AdminConfig    `json:"admin"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type     string `json:"type"`     // sqlite, mysql, postgres
	File     string `json:"file"`     // SQLite 文件路径
	Host     string `json:"host"`     // MySQL/PostgreSQL 主机
	Port     int    `json:"port"`     // MySQL/PostgreSQL 端口
	User     string `json:"user"`     // MySQL/PostgreSQL 用户名
	Password string `json:"password"` // MySQL/PostgreSQL 密码
	Database string `json:"database"` // MySQL/PostgreSQL 数据库名
}

// AdminConfig 管理员配置
type AdminConfig struct {
	Username              string `json:"username"`
	Email                 string `json:"email"`
	Nickname              string `json:"nickname"`
	Password              string `json:"password"`
	AllowUserRegistration bool   `json:"allowUserRegistration"`
}

// Initialize 执行系统初始化
func (h *SetupHandler) Initialize(c *gin.Context) {
	// 首先检查系统是否已经初始化
	adminCount, err := h.daoManager.User.CountAdminUsers()
	if err != nil {
		common.InternalServerErrorResponse(c, "检查系统状态失败")
		return
	}

	if adminCount > 0 {
		common.BadRequestResponse(c, "系统已经初始化，无法重复初始化")
		return
	}

	var req SetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}

	// 验证管理员信息
	if err := h.validateAdminConfig(req.Admin); err != nil {
		common.BadRequestResponse(c, err.Error())
		return
	}

	// 验证数据库配置
	if err := h.validateDatabaseConfig(req.Database); err != nil {
		common.BadRequestResponse(c, err.Error())
		return
	}

	// 更新数据库配置
	if err := h.updateDatabaseConfig(req.Database); err != nil {
		common.InternalServerErrorResponse(c, "更新数据库配置失败: "+err.Error())
		return
	}

	// 创建管理员用户
	if err := h.createAdminUser(req.Admin); err != nil {
		common.InternalServerErrorResponse(c, "创建管理员用户失败: "+err.Error())
		return
	}

	// 启用用户系统
	if err := h.enableUserSystem(req.Admin); err != nil {
		common.InternalServerErrorResponse(c, "启用用户系统失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, "系统初始化成功", map[string]interface{}{
		"message":        "系统初始化完成",
		"admin_username": req.Admin.Username,
		"database_type":  req.Database.Type,
	})
}

// validateAdminConfig 验证管理员配置
func (h *SetupHandler) validateAdminConfig(admin AdminConfig) error {
	if len(admin.Username) < 3 {
		return fmt.Errorf("用户名长度至少3个字符")
	}

	if len(admin.Password) < 6 {
		return fmt.Errorf("密码长度至少6个字符")
	}

	if admin.Email == "" {
		return fmt.Errorf("邮箱地址不能为空")
	}

	// 简单的邮箱格式验证
	if len(admin.Email) < 5 || !contains(admin.Email, "@") {
		return fmt.Errorf("邮箱格式无效")
	}

	return nil
}

// validateDatabaseConfig 验证数据库配置
func (h *SetupHandler) validateDatabaseConfig(db DatabaseConfig) error {
	switch db.Type {
	case "sqlite":
		if db.File == "" {
			return fmt.Errorf("SQLite 数据库文件路径不能为空")
		}
	case "mysql", "postgres":
		if db.Host == "" {
			return fmt.Errorf("数据库主机地址不能为空")
		}
		if db.Port <= 0 || db.Port > 65535 {
			return fmt.Errorf("数据库端口无效")
		}
		if db.User == "" {
			return fmt.Errorf("数据库用户名不能为空")
		}
		if db.Database == "" {
			return fmt.Errorf("数据库名不能为空")
		}
	default:
		return fmt.Errorf("不支持的数据库类型: %s", db.Type)
	}
	return nil
}

// updateDatabaseConfig 更新数据库配置
func (h *SetupHandler) updateDatabaseConfig(db DatabaseConfig) error {
	// 更新配置管理器中的数据库配置
	h.manager.Database.Type = db.Type

	switch db.Type {
	case "sqlite":
		h.manager.Database.Host = ""
		h.manager.Database.Port = 0
		h.manager.Database.User = ""
		h.manager.Database.Pass = ""
		h.manager.Database.Name = db.File
	case "mysql":
		h.manager.Database.Host = db.Host
		h.manager.Database.Port = db.Port
		h.manager.Database.User = db.User
		h.manager.Database.Pass = db.Password
		h.manager.Database.Name = db.Database
	case "postgres":
		h.manager.Database.Host = db.Host
		h.manager.Database.Port = db.Port
		h.manager.Database.User = db.User
		h.manager.Database.Pass = db.Password
		h.manager.Database.Name = db.Database
	}

	// 保存配置到文件
	return h.manager.Save()
}

// createAdminUser 创建管理员用户
func (h *SetupHandler) createAdminUser(admin AdminConfig) error {
	// 使用auth服务创建用户
	authService := auth.NewService(h.daoManager, h.manager)

	// 哈希密码
	hashedPassword, err := authService.HashPassword(admin.Password)
	if err != nil {
		return fmt.Errorf("密码哈希失败: %w", err)
	}

	// 设置默认昵称
	nickname := admin.Nickname
	if nickname == "" {
		nickname = admin.Username
	}

	// 创建管理员用户
	user := &models.User{
		Username:        admin.Username,
		Email:           admin.Email,
		PasswordHash:    hashedPassword,
		Nickname:        nickname,
		Role:            "admin", // 设置为管理员角色
		Status:          "active",
		EmailVerified:   true, // 管理员默认已验证邮箱
		MaxUploadSize:   h.manager.User.UserUploadSize,
		MaxStorageQuota: h.manager.User.UserStorageQuota,
	}

	return h.daoManager.User.Create(user)
}

// enableUserSystem 启用用户系统
func (h *SetupHandler) enableUserSystem(adminConfig AdminConfig) error {
	// 用户系统始终启用，无需设置

	// 根据管理员选择设置用户注册权限
	if adminConfig.AllowUserRegistration {
		h.manager.User.AllowUserRegistration = 1
	} else {
		h.manager.User.AllowUserRegistration = 0
	}

	// 保存配置
	return h.manager.Save()
}

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
