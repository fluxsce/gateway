package naming

import (
	"context"
	"time"

	"gateway/internal/servicecenterv3/contract"
	"gateway/internal/servicecenterv3/infra/live"
	"gateway/internal/servicecenterv3/model"
	"gateway/pkg/logger"
)

const (
	liveOpTimeout      = 2 * time.Second
	liveListTimeout    = 8 * time.Second
	liveCacheTrust     = 3 * time.Second
	liveDeleteAttempts = 3
	liveFailThreshold  = 3
	defaultLiveSync    = 2 * time.Second
)

// SetLive 挂上集群活视图。nil 表示只走本机 Cache + 库。
func (a *App) SetLive(s *live.Store) {
	if a == nil {
		return
	}
	a.live = s
	if s == nil {
		a.liveDown.Store(false)
		a.liveFails.Store(0)
	}
}

func (a *App) hasLive() bool {
	return a != nil && a.live != nil
}

func (a *App) liveReadable() bool {
	return a.hasLive() && !a.liveDown.Load()
}

func (a *App) noteLiveOK() {
	a.liveFails.Store(0)
	if a.liveDown.Swap(false) {
		logger.Info("服务注册发现活视图已恢复")
	}
}

func (a *App) noteLiveErr(op, nodeID string, err error) {
	logger.Warn("注册发现活视图失败", "op", op, "nodeId", nodeID, "error", err)
	if a.liveFails.Add(1) >= liveFailThreshold {
		if !a.liveDown.Swap(true) {
			logger.Warn("服务注册发现活视图降级，改走本机 Cache 与集群事件")
		}
	}
}

func (a *App) liveTTL(inst *model.Node) time.Duration {
	timeout := 15 * time.Second
	if a.cfg != nil && a.cfg.HealthCheckTimeout > 0 {
		timeout = time.Duration(a.cfg.HealthCheckTimeout) * time.Second
	}
	if inst != nil && !inst.Ephemeral {
		return 24 * time.Hour
	}
	return 2 * timeout
}

func (a *App) livePut(ctx context.Context, inst *model.Node) {
	if !a.hasLive() || inst == nil {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, liveOpTimeout)
	defer cancel()
	if err := a.live.Put(ctx, inst, a.liveTTL(inst)); err != nil {
		a.noteLiveErr("put", inst.NodeID, err)
		return
	}
	a.noteLiveOK()
}

// liveWriteOnRegister 首次写入整包 Put；本机已有且身份未变只续 TTL。
func (a *App) liveWriteOnRegister(ctx context.Context, inst, prev *model.Node) {
	if prev == nil || liveRegisterNeedsPut(prev, inst) {
		a.livePut(ctx, inst)
		return
	}
	a.liveTouch(ctx, inst)
}

func liveRegisterNeedsPut(prev, inst *model.Node) bool {
	if prev == nil || inst == nil {
		return true
	}
	return prev.OwnerGatewayID != inst.OwnerGatewayID ||
		prev.ConnectionID != inst.ConnectionID ||
		prev.Status != inst.Status ||
		prev.IP != inst.IP ||
		prev.Port != inst.Port ||
		prev.NamespaceID != inst.NamespaceID ||
		prev.GroupName != inst.GroupName ||
		prev.ServiceName != inst.ServiceName
}

// liveTouch 心跳只续 TTL。键丢了才回退整包 Put。持久节点不续期。
func (a *App) liveTouch(ctx context.Context, inst *model.Node) {
	if !a.hasLive() || inst == nil {
		return
	}
	if !inst.Ephemeral {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, liveOpTimeout)
	defer cancel()
	ok, err := a.live.Touch(ctx, inst.TenantID, inst.NodeID, a.liveTTL(inst))
	if err != nil {
		a.noteLiveErr("touch", inst.NodeID, err)
		return
	}
	if !ok {
		a.livePut(ctx, inst)
		return
	}
	a.noteLiveOK()
}

func (a *App) liveViewFresh() bool {
	if a == nil {
		return false
	}
	ts := a.lastLiveRefresh.Load()
	if ts == 0 {
		return false
	}
	return time.Since(time.Unix(0, ts)) < liveCacheTrust
}

func (a *App) markLiveRefreshed() {
	a.lastLiveRefresh.Store(time.Now().UnixNano())
}

