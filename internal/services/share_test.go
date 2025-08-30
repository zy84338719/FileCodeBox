package services

import (
	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/models"
	"github.com/zy84338719/filecodebox/internal/storage"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// 自动迁移
	db.AutoMigrate(&models.FileCode{}, &models.UploadChunk{}, &models.KeyValue{})
	return db
}

func TestShareText(t *testing.T) {
	db := setupTestDB()
	cfg := &config.Config{
		DataPath: "./test_data",
	}
	storageManager := storage.NewStorageManager(cfg)
	service := NewShareService(db, storageManager, cfg)

	// 测试分享文本
	fileCode, err := service.ShareText("Hello World", 1, "day")
	if err != nil {
		t.Fatalf("ShareText failed: %v", err)
	}

	if fileCode.Text != "Hello World" {
		t.Errorf("Expected text 'Hello World', got '%s'", fileCode.Text)
	}

	if fileCode.Code == "" {
		t.Error("Expected non-empty code")
	}

	// 测试获取文本
	retrievedFile, err := service.GetFileByCode(fileCode.Code, true)
	if err != nil {
		t.Fatalf("GetFileByCode failed: %v", err)
	}

	if retrievedFile.Text != "Hello World" {
		t.Errorf("Expected text 'Hello World', got '%s'", retrievedFile.Text)
	}
}

func TestGenerateCode(t *testing.T) {
	db := setupTestDB()
	cfg := &config.Config{}
	storageManager := storage.NewStorageManager(cfg)
	service := NewShareService(db, storageManager, cfg)

	code1 := service.generateCode()
	code2 := service.generateCode()

	if code1 == code2 {
		t.Error("Generated codes should be unique")
	}

	if len(code1) != 12 {
		t.Errorf("Expected code length 12, got %d", len(code1))
	}
}

func TestParseExpireInfo(t *testing.T) {
	db := setupTestDB()
	cfg := &config.Config{}
	storageManager := storage.NewStorageManager(cfg)
	service := NewShareService(db, storageManager, cfg)

	// 测试天数过期
	expiredAt, expiredCount, usedCount := service.parseExpireInfo(7, "day")
	if expiredAt == nil {
		t.Error("Expected expiredAt to be set for day expire style")
	}
	if expiredCount != -1 {
		t.Errorf("Expected expiredCount -1, got %d", expiredCount)
	}
	if usedCount != 0 {
		t.Errorf("Expected usedCount 0, got %d", usedCount)
	}

	// 测试次数过期
	expiredAt, expiredCount, _ = service.parseExpireInfo(5, "count")
	if expiredAt != nil {
		t.Error("Expected expiredAt to be nil for count expire style")
	}
	if expiredCount != 5 {
		t.Errorf("Expected expiredCount 5, got %d", expiredCount)
	}

	// 测试永不过期
	expiredAt, expiredCount, _ = service.parseExpireInfo(0, "forever")
	if expiredAt != nil {
		t.Error("Expected expiredAt to be nil for forever expire style")
	}
	if expiredCount != -1 {
		t.Errorf("Expected expiredCount -1, got %d", expiredCount)
	}
}

func TestFileExpiration(t *testing.T) {
	// 测试过期文件
	pastTime := time.Now().Add(-1 * time.Hour)
	fileCode := &models.FileCode{
		ExpiredAt:    &pastTime,
		ExpiredCount: -1,
	}

	if !fileCode.IsExpired() {
		t.Error("Expected file to be expired")
	}

	// 测试未过期文件
	futureTime := time.Now().Add(1 * time.Hour)
	fileCode.ExpiredAt = &futureTime

	if fileCode.IsExpired() {
		t.Error("Expected file to not be expired")
	}

	// 测试次数过期
	fileCode.ExpiredAt = nil
	fileCode.ExpiredCount = 0

	if !fileCode.IsExpired() {
		t.Error("Expected file to be expired when count is 0")
	}

	// 测试有剩余次数
	fileCode.ExpiredCount = 5
	if fileCode.IsExpired() {
		t.Error("Expected file to not be expired when count > 0")
	}
}
