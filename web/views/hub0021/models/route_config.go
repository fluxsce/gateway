package models

import (
	"time"
)

// RouteConfig 路由配置模型，对应数据库HUB_GW_ROUTE_CONFIG表
type RouteConfig struct {
	TenantId          string `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                                     // 租户ID，联合主键
	RouteConfigId     string `json:"routeConfigId" form:"routeConfigId" query:"routeConfigId" db:"routeConfigId"`                 // 路由配置ID，联合主键
	GatewayInstanceId string `json:"gatewayInstanceId" form:"gatewayInstanceId" query:"gatewayInstanceId" db:"gatewayInstanceId"` // 关联的网关实例ID
	RouteName         string `json:"routeName" form:"routeName" query:"routeName" db:"routeName"`                                 // 路由名称
	RoutePath         string `json:"routePath" form:"routePath" query:"routePath" db:"routePath"`                                 // 路由路径

	// HTTP方法和域名配置
	AllowedMethods string `json:"allowedMethods" form:"allowedMethods" query:"allowedMethods" db:"allowedMethods"` // 允许的HTTP方法,JSON数组格式["GET","POST"]
	AllowedHosts   string `json:"allowedHosts" form:"allowedHosts" query:"allowedHosts" db:"allowedHosts"`         // 允许的域名,逗号分隔

	// 路由匹配配置
	MatchType       int    `json:"matchType" form:"matchType" query:"matchType" db:"matchType"`                   // 匹配类型(0精确匹配,1前缀匹配,2正则匹配)
	RoutePriority   int    `json:"routePriority" form:"routePriority" query:"routePriority" db:"routePriority"`   // 路由优先级,数值越小优先级越高
	StripPathPrefix string `json:"stripPathPrefix" form:"stripPathPrefix" query:"stripPathPrefix" db:"stripPathPrefix"` // 是否剥离路径前缀(N否,Y是)
	RewritePath     string `json:"rewritePath" form:"rewritePath" query:"rewritePath" db:"rewritePath"`           // 重写路径

	// WebSocket和超时配置
	EnableWebsocket string `json:"enableWebsocket" form:"enableWebsocket" query:"enableWebsocket" db:"enableWebsocket"` // 是否支持WebSocket(N否,Y是)
	TimeoutMs       int    `json:"timeoutMs" form:"timeoutMs" query:"timeoutMs" db:"timeoutMs"`                         // 超时时间(毫秒)

	// 重试配置
	RetryCount      int `json:"retryCount" form:"retryCount" query:"retryCount" db:"retryCount"`             // 重试次数
	RetryIntervalMs int `json:"retryIntervalMs" form:"retryIntervalMs" query:"retryIntervalMs" db:"retryIntervalMs"` // 重试间隔(毫秒)

	// 服务关联字段，直接关联服务定义表
	ServiceDefinitionId string `json:"serviceDefinitionId" form:"serviceDefinitionId" query:"serviceDefinitionId" db:"serviceDefinitionId"` // 关联的服务定义ID

	// 日志配置关联字段
	LogConfigId string `json:"logConfigId" form:"logConfigId" query:"logConfigId" db:"logConfigId"` // 关联的日志配置ID(路由级日志配置)

	// 路由元数据，用于存储额外配置信息
	RouteMetadata string `json:"routeMetadata" form:"routeMetadata" query:"routeMetadata" db:"routeMetadata"` // 路由元数据,JSON格式,存储Methods等配置

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
func (RouteConfig) TableName() string {
	return "HUB_GW_ROUTE_CONFIG"
}

// RouteAssertion 路由断言模型，对应数据库HUB_GW_ROUTE_ASSERTION表
type RouteAssertion struct {
	TenantId          string `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                                     // 租户ID，联合主键
	RouteAssertionId  string `json:"routeAssertionId" form:"routeAssertionId" query:"routeAssertionId" db:"routeAssertionId"`     // 路由断言ID，联合主键
	RouteConfigId     string `json:"routeConfigId" form:"routeConfigId" query:"routeConfigId" db:"routeConfigId"`                 // 关联的路由配置ID
	AssertionName     string `json:"assertionName" form:"assertionName" query:"assertionName" db:"assertionName"`                 // 断言名称
	AssertionType     string `json:"assertionType" form:"assertionType" query:"assertionType" db:"assertionType"`                 // 断言类型(PATH,HEADER,QUERY,COOKIE,IP)
	AssertionOperator string `json:"assertionOperator" form:"assertionOperator" query:"assertionOperator" db:"assertionOperator"` // 断言操作符(EQUAL,NOT_EQUAL,CONTAINS,MATCHES等)

	// 断言条件配置
	FieldName     string `json:"fieldName" form:"fieldName" query:"fieldName" db:"fieldName"`             // 字段名称(header/query名称)
	ExpectedValue string `json:"expectedValue" form:"expectedValue" query:"expectedValue" db:"expectedValue"` // 期望值
	PatternValue  string `json:"patternValue" form:"patternValue" query:"patternValue" db:"patternValue"`   // 匹配模式(正则表达式等)

	// 断言执行配置
	CaseSensitive  string `json:"caseSensitive" form:"caseSensitive" query:"caseSensitive" db:"caseSensitive"`   // 是否区分大小写(N否,Y是)
	AssertionOrder int    `json:"assertionOrder" form:"assertionOrder" query:"assertionOrder" db:"assertionOrder"` // 断言执行顺序
	IsRequired     string `json:"isRequired" form:"isRequired" query:"isRequired" db:"isRequired"`               // 是否必须匹配(N否,Y是)
	AssertionDesc  string `json:"assertionDesc" form:"assertionDesc" query:"assertionDesc" db:"assertionDesc"`   // 断言描述

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
func (RouteAssertion) TableName() string {
	return "HUB_GW_ROUTE_ASSERTION"
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

// RouterConfig Router配置模型，对应数据库HUB_GW_ROUTER_CONFIG表
type RouterConfig struct {
	TenantId          string `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                                     // 租户ID，联合主键
	RouterConfigId    string `json:"routerConfigId" form:"routerConfigId" query:"routerConfigId" db:"routerConfigId"`             // Router配置ID，联合主键
	GatewayInstanceId string `json:"gatewayInstanceId" form:"gatewayInstanceId" query:"gatewayInstanceId" db:"gatewayInstanceId"` // 关联的网关实例ID
	RouterName        string `json:"routerName" form:"routerName" query:"routerName" db:"routerName"`                             // Router名称
	RouterDesc        string `json:"routerDesc" form:"routerDesc" query:"routerDesc" db:"routerDesc"`                             // Router描述

	// Router基础配置
	DefaultPriority       int `json:"defaultPriority" form:"defaultPriority" query:"defaultPriority" db:"defaultPriority"`                   // 默认路由优先级
	EnableRouteCache      string `json:"enableRouteCache" form:"enableRouteCache" query:"enableRouteCache" db:"enableRouteCache"`             // 是否启用路由缓存(N否,Y是)
	RouteCacheTtlSeconds  int    `json:"routeCacheTtlSeconds" form:"routeCacheTtlSeconds" query:"routeCacheTtlSeconds" db:"routeCacheTtlSeconds"` // 路由缓存TTL(秒)
	MaxRoutes             *int   `json:"maxRoutes" form:"maxRoutes" query:"maxRoutes" db:"maxRoutes"`                                         // 最大路由数量限制
	RouteMatchTimeout     *int   `json:"routeMatchTimeout" form:"routeMatchTimeout" query:"routeMatchTimeout" db:"routeMatchTimeout"`         // 路由匹配超时时间(毫秒)

	// Router高级配置
	EnableStrictMode      string `json:"enableStrictMode" form:"enableStrictMode" query:"enableStrictMode" db:"enableStrictMode"`             // 是否启用严格模式(N否,Y是)
	EnableMetrics         string `json:"enableMetrics" form:"enableMetrics" query:"enableMetrics" db:"enableMetrics"`                         // 是否启用路由指标收集(N否,Y是)
	EnableTracing         string `json:"enableTracing" form:"enableTracing" query:"enableTracing" db:"enableTracing"`                         // 是否启用链路追踪(N否,Y是)
	CaseSensitive         string `json:"caseSensitive" form:"caseSensitive" query:"caseSensitive" db:"caseSensitive"`                         // 路径匹配是否区分大小写(N否,Y是)
	RemoveTrailingSlash   string `json:"removeTrailingSlash" form:"removeTrailingSlash" query:"removeTrailingSlash" db:"removeTrailingSlash"` // 是否移除路径尾部斜杠(N否,Y是)

	// 路由处理配置
	EnableGlobalFilters   string `json:"enableGlobalFilters" form:"enableGlobalFilters" query:"enableGlobalFilters" db:"enableGlobalFilters"`       // 是否启用全局过滤器(N否,Y是)
	FilterExecutionMode   string `json:"filterExecutionMode" form:"filterExecutionMode" query:"filterExecutionMode" db:"filterExecutionMode"`       // 过滤器执行模式(SEQUENTIAL顺序,PARALLEL并行)
	MaxFilterChainDepth   *int   `json:"maxFilterChainDepth" form:"maxFilterChainDepth" query:"maxFilterChainDepth" db:"maxFilterChainDepth"`       // 最大过滤器链深度

	// 性能优化配置
	EnableRoutePooling    string `json:"enableRoutePooling" form:"enableRoutePooling" query:"enableRoutePooling" db:"enableRoutePooling"`          // 是否启用路由对象池(N否,Y是)
	RoutePoolSize         *int   `json:"routePoolSize" form:"routePoolSize" query:"routePoolSize" db:"routePoolSize"`                              // 路由对象池大小
	EnableAsyncProcessing string `json:"enableAsyncProcessing" form:"enableAsyncProcessing" query:"enableAsyncProcessing" db:"enableAsyncProcessing"` // 是否启用异步处理(N否,Y是)

	// 错误处理配置
	EnableFallback      string `json:"enableFallback" form:"enableFallback" query:"enableFallback" db:"enableFallback"`             // 是否启用降级处理(N否,Y是)
	FallbackRoute       string `json:"fallbackRoute" form:"fallbackRoute" query:"fallbackRoute" db:"fallbackRoute"`                 // 降级路由路径
	NotFoundStatusCode  int    `json:"notFoundStatusCode" form:"notFoundStatusCode" query:"notFoundStatusCode" db:"notFoundStatusCode"` // 路由未找到时的状态码
	NotFoundMessage     string `json:"notFoundMessage" form:"notFoundMessage" query:"notFoundMessage" db:"notFoundMessage"`         // 路由未找到时的提示消息

	// 自定义配置
	RouterMetadata string `json:"routerMetadata" form:"routerMetadata" query:"routerMetadata" db:"routerMetadata"` // Router元数据,JSON格式
	CustomConfig   string `json:"customConfig" form:"customConfig" query:"customConfig" db:"customConfig"`         // 自定义配置,JSON格式

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
func (RouterConfig) TableName() string {
	return "HUB_GW_ROUTER_CONFIG"
}

// RouteConfigWithService 路由配置和服务定义的组合VO，用于关联查询时的返回
type RouteConfigWithService struct {
	// 路由配置信息
	TenantId            string     `json:"tenantId" db:"tenantId"`
	RouteConfigId       string     `json:"routeConfigId" db:"routeConfigId"`
	GatewayInstanceId   string     `json:"gatewayInstanceId" db:"gatewayInstanceId"`
	RouteName           string     `json:"routeName" db:"routeName"`
	RoutePath           string     `json:"routePath" db:"routePath"`
	AllowedMethods      string     `json:"allowedMethods" db:"allowedMethods"`
	AllowedHosts        string     `json:"allowedHosts" db:"allowedHosts"`
	MatchType           int        `json:"matchType" db:"matchType"`
	RoutePriority       int        `json:"routePriority" db:"routePriority"`
	StripPathPrefix     string     `json:"stripPathPrefix" db:"stripPathPrefix"`
	RewritePath         string     `json:"rewritePath" db:"rewritePath"`
	EnableWebsocket     string     `json:"enableWebsocket" db:"enableWebsocket"`
	TimeoutMs           int        `json:"timeoutMs" db:"timeoutMs"`
	RetryCount          int        `json:"retryCount" db:"retryCount"`
	RetryIntervalMs     int        `json:"retryIntervalMs" db:"retryIntervalMs"`
	ServiceDefinitionId string     `json:"serviceDefinitionId" db:"serviceDefinitionId"`
	LogConfigId         string     `json:"logConfigId" db:"logConfigId"`
	RouteMetadata       string     `json:"routeMetadata" db:"routeMetadata"`
	
	// 服务定义信息（关联查询）
	ServiceName         *string    `json:"serviceName" db:"serviceName"`
	ServiceDesc         *string    `json:"serviceDesc" db:"serviceDesc"`
	ServiceType         *int       `json:"serviceType" db:"serviceType"`
	LoadBalanceStrategy *string    `json:"loadBalanceStrategy" db:"loadBalanceStrategy"`
	
	// 预留字段
	Reserved1           string     `json:"reserved1" db:"reserved1"`
	Reserved2           string     `json:"reserved2" db:"reserved2"`
	Reserved3           *int       `json:"reserved3" db:"reserved3"`
	Reserved4           *int       `json:"reserved4" db:"reserved4"`
	Reserved5           *time.Time `json:"reserved5" db:"reserved5"`
	
	// 扩展属性
	ExtProperty         string     `json:"extProperty" db:"extProperty"`
	
	// 标准字段
	AddTime             time.Time  `json:"addTime" db:"addTime"`
	AddWho              string     `json:"addWho" db:"addWho"`
	EditTime            time.Time  `json:"editTime" db:"editTime"`
	EditWho             string     `json:"editWho" db:"editWho"`
	OprSeqFlag          string     `json:"oprSeqFlag" db:"oprSeqFlag"`
	CurrentVersion      int        `json:"currentVersion" db:"currentVersion"`
	ActiveFlag          string     `json:"activeFlag" db:"activeFlag"`
	NoteText            string     `json:"noteText" db:"noteText"`
} 