package admin

import (
	"context"
	"errors"
	"runtime"
	"time"

	"github.com/zy84338719/fileCodeBox/internal/pkg/auth"
	"github.com/zy84338719/fileCodeBox/internal/repo/db/dao"
	"github.com/zy84338719/fileCodeBox/internal/repo/db/model"
	"golang.org/x/crypto/bcrypt"
)

type AdminStats struct {
	TotalUsers     int64 `json:"total_users"`
	TotalFiles     int64 `json:"total_files"`
	TotalSize      int64 `json:"total_size"`
	TodayUploads   int64 `json:"today_uploads"`
	TodayDownloads int64 `json:"today_downloads"`
	ExpiredFiles   int64 `json:"expired_files"`
	AnonymousFiles int64 `json:"anonymous_files"`
	UserFiles      int64 `json:"user_files"`
}

type SystemConfig struct {
	Base struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Port        int    `json:"port"`
	} `json:"base"`
	
	Storage struct {
		Type    string `json:"type"`
		MaxSize int64  `json:"max_size"`
	} `json:"storage"`
	
	Transfer struct {
		MaxCount     int `json:"max_count"`
		ExpireDefault int `json:"expire_default"`
	} `json:"transfer"`
}

type Service struct {
	userRepo           *dao.UserRepository
	fileCodeRepo       *dao.FileCodeRepository
	transferLogRepo    *dao.TransferLogRepository
	adminOperationRepo *dao.AdminOperationLogRepository
	chunkRepo          *dao.ChunkRepository
	config             *SystemConfig
}

func NewService() *Service {
	return &Service{
		userRepo:           dao.NewUserRepository(),
		fileCodeRepo:       dao.NewFileCodeRepository(),
		transferLogRepo:    dao.NewTransferLogRepository(),
		adminOperationRepo: dao.NewAdminOperationLogRepository(),
		chunkRepo:          dao.NewChunkRepository(),
		config:             &SystemConfig{}, // 默认配置
	}
}

// SetConfig 设置配置
func (s *Service) SetConfig(config *SystemConfig) {
	s.config = config
}

// GetStats 获取管理员统计信息
func (s *Service) GetStats(ctx context.Context) (*AdminStats, error) {
	stats := &AdminStats{}

	// 获取文件统计
	totalFiles, err := s.fileCodeRepo.Count(ctx)
	if err == nil {
		stats.TotalFiles = totalFiles
	}

	totalSize, err := s.fileCodeRepo.GetTotalSize(ctx)
	if err == nil {
		stats.TotalSize = totalSize
	}

	todayUploads, err := s.fileCodeRepo.CountTodayUploads(ctx)
	if err == nil {
		stats.TodayUploads = todayUploads
	}

	// 获取用户统计
	users, err := s.userRepo.Count(ctx)
	if err == nil {
		stats.TotalUsers = users
	}

	// 获取今日下载（从 transfer_log 统计）
	// todayDownloads, err := s.transferLogRepo.CountTodayDownloads(ctx)
	// if err == nil {
	// 	stats.TodayDownloads = todayDownloads
	// }

	// 统计匿名上传和用户上传
	// anonymousFiles, _ := s.fileCodeRepo.CountByUploadType(ctx, "anonymous")
	// userFiles, _ := s.fileCodeRepo.CountByUploadType(ctx, "authenticated")
	// stats.AnonymousFiles = anonymousFiles
	// stats.UserFiles = userFiles

	return stats, nil
}

