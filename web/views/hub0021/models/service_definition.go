package models

import (
	"time"
)

// ServiceDefinition 服务定义模型，对应数据库HUB_GATEWAY_SERVICE_DEFINITION表
type ServiceDefinition struct {
	TenantId            string `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                                         // 租户ID，联合主键
	ServiceDefinitionId string `json:"serviceDefinitionId" form:"serviceDefinitionId" query:"serviceDefinitionId" db:"serviceDefinitionId"` // 服务定义ID，联合主键
	ServiceName         string `json:"serviceName" form:"serviceName" query:"serviceName" db:"serviceName"`                             // 服务名称
	ServiceDesc         string `json:"serviceDesc" form:"serviceDesc" query:"serviceDesc" db:"serviceDesc"`                             // 服务描述
	ServiceType         int    `json:"serviceType" form:"serviceType" query:"serviceType" db:"serviceType"`                             // 服务类型(0静态配置,1服务发现)

	// 代理配置关联字段
	ProxyConfigId string `json:"proxyConfigId" form:"proxyConfigId" query:"proxyConfigId" db:"proxyConfigId"` // 关联的代理配置ID

	// 根据ServiceConfig.Strategy字段设计负载均衡策略
	LoadBalanceStrategy string `json:"loadBalanceStrategy" form:"loadBalanceStrategy" query:"loadBalanceStrategy" db:"loadBalanceStrategy"` // 负载均衡策略(round-robin,random,ip-hash,least-conn,weighted-round-robin,consistent-hash)

	// 服务发现配置
	DiscoveryType   string `json:"discoveryType" form:"discoveryType" query:"discoveryType" db:"discoveryType"`         // 服务发现类型(CONSUL,EUREKA,NACOS等)
	DiscoveryConfig string `json:"discoveryConfig" form:"discoveryConfig" query:"discoveryConfig" db:"discoveryConfig"` // 服务发现配置,JSON格式

	// 根据LoadBalancerConfig结构设计负载均衡配置
	SessionAffinity string `json:"sessionAffinity" form:"sessionAffinity" query:"sessionAffinity" db:"sessionAffinity"`         // 是否启用会话亲和性(N否,Y是)
	StickySession   string `json:"stickySession" form:"stickySession" query:"stickySession" db:"stickySession"`                 // 是否启用粘性会话(N否,Y是)
	MaxRetries      int    `json:"maxRetries" form:"maxRetries" query:"maxRetries" db:"maxRetries"`                             // 最大重试次数
	RetryTimeoutMs  int    `json:"retryTimeoutMs" form:"retryTimeoutMs" query:"retryTimeoutMs" db:"retryTimeoutMs"`             // 重试超时时间(毫秒)
	EnableCircuitBreaker string `json:"enableCircuitBreaker" form:"enableCircuitBreaker" query:"enableCircuitBreaker" db:"enableCircuitBreaker"` // 是否启用熔断器(N否,Y是)

	// 根据HealthConfig结构设计健康检查配置
	HealthCheckEnabled        string `json:"healthCheckEnabled" form:"healthCheckEnabled" query:"healthCheckEnabled" db:"healthCheckEnabled"`                     // 是否启用健康检查(N否,Y是)
	HealthCheckPath           string `json:"healthCheckPath" form:"healthCheckPath" query:"healthCheckPath" db:"healthCheckPath"`                                 // 健康检查路径
	HealthCheckMethod         string `json:"healthCheckMethod" form:"healthCheckMethod" query:"healthCheckMethod" db:"healthCheckMethod"`                         // 健康检查方法
	HealthCheckIntervalSeconds int    `json:"healthCheckIntervalSeconds" form:"healthCheckIntervalSeconds" query:"healthCheckIntervalSeconds" db:"healthCheckIntervalSeconds"` // 健康检查间隔(秒)
	HealthCheckTimeoutMs      int    `json:"healthCheckTimeoutMs" form:"healthCheckTimeoutMs" query:"healthCheckTimeoutMs" db:"healthCheckTimeoutMs"`           // 健康检查超时(毫秒)
	HealthyThreshold          int    `json:"healthyThreshold" form:"healthyThreshold" query:"healthyThreshold" db:"healthyThreshold"`                             // 健康阈值
	UnhealthyThreshold        int    `json:"unhealthyThreshold" form:"unhealthyThreshold" query:"unhealthyThreshold" db:"unhealthyThreshold"`                     // 不健康阈值
	ExpectedStatusCodes       string `json:"expectedStatusCodes" form:"expectedStatusCodes" query:"expectedStatusCodes" db:"expectedStatusCodes"`                 // 期望的状态码,逗号分隔
	HealthCheckHeaders        string `json:"healthCheckHeaders" form:"healthCheckHeaders" query:"healthCheckHeaders" db:"healthCheckHeaders"`                     // 健康检查请求头,JSON格式

	// 负载均衡器配置(JSON格式存储完整的LoadBalancerConfig)
	LoadBalancerConfig string `json:"loadBalancerConfig" form:"loadBalancerConfig" query:"loadBalancerConfig" db:"loadBalancerConfig"` // 负载均衡器完整配置,JSON格式
	ServiceMetadata    string `json:"serviceMetadata" form:"serviceMetadata" query:"serviceMetadata" db:"serviceMetadata"`             // 服务元数据,JSON格式

	// 预留字段
	Reserved1 string     `json:"reserved1" form:"reserved1" query:"reserved1" db:"reserved1"` // 预留字段1
	Reserved2 string     `json:"reserved2" form:"reserved2" query:"reserved2" db:"reserved2"` // 预留字段2
	Reserved3 *int       `json:"reserved3" form:"reserved3" query:"reserved3" db:"reserved3"` // 预留字段3
	Reserved4 *int       `json:"reserved4" form:"reserved4" query:"reserved4" db:"reserved4"` // 预留字段4
	Reserved5 *time.Time `json:"reserved5" form:"reserved5" query:"reserved5" db:"reserved5"` // 预留字段5

	// 扩展属性
	ExtProperty string `json:"extProperty" form:"extProperty" query:"extProperty" db:"extProperty"` // 扩展属性,JSON格式

	// 标准字段
	AddTime        time.Time `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                             // 创建时间
	AddWho         string    `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                 // 创建人ID
	EditTime       time.Time `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                         // 最后修改时间
	EditWho        string    `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                             // 最后修改人ID
	OprSeqFlag     string    `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                 // 操作序列标识
	CurrentVersion int       `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"` // 当前版本号
	ActiveFlag     string    `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`                 // 活动状态标记(N非活动,Y活动)
	NoteText       string    `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`                         // 备注信息
}

