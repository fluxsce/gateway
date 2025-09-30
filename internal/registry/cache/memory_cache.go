package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gateway/internal/registry/core"
	"gateway/pkg/logger"
)

// MemoryCache 基于 sync.Map 的高并发内存缓存实现
// 采用 TenantId -> ServiceGroupId -> ServiceName -> ServiceInstance[] 的层次结构
// 集成外部注册中心管理，根据服务的注册类型自动同步到外部注册中心
type MemoryCache struct {
	// 租户映射: tenantId -> map[serviceGroupId]*core.ServiceGroup
	tenants sync.Map

	// 实例ID索引: instanceId -> {tenantId, serviceGroupId, serviceName}
	// 用于快速定位实例所属的服务和服务组
	instanceIndex sync.Map

	// 负载均衡器工厂，全局唯一实例
	lbFactory *LoadBalancerFactory

	// 外部注册中心管理器
	externalManager *ExternalRegistryCacheManager

	// 统计信息
	stats struct {
		sync.RWMutex
		hitCount  int64
		missCount int64
	}
}

// instanceLocation 实例位置信息
type instanceLocation struct {
	tenantId       string
	serviceGroupId string
	serviceName    string
}

// NewMemoryCache 创建基于 sync.Map 的高并发缓存实例
func NewMemoryCache() core.CacheStorage {
	return &MemoryCache{
		lbFactory:       NewLoadBalancerFactory(),
		externalManager: NewExternalRegistryCacheManager(),
	}
}

// ========== 内部辅助方法 ==========

// getTenantGroups 获取租户下的所有服务组映射
func (c *MemoryCache) getTenantGroups(tenantId string) (map[string]*core.ServiceGroup, bool) {
	value, exists := c.tenants.Load(tenantId)
	if !exists {
		return nil, false
	}
	return value.(map[string]*core.ServiceGroup), true
}

// getOrCreateTenantGroups 获取或创建租户下的服务组映射
func (c *MemoryCache) getOrCreateTenantGroups(tenantId string) map[string]*core.ServiceGroup {
	value, exists := c.tenants.Load(tenantId)
	if exists {
		return value.(map[string]*core.ServiceGroup)
	}

	// 创建新的映射
	groups := make(map[string]*core.ServiceGroup)

	// 使用 LoadOrStore 确保并发安全
	actual, loaded := c.tenants.LoadOrStore(tenantId, groups)
	if loaded {
		return actual.(map[string]*core.ServiceGroup)
	}
	return groups
}

// recordHit 记录缓存命中
func (c *MemoryCache) recordHit() {
	c.stats.Lock()
	defer c.stats.Unlock()
	c.stats.hitCount++
}

// recordMiss 记录缓存未命中
func (c *MemoryCache) recordMiss() {
	c.stats.Lock()
	defer c.stats.Unlock()
	c.stats.missCount++
}

// ========== CacheStorage接口实现 ==========

// GetServiceGroup 获取服务组
func (c *MemoryCache) GetServiceGroup(ctx context.Context, tenantId, serviceGroupId string) (*core.ServiceGroup, error) {
	groups, exists := c.getTenantGroups(tenantId)
	if !exists {
		c.recordMiss()
		return nil, fmt.Errorf("租户缓存未命中: %s", tenantId)
	}

	group, exists := groups[serviceGroupId]
	if !exists {
		c.recordMiss()
		return nil, fmt.Errorf("服务组缓存未命中: %s", serviceGroupId)
	}

	c.recordHit()
	logger.DebugWithTrace(ctx, "获取服务组缓存",
		"tenantId", tenantId,
		"serviceGroupId", serviceGroupId)

	return group, nil
}

