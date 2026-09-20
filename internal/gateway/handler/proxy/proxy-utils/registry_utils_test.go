package proxyutils

import (
	"errors"
	"testing"
	"time"

	"gateway/internal/gateway/handler/service"
	"gateway/internal/servicecenterv3/contract"
	"gateway/internal/servicecenterv3/model"
)

func TestLastGoodProtectsEmptyDiscovery(t *testing.T) {
	key := "t||ns|g|svc-lastgood"
	rememberLastGood(key, []*service.NodeConfig{{ID: "n1", URL: "http://10.0.0.1:80", Weight: 1, Health: true, Enabled: true}})
	got, ok := recallLastGood(key)
	if !ok || len(got) != 1 || got[0].ID != "n1" {
		t.Fatalf("expected last good node, got %+v ok=%v", got, ok)
	}
	got[0].ID = "mutated"
	again, _ := recallLastGood(key)
	if again[0].ID != "n1" {
		t.Fatal("recall must return a copy")
	}
}

func TestForgetLastGoodOnEmptyHealthy(t *testing.T) {
	key := "t||ns|g|svc-forget"
	rememberLastGood(key, []*service.NodeConfig{{ID: "n1", URL: "http://10.0.0.1:80", Weight: 1, Health: true, Enabled: true}})
	forgetLastGood(key)
	if _, ok := recallLastGood(key); ok {
		t.Fatal("empty healthy list must drop last good")
	}
}

func TestLastGoodExpires(t *testing.T) {
	key := "t||ns|g|svc-expired"
	lastGood.Store(key, lastGoodEntry{
		nodes: []*service.NodeConfig{{ID: "old"}},
		at:    time.Now().Add(-2 * lastGoodTTL),
	})
	if _, ok := recallLastGood(key); ok {
		t.Fatal("expired last good must not be reused")
	}
}

func TestUseLastGoodOnError(t *testing.T) {
	if !useLastGoodOnError(contract.ErrViewNotReady) || !useLastGoodOnError(contract.ErrCenterNotRunning) {
		t.Fatal("restart/catch-up errors should reuse last good")
	}
	if useLastGoodOnError(errors.New("other")) {
		t.Fatal("unrelated errors must not reuse last good")
	}
	if useLastGoodOnError(contract.ErrServiceNotFound) {
		t.Fatal("service gone must not reuse last good")
	}
}

func TestConvertInstanceToNodeConfig(t *testing.T) {
	n := convertInstanceToNodeConfig(&model.Node{
		NodeID:    "n1",
		TenantID:      "t",
		NamespaceID:   "ns",
		GroupName:     "g",
		ServiceName:   "s",
		IP:            "10.0.0.1",
		Port:          8080,
		Weight:        2,
		Status:        model.NodeUP,
		HealthyStatus: model.Healthy,
		Metadata:      map[string]string{"contextPath": "/api"},
	}, "http")
	if n == nil || n.URL != "http://10.0.0.1:8080/api" || n.Weight != 2 || !n.Health {
		t.Fatalf("unexpected node %+v", n)
	}
}
