package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/dao"
	"github.com/zy84338719/filecodebox/internal/models"
	"github.com/zy84338719/filecodebox/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
)

// AdminService 管理员服务
type AdminService struct {
	config         *config.Config
	storageManager *storage.StorageManager
	daoManager     *dao.DAOManager
	authService    *AuthService
}

// NewAdminService 创建管理员服务
func NewAdminService(db *gorm.DB, config *config.Config, storageManager *storage.StorageManager) *AdminService {
	return &AdminService{
		config:         config,
		storageManager: storageManager,
		daoManager:     dao.NewDAOManager(db),
		authService:    NewAuthService(db, config),
	}
}

// GetStats 获取统计信息
func (s *AdminService) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 用户统计信息
	// 总用户数
	totalUsers, err := s.daoManager.User.Count()
	if err != nil {
		return nil, err
	}
	stats["total_users"] = totalUsers

	// 活跃用户数
	activeUsers, err := s.daoManager.User.CountActive()
	if err != nil {
		return nil, err
	}
	stats["active_users"] = activeUsers

	// 今日注册用户数
	todayRegistrations, err := s.daoManager.User.CountTodayRegistrations()
	if err != nil {
		return nil, err
	}
	stats["today_registrations"] = todayRegistrations

	// 今日上传文件数
	todayUploads, err := s.daoManager.FileCode.CountToday()
	if err != nil {
		return nil, err
	}
	stats["today_uploads"] = todayUploads

	// 文件统计信息
	// 总文件数（不包括已删除的）
	totalFiles, err := s.daoManager.FileCode.Count()
	if err != nil {
		return nil, err
	}
	stats["total_files"] = totalFiles

	// 活跃文件数（未过期且未删除）
	activeFiles, err := s.daoManager.FileCode.CountActive()
	if err != nil {
		return nil, err
	}
	stats["active_files"] = activeFiles

	// 总大小（不包括已删除的）
	totalSize, err := s.daoManager.FileCode.GetTotalSize()
	if err != nil {
		return nil, err
	}
	stats["total_size"] = totalSize

	// 系统启动时间
	sysStart, err := s.daoManager.KeyValue.GetByKey("sys_start")
	if err == nil {
		stats["sys_start"] = sysStart.Value
	} else {
		// 如果没有记录，创建一个
		startTime := fmt.Sprintf("%d", time.Now().UnixMilli())
		err := s.daoManager.KeyValue.SetValue("sys_start", startTime)
		if err != nil {
			return nil, err
		}
		stats["sys_start"] = startTime
	}

	return stats, nil
}

// GetFiles 获取文件列表
func (s *AdminService) GetFiles(page, pageSize int, search string) ([]models.FileCode, int64, error) {
	return s.daoManager.FileCode.List(page, pageSize, search)
}

// DeleteFile 删除文件
func (s *AdminService) DeleteFile(id uint) error {
	fileCode, err := s.daoManager.FileCode.GetByID(id)
	if err != nil {
		return err
	}

	// 删除实际文件
	storageInterface := s.storageManager.GetStorage()
	if err := storageInterface.DeleteFile(fileCode); err != nil {
		// 记录错误，但不阻止数据库删除
		fmt.Printf("Warning: Failed to delete physical file: %v\n", err)
	}

	return s.daoManager.FileCode.DeleteByFileCode(fileCode)
}

// DeleteFileByCode 通过代码删除文件
func (s *AdminService) DeleteFileByCode(code string) error {
	fileCode, err := s.daoManager.FileCode.GetByCode(code)
	if err != nil {
		return err
	}

	// 删除实际文件
	storageInterface := s.storageManager.GetStorage()
	if err := storageInterface.DeleteFile(fileCode); err != nil {
		// 记录错误，但不阻止数据库删除
		fmt.Printf("Warning: Failed to delete physical file: %v\n", err)
	}

	return s.daoManager.FileCode.DeleteByFileCode(fileCode)
}

