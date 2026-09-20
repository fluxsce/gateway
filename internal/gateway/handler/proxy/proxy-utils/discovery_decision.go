package proxyutils

import (
	"strconv"
	"strings"

	"gateway/internal/gateway/handler/service"
)

const (
	DiscoveryTypeInternal = "INTERNAL"
	DiscoveryTypeStatic   = "STATIC"
	maxDecisionLen        = 500
)

// DiscoveryDecision 一次选点结果，写入后端追踪 loadBalancerDecision。
// 跟这次转发走，不放共享 ctx，避免群发并行互相覆盖。
type DiscoveryDecision struct {
	Type       string
	Strategy   string
	NodeID     string
	NodeURL    string
	Candidates int
	LastGood   bool
	Reason     string
}

// CollectResult 服务中心一次拉名单的结果。LastGood 表示沿用了未过期的上次健康名单。
type CollectResult struct {
	Nodes    []*service.NodeConfig
	LastGood bool
	Reason   string
}

// LoadBalanceStrategy 取服务已配置的负载均衡策略。
func LoadBalanceStrategy(cfg *service.ServiceConfig) string {
	if cfg == nil {
		return ""
	}
	if cfg.Strategy != "" {
		return string(cfg.Strategy)
	}
	if cfg.LoadBalancer != nil && cfg.LoadBalancer.Strategy != "" {
		return string(cfg.LoadBalancer.Strategy)
	}
	return ""
}

// FillSelected 填入负载均衡选中的节点。
func FillSelected(d *DiscoveryDecision, node *service.NodeConfig) {
	if d == nil || node == nil {
		return
	}
	d.NodeID = node.ID
	d.NodeURL = node.URL
}

// Format 生成从表 loadBalancerDecision 的一行摘要，长度不超过列宽。
func (d DiscoveryDecision) Format() string {
	if d.Type == "" {
		return ""
	}
	var b strings.Builder
	b.WriteString(d.Type)
	if d.NodeID == "" && d.NodeURL == "" {
		b.WriteString(" fail")
	} else if d.LastGood {
		b.WriteString(" lastGood")
	}
	if d.NodeID != "" {
		b.WriteString(" node=")
		b.WriteString(d.NodeID)
	}
	if d.NodeURL != "" {
		b.WriteString(" url=")
		b.WriteString(d.NodeURL)
	}
	b.WriteString(" candidates=")
	b.WriteString(strconv.Itoa(d.Candidates))
	if d.Reason != "" {
		b.WriteString(" reason=")
		b.WriteString(d.Reason)
	}
	return truncateDecision(b.String())
}

func truncateDecision(s string) string {
	runes := []rune(s)
	if len(runes) <= maxDecisionLen {
		return s
	}
	return string(runes[:maxDecisionLen])
}
