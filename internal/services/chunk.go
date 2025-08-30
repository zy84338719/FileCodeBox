package services

import (
	"crypto/sha256"
	"fmt"
	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/models"
	"github.com/zy84338719/filecodebox/internal/storage"
	"mime/multipart"
	"path/filepath"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ChunkService 分片服务
type ChunkService struct {
	db      *gorm.DB
	storage *storage.StorageManager
	config  *config.Config
}

func NewChunkService(db *gorm.DB, storageManager *storage.StorageManager, config *config.Config) *ChunkService {
	return &ChunkService{
		db:      db,
		storage: storageManager,
		config:  config,
	}
}

// InitChunkUploadResult 初始化上传结果
type InitChunkUploadResult struct {
	Existed        bool   `json:"existed"`
	UploadID       string `json:"upload_id"`
	ChunkSize      int    `json:"chunk_size"`
	TotalChunks    int    `json:"total_chunks"`
	UploadedChunks []int  `json:"uploaded_chunks"`
	FileCode       string `json:"file_code,omitempty"`       // 如果文件已存在，返回文件代码
	ResumePosition int64  `json:"resume_position,omitempty"` // 断点续传位置
	Progress       string `json:"progress,omitempty"`        // 上传进度
}

// InitChunkUpload 初始化分片上传，支持断点续传
func (s *ChunkService) InitChunkUpload(fileName string, fileSize int64, chunkSize int, fileHash string) (*InitChunkUploadResult, error) {
	// 0. 验证文件大小
	if fileSize > s.config.UploadSize {
		maxSizeMB := float64(s.config.UploadSize) / (1024 * 1024)
		return nil, fmt.Errorf("文件大小超过限制，最大为%.2fMB", maxSizeMB)
	}

	// 1. 检查文件是否已存在（秒传功能）
	var existingFile models.FileCode
	err := s.db.Where("file_hash = ? AND size = ? AND deleted_at IS NULL", fileHash, fileSize).First(&existingFile).Error
	if err == nil && !existingFile.IsExpired() {
		return &InitChunkUploadResult{
			Existed:  true,
			FileCode: existingFile.Code,
		}, nil
	}

	// 2. 检查是否有未完成的上传会话（断点续传）
	var existingChunk models.UploadChunk
	err = s.db.Where("chunk_hash = ? AND file_size = ? AND chunk_index = -1", fileHash, fileSize).First(&existingChunk).Error

	var uploadID string
	var totalChunks int

	if err == nil {
		// 断点续传场景
		uploadID = existingChunk.UploadID
		totalChunks = existingChunk.TotalChunks
	} else {
		// 新上传场景
		uploadID = uuid.New().String()
		totalChunks = int((fileSize + int64(chunkSize) - 1) / int64(chunkSize))

		chunk := &models.UploadChunk{
			UploadID:    uploadID,
			ChunkIndex:  -1, // 标记为控制记录
			TotalChunks: totalChunks,
			FileSize:    fileSize,
			ChunkSize:   chunkSize,
			ChunkHash:   fileHash,
			FileName:    fileName,
		}

		if err := s.db.Create(chunk).Error; err != nil {
			return nil, err
		}
	}

	// 3. 获取已上传的分片列表
	var uploadedChunks []int
	err = s.db.Model(&models.UploadChunk{}).
		Where("upload_id = ? AND completed = true AND chunk_index >= 0", uploadID).
		Pluck("chunk_index", &uploadedChunks).Error
	if err != nil {
		return nil, err
	}

	// 4. 计算断点续传信息
	resumePosition := int64(len(uploadedChunks)) * int64(chunkSize)
	if resumePosition > fileSize {
		resumePosition = fileSize
	}

	progressPercent := float64(len(uploadedChunks)) / float64(totalChunks) * 100
	progress := fmt.Sprintf("%.2f%%", progressPercent)

	return &InitChunkUploadResult{
		Existed:        false,
		UploadID:       uploadID,
		ChunkSize:      chunkSize,
		TotalChunks:    totalChunks,
		UploadedChunks: uploadedChunks,
		ResumePosition: resumePosition,
		Progress:       progress,
	}, nil
}

// UploadChunk 上传分片，支持断点续传
func (s *ChunkService) UploadChunk(uploadID string, chunkIndex int, file *multipart.FileHeader) (string, error) {
	// 获取上传会话信息
	var chunkInfo models.UploadChunk
	err := s.db.Where("upload_id = ? AND chunk_index = -1", uploadID).First(&chunkInfo).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", fmt.Errorf("上传会话不存在或已过期")
		}
		return "", err
	}

	// 检查分片索引有效性
	if chunkIndex < 0 || chunkIndex >= chunkInfo.TotalChunks {
		return "", fmt.Errorf("无效的分片索引: %d，总分片数: %d", chunkIndex, chunkInfo.TotalChunks)
	}

	// 检查该分片是否已上传
	var existingChunk models.UploadChunk
	err = s.db.Where("upload_id = ? AND chunk_index = ? AND completed = true", uploadID, chunkIndex).First(&existingChunk).Error
	if err == nil {
		// 分片已存在，返回已有的哈希值（实现幂等性）
		return existingChunk.ChunkHash, nil
	}

	// 读取分片数据并计算哈希
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %v", err)
	}
	defer src.Close()

	data := make([]byte, file.Size)
	_, err = src.Read(data)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %v", err)
	}

	chunkHash := fmt.Sprintf("%x", sha256.Sum256(data))

	// 保存分片到存储（使用路径管理器）
	storageInterface := s.storage.GetStorage()
	err = storageInterface.SaveChunk(uploadID, chunkIndex, data, chunkHash)
	if err != nil {
		return "", fmt.Errorf("保存分片失败: %v", err)
	}

	// 更新或创建分片记录
	chunkRecord := &models.UploadChunk{
		UploadID:    uploadID,
		ChunkIndex:  chunkIndex,
		ChunkHash:   chunkHash,
		TotalChunks: chunkInfo.TotalChunks,
		FileSize:    chunkInfo.FileSize,
		ChunkSize:   chunkInfo.ChunkSize,
		FileName:    chunkInfo.FileName,
		Completed:   true,
	}

	err = s.db.Where("upload_id = ? AND chunk_index = ?", uploadID, chunkIndex).
		Assign(chunkRecord).
		FirstOrCreate(chunkRecord).Error
	if err != nil {
		return "", fmt.Errorf("保存分片记录失败: %v", err)
	}

	return chunkHash, nil
}

