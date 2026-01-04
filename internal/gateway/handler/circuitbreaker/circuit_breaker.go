package circuitbreaker

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"gateway/internal/gateway/core"
)

// circuitBreakerImpl 熔断器实现
// 注意：
// 1. 当前实现中，storage 字段未被使用，状态直接存储在 circuits map 中
// 2. RecordSuccess 和 RecordFailure 方法存在但未被调用，统计信息不会自动更新
// 3. 未实现滑动窗口统计，当前是累积统计
// 4. shouldTrip 仅检查失败率，未检查慢调用率
type circuitBreakerImpl struct {
	config       *CircuitBreakerConfig          // 熔断配置
	circuits     map[string]*CircuitBreakerInfo // key -> 熔断器状态（当前使用此存储）
	mu           sync.RWMutex                   // 读写锁，保护 circuits map 和 config
	keyGenerator CircuitBreakerKeyGenerator     // Key生成器
	storage      CircuitBreakerStorage          // 存储接口（当前未使用）
	listeners    []CircuitBreakerListener       // 状态变更监听器列表
}

// NewCircuitBreaker 创建熔断器
// config: 熔断配置，不能为 nil
// 返回: CircuitBreakerHandler 实例
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
// 在请求处理前调用，检查是否允许请求通过
// 返回值：
//   - true: 允许请求通过，需要在请求完成后调用 RecordSuccess 或 RecordFailure 记录结果
//   - false: 拒绝请求，熔断器已打开，请求应直接返回错误
//
// 注意：当前实现中，RecordSuccess 和 RecordFailure 未被调用，统计信息不会更新
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

		// 通知监听器（在锁外调用，避免死锁）
		cb.notifyCallRejected(key, cb.GetState(key))
		return false
	}

	// 记录请求开始时间和key，用于后续记录统计信息
	ctx.Set("circuit_breaker_key", key)
	ctx.Set("circuit_breaker_start_time", time.Now())

	return true
}

// GetConfig 获取熔断配置
func (cb *circuitBreakerImpl) GetConfig() *CircuitBreakerConfig {
	return cb.config
}

// UpdateConfig 更新熔断配置
// 注意：配置更新后，已存在的熔断器状态不会重置
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
// 返回所有熔断器的汇总统计信息
// 注意：此方法会汇总所有key的统计信息，而不是单个key的信息
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

// Reset 重置所有熔断器状态
// 清除所有已创建的熔断器状态
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

// GetState 获取指定key的熔断器状态
// key: 熔断器key
// 返回: 熔断器状态，如果key不存在则返回 StateClosed
func (cb *circuitBreakerImpl) GetState(key string) CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	if circuit, exists := cb.circuits[key]; exists {
		return circuit.State
	}
	return StateClosed
}

// ForceOpen 强制打开熔断器（用于手动触发熔断）
// key: 熔断器key
func (cb *circuitBreakerImpl) ForceOpen(key string) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	circuit := cb.getOrCreateCircuit(key)
	oldState := circuit.State
	circuit.State = StateOpen
	circuit.OpenTime = time.Now().Unix()

	cb.notifyStateChangeSafe(key, oldState, StateOpen, circuit)
	return nil
}

// ForceClose 强制关闭熔断器（用于手动恢复服务）
// key: 熔断器key
func (cb *circuitBreakerImpl) ForceClose(key string) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	circuit := cb.getOrCreateCircuit(key)
	oldState := circuit.State
	circuit.State = StateClosed

	cb.notifyStateChangeSafe(key, oldState, StateClosed, circuit)
	return nil
}

// allowRequest 检查是否允许请求通过
// key: 熔断器key
// 返回值：
//   - true: 允许请求通过
//   - false: 拒绝请求（熔断器已打开）
//
// 状态转换逻辑：
//  1. Closed -> Open: 当失败率达到阈值时（在 RecordFailure 中触发）
//  2. Open -> HalfOpen: 当 OpenTimeoutSeconds 时间后（在此方法中触发）
//  3. HalfOpen -> Closed: 当半开状态下成功请求达到阈值时（在 RecordSuccess 中触发）
//  4. HalfOpen -> Open: 当半开状态下失败时（在 RecordFailure 中触发）
func (cb *circuitBreakerImpl) allowRequest(key string) bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	circuit := cb.getOrCreateCircuit(key)
	now := time.Now()

	switch circuit.State {
	case StateOpen:
		// 检查是否可以转为半开状态（经过 OpenTimeoutSeconds 后）
		if now.Unix()-circuit.OpenTime >= cb.config.OpenTimeoutSeconds {
			circuit.State = StateHalfOpen
			circuit.HalfOpenCount = 0 // 重置半开状态计数
			cb.notifyStateChangeSafe(key, StateOpen, StateHalfOpen, circuit)
			return true
		}
		return false
	case StateHalfOpen:
		// 半开状态下，只允许有限的请求（用于检测服务是否恢复）
		return circuit.HalfOpenCount < int64(cb.config.HalfOpenMaxRequests)
	default:
		// Closed 状态，允许所有请求
		return true
	}
}

