// Package storage 提供文件存储的抽象层和不同存储策略的实现
package storage

import (
	"mime/multipart"

	"github.com/zy84338719/filecodebox/internal/models"

	"github.com/gin-gonic/gin"
)

// StrategyBasedStorage 基于策略的存储适配器
type StrategyBasedStorage struct {
	operator *StorageOperator
}

// NewStrategyBasedStorage 创建基于策略的存储
func NewStrategyBasedStorage(strategy StorageStrategy, pathManager *PathManager) *StrategyBasedStorage {
	return &StrategyBasedStorage{
		operator: NewStorageOperator(strategy, pathManager),
	}
}

// SaveFile 实现 StorageInterface
func (sbs *StrategyBasedStorage) SaveFile(file *multipart.FileHeader, savePath string) error {
	return sbs.operator.SaveFile(file, savePath)
}

// SaveChunk 实现 StorageInterface
func (sbs *StrategyBasedStorage) SaveChunk(uploadID string, chunkIndex int, data []byte, chunkHash string) error {
	return sbs.operator.SaveChunk(uploadID, chunkIndex, data, chunkHash)
}

// MergeChunks 实现 StorageInterface
func (sbs *StrategyBasedStorage) MergeChunks(uploadID string, chunk *models.UploadChunk, savePath string) error {
	return sbs.operator.MergeChunks(uploadID, chunk, savePath)
}

// CleanChunks 实现 StorageInterface
func (sbs *StrategyBasedStorage) CleanChunks(uploadID string) error {
	return sbs.operator.CleanChunks(uploadID)
}

// GetFileResponse 实现 StorageInterface
func (sbs *StrategyBasedStorage) GetFileResponse(c *gin.Context, fileCode *models.FileCode) error {
	return sbs.operator.GetFileResponse(c, fileCode)
}

// GetFileURL 实现 StorageInterface
func (sbs *StrategyBasedStorage) GetFileURL(fileCode *models.FileCode) (string, error) {
	return sbs.operator.GetFileURL(fileCode)
}

// DeleteFile 实现 StorageInterface
func (sbs *StrategyBasedStorage) DeleteFile(fileCode *models.FileCode) error {
	return sbs.operator.DeleteFile(fileCode)
}

// TestConnection 测试连接
func (sbs *StrategyBasedStorage) TestConnection() error {
	return sbs.operator.strategy.TestConnection()
}
