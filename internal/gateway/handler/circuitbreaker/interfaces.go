package circuitbreaker

import (
	"gateway/internal/gateway/core"
)

// CircuitBreakerHandler 熔断处理器接口
// 熔断器用于保护后端服务，当服务出现大量失败或慢调用时，会自动熔断请求，防止雪崩效应
// 熔断器有三种状态：
// 1. Closed（关闭）：正常状态，允许所有请求通过
// 2. Open（打开）：熔断状态，拒绝所有请求，直接返回错误
// 3. HalfOpen（半开）：试探状态，允许有限数量的请求通过，用于检测服务是否恢复
type CircuitBreakerHandler interface {
	// Handle 处理熔断逻辑
	// 在请求处理前调用，检查是否允许请求通过
	// 返回值：
	//   - true: 允许请求通过，需要在请求完成后调用 RecordSuccess 或 RecordFailure
	//   - false: 拒绝请求，熔断器已打开，请求应直接返回错误
	Handle(ctx *core.Context) bool

	// GetConfig 获取熔断配置
	GetConfig() *CircuitBreakerConfig

	// UpdateConfig 更新熔断配置
	UpdateConfig(config *CircuitBreakerConfig) error

	// GetInfo 获取熔断器信息和统计
	// 返回所有熔断器的汇总统计信息
	GetInfo() *CircuitBreakerInfo

	// Reset 重置所有熔断器状态
	Reset() error

	// IsEnabled 检查是否启用
	IsEnabled() bool

	// GetState 获取指定key的熔断器状态
	GetState(key string) CircuitBreakerState

	// ForceOpen 强制打开熔断器（用于手动触发熔断）
	ForceOpen(key string) error

	// ForceClose 强制关闭熔断器（用于手动恢复服务）
	ForceClose(key string) error
}

// CircuitBreakerState 熔断器状态
type CircuitBreakerState string

const (
	// StateClosed 关闭状态 - 正常工作，允许所有请求通过
	StateClosed CircuitBreakerState = "closed"

	// StateOpen 开启状态 - 熔断触发，拒绝所有请求，直接返回错误
	StateOpen CircuitBreakerState = "open"

	// StateHalfOpen 半开状态 - 尝试恢复，允许有限数量的请求通过以检测服务是否恢复
	StateHalfOpen CircuitBreakerState = "half_open"
)

// CircuitBreakerConfig 熔断配置
type CircuitBreakerConfig struct {
	// 基础配置
	Enabled     bool   `json:"enabled" yaml:"enabled" mapstructure:"enabled"`                // 是否启用熔断
	KeyStrategy string `json:"key_strategy" yaml:"key_strategy" mapstructure:"key_strategy"` // 熔断Key策略(ip, service, api等)

	// 阈值配置
	ErrorRatePercent    int   `json:"error_rate_percent" yaml:"error_rate_percent" mapstructure:"error_rate_percent"`             // 错误率阈值(百分比)，超过此阈值触发熔断
	MinimumRequests     int   `json:"minimum_requests" yaml:"minimum_requests" mapstructure:"minimum_requests"`                   // 最小请求数，达到此数量后才进行熔断判断
	HalfOpenMaxRequests int   `json:"half_open_max_requests" yaml:"half_open_max_requests" mapstructure:"half_open_max_requests"` // 半开状态最大请求数，用于检测服务是否恢复
	SlowCallThreshold   int64 `json:"slow_call_threshold" yaml:"slow_call_threshold" mapstructure:"slow_call_threshold"`          // 慢调用阈值(毫秒)，超过此时间视为慢调用
	SlowCallRatePercent int   `json:"slow_call_rate_percent" yaml:"slow_call_rate_percent" mapstructure:"slow_call_rate_percent"` // 慢调用率阈值(百分比)，超过此阈值触发熔断

	// 时间配置
	OpenTimeoutSeconds int64 `json:"open_timeout_seconds" yaml:"open_timeout_seconds" mapstructure:"open_timeout_seconds"` // 熔断器打开持续时间(秒)，超过此时间后转为半开状态
	WindowSizeSeconds  int64 `json:"window_size_seconds" yaml:"window_size_seconds" mapstructure:"window_size_seconds"`    // 统计窗口大小(秒)，用于滑动窗口统计（当前未实现）

	// 错误处理配置
	ErrorStatusCode int    `json:"error_status_code" yaml:"error_status_code" mapstructure:"error_status_code"` // 熔断时返回的HTTP状态码
	ErrorMessage    string `json:"error_message" yaml:"error_message" mapstructure:"error_message"`             // 熔断时返回的错误信息

	// 存储配置
	StorageType   string            `json:"storage_type" yaml:"storage_type" mapstructure:"storage_type"`       // 存储类型(memory, redis)，当前仅支持memory
	StorageConfig map[string]string `json:"storage_config" yaml:"storage_config" mapstructure:"storage_config"` // 存储配置
}

