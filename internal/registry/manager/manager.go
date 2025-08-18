package manager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gateway/internal/registry/config"
	"gateway/internal/registry/core"
	"gateway/internal/registry/event"
	"gateway/internal/registry/health"
	"gateway/internal/registry/service"
	registrydb "gateway/internal/registry/storage/database"
	"gateway/pkg/database"
)

// Manager 注册中心管理器实现
type Manager struct {
	config          *config.Config
	db              database.Database
	storage         core.Storage
	externalStorage core.ExternalStorage
	registry        core.Registry
	eventPublisher  core.EventPublisher
	healthChecker   core.HealthChecker
	running         bool
	startTime       time.Time
	mutex           sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
}

// NewManager 创建注册中心管理器
func NewManager(cfg *config.Config, db database.Database) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	return &Manager{
		config: cfg,
		db:     db,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Initialize 初始化
func (m *Manager) Initialize() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.running {
		return fmt.Errorf("manager already initialized")
	}

	// 初始化存储层
	m.storage = registrydb.NewStorage(m.db)
	m.externalStorage = registrydb.NewExternalStorage(m.db)

	// 初始化事件发布器
	eventConfig := &event.EventConfig{
		BufferSize:        m.config.Event.BufferSize,
		WorkerCount:       m.config.Event.WorkerCount,
		BatchSize:         m.config.Event.BatchSize,
		BatchTimeout:      m.config.Event.BatchTimeout,
		MaxSubscribers:    m.config.Event.MaxSubscribers,
		SubscriberTimeout: m.config.Event.SubscriberTimeout,
		EnablePersistence: m.config.Event.EnablePersistence,
		RetentionPeriod:   m.config.Event.RetentionPeriod,
		CleanupInterval:   m.config.Event.CleanupInterval,
	}
	m.eventPublisher = event.NewPublisher(eventConfig, m.storage)

	// 初始化健康检查器
	healthConfig := &health.HealthCheckConfig{
		Enabled:          m.config.HealthCheck.Enabled,
		Interval:         m.config.HealthCheck.Interval,
		Timeout:          m.config.HealthCheck.Timeout,
		MaxRetries:       m.config.HealthCheck.MaxRetries,
		RetryInterval:    m.config.HealthCheck.RetryInterval,
		ConcurrentChecks: m.config.HealthCheck.ConcurrentChecks,
		FailureThreshold: m.config.HealthCheck.FailureThreshold,
		SuccessThreshold: m.config.HealthCheck.SuccessThreshold,
		DefaultPath:      m.config.HealthCheck.DefaultPath,
		EnableTCPCheck:   m.config.HealthCheck.EnableTCPCheck,
		EnableHTTPCheck:  m.config.HealthCheck.EnableHTTPCheck,
	}
	m.healthChecker = health.NewChecker(healthConfig, m.storage, m.eventPublisher)

	// 初始化注册中心服务
	m.registry = service.NewRegistryService(m.storage, m.eventPublisher)

	return nil
}

// Start 启动
func (m *Manager) Start() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.running {
		return nil
	}

	// 启动事件发布器
	if err := m.eventPublisher.Start(); err != nil {
		return fmt.Errorf("start event publisher failed: %w", err)
	}

	// 启动注册中心服务
	if err := m.registry.Start(); err != nil {
		return fmt.Errorf("start registry service failed: %w", err)
	}

	// 启动健康检查器
	if err := m.healthChecker.Start(m.ctx); err != nil {
		return fmt.Errorf("start health checker failed: %w", err)
	}

	// 加载现有实例到健康检查器
	go m.loadInstancesForHealthCheck()

	// 启动监控协程
	go m.monitorServices()

	m.running = true
	m.startTime = time.Now()

	return nil
}

// Stop 停止
func (m *Manager) Stop() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.running {
		return nil
	}

	// 取消上下文
	m.cancel()

	// 停止健康检查器
	if err := m.healthChecker.Stop(); err != nil {
		fmt.Printf("stop health checker failed: %v\n", err)
	}

	// 停止注册中心服务
	if err := m.registry.Close(); err != nil {
		fmt.Printf("stop registry service failed: %v\n", err)
	}

	// 停止事件发布器
	if err := m.eventPublisher.Close(); err != nil {
		fmt.Printf("stop event publisher failed: %v\n", err)
	}

	m.running = false

	return nil
}