// SetServiceGroup 设置服务组
func (c *MemoryCache) SetServiceGroup(ctx context.Context, tenantId string, serviceGroup *core.ServiceGroup) error {
	if serviceGroup == nil {
		return fmt.Errorf("服务组不能为空")
	}

	if serviceGroup.TenantId != tenantId {
		return fmt.Errorf("租户ID不匹配")
	}

	// 获取现有的租户服务组映射
	value, exists := c.tenants.Load(tenantId)

	// 创建新的服务组映射（深拷贝）
	var newGroups map[string]*core.ServiceGroup
	if exists {
		// 复制现有的服务组映射
		oldGroups := value.(map[string]*core.ServiceGroup)
		newGroups = make(map[string]*core.ServiceGroup, len(oldGroups)+1)

		// 复制所有服务组，除了要更新的服务组
		for groupId, oldGroup := range oldGroups {
			if groupId == serviceGroup.ServiceGroupId {
				// 稍后会单独处理目标服务组
				continue
			}
			// 直接复制非目标服务组的指针
			newGroups[groupId] = oldGroup
		}
	} else {
		// 创建新的空映射
		newGroups = make(map[string]*core.ServiceGroup)
	}

	// 处理目标服务组
	newServiceGroup := serviceGroup.ShallowCopy()

	// 处理服务列表 - 修复逻辑
	// SetServiceGroup 方法应该完全按照传入的服务组对象来设置，不应该自动保留旧的服务列表
	// 如果调用者希望保留服务，应该在调用前自己获取并设置到 serviceGroup.Services 中
	if serviceGroup.Services != nil {
		// 使用提供的服务列表（可能是空map，这是合法的）
		newServiceGroup.Services = make(map[string]*core.Service, len(serviceGroup.Services))
		for serviceName, service := range serviceGroup.Services {
			newServiceGroup.Services[serviceName] = service
		}
	} else {
		// 如果没有提供服务列表，创建空的服务map
		// 这样可以明确表示该服务组当前没有服务
		newServiceGroup.Services = make(map[string]*core.Service)
	}

	// 更新服务组到租户映射
	newGroups[serviceGroup.ServiceGroupId] = newServiceGroup

	// 原子性更新租户映射
	c.tenants.Store(tenantId, newGroups)

	logger.DebugWithTrace(ctx, "设置服务组缓存",
		"tenantId", tenantId,
		"serviceGroupId", serviceGroup.ServiceGroupId,
		"serviceCount", len(newServiceGroup.Services))

	return nil
}

// DeleteServiceGroup 删除服务组
func (c *MemoryCache) DeleteServiceGroup(ctx context.Context, tenantId, serviceGroupId string) error {
	groups, exists := c.getTenantGroups(tenantId)
	if !exists {
		return nil // 租户不存在，视为删除成功
	}

	// 获取服务组
	group, exists := groups[serviceGroupId]
	if !exists {
		return nil // 服务组不存在，视为删除成功
	}

	// 删除该服务组下所有实例的索引
	if group.Services != nil {
		for _, service := range group.Services {
			if service.Instances != nil {
				for _, instance := range service.Instances {
					c.instanceIndex.Delete(instance.ServiceInstanceId)
				}
			}
		}
	}

	// 删除服务组
	delete(groups, serviceGroupId)

	// 如果租户没有任何服务组，删除租户
	if len(groups) == 0 {
		c.tenants.Delete(tenantId)
	}

	logger.DebugWithTrace(ctx, "删除服务组缓存",
		"tenantId", tenantId,
		"serviceGroupId", serviceGroupId)

	return nil
}

// ListServiceGroups 列出租户下的所有服务组
func (c *MemoryCache) ListServiceGroups(ctx context.Context, tenantId string) ([]*core.ServiceGroup, error) {
	groups, exists := c.getTenantGroups(tenantId)
	if !exists {
		// 租户不存在，返回空列表而不是错误
		logger.DebugWithTrace(ctx, "租户不存在，返回空服务组列表",
			"tenantId", tenantId)
		return make([]*core.ServiceGroup, 0), nil
	}

	result := make([]*core.ServiceGroup, 0, len(groups))
	for _, group := range groups {
		result = append(result, group)
	}

	c.recordHit()
	logger.DebugWithTrace(ctx, "列出服务组缓存",
		"tenantId", tenantId,
		"count", len(result))

	return result, nil
}

// GetService 获取服务
func (c *MemoryCache) GetService(ctx context.Context, tenantId, serviceGroupId, serviceName string) (*core.Service, error) {
	group, err := c.GetServiceGroup(ctx, tenantId, serviceGroupId)
	if err != nil {
		return nil, err
	}

	if group.Services == nil {
		c.recordMiss()
		return nil, fmt.Errorf("服务缓存未命中: %s", serviceName)
	}

	service, exists := group.Services[serviceName]
	if !exists {
		c.recordMiss()
		return nil, fmt.Errorf("服务缓存未命中: %s", serviceName)
	}

	c.recordHit()
	logger.DebugWithTrace(ctx, "获取服务缓存",
		"tenantId", tenantId,
		"serviceGroupId", serviceGroupId,
		"serviceName", serviceName)

	return service, nil
}