// getOrCreateCircuit 获取或创建熔断器
// key: 熔断器key
// 注意：此方法必须在已持有写锁的情况下调用
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
// key: 熔断器key
// responseTime: 响应时间（毫秒）
// 注意：此方法当前未被调用，需要在请求成功完成后手动调用
// 状态转换：
//   - 如果处于 HalfOpen 状态，增加 HalfOpenCount
//   - 如果 HalfOpenCount 达到阈值，转为 Closed 状态并重置统计
func (cb *circuitBreakerImpl) RecordSuccess(key string, responseTime int64) {
	cb.mu.Lock()

	circuit := cb.getOrCreateCircuit(key)
	circuit.TotalRequests++
	circuit.SuccessRequests++
	circuit.LastRequestTime = time.Now().Unix()

	// 检查是否为慢调用（响应时间超过阈值）
	if responseTime > cb.config.SlowCallThreshold {
		circuit.SlowRequests++
	}

	// 半开状态下的成功处理
	if circuit.State == StateHalfOpen {
		circuit.HalfOpenCount++
		// 如果半开状态下的成功请求达到阈值，转为关闭状态
		if circuit.HalfOpenCount >= int64(cb.config.HalfOpenMaxRequests) {
			oldState := circuit.State
			circuit.State = StateClosed
			// 重置统计信息，重新开始统计
			circuit.TotalRequests = 0
			circuit.SuccessRequests = 0
			circuit.FailureRequests = 0
			circuit.SlowRequests = 0
			cb.notifyStateChangeSafe(key, oldState, StateClosed, circuit)
		}
	}

	// 复制监听器列表，在锁外通知
	listeners := make([]CircuitBreakerListener, len(cb.listeners))
	copy(listeners, cb.listeners)
	cb.mu.Unlock()

	// 在锁外通知监听器（异步执行，不阻塞）
	for _, listener := range listeners {
		go listener.OnCallSuccess(key, responseTime)
	}
}

// RecordFailure 记录失败调用
// key: 熔断器key
// responseTime: 响应时间（毫秒）
// err: 错误信息
// 注意：此方法当前未被调用，需要在请求失败后手动调用
// 状态转换：
//   - 如果处于 HalfOpen 状态，立即转为 Open 状态
//   - 如果处于 Closed 状态，检查失败率，如果达到阈值则转为 Open 状态
func (cb *circuitBreakerImpl) RecordFailure(key string, responseTime int64, err error) {
	cb.mu.Lock()

	circuit := cb.getOrCreateCircuit(key)
	circuit.TotalRequests++
	circuit.FailureRequests++
	circuit.LastFailureTime = time.Now().Unix()
	circuit.LastRequestTime = time.Now().Unix()

	// 检查是否为慢调用（响应时间超过阈值）
	if responseTime > cb.config.SlowCallThreshold {
		circuit.SlowRequests++
	}

	// 半开状态下的失败：立即转为打开状态
	if circuit.State == StateHalfOpen {
		oldState := circuit.State
		circuit.State = StateOpen
		circuit.OpenTime = time.Now().Unix()
		circuit.HalfOpenCount = 0
		cb.notifyStateChangeSafe(key, oldState, StateOpen, circuit)
	} else if circuit.State == StateClosed {
		// 关闭状态下的失败：检查是否需要开启熔断
		if cb.shouldTrip(circuit) {
			oldState := circuit.State
			circuit.State = StateOpen
			circuit.OpenTime = time.Now().Unix()
			cb.notifyStateChangeSafe(key, oldState, StateOpen, circuit)
		}
	}

	// 复制监听器列表，在锁外通知
	listeners := make([]CircuitBreakerListener, len(cb.listeners))
	copy(listeners, cb.listeners)
	cb.mu.Unlock()

	// 在锁外通知监听器（异步执行，不阻塞）
	for _, listener := range listeners {
		go listener.OnCallFailure(key, responseTime, err)
	}
}

// shouldTrip 检查是否应该触发熔断
// circuit: 熔断器状态信息
// 返回值：
//   - true: 应该触发熔断（转为 Open 状态）
//   - false: 不触发熔断
//
// 判断条件：
//  1. 总请求数必须达到 MinimumRequests
//  2. 失败率必须达到 ErrorRatePercent
//
// 注意：当前实现仅检查失败率，未检查慢调用率（SlowCallRatePercent）
func (cb *circuitBreakerImpl) shouldTrip(circuit *CircuitBreakerInfo) bool {
	// 检查最小请求数（避免在请求量较少时误触发熔断）
	if circuit.TotalRequests < int64(cb.config.MinimumRequests) {
		return false
	}

	// 检查失败率
	failureRate := float64(circuit.FailureRequests) / float64(circuit.TotalRequests) * 100
	if failureRate >= float64(cb.config.ErrorRatePercent) {
		return true
	}

	// TODO: 检查慢调用率
	// slowRate := float64(circuit.SlowRequests) / float64(circuit.TotalRequests) * 100
	// if slowRate >= float64(cb.config.SlowCallRatePercent) {
	// 	return true
	// }

	return false
}

