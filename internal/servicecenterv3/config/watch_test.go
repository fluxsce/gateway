package config

import (
	"context"
	"testing"

	"gateway/internal/servicecenterv3/contract"
	"gateway/internal/servicecenterv3/infra/cache"
	"gateway/internal/servicecenterv3/model"
)

func TestWatchOnClientLostDoesNotLeak(t *testing.T) {
	c := cache.New("center-1")
	c.SetNamespace(&model.Namespace{
		TenantID:           "t1",
		CenterInstanceName: "center-1",
		NamespaceID:        "ns",
		Active:             true,
	})
	app := New("center-1", nil, c)
	ch := make(chan model.ConfigEvent, 1)
	cc := contract.CallContext{
		TenantID:           "t1",
		CenterInstanceName: "center-1",
		NamespaceID:        "ns",
		ConnectionID:       "watch-1",
	}
	if err := app.Watch(context.Background(), cc, "g", []string{"data"}, ch); err != nil {
		t.Fatal(err)
	}
	app.OnClientLost(context.Background(), "watch-1")
	app.subs.publish(model.ConfigEvent{NamespaceID: "ns", GroupName: "g", DataID: "data", Type: model.EventConfigPublished})
	select {
	case <-ch:
		t.Fatal("lost watch connection still received event")
	default:
	}
}

func TestRemoveByConnectionClearsWatchSlotPointers(t *testing.T) {
	w := newWatcher()
	ch := make(chan model.ConfigEvent, 1)
	w.add("ns", "g", []string{"data"}, ch, "c1")
	w.removeByConnection("c1")
	if len(w.items) != 0 {
		t.Fatalf("want empty items, got %d", len(w.items))
	}
	for i, it := range w.items[:cap(w.items)] {
		if it != nil {
			t.Fatalf("slot %d still holds %v", i, it)
		}
	}
}
