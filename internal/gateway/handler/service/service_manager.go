package service

import (
	"gateway/internal/gateway/core"
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

	// GetServices 获取所有服务（返回内部 services map，用于共享健康检查器等需要直接访问的场景）
	GetServices() map[string]*Service

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
	services            map[string]*Service         // serviceID -> Service
	sharedHealthChecker *SharedHealthCheckerManager // 共享健康检查器管理器
	mu                  sync.RWMutex
	useSharedChecker    bool // 是否使用共享健康检查器（默认启用）
}

// NewServiceManager 创建服务管理器
// 默认启用共享健康检查器，以提高大规模场景下的性能
func NewServiceManager() ServiceManager {
	manager := &DefaultServiceManager{
		services:         make(map[string]*Service),
		useSharedChecker: true,
	}

	// 创建共享健康检查器管理器
	// 参数说明：
	// - serviceManager: manager（传入自身，用于直接访问和更新服务节点状态）
	// - workers: 100（并发检查的 goroutine 数量，可根据实际情况调整）
	// 注意：健康检查的间隔、超时等配置从每个服务的 ServiceConfig.HealthCheck 获取
	manager.sharedHealthChecker = NewSharedHealthCheckerManager(manager, 40)

	return manager
}

// NewServiceManagerWithOptions 创建服务管理器（带选项）
// useSharedChecker: 是否使用共享健康检查器（默认 true，建议保持启用）
// workers: 工作池大小（仅在 useSharedChecker=true 时有效，默认 100）
func NewServiceManagerWithOptions(useSharedChecker bool, workers int) ServiceManager {
	manager := &DefaultServiceManager{
		services:         make(map[string]*Service),
		useSharedChecker: useSharedChecker,
	}

	if useSharedChecker {
		if workers <= 0 {
			workers = 20
		}
		manager.sharedHealthChecker = NewSharedHealthCheckerManager(manager, workers)
	}

	return manager
}