// GetUsers 获取用户列表
func (s *Service) GetUsers(ctx context.Context, page, pageSize int) ([]*model.UserResp, int64, error) {
	users, total, err := s.userRepo.List(ctx, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	resps := make([]*model.UserResp, len(users))
	for i, user := range users {
		resps[i] = user.ToResp()
	}

	return resps, total, nil
}

// DeleteUser 删除用户
func (s *Service) DeleteUser(ctx context.Context, userID uint) error {
	// 1. 删除用户的所有文件
	// files, _ := s.fileCodeRepo.GetByUserID(ctx, userID)
	// for _, file := range files {
	// 	s.DeleteFile(ctx, file.ID)
	// }

	// 2. 删除用户记录
	return s.userRepo.Delete(ctx, userID)
}

// GetFiles 获取文件列表
func (s *Service) GetFiles(ctx context.Context, page, pageSize int, search string) ([]*model.FileCode, int64, error) {
	return s.fileCodeRepo.List(ctx, page, pageSize, search)
}

// DeleteFile 删除文件
func (s *Service) DeleteFile(ctx context.Context, fileID uint) error {
	// 1. 获取文件信息
	// file, err := s.fileCodeRepo.GetByID(ctx, fileID)
	// if err != nil {
	// 	return err
	// }

	// 2. 删除物理文件（如果实现了存储服务）
	// if s.storageService != nil && file.FilePath != "" {
	// 	s.storageService.DeleteFile(ctx, file.FilePath)
	// }

	// 3. 删除数据库记录
	return s.fileCodeRepo.Delete(ctx, fileID)
}

// GetTransferLogs 获取传输日志
func (s *Service) GetTransferLogs(ctx context.Context, query model.TransferLogQuery) ([]*model.TransferLog, int64, error) {
	return s.transferLogRepo.List(ctx, query)
}

// CleanupExpiredFiles 清理过期文件
func (s *Service) CleanupExpiredFiles(ctx context.Context) (int, error) {
	// 获取过期文件
	expiredFiles, err := s.fileCodeRepo.GetExpiredFiles(ctx)
	if err != nil {
		return 0, err
	}

	// 删除过期文件
	deletedCount, err := s.fileCodeRepo.DeleteExpiredFiles(ctx, expiredFiles)
	if err != nil {
		return 0, err
	}

	// TODO: 记录管理员操作日志
	// s.logAdminOperation(ctx, "maintenance.clean_expired_files", fmt.Sprintf("Cleaned up %d expired files", deletedCount), true)

	return deletedCount, nil
}

// CleanupIncompleteUploads 清理未完成的上传
func (s *Service) CleanupIncompleteUploads(ctx context.Context, olderThanHours int) (int, error) {
	// 获取未完成的上传
	incompleteUploads, err := s.chunkRepo.GetIncompleteUploads(ctx, olderThanHours)
	if err != nil {
		return 0, err
	}

	uploadIDs := make([]string, len(incompleteUploads))
	for i, upload := range incompleteUploads {
		uploadIDs[i] = upload.UploadID
	}

	// 删除未完成的上传记录
	deletedCount, err := s.chunkRepo.DeleteChunksByUploadIDs(ctx, uploadIDs)
	if err != nil {
		return 0, err
	}

	// TODO: 记录管理员操作日志
	// s.logAdminOperation(ctx, "maintenance.clean_incomplete_uploads", fmt.Sprintf("Cleaned up %d incomplete uploads", deletedCount), true)

	return deletedCount, nil
}

// TODO: 创建 AdminOperationLogRepository 和相关方法
// func (s *Service) logAdminOperation(ctx context.Context, action, target string, success bool) {
// 	// 记录管理员操作日志
// }

// GetConfig 获取系统配置
func (s *Service) GetConfig(ctx context.Context) (*SystemConfig, error) {
	// TODO: 从数据库或配置文件读取
	// 暂时返回默认配置
	if s.config == nil {
		s.config = &SystemConfig{
			Base: struct {
				Name        string `json:"name"`
				Description string `json:"description"`
				Port        int    `json:"port"`
			}{
				Name:        "FileCodeBox",
				Description: "文件分享平台",
				Port:        8888,
			},
			Storage: struct {
				Type    string `json:"type"`
				MaxSize int64  `json:"max_size"`
			}{
				Type:    "local",
				MaxSize: 1024 * 1024 * 1024, // 1GB
			},
			Transfer: struct {
				MaxCount     int `json:"max_count"`
				ExpireDefault int `json:"expire_default"`
			}{
				MaxCount:      100,
				ExpireDefault: 7, // 7天
			},
		}
	}
	return s.config, nil
}

// UpdateConfig 更新系统配置
func (s *Service) UpdateConfig(ctx context.Context, newConfig *SystemConfig) error {
	// TODO: 验证配置有效性
	// TODO: 保存到数据库或配置文件
	
	s.config = newConfig
	return nil
}

// UpdateUserStatus 更新用户状态
func (s *Service) UpdateUserStatus(ctx context.Context, userID uint, status string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	user.Status = status
	return s.userRepo.Update(ctx, user)
}

// GenerateTokenForAdmin 生成管理员登录 token
func (s *Service) GenerateTokenForAdmin(ctx context.Context, username, password string) (string, error) {
	// 查找用户
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return "", errors.New("用户名或密码错误")
	}

	// 检查是否为管理员
	if user.Role != "admin" {
		return "", errors.New("权限不足")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("用户名或密码错误")
	}

	// 检查用户状态
	if user.Status != "active" {
		return "", errors.New("用户已被禁用")
	}

	// 生成 token (24小时过期)
	token, err := auth.GenerateToken(user.ID, user.Username, "admin")
	if err != nil {
		return "", errors.New("生成 token 失败")
	}

	// 更新最后登录时间
	now := time.Now()
	user.LastLoginAt = &now
	_ = s.userRepo.Update(ctx, user)

	return token, nil
}

