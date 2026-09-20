package naming

import (
	"context"
	"testing"
	"time"

	"gateway/internal/servicecenterv3/contract"
	"gateway/internal/servicecenterv3/infra/cache"
	"gateway/internal/servicecenterv3/model"
)

func resetSyncHooks(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		LocalGatewayID = ""
		NamingHook = nil
	})
}

func TestApplyRemoteDoesNotReplicate(t *testing.T) {
	resetSyncHooks(t)
	hooked := 0
	NamingHook = func(model.NamingEvent) { hooked++ }
	app := New("center-1", nil, cache.New("center-1"))
	inst := &model.Node{
		NodeID:         "n1",
		NamespaceID:    "ns",
		GroupName:      "g",
		ServiceName:    "svc",
		IP:             "10.0.0.1",
		Port:           80,
		OwnerGatewayID: "peer",
		Ephemeral:      true,
		Status:         model.NodeUP,
		HealthyStatus:  model.Healthy,
	}
	ch := make(chan model.NamingEvent, 4)
	app.subs.add("ns", "g", []string{"svc"}, ch, "w")
	app.ApplyRemote(model.NamingEvent{
		Type:        model.EventNodeRegistered,
		NamespaceID: "ns",
		GroupName:   "g",
		ServiceName: "svc",
		ChangedNode: inst,
	})
	if hooked != 0 {
		t.Fatalf("ApplyRemote must not call NamingHook, got %d", hooked)
	}
	got, ok := app.cache.GetNode("n1")
	if !ok || got.OwnerGatewayID != "peer" {
		t.Fatalf("expected remote node in cache, got %+v", got)
	}
	select {
	case ev := <-ch:
		if ev.Type != model.EventNodeRegistered {
			t.Fatalf("subscriber got %s", ev.Type)
		}
	default:
		t.Fatal("local subscriber should see remote register")
	}
}

func TestApplyRemoteIgnoresStaleEvictFromOldOwner(t *testing.T) {
	resetSyncHooks(t)
	LocalGatewayID = "gw2"
	app := New("c", nil, cache.New("c"))
	now := time.Now()
	app.cache.PutNode(&model.Node{
		NodeID:         "n1",
		NamespaceID:    "ns",
		GroupName:      "g",
		ServiceName:    "svc",
		IP:             "10.0.0.1",
		Port:           80,
		Ephemeral:      true,
		OwnerGatewayID: "gw2",
		RegisterTime:   now,
		LastBeatTime:   now,
	})
	ch := make(chan model.NamingEvent, 2)
	app.subs.add("ns", "g", []string{"svc"}, ch, "w")
	app.ApplyRemote(model.NamingEvent{
		Type:      model.EventNodeEvicted,
		Timestamp: now.Add(-time.Second),
		ChangedNode: &model.Node{
			NodeID:         "n1",
			NamespaceID:    "ns",
			GroupName:      "g",
			ServiceName:    "svc",
			OwnerGatewayID: "gw1",
			Ephemeral:      true,
		},
	})
	if _, ok := app.cache.GetNode("n1"); !ok {
		t.Fatal("new owner node must survive stale evict from old owner")
	}
	select {
	case ev := <-ch:
		t.Fatalf("stale evict must not push, got %s", ev.Type)
	default:
	}
}

func TestApplyRemoteNewerOwnerTakesOverThenRejectsOldRegister(t *testing.T) {
	resetSyncHooks(t)
	LocalGatewayID = "gw3"
	app := New("c", nil, cache.New("c"))
	old := time.Now().Add(-2 * time.Second)
	app.cache.PutNode(&model.Node{
		NodeID:         "n1",
		NamespaceID:    "ns",
		GroupName:      "g",
		ServiceName:    "svc",
		IP:             "10.0.0.1",
		Port:           80,
		Ephemeral:      true,
		OwnerGatewayID: "gw1",
		RegisterTime:   old,
		LastBeatTime:   old,
	})
	newer := time.Now()
	app.ApplyRemote(model.NamingEvent{
		Type:      model.EventNodeRegistered,
		Timestamp: newer,
		ChangedNode: &model.Node{
			NodeID:         "n1",
			NamespaceID:    "ns",
			GroupName:      "g",
			ServiceName:    "svc",
			IP:             "10.0.0.1",
			Port:           81,
			Ephemeral:      true,
			OwnerGatewayID: "gw2",
			RegisterTime:   newer,
			LastBeatTime:   newer,
		},
	})
	got, _ := app.cache.GetNode("n1")
	if got.OwnerGatewayID != "gw2" || got.Port != 81 {
		t.Fatalf("newer owner should take over, got %+v", got)
	}
	app.ApplyRemote(model.NamingEvent{
		Type:      model.EventNodeRegistered,
		Timestamp: old,
		ChangedNode: &model.Node{
			NodeID:         "n1",
			NamespaceID:    "ns",
			GroupName:      "g",
			ServiceName:    "svc",
			IP:             "10.0.0.1",
			Port:           80,
			Ephemeral:      true,
			OwnerGatewayID: "gw1",
			RegisterTime:   old,
			LastBeatTime:   old,
		},
	})
	got, _ = app.cache.GetNode("n1")
	if got.OwnerGatewayID != "gw2" || got.Port != 81 {
		t.Fatalf("stale register from old owner must not overwrite, got %+v", got)
	}
}

