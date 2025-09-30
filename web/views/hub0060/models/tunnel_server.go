package models

import "time"

// TunnelServer 隧道服务器模型
// 对应数据库表：HUB_TUNNEL_SERVER
// 用途：管理隧道服务器信息
type TunnelServer struct {
	// 主键信息
	TunnelServerId    string `json:"tunnelServerId" form:"tunnelServerId" query:"tunnelServerId" db:"tunnelServerId"`             // 隧道服务器ID，主键
	TenantId          string `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                                     // 租户ID
	ServerName        string `json:"serverName" form:"serverName" query:"serverName" db:"serverName"`                             // 服务器名称
	ServerDescription string `json:"serverDescription" form:"serverDescription" query:"serverDescription" db:"serverDescription"` // 服务器描述

	// 服务器配置
	ControlAddress    string `json:"controlAddress" form:"controlAddress" query:"controlAddress" db:"controlAddress"`             // 控制端口监听地址
	ControlPort       int    `json:"controlPort" form:"controlPort" query:"controlPort" db:"controlPort"`                         // 控制端口（接受客户端连接）
	DashboardPort     int    `json:"dashboardPort" form:"dashboardPort" query:"dashboardPort" db:"dashboardPort"`                 // 管理面板端口
	VhostHttpPort     int    `json:"vhostHttpPort" form:"vhostHttpPort" query:"vhostHttpPort" db:"vhostHttpPort"`                 // 虚拟主机HTTP端口
	VhostHttpsPort    int    `json:"vhostHttpsPort" form:"vhostHttpsPort" query:"vhostHttpsPort" db:"vhostHttpsPort"`             // 虚拟主机HTTPS端口
	MaxClients        int    `json:"maxClients" form:"maxClients" query:"maxClients" db:"maxClients"`                             // 最大客户端连接数
	TokenAuth         string `json:"tokenAuth" form:"tokenAuth" query:"tokenAuth" db:"tokenAuth"`                                 // 启用Token认证(N禁用,Y启用)
	AuthToken         string `json:"authToken" form:"authToken" query:"authToken" db:"authToken"`                                 // 客户端认证Token
	TlsEnable         string `json:"tlsEnable" form:"tlsEnable" query:"tlsEnable" db:"tlsEnable"`                                 // TLS启用状态(N禁用,Y启用)
	TlsCertFile       string `json:"tlsCertFile" form:"tlsCertFile" query:"tlsCertFile" db:"tlsCertFile"`                         // TLS证书文件路径
	TlsKeyFile        string `json:"tlsKeyFile" form:"tlsKeyFile" query:"tlsKeyFile" db:"tlsKeyFile"`                             // TLS私钥文件路径
	HeartbeatInterval int    `json:"heartbeatInterval" form:"heartbeatInterval" query:"heartbeatInterval" db:"heartbeatInterval"` // 心跳间隔(秒)
	HeartbeatTimeout  int    `json:"heartbeatTimeout" form:"heartbeatTimeout" query:"heartbeatTimeout" db:"heartbeatTimeout"`     // 心跳超时(秒)
	LogLevel          string `json:"logLevel" form:"logLevel" query:"logLevel" db:"logLevel"`                                     // 日志级别(debug,info,warn,error)
	MaxPortsPerClient int    `json:"maxPortsPerClient" form:"maxPortsPerClient" query:"maxPortsPerClient" db:"maxPortsPerClient"` // 每个客户端最大端口数
	AllowPorts        string `json:"allowPorts" form:"allowPorts" query:"allowPorts" db:"allowPorts"`                             // 允许的端口范围，JSON格式

	// 状态信息
	ServerStatus  string     `json:"serverStatus" form:"serverStatus" query:"serverStatus" db:"serverStatus"`     // 服务器状态(running,stopped,error)
	StartTime     *time.Time `json:"startTime" form:"startTime" query:"startTime" db:"startTime"`                 // 服务启动时间
	ConfigVersion string     `json:"configVersion" form:"configVersion" query:"configVersion" db:"configVersion"` // 配置版本号

	// 通用审计字段
	AddTime        time.Time `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                             // 创建时间
	AddWho         string    `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                 // 创建人ID
	EditTime       time.Time `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                         // 最后修改时间
	EditWho        string    `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                             // 最后修改人ID
	OprSeqFlag     string    `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                 // 操作序列标识
	CurrentVersion int       `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"` // 当前版本号
	ActiveFlag     string    `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`                 // 活动状态标记(N非活动,Y活动)
	NoteText       string    `json:"noteText,omitempty" form:"noteText" query:"noteText" db:"noteText"`               // 备注信息
	ExtProperty    string    `json:"extProperty" form:"extProperty" query:"extProperty" db:"extProperty"`             // 扩展属性，JSON格式
}

