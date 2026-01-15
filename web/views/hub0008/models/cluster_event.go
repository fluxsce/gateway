package models

import (
	"time"
)

// ClusterEvent 集群事件模型，对应数据库 HUB_CLUSTER_EVENT 表
type ClusterEvent struct {
	// 主键和租户信息
	EventId  string `json:"eventId" db:"eventId"`   // 事件ID
	TenantId string `json:"tenantId" db:"tenantId"` // 租户ID

	// 事件来源(发布者)
	SourceNodeId string `json:"sourceNodeId" db:"sourceNodeId"` // 发布节点ID(hostname:port)
	SourceNodeIp string `json:"sourceNodeIp" db:"sourceNodeIp"` // 发布节点IP

	// 事件信息
	EventType    string `json:"eventType" db:"eventType"`       // 事件类型
	EventAction  string `json:"eventAction" db:"eventAction"`   // 事件动作
	EventPayload string `json:"eventPayload" db:"eventPayload"` // 事件数据(JSON格式)

	// 事件时间
	EventTime  time.Time  `json:"eventTime" db:"eventTime"`   // 事件发生时间
	ExpireTime *time.Time `json:"expireTime" db:"expireTime"` // 事件过期时间

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
	Reserved1 string `json:"reserved1" db:"reserved1"` // 预留字段1
	Reserved2 string `json:"reserved2" db:"reserved2"` // 预留字段2
	Reserved3 string `json:"reserved3" db:"reserved3"` // 预留字段3
	Reserved4 string `json:"reserved4" db:"reserved4"` // 预留字段4
	Reserved5 string `json:"reserved5" db:"reserved5"` // 预留字段5
}

// ClusterEventQuery 集群事件查询条件
type ClusterEventQuery struct {
	EventType    string `json:"eventType" form:"eventType" query:"eventType"`          // 事件类型（精确查询）
	EventAction  string `json:"eventAction" form:"eventAction" query:"eventAction"`    // 事件动作（精确查询）
	SourceNodeId string `json:"sourceNodeId" form:"sourceNodeId" query:"sourceNodeId"` // 发布节点ID（精确查询）
	SourceNodeIp string `json:"sourceNodeIp" form:"sourceNodeIp" query:"sourceNodeIp"` // 发布节点IP（精确查询）
	ActiveFlag   string `json:"activeFlag" form:"activeFlag" query:"activeFlag"`       // 活动状态标记：Y/N，空表示全部
	StartTime    string `json:"startTime" form:"startTime" query:"startTime"`          // 事件开始时间
	EndTime      string `json:"endTime" form:"endTime" query:"endTime"`                // 事件结束时间
}

// ClusterEventAck 集群事件确认模型，对应数据库 HUB_CLUSTER_EVENT_ACK 表
type ClusterEventAck struct {
	// 主键和租户信息
	AckId    string `json:"ackId" db:"ackId"`       // 确认ID
	TenantId string `json:"tenantId" db:"tenantId"` // 租户ID
	EventId  string `json:"eventId" db:"eventId"`   // 事件ID

	// 处理节点
	NodeId string `json:"nodeId" db:"nodeId"` // 处理节点ID(hostname:port)
	NodeIp string `json:"nodeIp" db:"nodeIp"` // 处理节点IP

	// 处理状态
	AckStatus     string     `json:"ackStatus" db:"ackStatus"`         // 确认状态(PENDING/SUCCESS/FAILED/SKIPPED)
	ProcessTime   *time.Time `json:"processTime" db:"processTime"`     // 处理时间
	ResultMessage string     `json:"resultMessage" db:"resultMessage"` // 结果信息
	RetryCount    int        `json:"retryCount" db:"retryCount"`       // 重试次数

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
	Reserved1 string `json:"reserved1" db:"reserved1"` // 预留字段1
	Reserved2 string `json:"reserved2" db:"reserved2"` // 预留字段2
	Reserved3 string `json:"reserved3" db:"reserved3"` // 预留字段3
}

// ClusterEventAckQuery 集群事件确认查询条件
type ClusterEventAckQuery struct {
	EventId    string `json:"eventId" form:"eventId" query:"eventId"`          // 事件ID（精确查询）
	NodeId     string `json:"nodeId" form:"nodeId" query:"nodeId"`             // 处理节点ID（精确查询）
	NodeIp     string `json:"nodeIp" form:"nodeIp" query:"nodeIp"`             // 处理节点IP（精确查询）
	AckStatus  string `json:"ackStatus" form:"ackStatus" query:"ackStatus"`    // 确认状态（精确查询）
	ActiveFlag string `json:"activeFlag" form:"activeFlag" query:"activeFlag"` // 活动状态标记：Y/N，空表示全部
	StartTime  string `json:"startTime" form:"startTime" query:"startTime"`    // 处理开始时间
	EndTime    string `json:"endTime" form:"endTime" query:"endTime"`          // 处理结束时间
}

// TableName 返回表名
func (ClusterEvent) TableName() string {
	return "HUB_CLUSTER_EVENT"
}

// TableName 返回表名
func (ClusterEventAck) TableName() string {
	return "HUB_CLUSTER_EVENT_ACK"
}
