package admin

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/zy84338719/filecodebox/internal/models"
	"github.com/zy84338719/filecodebox/internal/models/service"
	"github.com/zy84338719/filecodebox/internal/utils"

	"github.com/sirupsen/logrus"
)

// CleanupExpiredFiles 清理过期文件
func (s *Service) CleanupExpiredFiles() (int, error) {
	// 获取过期文件
	expiredFiles, err := s.repositoryManager.FileCode.GetExpiredFiles()
	if err != nil {
		return 0, err
	}

	count := 0
	for _, file := range expiredFiles {
		file := file
		// 删除实际文件
		result := s.storageService.DeleteFileWithResult(&file)
		if !result.Success {
			logrus.WithError(result.Error).
				WithField("code", file.Code).
				Warn("failed to delete expired file from storage")
		}

		// 删除数据库记录
		if err := s.repositoryManager.FileCode.DeleteByFileCode(&file); err != nil {
			logrus.WithError(err).
				WithField("code", file.Code).
				Warn("failed to delete expired file record")
		} else {
			count++
		}
	}

	return count, nil
}

// CleanupInvalidFiles 清理无效文件（数据库有记录但文件不存在）
func (s *Service) CleanupInvalidFiles() (int, error) {
	cleaned := 0
	page := 1
	const pageSize = 200

	for {
		files, total, err := s.repositoryManager.FileCode.List(page, pageSize, "")
		if err != nil {
			return cleaned, err
		}

		if len(files) == 0 {
			break
		}

		for idx := range files {
			file := files[idx]
			if file.Text != "" {
				continue
			}

			filePath := file.GetFilePath()
			if strings.TrimSpace(filePath) == "" {
				if err := s.repositoryManager.FileCode.DeleteByFileCode(&file); err != nil {
					logrus.WithError(err).
						WithField("code", file.Code).
						Warn("failed to delete file record with empty path")
					continue
				}
				cleaned++
				continue
			}

			if s.storageService.FileExists(&file) {
				continue
			}

			if err := s.repositoryManager.FileCode.DeleteByFileCode(&file); err != nil {
				logrus.WithError(err).
					WithField("code", file.Code).
					Warn("failed to delete orphan file record")
				continue
			}
			cleaned++
		}

		if int64(page*pageSize) >= total {
			break
		}
		page++
	}

	return cleaned, nil
}

// CleanupOrphanedFiles 清理孤儿文件（文件存在但数据库无记录）
func (s *Service) CleanupOrphanedFiles() (int, error) {
	// 这个功能需要存储服务支持列出所有文件
	// 目前暂不实现，可根据具体存储策略后续添加
	return 0, fmt.Errorf("orphaned file cleanup not implemented yet")
}

// CleanTempFiles 清理临时文件
func (s *Service) CleanTempFiles() (int, error) {
	cutoff := time.Now().Add(-24 * time.Hour)
	oldChunks, err := s.repositoryManager.Chunk.GetOldChunks(cutoff)
	if err != nil {
		return 0, err
	}

	if len(oldChunks) == 0 {
		return 0, nil
	}

	uploadIDSet := make(map[string]struct{})
	for _, chunk := range oldChunks {
		uploadID := chunk.UploadID
		if strings.TrimSpace(uploadID) == "" {
			continue
		}
		if _, exists := uploadIDSet[uploadID]; exists {
			continue
		}
		uploadIDSet[uploadID] = struct{}{}

		result := s.storageService.CleanChunksWithResult(uploadID)
		if !result.Success {
			logrus.WithError(result.Error).
				WithField("upload_id", uploadID).
				Warn("failed to clean temporary upload chunks")
		}
	}

	uploadIDs := make([]string, 0, len(uploadIDSet))
	for uploadID := range uploadIDSet {
		uploadIDs = append(uploadIDs, uploadID)
	}
	sort.Strings(uploadIDs)

	cleaned, err := s.repositoryManager.Chunk.DeleteChunksByUploadIDs(uploadIDs)
	if err != nil {
		return cleaned, err
	}

	return cleaned, nil
}

