package core

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// ServiceGroup 服务分组信息
// 对应数据库表：HUB_REGISTRY_SERVICE_GROUP
type ServiceGroup struct {
	// 主键信息
	ServiceGroupId string `json:"serviceGroupId" db:"serviceGroupId"` // 服务分组ID，主键
	TenantId       string `json:"tenantId" db:"tenantId"`             // 租户ID，用于多租户数据隔离

	// 分组基本信息
	GroupName        string `json:"groupName" db:"groupName"`               // 分组名称
	GroupDescription string `json:"groupDescription" db:"groupDescription"` // 分组描述
	GroupType        string `json:"groupType" db:"groupType"`               // 分组类型

	// 授权信息
	OwnerUserId          string `json:"ownerUserId" db:"ownerUserId"`                   // 分组所有者用户ID
	AdminUserIds         string `json:"adminUserIds" db:"adminUserIds"`                 // 管理员用户ID列表，JSON格式
	ReadUserIds          string `json:"readUserIds" db:"readUserIds"`                   // 只读用户ID列表，JSON格式
	AccessControlEnabled string `json:"accessControlEnabled" db:"accessControlEnabled"` // 是否启用访问控制

	// 默认配置
	DefaultProtocolType               string `json:"defaultProtocolType" db:"defaultProtocolType"`                             // 默认协议类型
	DefaultLoadBalanceStrategy        string `json:"defaultLoadBalanceStrategy" db:"defaultLoadBalanceStrategy"`               // 默认负载均衡策略
	DefaultHealthCheckUrl             string `json:"defaultHealthCheckUrl" db:"defaultHealthCheckUrl"`                         // 默认健康检查URL
	DefaultHealthCheckIntervalSeconds int    `json:"defaultHealthCheckIntervalSeconds" db:"defaultHealthCheckIntervalSeconds"` // 默认健康检查间隔

	// 通用字段
	AddTime        time.Time `json:"addTime" db:"addTime"`               // 创建时间
	AddWho         string    `json:"addWho" db:"addWho"`                 // 创建人ID
	EditTime       time.Time `json:"editTime" db:"editTime"`             // 最后修改时间
	EditWho        string    `json:"editWho" db:"editWho"`               // 最后修改人ID
	OprSeqFlag     string    `json:"oprSeqFlag" db:"oprSeqFlag"`         // 操作序列标识
	CurrentVersion int       `json:"currentVersion" db:"currentVersion"` // 当前版本号
	ActiveFlag     string    `json:"activeFlag" db:"activeFlag"`         // 活动状态标记
	NoteText       string    `json:"noteText" db:"noteText"`             // 备注信息
	ExtProperty    string    `json:"extProperty" db:"extProperty"`       // 扩展属性

	// 预留字段
	Reserved1  string `json:"reserved1" db:"reserved1"`   // 预留字段1
	Reserved2  string `json:"reserved2" db:"reserved2"`   // 预留字段2
	Reserved3  string `json:"reserved3" db:"reserved3"`   // 预留字段3
	Reserved4  string `json:"reserved4" db:"reserved4"`   // 预留字段4
	Reserved5  string `json:"reserved5" db:"reserved5"`   // 预留字段5
	Reserved6  string `json:"reserved6" db:"reserved6"`   // 预留字段6
	Reserved7  string `json:"reserved7" db:"reserved7"`   // 预留字段7
	Reserved8  string `json:"reserved8" db:"reserved8"`   // 预留字段8
	Reserved9  string `json:"reserved9" db:"reserved9"`   // 预留字段9
	Reserved10 string `json:"reserved10" db:"reserved10"` // 预留字段10

	// 内存缓存专用字段，非数据库字段
	Services map[string]*Service `json:"services,omitempty" db:"-"` // 该服务组下的所有服务，key为serviceName
}

