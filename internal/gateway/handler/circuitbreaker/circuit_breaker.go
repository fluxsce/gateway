package circuitbreaker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// circuitBreakerImpl 熔断器实现。
// 状态经 CircuitBreakerStorage（pkg/cache）持久化，进程内用 mu 串行读写，避免并发覆盖。
type circuitBreakerImpl struct {
	config    *CircuitBreakerConfig    // 当前熔断阈值与存储声明
	mu        sync.Mutex               // 保护读改写状态，与 storage 配套使用
	storage   CircuitBreakerStorage    // cache 存储，memory/redis 均可
	listeners []CircuitBreakerListener // 状态变更与调用结果回调
	now       func() time.Time         // 状态机时钟，测试可替换
}

// NewCircuitBreaker 创建熔断器，状态写入 cache，便于后续用 redis 替换 memory。
func NewCircuitBreaker(config *CircuitBreakerConfig) (CircuitBreakerHandler, error) {
	if config == nil {
		return nil, fmt.Errorf("配置不能为空")
	}
	storage, err := newCacheCircuitBreakerStorage(config)
	if err != nil {
		return nil, err
	}
	return &circuitBreakerImpl{
		config:    config,
		storage:   storage,
		listeners: make([]CircuitBreakerListener, 0),
		now:       time.Now,
	}, nil
}

// Check 判断指定节点键是否放行，不 Abort、不写 context。
// 半开态在放行时占用探测名额，避免并发探测超过 HalfOpenMaxRequests。
func (cb *circuitBreakerImpl) Check(key string) bool {
	if !cb.IsEnabled() || key == "" {
		return true
	}
	allowed, state := cb.allowRequest(key)
	if !allowed {
		cb.notifyCallRejected(key, state)
	}
	return allowed
}

// RecordKey 按指定节点键回写一次结果，供摘除使用。
func RecordKey(cb CircuitBreakerHandler, key string, success bool, responseTime time.Duration, err error) {
	if cb == nil || !cb.IsEnabled() || key == "" {
		return
	}
	ms := responseTime.Milliseconds()
	if success {
		cb.RecordSuccess(key, ms)
		return
	}
	cb.RecordFailure(key, ms, err)
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

// GetInfo 汇总缓存中已有熔断键的统计。
func (cb *circuitBreakerImpl) GetInfo() *CircuitBreakerInfo {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	info := &CircuitBreakerInfo{
		State:       StateClosed,
		WindowStart: time.Now().Unix(),
		WindowEnd:   time.Now().Unix(),
	}
	// 仅 cache 实现能按前缀扫键；其他存储只返回空汇总
	if cleaner, ok := cb.storage.(*cacheCircuitBreakerStorage); ok {
		keys, err := cleaner.c.Keys(context.Background(), circuitBreakerCachePrefix+"*")
		if err == nil {
			for _, cacheKey := range keys {
				raw, getErr := cleaner.c.Get(context.Background(), cacheKey)
				if getErr != nil || len(raw) == 0 {
					continue
				}
				circuit := &CircuitBreakerInfo{}
				if unmarshalErr := json.Unmarshal(raw, circuit); unmarshalErr != nil {
					continue
				}
				info.TotalRequests += circuit.TotalRequests
				info.SuccessRequests += circuit.SuccessRequests
				info.FailureRequests += circuit.FailureRequests
				info.SlowRequests += circuit.SlowRequests
				info.HalfOpenCount += circuit.HalfOpenCount
				if circuit.State == StateOpen {
					info.OpenCount++
				}
			}
		}
	}
	if info.TotalRequests > 0 {
		info.FailureRate = float64(info.FailureRequests) / float64(info.TotalRequests) * 100
		info.SlowRate = float64(info.SlowRequests) / float64(info.TotalRequests) * 100
	}
	return info
}

// Reset 清除缓存中的全部熔断状态。
func (cb *circuitBreakerImpl) Reset() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.storage.Cleanup()
}

