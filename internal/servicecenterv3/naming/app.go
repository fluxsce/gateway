package naming

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"gateway/internal/servicecenterv3/contract"
	"gateway/internal/servicecenterv3/infra/alert"
	"gateway/internal/servicecenterv3/infra/cache"
	"gateway/internal/servicecenterv3/infra/live"
	"gateway/internal/servicecenterv3/infra/store"
	"gateway/internal/servicecenterv3/model"
	"gateway/pkg/logger"
	"gateway/pkg/utils/random"
)

// errNodeBeatFresh 库行心跳比驱逐快照新，节点仍存活，不得从 Cache 摘掉。
var errNodeBeatFresh = errors.New("servicecenterv3: node beat is newer than evict snapshot")

// App 实现 contract.Naming，绑定单一 CenterInstanceName。
type App struct {
	centerName       string
	cfg              *model.CenterInstance // 可被 Reload 替换，供告警与配额读取
	store            *store.Store
	cache            *cache.Cache
	live             *live.Store
	subs             *subscriber
	connIndex        sync.Map // connectionID -> []nodeID
	ownerSeen        sync.Map // remote OwnerGatewayID -> last event time.Time
	lastPersistBeat  sync.Map // nodeID -> time.Time，心跳写库节流
	beatPersistEvery time.Duration
	viewWarm         atomic.Bool
	liveDown         atomic.Bool
	liveFails        atomic.Int32
	lastLiveRefresh  atomic.Int64 // unix nano；对账成功后发现信 L1
	liveSyncStop     chan struct{}
	liveSyncWG       sync.WaitGroup
}

// New 构造命名应用。c 必须是该中心实例私有的 Cache，不得跨实例复用。
func New(centerName string, st *store.Store, c *cache.Cache) *App {
	a := &App{
		centerName:       centerName,
		store:            st,
		cache:            c,
		subs:             newSubscriber(),
		beatPersistEvery: 10 * time.Second,
	}
	a.viewWarm.Store(true)
	return a
}

// SetCenter 绑定中心实例定义（告警、配额）。Reload 后再次调用以换指针。
func (a *App) SetCenter(cfg *model.CenterInstance) {
	a.cfg = cfg
	if cfg != nil && cfg.InstanceName != "" {
		a.centerName = cfg.InstanceName
	}
}

// ensureNamespace 校验命名空间存在、属于当前中心实例且未停用；未命中缓存时回源库。
func (a *App) ensureNamespace(ctx context.Context, cc contract.CallContext) error {
	if err := cc.RequireNamespace(); err != nil {
		return err
	}
	if !cc.AllowsNamespace(cc.NamespaceID) {
		return contract.ErrNamespaceForbidden
	}
	ns, ok := a.cache.GetNamespace(cc.NamespaceID)
	if !ok {
		loaded, err := a.store.Namespace.Get(ctx, cc.TenantID, cc.NamespaceID)
		if err != nil {
			return err
		}
		if loaded == nil {
			return contract.ErrNamespaceNotFound
		}
		a.cache.SetNamespace(loaded)
		ns = loaded
	}
	if ns.CenterInstanceName != cc.CenterInstanceName {
		return contract.ErrNamespaceNotFound
	}
	if !ns.Active {
		return contract.ErrNamespaceDisabled
	}
	return nil
}

// RegisterService 注册或覆盖服务定义，写穿数据库并推送变更。
func (a *App) RegisterService(ctx context.Context, cc contract.CallContext, svc *model.Service) error {
	if err := a.ensureNamespace(ctx, cc); err != nil {
		return err
	}
	if svc == nil || svc.ServiceName == "" {
		return contract.ErrInvalidArgument
	}
	svc.TenantID = cc.TenantID
	svc.CenterInstanceName = cc.CenterInstanceName
	svc.NamespaceID = cc.NamespaceID
	if svc.GroupName == "" {
		svc.GroupName = "DEFAULT_GROUP"
	}
	_, exists := a.cache.GetService(cc.NamespaceID, svc.GroupName, svc.ServiceName)
	if !exists {
		if loaded, err := a.lookupPersistedService(ctx, cc, svc.GroupName, svc.ServiceName); err != nil {
			return err
		} else if loaded != nil {
			exists = true
		}
	}
	if !exists {
		if err := a.checkServiceQuota(cc); err != nil {
			return err
		}
	}
	svc.AutoCreated = false
	// 节点只走 RegisterNode。带 Nodes 写进 SetService 会留下空 NodeID 幽灵，发现名单会变成 2 条。
	svc.Nodes = nil
	if a.store != nil && a.store.Service != nil {
		if err := a.store.Service.Upsert(ctx, svc, cc.OperatorID); err != nil {
			return err
		}
	}
	a.cache.SetService(svc)
	eventType := model.EventServiceAdded
	if exists {
		eventType = model.EventServiceUpdated
	}
	a.emit(model.NamingEvent{
		Type:               eventType,
		Timestamp:          time.Now(),
		CenterInstanceName: cc.CenterInstanceName,
		NamespaceID:        cc.NamespaceID,
		GroupName:          svc.GroupName,
		ServiceName:        svc.ServiceName,
		Service:            svc,
	})
	return nil
}

// UpdateService 更新已存在的服务定义；服务不存在时返回 ErrServiceNotFound。
func (a *App) UpdateService(ctx context.Context, cc contract.CallContext, svc *model.Service) error {
	if err := a.ensureNamespace(ctx, cc); err != nil {
		return err
	}
	if svc == nil || svc.ServiceName == "" {
		return contract.ErrInvalidArgument
	}
	svc.TenantID = cc.TenantID
	svc.CenterInstanceName = cc.CenterInstanceName
	svc.NamespaceID = cc.NamespaceID
	svc.AutoCreated = false
	if err := a.store.Service.Upsert(ctx, svc, cc.OperatorID); err != nil {
		return err
	}
	a.cache.SetService(svc)
	a.emit(model.NamingEvent{
		Type:               model.EventServiceUpdated,
		Timestamp:          time.Now(),
		CenterInstanceName: cc.CenterInstanceName,
		NamespaceID:        cc.NamespaceID,
		GroupName:          svc.GroupName,
		ServiceName:        svc.ServiceName,
		Service:            svc,
	})
	return nil
}