// TunnelClient 隧道客户端模型
// 对应数据库表：HUB_TUNNEL_CLIENT
// 用途：管理隧道客户端信息
type TunnelClient struct {
	// 主键信息
	TunnelClientId    string `json:"tunnelClientId" form:"tunnelClientId" query:"tunnelClientId" db:"tunnelClientId"`             // 隧道客户端ID，主键
	TenantId          string `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                                     // 租户ID
	UserId            string `json:"userId" form:"userId" query:"userId" db:"userId"`                                             // 用户ID，关联外部用户系统
	ClientName        string `json:"clientName" form:"clientName" query:"clientName" db:"clientName"`                             // 客户端名称
	ClientDescription string `json:"clientDescription" form:"clientDescription" query:"clientDescription" db:"clientDescription"` // 客户端描述
	ClientVersion     string `json:"clientVersion" form:"clientVersion" query:"clientVersion" db:"clientVersion"`                 // 客户端版本
	OperatingSystem   string `json:"operatingSystem" form:"operatingSystem" query:"operatingSystem" db:"operatingSystem"`         // 操作系统
	ClientIpAddress   string `json:"clientIpAddress" form:"clientIpAddress" query:"clientIpAddress" db:"clientIpAddress"`         // 客户端IP地址
	ClientMacAddress  string `json:"clientMacAddress" form:"clientMacAddress" query:"clientMacAddress" db:"clientMacAddress"`     // 客户端MAC地址

	// 服务器连接配置
	ServerAddress     string `json:"serverAddress" form:"serverAddress" query:"serverAddress" db:"serverAddress"`                 // 服务器地址
	ServerPort        int    `json:"serverPort" form:"serverPort" query:"serverPort" db:"serverPort"`                             // 服务器控制端口
	AuthToken         string `json:"authToken" form:"authToken" query:"authToken" db:"authToken"`                                 // 认证令牌
	TlsEnable         string `json:"tlsEnable" form:"tlsEnable" query:"tlsEnable" db:"tlsEnable"`                                 // 启用TLS(N禁用,Y启用)
	AutoReconnect     string `json:"autoReconnect" form:"autoReconnect" query:"autoReconnect" db:"autoReconnect"`                 // 自动重连(N禁用,Y启用)
	MaxRetries        int    `json:"maxRetries" form:"maxRetries" query:"maxRetries" db:"maxRetries"`                             // 最大重试次数
	RetryInterval     int    `json:"retryInterval" form:"retryInterval" query:"retryInterval" db:"retryInterval"`                 // 重试间隔(秒)
	HeartbeatInterval int    `json:"heartbeatInterval" form:"heartbeatInterval" query:"heartbeatInterval" db:"heartbeatInterval"` // 心跳间隔(秒)
	HeartbeatTimeout  int    `json:"heartbeatTimeout" form:"heartbeatTimeout" query:"heartbeatTimeout" db:"heartbeatTimeout"`     // 心跳超时(秒)

	// 状态信息
	ConnectionStatus   string     `json:"connectionStatus" form:"connectionStatus" query:"connectionStatus" db:"connectionStatus"`         // 连接状态(connected,disconnected,connecting,error)
	LastConnectTime    *time.Time `json:"lastConnectTime" form:"lastConnectTime" query:"lastConnectTime" db:"lastConnectTime"`             // 最后连接时间
	LastDisconnectTime *time.Time `json:"lastDisconnectTime" form:"lastDisconnectTime" query:"lastDisconnectTime" db:"lastDisconnectTime"` // 最后断开时间
	TotalConnectTime   int64      `json:"totalConnectTime" form:"totalConnectTime" query:"totalConnectTime" db:"totalConnectTime"`         // 总连接时长(秒)
	ReconnectCount     int        `json:"reconnectCount" form:"reconnectCount" query:"reconnectCount" db:"reconnectCount"`                 // 重连次数
	ServiceCount       int        `json:"serviceCount" form:"serviceCount" query:"serviceCount" db:"serviceCount"`                         // 注册的服务数量
	LastHeartbeat      *time.Time `json:"lastHeartbeat" form:"lastHeartbeat" query:"lastHeartbeat" db:"lastHeartbeat"`                     // 最后心跳时间
	ClientConfig       string     `json:"clientConfig" form:"clientConfig" query:"clientConfig" db:"clientConfig"`                         // 客户端配置，JSON格式

	// 通用审计字段
	AddTime        time.Time `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                             // 创建时间
	AddWho         string    `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                 // 创建人ID
	EditTime       time.Time `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                         // 最后修改时间
	EditWho        string    `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                             // 最后修改人ID
	OprSeqFlag     string    `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                 // 操作序列标识
	CurrentVersion int       `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"` // 当前版本号
	ActiveFlag     string    `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`                 // 活动状态标记(N非活动,Y活动)
	NoteText       string    `json:"noteText,omitempty" form:"noteText" query:"noteText" db:"noteText"`               // 备注信息
	ExtProperty    string    `json:"extProperty" form:"extProperty" query:"extProperty" db:"extProperty"`             // 扩展属性，JSON格式
}

