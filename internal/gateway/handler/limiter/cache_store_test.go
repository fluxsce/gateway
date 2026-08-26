package limiter

import (
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"gateway/internal/gateway/core"
	"gateway/pkg/cache"
	"gateway/pkg/cache/memory"
)

// TestMain 测试用全局 default 缓存；生产路径不会在限流模块里 NewMemoryCache。
func TestMain(m *testing.M) {
	if cache.GetDefaultCache() == nil {
		mem, err := memory.NewMemoryCache(&memory.MemoryConfig{
			Enabled:           true,
			EnableLazyCleanup: true,
			CleanupInterval:   time.Minute,
		})
		if err != nil {
			panic(err)
		}
		if err := cache.AddCache("default", mem); err != nil && cache.GetDefaultCache() == nil {
			panic(err)
		}
	}
	os.Exit(m.Run())
}

// TestFixedWindowLimiterSharesCacheByConfigID 同一限流配置 ID 应共享 cache 计数。
func TestFixedWindowLimiterSharesCacheByConfigID(t *testing.T) {
	cfg := &RateLimitConfig{
		ID:          "share-fw-cache",
		Enabled:     true,
		Rate:        2,
		WindowSize:  60,
		KeyStrategy: "ip",
		Algorithm:   AlgorithmFixedWindow,
	}

	first, err := NewFixedWindowLimiter(cfg)
	if err != nil {
		t.Fatalf("创建第一个限流器失败: %v", err)
	}
	second, err := NewFixedWindowLimiter(cfg)
	if err != nil {
		t.Fatalf("创建第二个限流器失败: %v", err)
	}

	req := httptest.NewRequest("GET", "/share", nil)
	req.RemoteAddr = "10.9.8.7:12345"

	for i := 0; i < 2; i++ {
		ctx := core.NewContext(httptest.NewRecorder(), req)
		if !first.Handle(ctx) {
			t.Fatalf("第%d次请求应通过第一个限流器", i+1)
		}
	}

	ctx := core.NewContext(httptest.NewRecorder(), req)
	if second.Handle(ctx) {
		t.Fatal("共享 cache 后第三个请求应被第二个限流器拒绝")
	}
}

// TestFixedWindowLimiterIsolatesAnonymousInstances 无 ID 的限流器互不抢计数。
func TestFixedWindowLimiterIsolatesAnonymousInstances(t *testing.T) {
	mk := func() LimiterHandler {
		h, err := NewFixedWindowLimiter(&RateLimitConfig{
			Enabled:     true,
			Rate:        1,
			WindowSize:  60,
			KeyStrategy: "ip",
		})
		if err != nil {
			t.Fatalf("创建限流器失败: %v", err)
		}
		return h
	}

	a := mk()
	b := mk()
	req := httptest.NewRequest("GET", "/anon", nil)
	req.RemoteAddr = "10.1.1.1:1"

	if !a.Handle(core.NewContext(httptest.NewRecorder(), req)) {
		t.Fatal("匿名实例 A 第一次应通过")
	}
	if !b.Handle(core.NewContext(httptest.NewRecorder(), req)) {
		t.Fatal("匿名实例 B 不应吃到 A 的计数")
	}
}
