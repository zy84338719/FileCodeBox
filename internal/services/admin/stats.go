package admin

import (
	"fmt"
	"time"

	"github.com/zy84338719/filecodebox/internal/models/web"
)

// GetStats 获取统计信息
func (s *Service) GetStats() (*web.AdminStatsResponse, error) {
	stats := &web.AdminStatsResponse{}

	// 用户统计信息
	// 总用户数
	totalUsers, err := s.repositoryManager.User.Count()
	if err != nil {
		return nil, err
	}
	stats.TotalUsers = totalUsers

	// 活跃用户数
	activeUsers, err := s.repositoryManager.User.CountActive()
	if err != nil {
		return nil, err
	}
	stats.ActiveUsers = activeUsers

	// 今日注册用户数
	todayRegistrations, err := s.repositoryManager.User.CountTodayRegistrations()
	if err != nil {
		return nil, err
	}
	stats.TodayRegistrations = todayRegistrations

	// 今日上传文件数
	todayUploads, err := s.repositoryManager.FileCode.CountToday()
	if err != nil {
		return nil, err
	}
	stats.TodayUploads = todayUploads

	// 文件统计信息
	// 总文件数（不包括已删除的）
	totalFiles, err := s.repositoryManager.FileCode.Count()
	if err != nil {
		return nil, err
	}
	stats.TotalFiles = totalFiles

	// 活跃文件数（未过期且未删除）
	activeFiles, err := s.repositoryManager.FileCode.CountActive()
	if err != nil {
		return nil, err
	}
	stats.ActiveFiles = activeFiles

	// 总大小（不包括已删除的）
	totalSize, err := s.repositoryManager.FileCode.GetTotalSize()
	if err != nil {
		return nil, err
	}
	stats.TotalSize = totalSize

	// 系统启动时间
	sysStart, err := s.repositoryManager.KeyValue.GetByKey("sys_start")
	if err == nil {
		stats.SysStart = sysStart.Value
	} else {
		// 如果没有记录，创建一个
		startTime := fmt.Sprintf("%d", time.Now().UnixMilli())
		err := s.repositoryManager.KeyValue.SetValue("sys_start", startTime)
		if err != nil {
			return nil, fmt.Errorf("设置系统启动时间失败: %v", err)
		}
		stats.SysStart = startTime
	}

	return stats, nil
}
