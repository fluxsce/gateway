// Package static 实现负载均衡器
// 负责根据不同的负载均衡算法选择最优节点
package static

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"gateway/internal/tunnel/types"
)

// LoadBalancer 负载均衡器接口
type LoadBalancer interface {
	// SelectNode 选择最优节点
	SelectNode(ctx context.Context, server *types.TunnelStaticServer) (*types.TunnelStaticNode, error)
	// GetAlgorithm 获取负载均衡算法名称
	GetAlgorithm() string
}

// simpleLoadBalancer 简单的负载均衡器实现（轮询算法）
type simpleLoadBalancer struct {
	currentIndex int
	mutex        sync.Mutex
}

// SelectNode 选择最优节点（轮询算法）
func (lb *simpleLoadBalancer) SelectNode(ctx context.Context, server *types.TunnelStaticServer) (*types.TunnelStaticNode, error) {
	if server.Nodes == nil || len(server.Nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}

	// 获取活跃且健康的节点
	activeNodes := make([]*types.TunnelStaticNode, 0)
	for _, node := range server.Nodes {
		if node.NodeStatus == "active" && node.HealthCheckStatus != nil && *node.HealthCheckStatus == "healthy" {
			activeNodes = append(activeNodes, node)
		}
	}

	if len(activeNodes) == 0 {
		return nil, fmt.Errorf("no active nodes available")
	}

	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	node := activeNodes[lb.currentIndex%len(activeNodes)]
	lb.currentIndex++

	return node, nil
}

// GetAlgorithm 获取负载均衡算法
func (lb *simpleLoadBalancer) GetAlgorithm() string {
	return "round_robin"
}

// leastConnLoadBalancer 最少连接数负载均衡器实现
// 选择当前连接数最少的节点
type leastConnLoadBalancer struct {
	mutex sync.Mutex
}

// SelectNode 选择最优节点（最少连接数算法）
func (lb *leastConnLoadBalancer) SelectNode(ctx context.Context, server *types.TunnelStaticServer) (*types.TunnelStaticNode, error) {
	if server.Nodes == nil || len(server.Nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}

	// 获取活跃且健康的节点
	activeNodes := make([]*types.TunnelStaticNode, 0)
	for _, node := range server.Nodes {
		if node.NodeStatus == "active" && node.HealthCheckStatus != nil && *node.HealthCheckStatus == "healthy" {
			activeNodes = append(activeNodes, node)
		}
	}

	if len(activeNodes) == 0 {
		return nil, fmt.Errorf("no active nodes available")
	}

	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	// 找到连接数最少的节点
	var selectedNode *types.TunnelStaticNode
	minConnections := int(^uint(0) >> 1) // 最大整数值

	// 收集所有连接数最少的节点（可能有多个）
	candidates := make([]*types.TunnelStaticNode, 0)

	for _, node := range activeNodes {
		connections := node.CurrentConnectionCount
		if connections < minConnections {
			minConnections = connections
			candidates = []*types.TunnelStaticNode{node} // 重置候选列表
			selectedNode = node
		} else if connections == minConnections {
			candidates = append(candidates, node)
		}
	}

	// 如果所有节点连接数相同，或者有多个节点连接数相同，随机选择一个
	if len(candidates) > 1 {
		selectedNode = candidates[rand.Intn(len(candidates))]
	}

	return selectedNode, nil
}

// GetAlgorithm 获取负载均衡算法
func (lb *leastConnLoadBalancer) GetAlgorithm() string {
	return "least_conn"
}

// randomLoadBalancer 随机负载均衡器实现
// 随机选择一个节点
type randomLoadBalancer struct {
	rand  *rand.Rand
	mutex sync.Mutex
}

// SelectNode 选择最优节点（随机算法）
func (lb *randomLoadBalancer) SelectNode(ctx context.Context, server *types.TunnelStaticServer) (*types.TunnelStaticNode, error) {
	if server.Nodes == nil || len(server.Nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}

	// 获取活跃且健康的节点
	activeNodes := make([]*types.TunnelStaticNode, 0)
	for _, node := range server.Nodes {
		if node.NodeStatus == "active" && node.HealthCheckStatus != nil && *node.HealthCheckStatus == "healthy" {
			activeNodes = append(activeNodes, node)
		}
	}

	if len(activeNodes) == 0 {
		return nil, fmt.Errorf("no active nodes available")
	}

	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	// 随机选择一个节点
	index := lb.rand.Intn(len(activeNodes))
	return activeNodes[index], nil
}

// GetAlgorithm 获取负载均衡算法
func (lb *randomLoadBalancer) GetAlgorithm() string {
	return "random"
}

// createLoadBalancer 根据算法类型创建负载均衡器
//
// 参数:
//   - algorithm: 负载均衡算法名称（round_robin, least_conn, random等）
//
// 返回:
//   - LoadBalancer: 负载均衡器实例
func createLoadBalancer(algorithm string) LoadBalancer {
	if algorithm == "" {
		algorithm = "round_robin" // 默认使用轮询算法
	}

	// 根据算法类型创建负载均衡器
	switch algorithm {
	case "round_robin", "roundrobin":
		return &simpleLoadBalancer{
			currentIndex: 0,
		}
	case "least_conn", "leastconn":
		return &leastConnLoadBalancer{}
	case "random":
		return &randomLoadBalancer{
			rand: rand.New(rand.NewSource(time.Now().UnixNano())),
		}
	default:
		// 默认使用轮询算法
		return &simpleLoadBalancer{
			currentIndex: 0,
		}
	}
}
