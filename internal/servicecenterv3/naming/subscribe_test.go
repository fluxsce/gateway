package naming

import (
	"context"
	"sync"
	"testing"
	"time"

	"gateway/internal/servicecenterv3/contract"
	"gateway/internal/servicecenterv3/infra/cache"
	"gateway/internal/servicecenterv3/model"
)

func TestListSubscribersResolvesPeerService(t *testing.T) {
	app := New("center-1", nil, cache.New("center-1"))
	app.cache.PutNode(&model.Node{
		NodeID:       "n-pay",
		NamespaceID:  "ns",
		GroupName:    "g",
		ServiceName:  "pay",
		ConnectionID: "conn-pay",
	})
	app.trackConnection("conn-pay", "n-pay")
	ch := make(chan model.NamingEvent, 1)
	app.subs.add("ns", "g", []string{"order"}, ch, "conn-pay")
	got := app.ListSubscribers("ns", "g", "order")
	if len(got) != 1 || got[0].ServiceName != "pay" {
		t.Fatalf("inbound peer should be pay, got %+v", got)
	}
}

func TestListSubscriptionsExcludesSelfService(t *testing.T) {
	app := New("center-1", nil, cache.New("center-1"))
	app.cache.PutNode(&model.Node{
		NodeID:       "n1",
		NamespaceID:  "ns",
		GroupName:    "g",
		ServiceName:  "order",
		ConnectionID: "conn-a",
	})
	ch := make(chan model.NamingEvent, 1)
	app.subs.add("ns", "g", []string{"order", "pay"}, ch, "conn-a")
	got := app.ListSubscriptions("ns", "g", "order")
	if len(got) != 1 || got[0].ServiceName != "pay" {
		t.Fatalf("want only pay, got %+v", got)
	}
}

func TestListByConnectionsShowsOutboundServices(t *testing.T) {
	s := newSubscriber()
	ch := make(chan model.NamingEvent, 1)
	s.add("ns", "g", []string{"order", "pay"}, ch, "conn-a")
	s.addNamespace("ns", "", ch, "conn-b")
	got := s.listByConnections(map[string]struct{}{"conn-a": {}})
	if len(got) != 2 {
		t.Fatalf("want 2 outbound services, got %d", len(got))
	}
	if len(s.listByConnections(map[string]struct{}{"conn-b": {}})) != 1 {
		t.Fatal("namespace subscribe should appear as one outbound row")
	}
}

func TestListMatchingIncludesNamespaceSubscribe(t *testing.T) {
	s := newSubscriber()
	ch := make(chan model.NamingEvent, 1)
	s.add("ns", "g", []string{"svc"}, ch, "c-svc")
	s.addNamespace("ns", "", ch, "c-ns")
	got := s.listMatching("ns", "g", "svc")
	if len(got) != 2 {
		t.Fatalf("want 2 subscribers, got %d", len(got))
	}
	if len(s.listMatching("ns", "g", "other")) != 1 {
		t.Fatal("namespace subscribe should still match other services")
	}
}

func TestSubscriberMatchAndOffline(t *testing.T) {
	s := newSubscriber()
	ch := make(chan model.NamingEvent, 8)
	s.add("ns", "g", []string{"svc"}, ch, "c-watch")
	s.publish(model.NamingEvent{NamespaceID: "ns", GroupName: "g", ServiceName: "svc", Type: model.EventNodeRegistered})
	s.publish(model.NamingEvent{NamespaceID: "ns", GroupName: "g", ServiceName: "other", Type: model.EventNodeRegistered})
	if got := len(ch); got != 1 {
		t.Fatalf("want 1 matched event, got %d", got)
	}
}

func TestOnClientLostNotifiesOthersAndCleansSubscribeOnlyClient(t *testing.T) {
	app := New("center-1", nil, cache.New("center-1"))
	watchCh := make(chan model.NamingEvent, 8)
	app.subs.add("ns", "g", []string{"svc"}, watchCh, "watcher")
	lostCh := make(chan model.NamingEvent, 2)
	app.subs.add("ns", "g", []string{"svc"}, lostCh, "lost")

	inst := &model.Node{
		NodeID:      "n1",
		TenantID:    "t1",
		NamespaceID: "ns",
		GroupName:   "g",
		ServiceName: "svc",
		Ephemeral:   true,
		IP:          "10.0.0.1",
		Port:        80,
	}
	app.cache.PutNode(inst)
	app.trackConnection("lost", "n1")

	app.OnClientLost(context.Background(), "lost", model.DisconnectClientLost)

	select {
	case ev := <-watchCh:
		if ev.Type != model.EventNodeEvicted {
			t.Fatalf("watcher should see evict, got %s", ev.Type)
		}
		if ev.ChangedNode == nil || ev.ChangedNode.NodeID != "n1" {
			t.Fatalf("changed node: %+v", ev.ChangedNode)
		}
	case <-time.After(time.Second):
		t.Fatal("watcher did not receive offline event")
	}
	if _, ok := app.cache.GetNode("n1"); ok {
		t.Fatal("ephemeral node should be removed")
	}
	app.subs.publish(model.NamingEvent{NamespaceID: "ns", GroupName: "g", ServiceName: "svc", Type: model.EventNodeUpdated})
	select {
	case <-lostCh:
		t.Fatal("lost connection should not still receive events")
	default:
	}
}