// CompleteUpload 完成上传
func (s *ChunkService) CompleteUpload(uploadID string, expireValue int, expireStyle string) (*models.FileCode, error) {
	// 获取上传基本信息
	var chunkInfo models.UploadChunk
	err := s.db.Where("upload_id = ? AND chunk_index = -1", uploadID).First(&chunkInfo).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("上传会话不存在")
		}
		return nil, err
	}

	// 验证所有分片是否完整
	var completedCount int64
	err = s.db.Model(&models.UploadChunk{}).
		Where("upload_id = ? AND completed = true AND chunk_index >= 0", uploadID).
		Count(&completedCount).Error
	if err != nil {
		return nil, err
	}

	if int(completedCount) != chunkInfo.TotalChunks {
		return nil, fmt.Errorf("分片不完整")
	}

	// 获取文件路径
	path, suffix, prefix, uuidFileName := storage.GenerateFileInfo(chunkInfo.FileName, uploadID)

	// 使用path和uuidFileName构建savePath
	savePath := filepath.Join(path, uuidFileName)

	// 合并文件
	storageInterface := s.storage.GetStorage()
	err = storageInterface.MergeChunks(uploadID, &chunkInfo, savePath)
	if err != nil {
		return nil, fmt.Errorf("合并文件失败: %v", err)
	}

	// 生成代码和过期信息
	shareService := &ShareService{db: s.db, config: s.config}
	code := shareService.generateCode()
	expiredAt, expiredCount, usedCount := shareService.parseExpireInfo(expireValue, expireStyle)

	// 创建文件记录
	fileCode := &models.FileCode{
		Code:         code,
		FileHash:     chunkInfo.ChunkHash,
		IsChunked:    true,
		UploadID:     uploadID,
		Size:         chunkInfo.FileSize,
		ExpiredAt:    expiredAt,
		ExpiredCount: expiredCount,
		UsedCount:    usedCount,
		FilePath:     path,
		UUIDFileName: uuidFileName,
		Prefix:       prefix,
		Suffix:       suffix,
	}

	if err := s.db.Create(fileCode).Error; err != nil {
		return nil, err
	}

	// 清理临时文件
	err = storageInterface.CleanChunks(uploadID)
	if err != nil {
		// 记录日志，但不返回错误
		fmt.Printf("清理临时文件失败: %v\n", err)
	}

	return fileCode, nil
}

