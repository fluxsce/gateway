package manager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gateway/internal/registry/cache"
	"gateway/internal/registry/core"
	"gateway/internal/registry/service"
	"gateway/pkg/logger"
	"gateway/pkg/utils/random"
)

// RegistryManager 注册中心管理器实现
// 负责各组件的生命周期管理，包括初始化、启动和停止
// 全局单例，对外提供统一的访问入口
type RegistryManager struct {
	eventPublisher core.EventPublisher
	healthChecker  core.HealthChecker
	cache          core.CacheStorage

	// 服务状态
	isReady   bool
	isRunning bool
	mutex     sync.RWMutex
}

// 全局单例
var (
	instance *RegistryManager
	once     sync.Once
)

// GetManager 获取注册中心管理器实例（全局单例）
// 如果实例不存在，则创建新实例
// 如果实例已存在，则直接返回现有实例
func GetManager(cache core.CacheStorage, eventPublisher core.EventPublisher) core.Manager {
	once.Do(func() {
		if cache == nil {
			panic("cache 不能为空")
		}

		// 创建健康检查器
		healthChecker := service.NewHealthMonitor(cache, eventPublisher)

		// 创建管理器实例
		instance = &RegistryManager{
			eventPublisher: eventPublisher,
			healthChecker:  healthChecker,
			cache:          cache,
			isReady:        false,
			isRunning:      false,
		}

		logger.Info("创建注册中心管理器全局单例")
	})

	return instance
}

// GetInstance 获取已初始化的实例
// 如果实例还未初始化，则尝试使用默认配置创建实例
func GetInstance() *RegistryManager {
	if instance == nil {
		// 使用默认配置创建实例
		cache := cache.NewMemoryCache()
		return GetManager(cache, nil).(*RegistryManager)
	}
	return instance
}

// ResetInstance 重置全局单例（仅用于测试）
func ResetInstance() {
	instance = nil
	once = sync.Once{}
}

// InitInstance 使用指定的缓存和事件发布器初始化实例
// 这是一个便捷方法，用于在应用启动时初始化注册中心管理器
func InitInstance(cache core.CacheStorage, eventPublisher core.EventPublisher) core.Manager {
	if instance == nil {
		return GetManager(cache, eventPublisher)
	}
	return instance
}

// Start 启动注册中心服务
// 按照依赖顺序启动各个组件
func (m *RegistryManager) Start(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.isRunning {
		logger.Info("注册中心管理器已经在运行中")
		return nil
	}

	// 1. 启动事件发布器（如果存在）
	if m.eventPublisher != nil {
		err := m.eventPublisher.Start(ctx)
		if err != nil {
			return fmt.Errorf("启动事件发布器失败: %w", err)
		}
		logger.InfoWithTrace(ctx, "事件发布器启动成功")
	}

	// 2. 启动健康检查器
	err := m.healthChecker.Start(ctx)
	if err != nil {
		return fmt.Errorf("启动健康检查器失败: %w", err)
	}
	logger.InfoWithTrace(ctx, "健康检查器启动成功")

	// 标记服务为运行状态
	m.isRunning = true
	m.isReady = true

	logger.InfoWithTrace(ctx, "注册中心管理器启动成功")
	return nil
}

// Stop 停止注册中心服务
// 优雅停止所有组件
func (m *RegistryManager) Stop(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.isRunning {
		logger.Info("注册中心管理器未在运行")
		return nil
	}

	var errs []error

	// 1. 停止健康检查器
	err := m.healthChecker.Stop(ctx)
	if err != nil {
		errs = append(errs, fmt.Errorf("停止健康检查器失败: %w", err))
		logger.WarnWithTrace(ctx, "停止健康检查器失败", "error", err)
	} else {
		logger.InfoWithTrace(ctx, "健康检查器停止成功")
	}

	// 2. 停止事件发布器（如果存在）
	if m.eventPublisher != nil {
		err = m.eventPublisher.Stop(ctx)
		if err != nil {
			errs = append(errs, fmt.Errorf("停止事件发布器失败: %w", err))
			logger.WarnWithTrace(ctx, "停止事件发布器失败", "error", err)
		} else {
			logger.InfoWithTrace(ctx, "事件发布器停止成功")
		}
	}

	// 标记服务为非运行状态
	m.isRunning = false
	m.isReady = false

	logger.InfoWithTrace(ctx, "注册中心管理器停止成功")

	// 如果有错误，返回第一个错误
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

// GetEventPublisher 获取事件发布器实例
func (m *RegistryManager) GetEventPublisher() core.EventPublisher {
	return m.eventPublisher
}

// GetHealthChecker 获取健康检查器实例
func (m *RegistryManager) GetHealthChecker() core.HealthChecker {
	return m.healthChecker
}

// IsReady 检查服务是否就绪
func (m *RegistryManager) IsReady() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.isReady
}

// GetCache 获取缓存实例
func (m *RegistryManager) GetCache() core.CacheStorage {
	return m.cache
}

