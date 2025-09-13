package repository

import (
	"time"

	"github.com/zy84338719/filecodebox/internal/models/db"
	"gorm.io/gorm"
)

// ChunkDAO 分片上传数据访问对象
type ChunkDAO struct {
	db *gorm.DB
}

// NewChunkDAO 创建新的分片DAO
func NewChunkDAO(db *gorm.DB) *ChunkDAO {
	return &ChunkDAO{db: db}
}

// Create 创建分片记录
func (dao *ChunkDAO) Create(chunk *db.UploadChunk) error {
	return dao.db.Create(chunk).Error
}

// GetByHash 根据哈希和文件大小获取分片信息
func (dao *ChunkDAO) GetByHash(chunkHash string, fileSize int64) (*db.UploadChunk, error) {
	var chunk db.UploadChunk
	err := dao.db.Where("chunk_hash = ? AND file_size = ? AND chunk_index = -1", chunkHash, fileSize).First(&chunk).Error
	if err != nil {
		return nil, err
	}
	return &chunk, nil
}

// GetByUploadID 根据上传ID获取控制记录
func (dao *ChunkDAO) GetByUploadID(uploadID string) (*db.UploadChunk, error) {
	var chunk db.UploadChunk
	err := dao.db.Where("upload_id = ? AND chunk_index = -1", uploadID).First(&chunk).Error
	if err != nil {
		return nil, err
	}
	return &chunk, nil
}

// GetChunkByIndex 根据上传ID和分片索引获取分片
func (dao *ChunkDAO) GetChunkByIndex(uploadID string, chunkIndex int) (*db.UploadChunk, error) {
	var chunk db.UploadChunk
	err := dao.db.Where("upload_id = ? AND chunk_index = ? AND completed = true", uploadID, chunkIndex).First(&chunk).Error
	if err != nil {
		return nil, err
	}
	return &chunk, nil
}

// UpdateUploadProgress 更新上传进度
func (dao *ChunkDAO) UpdateUploadProgress(uploadID string, completedChunks int) error {
	return dao.db.Model(&db.UploadChunk{}).
		Where("upload_id = ? AND chunk_index = -1", uploadID).
		Updates(map[string]interface{}{
			"completed":   completedChunks,
			"retry_count": gorm.Expr("retry_count + 1"),
		}).Error
}

// UpdateChunkCompleted 更新分片完成状态
func (dao *ChunkDAO) UpdateChunkCompleted(uploadID string, chunkIndex int, chunkHash string) error {
	return dao.db.Where("upload_id = ? AND chunk_index = ?", uploadID, chunkIndex).
		Updates(map[string]interface{}{
			"completed":  true,
			"chunk_hash": chunkHash,
			"status":     "completed",
		}).Error
}

// CountCompletedChunks 统计已完成的分片数量
func (dao *ChunkDAO) CountCompletedChunks(uploadID string) (int64, error) {
	var count int64
	err := dao.db.Model(&db.UploadChunk{}).
		Where("upload_id = ? AND chunk_index >= 0 AND completed = true", uploadID).
		Count(&count).Error
	return count, err
}

// DeleteByUploadID 删除上传相关的所有分片记录
func (dao *ChunkDAO) DeleteByUploadID(uploadID string) error {
	return dao.db.Where("upload_id = ?", uploadID).Delete(&db.UploadChunk{}).Error
}

// GetUploadList 获取上传列表（用于管理和清理）
func (dao *ChunkDAO) GetUploadList(page, pageSize int) ([]db.UploadChunk, int64, error) {
	var chunks []db.UploadChunk
	var total int64

	// 只获取控制记录（chunk_index = -1）
	query := dao.db.Model(&db.UploadChunk{}).Where("chunk_index = -1")

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&chunks).Error

	return chunks, total, err
}

// GetIncompleteUploads 获取未完成的上传（用于清理）
func (dao *ChunkDAO) GetIncompleteUploads(olderThan int) ([]db.UploadChunk, error) {
	var chunks []db.UploadChunk
	err := dao.db.Where("chunk_index = -1 AND status != 'completed' AND created_at < datetime('now', '-' || ? || ' hours')", olderThan).
		Find(&chunks).Error
	return chunks, err
}

// GetOldChunks 获取超过指定时间的未完成分片上传
func (dao *ChunkDAO) GetOldChunks(cutoffTime time.Time) ([]db.UploadChunk, error) {
	var oldChunks []db.UploadChunk
	err := dao.db.Where("created_at < ? AND chunk_index = -1", cutoffTime).Find(&oldChunks).Error
	return oldChunks, err
}

// DeleteChunksByUploadIDs 批量删除指定上传ID的所有分片记录
func (dao *ChunkDAO) DeleteChunksByUploadIDs(uploadIDs []string) (int, error) {
	if len(uploadIDs) == 0 {
		return 0, nil
	}

	count := 0
	for _, uploadID := range uploadIDs {
		if err := dao.db.Where("upload_id = ?", uploadID).Delete(&db.UploadChunk{}).Error; err != nil {
			continue // 记录错误但继续处理其他上传
		}
		count++
	}
	return count, nil
}

// GetUploadedChunkIndexes 获取已上传分片的索引列表
func (dao *ChunkDAO) GetUploadedChunkIndexes(uploadID string) ([]int, error) {
	var uploadedChunks []int
	err := dao.db.Model(&db.UploadChunk{}).
		Where("upload_id = ? AND completed = true AND chunk_index >= 0", uploadID).
		Pluck("chunk_index", &uploadedChunks).Error
	return uploadedChunks, err
}

// FirstOrCreateChunk 创建或更新分片记录
func (dao *ChunkDAO) FirstOrCreateChunk(chunk *db.UploadChunk) error {
	return dao.db.Where("upload_id = ? AND chunk_index = ?", chunk.UploadID, chunk.ChunkIndex).
		Assign(chunk).
		FirstOrCreate(chunk).Error
}
