package chunk

import (
	"errors"
	"fmt"

	"github.com/zy84338719/filecodebox/internal/models"
)

// InitiateUpload 初始化分片上传
func (s *Service) InitiateUpload(uploadID, fileName string, totalChunks int, fileSize int64) (*models.UploadChunk, error) {
	// 检查是否已存在相同的上传ID
	existing, err := s.repositoryManager.Chunk.GetByUploadID(uploadID)
	if err == nil && existing != nil {
		return nil, errors.New("upload ID already exists")
	}

	// 创建控制记录（chunk_index = -1）
	chunk := &models.UploadChunk{
		UploadID:    uploadID,
		ChunkIndex:  -1, // 控制记录标识
		TotalChunks: totalChunks,
		FileSize:    fileSize,
		FileName:    fileName,
		Status:      "pending",
	}

	err = s.repositoryManager.Chunk.Create(chunk)
	if err != nil {
		return nil, err
	}

	return chunk, nil
}

// UploadChunk 上传单个分片
func (s *Service) UploadChunk(uploadID string, chunkIndex int, chunkHash string, chunkSize int) (*models.UploadChunk, error) {
	// 检查上传ID是否存在
	controlChunk, err := s.repositoryManager.Chunk.GetByUploadID(uploadID)
	if err != nil {
		return nil, fmt.Errorf("upload ID not found: %v", err)
	}

	// 检查分片索引是否有效
	if chunkIndex < 0 || chunkIndex >= controlChunk.TotalChunks {
		return nil, errors.New("invalid chunk index")
	}

	// 检查分片是否已存在
	existingChunk, err := s.repositoryManager.Chunk.GetChunkByIndex(uploadID, chunkIndex)
	if err == nil && existingChunk.Completed {
		return existingChunk, nil // 分片已完成，直接返回
	}

	// 创建或更新分片记录
	chunk := &models.UploadChunk{
		UploadID:   uploadID,
		ChunkIndex: chunkIndex,
		ChunkHash:  chunkHash,
		ChunkSize:  chunkSize,
		Status:     "completed",
		Completed:  true,
	}

	if existingChunk != nil {
		// 更新现有记录
		err = s.repositoryManager.Chunk.UpdateChunkCompleted(uploadID, chunkIndex, chunkHash)
		if err != nil {
			return nil, err
		}
		chunk.ID = existingChunk.ID
	} else {
		// 创建新记录
		err = s.repositoryManager.Chunk.Create(chunk)
		if err != nil {
			return nil, err
		}
	}

	return chunk, nil
}

// CheckUploadProgress 检查上传进度
func (s *Service) CheckUploadProgress(uploadID string) (float64, error) {
	controlChunk, err := s.repositoryManager.Chunk.GetByUploadID(uploadID)
	if err != nil {
		return 0, err
	}

	completedChunks, err := s.repositoryManager.Chunk.CountCompletedChunks(uploadID)
	if err != nil {
		return 0, err
	}

	if controlChunk.TotalChunks == 0 {
		return 0, nil
	}

	return float64(completedChunks) / float64(controlChunk.TotalChunks) * 100, nil
}

// IsUploadComplete 检查上传是否完成
func (s *Service) IsUploadComplete(uploadID string) (bool, error) {
	progress, err := s.CheckUploadProgress(uploadID)
	if err != nil {
		return false, err
	}
	return progress >= 100.0, nil
}

// CompleteUpload 完成分片上传
func (s *Service) CompleteUpload(uploadID string) error {
	// 检查所有分片是否已完成
	isComplete, err := s.IsUploadComplete(uploadID)
	if err != nil {
		return err
	}

	if !isComplete {
		return errors.New("upload not complete")
	}

	// 获取完成的分片数量
	completedChunks, err := s.repositoryManager.Chunk.CountCompletedChunks(uploadID)
	if err != nil {
		return err
	}

	// 更新控制记录状态（先简化，只更新进度）
	return s.repositoryManager.Chunk.UpdateUploadProgress(uploadID, int(completedChunks))
}

// CancelUpload 取消分片上传
func (s *Service) CancelUpload(uploadID string) error {
	// 删除所有相关的分片记录
	return s.repositoryManager.Chunk.DeleteByUploadID(uploadID)
}

// GetUploadInfo 获取上传信息
func (s *Service) GetUploadInfo(uploadID string) (*models.UploadChunk, error) {
	return s.repositoryManager.Chunk.GetByUploadID(uploadID)
}

// ListChunks 列出分片
func (s *Service) ListChunks(uploadID string) ([]*models.UploadChunk, error) {
	// GetByUploadID只返回单个记录，需要获取所有分片
	// 暂时返回空数组，需要实现专门的方法
	return []*models.UploadChunk{}, nil
}

// CleanupExpiredUploads 清理过期的上传
func (s *Service) CleanupExpiredUploads() (int, error) {
	// 清理超过24小时未完成的上传
	expiredUploads, err := s.repositoryManager.Chunk.GetIncompleteUploads(24)
	if err != nil {
		return 0, err
	}

	deletedCount := 0
	for _, upload := range expiredUploads {
		err := s.repositoryManager.Chunk.DeleteByUploadID(upload.UploadID)
		if err != nil {
			fmt.Printf("Warning: Failed to delete expired upload %s: %v\n", upload.UploadID, err)
		} else {
			deletedCount++
		}
	}

	return deletedCount, nil
}

// InitChunkUpload 初始化分片上传 (兼容性方法)
func (s *Service) InitChunkUpload(uploadID, fileName string, totalChunks int, fileSize int64) (*models.UploadChunk, error) {
	return s.InitiateUpload(uploadID, fileName, totalChunks, fileSize)
}

// GetUploadStatus 获取上传状态 (兼容性方法)
func (s *Service) GetUploadStatus(uploadID string) (*models.ChunkUploadStatusData, error) {
	info, err := s.GetUploadInfo(uploadID)
	if err != nil {
		return nil, err
	}

	progress, err := s.CheckUploadProgress(uploadID)
	if err != nil {
		return nil, err
	}

	completedChunks, err := s.repositoryManager.Chunk.CountCompletedChunks(uploadID)
	if err != nil {
		return nil, err
	}

	return &models.ChunkUploadStatusData{
		UploadID:      info.UploadID,
		Status:        info.Status,
		Progress:      progress,
		TotalChunks:   info.TotalChunks,
		UploadedCount: int(completedChunks),
		FileName:      info.FileName,
		FileSize:      info.FileSize,
		CreatedAt:     info.CreatedAt,
		UpdatedAt:     info.UpdatedAt,
	}, nil
}

// VerifyChunk 验证分片 (占位符实现)
func (s *Service) VerifyChunk(uploadID string, chunkIndex int, expectedHash string) (bool, error) {
	chunk, err := s.repositoryManager.Chunk.GetChunkByIndex(uploadID, chunkIndex)
	if err != nil {
		return false, err
	}

	return chunk.ChunkHash == expectedHash, nil
}
