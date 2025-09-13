package user

import (
	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/repository"
	"github.com/zy84338719/filecodebox/internal/services/auth"
)

// Service 用户服务
type Service struct {
	repositoryManager *repository.RepositoryManager
	manager           *config.ConfigManager
	authService       *auth.Service
}

// NewService 创建用户服务
func NewService(repositoryManager *repository.RepositoryManager, manager *config.ConfigManager) *Service {
	return &Service{
		repositoryManager: repositoryManager,
		manager:           manager,
		authService:       auth.NewService(repositoryManager, manager),
	}
}

// GetAdminUserCount 获取管理员用户数量
func (s *Service) GetAdminUserCount() (int64, error) {
	return s.repositoryManager.User.CountAdminUsers()
}
