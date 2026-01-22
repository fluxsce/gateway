package dao

import (
	"context"
	"fmt"
	"time"

	"gateway/internal/alert/types"
	"gateway/pkg/database"
)

// LogDAO 告警日志数据访问对象
type LogDAO struct {
	db database.Database
}

// NewLogDAO 创建日志DAO
func NewLogDAO(db database.Database) *LogDAO {
	return &LogDAO{db: db}
}

// SaveLog 保存告警日志
func (d *LogDAO) SaveLog(ctx context.Context, log *types.AlertLog) error {
	_, err := d.db.Insert(ctx, "HUB_ALERT_LOG", log, true)
	if err != nil {
		return fmt.Errorf("保存告警日志失败: %w", err)
	}
	return nil
}

// BatchSaveLogs 批量保存告警日志
func (d *LogDAO) BatchSaveLogs(ctx context.Context, logs []*types.AlertLog) error {
	if len(logs) == 0 {
		return nil
	}
	_, err := d.db.BatchInsert(ctx, "HUB_ALERT_LOG", logs, true)
	if err != nil {
		return fmt.Errorf("批量保存告警日志失败: %w", err)
	}
	return nil
}

// GetLog 获取告警日志
func (d *LogDAO) GetLog(ctx context.Context, tenantId, alertLogId string) (*types.AlertLog, error) {
	query := "SELECT * FROM HUB_ALERT_LOG WHERE tenantId = ? AND alertLogId = ?"
	args := []interface{}{tenantId, alertLogId}

	var log types.AlertLog
	err := d.db.QueryOne(ctx, &log, query, args, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询告警日志失败: %w", err)
	}

	return &log, nil
}

// UpdateLog 更新告警日志
func (d *LogDAO) UpdateLog(ctx context.Context, log *types.AlertLog) error {
	whereClause := "tenantId = ? AND alertLogId = ?"
	whereArgs := []interface{}{log.TenantId, log.AlertLogId}
	_, err := d.db.Update(ctx, "HUB_ALERT_LOG", log, whereClause, whereArgs, true)
	if err != nil {
		return fmt.Errorf("更新告警日志失败: %w", err)
	}
	return nil
}

// GetPendingLogs 获取待发送的告警日志
func (d *LogDAO) GetPendingLogs(ctx context.Context, tenantId string, limit int) ([]*types.AlertLog, error) {
	query := "SELECT * FROM HUB_ALERT_LOG WHERE tenantId = ? AND sendStatus = 'PENDING' AND activeFlag = 'Y' ORDER BY alertTimestamp ASC LIMIT ?"
	args := []interface{}{tenantId, limit}

	var logs []*types.AlertLog
	err := d.db.Query(ctx, &logs, query, args, true)
	if err != nil {
		return nil, fmt.Errorf("查询待发送告警日志失败: %w", err)
	}

	return logs, nil
}

// CleanupOldLogs 清理旧的告警日志
func (d *LogDAO) CleanupOldLogs(ctx context.Context, tenantId string, beforeTime time.Time) (int64, error) {
	whereClause := "tenantId = ? AND alertTimestamp < ?"
	whereArgs := []interface{}{tenantId, beforeTime}
	affected, err := d.db.Delete(ctx, "HUB_ALERT_LOG", whereClause, whereArgs, true)
	if err != nil {
		return 0, fmt.Errorf("清理旧告警日志失败: %w", err)
	}
	return affected, nil
}
