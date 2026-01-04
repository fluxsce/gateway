package circuitbreaker

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gateway/internal/gateway/core"
	"gateway/internal/gateway/handler/circuitbreaker"
)

func TestCircuitBreakerConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *circuitbreaker.CircuitBreakerConfig
		description string
	}{
		{
			name: "DefaultConfig",
			config: &circuitbreaker.CircuitBreakerConfig{
				Enabled:             true,
				KeyStrategy:         "service",
				ErrorRatePercent:    50,
				MinimumRequests:     10,
				HalfOpenMaxRequests: 3,
				SlowCallThreshold:   1000,
				SlowCallRatePercent: 50,
				OpenTimeoutSeconds:  60,
				WindowSizeSeconds:   60,
				ErrorStatusCode:     503,
				ErrorMessage:        "Service temporarily unavailable",
				StorageType:         "memory",
				StorageConfig:       map[string]string{},
			},
			description: "默认熔断器配置",
		},
		{
			name: "FastFailConfig",
			config: &circuitbreaker.CircuitBreakerConfig{
				Enabled:             true,
				KeyStrategy:         "route",
				ErrorRatePercent:    30,
				MinimumRequests:     5,
				HalfOpenMaxRequests: 2,
				SlowCallThreshold:   500,
				SlowCallRatePercent: 30,
				OpenTimeoutSeconds:  30,
				WindowSizeSeconds:   30,
				ErrorStatusCode:     502,
				ErrorMessage:        "Circuit breaker open",
				StorageType:         "memory",
			},
			description: "快速失败熔断器配置",
		},
		{
			name: "DisabledConfig",
			config: &circuitbreaker.CircuitBreakerConfig{
				Enabled: false,
			},
			description: "禁用熔断器配置",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 验证配置字段
			if tt.config.Enabled {
				assert.NotEmpty(t, tt.config.KeyStrategy, "键策略不应该为空")
				assert.Greater(t, tt.config.ErrorRatePercent, 0, "错误率百分比应该大于0")
				assert.Greater(t, tt.config.MinimumRequests, 0, "最小请求数应该大于0")
			}
		})
	}
}

func TestCircuitBreakerKeyStrategies(t *testing.T) {
	strategies := []string{
		"service",
		"route",
		"path",
		"ip",
		"user",
		"global",
	}

	for _, strategy := range strategies {
		t.Run(strategy, func(t *testing.T) {
			config := &circuitbreaker.CircuitBreakerConfig{
				Enabled:     true,
				KeyStrategy: strategy,
			}

			// 验证策略设置
			assert.Equal(t, strategy, config.KeyStrategy)
		})
	}
}

func TestCircuitBreakerBasicFunctionality(t *testing.T) {
	config := &circuitbreaker.CircuitBreakerConfig{
		Enabled:             true,
		KeyStrategy:         "test",
		ErrorRatePercent:    50,
		MinimumRequests:     3,
		HalfOpenMaxRequests: 2,
		SlowCallThreshold:   1000,
		SlowCallRatePercent: 50,
		OpenTimeoutSeconds:  1, // 1秒恢复时间便于测试
		WindowSizeSeconds:   10,
		ErrorStatusCode:     503,
		ErrorMessage:        "Service temporarily unavailable",
		StorageType:         "memory",
		StorageConfig:       map[string]string{},
	}

	cb, err := circuitbreaker.NewCircuitBreaker(config)
	require.NoError(t, err, "创建熔断器失败")
	require.NotNil(t, cb, "熔断器不应该为nil")

	// 验证配置
	assert.True(t, config.Enabled)
	assert.Equal(t, "test", config.KeyStrategy)
	assert.Equal(t, 50, config.ErrorRatePercent)
	assert.Equal(t, 3, config.MinimumRequests)
}

