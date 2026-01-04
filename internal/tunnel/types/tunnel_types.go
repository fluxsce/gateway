// Package types 定义隧道管理系统的核心数据类型
// 基于FRP架构设计，支持静态端口映射和动态服务注册
package types

import (
	"time"
)

// TunnelServer 隧道服务器配置
// 对应数据库表: HUB_TUNNEL_SERVER
// 对应FRP: frps配置
type TunnelServer struct {
	TunnelServerId    string     `json:"tunnelServerId" form:"tunnelServerId" query:"tunnelServerId" db:"tunnelServerId"`
	TenantId          string     `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`
	ServerName        string     `json:"serverName" form:"serverName" query:"serverName" db:"serverName"`
	ServerDescription string     `json:"serverDescription" form:"serverDescription" query:"serverDescription" db:"serverDescription"`
	ControlAddress    string     `json:"controlAddress" form:"controlAddress" query:"controlAddress" db:"controlAddress"`
	ControlPort       int        `json:"controlPort" form:"controlPort" query:"controlPort" db:"controlPort"`
	DashboardPort     *int       `json:"dashboardPort" form:"dashboardPort" query:"dashboardPort" db:"dashboardPort"`
	VhostHttpPort     *int       `json:"vhostHttpPort" form:"vhostHttpPort" query:"vhostHttpPort" db:"vhostHttpPort"`
	VhostHttpsPort    *int       `json:"vhostHttpsPort" form:"vhostHttpsPort" query:"vhostHttpsPort" db:"vhostHttpsPort"`
	MaxClients        int        `json:"maxClients" form:"maxClients" query:"maxClients" db:"maxClients"`
	TokenAuth         string     `json:"tokenAuth" form:"tokenAuth" query:"tokenAuth" db:"tokenAuth"` // Y/N
	AuthToken         string     `json:"authToken" form:"authToken" query:"authToken" db:"authToken"`
	TlsEnable         string     `json:"tlsEnable" form:"tlsEnable" query:"tlsEnable" db:"tlsEnable"` // Y/N
	TlsCertFile       *string    `json:"tlsCertFile" form:"tlsCertFile" query:"tlsCertFile" db:"tlsCertFile"`
	TlsKeyFile        *string    `json:"tlsKeyFile" form:"tlsKeyFile" query:"tlsKeyFile" db:"tlsKeyFile"`
	HeartbeatInterval int        `json:"heartbeatInterval" form:"heartbeatInterval" query:"heartbeatInterval" db:"heartbeatInterval"`
	HeartbeatTimeout  int        `json:"heartbeatTimeout" form:"heartbeatTimeout" query:"heartbeatTimeout" db:"heartbeatTimeout"`
	LogLevel          string     `json:"logLevel" form:"logLevel" query:"logLevel" db:"logLevel"`
	MaxPortsPerClient *int       `json:"maxPortsPerClient" form:"maxPortsPerClient" query:"maxPortsPerClient" db:"maxPortsPerClient"`
	AllowPorts        *string    `json:"allowPorts" form:"allowPorts" query:"allowPorts" db:"allowPorts"` // JSON格式
	ServerStatus      string     `json:"serverStatus" form:"serverStatus" query:"serverStatus" db:"serverStatus"`
	StartTime         *time.Time `json:"startTime" form:"startTime" query:"startTime" db:"startTime"`
	ConfigVersion     *string    `json:"configVersion" form:"configVersion" query:"configVersion" db:"configVersion"`

	// 通用字段
	AddTime        time.Time `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`
	AddWho         string    `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`
	EditTime       time.Time `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`
	EditWho        string    `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`
	OprSeqFlag     string    `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`
	CurrentVersion int       `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"`
	ActiveFlag     string    `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`
	NoteText       *string   `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`
	ExtProperty    *string   `json:"extProperty" form:"extProperty" query:"extProperty" db:"extProperty"`
}

