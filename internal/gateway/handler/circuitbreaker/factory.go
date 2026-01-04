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

// CreateHandler 创建熔断处理器
// config: 熔断配置
// 返回: CircuitBreakerHandler 实例或错误
func (f *circuitBreakerFactory) CreateHandler(config *CircuitBreakerConfig) (CircuitBreakerHandler, error) {
	if config == nil {
		return nil, fmt.Errorf("熔断配置不能为空")
	}

	// 验证配置
	if err := f.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("熔断配置验证失败: %w", err)
	}

	// 创建熔断器
	return NewCircuitBreaker(config)
}

// ValidateConfig 验证配置
// 验证配置的有效性，如果某些配置无效会设置默认值（不会返回错误）
// 只有严重错误才会返回错误（如 MinimumRequests <= 0）
// 返回值：如果配置验证失败返回错误，否则返回 nil
func (f *circuitBreakerFactory) ValidateConfig(config *CircuitBreakerConfig) error {
	if config == nil {
		return fmt.Errorf("配置不能为空")
	}

	// 验证错误率阈值（0-100之间）
	if config.ErrorRatePercent <= 0 || config.ErrorRatePercent > 100 {
		config.ErrorRatePercent = 50 // 默认50%错误率阈值
	}

	// 验证最小请求数（必须大于0）
	if config.MinimumRequests <= 0 {
		return fmt.Errorf("最小请求数必须大于0，当前值: %d", config.MinimumRequests)
	}

	// 验证熔断器打开持续时间（必须大于0）
	if config.OpenTimeoutSeconds <= 0 {
		config.OpenTimeoutSeconds = 30 // 默认30秒
	}

	// 验证半开状态最大请求数（必须大于0）
	if config.HalfOpenMaxRequests <= 0 {
		config.HalfOpenMaxRequests = 5 // 默认半开状态最大5个请求
	}

	// 设置默认Key策略
	if config.KeyStrategy == "" {
		config.KeyStrategy = "service" // 默认基于服务熔断
	}

	// 设置默认错误状态码
	if config.ErrorStatusCode == 0 {
		config.ErrorStatusCode = 503 // 默认服务不可用
	}

	// 设置默认错误消息
	if config.ErrorMessage == "" {
		config.ErrorMessage = "Service Unavailable - Circuit Breaker Open"
	}

	// 设置默认窗口大小（当前未实现滑动窗口，此配置暂未使用）
	if config.WindowSizeSeconds <= 0 {
		config.WindowSizeSeconds = 60 // 默认60秒窗口
	}

	// 设置默认慢调用阈值（毫秒）
	if config.SlowCallThreshold <= 0 {
		config.SlowCallThreshold = 5000 // 默认5秒
	}

	// 验证慢调用率阈值（0-100之间）
	if config.SlowCallRatePercent <= 0 || config.SlowCallRatePercent > 100 {
		config.SlowCallRatePercent = 80 // 默认80%慢调用阈值
	}

	// 设置默认存储类型
	if config.StorageType == "" {
		config.StorageType = "memory"
	}

	// 初始化存储配置map（如果为nil）
	if config.StorageConfig == nil {
		config.StorageConfig = make(map[string]string)
	}

	return nil
}

// GetSupportedStorageTypes 获取支持的存储类型
// 返回值：支持的存储类型列表
// 注意：当前仅支持 memory，redis 存储未实现
func (f *circuitBreakerFactory) GetSupportedStorageTypes() []string {
	return []string{"memory", "redis"}
}