func TestApplyRemoteEmptyOwnerDoesNotOverwriteNewerOwner(t *testing.T) {
	resetSyncHooks(t)
	LocalGatewayID = "gw3"
	app := New("c", nil, cache.New("c"))
	newer := time.Now()
	app.cache.PutNode(&model.Node{
		NodeID:         "n1",
		NamespaceID:    "ns",
		GroupName:      "g",
		ServiceName:    "svc",
		IP:             "10.0.0.1",
		Port:           81,
		Ephemeral:      true,
		OwnerGatewayID: "gw2",
		RegisterTime:   newer,
		LastBeatTime:   newer,
	})
	app.ApplyRemote(model.NamingEvent{
		Type:      model.EventNodeUpdated,
		Timestamp: newer.Add(-time.Second),
		ChangedNode: &model.Node{
			NodeID:         "n1",
			NamespaceID:    "ns",
			GroupName:      "g",
			ServiceName:    "svc",
			IP:             "10.0.0.1",
			Port:           80,
			Ephemeral:      true,
			OwnerGatewayID: "",
			LastBeatTime:   newer.Add(-time.Second),
		},
	})
	got, _ := app.cache.GetNode("n1")
	if got.OwnerGatewayID != "gw2" || got.Port != 81 {
		t.Fatalf("delayed empty-owner release must not overwrite newer owner, got %+v", got)
	}
}

func TestApplyRemoteEmptyOwnerReleaseWhenStillPrevious(t *testing.T) {
	resetSyncHooks(t)
	LocalGatewayID = "gw3"
	app := New("c", nil, cache.New("c"))
	old := time.Now().Add(-2 * time.Second)
	app.cache.PutNode(&model.Node{
		NodeID:         "n1",
		NamespaceID:    "ns",
		GroupName:      "g",
		ServiceName:    "svc",
		IP:             "10.0.0.1",
		Port:           80,
		Ephemeral:      true,
		OwnerGatewayID: "gw1",
		RegisterTime:   old,
		LastBeatTime:   old,
	})
	release := time.Now()
	app.ApplyRemote(model.NamingEvent{
		Type:      model.EventNodeUpdated,
		Timestamp: release,
		ChangedNode: &model.Node{
			NodeID:         "n1",
			NamespaceID:    "ns",
			GroupName:      "g",
			ServiceName:    "svc",
			IP:             "10.0.0.1",
			Port:           80,
			Ephemeral:      true,
			OwnerGatewayID: "",
			LastBeatTime:   release,
		},
	})
	got, _ := app.cache.GetNode("n1")
	if got.OwnerGatewayID != "" {
		t.Fatalf("newer release should clear previous owner, got %+v", got)
	}
}

func TestApplyRemoteIgnoresEmptyOwnerEvictWhenClaimed(t *testing.T) {
	resetSyncHooks(t)
	LocalGatewayID = "gw3"
	app := New("c", nil, cache.New("c"))
	now := time.Now()
	app.cache.PutNode(&model.Node{
		NodeID:         "n1",
		NamespaceID:    "ns",
		GroupName:      "g",
		ServiceName:    "svc",
		Ephemeral:      true,
		OwnerGatewayID: "gw2",
		LastBeatTime:   now,
	})
	app.ApplyRemote(model.NamingEvent{
		Type:      model.EventNodeEvicted,
		Timestamp: now.Add(time.Second),
		ChangedNode: &model.Node{
			NodeID:         "n1",
			NamespaceID:    "ns",
			GroupName:      "g",
			ServiceName:    "svc",
			OwnerGatewayID: "",
			Ephemeral:      true,
		},
	})
	if _, ok := app.cache.GetNode("n1"); !ok {
		t.Fatal("unowned replica evict must not drop a claimed node")
	}
}

func TestApplyRemoteIgnoresServiceDeletedWhenNodesRemain(t *testing.T) {
	resetSyncHooks(t)
	app := New("c", nil, cache.New("c"))
	app.cache.PutNode(&model.Node{
		NodeID:         "n1",
		NamespaceID:    "ns",
		GroupName:      "g",
		ServiceName:    "svc",
		IP:             "10.0.0.1",
		Port:           80,
		Ephemeral:      true,
		OwnerGatewayID: "gw2",
	})
	app.ApplyRemote(model.NamingEvent{
		Type:        model.EventServiceDeleted,
		NamespaceID: "ns",
		GroupName:   "g",
		ServiceName: "svc",
	})
	if _, ok := app.cache.GetNode("n1"); !ok {
		t.Fatal("stale SERVICE_DELETED must not wipe live nodes")
	}
}

