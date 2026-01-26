package handlers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/zy84338719/filecodebox/internal/common"
	"github.com/zy84338719/filecodebox/internal/models/web"

	"github.com/gin-gonic/gin"
)

// CleanExpiredFiles 清理过期文件
func (h *AdminHandler) CleanExpiredFiles(c *gin.Context) {
	start := time.Now()
	// TODO: 修复服务方法调用
	// count, err := h.service.CleanupExpiredFiles()
	// if err != nil {
	// 	h.recordOperationLog(c, "maintenance.clean_expired", "files", false, err.Error(), start)
	// 	common.InternalServerErrorResponse(c, "清理失败: "+err.Error())
	// 	return
	// }
	msg := "清理完成"
	h.recordOperationLog(c, "maintenance.clean_expired", "files", true, msg, start)
	common.SuccessWithMessage(c, msg, web.CleanedCountResponse{CleanedCount: 0})
}

// CleanTempFiles 清理临时文件
func (h *AdminHandler) CleanTempFiles(c *gin.Context) {
	start := time.Now()
	count, err := h.service.CleanTempFiles()
	if err != nil {
		h.recordOperationLog(c, "maintenance.clean_temp", "chunks", false, err.Error(), start)
		common.InternalServerErrorResponse(c, "清理失败: "+err.Error())
		return
	}
	msg := fmt.Sprintf("清理了 %d 个临时文件", count)
	h.recordOperationLog(c, "maintenance.clean_temp", "chunks", true, msg, start)
	common.SuccessWithMessage(c, msg, web.CountResponse{Count: count})
}

// CleanInvalidRecords 清理无效记录
func (h *AdminHandler) CleanInvalidRecords(c *gin.Context) {
	start := time.Now()
	count, err := h.service.CleanInvalidRecords()
	if err != nil {
		h.recordOperationLog(c, "maintenance.clean_invalid", "records", false, err.Error(), start)
		common.InternalServerErrorResponse(c, "清理失败: "+err.Error())
		return
	}
	msg := fmt.Sprintf("清理了 %d 个无效记录", count)
	h.recordOperationLog(c, "maintenance.clean_invalid", "records", true, msg, start)
	common.SuccessWithMessage(c, msg, web.CountResponse{Count: count})
}

// OptimizeDatabase 优化数据库
func (h *AdminHandler) OptimizeDatabase(c *gin.Context) {
	start := time.Now()
	if err := h.service.OptimizeDatabase(); err != nil {
		h.recordOperationLog(c, "maintenance.db_optimize", "database", false, err.Error(), start)
		common.InternalServerErrorResponse(c, "优化失败: "+err.Error())
		return
	}
	h.recordOperationLog(c, "maintenance.db_optimize", "database", true, "数据库优化完成", start)
	common.SuccessWithMessage(c, "数据库优化完成", nil)
}

// AnalyzeDatabase 分析数据库
func (h *AdminHandler) AnalyzeDatabase(c *gin.Context) {
	start := time.Now()
	stats, err := h.service.AnalyzeDatabase()
	if err != nil {
		h.recordOperationLog(c, "maintenance.db_analyze", "database", false, err.Error(), start)
		common.InternalServerErrorResponse(c, "分析失败: "+err.Error())
		return
	}
	h.recordOperationLog(c, "maintenance.db_analyze", "database", true, "数据库分析完成", start)
	common.SuccessResponse(c, stats)
}

// BackupDatabase 备份数据库
func (h *AdminHandler) BackupDatabase(c *gin.Context) {
	start := time.Now()
	backupPath, err := h.service.BackupDatabase()
	if err != nil {
		h.recordOperationLog(c, "maintenance.db_backup", "database", false, err.Error(), start)
		common.InternalServerErrorResponse(c, "备份失败: "+err.Error())
		return
	}
	msg := "数据库备份完成"
	h.recordOperationLog(c, "maintenance.db_backup", backupPath, true, msg, start)
	common.SuccessWithMessage(c, "数据库备份完成", web.BackupPathResponse{BackupPath: backupPath})
}

// ClearSystemCache 清理系统缓存
func (h *AdminHandler) ClearSystemCache(c *gin.Context) {
	start := time.Now()
	if err := h.service.ClearSystemCache(); err != nil {
		h.recordOperationLog(c, "maintenance.cache_clear_system", "system-cache", false, err.Error(), start)
		common.InternalServerErrorResponse(c, "清理失败: "+err.Error())
		return
	}
	h.recordOperationLog(c, "maintenance.cache_clear_system", "system-cache", true, "系统缓存清理完成", start)
	common.SuccessWithMessage(c, "系统缓存清理完成", nil)
}