// GetRegistry 获取注册中心实例
func (m *Manager) GetRegistry() core.Registry {
	return m.registry
}

// GetStorage 获取存储实例
func (m *Manager) GetStorage() core.Storage {
	return m.storage
}

// GetExternalStorage 获取外部存储实例
func (m *Manager) GetExternalStorage() core.ExternalStorage {
	return m.externalStorage
}

// GetEventPublisher 获取事件发布器
func (m *Manager) GetEventPublisher() core.EventPublisher {
	return m.eventPublisher
}

// GetHealthChecker 获取健康检查器
func (m *Manager) GetHealthChecker() core.HealthChecker {
	return m.healthChecker
}

// GetStats 获取统计信息
func (m *Manager) GetStats() *core.ManagerStats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := &core.ManagerStats{
		Running:   m.running,
		Mode:      "standalone", // 默认为独立模式
		StartTime: m.startTime,
	}

	// 获取服务统计信息
	if m.running && m.storage != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// 这里需要一个默认的tenantId，实际使用中应该从配置或请求中获取
		defaultTenantId := "default"

		if storageStats, err := m.storage.GetStats(ctx, defaultTenantId); err == nil {
			stats.TenantId = storageStats.TenantId
			stats.ServiceCount = storageStats.ServiceCount
			stats.InstanceCount = storageStats.InstanceCount
		}

		if serviceNames, err := m.storage.GetServiceNames(ctx, defaultTenantId, ""); err == nil {
			stats.Services = serviceNames
		}
	}

	return stats
}

// IsRunning 获取运行状态
func (m *Manager) IsRunning() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.running
}

// ================== 内部方法 ==================

// loadInstancesForHealthCheck 加载实例到健康检查器
func (m *Manager) loadInstancesForHealthCheck() {
	if !m.config.HealthCheck.Enabled {
		return
	}

	// 等待一段时间让系统启动完成
	time.Sleep(5 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 这里需要一个默认的tenantId，实际使用中应该从配置中获取
	defaultTenantId := "default"

	if err := m.healthChecker.LoadInstances(ctx, defaultTenantId); err != nil {
		fmt.Printf("load instances for health check failed: %v\n", err)
	}
}

// monitorServices 监控服务
func (m *Manager) monitorServices() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.performMonitoring()
		case <-m.ctx.Done():
			return
		}
	}
}

// performMonitoring 执行监控
func (m *Manager) performMonitoring() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 这里需要一个默认的tenantId，实际使用中应该从配置中获取
	defaultTenantId := "default"

	// 检查不健康的实例
	unhealthyInstances, err := m.storage.GetUnhealthyInstances(ctx, defaultTenantId, 2*m.config.HealthCheck.Interval)
	if err != nil {
		fmt.Printf("get unhealthy instances failed: %v\n", err)
		return
	}

	// 处理不健康的实例
	for _, instance := range unhealthyInstances {
		// 如果实例长时间没有心跳，标记为下线
		if instance.LastHeartbeatTime != nil {
			timeSinceLastHeartbeat := time.Since(*instance.LastHeartbeatTime)
			if timeSinceLastHeartbeat > 5*m.config.HealthCheck.Interval {
				// 更新实例状态为下线
				if err := m.storage.UpdateInstanceStatus(ctx, instance.TenantId, instance.ServiceInstanceId, core.InstanceStatusDown); err != nil {
					fmt.Printf("update instance status failed: %v\n", err)
					continue
				}

				// 发布实例状态变更事件
				event := core.NewServiceEvent(
					instance.TenantId,
					core.EventTypeInstanceStatusChange,
					instance.ServiceName,
					instance.GroupName,
					"manager",
					fmt.Sprintf("Instance %s marked as DOWN due to no heartbeat", instance.ServiceInstanceId),
				)
				event.HostAddress = instance.HostAddress
				event.PortNumber = &instance.PortNumber

				if err := m.eventPublisher.Publish(ctx, event); err != nil {
					fmt.Printf("publish instance status change event failed: %v\n", err)
				}
			}
		}
	}
}

// GetConfig 获取配置
func (m *Manager) GetConfig() *config.Config {
	return m.config
}

