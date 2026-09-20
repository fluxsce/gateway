package naming

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"gateway/internal/servicecenterv3/contract"
	"gateway/internal/servicecenterv3/infra/alert"
	"gateway/internal/servicecenterv3/model"
	"gateway/pkg/logger"
)

// Evictor 按中心实例扫描心跳超时：临时节点剔除，持久节点标为 Unhealthy。
type Evictor struct {
	app      *App
	interval time.Duration
	timeout  time.Duration
	stopCh   chan struct{}
	running  atomic.Bool
	wg       sync.WaitGroup
}

// NewEvictor 构造剔除器。interval/timeout<=0 时使用 30s/15s 默认值。
func NewEvictor(app *App, interval, timeout time.Duration) *Evictor {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	return &Evictor{
		app:      app,
		interval: interval,
		timeout:  timeout,
		stopCh:   make(chan struct{}),
	}
}

// Start 启动后台扫描，重复调用是无操作。
func (e *Evictor) Start() {
	if e.running.Swap(true) {
		return
	}
	e.wg.Add(1)
	go e.loop()
}

// Stop 停止扫描并等待当前轮次结束，可再次 Start。
func (e *Evictor) Stop() {
	if !e.running.Swap(false) {
		return
	}
	close(e.stopCh)
	e.wg.Wait()
	e.stopCh = make(chan struct{})
}

// loop 按 interval 周期调用 sweep，直到 Stop。
func (e *Evictor) loop() {
	defer e.wg.Done()
	ticker := time.NewTicker(e.interval)
	defer ticker.Stop()
	e.sweep()
	for {
		select {
		case <-ticker.C:
			e.sweep()
		case <-e.stopCh:
			return
		}
	}
}

// sweep 扫描超时节点：临时节点驱逐，持久节点标为 Unhealthy。
func (e *Evictor) sweep() {
	now := time.Now()
	ctx := context.Background()
	if !e.app.liveViewFresh() {
		e.app.refreshLiveView(ctx)
	}
	var evicted []alert.NodeInfo
	unhealthy := 0
	for _, inst := range e.app.cache.AllNodes() {
		if !e.app.ownedLocally(inst) {
			if inst.Ephemeral {
				e.sweepReplicaEphemeral(ctx, inst)
			} else {
				e.sweepReplicaPersistent(inst)
			}
			continue
		}
		current, ok := e.app.cache.GetNode(inst.NodeID)
		if !ok {
			continue
		}
		if e.beatStillFresh(ctx, current, now) {
			continue
		}
		cc := contract.CallContext{
			TenantID:           current.TenantID,
			CenterInstanceName: e.app.centerName,
			NamespaceID:        current.NamespaceID,
		}
		if current.Ephemeral {
			if err := e.app.removeNode(ctx, cc, current, model.EventNodeEvicted); err != nil {
				continue
			}
			if _, still := e.app.cache.GetNode(current.NodeID); still {
				continue
			}
			logger.Info("驱逐心跳超时的临时节点", "nodeId", current.NodeID, "service", current.ServiceName)
			evicted = append(evicted, alert.NodeInfo{
				NodeID:      current.NodeID,
				ServiceName: current.ServiceName,
				NamespaceID: current.NamespaceID,
				GroupName:   current.GroupName,
				IP:          current.IP,
				Port:        current.Port,
			})
			continue
		}
		inst = current
		inst.HealthyStatus = model.Unhealthy
		_ = e.app.UpdateNode(ctx, cc, inst)
		unhealthy++
	}
	if len(evicted) > 0 {
		alert.NodeEviction(e.app.cfg, evicted)
	}
	if unhealthy > 0 {
		alert.HealthCheckFail(e.app.cfg, unhealthy, "")
	}
}

// sweepReplicaEphemeral 未接流的副本清理过期临时节点：可按库条件删行，只改本机 Cache，不广播 Evict。
func (e *Evictor) sweepReplicaEphemeral(ctx context.Context, inst *model.Node) {
	if inst == nil {
		return
	}
	if e.app.liveViewFresh() {
		return
	}
	if liveNode := e.app.liveGet(ctx, inst); liveNode != nil {
		e.app.mergeLiveNode(liveNode)
		return
	}
	grace := e.app.persistentHealthGrace()
	if e.app.store == nil || e.app.store.Node == nil {
		if inst.OwnerGatewayID != "" && inst.OwnerGatewayID != LocalGatewayID {
			return
		}
		if !persistentBeatFresh(inst.LastBeatTime, grace) {
			e.app.dropLocalNode(inst, model.EventNodeEvicted)
		}
		return
	}
	loaded, err := e.app.store.Node.Get(ctx, inst.TenantID, inst.NodeID)
	if err != nil {
		return
	}
	if loaded != nil && persistentBeatFresh(loaded.LastBeatTime, grace) {
		e.app.mergeStoredBeat(loaded)
		return
	}
	if loaded != nil {
		deleted, delErr := e.app.store.Node.DeleteIfBeatNotAfter(ctx, inst.TenantID, inst.NodeID, loaded.LastBeatTime)
		if delErr != nil {
			return
		}
		if !deleted {
			if again, getErr := e.app.store.Node.Get(ctx, inst.TenantID, inst.NodeID); getErr == nil && again != nil {
				e.app.mergeStoredBeat(again)
			}
			return
		}
	}
	e.app.dropLocalNode(inst, model.EventNodeEvicted)
}

// sweepReplicaPersistent 未接流的副本只把过期持久节点标成本机 Unhealthy，不写库、不广播。
func (e *Evictor) sweepReplicaPersistent(inst *model.Node) {
	if inst == nil {
		return
	}
	if persistentBeatFresh(inst.LastBeatTime, e.timeout) {
		return
	}
	current, ok := e.app.cache.GetNode(inst.NodeID)
	if !ok {
		return
	}
	if current.HealthyStatus == model.Unhealthy {
		return
	}
	current.HealthyStatus = model.Unhealthy
	e.app.cache.PutNode(current)
}

// beatStillFresh 以 Cache 现值为准，避免 AllNodes 拷贝过期后误驱逐；零心跳先对一下活视图。
func (e *Evictor) beatStillFresh(ctx context.Context, inst *model.Node, now time.Time) bool {
	if inst == nil {
		return false
	}
	if !inst.LastBeatTime.IsZero() && now.Sub(inst.LastBeatTime) <= e.timeout {
		return true
	}
	if !inst.LastBeatTime.IsZero() {
		return false
	}
	liveNode := e.app.liveGet(ctx, inst)
	if liveNode == nil {
		return false
	}
	e.app.mergeLiveNode(liveNode)
	return !liveNode.LastBeatTime.IsZero() && now.Sub(liveNode.LastBeatTime) <= e.timeout
}
