package manager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gateway/internal/servicecenter/cache"
	"gateway/internal/servicecenter/dao"
	"gateway/internal/servicecenter/server"
	pb "gateway/internal/servicecenter/server/proto"
	"gateway/internal/servicecenter/types"
	"gateway/pkg/database"
	"gateway/pkg/logger"
)

// ServiceCenterManager 服务中心管理器
// 管理共享的 DAO 层和服务中心实例连接池
// 注意：Cache 使用全局单例，无需在此管理
type ServiceCenterManager struct {
	// 数据库连接
	db database.Database

	// 服务中心实例连接池 - key: instanceName, value: *server.Server
	instances map[string]*server.Server
	mu        sync.RWMutex

	// 共享的 DAO 层（所有实例共享）
	namespaceDAO *dao.NamespaceDAO
	serviceDAO   *dao.ServiceDAO
	nodeDAO      *dao.NodeDAO // 服务节点 DAO（HUB_SERVICE_NODE）
	configDAO    *dao.ConfigDAO
	historyDAO   *dao.HistoryDAO
	instanceDAO  *dao.InstanceDAO // 服务中心实例配置 DAO（HUB_SERVICE_INSTANCE）

	// 健康检查器和同步逻辑
	healthCheckers map[string]*HealthChecker // key: instanceName, value: *HealthChecker
	hcMu           sync.RWMutex              // 保护 healthCheckers 的并发访问

	// 事件通知器（辅助类）
	eventNotifier *EventNotifier
}

// NewServiceCenterManager 创建服务中心管理器
func NewServiceCenterManager(db database.Database) *ServiceCenterManager {
	manager := &ServiceCenterManager{
		db:             db,
		instances:      make(map[string]*server.Server),
		healthCheckers: make(map[string]*HealthChecker),
	}

	// 初始化共享的 DAO 层
	manager.namespaceDAO = dao.NewNamespaceDAO(db)
	manager.serviceDAO = dao.NewServiceDAO(db)
	manager.nodeDAO = dao.NewNodeDAO(db)
	manager.configDAO = dao.NewConfigDAO(db)
	manager.historyDAO = dao.NewHistoryDAO(db)
	manager.instanceDAO = dao.NewInstanceDAO(db)

	// 初始化事件通知器
	manager.eventNotifier = NewEventNotifier(manager)

	logger.Info("服务中心管理器创建完成")
	return manager
}

// ========== 实例连接池管理 ==========

// ========== 内部辅助方法 ==========

// ForEachInstance 遍历所有实例（内部使用）
func (m *ServiceCenterManager) ForEachInstance(fn func(string, *server.Server) error) error {
	m.mu.RLock()
	instances := make(map[string]*server.Server)
	for name, srv := range m.instances {
		instances[name] = srv
	}
	m.mu.RUnlock()

	for name, srv := range instances {
		if err := fn(name, srv); err != nil {
			return err
		}
	}
	return nil
}

// LoadInstancesFromDB 从数据库加载实例配置并创建 Server
func (m *ServiceCenterManager) LoadInstancesFromDB(ctx context.Context, tenantId, environment string) error {
	// 查询指定租户和环境的所有实例配置
	configs, err := m.instanceDAO.ListInstances(ctx, tenantId, environment)
	if err != nil {
		return fmt.Errorf("加载实例配置失败: %w", err)
	}

	if len(configs) == 0 {
		logger.Warn("未找到任何服务中心实例配置",
			"tenantId", tenantId,
			"environment", environment)
		return nil
	}

	// 逐个创建 Server 并添加到实例池
	var errors []string
	for _, config := range configs {
		m.mu.Lock()
		// 检查实例是否已存在
		if _, exists := m.instances[config.InstanceName]; exists {
			m.mu.Unlock()
			errors = append(errors, fmt.Sprintf("%s: 已存在", config.InstanceName))
			logger.Warn("服务中心实例已存在", "instanceName", config.InstanceName)
			continue
		}

		// 创建 Server 并添加到实例池
		srv := server.NewServer(m.db, config)
		m.instances[config.InstanceName] = srv
		m.mu.Unlock()

		// 为每个实例创建健康检查器
		m.createHealthChecker(config.InstanceName, config.TenantID)

		logger.Info("服务中心实例已创建",
			"instanceName", config.InstanceName,
			"listenAddr", config.ListenAddress,
			"listenPort", config.ListenPort)
	}

	if len(errors) > 0 {
		return fmt.Errorf("部分实例创建失败: %v", errors)
	}

	logger.Info("所有服务中心实例加载完成", "count", len(configs))
	return nil
}

