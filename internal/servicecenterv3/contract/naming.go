package contract

import (
	"context"

	"gateway/internal/servicecenterv3/model"
)

// Naming 是注册、发现、心跳与订阅。实现按 CenterInstance 隔离缓存。
//
// 临时和持久节点都写库。心跳刷新本机 Cache、续集群活视图，并节流回写 LastBeatTime。
// 连接丢失或心跳超时剔除。集群变更走 HUB_CLUSTER_EVENT；发现再叠 Redis 活视图。
// ch 由调用方创建并负责关闭；实现不得 close(ch)。
type Naming interface {
	// RegisterService 注册或覆盖服务定义。
	RegisterService(ctx context.Context, cc CallContext, svc *model.Service) error
	// UpdateService 更新已存在的服务定义。
	UpdateService(ctx context.Context, cc CallContext, svc *model.Service) error
	// DeregisterService 数据面只摘本连接节点；管理面（ConnectionID 空）才删整服务。
	DeregisterService(ctx context.Context, cc CallContext, group, serviceName string) error
	// GetService 读取服务定义，节点列表为本机 Cache 叠集群活视图后的快照。
	GetService(ctx context.Context, cc CallContext, group, serviceName string) (*model.Service, error)
	// ListServices 列出命名空间下服务；group 空表示全部组。
	ListServices(ctx context.Context, cc CallContext, group string) ([]*model.Service, error)

	// RegisterNode 注册业务节点。临时节点绑定 connectionID 以便断连清理，并写库供各网关加载。
	RegisterNode(ctx context.Context, cc CallContext, node *model.Node) error
	// DeregisterNode 按 nodeId 下线。
	DeregisterNode(ctx context.Context, cc CallContext, nodeID string) error
	// UpdateNode 更新权重、元数据或状态。
	UpdateNode(ctx context.Context, cc CallContext, node *model.Node) error
	// ListNodes 列出业务节点；healthyOnly 为 true 时只返回健康且 UP 的节点。
	ListNodes(ctx context.Context, cc CallContext, group, serviceName string, healthyOnly bool) ([]*model.Node, error)

	// Heartbeat 刷新 LastBeatTime；snapshot 是 SDK 带上的注册快照：缓存或库已有则只续期，没有才补上。
	Heartbeat(ctx context.Context, cc CallContext, nodeID string, snapshot *model.Service) error

	// SubscribeServices 订阅指定服务的变更，事件写入 ch（锁外阻塞投递，不丢事件）。
	SubscribeServices(ctx context.Context, cc CallContext, group string, serviceNames []string, ch chan<- model.NamingEvent) error
	// SubscribeNamespace 订阅命名空间（可选组）下全部服务变更。
	SubscribeNamespace(ctx context.Context, cc CallContext, group string, ch chan<- model.NamingEvent) error
	// Unsubscribe 取消 ch 上的全部订阅，不关闭 ch。
	Unsubscribe(ctx context.Context, cc CallContext, ch chan<- model.NamingEvent)
	// UnsubscribeServices 取消 ch 上指定服务的订阅；names 空则取消该组下按服务名订阅的条目。
	UnsubscribeServices(ctx context.Context, cc CallContext, group string, names []string, ch chan<- model.NamingEvent)
	// BindClient 把订阅通道绑到数据面连接，断连时由 OnClientLost 取消。
	BindClient(ch chan<- model.NamingEvent, connectionID string)

	// OnClientLost 连接断开。reason 见 model.DisconnectClientLost / DisconnectServerDrain：
	// SDK 离开则剔除该连接上的临时节点；网关停机则保留节点、只交还 Owner，等 SDK 重连或心跳超时。
	OnClientLost(ctx context.Context, connectionID string, reason string)
}
