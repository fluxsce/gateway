package models

import (
	"time"
)

// ProxyConfig 代理配置模型，对应数据库HUB_GW_PROXY_CONFIG表
type ProxyConfig struct {
	TenantId          string `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                                     // 租户ID，联合主键
	ProxyConfigId     string `json:"proxyConfigId" form:"proxyConfigId" query:"proxyConfigId" db:"proxyConfigId"`                 // 代理配置ID，联合主键
	GatewayInstanceId string `json:"gatewayInstanceId" form:"gatewayInstanceId" query:"gatewayInstanceId" db:"gatewayInstanceId"` // 网关实例ID(代理配置仅支持实例级)
	ProxyName         string `json:"proxyName" form:"proxyName" query:"proxyName" db:"proxyName"`                                 // 代理名称

	// 根据ProxyType枚举值设计
	ProxyType string `json:"proxyType" form:"proxyType" query:"proxyType" db:"proxyType"` // 代理类型(http,websocket,tcp,udp)

	// 基础配置
	ProxyId         string `json:"proxyId" form:"proxyId" query:"proxyId" db:"proxyId"`                         // 代理ID(来自ProxyConfig.ID)
	ConfigPriority  int    `json:"configPriority" form:"configPriority" query:"configPriority" db:"configPriority"` // 配置优先级,数值越小优先级越高

	// 通用配置，JSON格式存储不同类型的具体配置
	ProxyConfig  string `json:"proxyConfig" form:"proxyConfig" query:"proxyConfig" db:"proxyConfig"`     // 代理具体配置,JSON格式,根据proxyType存储对应配置
	CustomConfig string `json:"customConfig" form:"customConfig" query:"customConfig" db:"customConfig"` // 自定义配置,JSON格式

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
	ActiveFlag     string    `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`                 // 活动状态标记(N非活动/禁用,Y活动/启用)
	NoteText       string    `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`                         // 备注信息
}

// TableName 返回表名
func (ProxyConfig) TableName() string {
	return "HUB_GW_PROXY_CONFIG"
}

// ServiceDefinition 服务定义模型，对应数据库HUB_GW_SERVICE_DEFINITION表
type ServiceDefinition struct {
	TenantId            string `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                                         // 租户ID，联合主键
	ServiceDefinitionId string `json:"serviceDefinitionId" form:"serviceDefinitionId" query:"serviceDefinitionId" db:"serviceDefinitionId"` // 服务定义ID，联合主键
	ServiceName         string `json:"serviceName" form:"serviceName" query:"serviceName" db:"serviceName"`                             // 服务名称
	ServiceDesc         string `json:"serviceDesc" form:"serviceDesc" query:"serviceDesc" db:"serviceDesc"`                             // 服务描述
	ServiceType         int    `json:"serviceType" form:"serviceType" query:"serviceType" db:"serviceType"`                             // 服务类型(0静态配置,1服务发现)
	ProxyConfigId     string `json:"proxyConfigId" form:"proxyConfigId" query:"proxyConfigId" db:"proxyConfigId"`                 // 代理配置ID，联合主键
	
	// 根据ServiceConfig.Strategy字段设计负载均衡策略
	LoadBalanceStrategy string `json:"loadBalanceStrategy" form:"loadBalanceStrategy" query:"loadBalanceStrategy" db:"loadBalanceStrategy"` // 负载均衡策略(round-robin,random,ip-hash,least-conn,weighted-round-robin,consistent-hash)

	// 服务发现配置
	DiscoveryType   string `json:"discoveryType" form:"discoveryType" query:"discoveryType" db:"discoveryType"`     // 服务发现类型(CONSUL,EUREKA,NACOS等)
	DiscoveryConfig string `json:"discoveryConfig" form:"discoveryConfig" query:"discoveryConfig" db:"discoveryConfig"` // 服务发现配置,JSON格式

	// 根据LoadBalancerConfig结构设计负载均衡配置
	SessionAffinity     string `json:"sessionAffinity" form:"sessionAffinity" query:"sessionAffinity" db:"sessionAffinity"`         // 是否启用会话亲和性(N否,Y是)
	StickySession       string `json:"stickySession" form:"stickySession" query:"stickySession" db:"stickySession"`                 // 是否启用粘性会话(N否,Y是)
	MaxRetries          int    `json:"maxRetries" form:"maxRetries" query:"maxRetries" db:"maxRetries"`                             // 最大重试次数
	RetryTimeoutMs      int    `json:"retryTimeoutMs" form:"retryTimeoutMs" query:"retryTimeoutMs" db:"retryTimeoutMs"`             // 重试超时时间(毫秒)
	EnableCircuitBreaker string `json:"enableCircuitBreaker" form:"enableCircuitBreaker" query:"enableCircuitBreaker" db:"enableCircuitBreaker"` // 是否启用熔断器(N否,Y是)

	// 根据HealthConfig结构设计健康检查配置
	HealthCheckEnabled        string `json:"healthCheckEnabled" form:"healthCheckEnabled" query:"healthCheckEnabled" db:"healthCheckEnabled"`                 // 是否启用健康检查(N否,Y是)
	HealthCheckPath           string `json:"healthCheckPath" form:"healthCheckPath" query:"healthCheckPath" db:"healthCheckPath"`                             // 健康检查路径
	HealthCheckMethod         string `json:"healthCheckMethod" form:"healthCheckMethod" query:"healthCheckMethod" db:"healthCheckMethod"`                     // 健康检查方法
	HealthCheckIntervalSeconds *int   `json:"healthCheckIntervalSeconds" form:"healthCheckIntervalSeconds" query:"healthCheckIntervalSeconds" db:"healthCheckIntervalSeconds"` // 健康检查间隔(秒)
	HealthCheckTimeoutMs      *int   `json:"healthCheckTimeoutMs" form:"healthCheckTimeoutMs" query:"healthCheckTimeoutMs" db:"healthCheckTimeoutMs"`       // 健康检查超时(毫秒)
	HealthyThreshold          *int   `json:"healthyThreshold" form:"healthyThreshold" query:"healthyThreshold" db:"healthyThreshold"`                       // 健康阈值
	UnhealthyThreshold        *int   `json:"unhealthyThreshold" form:"unhealthyThreshold" query:"unhealthyThreshold" db:"unhealthyThreshold"`               // 不健康阈值
	ExpectedStatusCodes       string `json:"expectedStatusCodes" form:"expectedStatusCodes" query:"expectedStatusCodes" db:"expectedStatusCodes"`           // 期望的状态码,逗号分隔
	HealthCheckHeaders        string `json:"healthCheckHeaders" form:"healthCheckHeaders" query:"healthCheckHeaders" db:"healthCheckHeaders"`               // 健康检查请求头,JSON格式

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
	return "HUB_GW_SERVICE_DEFINITION"
}