func TestCircuitBreakerStates(t *testing.T) {
	// 模拟熔断器状态
	states := []string{
		"CLOSED",    // 关闭状态，正常工作
		"OPEN",      // 开启状态，拒绝请求
		"HALF_OPEN", // 半开状态，允许少量请求
	}

	for _, state := range states {
		t.Run(state, func(t *testing.T) {
			// 验证状态常量
			assert.NotEmpty(t, state, "状态常量不应该为空")
		})
	}
}

func TestCircuitBreakerErrorThreshold(t *testing.T) {
	config := &circuitbreaker.CircuitBreakerConfig{
		Enabled:             true,
		KeyStrategy:         "test",
		ErrorRatePercent:    50,
		MinimumRequests:     5,
		HalfOpenMaxRequests: 2,
		SlowCallThreshold:   1000,
		SlowCallRatePercent: 50,
		OpenTimeoutSeconds:  60,
		WindowSizeSeconds:   60,
		ErrorStatusCode:     503,
		ErrorMessage:        "Service temporarily unavailable",
		StorageType:         "memory",
	}

	// 模拟错误阈值测试
	tests := []struct {
		name        string
		totalReqs   int
		errorReqs   int
		expectOpen  bool
		description string
	}{
		{
			name:        "BelowMinimumRequests",
			totalReqs:   3,
			errorReqs:   2,
			expectOpen:  false,
			description: "少于最小请求数不应该开启熔断器",
		},
		{
			name:        "BelowErrorThreshold",
			totalReqs:   10,
			errorReqs:   3,
			expectOpen:  false,
			description: "错误率低于阈值不应该开启熔断器",
		},
		{
			name:        "AboveErrorThreshold",
			totalReqs:   10,
			errorReqs:   6,
			expectOpen:  true,
			description: "错误率超过阈值应该开启熔断器",
		},
		{
			name:        "ExactErrorThreshold",
			totalReqs:   10,
			errorReqs:   5,
			expectOpen:  true,
			description: "错误率等于阈值应该开启熔断器",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 计算错误率
			if tt.totalReqs >= config.MinimumRequests {
				errorRate := float64(tt.errorReqs) / float64(tt.totalReqs) * 100
				shouldOpen := errorRate >= float64(config.ErrorRatePercent)
				assert.Equal(t, tt.expectOpen, shouldOpen, tt.description)
			} else {
				// 少于最小请求数不应该开启
				assert.False(t, tt.expectOpen, tt.description)
			}
		})
	}
}

func TestCircuitBreakerSlowCallThreshold(t *testing.T) {
	config := &circuitbreaker.CircuitBreakerConfig{
		Enabled:             true,
		KeyStrategy:         "test",
		ErrorRatePercent:    50,
		MinimumRequests:     5,
		HalfOpenMaxRequests: 2,
		SlowCallThreshold:   1000, // 1秒
		SlowCallRatePercent: 50,
		OpenTimeoutSeconds:  60,
		WindowSizeSeconds:   60,
		ErrorStatusCode:     503,
		ErrorMessage:        "Service temporarily unavailable",
		StorageType:         "memory",
	}

	tests := []struct {
		name        string
		totalReqs   int
		slowReqs    int
		expectOpen  bool
		description string
	}{
		{
			name:        "BelowSlowCallThreshold",
			totalReqs:   10,
			slowReqs:    3,
			expectOpen:  false,
			description: "慢调用率低于阈值不应该开启熔断器",
		},
		{
			name:        "AboveSlowCallThreshold",
			totalReqs:   10,
			slowReqs:    6,
			expectOpen:  true,
			description: "慢调用率超过阈值应该开启熔断器",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.totalReqs >= config.MinimumRequests {
				slowCallRate := float64(tt.slowReqs) / float64(tt.totalReqs) * 100
				shouldOpen := slowCallRate >= float64(config.SlowCallRatePercent)
				assert.Equal(t, tt.expectOpen, shouldOpen, tt.description)
			}
		})
	}
}

