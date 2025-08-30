package storage

import (
	"fmt"
	"github.com/zy84338719/filecodebox/internal/models"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
)

// StorageStrategy 存储策略接口 - 定义每个存储后端的差异化操作
type StorageStrategy interface {
	// 基础文件操作
	WriteFile(path string, data []byte) error
	ReadFile(path string) ([]byte, error)
	DeleteFile(path string) error
	FileExists(path string) bool

	// 上传文件操作
	SaveUploadFile(file *multipart.FileHeader, savePath string) error

	// 下载操作
	ServeFile(c *gin.Context, filePath string, fileName string) error
	GenerateFileURL(filePath string, fileName string) (string, error)

	// 连接测试
	TestConnection() error
}

// StorageOperator 通用存储操作器 - 包含公共逻辑
type StorageOperator struct {
	strategy    StorageStrategy
	pathManager *PathManager
}

// NewStorageOperator 创建存储操作器
func NewStorageOperator(strategy StorageStrategy, pathManager *PathManager) *StorageOperator {
	return &StorageOperator{
		strategy:    strategy,
		pathManager: pathManager,
	}
}

// SaveFile 保存文件 - 公共逻辑
func (so *StorageOperator) SaveFile(file *multipart.FileHeader, savePath string) error {
	return so.strategy.SaveUploadFile(file, savePath)
}

// SaveChunk 保存分片 - 公共逻辑
func (so *StorageOperator) SaveChunk(uploadID string, chunkIndex int, data []byte, chunkHash string) error {
	chunkPath := so.pathManager.GetChunkPath(uploadID, chunkIndex)
	return so.strategy.WriteFile(chunkPath, data)
}

// MergeChunks 合并分片 - 公共逻辑
func (so *StorageOperator) MergeChunks(uploadID string, chunk *models.UploadChunk, savePath string) error {
	var mergedData []byte

	// 读取并合并所有分片
	for i := 0; i < chunk.TotalChunks; i++ {
		chunkPath := so.pathManager.GetChunkPath(uploadID, i)
		chunkData, err := so.strategy.ReadFile(chunkPath)
		if err != nil {
			return fmt.Errorf("读取分片 %d 失败: %v", i, err)
		}
		mergedData = append(mergedData, chunkData...)
	}

	// 写入合并后的文件
	return so.strategy.WriteFile(savePath, mergedData)
}

// CleanChunks 清理分片 - 公共逻辑
func (so *StorageOperator) CleanChunks(uploadID string) error {
	chunkDir := so.pathManager.GetChunkDir(uploadID)
	return so.strategy.DeleteFile(chunkDir)
}

// GetFileResponse 获取文件响应 - 公共逻辑
func (so *StorageOperator) GetFileResponse(c *gin.Context, fileCode *models.FileCode) error {
	// 处理文本分享
	if fileCode.Text != "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "success",
			"detail": gin.H{
				"code": fileCode.Code,
				"name": fileCode.Prefix + fileCode.Suffix,
				"size": fileCode.Size,
				"text": fileCode.Text,
			},
		})
		return nil
	}

	filePath := fileCode.GetFilePath()
	if filePath == "" {
		return fmt.Errorf("文件路径为空")
	}

	// 构建完整的文件路径
	fullPath := so.pathManager.GetFullPath(filePath)

	// 检查文件是否存在
	if !so.strategy.FileExists(fullPath) {
		return fmt.Errorf("文件不存在")
	}

	// 委托给具体策略处理文件下载
	fileName := fileCode.Prefix + fileCode.Suffix
	return so.strategy.ServeFile(c, fullPath, fileName)
}

// GetFileURL 获取文件URL - 公共逻辑
func (so *StorageOperator) GetFileURL(fileCode *models.FileCode) (string, error) {
	if fileCode.Text != "" {
		return fileCode.Text, nil
	}

	filePath := fileCode.GetFilePath()
	fileName := fileCode.Prefix + fileCode.Suffix

	// 对于本地存储，传递相对路径即可；对于其他存储，可能需要完整路径
	return so.strategy.GenerateFileURL(filePath, fileName)
}

// DeleteFile 删除文件 - 公共逻辑
func (so *StorageOperator) DeleteFile(fileCode *models.FileCode) error {
	if fileCode.Text != "" {
		return nil // 文本不需要删除文件
	}

	filePath := fileCode.GetFilePath()
	if filePath == "" {
		return nil
	}

	// 构建完整的文件路径
	fullPath := so.pathManager.GetFullPath(filePath)
	return so.strategy.DeleteFile(fullPath)
}