// SetService 设置服务
func (c *MemoryCache) SetService(ctx context.Context, tenantId string, service *core.Service) error {
	if service == nil {
		return fmt.Errorf("服务不能为空")
	}

	if service.TenantId != tenantId {
		return fmt.Errorf("租户ID不匹配")
	}

	// 获取现有的租户服务组映射
	value, exists := c.tenants.Load(tenantId)

	// 创建新的服务组映射（深拷贝）
	var newGroups map[string]*core.ServiceGroup
	if exists {
		// 复制现有的服务组映射
		oldGroups := value.(map[string]*core.ServiceGroup)
		newGroups = make(map[string]*core.ServiceGroup, len(oldGroups))

		// 复制所有服务组
		for groupId, oldGroup := range oldGroups {
			if groupId == service.ServiceGroupId {
				// 稍后会单独处理目标服务组
				continue
			}
			// 直接复制非目标服务组的指针
			newGroups[groupId] = oldGroup
		}
	} else {
		// 创建新的空映射
		newGroups = make(map[string]*core.ServiceGroup)
	}

	// 处理目标服务组
	var newGroup *core.ServiceGroup
	var oldGroup *core.ServiceGroup
	if exists {
		oldGroups := value.(map[string]*core.ServiceGroup)
		oldGroup = oldGroups[service.ServiceGroupId]
	}

	if oldGroup != nil {
		// 使用浅拷贝复制现有服务组
		newGroup = oldGroup.ShallowCopy()
		// 创建新的服务映射
		newGroup.Services = make(map[string]*core.Service)

		// 复制所有服务
		for svcName, oldService := range oldGroup.Services {
			if svcName == service.ServiceName {
				// 稍后会单独处理目标服务
				continue
			}
			// 直接复制非目标服务的指针
			newGroup.Services[svcName] = oldService
		}
	} else {
		// 创建新的服务组
		newGroup = &core.ServiceGroup{
			TenantId:       tenantId,
			ServiceGroupId: service.ServiceGroupId,
			GroupName:      service.ServiceGroupId, // 默认使用ID作为名称
			Services:       make(map[string]*core.Service),
		}
	}

	// 处理目标服务
	var newService *core.Service

	// 使用浅拷贝创建新的服务对象
	newService = service.ShallowCopy()

	// 处理实例列表 - 修复逻辑
	// SetService 方法应该完全按照传入的服务对象来设置，不应该自动保留旧的实例列表
	// 如果调用者希望保留实例，应该在调用前自己获取并设置到 service.Instances 中
	if service.Instances != nil {
		// 使用提供的实例列表（可能是空列表，这是合法的）
		newService.Instances = make([]*core.ServiceInstance, len(service.Instances))
		copy(newService.Instances, service.Instances)
	} else {
		// 如果没有提供实例列表，创建空的实例列表
		// 这样可以明确表示该服务当前没有实例
		newService.Instances = make([]*core.ServiceInstance, 0)
	}

	// 更新服务到服务组
	newGroup.Services[service.ServiceName] = newService

	// 更新服务组到租户映射
	newGroups[service.ServiceGroupId] = newGroup

	// 原子性更新租户映射
	c.tenants.Store(tenantId, newGroups)

	logger.DebugWithTrace(ctx, "设置服务缓存",
		"tenantId", tenantId,
		"serviceGroupId", service.ServiceGroupId,
		"serviceName", service.ServiceName,
		"instanceCount", len(newService.Instances))

	// 如果是外部注册中心的服务，注册到外部注册中心
	if service.IsExternalRegistry() {
		if err := c.externalManager.RegisterService(ctx, service); err != nil {
			logger.WarnWithTrace(ctx, "注册服务到外部注册中心失败",
				"serviceName", service.ServiceName,
				"registryType", service.RegistryType,
				"error", err)
			// 外部注册失败不影响内存缓存操作的成功
		}
	}

	return nil
}

// DeleteService 删除服务
func (c *MemoryCache) DeleteService(ctx context.Context, tenantId, serviceGroupId, serviceName string) error {
	group, err := c.GetServiceGroup(ctx, tenantId, serviceGroupId)
	if err != nil {
		return nil // 服务组不存在，视为删除成功
	}

	if group.Services == nil {
		return nil // 服务不存在，视为删除成功
	}

	// 获取服务
	service, exists := group.Services[serviceName]
	if !exists {
		return nil // 服务不存在，视为删除成功
	}

	// 如果是外部注册中心的服务，从外部注册中心注销服务并清理资源
	if service.IsExternalRegistry() {
		if err := c.externalManager.DeregisterService(ctx, service); err != nil {
			logger.WarnWithTrace(ctx, "从外部注册中心注销服务失败",
				"serviceName", service.ServiceName,
				"registryType", service.RegistryType,
				"error", err)
		}
	}

	// 删除该服务下所有实例的索引
	if service.Instances != nil {
		for _, instance := range service.Instances {
			c.instanceIndex.Delete(instance.ServiceInstanceId)
		}
	}

	// 删除服务
	delete(group.Services, serviceName)

	// 清理该服务相关的负载均衡器状态
	serviceKey := ServiceKey{
		TenantId:       tenantId,
		ServiceGroupId: serviceGroupId,
		ServiceName:    serviceName,
	}

	// 删除服务的负载均衡器
	c.lbFactory.RemoveLoadBalancer(service.LoadBalanceStrategy, serviceKey)

	// 如果服务组没有任何服务，删除服务组
	if len(group.Services) == 0 {
		groups, _ := c.getTenantGroups(tenantId)
		delete(groups, serviceGroupId)

		// 如果租户没有任何服务组，删除租户
		if len(groups) == 0 {
			c.tenants.Delete(tenantId)
		}
	}

	logger.DebugWithTrace(ctx, "删除服务缓存",
		"tenantId", tenantId,
		"serviceGroupId", serviceGroupId,
		"serviceName", serviceName)

	return nil
}