// TunnelClient 隧道客户端信息
// 对应数据库表: HUB_TUNNEL_CLIENT
// 对应FRP: frpc客户端连接
type TunnelClient struct {
	TunnelClientId     string     `json:"tunnelClientId" form:"tunnelClientId" query:"tunnelClientId" db:"tunnelClientId"`
	TenantId           string     `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`
	UserId             string     `json:"userId" form:"userId" query:"userId" db:"userId"`
	ClientName         string     `json:"clientName" form:"clientName" query:"clientName" db:"clientName"`
	ClientDescription  *string    `json:"clientDescription" form:"clientDescription" query:"clientDescription" db:"clientDescription"`
	ClientVersion      *string    `json:"clientVersion" form:"clientVersion" query:"clientVersion" db:"clientVersion"`
	OperatingSystem    *string    `json:"operatingSystem" form:"operatingSystem" query:"operatingSystem" db:"operatingSystem"`
	ClientIpAddress    *string    `json:"clientIpAddress" form:"clientIpAddress" query:"clientIpAddress" db:"clientIpAddress"`
	ClientMacAddress   *string    `json:"clientMacAddress" form:"clientMacAddress" query:"clientMacAddress" db:"clientMacAddress"`
	ServerAddress      string     `json:"serverAddress" form:"serverAddress" query:"serverAddress" db:"serverAddress"`
	ServerPort         int        `json:"serverPort" form:"serverPort" query:"serverPort" db:"serverPort"`
	AuthToken          string     `json:"authToken" form:"authToken" query:"authToken" db:"authToken"`
	TlsEnable          string     `json:"tlsEnable" form:"tlsEnable" query:"tlsEnable" db:"tlsEnable"`                 // Y/N
	AutoReconnect      string     `json:"autoReconnect" form:"autoReconnect" query:"autoReconnect" db:"autoReconnect"` // Y/N
	MaxRetries         int        `json:"maxRetries" form:"maxRetries" query:"maxRetries" db:"maxRetries"`
	RetryInterval      int        `json:"retryInterval" form:"retryInterval" query:"retryInterval" db:"retryInterval"`
	HeartbeatInterval  int        `json:"heartbeatInterval" form:"heartbeatInterval" query:"heartbeatInterval" db:"heartbeatInterval"`
	HeartbeatTimeout   int        `json:"heartbeatTimeout" form:"heartbeatTimeout" query:"heartbeatTimeout" db:"heartbeatTimeout"`
	ConnectionStatus   string     `json:"connectionStatus" form:"connectionStatus" query:"connectionStatus" db:"connectionStatus"`
	LastConnectTime    *time.Time `json:"lastConnectTime" form:"lastConnectTime" query:"lastConnectTime" db:"lastConnectTime"`
	LastDisconnectTime *time.Time `json:"lastDisconnectTime" form:"lastDisconnectTime" query:"lastDisconnectTime" db:"lastDisconnectTime"`
	TotalConnectTime   int64      `json:"totalConnectTime" form:"totalConnectTime" query:"totalConnectTime" db:"totalConnectTime"`
	ReconnectCount     int        `json:"reconnectCount" form:"reconnectCount" query:"reconnectCount" db:"reconnectCount"`
	ServiceCount       int        `json:"serviceCount" form:"serviceCount" query:"serviceCount" db:"serviceCount"`
	LastHeartbeat      *time.Time `json:"lastHeartbeat" form:"lastHeartbeat" query:"lastHeartbeat" db:"lastHeartbeat"`
	ClientConfig       *string    `json:"clientConfig" form:"clientConfig" query:"clientConfig" db:"clientConfig"` // JSON格式

	// 通用字段
	AddTime        time.Time `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`
	AddWho         string    `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`
	EditTime       time.Time `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`
	EditWho        string    `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`
	OprSeqFlag     string    `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`
	CurrentVersion int       `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"`
	ActiveFlag     string    `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`
	NoteText       *string   `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`
	ExtProperty    *string   `json:"extProperty" form:"extProperty" query:"extProperty" db:"extProperty"`

	// 运行时字段（不存储到数据库）
	// 每个客户端维护自己注册的服务列表
	Services map[string]*TunnelService `json:"services" form:"-" query:"-" db:"-"` // serviceID -> TunnelService

	// 连接状态字段（运行时维护）
	Authenticated    bool      `json:"-" form:"-" query:"-" db:"-"` // 是否已认证
	LastActivityTime time.Time `json:"-" form:"-" query:"-" db:"-"` // 最后活动时间
}

