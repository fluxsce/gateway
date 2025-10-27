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
	TunnelServerId    string     `json:"tunnelServerId" db:"tunnelServerId"`
	TenantId          string     `json:"tenantId" db:"tenantId"`
	ServerName        string     `json:"serverName" db:"serverName"`
	ServerDescription string     `json:"serverDescription" db:"serverDescription"`
	ControlAddress    string     `json:"controlAddress" db:"controlAddress"`
	ControlPort       int        `json:"controlPort" db:"controlPort"`
	DashboardPort     *int       `json:"dashboardPort" db:"dashboardPort"`
	VhostHttpPort     *int       `json:"vhostHttpPort" db:"vhostHttpPort"`
	VhostHttpsPort    *int       `json:"vhostHttpsPort" db:"vhostHttpsPort"`
	MaxClients        int        `json:"maxClients" db:"maxClients"`
	TokenAuth         string     `json:"tokenAuth" db:"tokenAuth"` // Y/N
	AuthToken         string     `json:"authToken" db:"authToken"`
	TlsEnable         string     `json:"tlsEnable" db:"tlsEnable"` // Y/N
	TlsCertFile       *string    `json:"tlsCertFile" db:"tlsCertFile"`
	TlsKeyFile        *string    `json:"tlsKeyFile" db:"tlsKeyFile"`
	HeartbeatInterval int        `json:"heartbeatInterval" db:"heartbeatInterval"`
	HeartbeatTimeout  int        `json:"heartbeatTimeout" db:"heartbeatTimeout"`
	LogLevel          string     `json:"logLevel" db:"logLevel"`
	MaxPortsPerClient *int       `json:"maxPortsPerClient" db:"maxPortsPerClient"`
	AllowPorts        *string    `json:"allowPorts" db:"allowPorts"` // JSON格式
	ServerStatus      string     `json:"serverStatus" db:"serverStatus"`
	StartTime         *time.Time `json:"startTime" db:"startTime"`
	ConfigVersion     *string    `json:"configVersion" db:"configVersion"`

	// 通用字段
	AddTime        time.Time `json:"addTime" db:"addTime"`
	AddWho         string    `json:"addWho" db:"addWho"`
	EditTime       time.Time `json:"editTime" db:"editTime"`
	EditWho        string    `json:"editWho" db:"editWho"`
	OprSeqFlag     string    `json:"oprSeqFlag" db:"oprSeqFlag"`
	CurrentVersion int       `json:"currentVersion" db:"currentVersion"`
	ActiveFlag     string    `json:"activeFlag" db:"activeFlag"`
	NoteText       *string   `json:"noteText" db:"noteText"`
	ExtProperty    *string   `json:"extProperty" db:"extProperty"`
}

