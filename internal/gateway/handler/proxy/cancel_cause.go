package proxy

import (
	"context"
	"errors"

	"gateway/internal/gateway/constants"
)

// proxyCancelStatus 判断出站取消该记成哪个网关状态码。
// 调用方断开或超时会先取消进来的请求，记 499。代理总超时只取消出站 Context，记 504。
// 其它错误返回 0，由调用方按 502 处理。实例读超时关掉进来的连接时，也会得到 499。
func proxyCancelStatus(incoming, outbound context.Context, err error) int {
	if err == nil {
		return 0
	}
	if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		return 0
	}
	if incoming != nil && incoming.Err() != nil {
		return constants.GatewayStatusClientClosed
	}
	if outbound != nil && outbound.Err() != nil {
		return constants.GatewayStatusGatewayTimeout
	}
	return 0
}

func proxyFailureSummary(status int) string {
	switch status {
	case constants.GatewayStatusClientClosed:
		return "client closed request"
	case constants.GatewayStatusGatewayTimeout:
		return "gateway timeout"
	default:
		return "proxy request failed"
	}
}
