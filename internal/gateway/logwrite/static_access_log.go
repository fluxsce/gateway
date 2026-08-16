package logwrite

import (
	"gateway/internal/gateway/constants"
	"gateway/internal/gateway/core"
	"gateway/internal/gateway/logwrite/types"
)

const staticProxyType = "static"

// shouldSkipStaticAccessLog 判断是否跳过静态托管访问日志。
// 成功命中（file/index/spa/redirect，状态码 < 400）不落库，避免一个页面带出几十条 js/css 日志。
// 4xx/5xx 仍写入，便于排查缺文件、越权路径和磁盘异常。
func shouldSkipStaticAccessLog(gatewayCtx *core.Context) bool {
	if gatewayCtx == nil {
		return false
	}
	if !isStaticGatewayContext(gatewayCtx) {
		return false
	}
	status := constants.GatewayStatusOK
	if code, exists := gatewayCtx.GetInt(constants.GatewayStatusCode); exists {
		status = code
	}
	return status < 400
}

// isStaticGatewayContext 判断当前请求已由静态托管处理。
func isStaticGatewayContext(gatewayCtx *core.Context) bool {
	if proxyType, ok := gatewayCtx.GetString(constants.ContextKeyProxyType); ok && proxyType == staticProxyType {
		return true
	}
	result, ok := gatewayCtx.GetString(constants.ContextKeyStaticResult)
	return ok && result != ""
}

// isStaticAccessLog 判断访问日志来自静态托管。
func isStaticAccessLog(accessLog *types.AccessLog) bool {
	return accessLog != nil && accessLog.ProxyType == staticProxyType
}

// shouldAlertStaticAccessLog 静态托管只对 5xx 做状态码告警。
// 4xx（缺 favicon、爬虫探测、隐藏路径）和读大文件超时是常态，按请求告警会刷屏。
func shouldAlertStaticAccessLog(accessLog *types.AccessLog) bool {
	return isStaticAccessLog(accessLog) && accessLog.GatewayStatusCode >= 500
}
