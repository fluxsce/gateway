package models

import "time"

// TunnelClient 隧道客户端模型
// 对应数据库表：TUNNEL_CLIENT
// 用途：管理隧道客户端信息
type TunnelClient struct {
	// 主键信息
	TunnelClientId string `json:"tunnelClientId" form:"tunnelClientId" query:"tunnelClientId" db:"tunnelClientId"` // 隧道客户端ID，主键
	ClientName     string `json:"clientName" form:"clientName" query:"clientName" db:"clientName"`                 // 客户端名称
	ClientAddress  string `json:"clientAddress" form:"clientAddress" query:"clientAddress" db:"clientAddress"`     // 客户端地址

	// 关联服务器
	TunnelServerId string `json:"tunnelServerId" form:"tunnelServerId" query:"tunnelServerId" db:"tunnelServerId"` // 关联的隧道服务器ID
	ServerName     string `json:"serverName" form:"serverName" query:"serverName" db:"serverName"`                 // 服务器名称（冗余字段）

	// 客户端配置
	AuthToken         string `json:"authToken" form:"authToken" query:"authToken" db:"authToken"`                                 // 认证令牌
	HeartbeatInterval int    `json:"heartbeatInterval" form:"heartbeatInterval" query:"heartbeatInterval" db:"heartbeatInterval"` // 心跳间隔(秒)
	MaxRetries        int    `json:"maxRetries" form:"maxRetries" query:"maxRetries" db:"maxRetries"`                             // 最大重试次数
	RetryInterval     int    `json:"retryInterval" form:"retryInterval" query:"retryInterval" db:"retryInterval"`                 // 重试间隔(秒)

	// 状态信息
	ClientStatus    string     `json:"clientStatus" form:"clientStatus" query:"clientStatus" db:"clientStatus"`             // 客户端状态(CONNECTED,DISCONNECTED,RECONNECTING)
	LastConnectTime *time.Time `json:"lastConnectTime" form:"lastConnectTime" query:"lastConnectTime" db:"lastConnectTime"` // 最后连接时间
	LastHeartbeat   *time.Time `json:"lastHeartbeat" form:"lastHeartbeat" query:"lastHeartbeat" db:"lastHeartbeat"`         // 最后心跳时间
	ReconnectCount  int        `json:"reconnectCount" form:"reconnectCount" query:"reconnectCount" db:"reconnectCount"`     // 重连次数

	// 统计信息
	RegisteredServices int   `json:"registeredServices" form:"registeredServices" query:"registeredServices" db:"registeredServices"` // 注册的服务数量
	ActiveProxies      int   `json:"activeProxies" form:"activeProxies" query:"activeProxies" db:"activeProxies"`                     // 活跃代理数量
	TotalTraffic       int64 `json:"totalTraffic" form:"totalTraffic" query:"totalTraffic" db:"totalTraffic"`                         // 总流量(字节)

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

// TunnelClientStats 隧道客户端统计信息
type TunnelClientStats struct {
	TotalClients        int `json:"totalClients"`        // 总客户端数量
	ConnectedClients    int `json:"connectedClients"`    // 已连接客户端数量
	DisconnectedClients int `json:"disconnectedClients"` // 断开连接客户端数量
	TotalServices       int `json:"totalServices"`       // 总服务数量
	ActiveServices      int `json:"activeServices"`      // 活跃服务数量
}
