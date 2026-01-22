package dbloader

// GatewayInstanceRecord 网关实例数据库记录
type GatewayInstanceRecord struct {
	TenantId                     string  `db:"tenantId"`
	InstanceId                   string  `db:"gatewayInstanceId"`
	InstanceName                 string  `db:"instanceName"`
	InstanceDesc                 *string `db:"instanceDesc"`
	BindAddress                  string  `db:"bindAddress"`
	HTTPPort                     *int    `db:"httpPort"`
	HTTPSPort                    *int    `db:"httpsPort"`
	TLSEnabled                   string  `db:"tlsEnabled"`
	CertStorageType              string  `db:"certStorageType"`
	CertFilePath                 *string `db:"certFilePath"`
	KeyFilePath                  *string `db:"keyFilePath"`
	CertContent                  *string `db:"certContent"`
	KeyContent                   *string `db:"keyContent"`
	CertChainContent             *string `db:"certChainContent"`
	CertPassword                 *string `db:"certPassword"`
	MaxConnections               int     `db:"maxConnections"`
	ReadTimeoutMs                int     `db:"readTimeoutMs"`
	WriteTimeoutMs               int     `db:"writeTimeoutMs"`
	IdleTimeoutMs                int     `db:"idleTimeoutMs"`
	MaxHeaderBytes               int     `db:"maxHeaderBytes"`
	MaxWorkers                   int     `db:"maxWorkers"`
	KeepAliveEnabled             string  `db:"keepAliveEnabled"`
	TCPKeepAliveEnabled          string  `db:"tcpKeepAliveEnabled"`
	GracefulShutdownTimeoutMs    int     `db:"gracefulShutdownTimeoutMs"`
	EnableHTTP2                  string  `db:"enableHttp2"`
	TLSVersion                   *string `db:"tlsVersion"`
	TLSCipherSuites              *string `db:"tlsCipherSuites"`
	DisableGeneralOptionsHandler string  `db:"disableGeneralOptionsHandler"`
	LogConfigId                  *string `db:"logConfigId"`
	HealthStatus                 string  `db:"healthStatus"`
	LastHeartbeatTime            *string `db:"lastHeartbeatTime"`
	InstanceMetadata             *string `db:"instanceMetadata"`
	ActiveFlag                   string  `db:"activeFlag"`
}

// RouterConfigRecord Router配置数据库记录
type RouterConfigRecord struct {
	TenantId              string  `db:"tenantId"`
	RouterConfigId        string  `db:"routerConfigId"`
	GatewayInstanceId     string  `db:"gatewayInstanceId"`
	RouterName            string  `db:"routerName"`
	DefaultPriority       int     `db:"defaultPriority"`
	EnableRouteCache      string  `db:"enableRouteCache"`
	RouteCacheTtlSeconds  int     `db:"routeCacheTtlSeconds"`
	MaxRoutes             *int32  `db:"maxRoutes"`
	RouteMatchTimeout     *int32  `db:"routeMatchTimeout"`
	EnableStrictMode      string  `db:"enableStrictMode"`
	EnableMetrics         string  `db:"enableMetrics"`
	EnableTracing         string  `db:"enableTracing"`
	CaseSensitive         string  `db:"caseSensitive"`
	RemoveTrailingSlash   string  `db:"removeTrailingSlash"`
	EnableGlobalFilters   string  `db:"enableGlobalFilters"`
	FilterExecutionMode   string  `db:"filterExecutionMode"`
	MaxFilterChainDepth   *int32  `db:"maxFilterChainDepth"`
	EnableRoutePooling    string  `db:"enableRoutePooling"`
	RoutePoolSize         *int32  `db:"routePoolSize"`
	EnableAsyncProcessing string  `db:"enableAsyncProcessing"`
	EnableFallback        string  `db:"enableFallback"`
	FallbackRoute         *string `db:"fallbackRoute"`
	NotFoundStatusCode    int     `db:"notFoundStatusCode"`
	NotFoundMessage       string  `db:"notFoundMessage"`
	RouterMetadata        *string `db:"routerMetadata"`
	CustomConfig          *string `db:"customConfig"`
	ActiveFlag            string  `db:"activeFlag"`
}

