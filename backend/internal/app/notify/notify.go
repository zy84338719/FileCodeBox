// Package notify 实现系统通知（管理员公告）service。
package notify

import (
	"context"
	"errors"
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/zy84338719/fileCodeBox/backend/internal/repo/db/model"
)

// 错误
var (
	ErrNotifyNotFound = errors.New("notify not found")
	ErrInvalidParam   = errors.New("invalid parameter")
)

// Service 通知 service
type Service struct {
	db *gorm.DB
}

// NewService 创建 service
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// ListItem 列表项
type ListItem struct {
	ID        uint       `json:"id"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	Type      string     `json:"type"`
	Level     string     `json:"level"`
	Status    int        `json:"status"`
	StartAt   *time.Time `json:"start_at,omitempty"`
	EndAt     *time.Time `json:"end_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// ListData 列表数据
type ListData struct {
	Items    []ListItem `json:"items"`
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
}

// ListReq 列表请求
type ListReq struct {
	Page     int
	PageSize int
	Type     string
	Level    string
	Status   *int
}

// List 列表
func (s *Service) List(ctx context.Context, req ListReq) (*ListData, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 20
	}
	tx := s.db.WithContext(ctx).Model(&model.Notify{})
	if req.Type != "" {
		tx = tx.Where("type = ?", req.Type)
	}
	if req.Level != "" {
		tx = tx.Where("level = ?", req.Level)
	}
	if req.Status != nil {
		tx = tx.Where("status = ?", *req.Status)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, err
	}
	var items []model.Notify
	if err := tx.Order("created_at DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&items).Error; err != nil {
		return nil, err
	}
	result := &ListData{
		Items:    make([]ListItem, 0, len(items)),
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}
	for _, n := range items {
		result.Items = append(result.Items, toListItem(&n))
	}
	return result, nil
}

// Get 单条
func (s *Service) Get(ctx context.Context, id uint) (*ListItem, error) {
	var n model.Notify
	if err := s.db.WithContext(ctx).First(&n, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotifyNotFound
		}
		return nil, err
	}
	item := toListItem(&n)
	return &item, nil
}

// Active 当前活跃通知（公开 API）
func (s *Service) Active(ctx context.Context, typ string) ([]ListItem, error) {
	tx := s.db.WithContext(ctx).Model(&model.Notify{}).
		Where("status = ?", 1).
		Where("start_at IS NULL OR start_at <= ?", time.Now()).
		Where("end_at IS NULL OR end_at >= ?", time.Now())
	if typ != "" {
		tx = tx.Where("type = ?", typ)
	}
	var items []model.Notify
	if err := tx.Order("created_at DESC").Limit(50).Find(&items).Error; err != nil {
		return nil, err
	}
	result := make([]ListItem, 0, len(items))
	for i := range items {
		result = append(result, toListItem(&items[i]))
	}
	return result, nil
}

// CreateReq 创建请求
type CreateReq struct {
	Title    string
	Content  string
	Type     string
	Level    string
	Status   int
	StartAt  *time.Time
	EndAt    *time.Time
	AuthorID uint
}

// Create 创建
func (s *Service) Create(ctx context.Context, req CreateReq) (*ListItem, error) {
	if req.Title == "" || req.Content == "" {
		return nil, ErrInvalidParam
	}
	if req.Type == "" {
		req.Type = "system"
	}
	if req.Level == "" {
		req.Level = "info"
	}
	if req.Status == 0 {
		req.Status = 1
	}
	n := model.Notify{
		Title:    req.Title,
		Content:  req.Content,
		Type:     req.Type,
		Level:    req.Level,
		Status:   req.Status,
		StartAt:  req.StartAt,
		EndAt:    req.EndAt,
		AuthorID: req.AuthorID,
	}
	if err := s.db.WithContext(ctx).Create(&n).Error; err != nil {
		return nil, err
	}
	item := toListItem(&n)
	return &item, nil
}

// UpdateReq 更新请求
type UpdateReq struct {
	ID      uint
	Title   *string
	Content *string
	Type    *string
	Level   *string
	Status  *int
	StartAt *time.Time
	EndAt   *time.Time
}

// Update 更新
func (s *Service) Update(ctx context.Context, req UpdateReq) error {
	updates := map[string]interface{}{}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Content != nil {
		updates["content"] = *req.Content
	}
	if req.Type != nil {
		updates["type"] = *req.Type
	}
	if req.Level != nil {
		updates["level"] = *req.Level
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.StartAt != nil {
		updates["start_at"] = *req.StartAt
	}
	if req.EndAt != nil {
		updates["end_at"] = *req.EndAt
	}
	if len(updates) == 0 {
		return nil
	}
	res := s.db.WithContext(ctx).Model(&model.Notify{}).Where("id = ?", req.ID).Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotifyNotFound
	}
	return nil
}

// Delete 删除
func (s *Service) Delete(ctx context.Context, id uint) error {
	res := s.db.WithContext(ctx).Delete(&model.Notify{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotifyNotFound
	}
	return nil
}

// ============ 辅助 ============

func toListItem(n *model.Notify) ListItem {
	return ListItem{
		ID:        n.ID,
		Title:     n.Title,
		Content:   n.Content,
		Type:      n.Type,
		Level:     n.Level,
		Status:    n.Status,
		StartAt:   n.StartAt,
		EndAt:     n.EndAt,
		CreatedAt: n.CreatedAt,
		UpdatedAt: n.UpdatedAt,
	}
}

// ParseID 解析 path 中的 id
func ParseID(s string) (uint, error) {
	id, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, ErrInvalidParam
	}
	return uint(id), nil
}
