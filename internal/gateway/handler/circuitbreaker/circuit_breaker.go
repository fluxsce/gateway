package circuitbreaker

import (
	"fmt"
	"net"
	"sync"
	"time"

	"gateway/internal/gateway/core"
)

// circuitBreakerImpl 熔断器实现
type circuitBreakerImpl struct {
	config       *CircuitBreakerConfig
	circuits     map[string]*CircuitBreakerInfo
	mu           sync.RWMutex
	keyGenerator CircuitBreakerKeyGenerator
	storage      CircuitBreakerStorage
	listeners    []CircuitBreakerListener
}

// 使用interfaces.go中定义的CircuitBreakerInfo，无需重复定义

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker(config *CircuitBreakerConfig) (CircuitBreakerHandler, error) {
	if config == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	breaker := &circuitBreakerImpl{
		config:       config,
		circuits:     make(map[string]*CircuitBreakerInfo),
		keyGenerator: &defaultCircuitBreakerKeyGenerator{},
		storage:      &memoryCircuitBreakerStorage{data: make(map[string]*CircuitBreakerInfo)},
		listeners:    make([]CircuitBreakerListener, 0),
	}

	return breaker, nil
}

// Handle 处理熔断逻辑
func (cb *circuitBreakerImpl) Handle(ctx *core.Context) bool {
	if !cb.IsEnabled() {
		return true
	}

	key := cb.keyGenerator.GenerateKey(ctx, cb.config.KeyStrategy)

	// 检查熔断状态
	if !cb.allowRequest(key) {
		// 熔断开启，拒绝请求
		ctx.AddError(fmt.Errorf("circuit breaker is open for key: %s", key))
		ctx.Abort(cb.config.ErrorStatusCode, map[string]string{
			"error": cb.config.ErrorMessage,
		})

		// 通知监听器
		cb.notifyCallRejected(key, cb.GetState(key))
		return false
	}

	// 记录请求开始
	ctx.Set("circuit_breaker_key", key)
	ctx.Set("circuit_breaker_start_time", time.Now())

	return true
}

// GetConfig 获取熔断配置
func (cb *circuitBreakerImpl) GetConfig() *CircuitBreakerConfig {
	return cb.config
}

// UpdateConfig 更新熔断配置
func (cb *circuitBreakerImpl) UpdateConfig(config *CircuitBreakerConfig) error {
	if config == nil {
		return fmt.Errorf("配置不能为空")
	}

	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.config = config
	return nil
}

// GetInfo 获取熔断器信息和统计
func (cb *circuitBreakerImpl) GetInfo() *CircuitBreakerInfo {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	info := &CircuitBreakerInfo{
		State:           StateClosed,
		TotalRequests:   0,
		SuccessRequests: 0,
		FailureRequests: 0,
		SlowRequests:    0,
		OpenCount:       0,
		HalfOpenCount:   0,
		WindowStart:     time.Now().Unix(),
		WindowEnd:       time.Now().Unix(),
	}

	// 汇总所有熔断器的统计
	for _, circuit := range cb.circuits {
		info.TotalRequests += circuit.TotalRequests
		info.SuccessRequests += circuit.SuccessRequests
		info.FailureRequests += circuit.FailureRequests
		info.SlowRequests += circuit.SlowRequests
		info.HalfOpenCount += circuit.HalfOpenCount

		if circuit.State == StateOpen {
			info.OpenCount++
		}
	}

	// 计算率
	if info.TotalRequests > 0 {
		info.FailureRate = float64(info.FailureRequests) / float64(info.TotalRequests) * 100
		info.SlowRate = float64(info.SlowRequests) / float64(info.TotalRequests) * 100
	}

	return info
}

// Reset 重置熔断状态
func (cb *circuitBreakerImpl) Reset() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.circuits = make(map[string]*CircuitBreakerInfo)
	return nil
}

// IsEnabled 检查是否启用
func (cb *circuitBreakerImpl) IsEnabled() bool {
	return cb.config.Enabled
}

// GetState 获取熔断器状态
func (cb *circuitBreakerImpl) GetState(key string) CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	if circuit, exists := cb.circuits[key]; exists {
		return circuit.State
	}
	return StateClosed
}

