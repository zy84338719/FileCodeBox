package handlers

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	"github.com/zy84338719/filecodebox/internal/common"
	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/database"
	"github.com/zy84338719/filecodebox/internal/models"
	"github.com/zy84338719/filecodebox/internal/repository"
	"github.com/zy84338719/filecodebox/internal/services/auth"
	"gorm.io/gorm"
)

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
	Confirm               string `json:"confirm"`
	Password              string `json:"password"`
	AllowUserRegistration bool   `json:"allowUserRegistration"`
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
	// 在创建前检查用户是否已存在（按用户名或邮箱），保证初始化过程幂等
	if h.daoManager == nil {
		return fmt.Errorf("daoManager 未初始化")
	}

	// 检查用户名
	if _, err := h.daoManager.User.GetByUsername(admin.Username); err == nil {
		log.Printf("[createAdminUser] 管理员已存在（用户名）：%s，跳过创建", admin.Username)
		return nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("查询用户失败: %w", err)
	}

	// 检查邮箱
	if _, err := h.daoManager.User.GetByEmail(admin.Email); err == nil {
		log.Printf("[createAdminUser] 管理员已存在（邮箱）：%s，跳过创建", admin.Email)
		return nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("查询用户失败: %w", err)
	}

	// 使用auth服务创建用户并哈希密码
	authService := auth.NewService(h.daoManager, h.manager)
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

	err = h.daoManager.User.Create(user)
	if err != nil {
		// 如果是唯一约束冲突（用户已存在），视为成功（幂等行为）
		if contains(err.Error(), "UNIQUE constraint failed") || contains(err.Error(), "duplicate key value") {
			log.Printf("[createAdminUser] 用户已存在，忽略错误: %v", err)
			return nil
		}
		return err
	}
	return nil
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

// legacy flat mapping removed per request - only nested JSON supported now

// OnDatabaseInitialized 当数据库初始化完成时，handlers 包中的回调（由 main 设置）
var OnDatabaseInitialized func(daoManager *repository.RepositoryManager)

// initInProgress 用于防止并发初始化
var initInProgress int32 = 0

// onDBInitCalled 防止重复调用 OnDatabaseInitialized（多次 POST /setup 导致重复注册路由）
var onDBInitCalled int32 = 0