// IsEnabled 检查是否启用
func (cb *circuitBreakerImpl) IsEnabled() bool {
	return cb.config.Enabled
}

// GetState 获取指定key的熔断器状态
// key: 熔断器key
// 返回: 熔断器状态，如果key不存在则返回 StateClosed
func (cb *circuitBreakerImpl) GetState(key string) CircuitBreakerState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.loadCircuit(key).State
}

// ForceOpen 强制打开熔断器（用于手动触发熔断）
// key: 熔断器key
func (cb *circuitBreakerImpl) ForceOpen(key string) error {
	cb.mu.Lock()
	circuit := cb.loadCircuit(key)
	oldState := circuit.State
	circuit.State = StateOpen
	circuit.OpenTime = cb.currentUnix()
	_ = cb.storage.SetInfo(key, circuit)
	cb.mu.Unlock()
	cb.notifyStateChange(key, oldState, StateOpen, circuit)
	return nil
}

// ForceClose 强制关闭熔断器（用于手动恢复服务）
// key: 熔断器key
func (cb *circuitBreakerImpl) ForceClose(key string) error {
	cb.mu.Lock()
	circuit := cb.loadCircuit(key)
	oldState := circuit.State
	circuit.State = StateClosed
	_ = cb.storage.SetInfo(key, circuit)
	cb.mu.Unlock()
	cb.notifyStateChange(key, oldState, StateClosed, circuit)
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
func (cb *circuitBreakerImpl) allowRequest(key string) (bool, CircuitBreakerState) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	circuit := cb.loadCircuit(key)
	now := cb.currentUnix()

	switch circuit.State {
	case StateOpen:
		// 开闸超时后转入半开，本请求占用第一个探测名额
		if now-circuit.OpenTime >= cb.config.OpenTimeoutSeconds {
			circuit.State = StateHalfOpen
			circuit.HalfOpenCount = 1
			circuit.HalfOpenSuccess = 0
			_ = cb.storage.SetInfo(key, circuit)
			go cb.notifyStateChange(key, StateOpen, StateHalfOpen, circuit)
			return true, StateHalfOpen
		}
		return false, StateOpen
	case StateHalfOpen:
		if circuit.HalfOpenCount >= int64(cb.config.HalfOpenMaxRequests) {
			return false, StateHalfOpen
		}
		circuit.HalfOpenCount++
		_ = cb.storage.SetInfo(key, circuit)
		return true, StateHalfOpen
	default:
		return true, StateClosed
	}
}

// loadCircuit 从缓存读取状态；不存在则返回关闭态。调用方须持有 cb.mu。
func (cb *circuitBreakerImpl) loadCircuit(key string) *CircuitBreakerInfo {
	info, err := cb.storage.GetInfo(key)
	if err != nil || info == nil {
		return newClosedCircuit()
	}
	if info.State == "" {
		info.State = StateClosed
	}
	return info
}

// RecordSuccess 记录成功调用并回写 cache。
// 半开态下成功次数达到 HalfOpenMaxRequests 后转为关闭并清空窗口统计。
func (cb *circuitBreakerImpl) RecordSuccess(key string, responseTime int64) {
	cb.mu.Lock()
	circuit := cb.loadCircuit(key)
	now := cb.currentUnix()
	circuit.LastRequestTime = now
	slow := cb.config.SlowCallThreshold > 0 && responseTime > cb.config.SlowCallThreshold
	cb.recordSample(circuit, now, false, slow)

	var changedFrom CircuitBreakerState
	changed := false
	// 半开探测成功达到配额后关闸，并清空窗口以免旧失败率立刻再次开闸
	if circuit.State == StateHalfOpen {
		circuit.HalfOpenSuccess++
		if circuit.HalfOpenSuccess >= int64(cb.config.HalfOpenMaxRequests) {
			changedFrom = circuit.State
			circuit.State = StateClosed
			resetWindow(circuit)
			changed = true
		}
	} else if circuit.State == StateClosed && cb.shouldTrip(circuit) {
		// 全是慢成功、没有传输失败时也要能开闸
		changedFrom = circuit.State
		circuit.State = StateOpen
		circuit.OpenTime = now
		circuit.OpenCount++
		changed = true
	}
	_ = cb.storage.SetInfo(key, circuit)
	listeners := append([]CircuitBreakerListener(nil), cb.listeners...)
	cb.mu.Unlock()

	if changed {
		cb.notifyStateChange(key, changedFrom, circuit.State, circuit)
	}
	for _, listener := range listeners {
		go listener.OnCallSuccess(key, responseTime)
	}
}

