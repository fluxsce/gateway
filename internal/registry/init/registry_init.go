package init

import (
	"context"
	"fmt"

	"gateway/internal/registry/cache"
	"gateway/internal/registry/core"
	"gateway/internal/registry/event"
	"gateway/internal/registry/manager"
	"gateway/pkg/database"
	"gateway/pkg/logger"
)

// RegistryInitializer 注册中心初始化器
// 负责从数据库加载配置并初始化注册中心
type RegistryInitializer struct {
	db         database.Database
	tenantId   string
	activeFlag string
}

// NewRegistryInitializer 创建注册中心初始化器
func NewRegistryInitializer(db database.Database) *RegistryInitializer {
	return &RegistryInitializer{
		db:         db,
		tenantId:   "default", // 默认租户ID
		activeFlag: "Y",       // 默认只加载活跃的记录
	}
}

// SetTenantId 设置租户ID
func (i *RegistryInitializer) SetTenantId(tenantId string) *RegistryInitializer {
	i.tenantId = tenantId
	return i
}

// SetActiveFlag 设置活跃标志
func (i *RegistryInitializer) SetActiveFlag(activeFlag string) *RegistryInitializer {
	i.activeFlag = activeFlag
	return i
}

// Initialize 初始化注册中心
// 从数据库加载配置并初始化注册中心组件
// 返回管理器实例和布尔值，布尔值表示是否成功初始化了注册中心
func (i *RegistryInitializer) Initialize(ctx context.Context) (core.Manager, bool) {
	// 1. 创建缓存
	cacheStorage := i.createCacheStorage()

	// 2. 创建事件发布器
	eventPublisher, err := i.createEventPublisher()
	if err != nil {
		logger.ErrorWithTrace(ctx, "创建事件发布器失败", "error", err)
		return nil, false
	}

	// 3. 初始化注册中心管理器
	mgr := manager.InitInstance(cacheStorage, eventPublisher)

	// 4. 加载服务分组数据
	err = i.loadServiceGroups(ctx, mgr)
	if err != nil {
		logger.WarnWithTrace(ctx, "加载服务分组数据失败", "error", err)
		// 如果查询不到服务分组，直接返回false，表示不使用注册中心
		return mgr, false
	}

	// 5. 加载服务数据
	err = i.loadServices(ctx, mgr)
	if err != nil {
		logger.WarnWithTrace(ctx, "加载服务数据失败", "error", err)
		// 继续初始化，不中断流程
	}

	// 6. 加载服务实例数据
	err = i.loadServiceInstances(ctx, mgr)
	if err != nil {
		logger.WarnWithTrace(ctx, "加载服务实例数据失败", "error", err)
		// 继续初始化，不中断流程
	}

	logger.InfoWithTrace(ctx, "注册中心初始化完成")
	return mgr, true
}

// createCacheStorage 创建缓存存储
func (i *RegistryInitializer) createCacheStorage() core.CacheStorage {
	// 直接使用内存缓存
	return cache.NewMemoryCache()
}

// createEventPublisher 创建事件发布器
func (i *RegistryInitializer) createEventPublisher() (core.EventPublisher, error) {
	// 默认使用内存事件发布器
	return event.NewMemoryEventPublisher()
}

// loadServiceGroups 加载服务分组数据
// 如果查询不到服务分组，返回特定错误，表示不使用注册中心
func (i *RegistryInitializer) loadServiceGroups(ctx context.Context, mgr core.Manager) error {
	// 构建查询SQL - 使用全字段查询绑定
	query := `SELECT serviceGroupId, tenantId, groupName, groupDescription, groupType,
		ownerUserId, adminUserIds, readUserIds, accessControlEnabled,
		defaultProtocolType, defaultLoadBalanceStrategy, defaultHealthCheckUrl, defaultHealthCheckIntervalSeconds,
		addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty,
		reserved1, reserved2, reserved3, reserved4, reserved5, reserved6, reserved7, reserved8, reserved9, reserved10
	FROM HUB_REGISTRY_SERVICE_GROUP WHERE tenantId = ? AND activeFlag = ? ORDER BY addTime DESC`

	// 执行查询
	var serviceGroups []*core.ServiceGroup
	err := i.db.Query(ctx, &serviceGroups, query, []interface{}{i.tenantId, i.activeFlag}, true)
	if err != nil {
		return fmt.Errorf("查询服务分组数据失败: %w", err)
	}

	// 如果没有查询到服务分组，返回特定错误，表示不使用注册中心
	// if len(serviceGroups) == 0 {
	// 	logger.InfoWithTrace(ctx, "未查询到服务分组数据，不使用注册中心")
	// 	return fmt.Errorf("未查询到服务分组数据，不使用注册中心")
	// }

	// 初始化服务映射并缓存
	for _, group := range serviceGroups {
		// 仅初始化非数据库字段
		group.Services = make(map[string]*core.Service)

		// 缓存服务分组
		cache := mgr.(*manager.RegistryManager).GetCache()
		err = cache.SetServiceGroup(ctx, group.TenantId, group)
		if err != nil {
			logger.WarnWithTrace(ctx, "缓存服务分组失败",
				"serviceGroupId", group.ServiceGroupId,
				"error", err)
		}
	}

	logger.InfoWithTrace(ctx, "加载服务分组数据完成", "count", len(serviceGroups))
	return nil
}

