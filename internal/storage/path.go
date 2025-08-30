package storage

import (
	"fmt"
	"path/filepath"
	"time"
)

// PathManager 路径管理器，用于统一处理各种存储后端的路径
type PathManager struct {
	basePath string
}

// NewPathManager 创建路径管理器
func NewPathManager(basePath string) *PathManager {
	return &PathManager{
		basePath: basePath,
	}
}

// GetDateBasedPath 获取基于日期的文件路径
func (pm *PathManager) GetDateBasedPath(filename string) string {
	now := time.Now()
	dateDir := filepath.Join(now.Format("2006"), now.Format("01"), now.Format("02"))
	return filepath.Join(pm.basePath, dateDir, filename)
}

// GetChunkBasePath 获取分片存储的基础路径
func (pm *PathManager) GetChunkBasePath() string {
	return filepath.Join(pm.basePath, "chunks")
}

// GetChunkDir 获取特定上传ID的分片目录
func (pm *PathManager) GetChunkDir(uploadID string) string {
	return filepath.Join(pm.GetChunkBasePath(), uploadID)
}

// GetChunkPath 获取特定分片的完整路径
func (pm *PathManager) GetChunkPath(uploadID string, chunkIndex int) string {
	return filepath.Join(pm.GetChunkDir(uploadID), fmt.Sprintf("chunk_%d", chunkIndex))
}

// CleanPath 清理路径，确保路径格式正确
func (pm *PathManager) CleanPath(path string) string {
	return filepath.Clean(path)
}

// GetFullPath 获取基于基础路径的完整路径
func (pm *PathManager) GetFullPath(relativePath string) string {
	return filepath.Join(pm.basePath, relativePath)
}