// TunnelServerNode 隧道服务器节点配置 (静态端口映射)
// 对应数据库表: HUB_TUNNEL_SERVER_NODE
// 对应FRP: frps静态代理配置
type TunnelServerNode struct {
	ServerNodeId        string     `json:"serverNodeId" db:"serverNodeId"`
	TenantId            string     `json:"tenantId" db:"tenantId"`
	TunnelServerId      string     `json:"tunnelServerId" db:"tunnelServerId"`
	NodeName            string     `json:"nodeName" db:"nodeName"`
	NodeType            string     `json:"nodeType" db:"nodeType"`   // static, dynamic
	ProxyType           string     `json:"proxyType" db:"proxyType"` // tcp, udp, http, https, stcp, sudp
	ListenAddress       string     `json:"listenAddress" db:"listenAddress"`
	ListenPort          int        `json:"listenPort" db:"listenPort"`
	TargetAddress       string     `json:"targetAddress" db:"targetAddress"`
	TargetPort          int        `json:"targetPort" db:"targetPort"`
	CustomDomains       *string    `json:"customDomains" db:"customDomains"` // JSON格式
	SubDomain           *string    `json:"subDomain" db:"subDomain"`
	HttpUser            *string    `json:"httpUser" db:"httpUser"`
	HttpPassword        *string    `json:"httpPassword" db:"httpPassword"`
	HostHeaderRewrite   *string    `json:"hostHeaderRewrite" db:"hostHeaderRewrite"`
	Headers             *string    `json:"headers" db:"headers"`         // JSON格式
	Locations           *string    `json:"locations" db:"locations"`     // JSON格式
	Compression         string     `json:"compression" db:"compression"` // Y/N
	Encryption          string     `json:"encryption" db:"encryption"`   // Y/N
	SecretKey           *string    `json:"secretKey" db:"secretKey"`
	HealthCheckType     *string    `json:"healthCheckType" db:"healthCheckType"`
	HealthCheckUrl      *string    `json:"healthCheckUrl" db:"healthCheckUrl"`
	HealthCheckInterval *int       `json:"healthCheckInterval" db:"healthCheckInterval"`
	MaxConnections      *int       `json:"maxConnections" db:"maxConnections"`
	NodeStatus          string     `json:"nodeStatus" db:"nodeStatus"`
	LastHealthCheck     *time.Time `json:"lastHealthCheck" db:"lastHealthCheck"`
	ConnectionCount     int        `json:"connectionCount" db:"connectionCount"`
	TotalConnections    int64      `json:"totalConnections" db:"totalConnections"`
	TotalBytes          int64      `json:"totalBytes" db:"totalBytes"`
	CreatedTime         time.Time  `json:"createdTime" db:"createdTime"`

	// 通用字段
	AddTime        time.Time `json:"addTime" db:"addTime"`
	AddWho         string    `json:"addWho" db:"addWho"`
	EditTime       time.Time `json:"editTime" db:"editTime"`
	EditWho        string    `json:"editWho" db:"editWho"`
	OprSeqFlag     string    `json:"oprSeqFlag" db:"oprSeqFlag"`
	CurrentVersion int       `json:"currentVersion" db:"currentVersion"`
	ActiveFlag     string    `json:"activeFlag" db:"activeFlag"`
	NoteText       *string   `json:"noteText" db:"noteText"`
	ExtProperty    *string   `json:"extProperty" db:"extProperty"`
}

// TunnelClient 隧道客户端信息
// 对应数据库表: HUB_TUNNEL_CLIENT
// 对应FRP: frpc客户端连接
type TunnelClient struct {
	TunnelClientId     string     `json:"tunnelClientId" db:"tunnelClientId"`
	TenantId           string     `json:"tenantId" db:"tenantId"`
	UserId             string     `json:"userId" db:"userId"`
	ClientName         string     `json:"clientName" db:"clientName"`
	ClientDescription  *string    `json:"clientDescription" db:"clientDescription"`
	ClientVersion      *string    `json:"clientVersion" db:"clientVersion"`
	OperatingSystem    *string    `json:"operatingSystem" db:"operatingSystem"`
	ClientIpAddress    *string    `json:"clientIpAddress" db:"clientIpAddress"`
	ClientMacAddress   *string    `json:"clientMacAddress" db:"clientMacAddress"`
	ServerAddress      string     `json:"serverAddress" db:"serverAddress"`
	ServerPort         int        `json:"serverPort" db:"serverPort"`
	AuthToken          string     `json:"authToken" db:"authToken"`
	TlsEnable          string     `json:"tlsEnable" db:"tlsEnable"`         // Y/N
	AutoReconnect      string     `json:"autoReconnect" db:"autoReconnect"` // Y/N
	MaxRetries         int        `json:"maxRetries" db:"maxRetries"`
	RetryInterval      int        `json:"retryInterval" db:"retryInterval"`
	HeartbeatInterval  int        `json:"heartbeatInterval" db:"heartbeatInterval"`
	HeartbeatTimeout   int        `json:"heartbeatTimeout" db:"heartbeatTimeout"`
	ConnectionStatus   string     `json:"connectionStatus" db:"connectionStatus"`
	LastConnectTime    *time.Time `json:"lastConnectTime" db:"lastConnectTime"`
	LastDisconnectTime *time.Time `json:"lastDisconnectTime" db:"lastDisconnectTime"`
	TotalConnectTime   int64      `json:"totalConnectTime" db:"totalConnectTime"`
	ReconnectCount     int        `json:"reconnectCount" db:"reconnectCount"`
	ServiceCount       int        `json:"serviceCount" db:"serviceCount"`
	LastHeartbeat      *time.Time `json:"lastHeartbeat" db:"lastHeartbeat"`
	ClientConfig       *string    `json:"clientConfig" db:"clientConfig"` // JSON格式

	// 通用字段
	AddTime        time.Time `json:"addTime" db:"addTime"`
	AddWho         string    `json:"addWho" db:"addWho"`
	EditTime       time.Time `json:"editTime" db:"editTime"`
	EditWho        string    `json:"editWho" db:"editWho"`
	OprSeqFlag     string    `json:"oprSeqFlag" db:"oprSeqFlag"`
	CurrentVersion int       `json:"currentVersion" db:"currentVersion"`
	ActiveFlag     string    `json:"activeFlag" db:"activeFlag"`
	NoteText       *string   `json:"noteText" db:"noteText"`
	ExtProperty    *string   `json:"extProperty" db:"extProperty"`
}