// Service 服务信息
// 对应数据库表：HUB_REGISTRY_SERVICE
type Service struct {
	// 主键信息
	TenantId    string `json:"tenantId" db:"tenantId"`       // 租户ID
	ServiceName string `json:"serviceName" db:"serviceName"` // 服务名称，主键

	// 关联分组信息
	ServiceGroupId string `json:"serviceGroupId" db:"serviceGroupId"` // 服务分组ID
	GroupName      string `json:"groupName" db:"groupName"`           // 分组名称（冗余字段）

	// 服务基本信息
	ServiceDescription string `json:"serviceDescription" db:"serviceDescription"` // 服务描述

	// 注册管理配置
	RegistryType           string `json:"registryType" db:"registryType"`                     // 注册类型(INTERNAL:内部管理,NACOS:Nacos注册中心,CONSUL:Consul,EUREKA:Eureka,ETCD:ETCD,ZOOKEEPER:ZooKeeper)
	ExternalRegistryConfig string `json:"externalRegistryConfig" db:"externalRegistryConfig"` // 外部注册中心配置，JSON格式，仅当registryType非INTERNAL时使用

	// 服务配置
	ProtocolType        string `json:"protocolType" db:"protocolType"`               // 协议类型
	ContextPath         string `json:"contextPath" db:"contextPath"`                 // 上下文路径
	LoadBalanceStrategy string `json:"loadBalanceStrategy" db:"loadBalanceStrategy"` // 负载均衡策略

	// 健康检查配置
	HealthCheckUrl             string `json:"healthCheckUrl" db:"healthCheckUrl"`                         // 健康检查URL
	HealthCheckIntervalSeconds int    `json:"healthCheckIntervalSeconds" db:"healthCheckIntervalSeconds"` // 健康检查间隔
	HealthCheckTimeoutSeconds  int    `json:"healthCheckTimeoutSeconds" db:"healthCheckTimeoutSeconds"`   // 健康检查超时
	HealthCheckType            string `json:"healthCheckType" db:"healthCheckType"`                       // 健康检查类型(HTTP,TCP)
	HealthCheckMode            string `json:"healthCheckMode" db:"healthCheckMode"`                       // 健康检查模式(ACTIVE:主动探测,PASSIVE:客户端上报)

	// 元数据和标签
	MetadataJson string `json:"metadataJson" db:"metadataJson"` // 服务元数据，JSON格式
	TagsJson     string `json:"tagsJson" db:"tagsJson"`         // 服务标签，JSON格式

	// 通用字段
	AddTime        time.Time `json:"addTime" db:"addTime"`               // 创建时间
	AddWho         string    `json:"addWho" db:"addWho"`                 // 创建人ID
	EditTime       time.Time `json:"editTime" db:"editTime"`             // 最后修改时间
	EditWho        string    `json:"editWho" db:"editWho"`               // 最后修改人ID
	OprSeqFlag     string    `json:"oprSeqFlag" db:"oprSeqFlag"`         // 操作序列标识
	CurrentVersion int       `json:"currentVersion" db:"currentVersion"` // 当前版本号
	ActiveFlag     string    `json:"activeFlag" db:"activeFlag"`         // 活动状态标记
	NoteText       string    `json:"noteText" db:"noteText"`             // 备注信息
	ExtProperty    string    `json:"extProperty" db:"extProperty"`       // 扩展属性

	// 预留字段
	Reserved1  string `json:"reserved1" db:"reserved1"`   // 预留字段1
	Reserved2  string `json:"reserved2" db:"reserved2"`   // 预留字段2
	Reserved3  string `json:"reserved3" db:"reserved3"`   // 预留字段3
	Reserved4  string `json:"reserved4" db:"reserved4"`   // 预留字段4
	Reserved5  string `json:"reserved5" db:"reserved5"`   // 预留字段5
	Reserved6  string `json:"reserved6" db:"reserved6"`   // 预留字段6
	Reserved7  string `json:"reserved7" db:"reserved7"`   // 预留字段7
	Reserved8  string `json:"reserved8" db:"reserved8"`   // 预留字段8
	Reserved9  string `json:"reserved9" db:"reserved9"`   // 预留字段9
	Reserved10 string `json:"reserved10" db:"reserved10"` // 预留字段10

	// 内存缓存专用字段，非数据库字段
	Instances []*ServiceInstance `json:"instances,omitempty" db:"-"` // 该服务下的所有实例
}