// OptimizeDatabase 优化数据库
func (s *Service) OptimizeDatabase() error {
	db := s.repositoryManager.DB()
	if db == nil {
		return errors.New("数据库连接不可用")
	}

	switch strings.ToLower(s.manager.Database.Type) {
	case "sqlite":
		if err := db.Exec("VACUUM").Error; err != nil {
			return fmt.Errorf("执行 VACUUM 失败: %w", err)
		}
		if err := db.Exec("ANALYZE").Error; err != nil {
			return fmt.Errorf("执行 ANALYZE 失败: %w", err)
		}
	case "mysql":
		tables := []string{"file_codes", "upload_chunks", "users", "user_sessions", "transfer_logs"}
		for _, table := range tables {
			stmt := fmt.Sprintf("ANALYZE TABLE %s", table)
			if err := db.Exec(stmt).Error; err != nil {
				return fmt.Errorf("分析表 %s 失败: %w", table, err)
			}
		}
	case "postgres", "postgresql":
		if err := db.Exec("VACUUM ANALYZE").Error; err != nil {
			return fmt.Errorf("执行 VACUUM ANALYZE 失败: %w", err)
		}
	default:
		return fmt.Errorf("不支持的数据库类型: %s", s.manager.Database.Type)
	}

	return nil
}

// AnalyzeDatabase 分析数据库
func (s *Service) AnalyzeDatabase() (*models.DatabaseStats, error) {
	// 获取基本统计信息
	totalFiles, err := s.repositoryManager.FileCode.Count()
	if err != nil {
		return nil, err
	}

	totalUsers, err := s.repositoryManager.User.Count()
	if err != nil {
		return nil, err
	}

	// 获取存储使用情况
	totalSize, err := s.repositoryManager.FileCode.GetTotalSize()
	if err != nil {
		return nil, err
	}

	return &models.DatabaseStats{
		TotalFiles:   totalFiles,
		TotalUsers:   totalUsers,
		TotalSize:    totalSize,
		DatabaseSize: s.getDatabaseSizeHumanReadable(),
	}, nil
}

// GetSystemLogs 获取系统日志
func (s *Service) GetSystemLogs(lines int) ([]string, error) {
	if lines <= 0 {
		lines = 200
	}

	logPath, err := s.resolveLogPath("system")
	if err != nil {
		return nil, err
	}

	return tailFile(logPath, lines)
}

// BackupDatabase 备份数据库
func (s *Service) BackupDatabase() (string, error) {
	if strings.ToLower(s.manager.Database.Type) != "sqlite" {
		return "", errors.New("当前仅支持 SQLite 数据库备份，请使用外部工具备份其他数据库")
	}

	sourcePath, err := s.resolveSQLitePath()
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(sourcePath); err != nil {
		return "", fmt.Errorf("数据库文件不存在: %w", err)
	}

	backupDir := filepath.Join(s.ensureDataPath(), "backup")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("创建备份目录失败: %w", err)
	}

	backupPath := filepath.Join(backupDir, fmt.Sprintf("filecodebox_%s.db", time.Now().Format("20060102150405")))

	if err := copyFile(sourcePath, backupPath); err != nil {
		return "", fmt.Errorf("备份数据库失败: %w", err)
	}

	return backupPath, nil
}

// GetStorageStatus 获取存储状态
func (s *Service) GetStorageStatus() (*models.StorageStatus, error) {
	// 获取存储使用情况
	totalSize, err := s.repositoryManager.FileCode.GetTotalSize()
	if err != nil {
		return nil, err
	}

	// 获取当前存储类型
	storageType := s.manager.Storage.Type

	details := service.AdminStorageDetail{
		UsedSpace: totalSize,
		Type:      storageType,
	}

	available := true

	if storageType == "local" {
		basePath := s.storageRoot()
		if usage, err := utils.GetUsagePercent(basePath); err == nil {
			details.UsagePercent = usage
		}
		if total, free, usable, err := utils.GetDiskUsageStats(basePath); err == nil {
			details.TotalSpace = int64(total)
			details.AvailableSpace = int64(usable)
			if total > 0 {
				details.UsagePercent = (float64(total-free) / float64(total)) * 100
			}
		} else {
			available = false
			logrus.WithError(err).
				WithField("path", basePath).
				Warn("failed to collect disk usage for storage path")
		}
	}

	return &models.StorageStatus{
		Type:      storageType,
		Status:    "active",
		Available: available,
		Details:   details,
	}, nil
}

