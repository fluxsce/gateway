// Package server 提供负载均衡器的完整实现
// 负载均衡器负责在多个服务器节点间分配负载，支持多种负载均衡算法
package server

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"gateway/internal/tunnel/types"
	"gateway/pkg/logger"
)

// loadBalancer 负载均衡器实现
// 实现 LoadBalancer 接口，提供节点选择和负载分配功能
type loadBalancer struct {
	algorithm  string
	nodeStats  map[string]*nodeStatistics
	statsMutex sync.RWMutex
	roundRobin *roundRobinState
	random     *rand.Rand
}

// nodeStatistics 节点统计信息
type nodeStatistics struct {
	nodeID        string
	stats         *NodeStats
	lastUpdate    time.Time
	healthScore   float64
	weightedScore float64
	mutex         sync.RWMutex
}

// roundRobinState 轮询状态
type roundRobinState struct {
	currentIndex int
	mutex        sync.Mutex
}

// 负载均衡算法常量
const (
	AlgorithmRoundRobin     = "round_robin"
	AlgorithmWeightedRR     = "weighted_round_robin"
	AlgorithmLeastConn      = "least_connections"
	AlgorithmLeastLatency   = "least_latency"
	AlgorithmRandom         = "random"
	AlgorithmWeightedRandom = "weighted_random"
	AlgorithmHealthBased    = "health_based"
	AlgorithmIPHash         = "ip_hash"
)

// NewLoadBalancerImpl 创建新的负载均衡器实例
//
// 参数:
//   - algorithm: 负载均衡算法名称
//
// 返回:
//   - LoadBalancer: 负载均衡器接口实例
//
// 功能:
//   - 初始化负载均衡器
//   - 设置指定的负载均衡算法
//   - 创建节点统计映射表
//   - 初始化算法特定的状态
func NewLoadBalancerImpl(algorithm string) LoadBalancer {
	// 验证算法是否支持
	supportedAlgorithms := []string{
		AlgorithmRoundRobin, AlgorithmWeightedRR, AlgorithmLeastConn,
		AlgorithmLeastLatency, AlgorithmRandom, AlgorithmWeightedRandom,
		AlgorithmHealthBased, AlgorithmIPHash,
	}

	algorithmSupported := false
	for _, supported := range supportedAlgorithms {
		if algorithm == supported {
			algorithmSupported = true
			break
		}
	}

	if !algorithmSupported {
		logger.Warn("Unsupported load balancer algorithm, falling back to round_robin", map[string]interface{}{
			"algorithm": algorithm,
		})
		algorithm = AlgorithmRoundRobin
	}

	lb := &loadBalancer{
		algorithm: algorithm,
		nodeStats: make(map[string]*nodeStatistics),
		roundRobin: &roundRobinState{
			currentIndex: 0,
		},
		random: rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	logger.Info("Load balancer initialized", map[string]interface{}{
		"algorithm": algorithm,
	})

	return lb
}

// SelectNode 选择最优节点
//
// 参数:
//   - ctx: 上下文
//   - nodes: 可用节点列表
//
// 返回:
//   - *types.TunnelServerNode: 选择的节点
//   - error: 选择失败时返回错误
//
// 功能:
//   - 根据配置的负载均衡算法选择节点
//   - 过滤不健康的节点
//   - 考虑节点权重和性能指标
//   - 更新节点访问统计
func (lb *loadBalancer) SelectNode(ctx context.Context, nodes []*types.TunnelServerNode) (*types.TunnelServerNode, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}

	// 过滤健康的节点
	healthyNodes := lb.filterHealthyNodes(nodes)
	if len(healthyNodes) == 0 {
		logger.Warn("No healthy nodes available, using all nodes", nil)
		healthyNodes = nodes
	}

	// 根据算法选择节点
	var selectedNode *types.TunnelServerNode
	var err error

	switch lb.algorithm {
	case AlgorithmRoundRobin:
		selectedNode = lb.selectRoundRobin(healthyNodes)
	case AlgorithmWeightedRR:
		selectedNode = lb.selectWeightedRoundRobin(healthyNodes)
	case AlgorithmLeastConn:
		selectedNode = lb.selectLeastConnections(healthyNodes)
	case AlgorithmLeastLatency:
		selectedNode = lb.selectLeastLatency(healthyNodes)
	case AlgorithmRandom:
		selectedNode = lb.selectRandom(healthyNodes)
	case AlgorithmWeightedRandom:
		selectedNode = lb.selectWeightedRandom(healthyNodes)
	case AlgorithmHealthBased:
		selectedNode = lb.selectHealthBased(healthyNodes)
	case AlgorithmIPHash:
		selectedNode = lb.selectIPHash(ctx, healthyNodes)
	default:
		selectedNode = lb.selectRoundRobin(healthyNodes)
	}

	if selectedNode == nil {
		return nil, fmt.Errorf("failed to select node with algorithm %s", lb.algorithm)
	}

	// 更新节点访问统计
	lb.updateNodeAccess(selectedNode.ServerNodeId)

	logger.Debug("Node selected", map[string]interface{}{
		"nodeId":    selectedNode.ServerNodeId,
		"algorithm": lb.algorithm,
	})

	return selectedNode, err
}

