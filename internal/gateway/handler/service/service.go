package service

import (
	"fmt"
	"sync"
	"time"

	"gateway/internal/gateway/core"
	"gateway/internal/gateway/handler/circuitbreaker"
)

// Service 服务结构体，包含所有服务相关的功能
// 注意：此结构体管理服务的生命周期，包括：
// - 负载均衡器：负责选择后端节点
// - 熔断器：保护后端服务，防止雪崩
// - 健康检查器：定期检查后端节点健康状态（可以是独立的或共享的）
// - 节点管理：直接使用 config.Nodes 维护节点列表，通过节点状态字段区分健康/不健康节点
// 所有对节点的操作都需要通过 mutex 保护，确保并发安全
type Service struct {
	config           *ServiceConfig                       // 服务配置（包含节点列表，节点状态通过节点的 Health 和 Enabled 字段维护）
	loadBalancer     LoadBalancer                         // 负载均衡器
	circuitBreaker   circuitbreaker.CircuitBreakerHandler // 熔断器（可选）
	healthChecker    HealthChecker                        // 健康检查器（可选，仅在未使用共享检查器时使用）
	useSharedChecker bool                                 // 是否使用共享健康检查器（如果为 true，健康检查由 ServiceManager 的共享检查器处理）
	mutex            sync.RWMutex                         // 读写锁，保护所有共享状态（包括 config.Nodes）
	stats            ServiceStats                         // 服务统计信息
	lastAccessTime   time.Time                            // 最后访问时间，用于服务清理
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
// useSharedChecker: 是否使用共享健康检查器（如果为 true，健康检查由 ServiceManager 的共享检查器处理）
func NewService(config *ServiceConfig, useSharedChecker bool) (*Service, error) {
	if config == nil {
		return nil, fmt.Errorf("service config cannot be nil")
	}

	service := &Service{
		config:           config,
		lastAccessTime:   time.Now(),
		useSharedChecker: useSharedChecker,
	}

	// 初始化负载均衡器
	if err := service.initLoadBalancer(); err != nil {
		return nil, err
	}

	// 初始化熔断器
	if err := service.initCircuitBreaker(); err != nil {
		return nil, err
	}

	// 初始化健康检查器（如果使用共享检查器，则注册到共享检查器；否则创建独立检查器）
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
// 注意：当前熔断器逻辑尚未完全实现，暂时保留接口以便后续扩展
func (s *Service) initCircuitBreaker() error {
	// TODO: 熔断器功能待实现
	// 检查是否需要启用熔断器
	// if s.config.CircuitBreaker == nil || !s.config.CircuitBreaker.Enabled {
	// 	return nil
	// }

	// // 创建熔断器
	// var err error
	// s.circuitBreaker, err = circuitbreaker.NewCircuitBreaker(s.config.CircuitBreaker)
	// if err != nil {
	// 	return fmt.Errorf("create circuit breaker failed: %w", err)
	// }

	return nil
}

// initHealthChecker 初始化健康检查器
// 优先使用共享健康检查器（如果提供），否则使用独立的检查器（保持向后兼容）
func (s *Service) initHealthChecker() error {
	// 检查是否需要启用健康检查
	if s.config.HealthCheck == nil || !s.config.HealthCheck.Enabled {
		return nil
	}

	// 如果使用共享健康检查器，则不需要创建独立的检查器
	// 共享检查器会自动从 ServiceManager 获取所有服务和节点，直接更新节点健康状态
	if s.useSharedChecker {
		// 不需要任何操作，共享检查器会自动处理
		return nil
	}

	// 使用独立的健康检查器（向后兼容模式）
	// 性能警告：每个服务实例都会创建独立的健康检查器
	// 如果有大量服务（成千上万个），会导致以下问题：
	// 1. 每个服务都有一个独立的 healthCheckLoop goroutine（常驻运行）
	// 2. 每个检查周期会为每个节点启动一个 goroutine 进行健康检查
	// 3. 如果有 N 个服务，每个服务 M 个节点，常驻 goroutine 数量为 N，峰值 goroutine 数量为 N×M
	// 4. 大量 goroutine 会导致调度器压力、GC 压力和网络连接数爆炸
	s.healthChecker = NewHTTPHealthChecker(s.config.HealthCheck)

	// 注册节点健康状态回调
	s.healthChecker.RegisterCallback(s.onNodeHealthChanged)

	// 将所有启用的节点添加到健康检查器
	for _, node := range s.config.Nodes {
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
// 注意：此方法现在主要用于验证节点配置，节点状态通过节点的 Health 和 Enabled 字段维护
func (s *Service) initNodeStatus() {
	// 节点状态已经通过节点的 Health 和 Enabled 字段维护，不需要额外的初始化
	// 保留此方法是为了保持接口一致性，以及未来可能的扩展
}

// SelectNode 选择节点
func (s *Service) SelectNode(ctx *core.Context) (*NodeConfig, error) {

	// 使用负载均衡器选择节点（负载均衡器内部会过滤健康且启用的节点）
	selectedNode := s.loadBalancer.Select(s.config, ctx)

	// 更新统计信息
	isSuccess := selectedNode != nil
	s.updateStats(isSuccess, !isSuccess)

	if selectedNode == nil {
		return nil, ErrNoAvailableNode
	}

	return selectedNode, nil
}

// updateStats 更新服务统计信息（内部辅助方法）
// 注意：此方法需要写锁，调用前必须确保读锁已释放，使用 defer 确保锁一定会被释放
func (s *Service) updateStats(isSuccess, isFailure bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.stats.TotalRequests++
	if isSuccess {
		s.stats.SuccessRequests++
	} else if isFailure {
		s.stats.FailureRequests++
	}
	s.stats.LastRequestTime = time.Now()
	s.lastAccessTime = time.Now()
}

// AddNode 添加节点
func (s *Service) AddNode(node *NodeConfig) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if node == nil {
		return fmt.Errorf("node cannot be nil")
	}

	// 检查节点是否已存在
	for _, n := range s.config.Nodes {
		if n.ID == node.ID {
			return ErrNodeExists
		}
	}

	// 添加到节点列表（直接操作 config.Nodes）
	s.config.Nodes = append(s.config.Nodes, node)

	// 如果节点启用，添加到健康检查器（优先使用共享检查器）
	if node.Enabled {
		if s.useSharedChecker {
			// 使用共享检查器：不需要手动注册，检查器会自动从 ServiceManager 获取节点
			// 检查器会在下次检查周期自动发现新节点
		} else if s.healthChecker != nil {
			// 使用独立检查器（向后兼容）
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

	// 从节点列表中移除（直接操作 config.Nodes）
	found := false
	for i, node := range s.config.Nodes {
		if node.ID == nodeID {
			s.config.Nodes = append(s.config.Nodes[:i], s.config.Nodes[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return ErrNodeNotFound
	}

	// 从健康检查器中移除（优先使用共享检查器）
	if s.useSharedChecker {
		// 使用共享检查器：不需要手动注销，检查器会自动从 ServiceManager 获取节点
		// 如果节点被移除，检查器在下次检查时会自动发现节点不存在
	} else if s.healthChecker != nil {
		// 使用独立检查器（向后兼容）
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

	// 查找节点（在 config.Nodes 中查找）
	var node *NodeConfig
	for _, n := range s.config.Nodes {
		if n.ID == nodeID {
			node = n
			break
		}
	}

	if node == nil {
		return ErrNodeNotFound
	}

	// 更新健康状态（直接修改节点对象）
	node.Health = healthy

	return nil
}

// UpdateNodeStatus 更新节点启用状态
func (s *Service) UpdateNodeStatus(nodeID string, enabled bool) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 查找节点（在 config.Nodes 中查找）
	var node *NodeConfig
	for _, n := range s.config.Nodes {
		if n.ID == nodeID {
			node = n
			break
		}
	}

	if node == nil {
		return ErrNodeNotFound
	}

	// 更新启用状态（直接修改节点对象）
	node.Enabled = enabled

	// 根据启用状态更新健康检查器（优先使用共享检查器）
	if enabled {
		// 添加到健康检查器
		if s.useSharedChecker {
			// 使用共享检查器：不需要手动注册，检查器会自动从 ServiceManager 获取节点
			// 检查器会在下次检查周期自动发现新节点
		} else if s.healthChecker != nil {
			// 使用独立检查器（向后兼容）
			if httpChecker, ok := s.healthChecker.(*HTTPHealthChecker); ok {
				httpChecker.AddNode(node)
			} else if advancedChecker, ok := s.healthChecker.(*AdvancedHealthChecker); ok {
				advancedChecker.AddNode(node)
			}
		}
	} else {
		// 从健康检查器中移除
		if s.useSharedChecker {
			// 使用共享检查器：不需要手动注销，检查器会自动从 ServiceManager 获取节点
			// 如果节点被禁用，检查器在下次检查时会自动发现节点已禁用
		} else if s.healthChecker != nil {
			// 使用独立检查器（向后兼容）
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

	// 查找节点（在 config.Nodes 中查找）
	var node *NodeConfig
	for _, n := range s.config.Nodes {
		if n.ID == nodeID {
			node = n
			break
		}
	}

	if node == nil {
		return ErrNodeNotFound
	}

	// 更新权重（直接修改节点对象）
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
// 注意：此方法通过过滤 nodes 列表获取健康节点，实时计算，不需要维护额外的 map
func (s *Service) GetHealthyNodes() []*NodeConfig {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	nodes := make([]*NodeConfig, 0)
	for _, node := range s.config.Nodes {
		if node.Health && node.Enabled {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// GetUnhealthyNodes 获取不健康节点
// 注意：此方法通过过滤 nodes 列表获取不健康节点，实时计算，不需要维护额外的 map
func (s *Service) GetUnhealthyNodes() []*NodeConfig {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	nodes := make([]*NodeConfig, 0)
	for _, node := range s.config.Nodes {
		// 只返回启用但不健康的节点（禁用的节点不在此列）
		if node.Enabled && !node.Health {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// GetAllNodes 获取所有节点
func (s *Service) GetAllNodes() []*NodeConfig {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 返回节点的副本以避免并发问题
	nodes := make([]*NodeConfig, len(s.config.Nodes))
	copy(nodes, s.config.Nodes)
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
	stats["total_nodes"] = len(s.config.Nodes)

	// 实时统计健康/不健康节点数量
	healthyCount := 0
	unhealthyCount := 0
	for _, node := range s.config.Nodes {
		if node.Enabled {
			if node.Health {
				healthyCount++
			} else {
				unhealthyCount++
			}
		}
	}
	stats["healthy_nodes"] = healthyCount
	stats["unhealthy_nodes"] = unhealthyCount
	stats["total_requests"] = s.stats.TotalRequests
	stats["success_requests"] = s.stats.SuccessRequests
	stats["failure_requests"] = s.stats.FailureRequests
	stats["average_response_time_ms"] = s.stats.AverageResponseTime.Milliseconds()
	stats["max_response_time_ms"] = s.stats.MaxResponseTime.Milliseconds()
	stats["min_response_time_ms"] = s.stats.MinResponseTime.Milliseconds()
	stats["last_request_time"] = s.stats.LastRequestTime
	stats["last_access_time"] = s.lastAccessTime

	// TODO: 熔断器功能待实现
	// if s.circuitBreaker != nil {
	// 	stats["circuit_breaker_state"] = s.circuitBreaker.GetState("cb_service:" + s.config.ID)
	// }

	if s.loadBalancer != nil {
		stats["load_balancer_stats"] = s.loadBalancer.GetStats()
	}

	return stats
}

// Close 关闭服务相关资源
// 注意：此方法会停止健康检查器的goroutine，防止资源泄漏
// 必须确保在服务不再使用时调用此方法，否则健康检查goroutine会一直运行
// 返回最后一个错误（如果有多个错误，只返回最后一个）
func (s *Service) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var lastErr error

	// 停止健康检查器
	// 注意：如果使用共享检查器，不需要在这里停止（由 ServiceManager 统一管理）
	// 如果使用独立检查器，需要停止以防止 goroutine 泄漏
	if !s.useSharedChecker && s.healthChecker != nil {
		if err := s.healthChecker.Stop(); err != nil {
			lastErr = err
		}
		// 注意：不要将 healthChecker 设置为 nil，因为可能还有正在执行的回调
		// 但健康检查goroutine已经停止，不会再有新的检查
	}

	// TODO: 熔断器功能待实现
	// 清理熔断器资源（如果有需要的话）
	// if s.circuitBreaker != nil {
	// 	if cleaner, ok := s.circuitBreaker.(interface{ Close() error }); ok {
	// 		if err := cleaner.Close(); err != nil && lastErr == nil {
	// 			lastErr = err
	// 		}
	// 	}
	// }

	// 清理负载均衡器资源（重置内部状态）
	if s.loadBalancer != nil {
		s.loadBalancer.Reset()
	}

	// 注意：不清空 config.Nodes 列表，因为可能还有正在处理的请求引用这些节点
	// 节点状态通过节点的 Health 和 Enabled 字段维护，不需要额外的清理操作

	return lastErr
}

// onNodeHealthChanged 节点健康状态变化回调
// 注意：此方法仅用于独立健康检查器（向后兼容）
// 当使用共享健康检查器时，检查器会直接通过 ServiceManager 更新节点健康状态，不需要此回调
func (s *Service) onNodeHealthChanged(nodeID string, healthy bool) {
	_ = s.UpdateNodeHealth(nodeID, healthy)
}
