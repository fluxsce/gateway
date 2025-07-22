package service

import (
	"sync/atomic"

	"gateway/internal/gateway/core"
)

// RoundRobinBalancer 轮询负载均衡器
type RoundRobinBalancer struct {
	*BaseLoadBalancer
	counter uint64
}

// NewRoundRobinBalancer 创建轮询负载均衡器
func NewRoundRobinBalancer(config *LoadBalancerConfig) LoadBalancer {
	if config == nil {
		config = &DefaultConfig
	}

	return &RoundRobinBalancer{
		BaseLoadBalancer: NewBaseLoadBalancer(config),
		counter:          0,
	}
}

// Select 选择节点
func (r *RoundRobinBalancer) Select(service *ServiceConfig, ctx *core.Context) *NodeConfig {
	if len(service.Nodes) == 0 {
		return nil
	}

	// 过滤健康且启用的节点
	healthyNodes := make([]*NodeConfig, 0)
	for _, node := range service.Nodes {
		if node.Health && node.Enabled {
			healthyNodes = append(healthyNodes, node)
		}
	}

	if len(healthyNodes) == 0 {
		return nil
	}

	if len(healthyNodes) == 1 {
		return healthyNodes[0]
	}

	// 轮询选择
	index := atomic.AddUint64(&r.counter, 1) % uint64(len(healthyNodes))
	return healthyNodes[index]
}

// GetStrategy 获取策略
func (r *RoundRobinBalancer) GetStrategy() Strategy {
	return RoundRobin
}

// UpdateNodeWeight 更新节点权重
func (r *RoundRobinBalancer) UpdateNodeWeight(serviceID, nodeID string, weight int) error {
	// 轮询策略不使用权重
	return nil
}

// GetStats 获取负载均衡统计信息
func (r *RoundRobinBalancer) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"strategy": "round-robin",
		"counter":  atomic.LoadUint64(&r.counter),
	}
}

// Reset 重置负载均衡器状态
func (r *RoundRobinBalancer) Reset() {
	atomic.StoreUint64(&r.counter, 0)
}
