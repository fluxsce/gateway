package dao

import (
	"context"
	"fmt"
	"time"

	"gateway/internal/alert/types"
	"gateway/pkg/database"
)

// ConfigDAO 告警配置数据访问对象
type ConfigDAO struct {
	db database.Database
}

// NewConfigDAO 创建配置DAO
func NewConfigDAO(db database.Database) *ConfigDAO {
	return &ConfigDAO{db: db}
}

// SaveConfig 保存渠道配置
func (d *ConfigDAO) SaveConfig(ctx context.Context, config *types.AlertConfig) error {
	_, err := d.db.Insert(ctx, "HUB_ALERT_CONFIG", config, true)
	if err != nil {
		return fmt.Errorf("保存告警配置失败: %w", err)
	}
	return nil
}

// GetConfig 获取渠道配置
func (d *ConfigDAO) GetConfig(ctx context.Context, tenantId, channelName string) (*types.AlertConfig, error) {
	query := "SELECT * FROM HUB_ALERT_CONFIG WHERE tenantId = ? AND channelName = ?"
	args := []interface{}{tenantId, channelName}

	var config types.AlertConfig
	err := d.db.QueryOne(ctx, &config, query, args, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询告警配置失败: %w", err)
	}

	return &config, nil
}

// ListConfigs 列出渠道配置
func (d *ConfigDAO) ListConfigs(ctx context.Context, tenantId string, activeOnly bool) ([]*types.AlertConfig, error) {
	query := "SELECT * FROM HUB_ALERT_CONFIG WHERE tenantId = ?"
	args := []interface{}{tenantId}

	if activeOnly {
		query += " AND activeFlag = 'Y'"
	}

	query += " ORDER BY priorityLevel ASC, channelName ASC"

	var configs []*types.AlertConfig
	err := d.db.Query(ctx, &configs, query, args, true)
	if err != nil {
		return nil, fmt.Errorf("查询告警配置列表失败: %w", err)
	}

	return configs, nil
}

// GetDefaultConfig 获取默认渠道配置
func (d *ConfigDAO) GetDefaultConfig(ctx context.Context, tenantId string) (*types.AlertConfig, error) {
	query := "SELECT * FROM HUB_ALERT_CONFIG WHERE tenantId = ? AND defaultFlag = 'Y' AND activeFlag = 'Y' ORDER BY priorityLevel ASC LIMIT 1"
	args := []interface{}{tenantId}

	var config types.AlertConfig
	err := d.db.QueryOne(ctx, &config, query, args, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询默认告警配置失败: %w", err)
	}

	return &config, nil
}

// UpdateConfig 更新渠道配置
func (d *ConfigDAO) UpdateConfig(ctx context.Context, config *types.AlertConfig) error {
	whereClause := "tenantId = ? AND channelName = ?"
	whereArgs := []interface{}{config.TenantId, config.ChannelName}
	_, err := d.db.Update(ctx, "HUB_ALERT_CONFIG", config, whereClause, whereArgs, true)
	if err != nil {
		return fmt.Errorf("更新告警配置失败: %w", err)
	}
	return nil
}

// UpdateConfigStats 更新配置统计信息
func (d *ConfigDAO) UpdateConfigStats(ctx context.Context, tenantId, channelName string, success bool, errorMsg *string) error {
	// 获取配置
	config, err := d.GetConfig(ctx, tenantId, channelName)
	if err != nil {
		return err
	}
	if config == nil {
		return fmt.Errorf("配置不存在")
	}

	now := time.Now()

	// 更新统计信息
	config.TotalSentCount++
	if success {
		config.SuccessCount++
		nowTime := now
		config.LastSendTime = &nowTime
		config.LastSuccessTime = &nowTime
		config.LastErrorMessage = nil
	} else {
		config.FailureCount++
		nowTime := now
		config.LastSendTime = &nowTime
		config.LastFailureTime = &nowTime
		if errorMsg != nil {
			errorMsgStr := *errorMsg
			if len(errorMsgStr) > 1000 {
				errorMsgStr = errorMsgStr[:1000]
			}
			config.LastErrorMessage = &errorMsgStr
		}
	}
	config.EditTime = now
	config.EditWho = "system"

	return d.UpdateConfig(ctx, config)
}