// DeregisterService 数据面（有 ConnectionID）只摘本连接在该服务下的节点，不删别人的实例、不推 SERVICE_DELETED。
// 管理面（ConnectionID 空）才清全部节点、删服务定义并广播 SERVICE_DELETED。
func (a *App) DeregisterService(ctx context.Context, cc contract.CallContext, group, serviceName string) error {
	if err := a.ensureNamespace(ctx, cc); err != nil {
		// 管理面删除：命名空间停用仍要清运行时与目录，避免 0 节点空壳留在列表。
		if cc.ConnectionID != "" || !errors.Is(err, contract.ErrNamespaceDisabled) {
			return err
		}
	}
	if group == "" {
		group = "DEFAULT_GROUP"
	}
	if cc.ConnectionID != "" {
		a.removeConnectionServiceNodes(ctx, cc, group, serviceName)
		return nil
	}
	a.refreshLiveView(ctx)
	svc, ok := a.cache.GetService(cc.NamespaceID, group, serviceName)
	if ok {
		svc = a.overlayLiveService(ctx, cc.TenantID, svc)
	} else if liveSvc := a.serviceFromLive(ctx, cc.TenantID, cc.NamespaceID, group, serviceName); liveSvc != nil {
		svc = liveSvc
		ok = true
	}
	if ok {
		for _, inst := range append([]*model.Node{}, svc.Nodes...) {
			_ = a.removeNode(ctx, cc, inst, model.EventNodeDeregistered)
		}
	}
	if a.store != nil && a.store.Service != nil {
		_ = a.store.Service.Delete(ctx, cc.TenantID, cc.NamespaceID, group, serviceName)
	}
	a.cache.DeleteService(cc.NamespaceID, group, serviceName)
	a.emit(model.NamingEvent{
		Type:               model.EventServiceDeleted,
		Timestamp:          time.Now(),
		CenterInstanceName: cc.CenterInstanceName,
		NamespaceID:        cc.NamespaceID,
		GroupName:          group,
		ServiceName:        serviceName,
	})
	return nil
}

// removeConnectionServiceNodes 只注销本连接挂在该服务下的节点。
func (a *App) removeConnectionServiceNodes(ctx context.Context, cc contract.CallContext, group, serviceName string) {
	seen := map[string]struct{}{}
	var victims []*model.Node
	if svc, ok := a.cache.GetService(cc.NamespaceID, group, serviceName); ok {
		for _, inst := range svc.Nodes {
			if inst == nil || inst.ConnectionID != cc.ConnectionID {
				continue
			}
			victims = append(victims, inst)
			seen[inst.NodeID] = struct{}{}
		}
	}
	for _, id := range a.connectionNodeIDs(cc.ConnectionID) {
		if _, ok := seen[id]; ok {
			continue
		}
		inst, ok := a.cache.GetNode(id)
		if !ok || inst.NamespaceID != cc.NamespaceID || inst.GroupName != group || inst.ServiceName != serviceName {
			continue
		}
		victims = append(victims, inst)
	}
	for _, inst := range victims {
		_ = a.removeNode(ctx, cc, inst, model.EventNodeDeregistered)
	}
}

func (a *App) connectionNodeIDs(connectionID string) []string {
	if connectionID == "" {
		return nil
	}
	raw, ok := a.connIndex.Load(connectionID)
	if !ok {
		return nil
	}
	ids, _ := raw.([]string)
	return append([]string{}, ids...)
}

// GetService 先读缓存，再叠集群活视图；未命中则回源库并回填。
func (a *App) GetService(ctx context.Context, cc contract.CallContext, group, serviceName string) (*model.Service, error) {
	if err := a.ensureNamespace(ctx, cc); err != nil {
		return nil, err
	}
	if group == "" {
		group = "DEFAULT_GROUP"
	}
	if svc, ok := a.cache.GetService(cc.NamespaceID, group, serviceName); ok {
		return a.overlayLiveService(ctx, cc.TenantID, svc), nil
	}
	if liveSvc := a.serviceFromLive(ctx, cc.TenantID, cc.NamespaceID, group, serviceName); liveSvc != nil {
		return liveSvc, nil
	}
	if a.store == nil || a.store.Service == nil {
		if !a.IsViewWarm() {
			return nil, contract.ErrViewNotReady
		}
		return nil, contract.ErrServiceNotFound
	}
	svc, err := a.store.Service.Get(ctx, cc.TenantID, cc.NamespaceID, group, serviceName)
	if err != nil {
		return nil, err
	}
	if svc == nil {
		if !a.IsViewWarm() {
			return nil, contract.ErrViewNotReady
		}
		return nil, contract.ErrServiceNotFound
	}
	svc.CenterInstanceName = cc.CenterInstanceName
	nodes, err := a.store.Node.ListByService(ctx, cc.TenantID, cc.NamespaceID, group, serviceName)
	if err != nil {
		return nil, err
	}
	kept := make([]*model.Node, 0, len(nodes))
	for _, n := range nodes {
		if a.ingestStoredNode(n) {
			kept = append(kept, n)
		}
	}
	svc.Nodes = kept
	a.cache.SetService(svc)
	return a.overlayLiveService(ctx, cc.TenantID, svc), nil
}

