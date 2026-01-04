package service

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// HTTPHealthChecker HTTP健康检查器
type HTTPHealthChecker struct {
	config    *HealthConfig
	callbacks []HealthCheckCallback
	client    *http.Client
	mu        sync.RWMutex
	running   bool
	stopCh    chan struct{}
	nodes     map[string]*nodeHealthStatus
}

// nodeHealthStatus 节点健康状态
type nodeHealthStatus struct {
	node               *NodeConfig
	consecutiveSuccess int
	consecutiveFailure int
	lastCheck          time.Time
}

// NewHTTPHealthChecker 创建HTTP健康检查器
func NewHTTPHealthChecker(config *HealthConfig) HealthChecker {
	return &HTTPHealthChecker{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		nodes:  make(map[string]*nodeHealthStatus),
		stopCh: make(chan struct{}),
	}
}

// Start 启动健康检查
func (h *HTTPHealthChecker) Start() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.running {
		return fmt.Errorf("健康检查器已经在运行")
	}

	h.running = true
	h.stopCh = make(chan struct{})

	// 启动健康检查goroutine
	go h.healthCheckLoop()

	return nil
}

// Stop 停止健康检查
// 这个方法对于防止资源泄漏至关重要：
// 1. 设置running标志为false，防止新的健康检查被执行
// 2. 关闭stopCh通道，这会触发healthCheckLoop中的select语句退出循环
// 3. 当healthCheckLoop退出后，相关的goroutine会结束，释放资源
// 4. 如果不调用此方法，healthCheckLoop会一直运行，导致goroutine泄漏
func (h *HTTPHealthChecker) Stop() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.running {
		return nil
	}

	h.running = false
	close(h.stopCh)

	return nil
}

// CheckNode 检查单个节点健康状态
func (h *HTTPHealthChecker) CheckNode(node *NodeConfig) bool {
	if h.config == nil || !h.config.Enabled {
		return true
	}

	// 构建健康检查URL
	url := node.URL + h.config.Path

	// 创建请求
	req, err := http.NewRequest(h.config.Method, url, nil)
	if err != nil {
		return false
	}

	// 添加自定义头部
	for key, value := range h.config.Headers {
		req.Header.Set(key, value)
	}

	// 执行请求
	resp, err := h.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// 检查状态码
	for _, expectedCode := range h.config.ExpectedStatusCodes {
		if resp.StatusCode == expectedCode {
			return true
		}
	}

	return false
}

// RegisterCallback 注册健康状态变化回调
func (h *HTTPHealthChecker) RegisterCallback(callback HealthCheckCallback) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.callbacks = append(h.callbacks, callback)
}

// healthCheckLoop 健康检查循环
func (h *HTTPHealthChecker) healthCheckLoop() {
	ticker := time.NewTicker(h.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-h.stopCh:
			return
		case <-ticker.C:
			h.performHealthChecks()
		}
	}
}

// performHealthChecks 执行健康检查
// 注意：此方法会为每个节点启动一个 goroutine 进行健康检查
// 每个 goroutine 都有超时控制，防止长时间阻塞导致资源泄漏
func (h *HTTPHealthChecker) performHealthChecks() {
	h.mu.RLock()
	// 创建节点状态的副本，避免在检查过程中持有锁
	nodesCopy := make(map[string]*nodeHealthStatus)
	for k, v := range h.nodes {
		// 注意：这里复制的是指针，所以修改 status 中的字段会影响原始数据
		// 但我们需要通过锁保护对原始 nodes map 的访问
		nodesCopy[k] = v
	}
	h.mu.RUnlock()

	// 为每个节点启动健康检查 goroutine，使用 context 控制超时
	for nodeID, status := range nodesCopy {
		go func(id string, st *nodeHealthStatus) {
			// 使用 context 控制超时，防止健康检查卡住导致 goroutine 泄漏
			ctx, cancel := context.WithTimeout(context.Background(), h.config.Timeout*2)
			defer cancel()

			// 使用 channel 确保 goroutine 能够正常退出
			done := make(chan struct{})
			go func() {
				h.checkNodeHealth(st)
				close(done)
			}()

			select {
			case <-done:
				// 健康检查完成
			case <-ctx.Done():
				// 超时，记录但不影响其他节点的检查
			}
		}(nodeID, status)
	}
}

