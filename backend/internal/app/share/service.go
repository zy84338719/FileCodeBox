package share

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/zy84338719/fileCodeBox/backend/internal/pkg/utils"
	"github.com/zy84338719/fileCodeBox/backend/internal/repo/db/dao"
	"github.com/zy84338719/fileCodeBox/backend/internal/repo/db/model"
	"github.com/zy84338719/fileCodeBox/backend/internal/storage"
	"gorm.io/gorm"
)

type ShareTextReq struct {
	Text         string
	ExpiredAt    *time.Time
	ExpiredCount int
	RequireAuth  bool
	PasswordHash string
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
	PasswordHash string
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
	notifySvc    NotifyServiceInterface
}

// NotifyServiceInterface 取件通知接口（避免 share → notify 直接依赖）
type NotifyServiceInterface interface {
	CreateForUserSimple(ctx context.Context, userID uint, title, content, notifyType, level string) error
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
		notifySvc:    nil,
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

// SetNotifyService 注入 notify service（取件时给 owner 发通知）
func (s *Service) SetNotifyService(svc NotifyServiceInterface) {
	s.notifySvc = svc
}

// GenerateCode 生成分享代码（crypto/rand，8 位字母数字）。
func (s *Service) GenerateCode() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 8
	max := big.NewInt(int64(len(charset)))
	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			// 极端回退（crypto/rand 几乎不会失败）
			n = big.NewInt(int64(time.Now().UnixNano()) % int64(len(charset)))
		}
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

// createWithRetry 通用写库重试（code 唯一冲突时换码重试，最多 5 次）。
func (s *Service) createWithRetry(ctx context.Context, build func(code string) *model.FileCode) (*model.FileCode, error) {
	for i := 0; i < 5; i++ {
		fc := build(s.GenerateCode())
		if err := s.fileCodeRepo.Create(ctx, fc); err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				continue // 换 code 重试
			}
			return nil, err
		}
		return fc, nil
	}
	return nil, errors.New("生成分享码失败：多次冲突")
}