// loadServices 加载服务数据
func (i *RegistryInitializer) loadServices(ctx context.Context, mgr core.Manager) error {
	// 构建查询SQL - 使用全字段查询绑定
	query := `SELECT tenantId, serviceName, serviceGroupId, groupName, serviceDescription,
		protocolType, contextPath, loadBalanceStrategy, 
		healthCheckUrl, healthCheckIntervalSeconds, healthCheckTimeoutSeconds, healthCheckType, healthCheckMode,
		metadataJson, tagsJson, 
		addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty,
		reserved1, reserved2, reserved3, reserved4, reserved5, reserved6, reserved7, reserved8, reserved9, reserved10
	FROM HUB_REGISTRY_SERVICE WHERE tenantId = ? AND activeFlag = ? ORDER BY addTime DESC`

	// 执行查询
	var services []*core.Service
	err := i.db.Query(ctx, &services, query, []interface{}{i.tenantId, i.activeFlag}, true)
	if err != nil {
		return fmt.Errorf("查询服务数据失败: %w", err)
	}

	// 初始化实例列表并缓存
	for _, service := range services {
		// 仅初始化非数据库字段
		service.Instances = make([]*core.ServiceInstance, 0)

		// 缓存服务
		cache := mgr.(*manager.RegistryManager).GetCache()
		err = cache.SetService(ctx, service.TenantId, service)
		if err != nil {
			logger.WarnWithTrace(ctx, "缓存服务失败",
				"serviceName", service.ServiceName,
				"error", err)
		}
	}

	logger.InfoWithTrace(ctx, "加载服务数据完成", "count", len(services))
	return nil
}

// loadServiceInstances 加载服务实例数据
func (i *RegistryInitializer) loadServiceInstances(ctx context.Context, mgr core.Manager) error {
	// 构建查询SQL - 使用全字段查询绑定
	query := `SELECT serviceInstanceId, tenantId, serviceGroupId, serviceName, groupName,
		hostAddress, portNumber, contextPath,
		instanceStatus, healthStatus, weightValue,
		clientId, clientVersion, clientType, tempInstanceFlag, heartbeatFailCount,
		metadataJson, tagsJson, 
		registerTime, lastHeartbeatTime, lastHealthCheckTime,
		addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty,
		reserved1, reserved2, reserved3, reserved4, reserved5, reserved6, reserved7, reserved8, reserved9, reserved10
	FROM HUB_REGISTRY_SERVICE_INSTANCE WHERE tenantId = ? AND activeFlag = ? ORDER BY addTime DESC`

	// 执行查询
	var instances []*core.ServiceInstance
	err := i.db.Query(ctx, &instances, query, []interface{}{i.tenantId, i.activeFlag}, true)
	if err != nil {
		return fmt.Errorf("查询服务实例数据失败: %w", err)
	}

	// 缓存每个实例
	for _, instance := range instances {

		// 缓存单个实例
		cache := mgr.(*manager.RegistryManager).GetCache()
		err = cache.SetInstance(ctx, instance.TenantId, instance)
		if err != nil {
			logger.WarnWithTrace(ctx, "缓存服务实例失败",
				"instanceId", instance.ServiceInstanceId,
				"error", err)
		}
	}

	logger.InfoWithTrace(ctx, "加载服务实例数据完成", "count", len(instances))
	return nil
}
