package model

import "time"

// 事件类型常量。字符串值是订阅协议的一部分，新增只能追加。
const (
	EventServiceAdded     = "SERVICE_ADDED"
	EventServiceUpdated   = "SERVICE_UPDATED"
	EventServiceDeleted   = "SERVICE_DELETED"
	EventNodeRegistered   = "NODE_REGISTERED"
	EventNodeUpdated      = "NODE_UPDATED"
	EventNodeDeregistered = "NODE_DEREGISTERED"
	EventNodeEvicted      = "NODE_EVICTED"
	EventNodeOffline      = "NODE_OFFLINE"
	EventConfigPublished  = "CONFIG_PUBLISHED"
	EventConfigRolledBack = "CONFIG_ROLLED_BACK"
	EventConfigDeleted    = "CONFIG_DELETED"
	EventNamespaceAdded   = "NAMESPACE_ADDED"
	EventNamespaceUpdated = "NAMESPACE_UPDATED"
	EventNamespaceDeleted = "NAMESPACE_DELETED"
)

// NamingEvent 是命名变更通知。Service / ChangedNode 由实现给出快照拷贝。
type NamingEvent struct {
	Type               string
	Timestamp          time.Time
	CenterInstanceName string
	Environment        string // 中心实例环境，副本同步时定位运行时
	NamespaceID        string
	GroupName          string
	ServiceName        string
	Service            *Service
	ChangedNode        *Node // 这一条变更的业务节点；名单看 Service.Nodes
}

// NamespaceEvent 是管理面对命名空间的增删改，副本只改 L1，不写库。
type NamespaceEvent struct {
	Type               string
	Timestamp          time.Time
	CenterInstanceName string
	Environment        string
	NamespaceID        string
	Namespace          *Namespace
}

// ConfigEvent 是已发布配置变更通知。草稿变更不产生本事件。
type ConfigEvent struct {
	Type               string
	Timestamp          time.Time
	CenterInstanceName string
	Environment        string // 中心实例环境，副本同步时定位运行时
	NamespaceID        string
	GroupName          string
	DataID             string
	Release            *ConfigRelease
}

// 数据面断连原因。SDK 与网关 Stop 都表现为流结束，必须靠 reason 区分清理策略。
const (
	// DisconnectClientLost：SDK 应用断流（进程退出、主动关、网络断开）。临时节点立刻驱逐。
	DisconnectClientLost = "client_lost"
	// DisconnectServerDrain：本网关主动停监听（Stop / Reload / 优雅重启）。节点保留，只交还 Owner。
	DisconnectServerDrain = "server_shutdown"
)

// IsServerDrain 报告这次断连是不是网关主动停机，而不是 SDK 离开。
func IsServerDrain(reason string) bool {
	return reason == DisconnectServerDrain
}

// ServiceSubscriber 是本网关进程上的一条命名订阅关系。
type ServiceSubscriber struct {
	ConnectionID string
	ClientID     string
	ClientIP     string
	NamespaceID  string
	GroupName    string
	ServiceName  string // 对端服务：入站为订阅方服务，出站为被订服务；命名空间/分组级可为空
	Scope        string // service / group / namespace
	LastActive   time.Time
}

// ConnectionInfo 是一条数据面会话的只读视图。
type ConnectionInfo struct {
	ConnectionID string
	ClientID     string
	ClientIP     string
	NamespaceID  string // 握手默认命名空间；空表示尚未绑定
	LastActive   time.Time
}

// Overview 是单个中心实例的运行时计数快照。
type Overview struct {
	CenterInstanceName string
	ServiceCount       int
	NodeCount          int
	HealthyNodes       int
	ConfigCount        int
	ConnectionCount    int
	OwnerGatewayID     string   // 本网关进程 ID，不是业务 NodeID；单进程可为空
	ReplicaGatewayIDs  []string // 已知网关进程（本机 + 近期事件/Owner）
}
