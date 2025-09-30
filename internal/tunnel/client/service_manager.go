// Package client 提供服务管理器的完整实现
// 服务管理器负责客户端本地服务的生命周期管理
package client

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"gateway/internal/tunnel/types"
	"gateway/pkg/logger"
)

// serviceManager 服务管理器实现
// 实现 ServiceManager 接口，管理本地服务注册和状态
type serviceManager struct {
	client   *tunnelClient
	services map[string]*serviceInfo
	mutex    sync.RWMutex
}

// serviceInfo 服务信息
type serviceInfo struct {
	service        *types.TunnelService
	status         string
	registeredTime time.Time
	lastActiveTime time.Time
	mutex          sync.RWMutex
}

// NewServiceManager 创建服务管理器实例
//
// 参数:
//   - client: 隧道客户端实例
//
// 返回:
//   - ServiceManager: 服务管理器接口实例
//
// 功能:
//   - 初始化服务管理器
//   - 创建服务映射表和状态管理
func NewServiceManager(client *tunnelClient) ServiceManager {
	return &serviceManager{
		client:   client,
		services: make(map[string]*serviceInfo),
	}
}

// RegisterService 注册服务
func (sm *serviceManager) RegisterService(ctx context.Context, service *types.TunnelService) error {
	// 检查服务是否已存在
	sm.mutex.RLock()
	if _, exists := sm.services[service.TunnelServiceId]; exists {
		sm.mutex.RUnlock()
		return fmt.Errorf("service %s already registered", service.TunnelServiceId)
	}
	sm.mutex.RUnlock()

	// 向服务器发送注册请求
	registerMsg := &ControlMessage{
		Type:      MessageTypeRegisterService,
		RequestID: sm.generateRequestID(),
		Data: map[string]interface{}{
			"service": map[string]interface{}{
				"serviceId":      service.TunnelServiceId,
				"serviceName":    service.ServiceName,
				"serviceType":    service.ServiceType,
				"localAddress":   service.LocalAddress,
				"localPort":      service.LocalPort,
				"remotePort":     service.RemotePort,
				"customDomains":  service.CustomDomains,
				"subDomain":      service.SubDomain,
				"httpUser":       service.HttpUser,
				"httpPassword":   service.HttpPassword,
				"useEncryption":  service.UseEncryption,
				"useCompression": service.UseCompression,
				"secretKey":      service.SecretKey,
				"bandwidthLimit": service.BandwidthLimit,
				"maxConnections": service.MaxConnections,
			},
		},
		Timestamp: time.Now(),
	}

	if err := sm.client.controlConn.SendMessage(ctx, registerMsg); err != nil {
		return fmt.Errorf("failed to send register service message: %w", err)
	}

	// 创建服务信息
	serviceInfo := &serviceInfo{
		service:        service,
		status:         ServiceStatusInactive,
		registeredTime: time.Now(),
		lastActiveTime: time.Now(),
	}

	// 添加到服务列表
	sm.mutex.Lock()
	sm.services[service.TunnelServiceId] = serviceInfo
	sm.mutex.Unlock()

	logger.Info("Service registered with server", map[string]interface{}{
		"serviceId":   service.TunnelServiceId,
		"serviceName": service.ServiceName,
		"serviceType": service.ServiceType,
		"localPort":   service.LocalPort,
	})

	return nil
}

// UnregisterService 注销服务
func (sm *serviceManager) UnregisterService(ctx context.Context, serviceID string) error {
	// 查找服务
	sm.mutex.RLock()
	serviceInfo, exists := sm.services[serviceID]
	sm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("service %s not found", serviceID)
	}

	// 停止服务（如果正在运行）
	if serviceInfo.status == ServiceStatusActive {
		if err := sm.StopService(ctx, serviceID); err != nil {
			logger.Error("Failed to stop service during unregister", map[string]interface{}{
				"serviceId": serviceID,
				"error":     err.Error(),
			})
		}
	}

	// 向服务器发送注销请求
	unregisterMsg := &ControlMessage{
		Type:      MessageTypeUnregisterService,
		RequestID: sm.generateRequestID(),
		Data: map[string]interface{}{
			"serviceId": serviceID,
		},
		Timestamp: time.Now(),
	}

	if err := sm.client.controlConn.SendMessage(ctx, unregisterMsg); err != nil {
		return fmt.Errorf("failed to send unregister service message: %w", err)
	}

	// 从服务列表中移除
	sm.mutex.Lock()
	delete(sm.services, serviceID)
	sm.mutex.Unlock()

	logger.Info("Service unregistered from server", map[string]interface{}{
		"serviceId":   serviceID,
		"serviceName": serviceInfo.service.ServiceName,
	})

	return nil
}