// ListServices 列出命名空间下服务；group 空表示全部组。
func (a *App) ListServices(ctx context.Context, cc contract.CallContext, group string) ([]*model.Service, error) {
	if err := a.ensureNamespace(ctx, cc); err != nil {
		return nil, err
	}
	a.refreshLiveView(ctx)
	cached := a.cache.ListServices(cc.NamespaceID, group)
	if len(cached) > 0 {
		return cached, nil
	}
	if a.store == nil || a.store.Service == nil {
		return cached, nil
	}
	list, err := a.store.Service.ListByNamespace(ctx, cc.TenantID, cc.NamespaceID, group)
	if err != nil {
		return nil, err
	}
	for _, svc := range list {
		svc.CenterInstanceName = cc.CenterInstanceName
		nodes, _ := a.store.Node.ListByService(ctx, cc.TenantID, cc.NamespaceID, svc.GroupName, svc.ServiceName)
		kept := make([]*model.Node, 0, len(nodes))
		for _, n := range nodes {
			if a.ingestStoredNode(n) {
				kept = append(kept, n)
			}
		}
		svc.Nodes = kept
		a.cache.SetService(svc)
	}
	a.refreshLiveView(ctx)
	out := a.cache.ListServices(cc.NamespaceID, group)
	if len(out) == 0 {
		out = list
	}
	return out, nil
}

// RegisterNode 注册节点：写本机 Cache 和 HUB_SERVICE_NODE，再发集群变更事件。
func (a *App) RegisterNode(ctx context.Context, cc contract.CallContext, inst *model.Node) error {
	if err := a.ensureNamespace(ctx, cc); err != nil {
		return err
	}
	if inst == nil || inst.IP == "" || inst.Port <= 0 {
		return contract.ErrInvalidArgument
	}
	if inst.GroupName == "" {
		inst.GroupName = "DEFAULT_GROUP"
	}
	if inst.ServiceName == "" {
		return contract.ErrInvalidArgument
	}
	if inst.NodeID == "" {
		inst.NodeID = random.Generate32BitRandomString()
	} else if !validNodeID(inst.NodeID) {
		return contract.ErrInvalidArgument
	}
	inst.TenantID = cc.TenantID
	inst.CenterInstanceName = cc.CenterInstanceName
	inst.NamespaceID = cc.NamespaceID
	if err := a.rejectTakenNodeID(ctx, inst); err != nil {
		return err
	}
	now := time.Now()
	inst.RegisterTime = now
	inst.LastBeatTime = now
	if inst.Status == "" {
		inst.Status = model.NodeUP
	}
	if inst.HealthyStatus == "" {
		inst.HealthyStatus = model.Healthy
	}
	if inst.Weight <= 0 {
		inst.Weight = 1
	}
	if LocalGatewayID != "" {
		inst.OwnerGatewayID = LocalGatewayID
	}
	if err := a.ensureServiceForNode(ctx, cc, inst); err != nil {
		return err
	}
	if err := a.persistNode(ctx, inst, cc.OperatorID); err != nil {
		return err
	}
	prev, existed := a.cache.GetNode(inst.NodeID)
	a.cache.PutNode(inst)
	a.liveWriteOnRegister(ctx, inst, prev)
	if inst.ConnectionID != "" {
		a.trackConnection(inst.ConnectionID, inst.NodeID)
	}
	alert.NodeRegister(a.cfg, alert.NodeInfo{
		NodeID:      inst.NodeID,
		ServiceName: inst.ServiceName,
		NamespaceID: inst.NamespaceID,
		GroupName:   inst.GroupName,
		IP:          inst.IP,
		Port:        inst.Port,
		Reconnect:   existed,
	})
	svc, _ := a.cache.GetService(inst.NamespaceID, inst.GroupName, inst.ServiceName)
	a.emit(model.NamingEvent{
		Type:               model.EventNodeRegistered,
		Timestamp:          now,
		CenterInstanceName: cc.CenterInstanceName,
		NamespaceID:        inst.NamespaceID,
		GroupName:          inst.GroupName,
		ServiceName:        inst.ServiceName,
		Service:            svc,
		ChangedNode:        inst,
	})
	return nil
}

// DeregisterNode 按 nodeID 下线节点并推送 NODE_DEREGISTERED。
func (a *App) DeregisterNode(ctx context.Context, cc contract.CallContext, nodeID string) error {
	if err := cc.RequireCenter(); err != nil {
		return err
	}
	inst, ok := a.cache.GetNode(nodeID)
	if !ok {
		loaded, err := a.storeGet(ctx, cc.TenantID, nodeID)
		if err != nil {
			return err
		}
		if loaded == nil {
			loaded = a.liveGet(ctx, &model.Node{TenantID: cc.TenantID, NodeID: nodeID})
		}
		if loaded == nil {
			return contract.ErrNodeNotFound
		}
		inst = loaded
	}
	if cc.ConnectionID != "" && inst.ConnectionID != "" && inst.ConnectionID != cc.ConnectionID {
		return contract.ErrNodeNotFound
	}
	return a.removeNode(ctx, cc, inst, model.EventNodeDeregistered)
}

// UpdateNode 更新节点权重、元数据或状态；持久实例同步写库。
func (a *App) UpdateNode(ctx context.Context, cc contract.CallContext, inst *model.Node) error {
	if inst == nil || inst.NodeID == "" {
		return contract.ErrInvalidArgument
	}
	current, ok := a.cache.GetNode(inst.NodeID)
	if !ok {
		return contract.ErrNodeNotFound
	}
	if inst.IP != "" {
		current.IP = inst.IP
	}
	if inst.Port > 0 {
		current.Port = inst.Port
	}
	if inst.Weight > 0 {
		current.Weight = inst.Weight
	}
	if inst.Status != "" {
		current.Status = inst.Status
	}
	if inst.HealthyStatus != "" {
		current.HealthyStatus = inst.HealthyStatus
	}
	if inst.Metadata != nil {
		current.Metadata = inst.Metadata
	}
	current.LastBeatTime = time.Now()
	if err := a.persistNode(ctx, current, cc.OperatorID); err != nil {
		return err
	}
	a.cache.PutNode(current)
	liveErr := a.putLive(ctx, current)
	eventType := model.EventNodeUpdated
	if current.Status == model.NodeDown || current.Status == model.NodeOutOfService {
		eventType = model.EventNodeOffline
	}
	svc, _ := a.cache.GetService(current.NamespaceID, current.GroupName, current.ServiceName)
	a.emit(model.NamingEvent{
		Type:               eventType,
		Timestamp:          time.Now(),
		CenterInstanceName: cc.CenterInstanceName,
		NamespaceID:        current.NamespaceID,
		GroupName:          current.GroupName,
		ServiceName:        current.ServiceName,
		Service:            svc,
		ChangedNode:        current,
	})
	return liveErr
}