// RegisterService 注册服务
// 注册服务信息并发布相应事件
func (m *RegistryManager) RegisterService(ctx context.Context, service *core.Service) (*core.Service, error) {
	if service == nil {
		return nil, fmt.Errorf("服务不能为空")
	}

	// 检查服务是否已存在
	existingService, err := m.cache.GetService(ctx, service.TenantId, service.ServiceGroupId, service.ServiceName)
	isNewService := err != nil || existingService == nil

	// 设置服务到缓存
	err = m.cache.SetService(ctx, service.TenantId, service)
	if err != nil {
		return nil, fmt.Errorf("注册服务失败: %w", err)
	}

	// 发布服务注册或更新事件
	if m.eventPublisher != nil {
		eventType := core.EventTypeServiceUpdated
		if isNewService {
			eventType = core.EventTypeServiceRegistered
		}

		// 生成事件ID
		eventId := random.Generate32BitRandomString()

		// 构建事件数据JSON
		eventDataJson := fmt.Sprintf(`{
			"operation": "%s",
			"serviceName": "%s", 
			"tenantId": "%s",
			"serviceGroupId": "%s",
			"groupName": "%s",
			"protocolType": "%s",
			"loadBalanceStrategy": "%s",
			"healthCheckType": "%s",
			"healthCheckMode": "%s",
			"isNewService": %t
		}`, eventType, service.ServiceName, service.TenantId, service.ServiceGroupId,
			service.GroupName, service.ProtocolType, service.LoadBalanceStrategy,
			service.HealthCheckType, service.HealthCheckMode, isNewService)

		// 构建详细的事件消息
		var eventMessage string
		if isNewService {
			eventMessage = fmt.Sprintf("新服务注册成功: %s (租户: %s, 分组: %s, 协议: %s)",
				service.ServiceName, service.TenantId, service.GroupName, service.ProtocolType)
		} else {
			eventMessage = fmt.Sprintf("服务更新成功: %s (租户: %s, 分组: %s, 协议: %s)",
				service.ServiceName, service.TenantId, service.GroupName, service.ProtocolType)
		}

		now := time.Now()
		event := &core.ServiceEvent{
			ServiceEventId:    eventId,
			EventType:         eventType,
			TenantId:          service.TenantId,
			ServiceGroupId:    service.ServiceGroupId,
			ServiceInstanceId: "", // 服务级事件，实例ID为空
			GroupName:         service.GroupName,
			ServiceName:       service.ServiceName,
			HostAddress:       "", // 服务级事件，主机地址为空
			PortNumber:        0,  // 服务级事件，端口为0
			NodeIpAddress:     random.GetNodeIP(),
			EventTime:         service.AddTime,
			EventSource:       core.GetEventSourceFromContext(ctx, core.EventSourceRegistryManager),
			EventDataJson:     eventDataJson,
			EventMessage:      eventMessage,
			Service:           service, // 补充完整的服务对象
			// 通用字段
			AddTime:        now,
			AddWho:         "SYSTEM",
			EditTime:       now,
			EditWho:        "SYSTEM",
			OprSeqFlag:     eventId,
			CurrentVersion: 1,
			ActiveFlag:     "Y",
		}
		_ = m.eventPublisher.Publish(ctx, event)

		logger.InfoWithTrace(ctx, "服务事件已发布",
			"eventType", eventType,
			"serviceName", service.ServiceName,
			"eventId", eventId)
	}

	logger.InfoWithTrace(ctx, "服务注册成功",
		"serviceName", service.ServiceName,
		"tenantId", service.TenantId,
		"groupName", service.GroupName,
		"isNewService", isNewService)

	return service, nil
}

// RegisterInstance 注册服务实例
// 便捷方法，直接调用缓存存储
func (m *RegistryManager) RegisterInstance(ctx context.Context, instance *core.ServiceInstance) (*core.ServiceInstance, error) {
	if instance == nil {
		return nil, fmt.Errorf("服务实例不能为空")
	}

	// 设置实例到缓存
	err := m.cache.SetInstance(ctx, instance.TenantId, instance)
	if err != nil {
		return nil, fmt.Errorf("注册服务实例失败: %w", err)
	}

	// 发布实例注册事件
	if m.eventPublisher != nil {
		// 生成事件ID
		eventId := random.Generate32BitRandomString()

		// 构建事件数据JSON
		eventDataJson := fmt.Sprintf(`{
			"operation": "INSTANCE_REGISTERED",
			"serviceInstanceId": "%s",
			"serviceName": "%s", 
			"tenantId": "%s",
			"serviceGroupId": "%s",
			"groupName": "%s",
			"hostAddress": "%s",
			"portNumber": %d,
			"instanceStatus": "%s",
			"healthStatus": "%s",
			"weightValue": %d,
			"clientType": "%s",
			"tempInstanceFlag": "%s"
		}`, instance.ServiceInstanceId, instance.ServiceName, instance.TenantId,
			instance.ServiceGroupId, instance.GroupName, instance.HostAddress,
			instance.PortNumber, instance.InstanceStatus, instance.HealthStatus,
			instance.WeightValue, instance.ClientType, instance.TempInstanceFlag)

		// 构建详细的事件消息
		eventMessage := fmt.Sprintf("服务实例注册成功: %s (服务: %s, 地址: %s:%d, 状态: %s)",
			instance.ServiceInstanceId, instance.ServiceName,
			instance.HostAddress, instance.PortNumber, instance.InstanceStatus)

		now := time.Now()
		event := &core.ServiceEvent{
			ServiceEventId:    eventId,
			EventType:         core.EventTypeInstanceRegistered,
			TenantId:          instance.TenantId,
			ServiceGroupId:    instance.ServiceGroupId,
			ServiceInstanceId: instance.ServiceInstanceId,
			GroupName:         instance.GroupName,
			ServiceName:       instance.ServiceName,
			HostAddress:       instance.HostAddress,
			PortNumber:        instance.PortNumber,
			NodeIpAddress:     random.GetNodeIP(),
			EventTime:         instance.AddTime,
			EventSource:       core.GetEventSourceFromContext(ctx, core.EventSourceRegistryManager),
			EventDataJson:     eventDataJson,
			EventMessage:      eventMessage,
			Instance:          instance, // 补充完整的实例对象
			// 通用字段
			AddTime:        now,
			AddWho:         "SYSTEM",
			EditTime:       now,
			EditWho:        "SYSTEM",
			OprSeqFlag:     eventId,
			CurrentVersion: 1,
			ActiveFlag:     "Y",
		}
		_ = m.eventPublisher.Publish(ctx, event)
	}

	return instance, nil
}

