package notify

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/zy84338719/fileCodeBox/backend/internal/repo/db/model"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Notify{}))
	return db
}

func newTestService(t *testing.T) *Service {
	t.Helper()
	return NewService(newTestDB(t))
}

// TestCreate_CreateNotify 创建通知
func TestCreate_CreateNotify(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	item, err := svc.Create(ctx, CreateReq{
		Title:    "系统升级通知",
		Content:  "今晚 0 点系统升级",
		Type:     "maintenance",
		Level:    "warning",
		AuthorID: 1,
	})
	require.NoError(t, err)
	assert.NotZero(t, item.ID)
	assert.Equal(t, "系统升级通知", item.Title)
	assert.Equal(t, "maintenance", item.Type)
	assert.Equal(t, "warning", item.Level)
	assert.Equal(t, 1, item.Status, "默认 status=1 发布")
}

// TestCreate_Defaults 不传 type/level/status 用默认值
func TestCreate_Defaults(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	item, err := svc.Create(ctx, CreateReq{
		Title:   "默认",
		Content: "默认内容",
	})
	require.NoError(t, err)
	assert.Equal(t, "system", item.Type)
	assert.Equal(t, "info", item.Level)
	assert.Equal(t, 1, item.Status)
}

// TestCreate_InvalidParam title/content 必填
func TestCreate_InvalidParam(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	_, err := svc.Create(ctx, CreateReq{Title: "", Content: "x"})
	assert.ErrorIs(t, err, ErrInvalidParam)
	_, err = svc.Create(ctx, CreateReq{Title: "x", Content: ""})
	assert.ErrorIs(t, err, ErrInvalidParam)
}

// TestGet_GetByID 按 ID 查
func TestGet_GetByID(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	created, err := svc.Create(ctx, CreateReq{Title: "T", Content: "C"})
	require.NoError(t, err)

	got, err := svc.Get(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, "T", got.Title)
}

// TestGet_NotFound ID 不存在
func TestGet_NotFound(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	_, err := svc.Get(ctx, 99999)
	assert.ErrorIs(t, err, ErrNotifyNotFound)
}

// TestUpdate_PartialUpdate 部分字段更新
func TestUpdate_PartialUpdate(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	created, err := svc.Create(ctx, CreateReq{Title: "原", Content: "原内容", Type: "system"})
	require.NoError(t, err)

	newTitle := "新标题"
	newStatus := 2
	require.NoError(t, svc.Update(ctx, UpdateReq{
		ID:     created.ID,
		Title:  &newTitle,
		Status: &newStatus,
	}))

	got, err := svc.Get(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "新标题", got.Title)
	assert.Equal(t, 2, got.Status)
	assert.Equal(t, "system", got.Type, "未传 type 不变")
	assert.Equal(t, "原内容", got.Content, "未传 content 不变")
}

// TestUpdate_NotFound 更新不存在的 ID
func TestUpdate_NotFound(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	newTitle := "x"
	err := svc.Update(ctx, UpdateReq{ID: 99999, Title: &newTitle})
	assert.ErrorIs(t, err, ErrNotifyNotFound)
}

// TestDelete 删除
func TestDelete(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	created, err := svc.Create(ctx, CreateReq{Title: "T", Content: "C"})
	require.NoError(t, err)

	require.NoError(t, svc.Delete(ctx, created.ID))
	_, err = svc.Get(ctx, created.ID)
	assert.ErrorIs(t, err, ErrNotifyNotFound)
}

// TestDelete_NotFound 删除不存在的
func TestDelete_NotFound(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	err := svc.Delete(ctx, 99999)
	assert.ErrorIs(t, err, ErrNotifyNotFound)
}

// TestList_Pagination 列表分页
func TestList_Pagination(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	// 创建 5 条
	for i := 0; i < 5; i++ {
		_, err := svc.Create(ctx, CreateReq{Title: "T", Content: "C"})
		require.NoError(t, err)
	}
	data, err := svc.List(ctx, ListReq{Page: 1, PageSize: 3})
	require.NoError(t, err)
	assert.Equal(t, int64(5), data.Total)
	assert.Len(t, data.Items, 3)
	assert.Equal(t, 1, data.Page)
	assert.Equal(t, 3, data.PageSize)

	data, err = svc.List(ctx, ListReq{Page: 2, PageSize: 3})
	require.NoError(t, err)
	assert.Len(t, data.Items, 2, "第二页剩 2 条")
}