// GetFile 获取文件信息
func (s *AdminService) GetFile(id uint) (*models.FileCode, error) {
	return s.daoManager.FileCode.GetByID(id)
}

// GetFileByCode 通过代码获取文件信息
func (s *AdminService) GetFileByCode(code string) (*models.FileCode, error) {
	return s.daoManager.FileCode.GetByCode(code)
}

// UpdateFile 更新文件
func (s *AdminService) UpdateFile(id uint, text, name string, expTime time.Time) error {
	updates := map[string]interface{}{
		"text":       text,
		"expired_at": expTime,
	}
	return s.daoManager.FileCode.UpdateColumns(id, updates)
}

// UpdateFileByCode 通过代码更新文件
func (s *AdminService) UpdateFileByCode(code, text, name string, expTime time.Time) error {
	fileCode, err := s.daoManager.FileCode.GetByCode(code)
	if err != nil {
		return err
	}
	updates := map[string]interface{}{
		"text":       text,
		"expired_at": expTime,
	}
	return s.daoManager.FileCode.UpdateColumns(fileCode.ID, updates)
}

// DownloadFile 下载文件
func (s *AdminService) DownloadFile(c *gin.Context, id uint) error {
	fileCode, err := s.daoManager.FileCode.GetByID(id)
	if err != nil {
		return err
	}

	// 使用存储管理器处理文件下载
	return s.serveFile(c, fileCode)
}

// DownloadFileByCode 通过代码下载文件
func (s *AdminService) DownloadFileByCode(c *gin.Context, code string) error {
	fileCode, err := s.daoManager.FileCode.GetByCode(code)
	if err != nil {
		return err
	}

	// 使用存储管理器处理文件下载
	return s.serveFile(c, fileCode)
}

// serveFile 提供文件服务
func (s *AdminService) serveFile(c *gin.Context, fileCode *models.FileCode) error {
	storageInterface := s.storageManager.GetStorage()

	// 使用存储接口的GetFileResponse方法
	return storageInterface.GetFileResponse(c, fileCode)
}

// GetUsers 获取用户列表
func (s *AdminService) GetUsers(page, pageSize int, search string) ([]models.User, int64, error) {
	return s.daoManager.User.List(page, pageSize, search)
}

// CreateUser 创建用户 - 使用统一的认证服务
func (s *AdminService) CreateUser(username, email, password, nickname, role, status string) (*models.User, error) {
	// 哈希密码
	hashedPassword, err := s.authService.HashPassword(password)
	if err != nil {
		return nil, err
	}

	// 验证用户输入
	if err := s.authService.ValidateUsername(username); err != nil {
		return nil, err
	}
	if err := s.authService.ValidateEmail(email); err != nil {
		return nil, err
	}
	if err := s.authService.ValidatePassword(password, username, email); err != nil {
		return nil, err
	}

	// 规范化数据
	username = s.authService.NormalizeUsername(username)
	email = s.authService.NormalizeEmail(email)

	// 检查用户唯一性
	existingUser, err := s.daoManager.User.CheckExists(username, email)
	if err == nil {
		if existingUser.Username == username {
			return nil, errors.New("用户名已存在")
		}
		return nil, errors.New("邮箱已存在")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("检查用户唯一性失败: %w", err)
	}

	user := &models.User{
		Username:        username,
		Email:           email,
		PasswordHash:    hashedPassword,
		Nickname:        nickname,
		Role:            role,
		Status:          status,
		EmailVerified:   true, // 管理员创建的用户默认已验证
		MaxUploadSize:   s.config.UserUploadSize,
		MaxStorageQuota: s.config.UserStorageQuota,
	}

	if err := s.daoManager.User.Create(user); err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	return user, nil
}

