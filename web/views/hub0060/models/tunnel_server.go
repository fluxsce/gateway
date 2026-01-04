package models

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
	TotalServers     int `json:"totalServers"`     // 总服务器数量
	RunningServers   int `json:"runningServers"`   // 运行中服务器数量
	StoppedServers   int `json:"stoppedServers"`   // 已停止服务器数量
	ErrorServers     int `json:"errorServers"`     // 错误服务器数量
	TotalClients     int `json:"totalClients"`     // 总客户端数量
	TotalConnections int `json:"totalConnections"` // 总连接数（服务数量）
}

// TunnelClientStats 隧道客户端统计信息
type TunnelClientStats struct {
	TotalClients        int `json:"totalClients"`        // 总客户端数量
	ConnectedClients    int `json:"connectedClients"`    // 已连接客户端数量
	DisconnectedClients int `json:"disconnectedClients"` // 断开连接客户端数量
	TotalServices       int `json:"totalServices"`       // 总服务数量
	ActiveServices      int `json:"activeServices"`      // 活跃服务数量
}