func TestOnClientLostSkipsReboundConnection(t *testing.T) {
	app := New("center-1", nil, cache.New("center-1"))
	app.cache.SetNamespace(&model.Namespace{NamespaceID: "ns", CenterInstanceName: "center-1", Active: true})
	inst := &model.Node{
		NodeID:       "n1",
		TenantID:     "t1",
		NamespaceID:  "ns",
		GroupName:    "g",
		ServiceName:  "svc",
		Ephemeral:    true,
		IP:           "10.0.0.1",
		Port:         80,
		ConnectionID: "new-conn",
		Status:       model.NodeUP,
	}
	app.cache.PutNode(inst)
	app.trackConnection("old-conn", "n1")
	app.trackConnection("new-conn", "n1")
	app.OnClientLost(context.Background(), "old-conn", model.DisconnectClientLost)
	got, ok := app.cache.GetNode("n1")
	if !ok {
		t.Fatal("rebound ephemeral node must survive old stream cleanup")
	}
	if got.ConnectionID != "new-conn" {
		t.Fatalf("connection=%s", got.ConnectionID)
	}
}

func TestOnClientLostSubscribeOnlyDoesNotLeak(t *testing.T) {
	app := New("center-1", nil, cache.New("center-1"))
	ch := make(chan model.NamingEvent, 1)
	app.subs.add("ns", "g", []string{"svc"}, ch, "only-sub")
	app.OnClientLost(context.Background(), "only-sub", model.DisconnectClientLost)
	app.subs.publish(model.NamingEvent{NamespaceID: "ns", GroupName: "g", ServiceName: "svc", Type: model.EventServiceUpdated})
	select {
	case <-ch:
		t.Fatal("subscribe-only lost client still received event")
	default:
	}
}

func TestDeregisterUntracksConnectionIndex(t *testing.T) {
	app := New("center-1", nil, cache.New("center-1"))
	inst := &model.Node{
		NodeID:       "n1",
		TenantID:     "t1",
		NamespaceID:  "ns",
		GroupName:    "g",
		ServiceName:  "svc",
		Ephemeral:    true,
		IP:           "10.0.0.1",
		Port:         80,
		ConnectionID: "conn-1",
		Status:       model.NodeUP,
	}
	app.cache.PutNode(inst)
	app.trackConnection("conn-1", "n1")
	app.trackConnection("conn-1", "n1")
	if app.BoundNodeCount("conn-1") != 1 {
		t.Fatalf("duplicate track should be ignored, got %d", app.BoundNodeCount("conn-1"))
	}
	cc := contract.CallContext{TenantID: "t1", CenterInstanceName: "center-1"}
	if err := app.DeregisterNode(context.Background(), cc, "n1"); err != nil {
		t.Fatal(err)
	}
	if app.BoundNodeCount("conn-1") != 0 {
		t.Fatalf("deregister should untrack, got %d", app.BoundNodeCount("conn-1"))
	}
}

func TestRemoveByConnectionClearsSlotPointers(t *testing.T) {
	s := newSubscriber()
	ch := make(chan model.NamingEvent, 1)
	s.add("ns", "g", []string{"svc"}, ch, "c1")
	s.removeByConnection("c1")
	if len(s.items) != 0 {
		t.Fatalf("want empty items, got %d", len(s.items))
	}
	for i, it := range s.items[:cap(s.items)] {
		if it != nil {
			t.Fatalf("slot %d still holds %v", i, it)
		}
	}
}