func TestEvictorSkipsRemoteOwner(t *testing.T) {
	resetSyncHooks(t)
	LocalGatewayID = "n1"
	app := New("c", nil, cache.New("c"))
	app.cache.PutNode(&model.Node{
		NodeID:         "x",
		NamespaceID:    "ns",
		GroupName:      "g",
		ServiceName:    "s",
		Ephemeral:      true,
		OwnerGatewayID: "n2",
		LastBeatTime:   time.Now().Add(-time.Hour),
		Status:         model.NodeUP,
		HealthyStatus:  model.Healthy,
	})
	e := NewEvictor(app, time.Second, time.Millisecond)
	e.sweep()
	if _, ok := app.cache.GetNode("x"); !ok {
		t.Fatal("remote-owned node must not be evicted")
	}
}

func TestEvictorRechecksCacheBeatBeforeEvict(t *testing.T) {
	resetSyncHooks(t)
	LocalGatewayID = "gw1"
	app := New("c", nil, cache.New("c"))
	app.cache.PutNode(&model.Node{
		NodeID:         "pay-1",
		TenantID:       "t",
		NamespaceID:    "ns",
		GroupName:      "g",
		ServiceName:    "resident-pay",
		IP:             "127.0.0.1",
		Port:           18518,
		Ephemeral:      true,
		OwnerGatewayID: "gw1",
		LastBeatTime:   time.Now().Add(-time.Hour),
		Status:         model.NodeUP,
		HealthyStatus:  model.Healthy,
	})
	app.cache.TouchBeat("pay-1", time.Now(), "gw1")
	e := NewEvictor(app, time.Second, time.Second)
	e.sweep()
	if _, ok := app.cache.GetNode("pay-1"); !ok {
		t.Fatal("node that heartbeated after AllNodes snapshot must not be evicted")
	}
}

func TestEvictorEmptyOwnerStaleDropsLocalWithoutBroadcast(t *testing.T) {
	resetSyncHooks(t)
	LocalGatewayID = "gw1"
	var types []string
	NamingHook = func(ev model.NamingEvent) { types = append(types, ev.Type) }
	app := New("c", nil, cache.New("c"))
	ch := make(chan model.NamingEvent, 2)
	app.subs.add("ns", "g", []string{"s"}, ch, "w")
	app.cache.PutNode(&model.Node{
		NodeID:        "x",
		NamespaceID:   "ns",
		GroupName:     "g",
		ServiceName:   "s",
		Ephemeral:     true,
		LastBeatTime:  time.Now().Add(-time.Hour),
		Status:        model.NodeUP,
		HealthyStatus: model.Healthy,
	})
	e := NewEvictor(app, time.Second, time.Millisecond)
	e.sweep()
	if _, ok := app.cache.GetNode("x"); ok {
		t.Fatal("unowned stale ephemeral should drop from replica cache")
	}
	if len(types) != 0 {
		t.Fatalf("replica drop must not broadcast, got %v", types)
	}
	select {
	case ev := <-ch:
		if ev.Type != model.EventNodeEvicted {
			t.Fatalf("local subscriber should see evict, got %s", ev.Type)
		}
	default:
		t.Fatal("local subscriber should see replica drop")
	}
}

func TestOnClientLostServerDrainKeepsNodeAndClearsOwner(t *testing.T) {
	resetSyncHooks(t)
	LocalGatewayID = "gw1"
	var types []string
	NamingHook = func(ev model.NamingEvent) { types = append(types, ev.Type) }
	app := New("c", nil, cache.New("c"))
	now := time.Now().Add(-time.Second)
	app.cache.PutNode(&model.Node{
		NodeID:         "n1",
		TenantID:       "t",
		NamespaceID:    "ns",
		GroupName:      "g",
		ServiceName:    "s",
		IP:             "10.0.0.1",
		Port:           80,
		Ephemeral:      true,
		OwnerGatewayID: "gw1",
		ConnectionID:   "conn-1",
		Status:         model.NodeUP,
		HealthyStatus:  model.Healthy,
		LastBeatTime:   now,
	})
	app.trackConnection("conn-1", "n1")
	app.OnClientLost(context.Background(), "conn-1", model.DisconnectServerDrain)
	got, ok := app.cache.GetNode("n1")
	if !ok {
		t.Fatal("gateway drain must keep the ephemeral node for SDK reconnect")
	}
	if got.OwnerGatewayID != "" || got.ConnectionID != "" {
		t.Fatalf("drain should release owner and connection, got %+v", got)
	}
	if !got.LastBeatTime.After(now) {
		t.Fatal("drain should refresh LastBeatTime so replicas do not evict immediately")
	}
	if len(types) != 1 || types[0] != model.EventNodeUpdated {
		t.Fatalf("drain should replicate UPDATED not EVICTED, got %v", types)
	}
}

