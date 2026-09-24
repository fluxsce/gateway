package naming

import (
	"context"
	"sort"
	"time"

	"gateway/internal/cluster/publish"
	"gateway/internal/servicecenterv3/model"
	"gateway/pkg/logger"
)

// LocalGatewayID 是当前网关进程，与 GetNodeId() 对齐。
// 空表示单进程，Evictor 按心跳扫描全部业务节点。
var LocalGatewayID string

// NamingHook 仅测试用。生产路径在变更时直接 PublishNaming，ApplyRemote 不得触发。
var NamingHook func(model.NamingEvent)

var clusterPub = publish.NewServiceCenterEventPublisher()

func replicateNaming(ev model.NamingEvent) {
	if NamingHook != nil {
		NamingHook(ev)
		return
	}
	if err := clusterPub.PublishNaming(context.Background(), ev); err != nil {
		logger.Warn("发布服务中心命名集群事件失败", "error", err)
	}
}

// emit 推给本机订阅者，并直接发布集群事件。Environment 从中心实例补齐。
func (a *App) emit(ev model.NamingEvent) {
	if ev.CenterInstanceName == "" {
		ev.CenterInstanceName = a.centerName
	}
	if ev.Environment == "" {
		ev.Environment = a.environment()
	}
	if ev.Timestamp.IsZero() {
		ev.Timestamp = time.Now()
	}
	if inst := ev.ChangedNode; inst != nil && inst.ConnectionID != "" {
		cp := *inst
		cp.ConnectionID = ""
		ev.ChangedNode = &cp
	}
	a.subs.publish(ev)
	hookEv := ev
	if ev.Service != nil {
		cp := *ev.Service
		cp.Nodes = nil
		hookEv.Service = &cp
	}
	replicateNaming(hookEv)
}

func (a *App) environment() string {
	if a.cfg != nil {
		return a.cfg.Environment
	}
	return ""
}

// ApplyRemote 应用其它副本的命名变更：只改 Cache 并推给本机订阅者，不写库、不回复制。发起方已经写过库。
// 删除：本节点已接管则忽略。写入：其它 Owner 的包只有时间更新才接管。
func (a *App) ApplyRemote(ev model.NamingEvent) {
	if LocalGatewayID != "" {
		a.markRemoteActivity()
	}
	switch ev.Type {
	case model.EventServiceAdded, model.EventServiceUpdated:
		if ev.Service != nil {
			a.cache.SetService(ev.Service)
		}
	case model.EventServiceDeleted:
		if svc, ok := a.cache.GetService(ev.NamespaceID, ev.GroupName, ev.ServiceName); ok && len(svc.Nodes) > 0 {
			return
		}
		a.cache.DeleteService(ev.NamespaceID, ev.GroupName, ev.ServiceName)
	case model.EventNodeRegistered, model.EventNodeUpdated, model.EventNodeOffline:
		if inst := ev.ChangedNode; inst != nil {
			inst.ConnectionID = ""
			if current, ok := a.cache.GetNode(inst.NodeID); ok && staleRemoteNode(current, inst, ev.Timestamp, false) {
				return
			}
			a.cache.PutNode(inst)
			if inst.OwnerGatewayID != "" {
				a.noteOwner(inst.OwnerGatewayID)
			}
		}
	case model.EventNodeDeregistered, model.EventNodeEvicted:
		if inst := ev.ChangedNode; inst != nil {
			current, ok := a.cache.GetNode(inst.NodeID)
			if ok && staleRemoteNode(current, inst, ev.Timestamp, true) {
				return
			}
			if !ok {
				return
			}
			a.cache.RemoveNode(inst.NodeID)
			a.pruneEmptyAfterLastNode(inst.NamespaceID, inst.GroupName, inst.ServiceName, inst.TenantID, false)
		}
	}
	a.subs.publish(ev)
}

// nodeRevision 用注册或最近心跳中较新的时间作为副本比较键。
func nodeRevision(inst *model.Node) time.Time {
	if inst == nil {
		return time.Time{}
	}
	t := inst.RegisterTime
	if inst.LastBeatTime.After(t) {
		return inst.LastBeatTime
	}
	return t
}

