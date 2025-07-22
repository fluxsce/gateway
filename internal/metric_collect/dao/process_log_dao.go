package dao

import (
	"context"
	"fmt"
	"gohub/internal/metric_collect/types"
	"gohub/pkg/database"
	"time"
)

// ProcessLogDAO 进程信息日志数据访问对象
type ProcessLogDAO struct {
	db database.Database
}

// NewProcessLogDAO 创建进程信息日志DAO实例
func NewProcessLogDAO(db database.Database) *ProcessLogDAO {
	return &ProcessLogDAO{db: db}
}

// InsertProcessLog 插入进程信息日志
func (dao *ProcessLogDAO) InsertProcessLog(ctx context.Context, processLog *types.ProcessLog) error {
	// 设置通用字段默认值
	now := time.Now()
	processLog.AddTime = now
	processLog.EditTime = now
	processLog.CurrentVersion = 1
	processLog.ActiveFlag = types.ActiveFlagYes
	
	// 验证必填字段
	if processLog.MetricProcessLogId == "" {
		return fmt.Errorf("进程信息日志ID不能为空")
	}
	if processLog.TenantId == "" {
		return fmt.Errorf("租户ID不能为空")
	}
	if processLog.MetricServerId == "" {
		return fmt.Errorf("服务器ID不能为空")
	}
	if processLog.AddWho == "" {
		return fmt.Errorf("创建人ID不能为空")
	}
	if processLog.EditWho == "" {
		return fmt.Errorf("修改人ID不能为空")
	}
	if processLog.OprSeqFlag == "" {
		return fmt.Errorf("操作序列标识不能为空")
	}
	
	tableName := processLog.TableName()
	_, err := dao.db.Insert(ctx, tableName, processLog, true)
	if err != nil {
		return fmt.Errorf("插入进程信息日志失败: %w", err)
	}
	return nil
}

// BatchInsertProcessLog 批量插入进程信息日志
func (dao *ProcessLogDAO) BatchInsertProcessLog(ctx context.Context, processLogs []*types.ProcessLog) error {
	if len(processLogs) == 0 {
		return nil
	}
	
	// 设置通用字段默认值
	now := time.Now()
	for _, processLog := range processLogs {
		processLog.AddTime = now
		processLog.EditTime = now
		processLog.CurrentVersion = 1
		processLog.ActiveFlag = types.ActiveFlagYes
		
		// 验证必填字段
		if processLog.MetricProcessLogId == "" {
			return fmt.Errorf("进程信息日志ID不能为空")
		}
		if processLog.TenantId == "" {
			return fmt.Errorf("租户ID不能为空")
		}
		if processLog.MetricServerId == "" {
			return fmt.Errorf("服务器ID不能为空")
		}
		if processLog.AddWho == "" {
			return fmt.Errorf("创建人ID不能为空")
		}
		if processLog.EditWho == "" {
			return fmt.Errorf("修改人ID不能为空")
		}
		if processLog.OprSeqFlag == "" {
			return fmt.Errorf("操作序列标识不能为空")
		}
	}
	
	tableName := processLogs[0].TableName()
	
	// 转换为interface{}切片
	items := make([]interface{}, len(processLogs))
	for i, processLog := range processLogs {
		items[i] = processLog
	}
	
	_, err := dao.db.BatchInsert(ctx, tableName, items, true)
	if err != nil {
		return fmt.Errorf("批量插入进程信息日志失败: %w", err)
	}
	return nil
}

// DeleteProcessLog 删除进程信息日志（软删除）
func (dao *ProcessLogDAO) DeleteProcessLog(ctx context.Context, tenantId, metricProcessLogId string) error {
	if tenantId == "" {
		return fmt.Errorf("租户ID不能为空")
	}
	if metricProcessLogId == "" {
		return fmt.Errorf("进程信息日志ID不能为空")
	}
	
	tableName := (&types.ProcessLog{}).TableName()
	sql := fmt.Sprintf("UPDATE %s SET activeFlag = ? WHERE tenantId = ? AND metricProcessLogId = ?", tableName)
	
	_, err := dao.db.Exec(ctx, sql, []interface{}{types.ActiveFlagNo, tenantId, metricProcessLogId}, true)
	if err != nil {
		return fmt.Errorf("删除进程信息日志失败: %w", err)
	}
	return nil
}

// DeleteProcessLogByTime 根据时间范围删除进程信息日志（物理删除）
func (dao *ProcessLogDAO) DeleteProcessLogByTime(ctx context.Context, tenantId string, beforeTime time.Time) error {
	if tenantId == "" {
		return fmt.Errorf("租户ID不能为空")
	}
	
	tableName := (&types.ProcessLog{}).TableName()
	sql := fmt.Sprintf("DELETE FROM %s WHERE tenantId = ? AND collectTime < ?", tableName)
	
	_, err := dao.db.Exec(ctx, sql, []interface{}{tenantId, beforeTime}, true)
	if err != nil {
		return fmt.Errorf("根据时间删除进程信息日志失败: %w", err)
	}
	return nil
}

// DeleteProcessLogByServer 根据服务器ID删除进程信息日志（物理删除）
func (dao *ProcessLogDAO) DeleteProcessLogByServer(ctx context.Context, tenantId, metricServerId string) error {
	if tenantId == "" {
		return fmt.Errorf("租户ID不能为空")
	}
	if metricServerId == "" {
		return fmt.Errorf("服务器ID不能为空")
	}
	
	tableName := (&types.ProcessLog{}).TableName()
	sql := fmt.Sprintf("DELETE FROM %s WHERE tenantId = ? AND metricServerId = ?", tableName)
	
	_, err := dao.db.Exec(ctx, sql, []interface{}{tenantId, metricServerId}, true)
	if err != nil {
		return fmt.Errorf("根据服务器ID删除进程信息日志失败: %w", err)
	}
	return nil
} 