// checkNodeHealth 检查节点健康状态
// 注意：此方法会修改共享的 status 对象，需要通过锁保护
// 但由于 status 是从 nodes map 中获取的，我们需要通过锁保护对 nodes map 的访问
func (h *HTTPHealthChecker) checkNodeHealth(status *nodeHealthStatus) {
	// 执行健康检查（在锁外执行，避免长时间持有锁）
	healthy := h.CheckNode(status.node)

	// 需要加锁保护对 status 的修改，因为多个 goroutine 可能同时检查同一个节点
	h.mu.Lock()

	// 检查节点是否仍然存在（可能在检查过程中被移除）
	if _, exists := h.nodes[status.node.ID]; !exists {
		h.mu.Unlock()
		return
	}

	// 更新检查时间
	status.lastCheck = time.Now()

	// 记录之前的健康状态
	previousHealth := status.node.Health
	nodeID := status.node.ID
	newHealth := status.node.Health

	// 更新连续成功/失败计数
	if healthy {
		status.consecutiveSuccess++
		status.consecutiveFailure = 0

		// 连续成功达到阈值，标记为健康
		if status.consecutiveSuccess >= h.config.HealthyThreshold {
			status.node.Health = true
			newHealth = true
		}
	} else {
		status.consecutiveFailure++
		status.consecutiveSuccess = 0

		// 连续失败达到阈值，标记为不健康
		if status.consecutiveFailure >= h.config.UnhealthyThreshold {
			status.node.Health = false
			newHealth = false
		}
	}

	// 复制回调列表，准备在锁外执行回调
	var callbacks []HealthCheckCallback
	needNotify := previousHealth != newHealth
	if needNotify {
		callbacks = make([]HealthCheckCallback, len(h.callbacks))
		copy(callbacks, h.callbacks)
	}

	// 释放锁
	h.mu.Unlock()

	// 如果健康状态发生变化，触发回调（在锁外执行，避免死锁）
	if needNotify {
		// 在锁外通知回调，避免回调中可能再次获取锁导致死锁
		for _, callback := range callbacks {
			go func(cb HealthCheckCallback) {
				defer func() {
					if r := recover(); r != nil {
						// 忽略回调中的 panic，防止影响其他回调
					}
				}()
				cb(nodeID, newHealth)
			}(callback)
		}
	}
}

// notifyCallbacks 通知回调函数
// 注意：此方法已废弃，回调通知现在在 checkNodeHealth 中直接处理
// 保留此方法是为了向后兼容，但实际不会被调用
func (h *HTTPHealthChecker) notifyCallbacks(nodeID string, healthy bool) {
	h.mu.RLock()
	callbacks := make([]HealthCheckCallback, len(h.callbacks))
	copy(callbacks, h.callbacks)
	h.mu.RUnlock()

	// 为每个回调启动 goroutine，使用 context 控制超时
	for _, callback := range callbacks {
		go func(cb HealthCheckCallback) {
			// 使用 context 控制回调执行超时，防止回调卡住导致 goroutine 泄漏
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			done := make(chan struct{})
			go func() {
				defer func() {
					if r := recover(); r != nil {
						// 忽略回调中的 panic，防止影响其他回调
					}
					close(done)
				}()
				cb(nodeID, healthy)
			}()

			select {
			case <-done:
				// 回调执行完成
			case <-ctx.Done():
				// 回调超时，记录但不影响其他回调
			}
		}(callback)
	}
}