// TunnelService 隧道服务模型
// 对应数据库表：HUB_TUNNEL_SERVICE
// 用途：管理隧道服务信息
type TunnelService struct {
	// 主键信息
	TunnelServiceId    string `json:"tunnelServiceId" form:"tunnelServiceId" query:"tunnelServiceId" db:"tunnelServiceId"`             // 隧道服务ID，主键
	TenantId           string `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                                         // 租户ID
	TunnelClientId     string `json:"tunnelClientId" form:"tunnelClientId" query:"tunnelClientId" db:"tunnelClientId"`                 // 隧道客户端ID，关联HUB_TUNNEL_CLIENT
	UserId             string `json:"userId" form:"userId" query:"userId" db:"userId"`                                                 // 用户ID，关联外部用户系统
	ServiceName        string `json:"serviceName" form:"serviceName" query:"serviceName" db:"serviceName"`                             // 服务名称
	ServiceDescription string `json:"serviceDescription" form:"serviceDescription" query:"serviceDescription" db:"serviceDescription"` // 服务描述
	ServiceType        string `json:"serviceType" form:"serviceType" query:"serviceType" db:"serviceType"`                             // 服务类型(tcp,udp,http,https,stcp,sudp,xtcp)

	// 本地配置
	LocalAddress string `json:"localAddress" form:"localAddress" query:"localAddress" db:"localAddress"` // 本地地址
	LocalPort    int    `json:"localPort" form:"localPort" query:"localPort" db:"localPort"`             // 本地端口

	// 远程配置
	RemotePort    *int   `json:"remotePort" form:"remotePort" query:"remotePort" db:"remotePort"`             // 远程端口（服务端分配）
	CustomDomains string `json:"customDomains" form:"customDomains" query:"customDomains" db:"customDomains"` // 自定义域名列表，JSON格式
	SubDomain     string `json:"subDomain" form:"subDomain" query:"subDomain" db:"subDomain"`                 // 子域名前缀

	// HTTP配置
	HttpUser          string `json:"httpUser" form:"httpUser" query:"httpUser" db:"httpUser"`                                     // HTTP基础认证用户名
	HttpPassword      string `json:"httpPassword" form:"httpPassword" query:"httpPassword" db:"httpPassword"`                     // HTTP基础认证密码
	HostHeaderRewrite string `json:"hostHeaderRewrite" form:"hostHeaderRewrite" query:"hostHeaderRewrite" db:"hostHeaderRewrite"` // 重写Host头
	Headers           string `json:"headers" form:"headers" query:"headers" db:"headers"`                                         // 自定义HTTP头，JSON格式
	Locations         string `json:"locations" form:"locations" query:"locations" db:"locations"`                                 // HTTP路径配置，JSON格式

	// 安全配置
	UseEncryption  string `json:"useEncryption" form:"useEncryption" query:"useEncryption" db:"useEncryption"`     // 使用加密(N禁用,Y启用)
	UseCompression string `json:"useCompression" form:"useCompression" query:"useCompression" db:"useCompression"` // 使用压缩(N禁用,Y启用)
	SecretKey      string `json:"secretKey" form:"secretKey" query:"secretKey" db:"secretKey"`                     // 加密密钥

	// 限制配置
	BandwidthLimit  string `json:"bandwidthLimit" form:"bandwidthLimit" query:"bandwidthLimit" db:"bandwidthLimit"`     // 带宽限制
	MaxConnections  int    `json:"maxConnections" form:"maxConnections" query:"maxConnections" db:"maxConnections"`     // 最大连接数限制
	HealthCheckType string `json:"healthCheckType" form:"healthCheckType" query:"healthCheckType" db:"healthCheckType"` // 健康检查类型(tcp,http)
	HealthCheckUrl  string `json:"healthCheckUrl" form:"healthCheckUrl" query:"healthCheckUrl" db:"healthCheckUrl"`     // 健康检查URL

	// 状态信息
	ServiceStatus    string     `json:"serviceStatus" form:"serviceStatus" query:"serviceStatus" db:"serviceStatus"`             // 服务状态(active,inactive,error,offline)
	RegisteredTime   time.Time  `json:"registeredTime" form:"registeredTime" query:"registeredTime" db:"registeredTime"`         // 服务注册时间
	LastActiveTime   *time.Time `json:"lastActiveTime" form:"lastActiveTime" query:"lastActiveTime" db:"lastActiveTime"`         // 最后活跃时间
	ConnectionCount  int        `json:"connectionCount" form:"connectionCount" query:"connectionCount" db:"connectionCount"`     // 当前连接数
	TotalConnections int64      `json:"totalConnections" form:"totalConnections" query:"totalConnections" db:"totalConnections"` // 总连接数
	TotalTraffic     int64      `json:"totalTraffic" form:"totalTraffic" query:"totalTraffic" db:"totalTraffic"`                 // 总流量(字节)
	ServiceConfig    string     `json:"serviceConfig" form:"serviceConfig" query:"serviceConfig" db:"serviceConfig"`             // 服务配置，JSON格式

	// 通用审计字段
	AddTime        time.Time `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                             // 创建时间
	AddWho         string    `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                 // 创建人ID
	EditTime       time.Time `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                         // 最后修改时间
	EditWho        string    `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                             // 最后修改人ID
	OprSeqFlag     string    `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                 // 操作序列标识
	CurrentVersion int       `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"` // 当前版本号
	ActiveFlag     string    `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`                 // 活动状态标记(N非活动,Y活动)
	NoteText       string    `json:"noteText,omitempty" form:"noteText" query:"noteText" db:"noteText"`               // 备注信息
	ExtProperty    string    `json:"extProperty" form:"extProperty" query:"extProperty" db:"extProperty"`             // 扩展属性，JSON格式
}