// TunnelService 隧道服务配置 (动态注册的服务)
// 对应数据库表: HUB_TUNNEL_SERVICE
// 对应FRP: frpc服务配置 [web], [ssh] 等
type TunnelService struct {
	TunnelServiceId    string     `json:"tunnelServiceId" form:"tunnelServiceId" query:"tunnelServiceId" db:"tunnelServiceId"`
	TenantId           string     `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`
	TunnelClientId     string     `json:"tunnelClientId" form:"tunnelClientId" query:"tunnelClientId" db:"tunnelClientId"`
	UserId             string     `json:"userId" form:"userId" query:"userId" db:"userId"`
	ServiceName        string     `json:"serviceName" form:"serviceName" query:"serviceName" db:"serviceName"`
	ServiceDescription *string    `json:"serviceDescription" form:"serviceDescription" query:"serviceDescription" db:"serviceDescription"`
	ServiceType        string     `json:"serviceType" form:"serviceType" query:"serviceType" db:"serviceType"` // tcp, udp, http, https, stcp, sudp, xtcp
	LocalAddress       string     `json:"localAddress" form:"localAddress" query:"localAddress" db:"localAddress"`
	LocalPort          int        `json:"localPort" form:"localPort" query:"localPort" db:"localPort"`
	RemotePort         *int       `json:"remotePort" form:"remotePort" query:"remotePort" db:"remotePort"`
	CustomDomains      *string    `json:"customDomains" form:"customDomains" query:"customDomains" db:"customDomains"` // JSON格式
	SubDomain          *string    `json:"subDomain" form:"subDomain" query:"subDomain" db:"subDomain"`
	HttpUser           *string    `json:"httpUser" form:"httpUser" query:"httpUser" db:"httpUser"`
	HttpPassword       *string    `json:"httpPassword" form:"httpPassword" query:"httpPassword" db:"httpPassword"`
	HostHeaderRewrite  *string    `json:"hostHeaderRewrite" form:"hostHeaderRewrite" query:"hostHeaderRewrite" db:"hostHeaderRewrite"`
	Headers            *string    `json:"headers" form:"headers" query:"headers" db:"headers"`                             // JSON格式
	Locations          *string    `json:"locations" form:"locations" query:"locations" db:"locations"`                     // JSON格式
	UseEncryption      string     `json:"useEncryption" form:"useEncryption" query:"useEncryption" db:"useEncryption"`     // Y/N
	UseCompression     string     `json:"useCompression" form:"useCompression" query:"useCompression" db:"useCompression"` // Y/N
	SecretKey          *string    `json:"secretKey" form:"secretKey" query:"secretKey" db:"secretKey"`
	BandwidthLimit     *string    `json:"bandwidthLimit" form:"bandwidthLimit" query:"bandwidthLimit" db:"bandwidthLimit"`
	MaxConnections     *int       `json:"maxConnections" form:"maxConnections" query:"maxConnections" db:"maxConnections"`
	HealthCheckType    *string    `json:"healthCheckType" form:"healthCheckType" query:"healthCheckType" db:"healthCheckType"`
	HealthCheckUrl     *string    `json:"healthCheckUrl" form:"healthCheckUrl" query:"healthCheckUrl" db:"healthCheckUrl"`
	ServiceStatus      string     `json:"serviceStatus" form:"serviceStatus" query:"serviceStatus" db:"serviceStatus"`
	RegisteredTime     time.Time  `json:"registeredTime" form:"registeredTime" query:"registeredTime" db:"registeredTime"`
	LastActiveTime     *time.Time `json:"lastActiveTime" form:"lastActiveTime" query:"lastActiveTime" db:"lastActiveTime"`
	ConnectionCount    int        `json:"connectionCount" form:"connectionCount" query:"connectionCount" db:"connectionCount"`
	TotalConnections   int64      `json:"totalConnections" form:"totalConnections" query:"totalConnections" db:"totalConnections"`
	TotalTraffic       int64      `json:"totalTraffic" form:"totalTraffic" query:"totalTraffic" db:"totalTraffic"`
	ServiceConfig      *string    `json:"serviceConfig" form:"serviceConfig" query:"serviceConfig" db:"serviceConfig"` // JSON格式

	// 通用字段
	AddTime        time.Time `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`
	AddWho         string    `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`
	EditTime       time.Time `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`
	EditWho        string    `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`
	OprSeqFlag     string    `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`
	CurrentVersion int       `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"`
	ActiveFlag     string    `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`
	NoteText       *string   `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`
	ExtProperty    *string   `json:"extProperty" form:"extProperty" query:"extProperty" db:"extProperty"`
}

