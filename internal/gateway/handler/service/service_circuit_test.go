package service

import (
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"gateway/internal/gateway/core"
	"gateway/internal/gateway/handler/circuitbreaker"
)

// TestResolveServiceCircuitBreakerConfig 验证完整配置优先于 enableCircuitBreaker 开关。
func TestResolveServiceCircuitBreakerConfig(t *testing.T) {
	if resolveServiceCircuitBreakerConfig(nil) != nil {
		t.Fatal("空配置应返回 nil")
	}
	if resolveServiceCircuitBreakerConfig(&ServiceConfig{}) != nil {
		t.Fatal("未启用熔断应返回 nil")
	}

	full := circuitbreaker.DefaultCircuitBreakerConfig()
	full.MinimumRequests = 2
	got := resolveServiceCircuitBreakerConfig(&ServiceConfig{CircuitBreaker: full})
	if got == nil || got.MinimumRequests != 2 {
		t.Fatalf("应优先使用完整熔断配置: %+v", got)
	}

	fromFlag := resolveServiceCircuitBreakerConfig(&ServiceConfig{
		LoadBalancer: &LoadBalancerConfig{CircuitBreaker: true},
	})
	if fromFlag == nil || !fromFlag.Enabled {
		t.Fatalf("enableCircuitBreaker 应启用节点熔断: %+v", fromFlag)
	}
}

// TestNodeEjectionSkipsOpenNode 验证节点开闸后只跳过该实例，其他健康节点仍可选。
func TestNodeEjectionSkipsOpenNode(t *testing.T) {
	cb := circuitbreaker.DefaultCircuitBreakerConfig()
	cb.MinimumRequests = 2
	cb.ErrorRatePercent = 50
	cb.OpenTimeoutSeconds = 60
	cb.StorageType = "memory"

	svc, err := NewService(&ServiceConfig{
		ID:             "svc-node-eject",
		Name:           "node-eject",
		Strategy:       RoundRobin,
		CircuitBreaker: cb,
		Nodes: []*NodeConfig{
			{ID: "n1", URL: "http://127.0.0.1:9", Weight: 1, Health: true, Enabled: true},
			{ID: "n2", URL: "http://127.0.0.1:10", Weight: 1, Health: true, Enabled: true},
		},
	}, true)
	if err != nil {
		t.Fatalf("创建服务失败: %v", err)
	}

	for i := 0; i < 2; i++ {
		svc.RecordNodeCircuitResult("n1", false, 10*time.Millisecond, errString("node a failed"))
	}

	req := httptest.NewRequest("GET", "http://gateway/api", nil)
	ctx := core.NewContext(httptest.NewRecorder(), req)
	for i := 0; i < 4; i++ {
		node, selErr := svc.SelectNode(ctx)
		if selErr != nil {
			t.Fatalf("应仍能选到其他节点: %v", selErr)
		}
		if node.ID != "n2" {
			t.Fatalf("应跳过已开闸的 n1，实际选中 %s", node.ID)
		}
	}
}

// TestEjectedNodesFallbackToHealthy 验证健康节点全部被摘除时回退健康列表，避免和“无节点可重试”混淆。
func TestEjectedNodesFallbackToHealthy(t *testing.T) {
	cb := circuitbreaker.DefaultCircuitBreakerConfig()
	cb.MinimumRequests = 2
	cb.ErrorRatePercent = 50
	cb.OpenTimeoutSeconds = 60
	cb.StorageType = "memory"

	svc, err := NewService(&ServiceConfig{
		ID:             "svc-node-fallback",
		Name:           "node-fallback",
		Strategy:       RoundRobin,
		CircuitBreaker: cb,
		Nodes: []*NodeConfig{
			{ID: "n1", URL: "http://127.0.0.1:9", Weight: 1, Health: true, Enabled: true},
			{ID: "n2", URL: "http://127.0.0.1:10", Weight: 1, Health: true, Enabled: true},
		},
	}, true)
	if err != nil {
		t.Fatalf("创建服务失败: %v", err)
	}

	for i := 0; i < 2; i++ {
		svc.RecordNodeCircuitResult("n1", false, 10*time.Millisecond, errString("node a failed"))
		svc.RecordNodeCircuitResult("n2", false, 10*time.Millisecond, errString("node b failed"))
	}

	req := httptest.NewRequest("GET", "http://gateway/api", nil)
	ctx := core.NewContext(httptest.NewRecorder(), req)
	node, selErr := svc.SelectNode(ctx)
	if selErr != nil || node == nil {
		t.Fatalf("全部摘除时应回退健康节点，而不是无节点: %v", selErr)
	}
}

func TestConcurrentSelectSkipsOpenNode(t *testing.T) {
	cb := circuitbreaker.DefaultCircuitBreakerConfig()
	cb.MinimumRequests = 2
	cb.ErrorRatePercent = 50
	cb.OpenTimeoutSeconds = 60
	cb.StorageType = "memory"

	svc, err := NewService(&ServiceConfig{
		ID:             "svc-node-conc",
		Name:           "node-conc",
		Strategy:       RoundRobin,
		CircuitBreaker: cb,
		Nodes: []*NodeConfig{
			{ID: "n1", URL: "http://127.0.0.1:9", Weight: 1, Health: true, Enabled: true},
			{ID: "n2", URL: "http://127.0.0.1:10", Weight: 1, Health: true, Enabled: true},
		},
	}, true)
	if err != nil {
		t.Fatalf("创建服务失败: %v", err)
	}
	svc.RecordNodeCircuitResult("n1", false, 10*time.Millisecond, errString("node a failed"))
	svc.RecordNodeCircuitResult("n1", false, 10*time.Millisecond, errString("node a failed"))

	req := httptest.NewRequest("GET", "http://gateway/api", nil)
	ctx := core.NewContext(httptest.NewRecorder(), req)

	const n = 32
	errCh := make(chan error, n)
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			node, selErr := svc.SelectNode(ctx)
			if selErr != nil {
				errCh <- selErr
				return
			}
			if node.ID != "n2" {
				errCh <- errString("选中了已开闸节点 " + node.ID)
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for selErr := range errCh {
		t.Fatal(selErr)
	}
}

func TestUnhealthyNodeNotSelectedEvenWhenCircuitClosed(t *testing.T) {
	cb := circuitbreaker.DefaultCircuitBreakerConfig()
	cb.StorageType = "memory"
	svc, err := NewService(&ServiceConfig{
		ID:             "svc-node-health",
		Name:           "node-health",
		Strategy:       RoundRobin,
		CircuitBreaker: cb,
		Nodes: []*NodeConfig{
			{ID: "n1", URL: "http://127.0.0.1:9", Weight: 1, Health: false, Enabled: true},
			{ID: "n2", URL: "http://127.0.0.1:10", Weight: 1, Health: true, Enabled: true},
		},
	}, true)
	if err != nil {
		t.Fatalf("创建服务失败: %v", err)
	}
	req := httptest.NewRequest("GET", "http://gateway/api", nil)
	ctx := core.NewContext(httptest.NewRecorder(), req)
	node, selErr := svc.SelectNode(ctx)
	if selErr != nil || node == nil || node.ID != "n2" {
		t.Fatalf("不健康节点即使未熔断也不应入选: node=%v err=%v", node, selErr)
	}
}

type errString string

func (e errString) Error() string { return string(e) }
