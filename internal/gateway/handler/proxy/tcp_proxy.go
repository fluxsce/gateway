package proxy

import (
	"fmt"
	"net/http"

	"gohub/internal/gateway/core"
	"gohub/internal/gateway/handler/service"
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

// ProxyRequest 代理TCP请求到指定URL
func (t *TCPProxy) ProxyRequest(ctx *core.Context, targetURL string) error {
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
	// TODO: 关闭活跃的TCP连接
	return nil
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
