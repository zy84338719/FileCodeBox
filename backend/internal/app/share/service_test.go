package share

import (
	"context"
	"io"
	"mime/multipart"
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

// =====================================================================
// share service 单测
//
// 测试策略（遵循项目既有模式 notify_test.go / presign_test.go）：
//   - Shape B service：构造函数不接受 *gorm.DB，DAO 通过 db.GetDB() 全局获取
//   - 因此用 db.SetDatabaseInstance(sqlite内存) 注入测试 DB
//   - storage 用 hand-rolled mock（实现 storage.StorageInterface）
//   - userService / notifySvc 用 hand-rolled mock（实现对应接口）
// =====================================================================

// newTestDB 构造内存 sqlite + AutoMigrate FileCode（首个使用全局注入模式的测试）。
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, gormDB.AutoMigrate(&model.FileCode{}))
	// 注入到全局，使 DAO 的 db.GetDB() 指向测试库
	db.SetDatabaseInstance(gormDB)
	t.Cleanup(func() {
		// 还原全局为 nil，避免污染其他测试
		db.SetDatabaseInstance(nil)
	})
	return gormDB
}

// newTestService 构造注入 mock 依赖的 share service。
func newTestService(t *testing.T) (*Service, *mockStorage, *mockUserService, *mockNotifyService) {
	t.Helper()
	newTestDB(t)
	st := newMockStorage()
	usr := newMockUserService()
	notify := newMockNotifyService()
	svc := NewService("http://localhost:12345", st)
	svc.SetUserService(usr)
	svc.SetNotifyService(notify)
	return svc, st, usr, notify
}

// ===== mock: storage.StorageInterface =====

type mockStorage struct {
	deletedPath  string
	deleteCalled bool
	deleteErr    error
}

func newMockStorage() *mockStorage { return &mockStorage{} }

func (m *mockStorage) SaveFile(_ context.Context, _ *multipart.FileHeader, _ string) (*storage.FileOperationResult, error) {
	return &storage.FileOperationResult{}, nil
}
func (m *mockStorage) DeleteFile(_ context.Context, path string) error {
	m.deleteCalled = true
	m.deletedPath = path
	return m.deleteErr
}
func (m *mockStorage) GetFile(_ context.Context, _ string) ([]byte, error) {
	return nil, nil
}
func (m *mockStorage) FileExists(_ context.Context, _ string) bool { return false }
func (m *mockStorage) GetFileSize(_ context.Context, _ string) (int64, error) {
	return 0, nil
}
func (m *mockStorage) GetFileURL(_ context.Context, _ string) (string, error) {
	return "", nil
}
func (m *mockStorage) GetFileReader(_ context.Context, _ string) (io.ReadCloser, int64, error) {
	return nil, 0, nil
}
func (m *mockStorage) SaveChunk(_ context.Context, _ string, _ int, _ []byte) error {
	return nil
}
func (m *mockStorage) MergeChunks(_ context.Context, _ string, _ int, _ string) error {
	return nil
}
func (m *mockStorage) CleanChunks(_ context.Context, _ string) error { return nil }

// ===== mock: UserServiceInterface =====

type mockUserService struct {
	uploadsCalls int64
	storageDelta int64
}

func newMockUserService() *mockUserService { return &mockUserService{} }

func (m *mockUserService) UpdateUserStats(_ uint, statsType string, value int64) error {
	switch statsType {
	case "uploads":
		m.uploadsCalls += value
	case "storage":
		m.storageDelta += value
	}
	return nil
}

// ===== mock: NotifyServiceInterface =====

type mockNotifyService struct {
	called int
	lastUserID uint
}

func newMockNotifyService() *mockNotifyService { return &mockNotifyService{} }

func (m *mockNotifyService) CreateForUserSimple(_ context.Context, userID uint, _, _, _, _ string) error {
	m.called++
	m.lastUserID = userID
	return nil
}

// =====================================================================
// 测试用例
// =====================================================================