// ServiceInstance 服务实例信息
// 对应数据库表：HUB_REGISTRY_SERVICE_INSTANCE
type ServiceInstance struct {
	// 主键信息
	ServiceInstanceId string `json:"serviceInstanceId" db:"serviceInstanceId"` // 服务实例ID，主键
	TenantId          string `json:"tenantId" db:"tenantId"`                   // 租户ID

	// 关联信息
	ServiceGroupId string `json:"serviceGroupId" db:"serviceGroupId"` // 服务分组ID
	ServiceName    string `json:"serviceName" db:"serviceName"`       // 服务名称（冗余字段）
	GroupName      string `json:"groupName" db:"groupName"`           // 分组名称（冗余字段）

	// 网络连接信息
	HostAddress string `json:"hostAddress" db:"hostAddress"` // 主机地址
	PortNumber  int    `json:"portNumber" db:"portNumber"`   // 端口号
	ContextPath string `json:"contextPath" db:"contextPath"` // 上下文路径

	// 实例状态信息
	InstanceStatus string `json:"instanceStatus" db:"instanceStatus"` // 实例状态
	HealthStatus   string `json:"healthStatus" db:"healthStatus"`     // 健康状态

	// 负载均衡配置
	WeightValue int `json:"weightValue" db:"weightValue"` // 权重值

	// 客户端信息
	ClientId         string `json:"clientId" db:"clientId"`                 // 客户端ID
	ClientVersion    string `json:"clientVersion" db:"clientVersion"`       // 客户端版本
	ClientType       string `json:"clientType" db:"clientType"`             // 客户端类型
	TempInstanceFlag string `json:"tempInstanceFlag" db:"tempInstanceFlag"` // 临时实例标记(Y是临时实例,N否)

	// 健康检查统计
	HeartbeatFailCount int `json:"heartbeatFailCount" db:"heartbeatFailCount"` // 心跳检查失败次数，仅用于计数

	// 元数据和标签
	MetadataJson string `json:"metadataJson" db:"metadataJson"` // 实例元数据，JSON格式
	TagsJson     string `json:"tagsJson" db:"tagsJson"`         // 实例标签，JSON格式

	// 时间戳信息
	RegisterTime        time.Time  `json:"registerTime" db:"registerTime"`               // 注册时间
	LastHeartbeatTime   *time.Time `json:"lastHeartbeatTime" db:"lastHeartbeatTime"`     // 最后心跳时间
	LastHealthCheckTime *time.Time `json:"lastHealthCheckTime" db:"lastHealthCheckTime"` // 最后健康检查时间

	// 通用字段
	AddTime        time.Time `json:"addTime" db:"addTime"`               // 创建时间
	AddWho         string    `json:"addWho" db:"addWho"`                 // 创建人ID
	EditTime       time.Time `json:"editTime" db:"editTime"`             // 最后修改时间
	EditWho        string    `json:"editWho" db:"editWho"`               // 最后修改人ID
	OprSeqFlag     string    `json:"oprSeqFlag" db:"oprSeqFlag"`         // 操作序列标识
	CurrentVersion int       `json:"currentVersion" db:"currentVersion"` // 当前版本号
	ActiveFlag     string    `json:"activeFlag" db:"activeFlag"`         // 活动状态标记
	NoteText       string    `json:"noteText" db:"noteText"`             // 备注信息
	ExtProperty    string    `json:"extProperty" db:"extProperty"`       // 扩展属性

	// 预留字段
	Reserved1  string `json:"reserved1" db:"reserved1"`   // 预留字段1
	Reserved2  string `json:"reserved2" db:"reserved2"`   // 预留字段2
	Reserved3  string `json:"reserved3" db:"reserved3"`   // 预留字段3
	Reserved4  string `json:"reserved4" db:"reserved4"`   // 预留字段4
	Reserved5  string `json:"reserved5" db:"reserved5"`   // 预留字段5
	Reserved6  string `json:"reserved6" db:"reserved6"`   // 预留字段6
	Reserved7  string `json:"reserved7" db:"reserved7"`   // 预留字段7
	Reserved8  string `json:"reserved8" db:"reserved8"`   // 预留字段8
	Reserved9  string `json:"reserved9" db:"reserved9"`   // 预留字段9
	Reserved10 string `json:"reserved10" db:"reserved10"` // 预留字段10
}