func TestOnClientLostServerDrainSkipsRebound(t *testing.T) {
	resetSyncHooks(t)
	LocalGatewayID = "gw1"
	app := New("c", nil, cache.New("c"))
	app.cache.PutNode(&model.Node{
		NodeID:         "n1",
		NamespaceID:    "ns",
		GroupName:      "g",
		ServiceName:    "s",
		Ephemeral:      true,
		OwnerGatewayID: "gw1",
		ConnectionID:   "new-conn",
	})
	app.trackConnection("old-conn", "n1")
	app.OnClientLost(context.Background(), "old-conn", model.DisconnectServerDrain)
	got, ok := app.cache.GetNode("n1")
	if !ok || got.ConnectionID != "new-conn" || got.OwnerGatewayID != "gw1" {
		t.Fatalf("drain of old stream must not release rebound node, got %+v ok=%v", got, ok)
	}
}

func TestHeartbeatTakesOwnershipOnce(t *testing.T) {
	resetSyncHooks(t)
	LocalGatewayID = "n1"
	var events []string
	NamingHook = func(ev model.NamingEvent) { events = append(events, ev.Type) }
	app := New("c", nil, cache.New("c"))
	app.cache.PutNode(&model.Node{
		NodeID:         "x",
		NamespaceID:    "ns",
		GroupName:      "g",
		ServiceName:    "s",
		OwnerGatewayID: "n2",
		Status:         model.NodeUP,
		HealthyStatus:  model.Healthy,
		LastBeatTime:   time.Now().Add(-time.Second),
	})
	cc := contract.CallContext{CenterInstanceName: "c", TenantID: "t"}
	if err := app.Heartbeat(context.Background(), cc, "x", nil); err != nil {
		t.Fatal(err)
	}
	got, _ := app.cache.GetNode("x")
	if got.OwnerGatewayID != "n1" {
		t.Fatalf("want owner n1, got %s", got.OwnerGatewayID)
	}
	if len(events) != 1 || events[0] != model.EventNodeUpdated {
		t.Fatalf("want one UPDATED replicate, got %v", events)
	}
	if err := app.Heartbeat(context.Background(), cc, "x", nil); err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("second heartbeat must not re-replicate, got %v", events)
	}
}

func TestApplyRemoteIgnoresUnknownEvict(t *testing.T) {
	resetSyncHooks(t)
	app := New("c", nil, cache.New("c"))
	ch := make(chan model.NamingEvent, 2)
	app.subs.add("ns", "g", []string{"s"}, ch, "w")
	app.ApplyRemote(model.NamingEvent{
		Type: model.EventNodeEvicted,
		ChangedNode: &model.Node{
			NodeID: "ghost", NamespaceID: "ns", GroupName: "g", ServiceName: "s",
			OwnerGatewayID: "gw1", Ephemeral: true,
		},
	})
	select {
	case ev := <-ch:
		t.Fatalf("unknown evict must not push, got %s", ev.Type)
	default:
	}
}

func TestOnClientLostSkipsTransferredOwner(t *testing.T) {
	resetSyncHooks(t)
	LocalGatewayID = "n1"
	app := New("c", nil, cache.New("c"))
	app.cache.PutNode(&model.Node{
		NodeID:         "x",
		NamespaceID:    "ns",
		GroupName:      "g",
		ServiceName:    "s",
		Ephemeral:      true,
		OwnerGatewayID: "n2",
	})
	app.trackConnection("lost", "x")
	app.OnClientLost(context.Background(), "lost", model.DisconnectClientLost)
	if _, ok := app.cache.GetNode("x"); !ok {
		t.Fatal("transferred owner node must survive local disconnect")
	}
}

func TestSubscribePushesCacheSnapshot(t *testing.T) {
	resetSyncHooks(t)
	app := New("c", nil, cache.New("c"))
	app.cache.PutNode(&model.Node{
		NodeID:        "n1",
		NamespaceID:   "ns",
		GroupName:     "g",
		ServiceName:   "svc",
		IP:            "10.0.0.1",
		Port:          80,
		Status:        model.NodeUP,
		HealthyStatus: model.Healthy,
	})
	ch := make(chan model.NamingEvent, 4)
	cc := contract.CallContext{
		TenantID:           "t",
		CenterInstanceName: "c",
		NamespaceID:        "ns",
		ConnectionID:       "sub",
	}
	if err := app.SubscribeServices(context.Background(), cc, "g", []string{"svc"}, ch); err != nil {
		t.Fatal(err)
	}
	select {
	case ev := <-ch:
		if ev.Type != model.EventNodeRegistered || ev.ChangedNode == nil || ev.ChangedNode.NodeID != "n1" {
			t.Fatalf("want snapshot REGISTERED n1, got %+v", ev)
		}
	default:
		t.Fatal("subscribe should push current cache snapshot")
	}
}

