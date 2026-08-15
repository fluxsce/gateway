// Package circuitbreaker 提供节点摘除，与失败重试相互独立。
// 按实例统计失败/慢调用，供负载均衡跳过已开闸节点；全部开闸时回退健康列表。
// 状态经 CircuitBreakerStorage 写入 pkg/cache，默认进程内 memory。
package circuitbreaker

// CircuitBreakerHandler 熔断处理器接口。
// 按上游实例统计失败/慢调用并开闸，供负载均衡跳过该节点。
// 三种状态：Closed 放行、Open 跳过、HalfOpen 有限探测。
// 选节点前 Check(NodeCircuitKey)，转发后 RecordSuccess / RecordFailure。
type CircuitBreakerHandler interface {
	// Check 判断指定节点键是否放行。开闸时只跳过该节点，不 Abort 整请求。
	Check(key string) bool

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

	// RecordSuccess 记录一次成功调用并回写 cache 统计。
	// responseTime 单位为毫秒，超过 SlowCallThreshold 计入慢调用。
	RecordSuccess(key string, responseTime int64)

	// RecordFailure 记录一次失败调用并回写 cache 统计。
	// 半开失败立即开闸；关闭态在失败率或慢调用率达标后开闸。
	RecordFailure(key string, responseTime int64, err error)
}

// CircuitBreakerState 熔断器状态
type CircuitBreakerState string

const (
	// StateClosed 关闭状态 - 正常工作，允许所有请求通过
	StateClosed CircuitBreakerState = "closed"

	// StateOpen 开启状态 - 该节点被摘除，负载均衡跳过它
	StateOpen CircuitBreakerState = "open"

	// StateHalfOpen 半开状态 - 尝试恢复，允许有限数量的请求通过以检测服务是否恢复
	StateHalfOpen CircuitBreakerState = "half_open"
)

// CircuitBreakerConfig 熔断配置
type CircuitBreakerConfig struct {
	// 基础配置
	Enabled bool `json:"enabled" yaml:"enabled" mapstructure:"enabled"` // 是否启用节点熔断

	// 阈值配置
	ErrorRatePercent    int   `json:"error_rate_percent" yaml:"error_rate_percent" mapstructure:"error_rate_percent"`             // 错误率阈值(百分比)，超过此阈值触发熔断
	MinimumRequests     int   `json:"minimum_requests" yaml:"minimum_requests" mapstructure:"minimum_requests"`                   // 最小请求数，达到此数量后才进行熔断判断
	HalfOpenMaxRequests int   `json:"half_open_max_requests" yaml:"half_open_max_requests" mapstructure:"half_open_max_requests"` // 半开状态最大请求数，用于检测服务是否恢复
	SlowCallThreshold   int64 `json:"slow_call_threshold" yaml:"slow_call_threshold" mapstructure:"slow_call_threshold"`          // 慢调用阈值(毫秒)，超过此时间视为慢调用
	SlowCallRatePercent int   `json:"slow_call_rate_percent" yaml:"slow_call_rate_percent" mapstructure:"slow_call_rate_percent"` // 慢调用率阈值(百分比)，超过此阈值触发熔断

	// 时间配置
	OpenTimeoutSeconds int64 `json:"open_timeout_seconds" yaml:"open_timeout_seconds" mapstructure:"open_timeout_seconds"` // 熔断器打开持续时间(秒)，超过此时间后转为半开状态
	WindowSizeSeconds  int64 `json:"window_size_seconds" yaml:"window_size_seconds" mapstructure:"window_size_seconds"`    // 滑动窗口长度(秒)，只统计窗口内的失败/慢调用

	// 错误处理配置
	ErrorStatusCode int    `json:"error_status_code" yaml:"error_status_code" mapstructure:"error_status_code"` // 熔断时返回的HTTP状态码
	ErrorMessage    string `json:"error_message" yaml:"error_message" mapstructure:"error_message"`             // 熔断时返回的错误信息

	// 存储配置。状态经 pkg/cache 读写：memory 与 redis 都实现 Cache，替换 Manager 实例即可切换。
	// StorageType 仅作声明；实际实例由 StorageConfig["cache_name"]、default、circuit_breaker 依次解析。
	StorageType   string            `json:"storage_type" yaml:"storage_type" mapstructure:"storage_type"`       // 存储类型(memory, redis)
	StorageConfig map[string]string `json:"storage_config" yaml:"storage_config" mapstructure:"storage_config"` // cache_name 指定 Manager 中的实例名
}

