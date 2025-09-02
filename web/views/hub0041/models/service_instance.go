package models

import "time"

// ServiceInstance 服务实例信息模型
// 对应数据库表：HUB_REGISTRY_SERVICE_INSTANCE
// 用途：管理第三方应用注册的服务实例信息
type ServiceInstance struct {
	// 主键信息
	ServiceInstanceId string `json:"serviceInstanceId" form:"serviceInstanceId" query:"serviceInstanceId" db:"serviceInstanceId"` // 服务实例ID，主键
	TenantId          string `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                                     // 租户ID，用于多租户数据隔离

	// 关联服务和分组信息
	ServiceGroupId string `json:"serviceGroupId" form:"serviceGroupId" query:"serviceGroupId" db:"serviceGroupId"` // 服务分组ID，关联HUB_REGISTRY_SERVICE_GROUP表主键
	ServiceName    string `json:"serviceName" form:"serviceName" query:"serviceName" db:"serviceName"`             // 服务名称，冗余字段便于查询
	GroupName      string `json:"groupName" form:"groupName" query:"groupName" db:"groupName"`                     // 分组名称，冗余字段便于查询

	// 网络连接信息
	HostAddress string `json:"hostAddress" form:"hostAddress" query:"hostAddress" db:"hostAddress"` // 主机地址
	PortNumber  int    `json:"portNumber" form:"portNumber" query:"portNumber" db:"portNumber"`     // 端口号
	ContextPath string `json:"contextPath" form:"contextPath" query:"contextPath" db:"contextPath"` // 上下文路径

	// 实例状态信息
	InstanceStatus string `json:"instanceStatus" form:"instanceStatus" query:"instanceStatus" db:"instanceStatus"` // 实例状态(UP,DOWN,STARTING,OUT_OF_SERVICE)
	HealthStatus   string `json:"healthStatus" form:"healthStatus" query:"healthStatus" db:"healthStatus"`         // 健康状态(HEALTHY,UNHEALTHY,UNKNOWN)

	// 负载均衡配置
	WeightValue int `json:"weightValue" form:"weightValue" query:"weightValue" db:"weightValue"` // 权重值

	// 客户端信息
	ClientId         string `json:"clientId,omitempty" form:"clientId" query:"clientId" db:"clientId"`                       // 客户端ID
	ClientVersion    string `json:"clientVersion,omitempty" form:"clientVersion" query:"clientVersion" db:"clientVersion"`   // 客户端版本
	ClientType       string `json:"clientType" form:"clientType" query:"clientType" db:"clientType"`                         // 客户端类型(SERVICE,GATEWAY,ADMIN)
	TempInstanceFlag string `json:"tempInstanceFlag" form:"tempInstanceFlag" query:"tempInstanceFlag" db:"tempInstanceFlag"` // 临时实例标记(Y是临时实例,N否)

	// 健康检查统计
	HeartbeatFailCount int `json:"heartbeatFailCount" form:"heartbeatFailCount" query:"heartbeatFailCount" db:"heartbeatFailCount"` // 心跳检查失败次数，仅用于计数

	// 元数据和标签
	MetadataJson string `json:"metadataJson,omitempty" form:"metadataJson" query:"metadataJson" db:"metadataJson"` // 实例元数据，JSON格式
	TagsJson     string `json:"tagsJson,omitempty" form:"tagsJson" query:"tagsJson" db:"tagsJson"`                 // 实例标签，JSON格式

	// 时间戳信息
	RegisterTime        time.Time  `json:"registerTime" form:"registerTime" query:"registerTime" db:"registerTime"`                                       // 注册时间
	LastHeartbeatTime   *time.Time `json:"lastHeartbeatTime,omitempty" form:"lastHeartbeatTime" query:"lastHeartbeatTime" db:"lastHeartbeatTime"`         // 最后心跳时间
	LastHealthCheckTime *time.Time `json:"lastHealthCheckTime,omitempty" form:"lastHealthCheckTime" query:"lastHealthCheckTime" db:"lastHealthCheckTime"` // 最后健康检查时间

	// 通用审计字段
	AddTime        time.Time `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                             // 创建时间
	AddWho         string    `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                 // 创建人ID
	EditTime       time.Time `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                         // 最后修改时间
	EditWho        string    `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                             // 最后修改人ID
	OprSeqFlag     string    `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                 // 操作序列标识
	CurrentVersion int       `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"` // 当前版本号
	ActiveFlag     string    `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`                 // 活动状态标记(N非活动,Y活动)
	NoteText       string    `json:"noteText,omitempty" form:"noteText" query:"noteText" db:"noteText"`               // 备注信息
	ExtProperty    string    `json:"extProperty,omitempty" form:"extProperty" query:"extProperty" db:"extProperty"`   // 扩展属性，JSON格式

	// 预留字段
	Reserved1  string `json:"reserved1,omitempty" form:"reserved1" query:"reserved1" db:"reserved1"`     // 预留字段1
	Reserved2  string `json:"reserved2,omitempty" form:"reserved2" query:"reserved2" db:"reserved2"`     // 预留字段2
	Reserved3  string `json:"reserved3,omitempty" form:"reserved3" query:"reserved3" db:"reserved3"`     // 预留字段3
	Reserved4  string `json:"reserved4,omitempty" form:"reserved4" query:"reserved4" db:"reserved4"`     // 预留字段4
	Reserved5  string `json:"reserved5,omitempty" form:"reserved5" query:"reserved5" db:"reserved5"`     // 预留字段5
	Reserved6  string `json:"reserved6,omitempty" form:"reserved6" query:"reserved6" db:"reserved6"`     // 预留字段6
	Reserved7  string `json:"reserved7,omitempty" form:"reserved7" query:"reserved7" db:"reserved7"`     // 预留字段7
	Reserved8  string `json:"reserved8,omitempty" form:"reserved8" query:"reserved8" db:"reserved8"`     // 预留字段8
	Reserved9  string `json:"reserved9,omitempty" form:"reserved9" query:"reserved9" db:"reserved9"`     // 预留字段9
	Reserved10 string `json:"reserved10,omitempty" form:"reserved10" query:"reserved10" db:"reserved10"` // 预留字段10
}

