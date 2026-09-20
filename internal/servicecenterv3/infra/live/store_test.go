package live

import (
	"context"
	"testing"
	"time"

	"gateway/internal/servicecenterv3/model"
	"gateway/pkg/cache/memory"
)

func testStore(t *testing.T) *Store {
	t.Helper()
	c, err := memory.NewMemoryCache(&memory.MemoryConfig{
		CleanupInterval:   time.Hour,
		EnableLazyCleanup: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return New(c)
}

func TestListByCenterDropsStale(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	n1 := &model.Node{TenantID: "t", NodeID: "n1", CenterInstanceName: "c", NamespaceID: "ns", GroupName: "g", ServiceName: "order", IP: "10.0.0.1", Port: 80}
	n2 := &model.Node{TenantID: "t", NodeID: "n2", CenterInstanceName: "c", NamespaceID: "ns", GroupName: "g", ServiceName: "order", IP: "10.0.0.2", Port: 80}
	if err := s.Put(ctx, n1, time.Minute); err != nil {
		t.Fatal(err)
	}
	if err := s.Put(ctx, n2, time.Minute); err != nil {
		t.Fatal(err)
	}
	if err := s.c.Delete(ctx, nodeKey("t", "n2")); err != nil {
		t.Fatal(err)
	}
	got, err := s.ListByCenter(ctx, "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].NodeID != "n1" {
		t.Fatalf("want only n1, got %+v", got)
	}
	again, err := s.ListByCenter(ctx, "c")
	if err != nil || len(again) != 1 {
		t.Fatalf("stale service field should be gone, got %+v err=%v", again, err)
	}
}

func TestHierarchyMaps(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	if err := s.Put(ctx, &model.Node{
		TenantID: "t", NodeID: "n1", CenterInstanceName: "c",
		NamespaceID: "ns", GroupName: "g", ServiceName: "order",
		IP: "10.0.0.1", Port: 80,
	}, time.Minute); err != nil {
		t.Fatal(err)
	}
	if err := s.Put(ctx, &model.Node{
		TenantID: "t", NodeID: "n2", CenterInstanceName: "c",
		NamespaceID: "ns", GroupName: "g", ServiceName: "pay",
		IP: "10.0.0.2", Port: 81,
	}, time.Minute); err != nil {
		t.Fatal(err)
	}
	centers, err := s.c.HGetAll(ctx, centerKey("c"))
	if err != nil || centers["ns"] != present {
		t.Fatalf("center map should bind namespace, got %v err=%v", centers, err)
	}
	services, err := s.c.HGetAll(ctx, nsKey("c", "ns"))
	if err != nil {
		t.Fatal(err)
	}
	if services[serviceRef("g", "order")] != "t" || services[serviceRef("g", "pay")] != "t" {
		t.Fatalf("namespace map should index services, got %v", services)
	}
	order, err := s.c.HGetAll(ctx, serviceKey("t", "ns", "g", "order"))
	if err != nil || order["n1"] != present || len(order) != 1 {
		t.Fatalf("order should have n1 only, got %v err=%v", order, err)
	}
	got, err := s.ListByService(ctx, "t", "ns", "g", "order")
	if err != nil || len(got) != 1 || got[0].NodeID != "n1" {
		t.Fatalf("list order want n1, got %+v err=%v", got, err)
	}
	all, err := s.ListByCenter(ctx, "c")
	if err != nil || len(all) != 2 {
		t.Fatalf("center list want 2, got %d err=%v", len(all), err)
	}
	if err := s.Delete(ctx, &model.Node{
		TenantID: "t", NodeID: "n1", CenterInstanceName: "c",
		NamespaceID: "ns", GroupName: "g", ServiceName: "order",
	}); err != nil {
		t.Fatal(err)
	}
	left, _ := s.c.HGetAll(ctx, nsKey("c", "ns"))
	if _, ok := left[serviceRef("g", "order")]; ok {
		t.Fatalf("empty service should be pruned from namespace map, got %v", left)
	}
}

func TestTouchMissFallsThrough(t *testing.T) {
	s := testStore(t)
	ok, err := s.Touch(context.Background(), "t", "missing", time.Minute)
	if err != nil || ok {
		t.Fatalf("missing key want false, got ok=%v err=%v", ok, err)
	}
}
