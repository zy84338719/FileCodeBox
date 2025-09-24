package share

import (
	"time"

	"github.com/sirupsen/logrus"
	"github.com/zy84338719/filecodebox/internal/models"
)

func (s *Service) IsUploadLoginRequired() bool {
	if s == nil || s.manager == nil || s.manager.Transfer == nil || s.manager.Transfer.Upload == nil {
		return false
	}
	return s.manager.Transfer.Upload.IsLoginRequired()
}

func (s *Service) IsDownloadLoginRequired() bool {
	if s == nil || s.manager == nil || s.manager.Transfer == nil || s.manager.Transfer.Download == nil {
		return false
	}
	return s.manager.Transfer.Download.IsLoginRequired()
}

func (s *Service) recordTransferLog(operation string, fileCode *models.FileCode, userID *uint, ip string, duration time.Duration) {
	if s == nil || s.repositoryManager == nil || s.repositoryManager.TransferLog == nil || fileCode == nil {
		return
	}

	logEntry := &models.TransferLog{
		Operation:  operation,
		FileCodeID: fileCode.ID,
		FileCode:   fileCode.Code,
		FileName:   displayFileName(fileCode),
		FileSize:   fileCode.Size,
		IP:         ip,
		DurationMs: duration.Milliseconds(),
	}

	if userID != nil {
		idCopy := *userID
		logEntry.UserID = &idCopy
		if s.repositoryManager.User != nil {
			if user, err := s.repositoryManager.User.GetByID(*userID); err == nil {
				logEntry.Username = user.Username
			} else if err != nil {
				logrus.WithError(err).Warn("recordTransferLog: fetch user failed")
			}
		}
	}

	if err := s.repositoryManager.TransferLog.Create(logEntry); err != nil {
		logrus.WithError(err).Warn("recordTransferLog: create log failed")
	}
}

func displayFileName(fileCode *models.FileCode) string {
	if fileCode == nil {
		return ""
	}
	if fileCode.UUIDFileName != "" {
		return fileCode.UUIDFileName
	}
	return fileCode.Prefix + fileCode.Suffix
}

func (s *Service) RecordDownloadLog(fileCode *models.FileCode, userID *uint, ip string, duration time.Duration) {
	if !s.IsDownloadLoginRequired() {
		return
	}
	s.recordTransferLog("download", fileCode, userID, ip, duration)
}

func (s *Service) RecordUploadLog(fileCode *models.FileCode, userID *uint, ip string) {
	if !s.IsUploadLoginRequired() {
		return
	}
	s.recordTransferLog("upload", fileCode, userID, ip, 0)
}
