package proxyutils

import (
	"strings"
	"testing"

	"gateway/internal/gateway/handler/service"
)

func TestDiscoveryDecisionFormat(t *testing.T) {
	t.Parallel()
	got := DiscoveryDecision{
		Type:       DiscoveryTypeInternal,
		NodeID:     "n1",
		NodeURL:    "http://10.0.0.2:8080",
		Candidates: 3,
	}.Format()
	want := "INTERNAL node=n1 url=http://10.0.0.2:8080 candidates=3"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}

	got = DiscoveryDecision{
		Type:       DiscoveryTypeInternal,
		NodeID:     "n1",
		NodeURL:    "http://10.0.0.2:8080",
		Candidates: 2,
		LastGood:   true,
		Reason:     "服务中心未初始化",
	}.Format()
	want = "INTERNAL lastGood node=n1 url=http://10.0.0.2:8080 candidates=2 reason=服务中心未初始化"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}

	got = DiscoveryDecision{
		Type:       DiscoveryTypeStatic,
		Candidates: 0,
		Reason:     "未找到健康的服务节点",
	}.Format()
	want = "STATIC fail candidates=0 reason=未找到健康的服务节点"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestDiscoveryDecisionFormatTruncates(t *testing.T) {
	t.Parallel()
	got := DiscoveryDecision{
		Type:   DiscoveryTypeInternal,
		Reason: strings.Repeat("超", 600),
	}.Format()
	if len([]rune(got)) != maxDecisionLen {
		t.Fatalf("len=%d", len([]rune(got)))
	}
}

func TestLoadBalanceStrategy(t *testing.T) {
	t.Parallel()
	if got := LoadBalanceStrategy(nil); got != "" {
		t.Fatalf("nil = %q", got)
	}
	if got := LoadBalanceStrategy(&service.ServiceConfig{Strategy: service.IPHash}); got != "ip-hash" {
		t.Fatalf("strategy = %q", got)
	}
	if got := LoadBalanceStrategy(&service.ServiceConfig{
		LoadBalancer: &service.LoadBalancerConfig{Strategy: service.Random},
	}); got != "random" {
		t.Fatalf("lb strategy = %q", got)
	}
}