// ServiceEvent 服务事件信息
// 对应数据库表：HUB_REGISTRY_SERVICE_EVENT
type ServiceEvent struct {
	// 主键信息
	ServiceEventId string `json:"serviceEventId" db:"serviceEventId"` // 服务事件ID，主键
	TenantId       string `json:"tenantId" db:"tenantId"`             // 租户ID

	// 关联信息
	ServiceGroupId    string `json:"serviceGroupId" db:"serviceGroupId"`       // 服务分组ID
	ServiceInstanceId string `json:"serviceInstanceId" db:"serviceInstanceId"` // 服务实例ID

	// 事件基本信息（冗余字段）
	GroupName     string `json:"groupName" db:"groupName"`         // 分组名称
	ServiceName   string `json:"serviceName" db:"serviceName"`     // 服务名称
	HostAddress   string `json:"hostAddress" db:"hostAddress"`     // 主机地址
	PortNumber    int    `json:"portNumber" db:"portNumber"`       // 端口号
	NodeIpAddress string `json:"nodeIpAddress" db:"nodeIpAddress"` // 节点IP地址，记录程序运行的IP

	EventType   string `json:"eventType" db:"eventType"`     // 事件类型
	EventSource string `json:"eventSource" db:"eventSource"` // 事件来源

	// 事件数据
	EventDataJson string `json:"eventDataJson" db:"eventDataJson"` // 事件数据，JSON格式
	EventMessage  string `json:"eventMessage" db:"eventMessage"`   // 事件消息描述

	// 时间信息
	EventTime time.Time `json:"eventTime" db:"eventTime"` // 事件发生时间

	// 数据传递专用字段（非数据库字段）
	Service  *Service         `json:"service,omitempty" db:"-"`  // 关联的服务对象，用于事件处理时的数据传递
	Instance *ServiceInstance `json:"instance,omitempty" db:"-"` // 关联的实例对象，用于事件处理时的数据传递

	// 通用字段
	AddTime        time.Time `json:"addTime" db:"addTime"`               // 创建时间
	AddWho         string    `json:"addWho" db:"addWho"`                 // 创建人ID
	EditTime       time.Time `json:"editTime" db:"editTime"`             // 最后修改时间
	EditWho        string    `json:"editWho" db:"editWho"`               // 最后修改人ID
	OprSeqFlag     string    `json:"oprSeqFlag" db:"oprSeqFlag"`         // 操作序列标识
	CurrentVersion int       `json:"currentVersion" db:"currentVersion"` // 当前版本号
	ActiveFlag     string    `json:"activeFlag" db:"activeFlag"`         // 活动状态标记
	NoteText       string    `json:"noteText" db:"noteText"`             // 备注信息
	ExtProperty    string    `json:"extProperty" db:"extProperty"`       // 扩展属性

	// 预留字段
	Reserved1  string `json:"reserved1" db:"reserved1"`   // 预留字段1
	Reserved2  string `json:"reserved2" db:"reserved2"`   // 预留字段2
	Reserved3  string `json:"reserved3" db:"reserved3"`   // 预留字段3
	Reserved4  string `json:"reserved4" db:"reserved4"`   // 预留字段4
	Reserved5  string `json:"reserved5" db:"reserved5"`   // 预留字段5
	Reserved6  string `json:"reserved6" db:"reserved6"`   // 预留字段6
	Reserved7  string `json:"reserved7" db:"reserved7"`   // 预留字段7
	Reserved8  string `json:"reserved8" db:"reserved8"`   // 预留字段8
	Reserved9  string `json:"reserved9" db:"reserved9"`   // 预留字段9
	Reserved10 string `json:"reserved10" db:"reserved10"` // 预留字段10
}

// 常量定义

// 实例状态常量
const (
	InstanceStatusUp           = "UP"             // 运行中
	InstanceStatusDown         = "DOWN"           // 停止
	InstanceStatusStarting     = "STARTING"       // 启动中
	InstanceStatusOutOfService = "OUT_OF_SERVICE" // 暂停服务
)

// 健康状态常量
const (
	HealthStatusHealthy   = "HEALTHY"   // 健康
	HealthStatusUnhealthy = "UNHEALTHY" // 不健康
	HealthStatusUnknown   = "UNKNOWN"   // 未知
)

// 客户端类型常量
const (
	ClientTypeService = "SERVICE" // 服务
	ClientTypeGateway = "GATEWAY" // 网关
	ClientTypeAdmin   = "ADMIN"   // 管理端
)

