package core

import (
	"context"
	"time"
)

// Manager 注册中心管理器接口
// 负责各组件的生命周期管理，包括初始化、启动和停止
type Manager interface {
	// Start 启动注册中心服务
	// 按照依赖顺序启动各个组件
	Start(ctx context.Context) error

	// Stop 停止注册中心服务
	// 优雅停止所有组件
	Stop(ctx context.Context) error

	// GetEventPublisher 获取事件发布器实例
	GetEventPublisher() EventPublisher

	// GetHealthChecker 获取健康检查器实例
	GetHealthChecker() HealthChecker

	// IsReady 检查服务是否就绪
	IsReady() bool

	// GetCache 获取缓存存储实例
	GetCache() CacheStorage
}

// CacheStorage 缓存存储接口
// 定义内存缓存操作，用于提高查询性能
type CacheStorage interface {
	// 服务组相关操作
	GetServiceGroup(ctx context.Context, tenantId, serviceGroupId string) (*ServiceGroup, error)
	SetServiceGroup(ctx context.Context, tenantId string, serviceGroup *ServiceGroup) error
	DeleteServiceGroup(ctx context.Context, tenantId, serviceGroupId string) error
	ListServiceGroups(ctx context.Context, tenantId string) ([]*ServiceGroup, error)

	// 服务相关操作
	GetService(ctx context.Context, tenantId, serviceGroupId, serviceName string) (*Service, error)
	SetService(ctx context.Context, tenantId string, service *Service) error
	DeleteService(ctx context.Context, tenantId, serviceGroupId, serviceName string) error
	ListServices(ctx context.Context, tenantId, serviceGroupId string) ([]*Service, error)

	// 实例相关操作
	GetInstance(ctx context.Context, tenantId, instanceId string) (*ServiceInstance, error)
	SetInstance(ctx context.Context, tenantId string, instance *ServiceInstance) error
	DeleteInstance(ctx context.Context, tenantId, instanceId string) error
	ListInstances(ctx context.Context, tenantId, serviceGroupId, serviceName string) ([]*ServiceInstance, error)

	//服务发现相关操作
	DiscoverInstance(ctx context.Context, tenantId, serviceGroupId, serviceName string) (*ServiceInstance, error)
	UpdateInstanceHealth(ctx context.Context, tenantId, instanceId string, status string, checkTime time.Time) error

	// 缓存管理操作
	GetStats() CacheStats
	Clear() error
}

// EventPublisher 事件发布器接口
// 负责发布服务相关的事件，支持异步处理和事件订阅
type EventPublisher interface {
	// Publish 发布事件
	// 异步发布事件到订阅者
	Publish(ctx context.Context, event *ServiceEvent) error

	// Subscribe 订阅事件
	// 注册事件订阅者，订阅者通过接口方法处理事件
	Subscribe(subscriber EventSubscriber) error

	// Unsubscribe 取消订阅
	// 取消事件订阅者
	Unsubscribe(subscriber EventSubscriber) error

	// Start 启动事件系统
	Start(ctx context.Context) error

	// Stop 停止事件系统
	Stop(ctx context.Context) error
}

// HealthChecker 健康检查器接口
// 负责定期检查服务实例的健康状态
type HealthChecker interface {
	// Start 启动健康检查
	Start(ctx context.Context) error

	// Stop 停止健康检查
	Stop(ctx context.Context) error

	// SetCheckInterval 设置健康检查间隔
	SetCheckInterval(interval time.Duration)

	// GetStats 获取健康检查统计信息
	GetStats() HealthCheckStats
}

// EventSubscriber 事件订阅者接口
// 定义事件订阅者的基本行为，统一包含关闭方法
type EventSubscriber interface {
	// HandleEvent 处理事件
	HandleEvent(ctx context.Context, event *ServiceEvent) error

	// GetEventTypes 获取订阅的事件类型
	// 返回nil或空数组表示订阅所有事件类型
	GetEventTypes() []string

	// GetSubscriberName 获取订阅者名称，用于日志和调试
	GetSubscriberName() string

	// Close 关闭订阅者，清理资源
	// 对于不需要资源清理的订阅者，可以返回nil
	Close() error
}

// InstanceFilter 实例过滤器
// 用于在查询实例时进行条件过滤
type InstanceFilter struct {
	InstanceStatus []string `json:"instanceStatus,omitempty"` // 实例状态过滤
	HealthStatus   []string `json:"healthStatus,omitempty"`   // 健康状态过滤
	ClientType     []string `json:"clientType,omitempty"`     // 客户端类型过滤
	Tags           []string `json:"tags,omitempty"`           // 标签过滤
	MetadataKeys   []string `json:"metadataKeys,omitempty"`   // 元数据键过滤
	HostAddress    string   `json:"hostAddress,omitempty"`    // 主机地址过滤（模糊匹配）
	MinWeight      *int     `json:"minWeight,omitempty"`      // 最小权重过滤
	MaxWeight      *int     `json:"maxWeight,omitempty"`      // 最大权重过滤
}

// ServiceFilter 服务过滤器
// 用于在查询服务时进行条件过滤
type ServiceFilter struct {
	ProtocolType        []string `json:"protocolType,omitempty"`        // 协议类型过滤
	LoadBalanceStrategy []string `json:"loadBalanceStrategy,omitempty"` // 负载均衡策略过滤
	GroupName           string   `json:"groupName,omitempty"`           // 分组名称过滤
	Tags                []string `json:"tags,omitempty"`                // 标签过滤
	MetadataKeys        []string `json:"metadataKeys,omitempty"`        // 元数据键过滤
}

// EventFilter 事件过滤器
// 用于事件订阅时的条件过滤
type EventFilter struct {
	EventTypes   []string `json:"eventTypes,omitempty"`   // 事件类型过滤
	ServiceNames []string `json:"serviceNames,omitempty"` // 服务名称过滤
	GroupNames   []string `json:"groupNames,omitempty"`   // 分组名称过滤
	TenantIds    []string `json:"tenantIds,omitempty"`    // 租户ID过滤
}

// CacheStats 缓存统计信息
type CacheStats struct {
	HitCount     int64   `json:"hitCount"`     // 命中次数
	MissCount    int64   `json:"missCount"`    // 未命中次数
	HitRate      float64 `json:"hitRate"`      // 命中率
	TotalEntries int64   `json:"totalEntries"` // 总条目数
	MemoryUsage  int64   `json:"memoryUsage"`  // 内存使用量（字节）
}

// HealthCheckStats 健康检查统计信息
type HealthCheckStats struct {
	TotalChecks     int64   `json:"totalChecks"`     // 总检查次数
	SuccessChecks   int64   `json:"successChecks"`   // 成功检查次数
	FailedChecks    int64   `json:"failedChecks"`    // 失败检查次数
	SuccessRate     float64 `json:"successRate"`     // 成功率
	AvgResponseTime int64   `json:"avgResponseTime"` // 平均响应时间（毫秒）
	ActiveInstances int64   `json:"activeInstances"` // 活跃实例数
}