// RouteConfigRecord 路由配置数据库记录
type RouteConfigRecord struct {
	TenantId            string  `db:"tenantId"`
	RouteConfigId       string  `db:"routeConfigId"`
	GatewayInstanceId   string  `db:"gatewayInstanceId"`
	RouteName           string  `db:"routeName"`
	RoutePath           string  `db:"routePath"`
	AllowedMethods      *string `db:"allowedMethods"`
	AllowedHosts        *string `db:"allowedHosts"`
	MatchType           int     `db:"matchType"`
	RoutePriority       int     `db:"routePriority"`
	StripPathPrefix     string  `db:"stripPathPrefix"`
	RewritePath         *string `db:"rewritePath"`
	EnableWebsocket     string  `db:"enableWebsocket"`
	TimeoutMs           int     `db:"timeoutMs"`
	RetryCount          int     `db:"retryCount"`
	RetryIntervalMs     int     `db:"retryIntervalMs"`
	ServiceDefinitionId *string `db:"serviceDefinitionId"`
	LogConfigId         *string `db:"logConfigId"`
	RouteMetadata       *string `db:"routeMetadata"`
	ActiveFlag          string  `db:"activeFlag"`
}

// RouteAssertionRecord 路由断言数据库记录
type RouteAssertionRecord struct {
	TenantId          string  `db:"tenantId"`
	RouteAssertionId  string  `db:"routeAssertionId"`
	RouteConfigId     string  `db:"routeConfigId"`
	AssertionName     string  `db:"assertionName"`
	AssertionType     string  `db:"assertionType"`
	AssertionOperator string  `db:"assertionOperator"`
	FieldName         *string `db:"fieldName"`
	ExpectedValue     *string `db:"expectedValue"`
	PatternValue      *string `db:"patternValue"`
	CaseSensitive     string  `db:"caseSensitive"`
	AssertionOrder    int     `db:"assertionOrder"`
	IsRequired        string  `db:"isRequired"`
	AssertionDesc     *string `db:"assertionDesc"`
	ActiveFlag        string  `db:"activeFlag"`
}

// FilterConfigRecord 过滤器配置数据库记录
type FilterConfigRecord struct {
	TenantId       string  `db:"tenantId"`
	FilterConfigId string  `db:"filterConfigId"`
	FilterName     string  `db:"filterName"`
	FilterType     string  `db:"filterType"`
	FilterAction   string  `db:"filterAction"`
	FilterOrder    int     `db:"filterOrder"`
	FilterConfig   string  `db:"filterConfig"`
	ConfigId       *string `db:"configId"`
	ActiveFlag     string  `db:"activeFlag"`
}

// SecurityConfigRecord 安全配置数据库记录
type SecurityConfigRecord struct {
	TenantId          string  `db:"tenantId"`
	SecurityConfigId  string  `db:"securityConfigId"`
	GatewayInstanceId *string `db:"gatewayInstanceId"`
	RouteConfigId     *string `db:"routeConfigId"`
	ConfigName        string  `db:"configName"`
	ConfigDesc        *string `db:"configDesc"`
	ConfigPriority    int     `db:"configPriority"`
	CustomConfigJson  *string `db:"customConfigJson"`
	ActiveFlag        string  `db:"activeFlag"`
}

// IPAccessConfigRecord IP访问控制配置数据库记录
type IPAccessConfigRecord struct {
	TenantId           string  `db:"tenantId"`
	IpAccessConfigId   string  `db:"ipAccessConfigId"`
	SecurityConfigId   string  `db:"securityConfigId"`
	ConfigName         string  `db:"configName"`
	DefaultPolicy      string  `db:"defaultPolicy"`
	WhitelistIps       *string `db:"whitelistIps"`
	BlacklistIps       *string `db:"blacklistIps"`
	WhitelistCidrs     *string `db:"whitelistCidrs"`
	BlacklistCidrs     *string `db:"blacklistCidrs"`
	TrustXForwardedFor string  `db:"trustXForwardedFor"`
	TrustXRealIp       string  `db:"trustXRealIp"`
	ActiveFlag         string  `db:"activeFlag"`
}

// UserAgentAccessConfigRecord User-Agent访问控制配置数据库记录
type UserAgentAccessConfigRecord struct {
	TenantId                string  `db:"tenantId"`
	UseragentAccessConfigId string  `db:"useragentAccessConfigId"`
	SecurityConfigId        string  `db:"securityConfigId"`
	ConfigName              string  `db:"configName"`
	DefaultPolicy           string  `db:"defaultPolicy"`
	WhitelistPatterns       *string `db:"whitelistPatterns"`
	BlacklistPatterns       *string `db:"blacklistPatterns"`
	BlockEmptyUserAgent     string  `db:"blockEmptyUserAgent"`
	ActiveFlag              string  `db:"activeFlag"`
}

