package admin

import (
	"context"
	"io"
	"mime/multipart"
	"sync"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zy84338719/fileCodeBox/backend/internal/repo/db"
	"github.com/zy84338719/fileCodeBox/backend/internal/repo/db/model"
	"github.com/zy84338719/fileCodeBox/backend/internal/storage"
	"gorm.io/gorm"
)

// mockStorage 记录被删除的物理文件路径（实现完整 StorageInterface）
type mockStorage struct {
	mu      sync.Mutex
	deleted []string
}

func (m *mockStorage) SaveFile(_ context.Context, _ *multipart.FileHeader, _ string) (*storage.FileOperationResult, error) {
	return &storage.FileOperationResult{}, nil
}
func (m *mockStorage) DeleteFile(_ context.Context, path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deleted = append(m.deleted, path)
	return nil
}
func (m *mockStorage) GetFile(_ context.Context, _ string) ([]byte, error) { return nil, nil }
func (m *mockStorage) FileExists(_ context.Context, _ string) bool         { return false }
func (m *mockStorage) GetFileSize(_ context.Context, _ string) (int64, error) {
	return 0, nil
}
func (m *mockStorage) GetFileURL(_ context.Context, _ string) (string, error) {
	return "", nil
}
func (m *mockStorage) GetFileReader(_ context.Context, _ string) (io.ReadCloser, int64, error) {
	return nil, 0, nil
}
func (m *mockStorage) SaveChunk(_ context.Context, _ string, _ int, _ []byte) error { return nil }
func (m *mockStorage) MergeChunks(_ context.Context, _ string, _ int, _ string) error {
	return nil
}
func (m *mockStorage) CleanChunks(_ context.Context, _ string) error { return nil }

func newAdminTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	g, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, g.AutoMigrate(&model.FileCode{}))
	db.SetDatabaseInstance(g)
	t.Cleanup(func() { db.SetDatabaseInstance(nil) })
	return g
}

// TestCleanExpiredFiles_DeletesPhysicalFile 验证过期清理同时删 DB 记录和物理文件
func TestCleanExpiredFiles_DeletesPhysicalFile(t *testing.T) {
	newAdminTestDB(t)
	svc := NewService()
	st := &mockStorage{}
	svc.SetStorage(st)

	ctx := context.Background()
	// 过期文件（时间过期）
	past := time.Now().Add(-time.Hour)
	require.NoError(t, svc.fileCodeRepo.Create(ctx, &model.FileCode{
		Code: "EXP1", FilePath: "data/2024", UUIDFileName: "abc.txt",
		Size: 100, ExpiredAt: &past, ExpiredCount: -1,
	}))
	// 未过期文件（不应被删）
	require.NoError(t, svc.fileCodeRepo.Create(ctx, &model.FileCode{
		Code: "OK1", FilePath: "data/2024", UUIDFileName: "ok.txt",
		Size: 50, ExpiredCount: -1,
	}))

	deleted, freed, err := svc.CleanExpiredFiles(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(1), deleted)
	assert.Equal(t, int64(100), freed)

	// 物理文件被删
	st.mu.Lock()
	assert.Len(t, st.deleted, 1)
	assert.Contains(t, st.deleted[0], "abc.txt")
	st.mu.Unlock()

	// DB 记录已删（过期的不存在，未过期的还在）
	_, err = svc.fileCodeRepo.GetByCode(ctx, "EXP1")
	assert.Error(t, err)
	_, err = svc.fileCodeRepo.GetByCode(ctx, "OK1")
	assert.NoError(t, err)
}

// TestCleanExpiredFiles_NoStorage 无 storage 时不崩，仅删 DB
func TestCleanExpiredFiles_NoStorage(t *testing.T) {
	newAdminTestDB(t)
	svc := NewService() // 无 storage

	ctx := context.Background()
	past := time.Now().Add(-time.Hour)
	require.NoError(t, svc.fileCodeRepo.Create(ctx, &model.FileCode{
		Code: "EXP2", FilePath: "x", UUIDFileName: "y", Size: 10,
		ExpiredAt: &past, ExpiredCount: -1,
	}))

	deleted, _, err := svc.CleanExpiredFiles(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(1), deleted)
}
