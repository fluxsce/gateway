package limiter

import (
	"fmt"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gateway/internal/gateway/core"
	"gateway/internal/gateway/handler/limiter"
)

// TestNewFixedWindowLimiter 测试创建固定窗口限流器
func TestNewFixedWindowLimiter(t *testing.T) {
	tests := []struct {
		name        string
		config      *limiter.RateLimitConfig
		expectError bool
		description string
	}{
		{
			name: "正常创建",
			config: &limiter.RateLimitConfig{
				ID:              "test-fixed-window-1",
				Name:            "测试固定窗口1",
				Enabled:         true,
				Rate:            100,
				WindowSize:      60,
				KeyStrategy:     "ip",
				ErrorStatusCode: 429,
				ErrorMessage:    "Rate limit exceeded",
			},
			expectError: false,
			description: "使用正常配置创建限流器",
		},
		{
			name:        "nil配置",
			config:      nil,
			expectError: false,
			description: "nil配置应该使用默认配置",
		},
		{
			name: "Rate为0使用默认值",
			config: &limiter.RateLimitConfig{
				ID:          "test-fixed-window-2",
				Rate:        0, // 应该使用默认值
				WindowSize:  1,
				KeyStrategy: "ip",
			},
			expectError: false,
			description: "Rate为0时应该使用默认值",
		},
		{
			name: "WindowSize为0使用默认值",
			config: &limiter.RateLimitConfig{
				ID:          "test-fixed-window-3",
				Rate:        100,
				WindowSize:  0, // 应该使用默认值
				KeyStrategy: "ip",
			},
			expectError: false,
			description: "WindowSize为0时应该使用默认值",
		},
		{
			name: "KeyStrategy为空使用默认值",
			config: &limiter.RateLimitConfig{
				ID:          "test-fixed-window-4",
				Rate:        100,
				WindowSize:  1,
				KeyStrategy: "", // 应该使用默认值
			},
			expectError: false,
			description: "KeyStrategy为空时应该使用默认值",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, err := limiter.NewFixedWindowLimiter(tt.config)

			if tt.expectError {
				assert.Error(t, err, "应该返回错误")
				assert.Nil(t, handler, "限流器应该为nil")
			} else {
				assert.NoError(t, err, "不应该返回错误")
				require.NotNil(t, handler, "限流器不应该为nil")

				// 验证算法类型
				assert.Equal(t, limiter.AlgorithmFixedWindow, handler.GetAlgorithm())

				// 验证配置应用了默认值
				config := handler.GetConfig()
				assert.Greater(t, config.Rate, 0, "Rate应该大于0")
				assert.Greater(t, config.WindowSize, 0, "WindowSize应该大于0")
				assert.NotEmpty(t, config.KeyStrategy, "KeyStrategy不应该为空")
			}
		})
	}
}