// UpdateUser 更新用户 - 使用统一的认证服务
func (s *AdminService) UpdateUser(id uint, email, password, nickname, role, status string) error {
	// 检查用户是否存在
	user, err := s.daoManager.User.GetByID(id)
	if err != nil {
		return fmt.Errorf("用户不存在: %w", err)
	}

	// 准备更新数据
	updates := make(map[string]interface{})

	if email != "" {
		if err := s.authService.ValidateEmail(email); err != nil {
			return err
		}
		email = s.authService.NormalizeEmail(email)
		updates["email"] = email
	}

	if password != "" {
		if err := s.authService.ValidatePassword(password, user.Username, email); err != nil {
			return err
		}
		hashedPassword, err := s.authService.HashPassword(password)
		if err != nil {
			return err
		}
		updates["password_hash"] = hashedPassword
	}

	if nickname != "" {
		updates["nickname"] = nickname
	}
	if role != "" {
		updates["role"] = role
	}
	if status != "" {
		updates["status"] = status
	}

	return s.daoManager.User.UpdateColumns(id, updates)
}

// DeleteUser 删除用户
func (s *AdminService) DeleteUser(id uint) error {
	user, err := s.daoManager.User.GetByID(id)
	if err != nil {
		return err
	}
	// 开始事务
	tx := s.daoManager.BeginTransaction()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	err = s.daoManager.User.Delete(tx, user)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// GetUser 获取用户
func (s *AdminService) GetUser(id uint) (*models.User, error) {
	return s.daoManager.User.GetByID(id)
}

// UpdateConfig 更新配置
func (s *AdminService) UpdateConfig(configData map[string]interface{}) error {
	for key, value := range configData {
		// 将value转换为字符串
		var valueStr string
		switch v := value.(type) {
		case string:
			valueStr = v
		case int, int32, int64:
			valueStr = fmt.Sprintf("%d", v)
		case float32, float64:
			valueStr = fmt.Sprintf("%g", v)
		case bool:
			if v {
				valueStr = "1"
			} else {
				valueStr = "0"
			}
		default:
			// 对于复杂类型，序列化为JSON
			jsonBytes, err := json.Marshal(v)
			if err != nil {
				return fmt.Errorf("序列化配置值失败: %w", err)
			}
			valueStr = string(jsonBytes)
		}

		if err := s.daoManager.KeyValue.SetValue(key, valueStr); err != nil {
			return fmt.Errorf("保存配置失败: %w", err)
		}
	}
	return nil
}

// GetConfig 获取配置
func (s *AdminService) GetConfig() (map[string]interface{}, error) {
	config := make(map[string]interface{})

	// 获取所有配置
	allConfigs, err := s.daoManager.KeyValue.GetAll()
	if err != nil {
		return nil, err
	}

	for _, kv := range allConfigs {
		// 尝试转换为适当的类型
		value := s.parseConfigValue(kv.Key, kv.Value)
		config[kv.Key] = value
	}

	// 添加一些运行时配置
	config["name"] = s.config.Name
	config["description"] = s.config.Description
	config["keywords"] = s.config.Keywords
	config["port"] = s.config.Port
	config["host"] = s.config.Host
	config["data_path"] = s.config.DataPath
	config["production"] = s.config.Production
	config["notify_title"] = s.config.NotifyTitle
	config["notify_content"] = s.config.NotifyContent
	config["page_explain"] = s.config.PageExplain
	config["open_upload"] = s.config.OpenUpload
	config["upload_size"] = s.config.UploadSize
	config["enable_chunk"] = s.config.EnableChunk
	config["chunk_size"] = s.config.ChunkSize
	config["max_save_seconds"] = s.config.MaxSaveSeconds
	config["enable_concurrent_download"] = s.config.EnableConcurrentDownload
	config["max_concurrent_downloads"] = s.config.MaxConcurrentDownloads
	config["download_timeout"] = s.config.DownloadTimeout
	config["expire_style"] = s.config.ExpireStyle
	config["upload_minute"] = s.config.UploadMinute
	config["upload_count"] = s.config.UploadCount
	config["error_minute"] = s.config.ErrorMinute
	config["error_count"] = s.config.ErrorCount
	config["themes_select"] = s.config.ThemesSelect
	config["themes_choices"] = s.config.ThemesChoices
	config["opacity"] = s.config.Opacity
	config["background"] = s.config.Background
	config["file_storage"] = s.config.FileStorage
	config["storage_path"] = s.config.StoragePath
	config["s3_access_key_id"] = s.config.S3AccessKeyID
	config["s3_secret_access_key"] = s.config.S3SecretAccessKey
	config["s3_bucket_name"] = s.config.S3BucketName
	config["s3_endpoint_url"] = s.config.S3EndpointURL
	config["s3_region_name"] = s.config.S3RegionName
	config["s3_signature_version"] = s.config.S3SignatureVersion
	config["s3_hostname"] = s.config.S3Hostname
	config["s3_proxy"] = s.config.S3Proxy
	config["aws_session_token"] = s.config.AWSSessionToken
	config["webdav_hostname"] = s.config.WebDAVHostname
	config["webdav_root_path"] = s.config.WebDAVRootPath
	config["webdav_proxy"] = s.config.WebDAVProxy
	config["webdav_url"] = s.config.WebDAVURL
	config["webdav_password"] = s.config.WebDAVPassword
	config["webdav_username"] = s.config.WebDAVUsername
	config["onedrive_domain"] = s.config.OneDriveDomain
	config["onedrive_client_id"] = s.config.OneDriveClientID
	config["onedrive_username"] = s.config.OneDriveUsername
	config["onedrive_password"] = s.config.OneDrivePassword
	config["onedrive_root_path"] = s.config.OneDriveRootPath
	config["onedrive_proxy"] = s.config.OneDriveProxy
	config["admin_token"] = s.config.AdminToken
	config["robots_text"] = s.config.RobotsText
	config["enable_user_system"] = s.config.EnableUserSystem
	config["allow_user_registration"] = s.config.AllowUserRegistration
	config["require_email_verify"] = s.config.RequireEmailVerify
	config["user_upload_size"] = s.config.UserUploadSize
	config["user_storage_quota"] = s.config.UserStorageQuota
	config["session_expiry_hours"] = s.config.SessionExpiryHours
	config["max_sessions_per_user"] = s.config.MaxSessionsPerUser
	config["jwt_secret"] = s.config.JWTSecret

	return config, nil
}

// parseConfigValue 解析配置值
func (s *AdminService) parseConfigValue(key, value string) interface{} {
	// 根据key的类型来解析value
	switch key {
	case "port", "upload_size", "chunk_size", "max_save_seconds", "max_concurrent_downloads",
		"download_timeout", "upload_minute", "upload_count", "error_minute", "error_count",
		"s3_proxy", "webdav_proxy", "onedrive_proxy", "show_admin_address", "enable_user_system",
		"allow_user_registration", "require_email_verify", "user_upload_size", "user_storage_quota",
		"session_expiry_hours", "max_sessions_per_user", "open_upload", "enable_chunk",
		"enable_concurrent_download":
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	case "opacity":
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	case "production":
		return value == "true" || value == "1"
	case "expire_style", "themes_choices":
		// 这些可能是JSON数组
		var arr []interface{}
		if err := json.Unmarshal([]byte(value), &arr); err == nil {
			return arr
		}
	}

	// 默认返回字符串
	return value
}

// GenerateToken 生成管理员JWT令牌
func (s *AdminService) GenerateToken() (string, error) {
	// 创建JWT claims
	claims := jwt.MapClaims{
		"is_admin": true,
		"exp":      time.Now().Add(24 * time.Hour).Unix(), // 24小时过期
	}

	// 创建token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名token
	tokenString, err := token.SignedString([]byte(s.config.AdminToken))
	if err != nil {
		return "", fmt.Errorf("生成token失败: %w", err)
	}

	return tokenString, nil
}

// ValidateToken 验证管理员JWT令牌
func (s *AdminService) ValidateToken(tokenString string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 确保签名方法是HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.AdminToken), nil
	})

	if err != nil {
		return err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// 检查是否是管理员token
		if isAdmin, exists := claims["is_admin"]; !exists || !isAdmin.(bool) {
			return errors.New("不是管理员token")
		}
		return nil
	}

	return errors.New("无效的token")
}

