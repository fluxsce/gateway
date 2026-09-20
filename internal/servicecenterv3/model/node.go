package model

import "time"

// 业务节点状态。取值对齐 proto NodeStatus / InstanceStatus。
const (
	NodeUP           = "UP"
	NodeDown         = "DOWN"
	NodeStarting     = "STARTING"
	NodeOutOfService = "OUT_OF_SERVICE"
)

// 健康探测结果，与 Status 正交：Status 是人工/注册状态，HealthyStatus 是探测结果。
const (
	Healthy   = "HEALTHY"
	Unhealthy = "UNHEALTHY"
	Unknown   = "UNKNOWN"
)

// Node 是某个 Service 下的业务节点（ip:port），对应 proto Node、库表 HUB_SERVICE_NODE。
// 不要和下面两个 ID 混：
//   - CenterInstanceName：服务中心卡片（监听地址那一层）
//   - OwnerGatewayID：接到这条心跳的网关进程
type Node struct {
	NodeID             string            // 业务节点 ID = proto nodeId = 库表 nodeId；空则由实现生成
	TenantID           string            // 归属租户
	CenterInstanceName string            // 所属服务中心实例名，不是本节点 ID
	NamespaceID        string            // 命名空间
	GroupName          string            // 服务分组
	ServiceName        string            // 所属服务，节点必须挂在服务下
	IP                 string            // 发布地址
	Port               int               // 发布端口
	Weight             float64           // 负载权重
	Ephemeral          bool              // true 随连接或心跳过期，超时从库删除；false 超时标 Unhealthy
	Status             string            // 见 NodeUP / NodeDown / NodeOutOfService
	HealthyStatus      string            // 见 Healthy / Unhealthy / Unknown
	Metadata           map[string]string // 扩展元数据
	ConnectionID       string            // 注册该节点的数据面连接，用于断连清理；副本上为空
	OwnerGatewayID     string            // 持有心跳的网关进程 ID；空表示未认领
	RegisterTime       time.Time         // 首次注册时间
	LastBeatTime       time.Time         // 最近心跳（仅 Owner 网关更新）
}

// IsHealthy 表示对外可被发现：状态 UP 且探测 HEALTHY。
func (s *Node) IsHealthy() bool {
	return s != nil && s.Status == NodeUP && s.HealthyStatus == Healthy
}
