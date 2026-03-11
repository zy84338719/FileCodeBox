package model

import "time"

// FilePreview 文件预览信息（扩展FileCode模型）
type FilePreview struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	FileCodeID  uint      `gorm:"uniqueIndex;not null" json:"file_code_id"` // 关联的文件ID
	PreviewType string    `gorm:"size:20;not null" json:"preview_type"`     // 预览类型
	Thumbnail   string    `gorm:"size:255" json:"thumbnail"`                // 缩略图路径
	PreviewURL  string    `gorm:"size:255" json:"preview_url"`              // 预览URL
	Width       int       `json:"width"`                                    // 宽度
	Height      int       `json:"height"`                                   // 高度
	Duration    int       `json:"duration"`                                 // 时长（秒）
	PageCount   int       `json:"page_count"`                               // 页数
	TextContent string    `gorm:"type:text" json:"text_content"`            // 文本内容
	MimeType    string    `gorm:"size:50" json:"mime_type"`                 // MIME类型
	FileSize    int64     `json:"file_size"`                                // 文件大小
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 指定表名
func (FilePreview) TableName() string {
	return "file_previews"
}
