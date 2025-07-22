package service

import (
	"crypto/md5"
	"fmt"
	"sort"
	"sync"
	"time"

	"gateway/internal/gateway/core"
)

// ConsistentHashBalancer 一致性哈希负载均衡器
type ConsistentHashBalancer struct {
	*BaseLoadBalancer
	ring       map[uint32]string // hash -> nodeID
	sortedKeys []uint32
	replicas   int
	mu         sync.RWMutex
}

// NewConsistentHashBalancer 创建一致性哈希负载均衡器
func NewConsistentHashBalancer(config *LoadBalancerConfig) LoadBalancer {
	if config == nil {
		config = &DefaultConfig
	}

	return &ConsistentHashBalancer{
		BaseLoadBalancer: NewBaseLoadBalancer(config),
		ring:             make(map[uint32]string),
		replicas:         150, // 虚拟节点数量
	}
}

// Select 使用一致性哈希选择节点
func (c *ConsistentHashBalancer) Select(service *ServiceConfig, ctx *core.Context) *NodeConfig {
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

	// 构建哈希环
	c.buildRing(healthyNodes)

	// 获取哈希键
	key := c.getHashKey(ctx)
	hash := c.hashFunc(key)

	// 在环中查找节点
	nodeID := c.getNode(hash)
	if nodeID == "" {
		// 如果没找到，返回第一个节点
		return healthyNodes[0]
	}

	// 找到对应的节点
	for _, node := range healthyNodes {
		if node.ID == nodeID {
			return node
		}
	}

	return healthyNodes[0]
}

// GetStrategy 获取策略
func (c *ConsistentHashBalancer) GetStrategy() Strategy {
	return ConsistentHash
}

// UpdateNodeWeight 更新节点权重
func (c *ConsistentHashBalancer) UpdateNodeWeight(serviceID, nodeID string, weight int) error {
	// 一致性哈希会根据权重调整虚拟节点数量
	c.mu.Lock()
	defer c.mu.Unlock()
	// 清除环，下次选择时重新构建
	c.ring = make(map[uint32]string)
	c.sortedKeys = nil
	return nil
}

// GetStats 获取负载均衡统计信息
func (c *ConsistentHashBalancer) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["strategy"] = "consistent-hash"
	stats["replicas"] = c.replicas
	stats["ring_size"] = len(c.ring)
	return stats
}

// Reset 重置负载均衡器状态
func (c *ConsistentHashBalancer) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ring = make(map[uint32]string)
	c.sortedKeys = nil
}

// buildRing 构建哈希环
func (c *ConsistentHashBalancer) buildRing(nodes []*NodeConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.ring) > 0 {
		return // 已经构建过
	}

	for _, node := range nodes {
		weight := node.Weight
		if weight <= 0 {
			weight = 1
		}

		// 根据权重创建虚拟节点
		virtualNodes := c.replicas * weight
		for i := 0; i < virtualNodes; i++ {
			virtualKey := fmt.Sprintf("%s-%d", node.ID, i)
			hash := c.hashFunc(virtualKey)
			c.ring[hash] = node.ID
		}
	}

	// 排序哈希键
	c.sortedKeys = make([]uint32, 0, len(c.ring))
	for k := range c.ring {
		c.sortedKeys = append(c.sortedKeys, k)
	}
	sort.Slice(c.sortedKeys, func(i, j int) bool {
		return c.sortedKeys[i] < c.sortedKeys[j]
	})
}

// getNode 在环中查找节点
func (c *ConsistentHashBalancer) getNode(hash uint32) string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.sortedKeys) == 0 {
		return ""
	}

	// 二分查找第一个大于等于hash的节点
	idx := sort.Search(len(c.sortedKeys), func(i int) bool {
		return c.sortedKeys[i] >= hash
	})

	// 如果没找到，使用第一个节点（环形结构）
	if idx == len(c.sortedKeys) {
		idx = 0
	}

	return c.ring[c.sortedKeys[idx]]
}

// getHashKey 获取哈希键
func (c *ConsistentHashBalancer) getHashKey(ctx *core.Context) string {
	// 优先使用IP
	if ip := c.getClientIP(ctx); ip != "" {
		return ip
	}

	// 使用URL路径
	if path := ctx.Request.URL.Path; path != "" {
		return path
	}

	// 使用时间戳作为fallback
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// getClientIP 获取客户端IP
func (c *ConsistentHashBalancer) getClientIP(ctx *core.Context) string {
	// 优先从X-Forwarded-For获取
	if xff := ctx.Request.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	// 从X-Real-IP获取
	if xrip := ctx.Request.Header.Get("X-Real-IP"); xrip != "" {
		return xrip
	}

	// 从RemoteAddr获取
	return ctx.Request.RemoteAddr
}

// hashFunc 哈希函数
func (c *ConsistentHashBalancer) hashFunc(key string) uint32 {
	h := md5.Sum([]byte(key))
	return uint32(h[0])<<24 | uint32(h[1])<<16 | uint32(h[2])<<8 | uint32(h[3])
}