// TableName 返回表名
func (ServiceDefinition) TableName() string {
	return "HUB_GATEWAY_SERVICE_DEFINITION"
}

// ServiceDefinitionWithProxy 服务定义和代理配置的组合VO，用于按实例查询时的返回
type ServiceDefinitionWithProxy struct {
	// 服务定义信息
	TenantId            string `json:"tenantId" db:"tenantId"`
	ServiceDefinitionId string `json:"serviceDefinitionId" db:"serviceDefinitionId"`
	ServiceName         string `json:"serviceName" db:"serviceName"`
	ServiceDesc         string `json:"serviceDesc" db:"serviceDesc"`
	ServiceType         int    `json:"serviceType" db:"serviceType"`
	LoadBalanceStrategy string `json:"loadBalanceStrategy" db:"loadBalanceStrategy"`
	DiscoveryType       string `json:"discoveryType" db:"discoveryType"`
	DiscoveryConfig     string `json:"discoveryConfig" db:"discoveryConfig"`
	SessionAffinity     string `json:"sessionAffinity" db:"sessionAffinity"`
	StickySession       string `json:"stickySession" db:"stickySession"`
	MaxRetries          int    `json:"maxRetries" db:"maxRetries"`
	RetryTimeoutMs      int    `json:"retryTimeoutMs" db:"retryTimeoutMs"`
	EnableCircuitBreaker string `json:"enableCircuitBreaker" db:"enableCircuitBreaker"`
	HealthCheckEnabled  string `json:"healthCheckEnabled" db:"healthCheckEnabled"`
	HealthCheckPath     string `json:"healthCheckPath" db:"healthCheckPath"`
	HealthCheckMethod   string `json:"healthCheckMethod" db:"healthCheckMethod"`
	HealthCheckIntervalSeconds int    `json:"healthCheckIntervalSeconds" db:"healthCheckIntervalSeconds"`
	HealthCheckTimeoutMs      int    `json:"healthCheckTimeoutMs" db:"healthCheckTimeoutMs"`
	HealthyThreshold          int    `json:"healthyThreshold" db:"healthyThreshold"`
	UnhealthyThreshold        int    `json:"unhealthyThreshold" db:"unhealthyThreshold"`
	ExpectedStatusCodes       string `json:"expectedStatusCodes" db:"expectedStatusCodes"`
	HealthCheckHeaders        string `json:"healthCheckHeaders" db:"healthCheckHeaders"`
	LoadBalancerConfig        string `json:"loadBalancerConfig" db:"loadBalancerConfig"`
	ServiceMetadata           string `json:"serviceMetadata" db:"serviceMetadata"`

	// 代理配置信息
	ProxyConfigId   string `json:"proxyConfigId" db:"proxyConfigId"`
	ProxyName       string `json:"proxyName" db:"proxyName"`
	ProxyType       string `json:"proxyType" db:"proxyType"`
	ProxyId         string `json:"proxyId" db:"proxyId"`
	ConfigPriority  int    `json:"configPriority" db:"configPriority"`
	ProxyConfig     string `json:"proxyConfig" db:"proxyConfig"`
	ProxyCustomConfig string `json:"proxyCustomConfig" db:"proxyCustomConfig"`

	// 网关实例信息
	GatewayInstanceId string `json:"gatewayInstanceId" db:"gatewayInstanceId"`
	InstanceName      string `json:"instanceName" db:"instanceName"`
	InstanceDesc      string `json:"instanceDesc" db:"instanceDesc"`

	// 标准字段
	AddTime        time.Time `json:"addTime" db:"addTime"`
	AddWho         string    `json:"addWho" db:"addWho"`
	EditTime       time.Time `json:"editTime" db:"editTime"`
	EditWho        string    `json:"editWho" db:"editWho"`
	CurrentVersion int       `json:"currentVersion" db:"currentVersion"`
	ActiveFlag     string    `json:"activeFlag" db:"activeFlag"`
	NoteText       string    `json:"noteText" db:"noteText"`
} 