// ForceOpen 强制打开熔断器
func (cb *circuitBreakerImpl) ForceOpen(key string) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	circuit := cb.getOrCreateCircuit(key)
	oldState := circuit.State
	circuit.State = StateOpen
	circuit.OpenTime = time.Now().Unix()

	cb.notifyStateChange(key, oldState, StateOpen)
	return nil
}

// ForceClose 强制关闭熔断器
func (cb *circuitBreakerImpl) ForceClose(key string) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	circuit := cb.getOrCreateCircuit(key)
	oldState := circuit.State
	circuit.State = StateClosed

	cb.notifyStateChange(key, oldState, StateClosed)
	return nil
}

// allowRequest 检查是否允许请求通过
func (cb *circuitBreakerImpl) allowRequest(key string) bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	circuit := cb.getOrCreateCircuit(key)
	now := time.Now()

	switch circuit.State {
	case StateOpen:
		// 检查是否可以转为半开状态
		if now.Unix()-circuit.OpenTime >= cb.config.OpenTimeoutSeconds {
			circuit.State = StateHalfOpen
			circuit.HalfOpenCount = 0
			cb.notifyStateChange(key, StateOpen, StateHalfOpen)
			return true
		}
		return false
	case StateHalfOpen:
		// 半开状态下，只允许有限的请求
		return circuit.HalfOpenCount < int64(cb.config.HalfOpenMaxRequests)
	default:
		return true
	}
}

// getOrCreateCircuit 获取或创建熔断器
func (cb *circuitBreakerImpl) getOrCreateCircuit(key string) *CircuitBreakerInfo {
	if circuit, exists := cb.circuits[key]; exists {
		return circuit
	}

	circuit := &CircuitBreakerInfo{
		State:       StateClosed,
		WindowStart: time.Now().Unix(),
	}
	cb.circuits[key] = circuit
	return circuit
}

// RecordSuccess 记录成功调用
func (cb *circuitBreakerImpl) RecordSuccess(key string, responseTime int64) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	circuit := cb.getOrCreateCircuit(key)
	circuit.TotalRequests++
	circuit.SuccessRequests++

	// 检查是否为慢调用
	if responseTime > cb.config.SlowCallThreshold {
		circuit.SlowRequests++
	}

	// 半开状态下的成功
	if circuit.State == StateHalfOpen {
		circuit.HalfOpenCount++
		// 如果半开状态下的成功请求达到阈值，转为关闭状态
		if circuit.HalfOpenCount >= int64(cb.config.HalfOpenMaxRequests) {
			oldState := circuit.State
			circuit.State = StateClosed
			circuit.TotalRequests = 0
			circuit.SuccessRequests = 0
			circuit.FailureRequests = 0
			circuit.SlowRequests = 0
			cb.notifyStateChange(key, oldState, StateClosed)
		}
	}

	cb.notifyCallSuccess(key, responseTime)
}

// RecordFailure 记录失败调用
func (cb *circuitBreakerImpl) RecordFailure(key string, responseTime int64, err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	circuit := cb.getOrCreateCircuit(key)
	circuit.TotalRequests++
	circuit.FailureRequests++
	circuit.LastFailureTime = time.Now().Unix()

	// 检查是否为慢调用
	if responseTime > cb.config.SlowCallThreshold {
		circuit.SlowRequests++
	}

	// 检查是否需要开启熔断
	if cb.shouldTrip(circuit) {
		oldState := circuit.State
		circuit.State = StateOpen
		circuit.OpenTime = time.Now().Unix()
		cb.notifyStateChange(key, oldState, StateOpen)
	}

	cb.notifyCallFailure(key, responseTime, err)
}

// shouldTrip 检查是否应该触发熔断
func (cb *circuitBreakerImpl) shouldTrip(circuit *CircuitBreakerInfo) bool {
	// 检查最小请求数
	if circuit.TotalRequests < int64(cb.config.MinimumRequests) {
		return false
	}

	// 检查失败率
	failureRate := float64(circuit.FailureRequests) / float64(circuit.TotalRequests) * 100
	return failureRate >= float64(cb.config.ErrorRatePercent)
}

