package proxy

import (
	"context"
	"net/http"
	"time"

	"gateway/internal/gateway/core"
	"gateway/internal/gateway/handler/service"
)

// WebSocketUpgradeHandler 适配HTTP代理中的Upgrade入口，实际会话由WebSocketBridge统一管理。
type WebSocketUpgradeHandler struct {
	bridge *WebSocketBridge
}

// NewWebSocketUpgradeHandler 创建HTTP Upgrade适配器。
func NewWebSocketUpgradeHandler(serviceManager service.ServiceManager, config *WebSocketConfig) *WebSocketUpgradeHandler {
	return &WebSocketUpgradeHandler{
		bridge: NewWebSocketBridge(serviceManager, config),
	}
}

func newWebSocketUpgradeHandlerWithOverrides(serviceManager service.ServiceManager, overrides map[string]interface{}) *WebSocketUpgradeHandler {
	return &WebSocketUpgradeHandler{
		bridge: NewWebSocketBridgeWithOverrides(serviceManager, overrides),
	}
}

// InheritFromHTTPConfig 保留原有扩展点；WebSocket使用自己的滚动读写超时。
func (h *WebSocketUpgradeHandler) InheritFromHTTPConfig(_ *HTTPProxyConfig) {}

// IsWebSocketUpgrade 判断请求是否为WebSocket升级。
func (h *WebSocketUpgradeHandler) IsWebSocketUpgrade(req *http.Request) bool {
	return h.bridge.IsUpgradeRequest(req)
}

// HandleWebSocketUpgrade 通过共享Bridge代理WebSocket会话。
func (h *WebSocketUpgradeHandler) HandleWebSocketUpgrade(ctx *core.Context, proxyName, proxyType string) error {
	return h.bridge.Proxy(ctx, proxyName, proxyType)
}

// ShutdownContext 在给定排空期限内关闭全部WebSocket会话。
func (h *WebSocketUpgradeHandler) ShutdownContext(ctx context.Context) error {
	return h.bridge.Shutdown(ctx)
}

// Shutdown 保留原有按时长关闭接口，供兼容调用。
func (h *WebSocketUpgradeHandler) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return h.ShutdownContext(ctx)
}

// ForceClose 立即关闭全部WebSocket会话。
func (h *WebSocketUpgradeHandler) ForceClose() {
	h.bridge.ForceClose()
}

// GetStats 获取WebSocket统计快照。
func (h *WebSocketUpgradeHandler) GetStats() WebSocketStats {
	return h.bridge.GetStats()
}

// GetConnectionInfo 获取当前WebSocket会话信息。
func (h *WebSocketUpgradeHandler) GetConnectionInfo() []map[string]interface{} {
	return h.bridge.GetConnectionInfo()
}
