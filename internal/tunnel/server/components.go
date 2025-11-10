// Package server 提供服务端组件工厂方法
// 组件工厂负责创建和配置各种服务端组件实例
package server

import (
	"gateway/internal/tunnel/storage"
	"gateway/pkg/logger"
)

// ComponentFactory 组件工厂接口
// 提供统一的组件创建和管理接口
type ComponentFactory interface {
	// CreateControlServer 创建控制服务器实例
	CreateControlServer(tunnelServer TunnelServer) ControlServer

	// CreateProxyServer 创建反向代理服务器实例（隧道服务）
	CreateProxyServer(tunnelServer TunnelServer, controlServer ControlServer) ProxyServer

	// CreateServiceRegistry 创建服务注册器实例
	CreateServiceRegistry(storage storage.RepositoryManager) ServiceRegistry

	// CreateLoadBalancer 创建负载均衡器实例
	CreateLoadBalancer(algorithm string) LoadBalancer
}

// defaultComponentFactory 默认组件工厂实现
type defaultComponentFactory struct{}

// NewComponentFactory 创建默认组件工厂实例
//
// 返回:
//   - ComponentFactory: 组件工厂接口实例
//
// 功能:
//   - 创建默认的组件工厂
//   - 用于统一管理所有服务端组件的创建
func NewComponentFactory() ComponentFactory {
	logger.Info("Component factory initialized", nil)
	return &defaultComponentFactory{}
}

// CreateControlServer 创建控制服务器实例
//
// 参数:
//   - tunnelServer: 隧道服务器实例
//
// 返回:
//   - ControlServer: 控制服务器接口实例
//
// 功能:
//   - 创建完整的控制服务器实现
//   - 处理客户端控制连接和认证
func (f *defaultComponentFactory) CreateControlServer(tunnelServer TunnelServer) ControlServer {
	logger.Debug("Creating control server", nil)
	return NewControlServerImpl(tunnelServer)
}

// CreateProxyServer 创建反向代理服务器实例（隧道服务）
//
// 参数:
//   - tunnelServer: 隧道服务器实例
//   - controlServer: 控制服务器实例
//
// 返回:
//   - ProxyServer: 反向代理服务器接口实例
//
// 功能:
//   - 创建完整的反向代理服务器实现
//   - 支持隧道反向代理转发
func (f *defaultComponentFactory) CreateProxyServer(tunnelServer TunnelServer, controlServer ControlServer) ProxyServer {
	logger.Debug("Creating reverse proxy server", nil)
	return NewProxyServerImpl(tunnelServer, controlServer)
}

// CreateServiceRegistry 创建服务注册器实例
//
// 参数:
//   - storage: 存储管理器实例
//
// 返回:
//   - ServiceRegistry: 服务注册器接口实例
//
// 功能:
//   - 创建完整的服务注册器实现
//   - 管理服务注册和端口分配
func (f *defaultComponentFactory) CreateServiceRegistry(storage storage.RepositoryManager) ServiceRegistry {
	logger.Debug("Creating service registry", nil)
	return NewServiceRegistryImpl(storage)
}

// CreateLoadBalancer 创建负载均衡器实例
//
// 参数:
//   - algorithm: 负载均衡算法名称
//
// 返回:
//   - LoadBalancer: 负载均衡器接口实例
//
// 功能:
//   - 创建完整的负载均衡器实现
//   - 支持多种负载均衡算法
func (f *defaultComponentFactory) CreateLoadBalancer(algorithm string) LoadBalancer {
	logger.Debug("Creating load balancer", map[string]interface{}{
		"algorithm": algorithm,
	})
	return NewLoadBalancerImpl(algorithm)
}