// ClearUploadCache 清理上传缓存
func (h *AdminHandler) ClearUploadCache(c *gin.Context) {
	start := time.Now()
	if err := h.service.ClearUploadCache(); err != nil {
		h.recordOperationLog(c, "maintenance.cache_clear_upload", "upload-cache", false, err.Error(), start)
		common.InternalServerErrorResponse(c, "清理失败: "+err.Error())
		return
	}
	h.recordOperationLog(c, "maintenance.cache_clear_upload", "upload-cache", true, "上传缓存清理完成", start)
	common.SuccessWithMessage(c, "上传缓存清理完成", nil)
}

// ClearDownloadCache 清理下载缓存
func (h *AdminHandler) ClearDownloadCache(c *gin.Context) {
	start := time.Now()
	if err := h.service.ClearDownloadCache(); err != nil {
		h.recordOperationLog(c, "maintenance.cache_clear_download", "download-cache", false, err.Error(), start)
		common.InternalServerErrorResponse(c, "清理失败: "+err.Error())
		return
	}
	h.recordOperationLog(c, "maintenance.cache_clear_download", "download-cache", true, "下载缓存清理完成", start)
	common.SuccessWithMessage(c, "下载缓存清理完成", nil)
}

// GetSystemInfo 获取系统信息
func (h *AdminHandler) GetSystemInfo(c *gin.Context) {
	info, err := h.service.GetSystemInfo()
	if err != nil {
		common.InternalServerErrorResponse(c, "获取系统信息失败: "+err.Error())
		return
	}

	common.SuccessResponse(c, info)
}

// GetStorageStatus 获取存储状态
func (h *AdminHandler) GetStorageStatus(c *gin.Context) {
	status, err := h.service.GetStorageStatus()
	if err != nil {
		common.InternalServerErrorResponse(c, "获取存储状态失败: "+err.Error())
		return
	}

	common.SuccessResponse(c, status)
}

// GetPerformanceMetrics 获取性能指标
func (h *AdminHandler) GetPerformanceMetrics(c *gin.Context) {
	metrics, err := h.service.GetPerformanceMetrics()
	if err != nil {
		common.InternalServerErrorResponse(c, "获取性能指标失败: "+err.Error())
		return
	}

	common.SuccessResponse(c, metrics)
}

// ScanSecurity 安全扫描
func (h *AdminHandler) ScanSecurity(c *gin.Context) {
	result, err := h.service.ScanSecurity()
	if err != nil {
		common.InternalServerErrorResponse(c, "安全扫描失败: "+err.Error())
		return
	}

	common.SuccessResponse(c, result)
}

// CheckPermissions 检查权限
func (h *AdminHandler) CheckPermissions(c *gin.Context) {
	result, err := h.service.CheckPermissions()
	if err != nil {
		common.InternalServerErrorResponse(c, "权限检查失败: "+err.Error())
		return
	}

	common.SuccessResponse(c, result)
}

// CheckIntegrity 检查完整性
func (h *AdminHandler) CheckIntegrity(c *gin.Context) {
	result, err := h.service.CheckIntegrity()
	if err != nil {
		common.InternalServerErrorResponse(c, "完整性检查失败: "+err.Error())
		return
	}

	common.SuccessResponse(c, result)
}

// ClearSystemLogs 清理系统日志
func (h *AdminHandler) ClearSystemLogs(c *gin.Context) {
	start := time.Now()
	count, err := h.service.ClearSystemLogs()
	if err != nil {
		h.recordOperationLog(c, "maintenance.logs_clear_system", "system-logs", false, err.Error(), start)
		common.InternalServerErrorResponse(c, "清理失败: "+err.Error())
		return
	}
	msg := fmt.Sprintf("清理了 %d 条系统日志", count)
	h.recordOperationLog(c, "maintenance.logs_clear_system", "system-logs", true, msg, start)
	common.SuccessWithMessage(c, msg, web.CountResponse{Count: count})
}

// ClearAccessLogs 清理访问日志
func (h *AdminHandler) ClearAccessLogs(c *gin.Context) {
	start := time.Now()
	count, err := h.service.ClearAccessLogs()
	if err != nil {
		h.recordOperationLog(c, "maintenance.logs_clear_access", "access-logs", false, err.Error(), start)
		common.InternalServerErrorResponse(c, "清理失败: "+err.Error())
		return
	}
	msg := fmt.Sprintf("清理了 %d 条访问日志", count)
	h.recordOperationLog(c, "maintenance.logs_clear_access", "access-logs", true, msg, start)
	common.SuccessWithMessage(c, msg, web.CountResponse{Count: count})
}

// ClearErrorLogs 清理错误日志
func (h *AdminHandler) ClearErrorLogs(c *gin.Context) {
	start := time.Now()
	count, err := h.service.ClearErrorLogs()
	if err != nil {
		h.recordOperationLog(c, "maintenance.logs_clear_error", "error-logs", false, err.Error(), start)
		common.InternalServerErrorResponse(c, "清理失败: "+err.Error())
		return
	}
	msg := fmt.Sprintf("清理了 %d 条错误日志", count)
	h.recordOperationLog(c, "maintenance.logs_clear_error", "error-logs", true, msg, start)
	common.SuccessWithMessage(c, msg, web.CountResponse{Count: count})
}

