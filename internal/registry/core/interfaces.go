package core

import (
	"context"
	"fmt"
	"time"
)

// ================== 存储接口 ==================

// Storage 存储接口 - 独立注册中心存储
type Storage interface {
	// 服务分组管理
	SaveServiceGroup(ctx context.Context, group *ServiceGroup) error
	GetServiceGroup(ctx context.Context, tenantId, groupName string) (*ServiceGroup, error)
	DeleteServiceGroup(ctx context.Context, tenantId, groupName string) error
	ListServiceGroups(ctx context.Context, tenantId string) ([]*ServiceGroup, error)

	// 服务管理
	SaveService(ctx context.Context, service *Service) error
	GetService(ctx context.Context, tenantId, serviceName string) (*Service, error)
	DeleteService(ctx context.Context, tenantId, serviceName string) error
	ListServices(ctx context.Context, tenantId, groupName string) ([]*Service, error)

	// 服务实例管理
	SaveInstance(ctx context.Context, instance *ServiceInstance) error
	GetInstance(ctx context.Context, tenantId, instanceId string) (*ServiceInstance, error)
	DeleteInstance(ctx context.Context, tenantId, instanceId string) error
	ListInstances(ctx context.Context, tenantId, serviceName, groupName string) ([]*ServiceInstance, error)
	ListAllInstances(ctx context.Context, tenantId string) ([]*ServiceInstance, error)

	// 实例状态管理
	UpdateHeartbeat(ctx context.Context, tenantId, instanceId string) error
	UpdateInstanceHealth(ctx context.Context, tenantId, instanceId string, healthStatus string) error
	UpdateInstanceStatus(ctx context.Context, tenantId, instanceId string, instanceStatus string) error

	// 服务发现
	GetServiceNames(ctx context.Context, tenantId, groupName string) ([]string, error)
	GetInstances(ctx context.Context, tenantId, serviceName, groupName string, filters ...InstanceFilter) ([]*ServiceInstance, error)

	// 事件日志
	LogEvent(ctx context.Context, event *ServiceEvent) error
	GetEvents(ctx context.Context, tenantId string, filters ...EventFilter) ([]*ServiceEvent, error)

	// 健康检查
	GetUnhealthyInstances(ctx context.Context, tenantId string, timeout time.Duration) ([]*ServiceInstance, error)

	// 统计信息
	GetStats(ctx context.Context, tenantId string) (*StorageStats, error)
}

// ExternalStorage 外部注册中心存储接口
type ExternalStorage interface {
	// 配置管理
	SaveExternalConfig(ctx context.Context, config *ExternalRegistryConfig) error
	GetExternalConfig(ctx context.Context, tenantId, configId string) (*ExternalRegistryConfig, error)
	DeleteExternalConfig(ctx context.Context, tenantId, configId string) error
	ListExternalConfigs(ctx context.Context, tenantId, registryType, environment string) ([]*ExternalRegistryConfig, error)

	// 状态管理
	SaveExternalStatus(ctx context.Context, status *ExternalRegistryStatus) error
	GetExternalStatus(ctx context.Context, tenantId, configId string) (*ExternalRegistryStatus, error)
	UpdateExternalStatus(ctx context.Context, tenantId, configId string, updates map[string]interface{}) error

	// 连接管理
	Connect(ctx context.Context, config *ExternalRegistryConfig) error
	Disconnect(ctx context.Context, configId string) error
	IsConnected(configId string) bool

	// 健康检查
	HealthCheck(ctx context.Context, configId string) error
	GetConnectionStatus(configId string) string

	// 服务发现（代理模式）
	DiscoverServices(ctx context.Context, configId string, filters ...ServiceFilter) ([]*UnifiedServiceInstance, error)
	DiscoverInstances(ctx context.Context, configId, serviceName string, filters ...InstanceFilter) ([]*UnifiedServiceInstance, error)

	// 监控
	GetMetrics(ctx context.Context, configId string) (*ExternalRegistryMetrics, error)
}

// ================== 服务注册中心接口 ==================