// 协议类型常量
const (
	ProtocolTypeHTTP  = "HTTP"  // HTTP协议
	ProtocolTypeHTTPS = "HTTPS" // HTTPS协议
	ProtocolTypeTCP   = "TCP"   // TCP协议
	ProtocolTypeUDP   = "UDP"   // UDP协议
	ProtocolTypeGRPC  = "GRPC"  // GRPC协议
)

// 负载均衡策略常量
const (
	LoadBalanceRoundRobin         = "ROUND_ROBIN"          // 轮询
	LoadBalanceWeightedRoundRobin = "WEIGHTED_ROUND_ROBIN" // 加权轮询
	LoadBalanceLeastConnections   = "LEAST_CONNECTIONS"    // 最少连接数
	LoadBalanceIpHash             = "IP_HASH"              // IP哈希
	LoadBalanceRandom             = "RANDOM"               // 随机
)

// 注册类型常量
const (
	RegistryTypeInternal  = "INTERNAL"  // 内部管理（默认）
	RegistryTypeNacos     = "NACOS"     // Nacos注册中心
	RegistryTypeConsul    = "CONSUL"    // Consul注册中心
	RegistryTypeEureka    = "EUREKA"    // Eureka注册中心
	RegistryTypeEtcd      = "ETCD"      // ETCD注册中心
	RegistryTypeZookeeper = "ZOOKEEPER" // ZooKeeper注册中心
)

// 临时实例标记常量
const (
	TempInstanceFlagYes = "Y" // 是临时实例
	TempInstanceFlagNo  = "N" // 不是临时实例
)

// 事件类型常量
const (
	// 分组相关事件
	EventTypeServiceGroupCreated = "SERVICE_GROUP_CREATED" // 服务组创建
	EventTypeServiceGroupUpdated = "SERVICE_GROUP_UPDATED" // 服务组更新
	EventTypeServiceGroupDeleted = "SERVICE_GROUP_DELETED" // 服务组删除

	// 服务相关事件
	EventTypeServiceRegistered   = "SERVICE_REGISTERED"   // 服务注册
	EventTypeServiceUpdated      = "SERVICE_UPDATED"      // 服务更新
	EventTypeServiceDeregistered = "SERVICE_DEREGISTERED" // 服务注销

	// 实例相关事件
	EventTypeInstanceRegistered       = "INSTANCE_REGISTERED"        // 实例注册
	EventTypeInstanceDeregistered     = "INSTANCE_DEREGISTERED"      // 实例注销
	EventTypeInstanceUpdated          = "INSTANCE_UPDATED"           // 实例更新
	EventTypeInstanceHeartbeatUpdated = "INSTANCE_HEARTBEAT_UPDATED" // 实例心跳更新
	EventTypeInstanceHealthChange     = "INSTANCE_HEALTH_CHANGE"     // 实例健康状态变更
	EventTypeInstanceStatusChange     = "INSTANCE_STATUS_CHANGE"     // 实例状态变更
)

// 事件源常量
const (
	EventSourceRegistryManager = "RegistryManager" // 注册中心管理器
	EventSourceHealthMonitor   = "HealthMonitor"   // 健康监控器
	EventSourceWebController   = "WebController"   // Web控制器
	EventSourceSDKService      = "SDKService"      // SDK服务
	EventSourceSystem          = "SYSTEM"          // 系统
	EventSourceClient          = "CLIENT"          // 客户端
	EventSourceScheduler       = "SCHEDULER"       // 调度器
	EventSourceDatabase        = "DATABASE"        // 数据库
)

// Context 键常量
const (
	ContextKeyEventSource = "EventSource" // 事件源上下文键
)

// GetEventSourceFromContext 从 context 中获取事件源，如果没有设置则返回默认值
func GetEventSourceFromContext(ctx context.Context, defaultEventSource string) string {
	if eventSource, ok := ctx.Value(ContextKeyEventSource).(string); ok && eventSource != "" {
		return eventSource
	}
	return defaultEventSource
}

// WithEventSource 在 context 中设置事件源
func WithEventSource(ctx context.Context, eventSource string) context.Context {
	return context.WithValue(ctx, ContextKeyEventSource, eventSource)
}