// DeregisterService 注销服务
// 注销服务信息并发布相应事件
func (m *RegistryManager) DeregisterService(ctx context.Context, tenantId, serviceGroupId, serviceName string) error {
	if tenantId == "" || serviceGroupId == "" || serviceName == "" {
		return fmt.Errorf("租户ID、服务组ID和服务名不能为空")
	}

	// 获取服务信息用于发布事件
	service, _ := m.cache.GetService(ctx, tenantId, serviceGroupId, serviceName)

	// 从缓存中删除服务
	err := m.cache.DeleteService(ctx, tenantId, serviceGroupId, serviceName)
	if err != nil {
		return fmt.Errorf("注销服务失败: %w", err)
	}

	// 发布服务注销事件
	if m.eventPublisher != nil && service != nil {
		// 生成事件ID
		eventId := random.Generate32BitRandomString()

		// 构建事件数据JSON
		eventDataJson := fmt.Sprintf(`{
			"operation": "SERVICE_DEREGISTERED",
			"serviceName": "%s", 
			"tenantId": "%s",
			"serviceGroupId": "%s",
			"groupName": "%s",
			"protocolType": "%s",
			"loadBalanceStrategy": "%s",
			"healthCheckType": "%s",
			"healthCheckMode": "%s"
		}`, serviceName, tenantId, serviceGroupId, service.GroupName,
			service.ProtocolType, service.LoadBalanceStrategy,
			service.HealthCheckType, service.HealthCheckMode)

		// 构建详细的事件消息
		eventMessage := fmt.Sprintf("服务注销成功: %s (租户: %s, 分组: %s, 协议: %s)",
			serviceName, tenantId, service.GroupName, service.ProtocolType)

		now := time.Now()
		event := &core.ServiceEvent{
			ServiceEventId:    eventId,
			EventType:         core.EventTypeServiceDeregistered,
			TenantId:          tenantId,
			ServiceGroupId:    serviceGroupId,
			ServiceInstanceId: "", // 服务级事件，实例ID为空
			GroupName:         service.GroupName,
			ServiceName:       serviceName,
			HostAddress:       "", // 服务级事件，主机地址为空
			PortNumber:        0,  // 服务级事件，端口为0
			NodeIpAddress:     random.GetNodeIP(),
			EventTime:         service.EditTime,
			EventSource:       core.GetEventSourceFromContext(ctx, core.EventSourceRegistryManager),
			EventDataJson:     eventDataJson,
			EventMessage:      eventMessage,
			Service:           service, // 补充完整的服务对象
			// 通用字段
			AddTime:        now,
			AddWho:         "SYSTEM",
			EditTime:       now,
			EditWho:        "SYSTEM",
			OprSeqFlag:     eventId,
			CurrentVersion: 1,
			ActiveFlag:     "Y",
		}
		_ = m.eventPublisher.Publish(ctx, event)

		logger.InfoWithTrace(ctx, "服务注销事件已发布",
			"serviceName", serviceName,
			"eventId", eventId)
	}

	logger.InfoWithTrace(ctx, "服务注销成功",
		"serviceName", serviceName,
		"tenantId", tenantId,
		"serviceGroupId", serviceGroupId)

	return nil
}

// DeregisterInstance 注销服务实例
// 便捷方法，直接调用缓存存储
func (m *RegistryManager) DeregisterInstance(ctx context.Context, tenantId, instanceId string) error {
	if tenantId == "" || instanceId == "" {
		return fmt.Errorf("租户ID和实例ID不能为空")
	}

	// 获取实例信息用于发布事件
	instance, _ := m.cache.GetInstance(ctx, tenantId, instanceId)

	// 从缓存中删除实例
	err := m.cache.DeleteInstance(ctx, tenantId, instanceId)
	if err != nil {
		return fmt.Errorf("注销服务实例失败: %w", err)
	}

	// 发布实例注销事件
	if m.eventPublisher != nil && instance != nil {
		// 生成事件ID
		eventId := random.Generate32BitRandomString()

		// 构建事件数据JSON
		eventDataJson := fmt.Sprintf(`{
			"operation": "INSTANCE_DEREGISTERED",
			"serviceInstanceId": "%s",
			"serviceName": "%s", 
			"tenantId": "%s",
			"serviceGroupId": "%s",
			"groupName": "%s",
			"hostAddress": "%s",
			"portNumber": %d,
			"instanceStatus": "%s",
			"healthStatus": "%s",
			"weightValue": %d,
			"clientType": "%s",
			"tempInstanceFlag": "%s"
		}`, instanceId, instance.ServiceName, tenantId,
			instance.ServiceGroupId, instance.GroupName, instance.HostAddress,
			instance.PortNumber, instance.InstanceStatus, instance.HealthStatus,
			instance.WeightValue, instance.ClientType, instance.TempInstanceFlag)

		// 构建详细的事件消息
		eventMessage := fmt.Sprintf("服务实例注销成功: %s (服务: %s, 地址: %s:%d)",
			instanceId, instance.ServiceName, instance.HostAddress, instance.PortNumber)

		now := time.Now()
		event := &core.ServiceEvent{
			ServiceEventId:    eventId,
			EventType:         core.EventTypeInstanceDeregistered,
			TenantId:          tenantId,
			ServiceGroupId:    instance.ServiceGroupId,
			ServiceInstanceId: instanceId,
			GroupName:         instance.GroupName,
			ServiceName:       instance.ServiceName,
			HostAddress:       instance.HostAddress,
			PortNumber:        instance.PortNumber,
			NodeIpAddress:     random.GetNodeIP(),
			EventTime:         instance.EditTime,
			EventSource:       core.GetEventSourceFromContext(ctx, core.EventSourceRegistryManager),
			EventDataJson:     eventDataJson,
			EventMessage:      eventMessage,
			Instance:          instance, // 补充完整的实例对象
			// 通用字段
			AddTime:        now,
			AddWho:         "SYSTEM",
			EditTime:       now,
			EditWho:        "SYSTEM",
			OprSeqFlag:     eventId,
			CurrentVersion: 1,
			ActiveFlag:     "Y",
		}
		_ = m.eventPublisher.Publish(ctx, event)
	}

	return nil
}

// GetService 获取服务详情
// 便捷方法，直接调用缓存存储
func (m *RegistryManager) GetService(ctx context.Context, tenantId, serviceGroupId, serviceName string) (*core.Service, error) {
	if tenantId == "" || serviceGroupId == "" || serviceName == "" {
		return nil, fmt.Errorf("租户ID、服务组ID和服务名不能为空")
	}

	// 从缓存中获取服务
	service, err := m.cache.GetService(ctx, tenantId, serviceGroupId, serviceName)
	if err != nil {
		logger.DebugWithTrace(ctx, "获取服务失败",
			"tenantId", tenantId,
			"serviceGroupId", serviceGroupId,
			"serviceName", serviceName,
			"error", err)
		return nil, fmt.Errorf("获取服务失败: %w", err)
	}

	logger.DebugWithTrace(ctx, "获取服务成功",
		"tenantId", tenantId,
		"serviceGroupId", serviceGroupId,
		"serviceName", serviceName)

	return service, nil
}

