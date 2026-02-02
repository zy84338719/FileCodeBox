// Package services 提供与原有代码的兼容性
package services

import (
	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/models"
	"github.com/zy84338719/filecodebox/internal/repository"
	"github.com/zy84338719/filecodebox/internal/services/admin"
	"github.com/zy84338719/filecodebox/internal/services/auth"
	"github.com/zy84338719/filecodebox/internal/services/chunk"
	"github.com/zy84338719/filecodebox/internal/services/share"
	"github.com/zy84338719/filecodebox/internal/services/user"
	"github.com/zy84338719/filecodebox/internal/storage"
)

// 兼容性类型别名
type AdminService = admin.Service
type AuthService = auth.Service
type ChunkService = chunk.Service
type ShareService = share.Service
type UserService = user.Service

type APIKeyAuthResult = user.APIKeyAuthResult

type AdminUserUpdateParams = admin.UserUpdateParams

// 导出auth包中的类型
type AuthClaims = auth.AuthClaims

// 请求结构别名，使用统一的 models 定义
type ShareTextRequest = models.ShareTextRequest
type ShareFileRequest = models.ShareFileRequest

// 兼容性构造函数
func NewAdminService(repositoryManager *repository.RepositoryManager, manager *config.ConfigManager, storageService *storage.ConcreteStorageService) *AdminService {
	return admin.NewService(repositoryManager, manager, storageService)
}

func NewAuthService(repositoryManager *repository.RepositoryManager, manager *config.ConfigManager) *AuthService {
	return auth.NewService(repositoryManager, manager)
}

func NewChunkService(repositoryManager *repository.RepositoryManager, manager *config.ConfigManager, storageService *storage.ConcreteStorageService) *ChunkService {
	return chunk.NewService(repositoryManager, manager, storageService)
}

func NewShareService(repositoryManager *repository.RepositoryManager, manager *config.ConfigManager, storageService *storage.ConcreteStorageService, userService *UserService) *ShareService {
	shareService := share.NewService(repositoryManager, manager, storageService)
	if userService != nil {
		shareService.SetUserService(userService)
	}
	return shareService
}

func NewUserService(repositoryManager *repository.RepositoryManager, manager *config.ConfigManager) *UserService {
	return user.NewService(repositoryManager, manager)
}
