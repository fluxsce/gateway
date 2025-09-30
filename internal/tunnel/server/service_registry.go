// Package server 提供服务注册器的完整实现
// 服务注册器负责管理隧道服务的注册、注销、端口分配和配置验证
package server

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"gateway/internal/tunnel/storage"
	"gateway/internal/tunnel/types"
	"gateway/pkg/logger"
)

// serviceRegistry 服务注册器实现
// 实现 ServiceRegistry 接口，管理隧道服务注册
type serviceRegistry struct {
	storage       storage.RepositoryManager
	portAllocator *portAllocator
	services      map[string]*serviceInfo
	serviceMutex  sync.RWMutex
	clientIndex   map[string][]string // clientID -> serviceIDs
}

// serviceInfo 服务信息
type serviceInfo struct {
	service      *types.TunnelService
	lastActivity time.Time
	mutex        sync.RWMutex
}

// portAllocator 端口分配器
type portAllocator struct {
	allocatedPorts map[int]string // port -> serviceID
	portRange      portRange
	mutex          sync.RWMutex
}

// portRange 端口范围配置
type portRange struct {
	minPort int
	maxPort int
}

// NewServiceRegistryImpl 创建新的服务注册器实例
//
// 参数:
//   - storage: 存储管理器，用于持久化服务数据
//
// 返回:
//   - ServiceRegistry: 服务注册器接口实例
//
// 功能:
//   - 初始化服务注册器
//   - 创建端口分配器，默认端口范围 10000-20000
//   - 从数据库加载现有服务
func NewServiceRegistryImpl(storage storage.RepositoryManager) ServiceRegistry {
	allocator := &portAllocator{
		allocatedPorts: make(map[int]string),
		portRange: portRange{
			minPort: 10000,
			maxPort: 20000,
		},
	}

	registry := &serviceRegistry{
		storage:       storage,
		portAllocator: allocator,
		services:      make(map[string]*serviceInfo),
		clientIndex:   make(map[string][]string),
	}

	// 加载现有服务
	registry.loadExistingServices()

	return registry
}

// RegisterService 注册服务
//
// 参数:
//   - ctx: 上下文
//   - clientID: 客户端ID
//   - service: 要注册的服务对象
//
// 返回:
//   - error: 注册失败时返回错误
//
// 功能:
//   - 验证服务配置
//   - 分配远程端口（如果需要）
//   - 检查端口冲突
//   - 持久化服务到数据库
//   - 添加到内存索引
func (sr *serviceRegistry) RegisterService(ctx context.Context, clientID string, service *types.TunnelService) error {
	// 验证服务配置
	if err := sr.ValidateServiceConfig(ctx, service); err != nil {
		return fmt.Errorf("service configuration validation failed: %w", err)
	}

	// 设置服务基本信息
	service.TunnelClientId = clientID
	service.TunnelServiceId = sr.generateServiceID(service.ServiceName)
	service.ServiceStatus = types.ServiceStatusActive
	service.RegisteredTime = time.Now()
	service.LastActiveTime = &[]time.Time{time.Now()}[0]
	service.AddTime = time.Now()
	service.EditTime = time.Now()
	service.AddWho = "system"
	service.EditWho = "system"
	service.ActiveFlag = types.ActiveFlagYes

	// 分配远程端口
	if service.RemotePort == nil || *service.RemotePort == 0 {
		port, err := sr.AllocatePort(ctx, service.ServiceType, nil)
		if err != nil {
			return fmt.Errorf("failed to allocate port: %w", err)
		}
		service.RemotePort = &port
	} else {
		// 检查指定端口是否可用
		if err := sr.checkPortAvailable(*service.RemotePort, service.TunnelServiceId); err != nil {
			return fmt.Errorf("port %d not available: %w", *service.RemotePort, err)
		}
		sr.portAllocator.mutex.Lock()
		sr.portAllocator.allocatedPorts[*service.RemotePort] = service.TunnelServiceId
		sr.portAllocator.mutex.Unlock()
	}

	// 持久化到数据库
	if err := sr.storage.GetTunnelServiceRepository().Create(ctx, service); err != nil {
		// 回滚端口分配
		if service.RemotePort != nil {
			sr.ReleasePort(ctx, *service.RemotePort)
		}
		return fmt.Errorf("failed to create service in database: %w", err)
	}

	// 添加到内存索引
	serviceInfo := &serviceInfo{
		service:      service,
		lastActivity: time.Now(),
	}

	sr.serviceMutex.Lock()
	sr.services[service.TunnelServiceId] = serviceInfo
	sr.clientIndex[clientID] = append(sr.clientIndex[clientID], service.TunnelServiceId)
	sr.serviceMutex.Unlock()

	logger.Info("Service registered", map[string]interface{}{
		"serviceId":   service.TunnelServiceId,
		"serviceName": service.ServiceName,
		"clientId":    clientID,
		"serviceType": service.ServiceType,
		"remotePort":  *service.RemotePort,
	})

	return nil
}

