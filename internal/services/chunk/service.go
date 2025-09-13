package chunk

import (
	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/repository"
	"github.com/zy84338719/filecodebox/internal/storage"
)

// Service 分块服务
type Service struct {
	repositoryManager *repository.RepositoryManager
	manager           *config.ConfigManager
	storageService    *storage.ConcreteStorageService
}

// NewService 创建分块服务
func NewService(repositoryManager *repository.RepositoryManager, manager *config.ConfigManager, storageService *storage.ConcreteStorageService) *Service {
	return &Service{
		repositoryManager: repositoryManager,
		manager:           manager,
		storageService:    storageService,
	}
}
