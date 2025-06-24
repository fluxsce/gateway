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
func (h *HTTPHealthChecker) performHealthChecks() {
	h.mu.RLock()
	nodesCopy := make(map[string]*nodeHealthStatus)
	for k, v := range h.nodes {
		nodesCopy[k] = v
	}
	h.mu.RUnlock()

	for _, status := range nodesCopy {
		go h.checkNodeHealth(status)
	}
}

// checkNodeHealth 检查节点健康状态
func (h *HTTPHealthChecker) checkNodeHealth(status *nodeHealthStatus) {
	healthy := h.CheckNode(status.node)
	status.lastCheck = time.Now()

	previousHealth := status.node.Health

	if healthy {
		status.consecutiveSuccess++
		status.consecutiveFailure = 0

		// 连续成功达到阈值，标记为健康
		if status.consecutiveSuccess >= h.config.HealthyThreshold {
			status.node.Health = true
		}
	} else {
		status.consecutiveFailure++
		status.consecutiveSuccess = 0

		// 连续失败达到阈值，标记为不健康
		if status.consecutiveFailure >= h.config.UnhealthyThreshold {
			status.node.Health = false
		}
	}

	// 如果健康状态发生变化，触发回调
	if previousHealth != status.node.Health {
		h.notifyCallbacks(status.node.ID, status.node.Health)
	}
}

// notifyCallbacks 通知回调函数
func (h *HTTPHealthChecker) notifyCallbacks(nodeID string, healthy bool) {
	h.mu.RLock()
	callbacks := make([]HealthCheckCallback, len(h.callbacks))
	copy(callbacks, h.callbacks)
	h.mu.RUnlock()

	for _, callback := range callbacks {
		go func(cb HealthCheckCallback) {
			defer func() {
				if r := recover(); r != nil {
					// 忽略回调中的panic
				}
			}()
			cb(nodeID, healthy)
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
func (a *AdvancedHealthChecker) performHealthChecks() {
	a.mu.RLock()
	nodesCopy := make(map[string]*nodeHealthStatus)
	for k, v := range a.nodes {
		nodesCopy[k] = v
	}
	a.mu.RUnlock()

	for _, status := range nodesCopy {
		go a.checkNodeHealth(status)
	}
}

// checkNodeHealth 检查节点健康状态
func (a *AdvancedHealthChecker) checkNodeHealth(status *nodeHealthStatus) {
	healthy := a.CheckNode(status.node)
	status.lastCheck = time.Now()

	previousHealth := status.node.Health

	if healthy {
		status.consecutiveSuccess++
		status.consecutiveFailure = 0

		// 连续成功达到阈值，标记为健康
		if status.consecutiveSuccess >= a.config.HealthyThreshold {
			status.node.Health = true
		}
	} else {
		status.consecutiveFailure++
		status.consecutiveSuccess = 0

		// 连续失败达到阈值，标记为不健康
		if status.consecutiveFailure >= a.config.UnhealthyThreshold {
			status.node.Health = false
		}
	}

	// 如果健康状态发生变化，触发回调
	if previousHealth != status.node.Health {
		a.notifyCallbacks(status.node.ID, status.node.Health)
	}
}

// notifyCallbacks 通知回调函数
func (a *AdvancedHealthChecker) notifyCallbacks(nodeID string, healthy bool) {
	a.mu.RLock()
	callbacks := make([]HealthCheckCallback, len(a.callbacks))
	copy(callbacks, a.callbacks)
	a.mu.RUnlock()

	for _, callback := range callbacks {
		go func(cb HealthCheckCallback) {
			defer func() {
				if r := recover(); r != nil {
					// 忽略回调中的panic
				}
			}()
			cb(nodeID, healthy)
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