// LoadAllInstancesFromDB 从数据库加载指定租户的所有实例配置并创建 Server（所有环境）
// 同时从数据库恢复命名空间和服务到缓存
func (m *ServiceCenterManager) LoadAllInstancesFromDB(ctx context.Context, tenantId string) error {
	// 查询指定租户的所有实例配置（所有环境）
	configs, err := m.instanceDAO.ListAllInstances(ctx, tenantId)
	if err != nil {
		return fmt.Errorf("加载实例配置失败: %w", err)
	}

	if len(configs) == 0 {
		logger.Warn("未找到任何服务中心实例配置",
			"tenantId", tenantId)
		return nil
	}

	// 逐个创建 Server 并添加到实例池
	var errors []string
	for _, config := range configs {
		m.mu.Lock()
		// 检查实例是否已存在（使用 instanceName 作为唯一标识）
		if _, exists := m.instances[config.InstanceName]; exists {
			m.mu.Unlock()
			errors = append(errors, fmt.Sprintf("%s: 已存在", config.InstanceName))
			logger.Warn("服务中心实例已存在", "instanceName", config.InstanceName)
			continue
		}

		// 创建 Server 并添加到实例池
		srv := server.NewServer(m.db, config)
		m.instances[config.InstanceName] = srv
		m.mu.Unlock()

		// 为每个实例创建健康检查器
		m.createHealthChecker(config.InstanceName, config.TenantID)

		logger.Info("服务中心实例已创建",
			"instanceName", config.InstanceName,
			"environment", config.Environment,
			"listenAddr", config.ListenAddress,
			"listenPort", config.ListenPort)
	}

	if len(errors) > 0 {
		return fmt.Errorf("部分实例创建失败: %v", errors)
	}

	logger.Info("所有服务中心实例加载完成", "count", len(configs))

	// 从数据库恢复命名空间到缓存
	if err := m.loadNamespacesToCache(ctx, tenantId); err != nil {
		logger.Warn("加载命名空间到缓存失败", "error", err)
		// 命名空间加载失败不影响主流程，只记录警告
	}

	// 从数据库恢复服务和节点到缓存
	if err := m.loadServicesToCache(ctx, tenantId); err != nil {
		logger.Warn("加载服务和节点到缓存失败", "error", err)
		// 服务加载失败不影响主流程，只记录警告
	}

	return nil
}

// RemoveInstance 移除服务中心实例
func (m *ServiceCenterManager) RemoveInstance(ctx context.Context, instanceName string) error {
	if instanceName == "" {
		return fmt.Errorf("实例名称不能为空")
	}

	m.mu.Lock()
	srv, exists := m.instances[instanceName]
	if !exists {
		m.mu.Unlock()
		return fmt.Errorf("服务中心实例 '%s' 不存在", instanceName)
	}
	m.mu.Unlock()

	// 如果实例正在运行，先停止
	if srv != nil && srv.IsRunning() {
		srv.Stop(ctx)
	}

	// 停止并移除健康检查器
	m.stopHealthChecker(instanceName)

	// 从实例池中删除
	m.mu.Lock()
	delete(m.instances, instanceName)
	m.mu.Unlock()

	logger.Info("服务中心实例已从连接池移除", "instanceName", instanceName)
	return nil
}