// staleRemoteNode 判断对端事件是否过期，避免故障转移后旧 Owner 的包打掉新数据。
// 空 Owner 表示上一任放手，不能回退已经更新的认领；对已有 Owner 的删除必须来自同一 Owner。
func staleRemoteNode(current, remote *model.Node, evTime time.Time, isDelete bool) bool {
	if current == nil || remote == nil {
		return false
	}
	localRev := nodeRevision(current)
	remoteRev := evTime
	if r := nodeRevision(remote); r.After(remoteRev) {
		remoteRev = r
	}
	if LocalGatewayID != "" && current.OwnerGatewayID == LocalGatewayID && remote.OwnerGatewayID != LocalGatewayID {
		return true
	}
	if current.OwnerGatewayID != "" && remote.OwnerGatewayID != current.OwnerGatewayID {
		if isDelete {
			return true
		}
		return !remoteRev.After(localRev)
	}
	if isDelete && !evTime.IsZero() && !localRev.IsZero() && !evTime.After(localRev) {
		return true
	}
	return false
}

func (a *App) noteOwner(ownerGatewayID string) {
	if ownerGatewayID == "" {
		return
	}
	a.ownerSeen.Store(ownerGatewayID, time.Now())
}

// ReplicaGatewayIDs 返回当前已知的网关进程：本机、近期事件、以及业务节点的 OwnerGatewayID。
func (a *App) ReplicaGatewayIDs() []string {
	seen := map[string]struct{}{}
	if LocalGatewayID != "" {
		seen[LocalGatewayID] = struct{}{}
	}
	a.ownerSeen.Range(func(k, v any) bool {
		id, _ := k.(string)
		if id != "" {
			seen[id] = struct{}{}
		}
		return true
	})
	for _, inst := range a.cache.AllNodes() {
		if inst != nil && inst.OwnerGatewayID != "" {
			seen[inst.OwnerGatewayID] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for id := range seen {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

// IsViewWarm 报告命名视图是否已从库灌齐，空名单可以当真。
func (a *App) IsViewWarm() bool {
	return a != nil && a.viewWarm.Load()
}

// MarkViewWarm 在库加载完成后调用。
func (a *App) MarkViewWarm() {
	if a == nil {
		return
	}
	a.viewWarm.Store(true)
}

// ResetViewWarm 测试用：把视图标成未追齐。
func (a *App) ResetViewWarm() {
	if a == nil {
		return
	}
	a.viewWarm.Store(false)
}

func (a *App) markRemoteActivity() {
	a.MarkViewWarm()
}

// ownedLocally 报告本机是否是这颗节点的当前 Owner，因而可以删库并广播 Evict。
// 空 Owner 不是「谁都可以广播删除」：任意网关都能认领，但没接流的副本只按库清本地视图。
func (a *App) ownedLocally(inst *model.Node) bool {
	if inst == nil {
		return false
	}
	if LocalGatewayID == "" {
		return true
	}
	return inst.OwnerGatewayID == LocalGatewayID
}

// snapshotServices 把当前 Cache 快照投给刚订阅的通道，避免客户端只靠 Discover 撞上复制窗口。
func (a *App) snapshotServices(namespaceID, group string, serviceNames []string, ch chan<- model.NamingEvent) {
	if ch == nil {
		return
	}
	a.refreshLiveView(context.Background())
	now := time.Now()
	push := func(svc *model.Service) {
		if svc == nil {
			return
		}
		if len(svc.Nodes) == 0 {
			if !a.IsViewWarm() {
				return
			}
			deliverNaming(ch, model.NamingEvent{
				Type:               model.EventServiceAdded,
				Timestamp:          now,
				CenterInstanceName: a.centerName,
				Environment:        a.environment(),
				NamespaceID:        svc.NamespaceID,
				GroupName:          svc.GroupName,
				ServiceName:        svc.ServiceName,
				Service:            svc,
			})
			return
		}
		for _, inst := range svc.Nodes {
			deliverNaming(ch, model.NamingEvent{
				Type:               model.EventNodeRegistered,
				Timestamp:          now,
				CenterInstanceName: a.centerName,
				Environment:        a.environment(),
				NamespaceID:        inst.NamespaceID,
				GroupName:          inst.GroupName,
				ServiceName:        inst.ServiceName,
				Service:            svc,
				ChangedNode:        inst,
			})
		}
	}
	if len(serviceNames) == 0 {
		for _, svc := range a.cache.ListServices(namespaceID, group) {
			push(svc)
		}
		return
	}
	for _, name := range serviceNames {
		if name == "" {
			continue
		}
		svc, ok := a.cache.GetService(namespaceID, group, name)
		if ok {
			push(svc)
		}
	}
}