// GetDiskUsage 获取磁盘使用情况
func (s *Service) GetDiskUsage() (*models.DiskUsage, error) {
	basePath := s.storageRoot()
	total, free, available, err := utils.GetDiskUsageStats(basePath)
	if err != nil {
		errMsg := err.Error()
		return &models.DiskUsage{
			StorageType: s.manager.Storage.Type,
			Success:     false,
			Error:       &errMsg,
		}, err
	}

	used := total - free
	usagePercent := 0.0
	if total > 0 {
		usagePercent = (float64(used) / float64(total)) * 100
	}

	return &models.DiskUsage{
		TotalSpace:     int64(total),
		UsedSpace:      int64(used),
		AvailableSpace: int64(available),
		UsagePercent:   usagePercent,
		StorageType:    s.manager.Storage.Type,
		Success:        true,
		Error:          nil,
	}, nil
}

// GetFileCount 获取文件总数
func (s *Service) GetFileCount() (int64, error) {
	return s.repositoryManager.FileCode.Count()
}

// GetUserCount 获取用户总数
func (s *Service) GetUserCount() (int64, error) {
	return s.repositoryManager.User.Count()
}

// GetStorageUsage 获取存储使用情况
func (s *Service) GetStorageUsage() (int64, error) {
	return s.repositoryManager.FileCode.GetTotalSize()
}

// GetPerformanceMetrics 获取性能指标
func (s *Service) GetPerformanceMetrics() (*models.PerformanceMetrics, error) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	memoryUsage := fmt.Sprintf("%.2f MB", float64(mem.Alloc)/1024.0/1024.0)
	cpuUsage := fmt.Sprintf("goroutines: %d", runtime.NumGoroutine())
	responseTime := "-"

	dbStats := "-"
	if db := s.repositoryManager.DB(); db != nil {
		if sqlDB, err := db.DB(); err == nil {
			stats := sqlDB.Stats()
			dbStats = fmt.Sprintf("open=%d idle=%d inUse=%d waitCount=%d", stats.OpenConnections, stats.Idle, stats.InUse, stats.WaitCount)
		}
	}

	return &models.PerformanceMetrics{
		MemoryUsage:   memoryUsage,
		CPUUsage:      cpuUsage,
		ResponseTime:  responseTime,
		LastUpdated:   time.Now(),
		DatabaseStats: dbStats,
	}, nil
}

// ClearSystemCache 清理系统缓存
func (s *Service) ClearSystemCache() error {
	cacheDir := filepath.Join(s.ensureDataPath(), "cache")
	if _, err := os.Stat(cacheDir); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if err := os.RemoveAll(cacheDir); err != nil {
		return fmt.Errorf("清理系统缓存失败: %w", err)
	}
	return nil
}

// ClearUploadCache 清理上传缓存
func (s *Service) ClearUploadCache() error {
	chunkDir := filepath.Join(s.storageRoot(), "chunks")
	if err := os.RemoveAll(chunkDir); err != nil {
		return fmt.Errorf("清理上传缓存失败: %w", err)
	}
	return os.MkdirAll(chunkDir, 0755)
}

// ClearDownloadCache 清理下载缓存
func (s *Service) ClearDownloadCache() error {
	downloadDir := filepath.Join(s.storageRoot(), "downloads")
	if err := os.RemoveAll(downloadDir); err != nil {
		return fmt.Errorf("清理下载缓存失败: %w", err)
	}
	return os.MkdirAll(downloadDir, 0755)
}

// GetSystemInfo 获取系统信息
func (s *Service) GetSystemInfo() (*models.SystemInfo, error) {
	start := time.Now()
	if strings.TrimSpace(s.SysStart) != "" {
		if ms, err := strconv.ParseInt(s.SysStart, 10, 64); err == nil {
			start = time.UnixMilli(ms)
		}
	}

	uptime := time.Since(start).Truncate(time.Second)

	return &models.SystemInfo{
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
		GoVersion:    runtime.Version(),
		StartTime:    start,
		Uptime:       uptime.String(),
	}, nil
}