// ListNodes 列出服务下节点；healthyOnly 为 true 时只返回 UP 且 HEALTHY。
func (a *App) ListNodes(ctx context.Context, cc contract.CallContext, group, serviceName string, healthyOnly bool) ([]*model.Node, error) {
	svc, err := a.GetService(ctx, cc, group, serviceName)
	if err != nil {
		return nil, err
	}
	out := make([]*model.Node, 0, len(svc.Nodes))
	for _, inst := range svc.Nodes {
		if healthyOnly && !inst.IsHealthy() {
			continue
		}
		out = append(out, inst)
	}
	if healthyOnly && len(out) == 0 && !a.IsViewWarm() {
		return nil, contract.ErrViewNotReady
	}
	return out, nil
}

// Heartbeat 刷新本机缓存 LastBeatTime，续活视图 TTL，按节流回写库。缓存没有则先读活视图（还在则只续 TTL），再回源库，最后用心跳快照补注册。认领 Owner 才发集群事件。
func (a *App) Heartbeat(ctx context.Context, cc contract.CallContext, nodeID string, snapshot *model.Service) error {
	if nodeID == "" {
		return contract.ErrInvalidArgument
	}
	if ok, claimed := a.cache.TouchBeat(nodeID, time.Now(), LocalGatewayID); ok {
		if inst, found := a.cache.GetNode(nodeID); found {
			if claimed {
				a.livePut(ctx, inst)
			} else {
				a.liveTouch(ctx, inst)
			}
			a.persistBeatIfDue(ctx, inst, claimed)
			if claimed {
				svc, _ := a.cache.GetService(inst.NamespaceID, inst.GroupName, inst.ServiceName)
				a.emit(model.NamingEvent{
					Type:               model.EventNodeUpdated,
					Timestamp:          time.Now(),
					CenterInstanceName: cc.CenterInstanceName,
					NamespaceID:        inst.NamespaceID,
					GroupName:          inst.GroupName,
					ServiceName:        inst.ServiceName,
					Service:            svc,
					ChangedNode:        inst,
				})
			}
		}
		return nil
	}
	if live := a.liveGet(ctx, &model.Node{TenantID: cc.TenantID, NodeID: nodeID}); live != nil {
		prevOwner := live.OwnerGatewayID
		prevConn := live.ConnectionID
		a.claimHeartbeatNode(live, snapshot)
		a.cache.PutNode(live)
		ownerChanged := LocalGatewayID != "" && prevOwner != LocalGatewayID
		connChanged := live.ConnectionID != "" && live.ConnectionID != prevConn
		if ownerChanged || connChanged {
			a.livePut(ctx, live)
			a.persistBeatIfDue(ctx, live, true)
			if ownerChanged {
				svc, _ := a.cache.GetService(live.NamespaceID, live.GroupName, live.ServiceName)
				a.emit(model.NamingEvent{
					Type:               model.EventNodeUpdated,
					Timestamp:          time.Now(),
					CenterInstanceName: cc.CenterInstanceName,
					NamespaceID:        live.NamespaceID,
					GroupName:          live.GroupName,
					ServiceName:        live.ServiceName,
					Service:            svc,
					ChangedNode:        live,
				})
			}
		} else {
			a.liveTouch(ctx, live)
		}
		return nil
	}
	if loaded, err := a.storeGet(ctx, cc.TenantID, nodeID); err != nil {
		return err
	} else if loaded != nil {
		a.claimHeartbeatNode(loaded, snapshot)
		a.cache.PutNode(loaded)
		a.livePut(ctx, loaded)
		a.persistBeatIfDue(ctx, loaded, true)
		return nil
	}
	if inst := nodeFromHeartbeat(snapshot, nodeID); inst != nil {
		inst.TenantID = cc.TenantID
		inst.CenterInstanceName = cc.CenterInstanceName
		if inst.NamespaceID == "" {
			inst.NamespaceID = cc.NamespaceID
		}
		return a.RegisterNode(ctx, cc, inst)
	}
	return contract.ErrNodeNotFound
}

func (a *App) claimHeartbeatNode(inst *model.Node, snapshot *model.Service) {
	if inst == nil {
		return
	}
	inst.LastBeatTime = time.Now()
	inst.HealthyStatus = model.Healthy
	if LocalGatewayID != "" {
		inst.OwnerGatewayID = LocalGatewayID
	}
	if snap := nodeFromHeartbeat(snapshot, inst.NodeID); snap != nil && snap.ConnectionID != "" {
		inst.ConnectionID = snap.ConnectionID
		a.trackConnection(snap.ConnectionID, inst.NodeID)
	}
}

func nodeFromHeartbeat(snapshot *model.Service, nodeID string) *model.Node {
	if snapshot == nil {
		return nil
	}
	for _, inst := range snapshot.Nodes {
		if inst != nil && (inst.NodeID == nodeID || inst.NodeID == "") {
			if inst.NodeID == "" {
				inst.NodeID = nodeID
			}
			if inst.ServiceName == "" {
				inst.ServiceName = snapshot.ServiceName
			}
			if inst.GroupName == "" {
				inst.GroupName = snapshot.GroupName
			}
			if inst.NamespaceID == "" {
				inst.NamespaceID = snapshot.NamespaceID
			}
			return inst
		}
	}
	return nil
}

// SubscribeServices 订阅指定服务的变更；投递阻塞等待，通道满不丢事件。
func (a *App) SubscribeServices(_ context.Context, cc contract.CallContext, group string, serviceNames []string, ch chan<- model.NamingEvent) error {
	if err := cc.RequireNamespace(); err != nil {
		return err
	}
	if !cc.AllowsNamespace(cc.NamespaceID) {
		return contract.ErrNamespaceForbidden
	}
	if group == "" {
		group = "DEFAULT_GROUP"
	}
	a.subs.add(cc.NamespaceID, group, serviceNames, ch, cc.ConnectionID)
	a.snapshotServices(cc.NamespaceID, group, serviceNames, ch)
	alert.Subscribe(a.cfg, alert.SubscribeInfo{
		Action:       "SUBSCRIBE",
		SubscriberID: cc.ConnectionID,
		NamespaceID:  cc.NamespaceID,
		GroupName:    group,
		ServiceNames: serviceNames,
	})
	return nil
}