// ShareText 分享文本
func (s *Service) ShareText(ctx context.Context, req *ShareTextReq) (*ShareResp, error) {
	s.ensureRepository()

	fileCode, err := s.createWithRetry(ctx, func(code string) *model.FileCode {
		return &model.FileCode{
			Code:         code,
			Text:         req.Text,
			ExpiredAt:    req.ExpiredAt,
			ExpiredCount: req.ExpiredCount,
			RequireAuth:  req.RequireAuth,
			PasswordHash: req.PasswordHash,
			UserID:       req.UserID,
			UploadType:   req.UploadType,
			OwnerIP:      req.OwnerIP,
		}
	})
	if err != nil {
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
	return s.CreateShare(ctx, req)
}

// CreateShare 创建分享记录（ShareFile 的语义化别名，便于其他 service 调用）
// 行为：生成 code → 写 file_codes 表 → 返回 share_code / url
func (s *Service) CreateShare(ctx context.Context, req *ShareFileReq) (*ShareResp, error) {
	s.ensureRepository()

	fileCode, err := s.createWithRetry(ctx, func(code string) *model.FileCode {
		return &model.FileCode{
			Code:         code,
			FilePath:     req.FilePath,
			Size:         req.Size,
			Text:         req.Text,
			ExpiredAt:    req.ExpiredAt,
			ExpiredCount: req.ExpiredCount,
			RequireAuth:  req.RequireAuth,
			PasswordHash: req.PasswordHash,
			UserID:       req.UserID,
			UploadType:   req.UploadType,
			OwnerIP:      req.OwnerIP,
			FileHash:     req.FileHash,
			IsChunked:    req.IsChunked,
			UploadID:     req.UploadID,
		}
	})
	if err != nil {
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

	return s.modelToResp(fileCode), nil
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

	// 记录取件人 + 通知 owner（fire-and-forget，失败不影响主流程）
	viewerIP := "" // caller 已知 caller IP；这里仅占位，详细 IP 由 handler 注入
	_ = s.RecordViewerAndNotify(ctx, code, viewerIP, "")

	return fileCode, nil
}

// RecordViewerAndNotify 记录取件人 + 给 owner 发通知
//   - viewerIP: 取件人 IP
//   - notifyType/level: 通知 type / level
//
// 若 file_codes 找不到（anonymous 路径 share_code 是 file_name 占位），静默跳过
func (s *Service) RecordViewerAndNotify(ctx context.Context, code, viewerIP, viewerDetail string) error {
	s.ensureRepository()
	if err := s.fileCodeRepo.UpdateViewer(ctx, code, viewerIP); err != nil {
		// 文件可能不存在，忽略
		return nil
	}
	// 读最新记录判断 owner
	fc, err := s.fileCodeRepo.GetByCode(ctx, code)
	if err != nil || fc == nil {
		return nil
	}
	if fc.UserID == nil || s.notifySvc == nil {
		return nil
	}
	title := "您的分享已被取件"
	if fc.Text != "" {
		title = "您的文本分享已被查看"
	}
	content := fmt.Sprintf("分享码: %s\n取件人 IP: %s\n时间: %s", code, viewerIP, time.Now().Format("2006-01-02 15:04:05"))
	if viewerDetail != "" {
		content += "\n" + viewerDetail
	}
	return s.notifySvc.CreateForUserSimple(ctx, *fc.UserID, title, content, "share_retrieved", "info")
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

// UserShareListItem 用户分享列表项（包含 viewer 追踪字段）
type UserShareListItem struct {
	ID           uint       `json:"id"`
	Code         string     `json:"code"`
	Prefix       string     `json:"prefix"`
	Suffix       string     `json:"suffix"`
	FileName     string     `json:"file_name"`
	FilePath     string     `json:"file_path"`
	Size         int64      `json:"size"`
	Text         string     `json:"text"`
	ExpiredAt    *time.Time `json:"expired_at"`
	ExpiredCount int        `json:"expired_count"`
	UsedCount    int        `json:"used_count"`
	RequireAuth  bool       `json:"require_auth"`
	UploadType   string     `json:"upload_type"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
	ViewerIP     string     `json:"viewer_ip"`
	ViewerAt     *time.Time `json:"viewer_at"`
	ViewerCount  int        `json:"viewer_count"`
	IsExpired    bool       `json:"is_expired"`
	IsTextShare  bool       `json:"is_text_share"`
}

// deletedAtToPtr gorm.DeletedAt → *time.Time（nil 表示未删除）
func deletedAtToPtr(d gorm.DeletedAt) *time.Time {
	if d.Valid {
		return &d.Time
	}
	return nil
}

// ListUserShares 获取用户的分享列表（带筛选）
func (s *Service) ListUserShares(ctx context.Context, userID uint, filter dao.UserShareFilter) ([]*UserShareListItem, int64, error) {
	s.ensureRepository()
	files, total, err := s.fileCodeRepo.GetUserSharesWithFilter(ctx, userID, filter)
	if err != nil {
		return nil, 0, err
	}
	items := make([]*UserShareListItem, 0, len(files))
	for _, f := range files {
		items = append(items, toUserShareListItem(f))
	}
	return items, total, nil
}

// BatchDeleteUserShares 批量软删除
func (s *Service) BatchDeleteUserShares(ctx context.Context, userID uint, codes []string) (int, error) {
	s.ensureRepository()
	return s.fileCodeRepo.BatchSoftDeleteByCodes(ctx, userID, codes)
}

// BatchExtendUserShares 批量延期
func (s *Service) BatchExtendUserShares(ctx context.Context, userID uint, codes []string, newExpireAt *time.Time) (int, error) {
	s.ensureRepository()
	return s.fileCodeRepo.BatchExtendByCodes(ctx, userID, codes, newExpireAt)
}

// RestoreUserShare 恢复软删除的分享
func (s *Service) RestoreUserShare(ctx context.Context, userID uint, code string) error {
	s.ensureRepository()
	return s.fileCodeRepo.RestoreByCode(ctx, userID, code)
}

// HardDeleteUserShare 永久删除软删除的分享
func (s *Service) HardDeleteUserShare(ctx context.Context, userID uint, code string) error {
	s.ensureRepository()
	return s.fileCodeRepo.HardDeleteByCode(ctx, userID, code)
}

// RecordViewer 记录取件人 IP / 时间（用于 owner 查看取件历史）
// 取件流程中调用。返回 fileCode 供后续通知使用
func (s *Service) RecordViewer(ctx context.Context, code, viewerIP string) (*model.FileCode, error) {
	s.ensureRepository()
	if err := s.fileCodeRepo.UpdateViewer(ctx, code, viewerIP); err != nil {
		return nil, err
	}
	return s.fileCodeRepo.GetByCode(ctx, code)
}

// toUserShareListItem model → 列表项
func toUserShareListItem(f *model.FileCode) *UserShareListItem {
	item := &UserShareListItem{
		ID:           f.ID,
		Code:         f.Code,
		Prefix:       f.Prefix,
		Suffix:       f.Suffix,
		FilePath:     f.FilePath,
		Size:         f.Size,
		Text:         f.Text,
		ExpiredAt:    f.ExpiredAt,
		ExpiredCount: f.ExpiredCount,
		UsedCount:    f.UsedCount,
		RequireAuth:  f.RequireAuth,
		UploadType:   f.UploadType,
		CreatedAt:    f.CreatedAt,
		UpdatedAt:    f.UpdatedAt,
		DeletedAt:    deletedAtToPtr(f.DeletedAt),
		ViewerIP:     f.ViewerIP,
		ViewerAt:     f.ViewerAt,
		ViewerCount:  f.ViewerCount,
		IsExpired:    f.IsExpired(),
		IsTextShare:  f.Text != "",
	}
	// 提取文件名
	fileName := f.UUIDFileName
	if fileName == "" && f.FilePath != "" {
		parts := strings.Split(f.FilePath, "/")
		if len(parts) > 0 {
			fileName = parts[len(parts)-1]
		}
	}
	item.FileName = fileName
	return item
}