// ListServices 列出服务
// 便捷方法，根据服务组ID列出服务
func (m *RegistryManager) ListServices(ctx context.Context, tenantId, serviceGroupId string) ([]*core.Service, error) {
	if tenantId == "" {
		return nil, fmt.Errorf("租户ID不能为空")
	}

	// 如果指定了服务组ID，直接从该组获取服务
	if serviceGroupId != "" {
		return m.cache.ListServices(ctx, tenantId, serviceGroupId)
	}

	// 如果没有指定服务组ID，获取所有服务组下的服务
	var allServices []*core.Service

	serviceGroups, err := m.cache.ListServiceGroups(ctx, tenantId)
	if err != nil {
		return nil, fmt.Errorf("获取服务组列表失败: %w", err)
	}

	for _, group := range serviceGroups {
		services, err := m.cache.ListServices(ctx, tenantId, group.ServiceGroupId)
		if err != nil {
			// 记录错误但继续处理其他组
			logger.WarnWithTrace(ctx, "获取服务组服务列表失败",
				"tenantId", tenantId,
				"serviceGroupId", group.ServiceGroupId,
				"error", err)
			continue
		}
		allServices = append(allServices, services...)
	}

	return allServices, nil
}

// DiscoverInstance 发现一个健康的服务实例
// 使用负载均衡策略选择最合适的实例
func (m *RegistryManager) DiscoverInstance(ctx context.Context, tenantId, serviceGroupId, serviceName string) (*core.ServiceInstance, error) {
	if tenantId == "" || serviceName == "" {
		return nil, fmt.Errorf("租户ID和服务名不能为空")
	}

	if serviceGroupId == "" {
		return nil, fmt.Errorf("服务组ID不能为空")
	}

	// 调用缓存的服务发现方法
	instance, err := m.cache.DiscoverInstance(ctx, tenantId, serviceGroupId, serviceName)
	if err != nil {
		logger.WarnWithTrace(ctx, "服务发现失败",
			"tenantId", tenantId,
			"serviceGroupId", serviceGroupId,
			"serviceName", serviceName,
			"error", err)
		return nil, fmt.Errorf("服务发现失败: %w", err)
	}

	logger.DebugWithTrace(ctx, "服务发现成功",
		"tenantId", tenantId,
		"serviceGroupId", serviceGroupId,
		"serviceName", serviceName,
		"instanceId", instance.ServiceInstanceId,
		"address", fmt.Sprintf("%s:%d", instance.HostAddress, instance.PortNumber))

	return instance, nil
}

// GetServiceGroup 获取服务组信息
// 便捷方法，直接调用缓存存储
func (m *RegistryManager) GetServiceGroup(ctx context.Context, tenantId, serviceGroupId string) (*core.ServiceGroup, error) {
	if tenantId == "" || serviceGroupId == "" {
		return nil, fmt.Errorf("租户ID和服务组ID不能为空")
	}

	// 从缓存中直接获取服务组
	group, err := m.cache.GetServiceGroup(ctx, tenantId, serviceGroupId)
	if err != nil {
		logger.DebugWithTrace(ctx, "获取服务组失败",
			"tenantId", tenantId,
			"serviceGroupId", serviceGroupId,
			"error", err)
		return nil, fmt.Errorf("获取服务组失败: %w", err)
	}

	logger.DebugWithTrace(ctx, "获取服务组成功",
		"tenantId", tenantId,
		"serviceGroupId", serviceGroupId,
		"groupName", group.GroupName)

	return group, nil
}

// ListServiceGroups 列出租户下的所有服务组
// 便捷方法，直接调用缓存存储
func (m *RegistryManager) ListServiceGroups(ctx context.Context, tenantId string) ([]*core.ServiceGroup, error) {
	if tenantId == "" {
		return nil, fmt.Errorf("租户ID不能为空")
	}

	// 从缓存中获取服务组列表
	groups, err := m.cache.ListServiceGroups(ctx, tenantId)
	if err != nil {
		logger.DebugWithTrace(ctx, "获取服务组列表失败",
			"tenantId", tenantId,
			"error", err)
		return nil, fmt.Errorf("获取服务组列表失败: %w", err)
	}

	logger.DebugWithTrace(ctx, "获取服务组列表成功",
		"tenantId", tenantId,
		"count", len(groups))

	return groups, nil
}

// GetInstance 获取服务实例
// 便捷方法，直接调用缓存存储
func (m *RegistryManager) GetInstance(ctx context.Context, tenantId, instanceId string) (*core.ServiceInstance, error) {
	if tenantId == "" || instanceId == "" {
		return nil, fmt.Errorf("租户ID和实例ID不能为空")
	}

	// 从缓存中获取实例
	instance, err := m.cache.GetInstance(ctx, tenantId, instanceId)
	if err != nil {
		logger.DebugWithTrace(ctx, "获取服务实例失败",
			"tenantId", tenantId,
			"instanceId", instanceId,
			"error", err)
		return nil, fmt.Errorf("获取服务实例失败: %w", err)
	}

	logger.DebugWithTrace(ctx, "获取服务实例成功",
		"tenantId", tenantId,
		"instanceId", instanceId,
		"serviceName", instance.ServiceName,
		"address", fmt.Sprintf("%s:%d", instance.HostAddress, instance.PortNumber))

	return instance, nil
}

// ListInstances 列出服务下的所有实例
// 便捷方法，直接调用缓存存储
func (m *RegistryManager) ListInstances(ctx context.Context, tenantId, serviceGroupId, serviceName string) ([]*core.ServiceInstance, error) {
	if tenantId == "" || serviceGroupId == "" || serviceName == "" {
		return nil, fmt.Errorf("租户ID、服务组ID和服务名不能为空")
	}

	// 从缓存中获取实例列表
	instances, err := m.cache.ListInstances(ctx, tenantId, serviceGroupId, serviceName)
	if err != nil {
		logger.DebugWithTrace(ctx, "获取服务实例列表失败",
			"tenantId", tenantId,
			"serviceGroupId", serviceGroupId,
			"serviceName", serviceName,
			"error", err)
		return nil, fmt.Errorf("获取服务实例列表失败: %w", err)
	}

	logger.DebugWithTrace(ctx, "获取服务实例列表成功",
		"tenantId", tenantId,
		"serviceGroupId", serviceGroupId,
		"serviceName", serviceName,
		"count", len(instances))

	return instances, nil
}