// AddNode 添加节点到健康检查
func (h *HTTPHealthChecker) AddNode(node *NodeConfig) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.nodes[node.ID] = &nodeHealthStatus{
		node:      node,
		lastCheck: time.Now(),
	}
}

// RemoveNode 从健康检查中移除节点
func (h *HTTPHealthChecker) RemoveNode(nodeID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.nodes, nodeID)
}

// GetNodeStatus 获取节点状态
func (h *HTTPHealthChecker) GetNodeStatus(nodeID string) *nodeHealthStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()

	status, ok := h.nodes[nodeID]
	if !ok {
		return nil
	}

	// 返回副本以避免并发问题
	return &nodeHealthStatus{
		node:               status.node,
		consecutiveSuccess: status.consecutiveSuccess,
		consecutiveFailure: status.consecutiveFailure,
		lastCheck:          status.lastCheck,
	}
}

// GetHealthStats 获取健康检查统计信息
func (h *HTTPHealthChecker) GetHealthStats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["total_nodes"] = len(h.nodes)
	stats["running"] = h.running

	nodeStats := make(map[string]interface{})
	for nodeID, status := range h.nodes {
		nodeStats[nodeID] = map[string]interface{}{
			"healthy":             status.node.Health,
			"consecutive_success": status.consecutiveSuccess,
			"consecutive_failure": status.consecutiveFailure,
			"last_check":          status.lastCheck,
		}
	}
	stats["nodes"] = nodeStats

	return stats
}

// Close 关闭健康检查器并清理资源
// 实现此方法的目的：
// 1. 符合gateway.go中使用的interface{ Close() error }类型断言
// 2. 提供与Stop方法相同的功能，但遵循Go资源清理的命名惯例
// 3. 确保即使通过类型断言调用也能正确停止健康检查goroutine
// 4. 保持与其他组件（如代理处理器）的一致性
func (h *HTTPHealthChecker) Close() error {
	return h.Stop()
}

// NoOpHealthChecker 无操作健康检查器（用于禁用健康检查）
type NoOpHealthChecker struct{}

// NewNoOpHealthChecker 创建无操作健康检查器
func NewNoOpHealthChecker() HealthChecker {
	return &NoOpHealthChecker{}
}

// Start 启动健康检查（无操作）
func (n *NoOpHealthChecker) Start() error {
	return nil
}

// Stop 停止健康检查（无操作）
func (n *NoOpHealthChecker) Stop() error {
	return nil
}

// CheckNode 检查单个节点健康状态（始终返回true）
func (n *NoOpHealthChecker) CheckNode(node *NodeConfig) bool {
	return true
}

// RegisterCallback 注册健康状态变化回调（无操作）
func (n *NoOpHealthChecker) RegisterCallback(callback HealthCheckCallback) {
	// 无操作
}

// Close 关闭健康检查器（无操作）
// 实现此方法是为了与其他健康检查器保持一致性
// 并满足gateway.go中的类型断言检查
func (n *NoOpHealthChecker) Close() error {
	return nil
}

// AdvancedHealthChecker 高级健康检查器（支持自定义检查逻辑）
type AdvancedHealthChecker struct {
	config      *HealthConfig
	callbacks   []HealthCheckCallback
	client      *http.Client
	mu          sync.RWMutex
	running     bool
	stopCh      chan struct{}
	nodes       map[string]*nodeHealthStatus
	customCheck func(*NodeConfig) bool // 自定义检查函数
}

// NewAdvancedHealthChecker 创建高级健康检查器
func NewAdvancedHealthChecker(config *HealthConfig, customCheck func(*NodeConfig) bool) HealthChecker {
	return &AdvancedHealthChecker{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		nodes:       make(map[string]*nodeHealthStatus),
		stopCh:      make(chan struct{}),
		customCheck: customCheck,
	}
}

// Start 启动健康检查
func (a *AdvancedHealthChecker) Start() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.running {
		return fmt.Errorf("健康检查器已经在运行")
	}

	a.running = true
	a.stopCh = make(chan struct{})

	// 启动健康检查goroutine
	go a.healthCheckLoop()

	return nil
}

