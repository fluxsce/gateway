package limiter

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gohub/internal/gateway/core"
	"gohub/internal/gateway/handler/limiter"
)

func TestRateLimitConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *limiter.RateLimitConfig
		description string
	}{
		{
			name: "TokenBucketConfig",
			config: &limiter.RateLimitConfig{
				ID:              "test-token-bucket",
				Name:            "令牌桶限流测试",
				Enabled:         true,
				Algorithm:       limiter.AlgorithmTokenBucket,
				Rate:            100,
				Burst:           50,
				WindowSize:      1,
				KeyStrategy:     "ip",
				ErrorStatusCode: 429,
				ErrorMessage:    "Rate limit exceeded",
			},
			description: "令牌桶算法配置",
		},
		{
			name: "LeakyBucketConfig",
			config: &limiter.RateLimitConfig{
				ID:              "test-leaky-bucket",
				Name:            "漏桶限流测试",
				Enabled:         true,
				Algorithm:       limiter.AlgorithmLeakyBucket,
				Rate:            50,
				Burst:           25,
				WindowSize:      1,
				KeyStrategy:     "user",
				ErrorStatusCode: 429,
				ErrorMessage:    "Rate limit exceeded",
			},
			description: "漏桶算法配置",
		},
		{
			name: "SlidingWindowConfig",
			config: &limiter.RateLimitConfig{
				ID:              "test-sliding-window",
				Name:            "滑动窗口限流测试",
				Enabled:         true,
				Algorithm:       limiter.AlgorithmSlidingWindow,
				Rate:            200,
				Burst:           100,
				WindowSize:      1,
				KeyStrategy:     "path",
				ErrorStatusCode: 429,
				ErrorMessage:    "Rate limit exceeded",
			},
			description: "滑动窗口算法配置",
		},
		{
			name: "FixedWindowConfig",
			config: &limiter.RateLimitConfig{
				ID:              "test-fixed-window",
				Name:            "固定窗口限流测试",
				Enabled:         true,
				Algorithm:       limiter.AlgorithmFixedWindow,
				Rate:            75,
				Burst:           30,
				WindowSize:      1,
				KeyStrategy:     "service",
				ErrorStatusCode: 429,
				ErrorMessage:    "Rate limit exceeded",
			},
			description: "固定窗口算法配置",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 验证配置字段
			assert.NotEmpty(t, tt.config.ID, "ID不应该为空")
			assert.NotEmpty(t, tt.config.Name, "名称不应该为空")
			assert.True(t, tt.config.Enabled, "应该启用限流")
			assert.Greater(t, tt.config.Rate, 0, "速率应该大于0")
			assert.Greater(t, tt.config.Burst, 0, "突发大小应该大于0")
			assert.Greater(t, tt.config.WindowSize, 0, "窗口大小应该大于0")
			assert.Equal(t, 429, tt.config.ErrorStatusCode, "错误状态码应该是429")
		})
	}
}

func TestLimiterAlgorithms(t *testing.T) {
	algorithms := []limiter.RateLimitAlgorithm{
		limiter.AlgorithmTokenBucket,
		limiter.AlgorithmLeakyBucket,
		limiter.AlgorithmSlidingWindow,
		limiter.AlgorithmFixedWindow,
		limiter.AlgorithmNone,
	}

	for _, algorithm := range algorithms {
		t.Run(string(algorithm), func(t *testing.T) {
			// 验证算法常量不为空
			assert.NotEmpty(t, string(algorithm), "算法常量不应该为空")
		})
	}
}

