package tasks

import (
	"time"

	"github.com/zy84338719/filecodebox/internal/models"
	"github.com/zy84338719/filecodebox/internal/storage"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// TaskManager 任务管理器
type TaskManager struct {
	db          *gorm.DB
	storage     *storage.StorageManager
	cron        *cron.Cron
	pathManager *storage.PathManager
}

func NewTaskManager(db *gorm.DB, storageManager *storage.StorageManager, dataPath string) *TaskManager {
	// 创建路径管理器
	pathManager := storage.NewPathManager(dataPath)

	return &TaskManager{
		db:          db,
		storage:     storageManager,
		cron:        cron.New(),
		pathManager: pathManager,
	}
}

// Start 启动任务管理器
func (tm *TaskManager) Start() {
	// 每小时清理一次过期文件
	if _, err := tm.cron.AddFunc("0 * * * *", tm.cleanExpiredFiles); err != nil {
		logrus.WithError(err).Error("添加过期文件清理任务失败")
	}

	// 每天清理一次临时文件
	if _, err := tm.cron.AddFunc("0 2 * * *", tm.cleanTempFiles); err != nil {
		logrus.WithError(err).Error("添加临时文件清理任务失败")
	}

	tm.cron.Start()
	logrus.Info("任务管理器已启动")
}

// Stop 停止任务管理器
func (tm *TaskManager) Stop() {
	tm.cron.Stop()
	logrus.Info("任务管理器已停止")
}

// cleanExpiredFiles 清理过期文件
func (tm *TaskManager) cleanExpiredFiles() {
	logrus.Info("开始清理过期文件")

	var expiredFiles []models.FileCode
	now := time.Now()

	// 查找过期文件
	err := tm.db.Where("(expired_at IS NOT NULL AND expired_at < ?) OR expired_count = 0", now).Find(&expiredFiles).Error
	if err != nil {
		logrus.Error("查找过期文件失败:", err)
		return
	}

	count := 0
	storageInterface := tm.storage.GetStorage()

	for _, file := range expiredFiles {
		// 删除实际文件
		if err := storageInterface.DeleteFile(&file); err != nil {
			logrus.Warnf("删除文件失败 %s: %v", file.Code, err)
		}

		// 删除数据库记录
		if err := tm.db.Delete(&file).Error; err != nil {
			logrus.Warnf("删除数据库记录失败 %s: %v", file.Code, err)
		} else {
			count++
		}
	}

	logrus.Infof("清理过期文件完成，共清理 %d 个文件", count)
}

// cleanTempFiles 清理临时文件
func (tm *TaskManager) cleanTempFiles() {
	logrus.Info("开始清理临时文件")

	// 清理超过24小时的未完成上传
	cutoff := time.Now().Add(-24 * time.Hour)
	var oldChunks []models.UploadChunk

	err := tm.db.Where("created_at < ? AND chunk_index = -1", cutoff).Find(&oldChunks).Error
	if err != nil {
		logrus.Error("查找旧分片记录失败:", err)
		return
	}

	count := 0
	storageInterface := tm.storage.GetStorage()

	for _, chunk := range oldChunks {
		// 清理分片文件
		if err := storageInterface.CleanChunks(chunk.UploadID); err != nil {
			logrus.Warnf("清理分片文件失败 %s: %v", chunk.UploadID, err)
		}

		// 删除相关的所有分片记录
		if err := tm.db.Where("upload_id = ?", chunk.UploadID).Delete(&models.UploadChunk{}).Error; err != nil {
			logrus.Warnf("删除分片记录失败 %s: %v", chunk.UploadID, err)
		} else {
			count++
		}
	}

	logrus.Infof("清理临时文件完成，共清理 %d 个上传会话", count)
}
