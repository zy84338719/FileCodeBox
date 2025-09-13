package auth

import (
	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/repository"
)

// Service 认证服务
type Service struct {
	repositoryManager *repository.RepositoryManager
	manager           *config.ConfigManager
}

// NewService 创建认证服务
func NewService(repositoryManager *repository.RepositoryManager, manager *config.ConfigManager) *Service {
	return &Service{
		repositoryManager: repositoryManager,
		manager:           manager,
	}
}