// 查询请求模型

// TunnelServerQueryRequest 隧道服务器查询请求
type TunnelServerQueryRequest struct {
	ServerName    string `json:"serverName" form:"serverName"`       // 服务器名称过滤（模糊查询）
	ServerAddress string `json:"serverAddress" form:"serverAddress"` // 服务器地址过滤（模糊查询）
	ServerStatus  string `json:"serverStatus" form:"serverStatus"`   // 服务器状态过滤
	ActiveFlag    string `json:"activeFlag" form:"activeFlag"`       // 活动状态标记(Y活动,N非活动,空为全部)
	Keyword       string `json:"keyword" form:"keyword"`             // 关键字搜索（服务器名称、地址）

	// 分页参数
	PageIndex int `json:"pageIndex" form:"pageIndex"` // 页码，从1开始
	PageSize  int `json:"pageSize" form:"pageSize"`   // 每页数量，默认20
}

// TunnelClientQueryRequest 隧道客户端查询请求
type TunnelClientQueryRequest struct {
	ClientName     string `json:"clientName" form:"clientName"`         // 客户端名称过滤（模糊查询）
	TunnelServerId string `json:"tunnelServerId" form:"tunnelServerId"` // 隧道服务器ID过滤
	ClientStatus   string `json:"clientStatus" form:"clientStatus"`     // 客户端状态过滤
	ActiveFlag     string `json:"activeFlag" form:"activeFlag"`         // 活动状态标记(Y活动,N非活动,空为全部)
	Keyword        string `json:"keyword" form:"keyword"`               // 关键字搜索（客户端名称、地址）

	// 分页参数
	PageIndex int `json:"pageIndex" form:"pageIndex"` // 页码，从1开始
	PageSize  int `json:"pageSize" form:"pageSize"`   // 每页数量，默认20
}