// UpdateNodeStats 更新节点统计信息
//
// 参数:
//   - ctx: 上下文
//   - nodeID: 节点ID
//   - stats: 新的统计数据
//
// 返回:
//   - error: 更新失败时返回错误
//
// 功能:
//   - 更新节点的性能统计信息
//   - 计算健康评分和加权评分
//   - 用于后续的负载均衡决策
func (lb *loadBalancer) UpdateNodeStats(ctx context.Context, nodeID string, stats *NodeStats) error {
	lb.statsMutex.Lock()
	defer lb.statsMutex.Unlock()

	nodeStats, exists := lb.nodeStats[nodeID]
	if !exists {
		nodeStats = &nodeStatistics{
			nodeID:     nodeID,
			lastUpdate: time.Now(),
		}
		lb.nodeStats[nodeID] = nodeStats
	}

	nodeStats.mutex.Lock()
	nodeStats.stats = stats
	nodeStats.lastUpdate = time.Now()

	// 计算健康评分（0-100）
	healthScore := lb.calculateHealthScore(stats)
	nodeStats.healthScore = healthScore

	// 计算加权评分（综合考虑各种因素）
	weightedScore := lb.calculateWeightedScore(stats)
	nodeStats.weightedScore = weightedScore

	nodeStats.mutex.Unlock()

	logger.Debug("Node stats updated", map[string]interface{}{
		"nodeId":        nodeID,
		"healthScore":   healthScore,
		"weightedScore": weightedScore,
		"connections":   stats.ConnectionCount,
		"latency":       stats.AverageLatency,
		"errorRate":     stats.ErrorRate,
	})

	return nil
}

// GetAlgorithm 获取负载均衡算法
//
// 返回:
//   - string: 当前使用的负载均衡算法名称
func (lb *loadBalancer) GetAlgorithm() string {
	return lb.algorithm
}

// filterHealthyNodes 过滤健康的节点
func (lb *loadBalancer) filterHealthyNodes(nodes []*types.TunnelServerNode) []*types.TunnelServerNode {
	var healthyNodes []*types.TunnelServerNode

	for _, node := range nodes {
		if node.NodeStatus == types.NodeStatusActive {
			// 检查节点统计信息
			lb.statsMutex.RLock()
			nodeStats, exists := lb.nodeStats[node.ServerNodeId]
			lb.statsMutex.RUnlock()

			if !exists {
				// 新节点，认为是健康的
				healthyNodes = append(healthyNodes, node)
				continue
			}

			nodeStats.mutex.RLock()
			// 检查健康评分是否超过阈值（50分）
			if nodeStats.healthScore >= 50.0 {
				healthyNodes = append(healthyNodes, node)
			}
			nodeStats.mutex.RUnlock()
		}
	}

	return healthyNodes
}