// ListServices 列出服务组下的所有服务
func (c *MemoryCache) ListServices(ctx context.Context, tenantId, serviceGroupId string) ([]*core.Service, error) {
	// 获取租户的服务组映射
	groups, exists := c.getTenantGroups(tenantId)
	if !exists {
		// 租户不存在，返回空列表
		logger.DebugWithTrace(ctx, "租户不存在，返回空服务列表",
			"tenantId", tenantId,
			"serviceGroupId", serviceGroupId)
		return make([]*core.Service, 0), nil
	}

	// 获取指定的服务组
	group, exists := groups[serviceGroupId]
	if !exists {
		// 服务组不存在，返回空列表
		logger.DebugWithTrace(ctx, "服务组不存在，返回空服务列表",
			"tenantId", tenantId,
			"serviceGroupId", serviceGroupId)
		return make([]*core.Service, 0), nil
	}

	if group.Services == nil {
		return make([]*core.Service, 0), nil
	}

	result := make([]*core.Service, 0, len(group.Services))
	for _, service := range group.Services {
		result = append(result, service)
	}

	logger.DebugWithTrace(ctx, "列出服务缓存",
		"tenantId", tenantId,
		"serviceGroupId", serviceGroupId,
		"count", len(result))

	return result, nil
}

// GetInstance 获取服务实例
func (c *MemoryCache) GetInstance(ctx context.Context, tenantId, instanceId string) (*core.ServiceInstance, error) {
	// 通过索引快速定位实例
	value, exists := c.instanceIndex.Load(instanceId)
	if !exists {
		c.recordMiss()
		// 对于外部注册中心的实例，实例ID不在内存索引中是正常的
		logger.DebugWithTrace(ctx, "实例不在内存索引中，可能是外部注册中心实例",
			"tenantId", tenantId,
			"instanceId", instanceId)
		return nil, fmt.Errorf("实例缓存未命中: %s", instanceId)
	}

	loc := value.(instanceLocation)

	// 验证租户ID
	if loc.tenantId != tenantId {
		c.recordMiss()
		return nil, fmt.Errorf("实例租户不匹配: %s", instanceId)
	}

	// 获取服务信息以确定注册类型
	service, err := c.GetService(ctx, tenantId, loc.serviceGroupId, loc.serviceName)
	if err != nil {
		c.recordMiss()
		return nil, fmt.Errorf("获取服务信息失败: %w", err)
	}

	// 如果是外部注册中心的服务，实例不应该在内存中维护
	if service.IsExternalRegistry() {
		logger.WarnWithTrace(ctx, "外部注册中心的实例不应该在内存索引中",
			"tenantId", tenantId,
			"instanceId", instanceId,
			"serviceName", service.ServiceName,
			"registryType", service.RegistryType)
		// 删除错误的索引
		c.instanceIndex.Delete(instanceId)
		c.recordMiss()
		return nil, fmt.Errorf("外部注册中心实例不在内存缓存中: %s", instanceId)
	}

	// 内部注册中心的服务，从内存中查找实例
	for _, instance := range service.Instances {
		if instance.ServiceInstanceId == instanceId {
			c.recordHit()
			return instance, nil
		}
	}

	// 实例不存在，删除索引
	c.instanceIndex.Delete(instanceId)
	c.recordMiss()
	return nil, fmt.Errorf("实例缓存未命中: %s", instanceId)
}