func TestReplicaGatewayIDsIncludesLocalAndOwners(t *testing.T) {
	resetSyncHooks(t)
	LocalGatewayID = "local"
	app := New("c", nil, cache.New("c"))
	app.noteOwner("peer-a")
	app.cache.PutNode(&model.Node{
		NodeID:         "n1",
		OwnerGatewayID: "peer-b",
	})
	got := app.ReplicaGatewayIDs()
	want := map[string]struct{}{"local": {}, "peer-a": {}, "peer-b": {}}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for _, id := range got {
		if _, ok := want[id]; !ok {
			t.Fatalf("unexpected %s in %v", id, got)
		}
	}
}

func TestRegisterNodeCreatesParentService(t *testing.T) {
	app := New("c", nil, cache.New("c"))
	app.cache.SetNamespace(&model.Namespace{
		NamespaceID:        "ns",
		CenterInstanceName: "c",
		Active:             true,
	})
	cc := contract.CallContext{TenantID: "t", CenterInstanceName: "c", NamespaceID: "ns"}
	if err := app.RegisterNode(context.Background(), cc, &model.Node{
		GroupName:   "g",
		ServiceName: "order",
		IP:          "10.0.0.1",
		Port:        80,
		Ephemeral:   true,
	}); err != nil {
		t.Fatal(err)
	}
	svc, ok := app.cache.GetService("ns", "g", "order")
	if !ok {
		t.Fatal("registering a node must create the parent service")
	}
	if svc.ServiceName != "order" || len(svc.Nodes) != 1 {
		t.Fatalf("service=%+v instances=%d", svc, len(svc.Nodes))
	}
}

func TestRegisterNodeRejectsEmptyServiceName(t *testing.T) {
	app := New("c", nil, cache.New("c"))
	app.cache.SetNamespace(&model.Namespace{
		NamespaceID:        "ns",
		CenterInstanceName: "c",
		Active:             true,
	})
	err := app.RegisterNode(context.Background(), contract.CallContext{
		TenantID: "t", CenterInstanceName: "c", NamespaceID: "ns",
	}, &model.Node{IP: "10.0.0.1", Port: 80})
	if err != contract.ErrInvalidArgument {
		t.Fatalf("want INVALID_ARGUMENT, got %v", err)
	}
}

func TestIngestPersistentNodeMarksUnhealthyUntilHeartbeat(t *testing.T) {
	app := New("c", nil, cache.New("c"))
	n := &model.Node{
		NodeID:        "p1",
		NamespaceID:   "ns",
		GroupName:     "g",
		ServiceName:   "s",
		IP:            "10.0.0.1",
		Port:          80,
		Status:        model.NodeUP,
		HealthyStatus: model.Healthy,
		LastBeatTime:  time.Now().Add(-time.Hour),
		Ephemeral:     false,
	}
	app.ingestStoredNode(n)
	got, ok := app.cache.GetNode("p1")
	if !ok {
		t.Fatal("expected node in cache")
	}
	if got.HealthyStatus != model.Unhealthy {
		t.Fatalf("stale DB HEALTHY must not survive load, got %s", got.HealthyStatus)
	}
	if got.IsHealthy() {
		t.Fatal("loaded persistent node must not count as healthy before heartbeat")
	}
	app.cache.TouchBeat("p1", time.Now(), "local")
	live, _ := app.cache.GetNode("p1")
	if live.HealthyStatus != model.Healthy || !live.IsHealthy() {
		t.Fatalf("heartbeat should restore health, got %+v", live)
	}
}

func TestIngestPersistentNodeKeepsRecentHealthy(t *testing.T) {
	app := New("c", nil, cache.New("c"))
	n := &model.Node{
		NodeID:        "p1",
		NamespaceID:   "ns",
		GroupName:     "g",
		ServiceName:   "s",
		IP:            "10.0.0.1",
		Port:          80,
		Status:        model.NodeUP,
		HealthyStatus: model.Healthy,
		LastBeatTime:  time.Now().Add(-5 * time.Second),
		Ephemeral:     false,
	}
	app.ingestStoredNode(n)
	got, ok := app.cache.GetNode("p1")
	if !ok || !got.IsHealthy() {
		t.Fatalf("recent persistent beat should stay healthy for gateway routing, got %+v", got)
	}
}

