package service

import (
	"sync"

	"gateway/internal/gateway/core"
)

// WeightedRoundRobinBalancer 加权轮询负载均衡器
type WeightedRoundRobinBalancer struct {
	*BaseLoadBalancer
	weights map[string]int // nodeID -> current weight
	mu      sync.Mutex
}

// NewWeightedRoundRobinBalancer 创建加权轮询负载均衡器
func NewWeightedRoundRobinBalancer(config *LoadBalancerConfig) LoadBalancer {
	if config == nil {
		config = &DefaultConfig
	}

	return &WeightedRoundRobinBalancer{
		BaseLoadBalancer: NewBaseLoadBalancer(config),
		weights:          make(map[string]int),
	}
}

// Select 选择节点（加权轮询算法）
// 算法说明：
// 1. 每个节点维护一个当前权重（初始为0）
// 2. 每次选择时，所有节点的当前权重 += 节点权重
// 3. 选择当前权重最高的节点
// 4. 选中节点的当前权重 -= 总权重
// 这样可以保证权重高的节点被选中的频率更高，且分布均匀
func (w *WeightedRoundRobinBalancer) Select(service *ServiceConfig, ctx *core.Context) *NodeConfig {
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

	w.mu.Lock()
	defer w.mu.Unlock()

	// 初始化权重
	for _, node := range healthyNodes {
		if _, exists := w.weights[node.ID]; !exists {
			w.weights[node.ID] = 0
		}
	}

	// 计算总权重
	totalWeight := 0
	for _, node := range healthyNodes {
		weight := node.Weight
		if weight <= 0 {
			weight = 1
		}
		totalWeight += weight
	}

	// 增加当前权重
	for _, node := range healthyNodes {
		weight := node.Weight
		if weight <= 0 {
			weight = 1
		}
		w.weights[node.ID] += weight
	}

	// 选择权重最高的节点
	var selectedNode *NodeConfig
	maxWeight := 0
	for _, node := range healthyNodes {
		if w.weights[node.ID] > maxWeight {
			maxWeight = w.weights[node.ID]
			selectedNode = node
		}
	}

	// 减少选中节点的权重
	if selectedNode != nil {
		w.weights[selectedNode.ID] -= totalWeight
	}

	return selectedNode
}

// GetStrategy 获取策略
func (w *WeightedRoundRobinBalancer) GetStrategy() Strategy {
	return WeightedRoundRobin
}

// UpdateNodeWeight 更新节点权重
func (w *WeightedRoundRobinBalancer) UpdateNodeWeight(serviceID, nodeID string, weight int) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	// 重置权重，下次选择时重新计算
	delete(w.weights, nodeID)
	return nil
}

// GetStats 获取负载均衡统计信息
func (w *WeightedRoundRobinBalancer) GetStats() map[string]interface{} {
	w.mu.Lock()
	defer w.mu.Unlock()

	stats := make(map[string]interface{})
	stats["strategy"] = "weighted-round-robin"
	stats["current_weights"] = make(map[string]int)
	for k, v := range w.weights {
		stats["current_weights"].(map[string]int)[k] = v
	}
	return stats
}

// Reset 重置负载均衡器状态
func (w *WeightedRoundRobinBalancer) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.weights = make(map[string]int)
}