// SetInstance 设置服务实例
func (c *MemoryCache) SetInstance(ctx context.Context, tenantId string, instance *core.ServiceInstance) error {
	if instance == nil {
		return fmt.Errorf("实例不能为空")
	}

	if instance.TenantId != tenantId {
		return fmt.Errorf("租户ID不匹配")
	}

	// 先获取服务信息以确定注册类型
	service, err := c.GetService(ctx, tenantId, instance.ServiceGroupId, instance.ServiceName)
	if err != nil {
		logger.WarnWithTrace(ctx, "获取服务信息失败，无法确定注册类型",
			"instanceId", instance.ServiceInstanceId,
			"serviceName", instance.ServiceName,
			"error", err)
		return fmt.Errorf("获取服务信息失败: %w", err)
	}

	// 如果是外部注册中心的服务，只注册到外部，不在内存中维护实例状态
	if service.IsExternalRegistry() {
		return c.externalManager.RegisterInstance(ctx, instance, service)
	}

	// 内部注册中心的服务，正常处理内存缓存
	// 获取现有的租户服务组映射
	value, exists := c.tenants.Load(tenantId)

	// 创建新的服务组映射（深拷贝）
	var newGroups map[string]*core.ServiceGroup
	if exists {
		// 复制现有的服务组映射
		oldGroups := value.(map[string]*core.ServiceGroup)
		newGroups = make(map[string]*core.ServiceGroup, len(oldGroups))

		// 复制所有服务组
		for groupId, oldGroup := range oldGroups {
			if groupId == instance.ServiceGroupId {
				// 稍后会单独处理目标服务组
				continue
			}
			// 直接复制非目标服务组的指针
			newGroups[groupId] = oldGroup
		}
	} else {
		// 创建新的空映射
		newGroups = make(map[string]*core.ServiceGroup)
	}

	// 处理目标服务组
	var newGroup *core.ServiceGroup
	var oldGroup *core.ServiceGroup
	if exists {
		oldGroups := value.(map[string]*core.ServiceGroup)
		oldGroup, _ = oldGroups[instance.ServiceGroupId]
	}

	if oldGroup != nil {
		// 使用浅拷贝复制现有服务组
		newGroup = oldGroup.ShallowCopy()
		// 创建新的服务映射
		newGroup.Services = make(map[string]*core.Service)

		// 复制所有服务
		for svcName, oldService := range oldGroup.Services {
			if svcName == instance.ServiceName {
				// 稍后会单独处理目标服务
				continue
			}
			// 直接复制非目标服务的指针
			newGroup.Services[svcName] = oldService
		}
	} else {
		// 创建新的服务组
		newGroup = &core.ServiceGroup{
			TenantId:       tenantId,
			ServiceGroupId: instance.ServiceGroupId,
			GroupName:      instance.ServiceGroupId, // 默认使用ID作为名称
			Services:       make(map[string]*core.Service),
		}
	}

	// 处理目标服务
	var newService *core.Service
	var oldService *core.Service
	if oldGroup != nil {
		oldService = oldGroup.Services[instance.ServiceName]
	}

	if oldService != nil {
		// 使用浅拷贝复制现有服务
		newService = oldService.ShallowCopy()
		// 创建新的实例列表
		newService.Instances = make([]*core.ServiceInstance, 0, len(oldService.Instances))

		// 复制所有实例，替换或添加新实例
		instanceFound := false
		for _, oldInstance := range oldService.Instances {
			if oldInstance.ServiceInstanceId == instance.ServiceInstanceId {
				// 找到要更新的实例，添加新实例替换旧实例
				newService.Instances = append(newService.Instances, instance)
				instanceFound = true
			} else {
				// 保留其他实例
				newService.Instances = append(newService.Instances, oldInstance)
			}
		}

		// 如果没有找到要更新的实例，添加为新实例
		if !instanceFound {
			newService.Instances = append(newService.Instances, instance)
		}
	} else {
		// 创建新的服务
		newService = &core.Service{
			TenantId:       tenantId,
			ServiceGroupId: instance.ServiceGroupId,
			ServiceName:    instance.ServiceName,
			// 创建新的实例列表，只包含当前实例
			Instances: []*core.ServiceInstance{instance},
		}
	}

	// 更新服务到服务组
	newGroup.Services[instance.ServiceName] = newService

	// 更新服务组到租户映射
	newGroups[instance.ServiceGroupId] = newGroup

	// 原子性更新租户映射
	c.tenants.Store(tenantId, newGroups)

	// 更新实例索引
	c.instanceIndex.Store(instance.ServiceInstanceId, instanceLocation{
		tenantId:       tenantId,
		serviceGroupId: instance.ServiceGroupId,
		serviceName:    instance.ServiceName,
	})

	logger.DebugWithTrace(ctx, "设置实例缓存",
		"tenantId", tenantId,
		"serviceGroupId", instance.ServiceGroupId,
		"serviceName", instance.ServiceName,
		"instanceId", instance.ServiceInstanceId)

	return nil
}