// 测试：分享文本成功，记录写入 DB
func TestShareText_Success(t *testing.T) {
	svc, _, usr, _ := newTestService(t)
	uid := uint(1)

	resp, err := svc.ShareText(context.Background(), &ShareTextReq{
		Text:         "hello",
		ExpiredCount: -1, // -1 = 无限次数，否则 0 视为已过期
		UserID:       &uid,
		UploadType:   "authenticated",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Code)
	assert.Equal(t, "hello", resp.Text)
	assert.Equal(t, "authenticated", resp.UploadType)
	// 用户上传统计应 +1
	assert.Equal(t, int64(1), usr.uploadsCalls)

	// DB 应能查回
	fc, err := svc.GetFileByCode(context.Background(), resp.Code)
	require.NoError(t, err)
	assert.Equal(t, "hello", fc.Text)
}

// 测试：取过期分享（ExpiredCount=0 视为已用完）
func TestGetFileByCode_Expired(t *testing.T) {
	svc, _, _, _ := newTestService(t)

	resp, err := svc.ShareText(context.Background(), &ShareTextReq{
		Text:         "expiring",
		ExpiredCount: 0, // 0 → IsExpired()=true
	})
	require.NoError(t, err)

	_, err = svc.GetFileByCode(context.Background(), resp.Code)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

// 测试：取不存在的分享
func TestGetFileByCode_NotFound(t *testing.T) {
	svc, _, _, _ := newTestService(t)

	_, err := svc.GetFileByCode(context.Background(), "NOPE1234")
	assert.Error(t, err)
}

// 测试：按次数过期——UpdateFileUsage 递减 ExpiredCount
func TestUpdateFileUsage_DecrementCount(t *testing.T) {
	svc, _, _, _ := newTestService(t)

	resp, err := svc.ShareText(context.Background(), &ShareTextReq{
		Text:         "counted",
		ExpiredCount: 2,
	})
	require.NoError(t, err)

	// 第一次取件：2 → 1
	require.NoError(t, svc.UpdateFileUsage(context.Background(), resp.Code))
	fc, err := svc.GetFileByCode(context.Background(), resp.Code)
	require.NoError(t, err)
	assert.Equal(t, 1, fc.ExpiredCount)
	assert.Equal(t, 1, fc.UsedCount)

	// 第二次取件：1 → 0（此时变为已过期）
	require.NoError(t, svc.UpdateFileUsage(context.Background(), resp.Code))
	fc2, err := svc.GetFileByCode(context.Background(), resp.Code)
	// ExpiredCount=0 → IsExpired()=true → GetFileByCode 报错
	assert.Error(t, err)
	assert.Nil(t, fc2)
}

// 测试：无限次数分享（ExpiredCount=-1）可反复取件
func TestUpdateFileUsage_Unlimited(t *testing.T) {
	svc, _, _, _ := newTestService(t)

	resp, err := svc.ShareText(context.Background(), &ShareTextReq{
		Text:         "unlimited",
		ExpiredCount: -1,
	})
	require.NoError(t, err)

	for i := 0; i < 5; i++ {
		require.NoError(t, svc.UpdateFileUsage(context.Background(), resp.Code))
	}
	fc, err := svc.GetFileByCode(context.Background(), resp.Code)
	require.NoError(t, err)
	assert.Equal(t, -1, fc.ExpiredCount) // 不递减
	assert.Equal(t, 5, fc.UsedCount)
}

// 测试：按时间过期——ExpiredAt 在过去
func TestGetFileByCode_ExpiredByTime(t *testing.T) {
	svc, _, _, _ := newTestService(t)

	past := time.Now().Add(-1 * time.Hour)
	resp, err := svc.ShareText(context.Background(), &ShareTextReq{
		Text:         "past",
		ExpiredAt:    &past,
		ExpiredCount: -1, // 次数无限，但时间已过
	})
	require.NoError(t, err)

	_, err = svc.GetFileByCode(context.Background(), resp.Code)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

// 测试：删除分享——所有权校验（非 owner 拒绝）
func TestDeleteFileByCode_OwnershipDenied(t *testing.T) {
	svc, _, _, _ := newTestService(t)
	owner := uint(1)

	resp, err := svc.ShareText(context.Background(), &ShareTextReq{
		Text:         "mine",
		ExpiredCount: -1,
		UserID:       &owner,
	})
	require.NoError(t, err)

	// 另一个用户尝试删除 → 拒绝
	err = svc.DeleteFileByCode(context.Background(), resp.Code, 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "无权限")
}

// 测试：删除分享——owner 删除成功
func TestDeleteFileByCode_OwnerSuccess(t *testing.T) {
	svc, _, _, _ := newTestService(t)
	owner := uint(1)

	resp, err := svc.ShareText(context.Background(), &ShareTextReq{
		Text:         "mine",
		ExpiredCount: -1,
		UserID:       &owner,
	})
	require.NoError(t, err)

	require.NoError(t, svc.DeleteFileByCode(context.Background(), resp.Code, owner))
	// 软删除后查不到
	_, err = svc.GetFileByCode(context.Background(), resp.Code)
	assert.Error(t, err)
}

// 测试：GenerateCode 生成 8 位非空码，且多次不重复
func TestGenerateCode(t *testing.T) {
	svc, _, _, _ := newTestService(t)
	c1 := svc.GenerateCode()
	c2 := svc.GenerateCode()
	assert.Len(t, c1, 8)
	assert.Len(t, c2, 8)
	assert.NotEqual(t, c1, c2)
}
