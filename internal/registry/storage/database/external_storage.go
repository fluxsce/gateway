package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"gateway/internal/registry/core"
	"gateway/pkg/database"
	"gateway/pkg/utils/random"
)

// ExternalStorage 外部注册中心存储实现
type ExternalStorage struct {
	db          database.Database
	connections map[string]interface{} // 连接池
	connMutex   sync.RWMutex
}

// NewExternalStorage 创建外部注册中心存储实例
func NewExternalStorage(db database.Database) *ExternalStorage {
	return &ExternalStorage{
		db:          db,
		connections: make(map[string]interface{}),
	}
}

// ================== 配置管理 ==================

// SaveExternalConfig 保存外部配置
func (s *ExternalStorage) SaveExternalConfig(ctx context.Context, config *core.ExternalRegistryConfig) error {
	// 检查是否存在
	existing, err := s.GetExternalConfig(ctx, config.TenantId, config.ExternalConfigId)
	if err != nil && err != core.ErrConfigNotFound {
		return fmt.Errorf("check existing config failed: %w", err)
	}

	if existing != nil {
		// 更新
		return s.updateExternalConfig(ctx, config)
	} else {
		// 插入
		return s.insertExternalConfig(ctx, config)
	}
}

// insertExternalConfig 插入外部配置
func (s *ExternalStorage) insertExternalConfig(ctx context.Context, config *core.ExternalRegistryConfig) error {
	if config.ExternalConfigId == "" {
		config.ExternalConfigId = random.Generate32BitRandomString()
	}
	if config.OprSeqFlag == "" {
		config.OprSeqFlag = random.Generate32BitRandomString()
	}

	query := `INSERT INTO HUB_REGISTRY_EXTERNAL_CONFIG (
		externalConfigId, tenantId, configName, configDescription, registryType, environmentName,
		serverAddress, serverPort, serverPath, serverScheme,
		authEnabled, username, password, accessToken, secretKey,
		connectionTimeout, readTimeout, maxRetries, retryInterval,
		specificConfig, fieldMapping,
		failoverEnabled, failoverConfigId, failoverStrategy,
		syncEnabled, syncInterval, conflictResolution,
		addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.Exec(ctx, query, []interface{}{
		config.ExternalConfigId, config.TenantId, config.ConfigName, config.ConfigDescription, config.RegistryType, config.EnvironmentName,
		config.ServerAddress, config.ServerPort, config.ServerPath, config.ServerScheme,
		config.AuthEnabled, config.Username, config.Password, config.AccessToken, config.SecretKey,
		config.ConnectionTimeout, config.ReadTimeout, config.MaxRetries, config.RetryInterval,
		config.SpecificConfig, config.FieldMapping,
		config.FailoverEnabled, config.FailoverConfigId, config.FailoverStrategy,
		config.SyncEnabled, config.SyncInterval, config.ConflictResolution,
		config.AddTime, config.AddWho, config.EditTime, config.EditWho, config.OprSeqFlag, config.CurrentVersion, config.ActiveFlag, config.NoteText, config.ExtProperty,
	}, true)

	return err
}

// updateExternalConfig 更新外部配置
func (s *ExternalStorage) updateExternalConfig(ctx context.Context, config *core.ExternalRegistryConfig) error {
	config.EditTime = time.Now()
	config.CurrentVersion++

	query := `UPDATE HUB_REGISTRY_EXTERNAL_CONFIG SET
		configName = ?, configDescription = ?, registryType = ?, environmentName = ?,
		serverAddress = ?, serverPort = ?, serverPath = ?, serverScheme = ?,
		authEnabled = ?, username = ?, password = ?, accessToken = ?, secretKey = ?,
		connectionTimeout = ?, readTimeout = ?, maxRetries = ?, retryInterval = ?,
		specificConfig = ?, fieldMapping = ?,
		failoverEnabled = ?, failoverConfigId = ?, failoverStrategy = ?,
		syncEnabled = ?, syncInterval = ?, conflictResolution = ?,
		editTime = ?, editWho = ?, currentVersion = ?, activeFlag = ?, noteText = ?, extProperty = ?
	WHERE tenantId = ? AND externalConfigId = ?`

	_, err := s.db.Exec(ctx, query, []interface{}{
		config.ConfigName, config.ConfigDescription, config.RegistryType, config.EnvironmentName,
		config.ServerAddress, config.ServerPort, config.ServerPath, config.ServerScheme,
		config.AuthEnabled, config.Username, config.Password, config.AccessToken, config.SecretKey,
		config.ConnectionTimeout, config.ReadTimeout, config.MaxRetries, config.RetryInterval,
		config.SpecificConfig, config.FieldMapping,
		config.FailoverEnabled, config.FailoverConfigId, config.FailoverStrategy,
		config.SyncEnabled, config.SyncInterval, config.ConflictResolution,
		config.EditTime, config.EditWho, config.CurrentVersion, config.ActiveFlag, config.NoteText, config.ExtProperty,
		config.TenantId, config.ExternalConfigId,
	}, true)

	return err
}

// GetExternalConfig 获取外部配置
func (s *ExternalStorage) GetExternalConfig(ctx context.Context, tenantId, configId string) (*core.ExternalRegistryConfig, error) {
	query := `SELECT externalConfigId, tenantId, configName, configDescription, registryType, environmentName,
		serverAddress, serverPort, serverPath, serverScheme,
		authEnabled, username, password, accessToken, secretKey,
		connectionTimeout, readTimeout, maxRetries, retryInterval,
		specificConfig, fieldMapping,
		failoverEnabled, failoverConfigId, failoverStrategy,
		syncEnabled, syncInterval, conflictResolution,
		addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty
	FROM HUB_REGISTRY_EXTERNAL_CONFIG WHERE tenantId = ? AND externalConfigId = ? AND activeFlag = 'Y'`

	config := &core.ExternalRegistryConfig{}
	err := s.db.QueryOne(ctx, config, query, []interface{}{tenantId, configId}, true)
	if err == sql.ErrNoRows {
		return nil, core.ErrConfigNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query external config failed: %w", err)
	}

	return config, nil
}

// DeleteExternalConfig 删除外部配置
func (s *ExternalStorage) DeleteExternalConfig(ctx context.Context, tenantId, configId string) error {
	// 先断开连接
	s.Disconnect(ctx, configId)

	// 软删除配置
	query := `UPDATE HUB_REGISTRY_EXTERNAL_CONFIG SET activeFlag = 'N', editTime = ? WHERE tenantId = ? AND externalConfigId = ?`
	_, err := s.db.Exec(ctx, query, []interface{}{time.Now(), tenantId, configId}, true)
	return err
}

// ListExternalConfigs 列出外部配置
func (s *ExternalStorage) ListExternalConfigs(ctx context.Context, tenantId, registryType, environment string) ([]*core.ExternalRegistryConfig, error) {
	var query string
	var args []interface{}

	baseQuery := `SELECT externalConfigId, tenantId, configName, configDescription, registryType, environmentName,
		serverAddress, serverPort, serverPath, serverScheme,
		authEnabled, username, password, accessToken, secretKey,
		connectionTimeout, readTimeout, maxRetries, retryInterval,
		specificConfig, fieldMapping,
		failoverEnabled, failoverConfigId, failoverStrategy,
		syncEnabled, syncInterval, conflictResolution,
		addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty
	FROM HUB_REGISTRY_EXTERNAL_CONFIG WHERE tenantId = ? AND activeFlag = 'Y'`

	args = append(args, tenantId)

	if registryType != "" {
		baseQuery += " AND registryType = ?"
		args = append(args, registryType)
	}

	if environment != "" {
		baseQuery += " AND environmentName = ?"
		args = append(args, environment)
	}

	query = baseQuery + " ORDER BY configName"

	var configs []*core.ExternalRegistryConfig
	err := s.db.Query(ctx, &configs, query, args, true)
	if err != nil {
		return nil, fmt.Errorf("query external configs failed: %w", err)
	}

	return configs, nil
}

// ================== 状态管理 ==================

// SaveExternalStatus 保存外部状态
func (s *ExternalStorage) SaveExternalStatus(ctx context.Context, status *core.ExternalRegistryStatus) error {
	// 检查是否存在
	existing, err := s.GetExternalStatus(ctx, status.TenantId, status.ExternalConfigId)
	if err != nil && err != core.ErrConfigNotFound {
		return fmt.Errorf("check existing status failed: %w", err)
	}

	if existing != nil {
		// 更新
		return s.updateExternalStatus(ctx, status)
	} else {
		// 插入
		return s.insertExternalStatus(ctx, status)
	}
}

// insertExternalStatus 插入外部状态
func (s *ExternalStorage) insertExternalStatus(ctx context.Context, status *core.ExternalRegistryStatus) error {
	if status.ExternalStatusId == "" {
		status.ExternalStatusId = random.Generate32BitRandomString()
	}
	if status.OprSeqFlag == "" {
		status.OprSeqFlag = random.Generate32BitRandomString()
	}

	query := `INSERT INTO HUB_REGISTRY_EXTERNAL_STATUS (
		externalStatusId, tenantId, externalConfigId,
		connectionStatus, healthStatus, lastConnectTime, lastDisconnectTime, lastHealthCheckTime,
		responseTime, successCount, errorCount, timeoutCount,
		failoverStatus, failoverTime, failoverCount, recoverTime,
		syncStatus, lastSyncTime, syncSuccessCount, syncErrorCount,
		lastErrorMessage, lastErrorTime, errorDetails,
		addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.Exec(ctx, query, []interface{}{
		status.ExternalStatusId, status.TenantId, status.ExternalConfigId,
		status.ConnectionStatus, status.HealthStatus, status.LastConnectTime, status.LastDisconnectTime, status.LastHealthCheckTime,
		status.ResponseTime, status.SuccessCount, status.ErrorCount, status.TimeoutCount,
		status.FailoverStatus, status.FailoverTime, status.FailoverCount, status.RecoverTime,
		status.SyncStatus, status.LastSyncTime, status.SyncSuccessCount, status.SyncErrorCount,
		status.LastErrorMessage, status.LastErrorTime, status.ErrorDetails,
		status.AddTime, status.AddWho, status.EditTime, status.EditWho, status.OprSeqFlag, status.CurrentVersion, status.ActiveFlag, status.NoteText, status.ExtProperty,
	}, true)

	return err
}

// updateExternalStatus 更新外部状态
func (s *ExternalStorage) updateExternalStatus(ctx context.Context, status *core.ExternalRegistryStatus) error {
	status.EditTime = time.Now()
	status.CurrentVersion++

	query := `UPDATE HUB_REGISTRY_EXTERNAL_STATUS SET
		connectionStatus = ?, healthStatus = ?, lastConnectTime = ?, lastDisconnectTime = ?, lastHealthCheckTime = ?,
		responseTime = ?, successCount = ?, errorCount = ?, timeoutCount = ?,
		failoverStatus = ?, failoverTime = ?, failoverCount = ?, recoverTime = ?,
		syncStatus = ?, lastSyncTime = ?, syncSuccessCount = ?, syncErrorCount = ?,
		lastErrorMessage = ?, lastErrorTime = ?, errorDetails = ?,
		editTime = ?, editWho = ?, currentVersion = ?, activeFlag = ?, noteText = ?, extProperty = ?
	WHERE tenantId = ? AND externalConfigId = ?`

	_, err := s.db.Exec(ctx, query, []interface{}{
		status.ConnectionStatus, status.HealthStatus, status.LastConnectTime, status.LastDisconnectTime, status.LastHealthCheckTime,
		status.ResponseTime, status.SuccessCount, status.ErrorCount, status.TimeoutCount,
		status.FailoverStatus, status.FailoverTime, status.FailoverCount, status.RecoverTime,
		status.SyncStatus, status.LastSyncTime, status.SyncSuccessCount, status.SyncErrorCount,
		status.LastErrorMessage, status.LastErrorTime, status.ErrorDetails,
		status.EditTime, status.EditWho, status.CurrentVersion, status.ActiveFlag, status.NoteText, status.ExtProperty,
		status.TenantId, status.ExternalConfigId,
	}, true)

	return err
}

// GetExternalStatus 获取外部状态
func (s *ExternalStorage) GetExternalStatus(ctx context.Context, tenantId, configId string) (*core.ExternalRegistryStatus, error) {
	query := `SELECT externalStatusId, tenantId, externalConfigId,
		connectionStatus, healthStatus, lastConnectTime, lastDisconnectTime, lastHealthCheckTime,
		responseTime, successCount, errorCount, timeoutCount,
		failoverStatus, failoverTime, failoverCount, recoverTime,
		syncStatus, lastSyncTime, syncSuccessCount, syncErrorCount,
		lastErrorMessage, lastErrorTime, errorDetails,
		addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText, extProperty
	FROM HUB_REGISTRY_EXTERNAL_STATUS WHERE tenantId = ? AND externalConfigId = ? AND activeFlag = 'Y'`

	status := &core.ExternalRegistryStatus{}
	err := s.db.QueryOne(ctx, status, query, []interface{}{tenantId, configId}, true)
	if err == sql.ErrNoRows {
		return nil, core.ErrConfigNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query external status failed: %w", err)
	}

	return status, nil
}

// UpdateExternalStatus 更新外部状态
func (s *ExternalStorage) UpdateExternalStatus(ctx context.Context, tenantId, configId string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	// 构建动态更新SQL
	var setParts []string
	var args []interface{}

	for key, value := range updates {
		setParts = append(setParts, key+" = ?")
		args = append(args, value)
	}

	// 添加固定更新字段
	setParts = append(setParts, "editTime = ?", "currentVersion = currentVersion + 1")
	args = append(args, time.Now())

	// 添加WHERE条件参数
	args = append(args, tenantId, configId)

	query := fmt.Sprintf("UPDATE HUB_REGISTRY_EXTERNAL_STATUS SET %s WHERE tenantId = ? AND externalConfigId = ?",
		strings.Join(setParts, ", "))

	_, err := s.db.Exec(ctx, query, args, true)
	return err
}

// ================== 连接管理 ==================

// Connect 连接外部注册中心
func (s *ExternalStorage) Connect(ctx context.Context, config *core.ExternalRegistryConfig) error {
	s.connMutex.Lock()
	defer s.connMutex.Unlock()

	// 这里应该根据不同的注册中心类型创建相应的连接
	// 目前先做简单的模拟实现

	// 更新连接状态
	now := time.Now()
	updates := map[string]interface{}{
		"connectionStatus":    core.ConnectionStatusConnected,
		"lastConnectTime":     now,
		"healthStatus":        core.HealthStatusHealthy,
		"lastHealthCheckTime": now,
	}

	err := s.UpdateExternalStatus(ctx, config.TenantId, config.ExternalConfigId, updates)
	if err != nil {
		return fmt.Errorf("update connection status failed: %w", err)
	}

	// 保存连接到内存
	s.connections[config.ExternalConfigId] = struct{}{}

	return nil
}

// Disconnect 断开连接
func (s *ExternalStorage) Disconnect(ctx context.Context, configId string) error {
	s.connMutex.Lock()
	defer s.connMutex.Unlock()

	// 从内存中移除连接
	delete(s.connections, configId)

	// 更新连接状态
	// 这里需要获取tenantId，暂时跳过状态更新
	// 在实际实现中应该从配置中获取tenantId
	// updates := map[string]interface{}{
	//     "connectionStatus":    core.ConnectionStatusDisconnected,
	//     "lastDisconnectTime":  time.Now(),
	// }

	return nil
}

// IsConnected 检查是否已连接
func (s *ExternalStorage) IsConnected(configId string) bool {
	s.connMutex.RLock()
	defer s.connMutex.RUnlock()

	_, exists := s.connections[configId]
	return exists
}

// ================== 健康检查 ==================

// HealthCheck 健康检查
func (s *ExternalStorage) HealthCheck(ctx context.Context, configId string) error {
	// 这里应该实现具体的健康检查逻辑
	// 目前先做简单的模拟实现

	if !s.IsConnected(configId) {
		return core.ErrNotConnected
	}

	// 模拟健康检查成功
	return nil
}

// GetConnectionStatus 获取连接状态
func (s *ExternalStorage) GetConnectionStatus(configId string) string {
	if s.IsConnected(configId) {
		return core.ConnectionStatusConnected
	}
	return core.ConnectionStatusDisconnected
}

// ================== 服务发现（代理模式） ==================

// DiscoverServices 发现服务
func (s *ExternalStorage) DiscoverServices(ctx context.Context, configId string, filters ...core.ServiceFilter) ([]*core.UnifiedServiceInstance, error) {
	// 这里应该实现从外部注册中心发现服务的逻辑
	// 目前先返回空列表
	return []*core.UnifiedServiceInstance{}, nil
}

// DiscoverInstances 发现实例
func (s *ExternalStorage) DiscoverInstances(ctx context.Context, configId, serviceName string, filters ...core.InstanceFilter) ([]*core.UnifiedServiceInstance, error) {
	// 这里应该实现从外部注册中心发现实例的逻辑
	// 目前先返回空列表
	return []*core.UnifiedServiceInstance{}, nil
}

// ================== 监控 ==================

// GetMetrics 获取指标
func (s *ExternalStorage) GetMetrics(ctx context.Context, configId string) (*core.ExternalRegistryMetrics, error) {
	// 从状态表获取指标数据
	status, err := s.GetExternalStatus(ctx, "", configId) // 这里需要传入正确的tenantId
	if err != nil {
		return nil, fmt.Errorf("get external status failed: %w", err)
	}

	config, err := s.GetExternalConfig(ctx, "", configId) // 这里需要传入正确的tenantId
	if err != nil {
		return nil, fmt.Errorf("get external config failed: %w", err)
	}

	metrics := &core.ExternalRegistryMetrics{
		ConfigId:         configId,
		RegistryType:     config.RegistryType,
		ConnectionStatus: status.ConnectionStatus,
		HealthStatus:     status.HealthStatus,
		ResponseTime:     time.Duration(status.ResponseTime) * time.Millisecond,
		RequestCount:     status.SuccessCount + status.ErrorCount,
		SuccessCount:     status.SuccessCount,
		ErrorCount:       status.ErrorCount,
		LastCheckTime:    time.Now(),
	}

	if status.LastHealthCheckTime != nil {
		metrics.LastCheckTime = *status.LastHealthCheckTime
	}

	return metrics, nil
}