// TunnelStaticServer 静态隧道服务器配置
// 对应数据库表: HUB_TUNNEL_STATIC_SERVER
// 管理静态端口转发服务配置
type TunnelStaticServer struct {
	TunnelStaticServerId   string              `json:"tunnelStaticServerId" form:"tunnelStaticServerId" query:"tunnelStaticServerId" db:"tunnelStaticServerId"`         // 静态隧道服务器ID，主键
	TenantId               string              `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                                                         // 租户ID
	ServerName             string              `json:"serverName" form:"serverName" query:"serverName" db:"serverName"`                                                 // 服务器名称
	ServerDescription      *string             `json:"serverDescription" form:"serverDescription" query:"serverDescription" db:"serverDescription"`                     // 服务器描述
	ListenAddress          string              `json:"listenAddress" form:"listenAddress" query:"listenAddress" db:"listenAddress"`                                     // 监听地址
	ListenPort             int                 `json:"listenPort" form:"listenPort" query:"listenPort" db:"listenPort"`                                                 // 监听端口（公网端口）
	ServerType             string              `json:"serverType" form:"serverType" query:"serverType" db:"serverType"`                                                 // 服务器类型(tcp,udp,http,https)
	MaxConnections         int                 `json:"maxConnections" form:"maxConnections" query:"maxConnections" db:"maxConnections"`                                 // 最大连接数
	ConnectionTimeout      int                 `json:"connectionTimeout" form:"connectionTimeout" query:"connectionTimeout" db:"connectionTimeout"`                     // 连接超时时间(秒)
	ReadTimeout            int                 `json:"readTimeout" form:"readTimeout" query:"readTimeout" db:"readTimeout"`                                             // 读取超时时间(秒)
	WriteTimeout           int                 `json:"writeTimeout" form:"writeTimeout" query:"writeTimeout" db:"writeTimeout"`                                         // 写入超时时间(秒)
	TlsEnable              string              `json:"tlsEnable" form:"tlsEnable" query:"tlsEnable" db:"tlsEnable"`                                                     // 启用TLS(N禁用,Y启用)
	TlsCertFile            *string             `json:"tlsCertFile" form:"tlsCertFile" query:"tlsCertFile" db:"tlsCertFile"`                                             // TLS证书文件路径
	TlsKeyFile             *string             `json:"tlsKeyFile" form:"tlsKeyFile" query:"tlsKeyFile" db:"tlsKeyFile"`                                                 // TLS私钥文件路径
	TlsCaFile              *string             `json:"tlsCaFile" form:"tlsCaFile" query:"tlsCaFile" db:"tlsCaFile"`                                                     // TLS CA证书文件路径
	LogLevel               string              `json:"logLevel" form:"logLevel" query:"logLevel" db:"logLevel"`                                                         // 日志级别(debug,info,warn,error)
	LogFile                *string             `json:"logFile" form:"logFile" query:"logFile" db:"logFile"`                                                             // 日志文件路径
	ServerStatus           string              `json:"serverStatus" form:"serverStatus" query:"serverStatus" db:"serverStatus"`                                         // 服务器状态(running,stopped,error)
	StartTime              *time.Time          `json:"startTime" form:"startTime" query:"startTime" db:"startTime"`                                                     // 服务启动时间
	StopTime               *time.Time          `json:"stopTime" form:"stopTime" query:"stopTime" db:"stopTime"`                                                         // 服务停止时间
	CurrentConnectionCount int                 `json:"currentConnectionCount" form:"currentConnectionCount" query:"currentConnectionCount" db:"currentConnectionCount"` // 当前连接数
	TotalConnectionCount   int64               `json:"totalConnectionCount" form:"totalConnectionCount" query:"totalConnectionCount" db:"totalConnectionCount"`         // 总连接数
	TotalBytesReceived     int64               `json:"totalBytesReceived" form:"totalBytesReceived" query:"totalBytesReceived" db:"totalBytesReceived"`                 // 总接收字节数
	TotalBytesSent         int64               `json:"totalBytesSent" form:"totalBytesSent" query:"totalBytesSent" db:"totalBytesSent"`                                 // 总发送字节数
	HealthCheckType        *string             `json:"healthCheckType" form:"healthCheckType" query:"healthCheckType" db:"healthCheckType"`                             // 健康检查类型(tcp,http,https)
	HealthCheckUrl         *string             `json:"healthCheckUrl" form:"healthCheckUrl" query:"healthCheckUrl" db:"healthCheckUrl"`                                 // 健康检查URL
	HealthCheckInterval    *int                `json:"healthCheckInterval" form:"healthCheckInterval" query:"healthCheckInterval" db:"healthCheckInterval"`             // 健康检查间隔(秒)
	HealthCheckTimeout     *int                `json:"healthCheckTimeout" form:"healthCheckTimeout" query:"healthCheckTimeout" db:"healthCheckTimeout"`                 // 健康检查超时(秒)
	HealthCheckMaxFailures *int                `json:"healthCheckMaxFailures" form:"healthCheckMaxFailures" query:"healthCheckMaxFailures" db:"healthCheckMaxFailures"` // 健康检查最大失败次数
	LoadBalanceType        *string             `json:"loadBalanceType" form:"loadBalanceType" query:"loadBalanceType" db:"loadBalanceType"`                             // 负载均衡类型(roundrobin,leastconn,random)
	ServerConfig           *string             `json:"serverConfig" form:"serverConfig" query:"serverConfig" db:"serverConfig"`                                         // 服务器配置，JSON格式
	Nodes                  []*TunnelStaticNode `json:"nodes" db:"-"`                                                                                                    // 后端节点列表（不存储到数据库）

	// 通用字段
	AddTime        time.Time `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`
	AddWho         string    `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`
	EditTime       time.Time `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`
	EditWho        string    `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`
	OprSeqFlag     string    `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`
	CurrentVersion int       `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"`
	ActiveFlag     string    `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`
	NoteText       *string   `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`
	ExtProperty    *string   `json:"extProperty" form:"extProperty" query:"extProperty" db:"extProperty"`

	// 关联数据（不存储到数据库）
	NodeCount int `json:"nodeCount" form:"-" query:"-" db:"-"`
}