func TestBaseLimiterHandler(t *testing.T) {
	config := &limiter.RateLimitConfig{
		ID:              "test-base-limiter",
		Name:            "基础限流器测试",
		Enabled:         true,
		Algorithm:       limiter.AlgorithmTokenBucket,
		Rate:            100,
		Burst:           50,
		WindowSize:      1,
		KeyStrategy:     "ip",
		ErrorStatusCode: 429,
		ErrorMessage:    "Rate limit exceeded",
	}

	// 创建基础限流器处理器
	handler := limiter.NewBaseLimiterHandler(config)
	require.NotNil(t, handler, "处理器不应该为nil")

	// 验证配置
	assert.Equal(t, limiter.AlgorithmTokenBucket, handler.GetAlgorithm())
	assert.True(t, handler.IsEnabled())
	assert.Equal(t, "基础限流器测试", handler.GetName())
	assert.Equal(t, config, handler.GetConfig())

	// 验证配置
	assert.NoError(t, handler.Validate())

	// 创建测试请求
	req := httptest.NewRequest("GET", "/test", nil)
	writer := httptest.NewRecorder()

	// 创建测试上下文
	ctx := core.NewContext(writer, req)

	// 执行限流检查（基础实现应该总是返回true）
	result := handler.Handle(ctx)
	assert.True(t, result, "基础处理器应该总是允许请求")

	// 测试响应处理
	handler.OnResponse(ctx, nil)
}

func TestKeyExtractors(t *testing.T) {
	tests := []struct {
		name        string
		strategy    string
		setupCtx    func(*core.Context)
		expectedKey string
		description string
	}{
		{
			name:     "IPKey",
			strategy: "ip",
			setupCtx: func(ctx *core.Context) {
				// IP将从RemoteAddr提取
			},
			expectedKey: "ip:",
			description: "基于IP的限流键",
		},
		{
			name:     "UserKey",
			strategy: "user",
			setupCtx: func(ctx *core.Context) {
				ctx.Set("user_id", "test-user")
			},
			expectedKey: "user:test-user",
			description: "基于用户的限流键",
		},
		{
			name:     "PathKey",
			strategy: "path",
			setupCtx: func(ctx *core.Context) {
				// 路径从URL.Path提取
			},
			expectedKey: "path:/test",
			description: "基于路径的限流键",
		},
		{
			name:     "ServiceKey",
			strategy: "service",
			setupCtx: func(ctx *core.Context) {
				ctx.SetServiceID("test-service")
			},
			expectedKey: "service:test-service",
			description: "基于服务的限流键",
		},
		{
			name:     "RouteKey",
			strategy: "route",
			setupCtx: func(ctx *core.Context) {
				ctx.SetRouteID("test-route")
			},
			expectedKey: "route:test-route",
			description: "基于路由的限流键",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试请求
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "127.0.0.1:12345"
			writer := httptest.NewRecorder()

			// 创建测试上下文
			ctx := core.NewContext(writer, req)

			// 设置上下文
			if tt.setupCtx != nil {
				tt.setupCtx(ctx)
			}

			// 获取键提取器
			extractor := limiter.GetKeyExtractor(tt.strategy)
			require.NotNil(t, extractor, "键提取器不应该为nil")

			// 提取键
			key := extractor(ctx)

			// 验证键格式
			assert.Contains(t, key, tt.strategy+":", "限流键应该包含策略前缀")
		})
	}
}

func TestDefaultRateLimitConfig(t *testing.T) {
	defaultConfig := limiter.DefaultRateLimitConfig

	// 验证默认配置
	assert.Equal(t, "default-ratelimit", defaultConfig.ID)
	assert.Equal(t, "Default Rate Limiter", defaultConfig.Name)
	assert.True(t, defaultConfig.Enabled, "默认应该启用限流")
	assert.Equal(t, limiter.AlgorithmTokenBucket, defaultConfig.Algorithm, "默认算法应该是令牌桶")
	assert.Equal(t, 100, defaultConfig.Rate, "默认速率应该是100")
	assert.Equal(t, 50, defaultConfig.Burst, "默认突发应该是50")
	assert.Equal(t, 1, defaultConfig.WindowSize, "默认窗口大小应该是1秒")
	assert.Equal(t, "ip", defaultConfig.KeyStrategy, "默认键策略应该是ip")
	assert.Equal(t, 429, defaultConfig.ErrorStatusCode, "默认错误状态码应该是429")
	assert.Equal(t, "Rate limit exceeded", defaultConfig.ErrorMessage, "默认错误消息")
}

