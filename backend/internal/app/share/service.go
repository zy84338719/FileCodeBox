package share

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/zy84338719/fileCodeBox/internal/pkg/utils"
	"github.com/zy84338719/fileCodeBox/internal/repo/db/dao"
	"github.com/zy84338719/fileCodeBox/internal/repo/db/model"
	"github.com/zy84338719/fileCodeBox/internal/storage"
)

type ShareTextReq struct {
	Text         string
	ExpiredAt    *time.Time
	ExpiredCount int
	RequireAuth  bool
	UserID       *uint
	UploadType   string
	OwnerIP      string
}

type ShareFileReq struct {
	FilePath     string
	Size         int64
	Text         string
	ExpiredAt    *time.Time
	ExpiredCount int
	RequireAuth  bool
	UserID       *uint
	UploadType   string
	OwnerIP      string
	FileHash     string
	IsChunked    bool
	UploadID     string
}

type ShareResp struct {
	Code         string     `json:"code"`
	Prefix       string     `json:"prefix"`
	Suffix       string     `json:"suffix"`
	UUIDFileName string     `json:"uuid_file_name"`
	FilePath     string     `json:"file_path"`
	Size         int64      `json:"size"`
	Text         string     `json:"text"`
	ExpiredAt    *time.Time `json:"expired_at"`
	ExpiredCount int        `json:"expired_count"`
	UsedCount    int        `json:"used_count"`
	FileHash     string     `json:"file_hash"`
	IsChunked    bool       `json:"is_chunked"`
	UploadID     string     `json:"upload_id"`
	UserID       *uint      `json:"user_id"`
	UploadType   string     `json:"upload_type"`
	RequireAuth  bool       `json:"require_auth"`
	OwnerIP      string     `json:"owner_ip"`
	ShareURL     string     `json:"share_url"`      // 相对分享链接
	FullShareURL string     `json:"full_share_url"` // 完整分享链接
}

type Service struct {
	fileCodeRepo *dao.FileCodeRepository
	userService  UserServiceInterface
	storage      storage.StorageInterface
	baseURL      string // 基础 URL，用于生成分享链接
}

// UserServiceInterface 定义用户服务接口，避免循环依赖
type UserServiceInterface interface {
	UpdateUserStats(userID uint, statsType string, value int64) error
}

func NewService(baseURL string, storageService storage.StorageInterface) *Service {
	// 延迟初始化 repository，确保数据库已经准备好
	return &Service{
		fileCodeRepo: nil, // 延迟初始化
		userService:  nil,
		storage:      storageService,
		baseURL:      baseURL,
	}
}

// ensureRepository 确保repository已初始化
func (s *Service) ensureRepository() {
	if s.fileCodeRepo == nil {
		s.fileCodeRepo = dao.NewFileCodeRepository()
	}
}

func (s *Service) SetUserService(userService UserServiceInterface) {
	s.userService = userService
}

// GenerateCode 生成分享代码
func (s *Service) GenerateCode() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))

	code := make([]byte, 8)
	for i := range code {
		code[i] = charset[seededRand.Intn(len(charset))]
	}

	return string(code)
}

// ShareText 分享文本
func (s *Service) ShareText(ctx context.Context, req *ShareTextReq) (*ShareResp, error) {
	s.ensureRepository()

	code := s.GenerateCode()

	fileCode := &model.FileCode{
		Code:         code,
		Text:         req.Text,
		ExpiredAt:    req.ExpiredAt,
		ExpiredCount: req.ExpiredCount,
		RequireAuth:  req.RequireAuth,
		UserID:       req.UserID,
		UploadType:   req.UploadType,
		OwnerIP:      req.OwnerIP,
	}

	if err := s.fileCodeRepo.Create(ctx, fileCode); err != nil {
		return nil, err
	}

	// 更新用户统计
	if s.userService != nil && req.UserID != nil {
		if err := s.userService.UpdateUserStats(*req.UserID, "uploads", 1); err != nil {
			// 记录错误但不影响主流程
		}
	}

	return s.modelToResp(fileCode), nil
}

