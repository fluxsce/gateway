package models

import "time"

// ServiceNodeModel 服务节点模型
type ServiceNodeModel struct {
	TenantId            string     `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                                         // 租户ID
	ServiceNodeId       string     `json:"serviceNodeId" form:"serviceNodeId" query:"serviceNodeId" db:"serviceNodeId"`                     // 服务节点ID
	ServiceDefinitionId string     `json:"serviceDefinitionId" form:"serviceDefinitionId" query:"serviceDefinitionId" db:"serviceDefinitionId"` // 服务定义ID
	NodeId              string     `json:"nodeId" form:"nodeId" query:"nodeId" db:"nodeId"`                                                 // 节点标识ID
	NodeUrl             string     `json:"nodeUrl" form:"nodeUrl" query:"nodeUrl" db:"nodeUrl"`                                             // 节点完整URL
	NodeHost            string     `json:"nodeHost" form:"nodeHost" query:"nodeHost" db:"nodeHost"`                                         // 节点主机地址
	NodePort            int        `json:"nodePort" form:"nodePort" query:"nodePort" db:"nodePort"`                                         // 节点端口
	NodeProtocol        string     `json:"nodeProtocol" form:"nodeProtocol" query:"nodeProtocol" db:"nodeProtocol"`                         // 节点协议(HTTP,HTTPS)
	NodeWeight          int        `json:"nodeWeight" form:"nodeWeight" query:"nodeWeight" db:"nodeWeight"`                                 // 节点权重
	HealthStatus        string     `json:"healthStatus" form:"healthStatus" query:"healthStatus" db:"healthStatus"`                         // 健康状态(N不健康,Y健康)
	NodeMetadata        string     `json:"nodeMetadata" form:"nodeMetadata" query:"nodeMetadata" db:"nodeMetadata"`                         // 节点元数据,JSON格式
	NodeStatus          int        `json:"nodeStatus" form:"nodeStatus" query:"nodeStatus" db:"nodeStatus"`                                 // 节点运行状态(0下线,1在线,2维护)
	LastHealthCheckTime *time.Time `json:"lastHealthCheckTime" form:"lastHealthCheckTime" query:"lastHealthCheckTime" db:"lastHealthCheckTime"` // 最后健康检查时间
	HealthCheckResult   string     `json:"healthCheckResult" form:"healthCheckResult" query:"healthCheckResult" db:"healthCheckResult"`     // 健康检查结果详情
	Reserved1           string     `json:"reserved1" form:"reserved1" query:"reserved1" db:"reserved1"`                                     // 预留字段1
	Reserved2           string     `json:"reserved2" form:"reserved2" query:"reserved2" db:"reserved2"`                                     // 预留字段2
	Reserved3           int        `json:"reserved3" form:"reserved3" query:"reserved3" db:"reserved3"`                                     // 预留字段3
	Reserved4           int        `json:"reserved4" form:"reserved4" query:"reserved4" db:"reserved4"`                                     // 预留字段4
	Reserved5           *time.Time `json:"reserved5" form:"reserved5" query:"reserved5" db:"reserved5"`                                     // 预留字段5
	ExtProperty         string     `json:"extProperty" form:"extProperty" query:"extProperty" db:"extProperty"`                             // 扩展属性,JSON格式
	AddTime             *time.Time `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                                             // 创建时间
	AddWho              string     `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                                 // 创建人ID
	EditTime            *time.Time `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                                         // 最后修改时间
	EditWho             string     `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                                             // 最后修改人ID
	OprSeqFlag          string     `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                                 // 操作序列标识
	CurrentVersion      int        `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"`                 // 当前版本号
	ActiveFlag          string     `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`                                 // 活动状态标记(N非活动,Y活动)
	NoteText            string     `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`                                         // 备注信息
} 