func TestLimiterInterface(t *testing.T) {
	config := &limiter.RateLimitConfig{
		ID:              "test-interface",
		Name:            "接口测试",
		Enabled:         true,
		Algorithm:       limiter.AlgorithmTokenBucket,
		Rate:            100,
		Burst:           50,
		WindowSize:      1,
		KeyStrategy:     "ip",
		ErrorStatusCode: 429,
		ErrorMessage:    "Rate limit exceeded",
	}

	// 创建处理器
	handler := limiter.NewBaseLimiterHandler(config)

	// 验证处理器实现了LimiterHandler接口
	var _ limiter.LimiterHandler = handler

	// 测试接口方法
	assert.Equal(t, limiter.AlgorithmTokenBucket, handler.GetAlgorithm())
	assert.True(t, handler.IsEnabled())
	assert.Equal(t, "接口测试", handler.GetName())
	assert.NoError(t, handler.Validate())
	assert.Equal(t, config, handler.GetConfig())
}

func TestNilConfigHandling(t *testing.T) {
	// 测试nil配置处理
	handler := limiter.NewBaseLimiterHandler(nil)
	require.NotNil(t, handler, "即使配置为nil，处理器也不应该为nil")

	// 应该使用默认配置
	config := handler.GetConfig()
	require.NotNil(t, config, "配置不应该为nil")
	assert.Equal(t, limiter.DefaultRateLimitConfig.ID, config.ID)
}

func TestCustomConfig(t *testing.T) {
	customConfig := map[string]interface{}{
		"redis_host":     "localhost:6379",
		"redis_db":       0,
		"storage_prefix": "custom_limiter:",
	}

	config := &limiter.RateLimitConfig{
		ID:              "test-custom",
		Name:            "自定义配置测试",
		Enabled:         true,
		Algorithm:       limiter.AlgorithmTokenBucket,
		Rate:            100,
		Burst:           50,
		WindowSize:      1,
		KeyStrategy:     "ip",
		ErrorStatusCode: 429,
		ErrorMessage:    "Rate limit exceeded",
		CustomConfig:    customConfig,
	}

	handler := limiter.NewBaseLimiterHandler(config)
	resultConfig := handler.GetConfig()

	// 验证自定义配置
	assert.Equal(t, customConfig, resultConfig.CustomConfig)
	assert.Equal(t, "localhost:6379", resultConfig.CustomConfig["redis_host"])
	assert.Equal(t, 0, resultConfig.CustomConfig["redis_db"])
	assert.Equal(t, "custom_limiter:", resultConfig.CustomConfig["storage_prefix"])
}

// 基准测试
func BenchmarkBaseLimiterHandler(b *testing.B) {
	config := &limiter.RateLimitConfig{
		ID:              "bench-base",
		Name:            "基准测试",
		Enabled:         true,
		Algorithm:       limiter.AlgorithmTokenBucket,
		Rate:            10000,
		Burst:           5000,
		WindowSize:      1,
		KeyStrategy:     "ip",
		ErrorStatusCode: 429,
		ErrorMessage:    "Rate limit exceeded",
	}

	handler := limiter.NewBaseLimiterHandler(config)
	req := httptest.NewRequest("GET", "/test", nil)
	writer := httptest.NewRecorder()
	ctx := core.NewContext(writer, req)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.Handle(ctx)
	}
}

func BenchmarkKeyExtractor(b *testing.B) {
	req := httptest.NewRequest("GET", "/test/path", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	writer := httptest.NewRecorder()
	ctx := core.NewContext(writer, req)
	ctx.Set("user_id", "test-user")

	extractor := limiter.GetKeyExtractor("ip")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		extractor(ctx)
	}
}
