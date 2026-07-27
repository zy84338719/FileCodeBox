package anonymous

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestService(t *testing.T) (*Service, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { rdb.Close() })

	return NewService(rdb), mr
}

// TestCodeAlphabet 验证字符表（去掉易混淆字符 0/O/1/I/L）
func TestCodeAlphabet(t *testing.T) {
	for _, ch := range codeAlphabet {
		c := string(ch)
		// 这些字符不应该出现（0 O 1 I L）
		badChars := []string{"0", "O", "1", "I", "L", "o", "l"}
		for _, bad := range badChars {
			assert.NotEqual(t, bad, c, "字符表含易混淆字符: %s", c)
		}
	}
	assert.Equal(t, codeLength, 6, "code 长度应该是 6")
}

// TestGenerateCode_1000Uniqueness 1000 次生成无碰撞
func TestGenerateCode_1000Uniqueness(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()

	expireAt := time.Now().Add(1 * time.Hour)
	seen := make(map[string]bool)
	const N = 1000
	for i := 0; i < N; i++ {
		meta := CodeMeta{
			ShareCode:      "share_" + string(rune('A'+i%26)),
			FileName:       "f.txt",
			FileSize:       1024,
			ExpireAt:       expireAt,
			MaxPickupCount: 1,
		}
		code, err := svc.GenerateCode(ctx, meta)
		require.NoError(t, err)
		assert.Len(t, code, codeLength, "code 长度不符: %s", code)
		assert.False(t, seen[code], "code 重复: %s（第 %d 次）", code, i)
		seen[code] = true
	}
	assert.Equal(t, N, len(seen))
}

// TestGenerateCode_Concurrent 并发生成 + SETNX 防碰撞
func TestGenerateCode_Concurrent(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()

	expireAt := time.Now().Add(1 * time.Hour)
	const N = 100
	const workers = 10
	results := make(chan string, N)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for i := 0; i < N/workers; i++ {
				meta := CodeMeta{
					ShareCode: "concurrent_test",
					FileName:  "f.txt",
					FileSize:  1024,
					ExpireAt:  expireAt,
				}
				code, err := svc.GenerateCode(ctx, meta)
				if err != nil {
					t.Errorf("GenerateCode err: %v", err)
					return
				}
				results <- code
			}
		}(w)
	}
	wg.Wait()
	close(results)

	seen := make(map[string]bool)
	for code := range results {
		assert.False(t, seen[code], "code 重复: %s", code)
		seen[code] = true
	}
	assert.Equal(t, N, len(seen))
}

// TestRetrieve_HappyPath 校验 code → share 映射
func TestRetrieve_HappyPath(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()
	expireAt := time.Now().Add(1 * time.Hour)
	meta := CodeMeta{
		ShareCode:      "abc123",
		FileName:       "secret.pdf",
		FileSize:       4096,
		ContentType:    "application/pdf",
		ExpireAt:       expireAt,
		MaxPickupCount: 5,
	}
	code, err := svc.GenerateCode(ctx, meta)
	require.NoError(t, err)

	retrieved, err := svc.Retrieve(ctx, code, "")
	require.NoError(t, err)
	assert.Equal(t, "abc123", retrieved.ShareCode)
	assert.Equal(t, "secret.pdf", retrieved.FileName)
	assert.Equal(t, int64(4096), retrieved.FileSize)
	assert.Equal(t, "application/pdf", retrieved.ContentType)
	assert.Equal(t, int32(5), retrieved.MaxPickupCount)
}

// TestRetrieve_NotFound 取件码不存在
func TestRetrieve_NotFound(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()
	_, err := svc.Retrieve(ctx, "NOPE12", "")
	assert.ErrorIs(t, err, ErrCodeNotFound)
}

// TestRetrieve_Expired 过期
func TestRetrieve_Expired(t *testing.T) {
	svc, mr := newTestService(t)
	ctx := context.Background()
	meta := CodeMeta{
		ShareCode: "expired_share",
		FileName:  "old.txt",
		FileSize:  100,
		ExpireAt:  time.Now().Add(1 * time.Hour),
	}
	code, err := svc.GenerateCode(ctx, meta)
	require.NoError(t, err)

	// miniredis 加速时间到 1h 后 → meta 被 Redis 清理
	mr.FastForward(2 * time.Hour)

	_, err = svc.Retrieve(ctx, code, "")
	assert.ErrorIs(t, err, ErrCodeNotFound)
}

// TestRetrieve_Exhausted 超次
func TestRetrieve_Exhausted(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()
	meta := CodeMeta{
		ShareCode:      "limited",
		FileName:       "once.txt",
		FileSize:       1,
		ExpireAt:       time.Now().Add(1 * time.Hour),
		MaxPickupCount: 2, // 限 2 次
	}
	code, err := svc.GenerateCode(ctx, meta)
	require.NoError(t, err)

	// 第 1、2 次 OK
	for i := 0; i < 2; i++ {
		_, err = svc.Retrieve(ctx, code, "")
		require.NoError(t, err)
		require.NoError(t, svc.IncrementCount(ctx, code))
	}

	// 第 3 次 拒绝
	_, err = svc.Retrieve(ctx, code, "")
	assert.ErrorIs(t, err, ErrCodeExhausted)
}