// ShareTextWithAuth 带认证的文本分享（用于 Handler）
func (s *Service) ShareTextWithAuth(ctx context.Context, text string, expireValue int, expireStyle string, userID *uint, ownerIP string) (*ShareResp, error) {
	// 计算过期时间
	expireTime := utils.CalculateExpireTime(expireValue, expireStyle)
	expireCount := utils.CalculateExpireCount(expireStyle, expireValue)

	uploadType := "anonymous"
	if userID != nil {
		uploadType = "authenticated"
	}

	req := &ShareTextReq{
		Text:         text,
		ExpiredAt:    expireTime,
		ExpiredCount: expireCount,
		UserID:       userID,
		UploadType:   uploadType,
		OwnerIP:      ownerIP,
	}

	resp, err := s.ShareText(ctx, req)
	if err != nil {
		return nil, err
	}

	// 生成分享 URL
	resp.ShareURL = fmt.Sprintf("/share/%s", resp.Code)
	resp.FullShareURL = fmt.Sprintf("%s/share/%s", s.baseURL, resp.Code)

	return resp, nil
}

// ShareFile 分享文件
func (s *Service) ShareFile(ctx context.Context, req *ShareFileReq) (*ShareResp, error) {
	s.ensureRepository()

	code := s.GenerateCode()

	fileCode := &model.FileCode{
		Code:         code,
		FilePath:     req.FilePath,
		Size:         req.Size,
		Text:         req.Text,
		ExpiredAt:    req.ExpiredAt,
		ExpiredCount: req.ExpiredCount,
		RequireAuth:  req.RequireAuth,
		UserID:       req.UserID,
		UploadType:   req.UploadType,
		OwnerIP:      req.OwnerIP,
		FileHash:     req.FileHash,
		IsChunked:    req.IsChunked,
		UploadID:     req.UploadID,
	}

	if err := s.fileCodeRepo.Create(ctx, fileCode); err != nil {
		return nil, err
	}

	// 更新用户统计
	if s.userService != nil && req.UserID != nil {
		if err := s.userService.UpdateUserStats(*req.UserID, "uploads", 1); err != nil {
			// 记录错误但不影响主流程
		}
		if err := s.userService.UpdateUserStats(*req.UserID, "storage", req.Size); err != nil {
			// 记录错误但不影响主流程
		}
	}

	return &ShareResp{
		Code:         fileCode.Code,
		Prefix:       fileCode.Prefix,
		Suffix:       fileCode.Suffix,
		UUIDFileName: fileCode.UUIDFileName,
		FilePath:     fileCode.FilePath,
		Size:         fileCode.Size,
		Text:         fileCode.Text,
		ExpiredAt:    fileCode.ExpiredAt,
		ExpiredCount: fileCode.ExpiredCount,
		UsedCount:    fileCode.UsedCount,
		FileHash:     fileCode.FileHash,
		IsChunked:    fileCode.IsChunked,
		UploadID:     fileCode.UploadID,
		UserID:       fileCode.UserID,
		UploadType:   fileCode.UploadType,
		RequireAuth:  fileCode.RequireAuth,
		OwnerIP:      fileCode.OwnerIP,
	}, nil
}

// GetFileByCode 通过代码获取文件
func (s *Service) GetFileByCode(ctx context.Context, code string) (*model.FileCode, error) {
	s.ensureRepository()

	fileCode, err := s.fileCodeRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	// 检查文件是否过期
	if fileCode.IsExpired() {
		return nil, errors.New("file has expired")
	}

	return fileCode, nil
}

// GetFilesByUserID 获取用户的文件列表
func (s *Service) GetFilesByUserID(ctx context.Context, userID uint, page, pageSize int) ([]*model.FileCode, int64, error) {
	s.ensureRepository()
	return s.fileCodeRepo.GetFilesByUserIDWithPagination(ctx, userID, page, pageSize)
}

