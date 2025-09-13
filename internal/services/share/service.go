package share

import (
	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/repository"
	"github.com/zy84338719/filecodebox/internal/storage"
)

// UserServiceInterface 定义用户服务接口，避免循环依赖
type UserServiceInterface interface {
	UpdateUserStats(userID uint, statsType string, value int64) error
}

// Service 分享服务
type Service struct {
	repositoryManager *repository.RepositoryManager
	manager           *config.ConfigManager
	storageService    *storage.ConcreteStorageService
	userService       UserServiceInterface
}

// NewService 创建分享服务
func NewService(repositoryManager *repository.RepositoryManager, manager *config.ConfigManager, storageService *storage.ConcreteStorageService) *Service {
	return &Service{
		repositoryManager: repositoryManager,
		manager:           manager,
		storageService:    storageService,
		userService:       nil, // 将在初始化后设置
	}
}

// SetUserService 设置用户服务（用于避免循环依赖）
func (s *Service) SetUserService(userService UserServiceInterface) {
	s.userService = userService
}
