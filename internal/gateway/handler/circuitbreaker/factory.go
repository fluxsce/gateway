package circuitbreaker

import (
	"fmt"
)

// circuitBreakerFactory 熔断处理器工厂实现
type circuitBreakerFactory struct{}

// NewCircuitBreakerFactory 创建熔断处理器工厂
func NewCircuitBreakerFactory() CircuitBreakerFactory {
	return &circuitBreakerFactory{}
}

// CreateHandler 校验配置后创建熔断处理器，状态经 pkg/cache 持久化。
func (f *circuitBreakerFactory) CreateHandler(config *CircuitBreakerConfig) (CircuitBreakerHandler, error) {
	if config == nil {
		return nil, fmt.Errorf("熔断配置不能为空")
	}

	if err := f.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("熔断配置验证失败: %w", err)
	}

	return NewCircuitBreaker(config)
}

// ValidateConfig 校验熔断配置，非法阈值回退到 DefaultCircuitBreakerConfig 的同名字段。
func (f *circuitBreakerFactory) ValidateConfig(config *CircuitBreakerConfig) error {
	if config == nil {
		return fmt.Errorf("配置不能为空")
	}

	defaults := DefaultCircuitBreakerConfig()

	if config.ErrorRatePercent <= 0 || config.ErrorRatePercent > 100 {
		config.ErrorRatePercent = defaults.ErrorRatePercent
	}
	if config.MinimumRequests <= 0 {
		return fmt.Errorf("最小请求数必须大于0，当前值: %d", config.MinimumRequests)
	}
	if config.OpenTimeoutSeconds <= 0 {
		config.OpenTimeoutSeconds = defaults.OpenTimeoutSeconds
	}
	if config.HalfOpenMaxRequests <= 0 {
		config.HalfOpenMaxRequests = defaults.HalfOpenMaxRequests
	}
	if config.WindowSizeSeconds <= 0 {
		config.WindowSizeSeconds = defaults.WindowSizeSeconds
	}
	if config.SlowCallThreshold <= 0 {
		config.SlowCallThreshold = defaults.SlowCallThreshold
	}
	if config.SlowCallRatePercent <= 0 || config.SlowCallRatePercent > 100 {
		config.SlowCallRatePercent = defaults.SlowCallRatePercent
	}
	if config.ErrorStatusCode == 0 {
		config.ErrorStatusCode = defaults.ErrorStatusCode
	}
	if config.ErrorMessage == "" {
		config.ErrorMessage = defaults.ErrorMessage
	}
	if config.StorageType == "" || (config.StorageType != "memory" && config.StorageType != "redis") {
		config.StorageType = defaults.StorageType
	}
	if config.StorageConfig == nil {
		config.StorageConfig = make(map[string]string)
	}

	return nil
}

// GetSupportedStorageTypes 返回支持的存储类型。
// memory 与 redis 都通过 pkg/cache.Cache 读写，替换 Manager 中的实例即可切换。
func (f *circuitBreakerFactory) GetSupportedStorageTypes() []string {
	return []string{"memory", "redis"}
}
