package service

import (
	"fmt"
	"sync"
	"time"

	"gohub/internal/gateway/core"
	"gohub/internal/gateway/handler/circuitbreaker"
)

// Service 服务结构体，包含所有服务相关的功能
type Service struct {
	config         *ServiceConfig
	loadBalancer   LoadBalancer
	circuitBreaker circuitbreaker.CircuitBreakerHandler
	healthChecker  HealthChecker
	nodes          []*NodeConfig
	healthyNodes   map[string]*NodeConfig // nodeID -> NodeConfig
	unhealthyNodes map[string]*NodeConfig // nodeID -> NodeConfig
	mutex          sync.RWMutex
	stats          ServiceStats
	lastAccessTime time.Time
}

// ServiceStats 服务统计信息
type ServiceStats struct {
	TotalRequests       int64
	SuccessRequests     int64
	FailureRequests     int64
	LastRequestTime     time.Time
	AverageResponseTime time.Duration
	MaxResponseTime     time.Duration
	MinResponseTime     time.Duration
}

// NewService 创建新的服务实例
func NewService(config *ServiceConfig) (*Service, error) {
	if config == nil {
		return nil, fmt.Errorf("service config cannot be nil")
	}

	service := &Service{
		config:         config,
		nodes:          config.Nodes,
		healthyNodes:   make(map[string]*NodeConfig),
		unhealthyNodes: make(map[string]*NodeConfig),
		lastAccessTime: time.Now(),
	}

	// 初始化负载均衡器
	if err := service.initLoadBalancer(); err != nil {
		return nil, err
	}

	// 初始化熔断器
	if err := service.initCircuitBreaker(); err != nil {
		return nil, err
	}

	// 初始化健康检查器
	if err := service.initHealthChecker(); err != nil {
		return nil, err
	}

	// 初始化节点状态
	service.initNodeStatus()

	return service, nil
}

// initLoadBalancer 初始化负载均衡器
func (s *Service) initLoadBalancer() error {
	factory := NewLoadBalancerFactory()

	// 使用服务特定的负载均衡配置，如果没有则使用默认配置
	lbConfig := s.config.LoadBalancer
	if lbConfig == nil {
		lbConfig = &DefaultConfig
	}

	var err error
	s.loadBalancer, err = factory.CreateLoadBalancer(lbConfig)
	if err != nil {
		return fmt.Errorf("create load balancer failed: %w", err)
	}

	return nil
}

// initCircuitBreaker 初始化熔断器
func (s *Service) initCircuitBreaker() error {
	// 检查是否需要启用熔断器
	if s.config.CircuitBreaker == nil || !s.config.CircuitBreaker.Enabled {
		return nil
	}

	// 创建熔断器
	var err error
	s.circuitBreaker, err = circuitbreaker.NewCircuitBreaker(s.config.CircuitBreaker)
	if err != nil {
		return fmt.Errorf("create circuit breaker failed: %w", err)
	}

	return nil
}

// initHealthChecker 初始化健康检查器
func (s *Service) initHealthChecker() error {
	// 检查是否需要启用健康检查
	if s.config.LoadBalancer == nil ||
		s.config.LoadBalancer.HealthCheck == nil ||
		!s.config.LoadBalancer.HealthCheck.Enabled {
		return nil
	}

	// 创建健康检查器
	s.healthChecker = NewHTTPHealthChecker(s.config.LoadBalancer.HealthCheck)

	// 注册节点健康状态回调
	s.healthChecker.RegisterCallback(s.onNodeHealthChanged)

	// 将所有启用的节点添加到健康检查器
	for _, node := range s.nodes {
		if node.Enabled {
			if httpChecker, ok := s.healthChecker.(*HTTPHealthChecker); ok {
				httpChecker.AddNode(node)
			} else if advancedChecker, ok := s.healthChecker.(*AdvancedHealthChecker); ok {
				advancedChecker.AddNode(node)
			}
		}
	}

	// 启动健康检查
	if err := s.healthChecker.Start(); err != nil {
		return fmt.Errorf("start health checker failed: %w", err)
	}

	return nil
}

// initNodeStatus 初始化节点状态
func (s *Service) initNodeStatus() {
	for _, node := range s.nodes {
		if node.Enabled {
			if node.Health {
				s.healthyNodes[node.ID] = node
			} else {
				s.unhealthyNodes[node.ID] = node
			}
		}
	}
}