// UnregisterService 注销服务
//
// 参数:
//   - ctx: 上下文
//   - serviceID: 服务ID
//
// 返回:
//   - error: 注销失败时返回错误
//
// 功能:
//   - 从内存索引中移除服务
//   - 释放分配的端口
//   - 更新数据库中的服务状态
func (sr *serviceRegistry) UnregisterService(ctx context.Context, serviceID string) error {
	sr.serviceMutex.Lock()
	serviceInfo, exists := sr.services[serviceID]
	if exists {
		delete(sr.services, serviceID)

		// 从客户端索引中移除
		clientID := serviceInfo.service.TunnelClientId
		if serviceIDs, ok := sr.clientIndex[clientID]; ok {
			var newServiceIDs []string
			for _, id := range serviceIDs {
				if id != serviceID {
					newServiceIDs = append(newServiceIDs, id)
				}
			}
			if len(newServiceIDs) == 0 {
				delete(sr.clientIndex, clientID)
			} else {
				sr.clientIndex[clientID] = newServiceIDs
			}
		}
	}
	sr.serviceMutex.Unlock()

	if !exists {
		return fmt.Errorf("service %s not found", serviceID)
	}

	// 释放端口
	if serviceInfo.service.RemotePort != nil {
		if err := sr.ReleasePort(ctx, *serviceInfo.service.RemotePort); err != nil {
			logger.Error("Failed to release port", map[string]interface{}{
				"error":      err.Error(),
				"serviceId":  serviceID,
				"remotePort": *serviceInfo.service.RemotePort,
			})
		}
	}

	// 更新数据库状态
	if err := sr.storage.GetTunnelServiceRepository().UpdateStatus(ctx, serviceID, types.ServiceStatusOffline, nil); err != nil {
		logger.Error("Failed to update service status in database", map[string]interface{}{
			"error":     err.Error(),
			"serviceId": serviceID,
		})
	}

	logger.Info("Service unregistered", map[string]interface{}{
		"serviceId": serviceID,
	})

	return nil
}

// GetService 获取服务
//
// 参数:
//   - ctx: 上下文
//   - serviceID: 服务ID
//
// 返回:
//   - *types.TunnelService: 服务对象
//   - error: 获取失败时返回错误
//
// 功能:
//   - 从内存索引中查找服务
//   - 如果内存中不存在，从数据库加载
func (sr *serviceRegistry) GetService(ctx context.Context, serviceID string) (*types.TunnelService, error) {
	sr.serviceMutex.RLock()
	serviceInfo, exists := sr.services[serviceID]
	sr.serviceMutex.RUnlock()

	if exists {
		serviceInfo.mutex.RLock()
		service := serviceInfo.service
		serviceInfo.mutex.RUnlock()
		return service, nil
	}

	// 从数据库加载
	service, err := sr.storage.GetTunnelServiceRepository().GetByID(ctx, serviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get service from database: %w", err)
	}

	return service, nil
}

// GetServicesByClient 获取客户端的所有服务
//
// 参数:
//   - ctx: 上下文
//   - clientID: 客户端ID
//
// 返回:
//   - []*types.TunnelService: 服务列表
//   - error: 获取失败时返回错误
//
// 功能:
//   - 通过客户端索引快速查找服务
//   - 返回客户端注册的所有活跃服务
func (sr *serviceRegistry) GetServicesByClient(ctx context.Context, clientID string) ([]*types.TunnelService, error) {
	sr.serviceMutex.RLock()
	serviceIDs, exists := sr.clientIndex[clientID]
	sr.serviceMutex.RUnlock()

	if !exists {
		// 从数据库查询
		return sr.storage.GetTunnelServiceRepository().GetByClientID(ctx, clientID)
	}

	var services []*types.TunnelService
	for _, serviceID := range serviceIDs {
		if service, err := sr.GetService(ctx, serviceID); err == nil {
			services = append(services, service)
		}
	}

	return services, nil
}