// TunnelService 隧道服务配置 (动态注册的服务)
// 对应数据库表: HUB_TUNNEL_SERVICE
// 对应FRP: frpc服务配置 [web], [ssh] 等
type TunnelService struct {
	TunnelServiceId    string     `json:"tunnelServiceId" db:"tunnelServiceId"`
	TenantId           string     `json:"tenantId" db:"tenantId"`
	TunnelClientId     string     `json:"tunnelClientId" db:"tunnelClientId"`
	UserId             string     `json:"userId" db:"userId"`
	ServiceName        string     `json:"serviceName" db:"serviceName"`
	ServiceDescription *string    `json:"serviceDescription" db:"serviceDescription"`
	ServiceType        string     `json:"serviceType" db:"serviceType"` // tcp, udp, http, https, stcp, sudp, xtcp
	LocalAddress       string     `json:"localAddress" db:"localAddress"`
	LocalPort          int        `json:"localPort" db:"localPort"`
	RemotePort         *int       `json:"remotePort" db:"remotePort"`
	CustomDomains      *string    `json:"customDomains" db:"customDomains"` // JSON格式
	SubDomain          *string    `json:"subDomain" db:"subDomain"`
	HttpUser           *string    `json:"httpUser" db:"httpUser"`
	HttpPassword       *string    `json:"httpPassword" db:"httpPassword"`
	HostHeaderRewrite  *string    `json:"hostHeaderRewrite" db:"hostHeaderRewrite"`
	Headers            *string    `json:"headers" db:"headers"`               // JSON格式
	Locations          *string    `json:"locations" db:"locations"`           // JSON格式
	UseEncryption      string     `json:"useEncryption" db:"useEncryption"`   // Y/N
	UseCompression     string     `json:"useCompression" db:"useCompression"` // Y/N
	SecretKey          *string    `json:"secretKey" db:"secretKey"`
	BandwidthLimit     *string    `json:"bandwidthLimit" db:"bandwidthLimit"`
	MaxConnections     *int       `json:"maxConnections" db:"maxConnections"`
	HealthCheckType    *string    `json:"healthCheckType" db:"healthCheckType"`
	HealthCheckUrl     *string    `json:"healthCheckUrl" db:"healthCheckUrl"`
	ServiceStatus      string     `json:"serviceStatus" db:"serviceStatus"`
	RegisteredTime     time.Time  `json:"registeredTime" db:"registeredTime"`
	LastActiveTime     *time.Time `json:"lastActiveTime" db:"lastActiveTime"`
	ConnectionCount    int        `json:"connectionCount" db:"connectionCount"`
	TotalConnections   int64      `json:"totalConnections" db:"totalConnections"`
	TotalTraffic       int64      `json:"totalTraffic" db:"totalTraffic"`
	ServiceConfig      *string    `json:"serviceConfig" db:"serviceConfig"` // JSON格式

	// 通用字段
	AddTime        time.Time `json:"addTime" db:"addTime"`
	AddWho         string    `json:"addWho" db:"addWho"`
	EditTime       time.Time `json:"editTime" db:"editTime"`
	EditWho        string    `json:"editWho" db:"editWho"`
	OprSeqFlag     string    `json:"oprSeqFlag" db:"oprSeqFlag"`
	CurrentVersion int       `json:"currentVersion" db:"currentVersion"`
	ActiveFlag     string    `json:"activeFlag" db:"activeFlag"`
	NoteText       *string   `json:"noteText" db:"noteText"`
	ExtProperty    *string   `json:"extProperty" db:"extProperty"`
}

