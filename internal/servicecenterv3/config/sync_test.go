package config

import (
	"context"
	"testing"

	"gateway/internal/servicecenterv3/contract"
	"gateway/internal/servicecenterv3/infra/cache"
	"gateway/internal/servicecenterv3/model"
)

func TestApplyRemoteConfigDoesNotHook(t *testing.T) {
	t.Cleanup(func() { ConfigHook = nil })
	hooked := 0
	ConfigHook = func(model.ConfigEvent) { hooked++ }
	app := New("c", nil, cache.New("c"))
	ch := make(chan model.ConfigEvent, 2)
	app.subs.add("ns", "g", []string{"d"}, ch, "w")
	app.ApplyRemote(model.ConfigEvent{
		Type:        model.EventConfigPublished,
		NamespaceID: "ns",
		GroupName:   "g",
		DataID:      "d",
		Release: &model.ConfigRelease{
			NamespaceID: "ns",
			GroupName:   "g",
			DataID:      "d",
			Content:     "v1",
			Version:     1,
		},
	})
	if hooked != 0 {
		t.Fatalf("ApplyRemote must not call ConfigHook, got %d", hooked)
	}
	select {
	case ev := <-ch:
		if ev.Type != model.EventConfigPublished || ev.Release == nil || ev.Release.Content != "v1" {
			t.Fatalf("got %+v", ev)
		}
	default:
		t.Fatal("watch should see remote publish")
	}
}

func TestApplyRemoteConfigIgnoresStaleVersion(t *testing.T) {
	t.Cleanup(func() { ConfigHook = nil })
	app := New("c", nil, cache.New("c"))
	ch := make(chan model.ConfigEvent, 2)
	app.subs.add("ns", "g", []string{"d"}, ch, "w")
	app.ApplyRemote(model.ConfigEvent{
		Type:        model.EventConfigPublished,
		NamespaceID: "ns",
		GroupName:   "g",
		DataID:      "d",
		Release: &model.ConfigRelease{
			NamespaceID: "ns", GroupName: "g", DataID: "d",
			Version: 2,
		},
	})
	select {
	case ev := <-ch:
		t.Fatalf("stripped publish without DB must not push, got %s", ev.Type)
	default:
	}
}

func TestWatchPushesPublishedSnapshot(t *testing.T) {
	t.Cleanup(func() { ConfigHook = nil })
	c := cache.New("c")
	c.SetNamespace(&model.Namespace{
		TenantID:           "t",
		CenterInstanceName: "c",
		NamespaceID:        "ns",
		Active:             true,
	})
	app := New("c", nil, c)
	ch := make(chan model.ConfigEvent, 2)
	cc := contract.CallContext{
		TenantID:           "t",
		CenterInstanceName: "c",
		NamespaceID:        "ns",
		ConnectionID:       "w",
	}
	if err := app.Watch(context.Background(), cc, "g", []string{"d"}, ch); err != nil {
		t.Fatal(err)
	}
	select {
	case ev := <-ch:
		t.Fatalf("nil store must not invent a snapshot, got %+v", ev)
	default:
	}
	app.ApplyRemote(model.ConfigEvent{
		Type:        model.EventConfigPublished,
		NamespaceID: "ns",
		GroupName:   "g",
		DataID:      "d",
		Release: &model.ConfigRelease{
			NamespaceID: "ns",
			GroupName:   "g",
			DataID:      "d",
			Content:     "v1",
			Version:     1,
		},
	})
	select {
	case ev := <-ch:
		if ev.Type != model.EventConfigPublished || ev.DataID != "d" || ev.Release.Content != "v1" {
			t.Fatalf("want remote PUBLISHED d, got %+v", ev)
		}
	default:
		t.Fatal("watch should receive published change")
	}
}

func TestEmitStripsContentFromHook(t *testing.T) {
	t.Cleanup(func() { ConfigHook = nil })
	var hooked model.ConfigEvent
	ConfigHook = func(ev model.ConfigEvent) { hooked = ev }
	app := New("c", nil, cache.New("c"))
	ch := make(chan model.ConfigEvent, 1)
	app.subs.add("ns", "g", []string{"d"}, ch, "w")
	app.emit(model.ConfigEvent{
		Type:        model.EventConfigPublished,
		NamespaceID: "ns",
		GroupName:   "g",
		DataID:      "d",
		Release: &model.ConfigRelease{
			NamespaceID: "ns",
			GroupName:   "g",
			DataID:      "d",
			Content:     "huge-file",
			Version:     1,
		},
	})
	if hooked.Release == nil || hooked.Release.Content != "" {
		t.Fatalf("hook must not carry content, got %+v", hooked.Release)
	}
	select {
	case ev := <-ch:
		if ev.Release == nil || ev.Release.Content != "huge-file" {
			t.Fatalf("local watch still needs content, got %+v", ev.Release)
		}
	default:
		t.Fatal("local watch should receive content")
	}
}