// AllocatePort 分配端口
//
// 参数:
//   - ctx: 上下文
//   - serviceType: 服务类型
//   - preferPort: 首选端口（可为nil）
//
// 返回:
//   - int: 分配的端口号
//   - error: 分配失败时返回错误
//
// 功能:
//   - 如果指定首选端口且可用，则分配首选端口
//   - 否则在配置的端口范围内自动分配可用端口
//   - 检查端口是否已被占用
func (sr *serviceRegistry) AllocatePort(ctx context.Context, serviceType string, preferPort *int) (int, error) {
	sr.portAllocator.mutex.Lock()
	defer sr.portAllocator.mutex.Unlock()

	// 如果指定了首选端口
	if preferPort != nil && *preferPort > 0 {
		if err := sr.checkPortAvailableUnsafe(*preferPort, ""); err != nil {
			return 0, fmt.Errorf("preferred port %d not available: %w", *preferPort, err)
		}
		return *preferPort, nil
	}

	// 自动分配端口
	for port := sr.portAllocator.portRange.minPort; port <= sr.portAllocator.portRange.maxPort; port++ {
		if _, allocated := sr.portAllocator.allocatedPorts[port]; !allocated {
			// 检查端口是否真的可用
			if sr.isPortReallyAvailable(port) {
				return port, nil
			}
		}
	}

	return 0, fmt.Errorf("no available ports in range %d-%d",
		sr.portAllocator.portRange.minPort, sr.portAllocator.portRange.maxPort)
}

// ReleasePort 释放端口
//
// 参数:
//   - ctx: 上下文
//   - port: 要释放的端口号
//
// 返回:
//   - error: 释放失败时返回错误
//
// 功能:
//   - 从端口分配表中移除端口
//   - 记录端口释放日志
func (sr *serviceRegistry) ReleasePort(ctx context.Context, port int) error {
	sr.portAllocator.mutex.Lock()
	defer sr.portAllocator.mutex.Unlock()

	serviceID, exists := sr.portAllocator.allocatedPorts[port]
	if !exists {
		return fmt.Errorf("port %d is not allocated", port)
	}

	delete(sr.portAllocator.allocatedPorts, port)

	logger.Info("Port released", map[string]interface{}{
		"port":      port,
		"serviceId": serviceID,
	})

	return nil
}

// ValidateServiceConfig 验证服务配置
//
// 参数:
//   - ctx: 上下文
//   - service: 要验证的服务对象
//
// 返回:
//   - error: 验证失败时返回错误
//
// 功能:
//   - 检查必填字段
//   - 验证服务类型
//   - 验证端口范围
//   - 验证地址格式
func (sr *serviceRegistry) ValidateServiceConfig(ctx context.Context, service *types.TunnelService) error {
	// 检查必填字段
	if service.ServiceName == "" {
		return fmt.Errorf("service name is required")
	}

	if service.ServiceType == "" {
		return fmt.Errorf("service type is required")
	}

	if service.LocalAddress == "" {
		return fmt.Errorf("local address is required")
	}

	if service.LocalPort <= 0 || service.LocalPort > 65535 {
		return fmt.Errorf("invalid local port: %d", service.LocalPort)
	}

	// 验证服务类型
	validTypes := []string{
		types.ProxyTypeTCP, types.ProxyTypeUDP,
		types.ProxyTypeHTTP, types.ProxyTypeHTTPS,
		types.ProxyTypeSTCP, types.ProxyTypeSUDP, types.ProxyTypeXTCP,
	}

	typeValid := false
	for _, validType := range validTypes {
		if service.ServiceType == validType {
			typeValid = true
			break
		}
	}

	if !typeValid {
		return fmt.Errorf("invalid service type: %s", service.ServiceType)
	}

	// 验证地址格式
	if net.ParseIP(service.LocalAddress) == nil && service.LocalAddress != "localhost" {
		// 尝试解析域名
		if _, err := net.LookupHost(service.LocalAddress); err != nil {
			return fmt.Errorf("invalid local address: %s", service.LocalAddress)
		}
	}

	// 验证远程端口
	if service.RemotePort != nil {
		if *service.RemotePort <= 0 || *service.RemotePort > 65535 {
			return fmt.Errorf("invalid remote port: %d", *service.RemotePort)
		}
	}

	// HTTP/HTTPS特定验证
	if service.ServiceType == types.ProxyTypeHTTP || service.ServiceType == types.ProxyTypeHTTPS {
		if service.CustomDomains == nil && service.SubDomain == nil {
			return fmt.Errorf("HTTP/HTTPS services require custom domains or subdomain")
		}
	}

	return nil
}