// GetValidInstanceStatuses 获取所有有效的实例状态
func GetValidInstanceStatuses() []string {
	return []string{
		InstanceStatusUp,
		InstanceStatusDown,
		InstanceStatusStarting,
		InstanceStatusOutOfService,
	}
}

// GetValidHealthStatuses 获取所有有效的健康状态
func GetValidHealthStatuses() []string {
	return []string{
		HealthStatusHealthy,
		HealthStatusUnhealthy,
		HealthStatusUnknown,
	}
}

// GetValidClientTypes 获取所有有效的客户端类型
func GetValidClientTypes() []string {
	return []string{
		ClientTypeService,
		ClientTypeGateway,
		ClientTypeAdmin,
	}
}

// GetValidTempInstanceFlags 获取所有有效的临时实例标记
func GetValidTempInstanceFlags() []string {
	return []string{
		TempInstanceFlagYes,
		TempInstanceFlagNo,
	}
}

// GetValidProtocolTypes 获取所有有效的协议类型
func GetValidProtocolTypes() []string {
	return []string{
		ProtocolTypeHTTP,
		ProtocolTypeHTTPS,
		ProtocolTypeTCP,
		ProtocolTypeUDP,
		ProtocolTypeGRPC,
	}
}

// GetValidLoadBalanceStrategies 获取所有有效的负载均衡策略
func GetValidLoadBalanceStrategies() []string {
	return []string{
		LoadBalanceRoundRobin,
		LoadBalanceWeightedRoundRobin,
		LoadBalanceLeastConnections,
		LoadBalanceIpHash,
		LoadBalanceRandom,
	}
}

// GetValidRegistryTypes 获取所有有效的注册类型
func GetValidRegistryTypes() []string {
	return []string{
		RegistryTypeInternal,
		RegistryTypeNacos,
		RegistryTypeConsul,
		RegistryTypeEureka,
		RegistryTypeEtcd,
		RegistryTypeZookeeper,
	}
}

