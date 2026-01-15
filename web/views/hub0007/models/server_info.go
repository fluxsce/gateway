package models

import (
	"time"
)

// ServerInfo 系统节点信息模型，对应数据库HUB_METRIC_SERVER_INFO表
type ServerInfo struct {
	// 主键和租户信息
	MetricServerId string `json:"metricServerId" form:"metricServerId" query:"metricServerId" db:"metricServerId"` // 服务器ID，联合主键
	TenantId       string `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                         // 租户ID，联合主键

	// 服务器基本信息
	Hostname       string    `json:"hostname" form:"hostname" query:"hostname" db:"hostname"`                         // 主机名
	OsType         string    `json:"osType" form:"osType" query:"osType" db:"osType"`                                 // 操作系统类型
	OsVersion      string    `json:"osVersion" form:"osVersion" query:"osVersion" db:"osVersion"`                     // 操作系统版本
	KernelVersion  string    `json:"kernelVersion" form:"kernelVersion" query:"kernelVersion" db:"kernelVersion"`     // 内核版本
	Architecture   string    `json:"architecture" form:"architecture" query:"architecture" db:"architecture"`         // 系统架构
	BootTime       time.Time `json:"bootTime" form:"bootTime" query:"bootTime" db:"bootTime"`                         // 系统启动时间
	IpAddress      string    `json:"ipAddress" form:"ipAddress" query:"ipAddress" db:"ipAddress"`                     // 主IP地址
	MacAddress     string    `json:"macAddress" form:"macAddress" query:"macAddress" db:"macAddress"`                 // 主MAC地址
	ServerLocation string    `json:"serverLocation" form:"serverLocation" query:"serverLocation" db:"serverLocation"` // 服务器位置
	ServerType     string    `json:"serverType" form:"serverType" query:"serverType" db:"serverType"`                 // 服务器类型(physical/virtual/unknown)
	LastUpdateTime time.Time `json:"lastUpdateTime" form:"lastUpdateTime" query:"lastUpdateTime" db:"lastUpdateTime"` // 最后更新时间

	// 扩展信息（JSON格式）
	NetworkInfo  string `json:"networkInfo" form:"networkInfo" query:"networkInfo" db:"networkInfo"`     // 网络信息详情，JSON格式存储所有IP和MAC地址
	SystemInfo   string `json:"systemInfo" form:"systemInfo" query:"systemInfo" db:"systemInfo"`         // 系统详细信息，JSON格式存储温度、负载等扩展信息
	HardwareInfo string `json:"hardwareInfo" form:"hardwareInfo" query:"hardwareInfo" db:"hardwareInfo"` // 硬件信息，JSON格式存储CPU、内存、磁盘等硬件详情

	// 通用字段
	AddTime        time.Time `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                             // 创建时间
	AddWho         string    `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                 // 创建人ID
	EditTime       time.Time `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                         // 最后修改时间
	EditWho        string    `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                             // 最后修改人ID
	OprSeqFlag     string    `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                 // 操作序列标识
	CurrentVersion int       `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"` // 当前版本号
	ActiveFlag     string    `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`                 // 活动状态标记(N非活动,Y活动)
	NoteText       string    `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`                         // 备注信息
	ExtProperty    string    `json:"extProperty" form:"extProperty" query:"extProperty" db:"extProperty"`             // 扩展属性，JSON格式

	// 预留字段
	Reserved1  string `json:"reserved1" form:"reserved1" query:"reserved1" db:"reserved1"`     // 预留字段1
	Reserved2  string `json:"reserved2" form:"reserved2" query:"reserved2" db:"reserved2"`     // 预留字段2
	Reserved3  string `json:"reserved3" form:"reserved3" query:"reserved3" db:"reserved3"`     // 预留字段3
	Reserved4  string `json:"reserved4" form:"reserved4" query:"reserved4" db:"reserved4"`     // 预留字段4
	Reserved5  string `json:"reserved5" form:"reserved5" query:"reserved5" db:"reserved5"`     // 预留字段5
	Reserved6  string `json:"reserved6" form:"reserved6" query:"reserved6" db:"reserved6"`     // 预留字段6
	Reserved7  string `json:"reserved7" form:"reserved7" query:"reserved7" db:"reserved7"`     // 预留字段7
	Reserved8  string `json:"reserved8" form:"reserved8" query:"reserved8" db:"reserved8"`     // 预留字段8
	Reserved9  string `json:"reserved9" form:"reserved9" query:"reserved9" db:"reserved9"`     // 预留字段9
	Reserved10 string `json:"reserved10" form:"reserved10" query:"reserved10" db:"reserved10"` // 预留字段10
}

// TableName 返回表名
func (ServerInfo) TableName() string {
	return "HUB_METRIC_SERVER_INFO"
}

// ServerType 服务器类型常量
const (
	ServerTypePhysical = "physical" // 物理服务器
	ServerTypeVirtual  = "virtual"  // 虚拟服务器
	ServerTypeUnknown  = "unknown"  // 未知类型
)

// ServerInfoQuery 系统节点信息查询条件，对应前端搜索表单的查询参数
type ServerInfoQuery struct {
	Hostname       string `json:"hostname" form:"hostname" query:"hostname"`                   // 主机名（模糊查询）
	OsType         string `json:"osType" form:"osType" query:"osType"`                         // 操作系统类型
	ServerType     string `json:"serverType" form:"serverType" query:"serverType"`             // 服务器类型：physical/virtual/unknown，空表示全部
	IpAddress      string `json:"ipAddress" form:"ipAddress" query:"ipAddress"`                // IP地址（模糊查询）
	ServerLocation string `json:"serverLocation" form:"serverLocation" query:"serverLocation"` // 服务器位置（模糊查询）
	ActiveFlag     string `json:"activeFlag" form:"activeFlag" query:"activeFlag"`             // 活动标记：Y-活动，N-非活动，空表示全部（默认查询活动状态）
}

// ToMap 将ServerInfo模型转换为响应map
func (s *ServerInfo) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"metricServerId": s.MetricServerId,
		"tenantId":       s.TenantId,
		"hostname":       s.Hostname,
		"osType":         s.OsType,
		"osVersion":      s.OsVersion,
		"kernelVersion":  s.KernelVersion,
		"architecture":   s.Architecture,
		"bootTime":       s.BootTime,
		"ipAddress":      s.IpAddress,
		"macAddress":     s.MacAddress,
		"serverLocation": s.ServerLocation,
		"serverType":     s.ServerType,
		"lastUpdateTime": s.LastUpdateTime,
		"networkInfo":    s.NetworkInfo,
		"systemInfo":     s.SystemInfo,
		"hardwareInfo":   s.HardwareInfo,
		"addTime":        s.AddTime,
		"addWho":         s.AddWho,
		"editTime":       s.EditTime,
		"editWho":        s.EditWho,
		"oprSeqFlag":     s.OprSeqFlag,
		"currentVersion": s.CurrentVersion,
		"activeFlag":     s.ActiveFlag,
		"noteText":       s.NoteText,
		"extProperty":    s.ExtProperty,
		"reserved1":      s.Reserved1,
		"reserved2":      s.Reserved2,
		"reserved3":      s.Reserved3,
		"reserved4":      s.Reserved4,
		"reserved5":      s.Reserved5,
		"reserved6":      s.Reserved6,
		"reserved7":      s.Reserved7,
		"reserved8":      s.Reserved8,
		"reserved9":      s.Reserved9,
		"reserved10":     s.Reserved10,
	}
}