// CircuitWindowBucket 滑动窗口中的一个时间片统计。
type CircuitWindowBucket struct {
	Start    int64 `json:"start"`    // 桶起始 Unix 秒
	Total    int64 `json:"total"`    // 该片内请求数
	Failures int64 `json:"failures"` // 该片内失败数（连接错误或上游 5xx）
	Slow     int64 `json:"slow"`     // 该片内慢调用数
}

// CircuitBreakerInfo 熔断器完整信息(包含状态和统计)
type CircuitBreakerInfo struct {
	// 基本状态
	State CircuitBreakerState `json:"state"` // 当前状态

	// 请求统计，数值来自 WindowSizeSeconds 内的分桶汇总
	TotalRequests   int64                 `json:"total_requests"`           // 窗口内总请求数
	SuccessRequests int64                 `json:"success_requests"`         // 窗口内成功请求数
	FailureRequests int64                 `json:"failure_requests"`         // 窗口内失败请求数
	SlowRequests    int64                 `json:"slow_requests"`            // 窗口内慢请求数
	FailureRate     float64               `json:"failure_rate"`             // 窗口内失败率（百分比）
	SlowRate        float64               `json:"slow_rate"`                // 窗口内慢调用率（百分比）
	WindowBuckets   []CircuitWindowBucket `json:"window_buckets,omitempty"` // 时间分桶，过期桶在写入时丢弃

	// 状态计数
	OpenCount       int64 `json:"open_count"`        // 熔断器打开次数
	HalfOpenCount   int64 `json:"half_open_count"`   // 半开已占用的探测名额（Check 时占坑）
	HalfOpenSuccess int64 `json:"half_open_success"` // 半开探测已成功次数，达到配额后关闸

	// 时间信息
	StateChangeTime int64 `json:"state_change_time"` // 状态变更时间（Unix时间戳）
	WindowStart     int64 `json:"window_start"`      // 当前窗口最早桶的起始时间（Unix时间戳）
	WindowEnd       int64 `json:"window_end"`        // 当前窗口最晚桶的起始时间（Unix时间戳）
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

// CircuitBreakerStorage 熔断状态存储。
// 实现应走 pkg/cache，便于后期把 memory 实例换成 redis。
// 状态按节点键保存，热重载复用同一 cache，不因代际切换清零。
// 节点 ID 唯一生成，不会撞键。
type CircuitBreakerStorage interface {
	// GetInfo 获取熔断器完整信息；键不存在时返回 nil, nil。
	GetInfo(key string) (*CircuitBreakerInfo, error)

	// SetInfo 设置熔断器完整信息；info 为 nil 时删除该键。
	SetInfo(key string, info *CircuitBreakerInfo) error

	// IncrementSuccess 增加成功计数并回写。熔断状态机以 RecordSuccess 为准，本方法供扩展。
	IncrementSuccess(key string, responseTime int64) error

	// IncrementFailure 增加失败计数并回写。熔断状态机以 RecordFailure 为准，本方法供扩展。
	IncrementFailure(key string, responseTime int64) error

	// Reset 删除指定 key 的状态
	Reset(key string) error

	// Cleanup 清理全部熔断缓存键
	Cleanup() error

	// Close 关闭存储。共享 Cache 实现不应在此关闭 Manager 中的实例。
	Close() error
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

// DefaultCircuitBreakerConfig 返回默认节点熔断配置。
// 50% 错误率、最少 10 次、开闸 60 秒、慢调用 60 秒；存储声明为 memory，实际仍经 pkg/cache 解析。
func DefaultCircuitBreakerConfig() *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		Enabled:             true,
		ErrorRatePercent:    50,    // 50%错误率
		MinimumRequests:     10,    // 最少10个请求
		HalfOpenMaxRequests: 3,     // 半开状态最多3个请求
		SlowCallThreshold:   60000, // 60秒，按可能较长的上游处理时间，避免把正常慢接口当故障
		SlowCallRatePercent: 50,    // 50%慢调用率
		OpenTimeoutSeconds:  60,    // 熔断1分钟
		WindowSizeSeconds:   60,    // 统计窗口1分钟
		ErrorStatusCode:     503,   // 服务不可用
		ErrorMessage:        "Service temporarily unavailable due to circuit breaker",
		StorageType:         "memory", // 默认内存存储
		StorageConfig:       make(map[string]string),
	}
}