func TestIngestStaleEphemeralSkipped(t *testing.T) {
	app := New("c", nil, cache.New("c"))
	n := &model.Node{
		NodeID:        "e1",
		TenantID:      "t",
		NamespaceID:   "ns",
		GroupName:     "g",
		ServiceName:   "s",
		IP:            "10.0.0.1",
		Port:          80,
		Status:        model.NodeUP,
		HealthyStatus: model.Healthy,
		LastBeatTime:  time.Now().Add(-time.Hour),
		Ephemeral:     true,
	}
	if app.ingestStoredNode(n) {
		t.Fatal("stale ephemeral must not enter cache")
	}
	if _, ok := app.cache.GetNode("e1"); ok {
		t.Fatal("stale ephemeral must not be discoverable")
	}
}

func TestListNodesViewNotReadyWhenCold(t *testing.T) {
	app := New("c", nil, cache.New("c"))
	app.cache.SetNamespace(&model.Namespace{NamespaceID: "ns", CenterInstanceName: "c", Active: true})
	app.cache.SetService(&model.Service{NamespaceID: "ns", GroupName: "g", ServiceName: "s"})
	app.ResetViewWarm()
	cc := contract.CallContext{TenantID: "t", CenterInstanceName: "c", NamespaceID: "ns"}
	_, err := app.ListNodes(context.Background(), cc, "g", "s", true)
	if err != contract.ErrViewNotReady {
		t.Fatalf("cold empty view must be VIEW_NOT_READY, got %v", err)
	}
	app.MarkViewWarm()
	list, err := app.ListNodes(context.Background(), cc, "g", "s", true)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Fatalf("warm empty view should return empty list, got %d", len(list))
	}
}

func TestEvictorMarksPersistentUnhealthyWhenLastBeatZero(t *testing.T) {
	app := New("c", nil, cache.New("c"))
	app.cache.PutNode(&model.Node{
		NodeID:        "p1",
		TenantID:      "t",
		NamespaceID:   "ns",
		GroupName:     "g",
		ServiceName:   "s",
		IP:            "10.0.0.1",
		Port:          80,
		Ephemeral:     false,
		Status:        model.NodeUP,
		HealthyStatus: model.Healthy,
	})
	e := NewEvictor(app, time.Second, time.Millisecond)
	e.sweep()
	got, ok := app.cache.GetNode("p1")
	if !ok {
		t.Fatal("persistent node should remain")
	}
	if got.HealthyStatus != model.Unhealthy {
		t.Fatalf("zero lastBeat persistent should be unhealthy, got %s", got.HealthyStatus)
	}
}

func TestRegisterServiceOverwriteEmitsUpdated(t *testing.T) {
	resetSyncHooks(t)
	var types []string
	NamingHook = func(ev model.NamingEvent) { types = append(types, ev.Type) }
	app := New("c", nil, cache.New("c"))
	app.cache.SetNamespace(&model.Namespace{NamespaceID: "ns", CenterInstanceName: "c", Active: true})
	cc := contract.CallContext{TenantID: "t", CenterInstanceName: "c", NamespaceID: "ns"}
	svc := &model.Service{GroupName: "g", ServiceName: "order"}
	if err := app.RegisterService(context.Background(), cc, svc); err != nil {
		t.Fatal(err)
	}
	if err := app.RegisterService(context.Background(), cc, &model.Service{GroupName: "g", ServiceName: "order", Description: "v2"}); err != nil {
		t.Fatal(err)
	}
	if len(types) != 2 || types[0] != model.EventServiceAdded || types[1] != model.EventServiceUpdated {
		t.Fatalf("want ADDED then UPDATED, got %v", types)
	}
}

func TestRemoveLastEphemeralDropsAutoCreatedService(t *testing.T) {
	app := New("c", nil, cache.New("c"))
	app.cache.SetNamespace(&model.Namespace{NamespaceID: "ns", CenterInstanceName: "c", Active: true})
	cc := contract.CallContext{TenantID: "t", CenterInstanceName: "c", NamespaceID: "ns"}
	if err := app.RegisterNode(context.Background(), cc, &model.Node{
		NodeID: "n1", GroupName: "g", ServiceName: "order", IP: "10.0.0.1", Port: 80, Ephemeral: true,
	}); err != nil {
		t.Fatal(err)
	}
	if _, ok := app.cache.GetService("ns", "g", "order"); !ok {
		t.Fatal("auto-created service missing")
	}
	if err := app.DeregisterNode(context.Background(), cc, "n1"); err != nil {
		t.Fatal(err)
	}
	if _, ok := app.cache.GetService("ns", "g", "order"); ok {
		t.Fatal("empty auto-created service should be pruned")
	}
}