// APIAccessConfigRecord API访问控制配置数据库记录
type APIAccessConfigRecord struct {
	TenantId          string  `db:"tenantId"`
	ApiAccessConfigId string  `db:"apiAccessConfigId"`
	SecurityConfigId  string  `db:"securityConfigId"`
	ConfigName        string  `db:"configName"`
	DefaultPolicy     string  `db:"defaultPolicy"`
	WhitelistPaths    *string `db:"whitelistPaths"`
	BlacklistPaths    *string `db:"blacklistPaths"`
	AllowedMethods    string  `db:"allowedMethods"`
	BlockedMethods    *string `db:"blockedMethods"`
	ActiveFlag        string  `db:"activeFlag"`
}

// DomainAccessConfigRecord 域名访问控制配置数据库记录
type DomainAccessConfigRecord struct {
	TenantId             string  `db:"tenantId"`
	DomainAccessConfigId string  `db:"domainAccessConfigId"`
	SecurityConfigId     string  `db:"securityConfigId"`
	ConfigName           string  `db:"configName"`
	DefaultPolicy        string  `db:"defaultPolicy"`
	WhitelistDomains     *string `db:"whitelistDomains"`
	BlacklistDomains     *string `db:"blacklistDomains"`
	AllowSubdomains      string  `db:"allowSubdomains"`
	ActiveFlag           string  `db:"activeFlag"`
}

// AuthConfigRecord 认证配置数据库记录
type AuthConfigRecord struct {
	TenantId          string  `db:"tenantId"`
	AuthConfigId      string  `db:"authConfigId"`
	AuthName          string  `db:"authName"`
	AuthType          string  `db:"authType"`
	AuthStrategy      string  `db:"authStrategy"`
	AuthConfig        string  `db:"authConfig"`
	ExemptPaths       *string `db:"exemptPaths"`
	ExemptHeaders     *string `db:"exemptHeaders"`
	FailureStatusCode int     `db:"failureStatusCode"`
	FailureMessage    string  `db:"failureMessage"`
}

// CORSConfigRecord CORS配置数据库记录
type CORSConfigRecord struct {
	TenantId         string  `db:"tenantId"`
	CorsConfigId     string  `db:"corsConfigId"`
	ConfigName       string  `db:"configName"`
	AllowOrigins     string  `db:"allowOrigins"`
	AllowMethods     string  `db:"allowMethods"`
	AllowHeaders     *string `db:"allowHeaders"`
	ExposeHeaders    *string `db:"exposeHeaders"`
	AllowCredentials string  `db:"allowCredentials"`
	MaxAgeSeconds    int     `db:"maxAgeSeconds"`
}

// RateLimitConfigRecord 限流配置数据库记录
type RateLimitConfigRecord struct {
	TenantId            string `db:"tenantId"`
	RateLimitConfigId   string `db:"rateLimitConfigId"`
	LimitName           string `db:"limitName"`
	Algorithm           string `db:"algorithm"`
	KeyStrategy         string `db:"keyStrategy"`
	LimitRate           int    `db:"limitRate"`
	BurstCapacity       int    `db:"burstCapacity"`
	TimeWindowSeconds   int    `db:"timeWindowSeconds"`
	RejectionStatusCode int    `db:"rejectionStatusCode"`
	RejectionMessage    string `db:"rejectionMessage"`
	CustomConfig        string `db:"customConfig"`
}

