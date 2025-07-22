package dao

import (
	"context"
	"fmt"
	"gateway/internal/metric_collect/types"
	"gateway/pkg/database"
	"time"
)

// ProcessStatsLogDAO 进程统计日志数据访问对象
type ProcessStatsLogDAO struct {
	db database.Database
}

// NewProcessStatsLogDAO 创建进程统计日志DAO实例
func NewProcessStatsLogDAO(db database.Database) *ProcessStatsLogDAO {
	return &ProcessStatsLogDAO{db: db}
}

// InsertProcessStatsLog 插入进程统计日志
func (dao *ProcessStatsLogDAO) InsertProcessStatsLog(ctx context.Context, processStatsLog *types.ProcessStatsLog) error {
	// 设置通用字段默认值
	now := time.Now()
	processStatsLog.AddTime = now
	processStatsLog.EditTime = now
	processStatsLog.CurrentVersion = 1
	processStatsLog.ActiveFlag = types.ActiveFlagYes

	// 验证必填字段
	if processStatsLog.MetricProcessStatsLogId == "" {
		return fmt.Errorf("进程统计日志ID不能为空")
	}
	if processStatsLog.TenantId == "" {
		return fmt.Errorf("租户ID不能为空")
	}
	if processStatsLog.MetricServerId == "" {
		return fmt.Errorf("服务器ID不能为空")
	}
	if processStatsLog.AddWho == "" {
		return fmt.Errorf("创建人ID不能为空")
	}
	if processStatsLog.EditWho == "" {
		return fmt.Errorf("修改人ID不能为空")
	}
	if processStatsLog.OprSeqFlag == "" {
		return fmt.Errorf("操作序列标识不能为空")
	}

	tableName := processStatsLog.TableName()
	_, err := dao.db.Insert(ctx, tableName, processStatsLog, true)
	if err != nil {
		return fmt.Errorf("插入进程统计日志失败: %w", err)
	}
	return nil
}

// BatchInsertProcessStatsLog 批量插入进程统计日志
func (dao *ProcessStatsLogDAO) BatchInsertProcessStatsLog(ctx context.Context, processStatsLogs []*types.ProcessStatsLog) error {
	if len(processStatsLogs) == 0 {
		return nil
	}

	// 设置通用字段默认值
	now := time.Now()
	for _, processStatsLog := range processStatsLogs {
		processStatsLog.AddTime = now
		processStatsLog.EditTime = now
		processStatsLog.CurrentVersion = 1
		processStatsLog.ActiveFlag = types.ActiveFlagYes

		// 验证必填字段
		if processStatsLog.MetricProcessStatsLogId == "" {
			return fmt.Errorf("进程统计日志ID不能为空")
		}
		if processStatsLog.TenantId == "" {
			return fmt.Errorf("租户ID不能为空")
		}
		if processStatsLog.MetricServerId == "" {
			return fmt.Errorf("服务器ID不能为空")
		}
		if processStatsLog.AddWho == "" {
			return fmt.Errorf("创建人ID不能为空")
		}
		if processStatsLog.EditWho == "" {
			return fmt.Errorf("修改人ID不能为空")
		}
		if processStatsLog.OprSeqFlag == "" {
			return fmt.Errorf("操作序列标识不能为空")
		}
	}

	tableName := processStatsLogs[0].TableName()

	// 转换为interface{}切片
	items := make([]interface{}, len(processStatsLogs))
	for i, processStatsLog := range processStatsLogs {
		items[i] = processStatsLog
	}

	_, err := dao.db.BatchInsert(ctx, tableName, items, true)
	if err != nil {
		return fmt.Errorf("批量插入进程统计日志失败: %w", err)
	}
	return nil
}

// DeleteProcessStatsLog 删除进程统计日志（软删除）
func (dao *ProcessStatsLogDAO) DeleteProcessStatsLog(ctx context.Context, tenantId, metricProcessStatsLogId string) error {
	if tenantId == "" {
		return fmt.Errorf("租户ID不能为空")
	}
	if metricProcessStatsLogId == "" {
		return fmt.Errorf("进程统计日志ID不能为空")
	}

	tableName := (&types.ProcessStatsLog{}).TableName()
	sql := fmt.Sprintf("UPDATE %s SET activeFlag = ? WHERE tenantId = ? AND metricProcessStatsLogId = ?", tableName)

	_, err := dao.db.Exec(ctx, sql, []interface{}{types.ActiveFlagNo, tenantId, metricProcessStatsLogId}, true)
	if err != nil {
		return fmt.Errorf("删除进程统计日志失败: %w", err)
	}
	return nil
}

// DeleteProcessStatsLogByTime 根据时间范围删除进程统计日志（物理删除）
func (dao *ProcessStatsLogDAO) DeleteProcessStatsLogByTime(ctx context.Context, tenantId string, beforeTime time.Time) error {
	if tenantId == "" {
		return fmt.Errorf("租户ID不能为空")
	}

	tableName := (&types.ProcessStatsLog{}).TableName()
	sql := fmt.Sprintf("DELETE FROM %s WHERE tenantId = ? AND collectTime < ?", tableName)

	_, err := dao.db.Exec(ctx, sql, []interface{}{tenantId, beforeTime}, true)
	if err != nil {
		return fmt.Errorf("根据时间删除进程统计日志失败: %w", err)
	}
	return nil
}

// DeleteProcessStatsLogByServer 根据服务器ID删除进程统计日志（物理删除）
func (dao *ProcessStatsLogDAO) DeleteProcessStatsLogByServer(ctx context.Context, tenantId, metricServerId string) error {
	if tenantId == "" {
		return fmt.Errorf("租户ID不能为空")
	}
	if metricServerId == "" {
		return fmt.Errorf("服务器ID不能为空")
	}

	tableName := (&types.ProcessStatsLog{}).TableName()
	sql := fmt.Sprintf("DELETE FROM %s WHERE tenantId = ? AND metricServerId = ?", tableName)

	_, err := dao.db.Exec(ctx, sql, []interface{}{tenantId, metricServerId}, true)
	if err != nil {
		return fmt.Errorf("根据服务器ID删除进程统计日志失败: %w", err)
	}
	return nil
}