// Stop 停止健康检查
// 这个方法对于防止资源泄漏至关重要：
// 1. 设置running标志为false，防止新的健康检查被执行
// 2. 关闭stopCh通道，这会触发healthCheckLoop中的select语句退出循环
// 3. 当healthCheckLoop退出后，相关的goroutine会结束，释放资源
// 4. 如果不调用此方法，healthCheckLoop会一直运行，导致goroutine泄漏
func (a *AdvancedHealthChecker) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.running {
		return nil
	}

	a.running = false
	close(a.stopCh)

	return nil
}

// CheckNode 检查单个节点健康状态
func (a *AdvancedHealthChecker) CheckNode(node *NodeConfig) bool {
	if a.config == nil || !a.config.Enabled {
		return true
	}

	// 如果有自定义检查函数，优先使用
	if a.customCheck != nil {
		return a.customCheck(node)
	}

	// 否则使用默认的HTTP检查
	return a.httpCheck(node)
}

// RegisterCallback 注册健康状态变化回调
func (a *AdvancedHealthChecker) RegisterCallback(callback HealthCheckCallback) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.callbacks = append(a.callbacks, callback)
}

// httpCheck HTTP健康检查
func (a *AdvancedHealthChecker) httpCheck(node *NodeConfig) bool {
	url := node.URL + a.config.Path

	ctx, cancel := context.WithTimeout(context.Background(), a.config.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, a.config.Method, url, nil)
	if err != nil {
		return false
	}

	// 添加自定义头部
	for key, value := range a.config.Headers {
		req.Header.Set(key, value)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// 检查状态码
	for _, expectedCode := range a.config.ExpectedStatusCodes {
		if resp.StatusCode == expectedCode {
			return true
		}
	}

	return false
}

// healthCheckLoop 健康检查循环
func (a *AdvancedHealthChecker) healthCheckLoop() {
	ticker := time.NewTicker(a.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-a.stopCh:
			return
		case <-ticker.C:
			a.performHealthChecks()
		}
	}
}

// performHealthChecks 执行健康检查
// 注意：此方法会为每个节点启动一个 goroutine 进行健康检查
// 每个 goroutine 都有超时控制，防止长时间阻塞导致资源泄漏
func (a *AdvancedHealthChecker) performHealthChecks() {
	a.mu.RLock()
	// 创建节点状态的副本，避免在检查过程中持有锁
	nodesCopy := make(map[string]*nodeHealthStatus)
	for k, v := range a.nodes {
		// 注意：这里复制的是指针，所以修改 status 中的字段会影响原始数据
		// 但我们需要通过锁保护对原始 nodes map 的访问
		nodesCopy[k] = v
	}
	a.mu.RUnlock()

	// 为每个节点启动健康检查 goroutine，使用 context 控制超时
	for nodeID, status := range nodesCopy {
		go func(id string, st *nodeHealthStatus) {
			// 使用 context 控制超时，防止健康检查卡住导致 goroutine 泄漏
			ctx, cancel := context.WithTimeout(context.Background(), a.config.Timeout*2)
			defer cancel()

			// 使用 channel 确保 goroutine 能够正常退出
			done := make(chan struct{})
			go func() {
				a.checkNodeHealth(st)
				close(done)
			}()

			select {
			case <-done:
				// 健康检查完成
			case <-ctx.Done():
				// 超时，记录但不影响其他节点的检查
			}
		}(nodeID, status)
	}
}

