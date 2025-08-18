package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gateway/internal/registry/core"
	"gateway/pkg/utils/random"
)

// RegistryService 注册中心服务实现
type RegistryService struct {
	storage        core.Storage
	eventPublisher core.EventPublisher
	subscribers    map[string][]chan *core.ServiceEvent
	subMutex       sync.RWMutex
	running        bool
	mutex          sync.RWMutex
}

// NewRegistryService 创建注册中心服务实例
func NewRegistryService(storage core.Storage, eventPublisher core.EventPublisher) *RegistryService {
	return &RegistryService{
		storage:        storage,
		eventPublisher: eventPublisher,
		subscribers:    make(map[string][]chan *core.ServiceEvent),
	}
}

// ================== 服务实例管理 ==================

// Register 注册服务实例
func (r *RegistryService) Register(ctx context.Context, instance *core.ServiceInstance) error {
	r.mutex.RLock()
	if !r.running {
		r.mutex.RUnlock()
		return core.ErrRegistryNotRunning
	}
	r.mutex.RUnlock()

	// 验证参数
	if err := r.validateInstance(instance); err != nil {
		return fmt.Errorf("validate instance failed: %w", err)
	}

	// 设置默认值
	r.setInstanceDefaults(instance)

	// 检查服务是否存在，如果不存在则创建
	service, err := r.storage.GetService(ctx, instance.TenantId, instance.ServiceName)
	if err == core.ErrServiceNotFound {
		// 创建默认服务
		service = &core.Service{
			TenantId:                   instance.TenantId,
			ServiceName:                instance.ServiceName,
			GroupName:                  instance.GroupName,
			ServiceDescription:         fmt.Sprintf("Auto-created service for %s", instance.ServiceName),
			ProtocolType:               "HTTP",
			LoadBalanceStrategy:        "ROUND_ROBIN",
			HealthCheckUrl:             "/health",
			HealthCheckIntervalSeconds: 30,
			HealthCheckTimeoutSeconds:  5,
			AddTime:                    time.Now(),
			AddWho:                     instance.AddWho,
			EditTime:                   time.Now(),
			EditWho:                    instance.EditWho,
			OprSeqFlag:                 random.Generate32BitRandomString(),
			CurrentVersion:             1,
			ActiveFlag:                 core.FlagYes,
		}

		if err := r.storage.SaveService(ctx, service); err != nil {
			return fmt.Errorf("create service failed: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("get service failed: %w", err)
	}

	// 检查分组是否存在，如果不存在则创建
	group, err := r.storage.GetServiceGroup(ctx, instance.TenantId, instance.GroupName)
	if err == core.ErrGroupNotFound {
		// 创建默认分组
		group = &core.ServiceGroup{
			ServiceGroupId:                    random.Generate32BitRandomString(),
			TenantId:                          instance.TenantId,
			GroupName:                         instance.GroupName,
			GroupDescription:                  fmt.Sprintf("Auto-created group for %s", instance.GroupName),
			GroupType:                         core.GroupTypeBusiness,
			OwnerUserId:                       instance.AddWho,
			AccessControlEnabled:              core.FlagNo,
			DefaultProtocolType:               "HTTP",
			DefaultLoadBalanceStrategy:        "ROUND_ROBIN",
			DefaultHealthCheckUrl:             "/health",
			DefaultHealthCheckIntervalSeconds: 30,
			AddTime:                           time.Now(),
			AddWho:                            instance.AddWho,
			EditTime:                          time.Now(),
			EditWho:                           instance.EditWho,
			OprSeqFlag:                        random.Generate32BitRandomString(),
			CurrentVersion:                    1,
			ActiveFlag:                        core.FlagYes,
		}

		if err := r.storage.SaveServiceGroup(ctx, group); err != nil {
			return fmt.Errorf("create service group failed: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("get service group failed: %w", err)
	}

	// 保存实例
	if err := r.storage.SaveInstance(ctx, instance); err != nil {
		return fmt.Errorf("save instance failed: %w", err)
	}

	// 发布注册事件
	event := core.NewServiceEvent(
		instance.TenantId,
		core.EventTypeInstanceRegister,
		instance.ServiceName,
		instance.GroupName,
		"registry-service",
		fmt.Sprintf("Instance %s registered", instance.ServiceInstanceId),
	)
	event.HostAddress = instance.HostAddress
	event.PortNumber = &instance.PortNumber

	if err := r.eventPublisher.Publish(ctx, event); err != nil {
		// 事件发布失败不影响注册
		fmt.Printf("publish register event failed: %v\n", err)
	}

	return nil
}

// Deregister 注销服务实例
func (r *RegistryService) Deregister(ctx context.Context, tenantId, instanceId string) error {
	r.mutex.RLock()
	if !r.running {
		r.mutex.RUnlock()
		return core.ErrRegistryNotRunning
	}
	r.mutex.RUnlock()

	// 获取实例信息
	instance, err := r.storage.GetInstance(ctx, tenantId, instanceId)
	if err != nil {
		return fmt.Errorf("get instance failed: %w", err)
	}

	// 删除实例
	if err := r.storage.DeleteInstance(ctx, tenantId, instanceId); err != nil {
		return fmt.Errorf("delete instance failed: %w", err)
	}

	// 发布注销事件
	event := core.NewServiceEvent(
		tenantId,
		core.EventTypeInstanceDeregister,
		instance.ServiceName,
		instance.GroupName,
		"registry-service",
		fmt.Sprintf("Instance %s deregistered", instanceId),
	)
	event.HostAddress = instance.HostAddress
	event.PortNumber = &instance.PortNumber

	if err := r.eventPublisher.Publish(ctx, event); err != nil {
		// 事件发布失败不影响注销
		fmt.Printf("publish deregister event failed: %v\n", err)
	}

	return nil
}

// Heartbeat 心跳
func (r *RegistryService) Heartbeat(ctx context.Context, tenantId, instanceId string) error {
	r.mutex.RLock()
	if !r.running {
		r.mutex.RUnlock()
		return core.ErrRegistryNotRunning
	}
	r.mutex.RUnlock()

	// 获取实例信息
	instance, err := r.storage.GetInstance(ctx, tenantId, instanceId)
	if err != nil {
		return fmt.Errorf("get instance failed: %w", err)
	}

	// 更新心跳时间
	if err := r.storage.UpdateHeartbeat(ctx, tenantId, instanceId); err != nil {
		return fmt.Errorf("update heartbeat failed: %w", err)
	}

	// 发布心跳事件
	event := core.NewServiceEvent(
		tenantId,
		core.EventTypeInstanceHeartbeat,
		instance.ServiceName,
		instance.GroupName,
		"registry-service",
		fmt.Sprintf("Instance %s heartbeat", instanceId),
	)
	event.HostAddress = instance.HostAddress
	event.PortNumber = &instance.PortNumber

	if err := r.eventPublisher.Publish(ctx, event); err != nil {
		// 事件发布失败不影响心跳
		fmt.Printf("publish heartbeat event failed: %v\n", err)
	}

	return nil
}

// ================== 服务发现 ==================

// Discover 发现服务实例
func (r *RegistryService) Discover(ctx context.Context, tenantId, serviceName, groupName string, filters ...core.InstanceFilter) ([]*core.ServiceInstance, error) {
	r.mutex.RLock()
	if !r.running {
		r.mutex.RUnlock()
		return nil, core.ErrRegistryNotRunning
	}
	r.mutex.RUnlock()

	// 获取实例列表
	instances, err := r.storage.GetInstances(ctx, tenantId, serviceName, groupName, filters...)
	if err != nil {
		return nil, fmt.Errorf("get instances failed: %w", err)
	}

	return instances, nil
}

// GetInstance 获取服务实例
func (r *RegistryService) GetInstance(ctx context.Context, tenantId, instanceId string) (*core.ServiceInstance, error) {
	r.mutex.RLock()
	if !r.running {
		r.mutex.RUnlock()
		return nil, core.ErrRegistryNotRunning
	}
	r.mutex.RUnlock()

	instance, err := r.storage.GetInstance(ctx, tenantId, instanceId)
	if err != nil {
		return nil, fmt.Errorf("get instance failed: %w", err)
	}

	return instance, nil
}

// ListServices 列出服务
func (r *RegistryService) ListServices(ctx context.Context, tenantId, groupName string) ([]string, error) {
	r.mutex.RLock()
	if !r.running {
		r.mutex.RUnlock()
		return nil, core.ErrRegistryNotRunning
	}
	r.mutex.RUnlock()

	serviceNames, err := r.storage.GetServiceNames(ctx, tenantId, groupName)
	if err != nil {
		return nil, fmt.Errorf("get service names failed: %w", err)
	}

	return serviceNames, nil
}

// ================== 事件订阅 ==================

// Subscribe 订阅事件
func (r *RegistryService) Subscribe(ctx context.Context, tenantId, serviceName, groupName string) (<-chan *core.ServiceEvent, error) {
	r.mutex.RLock()
	if !r.running {
		r.mutex.RUnlock()
		return nil, core.ErrRegistryNotRunning
	}
	r.mutex.RUnlock()

	// 创建订阅通道
	eventChan := make(chan *core.ServiceEvent, 100)

	// 构建订阅键
	key := r.buildSubscriptionKey(tenantId, serviceName, groupName)

	r.subMutex.Lock()
	r.subscribers[key] = append(r.subscribers[key], eventChan)
	r.subMutex.Unlock()

	// 订阅事件发布器
	pubChan, err := r.eventPublisher.Subscribe(ctx, tenantId, serviceName, groupName)
	if err != nil {
		return nil, fmt.Errorf("subscribe to event publisher failed: %w", err)
	}

	// 启动事件转发协程
	go r.forwardEvents(pubChan, eventChan)

	return eventChan, nil
}

// Unsubscribe 取消订阅
func (r *RegistryService) Unsubscribe(ctx context.Context, tenantId, serviceName, groupName string) error {
	key := r.buildSubscriptionKey(tenantId, serviceName, groupName)

	r.subMutex.Lock()
	defer r.subMutex.Unlock()

	// 关闭所有订阅通道
	if channels, exists := r.subscribers[key]; exists {
		for _, ch := range channels {
			close(ch)
		}
		delete(r.subscribers, key)
	}

	// 取消事件发布器订阅
	return r.eventPublisher.Unsubscribe(ctx, tenantId, serviceName, groupName)
}

// ================== 健康状态管理 ==================

// UpdateHealth 更新健康状态
func (r *RegistryService) UpdateHealth(ctx context.Context, tenantId, instanceId string, healthStatus string) error {
	r.mutex.RLock()
	if !r.running {
		r.mutex.RUnlock()
		return core.ErrRegistryNotRunning
	}
	r.mutex.RUnlock()

	// 获取实例信息
	instance, err := r.storage.GetInstance(ctx, tenantId, instanceId)
	if err != nil {
		return fmt.Errorf("get instance failed: %w", err)
	}

	oldStatus := instance.HealthStatus

	// 更新健康状态
	if err := r.storage.UpdateInstanceHealth(ctx, tenantId, instanceId, healthStatus); err != nil {
		return fmt.Errorf("update instance health failed: %w", err)
	}

	// 如果状态发生变化，发布事件
	if oldStatus != healthStatus {
		event := core.NewServiceEvent(
			tenantId,
			core.EventTypeInstanceHealthChange,
			instance.ServiceName,
			instance.GroupName,
			"registry-service",
			fmt.Sprintf("Instance %s health changed from %s to %s", instanceId, oldStatus, healthStatus),
		)
		event.HostAddress = instance.HostAddress
		event.PortNumber = &instance.PortNumber

		if err := r.eventPublisher.Publish(ctx, event); err != nil {
			// 事件发布失败不影响状态更新
			fmt.Printf("publish health change event failed: %v\n", err)
		}
	}

	return nil
}

// ================== 生命周期管理 ==================

// Start 启动服务
func (r *RegistryService) Start() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.running {
		return nil
	}

	r.running = true
	return nil
}

// Close 关闭服务
func (r *RegistryService) Close() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.running {
		return nil
	}

	r.running = false

	// 关闭所有订阅通道
	r.subMutex.Lock()
	for key, channels := range r.subscribers {
		for _, ch := range channels {
			close(ch)
		}
		delete(r.subscribers, key)
	}
	r.subMutex.Unlock()

	// 关闭事件发布器
	if r.eventPublisher != nil {
		return r.eventPublisher.Close()
	}

	return nil
}

// ================== 辅助方法 ==================

// validateInstance 验证实例
func (r *RegistryService) validateInstance(instance *core.ServiceInstance) error {
	if instance == nil {
		return core.ErrInvalidParameter
	}

	if instance.TenantId == "" {
		return fmt.Errorf("tenantId is required")
	}

	if instance.ServiceName == "" {
		return fmt.Errorf("serviceName is required")
	}

	if instance.GroupName == "" {
		return fmt.Errorf("groupName is required")
	}

	if instance.HostAddress == "" {
		return fmt.Errorf("hostAddress is required")
	}

	if instance.PortNumber <= 0 || instance.PortNumber > 65535 {
		return fmt.Errorf("portNumber must be between 1 and 65535")
	}

	return nil
}

// setInstanceDefaults 设置实例默认值
func (r *RegistryService) setInstanceDefaults(instance *core.ServiceInstance) {
	now := time.Now()

	if instance.ServiceInstanceId == "" {
		instance.ServiceInstanceId = random.Generate32BitRandomString()
	}

	if instance.InstanceStatus == "" {
		instance.InstanceStatus = core.InstanceStatusUp
	}

	if instance.HealthStatus == "" {
		instance.HealthStatus = core.HealthStatusHealthy
	}

	if instance.WeightValue == 0 {
		instance.WeightValue = 100
	}

	if instance.ClientType == "" {
		instance.ClientType = core.ClientTypeService
	}

	if instance.RegisterTime.IsZero() {
		instance.RegisterTime = now
	}

	if instance.AddTime.IsZero() {
		instance.AddTime = now
	}

	if instance.EditTime.IsZero() {
		instance.EditTime = now
	}

	if instance.OprSeqFlag == "" {
		instance.OprSeqFlag = random.Generate32BitRandomString()
	}

	if instance.CurrentVersion == 0 {
		instance.CurrentVersion = 1
	}

	if instance.ActiveFlag == "" {
		instance.ActiveFlag = core.FlagYes
	}
}

// buildSubscriptionKey 构建订阅键
func (r *RegistryService) buildSubscriptionKey(tenantId, serviceName, groupName string) string {
	return fmt.Sprintf("%s:%s:%s", tenantId, groupName, serviceName)
}

// forwardEvents 转发事件
func (r *RegistryService) forwardEvents(source <-chan *core.ServiceEvent, target chan<- *core.ServiceEvent) {
	defer close(target)

	for event := range source {
		select {
		case target <- event:
		default:
			// 如果目标通道满了，跳过这个事件
			fmt.Printf("event channel full, dropping event: %s\n", event.EventType)
		}
	}
}

// IsRunning 检查是否运行中
func (r *RegistryService) IsRunning() bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.running
}