// IsValidInstanceStatus 检查实例状态是否有效
func IsValidInstanceStatus(status string) bool {
	validStatuses := GetValidInstanceStatuses()
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

// IsValidHealthStatus 检查健康状态是否有效
func IsValidHealthStatus(status string) bool {
	validStatuses := GetValidHealthStatuses()
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

// IsValidClientType 检查客户端类型是否有效
func IsValidClientType(clientType string) bool {
	validTypes := GetValidClientTypes()
	for _, validType := range validTypes {
		if clientType == validType {
			return true
		}
	}
	return false
}

// IsValidTempInstanceFlag 检查临时实例标记是否有效
func IsValidTempInstanceFlag(flag string) bool {
	validFlags := GetValidTempInstanceFlags()
	for _, validFlag := range validFlags {
		if flag == validFlag {
			return true
		}
	}
	return false
}

// IsValidRegistryType 检查注册类型是否有效
func IsValidRegistryType(registryType string) bool {
	validTypes := GetValidRegistryTypes()
	for _, validType := range validTypes {
		if registryType == validType {
			return true
		}
	}
	return false
}

// IsExternalRegistry 判断服务是否使用外部注册中心
func (s *Service) IsExternalRegistry() bool {
	return s.RegistryType != RegistryTypeInternal
}

// IsInternalRegistry 判断服务是否使用内部注册中心
func (s *Service) IsInternalRegistry() bool {
	return s.RegistryType == RegistryTypeInternal
}

// IsHealthy 判断实例是否健康
func (si *ServiceInstance) IsHealthy() bool {
	return si.HealthStatus == HealthStatusHealthy && si.InstanceStatus == InstanceStatusUp
}

// IsAvailable 判断实例是否可用于服务发现
func (si *ServiceInstance) IsAvailable() bool {
	return si.IsHealthy() && si.ActiveFlag == "Y"
}

// GetFullAddress 获取实例的完整地址
func (si *ServiceInstance) GetFullAddress() string {
	portStr := fmt.Sprintf("%d", si.PortNumber)
	if si.ContextPath != "" && si.ContextPath != "/" {
		return si.HostAddress + ":" + portStr + si.ContextPath
	}
	return si.HostAddress + ":" + portStr
}

// CreateCacheKey 创建缓存键
func CreateInstanceCacheKey(tenantId, instanceId string) string {
	return "instance:" + tenantId + ":" + instanceId
}

// CreateInstanceListCacheKey 创建实例列表缓存键
func CreateInstanceListCacheKey(tenantId, serviceName, groupName string) string {
	return "instances:" + tenantId + ":" + groupName + ":" + serviceName
}

// CreateServiceCacheKey 创建服务缓存键
func CreateServiceCacheKey(tenantId, serviceName string) string {
	return "service:" + tenantId + ":" + serviceName
}

// =============================================================================
// 对象拷贝方法
// =============================================================================

// DeepCopy 深拷贝服务分组对象
// 通过JSON序列化/反序列化实现完整的深拷贝，包括嵌套的Services map
func (sg *ServiceGroup) DeepCopy() (*ServiceGroup, error) {
	data, err := json.Marshal(sg)
	if err != nil {
		return nil, fmt.Errorf("序列化ServiceGroup失败: %w", err)
	}

	var copy ServiceGroup
	if err := json.Unmarshal(data, &copy); err != nil {
		return nil, fmt.Errorf("反序列化ServiceGroup失败: %w", err)
	}

	return &copy, nil
}

// ShallowCopy 浅拷贝服务分组对象
// 复制所有基本字段，但Services map使用相同的引用
func (sg *ServiceGroup) ShallowCopy() *ServiceGroup {
	if sg == nil {
		return nil
	}

	copy := *sg
	return &copy
}

// DeepCopy 深拷贝服务对象
// 通过JSON序列化/反序列化实现完整的深拷贝，包括嵌套的Instances切片
func (s *Service) DeepCopy() (*Service, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return nil, fmt.Errorf("序列化Service失败: %w", err)
	}

	var copy Service
	if err := json.Unmarshal(data, &copy); err != nil {
		return nil, fmt.Errorf("反序列化Service失败: %w", err)
	}

	return &copy, nil
}

// ShallowCopy 浅拷贝服务对象
// 复制所有基本字段，但Instances切片使用相同的引用
func (s *Service) ShallowCopy() *Service {
	if s == nil {
		return nil
	}

	copy := *s
	return &copy
}

// DeepCopy 深拷贝服务实例对象
func (si *ServiceInstance) DeepCopy() (*ServiceInstance, error) {
	data, err := json.Marshal(si)
	if err != nil {
		return nil, fmt.Errorf("序列化ServiceInstance失败: %w", err)
	}

	var copy ServiceInstance
	if err := json.Unmarshal(data, &copy); err != nil {
		return nil, fmt.Errorf("反序列化ServiceInstance失败: %w", err)
	}

	return &copy, nil
}

// ShallowCopy 浅拷贝服务实例对象
func (si *ServiceInstance) ShallowCopy() *ServiceInstance {
	if si == nil {
		return nil
	}

	copy := *si
	return &copy
}

// DeepCopy 深拷贝服务事件对象
// 通过JSON序列化/反序列化实现完整的深拷贝，包括嵌套的Service和Instance对象
func (se *ServiceEvent) DeepCopy() (*ServiceEvent, error) {
	data, err := json.Marshal(se)
	if err != nil {
		return nil, fmt.Errorf("序列化ServiceEvent失败: %w", err)
	}

	var copy ServiceEvent
	if err := json.Unmarshal(data, &copy); err != nil {
		return nil, fmt.Errorf("反序列化ServiceEvent失败: %w", err)
	}

	return &copy, nil
}

// =============================================================================
// 使用说明和最佳实践
// =============================================================================

/*
拷贝方法使用指南：

1. 深拷贝 (DeepCopy)：
   - 使用JSON序列化/反序列化实现完整的深拷贝
   - 适用于需要完全独立的对象副本的场景
   - 包含所有嵌套对象和切片的完整拷贝
   - 性能开销相对较大，但数据完全隔离

   使用示例：
   originalService := &Service{...}
   copiedService, err := originalService.DeepCopy()
   if err != nil {
       // 处理错误
   }

2. 浅拷贝 (ShallowCopy)：
   - 复制结构体的所有基本字段
   - 嵌套对象和切片使用相同的内存引用
   - 性能开销小，适用于只需要修改基本字段的场景
   - 注意：修改嵌套对象会影响原对象

   使用示例：
   originalInstance := &ServiceInstance{...}
   copiedInstance := originalInstance.ShallowCopy()
*/
