package models

import "time"

// TunnelService 隧道服务模型
// 对应数据库表：HUB_TUNNEL_SERVICE
// 用途：管理客户端动态注册的服务配置
type TunnelService struct {
	// 主键信息
	TunnelServiceId string `json:"tunnelServiceId" form:"tunnelServiceId" query:"tunnelServiceId" db:"tunnelServiceId"` // 服务ID，主键
	TenantId        string `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                             // 租户ID
	TunnelClientId  string `json:"tunnelClientId" form:"tunnelClientId" query:"tunnelClientId" db:"tunnelClientId"`     // 客户端ID，关联HUB_TUNNEL_CLIENT
	UserId          string `json:"userId" form:"userId" query:"userId" db:"userId"`                                     // 用户ID

	// 基本信息
	ServiceName        string `json:"serviceName" form:"serviceName" query:"serviceName" db:"serviceName"`                             // 服务名称
	ServiceDescription string `json:"serviceDescription" form:"serviceDescription" query:"serviceDescription" db:"serviceDescription"` // 服务描述
	ServiceType        string `json:"serviceType" form:"serviceType" query:"serviceType" db:"serviceType"`                             // 服务类型(tcp,udp,http,https,stcp,sudp,xtcp)

	// 本地配置
	LocalAddress string `json:"localAddress" form:"localAddress" query:"localAddress" db:"localAddress"` // 本地地址
	LocalPort    int    `json:"localPort" form:"localPort" query:"localPort" db:"localPort"`             // 本地端口

	// 远程配置
	RemotePort    *int   `json:"remotePort" form:"remotePort" query:"remotePort" db:"remotePort"`             // 远程端口（TCP/UDP类型使用）
	CustomDomains string `json:"customDomains" form:"customDomains" query:"customDomains" db:"customDomains"` // 自定义域名列表，JSON格式
	SubDomain     string `json:"subDomain" form:"subDomain" query:"subDomain" db:"subDomain"`                 // 子域名（HTTP/HTTPS类型使用）

	// 高级配置
	UseEncryption  string `json:"useEncryption" form:"useEncryption" query:"useEncryption" db:"useEncryption"`     // 启用加密(N禁用,Y启用)
	UseCompression string `json:"useCompression" form:"useCompression" query:"useCompression" db:"useCompression"` // 启用压缩(N禁用,Y启用)
	BandwidthLimit string `json:"bandwidthLimit" form:"bandwidthLimit" query:"bandwidthLimit" db:"bandwidthLimit"` // 带宽限制（如：10MB/s）
	MaxConnections int    `json:"maxConnections" form:"maxConnections" query:"maxConnections" db:"maxConnections"` // 最大连接数

	// 状态信息
	ServiceStatus  string     `json:"serviceStatus" form:"serviceStatus" query:"serviceStatus" db:"serviceStatus"`     // 服务状态(active,inactive,error,offline)
	RegisteredTime time.Time  `json:"registeredTime" form:"registeredTime" query:"registeredTime" db:"registeredTime"` // 注册时间
	LastActiveTime *time.Time `json:"lastActiveTime" form:"lastActiveTime" query:"lastActiveTime" db:"lastActiveTime"` // 最后活跃时间

	// 统计信息
	ConnectionCount  int   `json:"connectionCount" form:"connectionCount" query:"connectionCount" db:"connectionCount"`     // 当前连接数
	TotalConnections int64 `json:"totalConnections" form:"totalConnections" query:"totalConnections" db:"totalConnections"` // 总连接数
	TotalTraffic     int64 `json:"totalTraffic" form:"totalTraffic" query:"totalTraffic" db:"totalTraffic"`                 // 总流量（字节）

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

// TunnelServiceQueryRequest 服务查询请求
type TunnelServiceQueryRequest struct {
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

// ServiceTrafficRequest 服务流量查询请求
type ServiceTrafficRequest struct {
	TunnelServiceId string `json:"tunnelServiceId" form:"tunnelServiceId" binding:"required"` // 服务ID
	TimeRange       string `json:"timeRange" form:"timeRange"`                                // 时间范围（如：24h, 7d, 30d）
}

// ServiceTrafficResponse 服务流量响应
type ServiceTrafficResponse struct {
	TunnelServiceId   string  `json:"tunnelServiceId"`   // 服务ID
	ServiceName       string  `json:"serviceName"`       // 服务名称
	TotalConnections  int64   `json:"totalConnections"`  // 总连接数
	ActiveConnections int     `json:"activeConnections"` // 活跃连接数
	TotalTraffic      int64   `json:"totalTraffic"`      // 总流量（字节）
	AvgResponseTime   float64 `json:"avgResponseTime"`   // 平均响应时间（毫秒）
	TrafficByHour     []int64 `json:"trafficByHour"`     // 按小时统计的流量
}

// AllocatePortRequest 分配远程端口请求
type AllocatePortRequest struct {
	TunnelServiceId string `json:"tunnelServiceId" form:"tunnelServiceId" binding:"required"` // 服务ID
	PreferredPort   int    `json:"preferredPort" form:"preferredPort"`                        // 首选端口（可选）
}

// AllocatePortResponse 分配远程端口响应
type AllocatePortResponse struct {
	TunnelServiceId string `json:"tunnelServiceId"` // 服务ID
	RemotePort      int    `json:"remotePort"`      // 分配的远程端口
}

// ReleasePortRequest 释放远程端口请求
type ReleasePortRequest struct {
	TunnelServiceId string `json:"tunnelServiceId" form:"tunnelServiceId" binding:"required"` // 服务ID
}

// ServicesByClientRequest 按客户端查询服务请求
type ServicesByClientRequest struct {
	TunnelClientId string `json:"tunnelClientId" form:"tunnelClientId" binding:"required"` // 客户端ID
	ServiceStatus  string `json:"serviceStatus" form:"serviceStatus"`                      // 服务状态过滤（可选）
}

// BatchOperationRequest 批量操作请求
type BatchOperationRequest struct {
	ServiceIds []string `json:"serviceIds" form:"serviceIds" binding:"required"` // 服务ID列表
}

// BatchOperationResponse 批量操作响应
type BatchOperationResponse struct {
	SuccessCount int      `json:"successCount"` // 成功数量
	FailedCount  int      `json:"failedCount"`  // 失败数量
	FailedIds    []string `json:"failedIds"`    // 失败的服务ID列表
}
