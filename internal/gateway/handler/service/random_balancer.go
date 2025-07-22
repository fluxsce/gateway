package service

import (
	"math/rand"
	"time"

	"gateway/internal/gateway/core"
)

// RandomBalancer 随机负载均衡器
type RandomBalancer struct {
	*BaseLoadBalancer
	rand *rand.Rand
}

// NewRandomBalancer 创建随机负载均衡器
func NewRandomBalancer(config *LoadBalancerConfig) LoadBalancer {
	if config == nil {
		config = &DefaultConfig
	}

	return &RandomBalancer{
		BaseLoadBalancer: NewBaseLoadBalancer(config),
		rand:             rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Select 选择节点
func (r *RandomBalancer) Select(service *ServiceConfig, ctx *core.Context) *NodeConfig {
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

	// 随机选择
	index := r.rand.Intn(len(healthyNodes))
	return healthyNodes[index]
}

// GetStrategy 获取策略
func (r *RandomBalancer) GetStrategy() Strategy {
	return Random
}

// UpdateNodeWeight 更新节点权重
func (r *RandomBalancer) UpdateNodeWeight(serviceID, nodeID string, weight int) error {
	// 随机策略不使用权重
	return nil
}

// GetStats 获取负载均衡统计信息
func (r *RandomBalancer) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"strategy": "random",
	}
}

// Reset 重置负载均衡器状态
func (r *RandomBalancer) Reset() {
	// 随机策略无状态
}