// TestFixedWindowLimiter_Handle 测试固定窗口限流处理
func TestFixedWindowLimiter_Handle(t *testing.T) {
	t.Run("禁用限流器", func(t *testing.T) {
		config := &limiter.RateLimitConfig{
			Enabled:    false, // 禁用
			Rate:       1,
			WindowSize: 1,
		}
		handler, err := limiter.NewFixedWindowLimiter(config)
		require.NoError(t, err)

		req := httptest.NewRequest("GET", "/test", nil)
		writer := httptest.NewRecorder()
		ctx := core.NewContext(writer, req)

		// 即使速率很小，禁用时也应该通过
		for i := 0; i < 10; i++ {
			result := handler.Handle(ctx)
			assert.True(t, result, "禁用时所有请求都应该通过")
		}
	})

	t.Run("单个请求通过", func(t *testing.T) {
		config := &limiter.RateLimitConfig{
			Enabled:         true,
			Rate:            10,
			WindowSize:      1,
			KeyStrategy:     "ip",
			ErrorStatusCode: 429,
			ErrorMessage:    "Too many requests",
		}
		handler, err := limiter.NewFixedWindowLimiter(config)
		require.NoError(t, err)

		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		writer := httptest.NewRecorder()
		ctx := core.NewContext(writer, req)

		result := handler.Handle(ctx)
		assert.True(t, result, "第一个请求应该通过")

		// 验证上下文设置
		rateLimited, _ := ctx.Get("rate_limited")
		assert.False(t, rateLimited.(bool), "rate_limited应该为false")

		key, _ := ctx.Get("rate_limit_key")
		assert.Contains(t, key.(string), "ip:", "限流键应该包含ip前缀")

		algorithm, _ := ctx.Get("rate_limit_algorithm")
		assert.Equal(t, "fixed-window", algorithm.(string), "算法应该是fixed-window")
	})

	t.Run("窗口内请求达到限制", func(t *testing.T) {
		config := &limiter.RateLimitConfig{
			Enabled:         true,
			Rate:            5, // 只允许5个请求
			WindowSize:      1,
			KeyStrategy:     "ip",
			ErrorStatusCode: 429,
			ErrorMessage:    "Too many requests",
		}
		handler, err := limiter.NewFixedWindowLimiter(config)
		require.NoError(t, err)

		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.2:12345"

		// 前5个请求应该通过
		for i := 1; i <= 5; i++ {
			writer := httptest.NewRecorder()
			ctx := core.NewContext(writer, req)
			result := handler.Handle(ctx)
			assert.True(t, result, "第%d个请求应该通过", i)
		}

		// 第6个请求应该被拒绝
		writer := httptest.NewRecorder()
		ctx := core.NewContext(writer, req)
		result := handler.Handle(ctx)
		assert.False(t, result, "第6个请求应该被拒绝")
	})

	t.Run("窗口过期后重置", func(t *testing.T) {
		config := &limiter.RateLimitConfig{
			Enabled:         true,
			Rate:            3,
			WindowSize:      1, // 1秒窗口
			KeyStrategy:     "ip",
			ErrorStatusCode: 429,
			ErrorMessage:    "Too many requests",
		}
		handler, err := limiter.NewFixedWindowLimiter(config)
		require.NoError(t, err)

		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.3:12345"

		// 前3个请求应该通过
		for i := 1; i <= 3; i++ {
			writer := httptest.NewRecorder()
			ctx := core.NewContext(writer, req)
			result := handler.Handle(ctx)
			assert.True(t, result, "第%d个请求应该通过", i)
		}

		// 第4个请求应该被拒绝
		writer := httptest.NewRecorder()
		ctx := core.NewContext(writer, req)
		result := handler.Handle(ctx)
		assert.False(t, result, "第4个请求应该被拒绝")

		// 等待窗口过期
		time.Sleep(1100 * time.Millisecond)

		// 新窗口的请求应该通过
		for i := 1; i <= 3; i++ {
			writer := httptest.NewRecorder()
			ctx := core.NewContext(writer, req)
			result := handler.Handle(ctx)
			assert.True(t, result, "新窗口第%d个请求应该通过", i)
		}
	})

	t.Run("不同IP独立限流", func(t *testing.T) {
		config := &limiter.RateLimitConfig{
			Enabled:     true,
			Rate:        2,
			WindowSize:  1,
			KeyStrategy: "ip",
		}
		handler, err := limiter.NewFixedWindowLimiter(config)
		require.NoError(t, err)

		// IP1的请求
		req1 := httptest.NewRequest("GET", "/test", nil)
		req1.RemoteAddr = "192.168.1.10:12345"

		// IP2的请求
		req2 := httptest.NewRequest("GET", "/test", nil)
		req2.RemoteAddr = "192.168.1.20:12345"

		// IP1的前2个请求应该通过
		for i := 1; i <= 2; i++ {
			writer := httptest.NewRecorder()
			ctx := core.NewContext(writer, req1)
			result := handler.Handle(ctx)
			assert.True(t, result, "IP1第%d个请求应该通过", i)
		}

		// IP1的第3个请求应该被拒绝
		writer1 := httptest.NewRecorder()
		ctx1 := core.NewContext(writer1, req1)
		result1 := handler.Handle(ctx1)
		assert.False(t, result1, "IP1第3个请求应该被拒绝")

		// IP2的请求应该不受影响
		for i := 1; i <= 2; i++ {
			writer := httptest.NewRecorder()
			ctx := core.NewContext(writer, req2)
			result := handler.Handle(ctx)
			assert.True(t, result, "IP2第%d个请求应该通过", i)
		}
	})

	t.Run("基于路径的限流", func(t *testing.T) {
		config := &limiter.RateLimitConfig{
			Enabled:     true,
			Rate:        3,
			WindowSize:  1,
			KeyStrategy: "path",
		}
		handler, err := limiter.NewFixedWindowLimiter(config)
		require.NoError(t, err)

		// 路径1的请求
		req1 := httptest.NewRequest("GET", "/api/users", nil)
		// 路径2的请求
		req2 := httptest.NewRequest("GET", "/api/products", nil)

		// 路径1的3个请求应该通过
		for i := 1; i <= 3; i++ {
			writer := httptest.NewRecorder()
			ctx := core.NewContext(writer, req1)
			result := handler.Handle(ctx)
			assert.True(t, result, "路径1第%d个请求应该通过", i)
		}

		// 路径1的第4个请求应该被拒绝
		writer1 := httptest.NewRecorder()
		ctx1 := core.NewContext(writer1, req1)
		result1 := handler.Handle(ctx1)
		assert.False(t, result1, "路径1第4个请求应该被拒绝")

		// 路径2的请求不受影响
		for i := 1; i <= 3; i++ {
			writer := httptest.NewRecorder()
			ctx := core.NewContext(writer, req2)
			result := handler.Handle(ctx)
			assert.True(t, result, "路径2第%d个请求应该通过", i)
		}
	})
}

