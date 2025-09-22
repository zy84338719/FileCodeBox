package db

import "gorm.io/gorm"

// TransferLog 记录上传/下载操作日志
// 当系统要求登录上传/下载时，用于追踪用户行为
// Operation: upload 或 download
// DurationMs: 针对下载记录耗时，上传默认为0
// Username保留冗余信息，便于查询，即使用户被删除仍保留原始名字
type TransferLog struct {
	gorm.Model
	Operation  string `gorm:"size:20;index" json:"operation"`
	FileCodeID uint   `gorm:"index" json:"file_code_id"`
	FileCode   string `gorm:"size:255" json:"file_code"`
	FileName   string `gorm:"size:255" json:"file_name"`
	FileSize   int64  `json:"file_size"`
	UserID     *uint  `gorm:"index" json:"user_id"`
	Username   string `gorm:"size:100" json:"username"`
	IP         string `gorm:"size:45" json:"ip"`
	DurationMs int64  `json:"duration_ms"`
}

// TransferLogQuery 查询条件
type TransferLogQuery struct {
	Operation string `json:"operation"`
	UserID    *uint  `json:"user_id"`
	Search    string `json:"search"`
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
}