// notifyStateChangeSafe 安全地通知状态变更（避免死锁）
// 此方法复制 circuit 信息后，在锁外调用 notifyStateChange
// 注意：此方法必须在已持有写锁的情况下调用
func (cb *circuitBreakerImpl) notifyStateChangeSafe(key string, from, to CircuitBreakerState, circuit *CircuitBreakerInfo) {
	// 复制 circuit 信息，避免在锁外访问
	infoCopy := *circuit
	listeners := make([]CircuitBreakerListener, len(cb.listeners))
	copy(listeners, cb.listeners)

	// 在锁外通知监听器
	cb.mu.Unlock()
	defer cb.mu.Lock()

	for _, listener := range listeners {
		go listener.OnStateChange(key, from, to, &infoCopy)
	}
}

// notifyCallSuccess 通知调用成功
func (cb *circuitBreakerImpl) notifyCallSuccess(key string, responseTime int64) {
	listeners := make([]CircuitBreakerListener, len(cb.listeners))
	copy(listeners, cb.listeners)
	cb.mu.Unlock()
	defer cb.mu.Lock()

	for _, listener := range listeners {
		go listener.OnCallSuccess(key, responseTime)
	}
}

// notifyCallFailure 通知调用失败
func (cb *circuitBreakerImpl) notifyCallFailure(key string, responseTime int64, err error) {
	listeners := make([]CircuitBreakerListener, len(cb.listeners))
	copy(listeners, cb.listeners)
	cb.mu.Unlock()
	defer cb.mu.Lock()

	for _, listener := range listeners {
		go listener.OnCallFailure(key, responseTime, err)
	}
}

// notifyCallRejected 通知调用被拒绝
func (cb *circuitBreakerImpl) notifyCallRejected(key string, state CircuitBreakerState) {
	cb.mu.RLock()
	listeners := make([]CircuitBreakerListener, len(cb.listeners))
	copy(listeners, cb.listeners)
	cb.mu.RUnlock()

	for _, listener := range listeners {
		go listener.OnCallRejected(key, state)
	}
}

// defaultCircuitBreakerKeyGenerator 默认Key生成器
type defaultCircuitBreakerKeyGenerator struct{}

// GenerateKey 生成熔断key
// ctx: 请求上下文
// strategy: 策略类型（ip, service, api等）
// 返回值：生成的key字符串
func (g *defaultCircuitBreakerKeyGenerator) GenerateKey(ctx *core.Context, strategy string) string {
	switch strategy {
	case "ip":
		// 基于IP的熔断（按客户端IP分组）
		if clientIP := ctx.Request.Header.Get("X-Forwarded-For"); clientIP != "" {
			// 取第一个IP（如果有多层代理，X-Forwarded-For 可能包含多个IP）
			if ips := parseIPList(clientIP); len(ips) > 0 {
				return "cb_ip:" + ips[0]
			}
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
		// 基于服务的熔断（按服务ID分组）
		if serviceID, exists := ctx.GetString("service_id"); exists && serviceID != "" {
			return "cb_service:" + serviceID
		}
		return "cb_service:default"
	case "api":
		// 基于API路径的熔断（按API路径分组）
		return "cb_api:" + ctx.Request.URL.Path
	default:
		return "cb_default"
	}
}

// parseIPList 解析IP列表（从 X-Forwarded-For 等header中）
// X-Forwarded-For 格式：client, proxy1, proxy2（最左边的IP是原始客户端IP）
func parseIPList(ipList string) []string {
	ips := strings.Split(ipList, ",")
	result := make([]string, 0, len(ips))
	for _, ip := range ips {
		ip = strings.TrimSpace(ip)
		if ip != "" {
			result = append(result, ip)
		}
	}
	return result
}

// memoryCircuitBreakerStorage 内存存储实现
// 注意：当前实现中，此存储接口未被使用，状态直接存储在 circuitBreakerImpl.circuits 中
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
// 注意：当前实现为空，内存存储的数据会一直保留
func (s *memoryCircuitBreakerStorage) Cleanup() error {
	// TODO: 实现清理逻辑，删除长期未使用的熔断器状态
	return nil
}

// Close 关闭存储
func (s *memoryCircuitBreakerStorage) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data = nil
	return nil
}
