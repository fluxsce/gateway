package types

import "time"

// ServiceNode 服务节点实体
// 对应数据库表：HUB_SERVICE_NODE
type ServiceNode struct {
	// 主键和租户信息
	NodeId   string `json:"nodeId" db:"nodeId" form:"nodeId" query:"nodeId"`         // 节点ID，主键，最大长度32
	TenantId string `json:"tenantId" db:"tenantId" form:"tenantId" query:"tenantId"` // 租户ID，用于多租户数据隔离，最大长度32

	// 关联服务（通过联合主键关联HUB_SERVICE表）
	NamespaceId string `json:"namespaceId" db:"namespaceId" form:"namespaceId" query:"namespaceId"` // 命名空间ID，关联HUB_SERVICE表，最大长度32
	GroupName   string `json:"groupName" db:"groupName" form:"groupName" query:"groupName"`         // 分组名称，关联HUB_SERVICE表，最大长度64
	ServiceName string `json:"serviceName" db:"serviceName" form:"serviceName" query:"serviceName"` // 服务名称，关联HUB_SERVICE表，最大长度100

	// 网络连接信息
	IpAddress  string `json:"ipAddress" db:"ipAddress" form:"ipAddress" query:"ipAddress"`     // IP地址
	PortNumber int    `json:"portNumber" db:"portNumber" form:"portNumber" query:"portNumber"` // 端口号

	// 节点状态信息
	InstanceStatus string `json:"instanceStatus" db:"instanceStatus" form:"instanceStatus" query:"instanceStatus"` // 节点状态(UP:运行中,DOWN:已停止,STARTING:启动中,OUT_OF_SERVICE:暂停服务)
	HealthyStatus  string `json:"healthyStatus" db:"healthyStatus" form:"healthyStatus" query:"healthyStatus"`     // 健康状态(HEALTHY:健康,UNHEALTHY:不健康,UNKNOWN:未知)
	Ephemeral      string `json:"ephemeral" db:"ephemeral" form:"ephemeral"`                                       // 是否临时节点(Y:临时节点,N:持久节点)

	// 负载均衡配置
	Weight float64 `json:"weight" db:"weight" form:"weight"` // 权重值（DECIMAL(6,2)），范围0.01-10000.00，用于负载均衡

	// 节点元数据
	MetadataJson string `json:"metadataJson" db:"metadataJson" form:"metadataJson"` // 节点元数据，JSON格式，存储节点的扩展信息

	// 时间戳信息（对应数据库 DATETIME/DATE 类型）
	RegisterTime  time.Time  `json:"registerTime" db:"registerTime"`   // 注册时间（DATETIME/DATE NOT NULL）
	LastBeatTime  *time.Time `json:"lastBeatTime" db:"lastBeatTime"`   // 最后心跳时间（DATETIME/DATE DEFAULT NULL）
	LastCheckTime *time.Time `json:"lastCheckTime" db:"lastCheckTime"` // 最后健康检查时间（DATETIME/DATE DEFAULT NULL）

	// 通用字段（对应数据库 DATETIME/DATE 类型）
	AddTime        time.Time `json:"addTime" db:"addTime"`                                            // 创建时间（DATETIME/DATE NOT NULL）
	AddWho         string    `json:"addWho" db:"addWho" form:"addWho"`                                // 创建人ID
	EditTime       time.Time `json:"editTime" db:"editTime"`                                          // 最后修改时间（DATETIME/DATE NOT NULL）
	EditWho        string    `json:"editWho" db:"editWho" form:"editWho"`                             // 最后修改人ID
	OprSeqFlag     string    `json:"oprSeqFlag" db:"oprSeqFlag"`                                      // 操作序列标识
	CurrentVersion int       `json:"currentVersion" db:"currentVersion"`                              // 当前版本号
	ActiveFlag     string    `json:"activeFlag" db:"activeFlag" form:"activeFlag" query:"activeFlag"` // 活动状态标记(N非活动,Y活动)
	NoteText       string    `json:"noteText" db:"noteText" form:"noteText"`                          // 备注信息
	ExtProperty    string    `json:"extProperty" db:"extProperty" form:"extProperty"`                 // 扩展属性，JSON格式
}

// NodeStatus 节点状态常量
const (
	NodeStatusUp           = "UP"             // 运行中
	NodeStatusDown         = "DOWN"           // 已停止
	NodeStatusStarting     = "STARTING"       // 启动中
	NodeStatusOutOfService = "OUT_OF_SERVICE" // 暂停服务
)

// HealthyStatus 健康状态常量
const (
	HealthyStatusHealthy   = "HEALTHY"   // 健康
	HealthyStatusUnhealthy = "UNHEALTHY" // 不健康
	HealthyStatusUnknown   = "UNKNOWN"   // 未知
)