// selectRoundRobin 轮询选择
func (lb *loadBalancer) selectRoundRobin(nodes []*types.TunnelServerNode) *types.TunnelServerNode {
	if len(nodes) == 0 {
		return nil
	}

	lb.roundRobin.mutex.Lock()
	defer lb.roundRobin.mutex.Unlock()

	if lb.roundRobin.currentIndex >= len(nodes) {
		lb.roundRobin.currentIndex = 0
	}

	selectedNode := nodes[lb.roundRobin.currentIndex]
	lb.roundRobin.currentIndex++

	return selectedNode
}

// selectWeightedRoundRobin 加权轮询选择
func (lb *loadBalancer) selectWeightedRoundRobin(nodes []*types.TunnelServerNode) *types.TunnelServerNode {
	if len(nodes) == 0 {
		return nil
	}

	// 创建加权节点列表
	var weightedNodes []*types.TunnelServerNode

	for _, node := range nodes {
		weight := lb.getNodeWeight(node.ServerNodeId)
		// 根据权重重复添加节点
		for i := 0; i < weight; i++ {
			weightedNodes = append(weightedNodes, node)
		}
	}

	if len(weightedNodes) == 0 {
		return nodes[0] // 回退到第一个节点
	}

	return lb.selectRoundRobin(weightedNodes)
}

// selectLeastConnections 最少连接选择
func (lb *loadBalancer) selectLeastConnections(nodes []*types.TunnelServerNode) *types.TunnelServerNode {
	if len(nodes) == 0 {
		return nil
	}

	var bestNode *types.TunnelServerNode
	minConnections := int(^uint(0) >> 1) // 最大整数

	for _, node := range nodes {
		connections := lb.getNodeConnections(node.ServerNodeId)
		if connections < minConnections {
			minConnections = connections
			bestNode = node
		}
	}

	return bestNode
}

// selectLeastLatency 最低延迟选择
func (lb *loadBalancer) selectLeastLatency(nodes []*types.TunnelServerNode) *types.TunnelServerNode {
	if len(nodes) == 0 {
		return nil
	}

	var bestNode *types.TunnelServerNode
	minLatency := float64(999999) // 很大的数

	for _, node := range nodes {
		latency := lb.getNodeLatency(node.ServerNodeId)
		if latency < minLatency {
			minLatency = latency
			bestNode = node
		}
	}

	return bestNode
}

// selectRandom 随机选择
func (lb *loadBalancer) selectRandom(nodes []*types.TunnelServerNode) *types.TunnelServerNode {
	if len(nodes) == 0 {
		return nil
	}

	index := lb.random.Intn(len(nodes))
	return nodes[index]
}

// selectWeightedRandom 加权随机选择
func (lb *loadBalancer) selectWeightedRandom(nodes []*types.TunnelServerNode) *types.TunnelServerNode {
	if len(nodes) == 0 {
		return nil
	}

	// 计算总权重
	totalWeight := 0
	weights := make([]int, len(nodes))

	for i, node := range nodes {
		weight := lb.getNodeWeight(node.ServerNodeId)
		weights[i] = weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return lb.selectRandom(nodes)
	}

	// 随机选择
	randomValue := lb.random.Intn(totalWeight)
	currentWeight := 0

	for i, weight := range weights {
		currentWeight += weight
		if randomValue < currentWeight {
			return nodes[i]
		}
	}

	return nodes[len(nodes)-1] // 回退
}

// selectHealthBased 基于健康状况选择
func (lb *loadBalancer) selectHealthBased(nodes []*types.TunnelServerNode) *types.TunnelServerNode {
	if len(nodes) == 0 {
		return nil
	}

	// 按健康评分排序
	type nodeWithScore struct {
		node  *types.TunnelServerNode
		score float64
	}

	var nodesWithScores []nodeWithScore

	for _, node := range nodes {
		score := lb.getNodeHealthScore(node.ServerNodeId)
		nodesWithScores = append(nodesWithScores, nodeWithScore{
			node:  node,
			score: score,
		})
	}

	// 按评分降序排序
	sort.Slice(nodesWithScores, func(i, j int) bool {
		return nodesWithScores[i].score > nodesWithScores[j].score
	})

	// 选择评分最高的节点
	return nodesWithScores[0].node
}