// DeleteInstance 删除服务实例
func (c *MemoryCache) DeleteInstance(ctx context.Context, tenantId, instanceId string) error {
	// 先尝试获取服务信息以确定注册类型
	// 通过索引快速定位实例
	value, exists := c.instanceIndex.Load(instanceId)
	if !exists {
		// 实例不在内存索引中，可能是外部注册中心的实例
		// 外部实例由外部注册中心维护，内存缓存不处理
		logger.DebugWithTrace(ctx, "外部注册中心实例删除",
			"tenantId", tenantId,
			"instanceId", instanceId,
			"note", "外部实例由外部注册中心维护，内存缓存不处理")
		return nil
	}

	loc := value.(instanceLocation)

	// 验证租户ID
	if loc.tenantId != tenantId {
		return nil // 实例租户不匹配，视为删除成功
	}

	// 获取服务信息以确定注册类型
	service, err := c.GetService(ctx, tenantId, loc.serviceGroupId, loc.serviceName)
	if err == nil && service.IsExternalRegistry() {
		// 外部注册中心的服务，外部实例由外部注册中心维护，内存缓存不处理
		logger.DebugWithTrace(ctx, "外部注册中心实例删除",
			"tenantId", tenantId,
			"instanceId", instanceId,
			"serviceName", service.ServiceName,
			"registryType", service.RegistryType,
			"note", "外部实例由外部注册中心维护，内存缓存不处理")
		return nil
	}

	// 内部注册中心的服务，正常处理内存缓存删除
	// 删除实例索引（先删除索引，避免后续操作失败时出现不一致）
	c.instanceIndex.Delete(instanceId)

	// 获取现有的租户服务组映射
	value, exists = c.tenants.Load(tenantId)
	if !exists {
		return nil // 租户不存在，视为删除成功
	}

	oldGroups := value.(map[string]*core.ServiceGroup)
	oldGroup, exists := oldGroups[loc.serviceGroupId]
	if !exists {
		return nil // 服务组不存在，视为删除成功
	}

	oldService, exists := oldGroup.Services[loc.serviceName]
	if !exists {
		return nil // 服务不存在，视为删除成功
	}

	// 检查实例是否存在
	instanceFound := false
	for _, instance := range oldService.Instances {
		if instance.ServiceInstanceId == instanceId {
			instanceFound = true
			break
		}
	}

	if !instanceFound {
		return nil // 实例不存在，视为删除成功
	}

	// 创建新的服务组映射（深拷贝）
	newGroups := make(map[string]*core.ServiceGroup, len(oldGroups))

	// 复制所有服务组，除了目标服务组
	for groupId, group := range oldGroups {
		if groupId == loc.serviceGroupId {
			continue // 稍后会单独处理目标服务组
		}
		newGroups[groupId] = group // 直接复制非目标服务组的指针
	}

	// 使用浅拷贝创建新的服务组
	newGroup := oldGroup.ShallowCopy()
	newGroup.Services = make(map[string]*core.Service)

	// 复制所有服务，除了目标服务
	for svcName, service := range oldGroup.Services {
		if svcName == loc.serviceName {
			continue // 稍后会单独处理目标服务
		}
		newGroup.Services[svcName] = service // 直接复制非目标服务的指针
	}

	// 使用浅拷贝创建新的服务
	newService := oldService.ShallowCopy()
	newService.Instances = make([]*core.ServiceInstance, 0, len(oldService.Instances)-1)

	// 复制所有实例，除了要删除的实例
	for _, instance := range oldService.Instances {
		if instance.ServiceInstanceId != instanceId {
			newService.Instances = append(newService.Instances, instance)
		}
	}

	// 始终保留服务，即使没有实例
	// 因为实例可能是临时的，删除后还会重新注册
	// 只有通过 DeleteService 方法才应该删除服务本身
	newGroup.Services[loc.serviceName] = newService

	// 始终保留服务组，即使没有服务
	// 服务组是持久性配置，不应该因为临时实例删除而消失
	// 只有通过 DeleteServiceGroup 方法才应该删除服务组本身
	newGroups[loc.serviceGroupId] = newGroup

	// 原子性更新租户映射
	// 由于我们始终保留服务组和服务，租户映射也始终存在
	c.tenants.Store(tenantId, newGroups)

	logger.DebugWithTrace(ctx, "删除实例缓存",
		"tenantId", tenantId,
		"serviceGroupId", loc.serviceGroupId,
		"serviceName", loc.serviceName,
		"instanceId", instanceId)

	return nil
}