// ...existing code...

// CleanInvalidRecords 清理无效记录 (兼容性方法)
func (s *Service) CleanInvalidRecords() (int, error) {
	return s.CleanupInvalidFiles()
}

// ScanSecurity 安全扫描 (占位符实现)
func (s *Service) ScanSecurity() (*models.SecurityScanResult, error) {
	return &models.SecurityScanResult{
		Status:      "ok",
		Issues:      []string{},
		LastScanned: time.Now().Format("2006-01-02 15:04:05"),
		Passed:      true,
		Suggestions: []string{},
	}, nil
}

// CheckPermissions 检查权限 (占位符实现)
func (s *Service) CheckPermissions() (*models.PermissionCheckResult, error) {
	return &models.PermissionCheckResult{
		Status: "ok",
		Permissions: map[string]string{
			"read":  "granted",
			"write": "granted",
		},
		Issues: []string{},
	}, nil
}

// CheckIntegrity 检查完整性 (占位符实现)
func (s *Service) CheckIntegrity() (*models.IntegrityCheckResult, error) {
	return &models.IntegrityCheckResult{
		Status:       "ok",
		CheckedFiles: 0,
		CorruptFiles: 0,
		MissingFiles: 0,
		Issues:       []string{},
	}, nil
}

// ClearSystemLogs 清理系统日志 (占位符实现)
func (s *Service) ClearSystemLogs() (int, error) {
	return s.truncateLogFile("system")
}

// ClearAccessLogs 清理访问日志 (占位符实现)
func (s *Service) ClearAccessLogs() (int, error) {
	return s.truncateLogFile("access")
}

// ClearErrorLogs 清理错误日志 (占位符实现)
func (s *Service) ClearErrorLogs() (int, error) {
	return s.truncateLogFile("error")
}

// ExportLogs 导出日志 (占位符实现)
func (s *Service) ExportLogs(logType string) (string, error) {
	logPath, err := s.resolveLogPath(logType)
	if err != nil {
		return "", err
	}

	exportDir := filepath.Join(s.ensureDataPath(), "logs", "exports")
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		return "", fmt.Errorf("创建日志导出目录失败: %w", err)
	}

	fileName := fmt.Sprintf("%s_%s.log", logType, time.Now().Format("20060102150405"))
	destination := filepath.Join(exportDir, fileName)

	if err := copyFile(logPath, destination); err != nil {
		return "", fmt.Errorf("导出日志失败: %w", err)
	}

	return destination, nil
}

