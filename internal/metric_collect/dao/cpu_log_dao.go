package dao

import (
	"context"
	"fmt"
	"gohub/internal/metric_collect/types"
	"gohub/pkg/database"
	"time"
)

// CpuLogDAO CPU采集日志数据访问对象
type CpuLogDAO struct {
	db database.Database
}

// NewCpuLogDAO 创建CPU采集日志DAO实例
func NewCpuLogDAO(db database.Database) *CpuLogDAO {
	return &CpuLogDAO{db: db}
}

// InsertCpuLog 插入CPU采集日志
func (dao *CpuLogDAO) InsertCpuLog(ctx context.Context, cpuLog *types.CpuLog) error {
	// 设置通用字段默认值
	now := time.Now()
	cpuLog.AddTime = now
	cpuLog.EditTime = now
	cpuLog.CurrentVersion = 1
	cpuLog.ActiveFlag = types.ActiveFlagYes
	
	// 验证必填字段
	if cpuLog.MetricCpuLogId == "" {
		return fmt.Errorf("CPU采集日志ID不能为空")
	}
	if cpuLog.TenantId == "" {
		return fmt.Errorf("租户ID不能为空")
	}
	if cpuLog.MetricServerId == "" {
		return fmt.Errorf("服务器ID不能为空")
	}
	if cpuLog.AddWho == "" {
		return fmt.Errorf("创建人ID不能为空")
	}
	if cpuLog.EditWho == "" {
		return fmt.Errorf("修改人ID不能为空")
	}
	if cpuLog.OprSeqFlag == "" {
		return fmt.Errorf("操作序列标识不能为空")
	}
	
	tableName := cpuLog.TableName()
	_, err := dao.db.Insert(ctx, tableName, cpuLog, true)
	if err != nil {
		return fmt.Errorf("插入CPU采集日志失败: %w", err)
	}
	return nil
}

// BatchInsertCpuLog 批量插入CPU采集日志
func (dao *CpuLogDAO) BatchInsertCpuLog(ctx context.Context, cpuLogs []*types.CpuLog) error {
	if len(cpuLogs) == 0 {
		return nil
	}
	
	// 设置通用字段默认值
	now := time.Now()
	for _, cpuLog := range cpuLogs {
		cpuLog.AddTime = now
		cpuLog.EditTime = now
		cpuLog.CurrentVersion = 1
		cpuLog.ActiveFlag = types.ActiveFlagYes
		
		// 验证必填字段
		if cpuLog.MetricCpuLogId == "" {
			return fmt.Errorf("CPU采集日志ID不能为空")
		}
		if cpuLog.TenantId == "" {
			return fmt.Errorf("租户ID不能为空")
		}
		if cpuLog.MetricServerId == "" {
			return fmt.Errorf("服务器ID不能为空")
		}
		if cpuLog.AddWho == "" {
			return fmt.Errorf("创建人ID不能为空")
		}
		if cpuLog.EditWho == "" {
			return fmt.Errorf("修改人ID不能为空")
		}
		if cpuLog.OprSeqFlag == "" {
			return fmt.Errorf("操作序列标识不能为空")
		}
	}
	
	tableName := cpuLogs[0].TableName()
	
	// 转换为interface{}切片
	items := make([]interface{}, len(cpuLogs))
	for i, cpuLog := range cpuLogs {
		items[i] = cpuLog
	}
	
	_, err := dao.db.BatchInsert(ctx, tableName, items, true)
	if err != nil {
		return fmt.Errorf("批量插入CPU采集日志失败: %w", err)
	}
	return nil
}

// DeleteCpuLog 删除CPU采集日志（软删除）
func (dao *CpuLogDAO) DeleteCpuLog(ctx context.Context, tenantId, metricCpuLogId string) error {
	if tenantId == "" {
		return fmt.Errorf("租户ID不能为空")
	}
	if metricCpuLogId == "" {
		return fmt.Errorf("CPU采集日志ID不能为空")
	}
	
	tableName := (&types.CpuLog{}).TableName()
	sql := fmt.Sprintf("UPDATE %s SET activeFlag = ? WHERE tenantId = ? AND metricCpuLogId = ?", tableName)
	
	_, err := dao.db.Exec(ctx, sql, []interface{}{types.ActiveFlagNo, tenantId, metricCpuLogId}, true)
	if err != nil {
		return fmt.Errorf("删除CPU采集日志失败: %w", err)
	}
	return nil
}

// DeleteCpuLogByTime 根据时间范围删除CPU采集日志（物理删除）
func (dao *CpuLogDAO) DeleteCpuLogByTime(ctx context.Context, tenantId string, beforeTime time.Time) error {
	if tenantId == "" {
		return fmt.Errorf("租户ID不能为空")
	}
	
	tableName := (&types.CpuLog{}).TableName()
	sql := fmt.Sprintf("DELETE FROM %s WHERE tenantId = ? AND collectTime < ?", tableName)
	
	_, err := dao.db.Exec(ctx, sql, []interface{}{tenantId, beforeTime}, true)
	if err != nil {
		return fmt.Errorf("根据时间删除CPU采集日志失败: %w", err)
	}
	return nil
}

// DeleteCpuLogByServer 根据服务器ID删除CPU采集日志（物理删除）
func (dao *CpuLogDAO) DeleteCpuLogByServer(ctx context.Context, tenantId, metricServerId string) error {
	if tenantId == "" {
		return fmt.Errorf("租户ID不能为空")
	}
	if metricServerId == "" {
		return fmt.Errorf("服务器ID不能为空")
	}
	
	tableName := (&types.CpuLog{}).TableName()
	sql := fmt.Sprintf("DELETE FROM %s WHERE tenantId = ? AND metricServerId = ?", tableName)
	
	_, err := dao.db.Exec(ctx, sql, []interface{}{tenantId, metricServerId}, true)
	if err != nil {
		return fmt.Errorf("根据服务器ID删除CPU采集日志失败: %w", err)
	}
	return nil
} 