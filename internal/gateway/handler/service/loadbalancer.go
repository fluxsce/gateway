package service

import (
	"fmt"
	"time"

	"gohub/internal/gateway/core"
	"gohub/internal/gateway/handler/circuitbreaker"
)

// Strategy 负载均衡策略
type Strategy string

const (
	// RoundRobin 轮询策略
	RoundRobin Strategy = "round-robin"
	// Random 随机策略
	Random Strategy = "random"
	// IPHash 基于IP的哈希策略
	IPHash Strategy = "ip-hash"
	// LeastConn 最少连接策略
	LeastConn Strategy = "least-conn"
	// WeightedRoundRobin 加权轮询策略
	WeightedRoundRobin Strategy = "weighted-round-robin"
	// ConsistentHash 一致性哈希策略
	ConsistentHash Strategy = "consistent-hash"
)

// NodeConfig 服务节点配置
type NodeConfig struct {
	ID       string            `yaml:"id" json:"id" mapstructure:"id"`                   // 节点ID
	URL      string            `yaml:"url" json:"url" mapstructure:"url"`                // 节点URL
	Weight   int               `yaml:"weight" json:"weight" mapstructure:"weight"`       // 权重
	Metadata map[string]string `yaml:"metadata" json:"metadata" mapstructure:"metadata"` // 元数据
	Health   bool              `yaml:"health" json:"health" mapstructure:"health"`       // 健康状态
	Enabled  bool              `yaml:"enabled" json:"enabled" mapstructure:"enabled"`    // 是否启用
}

// ServiceConfig 服务配置定义
type ServiceConfig struct {
	ID             string                               `yaml:"id" json:"id" mapstructure:"id"`                                                                      // 服务ID
	Name           string                               `yaml:"name" json:"name" mapstructure:"name"`                                                                // 服务名称
	Strategy       Strategy                             `yaml:"strategy" json:"strategy" mapstructure:"strategy"`                                                    // 负载均衡策略
	Nodes          []*NodeConfig                        `yaml:"nodes" json:"nodes" mapstructure:"nodes"`                                                             // 服务节点列表
	CircuitBreaker *circuitbreaker.CircuitBreakerConfig `yaml:"circuit_breaker,omitempty" json:"circuit_breaker,omitempty" mapstructure:"circuit_breaker,omitempty"` // 该服务的熔断器配置
	// 负载均衡配置
	LoadBalancer *LoadBalancerConfig `yaml:"load_balancer,omitempty" json:"load_balancer,omitempty" mapstructure:"load_balancer,omitempty"` // 该服务的负载均衡配置
	// 服务元数据
	ServiceMetadata map[string]string `yaml:"service_metadata,omitempty" json:"service_metadata,omitempty" mapstructure:"service_metadata,omitempty"` // 服务级别的元数据配置
}

// LoadBalancer 负载均衡器接口
type LoadBalancer interface {
	// Select 选择节点
	Select(service *ServiceConfig, ctx *core.Context) *NodeConfig

	// GetStrategy 获取策略
	GetStrategy() Strategy

	// UpdateNodeWeight 更新节点权重
	UpdateNodeWeight(serviceID, nodeID string, weight int) error

	// GetStats 获取负载均衡统计信息
	GetStats() map[string]interface{}

	// Reset 重置负载均衡器状态
	Reset()

	// GetConfig 获取配置
	GetConfig() *LoadBalancerConfig
}

// HealthChecker 健康检查器接口
type HealthChecker interface {
	// Start 启动健康检查
	Start() error

	// Stop 停止健康检查
	Stop() error

	// CheckNode 检查单个节点健康状态
	CheckNode(node *NodeConfig) bool

	// RegisterCallback 注册健康状态变化回调
	RegisterCallback(callback HealthCheckCallback)
}

// HealthCheckCallback 健康检查回调函数
type HealthCheckCallback func(nodeID string, healthy bool)

// LoadBalancerConfig 负载均衡器配置
type LoadBalancerConfig struct {
	ID              string        `yaml:"id" json:"id" mapstructure:"id"`                                           // 负载均衡器ID
	Strategy        Strategy      `yaml:"strategy" json:"strategy" mapstructure:"strategy"`                         // 负载均衡策略
	HealthCheck     *HealthConfig `yaml:"health_check" json:"health_check" mapstructure:"health_check"`             // 健康检查配置
	SessionAffinity bool          `yaml:"session_affinity" json:"session_affinity" mapstructure:"session_affinity"` // 会话亲和性
	StickySession   bool          `yaml:"sticky_session" json:"sticky_session" mapstructure:"sticky_session"`       // 粘性会话
	MaxRetries      int           `yaml:"max_retries" json:"max_retries" mapstructure:"max_retries"`                // 最大重试次数
	RetryTimeout    time.Duration `yaml:"retry_timeout" json:"retry_timeout" mapstructure:"retry_timeout"`          // 重试超时
	CircuitBreaker  bool          `yaml:"circuit_breaker" json:"circuit_breaker" mapstructure:"circuit_breaker"`    // 是否启用熔断器
}

// HealthConfig 健康检查配置
type HealthConfig struct {
	ID                  string            `yaml:"id" json:"id" mapstructure:"id"`                                                          // 健康检查配置ID
	Enabled             bool              `yaml:"enabled" json:"enabled" mapstructure:"enabled"`                                           // 是否启用健康检查
	Path                string            `yaml:"path" json:"path" mapstructure:"path"`                                                    // 健康检查路径
	Method              string            `yaml:"method" json:"method" mapstructure:"method"`                                              // 健康检查方法
	Interval            time.Duration     `yaml:"interval" json:"interval" mapstructure:"interval"`                                        // 检查间隔
	Timeout             time.Duration     `yaml:"timeout" json:"timeout" mapstructure:"timeout"`                                           // 检查超时
	HealthyThreshold    int               `yaml:"healthy_threshold" json:"healthy_threshold" mapstructure:"healthy_threshold"`             // 健康阈值
	UnhealthyThreshold  int               `yaml:"unhealthy_threshold" json:"unhealthy_threshold" mapstructure:"unhealthy_threshold"`       // 不健康阈值
	ExpectedStatusCodes []int             `yaml:"expected_status_codes" json:"expected_status_codes" mapstructure:"expected_status_codes"` // 期望的状态码
	Headers             map[string]string `yaml:"headers,omitempty" json:"headers,omitempty" mapstructure:"headers,omitempty"`             // 健康检查请求头
}

// 默认配置
var DefaultConfig = LoadBalancerConfig{
	ID:              "default-loadbalancer",
	Strategy:        RoundRobin,
	SessionAffinity: false,
	StickySession:   false,
	MaxRetries:      3,
	RetryTimeout:    5 * time.Second,
	CircuitBreaker:  false,
	HealthCheck: &HealthConfig{
		Enabled:             false,
		Path:                "/health",
		Method:              "GET",
		Interval:            30 * time.Second,
		Timeout:             5 * time.Second,
		HealthyThreshold:    2,
		UnhealthyThreshold:  3,
		ExpectedStatusCodes: []int{200},
	},
}

// 错误定义
var (
	ErrServiceNotFound = fmt.Errorf("service not found")
	ErrNodeNotFound    = fmt.Errorf("node not found")
	ErrNoAvailableNode = fmt.Errorf("no available node")
	ErrInvalidStrategy = fmt.Errorf("invalid load balance strategy")
	ErrServiceExists   = fmt.Errorf("service already exists")
	ErrNodeExists      = fmt.Errorf("node already exists")
)