// SubscribeNamespace 订阅命名空间（可选组）下全部服务变更。
func (a *App) SubscribeNamespace(_ context.Context, cc contract.CallContext, group string, ch chan<- model.NamingEvent) error {
	if err := cc.RequireNamespace(); err != nil {
		return err
	}
	if !cc.AllowsNamespace(cc.NamespaceID) {
		return contract.ErrNamespaceForbidden
	}
	a.subs.addNamespace(cc.NamespaceID, group, ch, cc.ConnectionID)
	a.snapshotServices(cc.NamespaceID, group, nil, ch)
	alert.Subscribe(a.cfg, alert.SubscribeInfo{
		Action:       "SUBSCRIBE_NAMESPACE",
		SubscriberID: cc.ConnectionID,
		NamespaceID:  cc.NamespaceID,
		GroupName:    group,
	})
	return nil
}

// ListSubscribers 列出订了该服务的对端服务（由订阅连接上的注册节点反查）。
func (a *App) ListSubscribers(namespaceID, group, serviceName string) []model.ServiceSubscriber {
	if a == nil || a.subs == nil {
		return nil
	}
	return a.expandInboundPeerServices(
		subscribersFromMatches(a, a.subs.listMatching(namespaceID, group, serviceName)),
		namespaceID, group, serviceName,
	)
}

// ListSubscriptions 列出本服务节点所在连接订了哪些服务（本进程）。
func (a *App) ListSubscriptions(namespaceID, group, serviceName string) []model.ServiceSubscriber {
	if a == nil || a.subs == nil {
		return nil
	}
	matches := a.subs.listByConnections(a.connectionIDsOfService(namespaceID, group, serviceName))
	kept := matches[:0]
	for _, m := range matches {
		if m.scope == "service" && m.namespaceID == namespaceID && m.group == group && m.serviceName == serviceName {
			continue
		}
		kept = append(kept, m)
	}
	return subscribersFromMatches(a, kept)
}

func (a *App) connectionIDsOfService(namespaceID, group, serviceName string) map[string]struct{} {
	out := map[string]struct{}{}
	if a == nil || a.cache == nil {
		return out
	}
	svc, ok := a.cache.GetService(namespaceID, group, serviceName)
	if !ok || svc == nil {
		return out
	}
	for _, node := range svc.Nodes {
		if node != nil && node.ConnectionID != "" {
			out[node.ConnectionID] = struct{}{}
		}
	}
	return out
}

func (a *App) expandInboundPeerServices(items []model.ServiceSubscriber, namespaceID, group, serviceName string) []model.ServiceSubscriber {
	if len(items) == 0 {
		return nil
	}
	out := make([]model.ServiceSubscriber, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		for _, node := range a.servicesOfConnection(item.ConnectionID) {
			if node.NamespaceID == namespaceID && node.GroupName == group && node.ServiceName == serviceName {
				continue
			}
			key := item.ConnectionID + ":" + node.NamespaceID + ":" + node.GroupName + ":" + node.ServiceName
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			row := item
			row.NamespaceID = node.NamespaceID
			row.GroupName = node.GroupName
			row.ServiceName = node.ServiceName
			out = append(out, row)
		}
	}
	return out
}

func (a *App) servicesOfConnection(connectionID string) []*model.Node {
	if a == nil || a.cache == nil || connectionID == "" {
		return nil
	}
	seen := map[string]struct{}{}
	out := []*model.Node{}
	add := func(node *model.Node) {
		if node == nil || node.ServiceName == "" {
			return
		}
		key := node.NamespaceID + ":" + node.GroupName + ":" + node.ServiceName
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, node)
	}
	if raw, ok := a.connIndex.Load(connectionID); ok {
		ids, _ := raw.([]string)
		for _, id := range ids {
			if node, found := a.cache.GetNode(id); found {
				add(node)
			}
		}
	}
	if len(out) == 0 {
		for _, node := range a.cache.AllNodes() {
			if node != nil && node.ConnectionID == connectionID {
				add(node)
			}
		}
	}
	return out
}

func subscribersFromMatches(a *App, matches []subMatch) []model.ServiceSubscriber {
	if a == nil || a.subs == nil || len(matches) == 0 {
		return nil
	}
	out := make([]model.ServiceSubscriber, 0, len(matches))
	for _, m := range matches {
		out = append(out, model.ServiceSubscriber{
			ConnectionID: m.connectionID,
			NamespaceID:  m.namespaceID,
			GroupName:    m.group,
			ServiceName:  m.serviceName,
			Scope:        m.scope,
		})
	}
	return out
}

// Unsubscribe 取消 ch 上的全部订阅，不关闭 ch。
func (a *App) Unsubscribe(_ context.Context, _ contract.CallContext, ch chan<- model.NamingEvent) {
	a.subs.remove(ch)
}

// UnsubscribeServices 取消 ch 上指定服务的订阅。
func (a *App) UnsubscribeServices(_ context.Context, cc contract.CallContext, group string, names []string, ch chan<- model.NamingEvent) {
	if group == "" {
		group = "DEFAULT_GROUP"
	}
	a.subs.removeServices(ch, cc.NamespaceID, group, names)
}

