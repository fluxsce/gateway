package proxy

import (
	"context"
	"fmt"
	"net/http"

	"gateway/internal/gateway/core"
	"gateway/internal/gateway/handler/service"
)

// WebSocketProxy 是专项WebSocket代理入口，会话处理统一委托给WebSocketBridge。
type WebSocketProxy struct {
	*BaseProxyHandler
	serviceManager service.ServiceManager
	config         *WebSocketConfig
	bridge         *WebSocketBridge
}

// Handle 校验专项入口并通过共享Bridge代理WebSocket会话。
func (w *WebSocketProxy) Handle(ctx *core.Context) bool {
	if !w.IsEnabled() {
		return true
	}
	if !w.bridge.IsUpgradeRequest(ctx.Request) {
		ctx.AddError(fmt.Errorf("不是WebSocket升级请求"))
		ctx.Abort(http.StatusBadRequest, map[string]string{
			"error": "not a websocket upgrade request",
		})
		return false
	}
	// enableWebsocket=N 仍允许升级，与历史行为保持兼容；该字段仅作路由标记，不作为准入开关。
	if err := w.bridge.Proxy(ctx, w.GetName(), string(ProxyTypeWebSocket)); err != nil {
		ctx.AddError(fmt.Errorf("代理WebSocket请求失败: %w", err))
		if !ctx.IsResponded() {
			ctx.Abort(http.StatusBadGateway, map[string]string{
				"error": "websocket proxy failed",
			})
		}
		return false
	}
	return true
}

// GetWebSocketConfig 获取专项WebSocket配置副本。
func (w *WebSocketProxy) GetWebSocketConfig() *WebSocketConfig {
	config := *w.config
	return &config
}

// Validate 验证WebSocket配置。
func (w *WebSocketProxy) Validate() error {
	config := w.config
	if config.PingInterval < 0 {
		return fmt.Errorf("Ping间隔不能为负数")
	}
	if config.PongTimeout < 0 {
		return fmt.Errorf("Pong超时不能为负数")
	}
	if config.WriteTimeout < 0 || config.ReadTimeout < 0 {
		return fmt.Errorf("WebSocket读写超时不能为负数")
	}
	if config.MaxMessageSize < 0 {
		return fmt.Errorf("最大消息大小不能为负数")
	}
	if config.ReadBufferSize <= 0 || config.WriteBufferSize <= 0 {
		return fmt.Errorf("WebSocket缓冲区大小必须大于0")
	}
	return nil
}

// Shutdown 在给定期限内优雅关闭全部WebSocket会话。
func (w *WebSocketProxy) Shutdown(ctx context.Context) error {
	return w.bridge.Shutdown(ctx)
}

// ForceClose 立即关闭全部WebSocket会话。
func (w *WebSocketProxy) ForceClose() {
	w.bridge.ForceClose()
}

// Close 强制释放会话和服务管理器资源。
func (w *WebSocketProxy) Close() error {
	w.bridge.ForceClose()
	if closer, ok := w.serviceManager.(interface{ Close() error }); ok {
		return closer.Close()
	}
	return nil
}

// GetConnectionCount 获取当前WebSocket连接数。
func (w *WebSocketProxy) GetConnectionCount() int {
	return int(w.bridge.GetStats().ActiveConnections)
}

// GetConnectionStats 获取当前WebSocket统计和会话信息。
func (w *WebSocketProxy) GetConnectionStats() map[string]interface{} {
	stats := w.bridge.GetStats()
	return map[string]interface{}{
		"active_connections": stats.ActiveConnections,
		"total_connections":  stats.TotalConnections,
		"failed_upgrades":    stats.FailedUpgrades,
		"bytes_received":     stats.BytesReceived,
		"bytes_sent":         stats.BytesSent,
		"normal_closed":      stats.NormalClosed,
		"shutdown_closed":    stats.ShutdownClosed,
		"forced_closed":      stats.ForcedClosed,
		"error_closed":       stats.ErrorClosed,
		"connections":        w.bridge.GetConnectionInfo(),
	}
}

// NewWebSocketProxy 创建使用共享Bridge的专项WebSocket代理。
func NewWebSocketProxy(config ProxyConfig, serviceManager service.ServiceManager) (*WebSocketProxy, error) {
	wsConfig := DefaultWebSocketConfig
	if config.Config != nil {
		NewWebSocketConfigParser().ParseConfig(config.Config, &wsConfig)
	}
	proxy := &WebSocketProxy{
		BaseProxyHandler: NewBaseProxyHandler(config.Type, config.Enabled, config.Name),
		serviceManager:   serviceManager,
		config:           &wsConfig,
		bridge:           NewWebSocketBridgeWithOverrides(serviceManager, config.Config),
	}
	if err := proxy.Validate(); err != nil {
		return nil, err
	}
	return proxy, nil
}
