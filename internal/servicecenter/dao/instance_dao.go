package dao

import (
	"context"
	"fmt"

	"gateway/internal/servicecenter/types"
	"gateway/pkg/database"
)

// InstanceDAO 服务中心实例配置数据访问对象
// 管理 HUB_SERVICE_INSTANCE 表（服务中心实例配置）
type InstanceDAO struct {
	db database.Database
}

// NewInstanceDAO 创建服务中心实例配置DAO
func NewInstanceDAO(db database.Database) *InstanceDAO {
	return &InstanceDAO{db: db}
}

// GetInstance 获取指定实例的配置
func (d *InstanceDAO) GetInstance(ctx context.Context, tenantID, instanceName, environment string) (*types.InstanceConfig, error) {
	query := `
		SELECT tenantId, instanceName, environment, serverType,
		       listenAddress, listenPort,
		       maxRecvMsgSize, maxSendMsgSize,
		       keepAliveTime, keepAliveTimeout, keepAliveMinTime, permitWithoutStream,
		       maxConnectionIdle, maxConnectionAge, maxConnectionAgeGrace,
		       enableReflection, enableTLS,
		       certStorageType, certFilePath, keyFilePath,
		       certContent, keyContent, certChainContent, certPassword, enableMTLS,
		       maxConcurrentStreams, readBufferSize, writeBufferSize,
		       healthCheckInterval, healthCheckTimeout,
		       instanceStatus, statusMessage, lastStatusTime, lastHealthCheckTime,
		       addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion,
		       activeFlag, noteText, extProperty,
		       enableAuth, ipWhitelist, ipBlacklist
		FROM HUB_SERVICE_INSTANCE
		WHERE tenantId = ? AND instanceName = ? AND environment = ?
	`

	config := &types.InstanceConfig{}
	err := d.db.QueryOne(ctx, config, query, []interface{}{tenantID, instanceName, environment}, true)

	if err == database.ErrRecordNotFound {
		return nil, nil // 返回 nil 表示未找到
	}
	if err != nil {
		return nil, fmt.Errorf("查询实例配置失败: %w", err)
	}

	return config, nil
}

// ListInstances 获取指定租户和环境的所有实例配置
func (d *InstanceDAO) ListInstances(ctx context.Context, tenantID, environment string) ([]*types.InstanceConfig, error) {
	query := `
		SELECT tenantId, instanceName, environment, serverType,
		       listenAddress, listenPort,
		       maxRecvMsgSize, maxSendMsgSize,
		       keepAliveTime, keepAliveTimeout, keepAliveMinTime, permitWithoutStream,
		       maxConnectionIdle, maxConnectionAge, maxConnectionAgeGrace,
		       enableReflection, enableTLS,
		       certStorageType, certFilePath, keyFilePath,
		       certContent, keyContent, certChainContent, certPassword, enableMTLS,
		       maxConcurrentStreams, readBufferSize, writeBufferSize,
		       healthCheckInterval, healthCheckTimeout,
		       instanceStatus, statusMessage, lastStatusTime, lastHealthCheckTime,
		       addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion,
		       activeFlag, noteText, extProperty,
		       enableAuth, ipWhitelist, ipBlacklist
		FROM HUB_SERVICE_INSTANCE
		WHERE tenantId = ? AND environment = ? AND activeFlag = 'Y'
		ORDER BY instanceName
	`

	var configs []*types.InstanceConfig
	err := d.db.Query(ctx, &configs, query, []interface{}{tenantID, environment}, true)

	if err != nil {
		return nil, fmt.Errorf("查询实例配置列表失败: %w", err)
	}

	return configs, nil
}

// ListAllInstances 获取指定租户的所有实例配置（所有环境）
func (d *InstanceDAO) ListAllInstances(ctx context.Context, tenantID string) ([]*types.InstanceConfig, error) {
	query := `
		SELECT tenantId, instanceName, environment, serverType,
		       listenAddress, listenPort,
		       maxRecvMsgSize, maxSendMsgSize,
		       keepAliveTime, keepAliveTimeout, keepAliveMinTime, permitWithoutStream,
		       maxConnectionIdle, maxConnectionAge, maxConnectionAgeGrace,
		       enableReflection, enableTLS,
		       certStorageType, certFilePath, keyFilePath,
		       certContent, keyContent, certChainContent, certPassword, enableMTLS,
		       maxConcurrentStreams, readBufferSize, writeBufferSize,
		       healthCheckInterval, healthCheckTimeout,
		       instanceStatus, statusMessage, lastStatusTime, lastHealthCheckTime,
		       addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion,
		       activeFlag, noteText, extProperty,
		       enableAuth, ipWhitelist, ipBlacklist
		FROM HUB_SERVICE_INSTANCE
		WHERE tenantId = ? AND activeFlag = 'Y'
		ORDER BY environment, instanceName
	`

	var configs []*types.InstanceConfig
	err := d.db.Query(ctx, &configs, query, []interface{}{tenantID}, true)

	if err != nil {
		return nil, fmt.Errorf("查询所有实例配置列表失败: %w", err)
	}

	return configs, nil
}

// CreateInstance 创建新实例配置
func (d *InstanceDAO) CreateInstance(ctx context.Context, config *types.InstanceConfig) error {
	_, err := d.db.Insert(ctx, "HUB_SERVICE_INSTANCE", config, true)
	if err != nil {
		return fmt.Errorf("创建实例配置失败: %w", err)
	}
	return nil
}

// UpdateInstance 更新实例配置
func (d *InstanceDAO) UpdateInstance(ctx context.Context, config *types.InstanceConfig) error {
	where := "tenantId = ? AND instanceName = ? AND environment = ?"
	args := []interface{}{config.TenantID, config.InstanceName, config.Environment}

	_, err := d.db.Update(ctx, "HUB_SERVICE_INSTANCE", config, where, args, true, true)
	if err != nil {
		return fmt.Errorf("更新实例配置失败: %w", err)
	}
	return nil
}

// DeleteInstance 删除实例配置（物理删除）
func (d *InstanceDAO) DeleteInstance(ctx context.Context, tenantID, instanceName, environment string) error {
	where := "tenantId = ? AND instanceName = ? AND environment = ?"
	args := []interface{}{tenantID, instanceName, environment}

	_, err := d.db.Delete(ctx, "HUB_SERVICE_INSTANCE", where, args, true)
	if err != nil {
		return fmt.Errorf("删除实例配置失败: %w", err)
	}
	return nil
}

// UpdateInstanceStatus 更新实例状态信息
// updates: 要更新的字段 map，如 {"instanceStatus": "RUNNING", "statusMessage": "...", "lastStatusTime": time.Now()}
func (d *InstanceDAO) UpdateInstanceStatus(ctx context.Context, tenantID, instanceName, environment string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return fmt.Errorf("更新字段不能为空")
	}

	// 构建 SET 子句和参数
	setClauses := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+3)

	for field, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", field))
		args = append(args, value)
	}

	// WHERE 参数
	args = append(args, tenantID, instanceName, environment)

	// 构建 SQL（跨数据库兼容）
	query := fmt.Sprintf(`
		UPDATE HUB_SERVICE_INSTANCE
		SET %s
		WHERE tenantId = ? AND instanceName = ? AND environment = ?
	`, joinStrings(setClauses, ", "))

	_, err := d.db.Exec(ctx, query, args, true)
	if err != nil {
		return fmt.Errorf("更新实例状态失败: %w", err)
	}

	return nil
}

// joinStrings 连接字符串切片（辅助函数）
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
