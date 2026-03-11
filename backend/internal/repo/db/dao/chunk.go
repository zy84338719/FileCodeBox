package dao

import (
	"context"
	"time"

	"github.com/zy84338719/fileCodeBox/internal/repo/db"
	"github.com/zy84338719/fileCodeBox/internal/repo/db/model"
	"gorm.io/gorm"
)

type ChunkRepository struct {
}

func NewChunkRepository() *ChunkRepository {
	return &ChunkRepository{}
}

func (r *ChunkRepository) db() *gorm.DB {
	return db.GetDB()
}

func (r *ChunkRepository) Create(ctx context.Context, chunk *model.UploadChunk) error {
	return r.db().WithContext(ctx).Create(chunk).Error
}

func (r *ChunkRepository) GetByHash(ctx context.Context, chunkHash string, fileSize int64) (*model.UploadChunk, error) {
	var chunk model.UploadChunk
	err := r.db().WithContext(ctx).Where("chunk_hash = ? AND file_size = ? AND chunk_index = -1", chunkHash, fileSize).First(&chunk).Error
	if err != nil {
		return nil, err
	}
	return &chunk, nil
}

func (r *ChunkRepository) GetByUploadID(ctx context.Context, uploadID string) (*model.UploadChunk, error) {
	var chunk model.UploadChunk
	err := r.db().WithContext(ctx).Where("upload_id = ? AND chunk_index = -1", uploadID).First(&chunk).Error
	if err != nil {
		return nil, err
	}
	return &chunk, nil
}

func (r *ChunkRepository) GetChunkByIndex(ctx context.Context, uploadID string, chunkIndex int) (*model.UploadChunk, error) {
	var chunk model.UploadChunk
	err := r.db().WithContext(ctx).Where("upload_id = ? AND chunk_index = ? AND completed = true", uploadID, chunkIndex).First(&chunk).Error
	if err != nil {
		return nil, err
	}
	return &chunk, nil
}

func (r *ChunkRepository) UpdateUploadProgress(ctx context.Context, uploadID string, completedChunks int) error {
	return r.db().WithContext(ctx).Model(&model.UploadChunk{}).
		Where("upload_id = ? AND chunk_index = -1", uploadID).
		Updates(map[string]interface{}{
			"completed":   completedChunks,
			"retry_count": gorm.Expr("retry_count + 1"),
		}).Error
}

func (r *ChunkRepository) UpdateChunkCompleted(ctx context.Context, uploadID string, chunkIndex int, chunkHash string) error {
	return r.db().WithContext(ctx).Where("upload_id = ? AND chunk_index = ?", uploadID, chunkIndex).
		Updates(map[string]interface{}{
			"completed":  true,
			"chunk_hash": chunkHash,
			"status":     "completed",
		}).Error
}

func (r *ChunkRepository) CountCompletedChunks(ctx context.Context, uploadID string) (int64, error) {
	var count int64
	err := r.db().WithContext(ctx).Model(&model.UploadChunk{}).
		Where("upload_id = ? AND chunk_index >= 0 AND completed = true", uploadID).
		Count(&count).Error
	return count, err
}

func (r *ChunkRepository) DeleteByUploadID(ctx context.Context, uploadID string) error {
	return r.db().WithContext(ctx).Where("upload_id = ?", uploadID).Delete(&model.UploadChunk{}).Error
}

func (r *ChunkRepository) GetUploadList(ctx context.Context, page, pageSize int) ([]*model.UploadChunk, int64, error) {
	var chunks []*model.UploadChunk
	var total int64

	// 只获取控制记录（chunk_index = -1）
	query := r.db().WithContext(ctx).Model(&model.UploadChunk{}).Where("chunk_index = -1")

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&chunks).Error

	return chunks, total, err
}

func (r *ChunkRepository) GetIncompleteUploads(ctx context.Context, olderThan int) ([]*model.UploadChunk, error) {
	var chunks []*model.UploadChunk
	err := r.db().WithContext(ctx).Where("chunk_index = -1 AND status != 'completed' AND created_at < datetime('now', '-' || ? || ' hours')", olderThan).
		Find(&chunks).Error
	return chunks, err
}

func (r *ChunkRepository) GetOldChunks(ctx context.Context, cutoffTime time.Time) ([]*model.UploadChunk, error) {
	var oldChunks []*model.UploadChunk
	err := r.db().WithContext(ctx).Where("created_at < ? AND chunk_index = -1", cutoffTime).Find(&oldChunks).Error
	return oldChunks, err
}

func (r *ChunkRepository) DeleteChunksByUploadIDs(ctx context.Context, uploadIDs []string) (int, error) {
	if len(uploadIDs) == 0 {
		return 0, nil
	}

	count := 0
	for _, uploadID := range uploadIDs {
		if err := r.db().WithContext(ctx).Where("upload_id = ?", uploadID).Delete(&model.UploadChunk{}).Error; err != nil {
			continue // 记录错误但继续处理其他上传
		}
		count++
	}
	return count, nil
}

func (r *ChunkRepository) GetUploadedChunkIndexes(ctx context.Context, uploadID string) ([]int, error) {
	var uploadedChunks []int
	err := r.db().WithContext(ctx).Model(&model.UploadChunk{}).
		Where("upload_id = ? AND completed = true AND chunk_index >= 0", uploadID).
		Pluck("chunk_index", &uploadedChunks).Error
	return uploadedChunks, err
}

func (r *ChunkRepository) FirstOrCreateChunk(ctx context.Context, chunk *model.UploadChunk) error {
	return r.db().WithContext(ctx).Where("upload_id = ? AND chunk_index = ?", chunk.UploadID, chunk.ChunkIndex).
		Assign(chunk).
		FirstOrCreate(chunk).Error
}