// selectIPHash IP哈希选择
func (lb *loadBalancer) selectIPHash(ctx context.Context, nodes []*types.TunnelServerNode) *types.TunnelServerNode {
	if len(nodes) == 0 {
		return nil
	}

	// 尝试从上下文中获取客户端IP
	clientIP := "unknown"
	if ip := ctx.Value("clientIP"); ip != nil {
		if ipStr, ok := ip.(string); ok {
			clientIP = ipStr
		}
	}

	// 计算哈希值
	hash := lb.hashString(clientIP)
	index := hash % len(nodes)

	return nodes[index]
}

// getNodeWeight 获取节点权重
func (lb *loadBalancer) getNodeWeight(nodeID string) int {
	lb.statsMutex.RLock()
	defer lb.statsMutex.RUnlock()

	nodeStats, exists := lb.nodeStats[nodeID]
	if !exists {
		return 1 // 默认权重
	}

	nodeStats.mutex.RLock()
	defer nodeStats.mutex.RUnlock()

	// 基于健康评分计算权重（1-10）
	weight := int(nodeStats.healthScore/10) + 1
	if weight > 10 {
		weight = 10
	}
	if weight < 1 {
		weight = 1
	}

	return weight
}

// getNodeConnections 获取节点连接数
func (lb *loadBalancer) getNodeConnections(nodeID string) int {
	lb.statsMutex.RLock()
	defer lb.statsMutex.RUnlock()

	nodeStats, exists := lb.nodeStats[nodeID]
	if !exists {
		return 0
	}

	nodeStats.mutex.RLock()
	defer nodeStats.mutex.RUnlock()

	if nodeStats.stats == nil {
		return 0
	}

	return nodeStats.stats.ConnectionCount
}

// getNodeLatency 获取节点延迟
func (lb *loadBalancer) getNodeLatency(nodeID string) float64 {
	lb.statsMutex.RLock()
	defer lb.statsMutex.RUnlock()

	nodeStats, exists := lb.nodeStats[nodeID]
	if !exists {
		return 1000.0 // 默认高延迟
	}

	nodeStats.mutex.RLock()
	defer nodeStats.mutex.RUnlock()

	if nodeStats.stats == nil {
		return 1000.0
	}

	return nodeStats.stats.AverageLatency
}

// getNodeHealthScore 获取节点健康评分
func (lb *loadBalancer) getNodeHealthScore(nodeID string) float64 {
	lb.statsMutex.RLock()
	defer lb.statsMutex.RUnlock()

	nodeStats, exists := lb.nodeStats[nodeID]
	if !exists {
		return 100.0 // 新节点默认满分
	}

	nodeStats.mutex.RLock()
	defer nodeStats.mutex.RUnlock()

	return nodeStats.healthScore
}