// InitializeNoDB 用于在没有 daoManager 的情况下处理 /setup/initialize 请求
// 它会：验证请求、使用配置管理器初始化数据库、创建 daoManager、创建管理员用户，最后触发 OnDatabaseInitialized 回调
func InitializeNoDB(manager *config.ConfigManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 并发保护：避免多个初始化同时进行
		if !atomic.CompareAndSwapInt32(&initInProgress, 0, 1) {
			common.BadRequestResponse(c, "系统正在初始化，请稍候")
			return
		}
		defer atomic.StoreInt32(&initInProgress, 0)
		// 解析 JSON（仅接受嵌套结构），不再兼容 legacy 扁平字段
		var req SetupRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			common.BadRequestResponse(c, "请求参数错误: "+err.Error())
			return
		}
		// 继续使用 req 进行验证和初始化
		var desiredStoragePath string
		// 不再从请求体中读取 legacy storage_path；如果配置管理器已包含 storage path，则后续逻辑会处理

		if manager.Storage.Type == "local" {
			sp := manager.Storage.StoragePath
			// 若为相对路径，则相对于 manager.Base.DataPath
			if !filepath.IsAbs(sp) {
				if manager.Base != nil && manager.Base.DataPath != "" {
					sp = filepath.Join(manager.Base.DataPath, sp)
				} else {
					sp, _ = filepath.Abs(sp)
				}
			}

			// 尝试创建目录（如果不存在）
			if _, err := os.Stat(sp); os.IsNotExist(err) {
				if err := os.MkdirAll(sp, 0755); err != nil {
					common.InternalServerErrorResponse(c, "创建本地存储目录失败: "+err.Error())
					return
				}
			}

			// 检查是否可写：尝试在目录中创建一个临时文件
			testFile := filepath.Join(sp, ".perm_check")
			if f, err := os.Create(testFile); err != nil {
				common.InternalServerErrorResponse(c, "本地存储路径不可写: "+err.Error())
				return
			} else {
				if err := f.Close(); err != nil {
					log.Printf("warning: failed to close perm check file: %v", err)
				}
				_ = os.Remove(testFile)
			}

			desiredStoragePath = sp
		}

		// 验证管理员信息
		if len(req.Admin.Username) < 3 {
			common.BadRequestResponse(c, "用户名长度至少3个字符")
			return
		}
		if len(req.Admin.Password) < 6 {
			common.BadRequestResponse(c, "密码长度至少6个字符")
			return
		}
		// 验证密码确认（若提供）
		if req.Admin.Confirm != "" && req.Admin.Confirm != req.Admin.Password {
			common.BadRequestResponse(c, "两次输入的管理员密码不一致")
			return
		}
		if req.Admin.Email == "" || len(req.Admin.Email) < 5 || !contains(req.Admin.Email, "@") {
			common.BadRequestResponse(c, "邮箱格式无效")
			return
		}

		// 验证数据库配置（简单校验）
		switch req.Database.Type {
		case "sqlite":
			if req.Database.File == "" {
				common.BadRequestResponse(c, "SQLite 数据库文件路径不能为空")
				return
			}
		case "mysql", "postgres":
			if req.Database.Host == "" || req.Database.User == "" || req.Database.Database == "" {
				common.BadRequestResponse(c, "关系型数据库连接信息不完整")
				return
			}
		default:
			common.BadRequestResponse(c, "不支持的数据库类型: "+req.Database.Type)
			return
		}

		// 将数据库配置写入 manager
		manager.Database.Type = req.Database.Type
		switch req.Database.Type {
		case "sqlite":
			manager.Database.Name = req.Database.File
			manager.Database.Host = ""
			manager.Database.Port = 0
			manager.Database.User = ""
			manager.Database.Pass = ""
		case "mysql", "postgres":
			manager.Database.Host = req.Database.Host
			manager.Database.Port = req.Database.Port
			manager.Database.User = req.Database.User
			manager.Database.Pass = req.Database.Password
			manager.Database.Name = req.Database.Database
		}
		// 注意：此处不能在数据库未初始化前调用 manager.Save()，因为 Save 仅保存到数据库。
		// 后续将在数据库初始化并注入 manager 后再次保存配置。

		// 初始化数据库连接并执行自动迁移
		// Ensure Base config exists
		if manager.Base == nil {
			manager.Base = &config.BaseConfig{}
		}

		// For sqlite, determine DataPath from provided database file if not already set
		if manager.Database.Type == "sqlite" {
			dataFile := manager.Database.Name
			if dataFile == "" {
				dataFile = req.Database.File
			}
			var dataDir string
			if dataFile != "" {
				dataDir = filepath.Dir(dataFile)
				if dataDir == "." || dataDir == "" {
					dataDir = "./data"
				}
				// make absolute if possible
				if !filepath.IsAbs(dataDir) {
					if abs, err := filepath.Abs(dataDir); err == nil {
						dataDir = abs
					}
				}
				manager.Base.DataPath = dataDir
			} else if manager.Base.DataPath == "" {
				// fallback default
				manager.Base.DataPath = "./data"
				if abs, err := filepath.Abs(manager.Base.DataPath); err == nil {
					manager.Base.DataPath = abs
				}
			}

			// ensure directory exists before InitWithManager attempts to mkdir
			if err := os.MkdirAll(manager.Base.DataPath, 0750); err != nil {
				common.InternalServerErrorResponse(c, "创建SQLite数据目录失败: "+err.Error())
				return
			}
		}

		log.Printf("[InitializeNoDB] 开始调用 database.InitWithManager, dbType=%s, dataPath=%s", manager.Database.Type, manager.Base.DataPath)
		db, err := database.InitWithManager(manager)
		if err != nil {
			log.Printf("[InitializeNoDB] InitWithManager 失败: %v", err)
			common.InternalServerErrorResponse(c, "初始化数据库失败: "+err.Error())
			return
		}

		// 将 db 注入 manager 并初始化默认配置
		// Inject DB connection into manager. Initialization of config from DB is disabled.
		manager.SetDB(db)

		// 诊断检查：确认 manager 内部已设置 db
		if manager.GetDB() == nil {
			log.Printf("[InitializeNoDB] 警告: manager.GetDB() 返回 nil（注入失败）")
			common.InternalServerErrorResponse(c, "初始化失败：配置管理器未能获取数据库连接")
			return
		}

		// 创建 daoManager
		daoManager := repository.NewRepositoryManager(db)

		// 如果之前捕获了 desiredStoragePath，则此时 manager 已注入 DB，可以持久化 storage_path
		if desiredStoragePath != "" {
			manager.Storage.StoragePath = desiredStoragePath
			if err := manager.Save(); err != nil {
				log.Printf("[InitializeNoDB] 保存 storage_path 失败: %v", err)
				// 记录但不阻塞初始化流程
				if manager.Base != nil && manager.Base.DataPath != "" {
					_ = os.WriteFile(manager.Base.DataPath+"/init_save_storage_err.log", []byte(err.Error()), 0644)
				} else {
					_ = os.WriteFile("init_save_storage_err.log", []byte(err.Error()), 0644)
				}
			}
		}

		// 创建管理员用户（使用 SetupHandler.createAdminUser，包含幂等性处理）
		setupHandler := NewSetupHandler(daoManager, manager)
		if err := setupHandler.createAdminUser(req.Admin); err != nil {
			log.Printf("[InitializeNoDB] 创建管理员用户失败: %v", err)
			common.InternalServerErrorResponse(c, "创建管理员用户失败: "+err.Error())
			return
		}

		// 启用用户系统配置
		if req.Admin.AllowUserRegistration {
			manager.User.AllowUserRegistration = 1
		} else {
			manager.User.AllowUserRegistration = 0
		}
		if err := manager.Save(); err != nil {
			// 不阻塞初始化成功路径，但记录错误
			log.Printf("[InitializeNoDB] manager.Save() 返回错误（但不阻塞初始化）: %v", err)
			// 将错误写入数据目录下的日志文件以便排查
			if manager.Base != nil && manager.Base.DataPath != "" {
				_ = os.WriteFile(manager.Base.DataPath+"/init_save_err.log", []byte(err.Error()), 0644)
			} else {
				_ = os.WriteFile("init_save_err.log", []byte(err.Error()), 0644)
			}
		}

		// 触发回调以让主程序挂载其余路由并启动后台任务
		if OnDatabaseInitialized != nil {
			// 只允许调用一次，避免重复注册路由导致 gin panic
			if atomic.CompareAndSwapInt32(&onDBInitCalled, 0, 1) {
				OnDatabaseInitialized(daoManager)
			} else {
				log.Printf("[InitializeNoDB] OnDatabaseInitialized 已调用，跳过重复挂载")
			}
		}

		common.SuccessWithMessage(c, "系统初始化成功", map[string]interface{}{
			"message":       "系统初始化完成",
			"username":      req.Admin.Username,
			"database_type": req.Database.Type,
		})
	}
}