// UpdateConfig 更新配置
func (m *Manager) UpdateConfig(newConfig *config.Config) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.running {
		return fmt.Errorf("cannot update config while manager is running")
	}

	m.config = newConfig
	return nil
}

// Restart 重启管理器
func (m *Manager) Restart() error {
	if err := m.Stop(); err != nil {
		return fmt.Errorf("stop manager failed: %w", err)
	}

	// 等待一段时间确保完全停止
	time.Sleep(2 * time.Second)

	if err := m.Start(); err != nil {
		return fmt.Errorf("start manager failed: %w", err)
	}

	return nil
}

// GetHealthStatus 获取健康状态
func (m *Manager) GetHealthStatus() map[string]interface{} {
	status := map[string]interface{}{
		"running":   m.IsRunning(),
		"startTime": m.startTime,
		"uptime":    time.Since(m.startTime).String(),
	}

	if m.running {
		// 添加各组件状态
		if registryService, ok := m.registry.(*service.RegistryService); ok {
			status["registry"] = map[string]interface{}{
				"running": registryService.IsRunning(),
			}
		}

		if eventPublisher, ok := m.eventPublisher.(*event.Publisher); ok {
			status["eventPublisher"] = eventPublisher.GetStats()
		}

		if healthChecker, ok := m.healthChecker.(*health.Checker); ok {
			status["healthChecker"] = map[string]interface{}{
				"running":       healthChecker.IsRunning(),
				"instanceCount": healthChecker.GetInstanceCount(),
				"stats":         healthChecker.GetStats(),
			}
		}
	}

	return status
}

// RegisterInstance 注册实例（便捷方法）
func (m *Manager) RegisterInstance(ctx context.Context, instance *core.ServiceInstance) error {
	if !m.IsRunning() {
		return core.ErrRegistryNotRunning
	}

	// 注册实例
	if err := m.registry.Register(ctx, instance); err != nil {
		return err
	}

	// 添加到健康检查器
	if m.config.HealthCheck.Enabled {
		if err := m.healthChecker.AddInstance(instance); err != nil {
			fmt.Printf("add instance to health checker failed: %v\n", err)
		}
	}

	return nil
}

// DeregisterInstance 注销实例（便捷方法）
func (m *Manager) DeregisterInstance(ctx context.Context, tenantId, instanceId string) error {
	if !m.IsRunning() {
		return core.ErrRegistryNotRunning
	}

	// 从健康检查器移除
	if m.config.HealthCheck.Enabled {
		if err := m.healthChecker.RemoveInstance(instanceId); err != nil {
			fmt.Printf("remove instance from health checker failed: %v\n", err)
		}
	}

	// 注销实例
	return m.registry.Deregister(ctx, tenantId, instanceId)
}

// DiscoverInstances 发现实例（便捷方法）
func (m *Manager) DiscoverInstances(ctx context.Context, tenantId, serviceName, groupName string, filters ...core.InstanceFilter) ([]*core.ServiceInstance, error) {
	if !m.IsRunning() {
		return nil, core.ErrRegistryNotRunning
	}

	// 默认只返回健康的实例
	if len(filters) == 0 {
		filters = append(filters, core.GetHealthyInstancesFilter())
	}

	return m.registry.Discover(ctx, tenantId, serviceName, groupName, filters...)
}

// GetInstanceStats 获取实例统计信息
func (m *Manager) GetInstanceStats(ctx context.Context, tenantId string) (map[string]int, error) {
	if !m.IsRunning() {
		return nil, core.ErrRegistryNotRunning
	}

	instances, err := m.storage.ListAllInstances(ctx, tenantId)
	if err != nil {
		return nil, err
	}

	stats := map[string]int{
		"total":     0,
		"healthy":   0,
		"unhealthy": 0,
		"up":        0,
		"down":      0,
		"active":    0,
		"inactive":  0,
	}

	for _, instance := range instances {
		stats["total"]++

		if instance.IsHealthy() {
			stats["healthy"]++
		} else {
			stats["unhealthy"]++
		}

		if instance.IsUp() {
			stats["up"]++
		} else {
			stats["down"]++
		}

		if instance.IsActive() {
			stats["active"]++
		} else {
			stats["inactive"]++
		}
	}

	return stats, nil
}
