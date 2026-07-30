package model

import (
	"time"

	"gorm.io/gorm"
)

// Notify 通知公告
type Notify struct {
	gorm.Model
	Title        string     `gorm:"size:255;not null" json:"title"`
	Content      string     `gorm:"type:text;not null" json:"content"`
	Type         string     `gorm:"size:20;default:'system';index" json:"type"` // system/feature/maintenance/share_retrieved
	Level        string     `gorm:"size:20;default:'info';index" json:"level"`  // info/warning/error/success
	Status       int        `gorm:"default:1;index" json:"status"`              // 0=草稿 1=发布 2=下线
	StartAt      *time.Time `json:"start_at,omitempty"`                         // 生效时间
	EndAt        *time.Time `json:"end_at,omitempty"`                           // 失效时间
	AuthorID     uint       `gorm:"index" json:"author_id"`                     // 发布管理员
	TargetUserID *uint      `gorm:"index" json:"target_user_id,omitempty"`      // 0=广播, >0=定向给某用户
	ReadAt       *time.Time `gorm:"index" json:"read_at,omitempty"`             // 用户首次已读时间
}

// TableName 自定义表名
func (Notify) TableName() string {
	return "notifies"
}

// IsActive 当前是否活跃（在时间窗口内 + 状态发布）
func (n *Notify) IsActive() bool {
	if n.Status != 1 {
		return false
	}
	now := time.Now()
	if n.StartAt != nil && now.Before(*n.StartAt) {
		return false
	}
	if n.EndAt != nil && now.After(*n.EndAt) {
		return false
	}
	return true
}