func TestSubscribeBindsConnectionIDImmediately(t *testing.T) {
	app := New("center-1", nil, cache.New("center-1"))
	ch := make(chan model.NamingEvent, 1)
	cc := contract.CallContext{
		TenantID:           "t1",
		CenterInstanceName: "center-1",
		NamespaceID:        "ns",
		ConnectionID:       "conn-sub",
	}
	if err := app.SubscribeServices(context.Background(), cc, "g", []string{"svc"}, ch); err != nil {
		t.Fatal(err)
	}
	app.OnClientLost(context.Background(), "conn-sub", model.DisconnectClientLost)
	app.subs.publish(model.NamingEvent{NamespaceID: "ns", GroupName: "g", ServiceName: "svc", Type: model.EventNodeUpdated})
	select {
	case <-ch:
		t.Fatal("subscribe bound by connectionID should have been removed on lost")
	default:
	}
}

func TestPublishDoesNotDropWhenConsumerSlow(t *testing.T) {
	s := newSubscriber()
	ch := make(chan model.NamingEvent, 1)
	s.add("ns", "g", []string{"svc"}, ch, "c")
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 3; i++ {
			s.publish(model.NamingEvent{NamespaceID: "ns", GroupName: "g", ServiceName: "svc", Type: model.EventNodeUpdated})
		}
	}()
	got := 0
	deadline := time.After(2 * time.Second)
	for got < 3 {
		select {
		case <-ch:
			got++
		case <-deadline:
			t.Fatalf("expected 3 events without drop, got %d", got)
		}
	}
	<-done
}

func TestSubscriberConcurrentPublish(t *testing.T) {
	s := newSubscriber()
	chs := make([]chan model.NamingEvent, 8)
	for i := range chs {
		chs[i] = make(chan model.NamingEvent, 256)
		s.add("ns", "g", []string{"svc"}, chs[i], "c")
	}
	var pub sync.WaitGroup
	for i := 0; i < 50; i++ {
		pub.Add(1)
		go func() {
			defer pub.Done()
			s.publish(model.NamingEvent{NamespaceID: "ns", GroupName: "g", ServiceName: "svc", Type: model.EventNodeUpdated})
		}()
	}
	pub.Wait()
	total := 0
	for _, ch := range chs {
		total += len(ch)
	}
	if total != 50*len(chs) {
		t.Fatalf("want %d deliveries, got %d", 50*len(chs), total)
	}
}

func TestUnsubscribeDoesNotAffectOtherConnection(t *testing.T) {
	s := newSubscriber()
	first := make(chan model.NamingEvent, 4)
	second := make(chan model.NamingEvent, 4)
	s.add("ns", "g", []string{"svc"}, first, "conn-1")
	s.add("ns", "g", []string{"svc"}, second, "conn-2")
	s.remove(second)
	s.publish(model.NamingEvent{NamespaceID: "ns", GroupName: "g", ServiceName: "svc", Type: model.EventNodeUpdated})
	if len(first) != 1 {
		t.Fatalf("conn-1 should still receive, got %d", len(first))
	}
	if len(second) != 0 {
		t.Fatalf("conn-2 unsubscribe should not receive, got %d", len(second))
	}
}

func TestPublishDedupesSameChannel(t *testing.T) {
	s := newSubscriber()
	ch := make(chan model.NamingEvent, 4)
	s.add("ns", "g", []string{"svc"}, ch, "c1")
	s.add("ns", "g", []string{"svc"}, ch, "c1")
	s.publish(model.NamingEvent{NamespaceID: "ns", GroupName: "g", ServiceName: "svc", Type: model.EventNodeUpdated})
	if got := len(ch); got != 1 {
		t.Fatalf("same connection should get 1 event, got %d", got)
	}
}

func TestRemoveServicesKeepsOtherServiceAndOtherConnection(t *testing.T) {
	s := newSubscriber()
	first := make(chan model.NamingEvent, 4)
	second := make(chan model.NamingEvent, 4)
	s.add("ns", "g", []string{"order", "user"}, first, "c1")
	s.add("ns", "g", []string{"order"}, second, "c2")
	s.removeServices(first, "ns", "g", []string{"order"})
	s.publish(model.NamingEvent{NamespaceID: "ns", GroupName: "g", ServiceName: "order", Type: model.EventNodeUpdated})
	s.publish(model.NamingEvent{NamespaceID: "ns", GroupName: "g", ServiceName: "user", Type: model.EventNodeUpdated})
	if len(first) != 1 {
		t.Fatalf("c1 should only get user, got %d", len(first))
	}
	if (<-first).ServiceName != "user" {
		t.Fatal("c1 remaining service should be user")
	}
	if len(second) != 1 {
		t.Fatalf("c2 should still get order, got %d", len(second))
	}
}