// TestFixedWindowLimiter_Validate 测试配置验证
func TestFixedWindowLimiter_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      *limiter.RateLimitConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "有效配置",
			config: &limiter.RateLimitConfig{
				Rate:       100,
				WindowSize: 60,
			},
			expectError: false,
		},
		{
			name: "Rate为0",
			config: &limiter.RateLimitConfig{
				Rate:       0,
				WindowSize: 60,
			},
			expectError: true,
			errorMsg:    "固定窗口限流速率必须大于0",
		},
		{
			name: "Rate为负数",
			config: &limiter.RateLimitConfig{
				Rate:       -10,
				WindowSize: 60,
			},
			expectError: true,
			errorMsg:    "固定窗口限流速率必须大于0",
		},
		{
			name: "WindowSize为0",
			config: &limiter.RateLimitConfig{
				Rate:       100,
				WindowSize: 0,
			},
			expectError: true,
			errorMsg:    "固定窗口时间窗口大小必须大于0",
		},
		{
			name: "WindowSize为负数",
			config: &limiter.RateLimitConfig{
				Rate:       100,
				WindowSize: -5,
			},
			expectError: true,
			errorMsg:    "固定窗口时间窗口大小必须大于0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 强制设置配置，跳过NewFixedWindowLimiter的默认值应用
			handler := &limiter.FixedWindowLimiter{
				BaseLimiterHandler: limiter.NewBaseLimiterHandler(tt.config),
			}

			err := handler.Validate()

			if tt.expectError {
				assert.Error(t, err, "应该返回验证错误")
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg, "错误消息应该匹配")
				}
			} else {
				assert.NoError(t, err, "不应该返回验证错误")
			}
		})
	}
}

// TestFixedWindowLimiter_OnResponse 测试响应处理
func TestFixedWindowLimiter_OnResponse(t *testing.T) {
	config := &limiter.RateLimitConfig{
		Enabled:    true,
		Rate:       100,
		WindowSize: 1,
	}
	handler, err := limiter.NewFixedWindowLimiter(config)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/test", nil)
	writer := httptest.NewRecorder()
	ctx := core.NewContext(writer, req)

	// OnResponse不应该引发panic或错误
	assert.NotPanics(t, func() {
		handler.OnResponse(ctx, nil)
	}, "OnResponse不应该panic")

	assert.NotPanics(t, func() {
		handler.OnResponse(ctx, assert.AnError)
	}, "OnResponse处理错误时不应该panic")
}

