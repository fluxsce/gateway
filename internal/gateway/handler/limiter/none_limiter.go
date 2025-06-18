package limiter

import (
	"gohub/internal/gateway/core"
)

// NoneLimiter 无限制限流器
// 不执行任何限流操作，直接通过
type NoneLimiter struct {
	*BaseLimiterHandler
}

// NewNoneLimiter 创建无限制限流器
func NewNoneLimiter(config *RateLimitConfig) (LimiterHandler, error) {
	if config == nil {
		config = &RateLimitConfig{
			ID:        "none-limiter",
			Name:      "None Limiter",
			Enabled:   false,
			Algorithm: AlgorithmNone,
		}
	}

	config.Algorithm = AlgorithmNone

	return &NoneLimiter{
		BaseLimiterHandler: NewBaseLimiterHandler(config),
	}, nil
}

// Handle 处理限流，直接通过
func (n *NoneLimiter) Handle(ctx *core.Context) bool {
	return true
}

// Validate 验证配置
func (n *NoneLimiter) Validate() error {
	return nil
}