// SelectNode 选择节点
func (s *Service) SelectNode(ctx *core.Context) (*NodeConfig, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 记录请求统计信息
	defer func() {
		s.stats.TotalRequests++
		s.stats.LastRequestTime = time.Now()
		s.lastAccessTime = time.Now()
	}()

	// 检查熔断器状态
	if s.circuitBreaker != nil && !s.circuitBreaker.Handle(ctx) {
		s.stats.FailureRequests++
		return nil, fmt.Errorf("service %s circuit breaker is open", s.config.ID)
	}

	// 如果没有健康节点，返回错误
	if len(s.healthyNodes) == 0 {
		s.stats.FailureRequests++
		return nil, ErrNoAvailableNode
	}

	// 准备服务配置用于负载均衡
	serviceConfig := &ServiceConfig{
		ID:       s.config.ID,
		Name:     s.config.Name,
		Strategy: s.config.Strategy,
		Nodes:    make([]*NodeConfig, 0, len(s.healthyNodes)),
	}

	for _, node := range s.healthyNodes {
		serviceConfig.Nodes = append(serviceConfig.Nodes, node)
	}

	// 使用负载均衡器选择节点
	node := s.loadBalancer.Select(serviceConfig, ctx)
	if node == nil {
		s.stats.FailureRequests++
		return nil, ErrNoAvailableNode
	}

	s.stats.SuccessRequests++
	return node, nil
}

// AddNode 添加节点
func (s *Service) AddNode(node *NodeConfig) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if node == nil {
		return fmt.Errorf("node cannot be nil")
	}

	// 检查节点是否已存在
	for _, n := range s.nodes {
		if n.ID == node.ID {
			return ErrNodeExists
		}
	}

	// 添加到节点列表
	s.nodes = append(s.nodes, node)

	// 根据节点状态添加到相应的map
	if node.Enabled {
		if node.Health {
			s.healthyNodes[node.ID] = node
		} else {
			s.unhealthyNodes[node.ID] = node
		}

		// 添加到健康检查器
		if s.healthChecker != nil {
			if httpChecker, ok := s.healthChecker.(*HTTPHealthChecker); ok {
				httpChecker.AddNode(node)
			} else if advancedChecker, ok := s.healthChecker.(*AdvancedHealthChecker); ok {
				advancedChecker.AddNode(node)
			}
		}
	}

	return nil
}

// RemoveNode 移除节点
func (s *Service) RemoveNode(nodeID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 从节点列表中移除
	found := false
	for i, node := range s.nodes {
		if node.ID == nodeID {
			s.nodes = append(s.nodes[:i], s.nodes[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return ErrNodeNotFound
	}

	// 从健康/不健康节点map中移除
	delete(s.healthyNodes, nodeID)
	delete(s.unhealthyNodes, nodeID)

	// 从健康检查器中移除
	if s.healthChecker != nil {
		if httpChecker, ok := s.healthChecker.(*HTTPHealthChecker); ok {
			httpChecker.RemoveNode(nodeID)
		} else if advancedChecker, ok := s.healthChecker.(*AdvancedHealthChecker); ok {
			advancedChecker.RemoveNode(nodeID)
		}
	}

	return nil
}

// UpdateNodeHealth 更新节点健康状态
func (s *Service) UpdateNodeHealth(nodeID string, healthy bool) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 查找节点
	var node *NodeConfig
	for _, n := range s.nodes {
		if n.ID == nodeID {
			node = n
			break
		}
	}

	if node == nil {
		return ErrNodeNotFound
	}

	// 更新健康状态
	node.Health = healthy

	// 更新节点在map中的位置
	if healthy {
		delete(s.unhealthyNodes, nodeID)
		if node.Enabled {
			s.healthyNodes[nodeID] = node
		}
	} else {
		delete(s.healthyNodes, nodeID)
		if node.Enabled {
			s.unhealthyNodes[nodeID] = node
		}
	}

	return nil
}

// UpdateNodeStatus 更新节点启用状态
func (s *Service) UpdateNodeStatus(nodeID string, enabled bool) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 查找节点
	var node *NodeConfig
	for _, n := range s.nodes {
		if n.ID == nodeID {
			node = n
			break
		}
	}

	if node == nil {
		return ErrNodeNotFound
	}

	// 更新启用状态
	node.Enabled = enabled

	// 更新节点在map中的位置
	if enabled {
		if node.Health {
			s.healthyNodes[nodeID] = node
			delete(s.unhealthyNodes, nodeID)
		} else {
			s.unhealthyNodes[nodeID] = node
			delete(s.healthyNodes, nodeID)
		}

		// 添加到健康检查器
		if s.healthChecker != nil {
			if httpChecker, ok := s.healthChecker.(*HTTPHealthChecker); ok {
				httpChecker.AddNode(node)
			} else if advancedChecker, ok := s.healthChecker.(*AdvancedHealthChecker); ok {
				advancedChecker.AddNode(node)
			}
		}
	} else {
		delete(s.healthyNodes, nodeID)
		delete(s.unhealthyNodes, nodeID)

		// 从健康检查器中移除
		if s.healthChecker != nil {
			if httpChecker, ok := s.healthChecker.(*HTTPHealthChecker); ok {
				httpChecker.RemoveNode(nodeID)
			} else if advancedChecker, ok := s.healthChecker.(*AdvancedHealthChecker); ok {
				advancedChecker.RemoveNode(nodeID)
			}
		}
	}

	return nil
}

// UpdateNodeWeight 更新节点权重
func (s *Service) UpdateNodeWeight(nodeID string, weight int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 查找节点
	var node *NodeConfig
	for _, n := range s.nodes {
		if n.ID == nodeID {
			node = n
			break
		}
	}

	if node == nil {
		return ErrNodeNotFound
	}

	// 更新权重
	node.Weight = weight

	// 更新负载均衡器中的权重
	if s.loadBalancer != nil {
		return s.loadBalancer.UpdateNodeWeight(s.config.ID, nodeID, weight)
	}

	return nil
}

// RecordSuccess 记录成功调用
func (s *Service) RecordSuccess(responseTime time.Duration) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.stats.SuccessRequests++

	// 更新响应时间统计
	if responseTime > 0 {
		if s.stats.MaxResponseTime < responseTime {
			s.stats.MaxResponseTime = responseTime
		}

		if s.stats.MinResponseTime == 0 || s.stats.MinResponseTime > responseTime {
			s.stats.MinResponseTime = responseTime
		}

		// 简单计算平均响应时间
		if s.stats.AverageResponseTime == 0 {
			s.stats.AverageResponseTime = responseTime
		} else {
			s.stats.AverageResponseTime = (s.stats.AverageResponseTime + responseTime) / 2
		}
	}
}