// OnClientLost 先取消该连接的订阅。SDK 断流驱逐临时节点；网关停机只交还 Owner，节点留在发现表。
func (a *App) OnClientLost(ctx context.Context, connectionID string, reason string) {
	a.subs.removeByConnection(connectionID)
	raw, ok := a.connIndex.Load(connectionID)
	if !ok {
		return
	}
	ids, _ := raw.([]string)
	a.connIndex.Delete(connectionID)
	drain := model.IsServerDrain(reason)
	cc := contract.CallContext{CenterInstanceName: a.centerName}
	for _, id := range ids {
		inst, found := a.cache.GetNode(id)
		if !found || a.skipLostNode(inst, connectionID) {
			continue
		}
		cc.TenantID = inst.TenantID
		if drain {
			a.releaseOwnedNode(ctx, inst)
			continue
		}
		if !inst.Ephemeral {
			continue
		}
		_ = a.removeNode(ctx, cc, inst, model.EventNodeEvicted)
	}
}

// skipLostNode 跳过已迁走 Owner 或已绑到新连接的节点，避免旧流清理误伤重连。
func (a *App) skipLostNode(inst *model.Node, connectionID string) bool {
	if inst == nil {
		return true
	}
	if LocalGatewayID != "" && inst.OwnerGatewayID != "" && inst.OwnerGatewayID != LocalGatewayID {
		return true
	}
	if inst.ConnectionID != "" && inst.ConnectionID != connectionID {
		return true
	}
	return false
}

// releaseOwnedNode 网关停机：节点还在，清掉本机连接与 Owner，刷新心跳以免副本立刻当尸体删。
func (a *App) releaseOwnedNode(ctx context.Context, inst *model.Node) {
	if inst == nil {
		return
	}
	now := time.Now()
	inst.ConnectionID = ""
	inst.OwnerGatewayID = ""
	inst.LastBeatTime = now
	if err := a.persistNode(ctx, inst, ""); err != nil {
		logger.Warn("网关停机交还节点写库失败", "nodeId", inst.NodeID, "error", err)
	} else {
		a.lastPersistBeat.Store(inst.NodeID, now)
	}
	a.cache.PutNode(inst)
	a.livePut(ctx, inst)
	svc, _ := a.cache.GetService(inst.NamespaceID, inst.GroupName, inst.ServiceName)
	a.emit(model.NamingEvent{
		Type:               model.EventNodeUpdated,
		Timestamp:          now,
		CenterInstanceName: a.centerName,
		NamespaceID:        inst.NamespaceID,
		GroupName:          inst.GroupName,
		ServiceName:        inst.ServiceName,
		Service:            svc,
		ChangedNode:        inst,
	})
}

// LoadPersistent 从库加载本中心下的命名空间、服务与在册节点到缓存。
func (a *App) LoadPersistent(ctx context.Context, tenantID string) error {
	namespaces, err := a.store.Namespace.ListByCenter(ctx, tenantID, a.centerName)
	if err != nil {
		return err
	}
	for _, ns := range namespaces {
		a.cache.SetNamespace(ns)
		services, err := a.store.Service.ListByNamespace(ctx, tenantID, ns.NamespaceID, "")
		if err != nil {
			logger.Warn("加载服务失败", "namespaceId", ns.NamespaceID, "error", err)
			continue
		}
		for _, svc := range services {
			svc.CenterInstanceName = a.centerName
			nodes, err := a.store.Node.ListByService(ctx, tenantID, ns.NamespaceID, svc.GroupName, svc.ServiceName)
			if err != nil {
				continue
			}
			kept := make([]*model.Node, 0, len(nodes))
			for _, n := range nodes {
				if a.ingestStoredNode(n) {
					kept = append(kept, n)
				}
			}
			hadStored := len(nodes) > 0
			svc.Nodes = kept
			if len(kept) == 0 && svc.IsInternal() && (svc.AutoCreated || hadStored) {
				if a.store != nil && a.store.Service != nil {
					_ = a.store.Service.Delete(ctx, tenantID, ns.NamespaceID, svc.GroupName, svc.ServiceName)
				}
				continue
			}
			a.cache.SetService(svc)
		}
	}
	a.refreshLiveView(ctx)
	return nil
}

// ingestStoredNode 把库中的节点灌进缓存。
// 心跳仍在宽限期内的保持 HEALTHY；过期的持久节点标 Unhealthy，过期的临时节点从库删除且不进缓存。
func (a *App) ingestStoredNode(n *model.Node) bool {
	if n == nil {
		return false
	}
	n.CenterInstanceName = a.centerName
	n.OwnerGatewayID = ""
	n.ConnectionID = ""
	fresh := persistentBeatFresh(n.LastBeatTime, a.persistentHealthGrace())
	if n.Ephemeral && !fresh {
		if a.store != nil && a.store.Node != nil && n.TenantID != "" && n.NodeID != "" {
			_, _ = a.store.Node.DeleteIfBeatNotAfter(context.Background(), n.TenantID, n.NodeID, n.LastBeatTime)
		}
		return false
	}
	if fresh {
		if n.Status == "" {
			n.Status = model.NodeUP
		}
		if n.HealthyStatus == "" {
			n.HealthyStatus = model.Healthy
		}
	} else {
		n.HealthyStatus = model.Unhealthy
	}
	a.cache.PutNode(n)
	return true
}

func (a *App) persistentHealthGrace() time.Duration {
	timeout := 15 * time.Second
	if a.cfg != nil && a.cfg.HealthCheckTimeout > 0 {
		timeout = time.Duration(a.cfg.HealthCheckTimeout) * time.Second
	}
	return 2 * timeout
}

func persistentBeatFresh(lastBeat time.Time, grace time.Duration) bool {
	if lastBeat.IsZero() || grace <= 0 {
		return false
	}
	return time.Since(lastBeat) <= grace
}

// persistNode 临时和持久节点都写 HUB_SERVICE_NODE。
func (a *App) persistNode(ctx context.Context, inst *model.Node, operator string) error {
	if inst == nil || a.store == nil || a.store.Node == nil {
		return nil
	}
	return a.store.Node.Upsert(ctx, inst, operator)
}

func (a *App) storeGet(ctx context.Context, tenantID, nodeID string) (*model.Node, error) {
	if a.store == nil || a.store.Node == nil || nodeID == "" {
		return nil, nil
	}
	return a.store.Node.Get(ctx, tenantID, nodeID)
}