// TestList_FilterByType 按 type 过滤
func TestList_FilterByType(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	_, err := svc.Create(ctx, CreateReq{Title: "M1", Content: "x", Type: "maintenance"})
	require.NoError(t, err)
	_, err = svc.Create(ctx, CreateReq{Title: "M2", Content: "x", Type: "maintenance"})
	require.NoError(t, err)
	_, err = svc.Create(ctx, CreateReq{Title: "S1", Content: "x", Type: "system"})
	require.NoError(t, err)

	data, err := svc.List(ctx, ListReq{Page: 1, PageSize: 20, Type: "maintenance"})
	require.NoError(t, err)
	assert.Equal(t, int64(2), data.Total)
	for _, item := range data.Items {
		assert.Equal(t, "maintenance", item.Type)
	}
}

// TestList_DefaultPagination 默认分页参数
func TestList_DefaultPagination(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	data, err := svc.List(ctx, ListReq{Page: 0, PageSize: 0})
	require.NoError(t, err)
	assert.Equal(t, 1, data.Page, "Page=0 → 1")
	assert.Equal(t, 20, data.PageSize, "PageSize=0 → 20")
}

// TestList_PageSizeTooLarge 超过 100 → 用默认 20
func TestList_PageSizeTooLarge(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	data, err := svc.List(ctx, ListReq{Page: 1, PageSize: 9999})
	require.NoError(t, err)
	assert.Equal(t, 20, data.PageSize, "PageSize>100 走默认 20")
}

// TestActive_OnlyActive 当前活跃通知（公开 API）
func TestActive_OnlyActive(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	now := time.Now()

	// 1. 状态 1 (发布) + 无时间窗口 → 活跃
	_, err := svc.Create(ctx, CreateReq{Title: "A1", Content: "x", Status: 1})
	require.NoError(t, err)
	// 2. 状态 2 (下线) → 不活跃
	_, err = svc.Create(ctx, CreateReq{Title: "A2", Content: "x", Status: 2})
	require.NoError(t, err)
	// 3. 状态 1 + start_at 未来 → 不活跃
	future := now.Add(2 * time.Hour)
	_, err = svc.Create(ctx, CreateReq{Title: "A3", Content: "x", Status: 1, StartAt: &future})
	require.NoError(t, err)
	// 4. 状态 1 + end_at 过去 → 不活跃
	past := now.Add(-2 * time.Hour)
	_, err = svc.Create(ctx, CreateReq{Title: "A4", Content: "x", Status: 1, EndAt: &past})
	require.NoError(t, err)
	// 5. 状态 1 + 时间窗口内 → 活跃
	start := now.Add(-1 * time.Hour)
	end := now.Add(1 * time.Hour)
	_, err = svc.Create(ctx, CreateReq{Title: "A5", Content: "x", Status: 1, StartAt: &start, EndAt: &end})
	require.NoError(t, err)

	active, err := svc.Active(ctx, "")
	require.NoError(t, err)
	// 应该有 A1 和 A5
	assert.Len(t, active, 2)
	titles := make([]string, len(active))
	for i, a := range active {
		titles[i] = a.Title
	}
	assert.Contains(t, titles, "A1")
	assert.Contains(t, titles, "A5")
}

// TestActive_FilterByType 按 type 过滤
func TestActive_FilterByType(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	_, err := svc.Create(ctx, CreateReq{Title: "M", Content: "x", Type: "maintenance", Status: 1})
	require.NoError(t, err)
	_, err = svc.Create(ctx, CreateReq{Title: "S", Content: "x", Type: "system", Status: 1})
	require.NoError(t, err)

	active, err := svc.Active(ctx, "maintenance")
	require.NoError(t, err)
	assert.Len(t, active, 1)
	assert.Equal(t, "M", active[0].Title)
}

// TestParseID 解析 path 中的 id
func TestParseID(t *testing.T) {
	id, err := ParseID("42")
	require.NoError(t, err)
	assert.Equal(t, uint(42), id)

	_, err = ParseID("not-a-number")
	assert.ErrorIs(t, err, ErrInvalidParam)

	_, err = ParseID("")
	assert.ErrorIs(t, err, ErrInvalidParam)
}