// TunnelStaticNode 静态隧道节点配置
// 对应数据库表: HUB_TUNNEL_STATIC_NODE
// 管理静态隧道转发后端节点配置
type TunnelStaticNode struct {
	TunnelStaticNodeId     string     `json:"tunnelStaticNodeId" form:"tunnelStaticNodeId" query:"tunnelStaticNodeId" db:"tunnelStaticNodeId"`                 // 静态隧道节点ID，主键
	TenantId               string     `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                                                         // 租户ID
	TunnelStaticServerId   string     `json:"tunnelStaticServerId" form:"tunnelStaticServerId" query:"tunnelStaticServerId" db:"tunnelStaticServerId"`         // 静态隧道服务器ID，关联HUB_TUNNEL_STATIC_SERVER
	NodeName               string     `json:"nodeName" form:"nodeName" query:"nodeName" db:"nodeName"`                                                         // 节点名称
	NodeDescription        *string    `json:"nodeDescription" form:"nodeDescription" query:"nodeDescription" db:"nodeDescription"`                             // 节点描述
	TargetAddress          string     `json:"targetAddress" form:"targetAddress" query:"targetAddress" db:"targetAddress"`                                     // 目标地址（后端服务地址）
	TargetPort             int        `json:"targetPort" form:"targetPort" query:"targetPort" db:"targetPort"`                                                 // 目标端口（后端服务端口）
	ProxyType              string     `json:"proxyType" form:"proxyType" query:"proxyType" db:"proxyType"`                                                     // 代理类型(tcp,udp,http,https)
	MaxConnections         *int       `json:"maxConnections" form:"maxConnections" query:"maxConnections" db:"maxConnections"`                                 // 最大连接数
	ConnectionTimeout      *int       `json:"connectionTimeout" form:"connectionTimeout" query:"connectionTimeout" db:"connectionTimeout"`                     // 连接超时时间(秒)
	ReadTimeout            *int       `json:"readTimeout" form:"readTimeout" query:"readTimeout" db:"readTimeout"`                                             // 读取超时时间(秒)
	WriteTimeout           *int       `json:"writeTimeout" form:"writeTimeout" query:"writeTimeout" db:"writeTimeout"`                                         // 写入超时时间(秒)
	RetryCount             *int       `json:"retryCount" form:"retryCount" query:"retryCount" db:"retryCount"`                                                 // 重试次数
	RetryInterval          *int       `json:"retryInterval" form:"retryInterval" query:"retryInterval" db:"retryInterval"`                                     // 重试间隔(秒)
	Compression            string     `json:"compression" form:"compression" query:"compression" db:"compression"`                                             // 启用压缩(N禁用,Y启用)
	Encryption             string     `json:"encryption" form:"encryption" query:"encryption" db:"encryption"`                                                 // 启用加密(N禁用,Y启用)
	SecretKey              *string    `json:"secretKey" form:"secretKey" query:"secretKey" db:"secretKey"`                                                     // 加密密钥
	CustomHeaders          *string    `json:"customHeaders" form:"customHeaders" query:"customHeaders" db:"customHeaders"`                                     // 自定义HTTP头，JSON格式
	NodeStatus             string     `json:"nodeStatus" form:"nodeStatus" query:"nodeStatus" db:"nodeStatus"`                                                 // 节点状态(active,inactive,error)
	LastHealthCheck        *time.Time `json:"lastHealthCheck" form:"lastHealthCheck" query:"lastHealthCheck" db:"lastHealthCheck"`                             // 最后健康检查时间
	HealthCheckStatus      *string    `json:"healthCheckStatus" form:"healthCheckStatus" query:"healthCheckStatus" db:"healthCheckStatus"`                     // 健康检查状态(healthy,unhealthy,unknown)
	CurrentConnectionCount int        `json:"currentConnectionCount" form:"currentConnectionCount" query:"currentConnectionCount" db:"currentConnectionCount"` // 当前连接数
	TotalConnectionCount   int64      `json:"totalConnectionCount" form:"totalConnectionCount" query:"totalConnectionCount" db:"totalConnectionCount"`         // 总连接数
	TotalBytesReceived     int64      `json:"totalBytesReceived" form:"totalBytesReceived" query:"totalBytesReceived" db:"totalBytesReceived"`                 // 总接收字节数
	TotalBytesSent         int64      `json:"totalBytesSent" form:"totalBytesSent" query:"totalBytesSent" db:"totalBytesSent"`                                 // 总发送字节数
	FailureCount           int        `json:"failureCount" form:"failureCount" query:"failureCount" db:"failureCount"`                                         // 失败次数
	LastFailureTime        *time.Time `json:"lastFailureTime" form:"lastFailureTime" query:"lastFailureTime" db:"lastFailureTime"`                             // 最后失败时间
	NodeConfig             *string    `json:"nodeConfig" form:"nodeConfig" query:"nodeConfig" db:"nodeConfig"`                                                 // 节点配置，JSON格式

	// 通用字段
	AddTime        time.Time `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`
	AddWho         string    `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`
	EditTime       time.Time `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`
	EditWho        string    `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`
	OprSeqFlag     string    `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`
	CurrentVersion int       `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"`
	ActiveFlag     string    `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`
	NoteText       *string   `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`
	ExtProperty    *string   `json:"extProperty" form:"extProperty" query:"extProperty" db:"extProperty"`

	// 关联数据（不存储到数据库）
	ServerName string `json:"serverName" form:"-" query:"-" db:"-"`
}

