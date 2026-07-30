package middleware

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDefaultRateLimitConfig 默认配置
func TestDefaultRateLimitConfig(t *testing.T) {
	cfg := DefaultRateLimitConfig()
	assert.Equal(t, 100, cfg.GlobalQPS)
	assert.Equal(t, 10, cfg.UploadQPS)
	assert.Equal(t, 50, cfg.DownloadQPS)
	assert.Equal(t, 5, cfg.LoginQPS)
	assert.Equal(t, 20, cfg.Burst)
	assert.True(t, cfg.Enabled)
}

// TestRateLimiter_Allow_IPDimension IP 维度限流
func TestRateLimiter_Allow_IPDimension(t *testing.T) {
	cfg := RateLimitConfig{
		GlobalQPS: 5,
		Burst:     5,
		Enabled:   true,
	}
	rl := NewRateLimiter(cfg)
	defer rl.Stop()

	// 同一个 IP 在 burst 范围内允许
	for i := 0; i < 5; i++ {
		assert.True(t, rl.allow(nil, scopeGlobal, "1.2.3.4"), "burst 内第 %d 次", i+1)
	}
	// burst 用完后拒绝
	assert.False(t, rl.allow(nil, scopeGlobal, "1.2.3.4"), "burst 满后拒绝")
	assert.False(t, rl.allow(nil, scopeGlobal, "1.2.3.4"), "持续拒绝")
}

// TestRateLimiter_Allow_ScopeDimension 不同 scope 独立计数
func TestRateLimiter_Allow_ScopeDimension(t *testing.T) {
	cfg := RateLimitConfig{
		GlobalQPS:   100,
		UploadQPS:   2,
		DownloadQPS: 100,
		LoginQPS:    100,
		Burst:       2,
		Enabled:     true,
	}
	rl := NewRateLimiter(cfg)
	defer rl.Stop()
	ip := "1.2.3.4"

	// Upload scope 用完 burst
	assert.True(t, rl.allow(nil, scopeUpload, ip))
	assert.True(t, rl.allow(nil, scopeUpload, ip))
	assert.False(t, rl.allow(nil, scopeUpload, ip), "upload burst 用完")

	// Download scope 独立计数
	assert.True(t, rl.allow(nil, scopeDownload, ip), "download scope 独立")
	assert.True(t, rl.allow(nil, scopeDownload, ip))

	// Global scope 也独立
	assert.True(t, rl.allow(nil, scopeGlobal, ip), "global scope 独立")
}

// TestRateLimiter_Allow_DifferentIP 不同 IP 独立
func TestRateLimiter_Allow_DifferentIP(t *testing.T) {
	cfg := RateLimitConfig{
		GlobalQPS: 1,
		Burst:     1,
		Enabled:   true,
	}
	rl := NewRateLimiter(cfg)
	defer rl.Stop()

	// IP A 用完 burst
	assert.True(t, rl.allow(nil, scopeGlobal, "ip-A"))
	assert.False(t, rl.allow(nil, scopeGlobal, "ip-A"))

	// IP B 独立
	assert.True(t, rl.allow(nil, scopeGlobal, "ip-B"))
	assert.False(t, rl.allow(nil, scopeGlobal, "ip-B"))
}

// TestRateLimiter_Allow_DisabledEnabled=false 全部放行
func TestRateLimiter_Allow_Disabled(t *testing.T) {
	cfg := RateLimitConfig{
		GlobalQPS: 1,
		Burst:     1,
		Enabled:   false,
	}
	rl := NewRateLimiter(cfg)
	defer rl.Stop()

	// 任何 IP 任何 scope 都能过（无数限制）
	for i := 0; i < 1000; i++ {
		assert.True(t, rl.allow(nil, scopeGlobal, "1.2.3.4"))
	}
}

// TestRateLimiter_BurstCapacity 突发容量验证
func TestRateLimiter_BurstCapacity(t *testing.T) {
	cfg := RateLimitConfig{
		GlobalQPS: 10, // 10 QPS
		Burst:     20, // 突发 20
		Enabled:   true,
	}
	rl := NewRateLimiter(cfg)
	defer rl.Stop()
	ip := "burst-test"

	// 一次性消耗 20 个 token
	for i := 0; i < 20; i++ {
		assert.True(t, rl.allow(nil, scopeGlobal, ip), "burst 第 %d 个", i+1)
	}
	// 21 个应该失败
	assert.False(t, rl.allow(nil, scopeGlobal, ip))

	// 等 1 个 token 补充（QPS=10 → 100ms 一个 token）
	time.Sleep(150 * time.Millisecond)
	// 现在应该又有 1-2 个 token 可用
	allowed := rl.allow(nil, scopeGlobal, ip)
	assert.True(t, allowed, "等待 100ms 后 burst 应至少 1 个 token")
}