// AddService 添加服务
func (m *DefaultServiceManager) AddService(config *ServiceConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.services[config.ID]; exists {
		return ErrServiceExists
	}

	// 创建服务实例（传入是否使用共享检查器的标志）
	service, err := NewService(config, m.useSharedChecker)
	if err != nil {
		return err
	}

	m.services[config.ID] = service

	// 如果使用共享检查器且检查器未启动，则启动它
	if m.useSharedChecker && m.sharedHealthChecker != nil {
		m.sharedHealthChecker.Start()
	}

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

// GetService 获取服务配置
func (m *DefaultServiceManager) GetService(serviceID string) (*ServiceConfig, bool) {
	m.mu.RLock()
	service, exists := m.services[serviceID]
	m.mu.RUnlock()

	if !exists {
		return nil, false
	}

	// Service.GetConfig 直接返回 config 指针，不需要锁（config 是只读的）
	return service.GetConfig(), true
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

// GetServices 获取所有服务（返回内部 services map 的副本）
// 注意：此方法返回 map 的副本，调用者需要自行加锁保护对 Service 的访问
func (m *DefaultServiceManager) GetServices() map[string]*Service {
	m.mu.RLock()
	defer m.mu.RUnlock()

	services := make(map[string]*Service, len(m.services))
	for serviceID, service := range m.services {
		services[serviceID] = service
	}
	return services
}

// AddNode 添加节点到服务
func (m *DefaultServiceManager) AddNode(serviceID string, node *NodeConfig) error {
	m.mu.RLock()
	service, exists := m.services[serviceID]
	m.mu.RUnlock()

	if !exists {
		return ErrServiceNotFound
	}

	// Service.AddNode 内部已有锁保护，不需要持有 ServiceManager 的锁
	return service.AddNode(node)
}

// RemoveNode 从服务中移除节点
func (m *DefaultServiceManager) RemoveNode(serviceID, nodeID string) error {
	m.mu.RLock()
	service, exists := m.services[serviceID]
	m.mu.RUnlock()

	if !exists {
		return ErrServiceNotFound
	}

	// Service.RemoveNode 内部已有锁保护，不需要持有 ServiceManager 的锁
	return service.RemoveNode(nodeID)
}

// UpdateNodeHealth 更新节点健康状态
func (m *DefaultServiceManager) UpdateNodeHealth(serviceID, nodeID string, healthy bool) error {
	m.mu.RLock()
	service, exists := m.services[serviceID]
	m.mu.RUnlock()

	if !exists {
		return ErrServiceNotFound
	}

	// Service.UpdateNodeHealth 内部已有锁保护，不需要持有 ServiceManager 的锁
	return service.UpdateNodeHealth(nodeID, healthy)
}

// UpdateNodeStatus 更新节点状态
func (m *DefaultServiceManager) UpdateNodeStatus(serviceID, nodeID string, enabled bool) error {
	m.mu.RLock()
	service, exists := m.services[serviceID]
	m.mu.RUnlock()

	if !exists {
		return ErrServiceNotFound
	}

	// Service.UpdateNodeStatus 内部已有锁保护，不需要持有 ServiceManager 的锁
	return service.UpdateNodeStatus(nodeID, enabled)
}

// GetHealthyNodes 获取健康的节点
func (m *DefaultServiceManager) GetHealthyNodes(serviceID string) ([]*NodeConfig, error) {
	m.mu.RLock()
	service, exists := m.services[serviceID]
	m.mu.RUnlock()

	if !exists {
		return nil, ErrServiceNotFound
	}

	// Service.GetHealthyNodes 内部已有锁保护，不需要持有 ServiceManager 的锁
	return service.GetHealthyNodes(), nil
}

// GetUnhealthyNodes 获取不健康的节点
func (m *DefaultServiceManager) GetUnhealthyNodes(serviceID string) ([]*NodeConfig, error) {
	m.mu.RLock()
	service, exists := m.services[serviceID]
	m.mu.RUnlock()

	if !exists {
		return nil, ErrServiceNotFound
	}

	// Service.GetUnhealthyNodes 内部已有锁保护，不需要持有 ServiceManager 的锁
	return service.GetUnhealthyNodes(), nil
}

// GetAllNodes 获取所有节点
func (m *DefaultServiceManager) GetAllNodes(serviceID string) ([]*NodeConfig, error) {
	m.mu.RLock()
	service, exists := m.services[serviceID]
	m.mu.RUnlock()

	if !exists {
		return nil, ErrServiceNotFound
	}

	// Service.GetAllNodes 内部已有锁保护，不需要持有 ServiceManager 的锁
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

	// 创建新的服务实例（使用共享健康检查器标志）
	newService, err := NewService(config, m.useSharedChecker)
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
	service, exists := m.services[serviceID]
	m.mu.RUnlock()

	if !exists {
		return nil, ErrServiceNotFound
	}

	// Service.GetStats 内部已有锁保护，不需要持有 ServiceManager 的锁
	return service.GetStats(), nil
}

// SelectNode 为服务选择节点
func (m *DefaultServiceManager) SelectNode(serviceID string, ctx *core.Context) (*NodeConfig, error) {
	m.mu.RLock()
	service, exists := m.services[serviceID]
	m.mu.RUnlock()

	if !exists {
		return nil, ErrServiceNotFound
	}

	// Service.SelectNode 内部已有锁保护，不需要持有 ServiceManager 的锁
	return service.SelectNode(ctx)
}

// RecordServiceSuccess 记录服务调用成功
func (m *DefaultServiceManager) RecordServiceSuccess(serviceID string, responseTime time.Duration) {
	m.mu.RLock()
	service, exists := m.services[serviceID]
	m.mu.RUnlock()

	if exists {
		// Service.RecordSuccess 内部已有锁保护，不需要持有 ServiceManager 的锁
		service.RecordSuccess(responseTime)
	}
}

// RecordServiceFailure 记录服务调用失败
func (m *DefaultServiceManager) RecordServiceFailure(serviceID string) {
	m.mu.RLock()
	service, exists := m.services[serviceID]
	m.mu.RUnlock()

	if exists {
		// Service.RecordFailure 内部已有锁保护，不需要持有 ServiceManager 的锁
		service.RecordFailure()
	}
}

// Close 关闭管理器
func (m *DefaultServiceManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 关闭所有服务
	for _, service := range m.services {
		if err := service.Close(); err != nil {
			// 记录错误但继续关闭其他服务
		}
	}

	// 关闭共享健康检查器
	if m.sharedHealthChecker != nil {
		if err := m.sharedHealthChecker.Stop(); err != nil {
			return err
		}
	}

	return nil
}