// persistBeatIfDue 心跳按节流回写 LastBeatTime，不发集群事件。
func (a *App) persistBeatIfDue(ctx context.Context, inst *model.Node, force bool) {
	if inst == nil {
		return
	}
	if !force && a.beatPersistEvery > 0 {
		if last, ok := a.lastPersistBeat.Load(inst.NodeID); ok {
			if t, ok := last.(time.Time); ok && time.Since(t) < a.beatPersistEvery {
				return
			}
		}
	}
	if a.store != nil && a.store.Node != nil {
		if err := a.store.Node.Upsert(ctx, inst, ""); err != nil {
			logger.Warn("节点心跳写库失败", "nodeId", inst.NodeID, "error", err)
			return
		}
	}
	if !force {
		a.livePut(ctx, inst)
	}
	a.lastPersistBeat.Store(inst.NodeID, time.Now())
}

// lookupPersistedService 从库读服务定义；未装配存储时返回 nil。
func (a *App) lookupPersistedService(ctx context.Context, cc contract.CallContext, group, name string) (*model.Service, error) {
	if a.store == nil || a.store.Service == nil {
		return nil, nil
	}
	return a.store.Service.Get(ctx, cc.TenantID, cc.NamespaceID, group, name)
}

func (a *App) pruneEmptyAfterLastNode(namespaceID, group, name, tenantID string, replicate bool) {
	svc, ok := a.cache.GetService(namespaceID, group, name)
	if !ok || len(svc.Nodes) > 0 {
		return
	}
	if tenantID == "" {
		tenantID = svc.TenantID
	} else if svc.TenantID == "" {
		svc.TenantID = tenantID
	}
	if svc.AutoCreated {
		a.pruneEmptyAutoService(namespaceID, group, name, replicate)
		return
	}
	if !svc.IsInternal() {
		return
	}
	a.cache.DeleteService(namespaceID, group, name)
	if a.store != nil && a.store.Service != nil && tenantID != "" {
		_ = a.store.Service.Delete(context.Background(), tenantID, namespaceID, group, name)
	}
	if !replicate {
		return
	}
	a.emit(model.NamingEvent{
		Type:               model.EventServiceDeleted,
		Timestamp:          time.Now(),
		CenterInstanceName: a.centerName,
		NamespaceID:        namespaceID,
		GroupName:          group,
		ServiceName:        name,
	})
}

func (a *App) pruneEmptyAutoServices() {
	if a == nil || a.cache == nil {
		return
	}
	for _, svc := range a.cache.AllServices() {
		if svc == nil {
			continue
		}
		a.pruneEmptyAutoService(svc.NamespaceID, svc.GroupName, svc.ServiceName, false)
	}
}

// pruneEmptyAutoService 临时节点都离开后，删掉 REGISTER_NODE 隐含创建的空服务。
// 不向 SDK 推 SERVICE_DELETED：订阅按服务名还在，随后再 REGISTER_NODE 仍会推 NODE_ADDED。
func (a *App) pruneEmptyAutoService(namespaceID, group, name string, replicate bool) {
	svc, ok := a.cache.GetService(namespaceID, group, name)
	if !ok || !svc.AutoCreated || len(svc.Nodes) > 0 {
		return
	}
	tenantID := svc.TenantID
	a.cache.DeleteService(namespaceID, group, name)
	if a.store != nil && a.store.Service != nil && tenantID != "" {
		_ = a.store.Service.Delete(context.Background(), tenantID, namespaceID, group, name)
	}
	if !replicate {
		return
	}
	ev := model.NamingEvent{
		Type:               model.EventServiceDeleted,
		Timestamp:          time.Now(),
		CenterInstanceName: a.centerName,
		NamespaceID:        namespaceID,
		GroupName:          group,
		ServiceName:        name,
	}
	if ev.Environment == "" {
		ev.Environment = a.environment()
	}
	replicateNaming(ev)
}

// ensureServiceForNode 保证节点挂在服务下：没有服务定义时先补一条并推 SERVICE_ADDED。
func (a *App) ensureServiceForNode(ctx context.Context, cc contract.CallContext, inst *model.Node) error {
	if inst == nil || inst.ServiceName == "" {
		return contract.ErrInvalidArgument
	}
	if _, ok := a.cache.GetService(inst.NamespaceID, inst.GroupName, inst.ServiceName); ok {
		return nil
	}
	if loaded, err := a.lookupPersistedService(ctx, cc, inst.GroupName, inst.ServiceName); err != nil {
		return err
	} else if loaded != nil {
		loaded.CenterInstanceName = inst.CenterInstanceName
		a.cache.SetService(loaded)
		return nil
	}
	if err := a.checkServiceQuota(cc); err != nil {
		return err
	}
	svc := &model.Service{
		TenantID:           inst.TenantID,
		CenterInstanceName: inst.CenterInstanceName,
		NamespaceID:        inst.NamespaceID,
		GroupName:          inst.GroupName,
		ServiceName:        inst.ServiceName,
		AutoCreated:        true,
	}
	a.cache.SetService(svc)
	a.emit(model.NamingEvent{
		Type:               model.EventServiceAdded,
		Timestamp:          time.Now(),
		CenterInstanceName: cc.CenterInstanceName,
		NamespaceID:        inst.NamespaceID,
		GroupName:          inst.GroupName,
		ServiceName:        inst.ServiceName,
		Service:            svc,
	})
	return nil
}

