package limiter

import (
	"fmt"
)

// LimiterFactory 限流器工厂
type LimiterFactory struct{}

// NewLimiterFactory 创建限流器工厂
func NewLimiterFactory() *LimiterFactory {
	return &LimiterFactory{}
}

// CreateLimiter 根据配置创建限流器
func (f *LimiterFactory) CreateLimiter(config *RateLimitConfig) (LimiterHandler, error) {
	if config == nil {
		config = &DefaultRateLimitConfig
	}

	switch config.Algorithm {
	case AlgorithmTokenBucket:
		return NewTokenBucketLimiter(config)
	case AlgorithmFixedWindow:
		return NewFixedWindowLimiter(config)
	case AlgorithmSlidingWindow:
		return NewSlidingWindowLimiter(config)
	case AlgorithmLeakyBucket:
		return NewLeakyBucketLimiter(config)
	case AlgorithmNone:
		return NewNoneLimiter(config)
	default:
		// 默认使用令牌桶算法
		defaultConfig := *config
		defaultConfig.Algorithm = AlgorithmTokenBucket
		return NewTokenBucketLimiter(&defaultConfig)
	}
}

// GetSupportedAlgorithms 获取支持的限流算法列表
func (f *LimiterFactory) GetSupportedAlgorithms() []RateLimitAlgorithm {
	return []RateLimitAlgorithm{
		AlgorithmTokenBucket,
		AlgorithmFixedWindow,
		AlgorithmSlidingWindow,
		AlgorithmLeakyBucket,
		AlgorithmNone,
	}
}

// GetAlgorithmDescription 获取算法描述
func (f *LimiterFactory) GetAlgorithmDescription(algorithm RateLimitAlgorithm) string {
	descriptions := map[RateLimitAlgorithm]string{
		AlgorithmTokenBucket:   "令牌桶算法，支持突发流量",
		AlgorithmFixedWindow:   "固定窗口算法，按时间窗口统计",
		AlgorithmSlidingWindow: "滑动窗口算法，更平滑的限流",
		AlgorithmLeakyBucket:   "漏桶算法，平滑流量输出",
		AlgorithmNone:          "无限制，不进行任何限制",
	}

	if desc, exists := descriptions[algorithm]; exists {
		return desc
	}
	return "未知限流算法"
}

// ValidateConfig 验证配置
func (f *LimiterFactory) ValidateConfig(config *RateLimitConfig) error {
	if config == nil {
		return fmt.Errorf("限流器配置不能为空")
	}

	// 验证算法
	validAlgorithms := f.GetSupportedAlgorithms()
	algorithmValid := false
	for _, algorithm := range validAlgorithms {
		if config.Algorithm == algorithm {
			algorithmValid = true
			break
		}
	}
	if !algorithmValid {
		return fmt.Errorf("不支持的限流算法: %s", config.Algorithm)
	}

	// 验证参数
	if config.Algorithm != AlgorithmNone {
		if config.Rate <= 0 {
			return fmt.Errorf("限流速率必须大于0")
		}
	}

	return nil
}