// GetLogStats 获取日志统计 (占位符实现)
func (s *Service) GetLogStats() (*models.LogStats, error) {
	logPath, err := s.resolveLogPath("system")
	if err != nil {
		return nil, err
	}

	file, err := os.Open(logPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var total, errorCount, warnCount, infoCount int
	var lastLine string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		total++
		upper := strings.ToUpper(line)
		switch {
		case strings.Contains(upper, "ERROR"):
			errorCount++
		case strings.Contains(upper, "WARN"):
			warnCount++
		case strings.Contains(upper, "INFO"):
			infoCount++
		}
		lastLine = line
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	lastLogTime := ""
	if ts := extractTimestamp(lastLine); ts != "" {
		lastLogTime = ts
	}

	return &models.LogStats{
		TotalLogs:   total,
		ErrorLogs:   errorCount,
		WarningLogs: warnCount,
		InfoLogs:    infoCount,
		LastLogTime: lastLogTime,
		LogSize:     formatBytes(uint64(fileInfo.Size())),
	}, nil
}

// GetRunningTasks 获取运行中的任务
func (s *Service) GetRunningTasks() ([]*models.RunningTask, error) {
	// 简单的任务列表实现，实际应该集成真实的任务系统
	tasks := []*models.RunningTask{
		{
			ID:          "task001",
			Type:        "cleanup",
			Status:      "running",
			Progress:    50,
			StartTime:   time.Now().Add(-time.Hour).Format("2006-01-02 15:04:05"),
			Description: "这是一个示例运行任务",
		},
	}
	return tasks, nil
}

// CancelTask 取消任务
func (s *Service) CancelTask(taskID string) error {
	// 这里应该实现真实的任务取消逻辑
	logrus.WithField("task_id", taskID).Info("Cancelling task")
	return nil
}

// RetryTask 重试任务
func (s *Service) RetryTask(taskID string) error {
	// 这里应该实现真实的任务重试逻辑
	logrus.WithField("task_id", taskID).Info("Retrying task")
	return nil
}

// RestartSystem 重启系统
func (s *Service) RestartSystem() error {
	// 这里应该实现真实的系统重启逻辑
	// 注意：这是一个危险操作，需要仔细考虑
	logrus.Info("System restart request received")
	return nil
}

func (s *Service) storageRoot() string {
	if path := strings.TrimSpace(s.manager.Storage.StoragePath); path != "" {
		if filepath.IsAbs(path) {
			return path
		}
		return filepath.Join(s.ensureDataPath(), path)
	}
	return s.ensureDataPath()
}

func (s *Service) ensureDataPath() string {
	base := strings.TrimSpace(s.manager.Base.DataPath)
	if base == "" {
		base = "./data"
	}
	if abs, err := filepath.Abs(base); err == nil {
		return abs
	}
	return base
}

func (s *Service) resolveSQLitePath() (string, error) {
	path := strings.TrimSpace(s.manager.Database.Name)
	if path == "" {
		path = filepath.Join(s.ensureDataPath(), "filecodebox.db")
	} else if !filepath.IsAbs(path) {
		path = filepath.Join(s.ensureDataPath(), path)
	}
	return filepath.Abs(path)
}

func (s *Service) resolveLogPath(logType string) (string, error) {
	nameCandidates := map[string][]string{
		"system": {"system.log", "server.log", "filecodebox.log"},
		"access": {"access.log", "server.log"},
		"error":  {"error.log", "server.log"},
	}

	names, ok := nameCandidates[logType]
	if !ok {
		names = []string{"server.log"}
	}

	searchRoots := []string{
		s.ensureDataPath(),
		filepath.Join(s.ensureDataPath(), "logs"),
		"./",
		"./logs",
	}

	for _, root := range searchRoots {
		for _, name := range names {
			candidate := filepath.Join(root, name)
			if _, err := os.Stat(candidate); err == nil {
				return candidate, nil
			}
		}
	}

	return "", fmt.Errorf("未找到 %s 日志文件", logType)
}

func tailFile(path string, limit int) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	if limit <= 0 {
		limit = 200
	}

	buffer := make([]string, 0, limit)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		buffer = append(buffer, scanner.Text())
		if len(buffer) > limit {
			buffer = buffer[1:]
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return buffer, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}

	return out.Sync()
}

func (s *Service) truncateLogFile(logType string) (int, error) {
	logPath, err := s.resolveLogPath(logType)
	if err != nil {
		return 0, err
	}

	info, err := os.Stat(logPath)
	if err != nil {
		return 0, err
	}

	if err := os.Truncate(logPath, 0); err != nil {
		return 0, err
	}

	return int(info.Size()), nil
}

func (s *Service) getDatabaseSizeHumanReadable() string {
	if strings.ToLower(s.manager.Database.Type) != "sqlite" {
		return "-"
	}
	path, err := s.resolveSQLitePath()
	if err != nil {
		return "-"
	}
	info, err := os.Stat(path)
	if err != nil {
		return "-"
	}
	return formatBytes(uint64(info.Size()))
}

func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		exp++
		div *= unit
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

func extractTimestamp(line string) string {
	if line == "" {
		return ""
	}
	if idx := strings.Index(line, "time=\""); idx >= 0 {
		rest := line[idx+len("time=\""):]
		if end := strings.Index(rest, "\""); end > 0 {
			return rest[:end]
		}
	}
	fields := strings.Fields(line)
	if len(fields) > 0 {
		return fields[0]
	}
	return ""
}
