package models

import "time"

// TunnelMapping 隧道静态映射模型
// 对应数据库表：TUNNEL_MAPPING
// 用途：管理隧道静态端口映射和域名映射配置
type TunnelMapping struct {
	// 主键信息
	TunnelMappingId string `json:"tunnelMappingId" form:"tunnelMappingId" query:"tunnelMappingId" db:"tunnelMappingId"` // 隧道映射ID，主键
	MappingName     string `json:"mappingName" form:"mappingName" query:"mappingName" db:"mappingName"`                 // 映射名称
	MappingType     string `json:"mappingType" form:"mappingType" query:"mappingType" db:"mappingType"`                 // 映射类型(PORT,DOMAIN,SUBDOMAIN)

	// 关联信息
	TunnelServerId string `json:"tunnelServerId" form:"tunnelServerId" query:"tunnelServerId" db:"tunnelServerId"` // 关联的隧道服务器ID
	ServerName     string `json:"serverName" form:"serverName" query:"serverName" db:"serverName"`                 // 服务器名称（冗余字段）

	// 端口映射配置（当mappingType=PORT时使用）
	ExternalPort *int    `json:"externalPort" form:"externalPort" query:"externalPort" db:"externalPort"` // 外部端口
	InternalPort *int    `json:"internalPort" form:"internalPort" query:"internalPort" db:"internalPort"` // 内部端口
	Protocol     *string `json:"protocol" form:"protocol" query:"protocol" db:"protocol"`                 // 协议类型(TCP,UDP)

	// 域名映射配置（当mappingType=DOMAIN或SUBDOMAIN时使用）
	ExternalDomain *string `json:"externalDomain" form:"externalDomain" query:"externalDomain" db:"externalDomain"` // 外部域名
	InternalHost   *string `json:"internalHost" form:"internalHost" query:"internalHost" db:"internalHost"`         // 内部主机
	InternalPort2  *int    `json:"internalPort2" form:"internalPort2" query:"internalPort2" db:"internalPort2"`     // 内部端口（域名映射用）

	// SSL/TLS配置
	EnableSSL   bool    `json:"enableSSL" form:"enableSSL" query:"enableSSL" db:"enableSSL"`         // 是否启用SSL
	SSLCertPath *string `json:"sslCertPath" form:"sslCertPath" query:"sslCertPath" db:"sslCertPath"` // SSL证书路径
	SSLKeyPath  *string `json:"sslKeyPath" form:"sslKeyPath" query:"sslKeyPath" db:"sslKeyPath"`     // SSL私钥路径
	ForceHTTPS  bool    `json:"forceHTTPS" form:"forceHTTPS" query:"forceHTTPS" db:"forceHTTPS"`     // 是否强制HTTPS

	// 访问控制
	AllowedIPs    *string `json:"allowedIPs" form:"allowedIPs" query:"allowedIPs" db:"allowedIPs"`             // 允许访问的IP列表(JSON数组)
	DeniedIPs     *string `json:"deniedIPs" form:"deniedIPs" query:"deniedIPs" db:"deniedIPs"`                 // 拒绝访问的IP列表(JSON数组)
	BasicAuthUser *string `json:"basicAuthUser" form:"basicAuthUser" query:"basicAuthUser" db:"basicAuthUser"` // HTTP基础认证用户名
	BasicAuthPass *string `json:"basicAuthPass" form:"basicAuthPass" query:"basicAuthPass" db:"basicAuthPass"` // HTTP基础认证密码

	// 负载均衡配置
	LoadBalanceType     *string `json:"loadBalanceType" form:"loadBalanceType" query:"loadBalanceType" db:"loadBalanceType"`                 // 负载均衡类型(ROUND_ROBIN,WEIGHTED,LEAST_CONN)
	TargetHosts         *string `json:"targetHosts" form:"targetHosts" query:"targetHosts" db:"targetHosts"`                                 // 目标主机列表(JSON数组)
	HealthCheckPath     *string `json:"healthCheckPath" form:"healthCheckPath" query:"healthCheckPath" db:"healthCheckPath"`                 // 健康检查路径
	HealthCheckInterval int     `json:"healthCheckInterval" form:"healthCheckInterval" query:"healthCheckInterval" db:"healthCheckInterval"` // 健康检查间隔(秒)

	// 限流配置
	RateLimitEnabled  bool `json:"rateLimitEnabled" form:"rateLimitEnabled" query:"rateLimitEnabled" db:"rateLimitEnabled"`     // 是否启用限流
	RequestsPerSecond *int `json:"requestsPerSecond" form:"requestsPerSecond" query:"requestsPerSecond" db:"requestsPerSecond"` // 每秒请求数限制
	BurstSize         *int `json:"burstSize" form:"burstSize" query:"burstSize" db:"burstSize"`                                 // 突发请求数

	// 缓存配置
	CacheEnabled bool `json:"cacheEnabled" form:"cacheEnabled" query:"cacheEnabled" db:"cacheEnabled"` // 是否启用缓存
	CacheTTL     *int `json:"cacheTTL" form:"cacheTTL" query:"cacheTTL" db:"cacheTTL"`                 // 缓存TTL(秒)
	CacheSize    *int `json:"cacheSize" form:"cacheSize" query:"cacheSize" db:"cacheSize"`             // 缓存大小(MB)

	// 日志配置
	LogEnabled bool    `json:"logEnabled" form:"logEnabled" query:"logEnabled" db:"logEnabled"` // 是否启用日志
	LogLevel   *string `json:"logLevel" form:"logLevel" query:"logLevel" db:"logLevel"`         // 日志级别(DEBUG,INFO,WARN,ERROR)
	LogPath    *string `json:"logPath" form:"logPath" query:"logPath" db:"logPath"`             // 日志文件路径

	// 状态信息
	MappingStatus string     `json:"mappingStatus" form:"mappingStatus" query:"mappingStatus" db:"mappingStatus"` // 映射状态(ACTIVE,INACTIVE,ERROR)
	LastCheckTime *time.Time `json:"lastCheckTime" form:"lastCheckTime" query:"lastCheckTime" db:"lastCheckTime"` // 最后检查时间
	ErrorMessage  *string    `json:"errorMessage" form:"errorMessage" query:"errorMessage" db:"errorMessage"`     // 错误信息

	// 统计信息
	TotalRequests     int64 `json:"totalRequests" form:"totalRequests" query:"totalRequests" db:"totalRequests"`                 // 总请求数
	TotalTraffic      int64 `json:"totalTraffic" form:"totalTraffic" query:"totalTraffic" db:"totalTraffic"`                     // 总流量(字节)
	ActiveConnections int   `json:"activeConnections" form:"activeConnections" query:"activeConnections" db:"activeConnections"` // 活跃连接数

	// 通用审计字段
	AddTime        time.Time `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                             // 创建时间
	AddWho         string    `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                 // 创建人ID
	EditTime       time.Time `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                         // 最后修改时间
	EditWho        string    `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                             // 最后修改人ID
	OprSeqFlag     string    `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                 // 操作序列标识
	CurrentVersion int       `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"` // 当前版本号
	ActiveFlag     string    `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`                 // 活动状态标记(N非活动,Y活动)
	NoteText       string    `json:"noteText,omitempty" form:"noteText" query:"noteText" db:"noteText"`               // 备注信息
}

