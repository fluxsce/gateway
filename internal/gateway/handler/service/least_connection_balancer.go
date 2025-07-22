package service

import (
	"sync"

	"gateway/internal/gateway/core"
)

// LeastConnectionBalancer 最少连接负载均衡器
type LeastConnectionBalancer struct {
	*BaseLoadBalancer
	connections map[string]int // nodeID -> connection count
	mu          sync.Mutex
}

// NewLeastConnBalancer 创建最少连接负载均衡器
func NewLeastConnBalancer(config *LoadBalancerConfig) LoadBalancer {
	if config == nil {
		config = &DefaultConfig
	}

	return &LeastConnectionBalancer{
		BaseLoadBalancer: NewBaseLoadBalancer(config),
		connections:      make(map[string]int),
	}
}

// Select 选择连接数最少的节点
func (l *LeastConnectionBalancer) Select(service *ServiceConfig, ctx *core.Context) *NodeConfig {
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

	l.mu.Lock()
	defer l.mu.Unlock()

	// 初始化连接数
	for _, node := range healthyNodes {
		if _, exists := l.connections[node.ID]; !exists {
			l.connections[node.ID] = 0
		}
	}

	// 找到连接数最少的节点
	var selectedNode *NodeConfig
	minConnections := -1

	for _, node := range healthyNodes {
		connections := l.connections[node.ID]
		if minConnections == -1 || connections < minConnections {
			minConnections = connections
			selectedNode = node
		}
	}

	// 增加选中节点的连接数
	if selectedNode != nil {
		l.connections[selectedNode.ID]++
	}

	return selectedNode
}

// GetStrategy 获取策略
func (l *LeastConnectionBalancer) GetStrategy() Strategy {
	return LeastConn
}

// UpdateNodeWeight 更新节点权重
func (l *LeastConnectionBalancer) UpdateNodeWeight(serviceID, nodeID string, weight int) error {
	// 最少连接策略不使用权重
	return nil
}

// GetStats 获取负载均衡统计信息
func (l *LeastConnectionBalancer) GetStats() map[string]interface{} {
	l.mu.Lock()
	defer l.mu.Unlock()

	stats := make(map[string]interface{})
	stats["strategy"] = "least-conn"
	stats["connections"] = make(map[string]int)
	for k, v := range l.connections {
		stats["connections"].(map[string]int)[k] = v
	}
	return stats
}

// Reset 重置负载均衡器状态
func (l *LeastConnectionBalancer) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.connections = make(map[string]int)
}

// ReleaseConnection 释放连接（在请求完成后调用）
func (l *LeastConnectionBalancer) ReleaseConnection(nodeID string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if count, exists := l.connections[nodeID]; exists && count > 0 {
		l.connections[nodeID]--
	}
}
