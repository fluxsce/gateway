package proxy

import (
	"fmt"
	"net/http"
	"time"

	"gohub/internal/gateway/core"
	"gohub/internal/gateway/handler/service"
)

// WebSocketProxy WebSocket代理实现
type WebSocketProxy struct {
	*BaseProxyHandler
	serviceManager service.ServiceManager
	config         *WebSocketConfig
}

// Handle 处理WebSocket代理请求
func (w *WebSocketProxy) Handle(ctx *core.Context) bool {
	if !w.IsEnabled() {
		return true
	}

	// 检查是否为WebSocket升级请求
	if !w.isWebSocketUpgrade(ctx.Request) {
		ctx.AddError(fmt.Errorf("不是WebSocket升级请求"))
		ctx.Abort(http.StatusBadRequest, map[string]string{
			"error": "不是WebSocket升级请求",
		})
		return false
	}

	// 获取服务ID
	serviceID := ctx.GetServiceID()
	if serviceID == "" {
		ctx.AddError(fmt.Errorf("服务ID不能为空"))
		ctx.Abort(http.StatusBadRequest, map[string]string{
			"error": "服务ID不能为空",
		})
		return false
	}

	// 从负载均衡器获取目标节点
	node, err := w.serviceManager.SelectNode(serviceID, ctx)
	if err != nil {
		ctx.AddError(fmt.Errorf("选择目标节点失败: %w", err))
		ctx.Abort(http.StatusServiceUnavailable, map[string]string{
			"error": "服务不可用",
		})
		return false
	}

	// 代理WebSocket请求
	err = w.ProxyRequest(ctx, node.URL)
	if err != nil {
		ctx.AddError(fmt.Errorf("代理WebSocket请求失败: %w", err))
		ctx.Abort(http.StatusBadGateway, map[string]string{
			"error": "代理WebSocket请求失败",
		})
		return false
	}

	return true
}

// ProxyRequest 代理WebSocket请求到指定URL
func (w *WebSocketProxy) ProxyRequest(ctx *core.Context, targetURL string) error {
	// TODO: 实现WebSocket代理逻辑
	// 这里需要实现WebSocket的升级和转发逻辑
	// 包括连接升级、消息转发、连接管理等
	return fmt.Errorf("WebSocket代理功能尚未实现")
}

// isWebSocketUpgrade 检查是否为WebSocket升级请求
func (w *WebSocketProxy) isWebSocketUpgrade(req *http.Request) bool {
	// 检查必要的头部
	if req.Header.Get("Connection") != "Upgrade" {
		return false
	}
	if req.Header.Get("Upgrade") != "websocket" {
		return false
	}
	if req.Header.Get("Sec-WebSocket-Key") == "" {
		return false
	}
	if req.Header.Get("Sec-WebSocket-Version") != "13" {
		return false
	}
	return true
}

// GetWebSocketConfig 获取WebSocket代理配置
func (w *WebSocketProxy) GetWebSocketConfig() *WebSocketConfig {
	return w.config
}

// Validate 验证WebSocket代理配置
func (w *WebSocketProxy) Validate() error {
	wsConfig := w.GetWebSocketConfig()
	if wsConfig.PingInterval < 0 {
		return fmt.Errorf("Ping间隔不能为负数")
	}
	if wsConfig.PongTimeout < 0 {
		return fmt.Errorf("Pong超时不能为负数")
	}
	if wsConfig.WriteTimeout < 0 {
		return fmt.Errorf("写超时不能为负数")
	}
	if wsConfig.ReadTimeout < 0 {
		return fmt.Errorf("读超时不能为负数")
	}
	if wsConfig.MaxMessageSize < 0 {
		return fmt.Errorf("最大消息大小不能为负数")
	}

	return nil
}

// Close 关闭WebSocket代理
func (w *WebSocketProxy) Close() error {
	// TODO: 关闭活跃的WebSocket连接
	return nil
}

// NewWebSocketProxy 创建WebSocket代理
func NewWebSocketProxy(config ProxyConfig, serviceManager service.ServiceManager) (*WebSocketProxy, error) {
	// 解析WebSocket特定配置
	wsConfig := DefaultWebSocketConfig
	if config.Config != nil {
		parseWebSocketConfig(config.Config, &wsConfig)
	}

	return &WebSocketProxy{
		BaseProxyHandler: NewBaseProxyHandler(config.Type, config.Enabled, config.Name),
		serviceManager:   serviceManager,
		config:           &wsConfig,
	}, nil
}

// parseWebSocketConfig 解析WebSocket配置
func parseWebSocketConfig(configMap map[string]interface{}, wsConfig *WebSocketConfig) {
	if pingInterval, ok := configMap["ping_interval"]; ok {
		if intervalStr, ok := pingInterval.(string); ok {
			if d, err := time.ParseDuration(intervalStr); err == nil {
				wsConfig.PingInterval = d
			}
		}
	}

	if maxMessageSize, ok := configMap["max_message_size"]; ok {
		if size, ok := maxMessageSize.(int64); ok {
			wsConfig.MaxMessageSize = size
		}
	}

	if enabled, ok := configMap["enabled"]; ok {
		if b, ok := enabled.(bool); ok {
			wsConfig.Enabled = b
		}
	}
}