// ListInstances 列出服务下的所有实例
func (c *MemoryCache) ListInstances(ctx context.Context, tenantId, serviceGroupId, serviceName string) ([]*core.ServiceInstance, error) {
	// 获取租户的服务组映射
	groups, exists := c.getTenantGroups(tenantId)
	if !exists {
		// 租户不存在，返回空列表
		logger.DebugWithTrace(ctx, "租户不存在，返回空实例列表",
			"tenantId", tenantId,
			"serviceGroupId", serviceGroupId,
			"serviceName", serviceName)
		return make([]*core.ServiceInstance, 0), nil
	}

	// 获取指定的服务组
	group, exists := groups[serviceGroupId]
	if !exists {
		// 服务组不存在，返回空列表
		logger.DebugWithTrace(ctx, "服务组不存在，返回空实例列表",
			"tenantId", tenantId,
			"serviceGroupId", serviceGroupId,
			"serviceName", serviceName)
		return make([]*core.ServiceInstance, 0), nil
	}

	if group.Services == nil {
		// 服务组没有服务，返回空列表
		logger.DebugWithTrace(ctx, "服务组没有服务，返回空实例列表",
			"tenantId", tenantId,
			"serviceGroupId", serviceGroupId,
			"serviceName", serviceName)
		return make([]*core.ServiceInstance, 0), nil
	}

	// 获取指定的服务
	service, exists := group.Services[serviceName]
	if !exists {
		// 服务不存在，返回空列表
		logger.DebugWithTrace(ctx, "服务不存在，返回空实例列表",
			"tenantId", tenantId,
			"serviceGroupId", serviceGroupId,
			"serviceName", serviceName)
		return make([]*core.ServiceInstance, 0), nil
	}

	// 如果是外部注册中心的服务，直接从外部获取实例
	if service.IsExternalRegistry() {
		return c.externalManager.GetServiceInstances(ctx, service)
	}

	// 内部注册中心的服务，从内存缓存获取
	if service.Instances == nil {
		return make([]*core.ServiceInstance, 0), nil
	}

	// 创建浅拷贝副本返回，确保缓存数据不会被外部修改
	result := make([]*core.ServiceInstance, len(service.Instances))
	for i, instance := range service.Instances {
		result[i] = instance.ShallowCopy()
	}

	logger.DebugWithTrace(ctx, "列出实例缓存",
		"tenantId", tenantId,
		"serviceGroupId", serviceGroupId,
		"serviceName", serviceName,
		"count", len(result))

	return result, nil
}

// DiscoverInstance 发现一个健康的服务实例
func (c *MemoryCache) DiscoverInstance(ctx context.Context, tenantId, serviceGroupId, serviceName string) (*core.ServiceInstance, error) {
	// 获取服务配置
	service, err := c.GetService(ctx, tenantId, serviceGroupId, serviceName)
	if err != nil {
		return nil, fmt.Errorf("获取服务配置失败: %w", err)
	}

	// 如果是外部注册中心的服务，直接从外部获取实例
	if service.IsExternalRegistry() {
		return c.externalManager.DiscoverHealthyInstance(ctx, service)
	}

	// 获取所有实例（包括不健康的实例）
	allInstances, err := c.ListInstances(ctx, tenantId, serviceGroupId, serviceName)
	if err != nil {
		return nil, err
	}

	// 如果没有实例，返回错误
	if len(allInstances) == 0 {
		return nil, fmt.Errorf("没有找到服务实例: %s", serviceName)
	}

	strategy := service.LoadBalanceStrategy

	// 创建服务标识
	serviceKey := ServiceKey{
		TenantId:       tenantId,
		ServiceGroupId: serviceGroupId,
		ServiceName:    serviceName,
	}

	// 使用负载均衡器来选择健康实例（负载均衡器会自动筛选健康实例）
	lb := c.lbFactory.GetLoadBalancer(strategy, serviceKey)
	selectedInstance, err := lb.Select(ctx, allInstances)
	if err != nil {
		return nil, err
	}

	logger.DebugWithTrace(ctx, "发现健康服务实例",
		"tenantId", tenantId,
		"serviceGroupId", serviceGroupId,
		"serviceName", serviceName,
		"strategy", strategy,
		"instanceId", selectedInstance.ServiceInstanceId,
		"address", fmt.Sprintf("%s:%d", selectedInstance.HostAddress, selectedInstance.PortNumber))

	return selectedInstance, nil
}

