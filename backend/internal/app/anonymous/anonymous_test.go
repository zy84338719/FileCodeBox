package anonymous

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/glebarez/sqlite"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zy84338719/fileCodeBox/backend/internal/pkg/utils"
	"github.com/zy84338719/fileCodeBox/backend/internal/repo/db"
	"github.com/zy84338719/fileCodeBox/backend/internal/repo/db/dao"
	"github.com/zy84338719/fileCodeBox/backend/internal/repo/db/model"
	"gorm.io/gorm"
)

// newTestService 构造注入 miniredis + sqlite 内存库的 anonymous service。
func newTestService(t *testing.T) (*Service, *miniredis.Miniredis, *gorm.DB) {
	t.Helper()
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { rdb.Close() })

	g, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, g.AutoMigrate(&model.FileCode{}))
	db.SetDatabaseInstance(g)
	t.Cleanup(func() { db.SetDatabaseInstance(nil) })

	repo := dao.NewFileCodeRepository()
	return NewService(rdb, repo), mr, g
}

// TestCodeAlphabet 验证字符表（去掉易混淆字符 0/O/1/I/L）
func TestCodeAlphabet(t *testing.T) {
	for _, ch := range codeAlphabet {
		assert.NotContains(t, "0O1IL", string(ch))
	}
}

// TestGenerateCode_1000Uniqueness 生成 1000 个码不重复
func TestGenerateCode_1000Uniqueness(t *testing.T) {
	svc, _, _ := newTestService(t)
	ctx := context.Background()
	seen := map[string]bool{}
	for i := 0; i < 1000; i++ {
		c := randomCode()
		assert.Len(t, c, codeLength)
		assert.False(t, seen[c], "重复码: %s", c)
		seen[c] = true
	}
	_ = svc
	_ = ctx
}

// TestGenerateCode_Concurrent 并发生成不冲突（SETNX 保证）
func TestGenerateCode_Concurrent(t *testing.T) {
	svc, _, _ := newTestService(t)
	ctx := context.Background()
	exp := time.Now().Add(time.Hour)
	require.NoError(t, svc.fileCodeRepo.Create(ctx, &model.FileCode{Code: "SHARE_C", ExpiredCount: -1}))

	var wg sync.WaitGroup
	codes := make(chan string, 50)
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c, err := svc.GenerateCode(ctx, CodeMeta{ShareCode: "SHARE_C"}, exp)
			if err == nil {
				codes <- c
			}
		}()
	}
	wg.Wait()
	close(codes)

	seen := map[string]bool{}
	for c := range codes {
		assert.False(t, seen[c], "并发冲突: %s", c)
		seen[c] = true
	}
}

// TestRetrieve_HappyPath 正常取件
func TestRetrieve_HappyPath(t *testing.T) {
	svc, _, _ := newTestService(t)
	ctx := context.Background()
	require.NoError(t, svc.fileCodeRepo.Create(ctx, &model.FileCode{Code: "SHARE1", ExpiredCount: 3}))

	code, err := svc.GenerateCode(ctx, CodeMeta{
		ShareCode: "SHARE1", FileName: "f.txt", FileSize: 100, ContentType: "text/plain",
	}, time.Now().Add(time.Hour))
	require.NoError(t, err)
	assert.Len(t, code, 6)

	meta, err := svc.Retrieve(ctx, code, "")
	require.NoError(t, err)
	assert.Equal(t, "SHARE1", meta.ShareCode)
	assert.Equal(t, "f.txt", meta.FileName)
	assert.Equal(t, int64(100), meta.FileSize)

	// DB 次数已扣减
	fc, _ := svc.fileCodeRepo.GetByCode(ctx, "SHARE1")
	assert.Equal(t, 2, fc.ExpiredCount)
	assert.Equal(t, 1, fc.UsedCount)
}

// TestRetrieve_NotFound 取件码不存在
func TestRetrieve_NotFound(t *testing.T) {
	svc, _, _ := newTestService(t)
	_, err := svc.Retrieve(context.Background(), "NOEXIST", "")
	assert.ErrorIs(t, err, ErrCodeNotFound)
}

// TestRetrieve_Expired 时间过期
func TestRetrieve_Expired(t *testing.T) {
	svc, _, _ := newTestService(t)
	ctx := context.Background()
	past := time.Now().Add(-time.Hour)
	require.NoError(t, svc.fileCodeRepo.Create(ctx, &model.FileCode{Code: "SHARE_EXP", ExpiredAt: &past, ExpiredCount: -1}))
	code, err := svc.GenerateCode(ctx, CodeMeta{ShareCode: "SHARE_EXP"}, time.Now().Add(time.Hour))
	require.NoError(t, err)

	_, err = svc.Retrieve(ctx, code, "")
	assert.ErrorIs(t, err, ErrCodeExpired)
}

