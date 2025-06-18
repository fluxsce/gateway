package assertion

import (
	"gohub/internal/gateway/core"
	"net"
	"strings"
)

// IPAsserter IP地址断言器
// 根据客户端IP地址进行断言
type IPAsserter struct {
	BaseAssertion
}

// IPAsserterFromConfig 从配置创建IP地址断言器
func IPAsserterFromConfig(config AssertionConfig, operator ComparisonOperator) (Assertion, error) {
	return &IPAsserter{
		BaseAssertion: BaseAssertion{
			Type:          IPAssertion,
			FieldName:     "ip",
			ExpectedValue: config.Value,
			Operator:      operator,
			CaseSensitive: false,
			Description:   config.Description,
			Config:        config,
		},
	}, nil
}

// Evaluate 实现Assertion接口
func (a *IPAsserter) Evaluate(ctx *core.Context) (bool, error) {
	// 获取客户端IP
	clientIP := getClientIP(ctx)

	// 应用比较规则
	return a.compare(clientIP, a.ExpectedValue), nil
}

// getClientIP 获取客户端真实IP地址
// 优先级：X-Forwarded-For > X-Real-IP > RemoteAddr
func getClientIP(ctx *core.Context) string {
	// 检查 X-Forwarded-For 头部
	forwarded := ctx.Request.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// X-Forwarded-For 可能包含多个IP，取第一个
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if ip != "" {
				return ip
			}
		}
	}

	// 检查 X-Real-IP 头部
	realIP := ctx.Request.Header.Get("X-Real-IP")
	if realIP != "" {
		return strings.TrimSpace(realIP)
	}

	// 使用 RemoteAddr
	if ctx.Request.RemoteAddr != "" {
		host, _, err := net.SplitHostPort(ctx.Request.RemoteAddr)
		if err != nil {
			return ctx.Request.RemoteAddr
		}
		return host
	}

	return ""
}
