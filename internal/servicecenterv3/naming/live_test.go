package naming

import (
	"context"
	"testing"
	"time"

	"gateway/internal/servicecenterv3/contract"
	"gateway/internal/servicecenterv3/infra/cache"
	"gateway/internal/servicecenterv3/infra/live"
	"gateway/internal/servicecenterv3/model"
	"gateway/pkg/cache/memory"
)

func newSharedLive(t *testing.T) *live.Store {
	t.Helper()
	c, err := memory.NewMemoryCache(&memory.MemoryConfig{
		CleanupInterval:   time.Hour,
		EnableLazyCleanup: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return live.New(c)
}

func newLiveApp(t *testing.T, name, gatewayID string, view *live.Store) *App {
	t.Helper()
	app := New(name, nil, cache.New(name))
	app.SetLive(view)
	app.cache.SetNamespace(&model.Namespace{
		NamespaceID:        "ns",
		CenterInstanceName: name,
		Active:             true,
	})
	t.Cleanup(func() {
		if LocalGatewayID == gatewayID {
			LocalGatewayID = ""
		}
	})
	return app
}

func TestDiscoverSeesNodeRegisteredOnPeer(t *testing.T) {
	resetSyncHooks(t)
	view := newSharedLive(t)
	a := newLiveApp(t, "c", "gw1", view)
	b := newLiveApp(t, "c", "gw2", view)
	LocalGatewayID = "gw1"
	cc := contract.CallContext{TenantID: "t", CenterInstanceName: "c", NamespaceID: "ns"}
	if err := a.RegisterNode(context.Background(), cc, &model.Node{
		NodeID: "n1", GroupName: "g", ServiceName: "order",
		IP: "10.0.0.1", Port: 80, Ephemeral: true,
	}); err != nil {
		t.Fatal(err)
	}
	LocalGatewayID = "gw2"
	svc, err := b.GetService(context.Background(), cc, "g", "order")
	if err != nil {
		t.Fatal(err)
	}
	if len(svc.Nodes) != 1 || svc.Nodes[0].NodeID != "n1" || svc.Nodes[0].IP != "10.0.0.1" {
		t.Fatalf("peer discover want n1, got %+v", svc.Nodes)
	}
	if svc.Nodes[0].ConnectionID != "" {
		t.Fatal("replica must not copy peer ConnectionID")
	}
}

func TestHeartbeatTouchesTTLWithoutRewrite(t *testing.T) {
	resetSyncHooks(t)
	view := newSharedLive(t)
	app := newLiveApp(t, "c", "gw1", view)
	LocalGatewayID = "gw1"
	cc := contract.CallContext{TenantID: "t", CenterInstanceName: "c", NamespaceID: "ns"}
	if err := app.RegisterNode(context.Background(), cc, &model.Node{
		NodeID: "n1", GroupName: "g", ServiceName: "order",
		IP: "10.0.0.1", Port: 80, Ephemeral: true,
	}); err != nil {
		t.Fatal(err)
	}
	first, err := view.Get(context.Background(), "t", "n1")
	if err != nil || first == nil {
		t.Fatal(err)
	}
	app.beatPersistEvery = time.Hour
	app.lastPersistBeat.Store("n1", time.Now())
	time.Sleep(5 * time.Millisecond)
	if err := app.Heartbeat(context.Background(), cc, "n1", nil); err != nil {
		t.Fatal(err)
	}
	second, err := view.Get(context.Background(), "t", "n1")
	if err != nil || second == nil {
		t.Fatal("touch must keep live key")
	}
	if !second.LastBeatTime.Equal(first.LastBeatTime) {
		t.Fatalf("steady heartbeat must not rewrite live JSON, first=%v second=%v", first.LastBeatTime, second.LastBeatTime)
	}
}

func TestHeartbeatRefreshesLiveView(t *testing.T) {
	resetSyncHooks(t)
	view := newSharedLive(t)
	app := newLiveApp(t, "c", "gw1", view)
	LocalGatewayID = "gw1"
	cc := contract.CallContext{TenantID: "t", CenterInstanceName: "c", NamespaceID: "ns"}
	inst := &model.Node{
		NodeID: "n1", GroupName: "g", ServiceName: "order",
		IP: "10.0.0.1", Port: 80, Ephemeral: true,
	}
	if err := app.RegisterNode(context.Background(), cc, inst); err != nil {
		t.Fatal(err)
	}
	first, err := view.Get(context.Background(), "t", "n1")
	if err != nil || first == nil {
		t.Fatalf("live after register %+v err=%v", first, err)
	}
	time.Sleep(5 * time.Millisecond)
	if err := app.Heartbeat(context.Background(), cc, "n1", nil); err != nil {
		t.Fatal(err)
	}
	second, err := view.Get(context.Background(), "t", "n1")
	if err != nil || second == nil {
		t.Fatal(err)
	}
	if !second.LastBeatTime.After(first.LastBeatTime) {
		t.Fatalf("heartbeat should refresh live LastBeatTime first=%v second=%v", first.LastBeatTime, second.LastBeatTime)
	}
}

func TestHeartbeatReclaimsFromLiveTouchesOnly(t *testing.T) {
	resetSyncHooks(t)
	view := newSharedLive(t)
	a := newLiveApp(t, "c", "gw1", view)
	LocalGatewayID = "gw1"
	cc := contract.CallContext{TenantID: "t", CenterInstanceName: "c", NamespaceID: "ns"}
	if err := a.RegisterNode(context.Background(), cc, &model.Node{
		NodeID: "n1", GroupName: "g", ServiceName: "order",
		IP: "10.0.0.1", Port: 80, Ephemeral: true,
	}); err != nil {
		t.Fatal(err)
	}
	first, err := view.Get(context.Background(), "t", "n1")
	if err != nil || first == nil {
		t.Fatal(err)
	}
	b := newLiveApp(t, "c", "gw1", view)
	LocalGatewayID = "gw1"
	b.beatPersistEvery = time.Hour
	if err := b.Heartbeat(context.Background(), cc, "n1", nil); err != nil {
		t.Fatal(err)
	}
	second, err := view.Get(context.Background(), "t", "n1")
	if err != nil || second == nil {
		t.Fatal(err)
	}
	if !second.LastBeatTime.Equal(first.LastBeatTime) {
		t.Fatalf("same-owner reclaim must Expire only, first=%v second=%v", first.LastBeatTime, second.LastBeatTime)
	}
	if _, ok := b.cache.GetNode("n1"); !ok {
		t.Fatal("reclaim must adopt live node into cache")
	}
}

func TestHeartbeatClaimFromLiveRewritesOwner(t *testing.T) {
	resetSyncHooks(t)
	view := newSharedLive(t)
	a := newLiveApp(t, "c", "gw1", view)
	LocalGatewayID = "gw1"
	cc := contract.CallContext{TenantID: "t", CenterInstanceName: "c", NamespaceID: "ns"}
	if err := a.RegisterNode(context.Background(), cc, &model.Node{
		NodeID: "n1", GroupName: "g", ServiceName: "order",
		IP: "10.0.0.1", Port: 80, Ephemeral: true, ConnectionID: "old",
	}); err != nil {
		t.Fatal(err)
	}
	b := newLiveApp(t, "c", "gw2", view)
	LocalGatewayID = "gw2"
	snap := &model.Service{
		NamespaceID: "ns", GroupName: "g", ServiceName: "order",
		Nodes: []*model.Node{{
			NodeID: "n1", NamespaceID: "ns", GroupName: "g", ServiceName: "order",
			IP: "10.0.0.1", Port: 80, Ephemeral: true, ConnectionID: "new",
		}},
	}
	if err := b.Heartbeat(context.Background(), cc, "n1", snap); err != nil {
		t.Fatal(err)
	}
	got, err := view.Get(context.Background(), "t", "n1")
	if err != nil || got == nil {
		t.Fatal(err)
	}
	if got.OwnerGatewayID != "gw2" || got.ConnectionID != "new" {
		t.Fatalf("claim must rewrite live owner/conn, got %+v", got)
	}
}

func TestDeregisterRemovesPeerDiscover(t *testing.T) {
	resetSyncHooks(t)
	view := newSharedLive(t)
	a := newLiveApp(t, "c", "gw1", view)
	b := newLiveApp(t, "c", "gw2", view)
	LocalGatewayID = "gw1"
	cc := contract.CallContext{TenantID: "t", CenterInstanceName: "c", NamespaceID: "ns"}
	if err := a.RegisterNode(context.Background(), cc, &model.Node{
		NodeID: "n1", GroupName: "g", ServiceName: "order",
		IP: "10.0.0.1", Port: 80, Ephemeral: true,
	}); err != nil {
		t.Fatal(err)
	}
	if err := a.DeregisterNode(context.Background(), cc, "n1"); err != nil {
		t.Fatal(err)
	}
	LocalGatewayID = "gw2"
	_, err := b.GetService(context.Background(), cc, "g", "order")
	if err != contract.ErrServiceNotFound {
		t.Fatalf("peer should not see deregistered implied service, got %v", err)
	}
}

func TestServerDrainKeepsLiveView(t *testing.T) {
	resetSyncHooks(t)
	view := newSharedLive(t)
	a := newLiveApp(t, "c", "gw1", view)
	b := newLiveApp(t, "c", "gw2", view)
	LocalGatewayID = "gw1"
	cc := contract.CallContext{TenantID: "t", CenterInstanceName: "c", NamespaceID: "ns"}
	if err := a.RegisterNode(context.Background(), cc, &model.Node{
		NodeID: "n1", GroupName: "g", ServiceName: "order",
		IP: "10.0.0.1", Port: 80, Ephemeral: true, ConnectionID: "conn-1",
	}); err != nil {
		t.Fatal(err)
	}
	a.OnClientLost(context.Background(), "conn-1", model.DisconnectServerDrain)
	got, err := view.Get(context.Background(), "t", "n1")
	if err != nil || got == nil {
		t.Fatal("drain must keep live node")
	}
	if got.OwnerGatewayID != "" {
		t.Fatalf("drain should clear owner, got %s", got.OwnerGatewayID)
	}
	LocalGatewayID = "gw2"
	svc, err := b.GetService(context.Background(), cc, "g", "order")
	if err != nil || len(svc.Nodes) != 1 {
		t.Fatalf("peer should still discover drained node, err=%v nodes=%v", err, svc)
	}
}

func TestDeregisterServiceOnConnectionKeepsPeerNode(t *testing.T) {
	resetSyncHooks(t)
	view := newSharedLive(t)
	app := newLiveApp(t, "c", "gw1", view)
	LocalGatewayID = "gw1"
	base := contract.CallContext{TenantID: "t", CenterInstanceName: "c", NamespaceID: "ns"}
	if err := app.RegisterNode(context.Background(), base, &model.Node{
		NodeID: "n1", GroupName: "g", ServiceName: "order",
		IP: "10.0.0.1", Port: 80, Ephemeral: true, ConnectionID: "conn-a",
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.RegisterNode(context.Background(), base, &model.Node{
		NodeID: "n2", GroupName: "g", ServiceName: "order",
		IP: "10.0.0.2", Port: 80, Ephemeral: true, ConnectionID: "conn-b",
	}); err != nil {
		t.Fatal(err)
	}
	ccA := base
	ccA.ConnectionID = "conn-a"
	if err := app.DeregisterService(context.Background(), ccA, "g", "order"); err != nil {
		t.Fatal(err)
	}
	if _, ok := app.cache.GetNode("n1"); ok {
		t.Fatal("this connection's node should be gone")
	}
	if _, ok := app.cache.GetNode("n2"); !ok {
		t.Fatal("peer node of the same service must stay")
	}
	got, err := view.Get(context.Background(), "t", "n2")
	if err != nil || got == nil {
		t.Fatal("peer live node must stay")
	}
}

func TestDeregisterServiceAdminRemovesAllNodes(t *testing.T) {
	resetSyncHooks(t)
	view := newSharedLive(t)
	app := newLiveApp(t, "c", "gw1", view)
	LocalGatewayID = "gw1"
	cc := contract.CallContext{TenantID: "t", CenterInstanceName: "c", NamespaceID: "ns"}
	if err := app.RegisterNode(context.Background(), cc, &model.Node{
		NodeID: "n1", GroupName: "g", ServiceName: "order",
		IP: "10.0.0.1", Port: 80, Ephemeral: true, ConnectionID: "conn-a",
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.RegisterNode(context.Background(), cc, &model.Node{
		NodeID: "n2", GroupName: "g", ServiceName: "order",
		IP: "10.0.0.2", Port: 80, Ephemeral: true, ConnectionID: "conn-b",
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.DeregisterService(context.Background(), cc, "g", "order"); err != nil {
		t.Fatal(err)
	}
	if _, ok := app.cache.GetNode("n1"); ok {
		t.Fatal("admin should remove n1")
	}
	if _, ok := app.cache.GetNode("n2"); ok {
		t.Fatal("admin should remove n2")
	}
}

func TestOverlayDropsEphemeralGhostWhenLiveEmpty(t *testing.T) {
	resetSyncHooks(t)
	view := newSharedLive(t)
	app := newLiveApp(t, "c", "gw2", view)
	LocalGatewayID = "gw2"
	app.cache.SetService(&model.Service{TenantID: "t", NamespaceID: "ns", GroupName: "g", ServiceName: "order"})
	app.cache.PutNode(&model.Node{
		NodeID: "n1", TenantID: "t", NamespaceID: "ns", GroupName: "g", ServiceName: "order",
		IP: "10.0.0.1", Port: 80, Ephemeral: true, OwnerGatewayID: "gw1",
		LastBeatTime: time.Now(), Status: model.NodeUP, HealthyStatus: model.Healthy,
	})
	cc := contract.CallContext{TenantID: "t", CenterInstanceName: "c", NamespaceID: "ns"}
	svc, err := app.GetService(context.Background(), cc, "g", "order")
	if err != nil {
		t.Fatal(err)
	}
	if len(svc.Nodes) != 0 {
		t.Fatalf("live empty: replica must not keep ephemeral ghost, got %+v", svc.Nodes)
	}
}

func TestGetServiceOverlaysAfterRecentRefresh(t *testing.T) {
	resetSyncHooks(t)
	view := newSharedLive(t)
	a := newLiveApp(t, "c", "gw1", view)
	b := newLiveApp(t, "c", "gw2", view)
	LocalGatewayID = "gw2"
	b.refreshLiveView(context.Background())
	if !b.liveViewFresh() {
		t.Fatal("refresh should mark view fresh")
	}
	LocalGatewayID = "gw1"
	cc := contract.CallContext{TenantID: "t", CenterInstanceName: "c", NamespaceID: "ns"}
	if err := a.RegisterNode(context.Background(), cc, &model.Node{
		NodeID: "n1", GroupName: "g", ServiceName: "order",
		IP: "10.0.0.1", Port: 80, Ephemeral: true,
	}); err != nil {
		t.Fatal(err)
	}
	LocalGatewayID = "gw2"
	svc, err := b.GetService(context.Background(), cc, "g", "order")
	if err != nil || len(svc.Nodes) != 1 || svc.Nodes[0].NodeID != "n1" {
		t.Fatalf("discover must overlay live even if L1 just refreshed, err=%v nodes=%v", err, svc)
	}
}

func TestAdminDownVisibleOnPeerDiscover(t *testing.T) {
	resetSyncHooks(t)
	view := newSharedLive(t)
	a := newLiveApp(t, "c", "gw1", view)
	b := newLiveApp(t, "c", "gw2", view)
	LocalGatewayID = "gw1"
	cc := contract.CallContext{TenantID: "t", CenterInstanceName: "c", NamespaceID: "ns"}
	if err := a.RegisterNode(context.Background(), cc, &model.Node{
		NodeID: "n1", GroupName: "g", ServiceName: "order",
		IP: "10.0.0.1", Port: 80, Ephemeral: true,
	}); err != nil {
		t.Fatal(err)
	}
	if err := a.UpdateNode(context.Background(), cc, &model.Node{NodeID: "n1", Status: model.NodeDown}); err != nil {
		t.Fatal(err)
	}
	LocalGatewayID = "gw2"
	svc, err := b.GetService(context.Background(), cc, "g", "order")
	if err != nil || len(svc.Nodes) != 1 || svc.Nodes[0].Status != model.NodeDown {
		t.Fatalf("peer discover must see admin DOWN, err=%v nodes=%v", err, svc)
	}
	if svc.Nodes[0].IsHealthy() {
		t.Fatal("DOWN node must not be healthy")
	}
}

func TestAdminDownNotOverwrittenByStaleLiveUP(t *testing.T) {
	resetSyncHooks(t)
	view := newSharedLive(t)
	a := newLiveApp(t, "c", "gw1", view)
	LocalGatewayID = "gw1"
	cc := contract.CallContext{TenantID: "t", CenterInstanceName: "c", NamespaceID: "ns"}
	if err := a.RegisterNode(context.Background(), cc, &model.Node{
		NodeID: "n1", GroupName: "g", ServiceName: "order",
		IP: "10.0.0.1", Port: 80, Ephemeral: true,
	}); err != nil {
		t.Fatal(err)
	}
	if err := a.UpdateNode(context.Background(), cc, &model.Node{NodeID: "n1", Status: model.NodeDown}); err != nil {
		t.Fatal(err)
	}
	stale := &model.Node{
		NodeID: "n1", TenantID: "t", CenterInstanceName: "c", NamespaceID: "ns",
		GroupName: "g", ServiceName: "order", IP: "10.0.0.1", Port: 80,
		Ephemeral: true, Status: model.NodeUP, HealthyStatus: model.Healthy,
		LastBeatTime: time.Now().Add(-time.Minute),
	}
	if err := view.Put(context.Background(), stale, time.Minute); err != nil {
		t.Fatal(err)
	}
	svc, err := a.GetService(context.Background(), cc, "g", "order")
	if err != nil || len(svc.Nodes) != 1 {
		t.Fatalf("err=%v nodes=%v", err, svc)
	}
	if svc.Nodes[0].Status != model.NodeDown {
		t.Fatalf("stale live UP must not cover admin DOWN, got %s", svc.Nodes[0].Status)
	}
	if svc.Nodes[0].IsHealthy() {
		t.Fatal("admin DOWN must stay out of healthy discovery")
	}
	a.MarkViewWarm()
	list, err := a.ListNodes(context.Background(), cc, "g", "order", true)
	if err != nil || len(list) != 0 {
		t.Fatalf("healthy list must be empty after admin DOWN, err=%v n=%d", err, len(list))
	}
}

func TestDeregisterNodeReadsLiveWhenCacheMiss(t *testing.T) {
	resetSyncHooks(t)
	view := newSharedLive(t)
	app := newLiveApp(t, "c", "gw1", view)
	LocalGatewayID = "gw1"
	n := &model.Node{
		NodeID: "n1", TenantID: "t", CenterInstanceName: "c", NamespaceID: "ns",
		GroupName: "g", ServiceName: "order", IP: "10.0.0.1", Port: 80, Ephemeral: true,
	}
	if err := view.Put(context.Background(), n, time.Minute); err != nil {
		t.Fatal(err)
	}
	cc := contract.CallContext{TenantID: "t", CenterInstanceName: "c", NamespaceID: "ns"}
	if err := app.DeregisterNode(context.Background(), cc, "n1"); err != nil {
		t.Fatal(err)
	}
	got, err := view.Get(context.Background(), "t", "n1")
	if err != nil || got != nil {
		t.Fatalf("live node should be deleted, got %+v err=%v", got, err)
	}
}

func TestDeregisterNodeRejectsOtherConnection(t *testing.T) {
	resetSyncHooks(t)
	app := New("c", nil, cache.New("c"))
	app.cache.SetNamespace(&model.Namespace{NamespaceID: "ns", CenterInstanceName: "c", Active: true})
	cc := contract.CallContext{TenantID: "t", CenterInstanceName: "c", NamespaceID: "ns", ConnectionID: "conn-a"}
	if err := app.RegisterNode(context.Background(), cc, &model.Node{
		NodeID: "n1", GroupName: "g", ServiceName: "order",
		IP: "10.0.0.1", Port: 80, Ephemeral: true, ConnectionID: "conn-a",
	}); err != nil {
		t.Fatal(err)
	}
	other := cc
	other.ConnectionID = "conn-b"
	if err := app.DeregisterNode(context.Background(), other, "n1"); err != contract.ErrNodeNotFound {
		t.Fatalf("want NODE_NOT_FOUND, got %v", err)
	}
	if _, ok := app.cache.GetNode("n1"); !ok {
		t.Fatal("other connection must not remove this node")
	}
}

func TestRefreshLiveViewPushesLocalSubscribe(t *testing.T) {
	resetSyncHooks(t)
	view := newSharedLive(t)
	a := newLiveApp(t, "c", "gw1", view)
	b := newLiveApp(t, "c", "gw2", view)
	ch := make(chan model.NamingEvent, 4)
	b.subs.add("ns", "g", []string{"order"}, ch, "w")
	LocalGatewayID = "gw1"
	cc := contract.CallContext{TenantID: "t", CenterInstanceName: "c", NamespaceID: "ns"}
	if err := a.RegisterNode(context.Background(), cc, &model.Node{
		NodeID: "n1", GroupName: "g", ServiceName: "order",
		IP: "10.0.0.1", Port: 80, Ephemeral: true,
	}); err != nil {
		t.Fatal(err)
	}
	LocalGatewayID = "gw2"
	b.refreshLiveView(context.Background())
	select {
	case ev := <-ch:
		if ev.Type != model.EventNodeRegistered || ev.ChangedNode == nil || ev.ChangedNode.NodeID != "n1" {
			t.Fatalf("want local NODE_REGISTERED n1, got %+v", ev)
		}
	case <-time.After(time.Second):
		t.Fatal("replica subscriber should get live reconcile push")
	}
}

func TestRegisterNodeRejectsBadAndTakenID(t *testing.T) {
	resetSyncHooks(t)
	app := New("c", nil, cache.New("c"))
	app.cache.SetNamespace(&model.Namespace{NamespaceID: "ns", CenterInstanceName: "c", Active: true})
	cc := contract.CallContext{TenantID: "t", CenterInstanceName: "c", NamespaceID: "ns"}
	if err := app.RegisterNode(context.Background(), cc, &model.Node{
		NodeID: "bad id!", GroupName: "g", ServiceName: "order",
		IP: "10.0.0.1", Port: 80, Ephemeral: true,
	}); err != contract.ErrInvalidArgument {
		t.Fatalf("want INVALID_ARGUMENT, got %v", err)
	}
	if err := app.RegisterNode(context.Background(), cc, &model.Node{
		NodeID: "n1", GroupName: "g", ServiceName: "order",
		IP: "10.0.0.1", Port: 80, Ephemeral: true, ConnectionID: "c1",
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.RegisterNode(context.Background(), cc, &model.Node{
		NodeID: "n1", GroupName: "g", ServiceName: "pay",
		IP: "10.0.0.2", Port: 81, Ephemeral: true, ConnectionID: "c2",
	}); err != contract.ErrNodeExists {
		t.Fatalf("want NODE_EXISTS, got %v", err)
	}
	if err := app.RegisterNode(context.Background(), cc, &model.Node{
		NodeID: "n1", GroupName: "g", ServiceName: "order",
		IP: "10.0.0.1", Port: 80, Ephemeral: true, ConnectionID: "c3",
	}); err != nil {
		t.Fatal(err)
	}
}

func TestReplicaEvictorKeepsLiveNode(t *testing.T) {
	resetSyncHooks(t)
	view := newSharedLive(t)
	a := newLiveApp(t, "c", "gw1", view)
	b := newLiveApp(t, "c", "gw2", view)
	LocalGatewayID = "gw1"
	cc := contract.CallContext{TenantID: "t", CenterInstanceName: "c", NamespaceID: "ns"}
	if err := a.RegisterNode(context.Background(), cc, &model.Node{
		NodeID: "n1", GroupName: "g", ServiceName: "order",
		IP: "10.0.0.1", Port: 80, Ephemeral: true,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := b.GetService(context.Background(), cc, "g", "order"); err != nil {
		t.Fatal(err)
	}
	b.cache.PutNode(&model.Node{
		NodeID: "n1", TenantID: "t", NamespaceID: "ns", GroupName: "g", ServiceName: "order",
		IP: "10.0.0.1", Port: 80, Ephemeral: true, OwnerGatewayID: "gw1",
		LastBeatTime: time.Now().Add(-time.Hour),
	})
	LocalGatewayID = "gw2"
	ev := NewEvictor(b, time.Hour, 50*time.Millisecond)
	ev.sweep()
	if _, ok := b.cache.GetNode("n1"); !ok {
		t.Fatal("replica must not drop node still present in live view")
	}
}