// TunnelServiceQueryRequest 隧道服务查询请求
type TunnelServiceQueryRequest struct {
	ServiceName    string `json:"serviceName" form:"serviceName"`       // 服务名称过滤（模糊查询）
	ServiceType    string `json:"serviceType" form:"serviceType"`       // 服务类型过滤
	TunnelClientId string `json:"tunnelClientId" form:"tunnelClientId"` // 隧道客户端ID过滤
	ServiceStatus  string `json:"serviceStatus" form:"serviceStatus"`   // 服务状态过滤
	ActiveFlag     string `json:"activeFlag" form:"activeFlag"`         // 活动状态标记(Y活动,N非活动,空为全部)
	Keyword        string `json:"keyword" form:"keyword"`               // 关键字搜索（服务名称）

	// 分页参数
	PageIndex int `json:"pageIndex" form:"pageIndex"` // 页码，从1开始
	PageSize  int `json:"pageSize" form:"pageSize"`   // 每页数量，默认20
}

// 统计信息模型

// TunnelServerStats 隧道服务器统计信息
type TunnelServerStats struct {
	TotalServers   int `json:"totalServers"`   // 总服务器数量
	OnlineServers  int `json:"onlineServers"`  // 在线服务器数量
	OfflineServers int `json:"offlineServers"` // 离线服务器数量
	TotalClients   int `json:"totalClients"`   // 总客户端数量
	TotalServices  int `json:"totalServices"`  // 总服务数量
}

// TunnelClientStats 隧道客户端统计信息
type TunnelClientStats struct {
	TotalClients        int `json:"totalClients"`        // 总客户端数量
	ConnectedClients    int `json:"connectedClients"`    // 已连接客户端数量
	DisconnectedClients int `json:"disconnectedClients"` // 断开连接客户端数量
	TotalServices       int `json:"totalServices"`       // 总服务数量
	ActiveServices      int `json:"activeServices"`      // 活跃服务数量
}
