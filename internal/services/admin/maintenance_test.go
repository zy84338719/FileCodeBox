package admin_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/zy84338719/filecodebox/internal/models"
	dbmodels "github.com/zy84338719/filecodebox/internal/models/db"
	"gorm.io/gorm"
)

func TestCleanupInvalidFilesRemovesMissingRecords(t *testing.T) {
	svc, repo, manager := setupAdminTestService(t)

	// create valid file on disk
	validDir := filepath.Join("files", "2025")
	validName := "file.bin"
	validFullPath := filepath.Join(manager.Storage.StoragePath, validDir, validName)
	if err := os.MkdirAll(filepath.Dir(validFullPath), 0755); err != nil {
		t.Fatalf("failed to create storage dir: %v", err)
	}
	if err := os.WriteFile(validFullPath, []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write valid file: %v", err)
	}

	valid := &models.FileCode{
		Code:         "valid",
		FilePath:     validDir,
		UUIDFileName: validName,
		Size:         int64(len("data")),
	}
	if err := repo.FileCode.Create(valid); err != nil {
		t.Fatalf("failed to create valid record: %v", err)
	}

	missing := &models.FileCode{
		Code:         "missing",
		FilePath:     filepath.Join("files", "missing"),
		UUIDFileName: "ghost.bin",
		Size:         10,
	}
	if err := repo.FileCode.Create(missing); err != nil {
		t.Fatalf("failed to create missing record: %v", err)
	}

	cleaned, err := svc.CleanupInvalidFiles()
	if err != nil {
		t.Fatalf("CleanupInvalidFiles returned error: %v", err)
	}
	if cleaned != 1 {
		t.Fatalf("expected 1 record cleaned, got %d", cleaned)
	}

	if _, err := repo.FileCode.GetByCode("missing"); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected missing record to be removed, err=%v", err)
	}

	if _, err := repo.FileCode.GetByCode("valid"); err != nil {
		t.Fatalf("expected valid record to remain: %v", err)
	}
}

func TestCleanTempFilesRemovesOldChunks(t *testing.T) {
	svc, repo, manager := setupAdminTestService(t)

	uploadID := "session-123"
	chunk := &dbmodels.UploadChunk{
		UploadID:   uploadID,
		ChunkIndex: -1,
		Status:     "pending",
	}
	if err := repo.Chunk.Create(chunk); err != nil {
		t.Fatalf("failed to create chunk: %v", err)
	}

	// backdate the chunk so it qualifies as old
	oldTime := time.Now().Add(-48 * time.Hour)
	if err := repo.DB().Model(&dbmodels.UploadChunk{}).
		Where("upload_id = ? AND chunk_index = -1", uploadID).
		Update("created_at", oldTime).Error; err != nil {
		t.Fatalf("failed to backdate chunk: %v", err)
	}

	chunkDir := filepath.Join(manager.Storage.StoragePath, "chunks", uploadID)
	if err := os.MkdirAll(chunkDir, 0755); err != nil {
		t.Fatalf("failed to create chunk dir: %v", err)
	}

	cleaned, err := svc.CleanTempFiles()
	if err != nil {
		t.Fatalf("CleanTempFiles returned error: %v", err)
	}
	if cleaned != 1 {
		t.Fatalf("expected 1 upload cleaned, got %d", cleaned)
	}

	if _, err := repo.Chunk.GetByUploadID(uploadID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected chunk record to be removed, err=%v", err)
	}

	if _, err := os.Stat(chunkDir); !os.IsNotExist(err) {
		t.Fatalf("expected chunk directory to be removed, stat err=%v", err)
	}
}
