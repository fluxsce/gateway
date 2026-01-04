package proxy

import (
	"fmt"
	"net/http"

	"gateway/internal/gateway/core"
	"gateway/internal/gateway/handler/service"
)

// TCPProxy TCP代理实现
type TCPProxy struct {
	*BaseProxyHandler
	serviceManager service.ServiceManager
	config         *TCPConfig
}

// Handle 处理TCP代理请求
func (t *TCPProxy) Handle(ctx *core.Context) bool {
	if !t.IsEnabled() {
		return true
	}

	// TCP代理通常不通过HTTP处理，这里返回错误
	ctx.AddError(fmt.Errorf("TCP代理不支持HTTP请求"))
	ctx.Abort(http.StatusBadRequest, map[string]string{
		"error": "TCP代理不支持HTTP请求",
	})
	return false
}

// proxyRequest 代理TCP请求到指定URL（内部方法）
func (t *TCPProxy) proxyRequest(ctx *core.Context, targetURL string) error {
	// TODO: 实现TCP代理逻辑
	// TCP代理需要在传输层进行，不是通过HTTP处理
	return fmt.Errorf("TCP代理功能尚未实现")
}

// GetTCPConfig 获取TCP代理配置
func (t *TCPProxy) GetTCPConfig() *TCPConfig {
	return t.config
}

// Validate 验证TCP代理配置
func (t *TCPProxy) Validate() error {
	config := t.GetTCPConfig()
	if config.BufferSize <= 0 {
		return fmt.Errorf("缓冲区大小必须大于0")
	}
	if config.ReadTimeout < 0 {
		return fmt.Errorf("读超时不能为负数")
	}
	if config.WriteTimeout < 0 {
		return fmt.Errorf("写超时不能为负数")
	}
	return nil
}

// Close 关闭TCP代理
func (t *TCPProxy) Close() error {
	var lastErr error

	// TODO: 关闭活跃的TCP连接
	// 这里应该实现关闭所有活跃TCP连接的逻辑

	// 关闭服务管理器
	// 服务管理器包含健康检查器等需要清理的资源
	if t.serviceManager != nil {
		if closer, ok := t.serviceManager.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				lastErr = err
			}
		}
	}

	return lastErr
}

// NewTCPProxy 创建TCP代理
func NewTCPProxy(config ProxyConfig, serviceManager service.ServiceManager) (*TCPProxy, error) {
	// 解析TCP特定配置
	tcpConfig := DefaultTCPConfig
	if config.Config != nil {
		parseTCPConfig(config.Config, &tcpConfig)
	}

	return &TCPProxy{
		BaseProxyHandler: NewBaseProxyHandler(config.Type, config.Enabled, config.Name),
		serviceManager:   serviceManager,
		config:           &tcpConfig,
	}, nil
}

// parseTCPConfig 解析TCP配置
func parseTCPConfig(configMap map[string]interface{}, tcpConfig *TCPConfig) {
	if enabled, ok := configMap["enabled"]; ok {
		if b, ok := enabled.(bool); ok {
			tcpConfig.Enabled = b
		}
	}

	if keepAlive, ok := configMap["keep_alive"]; ok {
		if b, ok := keepAlive.(bool); ok {
			tcpConfig.KeepAlive = b
		}
	}

	if bufferSize, ok := configMap["buffer_size"]; ok {
		if size, ok := bufferSize.(int); ok {
			tcpConfig.BufferSize = size
		}
	}
}