// ==================== 维护工具 API ====================

// CleanExpiredFiles 清理过期文件
func (s *Service) CleanExpiredFiles(ctx context.Context) (int64, int64, error) {
	// 获取过期文件
	expiredFiles, err := s.fileCodeRepo.GetExpiredFiles(ctx)
	if err != nil {
		return 0, 0, err
	}

	// 删除过期文件并计算释放的空间
	deletedCount := int64(0)
	freedSpace := int64(0)
	for _, file := range expiredFiles {
		freedSpace += file.Size
	}

	// 删除数据库记录
	count, err := s.fileCodeRepo.DeleteExpiredFiles(ctx, expiredFiles)
	if err != nil {
		return deletedCount, 0, err
	}
	deletedCount = int64(count)

	return deletedCount, freedSpace, nil
}

// CleanTempFiles 清理临时文件
func (s *Service) CleanTempFiles(ctx context.Context) (int64, int64, error) {
	// 获取24小时前的临时文件
	incompleteUploads, err := s.chunkRepo.GetIncompleteUploads(ctx, 24)
	if err != nil {
		return 0, 0, err
	}

	uploadIDs := make([]string, 0, len(incompleteUploads))
	for _, upload := range incompleteUploads {
		uploadIDs = append(uploadIDs, upload.UploadID)
	}

	// 删除未完成的上传记录
	deletedCount, err := s.chunkRepo.DeleteChunksByUploadIDs(ctx, uploadIDs)
	if err != nil {
		return 0, 0, err
	}

	return int64(deletedCount), 0, nil
}

// SystemInfo 系统信息
type SystemInfo struct {
	Version      string `json:"version"`
	OS           string `json:"os"`
	Arch         string `json:"arch"`
	Uptime       string `json:"uptime"`
	Goroutines   int64  `json:"goroutines"`
	MemoryAlloc  int64  `json:"memory_alloc"`
	MemoryTotal  int64  `json:"memory_total"`
	MemorySys    int64  `json:"memory_sys"`
}

// GetSystemInfo 获取系统信息
func (s *Service) GetSystemInfo(ctx context.Context) (*SystemInfo, error) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	uptime := time.Since(time.Now()).Truncate(time.Second)

	return &SystemInfo{
		Version:     "1.0.0",
		OS:          runtime.GOOS,
		Arch:        runtime.GOARCH,
		Uptime:      uptime.String(),
		Goroutines:  int64(runtime.NumGoroutine()),
		MemoryAlloc: int64(mem.Alloc),
		MemoryTotal: int64(mem.TotalAlloc),
		MemorySys:   int64(mem.Sys),
	}, nil
}

// StorageStatus 存储状态
type StorageStatus struct {
	StorageType string  `json:"storage_type"`
	TotalSpace  int64   `json:"total_space"`
	UsedSpace   int64   `json:"used_space"`
	FreeSpace   int64   `json:"free_space"`
	FileCount   int64   `json:"file_count"`
	UsagePercent float64 `json:"usage_percent"`
}

// GetStorageStatus 获取存储状态
func (s *Service) GetStorageStatus(ctx context.Context) (*StorageStatus, error) {
	// 获取文件总数和总大小
	totalFiles, err := s.fileCodeRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	totalSize, err := s.fileCodeRepo.GetTotalSize(ctx)
	if err != nil {
		return nil, err
	}

	// 简化实现，假设使用本地存储
	storageType := "local"
	totalSpace := int64(100 * 1024 * 1024 * 1024) // 100GB 默认
	freeSpace := totalSpace - totalSize
	usagePercent := float64(0)
	if totalSpace > 0 {
		usagePercent = (float64(totalSize) / float64(totalSpace)) * 100
	}

	return &StorageStatus{
		StorageType: storageType,
		TotalSpace:  totalSpace,
		UsedSpace:   totalSize,
		FreeSpace:   freeSpace,
		FileCount:   totalFiles,
		UsagePercent: usagePercent,
	}, nil
}

// LogEntry 日志条目
type LogEntry struct {
	ID        uint64 `json:"id"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
	Module    string `json:"module"`
	UserID    string `json:"user_id"`
}

// GetSystemLogs 获取系统日志
func (s *Service) GetSystemLogs(ctx context.Context, level string, page, pageSize int) ([]*LogEntry, int64, error) {
	// 暂时返回模拟数据
	// TODO: 实现真实的日志查询
	logs := []*LogEntry{}
	total := int64(0)

	if level == "" || level == "info" {
		logs = append(logs, &LogEntry{
			ID:        1,
			Level:     "info",
			Message:   "系统启动成功",
			CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
			Module:    "system",
			UserID:    "",
		})
		total = 1
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	return logs, total, nil
}