// TunnelSession 隧道会话信息
// 对应数据库表: HUB_TUNNEL_SESSION
// 对应FRP: 控制连接会话
type TunnelSession struct {
	TunnelSessionId      string     `json:"tunnelSessionId" db:"tunnelSessionId"`
	TenantId             string     `json:"tenantId" db:"tenantId"`
	TunnelClientId       string     `json:"tunnelClientId" db:"tunnelClientId"`
	SessionToken         string     `json:"sessionToken" db:"sessionToken"`
	SessionType          string     `json:"sessionType" db:"sessionType"` // control, proxy
	ClientIpAddress      string     `json:"clientIpAddress" db:"clientIpAddress"`
	ClientPort           int        `json:"clientPort" db:"clientPort"`
	ServerIpAddress      string     `json:"serverIpAddress" db:"serverIpAddress"`
	ServerPort           int        `json:"serverPort" db:"serverPort"`
	SessionStatus        string     `json:"sessionStatus" db:"sessionStatus"`
	StartTime            time.Time  `json:"startTime" db:"startTime"`
	LastActivityTime     *time.Time `json:"lastActivityTime" db:"lastActivityTime"`
	EndTime              *time.Time `json:"endTime" db:"endTime"`
	SessionDuration      int64      `json:"sessionDuration" db:"sessionDuration"`
	HeartbeatInterval    *int       `json:"heartbeatInterval" db:"heartbeatInterval"`
	HeartbeatCount       int        `json:"heartbeatCount" db:"heartbeatCount"`
	LastHeartbeatTime    *time.Time `json:"lastHeartbeatTime" db:"lastHeartbeatTime"`
	ProxyCount           int        `json:"proxyCount" db:"proxyCount"`
	TotalDataTransferred int64      `json:"totalDataTransferred" db:"totalDataTransferred"`
	AverageLatency       float64    `json:"averageLatency" db:"averageLatency"`
	SessionMetadata      *string    `json:"sessionMetadata" db:"sessionMetadata"` // JSON格式

	// 通用字段
	AddTime        time.Time `json:"addTime" db:"addTime"`
	AddWho         string    `json:"addWho" db:"addWho"`
	EditTime       time.Time `json:"editTime" db:"editTime"`
	EditWho        string    `json:"editWho" db:"editWho"`
	OprSeqFlag     string    `json:"oprSeqFlag" db:"oprSeqFlag"`
	CurrentVersion int       `json:"currentVersion" db:"currentVersion"`
	ActiveFlag     string    `json:"activeFlag" db:"activeFlag"`
	NoteText       *string   `json:"noteText" db:"noteText"`
	ExtProperty    *string   `json:"extProperty" db:"extProperty"`
}

