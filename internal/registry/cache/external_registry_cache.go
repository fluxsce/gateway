package cache

import (
	"context"
	"fmt"
	"sync"

	"gateway/internal/registry/core"
	"gateway/pkg/logger"
)

// ExternalRegistryClient 外部注册中心客户端接口
type ExternalRegistryClient interface {
	// 服务管理
	RegisterService(ctx context.Context, service *core.Service) error
	DeregisterService(ctx context.Context, service *core.Service) error

	// 实例管理
	RegisterInstance(ctx context.Context, instance *core.ServiceInstance) error
	DeregisterInstance(ctx context.Context, instance *core.ServiceInstance) error
	UpdateInstance(ctx context.Context, instance *core.ServiceInstance) error

	// 服务发现
	DiscoverServices(ctx context.Context, groupName string) ([]*core.Service, error)
	DiscoverInstances(ctx context.Context, serviceName, groupName string) ([]*core.ServiceInstance, error)

	// 健康检查
	UpdateInstanceHealth(ctx context.Context, instanceId string, healthy bool) error

	Close() error // 完全关闭客户端并释放所有资源
}

// ExternalRegistryCacheManager 外部注册中心缓存管理器
type ExternalRegistryCacheManager struct {
	clients map[string]ExternalRegistryClient // 注册中心客户端映射: serviceKey -> client
	mutex   sync.RWMutex
}

// NewExternalRegistryCacheManager 创建外部注册中心缓存管理器
func NewExternalRegistryCacheManager() *ExternalRegistryCacheManager {
	return &ExternalRegistryCacheManager{
		clients: make(map[string]ExternalRegistryClient),
	}
}

// GetClient 获取外部注册中心客户端，如果不存在则根据服务创建
func (m *ExternalRegistryCacheManager) GetClient(service *core.Service) (ExternalRegistryClient, error) {
	if service.IsInternalRegistry() {
		return nil, fmt.Errorf("内部注册中心不需要外部客户端")
	}

	serviceKey := generateServiceCacheKey(service)

	// 先尝试获取现有客户端
	m.mutex.RLock()
	client, exists := m.clients[serviceKey]
	m.mutex.RUnlock()

	if exists {
		return client, nil
	}

	// 如果不存在，创建新客户端
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 双重检查，防止并发创建
	if client, exists := m.clients[serviceKey]; exists {
		return client, nil
	}

	// 根据注册类型创建客户端
	switch service.RegistryType {
	case core.RegistryTypeNacos:
		client = NewNacosRegistryClient()
	case core.RegistryTypeConsul:
		// TODO: 实现 NewConsulRegistryClient
		return nil, fmt.Errorf("Consul注册中心客户端暂未实现")
	case core.RegistryTypeEureka:
		// TODO: 实现 NewEurekaRegistryClient
		return nil, fmt.Errorf("Eureka注册中心客户端暂未实现")
	case core.RegistryTypeEtcd:
		// TODO: 实现 NewEtcdRegistryClient
		return nil, fmt.Errorf("ETCD注册中心客户端暂未实现")
	case core.RegistryTypeZookeeper:
		// TODO: 实现 NewZookeeperRegistryClient
		return nil, fmt.Errorf("ZooKeeper注册中心客户端暂未实现")
	default:
		return nil, fmt.Errorf("不支持的注册中心类型: %s", service.RegistryType)
	}

	// 缓存客户端
	m.clients[serviceKey] = client

	logger.Info("根据服务创建外部注册中心客户端",
		"serviceKey", serviceKey,
		"registryType", service.RegistryType)

	return client, nil
}

// Close 关闭所有外部注册中心客户端
func (m *ExternalRegistryCacheManager) Close() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for serviceKey, client := range m.clients {
		if err := client.Close(); err != nil {
			logger.Warn("关闭外部注册中心客户端失败",
				"serviceKey", serviceKey,
				"error", err)
		}
	}

	m.clients = make(map[string]ExternalRegistryClient)
	logger.Info("所有外部注册中心客户端已关闭")

	return nil
}

// RegisterService 根据注册类型注册服务
func (m *ExternalRegistryCacheManager) RegisterService(ctx context.Context, service *core.Service) error {
	if service.IsInternalRegistry() {
		return nil // 内部注册中心不需要处理
	}

	// 获取客户端并注册服务
	client, err := m.GetClient(service)
	if err != nil {
		return fmt.Errorf("获取注册中心客户端失败: %w", err)
	}

	// 注册服务
	if err := client.RegisterService(ctx, service); err != nil {
		return fmt.Errorf("注册服务到外部注册中心失败: %w", err)
	}

	logger.InfoWithTrace(ctx, "服务注册到外部注册中心成功",
		"serviceName", service.ServiceName,
		"registryType", service.RegistryType,
		"instanceCount", len(service.Instances))

	return nil
}