// 通知方法
func (cb *circuitBreakerImpl) notifyStateChange(key string, from, to CircuitBreakerState) {
	info := cb.GetInfo()
	for _, listener := range cb.listeners {
		go listener.OnStateChange(key, from, to, info)
	}
}

func (cb *circuitBreakerImpl) notifyCallSuccess(key string, responseTime int64) {
	for _, listener := range cb.listeners {
		go listener.OnCallSuccess(key, responseTime)
	}
}

func (cb *circuitBreakerImpl) notifyCallFailure(key string, responseTime int64, err error) {
	for _, listener := range cb.listeners {
		go listener.OnCallFailure(key, responseTime, err)
	}
}

func (cb *circuitBreakerImpl) notifyCallRejected(key string, state CircuitBreakerState) {
	for _, listener := range cb.listeners {
		go listener.OnCallRejected(key, state)
	}
}

// defaultCircuitBreakerKeyGenerator 默认Key生成器
type defaultCircuitBreakerKeyGenerator struct{}

// GenerateKey 生成熔断key
func (g *defaultCircuitBreakerKeyGenerator) GenerateKey(ctx *core.Context, strategy string) string {
	switch strategy {
	case "ip":
		// 基于IP的熔断
		if clientIP := ctx.Request.Header.Get("X-Forwarded-For"); clientIP != "" {
			return "cb_ip:" + clientIP
		}
		if clientIP := ctx.Request.Header.Get("X-Real-IP"); clientIP != "" {
			return "cb_ip:" + clientIP
		}
		if host, _, err := net.SplitHostPort(ctx.Request.RemoteAddr); err == nil {
			return "cb_ip:" + host
		}
		return "cb_ip:" + ctx.Request.RemoteAddr
	case "service":
		// 基于服务的熔断
		if serviceID, exists := ctx.GetString("service_id"); exists && serviceID != "" {
			return "cb_service:" + serviceID
		}
		return "cb_service:default"
	case "api":
		// 基于API路径的熔断
		return "cb_api:" + ctx.Request.URL.Path
	default:
		return "cb_default"
	}
}

// memoryCircuitBreakerStorage 内存存储实现
type memoryCircuitBreakerStorage struct {
	data map[string]*CircuitBreakerInfo
	mu   sync.RWMutex
}

// GetInfo 获取熔断器完整信息
func (s *memoryCircuitBreakerStorage) GetInfo(key string) (*CircuitBreakerInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if info, exists := s.data[key]; exists {
		return info, nil
	}

	return nil, fmt.Errorf("key not found: %s", key)
}

// SetInfo 设置熔断器完整信息
func (s *memoryCircuitBreakerStorage) SetInfo(key string, info *CircuitBreakerInfo) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = info
	return nil
}

// IncrementSuccess 增加成功计数
func (s *memoryCircuitBreakerStorage) IncrementSuccess(key string, responseTime int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	info, exists := s.data[key]
	if !exists {
		info = &CircuitBreakerInfo{
			State:       StateClosed,
			WindowStart: time.Now().Unix(),
		}
		s.data[key] = info
	}

	info.TotalRequests++
	info.SuccessRequests++
	return nil
}

// IncrementFailure 增加失败计数
func (s *memoryCircuitBreakerStorage) IncrementFailure(key string, responseTime int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	info, exists := s.data[key]
	if !exists {
		info = &CircuitBreakerInfo{
			State:       StateClosed,
			WindowStart: time.Now().Unix(),
		}
		s.data[key] = info
	}

	info.TotalRequests++
	info.FailureRequests++
	info.LastFailureTime = time.Now().Unix()
	return nil
}

// Reset 重置状态
func (s *memoryCircuitBreakerStorage) Reset(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, key)
	return nil
}

// Cleanup 清理过期数据
func (s *memoryCircuitBreakerStorage) Cleanup() error {
	// 内存存储目前不需要特殊清理逻辑
	return nil
}

// Close 关闭存储
func (s *memoryCircuitBreakerStorage) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data = nil
	return nil
}