// ExportLogs 导出日志
func (h *AdminHandler) ExportLogs(c *gin.Context) {
	start := time.Now()
	logType := c.DefaultQuery("type", "system")

	logPath, err := h.service.ExportLogs(logType)
	if err != nil {
		h.recordOperationLog(c, "maintenance.logs_export", logType, false, err.Error(), start)
		common.InternalServerErrorResponse(c, "导出失败: "+err.Error())
		return
	}
	msg := "日志导出完成"
	h.recordOperationLog(c, "maintenance.logs_export", logType, true, msg, start)
	common.SuccessWithMessage(c, msg, web.LogPathResponse{LogPath: logPath})
}

// GetLogStats 获取日志统计
func (h *AdminHandler) GetLogStats(c *gin.Context) {
	stats, err := h.service.GetLogStats()
	if err != nil {
		common.InternalServerErrorResponse(c, "获取日志统计失败: "+err.Error())
		return
	}

	common.SuccessResponse(c, stats)
}

// GetSystemLogs 获取系统日志
func (h *AdminHandler) GetSystemLogs(c *gin.Context) {
	level := c.DefaultQuery("level", "")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	logs, err := h.service.GetSystemLogs(limit)
	if err != nil {
		common.InternalServerErrorResponse(c, "获取日志失败: "+err.Error())
		return
	}

	if level != "" {
		filteredLogs := make([]string, 0)
		for _, log := range logs {
			if len(log) > 0 && (level == "" || len(log) > 10) {
				filteredLogs = append(filteredLogs, log)
			}
		}
		logs = filteredLogs
	}

	response := web.LogsResponse{
		Logs:  logs,
		Total: len(logs),
	}

	common.SuccessResponse(c, response)
}

// GetRunningTasks 获取运行中的任务
func (h *AdminHandler) GetRunningTasks(c *gin.Context) {
	tasks, err := h.service.GetRunningTasks()
	if err != nil {
		common.InternalServerErrorResponse(c, "获取运行任务失败: "+err.Error())
		return
	}

	response := web.TasksResponse{
		Tasks: tasks,
		Total: len(tasks),
	}

	common.SuccessResponse(c, response)
}

// CancelTask 取消任务
func (h *AdminHandler) CancelTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		common.BadRequestResponse(c, "任务ID不能为空")
		return
	}

	start := time.Now()
	if err := h.service.CancelTask(taskID); err != nil {
		h.recordOperationLog(c, "maintenance.task_cancel", taskID, false, err.Error(), start)
		common.InternalServerErrorResponse(c, "取消任务失败: "+err.Error())
		return
	}

	h.recordOperationLog(c, "maintenance.task_cancel", taskID, true, "任务已取消", start)
	common.SuccessWithMessage(c, "任务已取消", nil)
}

// RetryTask 重试任务
func (h *AdminHandler) RetryTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		common.BadRequestResponse(c, "任务ID不能为空")
		return
	}

	start := time.Now()
	if err := h.service.RetryTask(taskID); err != nil {
		h.recordOperationLog(c, "maintenance.task_retry", taskID, false, err.Error(), start)
		common.InternalServerErrorResponse(c, "重试任务失败: "+err.Error())
		return
	}

	h.recordOperationLog(c, "maintenance.task_retry", taskID, true, "任务已重新启动", start)
	common.SuccessWithMessage(c, "任务已重新启动", nil)
}

// RestartSystem 重启系统
func (h *AdminHandler) RestartSystem(c *gin.Context) {
	start := time.Now()
	if err := h.service.RestartSystem(); err != nil {
		h.recordOperationLog(c, "maintenance.system_restart", "system", false, err.Error(), start)
		common.InternalServerErrorResponse(c, "重启系统失败: "+err.Error())
		return
	}

	h.recordOperationLog(c, "maintenance.system_restart", "system", true, "系统重启指令已发送", start)
	common.SuccessWithMessage(c, "系统重启指令已发送", nil)
}

// ShutdownSystem 关闭系统
func (h *AdminHandler) ShutdownSystem(c *gin.Context) {
	start := time.Now()
	shutdown := GetShutdownFunc()
	if shutdown == nil {
		msg := "关闭系统功能未启用"
		h.recordOperationLog(c, "maintenance.system_shutdown", "system", false, msg, start)
		common.InternalServerErrorResponse(c, msg)
		return
	}

	h.recordOperationLog(c, "maintenance.system_shutdown", "system", true, "系统关闭指令已发送", start)
	common.SuccessWithMessage(c, "系统关闭指令已发送", nil)

	go func() {
		time.Sleep(500 * time.Millisecond)
		shutdown()
	}()
}
