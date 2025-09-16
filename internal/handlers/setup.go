package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
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
	Confirm               string `json:"confirm"`
	Password              string `json:"password"`
	AllowUserRegistration bool   `json:"allowUserRegistration"`
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

// OnDatabaseInitialized 当数据库初始化完成时，handlers 包中的回调（由 main 设置）
var OnDatabaseInitialized func(daoManager *repository.RepositoryManager)

// initInProgress 用于防止并发初始化
var initInProgress int32 = 0

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
		var req SetupRequest

		// 读取原始请求体，支持两种格式：嵌套 JSON（{database:{}, admin:{}}）和扁平表单风格 JSON（db_type, db_path, admin_token, admin_password 等）
		data, err := c.GetRawData()
		if err != nil {
			common.BadRequestResponse(c, "无法读取请求体: "+err.Error())
			return
		}

		// 先尝试解码为嵌套结构
		// 注意：如果前端提交的是扁平字段（如 db_type, admin_password 等），
		// json.Unmarshal 会成功但会留下空的嵌套结构。我们在解码成功后仍需
		// 检查是否需要回退到扁平映射解析。
		if err := json.Unmarshal(data, &req); err != nil || (req.Database.Type == "" && req.Admin.Username == "" && req.Admin.Password == "") {
			// 尝试扁平化映射
			var flat map[string]interface{}
			if err := json.Unmarshal(data, &flat); err != nil {
				common.BadRequestResponse(c, "请求参数错误: 无法解析 JSON")
				return
			}

			// 映射常见字段
			if v, ok := flat["db_type"].(string); ok {
				req.Database.Type = v
			}
			if v, ok := flat["db_path"].(string); ok {
				req.Database.File = v
			}
			if v, ok := flat["db_file"].(string); ok && req.Database.File == "" {
				req.Database.File = v
			}
			if v, ok := flat["db_host"].(string); ok {
				req.Database.Host = v
			}
			if v, ok := flat["db_user"].(string); ok {
				req.Database.User = v
			}
			if v, ok := flat["db_password"].(string); ok {
				req.Database.Password = v
			}
			if v, ok := flat["db_name"].(string); ok {
				req.Database.Database = v
			}
			if v, ok := flat["admin_password"].(string); ok {
				req.Admin.Password = v
			}
			// 读取密码确认
			if v, ok := flat["admin_password_confirm"].(string); ok {
				req.Admin.Confirm = v
			}
			if v, ok := flat["admin_username"].(string); ok {
				req.Admin.Username = v
			}
			if v, ok := flat["admin_email"].(string); ok {
				req.Admin.Email = v
			}
			if v, ok := flat["admin_nickname"].(string); ok {
				req.Admin.Nickname = v
			}
			// enable_user_system may be "true"/"false" 或 布尔
			if v, ok := flat["enable_user_system"].(string); ok {
				if b, err := strconv.ParseBool(v); err == nil && b {
					req.Admin.AllowUserRegistration = true
				}
			} else if v, ok := flat["enable_user_system"].(bool); ok {
				req.Admin.AllowUserRegistration = v
			}

			// 兼容前端 admin_token：视为系统管理员令牌（Manager.AdminToken）
			if v, ok := flat["admin_token"].(string); ok && v != "" {
				manager.AdminToken = v
			}

			// 兼容前端 site_name, storage_type, max_file_size，先写入 manager 内存结构（将在 Save 时持久化）
			if v, ok := flat["site_name"].(string); ok && v != "" {
				manager.Base.Name = v
			}
			if v, ok := flat["storage_type"].(string); ok && v != "" {
				manager.Storage.Type = v
			}
			if v, ok := flat["max_file_size"].(string); ok && v != "" {
				if n, err := strconv.Atoi(v); err == nil {
					manager.Transfer.Upload.UploadSize = int64(n)
				}
			} else if v, ok := flat["max_file_size"].(float64); ok {
				manager.Transfer.Upload.UploadSize = int64(v)
			}

			// 如果提供了 sqlite 文件路径（例如 ./data/filecodebox.db），将目录设置为 Base.DataPath
			if req.Database.File != "" && req.Database.Type == "sqlite" {
				dir := filepath.Dir(req.Database.File)
				if dir == "." || dir == "" {
					// 使用默认数据目录
				} else {
					manager.Base.DataPath = dir
				}
			}

			// 如果前端没有提供管理员用户名/email，使用合理的默认值
			if req.Admin.Username == "" {
				// 尝试从 manager.AdminToken 派生用户名，否则使用 "admin"
				if manager != nil && manager.AdminToken != "" {
					req.Admin.Username = manager.AdminToken
				} else {
					req.Admin.Username = "admin"
				}
			}
			if req.Admin.Email == "" {
				req.Admin.Email = "admin@localhost"
			}

		}

		// 继续使用 req 进行验证和初始化
		// 捕获并验证扁平字段中的 storage_path（如果提供并且 storage 类型为 local），但不要立刻写入 manager
		var desiredStoragePath string
		if manager.Storage.Type == "local" {
			// 尝试从原始请求体的扁平映射中读取 storage_path
			var flat map[string]interface{}
			_ = json.Unmarshal(data, &flat)
			if v, ok := flat["storage_path"].(string); ok {
				sp := v
				if sp == "" {
					common.BadRequestResponse(c, "本地存储时必须提供 storage_path")
					return
				}

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
					f.Close()
					_ = os.Remove(testFile)
				}

				desiredStoragePath = sp
			}
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
			OnDatabaseInitialized(daoManager)
		}

		common.SuccessWithMessage(c, "系统初始化成功", map[string]interface{}{
			"message":        "系统初始化完成",
			"admin_username": req.Admin.Username,
			"database_type":  req.Database.Type,
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

	data, err := c.GetRawData()
	if err != nil {
		common.BadRequestResponse(c, "无法读取请求体: "+err.Error())
		return
	}

	// 先尝试解析为嵌套结构
	if err := json.Unmarshal(data, &req); err != nil {
		// 解析为扁平 map 并映射
		var flat map[string]interface{}
		if err := json.Unmarshal(data, &flat); err != nil {
			common.BadRequestResponse(c, "请求参数错误: 无法解析 JSON")
			return
		}
		if v, ok := flat["db_type"].(string); ok {
			req.Database.Type = v
		}
		if v, ok := flat["db_path"].(string); ok {
			req.Database.File = v
		}
		if v, ok := flat["db_file"].(string); ok && req.Database.File == "" {
			req.Database.File = v
		}
		if v, ok := flat["db_host"].(string); ok {
			req.Database.Host = v
		}
		if v, ok := flat["db_user"].(string); ok {
			req.Database.User = v
		}
		if v, ok := flat["db_password"].(string); ok {
			req.Database.Password = v
		}
		if v, ok := flat["db_name"].(string); ok {
			req.Database.Database = v
		}
		if v, ok := flat["admin_password"].(string); ok {
			req.Admin.Password = v
		}
		if v, ok := flat["admin_username"].(string); ok {
			req.Admin.Username = v
		}
		if v, ok := flat["admin_email"].(string); ok {
			req.Admin.Email = v
		}
		// 存储路径（local 存储时使用）
		if v, ok := flat["storage_path"].(string); ok {
			h.manager.Storage.StoragePath = v
		}
		// 读取密码确认
		if v, ok := flat["admin_password_confirm"].(string); ok {
			req.Admin.Confirm = v
		}
		if v, ok := flat["admin_nickname"].(string); ok {
			req.Admin.Nickname = v
		}
		if v, ok := flat["enable_user_system"].(string); ok {
			if b, err := strconv.ParseBool(v); err == nil && b {
				req.Admin.AllowUserRegistration = true
			}
		} else if v, ok := flat["enable_user_system"].(bool); ok {
			req.Admin.AllowUserRegistration = v
		}

		// 兼容前端 admin_token：视为系统管理员令牌（Manager.AdminToken）
		if v, ok := flat["admin_token"].(string); ok && v != "" {
			h.manager.AdminToken = v
		}

		// 如果前端没有提供管理员用户名/email，使用合理的默认值
		if req.Admin.Username == "" {
			// 尝试从 manager.AdminToken 派生用户名，否则使用 "admin"
			if h.manager != nil && h.manager.AdminToken != "" {
				req.Admin.Username = h.manager.AdminToken
			} else {
				req.Admin.Username = "admin"
			}
		}
		if req.Admin.Email == "" {
			req.Admin.Email = "admin@localhost"
		}

		// 如果 storage_type 是 local，则处理 storage_path（扁平字段）
		if h.manager.Storage.Type == "local" {
			if v, ok := flat["storage_path"].(string); ok {
				sp := v
				if sp == "" {
					common.BadRequestResponse(c, "本地存储时必须提供 storage_path")
					return
				}

				// 若为相对路径，则相对于 h.manager.Base.DataPath
				if !filepath.IsAbs(sp) {
					if h.manager.Base != nil && h.manager.Base.DataPath != "" {
						sp = filepath.Join(h.manager.Base.DataPath, sp)
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
					f.Close()
					_ = os.Remove(testFile)
				}

				h.manager.Storage.StoragePath = sp
				if err := h.manager.Save(); err != nil {
					common.InternalServerErrorResponse(c, "保存存储配置失败: "+err.Error())
					return
				}
			} else {
				// 如果没有传 storage_path，前端应已校验，但服务器端也需要确保
				common.BadRequestResponse(c, "本地存储时必须提供 storage_path")
				return
			}
		}
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
		"message":        "系统初始化完成",
		"admin_username": req.Admin.Username,
	})
}