// TunnelConnection 隧道连接信息
// 对应数据库表: HUB_TUNNEL_CONNECTION
// 对应FRP: 实际的数据代理连接
type TunnelConnection struct {
	TunnelConnectionId string     `json:"tunnelConnectionId" db:"tunnelConnectionId"`
	TenantId           string     `json:"tenantId" db:"tenantId"`
	TunnelSessionId    string     `json:"tunnelSessionId" db:"tunnelSessionId"`
	TunnelServiceId    *string    `json:"tunnelServiceId" db:"tunnelServiceId"`
	ServerNodeId       *string    `json:"serverNodeId" db:"serverNodeId"`
	ConnectionType     string     `json:"connectionType" db:"connectionType"` // control, proxy
	ProxyType          string     `json:"proxyType" db:"proxyType"`
	SourceIpAddress    string     `json:"sourceIpAddress" db:"sourceIpAddress"`
	SourcePort         int        `json:"sourcePort" db:"sourcePort"`
	TargetIpAddress    string     `json:"targetIpAddress" db:"targetIpAddress"`
	TargetPort         int        `json:"targetPort" db:"targetPort"`
	ProxyIpAddress     *string    `json:"proxyIpAddress" db:"proxyIpAddress"`
	ProxyPort          *int       `json:"proxyPort" db:"proxyPort"`
	ConnectionStatus   string     `json:"connectionStatus" db:"connectionStatus"`
	StartTime          time.Time  `json:"startTime" db:"startTime"`
	EndTime            *time.Time `json:"endTime" db:"endTime"`
	ConnectionDuration int64      `json:"connectionDuration" db:"connectionDuration"`
	BytesReceived      int64      `json:"bytesReceived" db:"bytesReceived"`
	BytesSent          int64      `json:"bytesSent" db:"bytesSent"`
	PacketsReceived    int64      `json:"packetsReceived" db:"packetsReceived"`
	PacketsSent        int64      `json:"packetsSent" db:"packetsSent"`
	LastActivity       *time.Time `json:"lastActivity" db:"lastActivity"`
	ErrorCount         int        `json:"errorCount" db:"errorCount"`
	LastErrorMessage   *string    `json:"lastErrorMessage" db:"lastErrorMessage"`
	ConnectionLatency  float64    `json:"connectionLatency" db:"connectionLatency"`
	UserAgent          *string    `json:"userAgent" db:"userAgent"`
	Referer            *string    `json:"referer" db:"referer"`
	HttpMethod         *string    `json:"httpMethod" db:"httpMethod"`
	HttpStatus         *int       `json:"httpStatus" db:"httpStatus"`
	ConnectionMetadata *string    `json:"connectionMetadata" db:"connectionMetadata"` // JSON格式

	// 通用字段
	AddTime        time.Time `json:"addTime" db:"addTime"`
	AddWho         string    `json:"addWho" db:"addWho"`
	EditTime       time.Time `json:"editTime" db:"editTime"`
	EditWho        string    `json:"editWho" db:"editWho"`
	OprSeqFlag     string    `json:"oprSeqFlag" db:"oprSeqFlag"`
	CurrentVersion int       `json:"currentVersion" db:"currentVersion"`
	ActiveFlag     string    `json:"activeFlag" db:"activeFlag"`
	NoteText       *string   `json:"noteText" db:"noteText"`
	ExtProperty    *string   `json:"extProperty" db:"extProperty"`
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

	// 控制消息类型
	MessageTypeAuth              = "auth"
	MessageTypeHeartbeat         = "heartbeat"
	MessageTypeRegisterService   = "register_service"
	MessageTypeUnregisterService = "unregister_service"
	MessageTypeProxyRequest      = "proxy_request"
	MessageTypeResponse          = "response"
)

// ControlMessage 控制消息结构
type ControlMessage struct {
	Type      string                 `json:"type"`
	SessionID string                 `json:"sessionId"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}
