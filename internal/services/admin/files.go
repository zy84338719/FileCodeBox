package admin

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zy84338719/filecodebox/internal/models"

	"github.com/sirupsen/logrus"
)

// GetFiles 获取文件列表
func (s *Service) GetFiles(page, pageSize int, search string) ([]models.FileCode, int64, error) {
	return s.repositoryManager.FileCode.List(page, pageSize, search)
}

// GetFile 获取文件信息
func (s *Service) GetFile(id uint) (*models.FileCode, error) {
	return s.repositoryManager.FileCode.GetByID(id)
}

// GetFileByCode 通过代码获取文件信息
func (s *Service) GetFileByCode(code string) (*models.FileCode, error) {
	return s.repositoryManager.FileCode.GetByCode(code)
}

// DeleteFile 删除文件
func (s *Service) DeleteFile(id uint) error {
	fileCode, err := s.repositoryManager.FileCode.GetByID(id)
	if err != nil {
		return err
	}

	// 删除实际文件
	result := s.storageService.DeleteFileWithResult(fileCode)
	if !result.Success {
		// 记录错误，但不阻止数据库删除
		logrus.WithError(result.Error).
			WithField("code", fileCode.Code).
			Warn("failed to delete physical file while removing file record")
	}

	return s.repositoryManager.FileCode.DeleteByFileCode(fileCode)
}

// DeleteFileByCode 通过代码删除文件
func (s *Service) DeleteFileByCode(code string) error {
	fileCode, err := s.repositoryManager.FileCode.GetByCode(code)
	if err != nil {
		return err
	}

	// 删除实际文件
	result := s.storageService.DeleteFileWithResult(fileCode)
	if !result.Success {
		// 记录错误，但不阻止数据库删除
		logrus.WithError(result.Error).
			WithField("code", fileCode.Code).
			Warn("failed to delete physical file while removing file record")
	}

	return s.repositoryManager.FileCode.DeleteByFileCode(fileCode)
}

// UpdateFile 更新文件
func (s *Service) UpdateFile(id uint, text, name string, expTime time.Time) error {
	updates := map[string]interface{}{
		"text":       text,
		"expired_at": expTime,
	}
	return s.repositoryManager.FileCode.UpdateColumns(id, updates)
}

// UpdateFileByCode 通过代码更新文件
func (s *Service) UpdateFileByCode(code, text, name string, expTime time.Time) error {
	fileCode, err := s.repositoryManager.FileCode.GetByCode(code)
	if err != nil {
		return err
	}
	updates := map[string]interface{}{
		"text":       text,
		"expired_at": expTime,
	}
	return s.repositoryManager.FileCode.UpdateColumns(fileCode.ID, updates)
}

// DownloadFile 下载文件
func (s *Service) DownloadFile(c *gin.Context, id uint) error {
	fileCode, err := s.repositoryManager.FileCode.GetByID(id)
	if err != nil {
		return err
	}

	return s.serveFile(c, fileCode)
}

// DownloadFileByCode 通过代码下载文件
func (s *Service) DownloadFileByCode(c *gin.Context, code string) error {
	fileCode, err := s.repositoryManager.FileCode.GetByCode(code)
	if err != nil {
		return err
	}

	return s.serveFile(c, fileCode)
}

// serveFile 提供文件服务
func (s *Service) serveFile(c *gin.Context, fileCode *models.FileCode) error {
	// 使用存储服务的GetFileResponse方法
	return s.storageService.GetFileResponse(c, fileCode)
}

// ServeFile 提供文件服务 (导出版本)
func (s *Service) ServeFile(c *gin.Context, fileCode *models.FileCode) error {
	return s.serveFile(c, fileCode)
}
