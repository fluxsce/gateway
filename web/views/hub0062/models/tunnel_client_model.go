package models

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
	TotalClients        int `json:"totalClients" db:"totalClients"`               // 总客户端数量
	ConnectedClients    int `json:"connectedClients" db:"connectedClients"`       // 已连接客户端数量
	DisconnectedClients int `json:"disconnectedClients" db:"disconnectedClients"` // 已断开客户端数量
	ErrorClients        int `json:"errorClients" db:"errorClients"`               // 错误状态客户端数量
	TotalServices       int `json:"totalServices" db:"totalServices"`             // 总服务数量
	TotalReconnects     int `json:"totalReconnects" db:"totalReconnects"`         // 总重连次数
}

// TunnelServiceQueryRequest 服务查询请求
type TunnelServiceQueryRequest struct {
	TenantId       string `json:"tenantId" form:"tenantId"`             // 租户ID过滤
	TunnelClientId string `json:"tunnelClientId" form:"tunnelClientId"` // 客户端ID过滤
	ServiceName    string `json:"serviceName" form:"serviceName"`       // 服务名称过滤（模糊查询）
	ServiceType    string `json:"serviceType" form:"serviceType"`       // 服务类型过滤
	ServiceStatus  string `json:"serviceStatus" form:"serviceStatus"`   // 服务状态过滤
	UserId         string `json:"userId" form:"userId"`                 // 用户ID过滤
	ActiveFlag     string `json:"activeFlag" form:"activeFlag"`         // 活动状态标记(Y活动,N非活动,空为全部)
	Keyword        string `json:"keyword" form:"keyword"`               // 关键字搜索（服务名称、本地地址）

	// 分页参数
	PageIndex int `json:"pageIndex" form:"pageIndex"` // 页码，从1开始
	PageSize  int `json:"pageSize" form:"pageSize"`   // 每页数量，默认20
}

// TunnelServiceStats 服务统计信息
type TunnelServiceStats struct {
	TotalServices    int   `json:"totalServices" db:"totalServices"`       // 总服务数量
	ActiveServices   int   `json:"activeServices" db:"activeServices"`     // 活跃服务数量
	InactiveServices int   `json:"inactiveServices" db:"inactiveServices"` // 非活跃服务数量
	ErrorServices    int   `json:"errorServices" db:"errorServices"`       // 错误服务数量
	TotalConnections int64 `json:"totalConnections" db:"totalConnections"` // 总连接数
	TotalTraffic     int64 `json:"totalTraffic" db:"totalTraffic"`         // 总流量
}
