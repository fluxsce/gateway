// Package models 定义 hub0061 模块的数据模型
package models

// ============================================================
// 静态服务器相关模型
// ============================================================

// StaticServerQueryRequest 静态服务器查询请求
type StaticServerQueryRequest struct {
	PageIndex         int    `json:"pageIndex" form:"pageIndex" query:"pageIndex"`
	PageSize          int    `json:"pageSize" form:"pageSize" query:"pageSize"`
	ServerName        string `json:"serverName" form:"serverName" query:"serverName"`                      // 服务器名称（模糊匹配）
	ServerDescription string `json:"serverDescription" form:"serverDescription" query:"serverDescription"` // 服务器描述（模糊匹配）
	ListenAddress     string `json:"listenAddress" form:"listenAddress" query:"listenAddress"`             // 监听地址
	ListenPort        int    `json:"listenPort" form:"listenPort" query:"listenPort"`                      // 监听端口
	ServerStatus      string `json:"serverStatus" form:"serverStatus" query:"serverStatus"`                // 服务器状态过滤
	ServerType        string `json:"serverType" form:"serverType" query:"serverType"`                      // 服务器类型过滤
	ActiveFlag        string `json:"activeFlag" form:"activeFlag" query:"activeFlag"`                      // 活动标记过滤
}

// StaticServerStats 静态服务器统计信息
type StaticServerStats struct {
	TotalServers       int   `json:"totalServers"`       // 总服务器数
	RunningServers     int   `json:"runningServers"`     // 运行中服务器数
	StoppedServers     int   `json:"stoppedServers"`     // 已停止服务器数
	TotalConnections   int64 `json:"totalConnections"`   // 总连接数
	TotalBytesReceived int64 `json:"totalBytesReceived"` // 总接收字节数
	TotalBytesSent     int64 `json:"totalBytesSent"`     // 总发送字节数
}

// ============================================================
// 静态节点相关模型
// ============================================================

// StaticNodeQueryRequest 静态节点查询请求
type StaticNodeQueryRequest struct {
	PageIndex            int    `json:"pageIndex" form:"pageIndex" query:"pageIndex"`
	PageSize             int    `json:"pageSize" form:"pageSize" query:"pageSize"`
	TunnelStaticServerId string `json:"tunnelStaticServerId" form:"tunnelStaticServerId" query:"tunnelStaticServerId"` // 所属服务器ID
	NodeName             string `json:"nodeName" form:"nodeName" query:"nodeName"`                                     // 节点名称（模糊匹配）
	NodeDescription      string `json:"nodeDescription" form:"nodeDescription" query:"nodeDescription"`                // 节点描述（模糊匹配）
	TargetAddress        string `json:"targetAddress" form:"targetAddress" query:"targetAddress"`                      // 目标地址（模糊匹配）
	TargetPort           int    `json:"targetPort" form:"targetPort" query:"targetPort"`                               // 目标端口
	NodeStatus           string `json:"nodeStatus" form:"nodeStatus" query:"nodeStatus"`                               // 节点状态过滤
	ProxyType            string `json:"proxyType" form:"proxyType" query:"proxyType"`                                  // 代理类型过滤
	HealthCheckStatus    string `json:"healthCheckStatus" form:"healthCheckStatus" query:"healthCheckStatus"`          // 健康检查状态过滤
	ActiveFlag           string `json:"activeFlag" form:"activeFlag" query:"activeFlag"`                               // 活动标记过滤
}

// StaticNodeStats 静态节点统计信息
type StaticNodeStats struct {
	TotalNodes         int   `json:"totalNodes"`         // 总节点数
	ActiveNodes        int   `json:"activeNodes"`        // 活跃节点数
	InactiveNodes      int   `json:"inactiveNodes"`      // 非活跃节点数
	HealthyNodes       int   `json:"healthyNodes"`       // 健康节点数
	UnhealthyNodes     int   `json:"unhealthyNodes"`     // 不健康节点数
	TotalConnections   int64 `json:"totalConnections"`   // 总连接数
	TotalBytesReceived int64 `json:"totalBytesReceived"` // 总接收字节数
	TotalBytesSent     int64 `json:"totalBytesSent"`     // 总发送字节数
}
