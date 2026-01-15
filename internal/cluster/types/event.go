package types

import (
	"encoding/json"
	"time"
)

// ClusterEvent 集群事件
// 对应数据库表：HUB_CLUSTER_EVENT
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

// ClusterEventAck 集群事件确认
// 对应数据库表：HUB_CLUSTER_EVENT_ACK
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

// SetPayload 设置事件数据(自动序列化为JSON)
func (e *ClusterEvent) SetPayload(data interface{}) error {
	if data == nil {
		e.EventPayload = ""
		return nil
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	e.EventPayload = string(bytes)
	return nil
}

// GetPayload 获取事件数据(自动反序列化)
func (e *ClusterEvent) GetPayload(result interface{}) error {
	if e.EventPayload == "" {
		return nil
	}
	return json.Unmarshal([]byte(e.EventPayload), result)
}

// IsExpired 检查事件是否已过期
func (e *ClusterEvent) IsExpired() bool {
	if e.ExpireTime == nil {
		return false
	}
	return time.Now().After(*e.ExpireTime)
}