// UpdateInstanceHeartbeat 更新实例心跳时间
// 更新实例的最后心跳时间，并发布心跳更新事件
func (m *RegistryManager) UpdateInstanceHeartbeat(ctx context.Context, tenantId, instanceId string) error {
	if tenantId == "" || instanceId == "" {
		return fmt.Errorf("租户ID和实例ID不能为空")
	}

	// 获取实例信息
	instance, err := m.cache.GetInstance(ctx, tenantId, instanceId)
	if err != nil {
		return fmt.Errorf("获取实例信息失败: %w", err)
	}

	if instance == nil {
		return fmt.Errorf("实例不存在: %s", instanceId)
	}

	// 记录原始心跳时间和健康状态以便判断是否发生变化
	oldHeartbeatTime := instance.LastHeartbeatTime
	oldHealthStatus := instance.HealthStatus

	// 更新心跳时间、健康状态和最后检查时间
	now := time.Now()
	instance.LastHeartbeatTime = &now
	instance.LastHealthCheckTime = &now
	instance.EditTime = now

	// 心跳正常时，设置健康状态为健康，并重置失败次数
	instance.HealthStatus = core.HealthStatusHealthy
	instance.HeartbeatFailCount = 0

	// 更新缓存中的实例
	err = m.cache.SetInstance(ctx, tenantId, instance)
	if err != nil {
		return fmt.Errorf("更新实例心跳时间失败: %w", err)
	}

	// 发布心跳更新事件
	if m.eventPublisher != nil {
		// 生成事件ID
		eventId := random.Generate32BitRandomString()

		// 构建事件数据JSON
		var oldHeartbeatStr string
		if oldHeartbeatTime != nil {
			oldHeartbeatStr = oldHeartbeatTime.Format("2006-01-02 15:04:05")
		} else {
			oldHeartbeatStr = "null"
		}

		eventDataJson := fmt.Sprintf(`{
			"operation": "INSTANCE_HEARTBEAT_UPDATED",
			"serviceInstanceId": "%s",
			"serviceName": "%s", 
			"tenantId": "%s",
			"serviceGroupId": "%s",
			"groupName": "%s",
			"hostAddress": "%s",
			"portNumber": %d,
			"instanceStatus": "%s",
			"oldHealthStatus": "%s",
			"newHealthStatus": "%s",
			"heartbeatFailCount": %d,
			"oldHeartbeatTime": "%s",
			"newHeartbeatTime": "%s",
			"lastHealthCheckTime": "%s",
			"isFirstHeartbeat": %t,
			"healthStatusChanged": %t
		}`, instanceId, instance.ServiceName, tenantId,
			instance.ServiceGroupId, instance.GroupName, instance.HostAddress,
			instance.PortNumber, instance.InstanceStatus, oldHealthStatus, instance.HealthStatus,
			instance.HeartbeatFailCount, oldHeartbeatStr, now.Format("2006-01-02 15:04:05"),
			now.Format("2006-01-02 15:04:05"), oldHeartbeatTime == nil, oldHealthStatus != instance.HealthStatus)

		// 构建详细的事件消息
		var eventMessage string
		if oldHeartbeatTime != nil {
			if oldHealthStatus != instance.HealthStatus {
				eventMessage = fmt.Sprintf("实例心跳更新: %s (服务: %s, 地址: %s:%d, 健康状态: %s->%s, 失败次数重置: %d->0, 上次心跳: %s)",
					instanceId, instance.ServiceName, instance.HostAddress, instance.PortNumber,
					oldHealthStatus, instance.HealthStatus, instance.HeartbeatFailCount,
					oldHeartbeatTime.Format("2006-01-02 15:04:05"))
			} else {
				eventMessage = fmt.Sprintf("实例心跳更新: %s (服务: %s, 地址: %s:%d, 健康状态: %s, 失败次数: %d, 上次心跳: %s)",
					instanceId, instance.ServiceName, instance.HostAddress, instance.PortNumber,
					instance.HealthStatus, instance.HeartbeatFailCount,
					oldHeartbeatTime.Format("2006-01-02 15:04:05"))
			}
		} else {
			eventMessage = fmt.Sprintf("实例首次心跳: %s (服务: %s, 地址: %s:%d, 健康状态: %s, 失败次数: %d)",
				instanceId, instance.ServiceName, instance.HostAddress, instance.PortNumber,
				instance.HealthStatus, instance.HeartbeatFailCount)
		}

		event := &core.ServiceEvent{
			ServiceEventId:    eventId,
			EventType:         core.EventTypeInstanceHeartbeatUpdated,
			TenantId:          tenantId,
			ServiceGroupId:    instance.ServiceGroupId,
			ServiceInstanceId: instanceId,
			GroupName:         instance.GroupName,
			ServiceName:       instance.ServiceName,
			HostAddress:       instance.HostAddress,
			PortNumber:        instance.PortNumber,
			NodeIpAddress:     random.GetNodeIP(),
			EventTime:         now,
			EventSource:       core.GetEventSourceFromContext(ctx, core.EventSourceRegistryManager),
			EventDataJson:     eventDataJson,
			EventMessage:      eventMessage,
			Instance:          instance, // 补充完整的实例对象
			// 通用字段
			AddTime:        now,
			AddWho:         "SYSTEM",
			EditTime:       now,
			EditWho:        "SYSTEM",
			OprSeqFlag:     eventId,
			CurrentVersion: 1,
			ActiveFlag:     "Y",
		}

		// 发布事件
		publishErr := m.eventPublisher.Publish(ctx, event)
		if publishErr != nil {
			logger.WarnWithTrace(ctx, "发布心跳更新事件失败",
				"instanceId", instanceId,
				"error", publishErr)
		} else {
			logger.DebugWithTrace(ctx, "发布心跳更新事件成功",
				"instanceId", instanceId,
				"eventId", eventId)
		}
	}

	// 记录心跳更新日志
	if oldHeartbeatTime != nil {
		if oldHealthStatus != instance.HealthStatus {
			logger.InfoWithTrace(ctx, "实例心跳时间已更新，健康状态已变更",
				"instanceId", instanceId,
				"serviceName", instance.ServiceName,
				"oldHeartbeatTime", oldHeartbeatTime.Format("2006-01-02 15:04:05"),
				"newHeartbeatTime", now.Format("2006-01-02 15:04:05"),
				"oldHealthStatus", oldHealthStatus,
				"newHealthStatus", instance.HealthStatus,
				"heartbeatFailCount", instance.HeartbeatFailCount)
		} else {
			logger.DebugWithTrace(ctx, "实例心跳时间已更新",
				"instanceId", instanceId,
				"serviceName", instance.ServiceName,
				"oldHeartbeatTime", oldHeartbeatTime.Format("2006-01-02 15:04:05"),
				"newHeartbeatTime", now.Format("2006-01-02 15:04:05"),
				"healthStatus", instance.HealthStatus,
				"heartbeatFailCount", instance.HeartbeatFailCount)
		}
	} else {
		logger.InfoWithTrace(ctx, "实例首次心跳时间已设置",
			"instanceId", instanceId,
			"serviceName", instance.ServiceName,
			"heartbeatTime", now.Format("2006-01-02 15:04:05"),
			"healthStatus", instance.HealthStatus,
			"heartbeatFailCount", instance.HeartbeatFailCount)
	}

	return nil
}