// TunnelMappingTemplate 隧道映射模板模型
// 对应数据库表：TUNNEL_MAPPING_TEMPLATE
// 用途：管理常用的映射配置模板
type TunnelMappingTemplate struct {
	// 主键信息
	TemplateId   string `json:"templateId" form:"templateId" query:"templateId" db:"templateId"`         // 模板ID，主键
	TemplateName string `json:"templateName" form:"templateName" query:"templateName" db:"templateName"` // 模板名称
	TemplateType string `json:"templateType" form:"templateType" query:"templateType" db:"templateType"` // 模板类型(PORT,DOMAIN,SUBDOMAIN)
	Description  string `json:"description" form:"description" query:"description" db:"description"`     // 模板描述

	// 模板配置（JSON格式存储）
	ConfigTemplate string `json:"configTemplate" form:"configTemplate" query:"configTemplate" db:"configTemplate"` // 配置模板(JSON)
	DefaultValues  string `json:"defaultValues" form:"defaultValues" query:"defaultValues" db:"defaultValues"`     // 默认值(JSON)

	// 使用统计
	UseCount     int        `json:"useCount" form:"useCount" query:"useCount" db:"useCount"`                 // 使用次数
	LastUsedTime *time.Time `json:"lastUsedTime" form:"lastUsedTime" query:"lastUsedTime" db:"lastUsedTime"` // 最后使用时间

	// 通用审计字段
	AddTime        time.Time `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                             // 创建时间
	AddWho         string    `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                 // 创建人ID
	EditTime       time.Time `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                         // 最后修改时间
	EditWho        string    `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                             // 最后修改人ID
	OprSeqFlag     string    `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                 // 操作序列标识
	CurrentVersion int       `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"` // 当前版本号
	ActiveFlag     string    `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`                 // 活动状态标记(N非活动,Y活动)
	NoteText       string    `json:"noteText,omitempty" form:"noteText" query:"noteText" db:"noteText"`               // 备注信息
}

