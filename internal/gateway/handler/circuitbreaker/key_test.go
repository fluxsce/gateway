package circuitbreaker

import "testing"

func TestNodeCircuitKey(t *testing.T) {
	if got := NodeCircuitKey("order-svc", "node-a"); got != "cb_node:order-svc:node-a" {
		t.Fatalf("NodeCircuitKey = %s", got)
	}
	if got := NodeCircuitKey("", "node-a"); got != "" {
		t.Fatalf("缺少服务 ID 应返回空键，实际 %s", got)
	}
	if got := NodeCircuitKey("order-svc", ""); got != "" {
		t.Fatalf("缺少节点 ID 应返回空键，实际 %s", got)
	}
	if NodeCircuitKey("svc-a", "n1") == NodeCircuitKey("svc-b", "n1") {
		t.Fatal("不同服务的同名节点不能共用熔断键")
	}
}
