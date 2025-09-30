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

	// CreateForwardProxyServer 创建正向代理服务器实例
	CreateForwardProxyServer(tunnelServer TunnelServer) ForwardProxyServer

	// CreateSessionManager 创建会话管理器实例
	CreateSessionManager(storage storage.RepositoryManager) SessionManager

	// CreateServiceRegistry 创建服务注册器实例
	CreateServiceRegistry(storage storage.RepositoryManager) ServiceRegistry

	// CreateConnectionTracker 创建连接跟踪器实例
	CreateConnectionTracker(storage storage.RepositoryManager) ConnectionTracker

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
	return NewControlServer(tunnelServer)
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
	return NewProxyServer(tunnelServer, controlServer)
}

// CreateForwardProxyServer 创建正向代理服务器实例
//
// 参数:
//   - tunnelServer: 隧道服务器实例
//
// 返回:
//   - ForwardProxyServer: 正向代理服务器接口实例
//
// 功能:
//   - 创建完整的正向代理服务器实现
//   - 支持客户端到外部服务器的代理转发
func (f *defaultComponentFactory) CreateForwardProxyServer(tunnelServer TunnelServer) ForwardProxyServer {
	logger.Debug("Creating forward proxy server", nil)
	return NewForwardProxyServer(tunnelServer)
}

// CreateSessionManager 创建会话管理器实例
//
// 参数:
//   - storage: 存储管理器实例
//
// 返回:
//   - SessionManager: 会话管理器接口实例
//
// 功能:
//   - 创建完整的会话管理器实现
//   - 管理客户端会话生命周期
func (f *defaultComponentFactory) CreateSessionManager(storage storage.RepositoryManager) SessionManager {
	logger.Debug("Creating session manager", nil)
	return NewSessionManager(storage)
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
	return NewServiceRegistry(storage)
}

// CreateConnectionTracker 创建连接跟踪器实例
//
// 参数:
//   - storage: 存储管理器实例
//
// 返回:
//   - ConnectionTracker: 连接跟踪器接口实例
//
// 功能:
//   - 创建完整的连接跟踪器实现
//   - 提供连接监控和统计功能
func (f *defaultComponentFactory) CreateConnectionTracker(storage storage.RepositoryManager) ConnectionTracker {
	logger.Debug("Creating connection tracker", nil)
	return NewConnectionTracker(storage)
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
	return NewLoadBalancer(algorithm)
}

// 兼容性工厂方法 - 保持向后兼容

// NewControlServer 创建控制服务器实例（兼容性方法）
//
// 参数:
//   - tunnelServer: 隧道服务器实例
//
// 返回:
//   - ControlServer: 控制服务器接口实例
//
// 功能:
//   - 提供向后兼容的工厂方法
//   - 直接创建控制服务器实例
func NewControlServer(tunnelServer TunnelServer) ControlServer {
	return NewControlServerImpl(tunnelServer)
}

// NewProxyServer 创建反向代理服务器实例（兼容性方法）
//
// 参数:
//   - tunnelServer: 隧道服务器实例
//   - controlServer: 控制服务器实例
//
// 返回:
//   - ProxyServer: 反向代理服务器接口实例
//
// 功能:
//   - 提供向后兼容的工厂方法
//   - 直接创建反向代理服务器实例
func NewProxyServer(tunnelServer TunnelServer, controlServer ControlServer) ProxyServer {
	return NewProxyServerImpl(tunnelServer, controlServer)
}

// NewForwardProxyServer 创建正向代理服务器实例（兼容性方法）
//
// 参数:
//   - tunnelServer: 隧道服务器实例
//
// 返回:
//   - ForwardProxyServer: 正向代理服务器接口实例
//
// 功能:
//   - 提供向后兼容的工厂方法
//   - 直接创建正向代理服务器实例
func NewForwardProxyServer(tunnelServer TunnelServer) ForwardProxyServer {
	return NewForwardProxyServerImpl(tunnelServer)
}

// NewSessionManager 创建会话管理器实例（兼容性方法）
//
// 参数:
//   - storage: 存储管理器实例
//
// 返回:
//   - SessionManager: 会话管理器接口实例
//
// 功能:
//   - 提供向后兼容的工厂方法
//   - 直接创建会话管理器实例
func NewSessionManager(storage storage.RepositoryManager) SessionManager {
	return NewSessionManagerImpl(storage)
}

// NewServiceRegistry 创建服务注册器实例（兼容性方法）
//
// 参数:
//   - storage: 存储管理器实例
//
// 返回:
//   - ServiceRegistry: 服务注册器接口实例
//
// 功能:
//   - 提供向后兼容的工厂方法
//   - 直接创建服务注册器实例
func NewServiceRegistry(storage storage.RepositoryManager) ServiceRegistry {
	return NewServiceRegistryImpl(storage)
}

// NewConnectionTracker 创建连接跟踪器实例（兼容性方法）
//
// 参数:
//   - storage: 存储管理器实例
//
// 返回:
//   - ConnectionTracker: 连接跟踪器接口实例
//
// 功能:
//   - 提供向后兼容的工厂方法
//   - 直接创建连接跟踪器实例
func NewConnectionTracker(storage storage.RepositoryManager) ConnectionTracker {
	return NewConnectionTrackerImpl(storage)
}

// NewLoadBalancer 创建负载均衡器实例（兼容性方法）
//
// 参数:
//   - algorithm: 负载均衡算法名称
//
// 返回:
//   - LoadBalancer: 负载均衡器接口实例
//
// 功能:
//   - 提供向后兼容的工厂方法
//   - 直接创建负载均衡器实例
func NewLoadBalancer(algorithm string) LoadBalancer {
	return NewLoadBalancerImpl(algorithm)
}