// 查询请求模型

// TunnelMappingQueryRequest 隧道映射查询请求
type TunnelMappingQueryRequest struct {
	MappingName    string `json:"mappingName" form:"mappingName"`       // 映射名称过滤（模糊查询）
	MappingType    string `json:"mappingType" form:"mappingType"`       // 映射类型过滤
	TunnelServerId string `json:"tunnelServerId" form:"tunnelServerId"` // 隧道服务器ID过滤
	MappingStatus  string `json:"mappingStatus" form:"mappingStatus"`   // 映射状态过滤
	Protocol       string `json:"protocol" form:"protocol"`             // 协议类型过滤
	ActiveFlag     string `json:"activeFlag" form:"activeFlag"`         // 活动状态标记(Y活动,N非活动,空为全部)
	Keyword        string `json:"keyword" form:"keyword"`               // 关键字搜索（映射名称、外部域名）

	// 分页参数
	PageIndex int `json:"pageIndex" form:"pageIndex"` // 页码，从1开始
	PageSize  int `json:"pageSize" form:"pageSize"`   // 每页数量，默认20
}

// TunnelMappingTemplateQueryRequest 隧道映射模板查询请求
type TunnelMappingTemplateQueryRequest struct {
	TemplateName string `json:"templateName" form:"templateName"` // 模板名称过滤（模糊查询）
	TemplateType string `json:"templateType" form:"templateType"` // 模板类型过滤
	ActiveFlag   string `json:"activeFlag" form:"activeFlag"`     // 活动状态标记(Y活动,N非活动,空为全部)
	Keyword      string `json:"keyword" form:"keyword"`           // 关键字搜索（模板名称、描述）

	// 分页参数
	PageIndex int `json:"pageIndex" form:"pageIndex"` // 页码，从1开始
	PageSize  int `json:"pageSize" form:"pageSize"`   // 每页数量，默认20
}

// 统计信息模型

// TunnelMappingStats 隧道映射统计信息
type TunnelMappingStats struct {
	TotalMappings    int   `json:"totalMappings"`    // 总映射数量
	ActiveMappings   int   `json:"activeMappings"`   // 活跃映射数量
	InactiveMappings int   `json:"inactiveMappings"` // 非活跃映射数量
	ErrorMappings    int   `json:"errorMappings"`    // 错误映射数量
	PortMappings     int   `json:"portMappings"`     // 端口映射数量
	DomainMappings   int   `json:"domainMappings"`   // 域名映射数量
	TotalRequests    int64 `json:"totalRequests"`    // 总请求数
	TotalTraffic     int64 `json:"totalTraffic"`     // 总流量
}

// PortUsageInfo 端口使用信息
type PortUsageInfo struct {
	Port        int    `json:"port"`        // 端口号
	Protocol    string `json:"protocol"`    // 协议
	MappingName string `json:"mappingName"` // 映射名称
	Status      string `json:"status"`      // 状态
}

// DomainUsageInfo 域名使用信息
type DomainUsageInfo struct {
	Domain      string `json:"domain"`      // 域名
	MappingName string `json:"mappingName"` // 映射名称
	Status      string `json:"status"`      // 状态
	EnableSSL   bool   `json:"enableSSL"`   // 是否启用SSL
}