// DeleteFile 删除文件
func (s *Service) DeleteFile(ctx context.Context, fileID uint, userID *uint) error {
	s.ensureRepository()

	// 如果指定了用户ID，验证文件所有权
	if userID != nil {
		file, err := s.fileCodeRepo.GetByUserID(ctx, *userID, fileID)
		if err != nil {
			return err
		}

		// 更新用户统计（减少存储空间）
		if s.userService != nil && userID != nil {
			if err := s.userService.UpdateUserStats(*userID, "storage", -file.Size); err != nil {
				// 记录错误但不影响主流程
			}
		}
	}

	return s.fileCodeRepo.Delete(ctx, fileID)
}

// DeleteFileByCode 根据分享码删除文件
func (s *Service) DeleteFileByCode(ctx context.Context, code string, userID uint) error {
	s.ensureRepository()

	// 1. 根据 code 查询文件记录
	file, err := s.fileCodeRepo.GetByCode(ctx, code)
	if err != nil {
		return fmt.Errorf("分享不存在")
	}

	// 2. 验证文件所有权（userID匹配）
	if file.UserID == nil || *file.UserID != userID {
		return fmt.Errorf("无权限删除此分享")
	}

	// 3. 如果是文件分享，删除物理文件
	if file.FilePath != "" && file.UUIDFileName != "" {
		// 获取完整的文件路径
		filePath := file.GetFilePath()

		// 如果有存储服务，尝试删除物理文件
		if s.storage != nil {
			if err := s.storage.DeleteFile(ctx, filePath); err != nil {
				// 记录错误但不阻止数据库删除
				// 可以考虑添加日志记录
			}
		}
	}

	// 4. 删除数据库记录
	if err := s.fileCodeRepo.Delete(ctx, file.ID); err != nil {
		return fmt.Errorf("删除分享记录失败: %w", err)
	}

	// 5. 更新用户统计（减少存储空间）
	if s.userService != nil {
		if err := s.userService.UpdateUserStats(userID, "storage", -file.Size); err != nil {
			// 记录错误但不影响主流程
		}
	}

	return nil
}

// GetFileList 获取文件列表
func (s *Service) GetFileList(ctx context.Context, page, pageSize int, search string) ([]*model.FileCode, int64, error) {
	s.ensureRepository()
	return s.fileCodeRepo.List(ctx, page, pageSize, search)
}

// UpdateFileUsage 更新文件使用次数（下载次数）
func (s *Service) UpdateFileUsage(ctx context.Context, code string) error {
	s.ensureRepository()

	fileCode, err := s.fileCodeRepo.GetByCode(ctx, code)
	if err != nil {
		return err
	}

	// 检查剩余次数
	if fileCode.ExpiredCount > 0 {
		fileCode.ExpiredCount--
		if fileCode.ExpiredCount < 0 {
			fileCode.ExpiredCount = 0
		}
	}

	// 增加使用次数
	fileCode.UsedCount++

	return s.fileCodeRepo.Update(ctx, fileCode)
}

// GetFileWithUsage 获取文件并增加使用次数
func (s *Service) GetFileWithUsage(ctx context.Context, code string, password string) (*model.FileCode, error) {
	s.ensureRepository()

	fileCode, err := s.GetFileByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	// 检查是否需要密码
	if fileCode.RequireAuth && password == "" {
		return nil, errors.New("需要密码")
	}

	// TODO: 验证密码逻辑（当前仅检查非空）

	return fileCode, nil
}

// modelToResp 将模型转换为响应
func (s *Service) modelToResp(fileCode *model.FileCode) *ShareResp {
	return &ShareResp{
		Code:         fileCode.Code,
		Prefix:       fileCode.Prefix,
		Suffix:       fileCode.Suffix,
		UUIDFileName: fileCode.UUIDFileName,
		FilePath:     fileCode.FilePath,
		Size:         fileCode.Size,
		Text:         fileCode.Text,
		ExpiredAt:    fileCode.ExpiredAt,
		ExpiredCount: fileCode.ExpiredCount,
		UsedCount:    fileCode.UsedCount,
		FileHash:     fileCode.FileHash,
		IsChunked:    fileCode.IsChunked,
		UploadID:     fileCode.UploadID,
		UserID:       fileCode.UserID,
		UploadType:   fileCode.UploadType,
		RequireAuth:  fileCode.RequireAuth,
		OwnerIP:      fileCode.OwnerIP,
	}
}
