package tasks

import (
	"time"

	"github.com/zy84338719/filecodebox/internal/repository"
	"github.com/zy84338719/filecodebox/internal/storage"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

// TaskManager 任务管理器
type TaskManager struct {
	daoManager  *repository.RepositoryManager
	storage     *storage.StorageManager
	cron        *cron.Cron
	pathManager *storage.PathManager
}

func NewTaskManager(daoManager *repository.RepositoryManager, storageManager *storage.StorageManager, dataPath string) *TaskManager {
	// 创建路径管理器
	pathManager := storage.NewPathManager(dataPath)

	return &TaskManager{
		daoManager:  daoManager,
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

	// 使用 DAO 获取过期文件
	expiredFiles, err := tm.daoManager.FileCode.GetExpiredFiles()
	if err != nil {
		logrus.Error("查找过期文件失败:", err)
		return
	}

	if len(expiredFiles) == 0 {
		logrus.Info("没有发现过期文件")
		return
	}

	storageInterface := tm.storage.GetStorage()
	successCount := 0

	for _, file := range expiredFiles {
		// 删除实际文件
		if err := storageInterface.DeleteFile(&file); err != nil {
			logrus.Warnf("删除文件失败 %s: %v", file.Code, err)
		}
	}

	// 使用 DAO 批量删除数据库记录
	successCount, err = tm.daoManager.FileCode.DeleteExpiredFiles(expiredFiles)
	if err != nil {
		logrus.Errorf("批量删除过期文件失败: %v", err)
	}

	logrus.Infof("清理过期文件完成，共清理 %d 个文件", successCount)
}

// cleanTempFiles 清理临时文件
func (tm *TaskManager) cleanTempFiles() {
	logrus.Info("开始清理临时文件")

	// 清理超过24小时的未完成上传
	cutoff := time.Now().Add(-24 * time.Hour)

	// 使用 DAO 获取旧分片记录
	oldChunks, err := tm.daoManager.Chunk.GetOldChunks(cutoff)
	if err != nil {
		logrus.Error("查找旧分片记录失败:", err)
		return
	}

	if len(oldChunks) == 0 {
		logrus.Info("没有发现需要清理的临时文件")
		return
	}

	storageInterface := tm.storage.GetStorage()
	uploadIDs := make([]string, 0, len(oldChunks))

	for _, chunk := range oldChunks {
		// 清理分片文件
		if err := storageInterface.CleanChunks(chunk.UploadID); err != nil {
			logrus.Warnf("清理分片文件失败 %s: %v", chunk.UploadID, err)
		}
		uploadIDs = append(uploadIDs, chunk.UploadID)
	}

	// 使用 DAO 批量删除分片记录
	successCount, err := tm.daoManager.Chunk.DeleteChunksByUploadIDs(uploadIDs)
	if err != nil {
		logrus.Errorf("批量删除分片记录失败: %v", err)
	}

	logrus.Infof("清理临时文件完成，共清理 %d 个上传会话", successCount)
}