// checkNodeHealth 检查节点健康状态
// 注意：此方法会修改共享的 status 对象，需要通过锁保护
// 但由于 status 是从 nodes map 中获取的，我们需要通过锁保护对 nodes map 的访问
func (a *AdvancedHealthChecker) checkNodeHealth(status *nodeHealthStatus) {
	// 执行健康检查（在锁外执行，避免长时间持有锁）
	healthy := a.CheckNode(status.node)

	// 需要加锁保护对 status 的修改，因为多个 goroutine 可能同时检查同一个节点
	a.mu.Lock()

	// 检查节点是否仍然存在（可能在检查过程中被移除）
	if _, exists := a.nodes[status.node.ID]; !exists {
		a.mu.Unlock()
		return
	}

	// 更新检查时间
	status.lastCheck = time.Now()

	// 记录之前的健康状态
	previousHealth := status.node.Health
	nodeID := status.node.ID
	newHealth := status.node.Health

	// 更新连续成功/失败计数
	if healthy {
		status.consecutiveSuccess++
		status.consecutiveFailure = 0

		// 连续成功达到阈值，标记为健康
		if status.consecutiveSuccess >= a.config.HealthyThreshold {
			status.node.Health = true
			newHealth = true
		}
	} else {
		status.consecutiveFailure++
		status.consecutiveSuccess = 0

		// 连续失败达到阈值，标记为不健康
		if status.consecutiveFailure >= a.config.UnhealthyThreshold {
			status.node.Health = false
			newHealth = false
		}
	}

	// 复制回调列表，准备在锁外执行回调
	var callbacks []HealthCheckCallback
	needNotify := previousHealth != newHealth
	if needNotify {
		callbacks = make([]HealthCheckCallback, len(a.callbacks))
		copy(callbacks, a.callbacks)
	}

	// 释放锁
	a.mu.Unlock()

	// 如果健康状态发生变化，触发回调（在锁外执行，避免死锁）
	if needNotify {
		// 在锁外通知回调，避免回调中可能再次获取锁导致死锁
		for _, callback := range callbacks {
			go func(cb HealthCheckCallback) {
				defer func() {
					if r := recover(); r != nil {
						// 忽略回调中的 panic，防止影响其他回调
					}
				}()
				cb(nodeID, newHealth)
			}(callback)
		}
	}
}

// notifyCallbacks 通知回调函数
// 注意：此方法已废弃，回调通知现在在 checkNodeHealth 中直接处理
// 保留此方法是为了向后兼容，但实际不会被调用
func (a *AdvancedHealthChecker) notifyCallbacks(nodeID string, healthy bool) {
	a.mu.RLock()
	callbacks := make([]HealthCheckCallback, len(a.callbacks))
	copy(callbacks, a.callbacks)
	a.mu.RUnlock()

	// 为每个回调启动 goroutine，使用 context 控制超时
	for _, callback := range callbacks {
		go func(cb HealthCheckCallback) {
			// 使用 context 控制回调执行超时，防止回调卡住导致 goroutine 泄漏
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			done := make(chan struct{})
			go func() {
				defer func() {
					if r := recover(); r != nil {
						// 忽略回调中的 panic，防止影响其他回调
					}
					close(done)
				}()
				cb(nodeID, healthy)
			}()

			select {
			case <-done:
				// 回调执行完成
			case <-ctx.Done():
				// 回调超时，记录但不影响其他回调
			}
		}(callback)
	}
}

// AddNode 添加节点到健康检查
func (a *AdvancedHealthChecker) AddNode(node *NodeConfig) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.nodes[node.ID] = &nodeHealthStatus{
		node:      node,
		lastCheck: time.Now(),
	}
}

// RemoveNode 从健康检查中移除节点
func (a *AdvancedHealthChecker) RemoveNode(nodeID string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	delete(a.nodes, nodeID)
}

// Close 关闭健康检查器并清理资源
// 实现此方法的目的：
// 1. 符合gateway.go中使用的interface{ Close() error }类型断言
// 2. 提供与Stop方法相同的功能，但遵循Go资源清理的命名惯例
// 3. 确保即使通过类型断言调用也能正确停止健康检查goroutine
// 4. 保持与其他组件（如代理处理器）的一致性
func (a *AdvancedHealthChecker) Close() error {
	return a.Stop()
}
