package proxy

import (
	"fmt"

	"gohub/internal/gateway/handler/service"
)

// ProxyFactory 代理工厂
type ProxyFactory struct {
	serviceManager service.ServiceManager
}

// NewProxyFactory 创建代理工厂
func NewProxyFactory(serviceManager service.ServiceManager) *ProxyFactory {
	return &ProxyFactory{
		serviceManager: serviceManager,
	}
}

// CreateProxy 根据配置创建代理
func (f *ProxyFactory) CreateProxy(config ProxyConfig) (ProxyHandler, error) {
	if !config.Enabled {
		return nil, fmt.Errorf("代理未启用")
	}

	switch config.Type {
	case ProxyTypeHTTP:
		return f.createHTTPProxy(config)
	case ProxyTypeWebSocket:
		return f.createWebSocketProxy(config)
	case ProxyTypeTCP:
		return f.createTCPProxy(config)
	case ProxyTypeUDP:
		return f.createUDPProxy(config)
	default:
		return nil, fmt.Errorf("不支持的代理类型: %s", config.Type)
	}
}

// CreateProxyByType 根据代理类型创建默认代理
func (f *ProxyFactory) CreateProxyByType(proxyType ProxyType) (ProxyHandler, error) {
	config := ProxyConfig{
		Enabled: true,
		Type:    proxyType,
		Name:    fmt.Sprintf("Default %s Proxy", proxyType),
		Config:  make(map[string]interface{}),
	}

	return f.CreateProxy(config)
}

// createHTTPProxy 创建HTTP代理
func (f *ProxyFactory) createHTTPProxy(config ProxyConfig) (ProxyHandler, error) {
	return NewHTTPProxy(config, f.serviceManager)
}

// createWebSocketProxy 创建WebSocket代理
func (f *ProxyFactory) createWebSocketProxy(config ProxyConfig) (ProxyHandler, error) {
	return NewWebSocketProxy(config, f.serviceManager)
}

// createTCPProxy 创建TCP代理
func (f *ProxyFactory) createTCPProxy(config ProxyConfig) (ProxyHandler, error) {
	return NewTCPProxy(config, f.serviceManager)
}

// createUDPProxy 创建UDP代理
func (f *ProxyFactory) createUDPProxy(config ProxyConfig) (ProxyHandler, error) {
	return NewUDPProxy(config, f.serviceManager)
}

// GetSupportedTypes 获取支持的代理类型列表
func (f *ProxyFactory) GetSupportedTypes() []ProxyType {
	return []ProxyType{
		ProxyTypeHTTP,
		ProxyTypeWebSocket,
		ProxyTypeTCP,
		ProxyTypeUDP,
	}
}

// GetTypeDescription 获取代理类型描述
func (f *ProxyFactory) GetTypeDescription(proxyType ProxyType) string {
	descriptions := map[ProxyType]string{
		ProxyTypeHTTP:      "HTTP代理，支持HTTP/HTTPS协议",
		ProxyTypeWebSocket: "WebSocket代理，支持WebSocket协议",
		ProxyTypeTCP:       "TCP代理，支持TCP协议",
		ProxyTypeUDP:       "UDP代理，支持UDP协议",
	}

	if desc, exists := descriptions[proxyType]; exists {
		return desc
	}
	return "未知代理类型"
}

// ValidateConfig 验证配置
func (f *ProxyFactory) ValidateConfig(config ProxyConfig) error {
	if config.Name == "" {
		return fmt.Errorf("代理名称不能为空")
	}

	// 验证代理类型
	validTypes := f.GetSupportedTypes()
	typeValid := false
	for _, t := range validTypes {
		if config.Type == t {
			typeValid = true
			break
		}
	}
	if !typeValid {
		return fmt.Errorf("不支持的代理类型: %s", config.Type)
	}

	return nil
}