// CircuitBreakerInfo 熔断器完整信息(包含状态和统计)
type CircuitBreakerInfo struct {
	// 基本状态
	State CircuitBreakerState `json:"state"` // 当前状态

	// 请求统计（注意：当前实现是累积统计，未实现滑动窗口）
	TotalRequests   int64   `json:"total_requests"`   // 总请求数
	SuccessRequests int64   `json:"success_requests"` // 成功请求数
	FailureRequests int64   `json:"failure_requests"` // 失败请求数
	SlowRequests    int64   `json:"slow_requests"`    // 慢请求数
	FailureRate     float64 `json:"failure_rate"`     // 失败率（百分比）
	SlowRate        float64 `json:"slow_rate"`        // 慢调用率（百分比）

	// 状态计数
	OpenCount     int64 `json:"open_count"`      // 熔断器打开次数
	HalfOpenCount int64 `json:"half_open_count"` // 半开状态次数（用于半开状态下已处理的请求数）

	// 时间信息
	StateChangeTime int64 `json:"state_change_time"` // 状态变更时间（Unix时间戳）
	WindowStart     int64 `json:"window_start"`      // 窗口开始时间（Unix时间戳，当前未使用）
	WindowEnd       int64 `json:"window_end"`        // 窗口结束时间（Unix时间戳，当前未使用）
	LastRequestTime int64 `json:"last_request_time"` // 最后请求时间（Unix时间戳）
	LastFailureTime int64 `json:"last_failure_time"` // 最后失败时间（Unix时间戳）
	NextRetryTime   int64 `json:"next_retry_time"`   // 下次重试时间（Unix时间戳，当前未使用）
	OpenTime        int64 `json:"open_time"`         // 熔断器打开时间（Unix时间戳）
}

// CircuitBreakerFactory 熔断处理器工厂接口
type CircuitBreakerFactory interface {
	// CreateHandler 创建熔断处理器
	CreateHandler(config *CircuitBreakerConfig) (CircuitBreakerHandler, error)

	// ValidateConfig 验证配置
	// 验证配置的有效性，如果配置无效会设置默认值或返回错误
	ValidateConfig(config *CircuitBreakerConfig) error

	// GetSupportedStorageTypes 获取支持的存储类型
	GetSupportedStorageTypes() []string
}

// CircuitBreakerStorage 熔断存储接口
// 用于存储熔断器状态信息，支持内存存储和Redis存储（Redis存储未实现）
// 注意：当前实现中，circuits map 直接存储状态，storage 接口未被使用
type CircuitBreakerStorage interface {
	// GetInfo 获取熔断器完整信息
	GetInfo(key string) (*CircuitBreakerInfo, error)

	// SetInfo 设置熔断器完整信息
	SetInfo(key string, info *CircuitBreakerInfo) error

	// IncrementSuccess 增加成功计数
	IncrementSuccess(key string, responseTime int64) error

	// IncrementFailure 增加失败计数
	IncrementFailure(key string, responseTime int64) error

	// Reset 重置状态
	Reset(key string) error

	// Cleanup 清理过期数据
	Cleanup() error

	// Close 关闭存储
	Close() error
}

// CircuitBreakerKeyGenerator 熔断Key生成器接口
// 用于根据不同的策略生成熔断器的key，支持按IP、服务、API等维度进行熔断
type CircuitBreakerKeyGenerator interface {
	// GenerateKey 生成熔断key
	// strategy: 策略类型（ip, service, api等）
	GenerateKey(ctx *core.Context, strategy string) string
}

// CircuitBreakerListener 熔断器状态变更监听器接口
// 用于监听熔断器的状态变化和请求事件，可用于监控、日志记录等
type CircuitBreakerListener interface {
	// OnStateChange 状态变更时的回调
	// 当熔断器状态从一种状态转换到另一种状态时调用
	OnStateChange(key string, from, to CircuitBreakerState, info *CircuitBreakerInfo)

	// OnCallSuccess 调用成功时的回调
	// 当请求成功完成时调用
	OnCallSuccess(key string, responseTime int64)

	// OnCallFailure 调用失败时的回调
	// 当请求失败时调用
	OnCallFailure(key string, responseTime int64, err error)

	// OnCallRejected 调用被拒绝时的回调
	// 当请求被熔断器拒绝时调用
	OnCallRejected(key string, state CircuitBreakerState)
}

// DefaultCircuitBreakerConfig 默认熔断配置
func DefaultCircuitBreakerConfig() *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		Enabled:             true,
		KeyStrategy:         "api", // 默认按API分组
		ErrorRatePercent:    50,    // 50%错误率
		MinimumRequests:     10,    // 最少10个请求
		HalfOpenMaxRequests: 3,     // 半开状态最多3个请求
		SlowCallThreshold:   1000,  // 1秒慢调用阈值
		SlowCallRatePercent: 50,    // 50%慢调用率
		OpenTimeoutSeconds:  60,    // 熔断1分钟
		WindowSizeSeconds:   60,    // 统计窗口1分钟
		ErrorStatusCode:     503,   // 服务不可用
		ErrorMessage:        "Service temporarily unavailable due to circuit breaker",
		StorageType:         "memory", // 默认内存存储
		StorageConfig:       make(map[string]string),
	}
}
