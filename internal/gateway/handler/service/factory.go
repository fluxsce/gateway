package service

import (
	"fmt"
	"gohub/internal/gateway/core"
)

// LoadBalancerFactory 负载均衡器工厂
type LoadBalancerFactory struct{}

// NewLoadBalancerFactory 创建负载均衡器工厂
func NewLoadBalancerFactory() *LoadBalancerFactory {
	return &LoadBalancerFactory{}
}

// CreateLoadBalancer 根据配置创建负载均衡器
func (f *LoadBalancerFactory) CreateLoadBalancer(config *LoadBalancerConfig) (LoadBalancer, error) {
	if config == nil {
		config = &DefaultConfig
	}

	switch config.Strategy {
	case RoundRobin:
		return NewRoundRobinBalancer(config), nil
	case Random:
		return NewRandomBalancer(config), nil
	case IPHash:
		return NewIPHashBalancer(config), nil
	case LeastConn:
		return NewLeastConnBalancer(config), nil
	case WeightedRoundRobin:
		return NewWeightedRoundRobinBalancer(config), nil
	case ConsistentHash:
		return NewConsistentHashBalancer(config), nil
	default:
		return nil, ErrInvalidStrategy
	}
}

// CreateRoundRobinBalancer 创建轮询负载均衡器
func (f *LoadBalancerFactory) CreateRoundRobinBalancer(config *LoadBalancerConfig) LoadBalancer {
	return NewRoundRobinBalancer(config)
}

// CreateRandomBalancer 创建随机负载均衡器
func (f *LoadBalancerFactory) CreateRandomBalancer(config *LoadBalancerConfig) LoadBalancer {
	return NewRandomBalancer(config)
}

// CreateIPHashBalancer 创建IP哈希负载均衡器
func (f *LoadBalancerFactory) CreateIPHashBalancer(config *LoadBalancerConfig) LoadBalancer {
	return NewIPHashBalancer(config)
}

// CreateLeastConnBalancer 创建最少连接负载均衡器
func (f *LoadBalancerFactory) CreateLeastConnBalancer(config *LoadBalancerConfig) LoadBalancer {
	return NewLeastConnBalancer(config)
}

// CreateWeightedRoundRobinBalancer 创建加权轮询负载均衡器
func (f *LoadBalancerFactory) CreateWeightedRoundRobinBalancer(config *LoadBalancerConfig) LoadBalancer {
	return NewWeightedRoundRobinBalancer(config)
}

// CreateConsistentHashBalancer 创建一致性哈希负载均衡器
func (f *LoadBalancerFactory) CreateConsistentHashBalancer(config *LoadBalancerConfig) LoadBalancer {
	return NewConsistentHashBalancer(config)
}

// SelectNodeForService 为服务选择节点的辅助函数
func SelectNodeForService(service *ServiceConfig, strategy Strategy, ctx *core.Context) *NodeConfig {
	if len(service.Nodes) == 0 {
		return nil
	}

	// 过滤出健康且启用的节点
	var availableNodes []*NodeConfig
	for _, node := range service.Nodes {
		if node.Health && node.Enabled {
			availableNodes = append(availableNodes, node)
		}
	}

	if len(availableNodes) == 0 {
		return nil
	}

	// 根据策略选择节点
	switch strategy {
	case RoundRobin:
		return selectRoundRobin(availableNodes, service)
	case Random:
		return selectRandom(availableNodes)
	case IPHash:
		return selectIPHash(availableNodes, ctx)
	case WeightedRoundRobin:
		return selectWeightedRoundRobin(availableNodes)
	default:
		return availableNodes[0]
	}
}

// selectRoundRobin 轮询选择
func selectRoundRobin(nodes []*NodeConfig, service *ServiceConfig) *NodeConfig {
	// 这里简化实现，实际应该在服务级别维护轮询状态
	return nodes[0]
}

// selectRandom 随机选择
func selectRandom(nodes []*NodeConfig) *NodeConfig {
	if len(nodes) == 0 {
		return nil
	}
	// 简化实现，实际应该使用随机数
	return nodes[0]
}

// selectIPHash IP哈希选择
func selectIPHash(nodes []*NodeConfig, ctx *core.Context) *NodeConfig {
	if len(nodes) == 0 {
		return nil
	}
	// 简化实现，实际应该根据IP哈希
	return nodes[0]
}

// selectWeightedRoundRobin 加权轮询选择
func selectWeightedRoundRobin(nodes []*NodeConfig) *NodeConfig {
	if len(nodes) == 0 {
		return nil
	}

	// 简单实现：选择权重最高的节点
	maxWeight := 0
	var selectedNode *NodeConfig

	for _, node := range nodes {
		if node.Weight > maxWeight {
			maxWeight = node.Weight
			selectedNode = node
		}
	}

	if selectedNode != nil {
		return selectedNode
	}

	return nodes[0]
}

