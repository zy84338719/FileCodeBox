package storage

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =====================================================================
// 本地存储（StorageService）单测
// 使用 t.TempDir() 隔离文件系统，覆盖 保存/读取/删除/存在/大小 等核心路径。
// =====================================================================

func newTestStorage(t *testing.T) *StorageService {
	t.Helper()
	return NewStorageService(&StorageConfig{
		Type:     StorageTypeLocal,
		DataPath: t.TempDir(),
		BaseURL:  "http://localhost:12345",
	})
}

// makeFileHeader 构造一个内存中的 multipart.FileHeader，内容为 content。
func makeFileHeader(t *testing.T, filename string, content []byte) *multipart.FileHeader {
	t.Helper()
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	fw, err := w.CreateFormFile("file", filename)
	require.NoError(t, err)
	_, err = fw.Write(content)
	require.NoError(t, err)
	require.NoError(t, w.Close())

	req := httptest.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	_, fh, err := req.FormFile("file")
	require.NoError(t, err)
	return fh
}

// 测试：保存文件 → 读取 → 校验内容
func TestSaveAndRead(t *testing.T) {
	svc := newTestStorage(t)
	ctx := context.Background()
	content := []byte("hello storage")

	res, err := svc.SaveFile(ctx, makeFileHeader(t, "a.txt", content), "sub/a.txt")
	require.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, int64(len(content)), res.FileSize)

	got, err := svc.GetFile(ctx, "sub/a.txt")
	require.NoError(t, err)
	assert.Equal(t, content, got)
}

// 测试：保存文件 → FileExists=true，删除后 → FileExists=false
func TestSaveDeleteExists(t *testing.T) {
	svc := newTestStorage(t)
	ctx := context.Background()

	_, err := svc.SaveFile(ctx, makeFileHeader(t, "b.bin", []byte("data")), "b.bin")
	require.NoError(t, err)
	assert.True(t, svc.FileExists(ctx, "b.bin"))

	require.NoError(t, svc.DeleteFile(ctx, "b.bin"))
	assert.False(t, svc.FileExists(ctx, "b.bin"))

	// 删除不存在的文件应报错
	err = svc.DeleteFile(ctx, "b.bin")
	assert.Error(t, err)
}

// 测试：GetFileSize
func TestGetFileSize(t *testing.T) {
	svc := newTestStorage(t)
	ctx := context.Background()
	content := make([]byte, 2048)

	_, err := svc.SaveFile(ctx, makeFileHeader(t, "c.dat", content), "c.dat")
	require.NoError(t, err)

	size, err := svc.GetFileSize(ctx, "c.dat")
	require.NoError(t, err)
	assert.Equal(t, int64(2048), size)
}

// 测试：SaveFile 自动创建多层目录
func TestSaveFile_CreatesDirs(t *testing.T) {
	svc := newTestStorage(t)
	ctx := context.Background()

	deepPath := filepath.Join("a", "b", "c", "deep.txt")
	_, err := svc.SaveFile(ctx, makeFileHeader(t, "deep.txt", []byte("x")), deepPath)
	require.NoError(t, err)

	// 物理文件应存在
	fullPath := filepath.Join(svc.config.DataPath, deepPath)
	_, err = os.Stat(fullPath)
	require.NoError(t, err)
}

// 测试：GetFile 不存在的文件返回错误
func TestGetFile_NotExist(t *testing.T) {
	svc := newTestStorage(t)
	_, err := svc.GetFile(context.Background(), "nope.txt")
	assert.Error(t, err)
}
