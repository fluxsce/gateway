package service

import (
	"gohub/internal/gateway/core"
	"sync"
	"time"
)

// ServiceManager 服务管理器接口
type ServiceManager interface {
	// AddService 添加服务
	AddService(config *ServiceConfig) error

	// RemoveService 移除服务
	RemoveService(serviceID string) error

	// GetService 获取服务配置
	GetService(serviceID string) (*ServiceConfig, bool)

	// ListServices 列出所有服务配置
	ListServices() []*ServiceConfig

	// AddNode 添加节点到服务
	AddNode(serviceID string, node *NodeConfig) error

	// RemoveNode 从服务中移除节点
	RemoveNode(serviceID, nodeID string) error

	// UpdateNodeHealth 更新节点健康状态
	UpdateNodeHealth(serviceID, nodeID string, healthy bool) error

	// UpdateNodeStatus 更新节点状态
	UpdateNodeStatus(serviceID, nodeID string, enabled bool) error

	// GetHealthyNodes 获取健康的节点
	GetHealthyNodes(serviceID string) ([]*NodeConfig, error)

	// GetUnhealthyNodes 获取不健康的节点
	GetUnhealthyNodes(serviceID string) ([]*NodeConfig, error)

	// GetAllNodes 获取所有节点
	GetAllNodes(serviceID string) ([]*NodeConfig, error)

	// UpdateService 更新服务配置
	UpdateService(service *ServiceConfig) error

	// GetServiceStats 获取服务统计信息
	GetServiceStats(serviceID string) (map[string]interface{}, error)

	// SelectNode 为服务选择节点
	SelectNode(serviceID string, ctx *core.Context) (*NodeConfig, error)

	// RecordServiceSuccess 记录服务调用成功
	RecordServiceSuccess(serviceID string, responseTime time.Duration)

	// RecordServiceFailure 记录服务调用失败
	RecordServiceFailure(serviceID string)

	// Close 关闭管理器
	Close() error
}

// DefaultServiceManager 默认服务管理器实现
type DefaultServiceManager struct {
	services map[string]*Service // serviceID -> Service
	mu       sync.RWMutex
}

// NewServiceManager 创建服务管理器
func NewServiceManager() ServiceManager {
	return &DefaultServiceManager{
		services: make(map[string]*Service),
	}
}

// AddService 添加服务
func (m *DefaultServiceManager) AddService(config *ServiceConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.services[config.ID]; exists {
		return ErrServiceExists
	}

	// 创建服务实例
	service, err := NewService(config)
	if err != nil {
		return err
	}

	m.services[config.ID] = service
	return nil
}

// RemoveService 移除服务
func (m *DefaultServiceManager) RemoveService(serviceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	service, exists := m.services[serviceID]
	if !exists {
		return ErrServiceNotFound
	}

	// 关闭服务资源
	if err := service.Close(); err != nil {
		return err
	}

	delete(m.services, serviceID)
	return nil
}

// GetService 获取服务
func (m *DefaultServiceManager) GetService(serviceID string) (*ServiceConfig, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	service, exists := m.services[serviceID]
	if !exists {
		return nil, false
	}
	return service.GetConfig(), exists
}

// ListServices 列出所有服务
func (m *DefaultServiceManager) ListServices() []*ServiceConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	services := make([]*ServiceConfig, 0, len(m.services))
	for _, service := range m.services {
		services = append(services, service.GetConfig())
	}
	return services
}

// AddNode 添加节点到服务
func (m *DefaultServiceManager) AddNode(serviceID string, node *NodeConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	service, exists := m.services[serviceID]
	if !exists {
		return ErrServiceNotFound
	}

	return service.AddNode(node)
}

// RemoveNode 从服务中移除节点
func (m *DefaultServiceManager) RemoveNode(serviceID, nodeID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	service, exists := m.services[serviceID]
	if !exists {
		return ErrServiceNotFound
	}

	return service.RemoveNode(nodeID)
}

// UpdateNodeHealth 更新节点健康状态
func (m *DefaultServiceManager) UpdateNodeHealth(serviceID, nodeID string, healthy bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	service, exists := m.services[serviceID]
	if !exists {
		return ErrServiceNotFound
	}

	return service.UpdateNodeHealth(nodeID, healthy)
}

// UpdateNodeStatus 更新节点状态
func (m *DefaultServiceManager) UpdateNodeStatus(serviceID, nodeID string, enabled bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	service, exists := m.services[serviceID]
	if !exists {
		return ErrServiceNotFound
	}

	return service.UpdateNodeStatus(nodeID, enabled)
}

// GetHealthyNodes 获取健康的节点
func (m *DefaultServiceManager) GetHealthyNodes(serviceID string) ([]*NodeConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	service, exists := m.services[serviceID]
	if !exists {
		return nil, ErrServiceNotFound
	}

	return service.GetHealthyNodes(), nil
}

// GetUnhealthyNodes 获取不健康的节点
func (m *DefaultServiceManager) GetUnhealthyNodes(serviceID string) ([]*NodeConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	service, exists := m.services[serviceID]
	if !exists {
		return nil, ErrServiceNotFound
	}

	return service.GetUnhealthyNodes(), nil
}

// GetAllNodes 获取所有节点
func (m *DefaultServiceManager) GetAllNodes(serviceID string) ([]*NodeConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	service, exists := m.services[serviceID]
	if !exists {
		return nil, ErrServiceNotFound
	}

	return service.GetAllNodes(), nil
}

// UpdateService 更新服务配置
func (m *DefaultServiceManager) UpdateService(config *ServiceConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	service, exists := m.services[config.ID]
	if !exists {
		return ErrServiceNotFound
	}

	// 创建新的服务实例
	newService, err := NewService(config)
	if err != nil {
		return err
	}

	// 关闭旧服务
	if err := service.Close(); err != nil {
		return err
	}

	m.services[config.ID] = newService
	return nil
}

// GetServiceStats 获取服务统计信息
func (m *DefaultServiceManager) GetServiceStats(serviceID string) (map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	service, exists := m.services[serviceID]
	if !exists {
		return nil, ErrServiceNotFound
	}

	return service.GetStats(), nil
}

// SelectNode 为服务选择节点
func (m *DefaultServiceManager) SelectNode(serviceID string, ctx *core.Context) (*NodeConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	service, exists := m.services[serviceID]
	if !exists {
		return nil, ErrServiceNotFound
	}

	return service.SelectNode(ctx)
}

// RecordServiceSuccess 记录服务调用成功
func (m *DefaultServiceManager) RecordServiceSuccess(serviceID string, responseTime time.Duration) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if service, exists := m.services[serviceID]; exists {
		service.RecordSuccess(responseTime)
	}
}

// RecordServiceFailure 记录服务调用失败
func (m *DefaultServiceManager) RecordServiceFailure(serviceID string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if service, exists := m.services[serviceID]; exists {
		service.RecordFailure()
	}
}

// Close 关闭管理器
func (m *DefaultServiceManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var lastErr error
	for serviceID, service := range m.services {
		if err := service.Close(); err != nil {
			lastErr = err
		}
		delete(m.services, serviceID)
	}

	return lastErr
}
