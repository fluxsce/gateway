package dao

import (
	"context"
	"crypto/md5"
	"fmt"
	"time"

	"gateway/internal/servicecenter/types"
	"gateway/pkg/database"
)

// ConfigDAO 配置数据访问对象
type ConfigDAO struct {
	db database.Database
}

// NewConfigDAO 创建配置DAO
func NewConfigDAO(db database.Database) *ConfigDAO {
	return &ConfigDAO{db: db}
}

// GetConfig 获取配置
// GetConfig 获取配置（不过滤 activeFlag，支持查询已删除的配置）
func (d *ConfigDAO) GetConfig(ctx context.Context, tenantId, namespaceId, groupName, configDataId string) (*types.ConfigData, error) {
	query := "SELECT * FROM HUB_SERVICE_CONFIG_DATA WHERE tenantId = ? AND namespaceId = ? AND groupName = ? AND configDataId = ?"
	args := []interface{}{tenantId, namespaceId, groupName, configDataId}

	var config types.ConfigData
	err := d.db.QueryOne(ctx, &config, query, args, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询配置失败: %w", err)
	}

	return &config, nil
}

// SaveConfig 保存配置
func (d *ConfigDAO) SaveConfig(ctx context.Context, config *types.ConfigData) error {
	// 验证必填字段
	if config.ConfigDataId == "" {
		return fmt.Errorf("configDataId不能为空")
	}
	if config.NamespaceId == "" {
		return fmt.Errorf("namespaceId不能为空")
	}
	if config.GroupName == "" {
		return fmt.Errorf("groupName不能为空")
	}

	// 计算MD5值
	hash := md5.Sum([]byte(config.ConfigContent))
	config.Md5Value = fmt.Sprintf("%x", hash)

	// 设置时间字段默认值
	now := time.Now()

	// 查询配置是否存在
	existingConfig, err := d.GetConfig(ctx, config.TenantId, config.NamespaceId, config.GroupName, config.ConfigDataId)
	if err != nil {
		return fmt.Errorf("查询配置失败: %w", err)
	}

	if existingConfig == nil {
		// 配置不存在，执行插入
		if config.AddTime.IsZero() {
			config.AddTime = now
		}
		if config.EditTime.IsZero() {
			config.EditTime = now
		}
		if config.Version == 0 {
			config.Version = 1
		}
		_, err := d.db.Insert(ctx, "HUB_SERVICE_CONFIG_DATA", config, true)
		if err != nil {
			return fmt.Errorf("插入配置失败: %w", err)
		}
	} else {
		// 配置存在，执行更新
		// 版本号递增
		config.Version = existingConfig.Version + 1
		// 保持原创建时间和创建人
		config.AddTime = existingConfig.AddTime
		config.AddWho = existingConfig.AddWho
		// 更新时，EditTime 应该设置为当前时间
		config.EditTime = now
		where := "tenantId = ? AND namespaceId = ? AND groupName = ? AND configDataId = ?"
		args := []interface{}{config.TenantId, config.NamespaceId, config.GroupName, config.ConfigDataId}
		_, err := d.db.Update(ctx, "HUB_SERVICE_CONFIG_DATA", config, where, args, true, true)
		if err != nil {
			return fmt.Errorf("更新配置失败: %w", err)
		}
	}

	return nil
}

// DeleteConfig 删除配置（物理删除）
func (d *ConfigDAO) DeleteConfig(ctx context.Context, tenantId, namespaceId, groupName, configDataId string) error {
	query := "DELETE FROM HUB_SERVICE_CONFIG_DATA WHERE tenantId = ? AND namespaceId = ? AND groupName = ? AND configDataId = ?"
	args := []interface{}{tenantId, namespaceId, groupName, configDataId}

	_, err := d.db.Exec(ctx, query, args, true)
	if err != nil {
		return fmt.Errorf("删除配置失败: %w", err)
	}
	return nil
}

// ListConfigs 列出配置列表
func (d *ConfigDAO) ListConfigs(ctx context.Context, tenantId, namespaceId, groupName string) ([]*types.ConfigData, error) {
	query := "SELECT * FROM HUB_SERVICE_CONFIG_DATA WHERE tenantId = ? AND namespaceId = ? AND activeFlag = 'Y'"
	args := []interface{}{tenantId, namespaceId}

	if groupName != "" {
		query += " AND groupName = ?"
		args = append(args, groupName)
	}

	query += " ORDER BY addTime DESC"

	var configs []*types.ConfigData
	err := d.db.Query(ctx, &configs, query, args, true)
	if err != nil {
		return nil, fmt.Errorf("查询配置列表失败: %w", err)
	}

	return configs, nil
}