// TestRetrieve_Exhausted 次数耗尽
func TestRetrieve_Exhausted(t *testing.T) {
	svc, _, _ := newTestService(t)
	ctx := context.Background()
	require.NoError(t, svc.fileCodeRepo.Create(ctx, &model.FileCode{Code: "SHARE_EXH", ExpiredCount: 1}))
	code, err := svc.GenerateCode(ctx, CodeMeta{ShareCode: "SHARE_EXH"}, time.Now().Add(time.Hour))
	require.NoError(t, err)

	// 第一次成功（1→0）
	_, err = svc.Retrieve(ctx, code, "")
	require.NoError(t, err)
	// 第二次：ExpiredCount=0 → IsExpired=true
	_, err = svc.Retrieve(ctx, code, "")
	assert.ErrorIs(t, err, ErrCodeExpired)
}

// TestRetrieve_PasswordWrong 密码错误
func TestRetrieve_PasswordWrong(t *testing.T) {
	svc, _, _ := newTestService(t)
	ctx := context.Background()
	hash, err := utils.HashPassword("right")
	require.NoError(t, err)
	require.NoError(t, svc.fileCodeRepo.Create(ctx, &model.FileCode{Code: "SHARE_PW", ExpiredCount: -1, RequireAuth: true, PasswordHash: hash}))
	code, err := svc.GenerateCode(ctx, CodeMeta{ShareCode: "SHARE_PW", RequireAuth: true}, time.Now().Add(time.Hour))
	require.NoError(t, err)

	// 错误密码
	_, err = svc.Retrieve(ctx, code, "wrong")
	assert.ErrorIs(t, err, ErrPasswordWrong)

	// 正确密码
	meta, err := svc.Retrieve(ctx, code, "right")
	require.NoError(t, err)
	assert.Equal(t, "SHARE_PW", meta.ShareCode)
}

// TestRetrieve_NoPasswordButRequired 需要密码但未传
func TestRetrieve_NoPasswordButRequired(t *testing.T) {
	svc, _, _ := newTestService(t)
	ctx := context.Background()
	hash, _ := utils.HashPassword("secret")
	require.NoError(t, svc.fileCodeRepo.Create(ctx, &model.FileCode{Code: "SHARE_NP", ExpiredCount: -1, RequireAuth: true, PasswordHash: hash}))
	code, _ := svc.GenerateCode(ctx, CodeMeta{ShareCode: "SHARE_NP", RequireAuth: true}, time.Now().Add(time.Hour))

	// 空密码 → CheckPassword("", "") 对 bcrypt hash 返回 false
	_, err := svc.Retrieve(ctx, code, "")
	assert.ErrorIs(t, err, ErrPasswordWrong)
}

// TestRetrieve_Unlimited 无限次数可反复取
func TestRetrieve_Unlimited(t *testing.T) {
	svc, _, _ := newTestService(t)
	ctx := context.Background()
	require.NoError(t, svc.fileCodeRepo.Create(ctx, &model.FileCode{Code: "SHARE_INF", ExpiredCount: -1}))
	code, _ := svc.GenerateCode(ctx, CodeMeta{ShareCode: "SHARE_INF"}, time.Now().Add(time.Hour))

	for i := 0; i < 5; i++ {
		_, err := svc.Retrieve(ctx, code, "")
		require.NoError(t, err, "第 %d 次取件应成功", i+1)
	}
	fc, _ := svc.fileCodeRepo.GetByCode(ctx, "SHARE_INF")
	assert.Equal(t, -1, fc.ExpiredCount)
	assert.Equal(t, 5, fc.UsedCount)
}

// TestCancel 作废取件码
func TestCancel(t *testing.T) {
	svc, _, _ := newTestService(t)
	ctx := context.Background()
	require.NoError(t, svc.fileCodeRepo.Create(ctx, &model.FileCode{Code: "SHARE_CAN", ExpiredCount: -1}))
	code, _ := svc.GenerateCode(ctx, CodeMeta{ShareCode: "SHARE_CAN"}, time.Now().Add(time.Hour))

	require.NoError(t, svc.Cancel(ctx, code))
	_, err := svc.Retrieve(ctx, code, "")
	assert.ErrorIs(t, err, ErrCodeNotFound)
}

// TestGenerateCode_ExpiredAtPast expireAt 已过期 → 报错
func TestGenerateCode_ExpiredAtPast(t *testing.T) {
	svc, _, _ := newTestService(t)
	_, err := svc.GenerateCode(context.Background(), CodeMeta{ShareCode: "X"}, time.Now().Add(-time.Hour))
	assert.Error(t, err)
}