// UploadStatus 上传状态结构
type UploadStatus struct {
	UploadID       string `json:"upload_id"`
	FileName       string `json:"file_name"`
	FileSize       int64  `json:"file_size"`
	TotalChunks    int    `json:"total_chunks"`
	UploadedChunks []int  `json:"uploaded_chunks"`
	MissingChunks  []int  `json:"missing_chunks"`
	Progress       string `json:"progress"`
	ResumePosition int64  `json:"resume_position"`
	Status         string `json:"status"` // uploading, completed, failed, cancelled
}

// GetUploadStatus 获取上传状态
func (s *ChunkService) GetUploadStatus(uploadID string) (*UploadStatus, error) {
	// 获取上传基本信息
	var chunkInfo models.UploadChunk
	err := s.db.Where("upload_id = ? AND chunk_index = -1", uploadID).First(&chunkInfo).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("上传会话不存在")
		}
		return nil, err
	}

	// 获取已上传的分片列表
	var uploadedChunks []int
	err = s.db.Model(&models.UploadChunk{}).
		Where("upload_id = ? AND completed = true AND chunk_index >= 0", uploadID).
		Pluck("chunk_index", &uploadedChunks).Error
	if err != nil {
		return nil, err
	}

	// 计算缺失的分片
	var missingChunks []int
	uploadedSet := make(map[int]bool)
	for _, chunk := range uploadedChunks {
		uploadedSet[chunk] = true
	}

	for i := 0; i < chunkInfo.TotalChunks; i++ {
		if !uploadedSet[i] {
			missingChunks = append(missingChunks, i)
		}
	}

	// 计算进度和状态
	progressPercent := float64(len(uploadedChunks)) / float64(chunkInfo.TotalChunks) * 100
	progress := fmt.Sprintf("%.2f%%", progressPercent)

	resumePosition := int64(len(uploadedChunks)) * int64(chunkInfo.ChunkSize)
	if resumePosition > chunkInfo.FileSize {
		resumePosition = chunkInfo.FileSize
	}

	status := "uploading"
	if len(uploadedChunks) == chunkInfo.TotalChunks {
		status = "completed"
	}

	return &UploadStatus{
		UploadID:       uploadID,
		FileName:       chunkInfo.FileName,
		FileSize:       chunkInfo.FileSize,
		TotalChunks:    chunkInfo.TotalChunks,
		UploadedChunks: uploadedChunks,
		MissingChunks:  missingChunks,
		Progress:       progress,
		ResumePosition: resumePosition,
		Status:         status,
	}, nil
}

// VerifyChunk 验证分片完整性
func (s *ChunkService) VerifyChunk(uploadID string, chunkIndex int, expectedHash string) (bool, error) {
	var chunk models.UploadChunk
	err := s.db.Where("upload_id = ? AND chunk_index = ? AND completed = true", uploadID, chunkIndex).First(&chunk).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil // 分片不存在
		}
		return false, err
	}

	return chunk.ChunkHash == expectedHash, nil
}

// CancelUpload 取消上传
func (s *ChunkService) CancelUpload(uploadID string) error {
	// 删除所有相关的分片记录
	err := s.db.Where("upload_id = ?", uploadID).Delete(&models.UploadChunk{}).Error
	if err != nil {
		return fmt.Errorf("删除上传记录失败: %v", err)
	}

	// 清理临时文件
	storageInterface := s.storage.GetStorage()
	err = storageInterface.CleanChunks(uploadID)
	if err != nil {
		// 记录日志，但不返回错误
		fmt.Printf("清理临时文件失败: %v\n", err)
	}

	return nil
}