// RecordFailure 记录失败调用
func (s *Service) RecordFailure() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.stats.FailureRequests++
}

// GetHealthyNodes 获取健康节点
func (s *Service) GetHealthyNodes() []*NodeConfig {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	nodes := make([]*NodeConfig, 0, len(s.healthyNodes))
	for _, node := range s.healthyNodes {
		nodes = append(nodes, node)
	}
	return nodes
}

// GetUnhealthyNodes 获取不健康节点
func (s *Service) GetUnhealthyNodes() []*NodeConfig {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	nodes := make([]*NodeConfig, 0, len(s.unhealthyNodes))
	for _, node := range s.unhealthyNodes {
		nodes = append(nodes, node)
	}
	return nodes
}

// GetAllNodes 获取所有节点
func (s *Service) GetAllNodes() []*NodeConfig {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 返回节点的副本以避免并发问题
	nodes := make([]*NodeConfig, len(s.nodes))
	copy(nodes, s.nodes)
	return nodes
}

// GetConfig 获取服务配置
func (s *Service) GetConfig() *ServiceConfig {
	return s.config
}

// GetStats 获取服务统计信息
func (s *Service) GetStats() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	stats := make(map[string]interface{})
	stats["id"] = s.config.ID
	stats["name"] = s.config.Name
	stats["strategy"] = s.config.Strategy
	stats["total_nodes"] = len(s.nodes)
	stats["healthy_nodes"] = len(s.healthyNodes)
	stats["unhealthy_nodes"] = len(s.unhealthyNodes)
	stats["total_requests"] = s.stats.TotalRequests
	stats["success_requests"] = s.stats.SuccessRequests
	stats["failure_requests"] = s.stats.FailureRequests
	stats["average_response_time_ms"] = s.stats.AverageResponseTime.Milliseconds()
	stats["max_response_time_ms"] = s.stats.MaxResponseTime.Milliseconds()
	stats["min_response_time_ms"] = s.stats.MinResponseTime.Milliseconds()
	stats["last_request_time"] = s.stats.LastRequestTime
	stats["last_access_time"] = s.lastAccessTime

	if s.circuitBreaker != nil {
		stats["circuit_breaker_state"] = s.circuitBreaker.GetState("cb_service:" + s.config.ID)
	}

	if s.loadBalancer != nil {
		stats["load_balancer_stats"] = s.loadBalancer.GetStats()
	}

	return stats
}

// Close 关闭服务相关资源
func (s *Service) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var lastErr error

	// 停止健康检查器
	if s.healthChecker != nil {
		if err := s.healthChecker.Stop(); err != nil {
			lastErr = err
		}
	}

	// 清理熔断器资源（如果有需要的话）
	if s.circuitBreaker != nil {
		if cleaner, ok := s.circuitBreaker.(interface{ Close() error }); ok {
			if err := cleaner.Close(); err != nil && lastErr == nil {
				lastErr = err
			}
		}
	}

	// 清理负载均衡器资源
	if s.loadBalancer != nil {
		s.loadBalancer.Reset()
	}

	// 清空节点映射
	s.healthyNodes = make(map[string]*NodeConfig)
	s.unhealthyNodes = make(map[string]*NodeConfig)

	return lastErr
}

// onNodeHealthChanged 节点健康状态变化回调
func (s *Service) onNodeHealthChanged(nodeID string, healthy bool) {
	_ = s.UpdateNodeHealth(nodeID, healthy)
}