// loadExistingServices 加载现有服务
func (sr *serviceRegistry) loadExistingServices() {
	ctx := context.Background()

	// 从数据库加载所有活跃的服务
	services, err := sr.storage.GetTunnelServiceRepository().GetActiveServices(ctx, types.ActiveFlagYes)
	if err != nil {
		logger.Error("Failed to load existing services", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	loadedCount := 0
	for _, service := range services {
		// 重建内存索引
		serviceInfo := &serviceInfo{
			service:      service,
			lastActivity: time.Now(),
		}

		sr.serviceMutex.Lock()
		sr.services[service.TunnelServiceId] = serviceInfo

		// 重建客户端索引
		clientID := service.TunnelClientId
		sr.clientIndex[clientID] = append(sr.clientIndex[clientID], service.TunnelServiceId)

		// 重建端口分配表
		if service.RemotePort != nil && *service.RemotePort > 0 {
			sr.portAllocator.allocatedPorts[*service.RemotePort] = service.TunnelServiceId
		}
		sr.serviceMutex.Unlock()

		loadedCount++
	}

	logger.Info("Service registry initialized", map[string]interface{}{
		"loadedServices": loadedCount,
		"totalServices":  len(sr.services),
	})
}

// checkPortAvailable 检查端口是否可用（带锁）
func (sr *serviceRegistry) checkPortAvailable(port int, serviceID string) error {
	sr.portAllocator.mutex.RLock()
	defer sr.portAllocator.mutex.RUnlock()

	return sr.checkPortAvailableUnsafe(port, serviceID)
}

// checkPortAvailableUnsafe 检查端口是否可用（不带锁）
func (sr *serviceRegistry) checkPortAvailableUnsafe(port int, serviceID string) error {
	if allocatedServiceID, allocated := sr.portAllocator.allocatedPorts[port]; allocated {
		if allocatedServiceID != serviceID {
			return fmt.Errorf("port %d is already allocated to service %s", port, allocatedServiceID)
		}
	}

	// 检查端口是否真的可用
	if !sr.isPortReallyAvailable(port) {
		return fmt.Errorf("port %d is in use by another process", port)
	}

	return nil
}

// isPortReallyAvailable 检查端口是否真的可用
func (sr *serviceRegistry) isPortReallyAvailable(port int) bool {
	// 尝试监听端口
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	listener.Close()

	// 检查UDP端口
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return false
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return false
	}
	udpConn.Close()

	return true
}

// generateServiceID 生成服务ID
func (sr *serviceRegistry) generateServiceID(serviceName string) string {
	// 清理服务名称，只保留字母数字和下划线
	cleanName := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return '_'
	}, serviceName)

	return fmt.Sprintf("service_%s_%d", cleanName, time.Now().UnixNano())
}

// SetPortRange 设置端口分配范围
func (sr *serviceRegistry) SetPortRange(minPort, maxPort int) error {
	if minPort <= 0 || maxPort <= 0 || minPort >= maxPort {
		return fmt.Errorf("invalid port range: %d-%d", minPort, maxPort)
	}

	sr.portAllocator.mutex.Lock()
	sr.portAllocator.portRange.minPort = minPort
	sr.portAllocator.portRange.maxPort = maxPort
	sr.portAllocator.mutex.Unlock()

	logger.Info("Port range updated", map[string]interface{}{
		"minPort": minPort,
		"maxPort": maxPort,
	})

	return nil
}

// GetPortUsage 获取端口使用情况
func (sr *serviceRegistry) GetPortUsage() map[string]interface{} {
	sr.portAllocator.mutex.RLock()
	defer sr.portAllocator.mutex.RUnlock()

	totalPorts := sr.portAllocator.portRange.maxPort - sr.portAllocator.portRange.minPort + 1
	allocatedCount := len(sr.portAllocator.allocatedPorts)

	usage := map[string]interface{}{
		"totalPorts":     totalPorts,
		"allocatedPorts": allocatedCount,
		"availablePorts": totalPorts - allocatedCount,
		"usageRate":      float64(allocatedCount) / float64(totalPorts) * 100,
		"portRange": map[string]int{
			"min": sr.portAllocator.portRange.minPort,
			"max": sr.portAllocator.portRange.maxPort,
		},
	}

	return usage
}
