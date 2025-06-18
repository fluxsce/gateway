package proxy

import (
	"fmt"
	"net/http"

	"gohub/internal/gateway/core"
	"gohub/internal/gateway/handler/service"
)

// UDPProxy UDP代理实现
type UDPProxy struct {
	*BaseProxyHandler
	serviceManager service.ServiceManager
}

// Handle 处理UDP代理请求
func (u *UDPProxy) Handle(ctx *core.Context) bool {
	if !u.IsEnabled() {
		return true
	}

	// UDP代理通常不通过HTTP处理，这里返回错误
	ctx.AddError(fmt.Errorf("UDP代理不支持HTTP请求"))
	ctx.Abort(http.StatusBadRequest, map[string]string{
		"error": "UDP代理不支持HTTP请求",
	})
	return false
}

// ProxyRequest 代理UDP请求到指定URL
func (u *UDPProxy) ProxyRequest(ctx *core.Context, targetURL string) error {
	// TODO: 实现UDP代理逻辑
	// UDP代理需要在传输层进行，不是通过HTTP处理
	return fmt.Errorf("UDP代理功能尚未实现")
}

// Validate 验证UDP代理配置
func (u *UDPProxy) Validate() error {
	// UDP代理的基本验证
	return nil
}

// Close 关闭UDP代理
func (u *UDPProxy) Close() error {
	// TODO: 关闭活跃的UDP连接
	return nil
}

// NewUDPProxy 创建UDP代理
func NewUDPProxy(config ProxyConfig, serviceManager service.ServiceManager) (*UDPProxy, error) {
	// TODO: 解析UDP特定配置
	// udpConfig := DefaultUDPConfig
	// if config.Config != nil {
	//     parseUDPConfig(config.Config, &udpConfig)
	// }

	return &UDPProxy{
		BaseProxyHandler: NewBaseProxyHandler(config.Type, config.Enabled, config.Name),
		serviceManager:   serviceManager,
	}, nil
}