// RecordFailure 记录失败调用并回写 cache。
// 半开失败立即开闸；关闭态在失败率或慢调用率达到阈值时开闸。
func (cb *circuitBreakerImpl) RecordFailure(key string, responseTime int64, err error) {
	cb.mu.Lock()
	circuit := cb.loadCircuit(key)
	now := cb.currentUnix()
	circuit.LastFailureTime = now
	circuit.LastRequestTime = now
	slow := cb.config.SlowCallThreshold > 0 && responseTime > cb.config.SlowCallThreshold
	cb.recordSample(circuit, now, true, slow)

	var changedFrom CircuitBreakerState
	changed := false
	if circuit.State == StateHalfOpen {
		// 半开探测失败立即重新开闸
		changedFrom = circuit.State
		circuit.State = StateOpen
		circuit.OpenTime = now
		circuit.HalfOpenCount = 0
		circuit.HalfOpenSuccess = 0
		circuit.OpenCount++
		changed = true
	} else if circuit.State == StateClosed && cb.shouldTrip(circuit) {
		changedFrom = circuit.State
		circuit.State = StateOpen
		circuit.OpenTime = now
		circuit.OpenCount++
		changed = true
	}
	_ = cb.storage.SetInfo(key, circuit)
	listeners := append([]CircuitBreakerListener(nil), cb.listeners...)
	cb.mu.Unlock()

	if changed {
		cb.notifyStateChange(key, changedFrom, StateOpen, circuit)
	}
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
//  2. 失败率达到 ErrorRatePercent，或慢调用率达到 SlowCallRatePercent
func (cb *circuitBreakerImpl) shouldTrip(circuit *CircuitBreakerInfo) bool {
	// 只看滑动窗口内的样本，冷启动未达最小请求数不开闸
	if circuit.TotalRequests < int64(cb.config.MinimumRequests) {
		return false
	}

	if circuit.FailureRate >= float64(cb.config.ErrorRatePercent) {
		return true
	}
	if circuit.SlowRequests > 0 && cb.config.SlowCallRatePercent > 0 &&
		circuit.SlowRate >= float64(cb.config.SlowCallRatePercent) {
		return true
	}
	return false
}

// notifyStateChange 在锁外通知状态变更。
func (cb *circuitBreakerImpl) notifyStateChange(key string, from, to CircuitBreakerState, circuit *CircuitBreakerInfo) {
	infoCopy := *circuit
	cb.mu.Lock()
	listeners := append([]CircuitBreakerListener(nil), cb.listeners...)
	cb.mu.Unlock()
	for _, listener := range listeners {
		go listener.OnStateChange(key, from, to, &infoCopy)
	}
}

// notifyCallRejected 通知调用被拒绝。
func (cb *circuitBreakerImpl) notifyCallRejected(key string, state CircuitBreakerState) {
	cb.mu.Lock()
	listeners := append([]CircuitBreakerListener(nil), cb.listeners...)
	cb.mu.Unlock()
	for _, listener := range listeners {
		go listener.OnCallRejected(key, state)
	}
}

// memoryCircuitBreakerStorage 进程内 map 存储，仅保留兼容 CircuitBreakerStorage 接口。
// 现行路径使用 cacheCircuitBreakerStorage，不再把状态放在 circuitBreakerImpl 内存 map。
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