// TestFixedWindowLimiter_Concurrent 测试并发安全性
func TestFixedWindowLimiter_Concurrent(t *testing.T) {
	config := &limiter.RateLimitConfig{
		Enabled:     true,
		Rate:        100,
		WindowSize:  1,
		KeyStrategy: "ip",
	}
	handler, err := limiter.NewFixedWindowLimiter(config)
	require.NoError(t, err)

	const numGoroutines = 50
	const requestsPerGoroutine = 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	successCount := int32(0)
	failCount := int32(0)
	var mu sync.Mutex

	// 并发发送请求
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < requestsPerGoroutine; j++ {
				req := httptest.NewRequest("GET", "/test", nil)
				req.RemoteAddr = "192.168.1.100:12345"
				writer := httptest.NewRecorder()
				ctx := core.NewContext(writer, req)

				result := handler.Handle(ctx)

				mu.Lock()
				if result {
					successCount++
				} else {
					failCount++
				}
				mu.Unlock()

				// 增加一些随机延迟
				time.Sleep(time.Microsecond * time.Duration(goroutineID%10))
			}
		}(i)
	}

	wg.Wait()

	// 验证总请求数
	totalRequests := successCount + failCount
	assert.Equal(t, int32(numGoroutines*requestsPerGoroutine), totalRequests, "总请求数应该匹配")

	// 验证成功请求数不超过速率限制
	assert.LessOrEqual(t, successCount, int32(config.Rate), "成功请求数不应该超过速率限制")

	// 应该有一些请求通过
	assert.Greater(t, successCount, int32(0), "应该有请求通过")

	t.Logf("并发测试结果: 成功=%d, 失败=%d, 总计=%d", successCount, failCount, totalRequests)
}

// TestFixedWindowLimiter_BoundaryConditions 测试边界条件
func TestFixedWindowLimiter_BoundaryConditions(t *testing.T) {
	t.Run("Rate为1", func(t *testing.T) {
		config := &limiter.RateLimitConfig{
			Enabled:    true,
			Rate:       1, // 最小速率
			WindowSize: 1,
		}
		handler, err := limiter.NewFixedWindowLimiter(config)
		require.NoError(t, err)

		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.50:12345"

		// 第1个请求应该通过
		writer1 := httptest.NewRecorder()
		ctx1 := core.NewContext(writer1, req)
		result1 := handler.Handle(ctx1)
		assert.True(t, result1, "第1个请求应该通过")

		// 第2个请求应该被拒绝
		writer2 := httptest.NewRecorder()
		ctx2 := core.NewContext(writer2, req)
		result2 := handler.Handle(ctx2)
		assert.False(t, result2, "第2个请求应该被拒绝")
	})

	t.Run("极大的Rate值", func(t *testing.T) {
		config := &limiter.RateLimitConfig{
			Enabled:    true,
			Rate:       1000000, // 极大的速率
			WindowSize: 1,
		}
		handler, err := limiter.NewFixedWindowLimiter(config)
		require.NoError(t, err)

		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.60:12345"

		// 大量请求都应该通过
		for i := 0; i < 1000; i++ {
			writer := httptest.NewRecorder()
			ctx := core.NewContext(writer, req)
			result := handler.Handle(ctx)
			assert.True(t, result, "第%d个请求应该通过", i+1)
		}
	})

	t.Run("极小的窗口大小", func(t *testing.T) {
		config := &limiter.RateLimitConfig{
			Enabled:    true,
			Rate:       5,
			WindowSize: 1, // 1秒窗口
		}
		handler, err := limiter.NewFixedWindowLimiter(config)
		require.NoError(t, err)

		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.70:12345"

		// 快速发送请求
		for i := 1; i <= 5; i++ {
			writer := httptest.NewRecorder()
			ctx := core.NewContext(writer, req)
			result := handler.Handle(ctx)
			assert.True(t, result, "第%d个请求应该通过", i)
		}

		// 第6个请求应该被拒绝
		writer := httptest.NewRecorder()
		ctx := core.NewContext(writer, req)
		result := handler.Handle(ctx)
		assert.False(t, result, "第6个请求应该被拒绝")
	})
}

// TestFixedWindowLimiter_RaceCondition 测试竞态条件
func TestFixedWindowLimiter_RaceCondition(t *testing.T) {
	config := &limiter.RateLimitConfig{
		Enabled:     true,
		Rate:        50,
		WindowSize:  1,
		KeyStrategy: "ip",
	}
	handler, err := limiter.NewFixedWindowLimiter(config)
	require.NoError(t, err)

	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// 使用相同的IP并发访问
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()

			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "192.168.1.200:12345"
			writer := httptest.NewRecorder()
			ctx := core.NewContext(writer, req)

			// 不关心结果，只测试不会panic或死锁
			handler.Handle(ctx)
		}()
	}

	// 使用超时防止死锁
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// 成功完成
	case <-time.After(5 * time.Second):
		t.Fatal("测试超时，可能存在死锁")
	}
}