// ingestLiveNode 灌活视图。本机 Owner 的心跳/连接以 Cache 为准（心跳只续 Redis TTL，JSON 可能旧）。
// 副本也不得用更旧的 LastBeatTime 回退。Status / 地址仍取活视图，保证管理端下线立刻可见。
func (a *App) ingestLiveNode(n *model.Node) *model.Node {
	n = a.adoptLiveNode(n)
	if n == nil {
		return nil
	}
	prev, ok := a.cache.GetNode(n.NodeID)
	if !ok {
		return n
	}
	if !prev.LastBeatTime.IsZero() && prev.LastBeatTime.After(n.LastBeatTime) {
		n.LastBeatTime = prev.LastBeatTime
	}
	if !a.ownedLocally(prev) {
		return n
	}
	n.ConnectionID = prev.ConnectionID
	n.OwnerGatewayID = prev.OwnerGatewayID
	n.HealthyStatus = prev.HealthyStatus
	return n
}

func (a *App) liveDelete(ctx context.Context, inst *model.Node) error {
	if !a.hasLive() || inst == nil {
		return nil
	}
	var last error
	for i := 0; i < liveDeleteAttempts; i++ {
		try, cancel := context.WithTimeout(ctx, liveOpTimeout)
		err := a.live.Delete(try, inst)
		cancel()
		if err == nil {
			a.noteLiveOK()
			return nil
		}
		last = err
	}
	a.noteLiveErr("delete", inst.NodeID, last)
	return contract.ErrLiveUnavailable
}

func (a *App) liveGet(ctx context.Context, inst *model.Node) *model.Node {
	if !a.hasLive() || inst == nil || inst.NodeID == "" {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, liveOpTimeout)
	defer cancel()
	got, err := a.live.Get(ctx, inst.TenantID, inst.NodeID)
	if err != nil {
		a.noteLiveErr("get", inst.NodeID, err)
		return nil
	}
	a.noteLiveOK()
	return got
}

func (a *App) adoptLiveNode(n *model.Node) *model.Node {
	if n == nil {
		return nil
	}
	if LocalGatewayID != "" && n.OwnerGatewayID != LocalGatewayID {
		n.ConnectionID = ""
	}
	return n
}

// overlayLiveService 用活视图覆盖该服务的节点名单。活视图失败或降级时原样返回。
func (a *App) overlayLiveService(ctx context.Context, tenantID string, svc *model.Service) *model.Service {
	if svc == nil || !a.liveReadable() {
		return svc
	}
	if tenantID == "" {
		tenantID = svc.TenantID
	}
	ctx, cancel := context.WithTimeout(ctx, liveListTimeout)
	defer cancel()
	liveNodes, err := a.live.ListByService(ctx, tenantID, svc.NamespaceID, svc.GroupName, svc.ServiceName)
	if err != nil {
		a.noteLiveErr("list", svc.ServiceName, err)
		return svc
	}
	a.noteLiveOK()
	seen := make(map[string]struct{}, len(liveNodes))
	for _, n := range liveNodes {
		if n = a.ingestLiveNode(n); n == nil {
			continue
		}
		a.cache.PutNode(n)
		seen[n.NodeID] = struct{}{}
	}
	refreshed, ok := a.cache.GetService(svc.NamespaceID, svc.GroupName, svc.ServiceName)
	if !ok {
		return svc
	}
	kept := make([]*model.Node, 0, len(refreshed.Nodes))
	for _, n := range refreshed.Nodes {
		if n == nil {
			continue
		}
		if _, ok := seen[n.NodeID]; ok || a.ownedLocally(n) || !n.Ephemeral {
			kept = append(kept, n)
		}
	}
	refreshed.Nodes = kept
	return refreshed
}

func (a *App) serviceFromLive(ctx context.Context, tenantID, namespaceID, group, serviceName string) *model.Service {
	if !a.liveReadable() || serviceName == "" {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, liveListTimeout)
	defer cancel()
	nodes, err := a.live.ListByService(ctx, tenantID, namespaceID, group, serviceName)
	if err != nil {
		a.noteLiveErr("list", serviceName, err)
		return nil
	}
	if len(nodes) == 0 {
		return nil
	}
	a.noteLiveOK()
	svc := &model.Service{
		TenantID:           tenantID,
		CenterInstanceName: a.centerName,
		NamespaceID:        namespaceID,
		GroupName:          group,
		ServiceName:        serviceName,
		AutoCreated:        true,
	}
	a.cache.SetService(svc)
	for _, n := range nodes {
		if n = a.ingestLiveNode(n); n == nil {
			continue
		}
		a.cache.PutNode(n)
	}
	if got, ok := a.cache.GetService(namespaceID, group, serviceName); ok {
		return got
	}
	return svc
}