func TestAutoCreatedPruneDoesNotPushServiceDeletedToSubscribers(t *testing.T) {
	app := New("c", nil, cache.New("c"))
	app.cache.SetNamespace(&model.Namespace{NamespaceID: "ns", CenterInstanceName: "c", Active: true})
	ch := make(chan model.NamingEvent, 8)
	app.subs.add("ns", "g", []string{"order"}, ch, "w")
	cc := contract.CallContext{TenantID: "t", CenterInstanceName: "c", NamespaceID: "ns"}
	if err := app.RegisterNode(context.Background(), cc, &model.Node{
		NodeID: "n1", GroupName: "g", ServiceName: "order", IP: "10.0.0.1", Port: 80, Ephemeral: true,
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.DeregisterNode(context.Background(), cc, "n1"); err != nil {
		t.Fatal(err)
	}
	gotDeleted := false
	for len(ch) > 0 {
		if (<-ch).Type == model.EventServiceDeleted {
			gotDeleted = true
		}
	}
	if gotDeleted {
		t.Fatal("auto-created prune must not push SERVICE_DELETED to SDK subscribers")
	}
}

func TestRegisterServiceAttachedNodeDoesNotPlantGhost(t *testing.T) {
	app := New("c", nil, cache.New("c"))
	app.cache.SetNamespace(&model.Namespace{NamespaceID: "ns", CenterInstanceName: "c", Active: true})
	cc := contract.CallContext{TenantID: "t", CenterInstanceName: "c", NamespaceID: "ns"}
	if err := app.RegisterService(context.Background(), cc, &model.Service{
		GroupName: "g", ServiceName: "order",
		Nodes: []*model.Node{{IP: "10.0.0.1", Port: 80, Ephemeral: true}},
	}); err != nil {
		t.Fatal(err)
	}
	svc, ok := app.cache.GetService("ns", "g", "order")
	if !ok || len(svc.Nodes) != 0 {
		t.Fatalf("RegisterService must not plant nodes, got %+v ok=%v", svc, ok)
	}
	if err := app.RegisterNode(context.Background(), cc, &model.Node{
		NodeID: "n1", GroupName: "g", ServiceName: "order", IP: "10.0.0.1", Port: 80, Ephemeral: true,
	}); err != nil {
		t.Fatal(err)
	}
	svc, ok = app.cache.GetService("ns", "g", "order")
	if !ok || len(svc.Nodes) != 1 || svc.Nodes[0].NodeID != "n1" {
		t.Fatalf("want single registered node, got %+v ok=%v", svc, ok)
	}
}

func TestListServicesIncludesEmptyCatalog(t *testing.T) {
	app := New("c", nil, cache.New("c"))
	app.cache.SetNamespace(&model.Namespace{NamespaceID: "ns", CenterInstanceName: "c", Active: true})
	cc := contract.CallContext{TenantID: "t", CenterInstanceName: "c", NamespaceID: "ns"}
	if err := app.RegisterService(context.Background(), cc, &model.Service{GroupName: "g", ServiceName: "empty"}); err != nil {
		t.Fatal(err)
	}
	if err := app.RegisterNode(context.Background(), cc, &model.Node{
		NodeID: "n1", GroupName: "g", ServiceName: "live", IP: "10.0.0.1", Port: 80, Ephemeral: true,
	}); err != nil {
		t.Fatal(err)
	}
	list, err := app.ListServices(context.Background(), cc, "")
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]int{}
	for _, svc := range list {
		names[svc.ServiceName] = len(svc.Nodes)
	}
	if names["empty"] != 0 || names["live"] != 1 {
		t.Fatalf("list must keep empty catalog as 0 nodes, got %+v", names)
	}
}

func TestRefreshLiveViewPrunesEmptyAutoCreated(t *testing.T) {
	app := New("c", nil, cache.New("c"))
	app.cache.SetService(&model.Service{NamespaceID: "ns", GroupName: "g", ServiceName: "ghost", AutoCreated: true})
	app.refreshLiveView(context.Background())
	if _, ok := app.cache.GetService("ns", "g", "ghost"); ok {
		t.Fatal("empty auto-created shell should be pruned on live refresh")
	}
}

func TestRegisterServiceLastEphemeralDropsEmptyCatalog(t *testing.T) {
	app := New("c", nil, cache.New("c"))
	app.cache.SetNamespace(&model.Namespace{NamespaceID: "ns", CenterInstanceName: "c", Active: true})
	cc := contract.CallContext{TenantID: "t", CenterInstanceName: "c", NamespaceID: "ns"}
	if err := app.RegisterService(context.Background(), cc, &model.Service{GroupName: "g", ServiceName: "order"}); err != nil {
		t.Fatal(err)
	}
	if err := app.RegisterNode(context.Background(), cc, &model.Node{
		NodeID: "n1", GroupName: "g", ServiceName: "order", IP: "10.0.0.1", Port: 80, Ephemeral: true,
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.DeregisterNode(context.Background(), cc, "n1"); err != nil {
		t.Fatal(err)
	}
	if _, ok := app.cache.GetService("ns", "g", "order"); ok {
		t.Fatal("last ephemeral leave should drop empty catalog service")
	}
}

func TestLastEphemeralKeepsExternalCatalog(t *testing.T) {
	app := New("c", nil, cache.New("c"))
	app.cache.SetNamespace(&model.Namespace{NamespaceID: "ns", CenterInstanceName: "c", Active: true})
	app.cache.SetService(&model.Service{
		TenantID: "t", NamespaceID: "ns", GroupName: "g", ServiceName: "nacos-svc", ServiceType: "NACOS",
	})
	cc := contract.CallContext{TenantID: "t", CenterInstanceName: "c", NamespaceID: "ns"}
	if err := app.RegisterNode(context.Background(), cc, &model.Node{
		NodeID: "n1", GroupName: "g", ServiceName: "nacos-svc", IP: "10.0.0.1", Port: 80, Ephemeral: true,
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.DeregisterNode(context.Background(), cc, "n1"); err != nil {
		t.Fatal(err)
	}
	svc, ok := app.cache.GetService("ns", "g", "nacos-svc")
	if !ok || svc.ServiceType != "NACOS" {
		t.Fatalf("external catalog must stay after last ephemeral leave, got %+v ok=%v", svc, ok)
	}
}

func TestHeartbeatPersistsPersistentNodeBeat(t *testing.T) {
	resetSyncHooks(t)
	app := New("c", nil, cache.New("c"))
	app.beatPersistEvery = 0
	app.cache.SetNamespace(&model.Namespace{NamespaceID: "ns", CenterInstanceName: "c", Active: true})
	cc := contract.CallContext{TenantID: "t", CenterInstanceName: "c", NamespaceID: "ns"}
	if err := app.RegisterNode(context.Background(), cc, &model.Node{
		NodeID: "p1", GroupName: "g", ServiceName: "s", IP: "10.0.0.1", Port: 80, Ephemeral: false,
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.Heartbeat(context.Background(), cc, "p1", nil); err != nil {
		t.Fatal(err)
	}
	first, ok := app.lastPersistBeat.Load("p1")
	if !ok {
		t.Fatal("persistent heartbeat should record a persist attempt")
	}
	app.beatPersistEvery = time.Hour
	if err := app.Heartbeat(context.Background(), cc, "p1", nil); err != nil {
		t.Fatal(err)
	}
	second, _ := app.lastPersistBeat.Load("p1")
	if !first.(time.Time).Equal(second.(time.Time)) {
		t.Fatal("throttled heartbeat should not persist again")
	}
}

func TestApplyRemotePrunesEmptyAutoService(t *testing.T) {
	resetSyncHooks(t)
	app := New("c", nil, cache.New("c"))
	inst := &model.Node{
		NodeID: "n1", NamespaceID: "ns", GroupName: "g", ServiceName: "svc",
		IP: "10.0.0.1", Port: 80, Ephemeral: true,
	}
	app.ApplyRemote(model.NamingEvent{Type: model.EventNodeRegistered, NamespaceID: "ns", GroupName: "g", ServiceName: "svc", ChangedNode: inst})
	if svc, ok := app.cache.GetService("ns", "g", "svc"); !ok || !svc.AutoCreated {
		t.Fatalf("remote register should create auto service, got %+v ok=%v", svc, ok)
	}
	app.ApplyRemote(model.NamingEvent{Type: model.EventNodeEvicted, NamespaceID: "ns", GroupName: "g", ServiceName: "svc", ChangedNode: inst})
	if _, ok := app.cache.GetService("ns", "g", "svc"); ok {
		t.Fatal("remote evict of last ephemeral should prune auto service")
	}
}

func TestHeartbeatSnapshotRepairsMissingCache(t *testing.T) {
	resetSyncHooks(t)
	app := New("c", nil, cache.New("c"))
	app.cache.SetNamespace(&model.Namespace{NamespaceID: "ns", CenterInstanceName: "c", Active: true})
	cc := contract.CallContext{TenantID: "t", CenterInstanceName: "c", NamespaceID: "ns"}
	snap := &model.Service{
		NamespaceID: "ns", GroupName: "g", ServiceName: "order",
		Nodes: []*model.Node{{
			NodeID: "n1", NamespaceID: "ns", GroupName: "g", ServiceName: "order",
			IP: "10.0.0.1", Port: 80, Ephemeral: true,
		}},
	}
	if err := app.Heartbeat(context.Background(), cc, "n1", snap); err != nil {
		t.Fatal(err)
	}
	got, ok := app.cache.GetNode("n1")
	if !ok || got.IP != "10.0.0.1" {
		t.Fatalf("heartbeat snapshot should restore missing node, got %+v ok=%v", got, ok)
	}
}