// UpdateInstance 更新服务实例信息
// 更新实例信息到缓存，并发布实例更新事件
func (m *RegistryManager) UpdateInstance(ctx context.Context, instance *core.ServiceInstance) (*core.ServiceInstance, error) {
	if instance == nil {
		return nil, fmt.Errorf("服务实例不能为空")
	}

	if instance.TenantId == "" || instance.ServiceInstanceId == "" {
		return nil, fmt.Errorf("租户ID和实例ID不能为空")
	}

	// 获取更新前的实例信息用于事件发布（暂时未使用，保留用于后续扩展）
	_, _ = m.cache.GetInstance(ctx, instance.TenantId, instance.ServiceInstanceId)

	// 更新实例到缓存
	err := m.cache.SetInstance(ctx, instance.TenantId, instance)
	if err != nil {
		return nil, fmt.Errorf("更新服务实例失败: %w", err)
	}

	// 发布实例更新事件
	if m.eventPublisher != nil {
		// 生成事件ID
		eventId := random.Generate32BitRandomString()

		// 构建事件数据JSON
		eventDataJson := fmt.Sprintf(`{
			"operation": "INSTANCE_UPDATED",
			"serviceInstanceId": "%s",
			"serviceName": "%s", 
			"tenantId": "%s",
			"serviceGroupId": "%s",
			"groupName": "%s",
			"hostAddress": "%s",
			"portNumber": %d,
			"instanceStatus": "%s",
			"healthStatus": "%s",
			"weightValue": %d,
			"clientType": "%s",
			"tempInstanceFlag": "%s"
		}`, instance.ServiceInstanceId, instance.ServiceName, instance.TenantId,
			instance.ServiceGroupId, instance.GroupName, instance.HostAddress,
			instance.PortNumber, instance.InstanceStatus, instance.HealthStatus,
			instance.WeightValue, instance.ClientType, instance.TempInstanceFlag)

		// 构建详细的事件消息
		eventMessage := fmt.Sprintf("服务实例更新成功: %s (服务: %s, 地址: %s:%d, 状态: %s)",
			instance.ServiceInstanceId, instance.ServiceName,
			instance.HostAddress, instance.PortNumber, instance.InstanceStatus)

		now := time.Now()
		event := &core.ServiceEvent{
			ServiceEventId:    eventId,
			EventType:         core.EventTypeInstanceUpdated,
			TenantId:          instance.TenantId,
			ServiceGroupId:    instance.ServiceGroupId,
			ServiceInstanceId: instance.ServiceInstanceId,
			GroupName:         instance.GroupName,
			ServiceName:       instance.ServiceName,
			HostAddress:       instance.HostAddress,
			PortNumber:        instance.PortNumber,
			NodeIpAddress:     random.GetNodeIP(),
			EventTime:         now,
			EventSource:       core.GetEventSourceFromContext(ctx, core.EventSourceRegistryManager),
			EventDataJson:     eventDataJson,
			EventMessage:      eventMessage,
			Instance:          instance, // 补充完整的实例对象
			// 通用字段
			AddTime:        now,
			AddWho:         "SYSTEM",
			EditTime:       now,
			EditWho:        "SYSTEM",
			OprSeqFlag:     eventId,
			CurrentVersion: 1,
			ActiveFlag:     "Y",
		}

		// 发布事件
		publishErr := m.eventPublisher.Publish(ctx, event)
		if publishErr != nil {
			logger.WarnWithTrace(ctx, "发布实例更新事件失败",
				"instanceId", instance.ServiceInstanceId,
				"error", publishErr)
		} else {
			logger.InfoWithTrace(ctx, "发布实例更新事件成功",
				"instanceId", instance.ServiceInstanceId,
				"serviceName", instance.ServiceName,
				"eventId", eventId)
		}
	}

	// 记录实例更新日志
	logger.InfoWithTrace(ctx, "服务实例信息已更新",
		"instanceId", instance.ServiceInstanceId,
		"serviceName", instance.ServiceName,
		"hostAddress", instance.HostAddress,
		"portNumber", instance.PortNumber,
		"instanceStatus", instance.InstanceStatus,
		"healthStatus", instance.HealthStatus)

	return instance, nil
}