// CreateServiceManager 根据配置创建服务管理器
func (f *LoadBalancerFactory) CreateServiceManager() ServiceManager {
	return NewServiceManager()
}

// CreateManagerWithServices 创建带有服务的管理器
func (f *LoadBalancerFactory) CreateManagerWithServices(services []*ServiceConfig) (ServiceManager, error) {
	manager := NewServiceManager()

	// 添加服务
	for _, serviceConfig := range services {
		if err := manager.AddService(serviceConfig); err != nil {
			return nil, fmt.Errorf("添加服务 %s 失败: %w", serviceConfig.ID, err)
		}
	}

	return manager, nil
}

// CreateService 根据配置创建服务
func (f *LoadBalancerFactory) CreateService(config ServiceConfig) (*ServiceConfig, error) {
	if err := f.validateServiceConfig(&config); err != nil {
		return nil, fmt.Errorf("服务配置验证失败: %w", err)
	}

	// 应用默认配置
	if config.Strategy == "" {
		config.Strategy = RoundRobin
	}

	// 创建节点列表
	nodes := make([]*NodeConfig, 0, len(config.Nodes))
	for _, node := range config.Nodes {
		if err := f.validateNodeConfig(node); err != nil {
			return nil, fmt.Errorf("创建节点 %s 失败: %w", node.ID, err)
		}
		nodes = append(nodes, node)
	}

	service := &ServiceConfig{
		ID:       config.ID,
		Name:     config.Name,
		Strategy: config.Strategy,
		Nodes:    nodes,
	}

	return service, nil
}

// CreateNode 根据配置创建节点
func (f *LoadBalancerFactory) CreateNode(config NodeConfig) (*NodeConfig, error) {
	if err := f.validateNodeConfig(&config); err != nil {
		return nil, fmt.Errorf("创建节点 %s 失败: %w", config.ID, err)
	}

	// 应用默认配置
	if config.Weight <= 0 {
		config.Weight = 1
	}

	node := &NodeConfig{
		ID:       config.ID,
		URL:      config.URL,
		Weight:   config.Weight,
		Metadata: config.Metadata,
		Health:   config.Health,
		Enabled:  config.Enabled,
	}

	return node, nil
}

// CreateHealthChecker 根据配置创建健康检查器
func (f *LoadBalancerFactory) CreateHealthChecker(config *HealthConfig) HealthChecker {
	if config == nil || !config.Enabled {
		return NewNoOpHealthChecker()
	}

	return NewHTTPHealthChecker(config)
}

// GetSupportedStrategies 获取支持的负载均衡策略
func (f *LoadBalancerFactory) GetSupportedStrategies() []Strategy {
	return []Strategy{
		RoundRobin,
		Random,
		IPHash,
		LeastConn,
		WeightedRoundRobin,
		ConsistentHash,
	}
}

// GetStrategyDescription 获取策略描述
func (f *LoadBalancerFactory) GetStrategyDescription(strategy Strategy) string {
	descriptions := map[Strategy]string{
		RoundRobin:         "轮询策略 - 按顺序依次分配请求到各个节点",
		Random:             "随机策略 - 随机选择节点处理请求",
		IPHash:             "IP哈希策略 - 根据客户端IP哈希选择节点，保证同一IP的请求路由到同一节点",
		LeastConn:          "最少连接策略 - 选择当前连接数最少的节点",
		WeightedRoundRobin: "加权轮询策略 - 根据节点权重按比例分配请求",
		ConsistentHash:     "一致性哈希策略 - 使用一致性哈希算法选择节点，适用于缓存场景",
	}

	if desc, ok := descriptions[strategy]; ok {
		return desc
	}
	return "未知策略"
}

// validateServiceConfig 验证服务配置
func (f *LoadBalancerFactory) validateServiceConfig(config *ServiceConfig) error {
	if config.ID == "" {
		return fmt.Errorf("服务ID不能为空")
	}
	if config.Name == "" {
		return fmt.Errorf("服务名称不能为空")
	}
	if len(config.Nodes) == 0 {
		return fmt.Errorf("服务必须至少包含一个节点")
	}

	// 验证策略
	validStrategies := f.GetSupportedStrategies()
	validStrategy := false
	for _, strategy := range validStrategies {
		if config.Strategy == strategy {
			validStrategy = true
			break
		}
	}
	if !validStrategy {
		return fmt.Errorf("不支持的负载均衡策略: %s", config.Strategy)
	}

	return nil
}

// validateNodeConfig 验证节点配置
func (f *LoadBalancerFactory) validateNodeConfig(config *NodeConfig) error {
	if config.ID == "" {
		return fmt.Errorf("节点ID不能为空")
	}
	if config.URL == "" {
		return fmt.Errorf("节点URL不能为空")
	}
	if config.Weight < 0 {
		return fmt.Errorf("节点权重不能为负数")
	}

	return nil
}
