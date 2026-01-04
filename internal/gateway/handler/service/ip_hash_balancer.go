package service

import (
	"crypto/md5"
	"time"

	"gateway/internal/gateway/core"
)

// IPHashBalancer IP哈希负载均衡器
type IPHashBalancer struct {
	*BaseLoadBalancer
}

// NewIPHashBalancer 创建IP哈希负载均衡器
func NewIPHashBalancer(config *LoadBalancerConfig) LoadBalancer {
	if config == nil {
		config = &DefaultConfig
	}

	return &IPHashBalancer{
		BaseLoadBalancer: NewBaseLoadBalancer(config),
	}
}

// Select 选择节点
func (i *IPHashBalancer) Select(service *ServiceConfig, ctx *core.Context) *NodeConfig {
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

	// 获取客户端IP
	ip := i.getClientIP(ctx)
	if ip == "" {
		// 如果无法获取IP，使用随机策略
		idx := time.Now().UnixNano() % int64(len(healthyNodes))
		return healthyNodes[idx]
	}

	// 计算哈希
	hash := i.hash(ip)
	return healthyNodes[hash%len(healthyNodes)]
}

// GetStrategy 获取策略
func (i *IPHashBalancer) GetStrategy() Strategy {
	return IPHash
}

// UpdateNodeWeight 更新节点权重
func (i *IPHashBalancer) UpdateNodeWeight(serviceID, nodeID string, weight int) error {
	// IP哈希策略不使用权重
	return nil
}

// GetStats 获取负载均衡统计信息
func (i *IPHashBalancer) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"strategy": "ip-hash",
	}
}

// Reset 重置负载均衡器状态
func (i *IPHashBalancer) Reset() {
	// IP哈希策略无状态
}

// getClientIP 获取客户端IP
// 注意：X-Forwarded-For 可能包含多个IP（用逗号分隔），这里只取第一个
// 如果无法获取IP，返回空字符串，调用方会使用fallback策略
func (i *IPHashBalancer) getClientIP(ctx *core.Context) string {
	// 优先从X-Forwarded-For获取（可能包含多个IP，用逗号分隔）
	if xff := ctx.Request.Header.Get("X-Forwarded-For"); xff != "" {
		// 取第一个IP（最原始的客户端IP）
		// X-Forwarded-For格式：client, proxy1, proxy2
		for idx := 0; idx < len(xff); idx++ {
			if xff[idx] == ',' {
				return xff[:idx]
			}
		}
		return xff
	}

	// 从X-Real-IP获取（通常只包含一个IP）
	if xrip := ctx.Request.Header.Get("X-Real-IP"); xrip != "" {
		return xrip
	}

	// 从RemoteAddr获取（格式：IP:Port，需要提取IP部分）
	remoteAddr := ctx.Request.RemoteAddr
	if remoteAddr != "" {
		// 提取IP部分（去除端口）
		for i := 0; i < len(remoteAddr); i++ {
			if remoteAddr[i] == ':' {
				return remoteAddr[:i]
			}
		}
		return remoteAddr
	}

	return ""
}

// hash 计算字符串哈希值
func (i *IPHashBalancer) hash(s string) int {
	h := md5.Sum([]byte(s))
	hash := 0
	for _, b := range h {
		hash = hash*31 + int(b)
	}
	if hash < 0 {
		hash = -hash
	}
	return hash
}