// refreshLiveView 把本中心活视图灌进 Cache，并对本机订阅者补推新增/身份变更（不走 NamingHook）。
func (a *App) refreshLiveView(ctx context.Context) {
	if !a.hasLive() {
		a.pruneEmptyAutoServices()
		return
	}
	ctx, cancel := context.WithTimeout(ctx, liveListTimeout)
	defer cancel()
	nodes, err := a.live.ListByCenter(ctx, a.centerName)
	if err != nil {
		a.noteLiveErr("refresh", a.centerName, err)
		return
	}
	a.noteLiveOK()
	a.markLiveRefreshed()
	seen := make(map[string]struct{}, len(nodes))
	for _, n := range nodes {
		if n = a.ingestLiveNode(n); n == nil {
			continue
		}
		if n.CenterInstanceName != "" && n.CenterInstanceName != a.centerName {
			continue
		}
		prev, existed := a.cache.GetNode(n.NodeID)
		a.cache.PutNode(n)
		if n.OwnerGatewayID != "" {
			a.noteOwner(n.OwnerGatewayID)
		}
		seen[n.NodeID] = struct{}{}
		if !existed {
			a.publishLocalLive(model.EventNodeRegistered, n)
		} else if liveIdentityChanged(prev, n) {
			a.publishLocalLive(model.EventNodeUpdated, n)
		}
	}
	for _, local := range a.cache.AllNodes() {
		if local == nil || !local.Ephemeral {
			continue
		}
		if _, ok := seen[local.NodeID]; ok || a.ownedLocally(local) || persistentBeatFresh(local.LastBeatTime, a.persistentHealthGrace()) {
			continue
		}
		a.dropLocalNode(local, model.EventNodeEvicted)
	}
	a.pruneEmptyAutoServices()
}

func (a *App) publishLocalLive(eventType string, inst *model.Node) {
	if inst == nil || a.subs == nil {
		return
	}
	svc, _ := a.cache.GetService(inst.NamespaceID, inst.GroupName, inst.ServiceName)
	a.subs.publish(model.NamingEvent{
		Type:               eventType,
		Timestamp:          time.Now(),
		CenterInstanceName: a.centerName,
		Environment:        a.environment(),
		NamespaceID:        inst.NamespaceID,
		GroupName:          inst.GroupName,
		ServiceName:        inst.ServiceName,
		Service:            svc,
		ChangedNode:        inst,
	})
}

func (a *App) mergeLiveNode(n *model.Node) {
	if n = a.adoptLiveNode(n); n == nil {
		return
	}
	if current, ok := a.cache.GetNode(n.NodeID); ok {
		if current.OwnerGatewayID == LocalGatewayID && LocalGatewayID != "" {
			n.OwnerGatewayID = current.OwnerGatewayID
			n.ConnectionID = current.ConnectionID
		}
	}
	a.cache.PutNode(n)
}

// StartLiveSync 周期对账活视图并补推本机订阅，弥补集群事件延迟。
func (a *App) StartLiveSync(interval time.Duration) {
	if a == nil || !a.hasLive() {
		return
	}
	if interval <= 0 {
		interval = defaultLiveSync
	}
	a.StopLiveSync()
	a.liveSyncStop = make(chan struct{})
	a.liveSyncWG.Add(1)
	go func() {
		defer a.liveSyncWG.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		a.refreshLiveView(context.Background())
		for {
			select {
			case <-ticker.C:
				a.refreshLiveView(context.Background())
			case <-a.liveSyncStop:
				return
			}
		}
	}()
}

// StopLiveSync 停止活视图对账。
func (a *App) StopLiveSync() {
	if a == nil || a.liveSyncStop == nil {
		return
	}
	close(a.liveSyncStop)
	a.liveSyncWG.Wait()
	a.liveSyncStop = nil
}