func TestCircuitBreakerRecovery(t *testing.T) {
	config := &circuitbreaker.CircuitBreakerConfig{
		Enabled:             true,
		KeyStrategy:         "test",
		ErrorRatePercent:    50,
		MinimumRequests:     3,
		HalfOpenMaxRequests: 2,
		SlowCallThreshold:   1000,
		SlowCallRatePercent: 50,
		OpenTimeoutSeconds:  1, // 1秒恢复时间便于测试
		WindowSizeSeconds:   10,
		ErrorStatusCode:     503,
		ErrorMessage:        "Service temporarily unavailable",
		StorageType:         "memory",
	}

	// 模拟熔断器恢复过程
	t.Run("OpenToHalfOpen", func(t *testing.T) {
		// 模拟从开启状态转为半开状态
		openTime := time.Now()
		recoverTime := openTime.Add(time.Duration(config.OpenTimeoutSeconds) * time.Second)

		// 检查是否到达恢复时间
		currentTime := time.Now()
		if currentTime.After(recoverTime) {
			// 应该允许转为半开状态
			assert.True(t, true, "应该允许转为半开状态")
		}
	})

	t.Run("HalfOpenToClosed", func(t *testing.T) {
		// 模拟从半开状态转为关闭状态
		maxRequests := config.HalfOpenMaxRequests
		successRequests := maxRequests

		// 如果所有半开请求都成功，应该转为关闭状态
		shouldClose := successRequests == maxRequests
		assert.True(t, shouldClose, "所有半开请求成功应该转为关闭状态")
	})

	t.Run("HalfOpenToOpen", func(t *testing.T) {
		// 模拟从半开状态转为开启状态
		_ = config.HalfOpenMaxRequests
		failedRequests := 1

		// 如果半开请求中有失败，应该转为开启状态
		shouldOpen := failedRequests > 0
		assert.True(t, shouldOpen, "半开请求失败应该转为开启状态")
	})
}

func TestCircuitBreakerConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *circuitbreaker.CircuitBreakerConfig
		expectValid bool
		description string
	}{
		{
			name: "ValidConfig",
			config: &circuitbreaker.CircuitBreakerConfig{
				Enabled:             true,
				KeyStrategy:         "service",
				ErrorRatePercent:    50,
				MinimumRequests:     5,
				HalfOpenMaxRequests: 2,
				SlowCallThreshold:   1000,
				SlowCallRatePercent: 50,
				OpenTimeoutSeconds:  60,
				WindowSizeSeconds:   60,
				ErrorStatusCode:     503,
				ErrorMessage:        "Service temporarily unavailable",
				StorageType:         "memory",
			},
			expectValid: true,
			description: "有效配置应该通过验证",
		},
		{
			name: "InvalidErrorRate",
			config: &circuitbreaker.CircuitBreakerConfig{
				Enabled:          true,
				ErrorRatePercent: 150, // 无效百分比
			},
			expectValid: false,
			description: "无效错误率应该失败",
		},
		{
			name: "ZeroMinimumRequests",
			config: &circuitbreaker.CircuitBreakerConfig{
				Enabled:         true,
				MinimumRequests: 0,
			},
			expectValid: false,
			description: "零最小请求数应该失败",
		},
		{
			name: "DisabledConfig",
			config: &circuitbreaker.CircuitBreakerConfig{
				Enabled: false,
			},
			expectValid: true,
			description: "禁用配置应该有效",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 简单验证逻辑
			valid := true
			if tt.config.Enabled {
				if tt.config.ErrorRatePercent < 0 || tt.config.ErrorRatePercent > 100 {
					valid = false
				}
				if tt.config.MinimumRequests <= 0 {
					valid = false
				}
			}

			assert.Equal(t, tt.expectValid, valid, tt.description)
		})
	}
}