// Registry 服务注册中心接口
type Registry interface {
	// 启动
	Start() error

	// 服务实例管理
	Register(ctx context.Context, instance *ServiceInstance) error
	Deregister(ctx context.Context, tenantId, instanceId string) error
	Heartbeat(ctx context.Context, tenantId, instanceId string) error

	// 服务发现
	Discover(ctx context.Context, tenantId, serviceName, groupName string, filters ...InstanceFilter) ([]*ServiceInstance, error)
	GetInstance(ctx context.Context, tenantId, instanceId string) (*ServiceInstance, error)
	ListServices(ctx context.Context, tenantId, groupName string) ([]string, error)

	// 事件订阅
	Subscribe(ctx context.Context, tenantId, serviceName, groupName string) (<-chan *ServiceEvent, error)
	Unsubscribe(ctx context.Context, tenantId, serviceName, groupName string) error

	// 健康状态管理
	UpdateHealth(ctx context.Context, tenantId, instanceId string, healthStatus string) error

	// 关闭
	Close() error
}

// ================== 事件系统接口 ==================

// EventPublisher 事件发布器接口
type EventPublisher interface {
	// 启动
	Start() error

	// 发布事件
	Publish(ctx context.Context, event *ServiceEvent) error

	// 订阅事件
	Subscribe(ctx context.Context, tenantId, serviceName, groupName string) (<-chan *ServiceEvent, error)
	Unsubscribe(ctx context.Context, tenantId, serviceName, groupName string) error

	// 关闭
	Close() error
}

// ================== 健康检查接口 ==================

// HealthChecker 健康检查器接口
type HealthChecker interface {
	// 检查单个实例
	CheckInstance(ctx context.Context, instance *ServiceInstance) *HealthCheckResult

	// 批量检查实例
	CheckInstances(ctx context.Context, instances []*ServiceInstance) []*HealthCheckResult

	// 启动健康检查
	Start(ctx context.Context) error

	// 停止健康检查
	Stop() error

	// 添加实例到检查列表
	AddInstance(instance *ServiceInstance) error

	// 从检查列表移除实例
	RemoveInstance(instanceId string) error

	// 获取健康检查统计
	GetStats() *HealthCheckStats

	// 加载实例
	LoadInstances(ctx context.Context, tenantId string) error
}

// ================== 过滤器接口 ==================

// InstanceFilter 实例过滤器接口
type InstanceFilter interface {
	Filter(instances []*ServiceInstance) []*ServiceInstance
	Name() string
}

// ServiceFilter 服务过滤器接口
type ServiceFilter interface {
	Filter(services []*Service) []*Service
	Name() string
}

// EventFilter 事件过滤器接口
type EventFilter interface {
	Filter(events []*ServiceEvent) []*ServiceEvent
	Name() string
}

// ================== 管理器接口 ==================

// Manager 注册中心管理器接口
type Manager interface {
	// 初始化
	Initialize() error

	// 启动
	Start() error

	// 停止
	Stop() error

	// 获取注册中心实例
	GetRegistry() Registry

	// 获取存储实例
	GetStorage() Storage

	// 获取外部存储实例
	GetExternalStorage() ExternalStorage

	// 获取事件发布器
	GetEventPublisher() EventPublisher

	// 获取健康检查器
	GetHealthChecker() HealthChecker

	// 获取统计信息
	GetStats() *ManagerStats

	// 获取运行状态
	IsRunning() bool

	// 获取健康状态
	GetHealthStatus() map[string]interface{}
}

// ================== 数据结构 ==================