// ResetUserPassword 重置用户密码 - 使用统一的认证服务
func (s *AdminService) ResetUserPassword(userID uint, newPassword string) error {
	return s.authService.ResetUserPassword(userID, newPassword)
}

// CleanExpiredFiles 清理过期文件
func (s *AdminService) CleanExpiredFiles() (int, error) {
	// 获取过期文件
	expiredFiles, err := s.daoManager.FileCode.GetExpiredFiles()
	if err != nil {
		return 0, err
	}

	count := 0
	for _, file := range expiredFiles {
		// 删除实际文件
		storageInterface := s.storageManager.GetStorage()
		if err := storageInterface.DeleteFile(&file); err != nil {
			fmt.Printf("Warning: Failed to delete physical file: %v\n", err)
		}

		// 删除数据库记录
		if err := s.daoManager.FileCode.DeleteByFileCode(&file); err != nil {
			fmt.Printf("Warning: Failed to delete file record: %v\n", err)
		} else {
			count++
		}
	}

	return count, nil
}

// ServeFile 提供文件服务（导出方法）
func (s *AdminService) ServeFile(c *gin.Context, fileCode *models.FileCode) error {
	return s.serveFile(c, fileCode)
}

// GetUserStats 获取用户统计信息
func (s *AdminService) GetUserStats(userID uint) (map[string]interface{}, error) {
	user, err := s.daoManager.User.GetByID(userID)
	if err != nil {
		return nil, err
	}

	// 获取文件数量
	fileCount, err := s.daoManager.FileCode.CountByUserID(userID)
	if err != nil {
		return nil, err
	}

	// 获取今日上传数量
	files, err := s.daoManager.FileCode.GetFilesByUserID(userID)
	if err != nil {
		return nil, err
	}

	// 计算今日上传数量
	today := time.Now().Truncate(24 * time.Hour)
	var todayUploads int64
	for _, file := range files {
		if file.CreatedAt.After(today) {
			todayUploads++
		}
	}

	stats := map[string]interface{}{
		"total_uploads":      user.TotalUploads,
		"total_downloads":    user.TotalDownloads,
		"total_storage":      user.TotalStorage,
		"total_files":        fileCount,
		"today_uploads":      todayUploads,
		"max_upload_size":    user.MaxUploadSize,
		"max_storage_quota":  user.MaxStorageQuota,
		"storage_usage":      user.TotalStorage,
		"storage_percentage": float64(user.TotalStorage) / float64(user.MaxStorageQuota) * 100,
	}

	return stats, nil
}

// GetUserByID 获取用户信息（代理到基础方法）
func (s *AdminService) GetUserByID(userID uint) (*models.User, error) {
	return s.GetUser(userID)
}

// UpdateUserStatus 更新用户状态
func (s *AdminService) UpdateUserStatus(userID uint, isActive bool) error {
	status := "inactive"
	if isActive {
		status = "active"
	}

	updates := map[string]interface{}{
		"status": status,
	}
	return s.daoManager.User.UpdateColumns(userID, updates)
}

// GetUserFiles 获取用户文件列表
func (s *AdminService) GetUserFiles(userID uint, page, limit int) ([]models.FileCode, int64, error) {
	offset := (page - 1) * limit

	// 计算总数
	total, err := s.daoManager.FileCode.CountByUserID(userID)
	if err != nil {
		return nil, 0, err
	}

	// 获取文件列表
	files, err := s.daoManager.FileCode.GetFilesByUserID(userID)
	if err != nil {
		return nil, 0, err
	}

	// 手动分页
	start := offset
	end := start + limit
	if start > len(files) {
		return []models.FileCode{}, total, nil
	}
	if end > len(files) {
		end = len(files)
	}

	return files[start:end], total, nil
}
