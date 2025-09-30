package models

import "time"

// TunnelService 隧道服务模型
// 对应数据库表：TUNNEL_SERVICE
// 用途：管理隧道服务信息
type TunnelService struct {
	// 主键信息
	TunnelServiceId string `json:"tunnelServiceId" form:"tunnelServiceId" query:"tunnelServiceId" db:"tunnelServiceId"` // 隧道服务ID，主键
	ServiceName     string `json:"serviceName" form:"serviceName" query:"serviceName" db:"serviceName"`                 // 服务名称
	ServiceType     string `json:"serviceType" form:"serviceType" query:"serviceType" db:"serviceType"`                 // 服务类型(TCP,UDP,HTTP,HTTPS,STCP,SUDP,XTCP)

	// 关联信息
	TunnelClientId string `json:"tunnelClientId" form:"tunnelClientId" query:"tunnelClientId" db:"tunnelClientId"` // 关联的隧道客户端ID
	ClientName     string `json:"clientName" form:"clientName" query:"clientName" db:"clientName"`                 // 客户端名称（冗余字段）
	TunnelServerId string `json:"tunnelServerId" form:"tunnelServerId" query:"tunnelServerId" db:"tunnelServerId"` // 关联的隧道服务器ID

	// 本地配置
	LocalAddress string `json:"localAddress" form:"localAddress" query:"localAddress" db:"localAddress"` // 本地地址
	LocalPort    int    `json:"localPort" form:"localPort" query:"localPort" db:"localPort"`             // 本地端口

	// 远程配置
	RemotePort    *int    `json:"remotePort" form:"remotePort" query:"remotePort" db:"remotePort"`             // 远程端口
	CustomDomains string  `json:"customDomains" form:"customDomains" query:"customDomains" db:"customDomains"` // 自定义域名(JSON数组)
	SubDomain     *string `json:"subDomain" form:"subDomain" query:"subDomain" db:"subDomain"`                 // 子域名

	// HTTP配置
	HttpUser     *string `json:"httpUser" form:"httpUser" query:"httpUser" db:"httpUser"`                 // HTTP用户名
	HttpPassword *string `json:"httpPassword" form:"httpPassword" query:"httpPassword" db:"httpPassword"` // HTTP密码

	// 安全配置
	UseEncryption  bool    `json:"useEncryption" form:"useEncryption" query:"useEncryption" db:"useEncryption"`     // 是否使用加密
	UseCompression bool    `json:"useCompression" form:"useCompression" query:"useCompression" db:"useCompression"` // 是否使用压缩
	SecretKey      *string `json:"secretKey" form:"secretKey" query:"secretKey" db:"secretKey"`                     // 密钥

	// 限制配置
	BandwidthLimit *string `json:"bandwidthLimit" form:"bandwidthLimit" query:"bandwidthLimit" db:"bandwidthLimit"` // 带宽限制
	MaxConnections *int    `json:"maxConnections" form:"maxConnections" query:"maxConnections" db:"maxConnections"` // 最大连接数

	// 状态信息
	ServiceStatus    string `json:"serviceStatus" form:"serviceStatus" query:"serviceStatus" db:"serviceStatus"`             // 服务状态(ACTIVE,INACTIVE,ERROR)
	ConnectionCount  int    `json:"connectionCount" form:"connectionCount" query:"connectionCount" db:"connectionCount"`     // 当前连接数
	TotalConnections int64  `json:"totalConnections" form:"totalConnections" query:"totalConnections" db:"totalConnections"` // 总连接数
	TotalTraffic     int64  `json:"totalTraffic" form:"totalTraffic" query:"totalTraffic" db:"totalTraffic"`                 // 总流量(字节)

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

// TunnelServiceStats 隧道服务统计信息
type TunnelServiceStats struct {
	TotalServices    int   `json:"totalServices"`    // 总服务数量
	ActiveServices   int   `json:"activeServices"`   // 活跃服务数量
	InactiveServices int   `json:"inactiveServices"` // 非活跃服务数量
	ErrorServices    int   `json:"errorServices"`    // 错误服务数量
	TotalConnections int64 `json:"totalConnections"` // 总连接数
	TotalTraffic     int64 `json:"totalTraffic"`     // 总流量
}
