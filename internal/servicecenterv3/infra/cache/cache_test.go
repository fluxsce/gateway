package cache

import (
	"testing"

	"gateway/internal/servicecenterv3/model"
)

func TestCacheIsolatedByCenterName(t *testing.T) {
	a := New("dev")
	b := New("prod")
	a.PutNode(&model.Node{
		NodeID:    "n1",
		NamespaceID:   "ns",
		GroupName:     "DEFAULT_GROUP",
		ServiceName:   "svc",
		IP:            "127.0.0.1",
		Port:          8080,
		Status:        model.NodeUP,
		HealthyStatus: model.Healthy,
	})
	if _, ok := b.GetNode("n1"); ok {
		t.Fatal("prod cache should not see dev node")
	}
	if a.CenterName() == b.CenterName() {
		t.Fatal("center names should differ")
	}
	svc, ok := a.GetService("ns", "DEFAULT_GROUP", "svc")
	if !ok || len(svc.Nodes) != 1 {
		t.Fatalf("expected service with 1 node, got %+v", svc)
	}
}

func TestPutNodeIgnoresEmptyNodeID(t *testing.T) {
	c := New("dev")
	c.PutNode(&model.Node{NamespaceID: "ns", GroupName: "g", ServiceName: "svc", IP: "127.0.0.1", Port: 8080})
	if _, ok := c.GetService("ns", "g", "svc"); ok {
		t.Fatal("empty nodeId must not plant a service ghost")
	}
}

func TestRemoveNodeCompactsServiceSlice(t *testing.T) {
	c := New("dev")
	first := &model.Node{
		NodeID:  "n1",
		NamespaceID: "ns",
		GroupName:   "g",
		ServiceName: "svc",
		IP:          "127.0.0.1",
		Port:        8080,
	}
	second := &model.Node{
		NodeID:  "n2",
		NamespaceID: "ns",
		GroupName:   "g",
		ServiceName: "svc",
		IP:          "127.0.0.1",
		Port:        8081,
	}
	c.PutNode(first)
	c.PutNode(second)
	if got := c.RemoveNode("n1"); got == nil || got.NodeID != "n1" {
		t.Fatalf("removed: %+v", got)
	}
	svc, ok := c.GetService("ns", "g", "svc")
	if !ok || len(svc.Nodes) != 1 || svc.Nodes[0].NodeID != "n2" {
		t.Fatalf("expected remaining n2, got %+v", svc)
	}
	internal := c.services[serviceKey("ns", "g", "svc")]
	for i := len(internal.Nodes); i < cap(internal.Nodes); i++ {
		if internal.Nodes[:cap(internal.Nodes)][i] != nil {
			t.Fatalf("tail slot %d still holds node", i)
		}
	}
}