// ServiceEvent 服务事件信息
// 对应数据库表：HUB_REGISTRY_SERVICE_EVENT
type ServiceEvent struct {
	// 主键信息
	ServiceEventId string `json:"serviceEventId" db:"serviceEventId"` // 服务事件ID，主键
	TenantId       string `json:"tenantId" db:"tenantId"`             // 租户ID

	// 关联信息
	ServiceGroupId    string `json:"serviceGroupId" db:"serviceGroupId"`       // 服务分组ID
	ServiceInstanceId string `json:"serviceInstanceId" db:"serviceInstanceId"` // 服务实例ID

	// 事件基本信息（冗余字段）
	GroupName     string `json:"groupName" db:"groupName"`         // 分组名称
	ServiceName   string `json:"serviceName" db:"serviceName"`     // 服务名称
	HostAddress   string `json:"hostAddress" db:"hostAddress"`     // 主机地址
	PortNumber    int    `json:"portNumber" db:"portNumber"`       // 端口号
	NodeIpAddress string `json:"nodeIpAddress" db:"nodeIpAddress"` // 事件产生节点的IP地址

	EventType   string `json:"eventType" db:"eventType"`     // 事件类型
	EventSource string `json:"eventSource" db:"eventSource"` // 事件来源

	// 事件数据
	EventDataJson string `json:"eventDataJson" db:"eventDataJson"` // 事件数据，JSON格式
	EventMessage  string `json:"eventMessage" db:"eventMessage"`   // 事件消息描述

	// 时间信息
	EventTime time.Time `json:"eventTime" db:"eventTime"` // 事件发生时间

	// 通用字段
	AddTime        time.Time `json:"addTime" db:"addTime"`               // 创建时间
	AddWho         string    `json:"addWho" db:"addWho"`                 // 创建人ID
	EditTime       time.Time `json:"editTime" db:"editTime"`             // 最后修改时间
	EditWho        string    `json:"editWho" db:"editWho"`               // 最后修改人ID
	OprSeqFlag     string    `json:"oprSeqFlag" db:"oprSeqFlag"`         // 操作序列标识
	CurrentVersion int       `json:"currentVersion" db:"currentVersion"` // 当前版本号
	ActiveFlag     string    `json:"activeFlag" db:"activeFlag"`         // 活动状态标记
	NoteText       string    `json:"noteText" db:"noteText"`             // 备注信息
	ExtProperty    string    `json:"extProperty" db:"extProperty"`       // 扩展属性

	// 预留字段
	Reserved1  string `json:"reserved1" db:"reserved1"`   // 预留字段1
	Reserved2  string `json:"reserved2" db:"reserved2"`   // 预留字段2
	Reserved3  string `json:"reserved3" db:"reserved3"`   // 预留字段3
	Reserved4  string `json:"reserved4" db:"reserved4"`   // 预留字段4
	Reserved5  string `json:"reserved5" db:"reserved5"`   // 预留字段5
	Reserved6  string `json:"reserved6" db:"reserved6"`   // 预留字段6
	Reserved7  string `json:"reserved7" db:"reserved7"`   // 预留字段7
	Reserved8  string `json:"reserved8" db:"reserved8"`   // 预留字段8
	Reserved9  string `json:"reserved9" db:"reserved9"`   // 预留字段9
	Reserved10 string `json:"reserved10" db:"reserved10"` // 预留字段10
}
