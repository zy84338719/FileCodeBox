package admin_test

import (
	"path/filepath"
	"testing"

	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/models"
	"github.com/zy84338719/filecodebox/internal/repository"
	admin "github.com/zy84338719/filecodebox/internal/services/admin"
	"github.com/zy84338719/filecodebox/internal/storage"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupAdminTestService(t *testing.T) (*admin.Service, *repository.RepositoryManager, *config.ConfigManager) {
	t.Helper()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	if err := db.AutoMigrate(&models.User{}, &models.FileCode{}, &models.UploadChunk{}, &models.TransferLog{}, &models.AdminOperationLog{}); err != nil {
		t.Fatalf("failed to auto-migrate test database: %v", err)
	}

	repo := repository.NewRepositoryManager(db)

	manager := config.NewConfigManager()
	manager.Base.DataPath = tempDir
	manager.Storage.Type = "local"
	manager.Storage.StoragePath = tempDir
	manager.SetDB(db)

	storageService := storage.NewConcreteStorageService(manager)

	svc := admin.NewService(repo, manager, storageService)
	return svc, repo, manager
}
