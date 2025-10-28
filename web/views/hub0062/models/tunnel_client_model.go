package models

import "time"

// TunnelClient 隧道客户端模型
// 对应数据库表：HUB_TUNNEL_CLIENT
// 用途：管理FRP客户端连接、认证和状态
type TunnelClient struct {
	// 主键信息
	TunnelClientId string `json:"tunnelClientId" form:"tunnelClientId" query:"tunnelClientId" db:"tunnelClientId"` // 客户端ID，主键
	TenantId       string `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                         // 租户ID
	UserId         string `json:"userId" form:"userId" query:"userId" db:"userId"`                                 // 用户ID

	// 基本信息
	ClientName        string `json:"clientName" form:"clientName" query:"clientName" db:"clientName"`                             // 客户端名称
	ClientDescription string `json:"clientDescription" form:"clientDescription" query:"clientDescription" db:"clientDescription"` // 客户端描述
	ClientVersion     string `json:"clientVersion" form:"clientVersion" query:"clientVersion" db:"clientVersion"`                 // 客户端版本
	OperatingSystem   string `json:"operatingSystem" form:"operatingSystem" query:"operatingSystem" db:"operatingSystem"`         // 操作系统
	ClientIpAddress   string `json:"clientIpAddress" form:"clientIpAddress" query:"clientIpAddress" db:"clientIpAddress"`         // 客户端IP地址

	// 服务器连接配置
	ServerAddress string `json:"serverAddress" form:"serverAddress" query:"serverAddress" db:"serverAddress"` // 服务器地址
	ServerPort    int    `json:"serverPort" form:"serverPort" query:"serverPort" db:"serverPort"`             // 服务器端口

	// 认证配置
	AuthToken string `json:"authToken" form:"authToken" query:"authToken" db:"authToken"` // 认证令牌
	TlsEnable string `json:"tlsEnable" form:"tlsEnable" query:"tlsEnable" db:"tlsEnable"` // 启用TLS(N禁用,Y启用)

	// 重连配置
	AutoReconnect    string `json:"autoReconnect" form:"autoReconnect" query:"autoReconnect" db:"autoReconnect"`             // 自动重连(N禁用,Y启用)
	MaxRetries       int    `json:"maxRetries" form:"maxRetries" query:"maxRetries" db:"maxRetries"`                         // 最大重试次数
	RetryInterval    int    `json:"retryInterval" form:"retryInterval" query:"retryInterval" db:"retryInterval"`             // 重试间隔(秒)
	ReconnectCount   int    `json:"reconnectCount" form:"reconnectCount" query:"reconnectCount" db:"reconnectCount"`         // 重连次数
	TotalConnectTime int64  `json:"totalConnectTime" form:"totalConnectTime" query:"totalConnectTime" db:"totalConnectTime"` // 总连接时长(秒)

	// 心跳配置
	HeartbeatInterval int        `json:"heartbeatInterval" form:"heartbeatInterval" query:"heartbeatInterval" db:"heartbeatInterval"` // 心跳间隔(秒)
	HeartbeatTimeout  int        `json:"heartbeatTimeout" form:"heartbeatTimeout" query:"heartbeatTimeout" db:"heartbeatTimeout"`     // 心跳超时(秒)
	LastHeartbeat     *time.Time `json:"lastHeartbeat" form:"lastHeartbeat" query:"lastHeartbeat" db:"lastHeartbeat"`                 // 最后心跳时间

	// 连接状态
	ConnectionStatus   string     `json:"connectionStatus" form:"connectionStatus" query:"connectionStatus" db:"connectionStatus"`         // 连接状态(connected,disconnected,connecting,error)
	LastConnectTime    *time.Time `json:"lastConnectTime" form:"lastConnectTime" query:"lastConnectTime" db:"lastConnectTime"`             // 最后连接时间
	LastDisconnectTime *time.Time `json:"lastDisconnectTime" form:"lastDisconnectTime" query:"lastDisconnectTime" db:"lastDisconnectTime"` // 最后断开时间
	DisconnectReason   string     `json:"disconnectReason" form:"disconnectReason" query:"disconnectReason" db:"disconnectReason"`         // 断开原因

	// 服务统计
	ServiceCount int `json:"serviceCount" form:"serviceCount" query:"serviceCount" db:"serviceCount"` // 服务数量

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

// TunnelClientQueryRequest 客户端查询请求
type TunnelClientQueryRequest struct {
	ClientName       string `json:"clientName" form:"clientName"`             // 客户端名称过滤（模糊查询）
	ConnectionStatus string `json:"connectionStatus" form:"connectionStatus"` // 连接状态过滤
	UserId           string `json:"userId" form:"userId"`                     // 用户ID过滤
	ServerAddress    string `json:"serverAddress" form:"serverAddress"`       // 服务器地址过滤
	ActiveFlag       string `json:"activeFlag" form:"activeFlag"`             // 活动状态标记(Y活动,N非活动,空为全部)
	Keyword          string `json:"keyword" form:"keyword"`                   // 关键字搜索（客户端名称、IP地址）

	// 分页参数
	PageIndex int `json:"pageIndex" form:"pageIndex"` // 页码，从1开始
	PageSize  int `json:"pageSize" form:"pageSize"`   // 每页数量，默认20
}

// TunnelClientStats 客户端统计信息
type TunnelClientStats struct {
	TotalClients        int `json:"totalClients"`        // 总客户端数量
	ConnectedClients    int `json:"connectedClients"`    // 已连接客户端数量
	DisconnectedClients int `json:"disconnectedClients"` // 已断开客户端数量
	ErrorClients        int `json:"errorClients"`        // 错误状态客户端数量
	TotalServices       int `json:"totalServices"`       // 总服务数量
	TotalReconnects     int `json:"totalReconnects"`     // 总重连次数
}

// ClientStatusResponse 客户端状态响应
type ClientStatusResponse struct {
	TunnelClientId   string     `json:"tunnelClientId"`   // 客户端ID
	ClientName       string     `json:"clientName"`       // 客户端名称
	ConnectionStatus string     `json:"connectionStatus"` // 连接状态
	LastConnectTime  *time.Time `json:"lastConnectTime"`  // 最后连接时间
	LastHeartbeat    *time.Time `json:"lastHeartbeat"`    // 最后心跳时间
	ServiceCount     int        `json:"serviceCount"`     // 服务数量
	TotalConnectTime int64      `json:"totalConnectTime"` // 总连接时长(秒)
	ReconnectCount   int        `json:"reconnectCount"`   // 重连次数
}

// ResetAuthTokenRequest 重置认证令牌请求
type ResetAuthTokenRequest struct {
	TunnelClientId string `json:"tunnelClientId" form:"tunnelClientId" binding:"required"` // 客户端ID
}

// ResetAuthTokenResponse 重置认证令牌响应
type ResetAuthTokenResponse struct {
	TunnelClientId string `json:"tunnelClientId"` // 客户端ID
	AuthToken      string `json:"authToken"`      // 新的认证令牌
}

// DisconnectClientRequest 断开客户端连接请求
type DisconnectClientRequest struct {
	TunnelClientId string `json:"tunnelClientId" form:"tunnelClientId" binding:"required"` // 客户端ID
	Reason         string `json:"reason" form:"reason"`                                    // 断开原因
}

// BatchOperationRequest 批量操作请求
type BatchOperationRequest struct {
	ClientIds []string `json:"clientIds" form:"clientIds" binding:"required"` // 客户端ID列表
}

// BatchOperationResponse 批量操作响应
type BatchOperationResponse struct {
	SuccessCount int      `json:"successCount"` // 成功数量
	FailedCount  int      `json:"failedCount"`  // 失败数量
	FailedIds    []string `json:"failedIds"`    // 失败的客户端ID列表
}
