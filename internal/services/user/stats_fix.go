package user

import (
	"fmt"
)

// RecalculateUserStats 重新计算用户统计数据
func (s *Service) RecalculateUserStats(userID uint) error {
	user, err := s.repositoryManager.User.GetByID(userID)
	if err != nil {
		return err
	}

	// 获取用户的所有文件
	files, err := s.repositoryManager.FileCode.GetFilesByUserID(userID)
	if err != nil {
		return err
	}

	// 重新计算统计数据
	var totalStorage int64 = 0
	totalUploads := len(files)

	for _, file := range files {
		totalStorage += file.Size
	}

	// 更新用户统计
	user.TotalUploads = totalUploads
	user.TotalStorage = totalStorage
	// 保持下载次数不变，因为我们没有下载历史记录

	return s.repositoryManager.User.Update(user)
}

// RecalculateAllUsersStats 重新计算所有用户的统计数据
func (s *Service) RecalculateAllUsersStats() error {
	// 获取所有用户（使用大分页获取所有用户）
	users, _, err := s.repositoryManager.User.GetAllUsers(1, 10000)
	if err != nil {
		return err
	}

	for _, user := range users {
		err := s.RecalculateUserStats(user.ID)
		if err != nil {
			fmt.Printf("Warning: Failed to recalculate stats for user %d: %v\n", user.ID, err)
		}
	}

	return nil
}