// UpdateInstanceHealthStatus 更新服务实例健康状态
// 更新实例健康状态到缓存，并发布健康状态变更事件
func (m *RegistryManager) UpdateInstanceHealthStatus(ctx context.Context, tenantId, instanceId, healthStatus string, healthCheckTime time.Time) error {
	if tenantId == "" || instanceId == "" {
		return fmt.Errorf("租户ID和实例ID不能为空")
	}

	// 获取实例信息
	instance, err := m.cache.GetInstance(ctx, tenantId, instanceId)
	if err != nil {
		return fmt.Errorf("获取实例信息失败: %w", err)
	}

	if instance == nil {
		return fmt.Errorf("实例不存在: %s", instanceId)
	}

	// 记录原始健康状态以便判断是否发生变化
	oldHealthStatus := instance.HealthStatus

	// 更新健康状态和检查时间
	instance.HealthStatus = healthStatus
	instance.LastHealthCheckTime = &healthCheckTime
	instance.EditTime = time.Now()

	// 更新缓存中的实例
	err = m.cache.SetInstance(ctx, tenantId, instance)
	if err != nil {
		return fmt.Errorf("更新实例健康状态失败: %w", err)
	}

	// 发布健康状态变更事件
	if m.eventPublisher != nil {
		// 生成事件ID
		eventId := random.Generate32BitRandomString()

		// 构建事件数据JSON
		eventDataJson := fmt.Sprintf(`{
			"operation": "HEALTH_STATUS_UPDATED",
			"serviceInstanceId": "%s",
			"serviceName": "%s", 
			"tenantId": "%s",
			"serviceGroupId": "%s",
			"groupName": "%s",
			"hostAddress": "%s",
			"portNumber": %d,
			"oldHealthStatus": "%s",
			"newHealthStatus": "%s",
			"healthCheckTime": "%s"
		}`, instanceId, instance.ServiceName, tenantId,
			instance.ServiceGroupId, instance.GroupName, instance.HostAddress,
			instance.PortNumber, oldHealthStatus, healthStatus,
			healthCheckTime.Format("2006-01-02 15:04:05"))

		// 构建详细的事件消息
		var eventMessage string
		if oldHealthStatus != healthStatus {
			eventMessage = fmt.Sprintf("实例健康状态变更: %s (服务: %s, 地址: %s:%d, %s -> %s)",
				instanceId, instance.ServiceName, instance.HostAddress, instance.PortNumber,
				oldHealthStatus, healthStatus)
		} else {
			eventMessage = fmt.Sprintf("实例健康状态检查: %s (服务: %s, 地址: %s:%d, 状态: %s)",
				instanceId, instance.ServiceName, instance.HostAddress, instance.PortNumber,
				healthStatus)
		}

		now := time.Now()
		event := &core.ServiceEvent{
			ServiceEventId:    eventId,
			EventType:         core.EventTypeInstanceHealthChange,
			TenantId:          tenantId,
			ServiceGroupId:    instance.ServiceGroupId,
			ServiceInstanceId: instanceId,
			GroupName:         instance.GroupName,
			ServiceName:       instance.ServiceName,
			HostAddress:       instance.HostAddress,
			PortNumber:        instance.PortNumber,
			NodeIpAddress:     random.GetNodeIP(),
			EventTime:         now,
			EventSource:       core.GetEventSourceFromContext(ctx, core.EventSourceRegistryManager),
			EventDataJson:     eventDataJson,
			EventMessage:      eventMessage,
			Instance:          instance, // 补充完整的实例对象
			// 通用字段
			AddTime:        now,
			AddWho:         "SYSTEM",
			EditTime:       now,
			EditWho:        "SYSTEM",
			OprSeqFlag:     eventId,
			CurrentVersion: 1,
			ActiveFlag:     "Y",
		}

		// 发布事件
		publishErr := m.eventPublisher.Publish(ctx, event)
		if publishErr != nil {
			logger.WarnWithTrace(ctx, "发布健康状态变更事件失败",
				"instanceId", instanceId,
				"oldStatus", oldHealthStatus,
				"newStatus", healthStatus,
				"error", publishErr)
		} else {
			logger.InfoWithTrace(ctx, "发布健康状态变更事件成功",
				"instanceId", instanceId,
				"serviceName", instance.ServiceName,
				"oldStatus", oldHealthStatus,
				"newStatus", healthStatus,
				"eventId", eventId)
		}
	}

	// 记录健康状态更新日志
	if oldHealthStatus != healthStatus {
		logger.InfoWithTrace(ctx, "实例健康状态已变更",
			"instanceId", instanceId,
			"serviceName", instance.ServiceName,
			"oldStatus", oldHealthStatus,
			"newStatus", healthStatus,
			"checkTime", healthCheckTime.Format("2006-01-02 15:04:05"))
	} else {
		logger.DebugWithTrace(ctx, "实例健康状态已检查",
			"instanceId", instanceId,
			"serviceName", instance.ServiceName,
			"healthStatus", healthStatus,
			"checkTime", healthCheckTime.Format("2006-01-02 15:04:05"))
	}

	return nil
}

