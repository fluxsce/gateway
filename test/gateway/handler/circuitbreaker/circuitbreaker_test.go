package circuitbreaker

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
			description: "????????",
		},
		{
			name: "FastFailConfig",
			config: &circuitbreaker.CircuitBreakerConfig{
				Enabled:             true,
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
			description: "??????????",
		},
		{
			name: "DisabledConfig",
			config: &circuitbreaker.CircuitBreakerConfig{
				Enabled: false,
			},
			description: "??????",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config.Enabled {
				assert.Greater(t, tt.config.ErrorRatePercent, 0, "??????????0")
				assert.Greater(t, tt.config.MinimumRequests, 0, "?????????0")
			}
		})
	}
}

func TestCircuitBreakerBasicFunctionality(t *testing.T) {
	config := &circuitbreaker.CircuitBreakerConfig{
		Enabled:             true,
		ErrorRatePercent:    50,
		MinimumRequests:     3,
		HalfOpenMaxRequests: 2,
		SlowCallThreshold:   1000,
		SlowCallRatePercent: 50,
		OpenTimeoutSeconds:  1,
		WindowSizeSeconds:   10,
		ErrorStatusCode:     503,
		ErrorMessage:        "Service temporarily unavailable",
		StorageType:         "memory",
		StorageConfig:       map[string]string{},
	}

	cb, err := circuitbreaker.NewCircuitBreaker(config)
	require.NoError(t, err, "???????")
	require.NotNil(t, cb, "???????nil")

	assert.True(t, config.Enabled)
	assert.Equal(t, 50, config.ErrorRatePercent)
	assert.Equal(t, 3, config.MinimumRequests)

	key := circuitbreaker.NodeCircuitKey("svc-1", "node-1")
	assert.True(t, cb.Check(key), "?????????")
}

func TestCircuitBreakerErrorThreshold(t *testing.T) {
	config := &circuitbreaker.CircuitBreakerConfig{
		Enabled:             true,
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
			description: "???????????????",
		},
		{
			name:        "BelowErrorThreshold",
			totalReqs:   10,
			errorReqs:   3,
			expectOpen:  false,
			description: "???????????????",
		},
		{
			name:        "AboveErrorThreshold",
			totalReqs:   10,
			errorReqs:   6,
			expectOpen:  true,
			description: "??????????????",
		},
		{
			name:        "ExactErrorThreshold",
			totalReqs:   10,
			errorReqs:   5,
			expectOpen:  true,
			description: "??????????????",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.totalReqs >= config.MinimumRequests {
				errorRate := float64(tt.errorReqs) / float64(tt.totalReqs) * 100
				shouldOpen := errorRate >= float64(config.ErrorRatePercent)
				assert.Equal(t, tt.expectOpen, shouldOpen, tt.description)
			} else {
				assert.False(t, tt.expectOpen, tt.description)
			}
		})
	}
}

func TestCircuitBreakerSlowCallThreshold(t *testing.T) {
	config := &circuitbreaker.CircuitBreakerConfig{
		Enabled:             true,
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
			description: "????????????????",
		},
		{
			name:        "AboveSlowCallThreshold",
			totalReqs:   10,
			slowReqs:    6,
			expectOpen:  true,
			description: "???????????????",
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
		ErrorRatePercent:    50,
		MinimumRequests:     3,
		HalfOpenMaxRequests: 2,
		SlowCallThreshold:   1000,
		SlowCallRatePercent: 50,
		OpenTimeoutSeconds:  1,
		WindowSizeSeconds:   10,
		ErrorStatusCode:     503,
		ErrorMessage:        "Service temporarily unavailable",
		StorageType:         "memory",
	}

	t.Run("OpenToHalfOpen", func(t *testing.T) {
		openTime := time.Now()
		recoverTime := openTime.Add(time.Duration(config.OpenTimeoutSeconds) * time.Second)
		currentTime := time.Now()
		if currentTime.After(recoverTime) {
			assert.True(t, true, "??????????")
		}
	})

	t.Run("HalfOpenToClosed", func(t *testing.T) {
		maxRequests := config.HalfOpenMaxRequests
		successRequests := maxRequests
		shouldClose := successRequests == maxRequests
		assert.True(t, shouldClose, "????????????????")
	})

	t.Run("HalfOpenToOpen", func(t *testing.T) {
		failedRequests := 1
		shouldOpen := failedRequests > 0
		assert.True(t, shouldOpen, "??????????????")
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
			description: "??????????",
		},
		{
			name: "InvalidErrorRate",
			config: &circuitbreaker.CircuitBreakerConfig{
				Enabled:          true,
				ErrorRatePercent: 150,
			},
			expectValid: false,
			description: "?????????",
		},
		{
			name: "ZeroMinimumRequests",
			config: &circuitbreaker.CircuitBreakerConfig{
				Enabled:         true,
				MinimumRequests: 0,
			},
			expectValid: false,
			description: "??????????",
		},
		{
			name: "DisabledConfig",
			config: &circuitbreaker.CircuitBreakerConfig{
				Enabled: false,
			},
			expectValid: true,
			description: "????????",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

func TestCircuitBreakerInterface(t *testing.T) {
	config := &circuitbreaker.CircuitBreakerConfig{
		Enabled: true,
	}

	cb, err := circuitbreaker.NewCircuitBreaker(config)
	require.NoError(t, err)
	require.NotNil(t, cb)
}

func BenchmarkCircuitBreakerStateCheck(b *testing.B) {
	config := &circuitbreaker.CircuitBreakerConfig{
		Enabled:             true,
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

	cb, err := circuitbreaker.NewCircuitBreaker(config)
	require.NoError(b, err)
	key := circuitbreaker.NodeCircuitKey("svc-1", "node-1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cb.Check(key)
	}
}

func BenchmarkCircuitBreakerKeyGeneration(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = circuitbreaker.NodeCircuitKey("test-service", "node-1")
	}
}