// ServiceConfigRecord 服务配置数据库记录
type ServiceConfigRecord struct {
	TenantId                   string  `db:"tenantId"`
	ServiceDefinitionId        string  `db:"serviceDefinitionId"`
	ServiceName                string  `db:"serviceName"`
	ServiceDesc                *string `db:"serviceDesc"`
	ServiceType                int     `db:"serviceType"`
	LoadBalanceStrategy        string  `db:"loadBalanceStrategy"`
	DiscoveryType              *string `db:"discoveryType"`
	DiscoveryConfig            *string `db:"discoveryConfig"`
	SessionAffinity            string  `db:"sessionAffinity"`
	StickySession              string  `db:"stickySession"`
	MaxRetries                 int     `db:"maxRetries"`
	RetryTimeoutMs             int     `db:"retryTimeoutMs"`
	EnableCircuitBreaker       string  `db:"enableCircuitBreaker"`
	HealthCheckEnabled         string  `db:"healthCheckEnabled"`
	HealthCheckPath            string  `db:"healthCheckPath"`
	HealthCheckMethod          string  `db:"healthCheckMethod"`
	HealthCheckIntervalSeconds *int32  `db:"healthCheckIntervalSeconds"`
	HealthCheckTimeoutMs       *int32  `db:"healthCheckTimeoutMs"`
	HealthyThreshold           *int32  `db:"healthyThreshold"`
	UnhealthyThreshold         *int32  `db:"unhealthyThreshold"`
	ExpectedStatusCodes        string  `db:"expectedStatusCodes"`
	HealthCheckHeaders         *string `db:"healthCheckHeaders"`
	LoadBalancerConfig         *string `db:"loadBalancerConfig"`
	ServiceMetadata            *string `db:"serviceMetadata"`
	ActiveFlag                 string  `db:"activeFlag"`
}

// ServiceNodeRecord 服务节点数据库记录
type ServiceNodeRecord struct {
	TenantId            string  `db:"tenantId"`
	ServiceNodeId       string  `db:"serviceNodeId"`
	ServiceDefinitionId string  `db:"serviceDefinitionId"`
	NodeId              string  `db:"nodeId"`
	NodeUrl             string  `db:"nodeUrl"`
	NodeHost            string  `db:"nodeHost"`
	NodePort            int     `db:"nodePort"`
	NodeProtocol        string  `db:"nodeProtocol"`
	NodeWeight          int     `db:"nodeWeight"`
	HealthStatus        string  `db:"healthStatus"`
	NodeMetadata        *string `db:"nodeMetadata"`
	NodeStatus          int     `db:"nodeStatus"`
	ActiveFlag          string  `db:"activeFlag"`
}

// ProxyConfigRecord 代理配置数据库记录
type ProxyConfigRecord struct {
	TenantId      string  `db:"tenantId"`
	ProxyConfigId string  `db:"proxyConfigId"`
	ProxyName     string  `db:"proxyName"`
	ProxyType     string  `db:"proxyType"`
	ProxyConfig   string  `db:"proxyConfig"`
	CustomConfig  *string `db:"customConfig"`
}

// LogConfigRecord 日志配置数据库记录
type LogConfigRecord struct {
	TenantId                   string `db:"tenantId"`
	LogConfigId                string `db:"logConfigId"`
	ConfigName                 string `db:"configName"`
	ConfigDesc                 string `db:"configDesc"`
	LogFormat                  string `db:"logFormat"`
	RecordRequestBody          string `db:"recordRequestBody"`
	RecordResponseBody         string `db:"recordResponseBody"`
	RecordHeaders              string `db:"recordHeaders"`
	MaxBodySizeBytes           int    `db:"maxBodySizeBytes"`
	OutputTargets              string `db:"outputTargets"`
	FileConfig                 string `db:"fileConfig"`
	DatabaseConfig             string `db:"databaseConfig"`
	MongoConfig                string `db:"mongoConfig"`
	ElasticsearchConfig        string `db:"elasticsearchConfig"`
	ClickhouseConfig           string `db:"clickhouseConfig"`
	EnableAsyncLogging         string `db:"enableAsyncLogging"`
	AsyncQueueSize             int    `db:"asyncQueueSize"`
	AsyncFlushIntervalMs       int    `db:"asyncFlushIntervalMs"`
	EnableBatchProcessing      string `db:"enableBatchProcessing"`
	BatchSize                  int    `db:"batchSize"`
	BatchTimeoutMs             int    `db:"batchTimeoutMs"`
	LogRetentionDays           int    `db:"logRetentionDays"`
	EnableFileRotation         string `db:"enableFileRotation"`
	MaxFileSizeMB              *int   `db:"maxFileSizeMB"`
	MaxFileCount               *int   `db:"maxFileCount"`
	RotationPattern            string `db:"rotationPattern"`
	EnableSensitiveDataMasking string `db:"enableSensitiveDataMasking"`
	SensitiveFields            string `db:"sensitiveFields"`
	MaskingPattern             string `db:"maskingPattern"`
	BufferSize                 int    `db:"bufferSize"`
	FlushThreshold             int    `db:"flushThreshold"`
	ConfigPriority             int    `db:"configPriority"`
	ActiveFlag                 string `db:"activeFlag"`
	ExtProperty                string `db:"extProperty"`
}