// Initialize 在数据库已经可用的情况下处理 /setup/initialize 请求
// 该方法用于通过已存在的 daoManager 来完成系统初始化（保存配置、创建管理员等）
func (h *SetupHandler) Initialize(c *gin.Context) {
	// 并发保护：避免多个初始化同时进行
	if !atomic.CompareAndSwapInt32(&initInProgress, 0, 1) {
		common.BadRequestResponse(c, "系统正在初始化，请稍候")
		return
	}
	defer atomic.StoreInt32(&initInProgress, 0)

	if h == nil || h.manager == nil {
		common.InternalServerErrorResponse(c, "服务器未正确初始化")
		return
	}
	if h.daoManager == nil {
		common.InternalServerErrorResponse(c, "数据库管理器未初始化")
		return
	}

	var req SetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}

	// 验证管理员信息
	if len(req.Admin.Username) < 3 {
		common.BadRequestResponse(c, "用户名长度至少3个字符")
		return
	}
	if len(req.Admin.Password) < 6 {
		common.BadRequestResponse(c, "密码长度至少6个字符")
		return
	}
	if req.Admin.Confirm != "" && req.Admin.Confirm != req.Admin.Password {
		common.BadRequestResponse(c, "两次输入的管理员密码不一致")
		return
	}
	if req.Admin.Email == "" || len(req.Admin.Email) < 5 || !contains(req.Admin.Email, "@") {
		common.BadRequestResponse(c, "邮箱格式无效")
		return
	}

	// 更新数据库配置并保存（manager.Save 会将配置写入数据库，因为 manager 已注入 DB）
	if err := h.updateDatabaseConfig(req.Database); err != nil {
		common.InternalServerErrorResponse(c, "保存数据库配置失败: "+err.Error())
		return
	}

	// 创建管理员用户
	if err := h.createAdminUser(req.Admin); err != nil {
		common.InternalServerErrorResponse(c, "创建管理员用户失败: "+err.Error())
		return
	}

	// 启用用户系统设置
	if err := h.enableUserSystem(req.Admin); err != nil {
		// 记录但不阻塞主要流程
		log.Printf("enableUserSystem 返回错误: %v", err)
	}

	common.SuccessWithMessage(c, "系统初始化成功", map[string]interface{}{
		"message":  "系统初始化完成",
		"username": req.Admin.Username,
	})
}