// ServiceNode 服务节点模型，对应数据库HUB_GW_SERVICE_NODE表
type ServiceNode struct {
	TenantId            string `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                                         // 租户ID，联合主键
	ServiceNodeId       string `json:"serviceNodeId" form:"serviceNodeId" query:"serviceNodeId" db:"serviceNodeId"`                     // 服务节点ID，联合主键
	GatewayInstanceId   string `json:"gatewayInstanceId" form:"gatewayInstanceId" query:"gatewayInstanceId" db:"gatewayInstanceId"`     // 关联的网关实例ID
	ServiceDefinitionId string `json:"serviceDefinitionId" form:"serviceDefinitionId" query:"serviceDefinitionId" db:"serviceDefinitionId"` // 关联的服务定义ID
	NodeId              string `json:"nodeId" form:"nodeId" query:"nodeId" db:"nodeId"`                                                 // 节点标识ID

	// 根据NodeConfig.URL字段设计,分解为host+port+protocol便于查询和管理
	NodeUrl      string `json:"nodeUrl" form:"nodeUrl" query:"nodeUrl" db:"nodeUrl"`             // 节点完整URL(来自NodeConfig.URL)
	NodeHost     string `json:"nodeHost" form:"nodeHost" query:"nodeHost" db:"nodeHost"`         // 节点主机地址(从URL解析)
	NodePort     int    `json:"nodePort" form:"nodePort" query:"nodePort" db:"nodePort"`         // 节点端口(从URL解析)
	NodeProtocol string `json:"nodeProtocol" form:"nodeProtocol" query:"nodeProtocol" db:"nodeProtocol"` // 节点协议(HTTP,HTTPS,从URL解析)

	// 根据NodeConfig.Weight字段设计
	NodeWeight int `json:"nodeWeight" form:"nodeWeight" query:"nodeWeight" db:"nodeWeight"` // 节点权重(来自NodeConfig.Weight)

	// 根据NodeConfig.Health字段设计
	HealthStatus string `json:"healthStatus" form:"healthStatus" query:"healthStatus" db:"healthStatus"` // 健康状态(N不健康,Y健康,来自NodeConfig.Health)

	// 根据NodeConfig.Enabled字段设计
	NodeEnabled string `json:"nodeEnabled" form:"nodeEnabled" query:"nodeEnabled" db:"nodeEnabled"` // 节点是否启用(N禁用,Y启用,来自NodeConfig.Enabled)

	// 根据NodeConfig.Metadata字段设计
	NodeMetadata string `json:"nodeMetadata" form:"nodeMetadata" query:"nodeMetadata" db:"nodeMetadata"` // 节点元数据,JSON格式(来自NodeConfig.Metadata)

	// 运行时状态字段(非NodeConfig结构,但运维需要)
	NodeStatus            int        `json:"nodeStatus" form:"nodeStatus" query:"nodeStatus" db:"nodeStatus"`                                     // 节点运行状态(0下线,1在线,2维护)
	LastHealthCheckTime   *time.Time `json:"lastHealthCheckTime" form:"lastHealthCheckTime" query:"lastHealthCheckTime" db:"lastHealthCheckTime"` // 最后健康检查时间
	HealthCheckResult     string     `json:"healthCheckResult" form:"healthCheckResult" query:"healthCheckResult" db:"healthCheckResult"`         // 健康检查结果详情

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
func (ServiceNode) TableName() string {
	return "HUB_GW_SERVICE_NODE"
}

// GatewayInstance 网关实例模型，对应数据库HUB_GW_INSTANCE表
type GatewayInstance struct {
	TenantId          string `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                                     // 租户ID，联合主键
	GatewayInstanceId string `json:"gatewayInstanceId" form:"gatewayInstanceId" query:"gatewayInstanceId" db:"gatewayInstanceId"` // 网关实例ID，联合主键
	InstanceName      string `json:"instanceName" form:"instanceName" query:"instanceName" db:"instanceName"`                     // 实例名称
	InstanceDesc      string `json:"instanceDesc" form:"instanceDesc" query:"instanceDesc" db:"instanceDesc"`                     // 实例描述
	BindAddress       string `json:"bindAddress" form:"bindAddress" query:"bindAddress" db:"bindAddress"`                         // 绑定地址

	// HTTP/HTTPS 端口配置
	HttpPort   *int   `json:"httpPort" form:"httpPort" query:"httpPort" db:"httpPort"`         // HTTP监听端口
	HttpsPort  *int   `json:"httpsPort" form:"httpsPort" query:"httpsPort" db:"httpsPort"`     // HTTPS监听端口
	TlsEnabled string `json:"tlsEnabled" form:"tlsEnabled" query:"tlsEnabled" db:"tlsEnabled"` // 是否启用TLS(N否,Y是)

	// 证书配置 - 支持文件路径和数据库存储
	CertStorageType   string `json:"certStorageType" form:"certStorageType" query:"certStorageType" db:"certStorageType"`         // 证书存储类型(FILE文件,DATABASE数据库)
	CertFilePath      string `json:"certFilePath" form:"certFilePath" query:"certFilePath" db:"certFilePath"`                     // 证书文件路径
	KeyFilePath       string `json:"keyFilePath" form:"keyFilePath" query:"keyFilePath" db:"keyFilePath"`                         // 私钥文件路径
	CertContent       string `json:"certContent" form:"certContent" query:"certContent" db:"certContent"`                         // 证书内容(PEM格式)
	KeyContent        string `json:"keyContent" form:"keyContent" query:"keyContent" db:"keyContent"`                             // 私钥内容(PEM格式)
	CertChainContent  string `json:"certChainContent" form:"certChainContent" query:"certChainContent" db:"certChainContent"`     // 证书链内容(PEM格式)
	CertPassword      string `json:"certPassword" form:"certPassword" query:"certPassword" db:"certPassword"`                     // 证书密码(加密存储)

	// Go HTTP Server 核心配置
	MaxConnections int `json:"maxConnections" form:"maxConnections" query:"maxConnections" db:"maxConnections"`         // 最大连接数
	ReadTimeoutMs  int `json:"readTimeoutMs" form:"readTimeoutMs" query:"readTimeoutMs" db:"readTimeoutMs"`             // 读取超时时间(毫秒)
	WriteTimeoutMs int `json:"writeTimeoutMs" form:"writeTimeoutMs" query:"writeTimeoutMs" db:"writeTimeoutMs"`         // 写入超时时间(毫秒)
	IdleTimeoutMs  int `json:"idleTimeoutMs" form:"idleTimeoutMs" query:"idleTimeoutMs" db:"idleTimeoutMs"`             // 空闲连接超时时间(毫秒)
	MaxHeaderBytes int `json:"maxHeaderBytes" form:"maxHeaderBytes" query:"maxHeaderBytes" db:"maxHeaderBytes"`         // 最大请求头字节数(默认1MB)

	// 性能和并发配置
	MaxWorkers                    int    `json:"maxWorkers" form:"maxWorkers" query:"maxWorkers" db:"maxWorkers"`                                                         // 最大工作协程数
	KeepAliveEnabled              string `json:"keepAliveEnabled" form:"keepAliveEnabled" query:"keepAliveEnabled" db:"keepAliveEnabled"`                                 // 是否启用Keep-Alive(N否,Y是)
	TcpKeepAliveEnabled           string `json:"tcpKeepAliveEnabled" form:"tcpKeepAliveEnabled" query:"tcpKeepAliveEnabled" db:"tcpKeepAliveEnabled"`                     // 是否启用TCP Keep-Alive(N否,Y是)
	GracefulShutdownTimeoutMs     int    `json:"gracefulShutdownTimeoutMs" form:"gracefulShutdownTimeoutMs" query:"gracefulShutdownTimeoutMs" db:"gracefulShutdownTimeoutMs"` // 优雅关闭超时时间(毫秒)

	// TLS安全配置
	EnableHttp2                   string `json:"enableHttp2" form:"enableHttp2" query:"enableHttp2" db:"enableHttp2"`                                                     // 是否启用HTTP/2(N否,Y是)
	TlsVersion                    string `json:"tlsVersion" form:"tlsVersion" query:"tlsVersion" db:"tlsVersion"`                                                         // TLS协议版本(1.0,1.1,1.2,1.3)
	TlsCipherSuites               string `json:"tlsCipherSuites" form:"tlsCipherSuites" query:"tlsCipherSuites" db:"tlsCipherSuites"`                                     // TLS密码套件列表,逗号分隔
	DisableGeneralOptionsHandler  string `json:"disableGeneralOptionsHandler" form:"disableGeneralOptionsHandler" query:"disableGeneralOptionsHandler" db:"disableGeneralOptionsHandler"` // 是否禁用默认OPTIONS处理器(N否,Y是)

	// 日志配置关联字段
	LogConfigId       string     `json:"logConfigId" form:"logConfigId" query:"logConfigId" db:"logConfigId"`                         // 关联的日志配置ID
	HealthStatus      string     `json:"healthStatus" form:"healthStatus" query:"healthStatus" db:"healthStatus"`                     // 健康状态(N不健康,Y健康)
	LastHeartbeatTime *time.Time `json:"lastHeartbeatTime" form:"lastHeartbeatTime" query:"lastHeartbeatTime" db:"lastHeartbeatTime"` // 最后心跳时间
	InstanceMetadata  string     `json:"instanceMetadata" form:"instanceMetadata" query:"instanceMetadata" db:"instanceMetadata"`     // 实例元数据,JSON格式

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
func (GatewayInstance) TableName() string {
	return "HUB_GW_INSTANCE"
} 