// UpdateInstanceHealth 更新实例健康状态
// 这个方法直接修改现有实例的健康状态，避免不必要的深拷贝操作
// 注意：这个方法直接修改内存中的实例对象，调用者需要确保线程安全
func (c *MemoryCache) UpdateInstanceHealth(ctx context.Context, tenantId, instanceId string, status string, checkTime time.Time) error {
	// 通过索引快速定位实例
	value, exists := c.instanceIndex.Load(instanceId)
	if !exists {
		// 对于外部注册中心的实例，不在内存中维护，直接返回成功
		logger.DebugWithTrace(ctx, "实例不在内存索引中，可能是外部注册中心实例，跳过健康状态更新",
			"tenantId", tenantId,
			"instanceId", instanceId,
			"status", status)
		return nil
	}

	loc := value.(instanceLocation)

	// 验证租户ID
	if loc.tenantId != tenantId {
		return fmt.Errorf("实例租户不匹配: %s", instanceId)
	}

	// 获取服务信息以确定注册类型
	service, err := c.GetService(ctx, tenantId, loc.serviceGroupId, loc.serviceName)
	if err != nil {
		return fmt.Errorf("获取服务信息失败: %w", err)
	}

	// 如果是外部注册中心的服务，不更新内存中的实例状态
	if service.IsExternalRegistry() {
		logger.DebugWithTrace(ctx, "外部注册中心服务实例健康状态由外部维护",
			"tenantId", tenantId,
			"instanceId", instanceId,
			"serviceName", service.ServiceName,
			"registryType", service.RegistryType,
			"status", status)
		return nil
	}

	// 内部注册中心的服务，更新内存中的实例健康状态
	// 获取服务组
	groups, exists := c.getTenantGroups(tenantId)
	if !exists {
		return fmt.Errorf("租户缓存未命中: %s", tenantId)
	}

	// 获取服务组
	group, exists := groups[loc.serviceGroupId]
	if !exists {
		return fmt.Errorf("服务组缓存未命中: %s", loc.serviceGroupId)
	}

	// 获取服务
	serviceFromGroup, exists := group.Services[loc.serviceName]
	if !exists {
		return fmt.Errorf("服务缓存未命中: %s", loc.serviceName)
	}

	// 查找并更新实例
	for _, instance := range serviceFromGroup.Instances {
		if instance.ServiceInstanceId == instanceId {
			// 记录原始状态用于调试
			oldStatus := instance.HealthStatus

			// 更新健康状态
			instance.HealthStatus = status
			instance.LastHealthCheckTime = &checkTime

			logger.DebugWithTrace(ctx, "更新实例健康状态",
				"tenantId", tenantId,
				"instanceId", instanceId,
				"oldStatus", oldStatus,
				"newStatus", status,
				"checkTime", checkTime)

			return nil
		}
	}

	return fmt.Errorf("实例缓存未命中: %s", instanceId)
}

// GetStats 获取缓存统计信息
func (c *MemoryCache) GetStats() core.CacheStats {
	c.stats.RLock()
	defer c.stats.RUnlock()

	tenantCount := 0
	groupCount := 0
	serviceCount := 0
	instanceCount := 0

	c.tenants.Range(func(_, value interface{}) bool {
		tenantCount++
		groups := value.(map[string]*core.ServiceGroup)

		for _, group := range groups {
			groupCount++

			if group.Services != nil {
				for _, service := range group.Services {
					serviceCount++

					if service.Instances != nil {
						instanceCount += len(service.Instances)
					}
				}
			}
		}

		return true
	})

	totalEntries := int64(tenantCount + groupCount + serviceCount + instanceCount)
	hitRate := float64(0)
	if c.stats.hitCount+c.stats.missCount > 0 {
		hitRate = float64(c.stats.hitCount) / float64(c.stats.hitCount+c.stats.missCount)
	}

	return core.CacheStats{
		HitCount:     c.stats.hitCount,
		MissCount:    c.stats.missCount,
		HitRate:      hitRate,
		TotalEntries: totalEntries,
	}
}

// Clear 清空所有缓存
func (c *MemoryCache) Clear() error {
	tenantCount := 0

	c.tenants.Range(func(key, _ interface{}) bool {
		c.tenants.Delete(key)
		tenantCount++
		return true
	})

	c.instanceIndex.Range(func(key, _ interface{}) bool {
		c.instanceIndex.Delete(key)
		return true
	})

	// 关闭外部注册中心连接
	if closeErr := c.externalManager.Close(); closeErr != nil {
		logger.Warn("关闭外部注册中心连接失败", "error", closeErr)
	}

	logger.Info("清空缓存", "tenantCount", tenantCount)
	return nil
}