// UnifiedServiceInstance 统一服务实例模型（用于外部注册中心）
type UnifiedServiceInstance struct {
	// 基础信息
	ID          string `json:"id"`
	ServiceName string `json:"serviceName"`
	GroupName   string `json:"groupName,omitempty"`
	Namespace   string `json:"namespace,omitempty"`

	// 网络信息
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`

	// 状态信息
	Status string `json:"status"`
	Health string `json:"health"`
	Weight int    `json:"weight"`

	// 元数据
	Metadata map[string]string `json:"metadata"`
	Tags     []string          `json:"tags"`

	// 时间戳
	RegisterTime   time.Time `json:"registerTime"`
	LastUpdateTime time.Time `json:"lastUpdateTime"`

	// 注册中心特有字段
	RegistryType string      `json:"registryType"`
	OriginalData interface{} `json:"originalData,omitempty"`
}

// HealthCheckResult 健康检查结果
type HealthCheckResult struct {
	InstanceId    string        `json:"instanceId"`
	Status        string        `json:"status"`
	ResponseTime  time.Duration `json:"responseTime"`
	Error         error         `json:"error,omitempty"`
	CheckTime     time.Time     `json:"checkTime"`
	StatusChanged bool          `json:"statusChanged"`
}

// HealthCheckStats 健康检查统计
type HealthCheckStats struct {
	TotalInstances      int           `json:"totalInstances"`
	HealthyInstances    int           `json:"healthyInstances"`
	UnhealthyInstances  int           `json:"unhealthyInstances"`
	LastCheckTime       time.Time     `json:"lastCheckTime"`
	AverageResponseTime time.Duration `json:"averageResponseTime"`
	CheckCount          int64         `json:"checkCount"`
	ErrorCount          int64         `json:"errorCount"`
}

// StorageStats 存储统计
type StorageStats struct {
	TenantId       string    `json:"tenantId"`
	GroupCount     int       `json:"groupCount"`
	ServiceCount   int       `json:"serviceCount"`
	InstanceCount  int       `json:"instanceCount"`
	EventCount     int64     `json:"eventCount"`
	LastUpdateTime time.Time `json:"lastUpdateTime"`
}

// ManagerStats 管理器统计
type ManagerStats struct {
	Running       bool      `json:"running"`
	TenantId      string    `json:"tenantId"`
	Mode          string    `json:"mode"`
	StartTime     time.Time `json:"startTime"`
	ServiceCount  int       `json:"serviceCount"`
	InstanceCount int       `json:"instanceCount"`
	Services      []string  `json:"services"`
}

// ExternalRegistryMetrics 外部注册中心指标
type ExternalRegistryMetrics struct {
	ConfigId         string        `json:"configId"`
	RegistryType     string        `json:"registryType"`
	ConnectionStatus string        `json:"connectionStatus"`
	HealthStatus     string        `json:"healthStatus"`
	ResponseTime     time.Duration `json:"responseTime"`
	RequestCount     int64         `json:"requestCount"`
	SuccessCount     int64         `json:"successCount"`
	ErrorCount       int64         `json:"errorCount"`
	LastCheckTime    time.Time     `json:"lastCheckTime"`
}

// ================== 回调函数类型 ==================

// InstanceChangeCallback 实例变更回调
type InstanceChangeCallback func(event *ServiceEvent)

// ServiceChangeCallback 服务变更回调
type ServiceChangeCallback func(serviceName string, instances []*UnifiedServiceInstance)

// HealthChangeCallback 健康状态变更回调
type HealthChangeCallback func(instanceId string, oldStatus, newStatus string)

// ConfigChangeCallback 配置变更回调
type ConfigChangeCallback func(configId string, changeType string, config *ExternalRegistryConfig)

// ================== 错误定义 ==================

// 常见错误
var (
	ErrInstanceNotFound       = fmt.Errorf("service instance not found")
	ErrServiceNotFound        = fmt.Errorf("service not found")
	ErrGroupNotFound          = fmt.Errorf("service group not found")
	ErrConfigNotFound         = fmt.Errorf("external config not found")
	ErrInstanceExists         = fmt.Errorf("service instance already exists")
	ErrServiceExists          = fmt.Errorf("service already exists")
	ErrGroupExists            = fmt.Errorf("service group already exists")
	ErrInvalidParameter       = fmt.Errorf("invalid parameter")
	ErrConnectionFailed       = fmt.Errorf("connection failed")
	ErrNotConnected           = fmt.Errorf("not connected")
	ErrTooManySubscribers     = fmt.Errorf("too many subscribers")
	ErrRegistryNotRunning     = fmt.Errorf("registry not running")
	ErrHealthCheckFailed      = fmt.Errorf("health check failed")
	ErrConfigValidationFailed = fmt.Errorf("config validation failed")
)

// ================== 工具函数 ==================

// NewServiceEvent 创建服务事件
func NewServiceEvent(tenantId, eventType, serviceName, groupName, source, message string) *ServiceEvent {
	return &ServiceEvent{
		TenantId:     tenantId,
		EventType:    eventType,
		ServiceName:  serviceName,
		GroupName:    groupName,
		EventSource:  source,
		EventMessage: message,
		EventTime:    time.Now(),
		AddTime:      time.Now(),
		EditTime:     time.Now(),
		ActiveFlag:   FlagYes,
	}
}

// NewHealthCheckResult 创建健康检查结果
func NewHealthCheckResult(instanceId, status string, responseTime time.Duration, err error) *HealthCheckResult {
	return &HealthCheckResult{
		InstanceId:   instanceId,
		Status:       status,
		ResponseTime: responseTime,
		Error:        err,
		CheckTime:    time.Now(),
	}
}