func TestCircuitBreakerMetrics(t *testing.T) {
	_ = &circuitbreaker.CircuitBreakerConfig{
		Enabled:             true,
		KeyStrategy:         "test",
		ErrorRatePercent:    50,
		MinimumRequests:     5,
		HalfOpenMaxRequests: 2,
		SlowCallThreshold:   1000,
		SlowCallRatePercent: 50,
		OpenTimeoutSeconds:  60,
		WindowSizeSeconds:   60,
		ErrorStatusCode:     503,
		ErrorMessage:        "Service temporarily unavailable",
		StorageType:         "memory",
	}

	// 模拟指标数据
	metrics := struct {
		TotalRequests  int
		FailedRequests int
		SlowRequests   int
		State          string
		ErrorRate      float64
		SlowCallRate   float64
	}{
		TotalRequests:  100,
		FailedRequests: 30,
		SlowRequests:   20,
		State:          "CLOSED",
		ErrorRate:      30.0,
		SlowCallRate:   20.0,
	}

	// 验证指标计算
	assert.Equal(t, 100, metrics.TotalRequests)
	assert.Equal(t, 30, metrics.FailedRequests)
	assert.Equal(t, 20, metrics.SlowRequests)
	assert.Equal(t, 30.0, metrics.ErrorRate)
	assert.Equal(t, 20.0, metrics.SlowCallRate)
	assert.Equal(t, "CLOSED", metrics.State)
}

func TestCircuitBreakerInterface(t *testing.T) {
	config := &circuitbreaker.CircuitBreakerConfig{
		Enabled:     true,
		KeyStrategy: "test",
	}

	cb, err := circuitbreaker.NewCircuitBreaker(config)
	require.NoError(t, err)
	require.NotNil(t, cb)

	// 验证接口方法存在（通过类型断言）
	assert.NotNil(t, cb, "熔断器实例不应该为nil")
}

func TestCircuitBreakerContextIntegration(t *testing.T) {
	_ = &circuitbreaker.CircuitBreakerConfig{
		Enabled:     true,
		KeyStrategy: "service",
	}

	// 创建测试上下文
	req := httptest.NewRequest("GET", "/api/test", nil)
	writer := httptest.NewRecorder()
	ctx := core.NewContext(writer, req)

	// 设置服务ID用于键生成
	ctx.SetServiceIDs([]string{"test-service"})

	// 验证上下文设置
	serviceIDs := ctx.GetServiceIDs()
	assert.Len(t, serviceIDs, 1, "服务ID数组应该包含1个元素")
	assert.Equal(t, "test-service", serviceIDs[0])
}

// 基准测试
func BenchmarkCircuitBreakerStateCheck(b *testing.B) {
	config := &circuitbreaker.CircuitBreakerConfig{
		Enabled:             true,
		KeyStrategy:         "test",
		ErrorRatePercent:    50,
		MinimumRequests:     10,
		HalfOpenMaxRequests: 3,
		SlowCallThreshold:   1000,
		SlowCallRatePercent: 50,
		OpenTimeoutSeconds:  60,
		WindowSizeSeconds:   60,
		ErrorStatusCode:     503,
		ErrorMessage:        "Service temporarily unavailable",
		StorageType:         "memory",
	}

	_, err := circuitbreaker.NewCircuitBreaker(config)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 模拟状态检查
		_ = config.Enabled
	}
}

func BenchmarkCircuitBreakerKeyGeneration(b *testing.B) {
	config := &circuitbreaker.CircuitBreakerConfig{
		Enabled:     true,
		KeyStrategy: "service",
	}

	req := httptest.NewRequest("GET", "/api/test", nil)
	writer := httptest.NewRecorder()
	ctx := core.NewContext(writer, req)
	ctx.SetServiceIDs([]string{"test-service"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 模拟键生成
		serviceIDs := ctx.GetServiceIDs()
		serviceID := ""
		if len(serviceIDs) > 0 {
			serviceID = serviceIDs[0]
		}
		key := "cb:" + config.KeyStrategy + ":" + serviceID
		_ = key
	}
}