// SetServiceGroup 设置（创建或更新）服务组
// 创建新服务组或更新现有服务组，并发布相应事件
func (m *RegistryManager) SetServiceGroup(ctx context.Context, serviceGroup *core.ServiceGroup) (*core.ServiceGroup, error) {
	if serviceGroup == nil {
		return nil, fmt.Errorf("service group cannot be empty")
	}

	if serviceGroup.TenantId == "" || serviceGroup.ServiceGroupId == "" || serviceGroup.GroupName == "" {
		return nil, fmt.Errorf("tenantId, serviceGroupId and groupName cannot be empty")
	}

	// Check if service group already exists
	existingGroup, err := m.cache.GetServiceGroup(ctx, serviceGroup.TenantId, serviceGroup.ServiceGroupId)
	isNewGroup := err != nil || existingGroup == nil

	// Set default values for new groups
	now := time.Now()
	if isNewGroup {
		serviceGroup.AddTime = now
		serviceGroup.AddWho = "SYSTEM"
		serviceGroup.CurrentVersion = 1
		serviceGroup.ActiveFlag = "Y"
	} else {
		// Preserve original creation info for updates
		serviceGroup.AddTime = existingGroup.AddTime
		serviceGroup.AddWho = existingGroup.AddWho
		serviceGroup.CurrentVersion = existingGroup.CurrentVersion + 1
	}
	serviceGroup.EditTime = now
	serviceGroup.EditWho = "SYSTEM"

	// Set service group to cache
	err = m.cache.SetServiceGroup(ctx, serviceGroup.TenantId, serviceGroup)
	if err != nil {
		return nil, fmt.Errorf("failed to set service group: %w", err)
	}

	// Publish service group creation or update event
	if m.eventPublisher != nil {
		eventType := core.EventTypeServiceGroupUpdated
		if isNewGroup {
			eventType = core.EventTypeServiceGroupCreated
		}

		// Generate event ID
		eventId := random.Generate32BitRandomString()

		// Build event data JSON
		eventDataJson := fmt.Sprintf(`{
			"operation": "%s",
			"serviceGroupId": "%s",
			"groupName": "%s",
			"tenantId": "%s",
			"groupType": "%s",
			"groupDescription": "%s",
			"ownerUserId": "%s",
			"defaultProtocolType": "%s",
			"defaultLoadBalanceStrategy": "%s",
			"accessControlEnabled": "%s",
			"isNewGroup": %t
		}`, eventType, serviceGroup.ServiceGroupId, serviceGroup.GroupName, serviceGroup.TenantId,
			serviceGroup.GroupType, serviceGroup.GroupDescription, serviceGroup.OwnerUserId,
			serviceGroup.DefaultProtocolType, serviceGroup.DefaultLoadBalanceStrategy,
			serviceGroup.AccessControlEnabled, isNewGroup)

		// Build detailed event message
		var eventMessage string
		if isNewGroup {
			eventMessage = fmt.Sprintf("Service group created successfully: %s (tenant: %s, type: %s, owner: %s)",
				serviceGroup.GroupName, serviceGroup.TenantId, serviceGroup.GroupType, serviceGroup.OwnerUserId)
		} else {
			eventMessage = fmt.Sprintf("Service group updated successfully: %s (tenant: %s, type: %s, owner: %s)",
				serviceGroup.GroupName, serviceGroup.TenantId, serviceGroup.GroupType, serviceGroup.OwnerUserId)
		}

		event := &core.ServiceEvent{
			ServiceEventId:    eventId,
			EventType:         eventType,
			TenantId:          serviceGroup.TenantId,
			ServiceGroupId:    serviceGroup.ServiceGroupId,
			ServiceInstanceId: "", // Group-level event, instance ID is empty
			GroupName:         serviceGroup.GroupName,
			ServiceName:       "", // Group-level event, service name is empty
			HostAddress:       "", // Group-level event, host address is empty
			PortNumber:        0,  // Group-level event, port is 0
			NodeIpAddress:     random.GetNodeIP(),
			EventTime:         serviceGroup.EditTime,
			EventSource:       core.GetEventSourceFromContext(ctx, core.EventSourceRegistryManager),
			EventDataJson:     eventDataJson,
			EventMessage:      eventMessage,
			// Common fields
			AddTime:        now,
			AddWho:         "SYSTEM",
			EditTime:       now,
			EditWho:        "SYSTEM",
			OprSeqFlag:     eventId,
			CurrentVersion: 1,
			ActiveFlag:     "Y",
		}
		_ = m.eventPublisher.Publish(ctx, event)

		logger.InfoWithTrace(ctx, "Service group event published",
			"eventType", eventType,
			"groupName", serviceGroup.GroupName,
			"eventId", eventId)
	}

	logger.InfoWithTrace(ctx, "Service group operation completed successfully",
		"groupName", serviceGroup.GroupName,
		"tenantId", serviceGroup.TenantId,
		"serviceGroupId", serviceGroup.ServiceGroupId,
		"isNewGroup", isNewGroup)

	return serviceGroup, nil
}

// DeleteServiceGroup 删除服务组
// 删除服务组信息并发布相应事件
func (m *RegistryManager) DeleteServiceGroup(ctx context.Context, tenantId, serviceGroupId string) error {
	if tenantId == "" || serviceGroupId == "" {
		return fmt.Errorf("tenantId and serviceGroupId cannot be empty")
	}

	// Get service group information for event publishing
	serviceGroup, _ := m.cache.GetServiceGroup(ctx, tenantId, serviceGroupId)

	// Check if there are any services in this group before deletion
	if serviceGroup != nil {
		services, err := m.cache.ListServices(ctx, tenantId, serviceGroupId)
		if err == nil && len(services) > 0 {
			return fmt.Errorf("cannot delete service group '%s': it contains %d services. Please remove all services first",
				serviceGroup.GroupName, len(services))
		}
	}

	// Delete service group from cache
	err := m.cache.DeleteServiceGroup(ctx, tenantId, serviceGroupId)
	if err != nil {
		return fmt.Errorf("failed to delete service group: %w", err)
	}

	// Publish service group deletion event
	if m.eventPublisher != nil && serviceGroup != nil {
		// Generate event ID
		eventId := random.Generate32BitRandomString()

		// Build event data JSON
		eventDataJson := fmt.Sprintf(`{
			"operation": "SERVICE_GROUP_DELETED",
			"serviceGroupId": "%s",
			"groupName": "%s",
			"tenantId": "%s",
			"groupType": "%s",
			"groupDescription": "%s",
			"ownerUserId": "%s",
			"defaultProtocolType": "%s",
			"defaultLoadBalanceStrategy": "%s",
			"accessControlEnabled": "%s"
		}`, serviceGroupId, serviceGroup.GroupName, tenantId,
			serviceGroup.GroupType, serviceGroup.GroupDescription, serviceGroup.OwnerUserId,
			serviceGroup.DefaultProtocolType, serviceGroup.DefaultLoadBalanceStrategy,
			serviceGroup.AccessControlEnabled)

		// Build detailed event message
		eventMessage := fmt.Sprintf("Service group deleted successfully: %s (tenant: %s, type: %s, owner: %s)",
			serviceGroup.GroupName, tenantId, serviceGroup.GroupType, serviceGroup.OwnerUserId)

		now := time.Now()
		event := &core.ServiceEvent{
			ServiceEventId:    eventId,
			EventType:         core.EventTypeServiceGroupDeleted,
			TenantId:          tenantId,
			ServiceGroupId:    serviceGroupId,
			ServiceInstanceId: "", // Group-level event, instance ID is empty
			GroupName:         serviceGroup.GroupName,
			ServiceName:       "", // Group-level event, service name is empty
			HostAddress:       "", // Group-level event, host address is empty
			PortNumber:        0,  // Group-level event, port is 0
			NodeIpAddress:     random.GetNodeIP(),
			EventTime:         serviceGroup.EditTime,
			EventSource:       core.GetEventSourceFromContext(ctx, core.EventSourceRegistryManager),
			EventDataJson:     eventDataJson,
			EventMessage:      eventMessage,
			// Common fields
			AddTime:        now,
			AddWho:         "SYSTEM",
			EditTime:       now,
			EditWho:        "SYSTEM",
			OprSeqFlag:     eventId,
			CurrentVersion: 1,
			ActiveFlag:     "Y",
		}
		_ = m.eventPublisher.Publish(ctx, event)

		logger.InfoWithTrace(ctx, "Service group deletion event published",
			"groupName", serviceGroup.GroupName,
			"eventId", eventId)
	}

	logger.InfoWithTrace(ctx, "Service group deleted successfully",
		"serviceGroupId", serviceGroupId,
		"tenantId", tenantId)

	return nil
}
