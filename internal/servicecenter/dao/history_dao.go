package dao

import (
	"context"
	"fmt"

	"gateway/internal/servicecenter/types"
	"gateway/pkg/database"
)

// HistoryDAO 配置历史数据访问对象
type HistoryDAO struct {
	db database.Database
}

// NewHistoryDAO 创建配置历史DAO
func NewHistoryDAO(db database.Database) *HistoryDAO {
	return &HistoryDAO{db: db}
}

// CreateHistory 创建配置历史记录
func (d *HistoryDAO) CreateHistory(ctx context.Context, history *types.ConfigHistory) error {
	_, err := d.db.Insert(ctx, "HUB_SERVICE_CONFIG_HISTORY", history, true)
	if err != nil {
		return fmt.Errorf("创建配置历史失败: %w", err)
	}
	return nil
}

// GetConfigHistory 获取配置历史
func (d *HistoryDAO) GetConfigHistory(ctx context.Context, tenantId, namespaceId, groupName, configDataId string, limit int) ([]*types.ConfigHistory, error) {
	if limit <= 0 {
		limit = 50 // 默认50条
	}

	query := "SELECT * FROM HUB_SERVICE_CONFIG_HISTORY WHERE tenantId = ? AND namespaceId = ? AND groupName = ? AND configDataId = ? ORDER BY changedAt DESC LIMIT ?"
	args := []interface{}{tenantId, namespaceId, groupName, configDataId, limit}

	var histories []*types.ConfigHistory
	err := d.db.Query(ctx, &histories, query, args, true)
	if err != nil {
		return nil, fmt.Errorf("查询配置历史失败: %w", err)
	}

	return histories, nil
}

// GetHistoryByVersion 根据版本号获取配置历史
func (d *HistoryDAO) GetHistoryByVersion(ctx context.Context, tenantId, namespaceId, groupName, configDataId string, version int64) (*types.ConfigHistory, error) {
	query := "SELECT * FROM HUB_SERVICE_CONFIG_HISTORY WHERE tenantId = ? AND namespaceId = ? AND groupName = ? AND configDataId = ? AND newVersion = ? ORDER BY changedAt DESC LIMIT 1"
	args := []interface{}{tenantId, namespaceId, groupName, configDataId, version}

	var history types.ConfigHistory
	err := d.db.QueryOne(ctx, &history, query, args, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询配置历史失败: %w", err)
	}

	return &history, nil
}