// 常量定义
const (
	// 服务器状态
	ServerStatusRunning = "running"
	ServerStatusStopped = "stopped"
	ServerStatusError   = "error"

	// 节点类型
	NodeTypeStatic  = "static"
	NodeTypeDynamic = "dynamic"

	// 代理类型
	ProxyTypeTCP   = "tcp"
	ProxyTypeUDP   = "udp"
	ProxyTypeHTTP  = "http"
	ProxyTypeHTTPS = "https"
	ProxyTypeSTCP  = "stcp"
	ProxyTypeSUDP  = "sudp"
	ProxyTypeXTCP  = "xtcp"

	// 节点状态
	NodeStatusActive   = "active"
	NodeStatusInactive = "inactive"
	NodeStatusError    = "error"

	// 健康检查状态
	HealthCheckStatusHealthy   = "healthy"
	HealthCheckStatusUnhealthy = "unhealthy"
	HealthCheckStatusUnknown   = "unknown"

	// 负载均衡类型
	LoadBalanceTypeRoundRobin = "roundrobin"
	LoadBalanceTypeLeastConn  = "leastconn"
	LoadBalanceTypeRandom     = "random"

	// 连接状态
	ConnectionStatusConnected    = "connected"
	ConnectionStatusDisconnected = "disconnected"
	ConnectionStatusConnecting   = "connecting"
	ConnectionStatusError        = "error"

	// 服务状态
	ServiceStatusActive   = "active"
	ServiceStatusInactive = "inactive"
	ServiceStatusError    = "error"
	ServiceStatusOffline  = "offline"

	// 会话类型
	SessionTypeControl = "control"
	SessionTypeProxy   = "proxy"

	// 会话状态
	SessionStatusActive   = "active"
	SessionStatusInactive = "inactive"
	SessionStatusTimeout  = "timeout"
	SessionStatusClosed   = "closed"

	// 活动标记
	ActiveFlagYes = "Y"
	ActiveFlagNo  = "N"
)