// DeregisterService 根据注册类型注销服务并清理资源
func (m *ExternalRegistryCacheManager) DeregisterService(ctx context.Context, service *core.Service) error {
	if service.IsInternalRegistry() {
		return nil // 内部注册中心不需要处理
	}

	serviceKey := generateServiceCacheKey(service)

	// 获取客户端
	m.mutex.RLock()
	client, exists := m.clients[serviceKey]
	m.mutex.RUnlock()

	if !exists {
		// 客户端不存在，视为注销成功
		logger.DebugWithTrace(ctx, "外部注册中心客户端不存在，跳过服务注销",
			"serviceName", service.ServiceName,
			"registryType", service.RegistryType)
		return nil
	}
	// 注销服务
	if err := client.DeregisterService(ctx, service); err != nil {
		logger.WarnWithTrace(ctx, "从外部注册中心注销服务失败",
			"serviceName", service.ServiceName,
			"registryType", service.RegistryType,
			"error", err)
	}

	// 完全关闭客户端并清理所有资源
	if err := client.Close(); err != nil {
		logger.WarnWithTrace(ctx, "关闭外部注册中心客户端失败",
			"serviceName", service.ServiceName,
			"registryType", service.RegistryType,
			"error", err)
	}

	// 从缓存中移除客户端
	m.mutex.Lock()
	delete(m.clients, serviceKey)
	m.mutex.Unlock()

	logger.InfoWithTrace(ctx, "服务从外部注册中心注销并清理资源成功",
		"serviceName", service.ServiceName,
		"registryType", service.RegistryType,
		"instanceCount", len(service.Instances))

	return nil
}

// RegisterInstance 根据服务注册实例
func (m *ExternalRegistryCacheManager) RegisterInstance(ctx context.Context, instance *core.ServiceInstance, service *core.Service) error {
	if service.IsInternalRegistry() {
		return nil // 内部注册中心不需要处理
	}

	// 获取客户端并注册实例
	client, err := m.GetClient(service)
	if err != nil {
		return fmt.Errorf("获取注册中心客户端失败: %w", err)
	}
	// 注册实例
	if err := client.RegisterInstance(ctx, instance); err != nil {
		return fmt.Errorf("注册实例到外部注册中心失败: %w", err)
	}

	logger.InfoWithTrace(ctx, "实例注册到外部注册中心成功",
		"instanceId", instance.ServiceInstanceId,
		"registryType", service.RegistryType)

	return nil
}

// DeregisterInstance 根据服务注销实例
func (m *ExternalRegistryCacheManager) DeregisterInstance(ctx context.Context, instance *core.ServiceInstance, service *core.Service) error {
	if service.IsInternalRegistry() {
		return nil // 内部注册中心不需要处理
	}

	client, err := m.GetClient(service)
	if err != nil {
		return fmt.Errorf("获取注册中心客户端失败: %w", err)
	}

	// 注销实例
	if err := client.DeregisterInstance(ctx, instance); err != nil {
		return fmt.Errorf("从外部注册中心注销实例失败: %w", err)
	}

	logger.InfoWithTrace(ctx, "实例从外部注册中心注销成功",
		"instanceId", instance.ServiceInstanceId,
		"registryType", service.RegistryType)

	return nil
}

// GetServiceInstances 根据注册类型获取服务实例
func (m *ExternalRegistryCacheManager) GetServiceInstances(ctx context.Context, service *core.Service) ([]*core.ServiceInstance, error) {
	if service.IsInternalRegistry() {
		return make([]*core.ServiceInstance, 0), nil
	}

	// 获取外部注册中心客户端
	client, err := m.GetClient(service)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取外部注册中心客户端失败",
			"serviceName", service.ServiceName,
			"registryType", service.RegistryType,
			"error", err)
		return nil, fmt.Errorf("获取外部注册中心客户端失败: %w", err)
	}

	// 从外部注册中心获取实例
	instances, err := client.DiscoverInstances(ctx, service.ServiceName, service.GroupName)
	if err != nil {
		logger.ErrorWithTrace(ctx, "从外部注册中心获取实例失败",
			"serviceName", service.ServiceName,
			"registryType", service.RegistryType,
			"error", err)
		return nil, fmt.Errorf("从外部注册中心获取实例失败: %w", err)
	}

	logger.DebugWithTrace(ctx, "从外部注册中心获取实例成功",
		"serviceName", service.ServiceName,
		"registryType", service.RegistryType,
		"instanceCount", len(instances))

	return instances, nil
}

// DiscoverHealthyInstance 根据注册类型发现健康实例
func (m *ExternalRegistryCacheManager) DiscoverHealthyInstance(ctx context.Context, service *core.Service) (*core.ServiceInstance, error) {
	if service.IsInternalRegistry() {
		return nil, fmt.Errorf("不是外部注册中心服务")
	}

	// 从外部注册中心获取所有实例
	instances, err := m.GetServiceInstances(ctx, service)
	if err != nil {
		return nil, fmt.Errorf("从外部注册中心获取实例失败: %w", err)
	}

	if len(instances) == 0 {
		return nil, fmt.Errorf("外部注册中心中没有实例")
	}

	// 简单选择第一个健康实例
	for _, instance := range instances {
		if instance.HealthStatus == core.HealthStatusHealthy {
			logger.DebugWithTrace(ctx, "从外部注册中心发现健康实例",
				"serviceName", service.ServiceName,
				"registryType", service.RegistryType,
				"instanceId", instance.ServiceInstanceId,
				"address", fmt.Sprintf("%s:%d", instance.HostAddress, instance.PortNumber))
			return instance, nil
		}
	}

	return nil, fmt.Errorf("外部注册中心中没有健康实例")
}

// generateServiceCacheKey 生成服务的缓存键
func generateServiceCacheKey(service *core.Service) string {
	// 使用租户ID、服务组ID、服务名称和注册类型组合作为缓存键
	return fmt.Sprintf("%s:%s:%s", service.TenantId, service.ServiceGroupId, service.ServiceName)
}
