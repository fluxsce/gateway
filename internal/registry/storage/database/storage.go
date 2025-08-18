package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gateway/internal/registry/core"
	"gateway/pkg/database"
	"gateway/pkg/utils/random"
)

// Storage 数据库存储实现
type Storage struct {
	db database.Database
}

// NewStorage 创建数据库存储实例
func NewStorage(db database.Database) *Storage {
	return &Storage{
		db: db,
	}
}

// ================== 服务分组管理 ==================

// SaveServiceGroup 保存服务分组
func (s *Storage) SaveServiceGroup(ctx context.Context, group *core.ServiceGroup) error {
	// 检查是否存在
	existing, err := s.GetServiceGroup(ctx, group.TenantId, group.GroupName)
	if err != nil && err != core.ErrGroupNotFound {
		return fmt.Errorf("check existing group failed: %w", err)
	}

	if existing != nil {
		// 更新
		return s.updateServiceGroup(ctx, group)
	} else {
		// 插入
		return s.insertServiceGroup(ctx, group)
	}
}

// insertServiceGroup 插入服务分组
func (s *Storage) insertServiceGroup(ctx context.Context, group *core.ServiceGroup) error {
	if group.ServiceGroupId == "" {
		group.ServiceGroupId = random.Generate32BitRandomString()
	}
	if group.OprSeqFlag == "" {
		group.OprSeqFlag = random.Generate32BitRandomString()
	}

	query := `INSERT INTO HUB_REGISTRY_SERVICE_GROUP (
		serviceGroupId, tenantId, groupName, groupDescription, groupType,
		ownerUserId, adminUserIds, readUserIds, accessControlEnabled,
		defaultProtocolType, defaultLoadBalanceStrategy, defaultHealthCheckUrl, defaultHealthCheckIntervalSeconds,
		addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.Exec(ctx, query, []interface{}{
		group.ServiceGroupId, group.TenantId, group.GroupName, group.GroupDescription, group.GroupType,
		group.OwnerUserId, group.AdminUserIds, group.ReadUserIds, group.AccessControlEnabled,
		group.DefaultProtocolType, group.DefaultLoadBalanceStrategy, group.DefaultHealthCheckUrl, group.DefaultHealthCheckIntervalSeconds,
		group.AddTime, group.AddWho, group.EditTime, group.EditWho, group.OprSeqFlag, group.CurrentVersion, group.ActiveFlag, group.NoteText, group.ExtProperty,
	}, true)

	return err
}

// updateServiceGroup 更新服务分组
func (s *Storage) updateServiceGroup(ctx context.Context, group *core.ServiceGroup) error {
	group.EditTime = time.Now()
	group.CurrentVersion++

	query := `UPDATE HUB_REGISTRY_SERVICE_GROUP SET
		groupDescription = ?, groupType = ?, ownerUserId = ?, adminUserIds = ?, readUserIds = ?, accessControlEnabled = ?,
		defaultProtocolType = ?, defaultLoadBalanceStrategy = ?, defaultHealthCheckUrl = ?, defaultHealthCheckIntervalSeconds = ?,
		editTime = ?, editWho = ?, currentVersion = ?, activeFlag = ?, noteText = ?, extProperty = ?
	WHERE tenantId = ? AND groupName = ?`

	_, err := s.db.Exec(ctx, query, []interface{}{
		group.GroupDescription, group.GroupType, group.OwnerUserId, group.AdminUserIds, group.ReadUserIds, group.AccessControlEnabled,
		group.DefaultProtocolType, group.DefaultLoadBalanceStrategy, group.DefaultHealthCheckUrl, group.DefaultHealthCheckIntervalSeconds,
		group.EditTime, group.EditWho, group.CurrentVersion, group.ActiveFlag, group.NoteText, group.ExtProperty,
		group.TenantId, group.GroupName,
	}, true)

	return err
}

// GetServiceGroup 获取服务分组
func (s *Storage) GetServiceGroup(ctx context.Context, tenantId, groupName string) (*core.ServiceGroup, error) {
	query := `SELECT serviceGroupId, tenantId, groupName, groupDescription, groupType,
		ownerUserId, adminUserIds, readUserIds, accessControlEnabled,
		defaultProtocolType, defaultLoadBalanceStrategy, defaultHealthCheckUrl, defaultHealthCheckIntervalSeconds,
		addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty
	FROM HUB_REGISTRY_SERVICE_GROUP WHERE tenantId = ? AND groupName = ? AND activeFlag = 'Y'`

	group := &core.ServiceGroup{}
	err := s.db.QueryOne(ctx, group, query, []interface{}{tenantId, groupName}, true)
	if err == sql.ErrNoRows {
		return nil, core.ErrGroupNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query service group failed: %w", err)
	}

	return group, nil
}

// DeleteServiceGroup 删除服务分组
func (s *Storage) DeleteServiceGroup(ctx context.Context, tenantId, groupName string) error {
	query := `UPDATE HUB_REGISTRY_SERVICE_GROUP SET activeFlag = 'N', editTime = ? WHERE tenantId = ? AND groupName = ?`
	_, err := s.db.Exec(ctx, query, []interface{}{time.Now(), tenantId, groupName}, true)
	return err
}

// ListServiceGroups 列出服务分组
func (s *Storage) ListServiceGroups(ctx context.Context, tenantId string) ([]*core.ServiceGroup, error) {
	query := `SELECT serviceGroupId, tenantId, groupName, groupDescription, groupType,
		ownerUserId, adminUserIds, readUserIds, accessControlEnabled,
		defaultProtocolType, defaultLoadBalanceStrategy, defaultHealthCheckUrl, defaultHealthCheckIntervalSeconds,
		addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty
	FROM HUB_REGISTRY_SERVICE_GROUP WHERE tenantId = ? AND activeFlag = 'Y' ORDER BY groupName`

	var groups []*core.ServiceGroup
	err := s.db.Query(ctx, &groups, query, []interface{}{tenantId}, true)
	if err != nil {
		return nil, fmt.Errorf("query service groups failed: %w", err)
	}

	return groups, nil
}

// ================== 服务管理 ==================

// SaveService 保存服务
func (s *Storage) SaveService(ctx context.Context, service *core.Service) error {
	// 检查是否存在
	existing, err := s.GetService(ctx, service.TenantId, service.ServiceName)
	if err != nil && err != core.ErrServiceNotFound {
		return fmt.Errorf("check existing service failed: %w", err)
	}

	if existing != nil {
		// 更新
		return s.updateService(ctx, service)
	} else {
		// 插入
		return s.insertService(ctx, service)
	}
}

// insertService 插入服务
func (s *Storage) insertService(ctx context.Context, service *core.Service) error {
	if service.OprSeqFlag == "" {
		service.OprSeqFlag = random.Generate32BitRandomString()
	}

	query := `INSERT INTO HUB_REGISTRY_SERVICE (
		tenantId, serviceName, groupName, serviceDescription, protocolType, contextPath, loadBalanceStrategy,
		healthCheckUrl, healthCheckIntervalSeconds, healthCheckTimeoutSeconds,
		metadataJson, tagsJson,
		addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.Exec(ctx, query, []interface{}{
		service.TenantId, service.ServiceName, service.GroupName, service.ServiceDescription, service.ProtocolType, service.ContextPath, service.LoadBalanceStrategy,
		service.HealthCheckUrl, service.HealthCheckIntervalSeconds, service.HealthCheckTimeoutSeconds,
		service.MetadataJson, service.TagsJson,
		service.AddTime, service.AddWho, service.EditTime, service.EditWho, service.OprSeqFlag, service.CurrentVersion, service.ActiveFlag, service.NoteText, service.ExtProperty,
	}, true)

	return err
}

// updateService 更新服务
func (s *Storage) updateService(ctx context.Context, service *core.Service) error {
	service.EditTime = time.Now()
	service.CurrentVersion++

	query := `UPDATE HUB_REGISTRY_SERVICE SET
		groupName = ?, serviceDescription = ?, protocolType = ?, contextPath = ?, loadBalanceStrategy = ?,
		healthCheckUrl = ?, healthCheckIntervalSeconds = ?, healthCheckTimeoutSeconds = ?,
		metadataJson = ?, tagsJson = ?,
		editTime = ?, editWho = ?, currentVersion = ?, activeFlag = ?, noteText = ?, extProperty = ?
	WHERE tenantId = ? AND serviceName = ?`

	_, err := s.db.Exec(ctx, query, []interface{}{
		service.GroupName, service.ServiceDescription, service.ProtocolType, service.ContextPath, service.LoadBalanceStrategy,
		service.HealthCheckUrl, service.HealthCheckIntervalSeconds, service.HealthCheckTimeoutSeconds,
		service.MetadataJson, service.TagsJson,
		service.EditTime, service.EditWho, service.CurrentVersion, service.ActiveFlag, service.NoteText, service.ExtProperty,
		service.TenantId, service.ServiceName,
	}, true)

	return err
}

// GetService 获取服务
func (s *Storage) GetService(ctx context.Context, tenantId, serviceName string) (*core.Service, error) {
	query := `SELECT tenantId, serviceName, groupName, serviceDescription, protocolType, contextPath, loadBalanceStrategy,
		healthCheckUrl, healthCheckIntervalSeconds, healthCheckTimeoutSeconds,
		metadataJson, tagsJson,
		addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty
	FROM HUB_REGISTRY_SERVICE WHERE tenantId = ? AND serviceName = ? AND activeFlag = 'Y'`

	service := &core.Service{}
	err := s.db.QueryOne(ctx, service, query, []interface{}{tenantId, serviceName}, true)
	if err == sql.ErrNoRows {
		return nil, core.ErrServiceNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query service failed: %w", err)
	}

	return service, nil
}

// DeleteService 删除服务
func (s *Storage) DeleteService(ctx context.Context, tenantId, serviceName string) error {
	query := `UPDATE HUB_REGISTRY_SERVICE SET activeFlag = 'N', editTime = ? WHERE tenantId = ? AND serviceName = ?`
	_, err := s.db.Exec(ctx, query, []interface{}{time.Now(), tenantId, serviceName}, true)
	return err
}

// ListServices 列出服务
func (s *Storage) ListServices(ctx context.Context, tenantId, groupName string) ([]*core.Service, error) {
	var query string
	var args []interface{}

	if groupName != "" {
		query = `SELECT tenantId, serviceName, groupName, serviceDescription, protocolType, contextPath, loadBalanceStrategy,
			healthCheckUrl, healthCheckIntervalSeconds, healthCheckTimeoutSeconds,
			metadataJson, tagsJson,
			addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty
		FROM HUB_REGISTRY_SERVICE WHERE tenantId = ? AND groupName = ? AND activeFlag = 'Y' ORDER BY serviceName`
		args = []interface{}{tenantId, groupName}
	} else {
		query = `SELECT tenantId, serviceName, groupName, serviceDescription, protocolType, contextPath, loadBalanceStrategy,
			healthCheckUrl, healthCheckIntervalSeconds, healthCheckTimeoutSeconds,
			metadataJson, tagsJson,
			addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty
		FROM HUB_REGISTRY_SERVICE WHERE tenantId = ? AND activeFlag = 'Y' ORDER BY serviceName`
		args = []interface{}{tenantId}
	}

	var services []*core.Service
	err := s.db.Query(ctx, &services, query, args, true)
	if err != nil {
		return nil, fmt.Errorf("query services failed: %w", err)
	}

	return services, nil
}

// ================== 服务实例管理 ==================

// SaveInstance 保存服务实例
func (s *Storage) SaveInstance(ctx context.Context, instance *core.ServiceInstance) error {
	// 检查是否存在
	existing, err := s.GetInstance(ctx, instance.TenantId, instance.ServiceInstanceId)
	if err != nil && err != core.ErrInstanceNotFound {
		return fmt.Errorf("check existing instance failed: %w", err)
	}

	if existing != nil {
		// 更新
		return s.updateInstance(ctx, instance)
	} else {
		// 插入
		return s.insertInstance(ctx, instance)
	}
}

// insertInstance 插入服务实例
func (s *Storage) insertInstance(ctx context.Context, instance *core.ServiceInstance) error {
	if instance.ServiceInstanceId == "" {
		instance.ServiceInstanceId = random.Generate32BitRandomString()
	}
	if instance.OprSeqFlag == "" {
		instance.OprSeqFlag = random.Generate32BitRandomString()
	}

	query := `INSERT INTO HUB_REGISTRY_SERVICE_INSTANCE (
		serviceInstanceId, tenantId, serviceName, groupName,
		hostAddress, portNumber, contextPath, instanceStatus, healthStatus, weightValue,
		clientId, clientVersion, clientType, metadataJson, tagsJson,
		registerTime, lastHeartbeatTime, lastHealthCheckTime,
		addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.Exec(ctx, query, []interface{}{
		instance.ServiceInstanceId, instance.TenantId, instance.ServiceName, instance.GroupName,
		instance.HostAddress, instance.PortNumber, instance.ContextPath, instance.InstanceStatus, instance.HealthStatus, instance.WeightValue,
		instance.ClientId, instance.ClientVersion, instance.ClientType, instance.MetadataJson, instance.TagsJson,
		instance.RegisterTime, instance.LastHeartbeatTime, instance.LastHealthCheckTime,
		instance.AddTime, instance.AddWho, instance.EditTime, instance.EditWho, instance.OprSeqFlag, instance.CurrentVersion, instance.ActiveFlag, instance.NoteText, instance.ExtProperty,
	}, true)

	return err
}

// updateInstance 更新服务实例
func (s *Storage) updateInstance(ctx context.Context, instance *core.ServiceInstance) error {
	instance.EditTime = time.Now()
	instance.CurrentVersion++

	query := `UPDATE HUB_REGISTRY_SERVICE_INSTANCE SET
		serviceName = ?, groupName = ?, hostAddress = ?, portNumber = ?, contextPath = ?, instanceStatus = ?, healthStatus = ?, weightValue = ?,
		clientId = ?, clientVersion = ?, clientType = ?, metadataJson = ?, tagsJson = ?,
		lastHeartbeatTime = ?, lastHealthCheckTime = ?,
		editTime = ?, editWho = ?, currentVersion = ?, activeFlag = ?, noteText = ?, extProperty = ?
	WHERE tenantId = ? AND serviceInstanceId = ?`

	_, err := s.db.Exec(ctx, query, []interface{}{
		instance.ServiceName, instance.GroupName, instance.HostAddress, instance.PortNumber, instance.ContextPath, instance.InstanceStatus, instance.HealthStatus, instance.WeightValue,
		instance.ClientId, instance.ClientVersion, instance.ClientType, instance.MetadataJson, instance.TagsJson,
		instance.LastHeartbeatTime, instance.LastHealthCheckTime,
		instance.EditTime, instance.EditWho, instance.CurrentVersion, instance.ActiveFlag, instance.NoteText, instance.ExtProperty,
		instance.TenantId, instance.ServiceInstanceId,
	}, true)

	return err
}

// GetInstance 获取服务实例
func (s *Storage) GetInstance(ctx context.Context, tenantId, instanceId string) (*core.ServiceInstance, error) {
	query := `SELECT serviceInstanceId, tenantId, serviceName, groupName,
		hostAddress, portNumber, contextPath, instanceStatus, healthStatus, weightValue,
		clientId, clientVersion, clientType, metadataJson, tagsJson,
		registerTime, lastHeartbeatTime, lastHealthCheckTime,
		addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty
	FROM HUB_REGISTRY_SERVICE_INSTANCE WHERE tenantId = ? AND serviceInstanceId = ? AND activeFlag = 'Y'`

	instance := &core.ServiceInstance{}
	err := s.db.QueryOne(ctx, instance, query, []interface{}{tenantId, instanceId}, true)
	if err == sql.ErrNoRows {
		return nil, core.ErrInstanceNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query service instance failed: %w", err)
	}

	return instance, nil
}

// DeleteInstance 删除服务实例
func (s *Storage) DeleteInstance(ctx context.Context, tenantId, instanceId string) error {
	query := `UPDATE HUB_REGISTRY_SERVICE_INSTANCE SET activeFlag = 'N', editTime = ? WHERE tenantId = ? AND serviceInstanceId = ?`
	_, err := s.db.Exec(ctx, query, []interface{}{time.Now(), tenantId, instanceId}, true)
	return err
}

// ListInstances 列出服务实例
func (s *Storage) ListInstances(ctx context.Context, tenantId, serviceName, groupName string) ([]*core.ServiceInstance, error) {
	var query string
	var args []interface{}

	if serviceName != "" && groupName != "" {
		query = `SELECT serviceInstanceId, tenantId, serviceName, groupName,
			hostAddress, portNumber, contextPath, instanceStatus, healthStatus, weightValue,
			clientId, clientVersion, clientType, metadataJson, tagsJson,
			registerTime, lastHeartbeatTime, lastHealthCheckTime,
			addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty
		FROM HUB_REGISTRY_SERVICE_INSTANCE WHERE tenantId = ? AND serviceName = ? AND groupName = ? AND activeFlag = 'Y' ORDER BY hostAddress, portNumber`
		args = []interface{}{tenantId, serviceName, groupName}
	} else if serviceName != "" {
		query = `SELECT serviceInstanceId, tenantId, serviceName, groupName,
			hostAddress, portNumber, contextPath, instanceStatus, healthStatus, weightValue,
			clientId, clientVersion, clientType, metadataJson, tagsJson,
			registerTime, lastHeartbeatTime, lastHealthCheckTime,
			addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty
		FROM HUB_REGISTRY_SERVICE_INSTANCE WHERE tenantId = ? AND serviceName = ? AND activeFlag = 'Y' ORDER BY hostAddress, portNumber`
		args = []interface{}{tenantId, serviceName}
	} else {
		query = `SELECT serviceInstanceId, tenantId, serviceName, groupName,
			hostAddress, portNumber, contextPath, instanceStatus, healthStatus, weightValue,
			clientId, clientVersion, clientType, metadataJson, tagsJson,
			registerTime, lastHeartbeatTime, lastHealthCheckTime,
			addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty
		FROM HUB_REGISTRY_SERVICE_INSTANCE WHERE tenantId = ? AND activeFlag = 'Y' ORDER BY serviceName, hostAddress, portNumber`
		args = []interface{}{tenantId}
	}

	var instances []*core.ServiceInstance
	err := s.db.Query(ctx, &instances, query, args, true)
	if err != nil {
		return nil, fmt.Errorf("query service instances failed: %w", err)
	}

	return instances, nil
}

// ListAllInstances 列出所有服务实例
func (s *Storage) ListAllInstances(ctx context.Context, tenantId string) ([]*core.ServiceInstance, error) {
	return s.ListInstances(ctx, tenantId, "", "")
}

// ================== 实例状态管理 ==================

// UpdateHeartbeat 更新心跳
func (s *Storage) UpdateHeartbeat(ctx context.Context, tenantId, instanceId string) error {
	now := time.Now()
	query := `UPDATE HUB_REGISTRY_SERVICE_INSTANCE SET 
		lastHeartbeatTime = ?, editTime = ?, currentVersion = currentVersion + 1 
	WHERE tenantId = ? AND serviceInstanceId = ? AND activeFlag = 'Y'`

	_, err := s.db.Exec(ctx, query, []interface{}{now, now, tenantId, instanceId}, true)
	return err
}

// UpdateInstanceHealth 更新实例健康状态
func (s *Storage) UpdateInstanceHealth(ctx context.Context, tenantId, instanceId string, healthStatus string) error {
	now := time.Now()
	query := `UPDATE HUB_REGISTRY_SERVICE_INSTANCE SET 
		healthStatus = ?, lastHealthCheckTime = ?, editTime = ?, currentVersion = currentVersion + 1 
	WHERE tenantId = ? AND serviceInstanceId = ? AND activeFlag = 'Y'`

	_, err := s.db.Exec(ctx, query, []interface{}{healthStatus, now, now, tenantId, instanceId}, true)
	return err
}

// UpdateInstanceStatus 更新实例状态
func (s *Storage) UpdateInstanceStatus(ctx context.Context, tenantId, instanceId string, instanceStatus string) error {
	now := time.Now()
	query := `UPDATE HUB_REGISTRY_SERVICE_INSTANCE SET 
		instanceStatus = ?, editTime = ?, currentVersion = currentVersion + 1 
	WHERE tenantId = ? AND serviceInstanceId = ? AND activeFlag = 'Y'`

	_, err := s.db.Exec(ctx, query, []interface{}{instanceStatus, now, tenantId, instanceId}, true)
	return err
}

// ================== 服务发现 ==================

// GetServiceNames 获取服务名称列表
func (s *Storage) GetServiceNames(ctx context.Context, tenantId, groupName string) ([]string, error) {
	var query string
	var args []interface{}

	if groupName != "" {
		query = `SELECT DISTINCT serviceName FROM HUB_REGISTRY_SERVICE WHERE tenantId = ? AND groupName = ? AND activeFlag = 'Y' ORDER BY serviceName`
		args = []interface{}{tenantId, groupName}
	} else {
		query = `SELECT DISTINCT serviceName FROM HUB_REGISTRY_SERVICE WHERE tenantId = ? AND activeFlag = 'Y' ORDER BY serviceName`
		args = []interface{}{tenantId}
	}

	var names []string
	err := s.db.Query(ctx, &names, query, args, true)
	if err != nil {
		return nil, fmt.Errorf("query service names failed: %w", err)
	}

	return names, nil
}

// GetInstances 获取实例（带过滤器）
func (s *Storage) GetInstances(ctx context.Context, tenantId, serviceName, groupName string, filters ...core.InstanceFilter) ([]*core.ServiceInstance, error) {
	instances, err := s.ListInstances(ctx, tenantId, serviceName, groupName)
	if err != nil {
		return nil, err
	}

	// 应用过滤器
	return core.ApplyInstanceFilters(instances, filters...), nil
}

// ================== 事件日志 ==================

// LogEvent 记录事件
func (s *Storage) LogEvent(ctx context.Context, event *core.ServiceEvent) error {
	if event.OprSeqFlag == "" {
		event.OprSeqFlag = random.Generate32BitRandomString()
	}

	query := `INSERT INTO HUB_REGISTRY_SERVICE_EVENT (
		tenantId, groupName, serviceName, hostAddress, portNumber, eventType, eventSource,
		eventDataJson, eventMessage, eventTime,
		addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.Exec(ctx, query, []interface{}{
		event.TenantId, event.GroupName, event.ServiceName, event.HostAddress, event.PortNumber, event.EventType, event.EventSource,
		event.EventDataJson, event.EventMessage, event.EventTime,
		event.AddTime, event.AddWho, event.EditTime, event.EditWho, event.OprSeqFlag, event.CurrentVersion, event.ActiveFlag, event.NoteText, event.ExtProperty,
	}, true)

	return err
}

// GetEvents 获取事件（带过滤器）
func (s *Storage) GetEvents(ctx context.Context, tenantId string, filters ...core.EventFilter) ([]*core.ServiceEvent, error) {
	query := `SELECT serviceEventId, tenantId, groupName, serviceName, hostAddress, portNumber, eventType, eventSource,
		eventDataJson, eventMessage, eventTime,
		addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty
	FROM HUB_REGISTRY_SERVICE_EVENT WHERE tenantId = ? AND activeFlag = 'Y' ORDER BY eventTime DESC LIMIT 1000`

	var events []*core.ServiceEvent
	err := s.db.Query(ctx, &events, query, []interface{}{tenantId}, true)
	if err != nil {
		return nil, fmt.Errorf("query service events failed: %w", err)
	}

	// 应用过滤器
	return core.ApplyEventFilters(events, filters...), nil
}

// ================== 健康检查 ==================

// GetUnhealthyInstances 获取不健康的实例
func (s *Storage) GetUnhealthyInstances(ctx context.Context, tenantId string, timeout time.Duration) ([]*core.ServiceInstance, error) {
	cutoffTime := time.Now().Add(-timeout)

	query := `SELECT serviceInstanceId, tenantId, serviceName, groupName,
		hostAddress, portNumber, contextPath, instanceStatus, healthStatus, weightValue,
		clientId, clientVersion, clientType, metadataJson, tagsJson,
		registerTime, lastHeartbeatTime, lastHealthCheckTime,
		addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty
	FROM HUB_REGISTRY_SERVICE_INSTANCE 
	WHERE tenantId = ? AND activeFlag = 'Y' AND (
		healthStatus != 'HEALTHY' OR 
		lastHeartbeatTime IS NULL OR 
		lastHeartbeatTime < ?
	) ORDER BY serviceName, hostAddress, portNumber`

	var instances []*core.ServiceInstance
	err := s.db.Query(ctx, &instances, query, []interface{}{tenantId, cutoffTime}, true)
	if err != nil {
		return nil, fmt.Errorf("query unhealthy instances failed: %w", err)
	}

	return instances, nil
}

// ================== 统计信息 ==================

// GetStats 获取统计信息
func (s *Storage) GetStats(ctx context.Context, tenantId string) (*core.StorageStats, error) {
	stats := &core.StorageStats{
		TenantId:       tenantId,
		LastUpdateTime: time.Now(),
	}

	// 获取分组数量
	err := s.db.QueryOne(ctx, &stats.GroupCount, "SELECT COUNT(*) FROM HUB_REGISTRY_SERVICE_GROUP WHERE tenantId = ? AND activeFlag = 'Y'", []interface{}{tenantId}, true)
	if err != nil {
		return nil, fmt.Errorf("get group count failed: %w", err)
	}

	// 获取服务数量
	err = s.db.QueryOne(ctx, &stats.ServiceCount, "SELECT COUNT(*) FROM HUB_REGISTRY_SERVICE WHERE tenantId = ? AND activeFlag = 'Y'", []interface{}{tenantId}, true)
	if err != nil {
		return nil, fmt.Errorf("get service count failed: %w", err)
	}

	// 获取实例数量
	err = s.db.QueryOne(ctx, &stats.InstanceCount, "SELECT COUNT(*) FROM HUB_REGISTRY_SERVICE_INSTANCE WHERE tenantId = ? AND activeFlag = 'Y'", []interface{}{tenantId}, true)
	if err != nil {
		return nil, fmt.Errorf("get instance count failed: %w", err)
	}

	// 获取事件数量
	err = s.db.QueryOne(ctx, &stats.EventCount, "SELECT COUNT(*) FROM HUB_REGISTRY_SERVICE_EVENT WHERE tenantId = ? AND activeFlag = 'Y'", []interface{}{tenantId}, true)
	if err != nil {
		return nil, fmt.Errorf("get event count failed: %w", err)
	}

	return stats, nil
}