// GetService 获取服务
func (sm *serviceManager) GetService(ctx context.Context, serviceID string) (*types.TunnelService, error) {
	sm.mutex.RLock()
	serviceInfo, exists := sm.services[serviceID]
	sm.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("service %s not found", serviceID)
	}

	serviceInfo.mutex.RLock()
	service := serviceInfo.service
	serviceInfo.mutex.RUnlock()

	return service, nil
}

// GetAllServices 获取所有服务
func (sm *serviceManager) GetAllServices(ctx context.Context) ([]*types.TunnelService, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	services := make([]*types.TunnelService, 0, len(sm.services))
	for _, serviceInfo := range sm.services {
		serviceInfo.mutex.RLock()
		services = append(services, serviceInfo.service)
		serviceInfo.mutex.RUnlock()
	}

	return services, nil
}

// StartService 启动服务
func (sm *serviceManager) StartService(ctx context.Context, serviceID string) error {
	// 查找服务
	sm.mutex.RLock()
	serviceInfo, exists := sm.services[serviceID]
	sm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("service %s not found", serviceID)
	}

	serviceInfo.mutex.Lock()
	if serviceInfo.status == ServiceStatusActive {
		serviceInfo.mutex.Unlock()
		return fmt.Errorf("service %s is already active", serviceID)
	}

	// 验证本地服务可用性
	if err := sm.checkLocalServiceAvailability(serviceInfo.service); err != nil {
		serviceInfo.mutex.Unlock()
		return fmt.Errorf("local service not available: %w", err)
	}

	serviceInfo.status = ServiceStatusStarting
	serviceInfo.mutex.Unlock()

	// 这里可以添加启动本地代理监听的逻辑
	// 目前简化处理，直接标记为活跃

	serviceInfo.mutex.Lock()
	serviceInfo.status = ServiceStatusActive
	serviceInfo.lastActiveTime = time.Now()
	serviceInfo.mutex.Unlock()

	logger.Info("Service started", map[string]interface{}{
		"serviceId":   serviceID,
		"serviceName": serviceInfo.service.ServiceName,
	})

	return nil
}

// StopService 停止服务
func (sm *serviceManager) StopService(ctx context.Context, serviceID string) error {
	// 查找服务
	sm.mutex.RLock()
	serviceInfo, exists := sm.services[serviceID]
	sm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("service %s not found", serviceID)
	}

	serviceInfo.mutex.Lock()
	if serviceInfo.status != ServiceStatusActive {
		serviceInfo.mutex.Unlock()
		return fmt.Errorf("service %s is not active", serviceID)
	}

	serviceInfo.status = ServiceStatusStopping
	serviceInfo.mutex.Unlock()

	// 这里可以添加停止本地代理监听的逻辑
	// 目前简化处理，直接标记为非活跃

	serviceInfo.mutex.Lock()
	serviceInfo.status = ServiceStatusInactive
	serviceInfo.mutex.Unlock()

	logger.Info("Service stopped", map[string]interface{}{
		"serviceId":   serviceID,
		"serviceName": serviceInfo.service.ServiceName,
	})

	return nil
}

