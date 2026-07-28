package chunk

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zy84338719/fileCodeBox/backend/internal/repo/db"
	"github.com/zy84338719/fileCodeBox/backend/internal/repo/db/model"
	"gorm.io/gorm"
)

// =====================================================================
// chunk service 单测（Shape B：DAO 通过全局 db.GetDB() 访问）。
// 注意：与 share/user 不同，chunk 的 NewService() 在构造时即创建 repo，
// 但 repo 内部仍惰性调用 db.GetDB()，故 SetDatabaseInstance 注入依然有效。
// =====================================================================

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, gormDB.AutoMigrate(&model.UploadChunk{}))
	db.SetDatabaseInstance(gormDB)
	t.Cleanup(func() { db.SetDatabaseInstance(nil) })
	return gormDB
}

func newTestService(t *testing.T) *Service {
	t.Helper()
	newTestDB(t)
	return NewService()
}

// 测试：初始化上传，返回控制记录（chunk_index=-1）
func TestInitiateUpload_Success(t *testing.T) {
	svc := newTestService(t)
	resp, err := svc.InitiateUpload(context.Background(), &InitiateUploadReq{
		UploadID:    "upload-001",
		FileName:    "big.zip",
		TotalChunks: 3,
		FileSize:    30,
		ChunkSize:   10,
	})
	require.NoError(t, err)
	assert.Equal(t, "upload-001", resp.UploadID)
	assert.Equal(t, -1, resp.ChunkIndex) // 控制记录
	assert.Equal(t, 3, resp.TotalChunks)
	assert.Equal(t, "pending", resp.Status)
}

// 测试：重复初始化同一 UploadID 拒绝
func TestInitiateUpload_Duplicate(t *testing.T) {
	svc := newTestService(t)
	req := &InitiateUploadReq{UploadID: "dup", TotalChunks: 2}
	_, err := svc.InitiateUpload(context.Background(), req)
	require.NoError(t, err)

	_, err = svc.InitiateUpload(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

// 测试：上传单个分片 → 标记完成
func TestUploadChunk_Success(t *testing.T) {
	svc := newTestService(t)
	_, err := svc.InitiateUpload(context.Background(), &InitiateUploadReq{
		UploadID: "u1", TotalChunks: 2, FileName: "f",
	})
	require.NoError(t, err)

	resp, err := svc.UploadChunk(context.Background(), &UploadChunkReq{
		UploadID:   "u1",
		ChunkIndex: 0,
		ChunkHash:  "hash0",
		ChunkSize:  10,
	})
	require.NoError(t, err)
	assert.Equal(t, 0, resp.ChunkIndex)
	assert.True(t, resp.Completed)
	assert.Equal(t, "completed", resp.Status)
}

// 测试：无效分片索引（越界）拒绝
func TestUploadChunk_InvalidIndex(t *testing.T) {
	svc := newTestService(t)
	_, err := svc.InitiateUpload(context.Background(), &InitiateUploadReq{
		UploadID: "u2", TotalChunks: 2,
	})
	require.NoError(t, err)

	_, err = svc.UploadChunk(context.Background(), &UploadChunkReq{
		UploadID:   "u2",
		ChunkIndex: 5, // 越界（TotalChunks=2）
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid chunk index")
}

// 测试：上传分片到不存在的 UploadID 拒绝
func TestUploadChunk_UploadNotFound(t *testing.T) {
	svc := newTestService(t)
	_, err := svc.UploadChunk(context.Background(), &UploadChunkReq{
		UploadID:   "ghost",
		ChunkIndex: 0,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// 测试：上传进度检查
func TestCheckUploadProgress(t *testing.T) {
	svc := newTestService(t)
	_, err := svc.InitiateUpload(context.Background(), &InitiateUploadReq{
		UploadID: "u3", TotalChunks: 3, FileSize: 30,
	})
	require.NoError(t, err)

	// 上传 2/3 分片
	_, err = svc.UploadChunk(context.Background(), &UploadChunkReq{UploadID: "u3", ChunkIndex: 0})
	require.NoError(t, err)
	_, err = svc.UploadChunk(context.Background(), &UploadChunkReq{UploadID: "u3", ChunkIndex: 1})
	require.NoError(t, err)

	progress, err := svc.CheckUploadProgress(context.Background(), "u3")
	require.NoError(t, err)
	assert.Equal(t, 3, progress.TotalChunks)
	assert.Equal(t, int64(2), progress.CompletedChunks)
	assert.NotEqual(t, "completed", progress.Status) // 2/3 未全部完成
}

// 测试：获取已上传分片索引列表
func TestGetUploadedChunkIndexes(t *testing.T) {
	svc := newTestService(t)
	_, err := svc.InitiateUpload(context.Background(), &InitiateUploadReq{
		UploadID: "u4", TotalChunks: 3,
	})
	require.NoError(t, err)
	_, _ = svc.UploadChunk(context.Background(), &UploadChunkReq{UploadID: "u4", ChunkIndex: 1})
	_, _ = svc.UploadChunk(context.Background(), &UploadChunkReq{UploadID: "u4", ChunkIndex: 2})

	indexes, err := svc.GetUploadedChunkIndexes(context.Background(), "u4")
	require.NoError(t, err)
	assert.Contains(t, indexes, 1)
	assert.Contains(t, indexes, 2)
	assert.NotContains(t, indexes, 0)
}
