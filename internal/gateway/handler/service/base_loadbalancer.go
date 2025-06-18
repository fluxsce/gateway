package service

// BaseLoadBalancer 基础负载均衡器
type BaseLoadBalancer struct {
	config *LoadBalancerConfig
}

// NewBaseLoadBalancer 创建基础负载均衡器
func NewBaseLoadBalancer(config *LoadBalancerConfig) *BaseLoadBalancer {
	if config == nil {
		config = &DefaultConfig
	}
	return &BaseLoadBalancer{
		config: config,
	}
}

// GetConfig 获取配置
func (b *BaseLoadBalancer) GetConfig() *LoadBalancerConfig {
	return b.config
}
