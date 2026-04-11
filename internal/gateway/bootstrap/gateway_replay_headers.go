package bootstrap

import (
	"net/http"
	"strings"

	"gateway/internal/gateway/constants"
	"gateway/internal/gateway/core"
)

// applyGatewayReplayHeaders 从直连网关端口的请求中读取重发标识与原始 trace，写入 core.Context。
// X-Gateway-Replay 约定取值仅为 Y（大小写不敏感）；is_gateway_replay 写入字符串 Y。
// 读完后从 r.Header 删除对应头，避免透传到上游；应在 HandleWithContext 之前调用。
func applyGatewayReplayHeaders(r *http.Request, ctx *core.Context) {
	if r == nil || ctx == nil {
		return
	}
	trace := strings.TrimSpace(r.Header.Get(constants.HeaderXGatewayReplayTraceID))
	if trace != "" {
		ctx.Set(constants.ContextKeyPresetTraceID, trace)
	}
	if isGatewayReplayMarker(r.Header.Get(constants.HeaderXGatewayReplay)) {
		ctx.Set(constants.ContextKeyIsGatewayReplay, "Y")
	}
	r.Header.Del(constants.HeaderXGatewayReplay)
	r.Header.Del(constants.HeaderXGatewayReplayTraceID)
}

// isGatewayReplayMarker 判断 X-Gateway-Replay 是否为约定值 Y（大小写不敏感）。
func isGatewayReplayMarker(s string) bool {
	return strings.EqualFold(strings.TrimSpace(s), "Y")
}