// GetInstance 获取指定的服务中心实例
// 如果实例不存在，返回 nil
func (m *ServiceCenterManager) GetInstance(instanceName string) *server.Server {
	if instanceName == "" {
		return nil
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	srv, exists := m.instances[instanceName]
	if !exists {
		return nil // 实例不存在，返回 nil
	}

	return srv
}

// ========== 实例生命周期管理 ==========

// StartInstance 启动指定的服务中心实例（同步方法）
//
// 处理流程：
//  1. 如果实例不存在，尝试从数据库加载配置并创建 Server
//  2. 检查实例是否已在运行
//  3. 直接同步调用 Server.Start()（启动处理逻辑在 Server 内部）
//
// 注意：
//   - 此方法是同步的，会等待服务器启动完成
//   - Server.Start() 内部处理 goroutine、WaitGroup、错误检查等
//   - manager 只负责同步调用，不处理启动细节
func (m *ServiceCenterManager) StartInstance(ctx context.Context, tenantId, instanceName, environment string) error {
	// 获取实例
	srv := m.GetInstance(instanceName)
	if srv == nil {
		// 实例不存在，尝试从数据库加载配置并创建 Server
		config, err := m.instanceDAO.GetInstance(ctx, tenantId, instanceName, environment)
		if err != nil {
			return fmt.Errorf("加载实例配置失败: %w", err)
		}
		if config == nil {
			return fmt.Errorf("实例配置不存在: %s", instanceName)
		}

		// 创建 Server
		m.mu.Lock()
		if _, exists := m.instances[instanceName]; exists {
			m.mu.Unlock()
			return fmt.Errorf("服务中心实例 '%s' 已存在", instanceName)
		}
		srv = server.NewServer(m.db, config)
		m.instances[instanceName] = srv
		m.mu.Unlock()

		// 为实例创建健康检查器
		m.createHealthChecker(instanceName, config.TenantID)

		logger.Info("从数据库创建实例成功", "instanceName", instanceName)
	}

	// 检查是否已在运行
	if srv.IsRunning() {
		return fmt.Errorf("服务中心实例 '%s' 已在运行中", instanceName)
	}

	logger.Info("启动服务中心实例", "instanceName", instanceName)

	// 直接同步调用 Server.Start()（启动处理逻辑在 Server 内部）
	if err := srv.Start(ctx); err != nil {
		return fmt.Errorf("启动 gRPC 服务器失败: %w", err)
	}

	logger.Info("服务中心实例启动成功", "instanceName", instanceName)
	return nil
}

// StopInstance 停止指定的服务中心实例（同步方法）
//
// 处理流程：
//  1. 获取实例
//  2. 检查实例是否正在运行
//  3. 直接同步调用 Server.Stop()（停止处理逻辑在 Server 内部）
//
// 注意：
//   - 此方法是同步的，会等待服务器停止完成
//   - Server.Stop() 内部处理停止信号、WaitGroup、优雅关闭等
//   - manager 只负责同步调用，不处理停止细节
func (m *ServiceCenterManager) StopInstance(ctx context.Context, instanceName string) error {
	srv := m.GetInstance(instanceName)
	if srv == nil {
		// 实例不存在，直接返回
		return nil
	}

	// 检查是否正在运行
	if !srv.IsRunning() {
		return nil // 未运行，直接返回（参考网关模式）
	}

	logger.Info("正在停止服务中心实例...", "instanceName", instanceName)

	// 直接同步调用 Server.Stop()（停止处理逻辑在 Server 内部）
	srv.Stop(ctx)

	// 停止健康检查器
	m.stopHealthChecker(instanceName)

	// 停止后从实例池中删除
	m.mu.Lock()
	delete(m.instances, instanceName)
	m.mu.Unlock()

	logger.Info("服务中心实例已停止并从连接池移除", "instanceName", instanceName)
	return nil
}

// ========== 配置管理（支持前端动态操作）==========

// ReloadInstance 重载指定实例的配置（从数据库重新加载）
//
// 处理流程：
//  1. 获取 Server 实例
//  2. 从 Server 获取当前配置
//  3. 从数据库重新加载配置
//  4. 如果实例正在运行，尝试调用 Server.Reload() 进行热重载
//  5. 如果 Server.Reload() 返回错误，直接返回错误（不执行重启）
//
// 注意：
//   - 某些配置变更（如监听地址、TLS）需要重启才能生效
//   - 某些配置变更（如认证、IP 白名单）可以热重载，无需重启
//   - 重载失败直接返回错误，不执行重启流程
func (m *ServiceCenterManager) ReloadInstance(ctx context.Context, instanceName string) error {
	srv := m.GetInstance(instanceName)
	if srv == nil {
		return fmt.Errorf("服务中心实例 '%s' 不存在", instanceName)
	}

	// 从 Server 获取当前配置
	currentConfig := srv.GetConfig()
	if currentConfig == nil {
		return fmt.Errorf("无法获取当前配置: %s", instanceName)
	}

	// 从数据库重新加载配置
	newConfig, err := m.instanceDAO.GetInstance(ctx,
		currentConfig.TenantID,
		currentConfig.InstanceName,
		currentConfig.Environment)
	if err != nil {
		return fmt.Errorf("重载配置失败: %w", err)
	}

	if newConfig == nil {
		return fmt.Errorf("实例配置不存在: %s", instanceName)
	}

	// 如果实例正在运行，尝试热重载
	if srv.IsRunning() {
		// 直接调用 Server.Reload() 方法
		if err := srv.Reload(ctx, newConfig); err != nil {
			// 热重载失败，直接返回错误
			return fmt.Errorf("热重载失败: %w", err)
		}

		logger.Info("服务中心实例配置重载成功（热重载，无需重启）", "instanceName", instanceName)
		return nil
	}

	// 实例未运行，无法重载（需要先启动）
	logger.Info("服务中心实例未运行，无法重载配置", "instanceName", instanceName)
	return nil
}

// ========== 手动触发订阅事件 ==========

// NotifyServiceChange 手动触发服务变更事件通知
//
// 处理流程：
//  1. 获取指定实例的 gRPC 服务器
//  2. 获取 RegistryHandler 和 ServiceSubscriber
//  3. 调用 NotifyServiceChange 通知所有订阅者
//
// 使用场景：
//   - 数据库变更后需要主动推送事件给客户端
//   - 外部系统同步数据后需要通知订阅者
//   - 手动触发服务变更通知
func (m *ServiceCenterManager) NotifyServiceChange(ctx context.Context, instanceName, tenantId, namespaceId, groupName, serviceName string, event *pb.ServiceChangeEvent) error {
	srv := m.GetInstance(instanceName)
	if srv == nil {
		return fmt.Errorf("服务中心实例 '%s' 不存在", instanceName)
	}

	// 获取 RegistryHandler
	registryHandler := srv.GetRegistryHandler()
	if registryHandler == nil {
		return fmt.Errorf("服务中心实例 '%s' 的 RegistryHandler 未初始化", instanceName)
	}

	// 获取 ServiceSubscriber 并触发事件
	serviceSubMgr := registryHandler.GetServiceSubscriber()
	if serviceSubMgr == nil {
		return fmt.Errorf("服务中心实例 '%s' 的 ServiceSubscriber 未初始化", instanceName)
	}

	// 确保事件时间戳已设置
	if event.Timestamp == "" {
		event.Timestamp = time.Now().Format("2006-01-02 15:04:05")
	}

	// 触发事件通知
	serviceSubMgr.NotifyServiceChange(tenantId, namespaceId, groupName, serviceName, event)

	logger.Info("手动触发服务变更事件通知",
		"instanceName", instanceName,
		"namespaceId", namespaceId,
		"groupName", groupName,
		"serviceName", serviceName,
		"eventType", event.EventType)

	return nil
}

// NotifyConfigChange 手动触发配置变更事件通知
//
// 处理流程：
//  1. 获取指定实例的 gRPC 服务器
//  2. 获取 ConfigHandler 和 ConfigWatcher
//  3. 调用 NotifyConfigChange 通知所有监听者
//
// 使用场景：
//   - 数据库变更后需要主动推送事件给客户端
//   - 外部系统同步数据后需要通知监听者
//   - 手动触发配置变更通知
func (m *ServiceCenterManager) NotifyConfigChange(ctx context.Context, instanceName, tenantId, namespaceId, groupName, configDataId string, event *pb.ConfigChangeEvent) error {
	srv := m.GetInstance(instanceName)
	if srv == nil {
		return fmt.Errorf("服务中心实例 '%s' 不存在", instanceName)
	}

	// 获取 ConfigHandler
	configHandler := srv.GetConfigHandler()
	if configHandler == nil {
		return fmt.Errorf("服务中心实例 '%s' 的 ConfigHandler 未初始化", instanceName)
	}

	// 获取 ConfigWatcher 并触发事件
	configWatcher := configHandler.GetConfigWatcher()
	if configWatcher == nil {
		return fmt.Errorf("服务中心实例 '%s' 的 ConfigWatcher 未初始化", instanceName)
	}

	// 确保事件时间戳已设置
	if event.Timestamp == "" {
		event.Timestamp = time.Now().Format("2006-01-02 15:04:05")
	}

	// 触发事件通知
	configWatcher.NotifyConfigChange(tenantId, namespaceId, groupName, configDataId, event)

	logger.Info("手动触发配置变更事件通知",
		"instanceName", instanceName,
		"namespaceId", namespaceId,
		"groupName", groupName,
		"configDataId", configDataId,
		"eventType", event.EventType)

	return nil
}

// ========== 缓存恢复（初始化时从数据库加载） ==========

// loadNamespacesToCache 从数据库加载命名空间到缓存
func (m *ServiceCenterManager) loadNamespacesToCache(ctx context.Context, tenantId string) error {
	// 查询所有命名空间
	namespaces, err := m.namespaceDAO.ListNamespaces(ctx, tenantId)
	if err != nil {
		return fmt.Errorf("查询命名空间列表失败: %w", err)
	}

	if len(namespaces) == 0 {
		logger.Info("未找到任何命名空间", "tenantId", tenantId)
		return nil
	}

	// 添加到缓存
	globalCache := cache.GetGlobalCache()
	count := 0
	for _, namespace := range namespaces {
		globalCache.SetNamespace(ctx, namespace)
		count++
	}

	logger.Info("命名空间已加载到缓存", "tenantId", tenantId, "count", count)
	return nil
}

// loadServicesToCache 从数据库加载服务和节点到缓存
func (m *ServiceCenterManager) loadServicesToCache(ctx context.Context, tenantId string) error {
	// 先获取所有命名空间
	namespaces, err := m.namespaceDAO.ListNamespaces(ctx, tenantId)
	if err != nil {
		return fmt.Errorf("查询命名空间列表失败: %w", err)
	}

	if len(namespaces) == 0 {
		logger.Info("未找到任何命名空间，跳过服务加载", "tenantId", tenantId)
		return nil
	}

	globalCache := cache.GetGlobalCache()
	serviceCount := 0
	nodeCount := 0

	// 遍历每个命名空间，加载其下的服务和节点
	for _, namespace := range namespaces {
		// 查询该命名空间下的所有服务（从 HUB_SERVICE 表）
		query := `SELECT * FROM HUB_SERVICE 
			WHERE tenantId = ? AND namespaceId = ? AND activeFlag = 'Y'
			ORDER BY groupName, serviceName`
		args := []interface{}{tenantId, namespace.NamespaceId}

		var services []*types.Service
		err := m.db.Query(ctx, &services, query, args, true)
		if err != nil {
			logger.Warn("查询服务列表失败", "error", err, "namespaceId", namespace.NamespaceId)
			continue
		}

		// 为每个服务加载节点并构建服务对象
		for _, service := range services {
			// 查询该服务的所有节点
			nodes, err := m.nodeDAO.DiscoverNodes(ctx, tenantId, namespace.NamespaceId, service.GroupName, service.ServiceName)
			if err != nil {
				logger.Warn("查询服务节点失败", "error", err,
					"namespaceId", namespace.NamespaceId,
					"groupName", service.GroupName,
					"serviceName", service.ServiceName)
				// 即使节点查询失败，也添加服务（节点列表为空）
				service.Nodes = []*types.ServiceNode{}
			} else {
				// 设置节点列表
				service.Nodes = nodes
				nodeCount += len(nodes)
			}

			// 添加到缓存
			globalCache.SetService(ctx, service)
			serviceCount++
		}
	}

	logger.Info("服务和节点已加载到缓存",
		"tenantId", tenantId,
		"serviceCount", serviceCount,
		"nodeCount", nodeCount)
	return nil
}

// ========== 命名空间缓存管理 ==========

// AddNamespaceToCache 添加命名空间到缓存
//
// 处理流程：
//  1. 从数据库加载命名空间信息
//  2. 将命名空间添加到全局缓存
//
// 使用场景：
//   - 命名空间创建后需要同步到缓存
//   - 命名空间更新后需要刷新缓存
func (m *ServiceCenterManager) AddNamespaceToCache(ctx context.Context, tenantId, namespaceId string) error {
	if tenantId == "" || namespaceId == "" {
		return fmt.Errorf("tenantId和namespaceId不能为空")
	}

	// 从数据库加载命名空间信息
	namespace, err := m.namespaceDAO.GetNamespace(ctx, tenantId, namespaceId)
	if err != nil {
		return fmt.Errorf("加载命名空间信息失败: %w", err)
	}

	if namespace == nil {
		return fmt.Errorf("命名空间不存在: %s", namespaceId)
	}

	// 添加到全局缓存
	globalCache := cache.GetGlobalCache()
	globalCache.SetNamespace(ctx, namespace)

	logger.Info("命名空间已添加到缓存",
		"tenantId", tenantId,
		"namespaceId", namespaceId,
		"namespaceName", namespace.NamespaceName)

	return nil
}

// UpdateNamespaceInCache 更新命名空间缓存
//
// 处理流程：
//  1. 从数据库重新加载命名空间信息
//  2. 更新全局缓存中的命名空间
//
// 使用场景：
//   - 命名空间信息更新后需要刷新缓存
func (m *ServiceCenterManager) UpdateNamespaceInCache(ctx context.Context, tenantId, namespaceId string) error {
	if tenantId == "" || namespaceId == "" {
		return fmt.Errorf("tenantId和namespaceId不能为空")
	}

	// 从数据库重新加载命名空间信息
	namespace, err := m.namespaceDAO.GetNamespace(ctx, tenantId, namespaceId)
	if err != nil {
		return fmt.Errorf("加载命名空间信息失败: %w", err)
	}

	if namespace == nil {
		// 如果命名空间不存在，从缓存中删除
		globalCache := cache.GetGlobalCache()
		globalCache.DeleteNamespace(ctx, tenantId, namespaceId)
		logger.Info("命名空间不存在，已从缓存删除",
			"tenantId", tenantId,
			"namespaceId", namespaceId)
		return nil
	}

	// 更新全局缓存
	globalCache := cache.GetGlobalCache()
	globalCache.SetNamespace(ctx, namespace)

	logger.Info("命名空间缓存已更新",
		"tenantId", tenantId,
		"namespaceId", namespaceId,
		"namespaceName", namespace.NamespaceName)

	return nil
}

// DeleteNamespaceFromCache 从缓存删除命名空间
//
// 处理流程：
//  1. 从全局缓存删除命名空间
//  2. 同时删除该命名空间下的所有服务和节点缓存
//
// 使用场景：
//   - 命名空间删除后需要清理缓存
//   - 命名空间被禁用时需要清理缓存
func (m *ServiceCenterManager) DeleteNamespaceFromCache(ctx context.Context, tenantId, namespaceId string) error {
	if tenantId == "" || namespaceId == "" {
		return fmt.Errorf("tenantId和namespaceId不能为空")
	}

	globalCache := cache.GetGlobalCache()

	// 删除命名空间缓存（会自动删除该命名空间下的所有服务和节点）
	globalCache.DeleteNamespace(ctx, tenantId, namespaceId)

	logger.Info("命名空间及相关服务已从缓存删除",
		"tenantId", tenantId,
		"namespaceId", namespaceId)

	return nil
}

// ========== 服务缓存管理 ==========

// AddServiceToCache 添加服务到缓存（自动通知订阅者）
//
// 处理流程：
//  1. 将传入的服务（包含节点）添加到全局缓存
//  2. 自动触发 SERVICE_ADDED 事件通知所有订阅者
//
// 使用场景：
//   - 服务创建后需要同步到缓存
//   - 服务更新后需要刷新缓存
//
// 参数:
//   - service: 服务对象（应包含节点列表）
func (m *ServiceCenterManager) AddServiceToCache(ctx context.Context, service *types.Service) error {
	if service == nil {
		return fmt.Errorf("服务对象不能为空")
	}

	if service.TenantId == "" || service.NamespaceId == "" || service.GroupName == "" || service.ServiceName == "" {
		return fmt.Errorf("服务的tenantId、namespaceId、groupName和serviceName不能为空")
	}

	// 确保节点列表不为 nil
	if service.Nodes == nil {
		service.Nodes = []*types.ServiceNode{}
	}

	// 添加到全局缓存
	globalCache := cache.GetGlobalCache()
	globalCache.SetService(ctx, service)

	logger.Info("服务已添加到缓存",
		"tenantId", service.TenantId,
		"namespaceId", service.NamespaceId,
		"groupName", service.GroupName,
		"serviceName", service.ServiceName,
		"nodeCount", len(service.Nodes))

	// 自动通知订阅者
	m.eventNotifier.NotifyServiceChange(ctx, service.TenantId, service.NamespaceId, service.GroupName, service.ServiceName, "SERVICE_ADDED")

	return nil
}

// UpdateServiceInCache 更新服务缓存（自动通知订阅者）
//
// 处理流程：
//  1. 从缓存获取现有节点列表（自动保留）
//  2. 将传入的服务信息更新到全局缓存（节点列表会自动保留）
//  3. 自动触发 SERVICE_UPDATED 事件通知所有订阅者
//
// 使用场景：
//   - 服务信息更新后需要刷新缓存
//   - 服务节点变更由其他方法单独处理（如 UpdateNodeInCache）
//
// 注意：
//   - 此方法不会覆盖现有节点列表，节点列表会自动从缓存中保留
//   - 如果需要更新节点，应使用 UpdateNodeInCache 方法
//
// 参数:
//   - service: 服务对象（节点列表会被忽略，自动从缓存保留）
func (m *ServiceCenterManager) UpdateServiceInCache(ctx context.Context, service *types.Service) error {
	if service == nil {
		return fmt.Errorf("服务对象不能为空")
	}

	if service.TenantId == "" || service.NamespaceId == "" || service.GroupName == "" || service.ServiceName == "" {
		return fmt.Errorf("服务的tenantId、namespaceId、groupName和serviceName不能为空")
	}

	// 更新全局缓存（SetService 会自动保留现有节点列表）
	globalCache := cache.GetGlobalCache()
	// 注意：SetService 内部会自动从缓存中获取并保留现有节点列表
	globalCache.SetService(ctx, service)

	// 从缓存获取更新后的服务信息（包含自动保留的节点）
	updatedService, found := globalCache.GetService(ctx, service.TenantId, service.NamespaceId, service.GroupName, service.ServiceName)
	nodeCount := 0
	if found && updatedService != nil {
		nodeCount = len(updatedService.Nodes)
	}

	logger.Info("服务缓存已更新",
		"tenantId", service.TenantId,
		"namespaceId", service.NamespaceId,
		"groupName", service.GroupName,
		"serviceName", service.ServiceName,
		"nodeCount", nodeCount)

	// 自动通知订阅者
	m.eventNotifier.NotifyServiceChange(ctx, service.TenantId, service.NamespaceId, service.GroupName, service.ServiceName, "SERVICE_UPDATED")

	return nil
}

// DeleteServiceFromCache 从缓存删除服务（自动通知订阅者）
//
// 处理流程：
//  1. 从全局缓存删除服务
//  2. 同时删除该服务的所有节点缓存
//  3. 自动触发 SERVICE_DELETED 事件通知所有订阅者
//
// 使用场景：
//   - 服务删除后需要清理缓存
//   - 服务被禁用时需要清理缓存
func (m *ServiceCenterManager) DeleteServiceFromCache(ctx context.Context, tenantId, namespaceId, groupName, serviceName string) error {
	if tenantId == "" || namespaceId == "" || groupName == "" || serviceName == "" {
		return fmt.Errorf("tenantId、namespaceId、groupName和serviceName不能为空")
	}

	globalCache := cache.GetGlobalCache()

	// 删除服务缓存（会自动删除该服务的所有节点）
	globalCache.DeleteService(ctx, tenantId, namespaceId, groupName, serviceName)

	logger.Info("服务及相关节点已从缓存删除",
		"tenantId", tenantId,
		"namespaceId", namespaceId,
		"groupName", groupName,
		"serviceName", serviceName)

	// 自动通知订阅者
	m.eventNotifier.NotifyServiceChange(ctx, tenantId, namespaceId, groupName, serviceName, "SERVICE_DELETED")

	return nil
}

// ========== 节点缓存管理 ==========

// AddNodeToCache 添加节点到缓存（自动通知订阅者）
//
// 处理流程：
//  1. 验证节点信息
//  2. 添加到全局缓存
//  3. 自动触发 NODE_ADDED 事件通知所有订阅者
//
// 使用场景：
//   - 新节点注册后需要同步到缓存
//
// 参数:
//   - node: 节点对象，不能为空
func (m *ServiceCenterManager) AddNodeToCache(ctx context.Context, node *types.ServiceNode) error {
	if node == nil {
		return fmt.Errorf("节点对象不能为空")
	}

	if node.TenantId == "" || node.NodeId == "" {
		return fmt.Errorf("节点的tenantId和nodeId不能为空")
	}

	if node.NamespaceId == "" || node.GroupName == "" || node.ServiceName == "" {
		return fmt.Errorf("节点的namespaceId、groupName和serviceName不能为空")
	}

	// 添加到全局缓存
	globalCache := cache.GetGlobalCache()
	globalCache.AddNode(ctx, node)

	logger.Info("节点已添加到缓存",
		"tenantId", node.TenantId,
		"namespaceId", node.NamespaceId,
		"groupName", node.GroupName,
		"serviceName", node.ServiceName,
		"nodeId", node.NodeId,
		"ipAddress", node.IpAddress,
		"portNumber", node.PortNumber)

	// 自动通知订阅者
	m.eventNotifier.NotifyServiceChange(ctx, node.TenantId, node.NamespaceId, node.GroupName, node.ServiceName, "NODE_ADDED")

	return nil
}

// UpdateNodeInCache 更新节点到缓存（自动通知订阅者）
//
// 处理流程：
//  1. 验证节点信息
//  2. 更新全局缓存中的节点
//  3. 自动触发 NODE_UPDATED 事件通知所有订阅者
//
// 使用场景：
//   - 节点信息更新后需要刷新缓存（如IP、端口、权重、元数据等）
//   - 节点状态变更后需要刷新缓存（如上线、下线）
//
// 参数:
//   - node: 节点对象，不能为空
func (m *ServiceCenterManager) UpdateNodeInCache(ctx context.Context, node *types.ServiceNode) error {
	if node == nil {
		return fmt.Errorf("节点对象不能为空")
	}

	if node.TenantId == "" || node.NodeId == "" {
		return fmt.Errorf("节点的tenantId和nodeId不能为空")
	}

	if node.NamespaceId == "" || node.GroupName == "" || node.ServiceName == "" {
		return fmt.Errorf("节点的namespaceId、groupName和serviceName不能为空")
	}

	// 更新全局缓存
	globalCache := cache.GetGlobalCache()
	globalCache.UpdateNode(ctx, node)

	logger.Info("节点缓存已更新",
		"tenantId", node.TenantId,
		"namespaceId", node.NamespaceId,
		"groupName", node.GroupName,
		"serviceName", node.ServiceName,
		"nodeId", node.NodeId,
		"ipAddress", node.IpAddress,
		"portNumber", node.PortNumber,
		"instanceStatus", node.InstanceStatus,
		"healthyStatus", node.HealthyStatus)

	// 自动通知订阅者
	m.eventNotifier.NotifyServiceChange(ctx, node.TenantId, node.NamespaceId, node.GroupName, node.ServiceName, "NODE_UPDATED")

	return nil
}

// DeleteNodeFromCache 从缓存删除节点（自动通知订阅者）
//
// 处理流程：
//  1. 从缓存获取节点信息（用于获取服务信息）
//  2. 从缓存删除节点
//  3. 自动触发 NODE_REMOVED 事件通知所有订阅者
//
// 使用场景：
//   - 物理删除节点（从数据库删除）
//   - 节点注销后清理缓存
//
// 参数:
//   - tenantId: 租户ID
//   - nodeId: 节点ID
func (m *ServiceCenterManager) DeleteNodeFromCache(ctx context.Context, tenantId, nodeId string) error {
	if tenantId == "" || nodeId == "" {
		return fmt.Errorf("tenantId和nodeId不能为空")
	}

	// 从缓存获取节点信息（需要获取服务信息用于通知）
	globalCache := cache.GetGlobalCache()
	currentNode, found := globalCache.GetNode(ctx, tenantId, nodeId)
	if !found || currentNode == nil {
		return fmt.Errorf("节点不存在: nodeId=%s", nodeId)
	}

	// 保存服务信息用于后续通知
	namespaceId := currentNode.NamespaceId
	groupName := currentNode.GroupName
	serviceName := currentNode.ServiceName

	// 从缓存删除节点
	globalCache.RemoveNode(ctx, tenantId, namespaceId, groupName, serviceName, nodeId)

	logger.Info("节点已从缓存删除",
		"tenantId", tenantId,
		"nodeId", nodeId,
		"namespaceId", namespaceId,
		"groupName", groupName,
		"serviceName", serviceName)

	// 自动通知订阅者（节点删除事件）
	m.eventNotifier.NotifyServiceChange(ctx, tenantId, namespaceId, groupName, serviceName, "NODE_REMOVED")

	return nil
}

// OfflineNodeInCache 下线节点（设置状态为DOWN，自动通知订阅者）
//
// 处理流程：
//  1. 从缓存获取节点信息
//  2. 更新节点状态为DOWN，健康状态为UNHEALTHY
//  3. 更新缓存
//  4. 自动触发 NODE_UPDATED 事件通知所有订阅者
//
// 使用场景：
//   - 手动下线节点
//   - 节点故障后需要标记为下线
//
// 参数:
//   - tenantId: 租户ID
//   - nodeId: 节点ID
//   - operatorId: 操作人ID（可选，用于记录操作人）
func (m *ServiceCenterManager) OfflineNodeInCache(ctx context.Context, tenantId, nodeId, operatorId string) error {
	if tenantId == "" || nodeId == "" {
		return fmt.Errorf("tenantId和nodeId不能为空")
	}

	// 从缓存获取节点信息
	globalCache := cache.GetGlobalCache()
	currentNode, found := globalCache.GetNode(ctx, tenantId, nodeId)
	if !found || currentNode == nil {
		return fmt.Errorf("节点不存在: nodeId=%s", nodeId)
	}

	// 更新节点状态为 DOWN
	currentNode.InstanceStatus = types.NodeStatusDown
	currentNode.HealthyStatus = types.HealthyStatusUnhealthy
	currentNode.EditTime = time.Now()
	if operatorId != "" {
		currentNode.EditWho = operatorId
	}

	// 更新缓存
	globalCache.UpdateNode(ctx, currentNode)

	logger.Info("节点已下线",
		"tenantId", tenantId,
		"nodeId", nodeId,
		"namespaceId", currentNode.NamespaceId,
		"groupName", currentNode.GroupName,
		"serviceName", currentNode.ServiceName)

	// 自动通知订阅者（节点状态更新事件）
	m.eventNotifier.NotifyServiceChange(ctx, currentNode.TenantId, currentNode.NamespaceId, currentNode.GroupName, currentNode.ServiceName, "NODE_UPDATED")

	return nil
}

// Close 关闭管理器，释放所有资源
func (m *ServiceCenterManager) Close() error {
	ctx := context.Background()
	// 停止所有健康检查器
	m.hcMu.Lock()
	for instanceName := range m.healthCheckers {
		m.stopHealthCheckerLocked(instanceName)
	}
	m.healthCheckers = make(map[string]*HealthChecker)
	m.hcMu.Unlock()

	// 停止所有实例
	var errors []string
	m.ForEachInstance(func(instanceName string, srv *server.Server) error {
		if srv.IsRunning() {
			if err := m.StopInstance(ctx, instanceName); err != nil {
				errors = append(errors, fmt.Sprintf("%s: %v", instanceName, err))
				logger.Error("停止服务中心实例失败", err, "instanceName", instanceName)
			}
		}
		return nil
	})

	if len(errors) > 0 {
		logger.Warn("部分实例停止失败", "errors", errors)
	}

	// 注意：缓存是全局单例，不需要在此处关闭

	logger.Info("服务中心管理器已关闭")
	return nil
}

// ========== 健康检查器管理 ==========

// createHealthChecker 为指定实例创建健康检查器
func (m *ServiceCenterManager) createHealthChecker(instanceName, tenantId string) {
	m.hcMu.Lock()
	defer m.hcMu.Unlock()

	// 检查是否已存在
	if _, exists := m.healthCheckers[instanceName]; exists {
		logger.Warn("健康检查器已存在", "instanceName", instanceName)
		return
	}

	// 创建健康检查器（从实例配置获取间隔和超时时间）
	hc := NewHealthChecker(instanceName, tenantId, m)

	m.healthCheckers[instanceName] = hc

	// 启动健康检查器
	hc.Start()

	logger.Info("健康检查器已创建并启动", "instanceName", instanceName, "interval", hc.GetInterval())
}

// stopHealthChecker 停止指定实例的健康检查器
func (m *ServiceCenterManager) stopHealthChecker(instanceName string) {
	m.hcMu.Lock()
	defer m.hcMu.Unlock()
	m.stopHealthCheckerLocked(instanceName)
}

// stopHealthCheckerLocked 停止指定实例的健康检查器（已持有锁）
func (m *ServiceCenterManager) stopHealthCheckerLocked(instanceName string) {
	hc, exists := m.healthCheckers[instanceName]
	if !exists {
		return
	}

	hc.Stop()
	delete(m.healthCheckers, instanceName)

	logger.Info("健康检查器已停止", "instanceName", instanceName)
}