// TestRetrieve_PasswordWrong 密码错
func TestRetrieve_PasswordWrong(t *testing.T) {
	svc, _ := newTestService(t)
	svc.SetPasswordHasher(func(p string) string {
		return "hashed:" + p
	})
	ctx := context.Background()
	meta := CodeMeta{
		ShareCode:    "secret_share",
		FileName:     "protected.zip",
		FileSize:     100,
		ExpireAt:     time.Now().Add(1 * time.Hour),
		PasswordHash: "hashed:correct",
	}
	code, err := svc.GenerateCode(ctx, meta)
	require.NoError(t, err)

	// 密码错
	_, err = svc.Retrieve(ctx, code, "wrong")
	assert.ErrorIs(t, err, ErrPasswordWrong)

	// 密码对
	_, err = svc.Retrieve(ctx, code, "correct")
	require.NoError(t, err)
}

// TestRetrieve_RequiresPassword 但没传
func TestRetrieve_RequiresPasswordButMissing(t *testing.T) {
	svc, _ := newTestService(t)
	svc.SetPasswordHasher(func(p string) string { return "h:" + p })
	ctx := context.Background()
	meta := CodeMeta{
		ShareCode:    "needs_pwd",
		FileName:     "p.txt",
		FileSize:     1,
		ExpireAt:     time.Now().Add(1 * time.Hour),
		PasswordHash: "h:secret",
	}
	code, err := svc.GenerateCode(ctx, meta)
	require.NoError(t, err)

	_, err = svc.Retrieve(ctx, code, "")
	assert.ErrorIs(t, err, ErrPasswordWrong)
}

// TestRemainingCount 剩余次数
func TestRemainingCount(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()
	meta := CodeMeta{
		ShareCode:      "counter",
		FileName:       "c.txt",
		FileSize:       1,
		ExpireAt:       time.Now().Add(1 * time.Hour),
		MaxPickupCount: 5,
	}
	code, err := svc.GenerateCode(ctx, meta)
	require.NoError(t, err)

	// 初始：剩余 5
	rem, used, err := svc.RemainingCount(ctx, code)
	require.NoError(t, err)
	assert.Equal(t, int32(5), rem)
	assert.Equal(t, int32(0), used)

	// 用了 2 次
	require.NoError(t, svc.IncrementCount(ctx, code))
	require.NoError(t, svc.IncrementCount(ctx, code))

	rem, used, err = svc.RemainingCount(ctx, code)
	require.NoError(t, err)
	assert.Equal(t, int32(3), rem)
	assert.Equal(t, int32(2), used)
}

// TestRemainingCount_Unlimited MaxPickupCount=0 表示无限
func TestRemainingCount_Unlimited(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()
	meta := CodeMeta{
		ShareCode: "unlimited",
		FileName:  "u.txt",
		FileSize:  1,
		ExpireAt:  time.Now().Add(1 * time.Hour),
		// MaxPickupCount = 0（无限）
	}
	code, err := svc.GenerateCode(ctx, meta)
	require.NoError(t, err)

	rem, _, err := svc.RemainingCount(ctx, code)
	require.NoError(t, err)
	assert.Equal(t, int32(-1), rem, "无限时 RemainingCount 返回 -1")
}

// TestCancel 取消
func TestCancel(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()
	meta := CodeMeta{
		ShareCode: "to_cancel",
		FileName:  "x.txt",
		FileSize:  1,
		ExpireAt:  time.Now().Add(1 * time.Hour),
	}
	code, err := svc.GenerateCode(ctx, meta)
	require.NoError(t, err)

	// 取消
	require.NoError(t, svc.Cancel(ctx, code))

	// 再取件 → NotFound
	_, err = svc.Retrieve(ctx, code, "")
	assert.ErrorIs(t, err, ErrCodeNotFound)
}

// TestIncrementCount 计数 +1
func TestIncrementCount(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()
	meta := CodeMeta{
		ShareCode:      "inc_test",
		FileName:       "i.txt",
		FileSize:       1,
		ExpireAt:       time.Now().Add(1 * time.Hour),
		MaxPickupCount: 100,
	}
	code, err := svc.GenerateCode(ctx, meta)
	require.NoError(t, err)

	for i := 1; i <= 5; i++ {
		require.NoError(t, svc.IncrementCount(ctx, code))
	}
	rem, used, err := svc.RemainingCount(ctx, code)
	require.NoError(t, err)
	assert.Equal(t, int32(95), rem)
	assert.Equal(t, int32(5), used)
}

// TestDefaultPasswordHash 默认 hash 行为
func TestDefaultPasswordHash(t *testing.T) {
	empty := defaultPasswordHash("")
	assert.Empty(t, empty, "空密码 hash 为空")

	hashed := defaultPasswordHash("password")
	assert.True(t, strings.HasPrefix(hashed, "sha256:"), "默认 hash 应有 sha256: 前缀")
}
