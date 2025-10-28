package models

import "time"

// TunnelServerNode 隧道服务器节点模型（静态端口映射）
// 对应数据库表：HUB_TUNNEL_SERVER_NODE
// 用途：管理静态端口映射配置
type TunnelServerNode struct {
	// 主键信息
	ServerNodeId   string `json:"serverNodeId" form:"serverNodeId" query:"serverNodeId" db:"serverNodeId"`         // 服务器节点ID，主键
	TenantId       string `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                         // 租户ID
	TunnelServerId string `json:"tunnelServerId" form:"tunnelServerId" query:"tunnelServerId" db:"tunnelServerId"` // 隧道服务器ID
	NodeName       string `json:"nodeName" form:"nodeName" query:"nodeName" db:"nodeName"`                         // 节点名称
	NodeType       string `json:"nodeType" form:"nodeType" query:"nodeType" db:"nodeType"`                         // 节点类型(static,dynamic)

	// 代理配置
	ProxyType     string `json:"proxyType" form:"proxyType" query:"proxyType" db:"proxyType"`                 // 代理类型(tcp,udp,http,https,stcp,sudp)
	ListenAddress string `json:"listenAddress" form:"listenAddress" query:"listenAddress" db:"listenAddress"` // 监听地址
	ListenPort    int    `json:"listenPort" form:"listenPort" query:"listenPort" db:"listenPort"`             // 监听端口（公网端口）
	TargetAddress string `json:"targetAddress" form:"targetAddress" query:"targetAddress" db:"targetAddress"` // 目标地址（内网地址）
	TargetPort    int    `json:"targetPort" form:"targetPort" query:"targetPort" db:"targetPort"`             // 目标端口（内网端口）

	// HTTP配置
	CustomDomains     string `json:"customDomains" form:"customDomains" query:"customDomains" db:"customDomains"`                 // 自定义域名列表，JSON格式
	SubDomain         string `json:"subDomain" form:"subDomain" query:"subDomain" db:"subDomain"`                                 // 子域名
	HttpUser          string `json:"httpUser" form:"httpUser" query:"httpUser" db:"httpUser"`                                     // HTTP基础认证用户名
	HttpPassword      string `json:"httpPassword" form:"httpPassword" query:"httpPassword" db:"httpPassword"`                     // HTTP基础认证密码
	HostHeaderRewrite string `json:"hostHeaderRewrite" form:"hostHeaderRewrite" query:"hostHeaderRewrite" db:"hostHeaderRewrite"` // 重写Host头
	Headers           string `json:"headers" form:"headers" query:"headers" db:"headers"`                                         // 自定义HTTP头，JSON格式
	Locations         string `json:"locations" form:"locations" query:"locations" db:"locations"`                                 // HTTP路径配置，JSON格式

	// 安全配置
	Compression string `json:"compression" form:"compression" query:"compression" db:"compression"` // 启用压缩(N禁用,Y启用)
	Encryption  string `json:"encryption" form:"encryption" query:"encryption" db:"encryption"`     // 启用加密(N禁用,Y启用)
	SecretKey   string `json:"secretKey" form:"secretKey" query:"secretKey" db:"secretKey"`         // 加密密钥

	// 健康检查配置
	HealthCheckType     string `json:"healthCheckType" form:"healthCheckType" query:"healthCheckType" db:"healthCheckType"`                 // 健康检查类型(tcp,http)
	HealthCheckUrl      string `json:"healthCheckUrl" form:"healthCheckUrl" query:"healthCheckUrl" db:"healthCheckUrl"`                     // 健康检查URL
	HealthCheckInterval int    `json:"healthCheckInterval" form:"healthCheckInterval" query:"healthCheckInterval" db:"healthCheckInterval"` // 健康检查间隔(秒)

	// 限制配置
	MaxConnections int `json:"maxConnections" form:"maxConnections" query:"maxConnections" db:"maxConnections"` // 最大连接数

	// 状态信息
	NodeStatus       string     `json:"nodeStatus" form:"nodeStatus" query:"nodeStatus" db:"nodeStatus"`                         // 节点状态(active,inactive,error)
	LastHealthCheck  *time.Time `json:"lastHealthCheck" form:"lastHealthCheck" query:"lastHealthCheck" db:"lastHealthCheck"`     // 最后健康检查时间
	ConnectionCount  int        `json:"connectionCount" form:"connectionCount" query:"connectionCount" db:"connectionCount"`     // 当前连接数
	TotalConnections int64      `json:"totalConnections" form:"totalConnections" query:"totalConnections" db:"totalConnections"` // 总连接数
	TotalBytes       int64      `json:"totalBytes" form:"totalBytes" query:"totalBytes" db:"totalBytes"`                         // 总传输字节数
	CreatedTime      time.Time  `json:"createdTime" form:"createdTime" query:"createdTime" db:"createdTime"`                     // 节点创建时间

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

// ServerNodeQueryRequest 服务器节点查询请求
type ServerNodeQueryRequest struct {
	TunnelServerId string `json:"tunnelServerId" form:"tunnelServerId"` // 隧道服务器ID过滤
	NodeName       string `json:"nodeName" form:"nodeName"`             // 节点名称过滤（模糊查询）
	ProxyType      string `json:"proxyType" form:"proxyType"`           // 代理类型过滤
	NodeStatus     string `json:"nodeStatus" form:"nodeStatus"`         // 节点状态过滤
	NodeType       string `json:"nodeType" form:"nodeType"`             // 节点类型过滤
	ActiveFlag     string `json:"activeFlag" form:"activeFlag"`         // 活动状态标记(Y活动,N非活动,空为全部)
	Keyword        string `json:"keyword" form:"keyword"`               // 关键字搜索（节点名称、目标地址）

	// 分页参数
	PageIndex int `json:"pageIndex" form:"pageIndex"` // 页码，从1开始
	PageSize  int `json:"pageSize" form:"pageSize"`   // 每页数量，默认20
}

// ServerNodeStats 服务器节点统计信息
type ServerNodeStats struct {
	TotalNodes       int   `json:"totalNodes"`       // 总节点数量
	ActiveNodes      int   `json:"activeNodes"`      // 活跃节点数量
	InactiveNodes    int   `json:"inactiveNodes"`    // 非活跃节点数量
	TotalConnections int64 `json:"totalConnections"` // 总连接数
	TotalTraffic     int64 `json:"totalTraffic"`     // 总流量(字节)
}