// TestFixedWindowLimiter_MultipleKeys 测试多个限流键
func TestFixedWindowLimiter_MultipleKeys(t *testing.T) {
	config := &limiter.RateLimitConfig{
		Enabled:     true,
		Rate:        3,
		WindowSize:  1,
		KeyStrategy: "ip",
	}
	handler, err := limiter.NewFixedWindowLimiter(config)
	require.NoError(t, err)

	// 创建10个不同的IP
	ips := make([]string, 10)
	for i := 0; i < 10; i++ {
		ips[i] = fmt.Sprintf("192.168.1.%d:12345", i+1)
	}

	// 每个IP发送3个请求
	for _, ip := range ips {
		for i := 1; i <= 3; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = ip
			writer := httptest.NewRecorder()
			ctx := core.NewContext(writer, req)

			result := handler.Handle(ctx)
			assert.True(t, result, "IP %s 第%d个请求应该通过", ip, i)
		}
	}

	// 每个IP的第4个请求都应该被拒绝
	for _, ip := range ips {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = ip
		writer := httptest.NewRecorder()
		ctx := core.NewContext(writer, req)

		result := handler.Handle(ctx)
		assert.False(t, result, "IP %s 第4个请求应该被拒绝", ip)
	}
}

// BenchmarkFixedWindowLimiter_Handle 基准测试：Handle方法
func BenchmarkFixedWindowLimiter_Handle(b *testing.B) {
	config := &limiter.RateLimitConfig{
		Enabled:     true,
		Rate:        10000,
		WindowSize:  1,
		KeyStrategy: "ip",
	}
	handler, _ := limiter.NewFixedWindowLimiter(config)

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	writer := httptest.NewRecorder()
	ctx := core.NewContext(writer, req)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.Handle(ctx)
	}
}

// BenchmarkFixedWindowLimiter_HandleParallel 基准测试：并发Handle
func BenchmarkFixedWindowLimiter_HandleParallel(b *testing.B) {
	config := &limiter.RateLimitConfig{
		Enabled:     true,
		Rate:        10000,
		WindowSize:  1,
		KeyStrategy: "ip",
	}
	handler, _ := limiter.NewFixedWindowLimiter(config)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		writer := httptest.NewRecorder()
		ctx := core.NewContext(writer, req)

		for pb.Next() {
			handler.Handle(ctx)
		}
	})
}

// BenchmarkFixedWindowLimiter_MultipleKeys 基准测试：多个限流键
func BenchmarkFixedWindowLimiter_MultipleKeys(b *testing.B) {
	config := &limiter.RateLimitConfig{
		Enabled:     true,
		Rate:        10000,
		WindowSize:  1,
		KeyStrategy: "ip",
	}
	handler, _ := limiter.NewFixedWindowLimiter(config)

	// 创建100个不同的IP
	ips := make([]string, 100)
	for i := 0; i < 100; i++ {
		ips[i] = fmt.Sprintf("192.168.1.%d:12345", i+1)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = ips[i%100]
		writer := httptest.NewRecorder()
		ctx := core.NewContext(writer, req)

		handler.Handle(ctx)
	}
}

// ExampleNewFixedWindowLimiter 示例：创建固定窗口限流器
func ExampleNewFixedWindowLimiter() {
	// 创建配置：每分钟最多100个请求
	config := &limiter.RateLimitConfig{
		ID:              "api-limiter",
		Name:            "API限流器",
		Enabled:         true,
		Rate:            100,  // 每个窗口最多100个请求
		WindowSize:      60,   // 窗口大小60秒
		KeyStrategy:     "ip", // 按IP限流
		ErrorStatusCode: 429,  // Too Many Requests
		ErrorMessage:    "请求过于频繁，请稍后再试",
	}

	// 创建限流器
	handler, err := limiter.NewFixedWindowLimiter(config)
	if err != nil {
		panic(err)
	}

	// 使用限流器
	req := httptest.NewRequest("GET", "/api/users", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	writer := httptest.NewRecorder()
	ctx := core.NewContext(writer, req)

	// 检查是否通过限流
	if handler.Handle(ctx) {
		// 请求通过，继续处理
		println("请求通过")
	} else {
		// 请求被限流
		println("请求被限流")
	}
}
