package admin

import (
	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/repository"
	"github.com/zy84338719/filecodebox/internal/services/auth"
	"github.com/zy84338719/filecodebox/internal/storage"
)

// Service 管理员服务
type Service struct {
	manager           *config.ConfigManager
	storageService    *storage.ConcreteStorageService
	repositoryManager *repository.RepositoryManager
	authService       *auth.Service
}

// NewService 创建管理员服务
func NewService(repositoryManager *repository.RepositoryManager, manager *config.ConfigManager, storageService *storage.ConcreteStorageService) *Service {
	return &Service{
		manager:           manager,
		storageService:    storageService,
		repositoryManager: repositoryManager,
		authService:       auth.NewService(repositoryManager, manager),
	}
}