// removeNode 先按心跳条件删库，确认行已不在再清活视图和 Cache 并推送。
// 库行心跳更新则中止（不删 Redis），避免集群副本按空活视图误摘还活着的节点。
// 主动注销时活视图删失败则中止，避免 Redis 幽灵节点。
func (a *App) removeNode(ctx context.Context, cc contract.CallContext, inst *model.Node, eventType string) error {
	if inst == nil {
		return nil
	}
	if a.store != nil && a.store.Node != nil && inst.TenantID != "" && inst.NodeID != "" {
		deleted, delErr := a.store.Node.DeleteIfBeatNotAfter(ctx, inst.TenantID, inst.NodeID, inst.LastBeatTime)
		if delErr != nil {
			return delErr
		}
		if !deleted {
			loaded, getErr := a.store.Node.Get(ctx, inst.TenantID, inst.NodeID)
			if getErr != nil {
				return getErr
			}
			if loaded != nil {
				a.mergeStoredBeat(loaded)
				return errNodeBeatFresh
			}
		}
	}
	if err := a.liveDelete(ctx, inst); err != nil && eventType == model.EventNodeDeregistered {
		return err
	}
	a.cache.RemoveNode(inst.NodeID)
	a.lastPersistBeat.Delete(inst.NodeID)
	a.untrackConnection(inst.ConnectionID, inst.NodeID)
	if eventType == model.EventNodeDeregistered {
		alert.NodeUnregister(a.cfg, alert.NodeInfo{
			NodeID:      inst.NodeID,
			ServiceName: inst.ServiceName,
			NamespaceID: inst.NamespaceID,
			GroupName:   inst.GroupName,
			IP:          inst.IP,
			Port:        inst.Port,
		})
	}
	svc, _ := a.cache.GetService(inst.NamespaceID, inst.GroupName, inst.ServiceName)
	a.emit(model.NamingEvent{
		Type:               eventType,
		Timestamp:          time.Now(),
		CenterInstanceName: a.centerName,
		NamespaceID:        inst.NamespaceID,
		GroupName:          inst.GroupName,
		ServiceName:        inst.ServiceName,
		Service:            svc,
		ChangedNode:        inst,
	})
	a.pruneEmptyAfterLastNode(inst.NamespaceID, inst.GroupName, inst.ServiceName, inst.TenantID, true)
	return nil
}

// dropLocalNode 只改本机 Cache 并推给本机订阅者，不写库、不走 NamingHook。
// 供未接流的副本清理过期视图；权威删除仍由当前 Owner 的 removeNode 广播。
func (a *App) dropLocalNode(inst *model.Node, eventType string) {
	if inst == nil {
		return
	}
	a.cache.RemoveNode(inst.NodeID)
	a.lastPersistBeat.Delete(inst.NodeID)
	a.untrackConnection(inst.ConnectionID, inst.NodeID)
	a.pruneEmptyAfterLastNode(inst.NamespaceID, inst.GroupName, inst.ServiceName, inst.TenantID, false)
	ev := model.NamingEvent{
		Type:               eventType,
		Timestamp:          time.Now(),
		CenterInstanceName: a.centerName,
		Environment:        a.environment(),
		NamespaceID:        inst.NamespaceID,
		GroupName:          inst.GroupName,
		ServiceName:        inst.ServiceName,
		ChangedNode:        inst,
	}
	if inst := ev.ChangedNode; inst != nil && inst.ConnectionID != "" {
		cp := *inst
		cp.ConnectionID = ""
		ev.ChangedNode = &cp
	}
	svc, _ := a.cache.GetService(ev.NamespaceID, ev.GroupName, ev.ServiceName)
	ev.Service = svc
	a.subs.publish(ev)
}

// mergeStoredBeat 用库行刷新心跳，保留本机已有的 Owner / ConnectionID，避免库表没有 Owner 字段把认领抹掉。
func (a *App) mergeStoredBeat(loaded *model.Node) {
	if loaded == nil {
		return
	}
	if current, ok := a.cache.GetNode(loaded.NodeID); ok {
		if current.OwnerGatewayID != "" {
			loaded.OwnerGatewayID = current.OwnerGatewayID
		}
		if current.ConnectionID != "" {
			loaded.ConnectionID = current.ConnectionID
		}
	}
	a.cache.PutNode(loaded)
}

// trackConnection 记录连接与临时节点的绑定，供 OnClientLost 清理。同一 nodeID 不重复追加。
func (a *App) trackConnection(connectionID, nodeID string) {
	if connectionID == "" || nodeID == "" {
		return
	}
	raw, _ := a.connIndex.Load(connectionID)
	ids, _ := raw.([]string)
	for _, id := range ids {
		if id == nodeID {
			return
		}
	}
	copied := append(append([]string{}, ids...), nodeID)
	a.connIndex.Store(connectionID, copied)
}

// untrackConnection 在主动注销或心跳驱逐后从连接索引去掉该节点，避免长连接反复注册撑爆切片。
func (a *App) untrackConnection(connectionID, nodeID string) {
	if connectionID == "" || nodeID == "" {
		return
	}
	raw, ok := a.connIndex.Load(connectionID)
	if !ok {
		return
	}
	ids, _ := raw.([]string)
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if id != nodeID {
			out = append(out, id)
		}
	}
	if len(out) == 0 {
		a.connIndex.Delete(connectionID)
		return
	}
	a.connIndex.Store(connectionID, out)
}

// checkServiceQuota 在新建服务时校验命名空间 serviceQuota，0 表示不限制。
func (a *App) checkServiceQuota(cc contract.CallContext) error {
	ns, ok := a.cache.GetNamespace(cc.NamespaceID)
	if !ok || ns.ServiceQuota <= 0 {
		return nil
	}
	if a.cache.CountServicesInNamespace(cc.NamespaceID) >= ns.ServiceQuota {
		return contract.ErrQuotaExceeded
	}
	return nil
}

// BoundNodeCount 返回连接上登记的业务节点数，供断连告警。
func (a *App) BoundNodeCount(connectionID string) int {
	raw, ok := a.connIndex.Load(connectionID)
	if !ok {
		return 0
	}
	ids, _ := raw.([]string)
	return len(ids)
}

// Cache 返回本实例私有缓存，仅供同实例装配（Overview / Evictor）使用。
func (a *App) Cache() *cache.Cache { return a.cache }

// BindClient 实现 contract.Naming。
func (a *App) BindClient(ch chan<- model.NamingEvent, connectionID string) {
	a.subs.bindConnection(ch, connectionID)
}

// String 返回 naming[instanceName]，便于日志。
func (a *App) String() string {
	return fmt.Sprintf("naming[%s]", a.centerName)
}

var _ contract.Naming = (*App)(nil)