// calculateHealthScore 计算健康评分
func (lb *loadBalancer) calculateHealthScore(stats *NodeStats) float64 {
	score := 100.0

	// 错误率影响（错误率越高，分数越低）
	if stats.ErrorRate > 0 {
		score -= stats.ErrorRate * 10 // 每1%错误率扣10分
	}

	// CPU使用率影响
	if stats.CpuUsage > 80 {
		score -= (stats.CpuUsage - 80) * 2 // CPU超过80%每1%扣2分
	}

	// 内存使用率影响
	if stats.MemoryUsage > 80 {
		score -= (stats.MemoryUsage - 80) * 2 // 内存超过80%每1%扣2分
	}

	// 延迟影响
	if stats.AverageLatency > 100 {
		score -= (stats.AverageLatency - 100) / 10 // 延迟超过100ms每10ms扣1分
	}

	// 负载平均值影响
	if stats.LoadAverage > 2.0 {
		score -= (stats.LoadAverage - 2.0) * 10 // 负载超过2.0每0.1扣1分
	}

	// 确保分数在0-100范围内
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

// calculateWeightedScore 计算加权评分
func (lb *loadBalancer) calculateWeightedScore(stats *NodeStats) float64 {
	// 综合考虑各种因素的加权评分
	score := 0.0

	// 连接数权重（连接数越少越好）
	connectionScore := 100.0
	if stats.ConnectionCount > 0 {
		connectionScore = 100.0 / float64(stats.ConnectionCount+1)
	}
	score += connectionScore * 0.3

	// 延迟权重（延迟越低越好）
	latencyScore := 100.0
	if stats.AverageLatency > 0 {
		latencyScore = 1000.0 / (stats.AverageLatency + 10)
	}
	score += latencyScore * 0.3

	// 错误率权重（错误率越低越好）
	errorScore := 100.0 - stats.ErrorRate*10
	if errorScore < 0 {
		errorScore = 0
	}
	score += errorScore * 0.2

	// 资源使用率权重（使用率越低越好）
	resourceScore := 100.0 - (stats.CpuUsage+stats.MemoryUsage)/2
	if resourceScore < 0 {
		resourceScore = 0
	}
	score += resourceScore * 0.2

	return score
}

// updateNodeAccess 更新节点访问统计
func (lb *loadBalancer) updateNodeAccess(nodeID string) {
	lb.statsMutex.Lock()
	defer lb.statsMutex.Unlock()

	nodeStats, exists := lb.nodeStats[nodeID]
	if !exists {
		nodeStats = &nodeStatistics{
			nodeID:     nodeID,
			lastUpdate: time.Now(),
		}
		lb.nodeStats[nodeID] = nodeStats
	}

	nodeStats.mutex.Lock()
	nodeStats.lastUpdate = time.Now()
	nodeStats.mutex.Unlock()
}

// hashString 计算字符串哈希值
func (lb *loadBalancer) hashString(s string) int {
	hash := 0
	for _, c := range s {
		hash = hash*31 + int(c)
	}
	if hash < 0 {
		hash = -hash
	}
	return hash
}

// GetNodeStatistics 获取节点统计信息
func (lb *loadBalancer) GetNodeStatistics() map[string]interface{} {
	lb.statsMutex.RLock()
	defer lb.statsMutex.RUnlock()

	result := make(map[string]interface{})

	for nodeID, nodeStats := range lb.nodeStats {
		nodeStats.mutex.RLock()
		result[nodeID] = map[string]interface{}{
			"healthScore":   nodeStats.healthScore,
			"weightedScore": nodeStats.weightedScore,
			"lastUpdate":    nodeStats.lastUpdate,
		}
		if nodeStats.stats != nil {
			result[nodeID].(map[string]interface{})["stats"] = *nodeStats.stats
		}
		nodeStats.mutex.RUnlock()
	}

	return result
}

// SetAlgorithm 设置负载均衡算法
func (lb *loadBalancer) SetAlgorithm(algorithm string) error {
	supportedAlgorithms := []string{
		AlgorithmRoundRobin, AlgorithmWeightedRR, AlgorithmLeastConn,
		AlgorithmLeastLatency, AlgorithmRandom, AlgorithmWeightedRandom,
		AlgorithmHealthBased, AlgorithmIPHash,
	}

	algorithmSupported := false
	for _, supported := range supportedAlgorithms {
		if algorithm == supported {
			algorithmSupported = true
			break
		}
	}

	if !algorithmSupported {
		return fmt.Errorf("unsupported algorithm: %s", algorithm)
	}

	lb.algorithm = algorithm

	// 重置轮询状态
	lb.roundRobin.mutex.Lock()
	lb.roundRobin.currentIndex = 0
	lb.roundRobin.mutex.Unlock()

	logger.Info("Load balancer algorithm changed", map[string]interface{}{
		"algorithm": algorithm,
	})

	return nil
}
