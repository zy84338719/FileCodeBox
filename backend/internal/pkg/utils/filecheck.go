package utils

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/zy84338719/fileCodeBox/backend/internal/conf"
)

// ErrFileTooLarge 文件超过允许大小
var ErrFileTooLarge = errors.New("file size exceeds limit")

// GetMaxUploadSize 从全局配置获取上传大小上限。
// 配置未初始化（测试/启动早期）返回 0 表示不限，避免 nil 解引用。
func GetMaxUploadSize() int64 {
	cfg := conf.GetGlobalConfig()
	if cfg == nil {
		return 0
	}
	return cfg.Upload.UploadSize
}

// defaultBlockedExtensions 默认拒绝的可执行文件扩展名
var defaultBlockedExtensions = []string{
	".exe", ".bat", ".cmd", ".com", ".scr", ".msi",
	".sh", ".bash", ".ps1", ".vbs", ".js", ".jar",
	".dll", ".so", ".dylib", ".app",
}

// DefaultBlockedExtensions 返回默认黑名单扩展名（返回副本，调用方可追加）
func DefaultBlockedExtensions() []string {
	cp := make([]string, len(defaultBlockedExtensions))
	copy(cp, defaultBlockedExtensions)
	return cp
}

// IsBlockedExtension 判断文件扩展名是否在黑名单（大小写不敏感）。
// 无扩展名返回 false。
func IsBlockedExtension(filename string, blacklist []string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return false
	}
	for _, b := range blacklist {
		if strings.ToLower(b) == ext {
			return true
		}
	}
	return false
}

// CheckUploadSize 校验文件大小是否超限。
// maxSize <= 0 表示不限制。
func CheckUploadSize(fileSize, maxSize int64) error {
	if maxSize > 0 && fileSize > maxSize {
		return ErrFileTooLarge
	}
	return nil
}