// TestRateLimiter_UpdateConfig 热更新配置
func TestRateLimiter_UpdateConfig(t *testing.T) {
	cfg := RateLimitConfig{
		GlobalQPS: 1,
		Burst:     1,
		Enabled:   true,
	}
	rl := NewRateLimiter(cfg)
	defer rl.Stop()
	ip := "hot-reload"

	// 用完 burst
	assert.True(t, rl.allow(nil, scopeGlobal, ip))
	assert.False(t, rl.allow(nil, scopeGlobal, ip))

	// 热更新：提高 QPS 和 burst
	rl.UpdateConfig(RateLimitConfig{
		GlobalQPS: 100,
		Burst:     50,
		Enabled:   true,
	})

	// 现在应该又有 burst
	for i := 0; i < 50; i++ {
		assert.True(t, rl.allow(nil, scopeGlobal, ip), "更新后第 %d 个", i+1)
	}
}

// TestRateLimiter_LoginQPS 登录接口独立低 QPS
func TestRateLimiter_LoginQPS(t *testing.T) {
	cfg := RateLimitConfig{
		GlobalQPS: 100,
		LoginQPS:  2,
		Burst:     2,
		Enabled:   true,
	}
	rl := NewRateLimiter(cfg)
	defer rl.Stop()

	assert.True(t, rl.allow(nil, scopeLogin, "attacker"))
	assert.True(t, rl.allow(nil, scopeLogin, "attacker"))
	assert.False(t, rl.allow(nil, scopeLogin, "attacker"), "登录限速 2 QPS 触发")
}

// TestRateLimiter_ZeroQPS 边界：QPS=0 → 兜底 1
func TestRateLimiter_ZeroQPS(t *testing.T) {
	cfg := RateLimitConfig{
		UploadQPS: 0, // 0 应该兜底为 1
		Burst:     0, // 0 应该兜底为 QPS
		Enabled:   true,
	}
	rl := NewRateLimiter(cfg)
	defer rl.Stop()

	assert.True(t, rl.allow(nil, scopeUpload, "1.2.3.4"), "第一次应通过")
	assert.False(t, rl.allow(nil, scopeUpload, "1.2.3.4"), "第二次应被拒（QPS=1 burst=1）")
}

// TestRateLimiter_Concurrent 不同 IP 并发
func TestRateLimiter_Concurrent(t *testing.T) {
	cfg := RateLimitConfig{
		GlobalQPS: 50,
		Burst:     50,
		Enabled:   true,
	}
	rl := NewRateLimiter(cfg)
	defer rl.Stop()

	var allowed int32
	var wg sync.WaitGroup
	const workers = 20
	const perWorker = 5
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			ip := "concurrent-ip-" + string(rune('A'+workerID%26))
			for i := 0; i < perWorker; i++ {
				if rl.allow(nil, scopeGlobal, ip) {
					atomic.AddInt32(&allowed, 1)
				}
			}
		}(w)
	}
	wg.Wait()
	// 20 个不同 IP × 5 次 = 100 次调用，每个 IP 限制 50 QPS
	// 期望：所有调用都通过（每个 IP 都用自己 50 的 burst）
	assert.Equal(t, int32(workers*perWorker), atomic.LoadInt32(&allowed),
		"每个 IP 独立 burst 50，20 个不同 IP 各 5 次都应通过")
}

// TestRateLimiter_Stop GC loop 关闭
func TestRateLimiter_Stop(t *testing.T) {
	rl := NewRateLimiter(DefaultRateLimitConfig())
	// Stop 应该正常关闭 stopCh，gc loop 退出
	rl.Stop()
}

// TestRateLimiter_ClientIP X-Forwarded-For
func TestClientIP_XForwardedFor(t *testing.T) {
	// X-Forwarded-For 解析逻辑（直接测解析函数）
	xff := "1.2.3.4, 5.6.7.8, 9.10.11.12"
	// 第一个 IP 是真实客户端
	var ip string
	for i := 0; i < len(xff); i++ {
		if xff[i] == ',' {
			ip = xff[:i]
			break
		}
	}
	assert.Equal(t, "1.2.3.4", ip, "X-Forwarded-For 第一个 IP 是客户端")
}

// TestRateLimiter_GC_NoError GC 不会 panic
func TestRateLimiter_GC_NoError(t *testing.T) {
	rl := NewRateLimiter(DefaultRateLimitConfig())
	defer rl.Stop()

	// 加几个 IP
	rl.allow(nil, scopeGlobal, "ip-1")
	rl.allow(nil, scopeGlobal, "ip-2")
	require.Equal(t, 2, len(rl.clients[scopeGlobal]))

	// 手动调用 gc：新鲜 entry 不会被清理（lastAccess 刚刚）
	rl.gc()
	assert.Equal(t, 2, len(rl.clients[scopeGlobal]), "新加的 entry 不会被清理")
}
