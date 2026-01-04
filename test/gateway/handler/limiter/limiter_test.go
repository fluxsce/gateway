package limiter

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gateway/internal/gateway/core"
	"gateway/internal/gateway/handler/limiter"
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
				ctx.SetServiceIDs([]string{"test-service"})
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

// TestConcurrentRateLimitRequest 测试并发限流请求
// 10个线程同时请求指定URL，并打印返回结果
func TestConcurrentRateLimitRequest(t *testing.T) {
	url := "https://localhost:8443/a00webres/assets/cssmin/matrix-login.cssgz"
	concurrency := 10

	// 创建HTTP客户端，跳过TLS证书验证（用于本地测试）
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // 跳过证书验证，仅用于测试
			},
		},
		Timeout: 30 * time.Second,
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	results := make([]RequestResult, 0, concurrency)

	// 并发请求
	startTime := time.Now()
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			result := RequestResult{
				ThreadID:   id,
				StartTime:  time.Now(),
				StatusCode: 0,
				Error:      nil,
			}

			// 发送HTTP请求
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				result.Error = err
				result.EndTime = time.Now()
				mu.Lock()
				results = append(results, result)
				mu.Unlock()
				t.Logf("[线程 %d] 创建请求失败: %v", id, err)
				return
			}

			resp, err := client.Do(req)
			result.EndTime = time.Now()
			result.Duration = result.EndTime.Sub(result.StartTime)

			if err != nil {
				result.Error = err
				mu.Lock()
				results = append(results, result)
				mu.Unlock()
				t.Logf("[线程 %d] 请求失败: %v (耗时: %v)", id, err, result.Duration)
				return
			}
			defer resp.Body.Close()

			result.StatusCode = resp.StatusCode

			// 读取响应体
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				result.Error = err
			} else {
				result.ResponseBody = string(bodyBytes)
				result.BodyLength = len(bodyBytes)
			}

			// 打印结果
			mu.Lock()
			results = append(results, result)
			mu.Unlock()

			// 打印详细信息
			if result.Error != nil {
				t.Logf("[线程 %d] ❌ 失败 - 状态码: %d, 错误: %v, 耗时: %v",
					id, result.StatusCode, result.Error, result.Duration)
			} else {
				t.Logf("[线程 %d] ✅ 成功 - 状态码: %d, 响应长度: %d 字节, 耗时: %v",
					id, result.StatusCode, result.BodyLength, result.Duration)
				if result.BodyLength < 500 { // 只打印较短的响应体
					t.Logf("[线程 %d] 响应内容: %s", id, result.ResponseBody)
				} else {
					t.Logf("[线程 %d] 响应内容前100字符: %s...", id, result.ResponseBody[:100])
				}
			}
		}(i)
	}

	// 等待所有请求完成
	wg.Wait()
	totalDuration := time.Since(startTime)

	// 打印汇总信息
	t.Logf("\n========== 并发请求测试汇总 ==========")
	t.Logf("请求URL: %s", url)
	t.Logf("并发数: %d", concurrency)
	t.Logf("总耗时: %v", totalDuration)
	t.Logf("完成请求数: %d", len(results))

	// 统计结果
	successCount := 0
	failureCount := 0
	statusCodeCount := make(map[int]int)
	totalDurationSum := time.Duration(0)
	minDuration := time.Hour
	maxDuration := time.Duration(0)

	for _, result := range results {
		if result.Error == nil {
			successCount++
			statusCodeCount[result.StatusCode]++
		} else {
			failureCount++
		}
		if result.Duration > 0 {
			totalDurationSum += result.Duration
			if result.Duration < minDuration {
				minDuration = result.Duration
			}
			if result.Duration > maxDuration {
				maxDuration = result.Duration
			}
		}
	}

	t.Logf("成功请求: %d", successCount)
	t.Logf("失败请求: %d", failureCount)
	t.Logf("状态码分布:")
	for statusCode, count := range statusCodeCount {
		t.Logf("  %d: %d 次", statusCode, count)
	}
	if len(results) > 0 {
		avgDuration := totalDurationSum / time.Duration(len(results))
		t.Logf("平均请求耗时: %v", avgDuration)
		t.Logf("最短请求耗时: %v", minDuration)
		t.Logf("最长请求耗时: %v", maxDuration)
	}
	t.Logf("=====================================\n")

	// 验证至少有一个请求成功
	if successCount == 0 {
		t.Errorf("所有请求都失败了，可能服务器未启动或URL不正确")
	}
}

// RequestResult 请求结果结构
type RequestResult struct {
	ThreadID     int
	StartTime    time.Time
	EndTime      time.Time
	Duration     time.Duration
	StatusCode   int
	ResponseBody string
	BodyLength   int
	Error        error
}

// String 实现Stringer接口，用于打印结果
func (r RequestResult) String() string {
	if r.Error != nil {
		return fmt.Sprintf("线程[%d] 失败: %v (耗时: %v)", r.ThreadID, r.Error, r.Duration)
	}
	return fmt.Sprintf("线程[%d] 成功: 状态码=%d, 响应长度=%d字节 (耗时: %v)",
		r.ThreadID, r.StatusCode, r.BodyLength, r.Duration)
}