// ValidateService 验证服务配置
func (sm *serviceManager) ValidateService(ctx context.Context, service *types.TunnelService) error {
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

	// 验证带宽限制格式
	if service.BandwidthLimit != nil && *service.BandwidthLimit != "" {
		if !sm.isValidBandwidthLimit(*service.BandwidthLimit) {
			return fmt.Errorf("invalid bandwidth limit format: %s", *service.BandwidthLimit)
		}
	}

	// 验证连接数限制
	if service.MaxConnections != nil && *service.MaxConnections <= 0 {
		return fmt.Errorf("invalid max connections: %d", *service.MaxConnections)
	}

	return nil
}

// checkLocalServiceAvailability 检查本地服务可用性
func (sm *serviceManager) checkLocalServiceAvailability(service *types.TunnelService) error {
	address := net.JoinHostPort(service.LocalAddress, fmt.Sprintf("%d", service.LocalPort))

	switch service.ServiceType {
	case types.ProxyTypeTCP, types.ProxyTypeHTTP, types.ProxyTypeHTTPS:
		// 尝试TCP连接
		conn, err := net.DialTimeout("tcp", address, 5*time.Second)
		if err != nil {
			return fmt.Errorf("cannot connect to TCP service at %s: %w", address, err)
		}
		conn.Close()

	case types.ProxyTypeUDP:
		// 尝试UDP连接
		udpAddr, err := net.ResolveUDPAddr("udp", address)
		if err != nil {
			return fmt.Errorf("invalid UDP address %s: %w", address, err)
		}

		conn, err := net.DialUDP("udp", nil, udpAddr)
		if err != nil {
			return fmt.Errorf("cannot connect to UDP service at %s: %w", address, err)
		}
		conn.Close()
	}

	return nil
}

// isValidBandwidthLimit 验证带宽限制格式
func (sm *serviceManager) isValidBandwidthLimit(limit string) bool {
	if limit == "" {
		return true
	}

	// 支持的格式: 100KB, 1MB, 10MB/s, 等
	limit = strings.ToUpper(strings.TrimSpace(limit))

	// 简单的格式检查
	validSuffixes := []string{"B", "KB", "MB", "GB", "B/S", "KB/S", "MB/S", "GB/S"}
	for _, suffix := range validSuffixes {
		if strings.HasSuffix(limit, suffix) {
			return true
		}
	}

	return false
}

// generateRequestID 生成请求ID
func (sm *serviceManager) generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// GetServiceStatus 获取服务状态
func (sm *serviceManager) GetServiceStatus(serviceID string) string {
	sm.mutex.RLock()
	serviceInfo, exists := sm.services[serviceID]
	sm.mutex.RUnlock()

	if !exists {
		return ServiceStatusError
	}

	serviceInfo.mutex.RLock()
	status := serviceInfo.status
	serviceInfo.mutex.RUnlock()

	return status
}

// GetActiveServices 获取活跃服务列表
func (sm *serviceManager) GetActiveServices() []*types.TunnelService {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	var activeServices []*types.TunnelService
	for _, serviceInfo := range sm.services {
		serviceInfo.mutex.RLock()
		if serviceInfo.status == ServiceStatusActive {
			activeServices = append(activeServices, serviceInfo.service)
		}
		serviceInfo.mutex.RUnlock()
	}

	return activeServices
}

// UpdateServiceActivity 更新服务活动时间
func (sm *serviceManager) UpdateServiceActivity(serviceID string) {
	sm.mutex.RLock()
	serviceInfo, exists := sm.services[serviceID]
	sm.mutex.RUnlock()

	if exists {
		serviceInfo.mutex.Lock()
		serviceInfo.lastActiveTime = time.Now()
		serviceInfo.mutex.Unlock()
	}
}

// Close 关闭服务管理器
func (sm *serviceManager) Close() error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// 停止所有活跃服务
	for serviceID, serviceInfo := range sm.services {
		serviceInfo.mutex.RLock()
		status := serviceInfo.status
		serviceInfo.mutex.RUnlock()

		if status == ServiceStatusActive {
			if err := sm.StopService(context.Background(), serviceID); err != nil {
				logger.Error("Failed to stop service during close", map[string]interface{}{
					"serviceId": serviceID,
					"error":     err.Error(),
				})
			}
		}
	}

	// 清空服务列表
	sm.services = make(map[string]*serviceInfo)

	return nil
}
