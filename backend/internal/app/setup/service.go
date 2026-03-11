package setup

import (
	"context"
	"errors"
	"fmt"

	"github.com/zy84338719/fileCodeBox/backend/internal/repo/db/dao"
	"github.com/zy84338719/fileCodeBox/backend/internal/repo/db/model"
	"golang.org/x/crypto/bcrypt"
)

// InitializeSystemReq 系统初始化请求
type InitializeSystemReq struct {
	AdminUsername string
	AdminPassword string
	AdminEmail    string
}

type Service struct {
	userRepo *dao.UserRepository
}

func NewService() *Service {
	return &Service{
		userRepo: dao.NewUserRepository(),
	}
}

// IsSystemInitialized 检查系统是否已初始化
// 通过检查是否存在管理员用户来判断
func (s *Service) IsSystemInitialized(ctx context.Context) (bool, error) {
	count, err := s.userRepo.CountAdminUsers(ctx)
	if err != nil {
		return false, fmt.Errorf("查询管理员用户失败: %w", err)
	}
	return count > 0, nil
}

// InitializeSystem 初始化系统，创建第一个管理员
func (s *Service) InitializeSystem(ctx context.Context, req *InitializeSystemReq) error {
	// 检查是否已存在管理员
	count, err := s.userRepo.CountAdminUsers(ctx)
	if err != nil {
		return fmt.Errorf("检查管理员状态失败: %w", err)
	}
	if count > 0 {
		return errors.New("系统已初始化，禁止重复初始化")
	}

	// 检查用户名是否已存在
	if user, _ := s.userRepo.GetByUsername(ctx, req.AdminUsername); user != nil {
		return errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在
	if user, _ := s.userRepo.GetByEmail(ctx, req.AdminEmail); user != nil {
		return errors.New("邮箱已存在")
	}

	// 哈希密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码哈希失败: %w", err)
	}

	// 创建管理员用户
	admin := &model.User{
		Username:      req.AdminUsername,
		Email:         req.AdminEmail,
		PasswordHash:  string(hashedPassword),
		Nickname:      req.AdminUsername, // 默认使用用户名作为昵称
		Role:          "admin",
		Status:        "active",
		EmailVerified: true,
	}

	err = s.userRepo.Create(ctx, admin)
	if err != nil {
		return fmt.Errorf("创建管理员失败: %w", err)
	}

	return nil
}
