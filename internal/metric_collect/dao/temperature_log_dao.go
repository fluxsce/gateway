package dao

import (
	"context"
	"fmt"
	"gateway/internal/metric_collect/types"
	"gateway/pkg/database"
	"time"
)

// TemperatureLogDAO 温度信息日志数据访问对象
type TemperatureLogDAO struct {
	db database.Database
}

// NewTemperatureLogDAO 创建温度信息日志DAO实例
func NewTemperatureLogDAO(db database.Database) *TemperatureLogDAO {
	return &TemperatureLogDAO{db: db}
}

// InsertTemperatureLog 插入温度信息日志
func (dao *TemperatureLogDAO) InsertTemperatureLog(ctx context.Context, temperatureLog *types.TemperatureLog) error {
	// 设置通用字段默认值
	now := time.Now()
	temperatureLog.AddTime = now
	temperatureLog.EditTime = now
	temperatureLog.CurrentVersion = 1
	temperatureLog.ActiveFlag = types.ActiveFlagYes

	// 验证必填字段
	if temperatureLog.MetricTemperatureLogId == "" {
		return fmt.Errorf("温度信息日志ID不能为空")
	}
	if temperatureLog.TenantId == "" {
		return fmt.Errorf("租户ID不能为空")
	}
	if temperatureLog.MetricServerId == "" {
		return fmt.Errorf("服务器ID不能为空")
	}
	if temperatureLog.AddWho == "" {
		return fmt.Errorf("创建人ID不能为空")
	}
	if temperatureLog.EditWho == "" {
		return fmt.Errorf("修改人ID不能为空")
	}
	if temperatureLog.OprSeqFlag == "" {
		return fmt.Errorf("操作序列标识不能为空")
	}

	tableName := temperatureLog.TableName()
	_, err := dao.db.Insert(ctx, tableName, temperatureLog, true)
	if err != nil {
		return fmt.Errorf("插入温度信息日志失败: %w", err)
	}
	return nil
}

// BatchInsertTemperatureLog 批量插入温度信息日志
func (dao *TemperatureLogDAO) BatchInsertTemperatureLog(ctx context.Context, temperatureLogs []*types.TemperatureLog) error {
	if len(temperatureLogs) == 0 {
		return nil
	}

	// 设置通用字段默认值
	now := time.Now()
	for _, temperatureLog := range temperatureLogs {
		temperatureLog.AddTime = now
		temperatureLog.EditTime = now
		temperatureLog.CurrentVersion = 1
		temperatureLog.ActiveFlag = types.ActiveFlagYes

		// 验证必填字段
		if temperatureLog.MetricTemperatureLogId == "" {
			return fmt.Errorf("温度信息日志ID不能为空")
		}
		if temperatureLog.TenantId == "" {
			return fmt.Errorf("租户ID不能为空")
		}
		if temperatureLog.MetricServerId == "" {
			return fmt.Errorf("服务器ID不能为空")
		}
		if temperatureLog.AddWho == "" {
			return fmt.Errorf("创建人ID不能为空")
		}
		if temperatureLog.EditWho == "" {
			return fmt.Errorf("修改人ID不能为空")
		}
		if temperatureLog.OprSeqFlag == "" {
			return fmt.Errorf("操作序列标识不能为空")
		}
	}

	tableName := temperatureLogs[0].TableName()

	// 转换为interface{}切片
	items := make([]interface{}, len(temperatureLogs))
	for i, temperatureLog := range temperatureLogs {
		items[i] = temperatureLog
	}

	_, err := dao.db.BatchInsert(ctx, tableName, items, true)
	if err != nil {
		return fmt.Errorf("批量插入温度信息日志失败: %w", err)
	}
	return nil
}

// DeleteTemperatureLog 删除温度信息日志（软删除）
func (dao *TemperatureLogDAO) DeleteTemperatureLog(ctx context.Context, tenantId, metricTemperatureLogId string) error {
	if tenantId == "" {
		return fmt.Errorf("租户ID不能为空")
	}
	if metricTemperatureLogId == "" {
		return fmt.Errorf("温度信息日志ID不能为空")
	}

	tableName := (&types.TemperatureLog{}).TableName()
	sql := fmt.Sprintf("UPDATE %s SET activeFlag = ? WHERE tenantId = ? AND metricTemperatureLogId = ?", tableName)

	_, err := dao.db.Exec(ctx, sql, []interface{}{types.ActiveFlagNo, tenantId, metricTemperatureLogId}, true)
	if err != nil {
		return fmt.Errorf("删除温度信息日志失败: %w", err)
	}
	return nil
}

// DeleteTemperatureLogByTime 根据时间范围删除温度信息日志（物理删除）
func (dao *TemperatureLogDAO) DeleteTemperatureLogByTime(ctx context.Context, tenantId string, beforeTime time.Time) error {
	if tenantId == "" {
		return fmt.Errorf("租户ID不能为空")
	}

	tableName := (&types.TemperatureLog{}).TableName()
	sql := fmt.Sprintf("DELETE FROM %s WHERE tenantId = ? AND collectTime < ?", tableName)

	_, err := dao.db.Exec(ctx, sql, []interface{}{tenantId, beforeTime}, true)
	if err != nil {
		return fmt.Errorf("根据时间删除温度信息日志失败: %w", err)
	}
	return nil
}

// DeleteTemperatureLogByServer 根据服务器ID删除温度信息日志（物理删除）
func (dao *TemperatureLogDAO) DeleteTemperatureLogByServer(ctx context.Context, tenantId, metricServerId string) error {
	if tenantId == "" {
		return fmt.Errorf("租户ID不能为空")
	}
	if metricServerId == "" {
		return fmt.Errorf("服务器ID不能为空")
	}

	tableName := (&types.TemperatureLog{}).TableName()
	sql := fmt.Sprintf("DELETE FROM %s WHERE tenantId = ? AND metricServerId = ?", tableName)

	_, err := dao.db.Exec(ctx, sql, []interface{}{tenantId, metricServerId}, true)
	if err != nil {
		return fmt.Errorf("根据服务器ID删除温度信息日志失败: %w", err)
	}
	return nil
}
